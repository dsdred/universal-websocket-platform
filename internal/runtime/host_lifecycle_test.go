package runtime

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/listener"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

func TestHostBuildDoesNotAcquireDependencies(t *testing.T) {
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
	if err := host.Build(); !errors.Is(err, ErrHostAlreadyBuilt) {
		t.Fatalf("second Build() error = %v, want ErrHostAlreadyBuilt", err)
	}

	if got := factoryCreations.Load(); got != 0 {
		t.Fatalf("Factory creations during Build = %d, want 0", got)
	}
	if got := listenerCreations.Load(); got != 0 {
		t.Fatalf("Listener creations during Build = %d, want 0", got)
	}
	if host.runtimeListener != nil {
		t.Fatal("Build() published Listener before startup transaction")
	}

	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if got := factoryCreations.Load(); got != 1 {
		t.Fatalf("Factory creations during Start = %d, want 1", got)
	}
	if got := listenerCreations.Load(); got != 1 {
		t.Fatalf("Listener creations during Start = %d, want 1", got)
	}
	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestHostListenerStartErrorRestoresBuiltState(t *testing.T) {
	listenerStartErr := errors.New("listener start failed")
	runtimeListener := newControlledListener(listenerStartErr, false)
	host := newTestHost(t, fixedComposer(runtimeListener))
	var contextCreations atomic.Int32
	var contextCancellations atomic.Int32
	host.newRuntimeContext = trackingRuntimeContextFactory(&contextCreations, &contextCancellations)
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
	if host.Ready() {
		t.Fatal("Ready() = true after Listener.Start error")
	}
	if host.CanAccept() {
		t.Fatal("CanAccept() = true after Listener.Start error")
	}
	if host.RuntimeContext() != nil {
		t.Fatal("RuntimeContext() after Listener.Start error is not nil")
	}
	if host.runtimeCancel != nil {
		t.Fatal("runtime CancelFunc after Listener.Start error is not nil")
	}
	if got := contextCreations.Load(); got != 0 {
		t.Fatalf("Runtime context creations after Listener.Start error = %d, want 0", got)
	}
	if got := contextCancellations.Load(); got != 0 {
		t.Fatalf("Runtime context cancellations after Listener.Start error = %d, want 0", got)
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
	if got := runtimeListener.stopCalls.Load(); got != 1 {
		t.Fatalf("Listener rollback calls after failed Start = %d, want 1", got)
	}
}

func TestHostStopDuringStarting(t *testing.T) {
	const iterations = 100
	for iteration := range iterations {
		runtimeListener := newControlledListener(nil, true)
		host := newTestHost(t, fixedComposer(runtimeListener))
		stopping := make(chan struct{})
		var stoppingOnce sync.Once
		host.stateObserver = func(state hostState) {
			if state == hostStopping {
				stoppingOnce.Do(func() { close(stopping) })
			}
		}
		if err := host.Build(); err != nil {
			t.Fatalf("iteration %d: Build() error = %v", iteration, err)
		}

		startResult := make(chan error, 1)
		go func() {
			startResult <- host.Start(context.Background())
		}()
		waitForSignal(t, runtimeListener.startEntered, "Listener.Start entry")
		if host.Ready() {
			t.Fatalf("iteration %d: Ready() = true during Starting", iteration)
		}
		if host.CanAccept() {
			t.Fatalf("iteration %d: CanAccept() = true during Starting", iteration)
		}

		stopResult := make(chan error, 1)
		go func() {
			stopResult <- host.Stop(context.Background())
		}()
		waitForSignal(t, stopping, "Host Stopping transition")
		if host.Running() {
			t.Fatalf("iteration %d: Host became Running while Stop was pending", iteration)
		}
		if host.Ready() {
			t.Fatalf("iteration %d: Ready() = true after Stop began", iteration)
		}
		if host.CanAccept() {
			t.Fatalf("iteration %d: CanAccept() = true after Stop began", iteration)
		}
		select {
		case err := <-stopResult:
			t.Fatalf("iteration %d: Stop() returned before Listener.Start completed: %v", iteration, err)
		default:
		}

		close(runtimeListener.releaseStart)
		waitForResult(t, startResult, "Start")
		waitForResult(t, stopResult, "Stop")

		runtimeContext := host.RuntimeContext()
		if runtimeContext == nil {
			t.Fatalf("iteration %d: RuntimeContext() is nil after successful Listener.Start", iteration)
		}
		assertContextCanceled(t, runtimeContext, "Runtime context after concurrent Stop")
		if host.Running() {
			t.Fatalf("iteration %d: Running() = true after Stop", iteration)
		}
		if host.Ready() {
			t.Fatalf("iteration %d: Ready() = true after Stop", iteration)
		}
		if host.CanAccept() {
			t.Fatalf("iteration %d: CanAccept() = true after Stop", iteration)
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
	startEntered  chan struct{}
	releaseStart  chan struct{}
	startErr      error
	startOnce     sync.Once
	startObserver func()
	startCalls    atomic.Int32
	stopEntered   chan struct{}
	releaseStop   chan struct{}
	stopReturned  chan struct{}
	stopErr       error
	stopOnce      sync.Once
	returnOnce    sync.Once
	stopObserver  func()
	mu            sync.RWMutex
	running       bool
	stopCalls     atomic.Int32
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
	runtimeListener.startCalls.Add(1)
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
	if runtimeListener.startObserver != nil {
		runtimeListener.startObserver()
	}
	return nil
}

func (runtimeListener *controlledListener) Stop(context.Context) error {
	runtimeListener.stopCalls.Add(1)
	if runtimeListener.stopObserver != nil {
		runtimeListener.stopObserver()
	}
	if runtimeListener.stopEntered != nil {
		runtimeListener.stopOnce.Do(func() { close(runtimeListener.stopEntered) })
	}
	if runtimeListener.releaseStop != nil {
		<-runtimeListener.releaseStop
	}
	runtimeListener.mu.Lock()
	runtimeListener.running = false
	runtimeListener.mu.Unlock()
	if runtimeListener.stopReturned != nil {
		runtimeListener.returnOnce.Do(func() { close(runtimeListener.stopReturned) })
	}
	return runtimeListener.stopErr
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

func TestHostServeCompletionProductionLifetime(t *testing.T) {
	snapshot, diagnostics := buildSnapshotForTest(func(version *configurationversion.ConfigurationVersion) {
		version.Listener.Port = availablePort(t)
	})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	host, err := NewHost(snapshot, emptyResolver(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = host.Stop(context.Background()) })
	assertHostServeCompletion(t, host.listenerServeCompletion(), false, false)
	if err := host.Build(); err != nil {
		t.Fatal(err)
	}
	assertHostServeCompletion(t, host.listenerServeCompletion(), false, false)
	if err := host.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	view := host.listenerServeCompletion()
	assertHostServeCompletion(t, view, false, true)
	copied := view
	var readers sync.WaitGroup
	var invalid atomic.Int32
	for range 4 {
		readers.Add(1)
		go func() {
			defer readers.Done()
			seenReturn := false
			for range 1000 {
				returned, available := host.listenerServeCompletion().Observe()
				if !available || (seenReturn && !returned) {
					invalid.Add(1)
				}
				seenReturn = seenReturn || returned
			}
		}()
	}
	if err := host.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	readers.Wait()
	if invalid.Load() != 0 {
		t.Fatal("Host observation lost coverage or regressed")
	}
	for range 3 {
		assertHostServeCompletion(t, host.listenerServeCompletion(), true, true)
		assertHostServeCompletion(t, copied, true, true)
	}
	if host.Ready() || host.Running() || host.CanAccept() {
		t.Fatal("Stop lifecycle changed")
	}
	assertContextCanceled(t, host.RuntimeContext(), "stopped Host")

	second, err := NewHost(snapshot, emptyResolver(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = second.Stop(context.Background()) })
	if err := second.Build(); err != nil {
		t.Fatal(err)
	}
	if err := second.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	secondView := second.listenerServeCompletion()
	if secondView == view {
		t.Fatal("different Host reused source view")
	}
	assertHostServeCompletion(t, secondView, false, true)
	assertHostServeCompletion(t, copied, true, true)
	if err := second.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertHostServeCompletion(t, secondView, true, true)
	assertHostServeCompletion(t, copied, true, true)
}

func TestHostServeCompletionEarlyReturnBeforeCommit(t *testing.T) {
	host := newServeCompletionProductionHost(t)
	original := host.compose
	returnedBeforeCommit, releaseStart := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseStart) }) }
	var sourceView listener.ServeCompletion
	host.compose = func(ctx context.Context, snapshot runtimeconfig.Snapshot, resolver secretresolver.Resolver, handler message.Handler) (runtimeComposition, error) {
		components, err := original(ctx, snapshot, resolver, handler)
		if err != nil {
			return runtimeComposition{}, err
		}
		sourceView = components.serveCompletion
		actual := components.listener
		components.listener = &serveCompletionStartWrapper{Listener: actual, afterStart: func() {
			// The existing private composer seam forces a real Serve shutdown
			// before Host commit, without writing the retained source cell.
			if err := actual.Stop(context.Background()); err != nil {
				t.Errorf("source Stop: %v", err)
			}
			close(returnedBeforeCommit)
			<-releaseStart
		}}
		return components, nil
	}
	t.Cleanup(func() { release(); _ = host.Stop(context.Background()) })
	started := make(chan error, 1)
	go func() { started <- host.Start(context.Background()) }()
	waitForSignal(t, returnedBeforeCommit, "real source return before Host commit")
	assertHostServeCompletion(t, sourceView, true, true)
	assertHostServeCompletion(t, host.listenerServeCompletion(), false, false)
	if host.Ready() || host.CanAccept() {
		t.Fatal("Host published before successful Start returned")
	}
	release()
	waitForResult(t, started, "Host Start")
	if host.listenerServeCompletion() != sourceView {
		t.Fatal("Host did not publish its exact source view")
	}
	assertHostServeCompletion(t, host.listenerServeCompletion(), true, true)
	// Evidence does not silently rewrite the existing committed lifecycle.
	if !host.Ready() {
		t.Fatal("completion evidence changed Host readiness semantics")
	}
	if err := host.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertHostServeCompletion(t, sourceView, true, true)
}

func TestHostServeCompletionFailedStartupRetryIsolation(t *testing.T) {
	host := newServeCompletionProductionHost(t)
	original := host.compose
	firstFailure := errors.New("source Start refused before Serve")
	var views []listener.ServeCompletion
	host.compose = func(ctx context.Context, snapshot runtimeconfig.Snapshot, resolver secretresolver.Resolver, handler message.Handler) (runtimeComposition, error) {
		components, err := original(ctx, snapshot, resolver, handler)
		if err != nil {
			return runtimeComposition{}, err
		}
		views = append(views, components.serveCompletion)
		if len(views) == 1 {
			components.listener = &serveCompletionStartWrapper{Listener: components.listener, startErr: firstFailure}
		}
		return components, nil
	}
	if err := host.Start(context.Background()); !errors.Is(err, firstFailure) {
		t.Fatalf("first Start = %v", err)
	}
	assertHostServeCompletion(t, host.listenerServeCompletion(), false, false)
	assertHostServeCompletion(t, views[0], false, true)
	if host.RuntimeContext() != nil || host.runtimeListener != nil || host.Ready() || host.CanAccept() || currentHostState(host) != hostBuilt {
		t.Fatal("failed startup changed lifecycle/resource publication")
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(views) != 2 || views[0] == views[1] || host.listenerServeCompletion() != views[1] {
		t.Fatal("retry source binding incorrect")
	}
	assertHostServeCompletion(t, views[0], false, true)
	assertHostServeCompletion(t, views[1], false, true)
	if err := host.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertHostServeCompletion(t, views[0], false, true)
	assertHostServeCompletion(t, views[1], true, true)
}

func TestHostServeCompletionLegacyCompositionUnavailable(t *testing.T) {
	host := newTestHost(t, fixedComposer(newControlledListener(nil, false)))
	if err := host.Build(); err != nil {
		t.Fatal(err)
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertHostServeCompletion(t, host.listenerServeCompletion(), false, false)
	if err := host.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertHostServeCompletion(t, host.listenerServeCompletion(), false, false)
}

type serveCompletionStartWrapper struct {
	listener.Listener
	startErr   error
	afterStart func()
}

func (source *serveCompletionStartWrapper) Start(ctx context.Context) error {
	if source.startErr != nil {
		return source.startErr
	}
	if err := source.Listener.Start(ctx); err != nil {
		return err
	}
	if source.afterStart != nil {
		source.afterStart()
	}
	return nil
}

func newServeCompletionProductionHost(t *testing.T) *DefaultHost {
	t.Helper()
	snapshot, diagnostics := buildSnapshotForTest(func(version *configurationversion.ConfigurationVersion) {
		version.Listener.Port = availablePort(t)
	})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	host, err := NewHost(snapshot, emptyResolver(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = host.Stop(context.Background()) })
	if err := host.Build(); err != nil {
		t.Fatal(err)
	}
	return host
}

func assertHostServeCompletion(t *testing.T, view listener.ServeCompletion, wantReturned, wantAvailable bool) {
	t.Helper()
	if returned, available := view.Observe(); returned != wantReturned || available != wantAvailable {
		t.Fatalf("Observe() = (%t, %t), want (%t, %t)", returned, available, wantReturned, wantAvailable)
	}
}
