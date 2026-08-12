package runtimecommandidempotency

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

func TestManagedPrimitiveEarlyAndFinalGatesResolveExactBinding(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 41)
	_, err := boundary.ExecuteManagedStart(
		context.Background(), scope, "start-a", intent, 7, "generation-a",
		func(context.Context, OrchestrationAuthorizationRequest) error { return nil },
		func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
			if gate, gateErr := boundary.ResolveManagedStartFinal(binding, ManagedStartFinalDisposition(99)); gateErr != ErrInvalidSubmission || gate != GateBlocked {
				t.Fatalf("invalid disposition = %v, %v", gate, gateErr)
			}
			if gate, gateErr := boundary.ResolveManagedStartEarly(binding, "attempt-a"); gateErr != nil || gate != GateClear {
				t.Fatalf("early gate = %v, %v", gate, gateErr)
			}
			if gate, gateErr := boundary.ResolveManagedStartFinal(binding, FinalContinue); gateErr != nil || gate != GateClear {
				t.Fatalf("final gate = %v, %v", gate, gateErr)
			}
			if _, gateErr := boundary.ResolveManagedStartFinal(binding, FinalContinue); gateErr != ErrIndeterminateExecution {
				t.Fatalf("reused binding error = %v", gateErr)
			}
			return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestManagedFinalGateConvergesStopAdmittedDuringBinding(t *testing.T) {
	boundary := newTestBoundary(t)
	startScope := testScope(t, "domain-a", "instance-a", OperationStart)
	stopScope := testScope(t, "domain-a", "instance-a", OperationStop)
	startIntent := startIntent(t, 41)
	stopIntent := NewStopIntent()
	bound := make(chan struct{})
	stopAdmitted := make(chan struct{})
	var wg sync.WaitGroup
	var startErr, stopErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, startErr = boundary.ExecuteManagedStart(
			context.Background(), startScope, "start-a", startIntent, 7, "generation-a",
			func(context.Context, OrchestrationAuthorizationRequest) error { return nil },
			func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
				if gate, err := boundary.ResolveManagedStartEarly(binding, "attempt-a"); err != nil || gate != GateClear {
					return TerminalOutcome{}, err
				}
				close(bound)
				<-stopAdmitted
				gate, err := boundary.ResolveManagedStartFinal(binding, FinalContinue)
				if err != nil || gate != GateStopConverged {
					t.Errorf("final gate = %v, %v", gate, err)
				}
				return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
			},
		)
	}()
	<-bound
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, stopErr = boundary.Execute(
			context.Background(), stopScope, "stop-a", stopIntent,
			func(context.Context, Scope, Intent) error { return nil },
			func() (TerminalOutcome, error) {
				return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
			},
		)
	}()
	// Stop callback runs only after the final gate signal, so admission is
	// observed through the command record rather than callback entry.
	admitted := false
	for attempts := 0; attempts < 10000; attempts++ {
		boundary.storage.clientMu.RLock()
		ledger := boundary.storage.existingLedger(startScope.instanceScope())
		if ledger != nil {
			ledger.mu.Lock()
			for identity := range ledger.records {
				if identity.scope.operation == OperationStop {
					admitted = true
				}
			}
			ledger.mu.Unlock()
		}
		boundary.storage.clientMu.RUnlock()
		if admitted {
			close(stopAdmitted)
			break
		}
		runtime.Gosched()
	}
	if !admitted {
		t.Fatal("Stop was not admitted within bounded scheduler yields")
	}
	wg.Wait()
	if startErr != nil || stopErr != nil {
		t.Fatalf("start=%v stop=%v", startErr, stopErr)
	}
}

