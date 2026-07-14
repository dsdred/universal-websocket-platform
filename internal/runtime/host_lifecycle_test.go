package runtime

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/listener"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

func TestHostSecondBuildDoesNotRecomposeDependencies(t *testing.T) {
	runtimeListener := newControlledListener(nil, false)
	var listenerCreations atomic.Int32
	var factoryCreations atomic.Int32
	composer := func(
		runtimeconfig.Snapshot,
		secretresolver.Resolver,
		message.Handler,
	) (listener.Listener, error) {
		factoryCreations.Add(1)
		listenerCreations.Add(1)
		return runtimeListener, nil
	}
	host := newTestHost(t, composer)

	if err := host.Build(); err != nil {
		t.Fatalf("first Build() error = %v", err)
	}
	firstListener := host.runtimeListener
	if err := host.Build(); !errors.Is(err, ErrHostAlreadyBuilt) {
		t.Fatalf("second Build() error = %v, want ErrHostAlreadyBuilt", err)
	}

	if got := factoryCreations.Load(); got != 1 {
		t.Fatalf("Factory creations = %d, want 1", got)
	}
	if got := listenerCreations.Load(); got != 1 {
		t.Fatalf("Listener creations = %d, want 1", got)
	}
	if host.runtimeListener != firstListener {
		t.Fatal("second Build() replaced the composed Listener")
	}
}

func TestHostListenerStartErrorRestoresBuiltState(t *testing.T) {
	listenerStartErr := errors.New("listener start failed")
	runtimeListener := newControlledListener(listenerStartErr, false)
	host := newTestHost(t, fixedComposer(runtimeListener))
	if err := host.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	err := host.Start(context.Background())
	if !errors.Is(err, listenerStartErr) {
		t.Fatalf("Start() error = %v, want listener error", err)
	}
	if host.Running() {
		t.Fatal("Running() = true after Listener.Start error")
	}
	if got := currentHostState(host); got != hostBuilt {
		t.Fatalf("state = %v, want hostBuilt", got)
	}

	stopResult := make(chan error, 1)
	go func() {
		stopResult <- host.Stop(context.Background())
	}()
	select {
	case err := <-stopResult:
		if err != nil {
			t.Fatalf("Stop() after Start error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Stop() after Start error deadlocked")
	}
	if got := runtimeListener.stopCalls.Load(); got != 0 {
		t.Fatalf("Listener.Stop calls after failed Start = %d, want 0", got)
	}
}

func TestHostStopDuringStarting(t *testing.T) {
	const iterations = 100
	for iteration := range iterations {
		runtimeListener := newControlledListener(nil, true)
		host := newTestHost(t, fixedComposer(runtimeListener))
		if err := host.Build(); err != nil {
			t.Fatalf("iteration %d: Build() error = %v", iteration, err)
		}

		startResult := make(chan error, 1)
		go func() {
			startResult <- host.Start(context.Background())
		}()
		waitForSignal(t, runtimeListener.startEntered, "Listener.Start entry")

		stopResult := make(chan error, 1)
		go func() {
			stopResult <- host.Stop(context.Background())
		}()
		waitForHostState(t, host, hostStopping)
		if host.Running() {
			t.Fatalf("iteration %d: Host became Running while Stop was pending", iteration)
		}
		select {
		case err := <-stopResult:
			t.Fatalf("iteration %d: Stop() returned before Listener.Start completed: %v", iteration, err)
		default:
		}

		close(runtimeListener.releaseStart)
		waitForResult(t, startResult, "Start")
		waitForResult(t, stopResult, "Stop")

		if host.Running() {
			t.Fatalf("iteration %d: Running() = true after Stop", iteration)
		}
		if got := currentHostState(host); got != hostStopped {
			t.Fatalf("iteration %d: state = %v, want hostStopped", iteration, got)
		}
		if got := runtimeListener.stopCalls.Load(); got != 1 {
			t.Fatalf("iteration %d: Listener.Stop calls = %d, want 1", iteration, got)
		}
	}
}

type controlledListener struct {
	startEntered chan struct{}
	releaseStart chan struct{}
	startErr     error
	startOnce    sync.Once
	mu           sync.RWMutex
	running      bool
	stopCalls    atomic.Int32
}

func newControlledListener(startErr error, blockStart bool) *controlledListener {
	runtimeListener := &controlledListener{
		startEntered: make(chan struct{}),
		startErr:     startErr,
	}
	if blockStart {
		runtimeListener.releaseStart = make(chan struct{})
	}
	return runtimeListener
}

func (runtimeListener *controlledListener) Address() string {
	return "127.0.0.1:0"
}

func (runtimeListener *controlledListener) Running() bool {
	runtimeListener.mu.RLock()
	defer runtimeListener.mu.RUnlock()
	return runtimeListener.running
}

func (runtimeListener *controlledListener) Start(ctx context.Context) error {
	runtimeListener.startOnce.Do(func() { close(runtimeListener.startEntered) })
	if runtimeListener.releaseStart != nil {
		select {
		case <-runtimeListener.releaseStart:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if runtimeListener.startErr != nil {
		return runtimeListener.startErr
	}
	runtimeListener.mu.Lock()
	runtimeListener.running = true
	runtimeListener.mu.Unlock()
	return nil
}

func (runtimeListener *controlledListener) Stop(context.Context) error {
	runtimeListener.stopCalls.Add(1)
	runtimeListener.mu.Lock()
	runtimeListener.running = false
	runtimeListener.mu.Unlock()
	return nil
}

func fixedComposer(runtimeListener listener.Listener) runtimeComposer {
	return func(runtimeconfig.Snapshot, secretresolver.Resolver, message.Handler) (listener.Listener, error) {
		return runtimeListener, nil
	}
}

func newTestHost(t *testing.T, composer runtimeComposer) *DefaultHost {
	t.Helper()
	host, err := newHost(validSnapshot(), emptyResolver(t), nil, composer)
	if err != nil {
		t.Fatalf("newHost() error = %v", err)
	}
	return host
}

func currentHostState(host *DefaultHost) hostState {
	host.mu.RLock()
	defer host.mu.RUnlock()
	return host.state
}

func waitForHostState(t *testing.T, host *DefaultHost, want hostState) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if currentHostState(host) == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("Host did not enter state %v; current state = %v", want, currentHostState(host))
}

func waitForSignal(t *testing.T, signal <-chan struct{}, operation string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for %s", operation)
	}
}

func waitForResult(t *testing.T, result <-chan error, operation string) {
	t.Helper()
	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("%s error = %v", operation, err)
		}
	case <-time.After(time.Second):
		t.Fatalf("%s deadlocked", operation)
	}
}
