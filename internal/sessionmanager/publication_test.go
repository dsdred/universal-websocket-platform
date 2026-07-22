package sessionmanager

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/executionbinding"
	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
	"github.com/dsdred/universal-websocket-platform/internal/lifetimelease"
)

func TestCommitPublishesCompleteBoundResult(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	reservedID := reservationFromHandle(t, handle).registrationID
	owner, handoff, input := newPublicationInput(t)

	result, err := handle.Commit(input)
	if err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	if result.RegistrationID() != reservedID {
		t.Fatalf("RegistrationID() = %+v, want %+v", result.RegistrationID(), reservedID)
	}
	if result.CompletionAdapter() == nil || result.LifetimeLease() == nil {
		t.Fatal("CommitResult contains a nil bound capability")
	}

	manager.mu.RLock()
	registration := manager.registrations[reservedID]
	if registration == nil || registration.stop != owner {
		manager.mu.RUnlock()
		t.Fatal("committed Registration does not store the exact Stop capability")
	}
	manager.mu.RUnlock()
	assertLifetimeLeaseCount(t, manager, 1)

	observed, err := handoff.Waiter().Wait()
	if err != nil || !observed.Committed() {
		t.Fatalf("CommitHandoff.Wait() = (%+v, %v), want Committed", observed, err)
	}
	publishedResult, ok := observed.CommitResult()
	if !ok || publishedResult.RegistrationID() != result.RegistrationID() ||
		publishedResult.CompletionAdapter() != result.CompletionAdapter() ||
		publishedResult.LifetimeLease() != result.LifetimeLease() {
		t.Fatal("CommitHandoff did not publish the exact logical CommitResult")
	}
	if err := handoff.NotCommittedPublisher().Publish(); !errors.Is(err, ErrCommitHandoffAlreadyPublished) {
		t.Fatalf("NotCommitted after Commit error = %v, want ErrCommitHandoffAlreadyPublished", err)
	}
	if outcome := result.CompletionAdapter().CompleteBoundRegistration(); outcome != executionowner.CompleteOutcomeCompleted {
		t.Fatalf("CompletionAdapter outcome = %d, want CompleteOutcomeCompleted", outcome)
	}
	assertRegistrationCount(t, manager, 0)
	assertLifetimeLeaseCount(t, manager, 1)
	if outcome := result.LifetimeLease().Release(); outcome != lifetimelease.ReleaseOutcomeReleased {
		t.Fatalf("LifetimeLease.Release() = %d, want ReleaseOutcomeReleased", outcome)
	}
}

func TestCommitResultExposesOnlyOwnerBoundResult(t *testing.T) {
	resultType := reflect.TypeOf(CommitResult{})
	if resultType.NumField() != 3 {
		t.Fatalf("CommitResult field count = %d, want 3", resultType.NumField())
	}
	wantMethods := map[string]struct{}{
		"CompletionAdapter": {},
		"LifetimeLease":     {},
		"RegistrationID":    {},
	}
	if resultType.NumMethod() != len(wantMethods) {
		t.Fatalf("CommitResult method count = %d, want %d", resultType.NumMethod(), len(wantMethods))
	}
	for index := range resultType.NumMethod() {
		method := resultType.Method(index)
		if _, exists := wantMethods[method.Name]; !exists {
			t.Fatalf("CommitResult exposes unexpected method %q", method.Name)
		}
	}
}

func TestCommitInputContainsExactlyRequiredPublicationCapabilities(t *testing.T) {
	inputType := reflect.TypeOf(CommitInput{})
	if inputType.NumField() != 2 {
		t.Fatalf("CommitInput field count = %d, want 2", inputType.NumField())
	}
	wantTypes := []reflect.Type{
		reflect.TypeOf((*StopPublicationBinding)(nil)).Elem(),
		reflect.TypeOf(CommitHandoffPublisher{}),
	}
	for index, wantType := range wantTypes {
		field := inputType.Field(index)
		if field.IsExported() {
			t.Fatalf("CommitInput field %q is exported", field.Name)
		}
		if field.Type != wantType {
			t.Fatalf("CommitInput field %q type = %v, want %v", field.Name, field.Type, wantType)
		}
	}
}

