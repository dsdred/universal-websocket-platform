package executionowner_test

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
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

	if got := owner.State(); got != 0 {
		t.Fatalf("zero-value State() = %d, want invalid zero state", got)
	}
	if err := owner.Transition(executionowner.StatePreCommit, executionowner.StateCommitted); !errors.Is(err, executionowner.ErrUninitializedOwner) {
		t.Fatalf("zero-value Transition() error = %v, want ErrUninitializedOwner", err)
	}
}

func TestNilOwnerNotUsable(t *testing.T) {
	var owner *executionowner.Owner

	if got := owner.State(); got != 0 {
		t.Fatalf("nil State() = %d, want invalid zero state", got)
	}
	if err := owner.Transition(executionowner.StatePreCommit, executionowner.StateCommitted); !errors.Is(err, executionowner.ErrUninitializedOwner) {
		t.Fatalf("nil Transition() error = %v, want ErrUninitializedOwner", err)
	}
}

func TestAllowedTransitions(t *testing.T) {
	tests := []struct {
		name string
		path []executionowner.State
	}{
		{
			name: "normal lifecycle",
			path: []executionowner.State{
				executionowner.StatePreCommit,
				executionowner.StateCommitted,
				executionowner.StateStarting,
				executionowner.StateRunning,
				executionowner.StateTerminalizing,
				executionowner.StateTerminal,
			},
		},
		{
			name: "terminalize before Start",
			path: []executionowner.State{
				executionowner.StatePreCommit,
				executionowner.StateCommitted,
				executionowner.StateTerminalizing,
				executionowner.StateTerminal,
			},
		},
		{
			name: "terminalize after Start linearization",
			path: []executionowner.State{
				executionowner.StatePreCommit,
				executionowner.StateCommitted,
				executionowner.StateStarting,
				executionowner.StateTerminalizing,
				executionowner.StateTerminal,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner := executionowner.New()
			for index := 1; index < len(test.path); index++ {
				from := test.path[index-1]
				to := test.path[index]
				if err := owner.Transition(from, to); err != nil {
					t.Fatalf("Transition(%d, %d) error = %v", from, to, err)
				}
				if got := owner.State(); got != to {
					t.Fatalf("State() = %d, want %d", got, to)
				}
			}
		})
	}
}

func TestInvalidTransitionsDoNotChangeState(t *testing.T) {
	const unknown executionowner.State = 255
	tests := []struct {
		name      string
		current   executionowner.State
		from      executionowner.State
		to        executionowner.State
		wantParts []string
	}{
		{name: "skip state", current: executionowner.StatePreCommit, from: executionowner.StatePreCommit, to: executionowner.StateStarting},
		{name: "move backward", current: executionowner.StateRunning, from: executionowner.StateRunning, to: executionowner.StateStarting},
		{name: "same state", current: executionowner.StateCommitted, from: executionowner.StateCommitted, to: executionowner.StateCommitted},
		{name: "from Terminal", current: executionowner.StateTerminal, from: executionowner.StateTerminal, to: executionowner.StateTerminalizing},
		{name: "unknown target", current: executionowner.StateCommitted, from: executionowner.StateCommitted, to: unknown, wantParts: []string{"Committed", "State(255)"}},
		{name: "unknown expected source", current: executionowner.StatePreCommit, from: unknown, to: executionowner.StateCommitted, wantParts: []string{"PreCommit", "State(255)", "Committed"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner := ownerInState(t, test.current)
			err := owner.Transition(test.from, test.to)
			if !errors.Is(err, executionowner.ErrInvalidTransition) {
				t.Fatalf("Transition(%d, %d) error = %v, want ErrInvalidTransition", test.from, test.to, err)
			}
			for _, part := range test.wantParts {
				if !strings.Contains(err.Error(), part) {
					t.Errorf("Transition() error %q does not contain %q", err, part)
				}
			}
			if got := owner.State(); got != test.current {
				t.Fatalf("State() after rejected transition = %d, want %d", got, test.current)
			}
		})
	}
}

