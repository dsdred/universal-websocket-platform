package runtimecommandidempotency

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

func TestReplayFirstPrimitiveExistingIdentityHasZeroNewAuthority(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "replay-first-existing", OperationStart)
	intent := startIntent(t, 41)
	var decisions, providers, flows atomic.Int32
	decision := primitiveDecision(t, 7, &decisions)
	provider := countingProvider("generation-a", &providers)
	invoke := countingManagedStart(t, boundary, &flows)
	claimed, disposition, err := boundary.ExecuteReplayFirstManagedStart(
		context.Background(), scope, "command", intent, allowOrchestration,
		decision, neverRevalidate, provider, invoke,
	)
	if err != nil || disposition != ReplayFirstAdmitted || claimed.Kind() != AdmissionClaimed {
		t.Fatalf("claim = %#v/%v/%v", claimed, disposition, err)
	}
	replay, disposition, err := boundary.ExecuteReplayFirstManagedStart(
		context.Background(), scope, "command", intent, allowOrchestration,
		decision, neverRevalidate, provider, invoke,
	)
	if err != nil || disposition != ReplayFirstAdmitted || replay.Kind() != AdmissionReplay {
		t.Fatalf("replay = %#v/%v/%v", replay, disposition, err)
	}
	conflictIntent := startIntent(t, 42)
	if _, _, err := boundary.ExecuteReplayFirstManagedStart(
		context.Background(), scope, "command", conflictIntent, allowOrchestration,
		decision, neverRevalidate, provider, invoke,
	); !errors.Is(err, ErrCommandKeyConflict) {
		t.Fatalf("conflict error = %v", err)
	}
	if decisions.Load() != 1 || providers.Load() != 1 || flows.Load() != 1 {
		t.Fatalf("decision/provider/flow = %d/%d/%d", decisions.Load(), providers.Load(), flows.Load())
	}
}

func TestReplayFirstPrimitiveConcurrentAbsentHasOneProviderWinner(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "replay-first-race", OperationStart)
	intent := startIntent(t, 41)
	candidate, _ := NewExecutePrimitiveCandidate(7)
	ready := make(chan struct{})
	release := make(chan struct{})
	var decisions, providers, flows atomic.Int32
	decide := func(context.Context) (AbsentCandidate, error) {
		if decisions.Add(1) == 2 {
			close(ready)
		}
		<-release
		return candidate, nil
	}
	type result struct {
		admission Admission
		err       error
	}
	results := make(chan result, 2)
	for range 2 {
		go func() {
			admission, _, err := boundary.ExecuteReplayFirstManagedStart(
				context.Background(), scope, "command", intent, allowOrchestration,
				decide, neverRevalidate, countingProvider("generation-a", &providers),
				countingManagedStart(t, boundary, &flows),
			)
			results <- result{admission: admission, err: err}
		}()
	}
	<-ready
	close(release)
	claimed := 0
	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatal(result.err)
		}
		if result.admission.Kind() == AdmissionClaimed {
			claimed++
		}
	}
	if claimed != 1 || decisions.Load() != 2 || providers.Load() != 1 || flows.Load() != 1 {
		t.Fatalf("claimed/decision/provider/flow = %d/%d/%d/%d", claimed, decisions.Load(), providers.Load(), flows.Load())
	}
}

