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

func TestStartupTransactionRollsBackOnlyAcquiredResourcesInReverseOrder(t *testing.T) {
	acquisitionErr := errors.New("dependency acquisition failed")
	transaction := startupTransaction{}
	var rollbackOrder []string

	for _, name := range []string{"first", "second"} {
		name := name
		if err := transaction.acquire(func() (startupRollback, error) {
			return func(context.Context) error {
				rollbackOrder = append(rollbackOrder, name)
				return nil
			}, nil
		}); err != nil {
			t.Fatalf("acquire %s: %v", name, err)
		}
	}
	if err := transaction.acquire(func() (startupRollback, error) {
		return func(context.Context) error {
			rollbackOrder = append(rollbackOrder, "not-acquired")
			return nil
		}, acquisitionErr
	}); !errors.Is(err, acquisitionErr) {
		t.Fatalf("failed acquire error = %v, want acquisition error", err)
	}

	if err := transaction.rollback(context.Background()); err != nil {
		t.Fatalf("rollback() error = %v", err)
	}
	want := []string{"second", "first"}
	if len(rollbackOrder) != len(want) {
		t.Fatalf("rollback order = %v, want %v", rollbackOrder, want)
	}
	for index := range want {
		if rollbackOrder[index] != want[index] {
			t.Fatalf("rollback order = %v, want %v", rollbackOrder, want)
		}
	}
}

func TestHostDependencyAcquisitionErrorRestoresBuiltState(t *testing.T) {
	acquisitionErr := errors.New("dependency acquisition failed")
	var acquisitions atomic.Int32
	host := newTestHost(t, func(
		_ runtimeconfig.Snapshot,
		_ secretresolver.Resolver,
		_ message.Handler,
	) (listener.Listener, error) {
		acquisitions.Add(1)
		return nil, acquisitionErr
	})
	var contextCreations atomic.Int32
	var contextCancellations atomic.Int32
	host.newRuntimeContext = trackingRuntimeContextFactory(&contextCreations, &contextCancellations)
	if err := host.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	err := host.Start(context.Background())
	if !errors.Is(err, acquisitionErr) {
		t.Fatalf("Start() error = %v, want acquisition error", err)
	}
	if got := acquisitions.Load(); got != 1 {
		t.Fatalf("dependency acquisitions = %d, want 1", got)
	}
	if got := currentHostState(host); got != hostBuilt {
		t.Fatalf("Host state = %v, want hostBuilt", got)
	}
	if host.RuntimeContext() != nil || host.runtimeListener != nil || host.Running() {
		t.Fatal("failed dependency acquisition published Runtime resources")
	}
	if contextCreations.Load() != 0 || contextCancellations.Load() != 0 {
		t.Fatal("failed dependency acquisition created Runtime Context")
	}
}

func TestHostRollbackErrorPreservesStartupError(t *testing.T) {
	listenerStartErr := errors.New("listener start failed")
	listenerRollbackErr := errors.New("listener rollback failed")
	runtimeListener := newControlledListener(listenerStartErr, false)
	runtimeListener.stopErr = listenerRollbackErr
	host := newTestHost(t, fixedComposer(runtimeListener))
	if err := host.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	err := host.Start(context.Background())
	if !errors.Is(err, listenerStartErr) {
		t.Fatalf("Start() error = %v, want startup error", err)
	}
	if !errors.Is(err, listenerRollbackErr) {
		t.Fatalf("Start() error = %v, want rollback error", err)
	}
	if got := currentHostState(host); got != hostBuilt {
		t.Fatalf("Host state = %v, want hostBuilt", got)
	}
	if host.RuntimeContext() != nil || host.runtimeListener != nil || host.Running() {
		t.Fatal("failed startup transaction published Runtime resources")
	}
}

func TestHostStopDuringRollback(t *testing.T) {
	const iterations = 100
	for iteration := range iterations {
		listenerStartErr := errors.New("listener start failed")
		runtimeListener := newControlledListener(listenerStartErr, false)
		runtimeListener.stopEntered = make(chan struct{})
		runtimeListener.releaseStop = make(chan struct{})
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
		go func() { startResult <- host.Start(context.Background()) }()
		waitForSignal(t, runtimeListener.stopEntered, "startup rollback entry")

		stopResult := make(chan error, 1)
		go func() { stopResult <- host.Stop(context.Background()) }()
		waitForSignal(t, stopping, "Host Stopping transition")
		close(runtimeListener.releaseStop)

		select {
		case err := <-startResult:
			if !errors.Is(err, listenerStartErr) {
				t.Fatalf("iteration %d: Start() error = %v, want startup error", iteration, err)
			}
		case <-time.After(time.Second):
			t.Fatalf("iteration %d: Start() deadlocked during rollback", iteration)
		}
		waitForResult(t, stopResult, "Stop during rollback")
		if got := currentHostState(host); got != hostBuilt {
			t.Fatalf("iteration %d: Host state = %v, want hostBuilt", iteration, got)
		}
		if host.RuntimeContext() != nil || host.runtimeListener != nil || host.Running() {
			t.Fatalf("iteration %d: rollback published Runtime resources", iteration)
		}
		if got := runtimeListener.stopCalls.Load(); got != 1 {
			t.Fatalf("iteration %d: rollback calls = %d, want 1", iteration, got)
		}
	}
}

func TestHostCanStartAfterRollback(t *testing.T) {
	listenerStartErr := errors.New("listener start failed")
	failedListener := newControlledListener(listenerStartErr, false)
	runningListener := newControlledListener(nil, false)
	var acquisitions atomic.Int32
	host := newTestHost(t, func(
		_ runtimeconfig.Snapshot,
		_ secretresolver.Resolver,
		_ message.Handler,
	) (listener.Listener, error) {
		if acquisitions.Add(1) == 1 {
			return failedListener, nil
		}
		return runningListener, nil
	})
	if err := host.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if err := host.Start(context.Background()); !errors.Is(err, listenerStartErr) {
		t.Fatalf("first Start() error = %v, want startup error", err)
	}
	if got := currentHostState(host); got != hostBuilt {
		t.Fatalf("state after rollback = %v, want hostBuilt", got)
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("second Start() error = %v", err)
	}
	if !host.Running() || host.RuntimeContext() == nil {
		t.Fatal("Host is not fully Running after retry")
	}
	if got := acquisitions.Load(); got != 2 {
		t.Fatalf("dependency acquisitions = %d, want 2", got)
	}
	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}
