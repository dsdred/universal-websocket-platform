package sessionmanager

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestCommitHandoffPublisherIsOpaqueOutsideSessionManager(t *testing.T) {
	publisherType := reflect.TypeOf(CommitHandoffPublisher{})
	if publisherType.NumMethod() != 0 {
		methods := make([]string, 0, publisherType.NumMethod())
		for index := range publisherType.NumMethod() {
			methods = append(methods, publisherType.Method(index).Name)
		}
		t.Fatalf("CommitHandoffPublisher exported methods = %v, want none", methods)
	}
}

func TestCommittedHandoffOutcomeIsPublishedOnlyByReservationCommit(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	_, handoff, input := newPublicationInput(t)

	result, err := handle.Commit(input)
	if err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	outcome, err := handoff.Waiter().Wait()
	if err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	published, ok := outcome.CommitResult()
	if !ok || published.RegistrationID() != result.RegistrationID() ||
		published.CompletionAdapter() != result.CompletionAdapter() ||
		published.LifetimeLease() != result.LifetimeLease() {
		t.Fatal("Reservation Commit did not publish its exact logical CommitResult")
	}
}

func TestCommitHandoffPublishesNotCommittedWithoutCommittedCapabilities(t *testing.T) {
	handoff := NewCommitHandoff()
	copied := *handoff

	if err := handoff.NotCommittedPublisher().Publish(); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if err := copied.NotCommittedPublisher().Publish(); err != nil {
		t.Fatalf("repeated Publish() error = %v", err)
	}
	outcome, err := copied.Waiter().Wait()
	if err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	if outcome.Committed() {
		t.Fatal("NotCommitted outcome reports Committed")
	}
	if result, ok := outcome.CommitResult(); ok || result != (CommitResult{}) {
		t.Fatalf("CommitResult() = (%+v, %t), want zero and false", result, ok)
	}
}

func TestCommitHandoffRejectsUninitializedFacets(t *testing.T) {
	var handoff CommitHandoff
	if err := handoff.NotCommittedPublisher().Publish(); !errors.Is(err, ErrInvalidCommitHandoff) {
		t.Fatalf("Publish() error = %v, want %v", err, ErrInvalidCommitHandoff)
	}
	if _, err := handoff.Waiter().Wait(); !errors.Is(err, ErrInvalidCommitHandoff) {
		t.Fatalf("Wait() error = %v, want %v", err, ErrInvalidCommitHandoff)
	}
	if err := handoff.CommitPublisher().validateFresh(); !errors.Is(err, ErrInvalidCommitHandoff) {
		t.Fatalf("validateFresh() error = %v, want %v", err, ErrInvalidCommitHandoff)
	}
}

func TestCommitAndNotCommittedPublicationHaveOneAtomicWinner(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		handle := mustReserve(t, manager, "session-1")
		_, handoff, input := newPublicationInput(t)
		start := make(chan struct{})
		commitResult := make(chan commitResultWithBundle, 1)
		notCommittedResult := make(chan error, 1)
		var ready sync.WaitGroup
		ready.Add(2)
		go func() {
			ready.Done()
			<-start
			result, err := handle.Commit(input)
			commitResult <- commitResultWithBundle{result: result, err: err}
		}()
		go func() {
			ready.Done()
			<-start
			notCommittedResult <- handoff.NotCommittedPublisher().Publish()
		}()
		ready.Wait()
		close(start)

		committed := <-commitResult
		notCommittedErr := <-notCommittedResult
		outcome, err := handoff.Waiter().Wait()
		if err != nil {
			t.Fatalf("iteration %d: Wait() error = %v", iteration, err)
		}
		if committed.err == nil {
			if !errors.Is(notCommittedErr, ErrCommitHandoffAlreadyPublished) {
				t.Fatalf("iteration %d: NotCommitted error = %v, want already published", iteration, notCommittedErr)
			}
			published, ok := outcome.CommitResult()
			if !ok || published.RegistrationID() != committed.result.RegistrationID() {
				t.Fatalf("iteration %d: committed outcome does not carry Commit result", iteration)
			}
			assertRegistrationCount(t, manager, 1)
			assertLifetimeLeaseCount(t, manager, 1)
			continue
		}

		if !errors.Is(committed.err, ErrInvalidCommitInput) || notCommittedErr != nil {
			t.Fatalf("iteration %d: Commit error = %v, NotCommitted error = %v", iteration, committed.err, notCommittedErr)
		}
		if outcome.Committed() {
			t.Fatalf("iteration %d: losing Commit published committed capabilities", iteration)
		}
		handle.AbortUnlessCommitted()
		assertRegistrationCount(t, manager, 0)
		assertLifetimeLeaseCount(t, manager, 0)
	}
}