func TestReplayFirstPrimitiveStopDuringProviderWaitsAndConverges(t *testing.T) {
	boundary := newTestBoundary(t)
	startScope := testScope(t, "domain-a", "late-primitive-stop", OperationStart)
	stopScope := testScope(t, "domain-a", "late-primitive-stop", OperationStop)
	providerEntered := make(chan struct{})
	releaseProvider := make(chan struct{})
	flowEntered := make(chan struct{})
	stopEntered := make(chan struct{})
	startResult := make(chan executionResult, 1)
	go func() {
		admission, _, err := boundary.ExecuteReplayFirstManagedStart(
			context.Background(), startScope, "start", startIntent(t, 41), allowOrchestration,
			primitiveDecision(t, 7, new(atomic.Int32)), neverRevalidate,
			func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
				close(providerEntered)
				<-releaseProvider
				return "generation-a", nil
			},
			func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
				close(flowEntered)
				gate, gateErr := boundary.ResolveManagedStartEarly(binding, "attempt-a")
				if gateErr != nil {
					return TerminalOutcome{}, gateErr
				}
				if gate != GateStopConverged {
					return TerminalOutcome{}, errors.New("pending Stop did not converge")
				}
				return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
			},
		)
		startResult <- executionResult{admission: admission, err: err}
	}()
	<-providerEntered
	stopResult := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(
			context.Background(), stopScope, "stop", NewStopIntent(), allow,
			func() (TerminalOutcome, error) {
				close(stopEntered)
				return NewTerminalOutcome(OutcomeSucceeded, "attempt-a")
			},
		)
		stopResult <- executionResult{admission: admission, err: err}
	}()
	waitForCommandRecord(t, boundary, stopScope, "stop")
	select {
	case <-flowEntered:
		t.Fatal("Start Flow ran before generation provider returned")
	default:
	}
	select {
	case <-stopEntered:
		t.Fatal("pending Stop ran before the managed Owner gate")
	default:
	}
	close(releaseProvider)
	stop := <-stopResult
	start := <-startResult
	if stop.err != nil || stop.admission.Record().State() != CommandStateTerminal {
		t.Fatalf("Stop = %#v/%v", stop.admission, stop.err)
	}
	if start.err != nil || start.admission.Record().State() != CommandStateTerminal {
		t.Fatalf("Start = %#v/%v", start.admission, start.err)
	}
}

func TestReplayFirstSatisfiedClaimsThenRevalidatesExactFacts(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "satisfied", OperationStart)
	intent := startIntent(t, 41)
	candidate, _ := NewSatisfiedCandidate(7, "attempt-a", 41)
	var revalidations, providers, flows atomic.Int32
	admission, disposition, err := boundary.ExecuteReplayFirstManagedStart(
		context.Background(), scope, "command", intent, allowOrchestration,
		func(context.Context) (AbsentCandidate, error) { return candidate, nil },
		func(_ context.Context, got AbsentCandidate) (CandidateRevalidation, error) {
			revalidations.Add(1)
			if got.ExpectedAggregateRevision() != 7 || got.LaunchAttemptID() != "attempt-a" ||
				got.ConfigurationVersionID() != 41 {
				t.Fatalf("revalidation facts = %#v", got)
			}
			ledger := boundary.storage.existingLedger(scope.instanceScope())
			ledger.mu.Lock()
			record := ledger.records[commandIdentity{scope: scope, key: "command"}]
			claimed := record != nil && record.state == CommandStateClaimed
			ledger.mu.Unlock()
			if !claimed {
				t.Fatal("revalidation ran before claim")
			}
			return CandidateRevalidated, nil
		}, countingProvider("generation-a", &providers), countingManagedStart(t, boundary, &flows),
	)
	if err != nil || disposition != ReplayFirstAdmitted || admission.Record().State() != CommandStateTerminal {
		t.Fatalf("satisfied = %#v/%v/%v", admission, disposition, err)
	}
	outcome, ok := admission.Record().Outcome()
	if !ok || outcome.LaunchAttemptID() != "attempt-a" || revalidations.Load() != 1 ||
		providers.Load() != 0 || flows.Load() != 0 {
		t.Fatalf("outcome/calls = %#v/%v %d/%d/%d", outcome, ok, revalidations.Load(), providers.Load(), flows.Load())
	}
}

func TestReplayFirstSatisfiedUnresolvedFactsRemainClaimed(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "satisfied-stale", OperationStart)
	intent := startIntent(t, 41)
	candidate, _ := NewSatisfiedCandidate(7, "attempt-a", 41)
	var decisions, revalidations atomic.Int32
	decide := func(context.Context) (AbsentCandidate, error) {
		decisions.Add(1)
		return candidate, nil
	}
	admission, _, err := boundary.ExecuteReplayFirstManagedStart(
		context.Background(), scope, "command", intent, allowOrchestration, decide,
		func(context.Context, AbsentCandidate) (CandidateRevalidation, error) {
			revalidations.Add(1)
			return CandidateUnresolved, nil
		}, countingProvider("generation-a", new(atomic.Int32)), countingManagedStart(t, boundary, new(atomic.Int32)),
	)
	if !errors.Is(err, ErrIndeterminateExecution) || admission.Record().State() != CommandStateClaimed {
		t.Fatalf("stale satisfied = %#v/%v", admission, err)
	}
	replay, _, err := boundary.ExecuteReplayFirstManagedStart(
		context.Background(), scope, "command", intent, allowOrchestration, decide,
		func(context.Context, AbsentCandidate) (CandidateRevalidation, error) {
			t.Fatal("in-progress revalidated")
			return 0, nil
		}, countingProvider("generation-b", new(atomic.Int32)), countingManagedStart(t, boundary, new(atomic.Int32)),
	)
	if err != nil || replay.Kind() != AdmissionInProgress || decisions.Load() != 1 || revalidations.Load() != 1 {
		t.Fatalf("replay = %#v/%v calls=%d/%d", replay, err, decisions.Load(), revalidations.Load())
	}
}

