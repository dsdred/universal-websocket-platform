package runtimecommandidempotency

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

func TestTrackedStartManagedParentAdmissionIsAtomicAndReplayHasNoAuthority(t *testing.T) {
	boundary := newTestBoundary(t)
	startScope, releaseStart, startDone := holdTrackedStart(t, boundary, "tracked-parent-atomic")
	parentScope := testScope(t, startScope.Domain(), "tracked-parent-atomic", OperationReplace)
	intent := replaceIntent(t, 42)
	var stopCalls atomic.Int32
	var parentCalls atomic.Int32

	admission, err := boundary.ExecuteManagedParentFromTrackedStart(
		context.Background(), parentScope, "replace-a", intent, allowOrchestration,
		func(execution *TrackedStartManagedParentExecution) error {
			parentCalls.Add(1)
			ledger := boundary.storage.ledger(parentScope.instanceScope())
			parentIdentity := commandIdentity{scope: parentScope, key: "replace-a"}
			stopIdentity, _ := newPhaseIdentity(parentIdentity, PhaseStopOld)
			startIdentity := commandIdentity{scope: startScope, key: "tracked-start"}
			ledger.mu.Lock()
			parent := ledger.parents[parentIdentity]
			stop := ledger.phases[stopIdentity]
			occupant := ledger.stopForStart[startIdentity]
			primitiveStops := 0
			for identity := range ledger.records {
				if identity.scope.operation == OperationStop {
					primitiveStops++
				}
			}
			atomicFacts := parent != nil && stop != nil && stop.identity.ordinal == 0 &&
				occupant.phase != nil && *occupant.phase == stopIdentity && primitiveStops == 0
			ledger.mu.Unlock()
			if !atomicFacts {
				return errors.New("parent, preclaimed StopOld, and exception occupant were not atomic")
			}

			ordinaryCalls := 0
			observed, ordinaryErr := execution.managed.InspectOrExecuteStopOld(func() (TerminalOutcome, error) {
				ordinaryCalls++
				return terminalOutcome(t, OutcomeSucceeded, "old-attempt"), nil
			})
			if ordinaryErr != nil || observed.Kind() != AdmissionInProgress || ordinaryCalls != 0 {
				return errors.New("ordinary phase path consumed preclaimed StopOld")
			}
			if _, prevented, startErr := execution.ContinueOrExecuteManagedStartTarget(
				context.Background(), 7, "generation-a",
				func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
					return TerminalOutcome{}, errors.New("StartTarget ran before StopOld")
				},
			); !errors.Is(startErr, ErrIllegalPhaseOrder) || prevented {
				return errors.New("StartTarget was not gated by preclaimed StopOld")
			}

			stopAdmission, stopErr := execution.ExecutePreclaimedStopOld(func() (TerminalOutcome, error) {
				stopCalls.Add(1)
				return terminalOutcome(t, OutcomeSucceeded, "old-attempt"), nil
			})
			if stopErr != nil || stopAdmission.Kind() != AdmissionClaimed ||
				stopAdmission.Record().State() != CommandStateTerminal {
				return stopErr
			}
			repeat, repeatErr := execution.ExecutePreclaimedStopOld(func() (TerminalOutcome, error) {
				stopCalls.Add(1)
				return TerminalOutcome{}, errors.New("duplicate StopOld")
			})
			if repeatErr != nil || repeat.Kind() != AdmissionReplay {
				return errors.New("repeated preclaimed StopOld did not observe terminal facts")
			}

			phase, prevented, startErr := execution.ContinueOrExecuteManagedStartTarget(
				context.Background(), 7, "generation-a",
				func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
					if gate, gateErr := boundary.ResolveManagedStartEarly(binding, "target-attempt"); gateErr != nil || gate != GateClear {
						return TerminalOutcome{}, gateErr
					}
					if gate, gateErr := boundary.ResolveManagedStartFinal(binding, FinalContinue); gateErr != nil || gate != GateClear {
						return TerminalOutcome{}, gateErr
					}
					return terminalOutcome(t, OutcomeSucceeded, "target-attempt"), nil
				},
			)
			if startErr != nil || prevented || phase.Kind() != AdmissionClaimed {
				return startErr
			}
			_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeSucceeded))
			return publishErr
		},
	)
	if err != nil || admission.Kind() != AdmissionClaimed ||
		admission.Record().State() != CommandStateTerminal {
		t.Fatalf("admission = %#v, %v", admission, err)
	}
	if parentCalls.Load() != 1 || stopCalls.Load() != 1 {
		t.Fatalf("parent calls=%d StopOld calls=%d", parentCalls.Load(), stopCalls.Load())
	}
	stopScope := testScope(t, startScope.Domain(), "tracked-parent-atomic", OperationStop)
	if _, stopErr := boundary.Execute(
		context.Background(), stopScope, "second-stop", NewStopIntent(), allow, success(t),
	); !errors.Is(stopErr, ErrInstanceBlocked) {
		t.Fatalf("second Stop entered before tracked Start completed: %v", stopErr)
	}

	replay, err := boundary.ExecuteManagedParentFromTrackedStart(
		context.Background(), parentScope, "replace-a", intent, allowOrchestration,
		func(*TrackedStartManagedParentExecution) error {
			t.Fatal("replay received tracked-Start capability")
			return nil
		},
	)
	if err != nil || replay.Kind() != AdmissionReplay {
		t.Fatalf("replay = %#v, %v", replay, err)
	}
	close(releaseStart)
	if err := <-startDone; err != nil {
		t.Fatal(err)
	}
}