func TestExecuteManagedParentCreatesExactLinkedStartTargetBinding(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationReplace)
	intent, _ := NewReplaceIntent(42)
	authorizeCalls := 0
	admission, err := boundary.ExecuteManagedParent(
		context.Background(), scope, "replace-a", intent,
		func(_ context.Context, request OrchestrationAuthorizationRequest) error {
			authorizeCalls++
			if request.Action() != OrchestrationActionReplaceExactTarget {
				t.Fatalf("authorization action = %v", request.Action())
			}
			return nil
		},
		func(parent *ManagedParentExecution) error {
			phase, prevented, phaseErr := parent.ContinueOrExecuteManagedStartTarget(
				context.Background(), 9, "generation-a",
				func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
					linked, ok := binding.LinkedExecutionIdentity()
					if !ok || linked.Phase().Ordinal() != 1 || linked.Parent().CommandKey() != "replace-a" {
						t.Fatalf("linked binding = %#v, %v", linked, ok)
					}
					if gate, gateErr := boundary.ResolveManagedStartEarly(binding, "attempt-a"); gateErr != nil || gate != GateClear {
						t.Fatalf("early gate = %v, %v", gate, gateErr)
					}
					if gate, gateErr := boundary.ResolveManagedStartFinal(binding, FinalContinue); gateErr != nil || gate != GateClear {
						t.Fatalf("final gate = %v, %v", gate, gateErr)
					}
					return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
				},
			)
			if phaseErr != nil || prevented || phase.Kind() != AdmissionClaimed {
				t.Fatalf("phase = %#v, prevented=%v, error=%v", phase, prevented, phaseErr)
			}
			parentOutcome, _ := NewParentTerminalOutcome(ParentOutcomeSucceeded)
			_, publishErr := parent.PublishTerminal(parentOutcome)
			return publishErr
		},
	)
	if err != nil || admission.Kind() != AdmissionClaimed || authorizeCalls != 1 {
		t.Fatalf("parent = %#v, auth=%d, error=%v", admission, authorizeCalls, err)
	}
	observed, err := boundary.ExecuteManagedParent(
		context.Background(), scope, "replace-a", intent,
		func(context.Context, OrchestrationAuthorizationRequest) error { authorizeCalls++; return nil },
		func(*ManagedParentExecution) error { t.Fatal("replay received authority"); return nil },
	)
	if err != nil || observed.Kind() != AdmissionReplay || authorizeCalls != 2 {
		t.Fatalf("replay = %#v, auth=%d, error=%v", observed, authorizeCalls, err)
	}
}

func TestManagedNoClaimAllCausesPrimitiveAndLinked(t *testing.T) {
	for _, tc := range []struct {
		name     string
		cause    runtimeorchestrationbinding.StartNoClaimCause
		category OutcomeCategory
	}{
		{"cancelled", runtimeorchestrationbinding.StartNoClaimCancelled, OutcomeRejected},
		{"rejected", runtimeorchestrationbinding.StartNoClaimRejected, OutcomeRejected},
		{"failed", runtimeorchestrationbinding.StartNoClaimFailed, OutcomeFailed},
	} {
		for _, linked := range []bool{false, true} {
			for _, withStop := range []bool{false, true} {
				name := "primitive"
				if linked {
					name = "linked"
				}
				if withStop {
					name += "/pending-stop"
				} else {
					name += "/no-stop"
				}
				t.Run(tc.name+"/"+name, func(t *testing.T) {
					boundary := newTestBoundary(t)
					invoke := func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
						var stopResult <-chan error
						var stopCallbacks *atomic.Int32
						if withStop {
							stopResult, stopCallbacks = admitManagedPendingStop(t, boundary, binding.Authorization())
						}
						if err := boundary.SignalManagedStartNoClaim(binding, tc.cause); err != nil {
							t.Fatal(err)
						}
						if withStop {
							if err := <-stopResult; err != nil || stopCallbacks.Load() != 0 {
								t.Fatalf("no-claim Stop = %v, callbacks=%d", err, stopCallbacks.Load())
							}
						}
						return NewTerminalOutcome(tc.category, "")
					}
					if !linked {
						scope := testScope(t, "domain-a", "instance-a", OperationStart)
						intent := startIntent(t, 41)
						_, err := boundary.ExecuteManagedStart(
							context.Background(), scope, "start-a", intent, 7, "generation-a",
							func(context.Context, OrchestrationAuthorizationRequest) error { return nil }, invoke,
						)
						if err != nil {
							t.Fatal(err)
						}
						return
					}
					scope := testScope(t, "domain-a", "instance-a", OperationReplace)
					intent, _ := NewReplaceIntent(41)
					_, err := boundary.ExecuteManagedParent(
						context.Background(), scope, "replace-a", intent,
						func(context.Context, OrchestrationAuthorizationRequest) error { return nil },
						func(parent *ManagedParentExecution) error {
							_, _, phaseErr := parent.ContinueOrExecuteManagedStartTarget(
								context.Background(), 7, "generation-a", invoke,
							)
							if phaseErr != nil {
								return phaseErr
							}
							parentCategory := ParentOutcomeRejected
							if tc.cause == runtimeorchestrationbinding.StartNoClaimCancelled {
								parentCategory = ParentOutcomeCancelled
							} else if tc.cause == runtimeorchestrationbinding.StartNoClaimFailed {
								parentCategory = ParentOutcomeFailed
							}
							outcome, _ := NewParentTerminalOutcome(parentCategory)
							_, publishErr := parent.PublishTerminal(outcome)
							return publishErr
						},
					)
					if err != nil {
						t.Fatal(err)
					}
				})
			}
		}
	}
}

