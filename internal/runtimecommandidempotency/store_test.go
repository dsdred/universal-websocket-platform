package runtimecommandidempotency

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

var errDenied = errors.New("denied")

type executionResult struct {
	admission Admission
	err       error
}

func TestExecuteClaimsBeforeDelegationAndReplaysTerminalOutcome(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 41)
	var authorized atomic.Int32
	var delegated atomic.Int32
	wantOutcome := terminalOutcome(t, OutcomeSucceeded, "attempt-a")
	authorize := func(context.Context, Scope, Intent) error {
		authorized.Add(1)
		return nil
	}

	claimed, err := boundary.Execute(context.Background(), scope, "command-a", intent, authorize,
		func() (TerminalOutcome, error) {
			delegated.Add(1)
			return wantOutcome, nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if claimed.Kind() != AdmissionClaimed || claimed.Record().State() != CommandStateTerminal ||
		claimed.Record().Revision() != 2 {
		t.Fatalf("unexpected claim result: %#v", claimed)
	}

	replay, err := boundary.Execute(context.Background(), scope, "command-a", intent, authorize,
		func() (TerminalOutcome, error) {
			delegated.Add(1)
			return TerminalOutcome{}, errors.New("must not run")
		})
	if err != nil {
		t.Fatal(err)
	}
	gotOutcome, ok := replay.Record().Outcome()
	if replay.Kind() != AdmissionReplay || !ok || gotOutcome != wantOutcome {
		t.Fatalf("unexpected replay: %#v outcome=%#v", replay, gotOutcome)
	}
	if delegated.Load() != 1 || authorized.Load() != 2 {
		t.Fatalf("delegations=%d authorizations=%d", delegated.Load(), authorized.Load())
	}
}

func TestConcurrentSameKeyIssuesOneSynchronousDelegation(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 42)
	const submissions = 32
	start := make(chan struct{})
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	results := make(chan executionResult, submissions)
	var delegated atomic.Int32

	for range submissions {
		go func() {
			<-start
			admission, err := boundary.Execute(context.Background(), scope, "same-key", intent, allow,
				func() (TerminalOutcome, error) {
					delegated.Add(1)
					entered <- struct{}{}
					<-release
					return terminalOutcome(t, OutcomeSucceeded, "attempt"), nil
				})
			results <- executionResult{admission: admission, err: err}
		}()
	}
	close(start)
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("claiming callback did not start")
	}

	var inProgress int
	for range submissions - 1 {
		result := <-results
		if result.err != nil || result.admission.Kind() != AdmissionInProgress {
			t.Fatalf("expected in-progress observation, got %#v err=%v", result.admission, result.err)
		}
		inProgress++
	}
	close(release)
	claimed := <-results
	if claimed.err != nil || claimed.admission.Kind() != AdmissionClaimed ||
		claimed.admission.Record().State() != CommandStateTerminal {
		t.Fatalf("unexpected claiming result: %#v err=%v", claimed.admission, claimed.err)
	}
	if delegated.Load() != 1 || inProgress != submissions-1 {
		t.Fatalf("delegations=%d in-progress=%d", delegated.Load(), inProgress)
	}
}

func TestSameKeyDifferentIntentConflictsWithoutMutation(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	started := make(chan struct{})
	release := make(chan struct{})
	result := executeAsync(boundary, context.Background(), scope, "same-key", startIntent(t, 1), started, release)
	<-started

	if _, err := boundary.Execute(context.Background(), scope, "same-key", startIntent(t, 2), allow, success(t)); !errors.Is(err, ErrCommandKeyConflict) {
		t.Fatalf("expected key conflict, got %v", err)
	}
	observed, err := boundary.Execute(context.Background(), scope, "same-key", startIntent(t, 1), allow, success(t))
	if err != nil || observed.Kind() != AdmissionInProgress || observed.Record().Revision() != 1 {
		t.Fatalf("conflict mutated record: admission=%#v err=%v", observed, err)
	}
	close(release)
	if completed := <-result; completed.err != nil {
		t.Fatal(completed.err)
	}
}

