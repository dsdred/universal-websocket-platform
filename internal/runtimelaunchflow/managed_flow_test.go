package runtimelaunchflow

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/configurationloader"
	"github.com/dsdred/universal-websocket-platform/internal/runtime"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

const (
	testWorkspaceID   uint64 = 11
	testConfiguration uint64 = 22
	testVersionID     uint64 = 33
	testInstanceID           = runtimeconfigload.RuntimeInstanceID("instance-a")
	testAttemptID            = runtimeconfigload.LaunchAttemptID("attempt-a")
	testGeneration           = runtimeorchestrationbinding.ExecutionGeneration("gen-a")
)

func validAuthorization(t *testing.T) runtimeorchestrationbinding.OrchestrationAuthorizationRequest {
	return validAuthorizationFor(t, testInstanceID)
}

func validAuthorizationFor(
	t *testing.T,
	instanceID runtimeconfigload.RuntimeInstanceID,
) runtimeorchestrationbinding.OrchestrationAuthorizationRequest {
	t.Helper()
	request, err := runtimeorchestrationbinding.NewOrchestrationAuthorizationRequest(
		"domain-a", testWorkspaceID, testConfiguration, instanceID,
		runtimeorchestrationbinding.OrchestrationActionActivateExactTarget, testVersionID,
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func validStartRendezvous(t *testing.T) runtimeorchestrationbinding.StartRendezvous {
	t.Helper()
	rendezvous, err := runtimeorchestrationbinding.NewStartRendezvous("rendezvous-a")
	if err != nil {
		t.Fatal(err)
	}
	return rendezvous
}

func testBinding(t *testing.T) ManagedStartBinding {
	t.Helper()
	binding, err := NewManagedStartBinding(
		validAuthorization(t), runtimeorchestrationbinding.AggregateRevision(1),
		testGeneration,
		validStartRendezvous(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return binding
}

type recordingContinuation struct {
	calls        atomic.Int32
	noClaimCalls atomic.Int32
	lastNoClaim  runtimeorchestrationbinding.StartNoClaimCause
	lastErr      error
	outcome      StartClaimOutcome
	lastBinding  ManagedStartBinding
	lastView     OwnerClaimView
	lastContext  context.Context
}

func (r *recordingContinuation) StartNoClaim(
	_ context.Context,
	_ ManagedStartBinding,
	cause runtimeorchestrationbinding.StartNoClaimCause,
) error {
	r.noClaimCalls.Add(1)
	r.lastNoClaim = cause
	return r.lastErr
}

func (r *recordingContinuation) AfterOwnerClaim(
	ctx context.Context,
	binding ManagedStartBinding,
	view OwnerClaimView,
) (StartClaimOutcome, error) {
	r.calls.Add(1)
	r.lastBinding = binding
	r.lastView = view
	r.lastContext = ctx
	outcome := r.outcome
	if outcome == 0 {
		outcome = StartClaimContinue
	}
	if r.lastErr != nil && r.outcome == 0 {
		outcome = StartClaimBlocked
	}
	return outcome, r.lastErr
}

type adversarialContinuation struct {
	outcome      StartClaimOutcome
	err          error
	panicAfter   bool
	panicNoClaim bool
}

type stoppingContinuation struct{ owner *runtimelifecycle.Owner }
type cancellationProbeContinuation struct {
	entered  chan struct{}
	release  chan struct{}
	observed error
}

func (c *cancellationProbeContinuation) StartNoClaim(context.Context, ManagedStartBinding, runtimeorchestrationbinding.StartNoClaimCause) error {
	return nil
}
func (c *cancellationProbeContinuation) AfterOwnerClaim(ctx context.Context, _ ManagedStartBinding, _ OwnerClaimView) (StartClaimOutcome, error) {
	close(c.entered)
	<-c.release
	c.observed = ctx.Err()
	return StartClaimBlocked, errors.New("blocked")
}

func (s stoppingContinuation) StartNoClaim(context.Context, ManagedStartBinding, runtimeorchestrationbinding.StartNoClaimCause) error {
	return nil
}
func (s stoppingContinuation) AfterOwnerClaim(context.Context, ManagedStartBinding, OwnerClaimView) (StartClaimOutcome, error) {
	if _, err := s.owner.Stop(context.Background()); err != nil {
		return StartClaimBlocked, err
	}
	return StartClaimStopConverged, nil
}

func (c adversarialContinuation) StartNoClaim(
	context.Context, ManagedStartBinding, runtimeorchestrationbinding.StartNoClaimCause,
) error {
	if c.panicNoClaim {
		panic("no claim")
	}
	return c.err
}

func (c adversarialContinuation) AfterOwnerClaim(
	context.Context, ManagedStartBinding, OwnerClaimView,
) (StartClaimOutcome, error) {
	if c.panicAfter {
		panic("after owner claim")
	}
	return c.outcome, c.err
}

func TestManagedStartBindingIsImmutableAndValidated(t *testing.T) {
	rendezvous := validStartRendezvous(t)

	valid, err := NewManagedStartBinding(
		validAuthorization(t), runtimeorchestrationbinding.AggregateRevision(1), testGeneration, rendezvous,
	)
	if err != nil {
		t.Fatal(err)
	}
	if valid.ExpectedAggregateRevision() != 1 || valid.ExecutionGeneration() != testGeneration {
		t.Fatalf("unexpected binding contents: %#v", valid)
	}
	_ = valid.Rendezvous()

	if _, err := NewManagedStartBinding(validAuthorization(t), 0, testGeneration, rendezvous); !errors.Is(err, ErrInvalidManagedBinding) {
		t.Fatalf("zero revision must fail: %v", err)
	}
	if _, err := NewManagedStartBinding(validAuthorization(t), 1, "", rendezvous); !errors.Is(err, ErrInvalidManagedBinding) {
		t.Fatalf("empty generation must fail: %v", err)
	}
	if _, err := NewManagedStartBinding(validAuthorization(t), 1, testGeneration, runtimeorchestrationbinding.StartRendezvous{}); !errors.Is(err, ErrInvalidManagedBinding) {
		t.Fatalf("zero rendezvous must fail: %v", err)
	}
}

func TestOwnerClaimViewIsImmutableAndValidated(t *testing.T) {
	request := runtimeconfigload.NewLoadRequest(
		testWorkspaceID, testConfiguration, testVersionID, testInstanceID, testAttemptID,
	)
	view, err := NewOwnerClaimView(request)
	if err != nil {
		t.Fatal(err)
	}
	if view.WorkspaceID() != testWorkspaceID ||
		view.ConfigurationID() != testConfiguration ||
		view.RuntimeInstanceID() != testInstanceID ||
		view.LaunchAttemptID() != testAttemptID ||
		view.TargetConfigurationVersionID() != testVersionID {
		t.Fatalf("unexpected view contents: %#v", view)
	}

	incomplete := runtimeconfigload.NewLoadRequest(
		0, testConfiguration, testVersionID, testInstanceID, testAttemptID,
	)
	if _, err := NewOwnerClaimView(incomplete); !errors.Is(err, ErrInvalidOwnerClaimView) {
		t.Fatalf("zero workspace must fail: %v", err)
	}
	incompleteConfig := runtimeconfigload.NewLoadRequest(
		testWorkspaceID, 0, testVersionID, testInstanceID, testAttemptID,
	)
	if _, err := NewOwnerClaimView(incompleteConfig); !errors.Is(err, ErrInvalidOwnerClaimView) {
		t.Fatalf("zero configuration must fail: %v", err)
	}
	incompleteInstance := runtimeconfigload.NewLoadRequest(
		testWorkspaceID, testConfiguration, testVersionID, "", testAttemptID,
	)
	if _, err := NewOwnerClaimView(incompleteInstance); !errors.Is(err, ErrInvalidOwnerClaimView) {
		t.Fatalf("empty instance must fail: %v", err)
	}
	incompleteAttempt := runtimeconfigload.NewLoadRequest(
		testWorkspaceID, testConfiguration, testVersionID, testInstanceID, "",
	)
	if _, err := NewOwnerClaimView(incompleteAttempt); !errors.Is(err, ErrInvalidOwnerClaimView) {
		t.Fatalf("empty attempt must fail: %v", err)
	}
	incompleteVersion := runtimeconfigload.NewLoadRequest(
		testWorkspaceID, testConfiguration, 0, testInstanceID, testAttemptID,
	)
	if _, err := NewOwnerClaimView(incompleteVersion); !errors.Is(err, ErrInvalidOwnerClaimView) {
		t.Fatalf("zero version must fail: %v", err)
	}
}

func TestManagedFlowConstructionBindsExactlyOneContinuation(t *testing.T) {
	managed, err := NewManaged(
		mustManagedOwner(t), mustManagedLoader(t), &recordingContinuation{},
	)
	if err != nil {
		t.Fatal(err)
	}
	_ = managed
}

func TestManagedFlowConstructionRejectsNilContinuation(t *testing.T) {
	if _, err := NewManaged(mustManagedOwner(t), mustManagedLoader(t), nil); !errors.Is(err, ErrInvalidContinuation) {
		t.Fatalf("nil continuation must fail: %v", err)
	}
}

func TestStartManagedValidatesBindingBeforeOwnerMutation(t *testing.T) {
	continuation := &recordingContinuation{}
	managed := mustManagedFlow(t, continuation)
	zero := ManagedStartBinding{}
	if _, err := managed.StartManaged(
		context.Background(),
		runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
		zero,
	); !errors.Is(err, ErrInvalidManagedBinding) {
		t.Fatalf("invalid binding must fail closed: %v", err)
	}
	if continuation.calls.Load() != 0 {
		t.Fatalf("continuation ran %d times on invalid binding, want 0", continuation.calls.Load())
	}
}

func TestStartManagedInvokesContinuationOnceAfterPrepareStartAndBeforeLoad(t *testing.T) {
	continuation := &recordingContinuation{}
	managed := mustManagedFlow(t, continuation)

	outcome, err := managed.StartManaged(
		context.Background(),
		runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
		testBinding(t),
	)
	if err != nil || outcome.Kind() != runtimelifecycle.StartRunning {
		t.Fatalf("managed start = %q/%v", outcome.Kind(), err)
	}
	if continuation.calls.Load() != 1 {
		t.Fatalf("continuation calls = %d, want 1", continuation.calls.Load())
	}
	if _, err := managed.flow.owner.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestStartManagedAcceptsExactLinkedBinding(t *testing.T) {
	authorization, err := runtimeorchestrationbinding.NewOrchestrationAuthorizationRequest(
		"domain-a", testWorkspaceID, testConfiguration, testInstanceID,
		runtimeorchestrationbinding.OrchestrationActionReplaceExactTarget, testVersionID,
	)
	if err != nil {
		t.Fatal(err)
	}
	parent, _ := runtimeorchestrationbinding.NewParentCommandIdentity(authorization, "replace-a")
	phase, _ := runtimeorchestrationbinding.DeriveStartTargetPhaseIdentity(parent)
	linked, _ := runtimeorchestrationbinding.NewLinkedExecutionIdentity(parent, phase)
	binding, err := runtimeorchestrationbinding.NewLinkedStartExecutionBinding(
		authorization, 1, testGeneration, linked, validStartRendezvous(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	continuation := &recordingContinuation{}
	managed := mustManagedFlow(t, continuation)
	outcome, err := managed.StartManaged(
		context.Background(),
		runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
		binding,
	)
	if err != nil || outcome.Kind() != runtimelifecycle.StartRunning || continuation.calls.Load() != 1 {
		t.Fatalf("linked managed start = %q/%v calls=%d", outcome.Kind(), err, continuation.calls.Load())
	}
	if _, err := managed.flow.owner.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestStartManagedBindingFailedReturnsSemanticOwnerOutcome(t *testing.T) {
	want := errors.New("binding failed")
	continuation := &recordingContinuation{outcome: StartClaimBindingFailed, lastErr: want}
	managed := mustManagedFlow(t, continuation)
	outcome, err := managed.StartManaged(
		context.Background(),
		runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
		testBinding(t),
	)
	if err != nil || outcome.Kind() != runtimelifecycle.StartPreparationFailed {
		t.Fatalf("binding failure = %q/%v", outcome.Kind(), err)
	}
}

func TestStartManagedInvalidContinuationPairBlocksWithContractError(t *testing.T) {
	continuation := &recordingContinuation{outcome: StartClaimContinue, lastErr: errors.New("invalid pair")}
	managed := mustManagedFlow(t, continuation)
	_, err := managed.StartManaged(
		context.Background(),
		runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
		testBinding(t),
	)
	if !errors.Is(err, ErrInvalidContinuation) {
		t.Fatalf("invalid pair error = %v", err)
	}
}

func TestStartManagedAdversarialContinuationMatrixBlocksWithoutLoad(t *testing.T) {
	wantErr := errors.New("invalid pair")
	tests := []struct {
		name         string
		continuation adversarialContinuation
	}{
		{"continue-error", adversarialContinuation{outcome: StartClaimContinue, err: wantErr}},
		{"stop-error", adversarialContinuation{outcome: StartClaimStopConverged, err: wantErr}},
		{"binding-failed-nil", adversarialContinuation{outcome: StartClaimBindingFailed}},
		{"blocked-nil", adversarialContinuation{outcome: StartClaimBlocked}},
		{"unknown", adversarialContinuation{outcome: StartClaimOutcome(99)}},
		{"panic", adversarialContinuation{panicAfter: true}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var loads atomic.Int32
			base := staticSource(availablePort(t))
			loader := configurationloader.New(sourceFunc(func(workspace, configuration, version uint64) (configurationloader.SourceObservation, error) {
				loads.Add(1)
				return base.LoadExact(workspace, configuration, version)
			}))
			managed, err := NewManaged(mustManagedOwner(t), loader, tc.continuation)
			if err != nil {
				t.Fatal(err)
			}
			_, err = managed.StartManaged(
				context.Background(),
				runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
				testBinding(t),
			)
			if !errors.Is(err, ErrInvalidContinuation) || loads.Load() != 0 {
				t.Fatalf("error=%v loads=%d", err, loads.Load())
			}
		})
	}
}

func TestStartManagedPostClaimContextDropsCancellationAndPreservesValues(t *testing.T) {
	type contextKey struct{}
	continuation := &recordingContinuation{outcome: StartClaimBlocked, lastErr: errors.New("blocked")}
	managed := mustManagedFlow(t, continuation)
	ctx, cancel := context.WithDeadline(
		context.WithValue(context.Background(), contextKey{}, "value-a"),
		time.Now().Add(time.Hour),
	)
	defer cancel()
	_, _ = managed.StartManaged(
		ctx,
		runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
		testBinding(t),
	)
	if continuation.lastContext == nil || continuation.lastContext.Value(contextKey{}) != "value-a" ||
		continuation.lastContext.Done() != nil {
		t.Fatalf("post-claim context did not preserve value/drop cancellation: %#v", continuation.lastContext)
	}
	if _, ok := continuation.lastContext.Deadline(); ok {
		t.Fatal("post-claim context retained caller deadline")
	}
}

func TestStartManagedNoClaimFailureSupersedesCallerCancellation(t *testing.T) {
	want := errors.New("no-claim signal failed")
	for _, tc := range []struct {
		name         string
		continuation adversarialContinuation
		want         error
	}{
		{"error", adversarialContinuation{err: want}, want},
		{"panic", adversarialContinuation{panicNoClaim: true}, ErrInvalidContinuation},
	} {
		t.Run(tc.name, func(t *testing.T) {
			managed, err := NewManaged(mustManagedOwner(t), mustManagedLoader(t), tc.continuation)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			_, err = managed.StartManaged(
				ctx,
				runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
				testBinding(t),
			)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
			if _, active := managed.flow.owner.Observe().ActiveAttempt(); active {
				t.Fatal("pre-claim signal failure mutated Owner")
			}
		})
	}
}

func TestStartManagedValidStopConvergedSkipsLoad(t *testing.T) {
	owner := mustManagedOwner(t)
	var loads atomic.Int32
	base := staticSource(availablePort(t))
	loader := configurationloader.New(sourceFunc(func(w, c, v uint64) (configurationloader.SourceObservation, error) {
		loads.Add(1)
		return base.LoadExact(w, c, v)
	}))
	managed, err := NewManaged(owner, loader, stoppingContinuation{owner: owner})
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := managed.StartManaged(
		context.Background(), runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID), testBinding(t),
	)
	if err != nil || outcome.Kind() != runtimelifecycle.StartStoppedBeforeRunning || loads.Load() != 0 {
		t.Fatalf("StopConverged = %q/%v loads=%d", outcome.Kind(), err, loads.Load())
	}
}

func TestStartManagedStopConvergedWithoutActualStopReturnsOwnerError(t *testing.T) {
	var loads atomic.Int32
	base := staticSource(availablePort(t))
	loader := configurationloader.New(sourceFunc(func(w, c, v uint64) (configurationloader.SourceObservation, error) {
		loads.Add(1)
		return base.LoadExact(w, c, v)
	}))
	continuation := &recordingContinuation{outcome: StartClaimStopConverged}
	managed, err := NewManaged(mustManagedOwner(t), loader, continuation)
	if err != nil {
		t.Fatal(err)
	}
	_, err = managed.StartManaged(
		context.Background(),
		runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
		testBinding(t),
	)
	if !errors.Is(err, runtimelifecycle.ErrInvalidPreparationResult) ||
		errors.Is(err, ErrInvalidContinuation) || loads.Load() != 0 {
		t.Fatalf("error=%v loads=%d", err, loads.Load())
	}
	if err != runtimelifecycle.ErrInvalidPreparationResult {
		t.Fatalf("Owner error was wrapped or replaced: %v", err)
	}
	if _, stopErr := managed.flow.owner.Stop(context.Background()); stopErr != nil {
		t.Fatalf("cleanup Stop: %v", stopErr)
	}
}

func TestStartManagedNoClaimRejectedAndFailedCausesAndSignalPrecedence(t *testing.T) {
	for _, tc := range []struct {
		name   string
		failed bool
	}{{"owner-conflict-rejected", false}, {"attempt-source-failed", true}} {
		t.Run(tc.name, func(t *testing.T) {
			continuation := &recordingContinuation{}
			var owner *runtimelifecycle.Owner
			if tc.failed {
				owner, _ = runtimelifecycle.NewOwner(testWorkspaceID, testConfiguration, testInstanceID, func() (runtimeconfigload.LaunchAttemptID, error) { return "", errors.New("source") }, managedDependencies(t))
			} else {
				owner = mustManagedOwner(t)
				_, _ = owner.PrepareStart(runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID))
			}
			managed, _ := NewManaged(owner, mustManagedLoader(t), continuation)
			_, _ = managed.StartManaged(context.Background(), runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID), testBinding(t))
			want := runtimeorchestrationbinding.StartNoClaimRejected
			if tc.failed {
				want = runtimeorchestrationbinding.StartNoClaimFailed
			}
			if continuation.noClaimCalls.Load() != 1 || continuation.lastNoClaim != want {
				t.Fatalf("calls=%d cause=%q", continuation.noClaimCalls.Load(), continuation.lastNoClaim)
			}
			precedence := errors.New("signal precedence")
			continuation.lastErr = precedence
			_, err := managed.StartManaged(context.Background(), runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID), testBinding(t))
			if err != precedence {
				t.Fatalf("precedence=%v", err)
			}
		})
	}
}

func TestStartManagedBindingFailedAndBlockedPerformZeroLoad(t *testing.T) {
	for _, tc := range []struct {
		name string
		out  StartClaimOutcome
	}{{"binding-failed", StartClaimBindingFailed}, {"blocked", StartClaimBlocked}} {
		t.Run(tc.name, func(t *testing.T) {
			var loads atomic.Int32
			base := staticSource(availablePort(t))
			loader := configurationloader.New(sourceFunc(func(w, c, v uint64) (configurationloader.SourceObservation, error) {
				loads.Add(1)
				return base.LoadExact(w, c, v)
			}))
			cause := errors.New(tc.name)
			managed, _ := NewManaged(mustManagedOwner(t), loader, &recordingContinuation{outcome: tc.out, lastErr: cause})
			_, _ = managed.StartManaged(context.Background(), runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID), testBinding(t))
			if loads.Load() != 0 {
				t.Fatalf("loads=%d", loads.Load())
			}
		})
	}
}

func TestStartManagedCancellationAfterPrepareIsSuppressed(t *testing.T) {
	probe := &cancellationProbeContinuation{entered: make(chan struct{}), release: make(chan struct{})}
	managed := mustManagedFlow(t, probe)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := managed.StartManaged(ctx, runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID), testBinding(t))
		done <- err
	}()
	<-probe.entered
	cancel()
	close(probe.release)
	<-done
	if probe.observed != nil {
		t.Fatalf("post-claim ctx error=%v", probe.observed)
	}
}