func TestTrackedStartManagedParentAndIndependentStopHaveOneWinner(t *testing.T) {
	t.Run("parent-first", func(t *testing.T) {
		boundary := newTestBoundary(t)
		startScope, releaseStart, startDone := holdTrackedStart(t, boundary, "parent-first")
		parentScope := testScope(t, startScope.Domain(), "parent-first", OperationReplace)
		entered := make(chan struct{})
		releaseParent := make(chan struct{})
		parentDone := make(chan error, 1)
		go func() {
			_, err := boundary.ExecuteManagedParentFromTrackedStart(
				context.Background(), parentScope, "parent", replaceIntent(t, 2), allowOrchestration,
				func(*TrackedStartManagedParentExecution) error {
					close(entered)
					<-releaseParent
					return errors.New("leave parent unresolved")
				},
			)
			parentDone <- err
		}()
		<-entered
		observed, observedErr := boundary.ExecuteManagedParentFromTrackedStart(
			context.Background(), parentScope, "parent", replaceIntent(t, 2), allowOrchestration,
			func(*TrackedStartManagedParentExecution) error {
				t.Fatal("same-parent observation received authority")
				return nil
			},
		)
		if observedErr != nil || observed.Kind() != AdmissionInProgress {
			t.Fatalf("same-parent observation = %#v, %v", observed, observedErr)
		}
		stopScope := testScope(t, startScope.Domain(), "parent-first", OperationStop)
		if _, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow, success(t)); !errors.Is(err, ErrInstanceBlocked) {
			t.Fatalf("independent Stop bypassed parent winner: %v", err)
		}
		close(releaseParent)
		if err := <-parentDone; !errors.Is(err, ErrIndeterminateExecution) {
			t.Fatalf("parent result = %v", err)
		}
		close(releaseStart)
		if err := <-startDone; err != nil {
			t.Fatal(err)
		}
	})

	t.Run("stop-first", func(t *testing.T) {
		boundary := newTestBoundary(t)
		startScope, releaseStart, startDone := holdTrackedStart(t, boundary, "stop-first")
		stopScope := testScope(t, startScope.Domain(), "stop-first", OperationStop)
		entered := make(chan struct{})
		releaseStop := make(chan struct{})
		stopDone := make(chan error, 1)
		go func() {
			_, err := boundary.Execute(context.Background(), stopScope, "stop", NewStopIntent(), allow,
				func() (TerminalOutcome, error) {
					close(entered)
					<-releaseStop
					return terminalOutcome(t, OutcomeSucceeded, "old-attempt"), nil
				})
			stopDone <- err
		}()
		<-entered
		parentScope := testScope(t, startScope.Domain(), "stop-first", OperationRollback)
		if _, err := boundary.ExecuteManagedParentFromTrackedStart(
			context.Background(), parentScope, "parent", rollbackIntent(t, 3), allowOrchestration,
			func(*TrackedStartManagedParentExecution) error {
				t.Fatal("parent callback ran after Stop won")
				return nil
			},
		); !errors.Is(err, ErrInstanceBlocked) {
			t.Fatalf("parent bypassed Stop winner: %v", err)
		}
		ledger := boundary.storage.ledger(parentScope.instanceScope())
		ledger.mu.Lock()
		parents, phases := len(ledger.parents), len(ledger.phases)
		ledger.mu.Unlock()
		if parents != 0 || phases != 0 {
			t.Fatalf("losing parent mutated facts: parents=%d phases=%d", parents, phases)
		}
		close(releaseStop)
		if err := <-stopDone; err != nil {
			t.Fatal(err)
		}
		close(releaseStart)
		if err := <-startDone; err != nil {
			t.Fatal(err)
		}
	})
}

