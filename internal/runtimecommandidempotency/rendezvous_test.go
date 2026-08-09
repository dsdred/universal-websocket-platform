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

func TestContinueGateStopFirstCreatesNoStartTarget(t *testing.T) {
	boundary := newTestBoundary(t)
	parentScope := testScope(t, "management", "stop-first", OperationReplace)
	stopScope := testScope(t, "management", "stop-first", OperationStop)
	parentEntered := make(chan struct{})
	continueNow := make(chan struct{})
	parentResult := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", replaceIntent(t, 31), allow,
			func(execution *ParentExecution) error {
				close(parentEntered)
				<-continueNow
				phase, prevented, continueErr := execution.ContinueOrExecuteStartTarget(context.Background(),
					func(*StartTargetExecution) (TerminalOutcome, error) {
						return TerminalOutcome{}, errors.New("Stop-first gate invoked StartTarget")
					})
				if continueErr != nil || !prevented || phase != (PhaseAdmission{}) {
					return errors.New("invalid Stop-first gate result")
				}
				_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeStopped))
				return publishErr
			})
		parentResult <- err
	}()
	<-parentEntered

	var stopCalls atomic.Int32
	stopResult := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow,
			func() (TerminalOutcome, error) {
				stopCalls.Add(1)
				return terminalOutcome(t, OutcomeSucceeded, "must-not-run"), nil
			})
		stopResult <- executionResult{admission: admission, err: err}
	}()
	waitForCommandRecord(t, boundary, stopScope, "stop")
	close(continueNow)
	if got := <-stopResult; got.err != nil || got.admission.Record().State() != CommandStateTerminal {
		t.Fatalf("Stop-first admission = %#v, err=%v", got.admission, got.err)
	}
	if err := <-parentResult; err != nil {
		t.Fatal(err)
	}
	if stopCalls.Load() != 0 {
		t.Fatalf("Stop-first invoked lifecycle %d times", stopCalls.Load())
	}
	ledger := boundary.storage.ledger(parentScope.instanceScope())
	ledger.mu.Lock()
	defer ledger.mu.Unlock()
	startID, _ := newPhaseIdentity(commandIdentity{scope: parentScope, key: "parent"}, PhaseStartTarget)
	if ledger.phases[startID] != nil {
		t.Fatal("Stop-first created a StartTarget record")
	}
}

func TestStartTargetFirstRendezvousUsesOriginalStopStackOnce(t *testing.T) {
	boundary := newTestBoundary(t)
	parentScope := testScope(t, "management", "start-first", OperationRollback)
	stopScope := testScope(t, "management", "start-first", OperationStop)
	startEntered := make(chan struct{})
	ownerNow := make(chan struct{})
	stopEntered := make(chan struct{}, 1)
	releaseStop := make(chan struct{})
	parentResult := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", rollbackIntent(t, 32), allow,
			func(execution *ParentExecution) error {
				phase, prevented, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
					func(start *StartTargetExecution) (TerminalOutcome, error) {
						close(startEntered)
						<-ownerNow
						stopped, signalErr := start.OwnerClaimed(runtimeconfigload.LaunchAttemptID("attempt-32"))
						if signalErr != nil || !stopped {
							return TerminalOutcome{}, errors.New("pending Stop did not converge")
						}
						return terminalOutcome(t, OutcomeRejected, "attempt-32"), nil
					})
				if phaseErr != nil || prevented || phase.Kind() != AdmissionClaimed {
					return phaseErr
				}
				_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeStopped))
				return publishErr
			})
		parentResult <- err
	}()
	<-startEntered

	stopResult := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow,
			func() (TerminalOutcome, error) {
				stopEntered <- struct{}{}
				<-releaseStop
				return terminalOutcome(t, OutcomeSucceeded, "attempt-32"), nil
			})
		stopResult <- executionResult{admission: admission, err: err}
	}()
	waitForCommandRecord(t, boundary, stopScope, "stop")
	if _, err := boundary.Execute(context.Background(), stopScope, "second-stop", NewStopIntent(), allow, success(t)); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("second pending Stop error = %v", err)
	}
	select {
	case <-stopEntered:
		t.Fatal("pending Stop invoked before OwnerClaimed")
	case <-time.After(30 * time.Millisecond):
	}
	close(ownerNow)
	select {
	case <-stopEntered:
	case <-time.After(2 * time.Second):
		t.Fatal("original Stop stack did not invoke after OwnerClaimed")
	}
	select {
	case err := <-parentResult:
		t.Fatalf("StartTarget did not wait for Stop convergence: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	close(releaseStop)
	if got := <-stopResult; got.err != nil || got.admission.Kind() != AdmissionClaimed {
		t.Fatalf("pending Stop result = %#v, %v", got.admission, got.err)
	}
	if err := <-parentResult; err != nil {
		t.Fatal(err)
	}
}

func TestDefinitiveStartNoClaimConsumesPendingStopWithoutLifecycle(t *testing.T) {
	boundary := newTestBoundary(t)
	parentScope := testScope(t, "management", "start-no-claim", OperationReplace)
	stopScope := testScope(t, "management", "start-no-claim", OperationStop)
	startEntered := make(chan struct{})
	returnStart := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	parentResult := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", replaceIntent(t, 33), allow,
			func(execution *ParentExecution) error {
				_, _, phaseErr := execution.ContinueOrExecuteStartTarget(ctx,
					func(start *StartTargetExecution) (TerminalOutcome, error) {
						close(startEntered)
						<-returnStart
						cancel()
						if err := start.StartNoClaim(StartNoClaimCancelled); err != nil {
							return TerminalOutcome{}, err
						}
						return terminalOutcome(t, OutcomeRejected, ""), nil
					})
				if phaseErr != nil {
					return phaseErr
				}
				for _, category := range []ParentOutcomeCategory{
					ParentOutcomeStopped, ParentOutcomeSucceeded, ParentOutcomeSatisfied,
				} {
					if _, publishErr := execution.PublishTerminal(parentOutcome(t, category)); !errors.Is(publishErr, ErrInstanceBlocked) {
						return errors.New("cancelled StartNoClaim accepted false parent outcome")
					}
				}
				_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeCancelled))
				return publishErr
			})
		parentResult <- err
	}()
	<-startEntered
	var stopCalls atomic.Int32
	stopResult := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow,
			func() (TerminalOutcome, error) {
				stopCalls.Add(1)
				return terminalOutcome(t, OutcomeSucceeded, "unexpected"), nil
			})
		stopResult <- executionResult{admission: admission, err: err}
	}()
	waitForCommandRecord(t, boundary, stopScope, "stop")
	close(returnStart)
	if got := <-stopResult; got.err != nil || got.admission.Record().State() != CommandStateTerminal {
		t.Fatalf("StartNoClaim Stop result = %#v, %v", got.admission, got.err)
	}
	if err := <-parentResult; err != nil {
		t.Fatal(err)
	}
	if stopCalls.Load() != 0 {
		t.Fatalf("StartNoClaim invoked Stop lifecycle %d times", stopCalls.Load())
	}
}

