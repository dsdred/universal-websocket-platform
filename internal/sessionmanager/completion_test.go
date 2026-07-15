package sessionmanager

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestManagerInitialCompleteRemovesCommittedRegistration(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))

	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() = false, want true")
	}

	assertRegistrationCount(t, manager, 0)
	assertReservationCount(t, manager, 0)
}

func TestManagerCompleteReleasesSessionID(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() = false, want true")
	}

	newReservation := mustReserve(t, manager, "session-1")
	newRegistrationID := reservationFromHandle(t, newReservation).registrationID

	if newRegistrationID == registrationID {
		t.Fatalf("RegistrationID reused: %+v", registrationID)
	}
	newReservation.Abort()
}

func TestManagerDoubleCompleteSucceedsOnlyOnce(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))

	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("first Complete() = false, want true")
	}
	if completed := manager.Complete(registrationID); completed {
		t.Fatal("second Complete() = true, want false")
	}

	assertRegistrationCount(t, manager, 0)
}

func TestManagerCompleteUnknownRegistrationID(t *testing.T) {
	manager := New()
	knownID := mustCommit(t, mustReserve(t, manager, "session-1"))

	if completed := manager.Complete(RegistrationID{value: knownID.value + 1}); completed {
		t.Fatal("Complete(unknown) = true, want false")
	}

	assertRegistration(t, manager, knownID, "session-1")
}

func TestManagerStaleCompleteCannotRemoveReusedSessionID(t *testing.T) {
	manager := New()
	staleID := mustCommit(t, mustReserve(t, manager, "session-1"))
	if completed := manager.Complete(staleID); !completed {
		t.Fatal("Complete(stale candidate) = false, want true")
	}
	freshID := mustCommit(t, mustReserve(t, manager, "session-1"))

	if completed := manager.Complete(staleID); completed {
		t.Fatal("Complete(stale) = true, want false")
	}

	if freshID == staleID {
		t.Fatalf("fresh RegistrationID = stale RegistrationID %+v", staleID)
	}
	assertRegistration(t, manager, freshID, "session-1")
}

func TestManagerConcurrentCompleteRemovesExactlyOnce(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
		start := make(chan struct{})
		results := make(chan bool, concurrentCalls)
		var waitGroup sync.WaitGroup

		for range concurrentCalls {
			waitGroup.Add(1)
			go func() {
				defer waitGroup.Done()
				<-start
				results <- manager.Complete(registrationID)
			}()
		}
		close(start)
		waitGroup.Wait()
		close(results)

		successes := 0
		for completed := range results {
			if completed {
				successes++
			}
		}
		if successes != 1 {
			t.Fatalf("iteration %d: successful Complete count = %d, want 1", iteration, successes)
		}
		assertRegistrationCount(t, manager, 0)
	}
}

func TestManagerConcurrentCommitAndCompleteHaveAtomicOutcomes(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		handle := mustReserve(t, manager, "session-1")
		registrationID := reservationFromHandle(t, handle).registrationID
		start := make(chan struct{})
		commitResults := make(chan commitResult, 1)
		completeResults := make(chan bool, 1)

		go func() {
			<-start
			committedID, err := handle.Commit()
			commitResults <- commitResult{registrationID: committedID, err: err}
		}()
		go func() {
			<-start
			completeResults <- manager.Complete(registrationID)
		}()
		close(start)

		commitResult := <-commitResults
		completed := <-completeResults
		if commitResult.err != nil || commitResult.registrationID != registrationID {
			t.Fatalf("iteration %d: Commit() = (%+v, %v)", iteration, commitResult.registrationID, commitResult.err)
		}
		if completed {
			assertRegistrationCount(t, manager, 0)
		} else {
			assertRegistration(t, manager, registrationID, "session-1")
			if finalCompleted := manager.Complete(registrationID); !finalCompleted {
				t.Fatalf("iteration %d: final Complete() = false, want true", iteration)
			}
		}
	}
}

func TestManagerConcurrentCompleteAndReservePreserveSessionIDExclusivity(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
		start := make(chan struct{})
		completeResults := make(chan bool, 1)
		reserveResults := make(chan reserveResult, 1)

		go func() {
			<-start
			completeResults <- manager.Complete(registrationID)
		}()
		go func() {
			<-start
			handle, err := manager.Reserve("session-1")
			reserveResults <- reserveResult{handle: handle, err: err}
		}()
		close(start)

		if completed := <-completeResults; !completed {
			t.Fatalf("iteration %d: Complete() = false, want true", iteration)
		}
		reserveResult := <-reserveResults
		switch {
		case reserveResult.err == nil:
			reserveResult.handle.Abort()
		case errors.Is(reserveResult.err, ErrSessionIDReserved):
			retry := mustReserve(t, manager, "session-1")
			retry.Abort()
		default:
			t.Fatalf("iteration %d: Reserve() error = %v", iteration, reserveResult.err)
		}
		assertReservationCount(t, manager, 0)
		assertRegistrationCount(t, manager, 0)
	}
}

func TestManagerConcurrentCompleteAndBeginShutdown(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
		start := make(chan struct{})
		completeResults := make(chan bool, 1)
		shutdownDone := make(chan struct{})

		go func() {
			<-start
			completeResults <- manager.Complete(registrationID)
		}()
		go func() {
			<-start
			manager.BeginShutdown()
			close(shutdownDone)
		}()
		close(start)

		if completed := <-completeResults; !completed {
			t.Fatalf("iteration %d: Complete() = false, want true", iteration)
		}
		<-shutdownDone
		if got := manager.State(); got != StateClosing {
			t.Fatalf("iteration %d: State() = %v, want StateClosing", iteration, got)
		}
		assertRegistrationCount(t, manager, 0)
	}
}

func TestManagerCompleteDoesNotChangeLifecycle(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))

	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() = false, want true")
	}
	if got := manager.State(); got != StateOpen {
		t.Fatalf("State() after Complete = %v, want StateOpen", got)
	}

	manager.BeginShutdown()
	if err := manager.Wait(context.Background()); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() = %v, want StateClosed", got)
	}
}

func TestManagerCompleteAfterClosedDoesNotDependOnShutdownState(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
	manager.BeginShutdown()
	if err := manager.Wait(context.Background()); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}

	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() in Closed = false, want true")
	}
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() after Complete = %v, want StateClosed", got)
	}
	assertRegistrationCount(t, manager, 0)
}

func mustCommit(t *testing.T, handle ReservationHandle) RegistrationID {
	t.Helper()
	registrationID, err := handle.Commit()
	if err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	return registrationID
}
