package runtimecommandidempotency

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

func TestExecuteManagedStartMapsCompleteBindingAndExpiresLookup(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 41)
	var gotAuthorization OrchestrationAuthorizationRequest
	var gotBinding runtimeorchestrationbinding.StartExecutionBinding
	authorize := func(_ context.Context, request OrchestrationAuthorizationRequest) error {
		gotAuthorization = request
		return nil
	}
	claimed, err := boundary.ExecuteManagedStart(
		context.Background(), scope, "command-a", intent, 7, "generation-a", authorize,
		func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
			gotBinding = binding
			if !boundary.managedStartRendezvousLive(scope, "command-a", binding.Rendezvous()) {
				t.Fatal("rendezvous not live during callback")
			}
			if boundary.managedStartRendezvousLive(scope, "wrong-key", binding.Rendezvous()) ||
				boundary.managedStartRendezvousLive(
					testScope(t, "domain-a", "instance-b", OperationStart),
					"command-a", binding.Rendezvous(),
				) {
				t.Fatal("rendezvous resolved for foreign identity")
			}
			return success(t)()
		},
	)
	if err != nil || claimed.Kind() != AdmissionClaimed ||
		claimed.Record().State() != CommandStateTerminal {
		t.Fatalf("admission = %#v, error = %v", claimed, err)
	}
	if gotAuthorization.OperationalDomain() != scope.Domain() ||
		gotAuthorization.WorkspaceID() != scope.WorkspaceID() ||
		gotAuthorization.ConfigurationID() != scope.ConfigurationID() ||
		gotAuthorization.RuntimeInstanceID() != scope.RuntimeInstanceID() ||
		gotAuthorization.Action() != OrchestrationActionActivateExactTarget ||
		gotAuthorization.TargetConfigurationVersionID() != intent.ConfigurationVersionID() {
		t.Fatalf("authorization = %#v", gotAuthorization)
	}
	if gotBinding.Authorization() != gotAuthorization ||
		gotBinding.ExpectedAggregateRevision() != 7 ||
		gotBinding.ExecutionGeneration() != "generation-a" {
		t.Fatalf("binding = %#v", gotBinding)
	}
	if _, linked := gotBinding.LinkedExecutionIdentity(); linked {
		t.Fatal("primitive binding carries linked identity")
	}
	if boundary.managedStartRendezvousLive(scope, "command-a", gotBinding.Rendezvous()) {
		t.Fatal("rendezvous remained live after callback")
	}
	forged, _ := runtimeorchestrationbinding.NewStartRendezvous("forged")
	if boundary.managedStartRendezvousLive(scope, "command-a", forged) {
		t.Fatal("forged rendezvous resolved")
	}
}

func TestExecuteManagedStartConcurrentSameKeyIssuesOneBinding(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 41)
	entered := make(chan runtimeorchestrationbinding.StartRendezvous, 1)
	release := make(chan struct{})
	completed := make(chan error, 1)
	var callbacks atomic.Int32
	go func() {
		_, err := boundary.ExecuteManagedStart(
			context.Background(), scope, "same-key", intent, 1, "generation-a",
			allowOrchestration,
			func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
				callbacks.Add(1)
				entered <- binding.Rendezvous()
				<-release
				return success(t)()
			},
		)
		completed <- err
	}()
	rendezvous := <-entered
	observed, err := boundary.ExecuteManagedStart(
		context.Background(), scope, "same-key", intent, 1, "generation-a",
		allowOrchestration,
		func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
			t.Fatal("in-progress submission received callback")
			return TerminalOutcome{}, nil
		},
	)
	if err != nil || observed.Kind() != AdmissionInProgress || callbacks.Load() != 1 ||
		!boundary.managedStartRendezvousLive(scope, "same-key", rendezvous) {
		t.Fatalf("observation=%#v error=%v callbacks=%d", observed, err, callbacks.Load())
	}
	close(release)
	if err := <-completed; err != nil {
		t.Fatal(err)
	}
	if boundary.managedStartRendezvousLive(scope, "same-key", rendezvous) {
		t.Fatal("rendezvous remained live after winning callback")
	}
}

func TestExecuteManagedStartAllocatesUniqueRendezvousPerNewClaim(t *testing.T) {
	boundary := newTestBoundary(t)
	var rendezvous [2]runtimeorchestrationbinding.StartRendezvous
	for index, instance := range []string{"instance-a", "instance-b"} {
		scope := testScope(t, "domain-a", instance, OperationStart)
		admission, err := boundary.ExecuteManagedStart(
			context.Background(), scope, CommandKey("command-"+instance),
			startIntent(t, uint64(41+index)), 1, "generation-a", allowOrchestration,
			func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
				rendezvous[index] = binding.Rendezvous()
				return success(t)()
			},
		)
		if err != nil || admission.Kind() != AdmissionClaimed {
			t.Fatalf("claim %d = %#v/%v", index, admission, err)
		}
	}
	if rendezvous[0] == (runtimeorchestrationbinding.StartRendezvous{}) ||
		rendezvous[1] == (runtimeorchestrationbinding.StartRendezvous{}) ||
		rendezvous[0] == rendezvous[1] {
		t.Fatalf("rendezvous identities are not unique: %#v %#v", rendezvous[0], rendezvous[1])
	}
}

