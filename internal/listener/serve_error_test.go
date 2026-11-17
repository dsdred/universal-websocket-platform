package listener

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	platformconnection "github.com/dsdred/universal-websocket-platform/internal/connection"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnostics"
)

func TestListenerExpectedServeShutdownIsNotReported(t *testing.T) {
	reporter := &listenerErrorRecorder{}
	runtimeListener := listenerWithTerminalErrorReporter(t, reporter.Report)
	if err := runtimeListener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := runtimeListener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if reporter.Count() != 0 {
		t.Fatalf("reported errors = %d, want 0", reporter.Count())
	}
}

func TestListenerExpectedClosedServeErrorsAreNotReported(t *testing.T) {
	for _, expected := range []error{http.ErrServerClosed, net.ErrClosed} {
		reporter := &listenerErrorRecorder{}
		runtimeListener := listenerWithTerminalErrorReporter(t, reporter.Report)
		runtimeListener.serveHTTP = func(*http.Server, net.Listener) error { return expected }
		if err := runtimeListener.Start(context.Background()); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		if err := runtimeListener.Stop(context.Background()); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
		if reporter.Count() != 0 {
			t.Fatalf("Serve error %v was reported", expected)
		}
	}
}

func TestListenerUnexpectedServeErrorIsReportedOnce(t *testing.T) {
	wantErr := errors.New("serve failure with credential-that-must-not-leak")
	reporter := &listenerErrorRecorder{}
	serveEntered := make(chan struct{})
	releaseServe := make(chan struct{})
	runtimeListener := listenerWithTerminalErrorReporter(t, reporter.Report)
	runtimeListener.serveHTTP = func(*http.Server, net.Listener) error {
		close(serveEntered)
		<-releaseServe
		return wantErr
	}
	if err := runtimeListener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	waitListenerSignal(t, serveEntered, "Serve entry")
	close(releaseServe)

	reported := reporter.Receive(t)
	if !errors.Is(reported, wantErr) {
		t.Fatalf("reported error = %v, want Serve sentinel", reported)
	}
	if strings.Contains(reported.Error(), "credential-that-must-not-leak") {
		t.Fatal("reported Serve error exposes credentials")
	}
	if err := runtimeListener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if reporter.Count() != 1 {
		t.Fatalf("reported errors = %d, want 1", reporter.Count())
	}
}

func TestListenerConcurrentStopAndServeFailureReportsOnce(t *testing.T) {
	wantErr := errors.New("unexpected Serve failure")
	reporter := &listenerErrorRecorder{}
	serveEntered := make(chan struct{})
	listenerClosed := make(chan struct{})
	releaseFailure := make(chan struct{})
	runtimeListener := listenerWithTerminalErrorReporter(t, reporter.Report)
	view := ServeCompletion{returned: runtimeListener.serveReturned}
	runtimeListener.serveHTTP = func(_ *http.Server, tcpListener net.Listener) error {
		close(serveEntered)
		_, acceptErr := tcpListener.Accept()
		if !errors.Is(acceptErr, net.ErrClosed) {
			return acceptErr
		}
		close(listenerClosed)
		<-releaseFailure
		return wantErr
	}
	if err := runtimeListener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	waitListenerSignal(t, serveEntered, "Serve entry")
	assertServeCompletion(t, view, false, true)

	stopResult := make(chan error, 1)
	go func() { stopResult <- runtimeListener.Stop(context.Background()) }()
	waitListenerSignal(t, listenerClosed, "Listener close during Stop")
	assertServeCompletion(t, view, false, true)
	close(releaseFailure)

	reported := reporter.Receive(t)
	assertServeCompletion(t, view, true, true)
	if !errors.Is(reported, wantErr) {
		t.Fatalf("reported error = %v, want Serve sentinel", reported)
	}
	select {
	case err := <-stopResult:
		if err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() deadlocked with Serve failure")
	}
	if reporter.Count() != 1 {
		t.Fatalf("reported errors = %d, want 1", reporter.Count())
	}
}

func TestListenerServeErrorReporterPanicDoesNotChangeLifecycle(t *testing.T) {
	serveEntered := make(chan struct{})
	releaseServe := make(chan struct{})
	runtimeListener := listenerWithTerminalErrorReporter(t, func(error) { panic("reporter failed") })
	runtimeListener.serveHTTP = func(*http.Server, net.Listener) error {
		close(serveEntered)
		<-releaseServe
		return errors.New("Serve failed")
	}
	if err := runtimeListener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	waitListenerSignal(t, serveEntered, "Serve entry")
	close(releaseServe)

	stopContext, cancelStop := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelStop()
	if err := runtimeListener.Stop(stopContext); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if runtimeListener.Running() {
		t.Fatal("Listener remains Running after reporter panic")
	}
}