func TestReplayFirstPrimitiveProviderFailuresAreClaimedAndNotRetried(t *testing.T) {
	cases := map[string]func(context.Context, *Boundary) (runtimeorchestrationbinding.ExecutionGeneration, error){
		"error": func(context.Context, *Boundary) (runtimeorchestrationbinding.ExecutionGeneration, error) {
			return "", errors.New("provider failed")
		},
		"empty": func(context.Context, *Boundary) (runtimeorchestrationbinding.ExecutionGeneration, error) {
			return "", nil
		},
		"panic": func(context.Context, *Boundary) (runtimeorchestrationbinding.ExecutionGeneration, error) {
			panic("provider")
		},
		"post-win-cancellation": func(ctx context.Context, _ *Boundary) (runtimeorchestrationbinding.ExecutionGeneration, error) {
			ctx.(interface{ cancel() }).cancel()
			return "generation-a", nil
		},
		"generation-loss": func(_ context.Context, boundary *Boundary) (runtimeorchestrationbinding.ExecutionGeneration, error) {
			_, _ = NewBoundary(boundary.storage)
			return "generation-a", nil
		},
	}
	for name, provide := range cases {
		t.Run(name, func(t *testing.T) {
			boundary := newTestBoundary(t)
			scope := testScope(t, "domain-a", "provider-"+name, OperationStart)
			ctx := context.Background()
			if name == "post-win-cancellation" {
				value := &cancelContext{Context: context.Background()}
				ctx = value
			}
			var providers, flows atomic.Int32
			provider := func(ctx context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
				providers.Add(1)
				return provide(ctx, boundary)
			}
			admission, _, err := boundary.ExecuteReplayFirstManagedStart(
				ctx, scope, "command", startIntent(t, 41), allowOrchestration,
				primitiveDecision(t, 7, new(atomic.Int32)), neverRevalidate, provider,
				countingManagedStart(t, boundary, &flows),
			)
			if err == nil || admission.Kind() != AdmissionClaimed ||
				admission.Record().State() != CommandStateClaimed || providers.Load() != 1 || flows.Load() != 0 {
				t.Fatalf("failure = %#v/%v providers=%d flows=%d", admission, err, providers.Load(), flows.Load())
			}
			ledger := boundary.storage.existingLedger(scope.instanceScope())
			ledger.mu.Lock()
			managed := len(ledger.managedStart)
			ledger.mu.Unlock()
			if managed != 0 {
				t.Fatalf("provider failure retained %d managed rendezvous", managed)
			}
			active := boundary
			if name == "generation-loss" {
				active, _ = NewBoundary(boundary.storage)
			}
			retry, _, retryErr := active.ExecuteReplayFirstManagedStart(
				context.Background(), scope, "command", startIntent(t, 41), allowOrchestration,
				func(context.Context) (AbsentCandidate, error) {
					t.Fatal("retry decided")
					return AbsentCandidate{}, nil
				},
				neverRevalidate,
				func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
					t.Fatal("retry provided generation")
					return "", nil
				}, func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
					t.Fatal("retry invoked flow")
					return TerminalOutcome{}, nil
				},
			)
			if retryErr != nil || retry.Kind() != AdmissionInProgress || providers.Load() != 1 {
				t.Fatalf("retry = %#v/%v providers=%d", retry, retryErr, providers.Load())
			}
		})
	}
}

