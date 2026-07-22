package session

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/coder/websocket"

	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
	"github.com/dsdred/universal-websocket-platform/internal/lifetimelease"
)

func TestOwnerSessionLifecyclePreservesCleanupPanicInTerminalResult(t *testing.T) {
	tests := []struct {
		name              string
		transportPanics   bool
		cancellationPanic bool
	}{
		{name: "transport-side panic", transportPanics: true},
		{name: "cancellation-side panic", cancellationPanic: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			core, err := newSessionCore(
				validPrincipal(),
				"192.0.2.1:4321",
				fixedSessionID("cleanup-panic"),
				nil,
				nil,
			)
			if err != nil {
				t.Fatalf("newSessionCore() error = %v", err)
			}
			done := make(chan struct{})
			var cancelOnce sync.Once
			prepared, err := prepareProvisionalSession(
				core,
				&cleanupPanicConnection{closePanics: test.transportPanics},
				cancellationDependency{
					done: done,
					cancel: func() {
						cancelOnce.Do(func() { close(done) })
						if test.cancellationPanic {
							panic("test cancellation panic")
						}
					},
				},
			)
			if err != nil {
				t.Fatalf("prepareProvisionalSession() error = %v", err)
			}
			if err := prepared.owner.Transition(
				executionowner.StatePreCommit,
				executionowner.StateCommitted,
			); err != nil {
				t.Fatalf("Transition() error = %v", err)
			}
			observer := &cleanupPanicObserver{}
			lease := &cleanupPanicLease{}

			executeErr := prepared.owner.Execute(
				context.Background(),
				prepared.lifecycle,
				cleanupPanicCompletion{},
				observer,
				lease,
			)
			if !errors.Is(executeErr, errCleanupPanicRun) {
				t.Fatalf("Execute() error = %v, want Run error", executeErr)
			}
			if observer.calls.Load() != 1 {
				t.Fatalf("Observer calls = %d, want 1", observer.calls.Load())
			}
			if observer.result.CleanupCategory() != executionowner.CleanupCategoryPanicked {
				t.Fatalf("CleanupCategory() = %v, want Panicked", observer.result.CleanupCategory())
			}
			if observer.result.CancellationCategory() != executionowner.CancellationCategoryConfirmed {
				t.Fatalf("CancellationCategory() = %v, want Confirmed", observer.result.CancellationCategory())
			}
			if lease.calls.Load() != 1 {
				t.Fatalf("Lease Release calls = %d, want 1", lease.calls.Load())
			}
		})
	}
}

var errCleanupPanicRun = errors.New("test Run failure")

type cleanupPanicConnection struct {
	closePanics bool
}

func (*cleanupPanicConnection) Read(context.Context) (websocket.MessageType, []byte, error) {
	return 0, nil, errCleanupPanicRun
}

func (*cleanupPanicConnection) Write(context.Context, websocket.MessageType, []byte) error {
	return nil
}

func (connection *cleanupPanicConnection) Close(websocket.StatusCode, string) error {
	if connection.closePanics {
		panic("test transport cleanup panic")
	}
	return nil
}

func (*cleanupPanicConnection) CloseNow() error {
	return nil
}

type cleanupPanicCompletion struct{}

func (cleanupPanicCompletion) CompleteBoundRegistration() executionowner.CompleteOutcome {
	return executionowner.CompleteOutcomeCompleted
}

type cleanupPanicObserver struct {
	calls  atomic.Int32
	result executionowner.TerminalResult
}

func (observer *cleanupPanicObserver) Observe(result executionowner.TerminalResult) {
	observer.result = result
	observer.calls.Add(1)
}

type cleanupPanicLease struct {
	calls atomic.Int32
}

func (lease *cleanupPanicLease) Release() lifetimelease.ReleaseOutcome {
	lease.calls.Add(1)
	return lifetimelease.ReleaseOutcomeReleased
}
