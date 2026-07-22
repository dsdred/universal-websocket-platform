package sessionmanager

import (
	"context"
	"errors"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/lifetimelease"
)

func TestReservationCommitPublishesRegistration(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	reservedID := reservationFromHandle(t, handle).registrationID

	committed, err := commitTestReservation(handle)
	if err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	if committed.RegistrationID() != reservedID {
		t.Fatalf("Commit() RegistrationID = %+v, want %+v", committed.RegistrationID(), reservedID)
	}

	assertReservationCount(t, manager, 0)
	assertRegistration(t, manager, committed.RegistrationID(), "session-1")
}

func TestReservationDoubleCommitIsIdempotent(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	input := newCommitTestInput()

	first, firstErr := handle.Commit(input)
	second, secondErr := handle.Commit(input)

	if firstErr != nil || secondErr != nil {
		t.Fatalf("Commit() errors = (%v, %v), want nil", firstErr, secondErr)
	}
	if first.RegistrationID() != second.RegistrationID() {
		t.Fatalf("Commit() IDs = (%+v, %+v), want equal", first.RegistrationID(), second.RegistrationID())
	}
	assertRegistrationCount(t, manager, 1)
}

func TestReservationCommitAfterCompleteReturnsRegistrationRemoved(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	registrationID := mustCommit(t, handle)
	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() = false, want true")
	}

	retriedID, err := commitTestReservation(handle)

	if retriedID != (CommitResult{}) || !errors.Is(err, ErrRegistrationRemoved) {
		t.Fatalf("Commit() after Complete = (%+v, %v), want zero ID and ErrRegistrationRemoved", retriedID, err)
	}
	assertRegistrationCount(t, manager, 0)
	if got := manager.State(); got != StateOpen {
		t.Fatalf("State() = %v, want StateOpen", got)
	}
}

func TestReservationCommitAfterSessionIDReuseDoesNotDescribeNewRegistration(t *testing.T) {
	manager := New()
	stale := mustReserve(t, manager, "session-1")
	staleID := mustCommit(t, stale)
	if completed := manager.Complete(staleID); !completed {
		t.Fatal("Complete() = false, want true")
	}
	fresh := mustReserve(t, manager, "session-1")
	freshID := mustCommit(t, fresh)

	retriedID, err := commitTestReservation(stale)

	if retriedID != (CommitResult{}) || !errors.Is(err, ErrRegistrationRemoved) {
		t.Fatalf("stale Commit() = (%+v, %v), want zero ID and ErrRegistrationRemoved", retriedID, err)
	}
	if freshID == staleID {
		t.Fatalf("fresh RegistrationID = stale RegistrationID %+v", staleID)
	}
	stale.Abort()
	stale.AbortUnlessCommitted()
	assertRegistration(t, manager, freshID, "session-1")
}

func TestReservationCommitAfterAbort(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	handle.Abort()

	registrationID, err := commitTestReservation(handle)

	if registrationID != (CommitResult{}) || !errors.Is(err, ErrReservationAborted) {
		t.Fatalf("Commit() = (%+v, %v), want zero ID and ErrReservationAborted", registrationID, err)
	}
	assertRegistrationCount(t, manager, 0)
}

func TestReservationAbortAfterCommitIsNoOp(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	result, err := commitTestReservation(handle)
	if err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	handle.Abort()
	handle.AbortUnlessCommitted()

	assertRegistration(t, manager, result.RegistrationID(), "session-1")
}

func TestReservationConcurrentCommitPublishesExactlyOnce(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	input := newCommitTestInput()
	start := make(chan struct{})
	results := make(chan commitResult, concurrentCalls)

	for range concurrentCalls {
		go func() {
			<-start
			result, err := handle.Commit(input)
			results <- commitResult{registrationID: result.RegistrationID(), err: err}
		}()
	}
	close(start)

	var committedID RegistrationID
	for range concurrentCalls {
		result := <-results
		if result.err != nil {
			t.Fatalf("concurrent Commit() error = %v", result.err)
		}
		if committedID == (RegistrationID{}) {
			committedID = result.registrationID
		} else if result.registrationID != committedID {
			t.Fatalf("concurrent Commit() ID = %+v, want %+v", result.registrationID, committedID)
		}
	}
	assertReservationCount(t, manager, 0)
	assertRegistration(t, manager, committedID, "session-1")
}

func TestReservationConcurrentAbortAndCommitHasOneTerminalOutcome(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		handle := mustReserve(t, manager, "session-1")
		start := make(chan struct{})
		commitResultChannel := make(chan commitResult, 1)
		abortDone := make(chan struct{})

		go func() {
			<-start
			result, err := commitTestReservation(handle)
			commitResultChannel <- commitResult{registrationID: result.RegistrationID(), err: err}
		}()
		go func() {
			<-start
			handle.Abort()
			close(abortDone)
		}()
		close(start)

		result := <-commitResultChannel
		<-abortDone
		switch {
		case result.err == nil:
			assertRegistration(t, manager, result.registrationID, "session-1")
		case errors.Is(result.err, ErrReservationAborted):
			assertRegistrationCount(t, manager, 0)
		default:
			t.Fatalf("iteration %d: Commit() error = %v", iteration, result.err)
		}
		assertReservationCount(t, manager, 0)
	}
}