func TestReplayFirstPrimitiveProviderGoexitLeavesClaimed(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "provider-goexit", OperationStart)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _, _ = boundary.ExecuteReplayFirstManagedStart(
			context.Background(), scope, "command", startIntent(t, 41), allowOrchestration,
			primitiveDecision(t, 7, new(atomic.Int32)), neverRevalidate,
			func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
				runtime.Goexit()
				return "", nil
			}, func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
				t.Error("Flow ran after provider Goexit")
				return TerminalOutcome{}, nil
			},
		)
	}()
	<-done
	ledger := boundary.storage.existingLedger(scope.instanceScope())
	ledger.mu.Lock()
	record := ledger.records[commandIdentity{scope: scope, key: "command"}]
	_, live := ledger.live[commandIdentity{scope: scope, key: "command"}]
	managed := len(ledger.managedStart)
	ledger.mu.Unlock()
	if record == nil || record.state != CommandStateClaimed || live || managed != 0 {
		t.Fatalf("record/live/managed = %#v/%v/%d", record, live, managed)
	}
}

func TestReplayFirstNoClaimAndPreclaimCancellationHaveZeroAuthority(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "no-claim", OperationStart)
	var decisions, providers, flows atomic.Int32
	_, disposition, err := boundary.ExecuteReplayFirstManagedStart(
		context.Background(), scope, "command", startIntent(t, 41), allowOrchestration,
		func(context.Context) (AbsentCandidate, error) { decisions.Add(1); return NewNoClaimCandidate(), nil },
		neverRevalidate, countingProvider("generation-a", &providers), countingManagedStart(t, boundary, &flows),
	)
	if err != nil || disposition != ReplayFirstNoClaim || decisions.Load() != 1 || providers.Load() != 0 || flows.Load() != 0 {
		t.Fatalf("no claim = %v/%v calls=%d/%d/%d", disposition, err, decisions.Load(), providers.Load(), flows.Load())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := boundary.ExecuteReplayFirstManagedStart(
		ctx, scope, "cancelled", startIntent(t, 41), allowOrchestration,
		func(context.Context) (AbsentCandidate, error) {
			t.Fatal("cancelled submission decided")
			return AbsentCandidate{}, nil
		},
		neverRevalidate, countingProvider("generation-a", &providers), countingManagedStart(t, boundary, &flows),
	); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel error = %v", err)
	}
}

func TestReplayFirstAbsentDecisionFailuresMutateNothing(t *testing.T) {
	for name, decide := range map[string]DecideAbsent{
		"error": func(context.Context) (AbsentCandidate, error) {
			return AbsentCandidate{}, errors.New("decision failed")
		},
		"panic":   func(context.Context) (AbsentCandidate, error) { panic("decision") },
		"invalid": func(context.Context) (AbsentCandidate, error) { return AbsentCandidate{}, nil },
	} {
		t.Run(name, func(t *testing.T) {
			boundary := newTestBoundary(t)
			scope := testScope(t, "domain-a", "decision-"+name, OperationStart)
			var providers, flows atomic.Int32
			if _, _, err := boundary.ExecuteReplayFirstManagedStart(
				context.Background(), scope, "command", startIntent(t, 41), allowOrchestration,
				decide, neverRevalidate, countingProvider("generation-a", &providers),
				countingManagedStart(t, boundary, &flows),
			); err == nil {
				t.Fatal("decision failure returned nil")
			}
			if ledger := boundary.storage.existingLedger(scope.instanceScope()); ledger != nil {
				ledger.mu.Lock()
				records := len(ledger.records)
				ledger.mu.Unlock()
				if records != 0 {
					t.Fatalf("decision failure created %d records", records)
				}
			}
			if providers.Load() != 0 || flows.Load() != 0 {
				t.Fatalf("provider/flow = %d/%d", providers.Load(), flows.Load())
			}
		})
	}
}

func TestReplayFirstExistingIdentityStillRequiresAuthorizationAndCancellation(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "existing-gates", OperationStart)
	intent := startIntent(t, 41)
	_, _, err := boundary.ExecuteReplayFirstManagedStart(
		context.Background(), scope, "command", intent, allowOrchestration,
		primitiveDecision(t, 7, new(atomic.Int32)), neverRevalidate,
		countingProvider("generation-a", new(atomic.Int32)), countingManagedStart(t, boundary, new(atomic.Int32)),
	)
	if err != nil {
		t.Fatal(err)
	}
	denied := errors.New("denied")
	if _, _, err := boundary.ExecuteReplayFirstManagedStart(
		context.Background(), scope, "command", intent,
		func(context.Context, OrchestrationAuthorizationRequest) error { return denied },
		func(context.Context) (AbsentCandidate, error) {
			t.Fatal("denied replay decided")
			return AbsentCandidate{}, nil
		},
		neverRevalidate, countingProvider("generation-b", new(atomic.Int32)), countingManagedStart(t, boundary, new(atomic.Int32)),
	); !errors.Is(err, denied) {
		t.Fatalf("authorization error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := boundary.ExecuteReplayFirstManagedStart(
		ctx, scope, "command", intent, allowOrchestration,
		func(context.Context) (AbsentCandidate, error) {
			t.Fatal("cancelled replay decided")
			return AbsentCandidate{}, nil
		},
		neverRevalidate, countingProvider("generation-b", new(atomic.Int32)), countingManagedStart(t, boundary, new(atomic.Int32)),
	); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}
}

