package executionowner_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
)

func TestExecuteNormalStartRunTerminalizing(t *testing.T) {
	owner := committedOwner(t)
	session := &lifecycleSession{
		start: func(context.Context) error {
			assertOwnerState(t, owner, executionowner.StateStarting)
			return nil
		},
		run: func(context.Context) error {
			assertOwnerState(t, owner, executionowner.StateRunning)
			return nil
		},
	}

	if err := owner.Execute(context.Background(), session); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	assertOwnerState(t, owner, executionowner.StateTerminalizing)
	if got := session.startCalls.Load(); got != 1 {
		t.Fatalf("Start() calls = %d, want 1", got)
	}
	if got := session.runCalls.Load(); got != 1 {
		t.Fatalf("Run() calls = %d, want 1", got)
	}
}

func TestExecuteStartErrorTerminalizesWithoutRun(t *testing.T) {
	wantErr := errors.New("start failure")
	owner := committedOwner(t)
	session := &lifecycleSession{start: func(context.Context) error { return wantErr }}

	err := owner.Execute(context.Background(), session)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want start failure", err)
	}
	assertOwnerState(t, owner, executionowner.StateTerminalizing)
	if got := session.startCalls.Load(); got != 1 {
		t.Fatalf("Start() calls = %d, want 1", got)
	}
	if got := session.runCalls.Load(); got != 0 {
		t.Fatalf("Run() calls = %d, want 0", got)
	}
}

func TestExecuteRunErrorTerminalizes(t *testing.T) {
	wantErr := errors.New("run failure")
	owner := committedOwner(t)
	session := &lifecycleSession{run: func(context.Context) error { return wantErr }}

	err := owner.Execute(context.Background(), session)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want run failure", err)
	}
	assertOwnerState(t, owner, executionowner.StateTerminalizing)
	if got := session.startCalls.Load(); got != 1 {
		t.Fatalf("Start() calls = %d, want 1", got)
	}
	if got := session.runCalls.Load(); got != 1 {
		t.Fatalf("Run() calls = %d, want 1", got)
	}
}

func TestExecuteCanceledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	owner := committedOwner(t)
	session := &lifecycleSession{}

	err := owner.Execute(ctx, session)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Execute() error = %v, want context.Canceled", err)
	}
	assertOwnerState(t, owner, executionowner.StateTerminalizing)
	if got := session.startCalls.Load(); got != 0 {
		t.Fatalf("Start() calls = %d, want 0", got)
	}
	if got := session.runCalls.Load(); got != 0 {
		t.Fatalf("Run() calls = %d, want 0", got)
	}
}

func TestExecuteStopIntentBeforeStart(t *testing.T) {
	owner := committedOwner(t)
	if !owner.RequestStop() {
		t.Fatal("RequestStop() = false, want true")
	}
	session := &lifecycleSession{}

	if err := owner.Execute(context.Background(), session); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	assertOwnerState(t, owner, executionowner.StateTerminalizing)
	if got := session.startCalls.Load(); got != 0 {
		t.Fatalf("Start() calls = %d, want 0", got)
	}
	if got := session.runCalls.Load(); got != 0 {
		t.Fatalf("Run() calls = %d, want 0", got)
	}
}

func TestExecuteCannotRunTwice(t *testing.T) {
	owner := committedOwner(t)
	session := &lifecycleSession{}

	if err := owner.Execute(context.Background(), session); err != nil {
		t.Fatalf("first Execute() error = %v", err)
	}
	err := owner.Execute(context.Background(), session)
	if !errors.Is(err, executionowner.ErrInvalidTransition) {
		t.Fatalf("second Execute() error = %v, want ErrInvalidTransition", err)
	}
	if got := session.startCalls.Load(); got != 1 {
		t.Fatalf("Start() calls = %d, want 1", got)
	}
	if got := session.runCalls.Load(); got != 1 {
		t.Fatalf("Run() calls = %d, want 1", got)
	}
}