func TestPendingStopStartNoClaimFailedMapsOnlyParentFailed(t *testing.T) {
	boundary := newTestBoundary(t)
	parentScope := testScope(t, "management", "start-no-claim-failed", OperationRollback)
	stopScope := testScope(t, "management", "start-no-claim-failed", OperationStop)
	startEntered := make(chan struct{})
	returnStart := make(chan struct{})
	parentResult := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", rollbackIntent(t, 55), allow,
			func(execution *ParentExecution) error {
				_, _, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
					func(start *StartTargetExecution) (TerminalOutcome, error) {
						close(startEntered)
						<-returnStart
						if signalErr := start.StartNoClaim(StartNoClaimFailed); signalErr != nil {
							return TerminalOutcome{}, signalErr
						}
						return terminalOutcome(t, OutcomeFailed, ""), nil
					})
				if phaseErr != nil {
					return phaseErr
				}
				for _, category := range []ParentOutcomeCategory{
					ParentOutcomeStopped, ParentOutcomeSucceeded, ParentOutcomeSatisfied,
					ParentOutcomeCancelled, ParentOutcomeRejected,
				} {
					if _, publishErr := execution.PublishTerminal(parentOutcome(t, category)); !errors.Is(publishErr, ErrInstanceBlocked) {
						return errors.New("failed StartNoClaim accepted false parent outcome")
					}
				}
				_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeFailed))
				return publishErr
			})
		parentResult <- err
	}()
	<-startEntered
	var stopCalls atomic.Int32
	stopResult := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow,
			func() (TerminalOutcome, error) {
				stopCalls.Add(1)
				return terminalOutcome(t, OutcomeSucceeded, "unexpected"), nil
			})
		stopResult <- executionResult{admission: admission, err: err}
	}()
	waitForCommandRecord(t, boundary, stopScope, "stop")
	close(returnStart)
	if got := <-stopResult; got.err != nil || got.admission.Record().State() != CommandStateTerminal {
		t.Fatalf("failed StartNoClaim Stop = %#v, %v", got.admission, got.err)
	}
	if err := <-parentResult; err != nil {
		t.Fatal(err)
	}
	if stopCalls.Load() != 0 {
		t.Fatalf("failed StartNoClaim invoked Stop callback %d times", stopCalls.Load())
	}
}

func TestExplicitRejectedCauseSurvivesLateContextCancellation(t *testing.T) {
	boundary := newTestBoundary(t)
	parentScope := testScope(t, "management", "rejected-late-cancel", OperationReplace)
	stopScope := testScope(t, "management", "rejected-late-cancel", OperationStop)
	ctx, cancel := context.WithCancel(context.Background())
	startEntered := make(chan struct{})
	signalNow := make(chan struct{})
	causeSelected := make(chan struct{})
	returnNow := make(chan struct{})
	parentResult := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", replaceIntent(t, 56), allow,
			func(execution *ParentExecution) error {
				_, _, phaseErr := execution.ContinueOrExecuteStartTarget(ctx,
					func(start *StartTargetExecution) (TerminalOutcome, error) {
						close(startEntered)
						<-signalNow
						if signalErr := start.StartNoClaim(StartNoClaimRejected); signalErr != nil {
							return TerminalOutcome{}, signalErr
						}
						close(causeSelected)
						<-returnNow
						return terminalOutcome(t, OutcomeRejected, ""), nil
					})
				if phaseErr != nil {
					return phaseErr
				}
				if _, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeCancelled)); !errors.Is(publishErr, ErrInstanceBlocked) {
					return errors.New("late context cancellation rewrote Rejected cause")
				}
				_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeRejected))
				return publishErr
			})
		parentResult <- err
	}()
	<-startEntered
	stopResult := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow, success(t))
		stopResult <- executionResult{admission: admission, err: err}
	}()
	waitForCommandRecord(t, boundary, stopScope, "stop")
	close(signalNow)
	if got := <-stopResult; got.err != nil {
		t.Fatal(got.err)
	}
	<-causeSelected
	cancel()
	close(returnNow)
	if err := <-parentResult; err != nil {
		t.Fatal(err)
	}
}

func TestPendingStopCancellationFirstLeavesOwnerClaimNoPending(t *testing.T) {
	boundary := newTestBoundary(t)
	parentScope := testScope(t, "management", "pending-cancel", OperationReplace)
	stopScope := testScope(t, "management", "pending-cancel", OperationStop)
	startEntered := make(chan struct{})
	ownerNow := make(chan struct{})
	ownerReturned := make(chan struct{})
	publishNow := make(chan struct{})
	parentResult := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", replaceIntent(t, 34), allow,
			func(execution *ParentExecution) error {
				_, _, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
					func(start *StartTargetExecution) (TerminalOutcome, error) {
						close(startEntered)
						<-ownerNow
						prevented, signalErr := start.OwnerClaimed("attempt-34")
						if prevented || signalErr != nil {
							return TerminalOutcome{}, errors.New("cancelled Stop remained pending")
						}
						close(ownerReturned)
						<-publishNow
						return terminalOutcome(t, OutcomeSucceeded, "attempt-34"), nil
					})
				if phaseErr != nil {
					return phaseErr
				}
				_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeSucceeded))
				return publishErr
			})
		parentResult <- err
	}()
	<-startEntered
	ctx, cancel := context.WithCancel(context.Background())
	var stopCalls atomic.Int32
	stopResult := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(ctx, stopScope, "stop", NewStopIntent(), allow,
			func() (TerminalOutcome, error) {
				stopCalls.Add(1)
				return terminalOutcome(t, OutcomeSucceeded, "unexpected"), nil
			})
		stopResult <- executionResult{admission: admission, err: err}
	}()
	waitForCommandRecord(t, boundary, stopScope, "stop")
	cancel()
	got := <-stopResult
	if !errors.Is(got.err, context.Canceled) || got.admission.Record().State() != CommandStateTerminal {
		t.Fatalf("cancelled Stop = %#v, %v", got.admission, got.err)
	}
	replay, replayErr := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow,
		func() (TerminalOutcome, error) {
			stopCalls.Add(1)
			return TerminalOutcome{}, errors.New("cancelled Stop replay invoked callback")
		})
	if replayErr != nil || replay.Kind() != AdmissionReplay {
		t.Fatalf("cancelled Stop replay = %#v, %v", replay, replayErr)
	}
	if _, err := boundary.Execute(context.Background(), stopScope, "second-pre-start", NewStopIntent(), allow, success(t)); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("cancelled Stop slot admitted second pre-Start Stop: %v", err)
	}
	close(ownerNow)
	<-ownerReturned
	if _, err := boundary.Execute(context.Background(), stopScope, "second-post-start", NewStopIntent(), allow, success(t)); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("cancelled Stop slot admitted second post-Start Stop: %v", err)
	}
	close(publishNow)
	if err := <-parentResult; err != nil {
		t.Fatal(err)
	}
	if stopCalls.Load() != 0 {
		t.Fatalf("cancelled pending Stop invoked lifecycle %d times", stopCalls.Load())
	}
}

