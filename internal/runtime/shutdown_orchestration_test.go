package runtime

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
	"github.com/dsdred/universal-websocket-platform/internal/lifetimelease"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
	"github.com/dsdred/universal-websocket-platform/internal/sessionmanager"
)

func TestHostShutdownOrchestratesSnapshotStopListenerAndManagerWait(t *testing.T) {
	manager := sessionmanager.New()
	runtimeListener := newControlledListener(nil, false)
	runtimeListener.stopEntered = make(chan struct{})
	runtimeListener.releaseStop = make(chan struct{})
	runtimeListener.stopReturned = make(chan struct{})
	host := startHostWithManager(t, runtimeListener, manager)
	runtimeContext := host.RuntimeContext()

	lateReserve := make(chan error, 1)
	bindings := []*shutdownStopBinding{
		newShutdownStopBinding(func() {
			_, err := manager.Reserve("late-session")
			lateReserve <- err
		}),
		newShutdownStopBinding(nil),
	}
	for _, binding := range bindings {
		binding.host = host
		binding.runtimeContext = runtimeContext
	}
	results := []sessionmanager.CommitResult{
		commitShutdownRegistration(t, manager, "session-1", bindings[0]),
		commitShutdownRegistration(t, manager, "session-2", bindings[1]),
	}

	var releaseListener sync.Once
	stopResult := make(chan error, 1)
	t.Cleanup(func() {
		releaseListener.Do(func() { close(runtimeListener.releaseStop) })
		completeShutdownRegistrations(results)
		_ = host.Stop(context.Background())
	})
	go func() { stopResult <- host.Stop(context.Background()) }()

	for index, binding := range bindings {
		waitForSignal(t, binding.requested, "Snapshot RequestStop")
		if got := binding.calls.Load(); got != 1 {
			t.Fatalf("Registration %d RequestStop calls = %d, want 1", index, got)
		}
		if binding.admissionOpen.Load() {
			t.Fatalf("Registration %d observed open admission during RequestStop", index)
		}
		if binding.runtimeCanceled.Load() {
			t.Fatalf("Registration %d observed root cancellation before RequestStop", index)
		}
	}
	if err := <-lateReserve; !errors.Is(err, sessionmanager.ErrManagerNotOpen) {
		t.Fatalf("Reserve() during Snapshot Stop error = %v, want ErrManagerNotOpen", err)
	}
	if got := manager.State(); got != sessionmanager.StateClosing {
		t.Fatalf("Manager state after BeginShutdown = %v, want Closing", got)
	}
	waitForSignal(t, runtimeListener.stopEntered, "Listener.Stop entry")
	assertContextCanceled(t, runtimeContext, "Runtime context before Listener.Stop")
	assertNoResult(t, stopResult, "Host.Stop while Listener.Stop is blocked")

	releaseListener.Do(func() { close(runtimeListener.releaseStop) })
	waitForSignal(t, runtimeListener.stopReturned, "Listener.Stop return")
	assertNoResult(t, stopResult, "Host.Stop while Manager accounting is active")

	completeShutdownRegistration(t, results[0])
	assertNoResult(t, stopResult, "Host.Stop while one Registration remains")
	completeShutdownRegistration(t, results[1])
	if err := waitForStopResult(t, stopResult); err != nil {
		t.Fatalf("Host.Stop() error = %v", err)
	}
	if got := manager.State(); got != sessionmanager.StateClosed {
		t.Fatalf("Manager state after Host.Stop = %v, want Closed", got)
	}
	if got := runtimeListener.stopCalls.Load(); got != 1 {
		t.Fatalf("Listener.Stop calls = %d, want 1", got)
	}
}