func TestExecuteBeforeCommitDoesNotStartSession(t *testing.T) {
	owner := executionowner.New()
	session := &lifecycleSession{}

	err := owner.Execute(context.Background(), session)
	if !errors.Is(err, executionowner.ErrInvalidTransition) {
		t.Fatalf("Execute() error = %v, want ErrInvalidTransition", err)
	}
	assertOwnerState(t, owner, executionowner.StatePreCommit)
	if got := session.startCalls.Load(); got != 0 {
		t.Fatalf("Start() calls = %d, want 0", got)
	}
	if got := session.runCalls.Load(); got != 0 {
		t.Fatalf("Run() calls = %d, want 0", got)
	}
}

func TestExternalTransitionCannotClaimPostCommitLifecycle(t *testing.T) {
	owner := committedOwner(t)

	err := owner.Transition(executionowner.StateCommitted, executionowner.StateStarting)
	if !errors.Is(err, executionowner.ErrInvalidTransition) {
		t.Fatalf("Transition(Committed, Starting) error = %v, want ErrInvalidTransition", err)
	}
	assertOwnerState(t, owner, executionowner.StateCommitted)
}

func TestPublicTransitionRejectsPostCommitLifecycleMatrix(t *testing.T) {
	t.Run("Committed", func(t *testing.T) {
		owner := committedOwner(t)
		assertRejectedTransitions(t, owner, executionowner.StateCommitted, []executionowner.State{
			executionowner.StateStarting,
			executionowner.StateTerminalizing,
			executionowner.StateCommitted,
			executionowner.StatePreCommit,
			255,
		})
	})

	t.Run("Starting", func(t *testing.T) {
		owner := committedOwner(t)
		startEntered := make(chan struct{})
		releaseStart := make(chan struct{})
		release := closeOnce(releaseStart)
		session := &lifecycleSession{start: func(context.Context) error {
			close(startEntered)
			<-releaseStart
			return nil
		}}
		result, wait := executeAsync(owner, session)
		defer func() {
			release()
			wait()
		}()
		waitForSignal(t, startEntered, result, "Start entry")

		assertRejectedTransitions(t, owner, executionowner.StateStarting, []executionowner.State{
			executionowner.StateRunning,
			executionowner.StateTerminalizing,
			executionowner.StateStarting,
			executionowner.StateCommitted,
			executionowner.StatePreCommit,
			255,
		})
	})

	t.Run("Running", func(t *testing.T) {
		owner := committedOwner(t)
		runEntered := make(chan struct{})
		releaseRun := make(chan struct{})
		release := closeOnce(releaseRun)
		session := &lifecycleSession{run: func(context.Context) error {
			close(runEntered)
			<-releaseRun
			return nil
		}}
		result, wait := executeAsync(owner, session)
		defer func() {
			release()
			wait()
		}()
		waitForSignal(t, runEntered, result, "Run entry")

		assertRejectedTransitions(t, owner, executionowner.StateRunning, []executionowner.State{
			executionowner.StateTerminalizing,
			executionowner.StateRunning,
			executionowner.StateStarting,
			executionowner.StateCommitted,
			255,
		})
	})

	t.Run("Terminalizing", func(t *testing.T) {
		owner := committedOwner(t)
		if err := owner.Execute(context.Background(), &lifecycleSession{}); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		assertRejectedTransitions(t, owner, executionowner.StateTerminalizing, []executionowner.State{
			executionowner.StateTerminal,
			executionowner.StateTerminalizing,
			executionowner.StateRunning,
			executionowner.StateCommitted,
			255,
		})
	})

	t.Run("UnknownSource", func(t *testing.T) {
		owner := committedOwner(t)
		err := owner.Transition(255, executionowner.StateCommitted)
		if !errors.Is(err, executionowner.ErrInvalidTransition) {
			t.Fatalf("Transition(unknown, Committed) error = %v, want ErrInvalidTransition", err)
		}
		assertOwnerState(t, owner, executionowner.StateCommitted)
	})
}