func TestReplayFirstParentSatisfiedClaimsBeforeExactRevalidation(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "parent-satisfied", OperationRollback)
	intent := rollbackIntent(t, 40)
	candidate, _ := NewSatisfiedCandidate(12, "attempt-old", 40)
	var revalidations, providers atomic.Int32
	admission, disposition, err := boundary.ExecuteReplayFirstManagedParent(
		context.Background(), scope, "parent", intent, allowOrchestration,
		func(context.Context) (AbsentCandidate, error) { return candidate, nil },
		func(_ context.Context, got AbsentCandidate) (CandidateRevalidation, error) {
			revalidations.Add(1)
			ledger := boundary.storage.existingLedger(scope.instanceScope())
			ledger.mu.Lock()
			record := ledger.parents[commandIdentity{scope: scope, key: "parent"}]
			claimed := record != nil && record.state == CommandStateClaimed
			ledger.mu.Unlock()
			if !claimed || got.ExpectedAggregateRevision() != 12 || got.LaunchAttemptID() != "attempt-old" {
				t.Fatalf("revalidation candidate = %#v claimed=%v", got, claimed)
			}
			return CandidateRevalidated, nil
		}, countingProvider("generation", &providers),
		func(*ReplayFirstManagedParentExecution) error {
			t.Fatal("satisfied parent received execution")
			return nil
		},
	)
	outcome, ok := admission.Record().Outcome()
	if err != nil || disposition != ReplayFirstAdmitted || admission.Record().State() != CommandStateTerminal ||
		!ok || outcome.Category() != ParentOutcomeSatisfied || revalidations.Load() != 1 || providers.Load() != 0 {
		t.Fatalf("satisfied parent = %#v/%v/%v outcome=%#v/%v calls=%d/%d",
			admission, disposition, err, outcome, ok, revalidations.Load(), providers.Load())
	}
}

func TestReplayFirstManagedParentUsesCapturedRevisionAndLateProvider(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "parent-late", OperationReplace)
	intent := replaceIntent(t, 42)
	candidate, _ := NewExecuteParentCandidate(9)
	var decisions, providers, flows atomic.Int32
	decision := func(context.Context) (AbsentCandidate, error) {
		decisions.Add(1)
		return candidate, nil
	}
	provider := countingProvider("generation-parent", &providers)
	var retained *ReplayFirstManagedParentExecution
	admission, disposition, err := boundary.ExecuteReplayFirstManagedParent(
		context.Background(), scope, "parent", intent, allowOrchestration,
		decision, neverRevalidate, provider,
		func(execution *ReplayFirstManagedParentExecution) error {
			retained = execution
			phase, prevented, phaseErr := execution.ContinueOrExecuteManagedStartTarget(
				context.Background(), func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
					flows.Add(1)
					if binding.ExpectedAggregateRevision() != 9 || binding.ExecutionGeneration() != "generation-parent" {
						t.Fatalf("binding = %#v", binding)
					}
					return completeManagedStart(t, boundary, binding)
				},
			)
			if phaseErr != nil || prevented || phase.Kind() != AdmissionClaimed {
				return phaseErr
			}
			_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeSucceeded))
			return publishErr
		},
	)
	if err != nil || disposition != ReplayFirstAdmitted || admission.Record().State() != CommandStateTerminal ||
		decisions.Load() != 1 || providers.Load() != 1 || flows.Load() != 1 {
		t.Fatalf("parent = %#v/%v/%v calls=%d/%d/%d", admission, disposition, err,
			decisions.Load(), providers.Load(), flows.Load())
	}
	if _, _, err := retained.ContinueOrExecuteManagedStartTarget(
		context.Background(), func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
			t.Fatal("retained execution invoked Flow")
			return TerminalOutcome{}, nil
		},
	); !errors.Is(err, ErrBoundaryExpired) || providers.Load() != 1 {
		t.Fatalf("retained execution = %v providers=%d", err, providers.Load())
	}
	replay, replayDisposition, replayErr := boundary.ExecuteReplayFirstManagedParent(
		context.Background(), scope, "parent", intent, allowOrchestration,
		decision, neverRevalidate, provider,
		func(*ReplayFirstManagedParentExecution) error {
			t.Fatal("replay invoked parent Flow")
			return nil
		},
	)
	if replayErr != nil || replayDisposition != ReplayFirstAdmitted || replay.Kind() != AdmissionReplay {
		t.Fatalf("parent replay = %#v/%v/%v", replay, replayDisposition, replayErr)
	}
	if _, _, conflictErr := boundary.ExecuteReplayFirstManagedParent(
		context.Background(), scope, "parent", replaceIntent(t, 43), allowOrchestration,
		decision, neverRevalidate, provider,
		func(*ReplayFirstManagedParentExecution) error {
			t.Fatal("conflict invoked parent Flow")
			return nil
		},
	); !errors.Is(conflictErr, ErrCommandKeyConflict) {
		t.Fatalf("parent conflict error = %v", conflictErr)
	}
	if decisions.Load() != 1 || providers.Load() != 1 || flows.Load() != 1 {
		t.Fatalf("parent replay authority calls = %d/%d/%d", decisions.Load(), providers.Load(), flows.Load())
	}
}