func TestHostShutdownEmptyAndAlreadyCompletedManager(t *testing.T) {
	t.Run("empty Runtime", func(t *testing.T) {
		manager := sessionmanager.New()
		runtimeListener := newControlledListener(nil, false)
		host := startHostWithManager(t, runtimeListener, manager)

		if err := host.Stop(context.Background()); err != nil {
			t.Fatalf("Host.Stop() error = %v", err)
		}
		if got := manager.State(); got != sessionmanager.StateClosed {
			t.Fatalf("Manager state = %v, want Closed", got)
		}
	})

	t.Run("Session completed before shutdown", func(t *testing.T) {
		manager := sessionmanager.New()
		runtimeListener := newControlledListener(nil, false)
		host := startHostWithManager(t, runtimeListener, manager)
		binding := newShutdownStopBinding(nil)
		result := commitShutdownRegistration(t, manager, "completed-session", binding)
		completeShutdownRegistration(t, result)

		if err := host.Stop(context.Background()); err != nil {
			t.Fatalf("Host.Stop() error = %v", err)
		}
		if got := binding.calls.Load(); got != 0 {
			t.Fatalf("completed Session RequestStop calls = %d, want 0", got)
		}
		if got := len(manager.BeginShutdown().Registrations()); got != 0 {
			t.Fatalf("shutdown Snapshot registrations = %d, want 0", got)
		}
	})
}

func TestHostConcurrentStopRunsOneShutdownOrchestration(t *testing.T) {
	manager := sessionmanager.New()
	runtimeListener := newControlledListener(nil, false)
	runtimeListener.stopEntered = make(chan struct{})
	runtimeListener.releaseStop = make(chan struct{})
	host := startHostWithManager(t, runtimeListener, manager)
	binding := newShutdownStopBinding(nil)
	result := commitShutdownRegistration(t, manager, "session-1", binding)

	const callers = 32
	start := make(chan struct{})
	results := make(chan error, callers)
	var callersDone sync.WaitGroup
	callersDone.Add(callers)
	for range callers {
		go func() {
			defer callersDone.Done()
			<-start
			results <- host.Stop(context.Background())
		}()
	}
	var releaseListener sync.Once
	t.Cleanup(func() {
		releaseListener.Do(func() { close(runtimeListener.releaseStop) })
		completeShutdownRegistrations([]sessionmanager.CommitResult{result})
		_ = host.Stop(context.Background())
		callersDone.Wait()
	})
	close(start)
	waitForSignal(t, binding.requested, "concurrent Snapshot RequestStop")
	waitForSignal(t, runtimeListener.stopEntered, "concurrent Listener.Stop entry")
	if got := binding.calls.Load(); got != 1 {
		t.Fatalf("RequestStop calls = %d, want 1", got)
	}
	assertNoResult(t, results, "concurrent Host.Stop before shutdown completion")

	completeShutdownRegistration(t, result)
	releaseListener.Do(func() { close(runtimeListener.releaseStop) })
	callersDone.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatalf("concurrent Host.Stop error = %v", err)
		}
	}
	if got := binding.calls.Load(); got != 1 {
		t.Fatalf("final RequestStop calls = %d, want 1", got)
	}
	if got := runtimeListener.stopCalls.Load(); got != 1 {
		t.Fatalf("Listener.Stop calls = %d, want 1", got)
	}
}

func TestHostShutdownPreservesListenerAndManagerWaitErrors(t *testing.T) {
	listenerErr := errors.New("listener shutdown failed")
	manager := sessionmanager.New()
	runtimeListener := newControlledListener(nil, false)
	runtimeListener.stopErr = listenerErr
	host := startHostWithManager(t, runtimeListener, manager)
	binding := newShutdownStopBinding(nil)
	result := commitShutdownRegistration(t, manager, "session-1", binding)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := host.Stop(ctx)
	if !errors.Is(err, listenerErr) || !errors.Is(err, context.Canceled) {
		t.Fatalf("Host.Stop() error = %v, want Listener and Manager Wait causes", err)
	}
	if got := manager.State(); got != sessionmanager.StateClosing {
		t.Fatalf("Manager state after bounded Wait = %v, want Closing", got)
	}
	if got := binding.calls.Load(); got != 1 {
		t.Fatalf("RequestStop calls = %d, want 1", got)
	}
	if repeated := host.Stop(context.Background()); !errors.Is(repeated, listenerErr) || !errors.Is(repeated, context.Canceled) {
		t.Fatalf("repeated Host.Stop() error = %v, want stored terminal result", repeated)
	}
	if got := binding.calls.Load(); got != 1 {
		t.Fatalf("repeated Stop changed RequestStop calls to %d", got)
	}

	completeShutdownRegistration(t, result)
	if got := manager.State(); got != sessionmanager.StateClosed {
		t.Fatalf("Manager state after eventual accounting completion = %v, want Closed", got)
	}
}

