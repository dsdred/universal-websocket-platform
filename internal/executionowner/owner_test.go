package executionowner

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestCreateOwnerSuccess(t *testing.T) {
	owner := New()
	if owner == nil {
		t.Fatal("New() returned a nil Owner")
	}
}

func TestInitialState(t *testing.T) {
	owner := New()
	if got := owner.State(); got != StatePreCommit {
		t.Fatalf("State() = %d, want StatePreCommit", got)
	}
	if owner.StopRequested() {
		t.Fatal("StopRequested() = true for a new Owner")
	}
}

func TestZeroValueNotUsable(t *testing.T) {
	var owner Owner

	if got := owner.State(); got != 0 {
		t.Fatalf("zero-value State() = %d, want invalid zero state", got)
	}
	if err := owner.transitionLifecycle(StatePreCommit, StateCommitted); !errors.Is(err, ErrUninitializedOwner) {
		t.Fatalf("zero-value Transition() error = %v, want ErrUninitializedOwner", err)
	}
	if owner.RequestStop() {
		t.Fatal("zero-value RequestStop() = true, want false")
	}
	if owner.StopRequested() {
		t.Fatal("zero-value StopRequested() = true, want false")
	}
}

func TestNilOwnerNotUsable(t *testing.T) {
	var owner *Owner

	if got := owner.State(); got != 0 {
		t.Fatalf("nil State() = %d, want invalid zero state", got)
	}
	if err := owner.transitionLifecycle(StatePreCommit, StateCommitted); !errors.Is(err, ErrUninitializedOwner) {
		t.Fatalf("nil Transition() error = %v, want ErrUninitializedOwner", err)
	}
	if owner.RequestStop() {
		t.Fatal("nil RequestStop() = true, want false")
	}
	if owner.StopRequested() {
		t.Fatal("nil StopRequested() = true, want false")
	}
}

func TestRequestStopIsOneShotAndDoesNotChangeLifecycle(t *testing.T) {
	owner := New()

	if !owner.RequestStop() {
		t.Fatal("first RequestStop() = false, want true")
	}
	if !owner.StopRequested() {
		t.Fatal("StopRequested() = false after successful RequestStop()")
	}
	if owner.RequestStop() {
		t.Fatal("repeated RequestStop() = true, want false")
	}
	if got := owner.State(); got != StatePreCommit {
		t.Fatalf("State() = %d after RequestStop(), want StatePreCommit", got)
	}
}

func TestStopRequestRemainsClosedAcrossLifecycleTransitions(t *testing.T) {
	owner := New()
	if !owner.RequestStop() {
		t.Fatal("first RequestStop() = false, want true")
	}

	path := []State{
		StatePreCommit,
		StateCommitted,
		StateStarting,
		StateRunning,
		StateTerminalizing,
		StateTerminal,
	}
	for index := 1; index < len(path); index++ {
		if err := owner.transitionLifecycle(path[index-1], path[index]); err != nil {
			t.Fatalf("Transition(%d, %d) error = %v", path[index-1], path[index], err)
		}
		if !owner.StopRequested() {
			t.Fatalf("StopRequested() = false after transition to %d", path[index])
		}
	}
	if owner.RequestStop() {
		t.Fatal("RequestStop() after Terminal = true, want false")
	}
}

func TestTwoConcurrentStopRequestsHaveOneWinner(t *testing.T) {
	owner := New()
	results := runConcurrentStopRequests(2, func(int) bool {
		return owner.RequestStop()
	})

	assertOneStopWinner(t, results)
	if !owner.StopRequested() {
		t.Fatal("StopRequested() = false after concurrent RequestStop() calls")
	}
}

func TestMassConcurrentStopRequestsHaveOneWinner(t *testing.T) {
	owner := New()
	results := runConcurrentStopRequests(128, func(int) bool {
		return owner.RequestStop()
	})

	assertOneStopWinner(t, results)
}

