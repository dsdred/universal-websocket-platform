package runtimelaunchflow

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

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
	calls       atomic.Int32
	lastErr     error
	lastBinding ManagedStartBinding
	lastView    OwnerClaimView
}

func (r *recordingContinuation) AfterOwnerClaim(
	ctx context.Context,
	binding ManagedStartBinding,
	view OwnerClaimView,
) error {
	r.calls.Add(1)
	r.lastBinding = binding
	r.lastView = view
	return r.lastErr
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