func TestExternalTransitionDuringStartCannotBypassStartFailure(t *testing.T) {
	wantErr := errors.New("start failure")
	owner := committedOwner(t)
	startEntered := make(chan struct{})
	releaseStart := make(chan struct{})
	session := &lifecycleSession{start: func(context.Context) error {
		close(startEntered)
		<-releaseStart
		return wantErr
	}}

	release := closeOnce(releaseStart)
	result, wait := executeAsync(owner, session)
	defer func() {
		release()
		wait()
	}()
	waitForSignal(t, startEntered, result, "Start entry")

	err := owner.Transition(executionowner.StateStarting, executionowner.StateRunning)
	if !errors.Is(err, executionowner.ErrInvalidTransition) {
		t.Fatalf("Transition(Starting, Running) error = %v, want ErrInvalidTransition", err)
	}
	assertOwnerState(t, owner, executionowner.StateStarting)
	release()
	if err := <-result; !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want start failure", err)
	}
	assertOwnerState(t, owner, executionowner.StateTerminalizing)
	if got := session.runCalls.Load(); got != 0 {
		t.Fatalf("Run() calls = %d, want 0", got)
	}
}

func TestConcurrentExternalTransitionsCannotMutateOwnedLifecycle(t *testing.T) {
	owner := committedOwner(t)
	startEntered := make(chan struct{})
	releaseStart := make(chan struct{})
	session := &lifecycleSession{start: func(context.Context) error {
		close(startEntered)
		<-releaseStart
		return nil
	}}

	release := closeOnce(releaseStart)
	result, wait := executeAsync(owner, session)
	defer func() {
		release()
		wait()
	}()
	waitForSignal(t, startEntered, result, "Start entry")

	const callers = 64
	errorsByCaller := make(chan error, callers)
	var ready sync.WaitGroup
	var done sync.WaitGroup
	ready.Add(callers)
	done.Add(callers)
	start := make(chan struct{})
	for index := range callers {
		go func() {
			defer done.Done()
			ready.Done()
			<-start
			to := executionowner.StateRunning
			if index%2 != 0 {
				to = executionowner.StateTerminalizing
			}
			errorsByCaller <- owner.Transition(executionowner.StateStarting, to)
		}()
	}
	ready.Wait()
	close(start)
	done.Wait()
	close(errorsByCaller)
	for err := range errorsByCaller {
		if !errors.Is(err, executionowner.ErrInvalidTransition) {
			t.Errorf("external Transition() error = %v, want ErrInvalidTransition", err)
		}
	}
	assertOwnerState(t, owner, executionowner.StateStarting)

	release()
	if err := <-result; err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	assertOwnerState(t, owner, executionowner.StateTerminalizing)
	if got := session.startCalls.Load(); got != 1 {
		t.Fatalf("Start() calls = %d, want 1", got)
	}
	if got := session.runCalls.Load(); got != 1 {
		t.Fatalf("Run() calls = %d, want 1", got)
	}
}

func TestStopDuringStartPublishesIntentWithoutChangingLifecycle(t *testing.T) {
	owner := committedOwner(t)
	startEntered := make(chan struct{})
	releaseStart := make(chan struct{})
	session := &lifecycleSession{start: func(context.Context) error {
		close(startEntered)
		<-releaseStart
		return nil
	}}

	release := closeOnce(releaseStart)
	result, wait := executeAsync(owner, session)
	defer func() {
		release()
		wait()
	}()
	waitForSignal(t, startEntered, result, "Start entry")
	if !owner.RequestStop() {
		t.Fatal("RequestStop() = false, want true")
	}
	assertOwnerState(t, owner, executionowner.StateStarting)

	release()
	if err := <-result; err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	assertOwnerState(t, owner, executionowner.StateTerminalizing)
	if got := session.runCalls.Load(); got != 0 {
		t.Fatalf("Run() calls = %d, want 0", got)
	}
}

