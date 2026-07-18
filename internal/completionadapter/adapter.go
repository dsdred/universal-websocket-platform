// Package completionadapter binds one Execution Owner completion capability to
// one Session Manager Registration.
package completionadapter

import (
	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
	"github.com/dsdred/universal-websocket-platform/internal/sessionmanager"
)

type adapterError string

const ErrInvalidBinding adapterError = "invalid Registration completion binding"

func (err adapterError) Error() string {
	return string(err)
}

type completionMutation interface {
	Complete(sessionmanager.RegistrationID) bool
}

type adapter struct {
	completion     completionMutation
	registrationID sessionmanager.RegistrationID
}

var _ executionowner.CompletionAdapter = (*adapter)(nil)

// New creates a completion capability bound to one Registration identity.
func New(
	manager *sessionmanager.Manager,
	registrationID sessionmanager.RegistrationID,
) (executionowner.CompletionAdapter, error) {
	if manager == nil || registrationID == (sessionmanager.RegistrationID{}) {
		return nil, ErrInvalidBinding
	}

	return &adapter{
		completion:     manager,
		registrationID: registrationID,
	}, nil
}

func (adapter *adapter) CompleteBoundRegistration() executionowner.CompleteOutcome {
	if adapter.completion.Complete(adapter.registrationID) {
		return executionowner.CompleteOutcomeCompleted
	}

	return executionowner.CompleteOutcomeAccountingAnomaly
}
