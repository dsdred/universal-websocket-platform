package sessionmanager

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"testing"
	"time"
)

func TestManagerReserveCreatesReservation(t *testing.T) {
	manager := New()

	handle, err := manager.Reserve("session-1")
	if err != nil {
		t.Fatalf("Reserve() error = %v", err)
	}
	reservation := reservationFromHandle(t, handle)
	if reservation.sessionID != "session-1" {
		t.Fatalf("Reservation SessionID = %q, want session-1", reservation.sessionID)
	}
	if reservation.registrationID.value == 0 {
		t.Fatal("Reservation RegistrationID is zero")
	}
	assertReservationCount(t, manager, 1)
}

func TestManagerRegistrationIDsAreUnique(t *testing.T) {
	manager := New()
	first := mustReserve(t, manager, "session-1")
	second := mustReserve(t, manager, "session-2")

	firstID := reservationFromHandle(t, first).registrationID
	secondID := reservationFromHandle(t, second).registrationID
	if firstID == secondID {
		t.Fatalf("RegistrationIDs are equal: %+v", firstID)
	}
}

func TestManagerConcurrentRegistrationIDsAreUnique(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		start := make(chan struct{})
		results := make(chan reserveResult, concurrentCalls)

		for call := range concurrentCalls {
			go func(call int) {
				<-start
				handle, err := manager.Reserve(SessionID(fmt.Sprintf("session-%d", call)))
				results <- reserveResult{handle: handle, err: err}
			}(call)
		}

		close(start)
		registrationIDs := make(map[RegistrationID]struct{}, concurrentCalls)
		for call := range concurrentCalls {
			result := receiveReserveResult(t, results, iteration, call)
			if result.err != nil {
				t.Fatalf("iteration %d, call %d: Reserve() error = %v", iteration, call, result.err)
			}
			registrationID := reservationFromHandle(t, result.handle).registrationID
			if _, duplicate := registrationIDs[registrationID]; duplicate {
				t.Fatalf("iteration %d: duplicate RegistrationID %+v", iteration, registrationID)
			}
			registrationIDs[registrationID] = struct{}{}
		}
	}
}

func TestManagerRejectsDuplicateSessionID(t *testing.T) {
	manager := New()
	mustReserve(t, manager, "session-1")

	handle, err := manager.Reserve("session-1")

	if handle != nil || !errors.Is(err, ErrSessionIDReserved) {
		t.Fatalf("Reserve() = (%v, %v), want nil and ErrSessionIDReserved", handle, err)
	}
	assertReservationCount(t, manager, 1)
}

func TestReservationAbortRemovesReservation(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")

	handle.Abort()

	assertReservationCount(t, manager, 0)
}

func TestReservationRepeatedAbortIsSafe(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")

	for range 10 {
		handle.Abort()
	}

	assertReservationCount(t, manager, 0)
}

func TestReservationAbortUnlessCommitted(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")

	handle.AbortUnlessCommitted()

	assertReservationCount(t, manager, 0)
	handle.AbortUnlessCommitted()
	assertReservationCount(t, manager, 0)
}

func TestReservationConcurrentAbort(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		handle := mustReserve(t, manager, "session-1")
		start := make(chan struct{})
		done := make(chan struct{})
		var waitGroup sync.WaitGroup

		for range concurrentCalls {
			waitGroup.Add(1)
			go func() {
				defer waitGroup.Done()
				<-start
				handle.Abort()
			}()
		}

		go func() {
			waitGroup.Wait()
			close(done)
		}()
		close(start)
		waitForCompletion(t, done, fmt.Sprintf("iteration %d concurrent Abort", iteration))
		assertReservationCount(t, manager, 0)
	}
}

func TestManagerConcurrentReserve(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		start := make(chan struct{})
		results := make(chan reserveResult, concurrentCalls)

		for call := range concurrentCalls {
			go func(call int) {
				<-start
				handle, err := manager.Reserve(SessionID(fmt.Sprintf("session-%d", call)))
				results <- reserveResult{handle: handle, err: err}
			}(call)
		}

		close(start)
		for call := range concurrentCalls {
			result := receiveReserveResult(t, results, iteration, call)
			if result.err != nil {
				t.Fatalf("iteration %d, call %d: Reserve() error = %v", iteration, call, result.err)
			}
		}
		assertReservationCount(t, manager, concurrentCalls)
	}
}

func TestManagerReserveAfterAbortReusesSessionID(t *testing.T) {
	manager := New()
	first := mustReserve(t, manager, "session-1")
	firstID := reservationFromHandle(t, first).registrationID
	first.Abort()

	second := mustReserve(t, manager, "session-1")
	secondID := reservationFromHandle(t, second).registrationID

	if firstID == secondID {
		t.Fatalf("RegistrationID reused after Abort: %+v", firstID)
	}
	assertReservationCount(t, manager, 1)
}

