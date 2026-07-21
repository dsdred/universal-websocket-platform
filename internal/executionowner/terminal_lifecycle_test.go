package executionowner

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/lifetimelease"
)

func TestTerminalLifecycleFollowsApprovedOrderExactlyOnce(t *testing.T) {
	owner := committedOwner(t)
	events := &terminalEvents{}
	installCapturedRuntimeCallback(t, owner, func() error {
		events.add("drain")
		return nil
	})
	session := &terminalSession{
		start: func(context.Context) error {
			events.add("start")
			return nil
		},
		run: func(context.Context) error {
			events.add("run")
			return nil
		},
	}
	session.cleanup = func(context.Context) (CleanupCategory, CancellationCategory) {
		if state := owner.State(); state != StateTerminalizing {
			t.Errorf("Owner state during Cleanup = %d, want Terminalizing", state)
		}
		events.add("cleanup")
		return CleanupCategorySucceeded, CancellationCategoryConfirmed
	}
	completion := &terminalCompletion{complete: func() CompleteOutcome {
		events.add("completion")
		return CompleteOutcomeCompleted
	}}
	observer := &terminalObserver{observe: func(result TerminalResult) {
		events.add("observer")
		if result.PrimaryCause() != TerminationCauseNaturalCompletion ||
			result.StartCategory() != StartCategorySucceeded ||
			result.RunCategory() != RunCategoryReturned ||
			result.CleanupCategory() != CleanupCategorySucceeded ||
			result.CompletionCategory() != CompletionCategoryCompleted {
			t.Errorf("Terminal Result = %+v, want normal completed path", result)
		}
	}}
	lease := &terminalLease{release: func() lifetimelease.ReleaseOutcome {
		events.add("lease")
		state := owner.state
		state.mu.RLock()
		terminal := state.current == StateTerminal
		sealed := state.control.sealed
		state.mu.RUnlock()
		if !terminal || !sealed {
			t.Errorf("lease observed Terminal/sealed = %t/%t, want true/true", terminal, sealed)
		}
		return lifetimelease.ReleaseOutcomeReleased
	}}

	if err := owner.Execute(
		context.Background(), session, completion, observer, lease,
	); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := events.snapshot(); !reflect.DeepEqual(got, []string{
		"start", "run", "cleanup", "completion", "observer", "drain", "lease",
	}) {
		t.Fatalf("terminal events = %v", got)
	}
	if owner.State() != StateTerminal {
		t.Fatalf("Owner state = %d, want Terminal", owner.State())
	}
	assertTerminalCalls(t, session, completion, observer, lease, 1)

	if err := owner.Execute(
		context.Background(), session, completion, observer, lease,
	); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("repeated Execute() error = %v, want ErrInvalidTransition", err)
	}
	assertTerminalCalls(t, session, completion, observer, lease, 1)
}

func TestConcurrentExecuteHasOneCompleteTerminalLifecycle(t *testing.T) {
	owner := committedOwner(t)
	startEntered := make(chan struct{})
	releaseStart := make(chan struct{})
	var enteredOnce sync.Once
	session := &terminalSession{start: func(context.Context) error {
		enteredOnce.Do(func() { close(startEntered) })
		<-releaseStart
		return nil
	}}
	completion := &terminalCompletion{}
	observer := &terminalObserver{}
	lease := &terminalLease{}

	const callers = 32
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
			results <- owner.Execute(
				context.Background(), session, completion, observer, lease,
			)
		}()
	}
	defer func() {
		select {
		case <-releaseStart:
		default:
			close(releaseStart)
		}
		done.Wait()
	}()

	ready.Wait()
	close(start)
	waitClosed(t, startEntered, "winning Start")
	for index := 0; index < callers-1; index++ {
		if err := waitTerminalExecution(t, results, "losing Execute"); !errors.Is(err, ErrInvalidTransition) {
			t.Fatalf("losing Execute() error = %v, want ErrInvalidTransition", err)
		}
	}
	close(releaseStart)
	if err := waitTerminalExecution(t, results, "winning Execute"); err != nil {
		t.Fatalf("winning Execute() error = %v", err)
	}
	done.Wait()

	if owner.State() != StateTerminal {
		t.Fatalf("Owner state = %d, want Terminal", owner.State())
	}
	assertTerminalCalls(t, session, completion, observer, lease, 1)
}

