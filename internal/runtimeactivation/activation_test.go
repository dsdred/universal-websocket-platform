package runtimeactivation

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/configuration"
	"github.com/dsdred/universal-websocket-platform/internal/configurationloader"
	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	runtimeplatform "github.com/dsdred/universal-websocket-platform/internal/runtime"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecommandidempotency"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeidentity"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelaunchflow"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle"
	"github.com/dsdred/universal-websocket-platform/internal/runtimemanagement"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationcontinuation"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

const (
	testDomain        = "runtime-operations"
	testWorkspace     = uint64(11)
	testConfiguration = uint64(22)
	versionOne        = uint64(31)
	versionTwo        = uint64(32)
)

type versionReader struct {
	mu       sync.Mutex
	versions map[uint64]configurationversion.ConfigurationVersion
	calls    int
	err      error
}

func (r *versionReader) Get(id uint64) (configurationversion.ConfigurationVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls++
	if r.err != nil {
		return configurationversion.ConfigurationVersion{}, r.err
	}
	version, ok := r.versions[id]
	if !ok {
		return configurationversion.ConfigurationVersion{}, errors.New("version absent")
	}
	return version, nil
}

type sourceFunc func(uint64, uint64, uint64) (configurationloader.SourceObservation, error)

func (f sourceFunc) LoadExact(workspace, configurationID, versionID uint64) (configurationloader.SourceObservation, error) {
	return f(workspace, configurationID, versionID)
}

type fixture struct {
	t              *testing.T
	target         runtimemanagement.Target
	versions       *versionReader
	identity       *runtimeidentity.Store
	owner          *runtimelifecycle.Owner
	boundary       *runtimecommandidempotency.Boundary
	invoker        *runtimemanagement.ManagedStartInvoker
	orchestrator   *Orchestrator
	authorizations atomic.Int32
	generations    atomic.Int32
}

func newFixture(t *testing.T) *fixture {
	return newFixtureNamed(t, "")
}

func newFixtureNamed(t *testing.T, suffix string) *fixture {
	t.Helper()
	instanceID := runtimeconfigload.RuntimeInstanceID("runtime-" + t.Name() + suffix)
	target, err := runtimemanagement.NewTarget(testWorkspace, testConfiguration, instanceID)
	if err != nil {
		t.Fatal(err)
	}
	identity := runtimeidentity.NewStore()
	if err := identity.CreateRuntimeInstance(testWorkspace, testConfiguration, instanceID); err != nil {
		t.Fatal(err)
	}
	var attempts atomic.Int32
	resolver, err := secretresolver.NewMemory(nil)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := runtimelifecycle.NewOwner(testWorkspace, testConfiguration, instanceID,
		func() (runtimeconfigload.LaunchAttemptID, error) {
			return runtimeconfigload.LaunchAttemptID(fmt.Sprintf("attempt-%d", attempts.Add(1))), nil
		}, &runtimeplatform.DependencyBindings{SecretResolver: resolver})
	if err != nil {
		t.Fatal(err)
	}
	boundary, err := runtimecommandidempotency.NewBoundary(runtimecommandidempotency.NewMemoryStorage())
	if err != nil {
		t.Fatal(err)
	}
	continuation, err := runtimeorchestrationcontinuation.New(boundary, identity)
	if err != nil {
		t.Fatal(err)
	}
	port := availablePort(t)
	loader := configurationloader.New(sourceFunc(func(workspace, configurationID, versionID uint64) (configurationloader.SourceObservation, error) {
		return sourceObservation(workspace, configurationID, versionID, port), nil
	}))
	flow, err := runtimelaunchflow.NewManaged(owner, loader, continuation)
	if err != nil {
		t.Fatal(err)
	}
	invoker, err := runtimemanagement.NewManagedStartInvoker(testDomain, target, flow)
	if err != nil {
		t.Fatal(err)
	}
	f := &fixture{t: t, target: target, versions: publishedVersions(), identity: identity, owner: owner, boundary: boundary, invoker: invoker}
	orchestrator, err := New(testDomain, target, f.versions, identity, owner, boundary, invoker,
		func(context.Context, runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error {
			f.authorizations.Add(1)
			return nil
		}, func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
			return runtimeorchestrationbinding.ExecutionGeneration(fmt.Sprintf("generation-%d", f.generations.Add(1))), nil
		})
	if err != nil {
		t.Fatal(err)
	}
	f.orchestrator = orchestrator
	t.Cleanup(func() { _, _ = owner.Stop(context.Background()) })
	return f
}