func TestStartManagedDoesNotStoreBindingAsFlowField(t *testing.T) {
	continuation := &recordingContinuation{}
	managed := mustManagedFlow(t, continuation)

	if _, err := managed.StartManaged(
		context.Background(),
		runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
		testBinding(t),
	); err != nil {
		_ = err
	}
	if continuation.calls.Load() != 1 {
		t.Fatalf("continuation calls = %d, want 1", continuation.calls.Load())
	}
}

func TestStartManagedFailsClosedOnNilContext(t *testing.T) {
	continuation := &recordingContinuation{}
	managed := mustManagedFlow(t, continuation)
	if _, err := managed.StartManaged(
		nil,
		runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
		testBinding(t),
	); !errors.Is(err, ErrInvalidStartContext) {
		t.Fatalf("nil context must fail: %v", err)
	}
	if continuation.calls.Load() != 0 {
		t.Fatalf("continuation ran %d times on nil context, want 0", continuation.calls.Load())
	}
}

func TestStartManagedPreClaimCancellationCausesZeroMutation(t *testing.T) {
	continuation := &recordingContinuation{}
	managed := mustManagedFlow(t, continuation)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := managed.StartManaged(
		ctx,
		runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
		testBinding(t),
	); !errors.Is(err, context.Canceled) {
		t.Fatalf("pre-claim cancellation must propagate: %v", err)
	}
	if continuation.calls.Load() != 0 {
		t.Fatalf("continuation ran %d times on cancelled ctx, want 0", continuation.calls.Load())
	}
	if continuation.noClaimCalls.Load() != 1 || continuation.lastNoClaim != runtimeorchestrationbinding.StartNoClaimCancelled {
		t.Fatalf("no-claim = %d/%q", continuation.noClaimCalls.Load(), continuation.lastNoClaim)
	}
}