func TestConcurrentExecuteAttemptsHaveOneExecutionPath(t *testing.T) {
	owner := committedOwner(t)
	startEntered := make(chan struct{})
	releaseStart := make(chan struct{})
	session := &lifecycleSession{start: func(context.Context) error {
		close(startEntered)
		<-releaseStart
		return nil
	}}

	release := closeOnce(releaseStart)
	firstResult, wait := executeAsync(owner, session)
	defer func() {
		release()
		wait()
	}()
	waitForSignal(t, startEntered, firstResult, "Start entry")

	const contenders = 64
	results := make(chan error, contenders)
	var ready sync.WaitGroup
	var done sync.WaitGroup
	ready.Add(contenders)
	done.Add(contenders)
	start := make(chan struct{})
	for range contenders {
		go func() {
			defer done.Done()
			ready.Done()
			<-start
			results <- owner.Execute(context.Background(), session)
		}()
	}
	ready.Wait()
	close(start)
	done.Wait()
	close(results)
	for err := range results {
		if !errors.Is(err, executionowner.ErrInvalidTransition) {
			t.Errorf("competing Execute() error = %v, want ErrInvalidTransition", err)
		}
	}

	release()
	if err := <-firstResult; err != nil {
		t.Fatalf("winning Execute() error = %v", err)
	}
	assertOwnerState(t, owner, executionowner.StateTerminalizing)
	if got := session.startCalls.Load(); got != 1 {
		t.Fatalf("Start() calls = %d, want 1", got)
	}
	if got := session.runCalls.Load(); got != 1 {
		t.Fatalf("Run() calls = %d, want 1", got)
	}
}