func TestManagerCreatedCompletionAdapterCannotCompleteAnotherRegistration(t *testing.T) {
	manager := New()
	firstHandle := mustReserve(t, manager, "session-1")
	_, _, firstInput := newPublicationInput(t)
	first, err := firstHandle.Commit(firstInput)
	if err != nil {
		t.Fatalf("first Commit() error = %v", err)
	}
	secondHandle := mustReserve(t, manager, "session-2")
	_, _, secondInput := newPublicationInput(t)
	second, err := secondHandle.Commit(secondInput)
	if err != nil {
		t.Fatalf("second Commit() error = %v", err)
	}

	if outcome := first.CompletionAdapter().CompleteBoundRegistration(); outcome != executionowner.CompleteOutcomeCompleted {
		t.Fatalf("first CompletionAdapter outcome = %d, want CompleteOutcomeCompleted", outcome)
	}
	if view, found := manager.Lookup("session-2"); !found || view.RegistrationID() != second.RegistrationID() {
		t.Fatalf("second Lookup() = (%+v, %t), want RegistrationID %+v", view, found, second.RegistrationID())
	}
}

func TestConcurrentCommitAndLookupObserveNoPartialRegistration(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		handle := mustReserve(t, manager, "session-1")
		owner, _, input := newPublicationInput(t)
		start := make(chan struct{})
		commits := make(chan commitResultWithBundle, 1)
		type lookupOutcome struct {
			view  RegistrationView
			found bool
		}
		lookups := make(chan lookupOutcome, 1)
		var ready sync.WaitGroup
		ready.Add(2)

		go func() {
			ready.Done()
			<-start
			result, err := handle.Commit(input)
			commits <- commitResultWithBundle{result: result, err: err}
		}()
		go func() {
			ready.Done()
			<-start
			view, found := manager.Lookup("session-1")
			lookups <- lookupOutcome{view: view, found: found}
		}()
		ready.Wait()
		close(start)

		commit := <-commits
		lookup := <-lookups
		if commit.err != nil {
			t.Fatalf("iteration %d: Commit() error = %v", iteration, commit.err)
		}
		if lookup.found && lookup.view.RegistrationID() != commit.result.RegistrationID() {
			t.Fatalf("iteration %d: Lookup ID = %+v, want %+v", iteration, lookup.view.RegistrationID(), commit.result.RegistrationID())
		}
		manager.mu.RLock()
		registration := manager.registrations[commit.result.RegistrationID()]
		complete := registration != nil && registration.stop == owner &&
			registration.commitResult.CompletionAdapter() == commit.result.CompletionAdapter() &&
			registration.commitResult.LifetimeLease() == commit.result.LifetimeLease()
		manager.mu.RUnlock()
		if !complete {
			t.Fatalf("iteration %d: committed Registration contains a partial bundle", iteration)
		}
	}
}

func TestConcurrentCommitPublishesOneLogicalResultThroughOneHandoff(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	const callers = 32
	_, handoff, input := newPublicationInput(t)
	start := make(chan struct{})
	results := make(chan commitResultWithBundle, callers)
	var ready sync.WaitGroup
	ready.Add(callers)
	for range callers {
		go func() {
			ready.Done()
			<-start
			result, err := handle.Commit(input)
			results <- commitResultWithBundle{result: result, err: err}
		}()
	}
	ready.Wait()
	close(start)

	var first CommitResult
	for range callers {
		observed := <-results
		if observed.err != nil {
			t.Fatalf("concurrent Commit() error = %v", observed.err)
		}
		if first.RegistrationID() == (RegistrationID{}) {
			first = observed.result
			continue
		}
		if observed.result.RegistrationID() != first.RegistrationID() ||
			observed.result.CompletionAdapter() != first.CompletionAdapter() ||
			observed.result.LifetimeLease() != first.LifetimeLease() {
			t.Fatal("concurrent Commit returned a different logical bundle")
		}
	}

	if err := handoff.NotCommittedPublisher().Publish(); !errors.Is(err, ErrCommitHandoffAlreadyPublished) {
		t.Fatalf("NotCommitted publication after Commit error = %v, want %v", err, ErrCommitHandoffAlreadyPublished)
	}
	outcome, err := handoff.Waiter().Wait()
	if err != nil {
		t.Fatalf("CommitHandoff.Wait() error = %v", err)
	}
	published, ok := outcome.CommitResult()
	if !ok || published.RegistrationID() != first.RegistrationID() ||
		published.CompletionAdapter() != first.CompletionAdapter() ||
		published.LifetimeLease() != first.LifetimeLease() {
		t.Fatal("committed Handoff contains a different logical result")
	}
	assertRegistrationCount(t, manager, 1)
	assertLifetimeLeaseCount(t, manager, 1)
}

