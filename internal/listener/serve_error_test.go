package listener

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	platformconnection "github.com/dsdred/universal-websocket-platform/internal/connection"
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

	stopResult := make(chan error, 1)
	go func() { stopResult <- runtimeListener.Stop(context.Background()) }()
	waitListenerSignal(t, listenerClosed, "Listener close during Stop")
	close(releaseFailure)

	reported := reporter.Receive(t)
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
