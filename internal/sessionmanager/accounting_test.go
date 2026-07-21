package sessionmanager

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestManagerWaitEmptyManagerClosesShutdown(t *testing.T) {
	manager := New()
	manager.BeginShutdown()
	if got := manager.State(); got != StateClosing {
		t.Fatalf("State() after BeginShutdown = %v, want StateClosing", got)
	}

	if err := manager.Wait(context.Background()); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() after Wait = %v, want StateClosed", got)
	}
}

func TestManagerWaitBlockedByReservationAndAbortWakesIt(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	manager.BeginShutdown()
	waitResult := startObservedWait(t, manager, context.Background())
	assertWaitBlocked(t, waitResult)

	handle.Abort()

	assertWaitResult(t, waitResult, nil)
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() = %v, want StateClosed", got)
	}
}

func TestManagerWaitBlockedByRegistrationAndCompleteWakesIt(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
	manager.BeginShutdown()
	waitResult := startObservedWait(t, manager, context.Background())
	assertWaitBlocked(t, waitResult)

	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() = false, want true")
	}

	assertWaitResult(t, waitResult, nil)
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() = %v, want StateClosed", got)
	}
}

func TestManagerCommitPreservesAccounting(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	assertAccountingCount(t, manager, 1)

	registrationID := mustCommit(t, handle)

	assertReservationCount(t, manager, 0)
	assertRegistration(t, manager, registrationID, "session-1")
	assertAccountingCount(t, manager, 1)
}

func TestManagerConcurrentCommitAndWaitPreserveRegistrationAccounting(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	registrationID := mustCommit(t, handle)
	manager.BeginShutdown()
	waitResult := startObservedWait(t, manager, context.Background())
	assertWaitBlocked(t, waitResult)
	commitResultChannel := make(chan commitResult, 1)

	go func() {
		committed, err := commitTestReservation(handle)
		commitResultChannel <- commitResult{registrationID: committed.RegistrationID(), err: err}
	}()
	commitResult := <-commitResultChannel
	if commitResult.err != nil || commitResult.registrationID != registrationID {
		t.Fatalf("repeated Commit() = (%+v, %v), want %+v and nil", commitResult.registrationID, commitResult.err, registrationID)
	}
	assertWaitBlocked(t, waitResult)
	assertAccountingCount(t, manager, 1)

	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() = false, want true")
	}
	assertWaitResult(t, waitResult, nil)
}

func TestManagerWaitContextCancellationKeepsClosing(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	manager.BeginShutdown()
	waitContext, cancelWait := context.WithCancel(context.Background())
	waitResult := startObservedWait(t, manager, waitContext)
	assertWaitBlocked(t, waitResult)

	cancelWait()

	assertWaitResult(t, waitResult, context.Canceled)
	if got := manager.State(); got != StateClosing {
		t.Fatalf("State() after canceled Wait = %v, want StateClosing", got)
	}
	assertReservationCount(t, manager, 1)
	handle.Abort()
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() after Abort = %v, want StateClosed", got)
	}
}

func TestManagerClosedOnlyAfterReservationAndRegistrationAccountingEmpty(t *testing.T) {
	manager := New()
	reservation := mustReserve(t, manager, "reserved-session")
	registrationID := mustCommit(t, mustReserve(t, manager, "registered-session"))
	manager.BeginShutdown()
	waitResult := startObservedWait(t, manager, context.Background())

	reservation.Abort()

	if got := manager.State(); got != StateClosing {
		t.Fatalf("State() with Registration = %v, want StateClosing", got)
	}
	assertAccountingCount(t, manager, 1)
	assertWaitBlocked(t, waitResult)

	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() = false, want true")
	}
	assertWaitResult(t, waitResult, nil)
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() after empty accounting = %v, want StateClosed", got)
	}
}

func TestManagerLookupVisibleUntilCompleteDuringShutdown(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
	manager.BeginShutdown()
	waitResult := startObservedWait(t, manager, context.Background())

	view, found := manager.Lookup("session-1")
	assertLookupResult(t, view, found, "session-1", registrationID)
	assertWaitBlocked(t, waitResult)

	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() = false, want true")
	}
	assertWaitResult(t, waitResult, nil)
	if view, found := manager.Lookup("session-1"); found || view != (RegistrationView{}) {
		t.Fatalf("Lookup() after Complete = (%+v, %t), want zero and false", view, found)
	}
}

type observedContext struct {
	context.Context
	doneObserved chan struct{}
	once         sync.Once
}

func (ctx *observedContext) Done() <-chan struct{} {
	ctx.once.Do(func() {
		close(ctx.doneObserved)
	})
	return ctx.Context.Done()
}

func startObservedWait(t *testing.T, manager *Manager, parent context.Context) <-chan error {
	t.Helper()
	ctx := &observedContext{
		Context:      parent,
		doneObserved: make(chan struct{}),
	}
	result := make(chan error, 1)
	go func() {
		result <- manager.Wait(ctx)
	}()
	waitForCompletion(t, ctx.doneObserved, "Wait to enter blocking observation")
	return result
}

func assertWaitBlocked(t *testing.T, result <-chan error) {
	t.Helper()
	select {
	case err := <-result:
		t.Fatalf("Wait() returned while accounting was nonempty: %v", err)
	default:
	}
}

func assertWaitResult(t *testing.T, result <-chan error, want error) {
	t.Helper()
	select {
	case err := <-result:
		if !errors.Is(err, want) {
			t.Fatalf("Wait() error = %v, want %v", err, want)
		}
	case <-time.After(time.Second):
		t.Fatal("Wait() did not complete")
	}
}

func assertAccountingCount(t *testing.T, manager *Manager, want int) {
	t.Helper()
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	if got := len(manager.reservations) + len(manager.registrations) + len(manager.lifetimeLeases); got != want {
		t.Fatalf("accounting count = %d, want %d", got, want)
	}
}