func admitManagedPendingStop(
	t *testing.T,
	boundary *Boundary,
	authorization runtimeorchestrationbinding.OrchestrationAuthorizationRequest,
) (<-chan error, *atomic.Int32) {
	t.Helper()
	result := make(chan error, 1)
	callbacks := &atomic.Int32{}
	stopScope, _ := NewScope(
		authorization.OperationalDomain(), authorization.WorkspaceID(), authorization.ConfigurationID(),
		authorization.RuntimeInstanceID(), OperationStop,
	)
	go func() {
		_, err := boundary.Execute(
			context.Background(), stopScope, "stop-a", NewStopIntent(),
			func(context.Context, Scope, Intent) error { return nil },
			func() (TerminalOutcome, error) {
				callbacks.Add(1)
				return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
			},
		)
		result <- err
	}()
	ledger := boundary.storage.existingLedger(stopScope.instanceScope())
	for attempts := 0; attempts < 10000; attempts++ {
		admitted := false
		if ledger != nil {
			ledger.mu.Lock()
			for identity := range ledger.records {
				if identity.scope.operation == OperationStop {
					admitted = true
				}
			}
			ledger.mu.Unlock()
		}
		if admitted {
			return result, callbacks
		}
		runtime.Gosched()
	}
	t.Fatal("pending Stop was not admitted within bounded scheduler yields")
	return nil, nil
}