func TestStartManagedContinuationErrorConvergesThroughAuthenticOwnerPreparation(t *testing.T) {
	want := errors.New("binding failure")
	continuation := &recordingContinuation{lastErr: want}
	managed := mustManagedFlow(t, continuation)
	owner := managed.flow.owner

	attemptFact, ok := owner.Observe().ActiveAttempt()
	if ok {
		t.Fatalf("unexpected pre-existing active attempt: %#v", attemptFact)
	}

	_, err := managed.StartManaged(
		context.Background(),
		runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
		testBinding(t),
	)
	if err != want {
		t.Fatalf("continuation failure = %v, want exact %v", err, want)
	}
	if continuation.calls.Load() != 1 {
		t.Fatalf("continuation calls = %d, want 1", continuation.calls.Load())
	}

	observation := owner.Observe()
	if fact, exists := observation.ActiveAttempt(); exists {
		if fact.Phase() == runtimelifecycle.AttemptHistorical {
			t.Logf("converged via Owner: phase %v", fact.Phase())
			return
		}
		t.Fatalf("continuation error left active non-converged claim: phase %v", fact.Phase())
	}
	if last, exists := observation.LastAttempt(); exists {
		if last.Phase() == runtimelifecycle.AttemptHistorical {
			// A historically converged last attempt proves Owner-side convergence.
			return
		}
	}
	t.Fatal("owner observation shows no converged state after continuation failure")
}

