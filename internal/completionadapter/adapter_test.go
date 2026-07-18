package completionadapter

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
	"github.com/dsdred/universal-websocket-platform/internal/lifetimelease"
	"github.com/dsdred/universal-websocket-platform/internal/sessionmanager"
)

func TestNewRejectsInvalidBinding(t *testing.T) {
	manager := sessionmanager.New()
	registrationID := committedRegistration(t, manager, "session-1")

	tests := []struct {
		name           string
		manager        *sessionmanager.Manager
		registrationID sessionmanager.RegistrationID
	}{
		{name: "nil Manager", registrationID: registrationID},
		{name: "zero Registration ID", manager: manager},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			adapter, err := New(test.manager, test.registrationID)
			if adapter != nil || !errors.Is(err, ErrInvalidBinding) {
				t.Fatalf("New() = (%v, %v), want nil and ErrInvalidBinding", adapter, err)
			}
		})
	}
}

func TestCompletionAdapterCompletesOnlyBoundRegistration(t *testing.T) {
	manager := sessionmanager.New()
	firstID := committedRegistration(t, manager, "session-1")
	secondID := committedRegistration(t, manager, "session-2")
	adapter := newAdapter(t, manager, firstID)

	if outcome := adapter.CompleteBoundRegistration(); outcome != executionowner.CompleteOutcomeCompleted {
		t.Fatalf("CompleteBoundRegistration() = %d, want CompleteOutcomeCompleted", outcome)
	}
	assertRegistrationAbsent(t, manager, "session-1")
	assertRegistrationPresent(t, manager, "session-2", secondID)
}

func TestRepeatedCompletionReturnsAccountingAnomaly(t *testing.T) {
	manager := sessionmanager.New()
	registrationID := committedRegistration(t, manager, "session-1")
	adapter := newAdapter(t, manager, registrationID)

	if outcome := adapter.CompleteBoundRegistration(); outcome != executionowner.CompleteOutcomeCompleted {
		t.Fatalf("first CompleteBoundRegistration() = %d, want CompleteOutcomeCompleted", outcome)
	}
	if outcome := adapter.CompleteBoundRegistration(); outcome != executionowner.CompleteOutcomeAccountingAnomaly {
		t.Fatalf("repeated CompleteBoundRegistration() = %d, want CompleteOutcomeAccountingAnomaly", outcome)
	}
	assertRegistrationAbsent(t, manager, "session-1")
}

func TestCompletionOfAlreadyRemovedRegistrationReturnsAccountingAnomaly(t *testing.T) {
	manager := sessionmanager.New()
	registrationID := committedRegistration(t, manager, "session-1")
	adapter := newAdapter(t, manager, registrationID)
	if !manager.Complete(registrationID) {
		t.Fatal("Manager.Complete() = false, want true")
	}

	if outcome := adapter.CompleteBoundRegistration(); outcome != executionowner.CompleteOutcomeAccountingAnomaly {
		t.Fatalf("CompleteBoundRegistration() = %d, want CompleteOutcomeAccountingAnomaly", outcome)
	}
}

func TestCompletionUpdatesCommittedAccountingOnce(t *testing.T) {
	manager := sessionmanager.New()
	registrationID := committedRegistration(t, manager, "session-1")
	adapter := newAdapter(t, manager, registrationID)
	manager.BeginShutdown()

	if outcome := adapter.CompleteBoundRegistration(); outcome != executionowner.CompleteOutcomeCompleted {
		t.Fatalf("CompleteBoundRegistration() = %d, want CompleteOutcomeCompleted", outcome)
	}
	if outcome := adapter.CompleteBoundRegistration(); outcome != executionowner.CompleteOutcomeAccountingAnomaly {
		t.Fatalf("repeated CompleteBoundRegistration() = %d, want CompleteOutcomeAccountingAnomaly", outcome)
	}
	if err := manager.Wait(context.Background()); err != nil {
		t.Fatalf("Manager.Wait() error = %v", err)
	}
	if got := manager.State(); got != sessionmanager.StateClosed {
		t.Fatalf("Manager.State() = %d, want StateClosed", got)
	}
}

func TestTwoConcurrentCompletionsHaveOneMutation(t *testing.T) {
	testConcurrentCompletion(t, 2)
}

func TestMassConcurrentCompletionsHaveOneMutation(t *testing.T) {
	testConcurrentCompletion(t, 128)
}

func TestCopiedCapabilitySharesOneBoundCompletion(t *testing.T) {
	manager := sessionmanager.New()
	registrationID := committedRegistration(t, manager, "session-1")
	original := newAdapter(t, manager, registrationID)
	copy1 := original
	copy2 := copy1
	adapters := []executionowner.CompletionAdapter{original, copy1, copy2}

	outcomes := runConcurrentCompletions(96, func(index int) executionowner.CompleteOutcome {
		return adapters[index%len(adapters)].CompleteBoundRegistration()
	})
	assertCompletionOutcomes(t, outcomes)
	assertRegistrationAbsent(t, manager, "session-1")
}