func TestConcurrentIdenticalTransitionHasOneWinner(t *testing.T) {
	owner := executionowner.New()
	const callers = 64

	results := runConcurrentTransitions(callers, func(int) error {
		return owner.Transition(executionowner.StatePreCommit, executionowner.StateCommitted)
	})
	assertOneTransitionWinner(t, results)
	if got := owner.State(); got != executionowner.StateCommitted {
		t.Fatalf("State() = %d, want StateCommitted", got)
	}
}

func TestConcurrentDifferentTransitionsHaveOneValidWinner(t *testing.T) {
	owner := ownerInState(t, executionowner.StateCommitted)
	results := runConcurrentTransitions(2, func(index int) error {
		if index == 0 {
			return owner.Transition(executionowner.StateCommitted, executionowner.StateStarting)
		}
		return owner.Transition(executionowner.StateCommitted, executionowner.StateTerminalizing)
	})
	assertOneTransitionWinner(t, results)

	got := owner.State()
	if got != executionowner.StateStarting && got != executionowner.StateTerminalizing {
		t.Fatalf("State() = %d, want StateStarting or StateTerminalizing", got)
	}
}

func TestConcurrentStateAndTransitionObserveOnlyValidStates(t *testing.T) {
	owner := executionowner.New()
	const readers = 128

	start := make(chan struct{})
	observed := make(chan executionowner.State, readers)
	var waitGroup sync.WaitGroup
	for range readers {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			observed <- owner.State()
		}()
	}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		<-start
		if err := owner.Transition(executionowner.StatePreCommit, executionowner.StateCommitted); err != nil {
			t.Errorf("Transition() error = %v", err)
		}
	}()

	close(start)
	waitGroup.Wait()
	close(observed)
	for state := range observed {
		if state != executionowner.StatePreCommit && state != executionowner.StateCommitted {
			t.Fatalf("State() observed invalid state %d", state)
		}
	}
}

func TestLifecycleTransitionsFromDifferentGoroutines(t *testing.T) {
	owner := executionowner.New()
	path := []executionowner.State{
		executionowner.StatePreCommit,
		executionowner.StateCommitted,
		executionowner.StateStarting,
		executionowner.StateRunning,
		executionowner.StateTerminalizing,
		executionowner.StateTerminal,
	}

	for index := 1; index < len(path); index++ {
		result := make(chan error, 1)
		var finished sync.WaitGroup
		finished.Add(1)
		from := path[index-1]
		to := path[index]
		go func() {
			defer finished.Done()
			result <- owner.Transition(from, to)
		}()
		err := <-result
		finished.Wait()
		if err != nil {
			t.Fatalf("Transition(%d, %d) error = %v", from, to, err)
		}
	}

	if got := owner.State(); got != executionowner.StateTerminal {
		t.Fatalf("State() = %d, want StateTerminal", got)
	}
}

func TestOwnerCopySharesLifecycle(t *testing.T) {
	original := executionowner.New()
	copy := *original

	if err := original.Transition(executionowner.StatePreCommit, executionowner.StateCommitted); err != nil {
		t.Fatalf("Transition() through original error = %v", err)
	}
	if got := copy.State(); got != executionowner.StateCommitted {
		t.Fatalf("copy State() = %d, want StateCommitted", got)
	}
	if err := copy.Transition(executionowner.StateCommitted, executionowner.StateStarting); err != nil {
		t.Fatalf("Transition() through copy error = %v", err)
	}
	if got := original.State(); got != executionowner.StateStarting {
		t.Fatalf("original State() = %d, want StateStarting", got)
	}

	secondCopy := copy
	if err := secondCopy.Transition(executionowner.StateStarting, executionowner.StateRunning); err != nil {
		t.Fatalf("Transition() through second copy error = %v", err)
	}
	if got := original.State(); got != executionowner.StateRunning {
		t.Fatalf("original State() after second-copy transition = %d, want StateRunning", got)
	}
}

