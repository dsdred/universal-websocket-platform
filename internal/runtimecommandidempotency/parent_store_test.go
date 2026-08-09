package runtimecommandidempotency

import (
	"context"
	"errors"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestParentValidationAuthorizationAndZeroPhaseReplay(t *testing.T) {
	boundary := newTestBoundary(t)
	replaceScope := testScope(t, "management", "parent-validation", OperationReplace)
	intent := replaceIntent(t, 7)
	if _, err := NewReplaceIntent(0); !errors.Is(err, ErrInvalidSubmission) {
		t.Fatalf("zero target version error = %v", err)
	}
	if _, err := boundary.Execute(context.Background(), replaceScope, "primitive", intent, allow, success(t)); !errors.Is(err, ErrInvalidSubmission) {
		t.Fatalf("primitive Execute accepted parent operation: %v", err)
	}
	startScope := testScope(t, "management", "parent-validation", OperationStart)
	if _, err := boundary.ExecuteParent(context.Background(), startScope, "parent", startIntent(t, 7), allow, func(*ParentExecution) error { return nil }); !errors.Is(err, ErrInvalidSubmission) {
		t.Fatalf("ExecuteParent accepted primitive operation: %v", err)
	}

	var authorized atomic.Int32
	authorize := func(context.Context, Scope, Intent) error {
		authorized.Add(1)
		return nil
	}
	outcome := parentOutcome(t, ParentOutcomeSatisfied)
	claimed, err := boundary.ExecuteParent(context.Background(), replaceScope, "parent", intent, authorize,
		func(execution *ParentExecution) error {
			view, publishErr := execution.PublishTerminal(outcome)
			if publishErr != nil {
				return publishErr
			}
			if view.State() != CommandStateTerminal || view.Revision() != 2 {
				t.Fatalf("published view = %#v", view)
			}
			return nil
		})
	if err != nil || claimed.Kind() != AdmissionClaimed || claimed.Record().State() != CommandStateTerminal {
		t.Fatalf("claimed = %#v, err = %v", claimed, err)
	}

	replay, err := boundary.ExecuteParent(context.Background(), replaceScope, "parent", intent, authorize,
		func(*ParentExecution) error { t.Fatal("replay received execution"); return nil })
	if err != nil || replay.Kind() != AdmissionReplay {
		t.Fatalf("replay = %#v, err = %v", replay, err)
	}
	if got := authorized.Load(); got != 2 {
		t.Fatalf("authorization calls = %d, want 2", got)
	}
}

func TestParentAuthorizationFailureCancellationAndIntentConflictDoNotMutate(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "management", "parent-zero-mutation", OperationRollback)
	intent := rollbackIntent(t, 8)
	denied := errors.New("denied")
	if _, err := boundary.ExecuteParent(context.Background(), scope, "parent", intent,
		func(context.Context, Scope, Intent) error { return denied }, func(*ParentExecution) error { return nil }); !errors.Is(err, denied) {
		t.Fatalf("denial error = %v", err)
	}
	if _, err := boundary.ExecuteParent(context.Background(), scope, "parent", intent,
		func(context.Context, Scope, Intent) error { panic("authorization secret") }, func(*ParentExecution) error { return nil }); !errors.Is(err, ErrInvalidSubmission) {
		t.Fatalf("authorization panic error = %v", err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := boundary.ExecuteParent(canceled, scope, "parent", intent, allow, func(*ParentExecution) error { return nil }); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}
	claimed, err := boundary.ExecuteParent(context.Background(), scope, "parent", intent, allow,
		func(execution *ParentExecution) error {
			_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeSatisfied))
			return publishErr
		})
	if err != nil || claimed.Kind() != AdmissionClaimed {
		t.Fatalf("claim after zero-mutation failures = %#v, %v", claimed, err)
	}
	if _, err := boundary.ExecuteParent(context.Background(), scope, "parent", rollbackIntent(t, 9), allow,
		func(*ParentExecution) error { return nil }); !errors.Is(err, ErrCommandKeyConflict) {
		t.Fatalf("intent conflict error = %v", err)
	}
}

