// Package executionowner defines the per-Session execution ownership boundary.
package executionowner

import (
	"fmt"
	"sync"
)

type ownerError string

const (
	// ErrInvalidTransition indicates that a lifecycle transition is not allowed
	// from the state observed at its linearization point.
	ErrInvalidTransition ownerError = "invalid execution owner lifecycle transition"
	// ErrUninitializedOwner indicates that an Owner was not created with New.
	ErrUninitializedOwner ownerError = "execution owner is uninitialized"
)

func (err ownerError) Error() string {
	return string(err)
}

// State is the read-only lifecycle state of an Owner.
type State uint8

const (
	// StatePreCommit is the initial dormant state before successful Commit.
	StatePreCommit State = iota + 1
	// StateCommitted begins at the successful Commit publication boundary.
	StateCommitted
	// StateStarting marks the Session Start linearization interval.
	StateStarting
	// StateRunning marks active Session execution.
	StateRunning
	// StateTerminalizing marks the single terminal cleanup path.
	StateTerminalizing
	// StateTerminal is the terminal lifecycle state.
	StateTerminal
)

// Owner represents one per-Session execution lifecycle. After Commit, Execute
// is the exclusive lifecycle writer and external control is limited to
// termination intent.
type Owner struct {
	state *ownerState
}

type ownerState struct {
	mu      sync.RWMutex
	current State
	control controlCell
}

// New creates a dormant Owner in the PreCommit state.
func New() *Owner {
	return &Owner{
		state: &ownerState{
			current: StatePreCommit,
			control: newControlCell(),
		},
	}
}

// State returns the current lifecycle state.
func (owner *Owner) State() State {
	if owner == nil || owner.state == nil {
		return 0
	}

	state := owner.state
	state.mu.RLock()
	defer state.mu.RUnlock()

	return state.current
}

// RequestStop records the first Stop request for this Owner. It does not
// transition the lifecycle or perform Session work.
func (owner *Owner) RequestStop() bool {
	return owner.requestStop()
}

// StopRequested reports whether this Owner has accepted a Stop request.
func (owner *Owner) StopRequested() bool {
	if owner == nil || owner.state == nil {
		return false
	}

	state := owner.state
	state.mu.RLock()
	defer state.mu.RUnlock()

	return state.control.primary == terminationExplicitStop
}

// Transition publishes the PreCommit-to-Committed ownership boundary.
// Post-Commit lifecycle transitions are internal to the execution path.
func (owner *Owner) Transition(from, to State) error {
	if owner == nil || owner.state == nil {
		return ErrUninitializedOwner
	}

	state := owner.state
	state.mu.Lock()
	defer state.mu.Unlock()

	if state.current != from {
		return fmt.Errorf(
			"%w: current=%s expected=%s requested=%s: current state does not match expected source",
			ErrInvalidTransition,
			stateName(state.current),
			stateName(from),
			stateName(to),
		)
	}
	if from != StatePreCommit || to != StateCommitted {
		return fmt.Errorf(
			"%w: current=%s expected=%s requested=%s: external post-Commit transition is not permitted",
			ErrInvalidTransition,
			stateName(state.current),
			stateName(from),
			stateName(to),
		)
	}

	state.current = to
	return nil
}

func (owner *Owner) transitionLifecycle(from, to State) error {
	if owner == nil || owner.state == nil {
		return ErrUninitializedOwner
	}

	state := owner.state
	state.mu.Lock()
	defer state.mu.Unlock()

	if state.current != from {
		return fmt.Errorf(
			"%w: current=%s expected=%s requested=%s: current state does not match expected source",
			ErrInvalidTransition,
			stateName(state.current),
			stateName(from),
			stateName(to),
		)
	}
	if !transitionAllowed(from, to) {
		return fmt.Errorf(
			"%w: current=%s expected=%s requested=%s: transition is not permitted",
			ErrInvalidTransition,
			stateName(state.current),
			stateName(from),
			stateName(to),
		)
	}

	state.current = to
	return nil
}

func transitionAllowed(from, to State) bool {
	switch from {
	case StatePreCommit:
		return to == StateCommitted
	case StateCommitted:
		return to == StateStarting || to == StateTerminalizing
	case StateStarting:
		return to == StateRunning || to == StateTerminalizing
	case StateRunning:
		return to == StateTerminalizing
	case StateTerminalizing:
		return to == StateTerminal
	default:
		return false
	}
}

func stateName(state State) string {
	switch state {
	case StatePreCommit:
		return "PreCommit"
	case StateCommitted:
		return "Committed"
	case StateStarting:
		return "Starting"
	case StateRunning:
		return "Running"
	case StateTerminalizing:
		return "Terminalizing"
	case StateTerminal:
		return "Terminal"
	default:
		return fmt.Sprintf("State(%d)", state)
	}
}