func TestReplayFirstParentStopDuringStartTargetProviderWaitsAndConverges(t *testing.T) {
	boundary := newTestBoundary(t)
	parentScope := testScope(t, "domain-a", "late-parent-stop", OperationReplace)
	stopScope := testScope(t, "domain-a", "late-parent-stop", OperationStop)
	candidate, _ := NewExecuteParentCandidate(9)
	providerEntered := make(chan struct{})
	releaseProvider := make(chan struct{})
	flowEntered := make(chan struct{})
	stopEntered := make(chan struct{})
	type parentResult struct {
		admission ParentAdmission
		err       error
	}
	parentDone := make(chan parentResult, 1)
	go func() {
		admission, _, err := boundary.ExecuteReplayFirstManagedParent(
			context.Background(), parentScope, "parent", replaceIntent(t, 42), allowOrchestration,
			func(context.Context) (AbsentCandidate, error) { return candidate, nil }, neverRevalidate,
			func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
				close(providerEntered)
				<-releaseProvider
				return "generation-parent", nil
			},
			func(execution *ReplayFirstManagedParentExecution) error {
				phase, prevented, phaseErr := execution.ContinueOrExecuteManagedStartTarget(
					context.Background(), func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
						close(flowEntered)
						gate, gateErr := boundary.ResolveManagedStartEarly(binding, "attempt-parent")
						if gateErr != nil {
							return TerminalOutcome{}, gateErr
						}
						if gate != GateStopConverged {
							return TerminalOutcome{}, errors.New("pending parent Stop did not converge")
						}
						return NewTerminalOutcome(OutcomeSucceeded, "attempt-parent")
					},
				)
				if phaseErr != nil {
					return phaseErr
				}
				if prevented || phase.Kind() != AdmissionClaimed {
					return errors.New("StartTarget was not newly claimed")
				}
				_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeSucceeded))
				return publishErr
			},
		)
		parentDone <- parentResult{admission: admission, err: err}
	}()
	<-providerEntered
	stopResult := make(chan executionResult, 1)
	go func() {
		admission, err := boundary.Execute(
			context.Background(), stopScope, "stop", NewStopIntent(), allow,
			func() (TerminalOutcome, error) {
				close(stopEntered)
				return NewTerminalOutcome(OutcomeSucceeded, "attempt-parent")
			},
		)
		stopResult <- executionResult{admission: admission, err: err}
	}()
	waitForCommandRecord(t, boundary, stopScope, "stop")
	select {
	case <-flowEntered:
		t.Fatal("StartTarget Flow ran before generation provider returned")
	default:
	}
	select {
	case <-stopEntered:
		t.Fatal("pending parent Stop ran before the managed Owner gate")
	default:
	}
	close(releaseProvider)
	stop := <-stopResult
	parent := <-parentDone
	if stop.err != nil || stop.admission.Record().State() != CommandStateTerminal {
		t.Fatalf("Stop = %#v/%v", stop.admission, stop.err)
	}
	if parent.err != nil || parent.admission.Record().State() != CommandStateTerminal {
		t.Fatalf("parent = %#v/%v", parent.admission, parent.err)
	}
}

