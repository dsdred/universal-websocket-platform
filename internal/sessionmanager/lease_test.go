package sessionmanager

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/lifetimelease"
)

func TestFailedCommitAndAbortDoNotCreateLifetimeLease(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	manager.BeginShutdown()

	result, err := handle.Commit()
	if result != (CommitResult{}) || !errors.Is(err, ErrManagerNotOpen) {
		t.Fatalf("Commit() = (%+v, %v), want zero result and ErrManagerNotOpen", result, err)
	}
	assertLifetimeLeaseCount(t, manager, 0)

	handle.Abort()
	assertLifetimeLeaseCount(t, manager, 0)
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() = %v, want StateClosed", got)
	}
}

func TestSuccessfulCommitPublishesRegistrationAndLifetimeLease(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	reservedID := reservationFromHandle(t, handle).registrationID

	result, err := handle.Commit()
	if err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	if result.RegistrationID() != reservedID {
		t.Fatalf("CommitResult.RegistrationID() = %+v, want %+v", result.RegistrationID(), reservedID)
	}
	if result.LifetimeLease() == nil {
		t.Fatal("CommitResult.LifetimeLease() = nil")
	}
	assertRegistration(t, manager, reservedID, "session-1")
	assertLifetimeLeaseCount(t, manager, 1)
}

func TestRepeatedCommitReturnsSameLifetimeLease(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")

	first, firstErr := handle.Commit()
	second, secondErr := handle.Commit()
	if firstErr != nil || secondErr != nil {
		t.Fatalf("Commit() errors = (%v, %v), want nil", firstErr, secondErr)
	}
	if first.RegistrationID() != second.RegistrationID() {
		t.Fatalf("Registration IDs differ: %+v and %+v", first.RegistrationID(), second.RegistrationID())
	}
	assertLifetimeLeaseCount(t, manager, 1)
	if outcome := first.LifetimeLease().Release(); outcome != lifetimelease.ReleaseOutcomeReleased {
		t.Fatalf("first Commit Lease Release() = %d, want ReleaseOutcomeReleased", outcome)
	}
	if outcome := second.LifetimeLease().Release(); outcome != lifetimelease.ReleaseOutcomeAccountingAnomaly {
		t.Fatalf("repeated Commit Lease Release() = %d, want ReleaseOutcomeAccountingAnomaly", outcome)
	}
}

func TestConcurrentRepeatedCommitPublishesOneLifetimeLease(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	start := make(chan struct{})
	results := make(chan CommitResult, concurrentCalls)
	errorsResult := make(chan error, concurrentCalls)
	var ready sync.WaitGroup
	ready.Add(concurrentCalls)

	for range concurrentCalls {
		go func() {
			ready.Done()
			<-start
			result, err := handle.Commit()
			results <- result
			errorsResult <- err
		}()
	}
	ready.Wait()
	close(start)

	leases := make([]lifetimelease.Lease, 0, concurrentCalls)
	for range concurrentCalls {
		result := <-results
		if err := <-errorsResult; err != nil {
			t.Fatalf("Commit() error = %v", err)
		}
		if result.LifetimeLease() == nil {
			t.Fatal("concurrent Commit returned nil Lifetime Lease")
		}
		leases = append(leases, result.LifetimeLease())
	}
	assertRegistrationCount(t, manager, 1)
	assertLifetimeLeaseCount(t, manager, 1)
	if outcome := leases[0].Release(); outcome != lifetimelease.ReleaseOutcomeReleased {
		t.Fatalf("first concurrent Commit Lease Release() = %d, want ReleaseOutcomeReleased", outcome)
	}
	for index, lease := range leases[1:] {
		if outcome := lease.Release(); outcome != lifetimelease.ReleaseOutcomeAccountingAnomaly {
			t.Fatalf("concurrent Commit Lease %d Release() = %d, want ReleaseOutcomeAccountingAnomaly", index+1, outcome)
		}
	}
}