func TestManagedEarlyPendingStopConvergesOwnerClaimed(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	_, err := boundary.ExecuteManagedStart(context.Background(), scope, "start-a", startIntent(t, 41), 7, "generation-a", allowOrchestration, func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
		stopResult, callbacks := admitManagedPendingStop(t, boundary, binding.Authorization())
		gate, gateErr := boundary.ResolveManagedStartEarly(binding, "attempt-a")
		if gateErr != nil || gate != GateStopConverged {
			return TerminalOutcome{}, gateErr
		}
		if stopErr := <-stopResult; stopErr != nil || callbacks.Load() != 1 {
			t.Fatalf("stop=%v callbacks=%d", stopErr, callbacks.Load())
		}
		return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestManagedNoClaimRejectsDuplicatePostEarlyAndForgedBinding(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 41)
	_, err := boundary.ExecuteManagedStart(
		context.Background(), scope, "start-a", intent, 7, "generation-a",
		func(context.Context, OrchestrationAuthorizationRequest) error { return nil },
		func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
			forgedHandle, _ := runtimeorchestrationbinding.NewStartRendezvous("forged")
			forged, _ := runtimeorchestrationbinding.NewPrimitiveStartExecutionBinding(
				binding.Authorization(), binding.ExpectedAggregateRevision(), binding.ExecutionGeneration(), forgedHandle,
			)
			if got := boundary.SignalManagedStartNoClaim(forged, runtimeorchestrationbinding.StartNoClaimRejected); got != ErrIndeterminateExecution {
				t.Fatalf("forged error = %v", got)
			}
			if got := boundary.SignalManagedStartNoClaim(binding, runtimeorchestrationbinding.StartNoClaimRejected); got != nil {
				t.Fatal(got)
			}
			if got := boundary.SignalManagedStartNoClaim(binding, runtimeorchestrationbinding.StartNoClaimRejected); got != ErrIndeterminateExecution {
				t.Fatalf("duplicate error = %v", got)
			}
			return NewTerminalOutcome(OutcomeRejected, "")
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	boundary = newTestBoundary(t)
	_, err = boundary.ExecuteManagedStart(
		context.Background(), scope, "start-b", intent, 7, "generation-a",
		func(context.Context, OrchestrationAuthorizationRequest) error { return nil },
		func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
			if gate, gateErr := boundary.ResolveManagedStartEarly(binding, "attempt-a"); gateErr != nil || gate != GateClear {
				t.Fatalf("early = %v, %v", gate, gateErr)
			}
			if got := boundary.SignalManagedStartNoClaim(binding, runtimeorchestrationbinding.StartNoClaimRejected); got != ErrIndeterminateExecution {
				t.Fatalf("post-early error = %v", got)
			}
			_, _ = boundary.ResolveManagedStartFinal(binding, FinalContinue)
			return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestManagedCallbackExitWithoutSignalBlocksPendingStopAndLeavesClaimed(t *testing.T) {
	for _, linked := range []bool{false, true} {
		for _, mode := range []string{"return", "panic", "goexit"} {
			name := mode + "/primitive"
			if linked {
				name = mode + "/linked"
			}
			t.Run(name, func(t *testing.T) {
				boundary := newTestBoundary(t)
				var stopResult <-chan error
				done := make(chan struct{})
				invoke := func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
					stopResult, _ = admitManagedPendingStop(t, boundary, binding.Authorization())
					switch mode {
					case "panic":
						panic("lost callback")
					case "goexit":
						runtime.Goexit()
					}
					return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
				}
				go func() {
					defer close(done)
					if !linked {
						scope := testScope(t, "domain-a", "instance-a", OperationStart)
						intent := startIntent(t, 41)
						_, _ = boundary.ExecuteManagedStart(
							context.Background(), scope, "start-a", intent, 7, "generation-a",
							func(context.Context, OrchestrationAuthorizationRequest) error { return nil }, invoke,
						)
						return
					}
					scope := testScope(t, "domain-a", "instance-a", OperationReplace)
					intent, _ := NewReplaceIntent(41)
					_, _ = boundary.ExecuteManagedParent(
						context.Background(), scope, "replace-a", intent,
						func(context.Context, OrchestrationAuthorizationRequest) error { return nil },
						func(parent *ManagedParentExecution) error {
							_, _, err := parent.ContinueOrExecuteManagedStartTarget(
								context.Background(), 7, "generation-a", invoke,
							)
							return err
						},
					)
				}()
				<-done
				if stopResult == nil {
					t.Fatal("callback exited before pending Stop admission")
				}
				if err := <-stopResult; err != ErrIndeterminateExecution {
					t.Fatalf("pending Stop error = %v", err)
				}
				if !linked {
					scope := testScope(t, "domain-a", "instance-a", OperationStart)
					intent := startIntent(t, 41)
					observation, err := boundary.ExecuteManagedStart(
						context.Background(), scope, "start-a", intent, 7, "generation-a",
						func(context.Context, OrchestrationAuthorizationRequest) error { return nil },
						func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
							t.Fatal("lost claim was reissued")
							return TerminalOutcome{}, nil
						},
					)
					if err != nil || observation.Kind() != AdmissionInProgress || observation.Record().State() != CommandStateClaimed {
						t.Fatalf("durable observation = %#v/%v", observation, err)
					}
				}
			})
		}
	}
}

func TestManagedFinalContinueLateStopPermitSurvivesStartCallbackReturn(t *testing.T) {
	boundary := newTestBoundary(t)
	startScope := testScope(t, "domain-a", "instance-a", OperationStart)
	stopScope := testScope(t, "domain-a", "instance-a", OperationStop)
	stopEntered := make(chan struct{})
	releaseStop := make(chan struct{})
	startDone := make(chan error, 1)
	stopDone := make(chan error, 1)
	go func() {
		_, err := boundary.ExecuteManagedStart(
			context.Background(), startScope, "start-a", startIntent(t, 41), 7, "generation-a",
			func(context.Context, OrchestrationAuthorizationRequest) error { return nil },
			func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
				if gate, gateErr := boundary.ResolveManagedStartEarly(binding, "attempt-a"); gateErr != nil || gate != GateClear {
					return TerminalOutcome{}, gateErr
				}
				if gate, gateErr := boundary.ResolveManagedStartFinal(binding, FinalContinue); gateErr != nil || gate != GateClear {
					return TerminalOutcome{}, gateErr
				}
				go func() {
					_, stopErr := boundary.Execute(
						context.Background(), stopScope, "stop-a", NewStopIntent(),
						func(context.Context, Scope, Intent) error { return nil },
						func() (TerminalOutcome, error) {
							close(stopEntered)
							<-releaseStop
							return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
						},
					)
					stopDone <- stopErr
				}()
				<-stopEntered
				return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
			},
		)
		startDone <- err
	}()
	if err := <-startDone; err != nil {
		t.Fatal(err)
	}
	close(releaseStop)
	if err := <-stopDone; err != nil {
		t.Fatalf("retained late Stop permit failed after Start return: %v", err)
	}
}

func TestExecuteManagedParentRollbackAuthorizationAndAdmissionFailures(t *testing.T) {
	boundary := newTestBoundary(t)
	rollback := testScope(t, "domain-a", "instance-a", OperationRollback)
	rollbackIntent, _ := NewRollbackIntent(44)
	seen := OrchestrationAuthorizationRequest{}
	_, _ = boundary.ExecuteManagedParent(
		context.Background(), rollback, "rollback-a", rollbackIntent,
		func(_ context.Context, request OrchestrationAuthorizationRequest) error {
			seen = request
			return errors.New("deny")
		},
		func(*ManagedParentExecution) error { t.Fatal("denied parent invoked"); return nil },
	)
	if seen.Action() != OrchestrationActionRollbackExactTarget || seen.TargetConfigurationVersionID() != 44 {
		t.Fatalf("rollback authorization = %#v", seen)
	}
	if ledger := boundary.storage.existingLedger(rollback.instanceScope()); ledger != nil {
		ledger.mu.Lock()
		mutated := len(ledger.parents) != 0
		ledger.mu.Unlock()
		if mutated {
			t.Fatal("denied authorization mutated parent ledger")
		}
	}
	for _, tc := range []struct {
		name   string
		scope  Scope
		intent Intent
		auth   AuthorizeOrchestration
	}{
		{"mismatch", rollback, func() Intent { value, _ := NewReplaceIntent(44); return value }(), allowOrchestration},
		{"nil-auth", rollback, rollbackIntent, nil},
		{"panic-auth", rollback, rollbackIntent, func(context.Context, OrchestrationAuthorizationRequest) error { panic("auth") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := boundary.ExecuteManagedParent(
				context.Background(), tc.scope, CommandKey("invalid-"+tc.name), tc.intent, tc.auth,
				func(*ManagedParentExecution) error { t.Fatal("invalid parent invoked"); return nil },
			); err == nil {
				t.Fatal("invalid managed parent accepted")
			}
		})
	}
}

func TestExecuteManagedParentCancellationConcurrentObservationAndLegacyNonAdoption(t *testing.T) {
	t.Run("cancel-after-auth", func(t *testing.T) {
		boundary := newTestBoundary(t)
		scope := testScope(t, "domain-a", "instance-a", OperationReplace)
		intent, _ := NewReplaceIntent(41)
		ctx, cancel := context.WithCancel(context.Background())
		_, err := boundary.ExecuteManagedParent(ctx, scope, "replace-a", intent,
			func(context.Context, OrchestrationAuthorizationRequest) error { cancel(); return nil },
			func(*ManagedParentExecution) error { t.Fatal("cancelled callback"); return nil })
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error=%v", err)
		}
		ledger := boundary.storage.existingLedger(scope.instanceScope())
		if ledger != nil {
			ledger.mu.Lock()
			count := len(ledger.parents)
			ledger.mu.Unlock()
			if count != 0 {
				t.Fatal("cancel mutated")
			}
		}
	})
	t.Run("concurrent-same-key", func(t *testing.T) {
		boundary := newTestBoundary(t)
		scope := testScope(t, "domain-a", "instance-a", OperationReplace)
		intent, _ := NewReplaceIntent(41)
		entered := make(chan struct{})
		release := make(chan struct{})
		done := make(chan error, 1)
		var calls atomic.Int32
		go func() {
			_, err := boundary.ExecuteManagedParent(context.Background(), scope, "replace-a", intent, allowOrchestration, func(*ManagedParentExecution) error {
				calls.Add(1)
				close(entered)
				<-release
				return errors.New("leave claimed")
			})
			done <- err
		}()
		<-entered
		observation, err := boundary.ExecuteManagedParent(context.Background(), scope, "replace-a", intent, allowOrchestration, func(*ManagedParentExecution) error { t.Fatal("duplicate capability"); return nil })
		if err != nil || observation.Kind() != AdmissionInProgress || calls.Load() != 1 {
			t.Fatalf("observation=%v/%v calls=%d", observation.Kind(), err, calls.Load())
		}
		close(release)
		<-done
	})
	t.Run("legacy-in-progress-non-adoption", func(t *testing.T) {
		boundary := newTestBoundary(t)
		scope := testScope(t, "domain-a", "instance-a", OperationReplace)
		intent, _ := NewReplaceIntent(41)
		entered := make(chan struct{})
		release := make(chan struct{})
		go func() {
			_, _ = boundary.ExecuteParent(context.Background(), scope, "replace-a", intent, func(context.Context, Scope, Intent) error { return nil }, func(*ParentExecution) error { close(entered); <-release; return errors.New("claimed") })
		}()
		<-entered
		observation, err := boundary.ExecuteManagedParent(context.Background(), scope, "replace-a", intent, allowOrchestration, func(*ManagedParentExecution) error { t.Fatal("legacy upgraded"); return nil })
		if err != nil || observation.Kind() != AdmissionInProgress {
			t.Fatalf("observation=%v/%v", observation.Kind(), err)
		}
		close(release)
	})
}