func TestPrePhaseStopCancellationIrreversiblyWinsCancelled(t *testing.T) {
	boundary := newTestBoundary(t)
	parentScope := testScope(t, "management", "pre-phase-cancel", OperationReplace)
	stopScope := testScope(t, "management", "pre-phase-cancel", OperationStop)
	parentEntered := make(chan struct{})
	continueNow := make(chan struct{})
	parentResult := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", replaceIntent(t, 51), allow,
			func(execution *ParentExecution) error {
				close(parentEntered)
				<-continueNow
				phase, prevented, continueErr := execution.ContinueOrExecuteStartTarget(
					context.Background(),
					func(*StartTargetExecution) (TerminalOutcome, error) {
						return TerminalOutcome{}, errors.New("pre-phase cancellation invoked StartTarget")
					},
				)
				if !errors.Is(continueErr, context.Canceled) || !prevented || phase != (PhaseAdmission{}) {
					return errors.New("pre-phase cancellation did not win Continue gate")
				}
				for _, category := range []ParentOutcomeCategory{
					ParentOutcomeSucceeded, ParentOutcomeSatisfied, ParentOutcomeStopped,
					ParentOutcomeRejected, ParentOutcomeFailed,
				} {
					if _, publishErr := execution.PublishTerminal(parentOutcome(t, category)); !errors.Is(publishErr, ErrInstanceBlocked) {
						return errors.New("pre-phase cancellation accepted non-Cancelled parent")
					}
				}
				_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeCancelled))
				return publishErr
			})
		parentResult <- err
	}()
	<-parentEntered
	ctx, cancel := context.WithCancel(context.Background())
	stopResult := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(ctx, stopScope, "stop", NewStopIntent(), allow, success(t))
		stopResult <- executionResult{admission: admission, err: err}
	}()
	waitForCommandRecord(t, boundary, stopScope, "stop")
	cancel()
	if got := <-stopResult; !errors.Is(got.err, context.Canceled) ||
		got.admission.Record().State() != CommandStateTerminal {
		t.Fatalf("pre-phase cancelled Stop = %#v, %v", got.admission, got.err)
	}
	if _, err := boundary.Execute(context.Background(), stopScope, "second", NewStopIntent(), allow, success(t)); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("pre-phase cancellation admitted second Stop: %v", err)
	}
	close(continueNow)
	if err := <-parentResult; err != nil {
		t.Fatal(err)
	}
	ledger := boundary.storage.ledger(parentScope.instanceScope())
	ledger.mu.Lock()
	defer ledger.mu.Unlock()
	startID, _ := newPhaseIdentity(commandIdentity{scope: parentScope, key: "parent"}, PhaseStartTarget)
	if ledger.phases[startID] != nil {
		t.Fatal("pre-phase cancellation created StartTarget")
	}
}

func TestReconstructionWakesPendingRendezvousFailClosed(t *testing.T) {
	storage := NewMemoryStorage()
	boundary := newBoundary(t, storage)
	parentScope := testScope(t, "management", "rendezvous-reconstruct", OperationRollback)
	stopScope := testScope(t, "management", "rendezvous-reconstruct", OperationStop)
	startEntered := make(chan struct{})
	ownerNow := make(chan struct{})
	parentResult := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", rollbackIntent(t, 35), allow,
			func(execution *ParentExecution) error {
				_, _, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
					func(start *StartTargetExecution) (TerminalOutcome, error) {
						close(startEntered)
						<-ownerNow
						_, signalErr := start.OwnerClaimed("attempt-35")
						return TerminalOutcome{}, signalErr
					})
				return phaseErr
			})
		parentResult <- err
	}()
	<-startEntered
	stopResult := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow, success(t))
		stopResult <- executionResult{admission: admission, err: err}
	}()
	waitForCommandRecord(t, boundary, stopScope, "stop")
	_ = newBoundary(t, storage)
	if got := <-stopResult; !errors.Is(got.err, ErrBoundaryExpired) {
		t.Fatalf("reconstructed pending Stop error = %v", got.err)
	}
	close(ownerNow)
	if err := <-parentResult; !errors.Is(err, ErrIndeterminateExecution) {
		t.Fatalf("reconstructed StartTarget error = %v", err)
	}
}

func TestPendingStopGoexitUnblocksStartAsBlocked(t *testing.T) {
	boundary := newTestBoundary(t)
	parentScope := testScope(t, "management", "pending-goexit", OperationReplace)
	stopScope := testScope(t, "management", "pending-goexit", OperationStop)
	startEntered := make(chan struct{})
	ownerNow := make(chan struct{})
	parentResult := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", replaceIntent(t, 36), allow,
			func(execution *ParentExecution) error {
				_, _, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
					func(start *StartTargetExecution) (TerminalOutcome, error) {
						close(startEntered)
						<-ownerNow
						_, signalErr := start.OwnerClaimed("attempt-36")
						return TerminalOutcome{}, signalErr
					})
				return phaseErr
			})
		parentResult <- err
	}()
	<-startEntered
	stopDone := make(chan struct{})
	go func() {
		defer close(stopDone)
		_, _ = boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow,
			func() (TerminalOutcome, error) {
				runtime.Goexit()
				return TerminalOutcome{}, nil
			})
	}()
	waitForCommandRecord(t, boundary, stopScope, "stop")
	close(ownerNow)
	select {
	case <-stopDone:
	case <-time.After(2 * time.Second):
		t.Fatal("pending Stop Goexit did not terminate")
	}
	if err := <-parentResult; !errors.Is(err, ErrIndeterminateExecution) {
		t.Fatalf("pending Stop Goexit parent error = %v", err)
	}
}