func listenerWithTerminalErrorReporter(t *testing.T, reportError func(error)) *DefaultListener {
	t.Helper()
	snapshot := validListenerSnapshot()
	snapshot.Port = availableTCPPort(t)
	built, err := NewBootstrapWithHandshakeAndTerminalErrorReporter(
		testWebSocketHandler{dispatcher: platformconnection.DefaultDispatcher{}},
		reportError,
	).Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	runtimeListener := built.(*DefaultListener)
	t.Cleanup(func() { _ = runtimeListener.Stop(context.Background()) })
	return runtimeListener
}

type listenerErrorRecorder struct {
	mu     sync.Mutex
	errors []error
	notify chan struct{}
}

func (recorder *listenerErrorRecorder) Report(err error) {
	recorder.mu.Lock()
	recorder.errors = append(recorder.errors, err)
	if recorder.notify != nil {
		close(recorder.notify)
		recorder.notify = nil
	}
	recorder.mu.Unlock()
}

func (recorder *listenerErrorRecorder) Count() int {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return len(recorder.errors)
}

func (recorder *listenerErrorRecorder) Receive(t *testing.T) error {
	t.Helper()
	recorder.mu.Lock()
	if len(recorder.errors) > 0 {
		err := recorder.errors[0]
		recorder.mu.Unlock()
		return err
	}
	recorder.notify = make(chan struct{})
	notify := recorder.notify
	recorder.mu.Unlock()

	waitListenerSignal(t, notify, "terminal error report")
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return recorder.errors[0]
}

func waitListenerSignal(t *testing.T, signal <-chan struct{}, operation string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for %s", operation)
	}
}

func TestServeCompletionRecordsEveryReturn(t *testing.T) {
	for name, cause := range map[string]error{
		"nil":                     nil,
		"unexpected":              errors.New("private Serve failure"),
		"server closed":           http.ErrServerClosed,
		"listener closed":         net.ErrClosed,
		"wrapped server closed":   fmt.Errorf("closed: %w", http.ErrServerClosed),
		"wrapped listener closed": fmt.Errorf("closed: %w", net.ErrClosed),
	} {
		t.Run(name, func(t *testing.T) {
			built, view, err := NewBootstrapWithHandshake(http.NotFoundHandler()).BuildWithServeCompletion(validListenerSnapshot())
			if err != nil {
				t.Fatal(err)
			}
			assertServeCompletion(t, view, false, true)
			copied := view
			runtimeListener := built.(*DefaultListener)
			diagnosticServe(runtimeListener, cause)
			for range 3 {
				assertServeCompletion(t, view, true, true)
				assertServeCompletion(t, copied, true, true)
			}
			if runtimeListener.Running() {
				t.Fatal("completion changed Listener lifecycle")
			}
		})
	}
}