func TestCompletionPanicIsContainedAndObserved(t *testing.T) {
	owner := committedOwner(t)
	completion := &terminalCompletion{complete: func() CompleteOutcome {
		panic("private completion payload")
	}}
	var observed TerminalResult
	observer := &terminalObserver{observe: func(result TerminalResult) {
		observed = result
	}}
	lease := &terminalLease{}

	if err := owner.Execute(
		context.Background(),
		&terminalSession{},
		completion,
		observer,
		lease,
	); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if observed.CompletionCategory() != CompletionCategoryPanicked {
		t.Fatalf("Completion category = %d, want Panicked", observed.CompletionCategory())
	}
	if owner.State() != StateTerminal || observer.calls.Load() != 1 || lease.calls.Load() != 1 {
		t.Fatalf("state/observer/lease = %d/%d/%d, want Terminal/1/1",
			owner.State(), observer.calls.Load(), lease.calls.Load())
	}
}

func TestStartAndRunPanicUseTheCommonTerminalLifecycle(t *testing.T) {
	tests := []struct {
		name        string
		session     func() *terminalSession
		wantStart   StartCategory
		wantRun     RunCategory
		wantPhase   RecoveredPanicPhase
		wantRunCall uint32
	}{
		{
			name: "Start",
			session: func() *terminalSession {
				return &terminalSession{start: func(context.Context) error {
					panic("private Start payload")
				}}
			},
			wantStart: StartCategoryPanicked,
			wantRun:   RunCategoryNotStarted,
			wantPhase: RecoveredPanicPhaseStart,
		},
		{
			name: "Run",
			session: func() *terminalSession {
				return &terminalSession{run: func(context.Context) error {
					panic("private Run payload")
				}}
			},
			wantStart:   StartCategorySucceeded,
			wantRun:     RunCategoryPanicked,
			wantPhase:   RecoveredPanicPhaseRun,
			wantRunCall: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner := committedOwner(t)
			session := test.session()
			var observed TerminalResult
			observer := &terminalObserver{observe: func(result TerminalResult) {
				observed = result
			}}
			lease := &terminalLease{}

			err := owner.Execute(
				context.Background(),
				session,
				&terminalCompletion{},
				observer,
				lease,
			)
			if !errors.Is(err, ErrSessionPanic) {
				t.Fatalf("Execute() error = %v, want ErrSessionPanic", err)
			}
			if observed.StartCategory() != test.wantStart ||
				observed.RunCategory() != test.wantRun ||
				observed.RecoveredPanicPhase() != test.wantPhase ||
				observed.PrimaryCause() != TerminationCauseRecoveredPanic {
				t.Fatalf("panic Terminal Result = %+v", observed)
			}
			if session.runCalls.Load() != test.wantRunCall {
				t.Fatalf("Run calls = %d, want %d", session.runCalls.Load(), test.wantRunCall)
			}
			if owner.State() != StateTerminal || session.cleanupCalls.Load() != 1 ||
				observer.calls.Load() != 1 || lease.calls.Load() != 1 {
				t.Fatalf("state/cleanup/observer/lease = %d/%d/%d/%d, want Terminal/1/1/1",
					owner.State(), session.cleanupCalls.Load(), observer.calls.Load(), lease.calls.Load())
			}
		})
	}
}