func TestAuthorizationAndPreClaimCancellationPerformZeroMutation(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 1)
	var delegated atomic.Int32
	invoke := func() (TerminalOutcome, error) {
		delegated.Add(1)
		return terminalOutcome(t, OutcomeSucceeded, "attempt"), nil
	}

	if _, err := boundary.Execute(context.Background(), scope, "denied", intent,
		func(context.Context, Scope, Intent) error { return errDenied }, invoke); !errors.Is(err, errDenied) {
		t.Fatalf("expected authorization error, got %v", err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := boundary.Execute(canceled, scope, "canceled", intent, allow, invoke); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if delegated.Load() != 0 {
		t.Fatalf("delegated %d times before claim", delegated.Load())
	}

	for _, key := range []CommandKey{"denied", "canceled"} {
		admission, err := boundary.Execute(context.Background(), scope, key, intent, allow, success(t))
		if err != nil || admission.Kind() != AdmissionClaimed {
			t.Fatalf("%s was mutated before authorized claim: admission=%#v err=%v", key, admission, err)
		}
	}
}

func TestAuthorizationRunsForInProgressAndReplay(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 1)
	started := make(chan struct{})
	release := make(chan struct{})
	result := executeAsync(boundary, context.Background(), scope, "command", intent, started, release)
	<-started
	deny := func(context.Context, Scope, Intent) error { return errDenied }
	if _, err := boundary.Execute(context.Background(), scope, "command", intent, deny, success(t)); !errors.Is(err, errDenied) {
		t.Fatalf("in-progress submission bypassed authorization: %v", err)
	}
	close(release)
	if completed := <-result; completed.err != nil {
		t.Fatal(completed.err)
	}
	if _, err := boundary.Execute(context.Background(), scope, "command", intent, deny, success(t)); !errors.Is(err, errDenied) {
		t.Fatalf("replay bypassed authorization: %v", err)
	}
}

func TestClaimingPathCannotAbandonPermitBeforeDelegationResolves(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	started := make(chan struct{})
	release := make(chan struct{})
	returned := executeAsyncWithCallback(boundary, context.Background(), scope, "command", startIntent(t, 1),
		func() (TerminalOutcome, error) {
			close(started)
			<-release
			return TerminalOutcome{}, errors.New("claiming path lost definitive outcome")
		})
	<-started
	select {
	case got := <-returned:
		t.Fatalf("Execute returned while its private permit was live: %#v", got)
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	got := <-returned
	if !errors.Is(got.err, ErrIndeterminateExecution) {
		t.Fatalf("expected unresolved claim, got %v", got.err)
	}
	observed, err := boundary.Execute(context.Background(), scope, "command", startIntent(t, 1), allow, success(t))
	if err != nil || observed.Kind() != AdmissionInProgress {
		t.Fatalf("lost outcome was not preserved as unresolved: %#v err=%v", observed, err)
	}
	stopScope := testScope(t, "domain-a", "instance-a", OperationStop)
	if _, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow, success(t)); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("unresolved claim allowed tracked-Start exception: %v", err)
	}
}

func TestGoexitExpiresLostPermitAndLeavesUnresolvedBarrier(t *testing.T) {
	boundary := newTestBoundary(t)
	startScope := testScope(t, "domain-a", "instance-a", OperationStart)
	finished := make(chan struct{})
	var returned atomic.Bool

	go func() {
		defer close(finished)
		_, _ = boundary.Execute(context.Background(), startScope, "start", startIntent(t, 1), allow,
			func() (TerminalOutcome, error) {
				runtime.Goexit()
				return TerminalOutcome{}, nil
			})
		returned.Store(true)
	}()

	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("Goexit callback did not terminate claiming goroutine")
	}
	if returned.Load() {
		t.Fatal("Execute returned after callback terminated its goroutine")
	}

	observed, err := boundary.Execute(context.Background(), startScope, "start", startIntent(t, 1), allow, success(t))
	if err != nil || observed.Kind() != AdmissionInProgress || observed.Record().State() != CommandStateClaimed {
		t.Fatalf("lost callback did not leave an unresolved Claim: %#v err=%v", observed, err)
	}
	stopScope := testScope(t, "domain-a", "instance-a", OperationStop)
	if _, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow, success(t)); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("lost permit allowed tracked-Start Stop exception: %v", err)
	}
}

