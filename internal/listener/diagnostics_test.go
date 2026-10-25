package listener

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnostics"
)

func diagnosticsListener(t *testing.T, producer runtimediagnostics.Producer, reporter func(error)) *DefaultListener {
	t.Helper()
	built, err := NewBootstrapWithHandshakeAndTerminalErrorReporter(http.NotFoundHandler(), reporter).
		WithDiagnosticsProducer(producer).Build(validListenerSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	return built.(*DefaultListener)
}

func diagnosticServe(listener *DefaultListener, err error) {
	listener.serveHTTP = func(*http.Server, net.Listener) error { return err }
	listener.wg.Add(1)
	listener.serve(nil, nil)
}

func assertListenerEvent(t *testing.T, reader *runtimediagnostics.Reader, sequence uint64) {
	t.Helper()
	event, ok := reader.TryRead()
	if !ok || event != (runtimediagnostics.Event{SchemaVersion: 1, Code: runtimediagnostics.ListenerServeFailed, Sequence: sequence}) {
		t.Fatalf("event = (%+v, %t), want Listener sequence %d", event, ok, sequence)
	}
}

func assertListenerMailboxEmpty(t *testing.T, reader *runtimediagnostics.Reader) {
	t.Helper()
	if event, ok := reader.TryRead(); ok || event != (runtimediagnostics.Event{}) {
		t.Fatalf("unexpected event: (%+v, %t)", event, ok)
	}
}

// I01: extend legacy direct-sentinel coverage to nil and wrapped sentinels.
func TestListenerDiagnosticsExpectedServeReturns(t *testing.T) {
	for index, cause := range []error{nil, http.ErrServerClosed, net.ErrClosed,
		fmt.Errorf("wrapped: %w", http.ErrServerClosed), fmt.Errorf("wrapped: %w", net.ErrClosed)} {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			producer, reader := runtimediagnostics.New()
			calls := 0
			listener := diagnosticsListener(t, *producer, func(error) { calls++ })
			diagnosticServe(listener, cause)
			if calls != 0 {
				t.Fatalf("legacy reports = %d", calls)
			}
			assertListenerMailboxEmpty(t, reader)
		})
	}
}

// I02: Stop does not reclassify a real unexpected Serve return.
func TestListenerDiagnosticsUnexpectedServeOnce(t *testing.T) {
	producer, reader := runtimediagnostics.New()
	cause := errors.New("sensitive Serve cause")
	reporter := &listenerErrorRecorder{}
	listener := diagnosticsListener(t, *producer, reporter.Report)
	listener.port = availableTCPPort(t)
	entered, closed, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	listener.serveHTTP = func(_ *http.Server, socket net.Listener) error {
		close(entered)
		_, err := socket.Accept()
		if !errors.Is(err, net.ErrClosed) {
			return err
		}
		close(closed)
		<-release
		return cause
	}
	if err := listener.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	waitListenerSignal(t, entered, "Serve entry")
	stopped := make(chan error, 1)
	go func() { stopped <- listener.Stop(context.Background()) }()
	waitListenerSignal(t, closed, "concurrent Stop close")
	close(release)
	if err := <-stopped; err != nil {
		t.Fatal(err)
	}
	if reported := reporter.Receive(t); !errors.Is(reported, cause) || reported.Error() != "HTTP Server Serve failed" {
		t.Fatalf("legacy cause/summary changed: %v", reported)
	}
	if reporter.Count() != 1 {
		t.Fatalf("legacy reports = %d", reporter.Count())
	}
	assertListenerEvent(t, reader, 1)
	assertListenerMailboxEmpty(t, reader)
}

// I03: consume the event while the synchronous legacy reporter is blocked.
// Release it before waiting for completion; this does not promise nonblocking Stop.
func TestListenerDiagnosticsBeforeLegacyReporter(t *testing.T) {
	producer, reader := runtimediagnostics.New()
	entered, release, completed := make(chan struct{}), make(chan struct{}), make(chan struct{})
	listener := diagnosticsListener(t, *producer, func(error) { close(entered); <-release })
	go func() { diagnosticServe(listener, errors.New("Serve failed")); close(completed) }()
	waitListenerSignal(t, entered, "legacy reporter")
	// Ensure a failed assertion cannot strand the legacy callback goroutine.
	defer func() { close(release); <-completed }()
	assertListenerEvent(t, reader, 1)
	assertListenerMailboxEmpty(t, reader)
}

func TestListenerDiagnosticsReporterPanic(t *testing.T) {
	producer, reader := runtimediagnostics.New()
	listener := diagnosticsListener(t, *producer, func(error) { panic("legacy reporter panic") })
	diagnosticServe(listener, errors.New("Serve failed"))
	listener.wg.Wait()
	assertListenerEvent(t, reader, 1)
	assertListenerMailboxEmpty(t, reader)
}

// I08: a full mailbox loses this occurrence without touching queued FIFO or legacy reporting.
func TestListenerDiagnosticsSaturationPreservesSourceBehavior(t *testing.T) {
	producer, reader := runtimediagnostics.New()
	for i := 0; i < 64; i++ {
		if !producer.Offer(runtimediagnostics.AuthenticationProviderFailed) {
			t.Fatal("sequential prefill failed")
		}
	}
	cause := errors.New("Serve failed")
	calls := 0
	listener := diagnosticsListener(t, *producer, func(err error) {
		calls++
		if !errors.Is(err, cause) {
			t.Error("legacy cause lost")
		}
	})
	diagnosticServe(listener, cause)
	for i := uint64(1); i <= 64; i++ {
		event, ok := reader.TryRead()
		if !ok || event.Sequence != i || event.Code != runtimediagnostics.AuthenticationProviderFailed {
			t.Fatalf("prefill FIFO changed: %+v, %t", event, ok)
		}
	}
	assertListenerMailboxEmpty(t, reader)
	if calls != 1 {
		t.Fatalf("legacy reports = %d", calls)
	}
	diagnosticServe(listener, cause)
	assertListenerEvent(t, reader, 65) // loss did not consume a sequence or schedule a retry
	zero := diagnosticsListener(t, runtimediagnostics.Producer{}, func(error) { calls++ })
	diagnosticServe(zero, cause)
	if calls != 3 {
		t.Fatalf("zero Producer changed legacy behavior: %d", calls)
	}
}