func TestStartTargetSignalExpiresAfterCallbackReturn(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "management", "signal-expiry", OperationReplace)
	var retained *StartTargetExecution
	_, err := boundary.ExecuteParent(context.Background(), scope, "parent", replaceIntent(t, 37), allow,
		func(execution *ParentExecution) error {
			_, _, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
				func(start *StartTargetExecution) (TerminalOutcome, error) {
					retained = start
					if signalErr := start.StartNoClaim(StartNoClaimRejected); signalErr != nil {
						return TerminalOutcome{}, signalErr
					}
					return terminalOutcome(t, OutcomeRejected, ""), nil
				})
			if phaseErr != nil {
				return phaseErr
			}
			_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeRejected))
			return publishErr
		})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := retained.OwnerClaimed("late-attempt"); !errors.Is(err, ErrInvalidSubmission) {
		t.Fatalf("retained signal error = %v", err)
	}
}

func TestOwnerClaimedFirstInvokesStopDespiteLaterCancellation(t *testing.T) {
	boundary := newTestBoundary(t)
	parentScope := testScope(t, "management", "owner-first", OperationReplace)
	stopScope := testScope(t, "management", "owner-first", OperationStop)
	startEntered := make(chan struct{})
	ownerNow := make(chan struct{})
	stopEntered := make(chan struct{})
	releaseStop := make(chan struct{})
	parentResult := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", replaceIntent(t, 38), allow,
			func(execution *ParentExecution) error {
				_, _, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
					func(start *StartTargetExecution) (TerminalOutcome, error) {
						close(startEntered)
						<-ownerNow
						stopped, signalErr := start.OwnerClaimed("attempt-38")
						if signalErr != nil || !stopped {
							return TerminalOutcome{}, errors.New("OwnerClaimed-first Stop did not converge")
						}
						return terminalOutcome(t, OutcomeRejected, "attempt-38"), nil
					})
				if phaseErr != nil {
					return phaseErr
				}
				_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeStopped))
				return publishErr
			})
		parentResult <- err
	}()
	<-startEntered
	ctx, cancel := context.WithCancel(context.Background())
	stopResult := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(ctx, stopScope, "stop", NewStopIntent(), allow,
			func() (TerminalOutcome, error) {
				close(stopEntered)
				<-releaseStop
				return terminalOutcome(t, OutcomeSucceeded, "attempt-38"), nil
			})
		stopResult <- executionResult{admission: admission, err: err}
	}()
	waitForCommandRecord(t, boundary, stopScope, "stop")
	close(ownerNow)
	<-stopEntered
	cancel()
	close(releaseStop)
	if got := <-stopResult; got.err != nil || got.admission.Record().State() != CommandStateTerminal {
		t.Fatalf("OwnerClaimed-first Stop = %#v, %v", got.admission, got.err)
	}
	if err := <-parentResult; err != nil {
		t.Fatal(err)
	}
}

func TestNonSuccessfulStopOutcomeBlocksStartDespiteTerminalStopRecord(t *testing.T) {
	for _, category := range []OutcomeCategory{OutcomeFailed, OutcomeRejected} {
		t.Run(string(category), func(t *testing.T) {
			boundary := newTestBoundary(t)
			instance := "stop-" + string(category)
			parentScope := testScope(t, "management", instance, OperationRollback)
			stopScope := testScope(t, "management", instance, OperationStop)
			startEntered := make(chan struct{})
			ownerNow := make(chan struct{})
			parentResult := make(chan error, 1)
			go func() {
				admission, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", rollbackIntent(t, 39), allow,
					func(execution *ParentExecution) error {
						_, _, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
							func(start *StartTargetExecution) (TerminalOutcome, error) {
								close(startEntered)
								<-ownerNow
								_, signalErr := start.OwnerClaimed("attempt-39")
								return TerminalOutcome{}, signalErr
							})
						return phaseErr
					})
				if !errors.Is(err, ErrIndeterminateExecution) || admission.Record().State() != CommandStateClaimed {
					parentResult <- errors.New("non-successful Stop did not leave parent unresolved")
					return
				}
				parentResult <- nil
			}()
			<-startEntered
			stopResult := make(chan executionResult, 1)
			go func() {
				admission, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow,
					func() (TerminalOutcome, error) {
						return terminalOutcome(t, category, "attempt-39"), nil
					})
				stopResult <- executionResult{admission: admission, err: err}
			}()
			waitForCommandRecord(t, boundary, stopScope, "stop")
			close(ownerNow)
			if got := <-stopResult; got.err != nil || got.admission.Record().State() != CommandStateTerminal {
				t.Fatalf("terminal Stop record = %#v, %v", got.admission, got.err)
			}
			if err := <-parentResult; err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestContinueCancellationCreatesNoPhaseAndOnlyCancelledParent(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "management", "continue-cancel", OperationReplace)
	admission, err := boundary.ExecuteParent(context.Background(), scope, "parent", replaceIntent(t, 40), allow,
		func(execution *ParentExecution) error {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			phase, prevented, continueErr := execution.ContinueOrExecuteStartTarget(ctx,
				func(*StartTargetExecution) (TerminalOutcome, error) {
					t.Fatal("cancelled Continue invoked StartTarget")
					return TerminalOutcome{}, nil
				})
			if !errors.Is(continueErr, context.Canceled) || !prevented || phase != (PhaseAdmission{}) {
				return errors.New("Continue cancellation did not win before phase claim")
			}
			if _, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeStopped)); !errors.Is(publishErr, ErrInstanceBlocked) {
				return errors.New("cancelled Continue allowed non-cancelled parent outcome")
			}
			if _, preventedAgain, replayErr := execution.ContinueOrExecuteStartTarget(
				context.Background(),
				func(*StartTargetExecution) (TerminalOutcome, error) {
					t.Fatal("cancelled Continue replay invoked StartTarget")
					return TerminalOutcome{}, nil
				},
			); !errors.Is(replayErr, context.Canceled) || !preventedAgain {
				return errors.New("cancelled Continue reopened its gate")
			}
			_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeCancelled))
			return publishErr
		})
	if err != nil || admission.Record().State() != CommandStateTerminal {
		t.Fatalf("cancelled Continue parent = %#v, %v", admission, err)
	}
	ledger := boundary.storage.ledger(scope.instanceScope())
	ledger.mu.Lock()
	defer ledger.mu.Unlock()
	startID, _ := newPhaseIdentity(commandIdentity{scope: scope, key: "parent"}, PhaseStartTarget)
	if ledger.phases[startID] != nil {
		t.Fatal("cancelled Continue created StartTarget phase")
	}
}