func TestStartManagedDifferentInstancesProgressIndependently(t *testing.T) {
	for i := 0; i < 4; i++ {
		instanceID := runtimeconfigload.RuntimeInstanceID("instance-" + itoa(i))
		attemptID := runtimeconfigload.LaunchAttemptID("attempt-" + itoa(i))
		continuation := &recordingContinuation{}
		owner, err := runtimelifecycle.NewOwner(
			testWorkspaceID,
			testConfiguration,
			instanceID,
			func() (runtimeconfigload.LaunchAttemptID, error) {
				return attemptID, nil
			},
			managedDependencies(t),
		)
		if err != nil {
			t.Fatal(err)
		}
		managed, err := NewManaged(owner, mustManagedLoader(t), continuation)
		if err != nil {
			t.Fatal(err)
		}
		binding, err := NewManagedStartBinding(
			validAuthorizationFor(t, instanceID), runtimeorchestrationbinding.AggregateRevision(uint64(i+1)),
			testGeneration,
			validStartRendezvous(t),
		)
		if err != nil {
			t.Fatal(err)
		}
		outcome, err := managed.StartManaged(
			context.Background(),
			runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
			binding,
		)
		if err != nil || outcome.Kind() != runtimelifecycle.StartRunning {
			t.Fatalf("instance %d managed start = %q/%v", i, outcome.Kind(), err)
		}
		if continuation.calls.Load() != 1 {
			t.Fatalf("instance %d continuation calls = %d, want 1", i, continuation.calls.Load())
		}
		if _, err := owner.Stop(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestStartManagedRejectsBindingRequestMismatchBeforeOwnerMutation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		workspace uint64
		config    uint64
		version   uint64
	}{
		{"workspace", testWorkspaceID + 1, testConfiguration, testVersionID},
		{"configuration", testWorkspaceID, testConfiguration + 1, testVersionID},
		{"version", testWorkspaceID, testConfiguration, testVersionID + 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			continuation := &recordingContinuation{}
			managed := mustManagedFlow(t, continuation)
			authorization, err := runtimeorchestrationbinding.NewOrchestrationAuthorizationRequest(
				"domain-a", tc.workspace, tc.config, testInstanceID,
				runtimeorchestrationbinding.OrchestrationActionActivateExactTarget, tc.version,
			)
			if err != nil {
				t.Fatal(err)
			}
			binding, err := NewManagedStartBinding(
				authorization, 1, testGeneration, validStartRendezvous(t),
			)
			if err != nil {
				t.Fatal(err)
			}
			_, err = managed.StartManaged(
				context.Background(),
				runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
				binding,
			)
			if !errors.Is(err, ErrInvalidManagedBinding) || continuation.calls.Load() != 0 {
				t.Fatalf("error=%v continuation=%d", err, continuation.calls.Load())
			}
			if _, active := managed.flow.owner.Observe().ActiveAttempt(); active {
				t.Fatal("mismatch mutated Owner")
			}
		})
	}
}

