package executionowner_test

import (
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
)

func TestCreateOwnerSuccess(t *testing.T) {
	owner := executionowner.New()
	if owner == nil {
		t.Fatal("New() returned a nil Owner")
	}
}

func TestInitialState(t *testing.T) {
	owner := executionowner.New()
	if got := owner.State(); got != executionowner.StatePreCommit {
		t.Fatalf("State() = %d, want StatePreCommit", got)
	}
}

func TestZeroValueNotUsable(t *testing.T) {
	var owner executionowner.Owner

	got := owner.State()
	validStates := []executionowner.State{
		executionowner.StatePreCommit,
		executionowner.StateCommitted,
		executionowner.StateStarting,
		executionowner.StateRunning,
		executionowner.StateTerminalizing,
		executionowner.StateTerminal,
	}
	for _, state := range validStates {
		if got == state {
			t.Fatalf("zero-value Owner reports valid lifecycle state %d", got)
		}
	}
}