func TestConcurrentTransitionsThroughCopiesShareLifecycle(t *testing.T) {
	original := executionowner.New()
	copy := *original

	results := runConcurrentTransitions(2, func(index int) error {
		if index == 0 {
			return original.Transition(executionowner.StatePreCommit, executionowner.StateCommitted)
		}
		return copy.Transition(executionowner.StatePreCommit, executionowner.StateCommitted)
	})
	assertOneTransitionWinner(t, results)
	if original.State() != executionowner.StateCommitted || copy.State() != executionowner.StateCommitted {
		t.Fatalf("copies do not observe one StateCommitted lifecycle")
	}
}

func TestCopyAfterMultipleTransitionsSharesLifecycle(t *testing.T) {
	original := ownerInState(t, executionowner.StateRunning)
	copy := *original

	if err := copy.Transition(executionowner.StateRunning, executionowner.StateTerminalizing); err != nil {
		t.Fatalf("Transition() through late copy error = %v", err)
	}
	if got := original.State(); got != executionowner.StateTerminalizing {
		t.Fatalf("original State() = %d, want StateTerminalizing", got)
	}
}

func TestTerminalIsIrreversibleUnderConcurrency(t *testing.T) {
	owner := ownerInState(t, executionowner.StateTerminal)
	const callers = 64

	results := runConcurrentTransitions(callers, func(index int) error {
		states := []executionowner.State{
			executionowner.StatePreCommit,
			executionowner.StateCommitted,
			executionowner.StateStarting,
			executionowner.StateRunning,
			executionowner.StateTerminalizing,
			executionowner.StateTerminal,
		}
		return owner.Transition(executionowner.StateTerminal, states[index%len(states)])
	})
	for _, err := range results {
		if !errors.Is(err, executionowner.ErrInvalidTransition) {
			t.Errorf("post-Terminal Transition() error = %v, want ErrInvalidTransition", err)
		}
	}
	if got := owner.State(); got != executionowner.StateTerminal {
		t.Fatalf("State() = %d, want immutable StateTerminal", got)
	}
}

func ownerInState(t *testing.T, target executionowner.State) *executionowner.Owner {
	t.Helper()

	owner := executionowner.New()
	path := []executionowner.State{
		executionowner.StatePreCommit,
		executionowner.StateCommitted,
		executionowner.StateStarting,
		executionowner.StateRunning,
		executionowner.StateTerminalizing,
		executionowner.StateTerminal,
	}
	if target == executionowner.StatePreCommit {
		return owner
	}
	for index := 1; index < len(path); index++ {
		if err := owner.Transition(path[index-1], path[index]); err != nil {
			t.Fatalf("advance Transition(%d, %d) error = %v", path[index-1], path[index], err)
		}
		if path[index] == target {
			return owner
		}
	}
	t.Fatalf("cannot construct Owner in state %d", target)
	return nil
}

func runConcurrentTransitions(callers int, transition func(int) error) []error {
	start := make(chan struct{})
	results := make([]error, callers)
	var ready sync.WaitGroup
	var finished sync.WaitGroup
	ready.Add(callers)
	finished.Add(callers)
	for index := range callers {
		go func() {
			defer finished.Done()
			ready.Done()
			<-start
			results[index] = transition(index)
		}()
	}
	ready.Wait()
	close(start)
	finished.Wait()
	return results
}

func assertOneTransitionWinner(t *testing.T, results []error) {
	t.Helper()

	var successes atomic.Uint32
	for _, err := range results {
		switch {
		case err == nil:
			successes.Add(1)
		case errors.Is(err, executionowner.ErrInvalidTransition):
		default:
			t.Errorf("Transition() error = %v, want nil or ErrInvalidTransition", err)
		}
	}
	if got := successes.Load(); got != 1 {
		t.Fatalf("successful transitions = %d, want 1", got)
	}
}