func TestStartManagedConvergesRuntimeInstanceMismatch(t *testing.T) {
	continuation := &recordingContinuation{}
	managed := mustManagedFlow(t, continuation)
	authorization := validAuthorizationFor(t, "foreign-instance")
	binding, err := NewManagedStartBinding(
		authorization, 1, testGeneration, validStartRendezvous(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = managed.StartManaged(
		context.Background(),
		runtimelifecycle.NewStartRequest(testWorkspaceID, testConfiguration, testVersionID),
		binding,
	)
	if !errors.Is(err, ErrInvalidManagedBinding) || continuation.calls.Load() != 0 {
		t.Fatalf("error=%v continuation=%d", err, continuation.calls.Load())
	}
	if _, active := managed.flow.owner.Observe().ActiveAttempt(); active {
		t.Fatal("instance mismatch left active attempt")
	}
}

func itoa(n int) string {
	if n >= 0 && n < 10 {
		return string(rune('0' + n))
	}
	return "many"
}

func mustManagedOwner(t *testing.T) *runtimelifecycle.Owner {
	t.Helper()
	return mustOwner(t, testInstanceID, testAttemptID, managedDependencies(t))
}

func mustManagedLoader(t *testing.T) *configurationloader.Loader {
	t.Helper()
	return configurationloader.New(staticSource(availablePort(t)))
}

func managedDependencies(t *testing.T) *runtime.DependencyBindings {
	t.Helper()
	resolver, err := secretresolver.NewMemory(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &runtime.DependencyBindings{SecretResolver: resolver}
}

func mustManagedFlow(t *testing.T, continuation StartClaimContinuation) *ManagedFlow {
	t.Helper()
	managed, err := NewManaged(mustManagedOwner(t), mustManagedLoader(t), continuation)
	if err != nil {
		t.Fatal(err)
	}
	return managed
}
