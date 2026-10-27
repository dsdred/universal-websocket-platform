package runtimecommandidempotency

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var stop *managedPendingStop
	var helperErr error
	defer func() {
		if stop != nil {
			if err := stop.cancelAndJoin(); err != nil {
				t.Error(err)
			}
		}
	}()
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	_, err := boundary.ExecuteManagedStart(ctx, scope, "start-a", startIntent(t, 41), 7, "generation-a", allowOrchestration,
		func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
			if gate, err := boundary.ResolveManagedStartEarly(binding, "attempt-a"); err != nil || gate != GateClear {
				return TerminalOutcome{}, fmt.Errorf("early gate = %v, %v", gate, err)
			}
			var err error
			stop, err = admitManagedPendingStop(ctx, boundary, binding)
			helperErr = err
			if err != nil {
				return TerminalOutcome{}, err
			}
			if gate, err := boundary.ResolveManagedStartFinal(binding, FinalContinue); err != nil || gate != GateStopConverged {
				return TerminalOutcome{}, fmt.Errorf("final gate = %v, %v", gate, err)
			}
			if err := stop.wait(ctx); err != nil {
				return TerminalOutcome{}, err
			}
			return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
		})
	if helperErr != nil {
		t.Fatal(helperErr)
	}
	if err != nil {
		t.Fatal(err)
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
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					var stop *managedPendingStop
					var helperErr error
					defer func() {
						if stop != nil {
							if err := stop.cancelAndJoin(); err != nil {
								t.Error(err)
							}
						}
					}()
					invoke := func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
						if withStop {
							var err error
							stop, err = admitManagedPendingStop(ctx, boundary, binding)
							helperErr = err
							if err != nil {
								return TerminalOutcome{}, err
							}
						}
						if err := boundary.SignalManagedStartNoClaim(binding, tc.cause); err != nil {
							return TerminalOutcome{}, err
						}
						if withStop {
							if err := stop.wait(ctx); err != nil || stop.callbacks.Load() != 0 {
								return TerminalOutcome{}, fmt.Errorf("no-claim Stop = %v, callbacks=%d", err, stop.callbacks.Load())
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
						if helperErr != nil {
							t.Fatal(helperErr)
						}
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
					if helperErr != nil {
						t.Fatal(helperErr)
					}
					if err != nil {
						t.Fatal(err)
					}
				})
			}
		}
	}
}

// managedPendingStop observes the existing admission notification, not callback entry.
// Result fields are immutable after done closes; cleanup never consumes a result twice.
type managedPendingStop struct {
	ledger    *commandLedger
	bridge    *startRendezvous
	identity  commandIdentity
	notify    <-chan struct{}
	admitted  chan struct{}
	ack       sync.Once
	done      chan struct{}
	cancel    context.CancelFunc
	admission Admission
	err       error
	callbacks atomic.Int32
}

func startManagedPendingStop(ctx context.Context, boundary *Boundary, binding runtimeorchestrationbinding.StartExecutionBinding, authorize Authorize) (*managedPendingStop, error) {
	authorization := binding.Authorization()
	scope, err := NewScope(authorization.OperationalDomain(), authorization.WorkspaceID(), authorization.ConfigurationID(), authorization.RuntimeInstanceID(), OperationStop)
	if err != nil {
		return nil, err
	}
	ledger := boundary.storage.existingLedger(scope.instanceScope())
	if ledger == nil {
		return nil, fmt.Errorf("managed Start ledger missing")
	}
	identity := commandIdentity{scope: scope, key: "stop-a"}
	ledger.mu.Lock()
	managed := ledger.managedStart[binding.Rendezvous()]
	if managed == nil || managed.binding != binding || managed.bridge == nil || ledger.records[identity] != nil || managed.bridge.stopConsumed || !managed.liveLocked(boundary, ledger) {
		ledger.mu.Unlock()
		return nil, fmt.Errorf("managed Stop fixture is not fresh")
	}
	// Subscribe under the admission lock BEFORE launching the Stop worker.
	bridge, notify := managed.bridge, managed.bridge.notify
	ledger.mu.Unlock()
	workerCtx, cancel := context.WithCancel(ctx)
	stop := &managedPendingStop{ledger: ledger, bridge: bridge, identity: identity, notify: notify, admitted: make(chan struct{}), done: make(chan struct{}), cancel: cancel}
	go func() {
		defer close(stop.done)
		stop.admission, stop.err = boundary.Execute(workerCtx, scope, "stop-a", NewStopIntent(), authorize,
			func() (TerminalOutcome, error) {
				// Keep the synthetic callback from completing a Claimed record before
				// the observer acknowledges it, including the already-OwnerClaimed case.
				select {
				case <-stop.admitted:
				case <-workerCtx.Done():
					return TerminalOutcome{}, workerCtx.Err()
				}
				stop.callbacks.Add(1)
				return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
			})
	}()
	return stop, nil
}