func TestPostClaimCancellationLeavesUnresolvedBarrierAndNoDuplicateDelegation(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 1)
	ctx, cancel := context.WithCancel(context.Background())
	claimed := make(chan struct{})
	result := executeAsyncWithCallback(boundary, ctx, scope, "command", intent,
		func() (TerminalOutcome, error) {
			close(claimed)
			<-ctx.Done()
			return TerminalOutcome{}, ctx.Err()
		})
	<-claimed
	cancel()
	got := <-result
	if !errors.Is(got.err, ErrIndeterminateExecution) {
		t.Fatalf("post-claim cancellation was not indeterminate: %v", got.err)
	}
	var delegated atomic.Int32
	replay, err := boundary.Execute(context.Background(), scope, "command", intent, allow,
		func() (TerminalOutcome, error) {
			delegated.Add(1)
			return terminalOutcome(t, OutcomeSucceeded, "attempt"), nil
		})
	if err != nil || replay.Kind() != AdmissionInProgress || delegated.Load() != 0 {
		t.Fatalf("retry delegated after cancellation: %#v err=%v delegated=%d", replay, err, delegated.Load())
	}
	otherScope := testScope(t, "domain-a", "instance-a", OperationStop)
	if _, err := boundary.Execute(context.Background(), otherScope, "other", NewStopIntent(), allow, success(t)); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("post-claim cancellation did not close barrier: %v", err)
	}
}

func TestTrackedStartAllowsExactlyOneDistinctStop(t *testing.T) {
	boundary := newTestBoundary(t)
	startScope := testScope(t, "domain-a", "instance-a", OperationStart)
	stopScope := testScope(t, "domain-a", "instance-a", OperationStop)
	startEntered, releaseStart := make(chan struct{}), make(chan struct{})
	startResult := executeAsync(boundary, context.Background(), startScope, "start", startIntent(t, 9), startEntered, releaseStart)
	<-startEntered
	stopEntered, releaseStop := make(chan struct{}), make(chan struct{})
	stopResult := executeAsync(boundary, context.Background(), stopScope, "stop", NewStopIntent(), stopEntered, releaseStop)
	<-stopEntered

	if _, err := boundary.Execute(context.Background(), stopScope, "other-stop", NewStopIntent(), allow, success(t)); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("second Stop was not blocked: %v", err)
	}
	if _, err := boundary.Execute(context.Background(), startScope, "other-start", startIntent(t, 10), allow, success(t)); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("second Start was not blocked: %v", err)
	}
	close(releaseStop)
	if got := <-stopResult; got.err != nil {
		t.Fatal(got.err)
	}
	if _, err := boundary.Execute(context.Background(), stopScope, "post-terminal-stop", NewStopIntent(), allow, success(t)); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("tracked Start issued a second Stop after first terminalized: %v", err)
	}
	close(releaseStart)
	if got := <-startResult; got.err != nil {
		t.Fatal(got.err)
	}
	if next, err := boundary.Execute(context.Background(), startScope, "next", startIntent(t, 11), allow, success(t)); err != nil || next.Kind() != AdmissionClaimed {
		t.Fatalf("terminal commands did not reopen barrier: %#v err=%v", next, err)
	}
}

func TestIndeterminateAndPanickingExecutionBecomeUnresolved(t *testing.T) {
	for _, test := range []struct {
		name   string
		invoke func() (TerminalOutcome, error)
	}{
		{name: "error", invoke: func() (TerminalOutcome, error) { return TerminalOutcome{}, errors.New("transport lost") }},
		{name: "panic", invoke: func() (TerminalOutcome, error) { panic("boom") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			boundary := newTestBoundary(t)
			scope := testScope(t, "domain-a", "instance-a", OperationStart)
			claimed, err := boundary.Execute(context.Background(), scope, "command", startIntent(t, 1), allow, test.invoke)
			if !errors.Is(err, ErrIndeterminateExecution) || claimed.Kind() != AdmissionClaimed {
				t.Fatalf("expected indeterminate Claim, admission=%#v err=%v", claimed, err)
			}
			observed, err := boundary.Execute(context.Background(), scope, "command", startIntent(t, 1), allow, success(t))
			if err != nil || observed.Kind() != AdmissionInProgress || observed.Record().State() != CommandStateClaimed {
				t.Fatalf("claim was not unresolved: %#v err=%v", observed, err)
			}
		})
	}
}

