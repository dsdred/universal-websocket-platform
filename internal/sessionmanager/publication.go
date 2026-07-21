package sessionmanager

import (
	"fmt"
	"reflect"

	"github.com/dsdred/universal-websocket-platform/internal/completionadapter"
	"github.com/dsdred/universal-websocket-platform/internal/executionbinding"
	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
)

// StopPublicationBinding is the stable non-owning Stop capability associated
// with one committed Registration.
type StopPublicationBinding interface {
	RequestStop() bool
}

// CommitInput contains exactly the two owner-side publication capabilities
// required by an atomic Commit.
type CommitInput struct {
	stop      StopPublicationBinding
	publisher executionbinding.CommitPublisher
}

// NewCommitInput creates an immutable complete publication input.
func NewCommitInput(
	stop StopPublicationBinding,
	publisher executionbinding.CommitPublisher,
) (CommitInput, error) {
	input := CommitInput{stop: stop, publisher: publisher}
	if err := input.validate(); err != nil {
		return CommitInput{}, err
	}
	return input, nil
}

func (input CommitInput) validate() error {
	if isNilStopBinding(input.stop) {
		return fmt.Errorf("%w: missing Stop publication binding", ErrInvalidCommitInput)
	}
	if err := input.publisher.ValidateFresh(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCommitInput, err)
	}
	return nil
}

func isNilStopBinding(stop StopPublicationBinding) bool {
	if stop == nil {
		return true
	}
	value := reflect.ValueOf(stop)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

type boundCompletionMutation struct {
	manager        *Manager
	registrationID RegistrationID
}

func (mutation *boundCompletionMutation) CompleteBoundRegistration() bool {
	return mutation.manager.Complete(mutation.registrationID)
}

func prepareCommitResult(
	manager *Manager,
	registrationID RegistrationID,
) (CommitResult, error) {
	completion, err := completionadapter.New(&boundCompletionMutation{
		manager:        manager,
		registrationID: registrationID,
	})
	if err != nil {
		return CommitResult{}, err
	}
	lease := &boundLifetimeLease{
		manager:        manager,
		registrationID: registrationID,
	}
	return CommitResult{
		registrationID: registrationID,
		completion:     completion,
		lifetimeLease:  lease,
	}, nil
}

type invalidCompletionAdapter struct{}

func (invalidCompletionAdapter) CompleteBoundRegistration() executionowner.CompleteOutcome {
	return executionowner.CompleteOutcomeAccountingAnomaly
}