func TestCallbackInstallationFailureBuildsOneResultBeforeUnconfirmedDrain(t *testing.T) {
	owner := committedOwner(t)
	installErr := errors.New("install failure")
	if err := owner.installRuntimeCancellation(runtimeCancellationObservation{
		root: context.Background(),
		register: func(func()) (callbackRegistration, error) {
			return callbackRegistration{}, installErr
		},
	}); !errors.Is(err, installErr) {
		t.Fatalf("installRuntimeCancellation() error = %v, want install failure", err)
	}
	session := &terminalSession{}
	var observed TerminalResult
	observer := &terminalObserver{observe: func(result TerminalResult) {
		observed = result
	}}
	lease := &terminalLease{}

	if err := owner.Execute(
		context.Background(),
		session,
		&terminalCompletion{},
		observer,
		lease,
	); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if observed.PrimaryCause() != TerminationCauseExecutionFailure ||
		observed.StartCategory() != StartCategoryNotAttempted ||
		!observed.RuntimeObservationAnomalies().Has(RuntimeObservationAnomalyInstallFailure) {
		t.Fatalf("installation-failure Terminal Result = %+v", observed)
	}
	if owner.State() != StateTerminalizing || session.startCalls.Load() != 0 ||
		session.runCalls.Load() != 0 || observer.calls.Load() != 1 || lease.calls.Load() != 0 {
		t.Fatalf("state/start/run/observer/lease = %d/%d/%d/%d/%d, want Terminalizing/0/0/1/0",
			owner.State(), session.startCalls.Load(), session.runCalls.Load(),
			observer.calls.Load(), lease.calls.Load())
	}
}

func TestObserverPanicIsContainedBeforeDrainAndRelease(t *testing.T) {
	owner := committedOwner(t)
	var drainCalls atomic.Uint32
	installCapturedRuntimeCallback(t, owner, func() error {
		drainCalls.Add(1)
		return nil
	})
	observer := &terminalObserver{observe: func(TerminalResult) {
		panic("private observer payload")
	}}
	lease := &terminalLease{}

	if err := owner.Execute(
		context.Background(),
		&terminalSession{},
		&terminalCompletion{},
		observer,
		lease,
	); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if owner.State() != StateTerminal || drainCalls.Load() != 1 || lease.calls.Load() != 1 {
		t.Fatalf("state/drain/lease = %d/%d/%d, want Terminal/1/1",
			owner.State(), drainCalls.Load(), lease.calls.Load())
	}
	if observer.calls.Load() != 1 {
		t.Fatalf("Observer calls = %d, want 1", observer.calls.Load())
	}
}

func TestBlockedObserverKeepsOwnerTerminalizingAndLeaseActive(t *testing.T) {
	owner := committedOwner(t)
	observerEntered := make(chan struct{})
	releaseObserver := make(chan struct{})
	observer := &terminalObserver{observe: func(TerminalResult) {
		close(observerEntered)
		<-releaseObserver
	}}
	lease := &terminalLease{}
	done := make(chan error, 1)
	joined := false
	go func() {
		done <- owner.Execute(
			context.Background(),
			&terminalSession{},
			&terminalCompletion{},
			observer,
			lease,
		)
	}()
	defer func() {
		select {
		case <-releaseObserver:
		default:
			close(releaseObserver)
		}
		if !joined {
			<-done
		}
	}()

	waitClosed(t, observerEntered, "Observer entry")
	state := owner.state
	state.mu.RLock()
	current := state.current
	drainStarted := state.control.drainStarted
	state.mu.RUnlock()
	if current != StateTerminalizing || drainStarted || lease.calls.Load() != 0 {
		t.Fatalf("state/drain/lease = %d/%t/%d, want Terminalizing/false/0",
			current, drainStarted, lease.calls.Load())
	}

	close(releaseObserver)
	if err := waitTerminalExecution(t, done, "Observer release"); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	joined = true
	if owner.State() != StateTerminal || lease.calls.Load() != 1 {
		t.Fatalf("state/lease = %d/%d, want Terminal/1", owner.State(), lease.calls.Load())
	}
}