func TestTrackedStartManagedParentLostCapabilityStaysUnresolved(t *testing.T) {
	for _, tc := range []struct {
		name   string
		invoke func(*TrackedStartManagedParentExecution) error
	}{
		{"return", func(*TrackedStartManagedParentExecution) error { return nil }},
		{"error", func(*TrackedStartManagedParentExecution) error { return errors.New("callback error") }},
		{"panic", func(*TrackedStartManagedParentExecution) error { panic("callback panic") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMemoryStorage()
			boundary := newBoundary(t, storage)
			startScope, releaseStart, startDone := holdTrackedStart(t, boundary, "lost-"+tc.name)
			parentScope := testScope(t, startScope.Domain(), "lost-"+tc.name, OperationReplace)
			claimed, err := boundary.ExecuteManagedParentFromTrackedStart(
				context.Background(), parentScope, "parent", replaceIntent(t, 4), allowOrchestration, tc.invoke,
			)
			if !errors.Is(err, ErrIndeterminateExecution) || claimed.Record().State() != CommandStateClaimed {
				t.Fatalf("claim = %#v, %v", claimed, err)
			}
			reconstructed := newBoundary(t, storage)
			observed, err := reconstructed.ExecuteManagedParentFromTrackedStart(
				context.Background(), parentScope, "parent", replaceIntent(t, 4), allowOrchestration,
				func(*TrackedStartManagedParentExecution) error {
					t.Fatal("reconstruction restored capability")
					return nil
				},
			)
			if err != nil || observed.Kind() != AdmissionInProgress {
				t.Fatalf("observation = %#v, %v", observed, err)
			}
			close(releaseStart)
			if err := <-startDone; !errors.Is(err, ErrBoundaryExpired) {
				t.Fatalf("stale tracked Start = %v", err)
			}
		})
	}
}

func TestTrackedStartManagedParentGoexitExpiresAuthority(t *testing.T) {
	boundary := newTestBoundary(t)
	startScope, releaseStart, startDone := holdTrackedStart(t, boundary, "goexit")
	parentScope := testScope(t, startScope.Domain(), "goexit", OperationRollback)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = boundary.ExecuteManagedParentFromTrackedStart(
			context.Background(), parentScope, "parent", rollbackIntent(t, 5), allowOrchestration,
			func(*TrackedStartManagedParentExecution) error {
				runtime.Goexit()
				return nil
			},
		)
	}()
	<-done
	observed, err := boundary.ExecuteManagedParentFromTrackedStart(
		context.Background(), parentScope, "parent", rollbackIntent(t, 5), allowOrchestration,
		func(*TrackedStartManagedParentExecution) error {
			t.Fatal("Goexit restored authority")
			return nil
		},
	)
	if err != nil || observed.Kind() != AdmissionInProgress {
		t.Fatalf("observation = %#v, %v", observed, err)
	}
	close(releaseStart)
	if err := <-startDone; err != nil {
		t.Fatal(err)
	}
}