func TestServeCompletionPrecedesErrorClassification(t *testing.T) {
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	releaseAll := func() { releaseOnce.Do(func() { close(release) }) }
	built, view, err := NewBootstrapWithHandshake(http.NotFoundHandler()).BuildWithServeCompletion(validListenerSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	runtimeListener := built.(*DefaultListener)
	cause := &blockingServeClassification{entered: entered, release: release}
	t.Cleanup(func() { releaseAll(); waitListenerSignal(t, done, "classification source exit") })
	go func() { diagnosticServe(runtimeListener, cause); close(done) }()
	waitListenerSignal(t, entered, "error classification")
	assertServeCompletion(t, view, true, true)
	select {
	case <-done:
		t.Fatal("classification blocker was bypassed")
	default:
	}
	releaseAll()
	waitListenerSignal(t, done, "source return after classification")
}

type blockingServeClassification struct {
	entered chan struct{}
	release <-chan struct{}
	once    sync.Once
}

func (*blockingServeClassification) Error() string { return "blocked classification" }
func (cause *blockingServeClassification) Is(error) bool {
	cause.once.Do(func() { close(cause.entered) })
	<-cause.release
	return false
}

func TestServeCompletionSurvivesBlockedReporterAndConcurrentStop(t *testing.T) {
	producer, reader := runtimediagnostics.New()
	for range 64 {
		if !producer.Offer(runtimediagnostics.AuthenticationProviderFailed) {
			t.Fatal("mailbox prefill failed")
		}
	}
	reporterEntered, releaseReporter := make(chan struct{}), make(chan struct{})
	closedSocket := make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseReporter) }) }
	snapshot := validListenerSnapshot()
	snapshot.Port = availableTCPPort(t)
	built, view, err := NewBootstrapWithHandshakeAndTerminalErrorReporter(http.NotFoundHandler(), func(error) {
		close(reporterEntered)
		<-releaseReporter
	}).WithDiagnosticsProducer(*producer).BuildWithServeCompletion(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	runtimeListener := built.(*DefaultListener)
	t.Cleanup(func() { release(); _ = runtimeListener.Stop(context.Background()) })
	runtimeListener.serveHTTP = func(*http.Server, net.Listener) error { return errors.New("Serve failed") }
	runtimeListener.closeTCP = func(socket net.Listener) error {
		err := socket.Close()
		close(closedSocket)
		return err
	}
	if err := runtimeListener.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	waitListenerSignal(t, reporterEntered, "blocked reporter")
	assertServeCompletion(t, view, true, true)
	for range 64 {
		event, ok := reader.TryRead()
		if !ok || event.Code != runtimediagnostics.AuthenticationProviderFailed {
			t.Fatal("lossy mailbox contents changed")
		}
	}
	assertListenerMailboxEmpty(t, reader)
	assertServeCompletion(t, view, true, true)

	stopResult := make(chan error, 1)
	go func() { stopResult <- runtimeListener.Stop(context.Background()) }()
	waitListenerSignal(t, closedSocket, "Stop socket close")
	select {
	case err := <-stopResult:
		t.Fatalf("Stop bypassed blocked legacy reporter: %v", err)
	default:
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := runtimeListener.Stop(canceled); !errors.Is(err, context.Canceled) {
		t.Fatalf("secondary Stop = %v", err)
	}
	var readers sync.WaitGroup
	var invalid atomic.Int32
	for range 4 {
		readers.Add(1)
		go func(copy ServeCompletion) {
			defer readers.Done()
			for range 1000 {
				if returned, available := copy.Observe(); !returned || !available {
					invalid.Add(1)
				}
			}
		}(view)
	}
	readers.Wait()
	if invalid.Load() != 0 {
		t.Fatal("concurrent observations changed retained evidence")
	}
	release()
	select {
	case err := <-stopResult:
		if err != nil {
			t.Fatalf("Stop = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not finish after reporter release")
	}
	assertServeCompletion(t, view, true, true)
	if runtimeListener.Running() {
		t.Fatal("Listener still running after Stop")
	}
}

func TestServeCompletionReporterPanicRetainsEvidence(t *testing.T) {
	built, view, err := NewBootstrapWithHandshakeAndTerminalErrorReporter(http.NotFoundHandler(), func(error) {
		panic("reporter panic")
	}).BuildWithServeCompletion(validListenerSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	runtimeListener := built.(*DefaultListener)
	diagnosticServe(runtimeListener, errors.New("Serve failed"))
	runtimeListener.wg.Wait()
	assertServeCompletion(t, view, true, true)
}

func TestServeCompletionUnderlyingPanicIsNotReturn(t *testing.T) {
	built, view, err := NewBootstrapWithHandshake(http.NotFoundHandler()).BuildWithServeCompletion(validListenerSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	runtimeListener := built.(*DefaultListener)
	runtimeListener.serveHTTP = func(*http.Server, net.Listener) error { panic("underlying Serve panic") }
	runtimeListener.wg.Add(1)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("underlying panic was swallowed")
			}
		}()
		runtimeListener.serve(nil, nil)
	}()
	runtimeListener.wg.Wait()
	assertServeCompletion(t, view, false, true)
}

func TestServeCompletionLegacyListenerHasNoCoverage(t *testing.T) {
	runtimeListener := &DefaultListener{serveHTTP: func(*http.Server, net.Listener) error { return nil }}
	runtimeListener.wg.Add(1)
	runtimeListener.serve(nil, nil)
	runtimeListener.wg.Wait()
	assertServeCompletion(t, ServeCompletion{}, false, false)
}

func assertServeCompletion(t *testing.T, view ServeCompletion, wantReturned, wantAvailable bool) {
	t.Helper()
	if returned, available := view.Observe(); returned != wantReturned || available != wantAvailable {
		t.Fatalf("Observe() = (%t, %t), want (%t, %t)", returned, available, wantReturned, wantAvailable)
	}
}