func TestContinueCancellationWithPendingStopStillPermitsOnlyCancelledParent(t *testing.T) {
	boundary := newTestBoundary(t)
	parentScope := testScope(t, "management", "continue-cancel-pending", OperationReplace)
	stopScope := testScope(t, "management", "continue-cancel-pending", OperationStop)
	parentEntered := make(chan struct{})
	continueNow := make(chan struct{})
	parentResult := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", replaceIntent(t, 54), allow,
			func(execution *ParentExecution) error {
				close(parentEntered)
				<-continueNow
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				_, prevented, continueErr := execution.ContinueOrExecuteStartTarget(ctx,
					func(*StartTargetExecution) (TerminalOutcome, error) {
						return TerminalOutcome{}, errors.New("cancelled Continue invoked Start")
					})
				if !errors.Is(continueErr, context.Canceled) || !prevented {
					return errors.New("Continue cancellation did not win")
				}
				if _, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeStopped)); !errors.Is(publishErr, ErrInstanceBlocked) {
					return errors.New("Continue cancellation accepted Stopped")
				}
				_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeCancelled))
				return publishErr
			})
		parentResult <- err
	}()
	<-parentEntered
	stopResult := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow, success(t))
		stopResult <- executionResult{admission: admission, err: err}
	}()
	waitForCommandRecord(t, boundary, stopScope, "stop")
	close(continueNow)
	if got := <-stopResult; got.err != nil {
		t.Fatal(got.err)
	}
	if err := <-parentResult; err != nil {
		t.Fatal(err)
	}
}

func TestPendingStopCancellationRaceHasOnlyLegalWinners(t *testing.T) {
	for i := 0; i < 20; i++ {
		boundary := newTestBoundary(t)
		instance := runtimeconfigload.RuntimeInstanceID("cancel-race-" + string(rune('a'+i)))
		parentScope := testScope(t, "management", string(instance), OperationReplace)
		stopScope := testScope(t, "management", string(instance), OperationStop)
		startEntered := make(chan struct{})
		race := make(chan struct{})
		var stoppedWinner atomic.Bool
		parentResult := make(chan error, 1)
		go func() {
			_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", replaceIntent(t, 41), allow,
				func(execution *ParentExecution) error {
					_, _, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
						func(start *StartTargetExecution) (TerminalOutcome, error) {
							close(startEntered)
							<-race
							stopped, signalErr := start.OwnerClaimed("attempt-41")
							if signalErr != nil {
								return TerminalOutcome{}, signalErr
							}
							if stopped {
								stoppedWinner.Store(true)
								return terminalOutcome(t, OutcomeRejected, "attempt-41"), nil
							}
							return terminalOutcome(t, OutcomeSucceeded, "attempt-41"), nil
						})
					if phaseErr != nil {
						return phaseErr
					}
					parentCategory := ParentOutcomeSucceeded
					if stoppedWinner.Load() {
						parentCategory = ParentOutcomeStopped
					}
					_, publishErr := execution.PublishTerminal(parentOutcome(t, parentCategory))
					return publishErr
				})
			parentResult <- err
		}()
		<-startEntered
		ctx, cancel := context.WithCancel(context.Background())
		var stopCalls atomic.Int32
		stopResult := make(chan executionResult, 1)
		go func() {
			admission, err := boundary.Execute(ctx, stopScope, "stop", NewStopIntent(), allow,
				func() (TerminalOutcome, error) {
					stopCalls.Add(1)
					return terminalOutcome(t, OutcomeSucceeded, "attempt-41"), nil
				})
			stopResult <- executionResult{admission: admission, err: err}
		}()
		waitForCommandRecord(t, boundary, stopScope, "stop")
		close(race)
		cancel()
		got := <-stopResult
		if got.err == nil {
			if stopCalls.Load() != 1 {
				t.Fatalf("owner-first winner invoked %d Stop callbacks", stopCalls.Load())
			}
		} else if errors.Is(got.err, context.Canceled) {
			if stopCalls.Load() != 0 {
				t.Fatalf("cancellation-first winner invoked %d Stop callbacks", stopCalls.Load())
			}
		} else {
			t.Fatalf("illegal cancellation race outcome: %v", got.err)
		}
		if err := <-parentResult; err != nil {
			t.Fatal(err)
		}
	}
}

func TestPendingRendezvousDoesNotBlockDifferentInstance(t *testing.T) {
	boundary := newTestBoundary(t)
	blockedParent := testScope(t, "management", "blocked-instance", OperationReplace)
	blockedStop := testScope(t, "management", "blocked-instance", OperationStop)
	otherParent := testScope(t, "management", "progress-instance", OperationReplace)
	startEntered := make(chan struct{})
	ownerNow := make(chan struct{})
	blockedResult := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), blockedParent, "parent", replaceIntent(t, 42), allow,
			func(execution *ParentExecution) error {
				_, _, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
					func(start *StartTargetExecution) (TerminalOutcome, error) {
						close(startEntered)
						<-ownerNow
						stopped, signalErr := start.OwnerClaimed("attempt-42")
						if signalErr != nil || !stopped {
							return TerminalOutcome{}, errors.New("pending Stop did not converge")
						}
						return terminalOutcome(t, OutcomeRejected, "attempt-42"), nil
					})
				if phaseErr != nil {
					return phaseErr
				}
				_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeStopped))
				return publishErr
			})
		blockedResult <- err
	}()
	<-startEntered
	stopResult := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(context.Background(), blockedStop, "stop", NewStopIntent(), allow,
			func() (TerminalOutcome, error) {
				return terminalOutcome(t, OutcomeSucceeded, "attempt-42"), nil
			})
		stopResult <- executionResult{admission: admission, err: err}
	}()
	waitForCommandRecord(t, boundary, blockedStop, "stop")

	other, err := boundary.ExecuteParent(context.Background(), otherParent, "parent", replaceIntent(t, 43), allow,
		func(execution *ParentExecution) error {
			_, phaseErr := executeStartTarget(execution, success(t))
			if phaseErr != nil {
				return phaseErr
			}
			_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeSucceeded))
			return publishErr
		})
	if err != nil || other.Record().State() != CommandStateTerminal {
		t.Fatalf("different Instance did not progress: %#v, %v", other, err)
	}
	close(ownerNow)
	if got := <-stopResult; got.err != nil {
		t.Fatal(got.err)
	}
	if err := <-blockedResult; err != nil {
		t.Fatal(err)
	}
}