func TestTrackedStartPreclaimedStopFailureExpiresAuthority(t *testing.T) {
	for _, tc := range []struct {
		name   string
		invoke func() (TerminalOutcome, error)
	}{
		{"error", func() (TerminalOutcome, error) { return TerminalOutcome{}, errors.New("stop error") }},
		{"panic", func() (TerminalOutcome, error) { panic("stop panic") }},
		{"invalid", func() (TerminalOutcome, error) { return TerminalOutcome{}, nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			boundary := newTestBoundary(t)
			startScope, releaseStart, startDone := holdTrackedStart(t, boundary, "stop-failure-"+tc.name)
			parentScope := testScope(t, startScope.Domain(), "stop-failure-"+tc.name, OperationReplace)
			claimed, err := boundary.ExecuteManagedParentFromTrackedStart(
				context.Background(), parentScope, "parent", replaceIntent(t, 7), allowOrchestration,
				func(execution *TrackedStartManagedParentExecution) error {
					phase, phaseErr := execution.ExecutePreclaimedStopOld(tc.invoke)
					if !errors.Is(phaseErr, ErrIndeterminateExecution) ||
						phase.Record().State() != CommandStateClaimed {
						return errors.New("preclaimed StopOld failure was not fail-closed")
					}
					observed, observedErr := execution.ExecutePreclaimedStopOld(func() (TerminalOutcome, error) {
						return terminalOutcome(t, OutcomeSucceeded, "duplicate"), nil
					})
					if observedErr != nil || observed.Kind() != AdmissionInProgress {
						return errors.New("failed StopOld restored execution authority")
					}
					return phaseErr
				},
			)
			if !errors.Is(err, ErrIndeterminateExecution) || claimed.Record().State() != CommandStateClaimed {
				t.Fatalf("parent = %#v, %v", claimed, err)
			}
			close(releaseStart)
			if err := <-startDone; err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestTrackedStartPreclaimedStopGoexitExpiresAuthority(t *testing.T) {
	boundary := newTestBoundary(t)
	startScope, releaseStart, startDone := holdTrackedStart(t, boundary, "stop-goexit")
	parentScope := testScope(t, startScope.Domain(), "stop-goexit", OperationReplace)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = boundary.ExecuteManagedParentFromTrackedStart(
			context.Background(), parentScope, "parent", replaceIntent(t, 8), allowOrchestration,
			func(execution *TrackedStartManagedParentExecution) error {
				_, _ = execution.ExecutePreclaimedStopOld(func() (TerminalOutcome, error) {
					runtime.Goexit()
					return TerminalOutcome{}, nil
				})
				return nil
			},
		)
	}()
	<-done
	observed, err := boundary.ExecuteManagedParentFromTrackedStart(
		context.Background(), parentScope, "parent", replaceIntent(t, 8), allowOrchestration,
		func(*TrackedStartManagedParentExecution) error {
			t.Fatal("phase Goexit restored parent authority")
			return nil
		},
	)
	if err != nil || observed.Kind() != AdmissionInProgress {
		t.Fatalf("observation = %#v, %v", observed, err)
	}
	close(releaseStart)
	if err := <-startDone; err != nil {
		t.Fatal(err)
	}
}

func TestTrackedStartManagedParentPreClaimFailuresMutateNothing(t *testing.T) {
	boundary := newTestBoundary(t)
	startScope, releaseStart, startDone := holdTrackedStart(t, boundary, "preclaim")
	parentScope := testScope(t, startScope.Domain(), "preclaim", OperationRollback)
	intent := rollbackIntent(t, 9)
	denied := errors.New("denied")
	if _, err := boundary.ExecuteManagedParentFromTrackedStart(
		context.Background(), parentScope, "invalid", replaceIntent(t, 9), allowOrchestration,
		func(*TrackedStartManagedParentExecution) error {
			t.Fatal("invalid callback")
			return nil
		},
	); !errors.Is(err, ErrInvalidSubmission) {
		t.Fatalf("invalid submission error = %v", err)
	}
	if _, err := boundary.ExecuteManagedParentFromTrackedStart(
		context.Background(), parentScope, "denied", intent,
		func(context.Context, OrchestrationAuthorizationRequest) error { return denied },
		func(*TrackedStartManagedParentExecution) error {
			t.Fatal("denied callback")
			return nil
		},
	); !errors.Is(err, denied) {
		t.Fatalf("denial error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if _, err := boundary.ExecuteManagedParentFromTrackedStart(
		ctx, parentScope, "cancelled", intent,
		func(context.Context, OrchestrationAuthorizationRequest) error { cancel(); return nil },
		func(*TrackedStartManagedParentExecution) error {
			t.Fatal("cancelled callback")
			return nil
		},
	); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}
	ledger := boundary.storage.ledger(parentScope.instanceScope())
	ledger.mu.Lock()
	parents, phases := len(ledger.parents), len(ledger.phases)
	occupants := len(ledger.stopForStart)
	ledger.mu.Unlock()
	if parents != 0 || phases != 0 || occupants != 0 {
		t.Fatalf("pre-claim failure mutated facts: parents=%d phases=%d occupants=%d", parents, phases, occupants)
	}
	close(releaseStart)
	if err := <-startDone; err != nil {
		t.Fatal(err)
	}
}

func TestTrackedStartManagedParentCapabilityExpiresAfterCallbackReturn(t *testing.T) {
	boundary := newTestBoundary(t)
	startScope, releaseStart, startDone := holdTrackedStart(t, boundary, "retained")
	parentScope := testScope(t, startScope.Domain(), "retained", OperationReplace)
	var retained *TrackedStartManagedParentExecution
	_, err := boundary.ExecuteManagedParentFromTrackedStart(
		context.Background(), parentScope, "parent", replaceIntent(t, 11), allowOrchestration,
		func(execution *TrackedStartManagedParentExecution) error {
			retained = execution
			return nil
		},
	)
	if !errors.Is(err, ErrIndeterminateExecution) {
		t.Fatalf("callback return error = %v", err)
	}
	var calls atomic.Int32
	if _, err := retained.ExecutePreclaimedStopOld(func() (TerminalOutcome, error) {
		calls.Add(1)
		return terminalOutcome(t, OutcomeSucceeded, "escaped"), nil
	}); !errors.Is(err, ErrBoundaryExpired) {
		t.Fatalf("retained capability error = %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("retained capability invoked StopOld %d times", calls.Load())
	}
	close(releaseStart)
	if err := <-startDone; err != nil {
		t.Fatal(err)
	}
}

func TestTrackedStartManagedParentRejectsStaleGenerationWithoutMutation(t *testing.T) {
	storage := NewMemoryStorage()
	stale := newBoundary(t, storage)
	startScope, releaseStart, startDone := holdTrackedStart(t, stale, "stale-generation")
	active := newBoundary(t, storage)
	parentScope := testScope(t, startScope.Domain(), "stale-generation", OperationReplace)
	intent := replaceIntent(t, 10)
	if _, err := stale.ExecuteManagedParentFromTrackedStart(
		context.Background(), parentScope, "parent", intent, allowOrchestration,
		func(*TrackedStartManagedParentExecution) error {
			t.Fatal("stale Boundary received authority")
			return nil
		},
	); !errors.Is(err, ErrBoundaryExpired) {
		t.Fatalf("stale Boundary error = %v", err)
	}
	if _, err := active.ExecuteManagedParentFromTrackedStart(
		context.Background(), parentScope, "parent", intent, allowOrchestration,
		func(*TrackedStartManagedParentExecution) error {
			t.Fatal("reconstruction restored tracked Start authority")
			return nil
		},
	); !errors.Is(err, ErrInstanceBlocked) {
		t.Fatalf("active Boundary error = %v", err)
	}
	ledger := storage.ledger(parentScope.instanceScope())
	ledger.mu.Lock()
	parents, phases := len(ledger.parents), len(ledger.phases)
	ledger.mu.Unlock()
	if parents != 0 || phases != 0 {
		t.Fatalf("stale admission mutated facts: parents=%d phases=%d", parents, phases)
	}
	close(releaseStart)
	if err := <-startDone; !errors.Is(err, ErrBoundaryExpired) {
		t.Fatalf("stale tracked Start = %v", err)
	}
}

func TestTrackedStartManagedParentsOnDifferentInstancesProgressIndependently(t *testing.T) {
	boundary := newTestBoundary(t)
	type heldStart struct {
		scope   Scope
		release chan struct{}
		done    <-chan error
	}
	starts := []heldStart{}
	for _, instance := range []string{"independent-a", "independent-b"} {
		scope, release, done := holdTrackedStart(t, boundary, instance)
		starts = append(starts, heldStart{scope: scope, release: release, done: done})
	}
	entered := make(chan string, 2)
	releaseParents := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, held := range starts {
		held := held
		wg.Add(1)
		go func() {
			defer wg.Done()
			instance := string(held.scope.RuntimeInstanceID())
			parentScope := testScope(t, held.scope.Domain(), instance, OperationReplace)
			_, err := boundary.ExecuteManagedParentFromTrackedStart(
				context.Background(), parentScope, "parent", replaceIntent(t, 6), allowOrchestration,
				func(*TrackedStartManagedParentExecution) error {
					entered <- instance
					<-releaseParents
					return errors.New("leave unresolved")
				},
			)
			errs <- err
		}()
	}
	<-entered
	<-entered
	close(releaseParents)
	wg.Wait()
	close(errs)
	for err := range errs {
		if !errors.Is(err, ErrIndeterminateExecution) {
			t.Fatalf("parent result = %v", err)
		}
	}
	for _, held := range starts {
		close(held.release)
		if err := <-held.done; err != nil {
			t.Fatal(err)
		}
	}
}

func holdTrackedStart(
	t *testing.T,
	boundary *Boundary,
	instance string,
) (Scope, chan struct{}, <-chan error) {
	t.Helper()
	scope := testScope(t, "domain-a", instance, OperationStart)
	intent := startIntent(t, 1)
	outcome := terminalOutcome(t, OutcomeSucceeded, runtimeconfigload.LaunchAttemptID("attempt-"+instance))
	entered := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := boundary.Execute(context.Background(), scope, "tracked-start", intent, allow,
			func() (TerminalOutcome, error) {
				close(entered)
				<-release
				return outcome, nil
			})
		done <- err
	}()
	<-entered
	return scope, release, done
}