func TestRepeatedCommitRequiresSameHandoffIdentity(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	owner, handoff, input := newPublicationInput(t)
	first, err := handle.Commit(input)
	if err != nil {
		t.Fatalf("first Commit() error = %v", err)
	}

	repeated, err := handle.Commit(input)
	if err != nil {
		t.Fatalf("repeated Commit() with same Handoff error = %v", err)
	}
	if repeated.RegistrationID() != first.RegistrationID() ||
		repeated.CompletionAdapter() != first.CompletionAdapter() ||
		repeated.LifetimeLease() != first.LifetimeLease() {
		t.Fatal("repeated Commit did not return the stored logical result")
	}
	manager.mu.RLock()
	storedStop := manager.registrations[first.RegistrationID()].stop
	manager.mu.RUnlock()
	if storedStop != owner {
		t.Fatal("repeated Commit replaced the stored Stop capability")
	}
	if err := handoff.NotCommittedPublisher().Publish(); !errors.Is(err, ErrCommitHandoffAlreadyPublished) {
		t.Fatalf("original Handoff NotCommitted error = %v, want already published", err)
	}

	_, _, differentInput := newPublicationInput(t)
	different, err := handle.Commit(differentInput)
	if different != (CommitResult{}) || !errors.Is(err, ErrInvalidCommitInput) {
		t.Fatalf("Commit() with different Handoff = (%+v, %v), want zero and ErrInvalidCommitInput", different, err)
	}
	assertLifetimeLeaseCount(t, manager, 1)
}

func TestConcurrentRepeatedCommitWithBlockedWaiterPreservesHandoffIdentity(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	_, handoff, input := newPublicationInput(t)
	_, _, differentInput := newPublicationInput(t)
	waiterStart := make(chan struct{})
	waiterResults := make(chan commitResultWithBundle, 2)
	var waiterReady sync.WaitGroup
	waiterReady.Add(2)
	for range 2 {
		go func() {
			waiterReady.Done()
			<-waiterStart
			outcome, waitErr := handoff.Waiter().Wait()
			if waitErr != nil {
				waiterResults <- commitResultWithBundle{err: waitErr}
				return
			}
			result, ok := outcome.CommitResult()
			if !ok {
				waiterResults <- commitResultWithBundle{err: ErrInvalidCommitHandoff}
				return
			}
			waiterResults <- commitResultWithBundle{result: result}
		}()
	}
	waiterReady.Wait()
	close(waiterStart)
	losingWaiter := <-waiterResults
	if !errors.Is(losingWaiter.err, executionbinding.ErrWaitAlreadyClaimed) {
		t.Fatalf("losing Wait() error = %v, want ErrWaitAlreadyClaimed", losingWaiter.err)
	}

	first, err := handle.Commit(input)
	if err != nil {
		t.Fatalf("initial Commit() error = %v", err)
	}
	const sameCallers = 16
	start := make(chan struct{})
	results := make(chan commitResultWithBundle, sameCallers+1)
	var ready sync.WaitGroup
	ready.Add(sameCallers + 1)
	for range sameCallers {
		go func() {
			ready.Done()
			<-start
			result, commitErr := handle.Commit(input)
			results <- commitResultWithBundle{result: result, err: commitErr}
		}()
	}
	go func() {
		ready.Done()
		<-start
		result, commitErr := handle.Commit(differentInput)
		results <- commitResultWithBundle{result: result, err: commitErr}
	}()
	ready.Wait()
	close(start)

	differentFailures := 0
	for range sameCallers + 1 {
		observed := <-results
		if observed.err != nil {
			if !errors.Is(observed.err, ErrInvalidCommitInput) ||
				!errors.Is(observed.err, ErrInvalidCommitHandoff) {
				t.Fatalf("repeated Commit() error = %v", observed.err)
			}
			differentFailures++
			continue
		}
		if observed.result.RegistrationID() != first.RegistrationID() ||
			observed.result.CompletionAdapter() != first.CompletionAdapter() ||
			observed.result.LifetimeLease() != first.LifetimeLease() {
			t.Fatal("same-identity repeated Commit returned a different logical result")
		}
	}
	if differentFailures != 1 {
		t.Fatalf("different-identity failures = %d, want 1", differentFailures)
	}
	waited := <-waiterResults
	if waited.err != nil || waited.result.RegistrationID() != first.RegistrationID() ||
		waited.result.CompletionAdapter() != first.CompletionAdapter() ||
		waited.result.LifetimeLease() != first.LifetimeLease() {
		t.Fatalf("waiter result = (%+v, %v), want exact initial CommitResult", waited.result, waited.err)
	}
	assertRegistrationCount(t, manager, 1)
	assertLifetimeLeaseCount(t, manager, 1)
}