func TestStopFirstParentTerminalCategoryGate(t *testing.T) {
	for _, category := range []ParentOutcomeCategory{
		ParentOutcomeSucceeded, ParentOutcomeSatisfied, ParentOutcomeStopped,
		ParentOutcomeCancelled, ParentOutcomeRejected, ParentOutcomeFailed,
	} {
		t.Run(string(category), func(t *testing.T) {
			boundary := newTestBoundary(t)
			parentScope := testScope(t, "management", "stop-first-"+string(category), OperationReplace)
			stopScope := testScope(t, "management", "stop-first-"+string(category), OperationStop)
			parentEntered := make(chan struct{})
			continueNow := make(chan struct{})
			parentResult := make(chan error, 1)
			go func() {
				_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", replaceIntent(t, 44), allow,
					func(execution *ParentExecution) error {
						close(parentEntered)
						<-continueNow
						_, prevented, continueErr := execution.ContinueOrExecuteStartTarget(
							context.Background(),
							func(*StartTargetExecution) (TerminalOutcome, error) {
								return TerminalOutcome{}, errors.New("Stop-first callback invoked")
							},
						)
						if continueErr != nil || !prevented {
							return errors.New("Stop-first gate did not prevent phase")
						}
						_, publishErr := execution.PublishTerminal(parentOutcome(t, category))
						allowed := category == ParentOutcomeStopped
						if allowed {
							return publishErr
						}
						if !errors.Is(publishErr, ErrInstanceBlocked) {
							return errors.New("Stop-first accepted illegal parent category")
						}
						_, publishErr = execution.PublishTerminal(parentOutcome(t, ParentOutcomeStopped))
						return publishErr
					})
				parentResult <- err
			}()
			<-parentEntered
			stopResult := make(chan executionResult, 1)
			go func() {
				admission, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow, success(t))
				stopResult <- executionResult{admission: admission, err: err}
			}()
			waitForCommandRecord(t, boundary, stopScope, "stop")
			close(continueNow)
			if got := <-stopResult; got.err != nil {
				t.Fatal(got.err)
			}
			if err := <-parentResult; err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestStartNoClaimOutcomeCompatibility(t *testing.T) {
	for _, test := range []struct {
		name     string
		cause    StartNoClaimCause
		category OutcomeCategory
		attempt  runtimeconfigload.LaunchAttemptID
		allowed  bool
	}{
		{name: "cancelled-rejected-empty", cause: StartNoClaimCancelled, category: OutcomeRejected, allowed: true},
		{name: "rejected-empty", cause: StartNoClaimRejected, category: OutcomeRejected, allowed: true},
		{name: "failed-empty", cause: StartNoClaimFailed, category: OutcomeFailed, allowed: true},
		{name: "rejected-cause-failed-outcome", cause: StartNoClaimRejected, category: OutcomeFailed},
		{name: "failed-cause-rejected-outcome", cause: StartNoClaimFailed, category: OutcomeRejected},
		{name: "succeeded-empty", cause: StartNoClaimRejected, category: OutcomeSucceeded},
		{name: "succeeded-attempt", cause: StartNoClaimRejected, category: OutcomeSucceeded, attempt: "attempt"},
		{name: "rejected-attempt", cause: StartNoClaimRejected, category: OutcomeRejected, attempt: "attempt"},
		{name: "failed-attempt", cause: StartNoClaimFailed, category: OutcomeFailed, attempt: "attempt"},
	} {
		t.Run(test.name, func(t *testing.T) {
			boundary := newTestBoundary(t)
			scope := testScope(t, "management", "no-claim-"+test.name, OperationReplace)
			admission, err := boundary.ExecuteParent(context.Background(), scope, "parent", replaceIntent(t, 45), allow,
				func(execution *ParentExecution) error {
					phase, _, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
						func(start *StartTargetExecution) (TerminalOutcome, error) {
							if signalErr := start.StartNoClaim(test.cause); signalErr != nil {
								return TerminalOutcome{}, signalErr
							}
							return terminalOutcome(t, test.category, test.attempt), nil
						})
					if test.allowed {
						if phaseErr != nil || phase.Record().State() != CommandStateTerminal {
							return errors.New("valid StartNoClaim outcome was rejected")
						}
						parentCategory := ParentOutcomeRejected
						switch test.cause {
						case StartNoClaimCancelled:
							parentCategory = ParentOutcomeCancelled
						case StartNoClaimFailed:
							parentCategory = ParentOutcomeFailed
						}
						_, publishErr := execution.PublishTerminal(parentOutcome(t, parentCategory))
						return publishErr
					}
					if !errors.Is(phaseErr, ErrIndeterminateExecution) || phase.Record().State() != CommandStateClaimed {
						return errors.New("invalid StartNoClaim outcome was accepted")
					}
					return phaseErr
				})
			if test.allowed {
				if err != nil || admission.Record().State() != CommandStateTerminal {
					t.Fatalf("allowed StartNoClaim = %#v, %v", admission, err)
				}
			} else if !errors.Is(err, ErrIndeterminateExecution) || admission.Record().State() != CommandStateClaimed {
				t.Fatalf("rejected StartNoClaim = %#v, %v", admission, err)
			}
		})
	}
}

func TestManyDistinctStopsConsumeOneStartTargetSlot(t *testing.T) {
	boundary := newTestBoundary(t)
	parentScope := testScope(t, "management", "many-stops", OperationRollback)
	stopScope := testScope(t, "management", "many-stops", OperationStop)
	startEntered := make(chan struct{})
	ownerNow := make(chan struct{})
	parentResult := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", rollbackIntent(t, 46), allow,
			func(execution *ParentExecution) error {
				_, _, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
					func(start *StartTargetExecution) (TerminalOutcome, error) {
						close(startEntered)
						<-ownerNow
						stopped, signalErr := start.OwnerClaimed("attempt-46")
						if signalErr != nil || !stopped {
							return TerminalOutcome{}, errors.New("winning Stop did not converge")
						}
						return terminalOutcome(t, OutcomeRejected, "attempt-46"), nil
					})
				if phaseErr != nil {
					return phaseErr
				}
				_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeStopped))
				return publishErr
			})
		parentResult <- err
	}()
	<-startEntered
	const stops = 32
	start := make(chan struct{})
	results := make(chan executionResult, stops)
	var callbacks atomic.Int32
	for i := 0; i < stops; i++ {
		key := CommandKey("stop-" + string(rune('a'+i)))
		go func() {
			<-start
			admission, err := boundary.Execute(context.Background(), stopScope, key, NewStopIntent(), allow,
				func() (TerminalOutcome, error) {
					callbacks.Add(1)
					return terminalOutcome(t, OutcomeSucceeded, "attempt-46"), nil
				})
			results <- executionResult{admission: admission, err: err}
		}()
	}
	close(start)
	blocked := 0
	for blocked < stops-1 {
		select {
		case got := <-results:
			if !errors.Is(got.err, ErrInstanceBlocked) {
				t.Fatalf("non-winning Stop = %#v, %v", got.admission, got.err)
			}
			blocked++
		case <-time.After(2 * time.Second):
			t.Fatal("distinct Stop contenders did not linearize")
		}
	}
	close(ownerNow)
	winner := <-results
	if winner.err != nil || winner.admission.Kind() != AdmissionClaimed || callbacks.Load() != 1 {
		t.Fatalf("winning Stop = %#v, %v, callbacks=%d", winner.admission, winner.err, callbacks.Load())
	}
	if err := <-parentResult; err != nil {
		t.Fatal(err)
	}
}

