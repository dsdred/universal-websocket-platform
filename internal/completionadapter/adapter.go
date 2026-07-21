// Package completionadapter binds one Execution Owner completion capability to
// one Session Manager Registration.
package completionadapter

import (
	"reflect"

	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
)

type adapterError string

const ErrInvalidBinding adapterError = "invalid Registration completion binding"

func (err adapterError) Error() string {
	return string(err)
}

// BoundMutation removes one Registration selected before adapter construction.
// It accepts no caller-supplied identity.
type BoundMutation interface {
	CompleteBoundRegistration() bool
}

type adapter struct {
	completion BoundMutation
}

var _ executionowner.CompletionAdapter = (*adapter)(nil)

// New creates a completion capability bound to one Registration identity.
func New(
	completion BoundMutation,
) (executionowner.CompletionAdapter, error) {
	if isNilMutation(completion) {
		return nil, ErrInvalidBinding
	}

	return &adapter{completion: completion}, nil
}

func (adapter *adapter) CompleteBoundRegistration() executionowner.CompleteOutcome {
	if adapter.completion.CompleteBoundRegistration() {
		return executionowner.CompleteOutcomeCompleted
	}

	return executionowner.CompleteOutcomeAccountingAnomaly
}

func isNilMutation(completion BoundMutation) bool {
	if completion == nil {
		return true
	}
	value := reflect.ValueOf(completion)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