func TestInvalidCommitInputPublishesNothing(t *testing.T) {
	tests := []struct {
		name  string
		input func(*testing.T) CommitInput
	}{
		{name: "zero input", input: func(*testing.T) CommitInput { return CommitInput{} }},
		{name: "missing Stop", input: func(t *testing.T) CommitInput {
			handoff := NewCommitHandoff()
			return CommitInput{publisher: handoff.CommitPublisher()}
		}},
		{name: "stale publisher", input: func(t *testing.T) CommitInput {
			handoff := NewCommitHandoff()
			if err := handoff.NotCommittedPublisher().Publish(); err != nil {
				t.Fatalf("Publish NotCommitted error = %v", err)
			}
			return CommitInput{stop: executionowner.New(), publisher: handoff.CommitPublisher()}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := New()
			handle := mustReserve(t, manager, "session-1")
			result, err := handle.Commit(test.input(t))
			if result != (CommitResult{}) || !errors.Is(err, ErrInvalidCommitInput) {
				t.Fatalf("Commit() = (%+v, %v), want zero and ErrInvalidCommitInput", result, err)
			}
			assertRegistrationCount(t, manager, 0)
			assertLifetimeLeaseCount(t, manager, 0)
			if view, found := manager.Lookup("session-1"); found || view != (RegistrationView{}) {
				t.Fatalf("Lookup() = (%+v, %t), want zero and false", view, found)
			}
			handle.AbortUnlessCommitted()
			assertSnapshotCount(t, manager.BeginShutdown(), 0)
		})
	}
}

func TestMissingStopDoesNotCommitSuppliedExecutionPublisher(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	handoff := NewCommitHandoff()
	input := CommitInput{publisher: handoff.CommitPublisher()}

	result, err := handle.Commit(input)
	if result != (CommitResult{}) || !errors.Is(err, ErrInvalidCommitInput) {
		t.Fatalf("Commit() = (%+v, %v), want zero and ErrInvalidCommitInput", result, err)
	}
	if err := handoff.CommitPublisher().validateFresh(); err != nil {
		t.Fatalf("rejected Commit changed supplied publisher: %v", err)
	}
	assertRegistrationCount(t, manager, 0)
	assertLifetimeLeaseCount(t, manager, 0)
}