func TestCommitAndBeginShutdownPublishRegistrationAndLifetimeLeaseAtomically(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		handle := mustReserve(t, manager, "session-1")
		start := make(chan struct{})
		type commitOutcome struct {
			result CommitResult
			err    error
		}
		commits := make(chan commitOutcome, 1)
		snapshots := make(chan ShutdownSnapshot, 1)
		var ready sync.WaitGroup
		ready.Add(2)

		go func() {
			ready.Done()
			<-start
			result, err := handle.Commit()
			commits <- commitOutcome{result: result, err: err}
		}()
		go func() {
			ready.Done()
			<-start
			snapshots <- manager.BeginShutdown()
		}()
		ready.Wait()
		close(start)

		commit := <-commits
		snapshot := <-snapshots
		switch {
		case commit.err == nil:
			assertSnapshotRegistration(t, snapshot, "session-1", commit.result.RegistrationID())
			assertLifetimeLeaseCount(t, manager, 1)
			if !manager.Complete(commit.result.RegistrationID()) {
				t.Fatalf("iteration %d: Complete() = false, want true", iteration)
			}
			if outcome := commit.result.LifetimeLease().Release(); outcome != lifetimelease.ReleaseOutcomeReleased {
				t.Fatalf("iteration %d: Release() = %d, want ReleaseOutcomeReleased", iteration, outcome)
			}
		case errors.Is(commit.err, ErrManagerNotOpen):
			if commit.result != (CommitResult{}) {
				t.Fatalf("iteration %d: rejected Commit result = %+v, want zero", iteration, commit.result)
			}
			assertSnapshotCount(t, snapshot, 0)
			assertLifetimeLeaseCount(t, manager, 0)
			handle.AbortUnlessCommitted()
		default:
			t.Fatalf("iteration %d: Commit() error = %v", iteration, commit.err)
		}
		if got := manager.State(); got != StateClosed {
			t.Fatalf("iteration %d: State() = %v, want StateClosed", iteration, got)
		}
	}
}

func TestDifferentRegistrationsReceiveDifferentLifetimeLeases(t *testing.T) {
	manager := New()
	first := commitWithLease(t, manager, "session-1")
	second := commitWithLease(t, manager, "session-2")

	if first.RegistrationID() == second.RegistrationID() {
		t.Fatalf("RegistrationID reused: %+v", first.RegistrationID())
	}
	assertLifetimeLeaseCount(t, manager, 2)
	if outcome := first.LifetimeLease().Release(); outcome != lifetimelease.ReleaseOutcomeReleased {
		t.Fatalf("first Lease Release() = %d, want ReleaseOutcomeReleased", outcome)
	}
	assertLifetimeLeaseCount(t, manager, 1)
	if outcome := second.LifetimeLease().Release(); outcome != lifetimelease.ReleaseOutcomeReleased {
		t.Fatalf("second Lease Release() = %d, want ReleaseOutcomeReleased", outcome)
	}
	assertLifetimeLeaseCount(t, manager, 0)
}

func TestCompleteRemovesRegistrationButKeepsLifetimeLease(t *testing.T) {
	manager := New()
	result := commitWithLease(t, manager, "session-1")
	manager.BeginShutdown()
	waitResult := startObservedWait(t, manager, context.Background())

	if !manager.Complete(result.RegistrationID()) {
		t.Fatal("Complete() = false, want true")
	}
	if view, found := manager.Lookup("session-1"); found || view != (RegistrationView{}) {
		t.Fatalf("Lookup() after Complete = (%+v, %t), want zero and false", view, found)
	}
	assertLifetimeLeaseCount(t, manager, 1)
	assertWaitBlocked(t, waitResult)
	if got := manager.State(); got != StateClosing {
		t.Fatalf("State() = %v, want StateClosing", got)
	}

	if outcome := result.LifetimeLease().Release(); outcome != lifetimelease.ReleaseOutcomeReleased {
		t.Fatalf("LifetimeLease.Release() = %d, want ReleaseOutcomeReleased", outcome)
	}
	assertWaitResult(t, waitResult, nil)
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() = %v, want StateClosed", got)
	}
}