func TestExecuteManagedParentRollbackSuccessCreatesLinkedOutcome(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationRollback)
	intent, _ := NewRollbackIntent(44)
	admission, err := boundary.ExecuteManagedParent(context.Background(), scope, "rollback-a", intent, allowOrchestration, func(parent *ManagedParentExecution) error {
		phase, prevented, phaseErr := parent.ContinueOrExecuteManagedStartTarget(context.Background(), 7, "generation-r", func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
			if binding.Authorization().Action() != OrchestrationActionRollbackExactTarget {
				t.Fatal("wrong action")
			}
			if gate, e := boundary.ResolveManagedStartEarly(binding, "attempt-r"); e != nil || gate != GateClear {
				return TerminalOutcome{}, e
			}
			if gate, e := boundary.ResolveManagedStartFinal(binding, FinalContinue); e != nil || gate != GateClear {
				return TerminalOutcome{}, e
			}
			return NewTerminalOutcome(OutcomeSucceeded, "attempt-r")
		})
		if phaseErr != nil || prevented || phase.Kind() != AdmissionClaimed {
			return phaseErr
		}
		out, _ := NewParentTerminalOutcome(ParentOutcomeSucceeded)
		_, e := parent.PublishTerminal(out)
		return e
	})
	if err != nil || admission.Kind() != AdmissionClaimed {
		t.Fatalf("rollback=%v/%v", admission.Kind(), err)
	}
}