func TestClientReconstructionPreservesFactsAndExpiresInFlightExecution(t *testing.T) {
	storage := NewMemoryStorage()
	first := newBoundary(t, storage)
	terminalScope := testScope(t, "domain-a", "terminal-instance", OperationStart)
	terminalIntent := startIntent(t, 1)
	want := terminalOutcome(t, OutcomeSucceeded, "attempt-terminal")
	if _, err := first.Execute(context.Background(), terminalScope, "terminal", terminalIntent, allow,
		func() (TerminalOutcome, error) { return want, nil }); err != nil {
		t.Fatal(err)
	}

	claimedScope := testScope(t, "domain-a", "claimed-instance", OperationStart)
	entered, release := make(chan struct{}), make(chan struct{})
	inFlight := executeAsync(first, context.Background(), claimedScope, "claimed", startIntent(t, 2), entered, release)
	<-entered
	second := newBoundary(t, storage)
	replay, err := second.Execute(context.Background(), terminalScope, "terminal", terminalIntent, allow, success(t))
	if err != nil || replay.Kind() != AdmissionReplay {
		t.Fatalf("terminal facts were not preserved: %#v err=%v", replay, err)
	}
	close(release)
	if got := <-inFlight; !errors.Is(got.err, ErrBoundaryExpired) {
		t.Fatalf("in-flight old client published after reconstruction: %v", got.err)
	}
	observed, err := second.Execute(context.Background(), claimedScope, "claimed", startIntent(t, 2), allow, success(t))
	if err != nil || observed.Kind() != AdmissionInProgress {
		t.Fatalf("claim fact was not preserved unresolved: %#v err=%v", observed, err)
	}
	stopScope := testScope(t, "domain-a", "claimed-instance", OperationStop)
	if _, err := second.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow, success(t)); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("reconstructed unresolved claim did not block: %v", err)
	}
}

func TestStaleBoundaryCannotClaimAfterClientReconstruction(t *testing.T) {
	storage := NewMemoryStorage()
	stale := newBoundary(t, storage)
	active := newBoundary(t, storage)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 1)
	var staleDelegations atomic.Int32
	if _, err := stale.Execute(context.Background(), scope, "command", intent, allow,
		func() (TerminalOutcome, error) {
			staleDelegations.Add(1)
			return terminalOutcome(t, OutcomeSucceeded, "attempt"), nil
		}); !errors.Is(err, ErrBoundaryExpired) {
		t.Fatalf("stale Boundary did not fail closed: %v", err)
	}
	if staleDelegations.Load() != 0 {
		t.Fatalf("stale Boundary delegated %d times", staleDelegations.Load())
	}
	claimed, err := active.Execute(context.Background(), scope, "command", intent, allow, success(t))
	if err != nil || claimed.Kind() != AdmissionClaimed {
		t.Fatalf("stale Boundary mutated storage: %#v err=%v", claimed, err)
	}
}

func TestDifferentInstancesExecuteIndependently(t *testing.T) {
	boundary := newTestBoundary(t)
	firstScope := testScope(t, "domain-a", "instance-a", OperationStart)
	secondScope := testScope(t, "domain-a", "instance-b", OperationStart)
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	results := make(chan executionResult, 2)
	for i, scope := range []Scope{firstScope, secondScope} {
		go func() {
			admission, err := boundary.Execute(context.Background(), scope, CommandKey(rune('a'+i)), startIntent(t, 1), allow,
				func() (TerminalOutcome, error) {
					entered <- struct{}{}
					<-release
					return terminalOutcome(t, OutcomeSucceeded, "attempt"), nil
				})
			results <- executionResult{admission: admission, err: err}
		}()
	}
	for range 2 {
		select {
		case <-entered:
		case <-time.After(2 * time.Second):
			t.Fatal("different Instance execution was serialized")
		}
	}
	close(release)
	for range 2 {
		if got := <-results; got.err != nil {
			t.Fatal(got.err)
		}
	}
}

func TestCommandIdentityIsIsolatedByScope(t *testing.T) {
	boundary := newTestBoundary(t)
	intent := startIntent(t, 1)
	for _, scope := range []Scope{
		testScope(t, "domain-a", "instance-a", OperationStart),
		testScope(t, "domain-b", "instance-a", OperationStart),
		testScope(t, "domain-a", "instance-b", OperationStart),
	} {
		admission, err := boundary.Execute(context.Background(), scope, "same-raw-key", intent, allow, success(t))
		if err != nil || admission.Kind() != AdmissionClaimed {
			t.Fatalf("scope collision: scope=%#v admission=%#v err=%v", scope, admission, err)
		}
	}
}