type shutdownStopBinding struct {
	calls           atomic.Int32
	requested       chan struct{}
	onFirst         func()
	once            sync.Once
	admissionOpen   atomic.Bool
	runtimeCanceled atomic.Bool
	host            *DefaultHost
	runtimeContext  context.Context
}

func newShutdownStopBinding(onFirst func()) *shutdownStopBinding {
	return &shutdownStopBinding{requested: make(chan struct{}), onFirst: onFirst}
}

func (binding *shutdownStopBinding) RequestStop() bool {
	first := binding.calls.Add(1) == 1
	if !first {
		return false
	}
	if binding.host != nil {
		binding.admissionOpen.Store(binding.host.CanAccept())
	}
	if binding.runtimeContext != nil {
		binding.runtimeCanceled.Store(binding.runtimeContext.Err() != nil)
	}
	if binding.onFirst != nil {
		binding.onFirst()
	}
	binding.once.Do(func() { close(binding.requested) })
	return true
}

func startHostWithManager(
	t *testing.T,
	runtimeListener *controlledListener,
	manager *sessionmanager.Manager,
) *DefaultHost {
	t.Helper()
	host := newTestHost(t, fixedComposer(runtimeListener))
	host.compose = func(
		runtimeconfig.Snapshot,
		secretresolver.Resolver,
		message.Handler,
	) (runtimeComposition, error) {
		return runtimeComposition{listener: runtimeListener, sessionManager: manager}, nil
	}
	if err := host.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	return host
}

func commitShutdownRegistration(
	t *testing.T,
	manager *sessionmanager.Manager,
	sessionID sessionmanager.SessionID,
	stop sessionmanager.StopPublicationBinding,
) sessionmanager.CommitResult {
	t.Helper()
	reservation, err := manager.Reserve(sessionID)
	if err != nil {
		t.Fatalf("Reserve(%q) error = %v", sessionID, err)
	}
	handoff := sessionmanager.NewCommitHandoff()
	input, err := sessionmanager.NewCommitInput(stop, handoff.CommitPublisher())
	if err != nil {
		t.Fatalf("NewCommitInput(%q) error = %v", sessionID, err)
	}
	result, err := reservation.Commit(input)
	if err != nil {
		t.Fatalf("Commit(%q) error = %v", sessionID, err)
	}
	return result
}

func completeShutdownRegistration(t *testing.T, result sessionmanager.CommitResult) {
	t.Helper()
	if got := result.CompletionAdapter().CompleteBoundRegistration(); got != executionowner.CompleteOutcomeCompleted {
		t.Fatalf("CompleteBoundRegistration() = %v, want Completed", got)
	}
	if got := result.LifetimeLease().Release(); got != lifetimelease.ReleaseOutcomeReleased {
		t.Fatalf("LifetimeLease.Release() = %v, want Released", got)
	}
}

func completeShutdownRegistrations(results []sessionmanager.CommitResult) {
	for _, result := range results {
		result.CompletionAdapter().CompleteBoundRegistration()
		result.LifetimeLease().Release()
	}
}

func assertNoResult(t *testing.T, result <-chan error, operation string) {
	t.Helper()
	select {
	case err := <-result:
		t.Fatalf("%s returned early: %v", operation, err)
	default:
	}
}

func waitForStopResult(t *testing.T, result <-chan error) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for Host.Stop")
		return nil
	}
}