func TestCompletionAndLookupLinearizeAtomically(t *testing.T) {
	manager := sessionmanager.New()
	registrationID := committedRegistration(t, manager, "session-1")
	adapter := newAdapter(t, manager, registrationID)
	start := make(chan struct{})
	completionResult := make(chan executionowner.CompleteOutcome, 1)
	type lookupResult struct {
		view  sessionmanager.RegistrationView
		found bool
	}
	lookupResults := make(chan lookupResult, 1)
	var ready sync.WaitGroup
	var finished sync.WaitGroup
	ready.Add(2)
	finished.Add(2)

	go func() {
		defer finished.Done()
		ready.Done()
		<-start
		completionResult <- adapter.CompleteBoundRegistration()
	}()
	go func() {
		defer finished.Done()
		ready.Done()
		<-start
		view, found := manager.Lookup("session-1")
		lookupResults <- lookupResult{view: view, found: found}
	}()

	ready.Wait()
	close(start)
	outcome := <-completionResult
	lookup := <-lookupResults
	finished.Wait()
	if outcome != executionowner.CompleteOutcomeCompleted {
		t.Fatalf("CompleteBoundRegistration() = %d, want CompleteOutcomeCompleted", outcome)
	}
	if lookup.found {
		if lookup.view.RegistrationID() != registrationID {
			t.Fatalf("Lookup RegistrationID = %+v, want %+v", lookup.view.RegistrationID(), registrationID)
		}
	} else if lookup.view != (sessionmanager.RegistrationView{}) {
		t.Fatalf("Lookup absent View = %+v, want zero", lookup.view)
	}
	assertRegistrationAbsent(t, manager, "session-1")
}

func TestCompletionDoesNotChangeOwnerLifecycleOrControlCell(t *testing.T) {
	manager := sessionmanager.New()
	registrationID := committedRegistration(t, manager, "session-1")
	adapter := newAdapter(t, manager, registrationID)
	owner := executionowner.New()

	if outcome := adapter.CompleteBoundRegistration(); outcome != executionowner.CompleteOutcomeCompleted {
		t.Fatalf("CompleteBoundRegistration() = %d, want CompleteOutcomeCompleted", outcome)
	}
	if got := owner.State(); got != executionowner.StatePreCommit {
		t.Fatalf("Owner.State() = %d, want StatePreCommit", got)
	}
	if owner.StopRequested() {
		t.Fatal("Owner.StopRequested() = true after completion")
	}
}

func testConcurrentCompletion(t *testing.T, callers int) {
	t.Helper()

	manager := sessionmanager.New()
	registrationID := committedRegistration(t, manager, "session-1")
	adapter := newAdapter(t, manager, registrationID)
	outcomes := runConcurrentCompletions(callers, func(int) executionowner.CompleteOutcome {
		return adapter.CompleteBoundRegistration()
	})

	assertCompletionOutcomes(t, outcomes)
	assertRegistrationAbsent(t, manager, "session-1")
}

func runConcurrentCompletions(
	callers int,
	complete func(int) executionowner.CompleteOutcome,
) []executionowner.CompleteOutcome {
	start := make(chan struct{})
	outcomes := make([]executionowner.CompleteOutcome, callers)
	var ready sync.WaitGroup
	var finished sync.WaitGroup
	ready.Add(callers)
	finished.Add(callers)
	for index := range callers {
		go func() {
			defer finished.Done()
			ready.Done()
			<-start
			outcomes[index] = complete(index)
		}()
	}
	ready.Wait()
	close(start)
	finished.Wait()
	return outcomes
}

func assertCompletionOutcomes(t *testing.T, outcomes []executionowner.CompleteOutcome) {
	t.Helper()

	completed := 0
	anomalies := 0
	for _, outcome := range outcomes {
		switch outcome {
		case executionowner.CompleteOutcomeCompleted:
			completed++
		case executionowner.CompleteOutcomeAccountingAnomaly:
			anomalies++
		default:
			t.Errorf("CompleteBoundRegistration() = %d, want a defined CompleteOutcome", outcome)
		}
	}
	if completed != 1 {
		t.Fatalf("Completed outcomes = %d, want 1", completed)
	}
	if anomalies != len(outcomes)-1 {
		t.Fatalf("AccountingAnomaly outcomes = %d, want %d", anomalies, len(outcomes)-1)
	}
}

func newAdapter(
	t *testing.T,
	manager *sessionmanager.Manager,
	registrationID sessionmanager.RegistrationID,
) executionowner.CompletionAdapter {
	t.Helper()

	adapter, err := New(manager, registrationID)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return adapter
}

func committedRegistration(
	t *testing.T,
	manager *sessionmanager.Manager,
	sessionID sessionmanager.SessionID,
) sessionmanager.RegistrationID {
	t.Helper()

	reservation, err := manager.Reserve(sessionID)
	if err != nil {
		t.Fatalf("Manager.Reserve(%q) error = %v", sessionID, err)
	}
	result, err := reservation.Commit()
	if err != nil {
		t.Fatalf("Reservation.Commit() error = %v", err)
	}
	if outcome := result.LifetimeLease().Release(); outcome != lifetimelease.ReleaseOutcomeReleased {
		t.Fatalf("LifetimeLease.Release() = %d, want ReleaseOutcomeReleased", outcome)
	}
	return result.RegistrationID()
}

func assertRegistrationPresent(
	t *testing.T,
	manager *sessionmanager.Manager,
	sessionID sessionmanager.SessionID,
	registrationID sessionmanager.RegistrationID,
) {
	t.Helper()

	view, found := manager.Lookup(sessionID)
	if !found || view.RegistrationID() != registrationID {
		t.Fatalf("Manager.Lookup(%q) = (%+v, %t), want RegistrationID %+v", sessionID, view, found, registrationID)
	}
}

func assertRegistrationAbsent(
	t *testing.T,
	manager *sessionmanager.Manager,
	sessionID sessionmanager.SessionID,
) {
	t.Helper()

	view, found := manager.Lookup(sessionID)
	if found || view != (sessionmanager.RegistrationView{}) {
		t.Fatalf("Manager.Lookup(%q) = (%+v, %t), want zero and false", sessionID, view, found)
	}
}