func TestConcurrentDifferentKeysHaveOneInstanceLinearizationPoint(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 1)
	const submissions = 24
	start := make(chan struct{})
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	results := make(chan executionResult, submissions)
	for i := range submissions {
		go func() {
			<-start
			admission, err := boundary.Execute(context.Background(), scope, CommandKey(rune('a'+i)), intent, allow,
				func() (TerminalOutcome, error) {
					entered <- struct{}{}
					<-release
					return terminalOutcome(t, OutcomeSucceeded, "attempt"), nil
				})
			results <- executionResult{admission: admission, err: err}
		}()
	}
	close(start)
	<-entered
	var blocked int
	for range submissions - 1 {
		got := <-results
		if !errors.Is(got.err, ErrInstanceBlocked) {
			t.Fatalf("expected blocked different key, got %#v err=%v", got.admission, got.err)
		}
		blocked++
	}
	close(release)
	claimed := <-results
	if claimed.err != nil || claimed.admission.Kind() != AdmissionClaimed || blocked != submissions-1 {
		t.Fatalf("claimed=%#v err=%v blocked=%d", claimed.admission, claimed.err, blocked)
	}
}

func TestValidationPrecedesAuthorization(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	var calls atomic.Int32
	authorize := func(context.Context, Scope, Intent) error { calls.Add(1); return nil }
	if _, err := boundary.Execute(context.Background(), scope, "", startIntent(t, 1), authorize, success(t)); !errors.Is(err, ErrInvalidSubmission) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("authorization called %d times for invalid submission", calls.Load())
	}
}

func TestStoredFactsExcludeAuthorityAndRawErrors(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	want := terminalOutcome(t, OutcomeFailed, "attempt-redacted")
	record, err := boundary.Execute(context.Background(), scope, "opaque", startIntent(t, 7), allow,
		func() (TerminalOutcome, error) { return want, nil })
	if err != nil {
		t.Fatal(err)
	}
	got, ok := record.Record().Outcome()
	if !ok || got != want || got.Category() != OutcomeFailed {
		t.Fatalf("semantic outcome was not stored exactly: %#v", record)
	}
	// Durable records expose no authorization callback, principal, credential,
	// context, transport response, or raw internal error field.
}

func executeAsync(
	boundary *Boundary,
	ctx context.Context,
	scope Scope,
	key CommandKey,
	intent Intent,
	entered chan<- struct{},
	release <-chan struct{},
) <-chan executionResult {
	return executeAsyncWithCallback(boundary, ctx, scope, key, intent,
		func() (TerminalOutcome, error) {
			entered <- struct{}{}
			<-release
			return TerminalOutcome{category: OutcomeSucceeded}, nil
		})
}

func executeAsyncWithCallback(
	boundary *Boundary,
	ctx context.Context,
	scope Scope,
	key CommandKey,
	intent Intent,
	invoke func() (TerminalOutcome, error),
) <-chan executionResult {
	result := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(ctx, scope, key, intent, allow, invoke)
		result <- executionResult{admission: admission, err: err}
	}()
	return result
}

func success(t *testing.T) func() (TerminalOutcome, error) {
	t.Helper()
	return func() (TerminalOutcome, error) {
		return terminalOutcome(t, OutcomeSucceeded, "attempt"), nil
	}
}

func allow(context.Context, Scope, Intent) error { return nil }

func newTestBoundary(t *testing.T) *Boundary { return newBoundary(t, NewMemoryStorage()) }

func newBoundary(t *testing.T, storage *MemoryStorage) *Boundary {
	t.Helper()
	boundary, err := NewBoundary(storage)
	if err != nil {
		t.Fatal(err)
	}
	return boundary
}

func testScope(t *testing.T, domain string, instance string, operation Operation) Scope {
	t.Helper()
	scope, err := NewScope(domain, 1, 2, runtimeconfigload.RuntimeInstanceID(instance), operation)
	if err != nil {
		t.Fatal(err)
	}
	return scope
}

func startIntent(t *testing.T, version uint64) Intent {
	t.Helper()
	intent, err := NewStartIntent(version)
	if err != nil {
		t.Fatal(err)
	}
	return intent
}

func terminalOutcome(
	t *testing.T,
	category OutcomeCategory,
	attempt runtimeconfigload.LaunchAttemptID,
) TerminalOutcome {
	t.Helper()
	outcome, err := NewTerminalOutcome(category, attempt)
	if err != nil {
		t.Fatal(err)
	}
	return outcome
}