func (stop *managedPendingStop) admissionSnapshot() (bool, <-chan struct{}) {
	stop.ledger.mu.Lock()
	defer stop.ledger.mu.Unlock()
	record := stop.ledger.records[stop.identity]
	state := stop.bridge.stopState
	admitted := record != nil && record.identity == stop.identity && record.state == CommandStateClaimed &&
		stop.bridge.stop == stop.identity && stop.bridge.stopConsumed && state != nil && record.revision == state.revision &&
		stop.ledger.live[stop.identity] == state
	return admitted, stop.bridge.notify
}

func (stop *managedPendingStop) awaitAdmission(ctx context.Context) error {
	notify := stop.notify
	acknowledge := func() { stop.ack.Do(func() { close(stop.admitted) }) }
	for {
		admitted, nextNotify := stop.admissionSnapshot()
		if admitted {
			acknowledge()
			return nil
		}
		select {
		case <-stop.done:
			if admitted, _ := stop.admissionSnapshot(); admitted {
				acknowledge()
				return nil
			}
			if stop.err != nil {
				return fmt.Errorf("Stop returned before admission acknowledgment (kind=%v): %w", stop.admission.Kind(), stop.err)
			}
			return fmt.Errorf("Stop returned before admission acknowledgment without error (kind=%v)", stop.admission.Kind())
		case <-ctx.Done():
			if admitted, _ := stop.admissionSnapshot(); admitted {
				acknowledge()
				return nil
			}
			return fmt.Errorf("Stop admission watchdog: %w", ctx.Err())
		case <-notify:
			notify = nextNotify
		}
	}
}

func admitManagedPendingStop(ctx context.Context, boundary *Boundary, binding runtimeorchestrationbinding.StartExecutionBinding) (*managedPendingStop, error) {
	stop, err := startManagedPendingStop(ctx, boundary, binding, func(context.Context, Scope, Intent) error { return nil })
	if err != nil {
		return stop, err
	}
	return stop, stop.awaitAdmission(ctx)
}

func (stop *managedPendingStop) wait(ctx context.Context) error {
	select {
	case <-stop.done:
		return stop.err
	case <-ctx.Done():
		return fmt.Errorf("Stop completion watchdog: %w", ctx.Err())
	}
}

func (stop *managedPendingStop) cancelAndJoin() error {
	stop.cancel()
	watchdog := time.NewTimer(10 * time.Second)
	defer watchdog.Stop()
	select {
	case <-stop.done:
		return nil
	case <-watchdog.C:
		return fmt.Errorf("Stop worker did not join after cancellation")
	}
}