func TestActivateExactPublishesRunningAndReplayDoesNotReinvoke(t *testing.T) {
	f := newFixture(t)
	request := mustRequest(t, "activate", versionOne)
	result, err := f.orchestrator.ActivateExact(context.Background(), request)
	if err != nil || result.Category() != ResultSucceeded {
		t.Fatalf("ActivateExact() = %#v/%v", result, err)
	}
	assertRunning(t, f.identity, f.owner, f.target.RuntimeInstanceID(), versionOne, "attempt-1")
	view, _ := f.identity.ReadRuntimeInstance(f.target.RuntimeInstanceID())
	revision := view.Revision()
	f.versions.err = errors.New("replay must not read version")
	replayed, err := f.orchestrator.ActivateExact(context.Background(), request)
	if err != nil || replayed.Category() != ResultSucceeded {
		t.Fatalf("ActivateExact(replay) = %#v/%v", replayed, err)
	}
	after, _ := f.identity.ReadRuntimeInstance(f.target.RuntimeInstanceID())
	if after.Revision() != revision || f.generations.Load() != 1 || f.authorizations.Load() != 2 {
		t.Fatalf("replay mutated/reinvoked: revision=%d/%d generations=%d authorizations=%d", revision, after.Revision(), f.generations.Load(), f.authorizations.Load())
	}
}

func TestSameTargetActivationAndRollbackAreSatisfiedWithoutMutation(t *testing.T) {
	f := newFixture(t)
	if _, err := f.orchestrator.ActivateExact(context.Background(), mustRequest(t, "activate", versionOne)); err != nil {
		t.Fatal(err)
	}
	before, _ := f.identity.ReadRuntimeInstance(f.target.RuntimeInstanceID())
	ownerBefore := f.owner.Observe()
	generationBefore := f.generations.Load()
	request := mustRequest(t, "activate-satisfied", versionOne)
	activation, err := f.orchestrator.ActivateExact(context.Background(), request)
	if err != nil || activation.Category() != ResultSatisfied {
		t.Fatalf("same-target activation = %#v/%v", activation, err)
	}
	versionCallsBeforeReplay := f.versions.calls
	f.versions.err = errors.New("satisfied replay must not read version")
	replayed, err := f.orchestrator.ActivateExact(context.Background(), request)
	if err != nil || replayed.Category() != ResultSatisfied {
		t.Fatalf("same-target activation replay = %#v/%v", replayed, err)
	}
	if f.versions.calls != versionCallsBeforeReplay || f.generations.Load() != generationBefore || f.owner.Observe() != ownerBefore {
		t.Fatalf("satisfied replay reinvoked decision/generation/lifecycle")
	}
	f.versions.err = nil
	rollback, err := f.orchestrator.RollbackExact(context.Background(), mustRequest(t, "rollback-satisfied", versionOne))
	if err != nil || rollback.Category() != ResultSatisfied {
		t.Fatalf("same-target rollback = %#v/%v", rollback, err)
	}
	after, _ := f.identity.ReadRuntimeInstance(f.target.RuntimeInstanceID())
	if before != after || f.generations.Load() != generationBefore {
		t.Fatalf("satisfied decisions mutated aggregate or allocated generation")
	}
}

func TestReplaceAndRollbackReleaseOldBeforeFreshAttempt(t *testing.T) {
	f := newFixture(t)
	if _, err := f.orchestrator.ActivateExact(context.Background(), mustRequest(t, "activate-v1", versionOne)); err != nil {
		t.Fatal(err)
	}
	replaced, err := f.orchestrator.ReplaceExact(context.Background(), mustRequest(t, "replace-v2", versionTwo))
	if err != nil || replaced.Category() != ResultSucceeded {
		t.Fatalf("ReplaceExact() = %#v/%v", replaced, err)
	}
	assertRunning(t, f.identity, f.owner, f.target.RuntimeInstanceID(), versionTwo, "attempt-2")
	rolledBack, err := f.orchestrator.RollbackExact(context.Background(), mustRequest(t, "rollback-v1", versionOne))
	if err != nil || rolledBack.Category() != ResultSucceeded {
		t.Fatalf("RollbackExact() = %#v/%v", rolledBack, err)
	}
	assertRunning(t, f.identity, f.owner, f.target.RuntimeInstanceID(), versionOne, "attempt-3")
	history, _ := f.identity.ReadLaunchAttemptHistory(f.target.RuntimeInstanceID())
	if len(history) != 3 || history[0].Phase() != runtimeidentity.AttemptPhaseStopped || history[1].Phase() != runtimeidentity.AttemptPhaseStopped || history[2].Phase() != runtimeidentity.AttemptPhaseRunning {
		t.Fatalf("attempt history = %#v", history)
	}
}

