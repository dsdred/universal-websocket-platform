package runtime

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestHostRuntimeContextLifecycle(t *testing.T) {
	runtimeListener := newControlledListener(nil, false)
	host := newTestHost(t, fixedComposer(runtimeListener))
	if host.Ready() {
		t.Fatal("Ready() = true before Build")
	}
	if runtimeContext := host.RuntimeContext(); runtimeContext != nil {
		t.Fatal("RuntimeContext() before Start is not nil")
	}
	if host.runtimeCancel != nil {
		t.Fatal("runtime CancelFunc before Build is not nil")
	}
	if err := host.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if runtimeContext := host.RuntimeContext(); runtimeContext != nil {
		t.Fatal("RuntimeContext() after Build is not nil")
	}
	if host.runtimeCancel != nil {
		t.Fatal("runtime CancelFunc after Build is not nil")
	}
	if host.Ready() {
		t.Fatal("Ready() = true after Build")
	}
	runtimeListener.startObserver = func() {
		if !runtimeListener.Running() {
			t.Error("Listener is not Running immediately before Start returns")
		}
		if host.RuntimeContext() != nil {
			t.Error("RuntimeContext was published before Listener.Start returned successfully")
		}
		if host.Ready() {
			t.Error("Ready() became true before Listener.Start returned successfully")
		}
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	runtimeContext := host.RuntimeContext()
	if runtimeContext == nil {
		t.Fatal("RuntimeContext() after Start is nil")
	}
	assertContextActive(t, runtimeContext, "Runtime context after Start")
	if !host.Running() || !host.Ready() || !runtimeListener.Running() {
		t.Fatal("successful Start did not atomically publish Running readiness")
	}
	if host.runtimeListener != runtimeListener {
		t.Fatal("Ready Host does not expose the committed Listener internally")
	}

	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	assertContextCanceled(t, runtimeContext, "Runtime context after Stop")
	if host.RuntimeContext() != runtimeContext {
		t.Fatal("RuntimeContext() after Stop did not return the original canceled context")
	}
	if host.Ready() || host.Running() {
		t.Fatal("Host remains Ready or Running after Stop")
	}
}

func TestHostRuntimeContextDoesNotUseStartupContext(t *testing.T) {
	runtimeListener := newControlledListener(nil, false)
	host := newTestHost(t, fixedComposer(runtimeListener))
	if err := host.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	startupContext, cancelStartup := context.WithCancel(context.Background())
	if err := host.Start(startupContext); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	runtimeContext := host.RuntimeContext()
	if runtimeContext == nil {
		t.Fatal("RuntimeContext() after Start is nil")
	}

	cancelStartup()
	assertContextActive(t, runtimeContext, "Runtime context after startup cancellation")
	if runtimeContext == startupContext {
		t.Fatal("Runtime context is the startup context")
	}
	if !host.Running() {
		t.Fatal("Host is not Running after startup context cancellation")
	}
	if !runtimeListener.Running() {
		t.Fatal("Listener stopped after startup context cancellation")
	}
	if got := runtimeListener.stopCalls.Load(); got != 0 {
		t.Fatalf("Listener.Stop calls after startup context cancellation = %d, want 0", got)
	}

	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	assertContextCanceled(t, runtimeContext, "Runtime context after Stop")
}

func TestHostRepeatedStopKeepsRuntimeContextCanceled(t *testing.T) {
	listenerStopErr := errors.New("listener stop failed")
	runtimeListener := newControlledListener(nil, false)
	runtimeListener.stopErr = listenerStopErr
	host := newTestHost(t, fixedComposer(runtimeListener))
	var contextCreations atomic.Int32
	var contextCancellations atomic.Int32
	host.newRuntimeContext = trackingRuntimeContextFactory(&contextCreations, &contextCancellations)
	if err := host.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	runtimeContext := host.RuntimeContext()

	if err := host.Stop(context.Background()); !errors.Is(err, listenerStopErr) {
		t.Fatalf("first Stop() error = %v, want Listener.Stop error", err)
	}
	if err := host.Stop(context.Background()); !errors.Is(err, listenerStopErr) {
		t.Fatalf("second Stop() error = %v, want same terminal error", err)
	}
	assertContextCanceled(t, runtimeContext, "Runtime context after repeated Stop")
	if host.Ready() {
		t.Fatal("Ready() = true after repeated Stop")
	}
	if host.RuntimeContext() != runtimeContext {
		t.Fatal("repeated Stop replaced RuntimeContext")
	}
	if got := contextCreations.Load(); got != 1 {
		t.Fatalf("Runtime context creations = %d, want 1", got)
	}
	if got := contextCancellations.Load(); got != 1 {
		t.Fatalf("Runtime context cancellations = %d, want 1", got)
	}
	if got := runtimeListener.stopCalls.Load(); got != 1 {
		t.Fatalf("Listener.Stop calls = %d, want 1", got)
	}
}

func trackingRuntimeContextFactory(
	creations *atomic.Int32,
	cancellations *atomic.Int32,
) runtimeContextFactory {
	return func() (context.Context, context.CancelFunc) {
		creations.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		return ctx, func() {
			cancellations.Add(1)
			cancel()
		}
	}
}

func assertContextActive(t *testing.T, ctx context.Context, operation string) {
	t.Helper()
	select {
	case <-ctx.Done():
		t.Fatalf("%s is canceled: %v", operation, ctx.Err())
	default:
	}
}

func assertContextCanceled(t *testing.T, ctx context.Context, operation string) {
	t.Helper()
	select {
	case <-ctx.Done():
		if ctx.Err() != context.Canceled {
			t.Fatalf("%s error = %v, want context.Canceled", operation, ctx.Err())
		}
	case <-time.After(time.Second):
		t.Fatalf("%s was not canceled", operation)
	}
}