func TestManagedEarlyPendingStopConvergesOwnerClaimed(t *testing.T) {
	boundary := newTestBoundary(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var stop *managedPendingStop
	var helperErr error
	defer func() {
		if stop != nil {
			if err := stop.cancelAndJoin(); err != nil {
				t.Error(err)
			}
		}
	}()
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	_, err := boundary.ExecuteManagedStart(ctx, scope, "start-a", startIntent(t, 41), 7, "generation-a", allowOrchestration, func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
		var err error
		stop, err = admitManagedPendingStop(ctx, boundary, binding)
		helperErr = err
		if err != nil {
			return TerminalOutcome{}, err
		}
		if gate, gateErr := boundary.ResolveManagedStartEarly(binding, "attempt-a"); gateErr != nil || gate != GateStopConverged {
			return TerminalOutcome{}, fmt.Errorf("early gate = %v, %v", gate, gateErr)
		}
		if stopErr := stop.wait(ctx); stopErr != nil || stop.callbacks.Load() != 1 {
			return TerminalOutcome{}, fmt.Errorf("stop=%v callbacks=%d", stopErr, stop.callbacks.Load())
		}
		return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
	})
	if helperErr != nil {
		t.Fatal(helperErr)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestManagedPendingStopAdmissionSynchronization(t *testing.T) {
	for _, scenario := range []string{"delayed-admission", "notification-before-wait", "early-rejection", "watchdog"} {
		t.Run(scenario, func(t *testing.T) {
			boundary := newTestBoundary(t)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			var stop *managedPendingStop
			defer func() {
				if stop != nil {
					if err := stop.cancelAndJoin(); err != nil {
						t.Error(err)
					}
				}
			}()
			denied := errors.New("controlled Stop authorization rejection")
			var proofErr error
			scope := testScope(t, "domain-a", "instance-a", OperationStart)
			_, executionErr := boundary.ExecuteManagedStart(ctx, scope, "start-a", startIntent(t, 41), 7, "generation-a", allowOrchestration,
				func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
					authorizationEntered := make(chan struct{})
					releaseAuthorization := make(chan struct{})
					authorize := func(workerCtx context.Context, _ Scope, _ Intent) error {
						close(authorizationEntered)
						if scenario == "early-rejection" {
							return denied
						}
						if scenario == "delayed-admission" || scenario == "watchdog" {
							select {
							case <-releaseAuthorization:
							case <-workerCtx.Done():
								return workerCtx.Err()
							}
						}
						return nil
					}
					var err error
					stop, err = startManagedPendingStop(ctx, boundary, binding, authorize)
					if err != nil {
						proofErr = err
						return TerminalOutcome{}, err
					}
					prove := func() (proofError error) {
						switch scenario {
						case "delayed-admission":
							waitCtx, cancelWait := context.WithCancel(ctx)
							defer cancelWait()
							waitDone := make(chan struct{})
							var admissionErr error
							go func() { defer close(waitDone); admissionErr = stop.awaitAdmission(waitCtx) }()
							defer func() {
								cancelWait()
								join := time.NewTimer(10 * time.Second)
								defer join.Stop()
								select {
								case <-waitDone:
								case <-join.C:
									proofError = fmt.Errorf("admission observer did not join")
								}
							}()
							select {
							case <-authorizationEntered:
							case <-ctx.Done():
								return ctx.Err()
							}
							if admitted, _ := stop.admissionSnapshot(); admitted {
								return fmt.Errorf("admission acknowledged while authorization held")
							}
							select {
							case <-waitDone:
								return fmt.Errorf("admission observer returned before authorization release: %v", admissionErr)
							default:
							}
							close(releaseAuthorization)
							select {
							case <-waitDone:
							case <-ctx.Done():
								return ctx.Err()
							}
							if admissionErr != nil {
								return admissionErr
							}
						case "notification-before-wait":
							// Consume no new notification: the pre-launch subscription is
							// already closed before awaitAdmission is called.
							select {
							case <-stop.notify:
							case <-ctx.Done():
								return ctx.Err()
							}
							if err := stop.awaitAdmission(ctx); err != nil {
								return err
							}
						case "early-rejection":
							if err := stop.awaitAdmission(ctx); !errors.Is(err, denied) {
								return fmt.Errorf("early rejection = %v", err)
							}
							if err := stop.wait(ctx); err != denied {
								return fmt.Errorf("Stop rejection result = %v", err)
							}
						case "watchdog":
							select {
							case <-authorizationEntered:
							case <-ctx.Done():
								return ctx.Err()
							}
							// An already-expired observer deadline deterministically exercises
							// the watchdog; the worker has a separate live context.
							waitCtx, cancelWait := context.WithDeadline(ctx, time.Now().Add(-time.Second))
							defer cancelWait()
							if err := stop.awaitAdmission(waitCtx); !errors.Is(err, context.DeadlineExceeded) {
								return fmt.Errorf("admission watchdog = %v", err)
							}
							if err := stop.cancelAndJoin(); err != nil {
								return err
							}
							if err := stop.wait(ctx); !errors.Is(err, context.Canceled) {
								return fmt.Errorf("cancelled Stop result = %v", err)
							}
						}
						if scenario == "early-rejection" || scenario == "watchdog" {
							stop.ledger.mu.Lock()
							record := stop.ledger.records[stop.identity]
							stop.ledger.mu.Unlock()
							if record != nil {
								return fmt.Errorf("failed admission created stop-a claim")
							}
						} else if admitted, _ := stop.admissionSnapshot(); !admitted {
							return fmt.Errorf("exact stop-a claim missing after acknowledgment")
						}
						if err := boundary.SignalManagedStartNoClaim(binding, runtimeorchestrationbinding.StartNoClaimRejected); err != nil {
							return err
						}
						if scenario == "delayed-admission" || scenario == "notification-before-wait" {
							if err := stop.wait(ctx); err != nil {
								return fmt.Errorf("no-claim Stop result = %v", err)
							}
						}
						if stop.callbacks.Load() != 0 {
							return fmt.Errorf("no-claim/failed admission invoked lifecycle callback")
						}
						if err := stop.cancelAndJoin(); err != nil {
							return err
						}
						select {
						case <-stop.done:
						default:
							return fmt.Errorf("Stop worker not joined")
						}
						return nil
					}
					proofErr = prove()
					if proofErr != nil {
						return TerminalOutcome{}, proofErr
					}
					return NewTerminalOutcome(OutcomeRejected, "")
				})
			if proofErr != nil {
				t.Fatal(proofErr)
			}
			if executionErr != nil {
				t.Fatal(executionErr)
			}
		})
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
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				startScope := testScope(t, "domain-a", "instance-a", OperationStart)
				parentScope := testScope(t, "domain-a", "instance-a", OperationReplace)
				intent := startIntent(t, 41)
				parentIntent, _ := NewReplaceIntent(41)
				var stop *managedPendingStop
				var helperErr error
				done := make(chan struct{})
				defer func() {
					cancel()
					join := time.NewTimer(10 * time.Second)
					defer join.Stop()
					select {
					case <-done:
					case <-join.C:
						t.Error("Start worker did not join")
						return
					}
					if stop != nil {
						if err := stop.cancelAndJoin(); err != nil {
							t.Error(err)
						}
					}
				}()
				invoke := func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
					stop, helperErr = admitManagedPendingStop(ctx, boundary, binding)
					if helperErr != nil {
						return TerminalOutcome{}, helperErr
					}
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
						_, _ = boundary.ExecuteManagedStart(ctx, startScope, "start-a", intent, 7, "generation-a", allowOrchestration, invoke)
						return
					}
					_, _ = boundary.ExecuteManagedParent(ctx, parentScope, "replace-a", parentIntent, allowOrchestration,
						func(parent *ManagedParentExecution) error {
							_, _, err := parent.ContinueOrExecuteManagedStartTarget(ctx, 7, "generation-a", invoke)
							return err
						})
				}()
				select {
				case <-done:
				case <-ctx.Done():
					t.Fatal("Start completion watchdog:", ctx.Err())
				}
				if helperErr != nil {
					t.Fatal(helperErr)
				}
				if stop == nil {
					t.Fatal("callback exited before pending Stop admission")
				}
				if err := stop.wait(ctx); err != ErrIndeterminateExecution {
					t.Fatalf("pending Stop error = %v", err)
				}
				if stop.callbacks.Load() != 0 {
					t.Fatal("lost callback authority invoked Stop lifecycle")
				}
				if !linked {
					calls := 0
					observation, err := boundary.ExecuteManagedStart(ctx, startScope, "start-a", intent, 7, "generation-a", allowOrchestration,
						func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
							calls++
							return TerminalOutcome{}, nil
						})
					if err != nil || observation.Kind() != AdmissionInProgress || observation.Record().State() != CommandStateClaimed || calls != 0 {
						t.Fatalf("durable observation = %#v/%v, reissues=%d", observation, err, calls)
					}
				} else {
					ledger := boundary.storage.existingLedger(parentScope.instanceScope())
					parentID := commandIdentity{scope: parentScope, key: "replace-a"}
					phaseID, _ := newPhaseIdentity(parentID, PhaseStartTarget)
					ledger.mu.Lock()
					parent, phase := ledger.parents[parentID], ledger.phases[phaseID]
					claimed := parent != nil && phase != nil && parent.state == CommandStateClaimed && phase.state == CommandStateClaimed && ledger.liveParents[parentID] == nil && ledger.livePhases[phaseID] == nil
					ledger.mu.Unlock()
					if !claimed {
						t.Fatal("lost linked callback did not preserve unresolved parent/phase claims")
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