func TestReplayFirstTrackedParentRechecksExactStartAndPreclaimsStopOld(t *testing.T) {
	boundary := newTestBoundary(t)
	startScope, releaseStart, startDone := holdTrackedStart(t, boundary, "tracked-late")
	parentScope := testScope(t, startScope.Domain(), "tracked-late", OperationReplace)
	candidate, _ := NewExecuteParentFromTrackedStartCandidate(11, startScope, "tracked-start", 1)
	var providers atomic.Int32
	admission, _, err := boundary.ExecuteReplayFirstManagedParent(
		context.Background(), parentScope, "parent", replaceIntent(t, 42), allowOrchestration,
		func(context.Context) (AbsentCandidate, error) { return candidate, nil }, neverRevalidate,
		countingProvider("generation-parent", &providers),
		func(execution *ReplayFirstManagedParentExecution) error {
			stop, stopErr := execution.ExecutePreclaimedStopOld(func() (TerminalOutcome, error) {
				return terminalOutcome(t, OutcomeSucceeded, "old-attempt"), nil
			})
			if stopErr != nil || stop.Kind() != AdmissionClaimed {
				return stopErr
			}
			phase, prevented, phaseErr := execution.ContinueOrExecuteManagedStartTarget(
				context.Background(), func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
					if binding.ExpectedAggregateRevision() != 11 {
						t.Fatalf("captured revision = %d", binding.ExpectedAggregateRevision())
					}
					return completeManagedStart(t, boundary, binding)
				},
			)
			if phaseErr != nil || prevented || phase.Kind() != AdmissionClaimed {
				return phaseErr
			}
			_, publishErr := execution.PublishTerminal(parentOutcome(t, ParentOutcomeSucceeded))
			return publishErr
		},
	)
	if err != nil || admission.Record().State() != CommandStateTerminal || providers.Load() != 1 {
		t.Fatalf("tracked parent = %#v/%v providers=%d", admission, err, providers.Load())
	}
	close(releaseStart)
	if err := <-startDone; err != nil {
		t.Fatal(err)
	}
}

func TestReplayFirstTrackedParentRejectsStaleStartCandidateAtomically(t *testing.T) {
	boundary := newTestBoundary(t)
	startScope, releaseStart, startDone := holdTrackedStart(t, boundary, "tracked-stale")
	parentScope := testScope(t, startScope.Domain(), "tracked-stale", OperationRollback)
	candidate, _ := NewExecuteParentFromTrackedStartCandidate(7, startScope, "wrong-start", 1)
	var providers atomic.Int32
	admission, _, err := boundary.ExecuteReplayFirstManagedParent(
		context.Background(), parentScope, "parent", rollbackIntent(t, 40), allowOrchestration,
		func(context.Context) (AbsentCandidate, error) { return candidate, nil }, neverRevalidate,
		countingProvider("generation-parent", &providers),
		func(*ReplayFirstManagedParentExecution) error {
			t.Fatal("stale candidate received authority")
			return nil
		},
	)
	if !errors.Is(err, ErrInstanceBlocked) || admission != (ParentAdmission{}) || providers.Load() != 0 {
		t.Fatalf("stale candidate = %#v/%v providers=%d", admission, err, providers.Load())
	}
	ledger := boundary.storage.existingLedger(parentScope.instanceScope())
	ledger.mu.Lock()
	_, exists := ledger.parents[commandIdentity{scope: parentScope, key: "parent"}]
	ledger.mu.Unlock()
	if exists {
		t.Fatal("stale candidate created parent")
	}
	close(releaseStart)
	if err := <-startDone; err != nil {
		t.Fatal(err)
	}
}