func TestExecuteManagedParentLegacyTerminalNonAdoption(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationReplace)
	intent, _ := NewReplaceIntent(41)
	legacy, err := boundary.ExecuteParent(context.Background(), scope, "replace-a", intent, func(context.Context, Scope, Intent) error { return nil }, func(parent *ParentExecution) error {
		out, _ := NewParentTerminalOutcome(ParentOutcomeSatisfied)
		_, e := parent.PublishTerminal(out)
		return e
	})
	if err != nil || legacy.Kind() != AdmissionClaimed {
		t.Fatalf("legacy=%v/%v", legacy.Kind(), err)
	}
	managed, err := boundary.ExecuteManagedParent(context.Background(), scope, "replace-a", intent, allowOrchestration, func(*ManagedParentExecution) error { t.Fatal("legacy terminal upgraded"); return nil })
	if err != nil || managed.Kind() != AdmissionReplay {
		t.Fatalf("managed=%v/%v", managed.Kind(), err)
	}
}

func TestManagedFinalBindingFailedRejectsWrongTerminalPrimitiveAndLinked(t *testing.T) {
	for _, linked := range []bool{false, true} {
		name := "primitive"
		if linked {
			name = "linked"
		}
		t.Run(name, func(t *testing.T) {
			boundary := newTestBoundary(t)
			invoke := func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
				if gate, e := boundary.ResolveManagedStartEarly(binding, "attempt-a"); e != nil || gate != GateClear {
					return TerminalOutcome{}, e
				}
				if gate, e := boundary.ResolveManagedStartFinal(binding, FinalBindingFailed); e != nil || gate != GateClear {
					return TerminalOutcome{}, e
				}
				return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
			}
			if !linked {
				scope := testScope(t, "domain-a", "instance-a", OperationStart)
				admission, e := boundary.ExecuteManagedStart(context.Background(), scope, "start-a", startIntent(t, 41), 7, "generation-a", allowOrchestration, invoke)
				if e != ErrIndeterminateExecution || admission.Record().State() != CommandStateClaimed {
					t.Fatalf("primitive=%v/%v", admission.Record().State(), e)
				}
				return
			}
			scope := testScope(t, "domain-a", "instance-a", OperationReplace)
			intent, _ := NewReplaceIntent(41)
			_, e := boundary.ExecuteManagedParent(context.Background(), scope, "replace-a", intent, allowOrchestration, func(parent *ManagedParentExecution) error {
				phase, _, pe := parent.ContinueOrExecuteManagedStartTarget(context.Background(), 7, "generation-a", invoke)
				if pe != ErrIndeterminateExecution || phase.Record().State() != CommandStateClaimed {
					t.Fatalf("linked=%v/%v", phase.Record().State(), pe)
				}
				return pe
			})
			if e != ErrIndeterminateExecution {
				t.Fatalf("parent error=%v", e)
			}
		})
	}
}