func TestUnconfirmedDrainPreventsSealTerminalAndLeaseRelease(t *testing.T) {
	owner := committedOwner(t)
	installCapturedRuntimeCallback(t, owner, func() error {
		return errors.New("unregister failure")
	})
	lease := &terminalLease{}

	if err := owner.Execute(
		context.Background(),
		&terminalSession{},
		&terminalCompletion{},
		&terminalObserver{},
		lease,
	); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	state := owner.state
	state.mu.RLock()
	current := state.current
	sealed := state.control.sealed
	admissionOpen := state.control.admissionOpen
	state.mu.RUnlock()
	if current != StateTerminalizing || sealed || admissionOpen || lease.calls.Load() != 0 {
		t.Fatalf("state/sealed/admission/lease = %d/%t/%t/%d, want Terminalizing/false/false/0",
			current, sealed, admissionOpen, lease.calls.Load())
	}
	if owner.RequestStop() {
		t.Fatal("RequestStop() after drain admission closure = true")
	}
}

func TestCancellationAnomalyReachesTerminalWithoutLeaseRelease(t *testing.T) {
	owner := committedOwner(t)
	session := &terminalSession{cleanup: func(context.Context) (CleanupCategory, CancellationCategory) {
		return CleanupCategorySucceeded, CancellationCategoryAnomaly
	}}
	lease := &terminalLease{}

	if err := owner.Execute(
		context.Background(),
		session,
		&terminalCompletion{},
		&terminalObserver{},
		lease,
	); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if owner.State() != StateTerminal || lease.calls.Load() != 0 {
		t.Fatalf("state/lease = %d/%d, want Terminal/0", owner.State(), lease.calls.Load())
	}
}

func TestLifetimeLeaseReleaseAdapterReturnsOutcome(t *testing.T) {
	tests := []struct {
		name  string
		lease lifetimelease.Lease
		want  lifetimelease.ReleaseOutcome
	}{
		{
			name: "released",
			lease: &terminalLease{release: func() lifetimelease.ReleaseOutcome {
				return lifetimelease.ReleaseOutcomeReleased
			}},
			want: lifetimelease.ReleaseOutcomeReleased,
		},
		{
			name: "accounting anomaly",
			lease: &terminalLease{release: func() lifetimelease.ReleaseOutcome {
				return lifetimelease.ReleaseOutcomeAccountingAnomaly
			}},
			want: lifetimelease.ReleaseOutcomeAccountingAnomaly,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := invokeLifetimeLeaseRelease(test.lease); got != test.want {
				t.Fatalf("invokeLifetimeLeaseRelease() = %d, want %d", got, test.want)
			}
		})
	}
}

func TestLifetimeLeaseReleaseAdapterNormalizesPanic(t *testing.T) {
	lease := &terminalLease{release: func() lifetimelease.ReleaseOutcome {
		panic("private lease payload")
	}}

	if got := invokeLifetimeLeaseRelease(lease); got != lifetimelease.ReleaseOutcomeAccountingAnomaly {
		t.Fatalf("invokeLifetimeLeaseRelease() = %d, want AccountingAnomaly", got)
	}
	if got := lease.calls.Load(); got != 1 {
		t.Fatalf("Lease calls = %d, want 1", got)
	}
}

func TestLifetimeLeaseReleasePanicDoesNotChangeTerminalLifecycle(t *testing.T) {
	owner := committedOwner(t)
	session := &terminalSession{}
	completion := &terminalCompletion{}
	observer := &terminalObserver{}
	lease := &terminalLease{release: func() lifetimelease.ReleaseOutcome {
		panic("private lease payload")
	}}

	if err := owner.Execute(
		context.Background(), session, completion, observer, lease,
	); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if owner.State() != StateTerminal {
		t.Fatalf("Owner state = %d, want Terminal", owner.State())
	}
	assertTerminalCalls(t, session, completion, observer, lease, 1)
}