func TestValidationAndAuthorizationFailBeforeCommandOrLifecycleMutation(t *testing.T) {
	f := newFixture(t)
	f.versions.versions[versionOne] = configurationversion.ConfigurationVersion{ID: versionOne, ConfigurationID: testConfiguration, State: configurationversion.Draft}
	request := mustRequest(t, "invalid-version", versionOne)
	if result, err := f.orchestrator.ActivateExact(context.Background(), request); result != (Result{}) || !errors.Is(err, ErrVersionNotPublished) {
		t.Fatalf("ActivateExact(Draft) = %#v/%v", result, err)
	}
	view, _ := f.identity.ReadRuntimeInstance(f.target.RuntimeInstanceID())
	if view.Revision() != 1 || f.generations.Load() != 0 || f.authorizations.Load() != 1 {
		t.Fatalf("validation failure mutated work: view=%#v generation=%d authorization=%d", view, f.generations.Load(), f.authorizations.Load())
	}
	f.versions.versions[versionOne] = publishedVersion(versionOne)
	result, err := f.orchestrator.ActivateExact(context.Background(), request)
	if err != nil || result.Category() != ResultSucceeded {
		t.Fatalf("retry after zero-mutation validation failure = %#v/%v", result, err)
	}
}

func TestCancelledBeforeClaimDoesNotAllocateOrMutate(t *testing.T) {
	f := newFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := f.orchestrator.ActivateExact(ctx, mustRequest(t, "cancelled", versionOne))
	if result != (Result{}) || !errors.Is(err, context.Canceled) {
		t.Fatalf("ActivateExact(cancelled) = %#v/%v", result, err)
	}
	view, _ := f.identity.ReadRuntimeInstance(f.target.RuntimeInstanceID())
	if view.Revision() != 1 || f.generations.Load() != 0 {
		t.Fatalf("cancelled request mutated state or allocated generation")
	}
}

