package executionowner

import (
	"context"
	"errors"
)

// ErrInvalidExecutionEnvironment indicates that the prepared root
// observation or derived execution dependency is incomplete.
var ErrInvalidExecutionEnvironment = errors.New("invalid execution environment")

// ExecutionEnvironment is the immutable boundary between the Host-owned root
// Runtime observation and the derived per-connection execution lifecycle.
// It exposes neither root cancellation authority nor mutable fields.
type ExecutionEnvironment struct {
	root      context.Context
	execution context.Context
	cancel    context.CancelFunc
}

// NewExecutionEnvironment validates and captures the two distinct context
// roles prepared before Commit.
func NewExecutionEnvironment(
	root context.Context,
	execution context.Context,
	cancel context.CancelFunc,
) (ExecutionEnvironment, error) {
	if root == nil || execution == nil || cancel == nil {
		return ExecutionEnvironment{}, ErrInvalidExecutionEnvironment
	}
	return ExecutionEnvironment{
		root:      root,
		execution: execution,
		cancel:    cancel,
	}, nil
}

func (environment ExecutionEnvironment) valid() bool {
	return environment.root != nil && environment.execution != nil && environment.cancel != nil
}