func TestParentAuthorizationRunsBeforeInProgressInspection(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "management", "parent-in-progress-auth", OperationReplace)
	intent := replaceIntent(t, 20)
	entered := make(chan struct{})
	release := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), scope, "parent", intent, allow,
			func(execution *ParentExecution) error {
				close(entered)
				<-release
				_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeSatisfied))
				return publishErr
			})
		result <- err
	}()
	<-entered
	denied := errors.New("fresh authorization denied")
	if _, err := boundary.ExecuteParent(context.Background(), scope, "parent", intent,
		func(context.Context, Scope, Intent) error { return denied },
		func(*ParentExecution) error { t.Fatal("denied in-progress request received capability"); return nil }); !errors.Is(err, denied) {
		t.Fatalf("in-progress authorization error = %v", err)
	}
	close(release)
	if err := <-result; err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentSameParentAndPhaseDelegateAtMostOnce(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "management", "parent-concurrency", OperationReplace)
	intent := replaceIntent(t, 10)
	var parentCalls atomic.Int32
	var phaseCalls atomic.Int32
	const callers = 24
	results := make(chan error, callers)
	start := make(chan struct{})
	for range callers {
		go func() {
			<-start
			_, err := boundary.ExecuteParent(context.Background(), scope, "same-parent", intent, allow,
				func(execution *ParentExecution) error {
					parentCalls.Add(1)
					var wg sync.WaitGroup
					phaseErrors := make(chan error, callers)
					for range callers {
						wg.Add(1)
						go func() {
							defer wg.Done()
							_, phaseErr := execution.inspectOrExecuteStartTarget(func() (TerminalOutcome, error) {
								phaseCalls.Add(1)
								return terminalOutcome(t, OutcomeSucceeded, "parent-concurrency-attempt"), nil
							})
							phaseErrors <- phaseErr
						}()
					}
					wg.Wait()
					close(phaseErrors)
					for phaseErr := range phaseErrors {
						if phaseErr != nil {
							return phaseErr
						}
					}
					_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeSucceeded))
					return publishErr
				})
			results <- err
		}()
	}
	close(start)
	for range callers {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	if got := parentCalls.Load(); got != 1 {
		t.Fatalf("parent callbacks = %d, want 1", got)
	}
	if got := phaseCalls.Load(); got != 1 {
		t.Fatalf("phase callbacks = %d, want 1", got)
	}
}

func TestParentPhaseFiniteOrderReplayAndTerminalGate(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "management", "parent-order", OperationReplace)
	intent := replaceIntent(t, 11)
	_, err := boundary.ExecuteParent(context.Background(), scope, "ordered", intent, allow,
		func(execution *ParentExecution) error {
			stop, stopErr := execution.InspectOrExecuteStopOld(func() (TerminalOutcome, error) {
				return terminalOutcome(t, OutcomeSucceeded, "old-attempt"), nil
			})
			if stopErr != nil || stop.Record().Ordinal() != 0 || stop.Record().Kind() != PhaseStopOld {
				t.Fatalf("StopOld = %#v, %v", stop, stopErr)
			}
			if _, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeSucceeded)); !errors.Is(publishErr, ErrInstanceBlocked) {
				t.Fatalf("parent terminalized before StartTarget: %v", publishErr)
			}
			start, startErr := execution.inspectOrExecuteStartTarget(func() (TerminalOutcome, error) {
				return terminalOutcome(t, OutcomeSucceeded, "target-attempt"), nil
			})
			if startErr != nil || start.Record().Ordinal() != 1 || start.Record().Kind() != PhaseStartTarget {
				t.Fatalf("StartTarget = %#v, %v", start, startErr)
			}
			if _, replayErr := execution.inspectOrExecuteStartTarget(func() (TerminalOutcome, error) {
				t.Fatal("phase replay invoked lifecycle")
				return TerminalOutcome{}, nil
			}); replayErr != nil {
				return replayErr
			}
			if _, orderErr := execution.InspectOrExecuteStopOld(success(t)); orderErr != nil {
				// Existing StopOld is a replay even after StartTarget and remains valid.
				return orderErr
			}
			_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeSucceeded))
			return publishErr
		})
	if err != nil {
		t.Fatal(err)
	}

	secondScope := testScope(t, "management", "parent-order-2", OperationRollback)
	_, err = boundary.ExecuteParent(context.Background(), secondScope, "start-first", rollbackIntent(t, 12), allow,
		func(execution *ParentExecution) error {
			if _, phaseErr := execution.inspectOrExecuteStartTarget(success(t)); phaseErr != nil {
				return phaseErr
			}
			if _, phaseErr := execution.InspectOrExecuteStopOld(success(t)); !errors.Is(phaseErr, ErrIllegalPhaseOrder) {
				t.Fatalf("StopOld after StartTarget error = %v", phaseErr)
			}
			_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeSucceeded))
			return publishErr
		})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIndeterminatePhaseAndAbandonedParentLeaveDurableBarrier(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "management", "parent-indeterminate", OperationReplace)
	intent := replaceIntent(t, 13)
	claimed, err := boundary.ExecuteParent(context.Background(), scope, "parent", intent, allow,
		func(execution *ParentExecution) error {
			_, phaseErr := execution.inspectOrExecuteStartTarget(func() (TerminalOutcome, error) {
				return TerminalOutcome{}, errors.New("raw lifecycle error")
			})
			return phaseErr
		})
	if !errors.Is(err, ErrIndeterminateExecution) || claimed.Record().State() != CommandStateClaimed {
		t.Fatalf("indeterminate claim = %#v, %v", claimed, err)
	}
	observed, err := boundary.ExecuteParent(context.Background(), scope, "parent", intent, allow,
		func(*ParentExecution) error { t.Fatal("unresolved replay received execution"); return nil })
	if err != nil || observed.Kind() != AdmissionInProgress {
		t.Fatalf("unresolved observation = %#v, %v", observed, err)
	}
	primitiveScope := testScope(t, "management", "parent-indeterminate", OperationStop)
	if _, err := boundary.Execute(context.Background(), primitiveScope, "stop", NewStopIntent(), allow, success(t)); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("primitive bypassed unresolved parent/phase: %v", err)
	}
}