func TestPendingStopCallbackFailuresWakeStartBlocked(t *testing.T) {
	for _, test := range []struct {
		name   string
		invoke func() (TerminalOutcome, error)
	}{
		{name: "error", invoke: func() (TerminalOutcome, error) {
			return TerminalOutcome{}, errors.New("stop callback error")
		}},
		{name: "panic", invoke: func() (TerminalOutcome, error) { panic("stop callback panic") }},
		{name: "invalid", invoke: func() (TerminalOutcome, error) { return TerminalOutcome{}, nil }},
	} {
		t.Run(test.name, func(t *testing.T) {
			boundary := newTestBoundary(t)
			instance := "stop-callback-" + test.name
			parentScope := testScope(t, "management", instance, OperationReplace)
			stopScope := testScope(t, "management", instance, OperationStop)
			startEntered := make(chan struct{})
			ownerNow := make(chan struct{})
			parentResult := make(chan error, 1)
			go func() {
				_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", replaceIntent(t, 47), allow,
					func(execution *ParentExecution) error {
						_, _, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
							func(start *StartTargetExecution) (TerminalOutcome, error) {
								close(startEntered)
								<-ownerNow
								_, signalErr := start.OwnerClaimed("attempt-47")
								return TerminalOutcome{}, signalErr
							})
						return phaseErr
					})
				parentResult <- err
			}()
			<-startEntered
			stopResult := make(chan executionResult, 1)
			go func() {
				admission, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow, test.invoke)
				stopResult <- executionResult{admission: admission, err: err}
			}()
			waitForCommandRecord(t, boundary, stopScope, "stop")
			close(ownerNow)
			if got := <-stopResult; !errors.Is(got.err, ErrIndeterminateExecution) ||
				got.admission.Record().State() != CommandStateClaimed {
				t.Fatalf("Stop callback failure = %#v, %v", got.admission, got.err)
			}
			if err := <-parentResult; !errors.Is(err, ErrIndeterminateExecution) {
				t.Fatalf("Start peer did not wake Blocked: %v", err)
			}
			other := testScope(t, "management", instance, OperationRollback)
			if _, err := boundary.ExecuteParent(context.Background(), other, "other", rollbackIntent(t, 48), allow,
				func(*ParentExecution) error { return nil }); !errors.Is(err, ErrInstanceBlocked) {
				t.Fatalf("failure reopened unresolved barrier: %v", err)
			}
		})
	}
}

func TestStartCallbackFailuresWakePendingStopBlocked(t *testing.T) {
	for _, test := range []struct {
		name   string
		invoke func() (TerminalOutcome, error)
	}{
		{name: "error", invoke: func() (TerminalOutcome, error) {
			return TerminalOutcome{}, errors.New("start callback error")
		}},
		{name: "panic", invoke: func() (TerminalOutcome, error) { panic("start callback panic") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			boundary := newTestBoundary(t)
			instance := "start-callback-" + test.name
			parentScope := testScope(t, "management", instance, OperationRollback)
			stopScope := testScope(t, "management", instance, OperationStop)
			startEntered := make(chan struct{})
			failNow := make(chan struct{})
			parentResult := make(chan error, 1)
			go func() {
				_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", rollbackIntent(t, 49), allow,
					func(execution *ParentExecution) error {
						_, _, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
							func(*StartTargetExecution) (TerminalOutcome, error) {
								close(startEntered)
								<-failNow
								return test.invoke()
							})
						return phaseErr
					})
				parentResult <- err
			}()
			<-startEntered
			stopResult := make(chan executionResult, 1)
			go func() {
				admission, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow, success(t))
				stopResult <- executionResult{admission: admission, err: err}
			}()
			waitForCommandRecord(t, boundary, stopScope, "stop")
			close(failNow)
			if err := <-parentResult; !errors.Is(err, ErrIndeterminateExecution) {
				t.Fatalf("Start callback failure = %v", err)
			}
			if got := <-stopResult; !errors.Is(got.err, ErrIndeterminateExecution) ||
				got.admission.Record().State() != CommandStateClaimed {
				t.Fatalf("pending Stop did not wake Blocked: %#v, %v", got.admission, got.err)
			}
		})
	}
}

func TestStartCallbackGoexitWakesPendingStopBlocked(t *testing.T) {
	boundary := newTestBoundary(t)
	parentScope := testScope(t, "management", "start-goexit-pending", OperationReplace)
	stopScope := testScope(t, "management", "start-goexit-pending", OperationStop)
	startEntered := make(chan struct{})
	goexitNow := make(chan struct{})
	parentDone := make(chan struct{})
	go func() {
		defer close(parentDone)
		_, _ = boundary.ExecuteParent(context.Background(), parentScope, "parent", replaceIntent(t, 50), allow,
			func(execution *ParentExecution) error {
				_, _, _ = execution.ContinueOrExecuteStartTarget(context.Background(),
					func(*StartTargetExecution) (TerminalOutcome, error) {
						close(startEntered)
						<-goexitNow
						runtime.Goexit()
						return TerminalOutcome{}, nil
					})
				return nil
			})
	}()
	<-startEntered
	stopResult := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow, success(t))
		stopResult <- executionResult{admission: admission, err: err}
	}()
	waitForCommandRecord(t, boundary, stopScope, "stop")
	close(goexitNow)
	<-parentDone
	if got := <-stopResult; !errors.Is(got.err, ErrIndeterminateExecution) ||
		got.admission.Record().State() != CommandStateClaimed {
		t.Fatalf("Start Goexit pending Stop = %#v, %v", got.admission, got.err)
	}
}