func TestConcurrentStopRequestedAndRequestStop(t *testing.T) {
	owner := New()
	const readers = 128

	start := make(chan struct{})
	observed := make(chan bool, readers)
	var waitGroup sync.WaitGroup
	for range readers {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			observed <- owner.StopRequested()
		}()
	}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		<-start
		if !owner.RequestStop() {
			t.Error("concurrent RequestStop() = false, want true")
		}
	}()

	close(start)
	waitGroup.Wait()
	close(observed)
	for range observed {
		// A reader may linearize before or after RequestStop. Both observations
		// are valid; completing this loop proves every reader returned.
	}
	if !owner.StopRequested() {
		t.Fatal("StopRequested() = false after concurrent operations")
	}
}

func TestOwnerCopiesShareOneControlCell(t *testing.T) {
	original := New()
	copy1 := *original
	copy2 := copy1

	if !copy1.RequestStop() {
		t.Fatal("RequestStop() through first copy = false, want true")
	}
	if !original.StopRequested() || !copy2.StopRequested() {
		t.Fatal("Owner copies do not observe one requested Stop state")
	}
	if original.RequestStop() || copy2.RequestStop() {
		t.Fatal("Owner copy accepted a second Stop request")
	}
}

func TestConcurrentStopRequestsThroughCopiesHaveOneWinner(t *testing.T) {
	original := New()
	copy1 := *original
	copy2 := copy1
	owners := []*Owner{original, &copy1, &copy2}

	results := runConcurrentStopRequests(96, func(index int) bool {
		return owners[index%len(owners)].RequestStop()
	})
	assertOneStopWinner(t, results)
	for index, owner := range owners {
		if !owner.StopRequested() {
			t.Errorf("owners[%d].StopRequested() = false", index)
		}
	}
}

func TestRequestStopAfterTerminalizingIsRejected(t *testing.T) {
	tests := []struct {
		name  string
		state State
	}{
		{name: "Terminalizing", state: StateTerminalizing},
		{name: "Terminal", state: StateTerminal},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner := ownerInState(t, test.state)
			if owner.RequestStop() {
				t.Fatalf("RequestStop() in state %d = true, want false", test.state)
			}
			if owner.StopRequested() {
				t.Fatalf("StopRequested() in state %d = true, want false", test.state)
			}
		})
	}
}

func TestRequestStopAndTerminalizingTransitionLinearizeConsistently(t *testing.T) {
	owner := ownerInState(t, StateRunning)
	start := make(chan struct{})
	requestResult := make(chan bool, 1)
	transitionResult := make(chan error, 1)
	var ready sync.WaitGroup
	var finished sync.WaitGroup
	ready.Add(2)
	finished.Add(2)

	go func() {
		defer finished.Done()
		ready.Done()
		<-start
		requestResult <- owner.RequestStop()
	}()
	go func() {
		defer finished.Done()
		ready.Done()
		<-start
		transitionResult <- owner.transitionLifecycle(StateRunning, StateTerminalizing)
	}()

	ready.Wait()
	close(start)
	requested := <-requestResult
	transitionErr := <-transitionResult
	finished.Wait()
	if transitionErr != nil {
		t.Fatalf("Transition() error = %v", transitionErr)
	}
	if owner.StopRequested() != requested {
		t.Fatalf("StopRequested() = %t, want RequestStop() result %t", owner.StopRequested(), requested)
	}
	if got := owner.State(); got != StateTerminalizing {
		t.Fatalf("State() = %d, want StateTerminalizing", got)
	}
}