func TestReturningOrPanickingParentWithoutPublicationLeavesUnresolved(t *testing.T) {
	for _, test := range []struct {
		name   string
		invoke func(*ParentExecution) error
	}{
		{name: "return", invoke: func(*ParentExecution) error { return nil }},
		{name: "error", invoke: func(*ParentExecution) error { return errors.New("raw parent error") }},
		{name: "panic", invoke: func(*ParentExecution) error { panic("raw parent panic") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			boundary := newTestBoundary(t)
			scope := testScope(t, "management", "abandoned-parent-"+test.name, OperationReplace)
			intent := replaceIntent(t, 19)
			claimed, err := boundary.ExecuteParent(context.Background(), scope, "parent", intent, allow, test.invoke)
			if !errors.Is(err, ErrIndeterminateExecution) || claimed.Record().State() != CommandStateClaimed {
				t.Fatalf("abandoned parent = %#v, %v", claimed, err)
			}
			observed, err := boundary.ExecuteParent(context.Background(), scope, "parent", intent, allow,
				func(*ParentExecution) error { t.Fatal("abandoned parent received another capability"); return nil })
			if err != nil || observed.Kind() != AdmissionInProgress {
				t.Fatalf("abandoned parent observation = %#v, %v", observed, err)
			}
		})
	}
}

func TestParentCapabilityRejectsUseAfterCallbackReturn(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "management", "parent-post-return", OperationReplace)
	intent := replaceIntent(t, 21)
	var retained *ParentExecution
	var copied ParentExecution
	_, err := boundary.ExecuteParent(context.Background(), scope, "parent", intent, allow,
		func(execution *ParentExecution) error {
			retained = execution
			copied = *execution
			return nil
		})
	if !errors.Is(err, ErrIndeterminateExecution) {
		t.Fatalf("callback return error = %v", err)
	}
	var invoked atomic.Int32
	if _, err := retained.inspectOrExecuteStartTarget(func() (TerminalOutcome, error) {
		invoked.Add(1)
		return terminalOutcome(t, OutcomeSucceeded, "escaped-attempt"), nil
	}); !errors.Is(err, ErrBoundaryExpired) {
		t.Fatalf("post-return phase claim error = %v", err)
	}
	if _, err := retained.PublishTerminal(parentOutcome(t, ParentOutcomeSatisfied)); !errors.Is(err, ErrBoundaryExpired) {
		t.Fatalf("post-return parent publication error = %v", err)
	}
	if _, err := copied.inspectOrExecuteStartTarget(success(t)); !errors.Is(err, ErrBoundaryExpired) {
		t.Fatalf("copied post-return capability error = %v", err)
	}
	if invoked.Load() != 0 {
		t.Fatalf("post-return capability invoked lifecycle %d times", invoked.Load())
	}
}

