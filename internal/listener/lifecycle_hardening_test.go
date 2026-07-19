package listener

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	platformconnection "github.com/dsdred/universal-websocket-platform/internal/connection"
)

func TestListenerConcurrentStopWaitsForAndSharesOneShutdown(t *testing.T) {
	runtimeListener := newLifecycleTestListener(t)
	if err := runtimeListener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	wantShutdownErr := errors.New("shutdown failed")
	wantCloseErr := errors.New("close failed")
	shutdownEntered := make(chan struct{})
	releaseShutdown := make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(releaseShutdown) }) })
	var shutdownCalls atomic.Int32
	var closeCalls atomic.Int32
	runtimeListener.shutdownHTTP = func(*http.Server, context.Context) error {
		shutdownCalls.Add(1)
		close(shutdownEntered)
		<-releaseShutdown
		return wantShutdownErr
	}
	runtimeListener.closeTCP = func(tcpListener net.Listener) error {
		closeCalls.Add(1)
		_ = tcpListener.Close()
		return wantCloseErr
	}

	primaryResult := make(chan error, 1)
	go func() { primaryResult <- runtimeListener.Stop(context.Background()) }()
	waitListenerSignal(t, shutdownEntered, "primary shutdown entry")

	secondaryContext := newObservedContext(context.Background())
	secondaryResult := make(chan error, 1)
	go func() { secondaryResult <- runtimeListener.Stop(secondaryContext) }()
	waitListenerSignal(t, secondaryContext.doneObserved, "secondary Stop wait entry")
	select {
	case err := <-secondaryResult:
		t.Fatalf("secondary Stop() returned before primary shutdown completed: %v", err)
	default:
	}

	releaseOnce.Do(func() { close(releaseShutdown) })
	for caller, result := range []<-chan error{primaryResult, secondaryResult} {
		err := waitListenerResult(t, result, "concurrent Stop")
		if !errors.Is(err, wantShutdownErr) || !errors.Is(err, wantCloseErr) {
			t.Fatalf("Stop caller %d error = %v, want both shutdown and close errors", caller, err)
		}
	}
	if got := shutdownCalls.Load(); got != 1 {
		t.Fatalf("Shutdown calls = %d, want 1", got)
	}
	if got := closeCalls.Load(); got != 1 {
		t.Fatalf("Close calls = %d, want 1", got)
	}

	repeatedErr := runtimeListener.Stop(context.Background())
	if !errors.Is(repeatedErr, wantShutdownErr) || !errors.Is(repeatedErr, wantCloseErr) {
		t.Fatalf("repeated Stop() error = %v, want stored combined error", repeatedErr)
	}
	if err := runtimeListener.Start(context.Background()); !errors.Is(err, ErrListenerAlreadyRunning) {
		t.Fatalf("Start() after Stop error = %v, want ErrListenerAlreadyRunning", err)
	}
}

func TestListenerConcurrentStopWaitCancellationDoesNotAlterShutdown(t *testing.T) {
	runtimeListener := newLifecycleTestListener(t)
	if err := runtimeListener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	wantErr := errors.New("shutdown failed")
	shutdownEntered := make(chan struct{})
	releaseShutdown := make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(releaseShutdown) }) })
	runtimeListener.shutdownHTTP = func(*http.Server, context.Context) error {
		close(shutdownEntered)
		<-releaseShutdown
		return wantErr
	}
	runtimeListener.closeTCP = func(tcpListener net.Listener) error {
		return tcpListener.Close()
	}

	primaryResult := make(chan error, 1)
	go func() { primaryResult <- runtimeListener.Stop(context.Background()) }()
	waitListenerSignal(t, shutdownEntered, "primary shutdown entry")

	waitContext, cancelWait := context.WithCancel(context.Background())
	cancelWait()
	if err := runtimeListener.Stop(waitContext); !errors.Is(err, context.Canceled) {
		t.Fatalf("secondary Stop() error = %v, want context.Canceled", err)
	}

	releaseOnce.Do(func() { close(releaseShutdown) })
	if err := waitListenerResult(t, primaryResult, "primary Stop"); !errors.Is(err, wantErr) {
		t.Fatalf("primary Stop() error = %v, want shutdown error", err)
	}
	if err := runtimeListener.Stop(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("repeated Stop() error = %v, want stored shutdown error", err)
	}
}

func TestListenerStopPreservesIndependentErrors(t *testing.T) {
	tests := []struct {
		name        string
		shutdownErr error
		closeErr    error
	}{
		{name: "shutdown error", shutdownErr: errors.New("shutdown failed")},
		{name: "close error", closeErr: errors.New("close failed")},
		{name: "both errors", shutdownErr: errors.New("shutdown failed"), closeErr: errors.New("close failed")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runtimeListener := newLifecycleTestListener(t)
			if err := runtimeListener.Start(context.Background()); err != nil {
				t.Fatalf("Start() error = %v", err)
			}
			runtimeListener.shutdownHTTP = func(*http.Server, context.Context) error {
				return test.shutdownErr
			}
			runtimeListener.closeTCP = func(tcpListener net.Listener) error {
				_ = tcpListener.Close()
				return test.closeErr
			}

			err := runtimeListener.Stop(context.Background())
			if test.shutdownErr != nil && !errors.Is(err, test.shutdownErr) {
				t.Fatalf("Stop() error = %v, want shutdown error", err)
			}
			if test.closeErr != nil && !errors.Is(err, test.closeErr) {
				t.Fatalf("Stop() error = %v, want close error", err)
			}
			if err != runtimeListener.Stop(context.Background()) {
				t.Fatal("repeated Stop() did not return the stored terminal error")
			}
			assertTCPPortAvailable(t, runtimeListener.Address())
		})
	}
}

func TestListenerCanceledShutdownCompletesAndStoresResult(t *testing.T) {
	runtimeListener := newLifecycleTestListener(t)
	if err := runtimeListener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	runtimeListener.shutdownHTTP = func(*http.Server, context.Context) error {
		return context.Canceled
	}
	runtimeListener.closeTCP = func(tcpListener net.Listener) error {
		return tcpListener.Close()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := runtimeListener.Stop(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Stop() error = %v, want context.Canceled", err)
	}
	if repeatedErr := runtimeListener.Stop(context.Background()); repeatedErr != err {
		t.Fatalf("repeated Stop() error = %v, want stored %v", repeatedErr, err)
	}
	assertTCPPortAvailable(t, runtimeListener.Address())
}

func newLifecycleTestListener(t *testing.T) *DefaultListener {
	t.Helper()
	snapshot := validListenerSnapshot()
	snapshot.Port = availableTCPPort(t)
	built, err := NewBootstrapWithHandshake(testWebSocketHandler{
		dispatcher: platformconnection.DefaultDispatcher{},
	}).Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	runtimeListener := built.(*DefaultListener)
	t.Cleanup(func() { _ = runtimeListener.Stop(context.Background()) })
	return runtimeListener
}

func waitListenerResult(t *testing.T, result <-chan error, operation string) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for %s", operation)
		return nil
	}
}

type observedContext struct {
	context.Context
	doneObserved chan struct{}
	once         sync.Once
}

func newObservedContext(ctx context.Context) *observedContext {
	return &observedContext{Context: ctx, doneObserved: make(chan struct{})}
}

func (ctx *observedContext) Done() <-chan struct{} {
	ctx.once.Do(func() { close(ctx.doneObserved) })
	return ctx.Context.Done()
}