func TestSuccessfulRendezvousTerminalReplayHasZeroCallbacksAndExactFacts(t *testing.T) {
	boundary := newTestBoundary(t)
	parentScope := testScope(t, "management", "terminal-rendezvous-replay", OperationRollback)
	stopScope := testScope(t, "management", "terminal-rendezvous-replay", OperationStop)
	startEntered := make(chan struct{})
	ownerNow := make(chan struct{})
	parentResult := make(chan error, 1)
	var parentCalls atomic.Int32
	var stopCalls atomic.Int32
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", rollbackIntent(t, 52), allow,
			func(execution *ParentExecution) error {
				parentCalls.Add(1)
				_, _, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
					func(start *StartTargetExecution) (TerminalOutcome, error) {
						close(startEntered)
						<-ownerNow
						stopped, signalErr := start.OwnerClaimed("attempt-52")
						if signalErr != nil || !stopped {
							return TerminalOutcome{}, errors.New("terminal Stop did not converge")
						}
						return terminalOutcome(t, OutcomeRejected, "attempt-52"), nil
					})
				if phaseErr != nil {
					return phaseErr
				}
				_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeStopped))
				return publishErr
			})
		parentResult <- err
	}()
	<-startEntered
	stopResult := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow,
			func() (TerminalOutcome, error) {
				stopCalls.Add(1)
				return terminalOutcome(t, OutcomeSucceeded, "attempt-52"), nil
			})
		stopResult <- executionResult{admission: admission, err: err}
	}()
	waitForCommandRecord(t, boundary, stopScope, "stop")
	close(ownerNow)
	stopAdmission := <-stopResult
	if stopAdmission.err != nil {
		t.Fatal(stopAdmission.err)
	}
	if err := <-parentResult; err != nil {
		t.Fatal(err)
	}
	parentReplay, err := boundary.ExecuteParent(context.Background(), parentScope, "parent", rollbackIntent(t, 52), allow,
		func(*ParentExecution) error {
			parentCalls.Add(1)
			return errors.New("parent replay callback invoked")
		})
	if err != nil || parentReplay.Kind() != AdmissionReplay {
		t.Fatalf("parent replay = %#v, %v", parentReplay, err)
	}
	stopReplay, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow,
		func() (TerminalOutcome, error) {
			stopCalls.Add(1)
			return TerminalOutcome{}, errors.New("Stop replay callback invoked")
		})
	if err != nil || stopReplay.Kind() != AdmissionReplay || parentCalls.Load() != 1 || stopCalls.Load() != 1 {
		t.Fatalf("terminal replay mutated work: parent=%d stop=%d err=%v", parentCalls.Load(), stopCalls.Load(), err)
	}
	stopOutcome, ok := stopReplay.Record().Outcome()
	if !ok || stopOutcome.Category() != OutcomeSucceeded || stopOutcome.LaunchAttemptID() != "attempt-52" {
		t.Fatalf("redacted Stop facts = %#v, %v", stopOutcome, ok)
	}
	parentOutcome, ok := parentReplay.Record().Outcome()
	if !ok || parentOutcome.Category() != ParentOutcomeStopped {
		t.Fatalf("redacted parent facts = %#v, %v", parentOutcome, ok)
	}
	ledger := boundary.storage.ledger(parentScope.instanceScope())
	ledger.mu.Lock()
	defer ledger.mu.Unlock()
	phaseID, _ := newPhaseIdentity(commandIdentity{scope: parentScope, key: "parent"}, PhaseStartTarget)
	phase := ledger.phases[phaseID]
	if phase == nil || phase.state != CommandStateTerminal || phase.outcome.category != OutcomeRejected ||
		phase.outcome.launchAttemptID != "attempt-52" {
		t.Fatalf("redacted phase facts = %#v", phase)
	}
}

func TestRetainedSignalBecomesStaleAcrossReconstruction(t *testing.T) {
	storage := NewMemoryStorage()
	boundary := newBoundary(t, storage)
	scope := testScope(t, "management", "retained-stale-signal", OperationReplace)
	startCapability := make(chan *StartTargetExecution, 1)
	release := make(chan struct{})
	parentResult := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteParent(context.Background(), scope, "parent", replaceIntent(t, 53), allow,
			func(execution *ParentExecution) error {
				_, _, phaseErr := execution.ContinueOrExecuteStartTarget(context.Background(),
					func(start *StartTargetExecution) (TerminalOutcome, error) {
						startCapability <- start
						<-release
						return TerminalOutcome{}, errors.New("stale callback exits unresolved")
					})
				return phaseErr
			})
		parentResult <- err
	}()
	retained := <-startCapability
	active := newBoundary(t, storage)
	if _, err := retained.OwnerClaimed("attempt-53"); !errors.Is(err, ErrBoundaryExpired) {
		t.Fatalf("stale active signal error = %v", err)
	}
	close(release)
	if err := <-parentResult; !errors.Is(err, ErrIndeterminateExecution) {
		t.Fatalf("stale parent result = %v", err)
	}
	if _, err := retained.OwnerClaimed("attempt-53"); !errors.Is(err, ErrInvalidSubmission) {
		t.Fatalf("post-return retained signal error = %v", err)
	}
	var callbacks atomic.Int32
	observed, err := active.ExecuteParent(context.Background(), scope, "parent", replaceIntent(t, 53), allow,
		func(*ParentExecution) error { callbacks.Add(1); return nil })
	if err != nil || observed.Kind() != AdmissionInProgress || callbacks.Load() != 0 {
		t.Fatalf("reconstruction restored capability: %#v, %v, callbacks=%d", observed, err, callbacks.Load())
	}
}

func waitForCommandRecord(t *testing.T, boundary *Boundary, scope Scope, key CommandKey) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	identity := commandIdentity{scope: scope, key: key}
	ledger := boundary.storage.ledger(scope.instanceScope())
	for time.Now().Before(deadline) {
		ledger.mu.Lock()
		exists := ledger.records[identity] != nil
		ledger.mu.Unlock()
		if exists {
			return
		}
		runtime.Gosched()
	}
	t.Fatalf("command %q was not admitted", key)
}