func TestAllowedTransitions(t *testing.T) {
	tests := []struct {
		name string
		path []State
	}{
		{
			name: "normal lifecycle",
			path: []State{
				StatePreCommit,
				StateCommitted,
				StateStarting,
				StateRunning,
				StateTerminalizing,
				StateTerminal,
			},
		},
		{
			name: "terminalize before Start",
			path: []State{
				StatePreCommit,
				StateCommitted,
				StateTerminalizing,
				StateTerminal,
			},
		},
		{
			name: "terminalize after Start linearization",
			path: []State{
				StatePreCommit,
				StateCommitted,
				StateStarting,
				StateTerminalizing,
				StateTerminal,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner := New()
			for index := 1; index < len(test.path); index++ {
				from := test.path[index-1]
				to := test.path[index]
				if err := owner.transitionLifecycle(from, to); err != nil {
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
	const unknown State = 255
	tests := []struct {
		name      string
		current   State
		from      State
		to        State
		wantParts []string
	}{
		{name: "skip state", current: StatePreCommit, from: StatePreCommit, to: StateStarting},
		{name: "move backward", current: StateRunning, from: StateRunning, to: StateStarting},
		{name: "same state", current: StateCommitted, from: StateCommitted, to: StateCommitted},
		{name: "from Terminal", current: StateTerminal, from: StateTerminal, to: StateTerminalizing},
		{name: "unknown target", current: StateCommitted, from: StateCommitted, to: unknown, wantParts: []string{"Committed", "State(255)"}},
		{name: "unknown expected source", current: StatePreCommit, from: unknown, to: StateCommitted, wantParts: []string{"PreCommit", "State(255)", "Committed"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner := ownerInState(t, test.current)
			err := owner.transitionLifecycle(test.from, test.to)
			if !errors.Is(err, ErrInvalidTransition) {
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
	owner := New()
	const callers = 64

	results := runConcurrentTransitions(callers, func(int) error {
		return owner.transitionLifecycle(StatePreCommit, StateCommitted)
	})
	assertOneTransitionWinner(t, results)
	if got := owner.State(); got != StateCommitted {
		t.Fatalf("State() = %d, want StateCommitted", got)
	}
}

func TestConcurrentDifferentTransitionsHaveOneValidWinner(t *testing.T) {
	owner := ownerInState(t, StateCommitted)
	results := runConcurrentTransitions(2, func(index int) error {
		if index == 0 {
			return owner.transitionLifecycle(StateCommitted, StateStarting)
		}
		return owner.transitionLifecycle(StateCommitted, StateTerminalizing)
	})
	assertOneTransitionWinner(t, results)

	got := owner.State()
	if got != StateStarting && got != StateTerminalizing {
		t.Fatalf("State() = %d, want StateStarting or StateTerminalizing", got)
	}
}

func TestConcurrentStateAndTransitionObserveOnlyValidStates(t *testing.T) {
	owner := New()
	const readers = 128

	start := make(chan struct{})
	observed := make(chan State, readers)
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
		if err := owner.transitionLifecycle(StatePreCommit, StateCommitted); err != nil {
			t.Errorf("Transition() error = %v", err)
		}
	}()

	close(start)
	waitGroup.Wait()
	close(observed)
	for state := range observed {
		if state != StatePreCommit && state != StateCommitted {
			t.Fatalf("State() observed invalid state %d", state)
		}
	}
}

func TestLifecycleTransitionsFromDifferentGoroutines(t *testing.T) {
	owner := New()
	path := []State{
		StatePreCommit,
		StateCommitted,
		StateStarting,
		StateRunning,
		StateTerminalizing,
		StateTerminal,
	}

	for index := 1; index < len(path); index++ {
		result := make(chan error, 1)
		var finished sync.WaitGroup
		finished.Add(1)
		from := path[index-1]
		to := path[index]
		go func() {
			defer finished.Done()
			result <- owner.transitionLifecycle(from, to)
		}()
		err := <-result
		finished.Wait()
		if err != nil {
			t.Fatalf("Transition(%d, %d) error = %v", from, to, err)
		}
	}

	if got := owner.State(); got != StateTerminal {
		t.Fatalf("State() = %d, want StateTerminal", got)
	}
}

func TestOwnerCopySharesLifecycle(t *testing.T) {
	original := New()
	copy := *original

	if err := original.transitionLifecycle(StatePreCommit, StateCommitted); err != nil {
		t.Fatalf("Transition() through original error = %v", err)
	}
	if got := copy.State(); got != StateCommitted {
		t.Fatalf("copy State() = %d, want StateCommitted", got)
	}
	if err := copy.transitionLifecycle(StateCommitted, StateStarting); err != nil {
		t.Fatalf("Transition() through copy error = %v", err)
	}
	if got := original.State(); got != StateStarting {
		t.Fatalf("original State() = %d, want StateStarting", got)
	}

	secondCopy := copy
	if err := secondCopy.transitionLifecycle(StateStarting, StateRunning); err != nil {
		t.Fatalf("Transition() through second copy error = %v", err)
	}
	if got := original.State(); got != StateRunning {
		t.Fatalf("original State() after second-copy transition = %d, want StateRunning", got)
	}
}

func TestConcurrentTransitionsThroughCopiesShareLifecycle(t *testing.T) {
	original := New()
	copy := *original

	results := runConcurrentTransitions(2, func(index int) error {
		if index == 0 {
			return original.transitionLifecycle(StatePreCommit, StateCommitted)
		}
		return copy.transitionLifecycle(StatePreCommit, StateCommitted)
	})
	assertOneTransitionWinner(t, results)
	if original.State() != StateCommitted || copy.State() != StateCommitted {
		t.Fatalf("copies do not observe one StateCommitted lifecycle")
	}
}

func TestCopyAfterMultipleTransitionsSharesLifecycle(t *testing.T) {
	original := ownerInState(t, StateRunning)
	copy := *original

	if err := copy.transitionLifecycle(StateRunning, StateTerminalizing); err != nil {
		t.Fatalf("Transition() through late copy error = %v", err)
	}
	if got := original.State(); got != StateTerminalizing {
		t.Fatalf("original State() = %d, want StateTerminalizing", got)
	}
}

func TestTerminalIsIrreversibleUnderConcurrency(t *testing.T) {
	owner := ownerInState(t, StateTerminal)
	const callers = 64

	results := runConcurrentTransitions(callers, func(index int) error {
		states := []State{
			StatePreCommit,
			StateCommitted,
			StateStarting,
			StateRunning,
			StateTerminalizing,
			StateTerminal,
		}
		return owner.transitionLifecycle(StateTerminal, states[index%len(states)])
	})
	for _, err := range results {
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("post-Terminal Transition() error = %v, want ErrInvalidTransition", err)
		}
	}
	if got := owner.State(); got != StateTerminal {
		t.Fatalf("State() = %d, want immutable StateTerminal", got)
	}
}

func ownerInState(t *testing.T, target State) *Owner {
	t.Helper()

	owner := New()
	path := []State{
		StatePreCommit,
		StateCommitted,
		StateStarting,
		StateRunning,
		StateTerminalizing,
		StateTerminal,
	}
	if target == StatePreCommit {
		return owner
	}
	for index := 1; index < len(path); index++ {
		if err := owner.transitionLifecycle(path[index-1], path[index]); err != nil {
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

func runConcurrentStopRequests(callers int, request func(int) bool) []bool {
	start := make(chan struct{})
	results := make([]bool, callers)
	var ready sync.WaitGroup
	var finished sync.WaitGroup
	ready.Add(callers)
	finished.Add(callers)
	for index := range callers {
		go func() {
			defer finished.Done()
			ready.Done()
			<-start
			results[index] = request(index)
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
		case errors.Is(err, ErrInvalidTransition):
		default:
			t.Errorf("Transition() error = %v, want nil or ErrInvalidTransition", err)
		}
	}
	if got := successes.Load(); got != 1 {
		t.Fatalf("successful transitions = %d, want 1", got)
	}
}

func assertOneStopWinner(t *testing.T, results []bool) {
	t.Helper()

	winners := 0
	for _, result := range results {
		if result {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("successful Stop requests = %d, want 1", winners)
	}
}