func TestReplayFirstParentProviderFailureLeavesPhaseAndParentClaimed(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "parent-provider-failure", OperationReplace)
	candidate, _ := NewExecuteParentCandidate(7)
	var decisions, providers, flows atomic.Int32
	admission, _, err := boundary.ExecuteReplayFirstManagedParent(
		context.Background(), scope, "parent", replaceIntent(t, 42), allowOrchestration,
		func(context.Context) (AbsentCandidate, error) { decisions.Add(1); return candidate, nil }, neverRevalidate,
		func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
			providers.Add(1)
			return "", errors.New("provider failed")
		}, func(execution *ReplayFirstManagedParentExecution) error {
			phase, _, phaseErr := execution.ContinueOrExecuteManagedStartTarget(
				context.Background(), func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
					flows.Add(1)
					return TerminalOutcome{}, nil
				},
			)
			if !errors.Is(phaseErr, ErrIndeterminateExecution) || phase.Kind() != AdmissionClaimed ||
				phase.Record().State() != CommandStateClaimed {
				t.Fatalf("phase = %#v/%v", phase, phaseErr)
			}
			return phaseErr
		},
	)
	if !errors.Is(err, ErrIndeterminateExecution) || admission.Record().State() != CommandStateClaimed ||
		decisions.Load() != 1 || providers.Load() != 1 || flows.Load() != 0 {
		t.Fatalf("parent = %#v/%v calls=%d/%d/%d", admission, err, decisions.Load(), providers.Load(), flows.Load())
	}
	ledger := boundary.storage.existingLedger(scope.instanceScope())
	ledger.mu.Lock()
	managed := len(ledger.managedStart)
	ledger.mu.Unlock()
	if managed != 0 {
		t.Fatalf("parent provider failure retained %d managed rendezvous", managed)
	}
	replay, _, replayErr := boundary.ExecuteReplayFirstManagedParent(
		context.Background(), scope, "parent", replaceIntent(t, 42), allowOrchestration,
		func(context.Context) (AbsentCandidate, error) {
			t.Fatal("replay decided")
			return AbsentCandidate{}, nil
		},
		neverRevalidate, countingProvider("generation-b", &providers),
		func(*ReplayFirstManagedParentExecution) error { t.Fatal("replay received authority"); return nil },
	)
	if replayErr != nil || replay.Kind() != AdmissionInProgress || providers.Load() != 1 {
		t.Fatalf("replay = %#v/%v providers=%d", replay, replayErr, providers.Load())
	}
}

func TestReplayFirstDifferentInstancesProgressIndependently(t *testing.T) {
	boundary := newTestBoundary(t)
	blocked := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	scopeA := testScope(t, "domain-a", "instance-a", OperationStart)
	go func() {
		_, _, err := boundary.ExecuteReplayFirstManagedStart(
			context.Background(), scopeA, "command", startIntent(t, 41), allowOrchestration,
			primitiveDecision(t, 7, new(atomic.Int32)), neverRevalidate,
			func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
				close(blocked)
				<-release
				return "generation-a", nil
			}, countingManagedStart(t, boundary, new(atomic.Int32)),
		)
		done <- err
	}()
	<-blocked
	scopeB := testScope(t, "domain-a", "instance-b", OperationStart)
	admission, _, err := boundary.ExecuteReplayFirstManagedStart(
		context.Background(), scopeB, "command", startIntent(t, 42), allowOrchestration,
		primitiveDecision(t, 8, new(atomic.Int32)), neverRevalidate,
		countingProvider("generation-b", new(atomic.Int32)), countingManagedStart(t, boundary, new(atomic.Int32)),
	)
	if err != nil || admission.Record().State() != CommandStateTerminal {
		t.Fatalf("instance B = %#v/%v", admission, err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

type cancelContext struct {
	context.Context
	cancelled atomic.Bool
}

func (c *cancelContext) cancel() { c.cancelled.Store(true) }

func (c *cancelContext) Err() error {
	if c.cancelled.Load() {
		return context.Canceled
	}
	return c.Context.Err()
}

func primitiveDecision(t *testing.T, revision runtimeorchestrationbinding.AggregateRevision, calls *atomic.Int32) DecideAbsent {
	t.Helper()
	candidate, err := NewExecutePrimitiveCandidate(revision)
	if err != nil {
		t.Fatal(err)
	}
	return func(context.Context) (AbsentCandidate, error) {
		calls.Add(1)
		return candidate, nil
	}
}

func countingProvider(generation runtimeorchestrationbinding.ExecutionGeneration, calls *atomic.Int32) ProvideExecutionGeneration {
	return func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
		calls.Add(1)
		return generation, nil
	}
}

func countingManagedStart(t *testing.T, boundary *Boundary, calls *atomic.Int32) func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
	t.Helper()
	return func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
		calls.Add(1)
		return completeManagedStart(t, boundary, binding)
	}
}

func neverRevalidate(context.Context, AbsentCandidate) (CandidateRevalidation, error) {
	return CandidateUnresolved, errors.New("unexpected revalidation")
}
