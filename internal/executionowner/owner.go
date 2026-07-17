// Package executionowner defines the per-Session execution ownership boundary.
package executionowner

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

// Owner represents one per-Session execution lifecycle.
//
// This skeleton exposes lifecycle identity only. It does not start execution or
// own Session, Runtime callback, registration, or lease resources.
type Owner struct {
	state State
}

// New creates a dormant Owner in the PreCommit state.
func New() *Owner {
	return &Owner{state: StatePreCommit}
}

// State returns the current lifecycle state.
func (owner *Owner) State() State {
	return owner.state
}