func TestReservationStaleAbortDoesNotRemoveReusedSessionID(t *testing.T) {
	manager := New()
	stale := mustReserve(t, manager, "session-1")
	stale.Abort()
	fresh := mustReserve(t, manager, "session-1")

	stale.Abort()

	assertReservationCount(t, manager, 1)
	fresh.Abort()
	assertReservationCount(t, manager, 0)
}

func TestReservationAbortCompletesShutdown(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	manager.BeginShutdown()
	waitContext, cancelWait := context.WithCancel(context.Background())
	cancelWait()
	if err := manager.Wait(waitContext); !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait() error = %v, want context.Canceled", err)
	}
	assertReservationCount(t, manager, 1)
	if got := manager.State(); got != StateClosing {
		t.Fatalf("State() before Abort = %v, want StateClosing", got)
	}

	handle.Abort()

	assertReservationCount(t, manager, 0)
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() = %v, want StateClosed", got)
	}
	if err := manager.Wait(context.Background()); err != nil {
		t.Fatalf("Wait() after Abort error = %v", err)
	}
}

func TestManagerReserveAfterBeginShutdown(t *testing.T) {
	manager := New()
	manager.BeginShutdown()

	handle, err := manager.Reserve("session-1")

	if handle != nil || !errors.Is(err, ErrManagerNotOpen) {
		t.Fatalf("Reserve() = (%v, %v), want nil and ErrManagerNotOpen", handle, err)
	}
}

func TestManagerBeginShutdownReserveRace(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		start := make(chan struct{})
		beginDone := make(chan struct{})
		reserveResultChannel := make(chan reserveResult, 1)

		go func() {
			<-start
			manager.BeginShutdown()
			close(beginDone)
		}()
		go func() {
			<-start
			handle, err := manager.Reserve("session-1")
			reserveResultChannel <- reserveResult{handle: handle, err: err}
		}()

		close(start)
		waitForCompletion(t, beginDone, "BeginShutdown race participant")
		result := receiveReserveResult(t, reserveResultChannel, iteration, 0)
		switch {
		case result.err == nil:
			result.handle.Abort()
			if got := manager.State(); got != StateClosed {
				t.Fatalf("iteration %d: State() = %v, want StateClosed", iteration, got)
			}
		case errors.Is(result.err, ErrManagerNotOpen):
			if got := manager.State(); got != StateClosing {
				t.Fatalf("iteration %d: State() = %v, want StateClosing", iteration, got)
			}
		default:
			t.Fatalf("iteration %d: Reserve() error = %v", iteration, result.err)
		}
		assertReservationCount(t, manager, 0)
	}
}

func TestManagerRejectsInvalidSessionID(t *testing.T) {
	manager := New()
	tests := []struct {
		name      string
		sessionID SessionID
	}{
		{name: "empty", sessionID: ""},
		{name: "too long", sessionID: SessionID(string(make([]byte, maxSessionIDBytes+1)))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handle, err := manager.Reserve(test.sessionID)
			if handle != nil || !errors.Is(err, ErrInvalidSessionID) {
				t.Fatalf("Reserve() = (%v, %v), want nil and ErrInvalidSessionID", handle, err)
			}
		})
	}
}

func TestManagerRegistrationIDExhaustion(t *testing.T) {
	manager := New()
	manager.nextRegistrationID = math.MaxUint64
	last := mustReserve(t, manager, "last-session")
	if got := reservationFromHandle(t, last).registrationID.value; got != math.MaxUint64 {
		t.Fatalf("last RegistrationID = %d, want %d", got, uint64(math.MaxUint64))
	}

	handle, err := manager.Reserve("exhausted-session")
	if handle != nil || !errors.Is(err, ErrRegistrationIDExhausted) {
		t.Fatalf("Reserve() = (%v, %v), want nil and ErrRegistrationIDExhausted", handle, err)
	}
}

type reserveResult struct {
	handle ReservationHandle
	err    error
}

func mustReserve(t *testing.T, manager *Manager, sessionID SessionID) ReservationHandle {
	t.Helper()
	handle, err := manager.Reserve(sessionID)
	if err != nil {
		t.Fatalf("Reserve(%q) error = %v", sessionID, err)
	}
	return handle
}

func reservationFromHandle(t *testing.T, handle ReservationHandle) *reservation {
	t.Helper()
	concrete, ok := handle.(*reservationHandle)
	if !ok {
		t.Fatalf("ReservationHandle type = %T", handle)
	}
	return concrete.reservation
}

func assertReservationCount(t *testing.T, manager *Manager, want int) {
	t.Helper()
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	if got := len(manager.reservations); got != want {
		t.Fatalf("Reservation count = %d, want %d", got, want)
	}
	if got := len(manager.reservedSessions); got != want {
		t.Fatalf("reserved SessionID count = %d, want %d", got, want)
	}
}

func receiveReserveResult(
	t *testing.T,
	results <-chan reserveResult,
	iteration int,
	call int,
) reserveResult {
	t.Helper()
	select {
	case result := <-results:
		return result
	case <-time.After(time.Second):
		t.Fatalf("iteration %d, call %d: Reserve() deadlocked", iteration, call)
		return reserveResult{}
	}
}