// I09: both Bootstrap and Build capture an opaque capability value.
func TestListenerDiagnosticsCaptureProducerValue(t *testing.T) {
	for _, replaceWithZero := range []bool{false, true} {
		t.Run(fmt.Sprint(replaceWithZero), func(t *testing.T) {
			first, reader1 := runtimediagnostics.New()
			second, reader2 := runtimediagnostics.New()
			caller := *first
			original := NewBootstrapWithHandshake(http.NotFoundHandler())
			bound := original.WithDiagnosticsProducer(caller)
			caller = *second
			if replaceWithZero {
				caller = runtimediagnostics.Producer{}
			}
			built, err := bound.Build(validListenerSnapshot())
			if err != nil {
				t.Fatal(err)
			}
			bound = bound.WithDiagnosticsProducer(caller)
			diagnosticServe(built.(*DefaultListener), errors.New("first"))
			assertListenerEvent(t, reader1, 1)
			assertListenerMailboxEmpty(t, reader1)
			assertListenerMailboxEmpty(t, reader2)
			unbound, err := original.Build(validListenerSnapshot())
			if err != nil {
				t.Fatal(err)
			}
			diagnosticServe(unbound.(*DefaultListener), errors.New("disabled original"))
			assertListenerMailboxEmpty(t, reader1)
			assertListenerMailboxEmpty(t, reader2)
		})
	}
}

// I10: integration operations remain safe under concurrent producers/readers.
// TryLock loss is allowed, so acceptance is bounded rather than required for every source.
func TestListenerDiagnosticsConcurrentSourcesAndReaders(t *testing.T) {
	producer, reader := runtimediagnostics.New()
	_, isolated := runtimediagnostics.New()
	const occurrences = 128
	listeners := make([]*DefaultListener, occurrences)
	for i := range listeners {
		listeners[i] = diagnosticsListener(t, *producer, nil)
	}
	var done atomic.Bool
	var readers sync.WaitGroup
	observed := make(chan []runtimediagnostics.Event, 2)
	for range 2 {
		readers.Add(1)
		go func() {
			defer readers.Done()
			var events []runtimediagnostics.Event
			for {
				if event, ok := reader.TryRead(); ok {
					events = append(events, event)
				} else if done.Load() {
					observed <- events
					return
				} else {
					runtime.Gosched()
				}
			}
		}()
	}
	var writers sync.WaitGroup
	for _, listener := range listeners {
		writers.Add(1)
		go func() { defer writers.Done(); diagnosticServe(listener, errors.New("Serve failed")) }()
	}
	writers.Wait()
	done.Store(true)
	readers.Wait()
	close(observed)
	seen := make(map[uint64]bool)
	check := func(event runtimediagnostics.Event) {
		if event.SchemaVersion != 1 || event.Code != runtimediagnostics.ListenerServeFailed || event.Sequence == 0 || seen[event.Sequence] {
			t.Fatalf("invalid or duplicate accepted event: %+v", event)
		}
		seen[event.Sequence] = true
	}
	for events := range observed {
		var previous uint64
		for _, event := range events {
			if event.Sequence <= previous {
				t.Fatal("per-reader FIFO changed")
			}
			previous = event.Sequence
			check(event)
		}
	}
	for {
		event, ok := reader.TryRead()
		if !ok {
			break
		}
		check(event)
	}
	if len(seen) > occurrences {
		t.Fatalf("accepted %d > source occurrences %d", len(seen), occurrences)
	}
	assertListenerMailboxEmpty(t, isolated)
}

// I11: construction, Listen and Stop failures have no new outcome code.
func TestListenerDiagnosticsDoNotAddOutcomeKinds(t *testing.T) {
	producer, reader := runtimediagnostics.New()
	bootstrap := NewBootstrapWithHandshake(http.NotFoundHandler()).WithDiagnosticsProducer(*producer)
	bad := validListenerSnapshot()
	bad.Host = ""
	if _, err := bootstrap.Build(bad); !errors.Is(err, ErrInvalidListenerConfiguration) {
		t.Fatalf("Build error = %v", err)
	}
	listener := diagnosticsListener(t, *producer, nil)
	if err := listener.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	listener.port = uint16(occupied.Addr().(*net.TCPAddr).Port)
	if err := listener.Start(context.Background()); err == nil {
		t.Fatal("Listen unexpectedly succeeded on occupied port")
	}
	assertListenerMailboxEmpty(t, reader)
	listener.port = availableTCPPort(t)
	listener.serveHTTP = func(*http.Server, net.Listener) error { return http.ErrServerClosed }
	stopCause := errors.New("shutdown failed")
	listener.shutdownHTTP = func(*http.Server, context.Context) error { return stopCause }
	if err := listener.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := listener.Stop(context.Background()); !errors.Is(err, stopCause) {
		t.Fatalf("Stop cause lost: %v", err)
	}
	assertListenerMailboxEmpty(t, reader)
}