func TestRepeatedLifetimeLeaseReleaseReturnsStableAnomaly(t *testing.T) {
	manager := New()
	result := commitWithLease(t, manager, "session-1")
	lease := result.LifetimeLease()

	if outcome := lease.Release(); outcome != lifetimelease.ReleaseOutcomeReleased {
		t.Fatalf("first Release() = %d, want ReleaseOutcomeReleased", outcome)
	}
	for call := range concurrentCalls {
		if outcome := lease.Release(); outcome != lifetimelease.ReleaseOutcomeAccountingAnomaly {
			t.Fatalf("repeated Release() call %d = %d, want ReleaseOutcomeAccountingAnomaly", call, outcome)
		}
	}
	assertLifetimeLeaseCount(t, manager, 0)
	assertRegistration(t, manager, result.RegistrationID(), "session-1")
}

func TestConcurrentLifetimeLeaseReleaseHasOneEffectiveMutation(t *testing.T) {
	manager := New()
	result := commitWithLease(t, manager, "session-1")
	original := result.LifetimeLease()
	copy1 := original
	copy2 := copy1
	leases := []lifetimelease.Lease{original, copy1, copy2}
	start := make(chan struct{})
	outcomes := make(chan lifetimelease.ReleaseOutcome, concurrentCalls)
	var ready sync.WaitGroup
	ready.Add(concurrentCalls)

	for index := range concurrentCalls {
		go func() {
			ready.Done()
			<-start
			outcomes <- leases[index%len(leases)].Release()
		}()
	}
	ready.Wait()
	close(start)

	released := 0
	anomalies := 0
	for range concurrentCalls {
		switch outcome := <-outcomes; outcome {
		case lifetimelease.ReleaseOutcomeReleased:
			released++
		case lifetimelease.ReleaseOutcomeAccountingAnomaly:
			anomalies++
		default:
			t.Fatalf("Release() = %d, want a defined outcome", outcome)
		}
	}
	if released != 1 || anomalies != concurrentCalls-1 {
		t.Fatalf("release outcomes = (%d released, %d anomalies), want (1, %d)", released, anomalies, concurrentCalls-1)
	}
	assertLifetimeLeaseCount(t, manager, 0)
}

func TestStaleLifetimeLeaseCannotReleaseReusedSessionRegistration(t *testing.T) {
	manager := New()
	stale := commitWithLease(t, manager, "session-1")
	if !manager.Complete(stale.RegistrationID()) {
		t.Fatal("Complete(stale) = false, want true")
	}
	fresh := commitWithLease(t, manager, "session-1")

	if outcome := stale.LifetimeLease().Release(); outcome != lifetimelease.ReleaseOutcomeReleased {
		t.Fatalf("stale Lease Release() = %d, want ReleaseOutcomeReleased", outcome)
	}
	assertRegistration(t, manager, fresh.RegistrationID(), "session-1")
	assertLifetimeLeaseCount(t, manager, 1)
	if outcome := stale.LifetimeLease().Release(); outcome != lifetimelease.ReleaseOutcomeAccountingAnomaly {
		t.Fatalf("repeated stale Lease Release() = %d, want ReleaseOutcomeAccountingAnomaly", outcome)
	}
	assertLifetimeLeaseCount(t, manager, 1)
}

func TestForeignLifetimeLeaseCannotAffectAccounting(t *testing.T) {
	manager := New()
	result := commitWithLease(t, manager, "session-1")
	foreign := &boundLifetimeLease{
		manager:        manager,
		registrationID: RegistrationID{value: result.RegistrationID().value + 1},
	}
	manager.BeginShutdown()
	if !manager.Complete(result.RegistrationID()) {
		t.Fatal("Complete() = false, want true")
	}
	waitResult := startObservedWait(t, manager, context.Background())

	if outcome := foreign.Release(); outcome != lifetimelease.ReleaseOutcomeAccountingAnomaly {
		t.Fatalf("foreign Release() = %d, want ReleaseOutcomeAccountingAnomaly", outcome)
	}
	assertLifetimeLeaseCount(t, manager, 1)
	assertWaitBlocked(t, waitResult)
	if outcome := result.LifetimeLease().Release(); outcome != lifetimelease.ReleaseOutcomeReleased {
		t.Fatalf("bound Release() = %d, want ReleaseOutcomeReleased", outcome)
	}
	assertWaitResult(t, waitResult, nil)
}