func TestAuthorizationPrecedesVersionLookupAndGeneration(t *testing.T) {
	f := newFixture(t)
	denied := errors.New("denied")
	orchestrator, err := New(testDomain, f.target, f.versions, f.identity, f.owner, f.boundary, f.invoker,
		func(context.Context, runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error {
			return denied
		},
		func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
			t.Fatal("generation called")
			return "", nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if result, callErr := orchestrator.ActivateExact(context.Background(), mustRequest(t, "denied", versionOne)); result != (Result{}) || !errors.Is(callErr, denied) {
		t.Fatalf("ActivateExact(denied) = %#v/%v", result, callErr)
	}
	if f.versions.calls != 0 {
		t.Fatalf("version lookup occurred before authorization: %d", f.versions.calls)
	}
}

func TestSameTargetStartingIsInProgressWithoutClaim(t *testing.T) {
	f := newFixture(t)
	preparation, err := f.owner.PrepareStart(runtimelifecycle.NewStartRequest(testWorkspace, testConfiguration, versionOne))
	if err != nil {
		t.Fatal(err)
	}
	attemptID := preparation.LoadRequest().LaunchAttemptID()
	claim, err := f.identity.ConditionalClaimLaunchAttempt(f.target.RuntimeInstanceID(), 1, attemptID, versionOne)
	if err != nil || !claim.Committed() {
		t.Fatalf("identity claim = %#v/%v", claim, err)
	}
	activation, err := f.orchestrator.ActivateExact(context.Background(), mustRequest(t, "starting-activate", versionOne))
	if err != nil || activation.Category() != ResultInProgress {
		t.Fatalf("same-target Starting activation = %#v/%v", activation, err)
	}
	rollback, err := f.orchestrator.RollbackExact(context.Background(), mustRequest(t, "starting-rollback", versionOne))
	if err != nil || rollback.Category() != ResultInProgress {
		t.Fatalf("same-target Starting rollback = %#v/%v", rollback, err)
	}
	view, _ := f.identity.ReadRuntimeInstance(f.target.RuntimeInstanceID())
	if view.Revision() != claim.Revision() || f.generations.Load() != 0 {
		t.Fatalf("same-target Starting decision mutated state")
	}
	_, _ = f.owner.StopExpectedAttempt(context.Background(), attemptID)
	_, _ = f.identity.ConditionalPublishTerminal(f.target.RuntimeInstanceID(), view.Revision(), attemptID, true)
}

func TestGenerationFailureLeavesClaimUnresolvedBeforeOwnerOrLoad(t *testing.T) {
	f := newFixture(t)
	generationErr := errors.New("generation unavailable")
	orchestrator, err := New(testDomain, f.target, f.versions, f.identity, f.owner, f.boundary, f.invoker,
		func(context.Context, runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error { return nil },
		func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
			return "", generationErr
		})
	if err != nil {
		t.Fatal(err)
	}
	result, err := orchestrator.ActivateExact(context.Background(), mustRequest(t, "generation-fails", versionOne))
	if result.Category() != ResultUnresolved || !errors.Is(err, runtimecommandidempotency.ErrIndeterminateExecution) {
		t.Fatalf("ActivateExact(generation failure) = %#v/%v", result, err)
	}
	view, _ := f.identity.ReadRuntimeInstance(f.target.RuntimeInstanceID())
	if view.Revision() != 1 || f.owner.Observe().ActualState() != runtimelifecycle.ActualStopped {
		t.Fatalf("generation failure reached identity or Owner")
	}
	second, secondErr := orchestrator.ActivateExact(context.Background(), mustRequest(t, "other-key", versionOne))
	if second != (Result{}) || !errors.Is(secondErr, runtimecommandidempotency.ErrInstanceBlocked) {
		t.Fatalf("unresolved barrier admitted another key: %#v/%v", second, secondErr)
	}
}

func TestStartupFailurePublishesFailedWithoutAutomaticRollback(t *testing.T) {
	f := newFixture(t)
	continuation, err := runtimeorchestrationcontinuation.New(f.boundary, f.identity)
	if err != nil {
		t.Fatal(err)
	}
	flow, err := runtimelaunchflow.NewManaged(f.owner, configurationloader.New(sourceFunc(func(uint64, uint64, uint64) (configurationloader.SourceObservation, error) {
		return configurationloader.SourceObservation{}, configurationloader.ErrSourceUnavailable
	})), continuation)
	if err != nil {
		t.Fatal(err)
	}
	invoker, err := runtimemanagement.NewManagedStartInvoker(testDomain, f.target, flow)
	if err != nil {
		t.Fatal(err)
	}
	orchestrator, err := New(testDomain, f.target, f.versions, f.identity, f.owner, f.boundary, invoker,
		func(context.Context, runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error { return nil },
		func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
			return "failure-generation", nil
		})
	if err != nil {
		t.Fatal(err)
	}
	result, err := orchestrator.ActivateExact(context.Background(), mustRequest(t, "startup-failure", versionOne))
	if err != nil || result.Category() != ResultFailed {
		t.Fatalf("ActivateExact(startup failure) = %#v/%v", result, err)
	}
	view, _ := f.identity.ReadRuntimeInstance(f.target.RuntimeInstanceID())
	history, _ := f.identity.ReadLaunchAttemptHistory(f.target.RuntimeInstanceID())
	if view.ActualState() != runtimeidentity.ActualStateFailed || len(history) != 1 || history[0].Phase() != runtimeidentity.AttemptPhaseFailed {
		t.Fatalf("failure facts = %#v/%#v", view, history)
	}
}

func TestTrackedStartingReplacementUsesPreclaimedStopAndFreshAttempt(t *testing.T) {
	f := newFixture(t)
	preparation, err := f.owner.PrepareStart(runtimelifecycle.NewStartRequest(testWorkspace, testConfiguration, versionOne))
	if err != nil {
		t.Fatal(err)
	}
	attemptID := preparation.LoadRequest().LaunchAttemptID()
	claim, err := f.identity.ConditionalClaimLaunchAttempt(f.target.RuntimeInstanceID(), 1, attemptID, versionOne)
	if err != nil || !claim.Committed() {
		t.Fatalf("identity claim = %#v/%v", claim, err)
	}
	scope, _ := runtimecommandidempotency.NewScope(testDomain, testWorkspace, testConfiguration, f.target.RuntimeInstanceID(), runtimecommandidempotency.OperationStart)
	intent, _ := runtimecommandidempotency.NewStartIntent(versionOne)
	entered, release, trackedDone := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() {
		_, _, trackedErr := f.boundary.ExecuteReplayFirstManagedStart(context.Background(), scope, "tracked-start", intent,
			func(context.Context, runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error { return nil },
			func(context.Context) (runtimecommandidempotency.AbsentCandidate, error) {
				return runtimecommandidempotency.NewExecutePrimitiveCandidate(1)
			},
			func(context.Context, runtimecommandidempotency.AbsentCandidate) (runtimecommandidempotency.CandidateRevalidation, error) {
				return runtimecommandidempotency.CandidateUnresolved, nil
			},
			func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
				return "tracked-generation", nil
			},
			func(runtimeorchestrationbinding.StartExecutionBinding) (runtimecommandidempotency.TerminalOutcome, error) {
				close(entered)
				<-release
				return runtimecommandidempotency.TerminalOutcome{}, runtimecommandidempotency.ErrIndeterminateExecution
			})
		trackedDone <- trackedErr
	}()
	<-entered
	request, err := NewTrackedStartRequest("replace-tracked", versionTwo, "tracked-start", 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err := f.orchestrator.ReplaceExact(context.Background(), request)
	if err != nil || result.Category() != ResultSucceeded {
		t.Fatalf("ReplaceExact(tracked Starting) = %#v/%v", result, err)
	}
	close(release)
	if trackedErr := <-trackedDone; !errors.Is(trackedErr, runtimecommandidempotency.ErrIndeterminateExecution) {
		t.Fatalf("tracked Start completion error = %v", trackedErr)
	}
	assertRunning(t, f.identity, f.owner, f.target.RuntimeInstanceID(), versionTwo, "attempt-2")
	history, _ := f.identity.ReadLaunchAttemptHistory(f.target.RuntimeInstanceID())
	if len(history) != 2 || history[0].Phase() != runtimeidentity.AttemptPhaseStopped || history[1].Phase() != runtimeidentity.AttemptPhaseRunning {
		t.Fatalf("tracked replacement history = %#v", history)
	}
}

type cancellingInvoker struct {
	cancel context.CancelFunc
	next   *runtimemanagement.ManagedStartInvoker
}

type cancellingOwner struct {
	next   *runtimelifecycle.Owner
	cancel context.CancelFunc
}

type unavailableStopOwner struct{ next *runtimelifecycle.Owner }

func (o unavailableStopOwner) Observe() runtimelifecycle.Observation { return o.next.Observe() }

func (unavailableStopOwner) StopExpectedAttempt(context.Context, runtimeconfigload.LaunchAttemptID) (runtimelifecycle.StopOutcome, error) {
	return runtimelifecycle.StopOutcome{}, errors.New("Stop outcome unavailable")
}

func (o cancellingOwner) Observe() runtimelifecycle.Observation { return o.next.Observe() }

func (o cancellingOwner) StopExpectedAttempt(ctx context.Context, attemptID runtimeconfigload.LaunchAttemptID) (runtimelifecycle.StopOutcome, error) {
	outcome, err := o.next.StopExpectedAttempt(ctx, attemptID)
	o.cancel()
	return outcome, err
}

func (i cancellingInvoker) InvokeManagedStart(ctx context.Context, request runtimelifecycle.StartRequest, binding runtimeorchestrationbinding.StartExecutionBinding) (runtimelifecycle.StartOutcome, error) {
	i.cancel()
	return i.next.InvokeManagedStart(ctx, request, binding)
}

func TestDefinitivePreOwnerCancellationTerminalizesWithoutAttempt(t *testing.T) {
	f := newFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	orchestrator, err := New(testDomain, f.target, f.versions, f.identity, f.owner, f.boundary, cancellingInvoker{cancel: cancel, next: f.invoker},
		func(context.Context, runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error { return nil },
		func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
			return "cancel-generation", nil
		})
	if err != nil {
		t.Fatal(err)
	}
	result, err := orchestrator.ActivateExact(ctx, mustRequest(t, "cancel-after-claim", versionOne))
	if err != nil || result.Category() != ResultRejected {
		t.Fatalf("ActivateExact(pre-Owner cancellation) = %#v/%v", result, err)
	}
	view, _ := f.identity.ReadRuntimeInstance(f.target.RuntimeInstanceID())
	if view.Revision() != 1 {
		t.Fatalf("pre-Owner cancellation mutated aggregate: %#v", view)
	}
	replayed, replayErr := orchestrator.ActivateExact(context.Background(), mustRequest(t, "cancel-after-claim", versionOne))
	if replayErr != nil || replayed.Category() != ResultRejected {
		t.Fatalf("cancelled replay = %#v/%v", replayed, replayErr)
	}
}

func TestCancellationAfterReleaseTerminalizesParentWithoutStartPhase(t *testing.T) {
	f := newFixture(t)
	if _, err := f.orchestrator.ActivateExact(context.Background(), mustRequest(t, "activate-before-cancel", versionOne)); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	orchestrator, err := New(testDomain, f.target, f.versions, f.identity, cancellingOwner{next: f.owner, cancel: cancel}, f.boundary, f.invoker,
		func(context.Context, runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error { return nil },
		func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
			t.Fatal("generation called after cancellation")
			return "", nil
		})
	if err != nil {
		t.Fatal(err)
	}
	request := mustRequest(t, "cancel-after-release", versionTwo)
	result, err := orchestrator.ReplaceExact(ctx, request)
	if err != nil || result.Category() != ResultCancelled {
		t.Fatalf("ReplaceExact(cancel after release) = %#v/%v", result, err)
	}
	view, _ := f.identity.ReadRuntimeInstance(f.target.RuntimeInstanceID())
	if view.ActualState() != runtimeidentity.ActualStateStopped {
		t.Fatalf("cancel after release aggregate = %#v, want Stopped", view)
	}
	if _, active := view.ActiveAttempt(); active {
		t.Fatal("cancel after release created an active attempt")
	}
	replayed, replayErr := orchestrator.ReplaceExact(context.Background(), request)
	if replayErr != nil || replayed.Category() != ResultCancelled {
		t.Fatalf("cancelled parent replay = %#v/%v", replayed, replayErr)
	}
}

func TestUnprovenOldStopLeavesParentUnresolvedAndNeverStartsTarget(t *testing.T) {
	f := newFixture(t)
	if _, err := f.orchestrator.ActivateExact(context.Background(), mustRequest(t, "activate-before-unresolved-stop", versionOne)); err != nil {
		t.Fatal(err)
	}
	generations := atomic.Int32{}
	orchestrator, err := New(testDomain, f.target, f.versions, f.identity, unavailableStopOwner{next: f.owner}, f.boundary, f.invoker,
		func(context.Context, runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error { return nil },
		func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
			generations.Add(1)
			return "forbidden", nil
		})
	if err != nil {
		t.Fatal(err)
	}
	result, err := orchestrator.ReplaceExact(context.Background(), mustRequest(t, "unresolved-stop", versionTwo))
	if result.Category() != ResultUnresolved || !errors.Is(err, runtimecommandidempotency.ErrIndeterminateExecution) {
		t.Fatalf("ReplaceExact(unproven Stop) = %#v/%v", result, err)
	}
	view, _ := f.identity.ReadRuntimeInstance(f.target.RuntimeInstanceID())
	history, _ := f.identity.ReadLaunchAttemptHistory(f.target.RuntimeInstanceID())
	if view.ActualState() != runtimeidentity.ActualStateStopping || len(history) != 1 || generations.Load() != 0 {
		t.Fatalf("unproven Stop continued: view=%#v history=%d generations=%d", view, len(history), generations.Load())
	}
}

func TestDifferentInstancesProgressOnSharedCommandStorage(t *testing.T) {
	first := newFixtureNamed(t, "-first")
	second := newFixtureNamed(t, "-second")
	shared := runtimecommandidempotency.NewMemoryStorage()
	boundary, _ := runtimecommandidempotency.NewBoundary(shared)
	first.boundary, second.boundary = boundary, boundary
	// Each managed Flow must use the matching shared-storage boundary.
	first = rebuildWithBoundary(t, first, boundary)
	second = rebuildWithBoundary(t, second, boundary)
	var wait sync.WaitGroup
	wait.Add(2)
	errorsFound := make(chan error, 2)
	for _, current := range []*fixture{first, second} {
		go func(f *fixture) {
			defer wait.Done()
			result, err := f.orchestrator.ActivateExact(context.Background(), mustRequest(t, "shared", versionOne))
			if err != nil || result.Category() != ResultSucceeded {
				errorsFound <- fmt.Errorf("result=%#v err=%w", result, err)
			}
		}(current)
	}
	wait.Wait()
	close(errorsFound)
	for err := range errorsFound {
		t.Error(err)
	}
}

func rebuildWithBoundary(t *testing.T, f *fixture, boundary *runtimecommandidempotency.Boundary) *fixture {
	t.Helper()
	continuation, err := runtimeorchestrationcontinuation.New(boundary, f.identity)
	if err != nil {
		t.Fatal(err)
	}
	port := availablePort(t)
	flow, err := runtimelaunchflow.NewManaged(f.owner, configurationloader.New(sourceFunc(func(workspace, configurationID, versionID uint64) (configurationloader.SourceObservation, error) {
		return sourceObservation(workspace, configurationID, versionID, port), nil
	})), continuation)
	if err != nil {
		t.Fatal(err)
	}
	f.invoker, err = runtimemanagement.NewManagedStartInvoker(testDomain, f.target, flow)
	if err != nil {
		t.Fatal(err)
	}
	f.orchestrator, err = New(testDomain, f.target, f.versions, f.identity, f.owner, boundary, f.invoker,
		func(context.Context, runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error {
			f.authorizations.Add(1)
			return nil
		},
		func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
			return runtimeorchestrationbinding.ExecutionGeneration(fmt.Sprintf("generation-%d", f.generations.Add(1))), nil
		})
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func publishedVersions() *versionReader {
	return &versionReader{versions: map[uint64]configurationversion.ConfigurationVersion{
		versionOne: publishedVersion(versionOne), versionTwo: publishedVersion(versionTwo),
	}}
}

func publishedVersion(id uint64) configurationversion.ConfigurationVersion {
	return configurationversion.ConfigurationVersion{ID: id, ConfigurationID: testConfiguration, Number: uint32(id), State: configurationversion.Published}
}

func sourceObservation(workspace, configurationID, versionID uint64, port uint16) configurationloader.SourceObservation {
	return configurationloader.SourceObservation{
		WorkspaceID:   workspace,
		Configuration: configuration.Configuration{ID: configurationID, WorkspaceID: workspace, Name: "activation-test"},
		ConfigurationVersion: configurationversion.ConfigurationVersion{
			ID: versionID, ConfigurationID: configurationID, Number: uint32(versionID), State: configurationversion.Published,
			Listener: configurationversion.ListenerSettings{Host: "127.0.0.1", Port: port, TLS: configurationversion.TLSSettings{MinVersion: "1.2"}, Timeouts: configurationversion.TimeoutSettings{HandshakeSeconds: 10, ReadSeconds: 30, WriteSeconds: 20, IdleSeconds: 60}},
		},
		SchemaIdentity: "uwp.configuration", SchemaVersion: 1, RepresentationComplete: true,
	}
}

func mustRequest(t *testing.T, key string, versionID uint64) Request {
	t.Helper()
	request, err := NewRequest(runtimecommandidempotency.CommandKey(key), versionID)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func assertRunning(t *testing.T, identity *runtimeidentity.Store, owner *runtimelifecycle.Owner, instanceID runtimeconfigload.RuntimeInstanceID, versionID uint64, attemptID runtimeconfigload.LaunchAttemptID) {
	t.Helper()
	view, err := identity.ReadRuntimeInstance(instanceID)
	if err != nil || view.ActualState() != runtimeidentity.ActualStateRunning {
		t.Fatalf("identity Running = %#v/%v", view, err)
	}
	active, ok := view.ActiveAttempt()
	ownerActive, ownerOK := owner.Observe().ActiveAttempt()
	if !ok || !ownerOK || active != attemptID || ownerActive.LaunchAttemptID() != attemptID || ownerActive.ConfigurationVersionID() != versionID || !ownerActive.RunningPublished() {
		t.Fatalf("active identity/Owner mismatch: %q/%t %#v/%t", active, ok, ownerActive, ownerOK)
	}
}

func availablePort(t *testing.T) uint16 {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, text, err := net.SplitHostPort(listener.Addr().String())
	if closeErr := listener.Close(); err == nil {
		err = closeErr
	}
	port, parseErr := strconv.ParseUint(text, 10, 16)
	if err != nil || parseErr != nil {
		t.Fatalf("available port: %v/%v", err, parseErr)
	}
	return uint16(port)
}