func TestExecuteParentWaitsForInFlightPhaseAtCallbackReturn(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "management", "parent-in-flight-join", OperationReplace)
	intent := replaceIntent(t, 22)
	phaseEntered := make(chan struct{})
	releasePhase := make(chan struct{})
	phaseDone := make(chan error, 1)
	parentResult := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), scope, "parent", intent, allow,
			func(execution *ParentExecution) error {
				go func() {
					_, phaseErr := execution.inspectOrExecuteStartTarget(func() (TerminalOutcome, error) {
						close(phaseEntered)
						<-releasePhase
						return terminalOutcome(t, OutcomeSucceeded, "joined-attempt"), nil
					})
					phaseDone <- phaseErr
				}()
				<-phaseEntered
				return nil
			})
		parentResult <- err
	}()
	<-phaseEntered
	select {
	case err := <-parentResult:
		t.Fatalf("ExecuteParent returned before in-flight phase completed: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	close(releasePhase)
	if err := <-phaseDone; err != nil {
		t.Fatal(err)
	}
	if err := <-parentResult; !errors.Is(err, ErrIndeterminateExecution) {
		t.Fatalf("joined parent result = %v", err)
	}
}

func TestPostClaimCancellationCannotDuplicateParentOrPhase(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "management", "parent-cancellation", OperationRollback)
	intent := rollbackIntent(t, 23)
	ctx, cancel := context.WithCancel(context.Background())
	phaseEntered := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(ctx, scope, "parent", intent, allow,
			func(execution *ParentExecution) error {
				_, phaseErr := execution.inspectOrExecuteStartTarget(func() (TerminalOutcome, error) {
					close(phaseEntered)
					<-ctx.Done()
					return TerminalOutcome{}, ctx.Err()
				})
				return phaseErr
			})
		result <- err
	}()
	<-phaseEntered
	cancel()
	if err := <-result; !errors.Is(err, ErrIndeterminateExecution) {
		t.Fatalf("post-claim cancellation result = %v", err)
	}
	var delegated atomic.Int32
	observed, err := boundary.ExecuteParent(context.Background(), scope, "parent", intent, allow,
		func(*ParentExecution) error { delegated.Add(1); return nil })
	if err != nil || observed.Kind() != AdmissionInProgress || delegated.Load() != 0 {
		t.Fatalf("cancellation retry = %#v, %v, delegated=%d", observed, err, delegated.Load())
	}
	primitiveScope := testScope(t, "management", "parent-cancellation", OperationStop)
	if _, err := boundary.Execute(context.Background(), primitiveScope, "stop", NewStopIntent(), allow, success(t)); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("cancellation lost parent/phase barrier: %v", err)
	}
}

func TestParentAndPhaseStoredFactsAreRedacted(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "management", "parent-redaction", OperationReplace)
	intent := replaceIntent(t, 24)
	wantPhase := terminalOutcome(t, OutcomeFailed, "redacted-attempt")
	wantParent := parentOutcome(t, ParentOutcomeFailed)
	admission, err := boundary.ExecuteParent(context.Background(), scope, "parent", intent, allow,
		func(execution *ParentExecution) error {
			if _, phaseErr := execution.inspectOrExecuteStartTarget(func() (TerminalOutcome, error) {
				return wantPhase, nil
			}); phaseErr != nil {
				return phaseErr
			}
			_, publishErr := execution.PublishTerminal(wantParent)
			return publishErr
		})
	if err != nil {
		t.Fatal(err)
	}
	gotParent, ok := admission.Record().Outcome()
	if !ok || gotParent != wantParent {
		t.Fatalf("parent outcome = %#v, %v", gotParent, ok)
	}
	ledger := boundary.storage.ledger(scope.instanceScope())
	ledger.mu.Lock()
	defer ledger.mu.Unlock()
	for _, typ := range []reflect.Type{reflect.TypeOf(parentRecord{}), reflect.TypeOf(phaseRecord{})} {
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			name := strings.ToLower(field.Name)
			if field.Type.Kind() == reflect.Func || field.Type.Kind() == reflect.Interface ||
				strings.Contains(name, "author") || strings.Contains(name, "context") ||
				strings.Contains(name, "error") || strings.Contains(name, "host") ||
				strings.Contains(name, "snapshot") || strings.Contains(name, "token") {
				t.Fatalf("durable %s exposes forbidden field %s %s", typ.Name(), field.Name, field.Type)
			}
		}
	}
}