func TestTerminalResultValidationFailureDoesNotPublishTerminal(t *testing.T) {
	owner := committedOwner(t)
	session := &terminalSession{cleanup: func(context.Context) (CleanupCategory, CancellationCategory) {
		return 0, CancellationCategoryConfirmed
	}}
	observer := &terminalObserver{}
	lease := &terminalLease{}

	err := owner.Execute(
		context.Background(),
		session,
		&terminalCompletion{},
		observer,
		lease,
	)
	if !errors.Is(err, ErrInvalidTerminalResult) {
		t.Fatalf("Execute() error = %v, want ErrInvalidTerminalResult", err)
	}
	if owner.State() != StateTerminalizing || observer.calls.Load() != 0 || lease.calls.Load() != 0 {
		t.Fatalf("state/observer/lease = %d/%d/%d, want Terminalizing/0/0",
			owner.State(), observer.calls.Load(), lease.calls.Load())
	}
}

type terminalSession struct {
	startCalls   atomic.Uint32
	runCalls     atomic.Uint32
	cleanupCalls atomic.Uint32
	start        func(context.Context) error
	run          func(context.Context) error
	cleanup      func(context.Context) (CleanupCategory, CancellationCategory)
}

func (session *terminalSession) Start(ctx context.Context) error {
	session.startCalls.Add(1)
	if session.start != nil {
		return session.start(ctx)
	}
	return nil
}

func (session *terminalSession) Run(ctx context.Context) error {
	session.runCalls.Add(1)
	if session.run != nil {
		return session.run(ctx)
	}
	return nil
}

func (session *terminalSession) Cleanup(
	ctx context.Context,
) (CleanupCategory, CancellationCategory) {
	session.cleanupCalls.Add(1)
	if session.cleanup != nil {
		return session.cleanup(ctx)
	}
	return CleanupCategorySucceeded, CancellationCategoryConfirmed
}

type terminalCompletion struct {
	calls    atomic.Uint32
	complete func() CompleteOutcome
}

func (completion *terminalCompletion) CompleteBoundRegistration() CompleteOutcome {
	completion.calls.Add(1)
	if completion.complete != nil {
		return completion.complete()
	}
	return CompleteOutcomeCompleted
}

type terminalObserver struct {
	calls   atomic.Uint32
	observe func(TerminalResult)
}

func (observer *terminalObserver) Observe(result TerminalResult) {
	observer.calls.Add(1)
	if observer.observe != nil {
		observer.observe(result)
	}
}

type terminalLease struct {
	calls   atomic.Uint32
	release func() lifetimelease.ReleaseOutcome
}

func (lease *terminalLease) Release() lifetimelease.ReleaseOutcome {
	lease.calls.Add(1)
	if lease.release != nil {
		return lease.release()
	}
	return lifetimelease.ReleaseOutcomeReleased
}

type terminalEvents struct {
	mu     sync.Mutex
	values []string
}

func (events *terminalEvents) add(value string) {
	events.mu.Lock()
	events.values = append(events.values, value)
	events.mu.Unlock()
}

func (events *terminalEvents) snapshot() []string {
	events.mu.Lock()
	defer events.mu.Unlock()
	return append([]string(nil), events.values...)
}

func assertTerminalCalls(
	t *testing.T,
	session *terminalSession,
	completion *terminalCompletion,
	observer *terminalObserver,
	lease *terminalLease,
	want uint32,
) {
	t.Helper()
	if got := session.startCalls.Load(); got != want {
		t.Errorf("Start calls = %d, want %d", got, want)
	}
	if got := session.runCalls.Load(); got != want {
		t.Errorf("Run calls = %d, want %d", got, want)
	}
	if got := session.cleanupCalls.Load(); got != want {
		t.Errorf("Cleanup calls = %d, want %d", got, want)
	}
	if got := completion.calls.Load(); got != want {
		t.Errorf("Completion calls = %d, want %d", got, want)
	}
	if got := observer.calls.Load(); got != want {
		t.Errorf("Observer calls = %d, want %d", got, want)
	}
	if got := lease.calls.Load(); got != want {
		t.Errorf("Lease calls = %d, want %d", got, want)
	}
}

func waitTerminalExecution(
	t *testing.T,
	result <-chan error,
	description string,
) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", description)
		return nil
	}
}