func TestReservationCommitAndBeginShutdownShareLinearizationBoundary(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		handle := mustReserve(t, manager, "session-1")
		start := make(chan struct{})
		commitResultChannel := make(chan commitResult, 1)
		shutdownDone := make(chan struct{})

		go func() {
			<-start
			result, err := commitTestReservation(handle)
			commitResultChannel <- commitResult{registrationID: result.RegistrationID(), err: err}
		}()
		go func() {
			<-start
			manager.BeginShutdown()
			close(shutdownDone)
		}()
		close(start)

		result := <-commitResultChannel
		<-shutdownDone
		if result.err == nil {
			assertReservationCount(t, manager, 0)
			assertRegistration(t, manager, result.registrationID, "session-1")
			if got := manager.State(); got != StateClosing {
				t.Fatalf("iteration %d: State() = %v, want StateClosing", iteration, got)
			}
		} else if errors.Is(result.err, ErrManagerNotOpen) {
			assertReservationCount(t, manager, 1)
			assertRegistrationCount(t, manager, 0)
			handle.AbortUnlessCommitted()
			if got := manager.State(); got != StateClosed {
				t.Fatalf("iteration %d: State() = %v, want StateClosed", iteration, got)
			}
		} else {
			t.Fatalf("iteration %d: Commit() error = %v", iteration, result.err)
		}
	}
}

func TestReservationCommitAfterBeginShutdownRequiresAbort(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	manager.BeginShutdown()

	registrationID, err := commitTestReservation(handle)

	if registrationID != (CommitResult{}) || !errors.Is(err, ErrManagerNotOpen) {
		t.Fatalf("Commit() = (%+v, %v), want zero ID and ErrManagerNotOpen", registrationID, err)
	}
	assertReservationCount(t, manager, 1)
	assertRegistrationCount(t, manager, 0)

	handle.AbortUnlessCommitted()
	assertReservationCount(t, manager, 0)
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() = %v, want StateClosed", got)
	}
}

func TestCommittedSessionIDCannotBeReservedAgain(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	if _, err := commitTestReservation(handle); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	duplicate, err := manager.Reserve("session-1")

	if duplicate != nil || !errors.Is(err, ErrSessionIDReserved) {
		t.Fatalf("Reserve() = (%v, %v), want nil and ErrSessionIDReserved", duplicate, err)
	}
}

func TestCommittedRegistrationKeepsManagerClosingUntilComplete(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	result, err := commitTestReservation(handle)
	if err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	if outcome := result.LifetimeLease().Release(); outcome != lifetimelease.ReleaseOutcomeReleased {
		t.Fatalf("LifetimeLease.Release() = %d, want ReleaseOutcomeReleased", outcome)
	}

	manager.BeginShutdown()
	waitContext, cancelWait := context.WithCancel(context.Background())
	cancelWait()
	if err := manager.Wait(waitContext); !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait() error = %v, want context.Canceled", err)
	}
	if got := manager.State(); got != StateClosing {
		t.Fatalf("State() = %v, want StateClosing", got)
	}
	assertRegistration(t, manager, result.RegistrationID(), "session-1")
	if completed := manager.Complete(result.RegistrationID()); !completed {
		t.Fatal("Complete() = false, want true")
	}
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() after Complete = %v, want StateClosed", got)
	}
}

type commitResult struct {
	registrationID RegistrationID
	err            error
}

func assertRegistration(t *testing.T, manager *Manager, registrationID RegistrationID, sessionID SessionID) {
	t.Helper()
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	if len(manager.registrations) != 1 {
		t.Fatalf("Registration count = %d, want 1", len(manager.registrations))
	}
	registration, exists := manager.registrations[registrationID]
	if !exists {
		t.Fatalf("Registration %+v is not visible", registrationID)
	}
	if registration.registrationID != registrationID || registration.sessionID != sessionID {
		t.Fatalf("Registration = %+v, want ID %+v and SessionID %q", registration, registrationID, sessionID)
	}
	if got := manager.registeredSessions[sessionID]; got != registrationID {
		t.Fatalf("registered SessionID maps to %+v, want %+v", got, registrationID)
	}
}

func assertRegistrationCount(t *testing.T, manager *Manager, want int) {
	t.Helper()
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	if got := len(manager.registrations); got != want {
		t.Fatalf("Registration count = %d, want %d", got, want)
	}
	if got := len(manager.registeredSessions); got != want {
		t.Fatalf("registered SessionID count = %d, want %d", got, want)
	}
}
