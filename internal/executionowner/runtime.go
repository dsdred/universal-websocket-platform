package executionowner

import (
	"context"
	"errors"
	"fmt"
)

var ErrNilSessionLifecycle = errors.New("execution Session lifecycle is nil")

// SessionLifecycle is the lifecycle-only Session contract used by an Owner.
type SessionLifecycle interface {
	Start(context.Context) error
	Run(context.Context) error
}

// Execute runs the single post-Commit Session execution path in the calling
// goroutine. It stops at Terminalizing; terminal cleanup is a separate concern.
func (owner *Owner) Execute(ctx context.Context, session SessionLifecycle) error {
	if owner == nil || owner.state == nil {
		return ErrUninitializedOwner
	}
	if session == nil {
		return ErrNilSessionLifecycle
	}

	state := owner.state
	state.mu.Lock()
	if state.current != StateCommitted {
		err := invalidTransitionError(state.current, StateCommitted, StateStarting)
		state.mu.Unlock()
		return err
	}
	if state.control.primary == terminationExplicitStop || ctx.Err() != nil {
		state.current = StateTerminalizing
		err := ctx.Err()
		state.mu.Unlock()
		return err
	}
	state.current = StateStarting
	state.mu.Unlock()

	if err := session.Start(ctx); err != nil {
		owner.finishExecution(StateStarting)
		return err
	}

	state.mu.Lock()
	if state.current != StateStarting {
		err := invalidTransitionError(state.current, StateStarting, StateRunning)
		state.mu.Unlock()
		return err
	}
	if state.control.primary == terminationExplicitStop || ctx.Err() != nil {
		state.current = StateTerminalizing
		err := ctx.Err()
		state.mu.Unlock()
		return err
	}
	state.current = StateRunning
	state.mu.Unlock()

	runErr := session.Run(ctx)
	if err := owner.finishExecution(StateRunning); err != nil && runErr == nil {
		return err
	}
	return runErr
}

func (owner *Owner) finishExecution(from State) error {
	state := owner.state
	state.mu.Lock()
	defer state.mu.Unlock()

	if state.current != from {
		return invalidTransitionError(state.current, from, StateTerminalizing)
	}
	state.current = StateTerminalizing
	return nil
}

func invalidTransitionError(current, from, to State) error {
	return fmt.Errorf(
		"%w: current=%s expected=%s requested=%s: current state does not match expected source",
		ErrInvalidTransition,
		stateName(current),
		stateName(from),
		stateName(to),
	)
}