func TestSimultaneousExecuteClaimFromCommittedHasOneWinner(t *testing.T) {
	owner := committedOwner(t)
	const callers = 64

	startEntered := make(chan struct{})
	releaseStart := make(chan struct{})
	release := closeOnce(releaseStart)
	var startEnteredOnce sync.Once
	session := &lifecycleSession{start: func(context.Context) error {
		startEnteredOnce.Do(func() { close(startEntered) })
		<-releaseStart
		return nil
	}}

	start := make(chan struct{})
	results := make(chan error, callers)
	var ready sync.WaitGroup
	var done sync.WaitGroup
	ready.Add(callers)
	done.Add(callers)
	for range callers {
		go func() {
			defer done.Done()
			ready.Done()
			<-start
			results <- owner.Execute(context.Background(), session)
		}()
	}
	defer func() {
		release()
		done.Wait()
	}()

	ready.Wait()
	assertOwnerState(t, owner, executionowner.StateCommitted)
	close(start)
	select {
	case <-startEntered:
	case <-time.After(5 * time.Second):
		t.Fatal("no Execute() reached Start")
	}

	for index := 0; index < callers-1; index++ {
		select {
		case err := <-results:
			if !errors.Is(err, executionowner.ErrInvalidTransition) {
				t.Errorf("losing Execute() error = %v, want ErrInvalidTransition", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("only %d losing Execute() calls returned before winner release", index)
		}
	}
	assertOwnerState(t, owner, executionowner.StateStarting)

	release()
	select {
	case err := <-results:
		if err != nil {
			t.Fatalf("winning Execute() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("winning Execute() did not return")
	}
	done.Wait()
	if got := session.startCalls.Load(); got != 1 {
		t.Fatalf("Start() calls = %d, want 1", got)
	}
	if got := session.runCalls.Load(); got != 1 {
		t.Fatalf("Run() calls = %d, want 1", got)
	}
	assertOwnerState(t, owner, executionowner.StateTerminalizing)
}

func TestOwnerCopiesCannotCreateSecondExecutionPath(t *testing.T) {
	original := committedOwner(t)
	copy := *original
	startEntered := make(chan struct{})
	releaseStart := make(chan struct{})
	session := &lifecycleSession{start: func(context.Context) error {
		close(startEntered)
		<-releaseStart
		return nil
	}}

	release := closeOnce(releaseStart)
	firstResult, wait := executeAsync(original, session)
	defer func() {
		release()
		wait()
	}()
	waitForSignal(t, startEntered, firstResult, "Start entry")

	if err := copy.Execute(context.Background(), session); !errors.Is(err, executionowner.ErrInvalidTransition) {
		t.Fatalf("Execute() through copy error = %v, want ErrInvalidTransition", err)
	}
	release()
	if err := <-firstResult; err != nil {
		t.Fatalf("winning Execute() error = %v", err)
	}
	if got := session.startCalls.Load(); got != 1 {
		t.Fatalf("Start() calls = %d, want 1", got)
	}
	if got := session.runCalls.Load(); got != 1 {
		t.Fatalf("Run() calls = %d, want 1", got)
	}
}

func TestOwnerCopyCannotMutateOwnedLifecycle(t *testing.T) {
	original := committedOwner(t)
	copy := *original
	startEntered := make(chan struct{})
	releaseStart := make(chan struct{})
	session := &lifecycleSession{start: func(context.Context) error {
		close(startEntered)
		<-releaseStart
		return nil
	}}

	release := closeOnce(releaseStart)
	result, wait := executeAsync(original, session)
	defer func() {
		release()
		wait()
	}()
	waitForSignal(t, startEntered, result, "Start entry")
	if err := copy.Transition(executionowner.StateStarting, executionowner.StateRunning); !errors.Is(err, executionowner.ErrInvalidTransition) {
		t.Fatalf("copy Transition() error = %v, want ErrInvalidTransition", err)
	}
	assertOwnerState(t, &copy, executionowner.StateStarting)

	release()
	if err := <-result; err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	assertOwnerState(t, &copy, executionowner.StateTerminalizing)
	if got := session.startCalls.Load(); got != 1 {
		t.Fatalf("Start() calls = %d, want 1", got)
	}
	if got := session.runCalls.Load(); got != 1 {
		t.Fatalf("Run() calls = %d, want 1", got)
	}
}

type lifecycleSession struct {
	startCalls atomic.Uint32
	runCalls   atomic.Uint32
	start      func(context.Context) error
	run        func(context.Context) error
}

func (session *lifecycleSession) Start(ctx context.Context) error {
	session.startCalls.Add(1)
	if session.start != nil {
		return session.start(ctx)
	}
	return nil
}

func (session *lifecycleSession) Run(ctx context.Context) error {
	session.runCalls.Add(1)
	if session.run != nil {
		return session.run(ctx)
	}
	return nil
}

func committedOwner(t *testing.T) *executionowner.Owner {
	t.Helper()
	owner := executionowner.New()
	if err := owner.Transition(executionowner.StatePreCommit, executionowner.StateCommitted); err != nil {
		t.Fatalf("commit Transition() error = %v", err)
	}
	return owner
}

func assertOwnerState(t *testing.T, owner *executionowner.Owner, want executionowner.State) {
	t.Helper()
	if got := owner.State(); got != want {
		t.Fatalf("Owner State() = %d, want %d", got, want)
	}
}

func assertRejectedTransitions(
	t *testing.T,
	owner *executionowner.Owner,
	from executionowner.State,
	targets []executionowner.State,
) {
	t.Helper()
	for _, to := range targets {
		err := owner.Transition(from, to)
		if !errors.Is(err, executionowner.ErrInvalidTransition) {
			t.Errorf("Transition(%d, %d) error = %v, want ErrInvalidTransition", from, to, err)
		}
		assertOwnerState(t, owner, from)
	}
}

func executeAsync(
	owner *executionowner.Owner,
	session executionowner.SessionLifecycle,
) (<-chan error, func()) {
	result := make(chan error, 1)
	var done sync.WaitGroup
	done.Add(1)
	go func() {
		defer done.Done()
		result <- owner.Execute(context.Background(), session)
	}()
	return result, done.Wait
}

func closeOnce(channel chan struct{}) func() {
	var once sync.Once
	return func() {
		once.Do(func() { close(channel) })
	}
}

func waitForSignal(t *testing.T, signal <-chan struct{}, result <-chan error, description string) {
	t.Helper()
	select {
	case <-signal:
	case err := <-result:
		t.Fatalf("Execute() returned before %s: %v", description, err)
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", description)
	}
}