func TestCommitAndAbortPublishExactlyOneTerminalOutcome(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		handle := mustReserve(t, manager, "session-1")
		_, handoff, input := newPublicationInput(t)
		start := make(chan struct{})
		commits := make(chan commitResultWithBundle, 1)
		abortDone := make(chan struct{})
		var ready sync.WaitGroup
		ready.Add(2)

		go func() {
			ready.Done()
			<-start
			result, err := handle.Commit(input)
			commits <- commitResultWithBundle{result: result, err: err}
		}()
		go func() {
			ready.Done()
			<-start
			handle.Abort()
			close(abortDone)
		}()
		ready.Wait()
		close(start)

		commit := <-commits
		<-abortDone
		switch {
		case commit.err == nil:
			assertRegistrationCount(t, manager, 1)
			assertLifetimeLeaseCount(t, manager, 1)
			if err := handoff.NotCommittedPublisher().Publish(); !errors.Is(err, ErrCommitHandoffAlreadyPublished) {
				t.Fatalf("iteration %d: Handoff publication error = %v, want already published", iteration, err)
			}
		case errors.Is(commit.err, ErrReservationAborted):
			assertRegistrationCount(t, manager, 0)
			assertLifetimeLeaseCount(t, manager, 0)
			if err := handoff.CommitPublisher().validateFresh(); err != nil {
				t.Fatalf("iteration %d: aborted Commit changed Handoff: %v", iteration, err)
			}
		default:
			t.Fatalf("iteration %d: Commit() error = %v", iteration, commit.err)
		}
	}
}

func TestCommitAndBeginShutdownPublishCapabilityBundleAtomically(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		handle := mustReserve(t, manager, "session-1")
		owner, handoff, input := newPublicationInput(t)
		start := make(chan struct{})
		commits := make(chan commitResultWithBundle, 1)
		snapshots := make(chan ShutdownSnapshot, 1)
		var ready sync.WaitGroup
		ready.Add(2)

		go func() {
			ready.Done()
			<-start
			result, err := handle.Commit(input)
			commits <- commitResultWithBundle{result: result, err: err}
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
			registrations := snapshot.Registrations()
			if len(registrations) != 1 || registrations[0].RegistrationID() != commit.result.RegistrationID() {
				t.Fatalf("iteration %d: Snapshot does not contain committed Registration", iteration)
			}
			if !registrations[0].RequestStop() || !owner.StopRequested() {
				t.Fatalf("iteration %d: Snapshot does not hold the committed Stop capability", iteration)
			}
			if err := handoff.NotCommittedPublisher().Publish(); !errors.Is(err, ErrCommitHandoffAlreadyPublished) {
				t.Fatalf("iteration %d: Handoff publication error = %v, want already published", iteration, err)
			}
		case errors.Is(commit.err, ErrManagerNotOpen):
			assertSnapshotCount(t, snapshot, 0)
			if err := handoff.CommitPublisher().validateFresh(); err != nil {
				t.Fatalf("iteration %d: rejected Commit changed Handoff: %v", iteration, err)
			}
			if owner.StopRequested() {
				t.Fatalf("iteration %d: rejected Commit invoked Stop", iteration)
			}
			handle.AbortUnlessCommitted()
		default:
			t.Fatalf("iteration %d: Commit() error = %v", iteration, commit.err)
		}
	}
}

func TestCapturedSnapshotKeepsExactStopCapabilityAfterComplete(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	owner, _, input := newPublicationInput(t)
	result, err := handle.Commit(input)
	if err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	snapshot := manager.BeginShutdown()
	registration := snapshot.Registrations()[0]
	copy := registration

	if !manager.Complete(result.RegistrationID()) {
		t.Fatal("Complete() = false, want true")
	}
	if !registration.RequestStop() {
		t.Fatal("captured RequestStop() = false, want true")
	}
	if copy.RequestStop() {
		t.Fatal("copied RequestStop() = true after the shared first request")
	}
	if !owner.StopRequested() {
		t.Fatal("Owner does not observe the Snapshot Stop request")
	}
	if outcome := result.LifetimeLease().Release(); outcome != lifetimelease.ReleaseOutcomeReleased {
		t.Fatalf("LifetimeLease.Release() = %d, want ReleaseOutcomeReleased", outcome)
	}
}

type commitResultWithBundle struct {
	result CommitResult
	err    error
}

func newPublicationInput(t *testing.T) (*executionowner.Owner, *CommitHandoff, CommitInput) {
	t.Helper()
	owner := executionowner.New()
	handoff := NewCommitHandoff()
	input, err := NewCommitInput(owner, handoff.CommitPublisher())
	if err != nil {
		t.Fatalf("NewCommitInput() error = %v", err)
	}
	return owner, handoff, input
}