func TestParentGoexitExpiresCapabilityAndLeavesUnresolved(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "management", "parent-goexit", OperationRollback)
	intent := rollbackIntent(t, 14)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = boundary.ExecuteParent(context.Background(), scope, "parent", intent, allow,
			func(*ParentExecution) error {
				runtime.Goexit()
				return nil
			})
	}()
	<-done
	observed, err := boundary.ExecuteParent(context.Background(), scope, "parent", intent, allow,
		func(*ParentExecution) error { t.Fatal("Goexit claim received another execution"); return nil })
	if err != nil || observed.Kind() != AdmissionInProgress {
		t.Fatalf("Goexit observation = %#v, %v", observed, err)
	}
}

func TestPhaseGoexitExpiresBothCapabilitiesAndLeavesUnresolved(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "management", "phase-goexit", OperationReplace)
	intent := replaceIntent(t, 18)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = boundary.ExecuteParent(context.Background(), scope, "parent", intent, allow,
			func(execution *ParentExecution) error {
				_, _ = execution.inspectOrExecuteStartTarget(func() (TerminalOutcome, error) {
					runtime.Goexit()
					return TerminalOutcome{}, nil
				})
				return nil
			})
	}()
	<-done
	observed, err := boundary.ExecuteParent(context.Background(), scope, "parent", intent, allow,
		func(*ParentExecution) error { t.Fatal("phase Goexit restored parent capability"); return nil })
	if err != nil || observed.Kind() != AdmissionInProgress {
		t.Fatalf("phase Goexit observation = %#v, %v", observed, err)
	}
	primitiveScope := testScope(t, "management", "phase-goexit", OperationStop)
	if _, err := boundary.Execute(context.Background(), primitiveScope, "stop", NewStopIntent(), allow, success(t)); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("phase Goexit lost durable barrier: %v", err)
	}
}

func TestParentReconstructionPreservesFactsAndExpiresCapabilities(t *testing.T) {
	storage := NewMemoryStorage()
	first := newBoundary(t, storage)
	scope := testScope(t, "management", "parent-reconstruction", OperationReplace)
	intent := replaceIntent(t, 15)
	entered := make(chan struct{})
	release := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		_, err := first.ExecuteParent(context.Background(), scope, "parent", intent, allow,
			func(execution *ParentExecution) error {
				_, phaseErr := execution.inspectOrExecuteStartTarget(func() (TerminalOutcome, error) {
					close(entered)
					<-release
					return terminalOutcome(t, OutcomeSucceeded, "stale-attempt"), nil
				})
				return phaseErr
			})
		result <- err
	}()
	<-entered
	second := newBoundary(t, storage)
	close(release)
	if err := <-result; !errors.Is(err, ErrIndeterminateExecution) {
		t.Fatalf("stale parent result = %v", err)
	}
	observed, err := second.ExecuteParent(context.Background(), scope, "parent", intent, allow,
		func(*ParentExecution) error { t.Fatal("reconstruction restored capability"); return nil })
	if err != nil || observed.Kind() != AdmissionInProgress {
		t.Fatalf("reconstructed observation = %#v, %v", observed, err)
	}
	other := testScope(t, "management", "parent-reconstruction", OperationRollback)
	if _, err := second.ExecuteParent(context.Background(), other, "other", rollbackIntent(t, 16), allow,
		func(*ParentExecution) error { return nil }); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("reconstructed unresolved barrier = %v", err)
	}
}

func TestParentDifferentInstancesProgressIndependently(t *testing.T) {
	boundary := newTestBoundary(t)
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, instance := range []string{"parent-instance-a", "parent-instance-b"} {
		instance := instance
		wg.Add(1)
		go func() {
			defer wg.Done()
			scope := testScope(t, "management", instance, OperationReplace)
			_, err := boundary.ExecuteParent(context.Background(), scope, "parent", replaceIntent(t, 17), allow,
				func(execution *ParentExecution) error {
					if _, phaseErr := execution.inspectOrExecuteStartTarget(success(t)); phaseErr != nil {
						return phaseErr
					}
					_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeSucceeded))
					return publishErr
				})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func replaceIntent(t *testing.T, version uint64) Intent {
	t.Helper()
	intent, err := NewReplaceIntent(version)
	if err != nil {
		t.Fatal(err)
	}
	return intent
}

func rollbackIntent(t *testing.T, version uint64) Intent {
	t.Helper()
	intent, err := NewRollbackIntent(version)
	if err != nil {
		t.Fatal(err)
	}
	return intent
}

func parentOutcome(t *testing.T, category ParentOutcomeCategory) ParentTerminalOutcome {
	t.Helper()
	outcome, err := NewParentTerminalOutcome(category)
	if err != nil {
		t.Fatal(err)
	}
	return outcome
}