func TestExecuteManagedStartAuthorizesEveryObservationWithoutAdoption(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 41)
	var authorizations atomic.Int32
	var callbacks atomic.Int32
	authorize := func(context.Context, OrchestrationAuthorizationRequest) error {
		authorizations.Add(1)
		return nil
	}
	invoke := func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
		callbacks.Add(1)
		return success(t)()
	}
	claimed, err := boundary.ExecuteManagedStart(
		context.Background(), scope, "managed", intent, 1, "generation-a", authorize, invoke,
	)
	if err != nil || claimed.Kind() != AdmissionClaimed {
		t.Fatalf("claim = %#v/%v", claimed, err)
	}
	replay, err := boundary.ExecuteManagedStart(
		context.Background(), scope, "managed", intent, 1, "generation-a", authorize, invoke,
	)
	if err != nil || replay.Kind() != AdmissionReplay {
		t.Fatalf("replay = %#v/%v", replay, err)
	}
	if callbacks.Load() != 1 || authorizations.Load() != 2 {
		t.Fatalf("callbacks=%d authorizations=%d", callbacks.Load(), authorizations.Load())
	}

	legacy, err := boundary.Execute(
		context.Background(), testScope(t, "domain-a", "instance-b", OperationStart),
		"legacy", startIntent(t, 42), allow, success(t),
	)
	if err != nil || legacy.Kind() != AdmissionClaimed {
		t.Fatalf("legacy = %#v/%v", legacy, err)
	}
	legacyScope := testScope(t, "domain-a", "instance-b", OperationStart)
	observed, err := boundary.ExecuteManagedStart(
		context.Background(), legacyScope, "legacy", startIntent(t, 42), 1,
		"generation-a", authorize, invoke,
	)
	if err != nil || observed.Kind() != AdmissionReplay {
		t.Fatalf("legacy observation = %#v/%v", observed, err)
	}
	if callbacks.Load() != 1 {
		t.Fatalf("legacy record adopted callback authority: %d", callbacks.Load())
	}

	legacyLiveScope := testScope(t, "domain-a", "instance-c", OperationStart)
	legacyEntered := make(chan struct{})
	legacyRelease := make(chan struct{})
	legacyDone := make(chan error, 1)
	go func() {
		_, legacyErr := boundary.Execute(
			context.Background(), legacyLiveScope, "legacy-live", startIntent(t, 43), allow,
			func() (TerminalOutcome, error) {
				close(legacyEntered)
				<-legacyRelease
				return success(t)()
			},
		)
		legacyDone <- legacyErr
	}()
	<-legacyEntered
	inProgress, err := boundary.ExecuteManagedStart(
		context.Background(), legacyLiveScope, "legacy-live", startIntent(t, 43), 1,
		"generation-a", authorize, invoke,
	)
	if err != nil || inProgress.Kind() != AdmissionInProgress || callbacks.Load() != 1 {
		t.Fatalf("legacy in-progress = %#v/%v callbacks=%d", inProgress, err, callbacks.Load())
	}
	close(legacyRelease)
	if err := <-legacyDone; err != nil {
		t.Fatal(err)
	}
}