func TestConcurrentCompleteAndLifetimeLeaseReleasePreserveAccounting(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		result := commitWithLease(t, manager, "session-1")
		manager.BeginShutdown()
		start := make(chan struct{})
		completeResult := make(chan bool, 1)
		releaseResult := make(chan lifetimelease.ReleaseOutcome, 1)
		var ready sync.WaitGroup
		ready.Add(2)

		go func() {
			ready.Done()
			<-start
			completeResult <- manager.Complete(result.RegistrationID())
		}()
		go func() {
			ready.Done()
			<-start
			releaseResult <- result.LifetimeLease().Release()
		}()
		ready.Wait()
		close(start)

		if !<-completeResult {
			t.Fatalf("iteration %d: Complete() = false, want true", iteration)
		}
		if outcome := <-releaseResult; outcome != lifetimelease.ReleaseOutcomeReleased {
			t.Fatalf("iteration %d: Release() = %d, want ReleaseOutcomeReleased", iteration, outcome)
		}
		if err := manager.Wait(context.Background()); err != nil {
			t.Fatalf("iteration %d: Wait() error = %v", iteration, err)
		}
		if got := manager.State(); got != StateClosed {
			t.Fatalf("iteration %d: State() = %v, want StateClosed", iteration, got)
		}
	}
}

func TestLifetimeLeaseReleaseDoesNotChangeLookup(t *testing.T) {
	manager := New()
	result := commitWithLease(t, manager, "session-1")
	start := make(chan struct{})
	releaseResult := make(chan lifetimelease.ReleaseOutcome, 1)
	type lookupResult struct {
		view  RegistrationView
		found bool
	}
	lookup := make(chan lookupResult, 1)
	var ready sync.WaitGroup
	ready.Add(2)

	go func() {
		ready.Done()
		<-start
		releaseResult <- result.LifetimeLease().Release()
	}()
	go func() {
		ready.Done()
		<-start
		view, found := manager.Lookup("session-1")
		lookup <- lookupResult{view: view, found: found}
	}()
	ready.Wait()
	close(start)

	if outcome := <-releaseResult; outcome != lifetimelease.ReleaseOutcomeReleased {
		t.Fatalf("Release() = %d, want ReleaseOutcomeReleased", outcome)
	}
	observed := <-lookup
	assertLookupResult(t, observed.view, observed.found, "session-1", result.RegistrationID())
	assertRegistration(t, manager, result.RegistrationID(), "session-1")
}

func TestZeroLifetimeLeaseReturnsAccountingAnomaly(t *testing.T) {
	lease := (CommitResult{}).LifetimeLease()
	if outcome := lease.Release(); outcome != lifetimelease.ReleaseOutcomeAccountingAnomaly {
		t.Fatalf("zero CommitResult LifetimeLease.Release() = %d, want ReleaseOutcomeAccountingAnomaly", outcome)
	}
	var nilLease *boundLifetimeLease
	if outcome := nilLease.Release(); outcome != lifetimelease.ReleaseOutcomeAccountingAnomaly {
		t.Fatalf("nil boundLifetimeLease.Release() = %d, want ReleaseOutcomeAccountingAnomaly", outcome)
	}
}

func commitWithLease(t *testing.T, manager *Manager, sessionID SessionID) CommitResult {
	t.Helper()

	handle := mustReserve(t, manager, sessionID)
	result, err := handle.Commit()
	if err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	return result
}

func assertLifetimeLeaseCount(t *testing.T, manager *Manager, want int) {
	t.Helper()

	manager.mu.RLock()
	defer manager.mu.RUnlock()
	if got := len(manager.lifetimeLeases); got != want {
		t.Fatalf("Owner Lifetime Lease count = %d, want %d", got, want)
	}
}