func TestExecuteManagedStartValidationDenialAndCancellationCauseZeroMutation(t *testing.T) {
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 41)
	denied := errors.New("denied")
	for _, tc := range []struct {
		name       string
		ctx        context.Context
		scope      Scope
		revision   runtimeorchestrationbinding.AggregateRevision
		generation runtimeorchestrationbinding.ExecutionGeneration
		authorize  AuthorizeOrchestration
		want       error
	}{
		{"nil context", nil, scope, 1, "generation-a", allowOrchestration, ErrInvalidSubmission},
		{"stop scope", context.Background(), testScope(t, "domain-a", "instance-a", OperationStop), 1, "generation-a", allowOrchestration, ErrInvalidSubmission},
		{"zero revision", context.Background(), scope, 0, "generation-a", allowOrchestration, ErrInvalidSubmission},
		{"zero generation", context.Background(), scope, 1, "", allowOrchestration, ErrInvalidSubmission},
		{"nil authorizer", context.Background(), scope, 1, "generation-a", nil, ErrInvalidSubmission},
		{"panicking authorizer", context.Background(), scope, 1, "generation-a", func(context.Context, OrchestrationAuthorizationRequest) error { panic("policy") }, ErrInvalidSubmission},
		{"denied", context.Background(), scope, 1, "generation-a", func(context.Context, OrchestrationAuthorizationRequest) error { return denied }, denied},
	} {
		t.Run(tc.name, func(t *testing.T) {
			boundary := newTestBoundary(t)
			var callbacks atomic.Int32
			_, err := boundary.ExecuteManagedStart(
				tc.ctx, tc.scope, "command-a", intent, tc.revision, tc.generation,
				tc.authorize, func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
					callbacks.Add(1)
					return success(t)()
				},
			)
			if !errors.Is(err, tc.want) || callbacks.Load() != 0 {
				t.Fatalf("error=%v callbacks=%d", err, callbacks.Load())
			}
		})
	}
	boundary := newTestBoundary(t)
	if _, err := boundary.ExecuteManagedStart(
		context.Background(), scope, "nil-callback", intent, 1, "generation-a",
		allowOrchestration, nil,
	); !errors.Is(err, ErrInvalidSubmission) {
		t.Fatalf("nil callback error = %v", err)
	}
	var nilBoundary *Boundary
	if _, err := nilBoundary.ExecuteManagedStart(
		context.Background(), scope, "nil-boundary", intent, 1, "generation-a",
		allowOrchestration, func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
			return success(t)()
		},
	); !errors.Is(err, ErrInvalidSubmission) {
		t.Fatalf("nil boundary error = %v", err)
	}

	boundary = newTestBoundary(t)
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := boundary.ExecuteManagedStart(
		cancelled, scope, "cancelled", intent, 1, "generation-a", allowOrchestration,
		func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
			t.Fatal("cancelled submission invoked callback")
			return TerminalOutcome{}, nil
		},
	); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel error = %v", err)
	}
	claimed, err := boundary.ExecuteManagedStart(
		context.Background(), scope, "cancelled", intent, 1, "generation-a",
		allowOrchestration, func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
			return success(t)()
		},
	)
	if err != nil || claimed.Kind() != AdmissionClaimed {
		t.Fatalf("later claim = %#v/%v", claimed, err)
	}
}

func TestExecuteManagedStartFailureAndGenerationLossExpireAuthority(t *testing.T) {
	for _, tc := range []struct {
		name   string
		invoke func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error)
	}{
		{"error", func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
			return TerminalOutcome{}, errors.New("callback")
		}},
		{"panic", func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) { panic("callback") }},
		{"invalid outcome", func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
			return TerminalOutcome{}, nil
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			boundary := newTestBoundary(t)
			scope := testScope(t, "domain-a", "instance-a", OperationStart)
			var rendezvous runtimeorchestrationbinding.StartRendezvous
			admission, err := boundary.ExecuteManagedStart(
				context.Background(), scope, "command-a", startIntent(t, 41), 1,
				"generation-a", allowOrchestration,
				func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
					rendezvous = binding.Rendezvous()
					return tc.invoke(binding)
				},
			)
			if !errors.Is(err, ErrIndeterminateExecution) ||
				admission.Record().State() != CommandStateClaimed ||
				boundary.managedStartRendezvousLive(scope, "command-a", rendezvous) {
				t.Fatalf("admission=%#v error=%v live=%v", admission, err,
					boundary.managedStartRendezvousLive(scope, "command-a", rendezvous))
			}
		})
	}

	storage := NewMemoryStorage()
	boundary, _ := NewBoundary(storage)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	var rendezvous runtimeorchestrationbinding.StartRendezvous
	admission, err := boundary.ExecuteManagedStart(
		context.Background(), scope, "generation-loss", startIntent(t, 41), 1,
		"generation-a", allowOrchestration,
		func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
			rendezvous = binding.Rendezvous()
			if _, newErr := NewBoundary(storage); newErr != nil {
				t.Fatal(newErr)
			}
			return success(t)()
		},
	)
	if !errors.Is(err, ErrBoundaryExpired) || admission.Record().State() != CommandStateClaimed ||
		boundary.managedStartRendezvousLive(scope, "generation-loss", rendezvous) {
		t.Fatalf("generation loss = %#v/%v", admission, err)
	}
}

func TestExecuteManagedStartGoexitExpiresAuthorityAndLeavesClaimed(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	var rendezvous runtimeorchestrationbinding.StartRendezvous
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = boundary.ExecuteManagedStart(
			context.Background(), scope, "goexit", startIntent(t, 41), 1,
			"generation-a", allowOrchestration,
			func(binding runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
				rendezvous = binding.Rendezvous()
				runtime.Goexit()
				return TerminalOutcome{}, nil
			},
		)
	}()
	<-done
	if boundary.managedStartRendezvousLive(scope, "goexit", rendezvous) {
		t.Fatal("Goexit left rendezvous authority live")
	}
	observed, err := boundary.ExecuteManagedStart(
		context.Background(), scope, "goexit", startIntent(t, 41), 1,
		"generation-a", allowOrchestration,
		func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
			t.Fatal("unresolved replay invoked callback")
			return TerminalOutcome{}, nil
		},
	)
	if err != nil || observed.Kind() != AdmissionInProgress ||
		observed.Record().State() != CommandStateClaimed {
		t.Fatalf("observation = %#v/%v", observed, err)
	}
}

func allowOrchestration(context.Context, OrchestrationAuthorizationRequest) error { return nil }
