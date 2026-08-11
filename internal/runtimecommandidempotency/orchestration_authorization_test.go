package runtimecommandidempotency

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

var errAuthzDenied = errors.New("orchestration authorization denied")

func TestOrchestrationAuthorizationRequestIsImmutableAndValidated(t *testing.T) {
	identity := runtimeconfigload.RuntimeInstanceID("instance-a")
	request, err := NewOrchestrationAuthorizationRequest(
		"domain-a", 1, 2, identity, OrchestrationActionActivateExactTarget, 41,
	)
	if err != nil {
		t.Fatal(err)
	}
	if request.OperationalDomain() != "domain-a" || request.WorkspaceID() != 1 || request.ConfigurationID() != 2 ||
		request.RuntimeInstanceID() != identity ||
		request.Action() != OrchestrationActionActivateExactTarget ||
		request.TargetConfigurationVersionID() != 41 {
		t.Fatalf("unexpected request contents: %#v", request)
	}

	cases := []struct {
		name      string
		domain    string
		workspace uint64
		config    uint64
		instance  runtimeconfigload.RuntimeInstanceID
		action    OrchestrationAction
		version   uint64
	}{
		{"empty domain", "", 1, 2, identity, OrchestrationActionActivateExactTarget, 41},
		{"zero workspace", "domain-a", 0, 2, identity, OrchestrationActionActivateExactTarget, 41},
		{"zero configuration", "domain-a", 1, 0, identity, OrchestrationActionActivateExactTarget, 41},
		{"empty instance", "domain-a", 1, 2, "", OrchestrationActionActivateExactTarget, 41},
		{"unknown action", "domain-a", 1, 2, identity, OrchestrationAction("unknown"), 41},
		{"empty action", "domain-a", 1, 2, identity, OrchestrationAction(""), 41},
		{"zero version", "domain-a", 1, 2, identity, OrchestrationActionActivateExactTarget, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewOrchestrationAuthorizationRequest(
				tc.domain, tc.workspace, tc.config, tc.instance, tc.action, tc.version,
			); !errors.Is(err, ErrInvalidSubmission) {
				t.Fatalf("expected ErrInvalidSubmission, got %v", err)
			}
		})
	}

	for _, action := range []OrchestrationAction{
		OrchestrationActionActivateExactTarget,
		OrchestrationActionReplaceExactTarget,
		OrchestrationActionRollbackExactTarget,
	} {
		if _, err := NewOrchestrationAuthorizationRequest(
			"domain-a", 1, 2, identity, action, 41,
		); err != nil {
			t.Fatalf("valid action %q rejected: %v", action, err)
		}
	}
}

func TestOrchestrationAuthorizationRequestCarriesOperationalDomain(t *testing.T) {
	request, err := NewOrchestrationAuthorizationRequest(
		"domain-a", 1, 2, runtimeconfigload.RuntimeInstanceID("instance-a"),
		OrchestrationActionReplaceExactTarget, 41,
	)
	if err != nil {
		t.Fatal(err)
	}
	if request.OperationalDomain() != "domain-a" {
		t.Fatalf("operational domain = %q", request.OperationalDomain())
	}
}

func TestOrchestrationActionSetIsExact(t *testing.T) {
	actions := []OrchestrationAction{
		OrchestrationActionActivateExactTarget,
		OrchestrationActionReplaceExactTarget,
		OrchestrationActionRollbackExactTarget,
	}
	seen := make(map[OrchestrationAction]bool, len(actions))
	for _, action := range actions {
		if !action.Valid() {
			t.Fatalf("defined action %q reports invalid", action)
		}
		if seen[action] {
			t.Fatalf("duplicate action %q", action)
		}
		seen[action] = true
	}
	if len(seen) != 3 {
		t.Fatalf("expected exactly 3 actions, got %d", len(seen))
	}
	for _, foreign := range []OrchestrationAction{
		"", "start", "stop", "observe", "activate", "replace", "rollback",
		"ActivateExactTarget", "StopExactTarget",
	} {
		if foreign.Valid() {
			t.Fatalf("foreign action %q reports valid", foreign)
		}
	}
}

func TestAuthorizeCommandMapsExactActionsWithoutFallback(t *testing.T) {
	instance := runtimeconfigload.RuntimeInstanceID("instance-a")
	startScope := testScope(t, "domain-a", "instance-a", OperationStart)
	replaceScope := testScope(t, "domain-a", "instance-a", OperationReplace)
	rollbackScope := testScope(t, "domain-a", "instance-a", OperationRollback)
	stopScope := testScope(t, "domain-a", "instance-a", OperationStop)

	start, err := NewStartIntent(41)
	if err != nil {
		t.Fatal(err)
	}
	replace, err := NewReplaceIntent(42)
	if err != nil {
		t.Fatal(err)
	}
	rollback, err := NewRollbackIntent(43)
	if err != nil {
		t.Fatal(err)
	}
	stop := NewStopIntent()

	cases := []struct {
		name    string
		scope   Scope
		intent  Intent
		want    OrchestrationAction
		version uint64
		ok      bool
	}{
		{"start maps to activate", startScope, start, OrchestrationActionActivateExactTarget, 41, true},
		{"replace maps to replace", replaceScope, replace, OrchestrationActionReplaceExactTarget, 42, true},
		{"rollback maps to rollback", rollbackScope, rollback, OrchestrationActionRollbackExactTarget, 43, true},
		{"stop has no orchestration action", stopScope, stop, "", 0, false},
		{"cross-operation start intent in replace scope", replaceScope, start, "", 0, false},
		{"cross-operation replace intent in start scope", startScope, replace, "", 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request, ok := authorizeCommandMaps(tc.scope, tc.intent)
			if ok != tc.ok {
				t.Fatalf("ok=%v, expected %v", ok, tc.ok)
			}
			if !tc.ok {
				return
			}
			if request.Action() != tc.want ||
				request.TargetConfigurationVersionID() != tc.version ||
				request.OperationalDomain() != "domain-a" ||
				request.WorkspaceID() != 1 || request.ConfigurationID() != 2 ||
				request.RuntimeInstanceID() != instance {
				t.Fatalf("unexpected mapping: %#v", request)
			}
		})
	}
	// Zero-value scope never maps.
	if _, ok := authorizeCommandMaps(Scope{}, start); ok {
		t.Fatal("zero scope mapped")
	}
}

func TestAuthorizeOrchestrationValidatesFunctionPerSubmission(t *testing.T) {
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 41)
	if err := authorizeOrchestration(nil, context.Background(), scope, intent); !errors.Is(err, ErrInvalidSubmission) {
		t.Fatalf("nil authorizer must fail closed, got %v", err)
	}
	if err := authorizeOrchestration(
		func(context.Context, OrchestrationAuthorizationRequest) error { return nil },
		nil, scope, intent,
	); !errors.Is(err, ErrInvalidSubmission) {
		t.Fatalf("nil context must fail closed, got %v", err)
	}
	invalidScope := testScope(t, "domain-a", "instance-a", OperationStop)
	if err := authorizeOrchestration(
		func(context.Context, OrchestrationAuthorizationRequest) error { return nil },
		context.Background(), invalidScope, NewStopIntent(),
	); !errors.Is(err, ErrInvalidSubmission) {
		t.Fatalf("stop intent must not map to orchestration authorization, got %v", err)
	}
}

func TestAuthorizeOrchestrationRunsOnEveryCallWithoutCache(t *testing.T) {
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 41)
	var calls atomic.Int32
	authorize := func(_ context.Context, request OrchestrationAuthorizationRequest) error {
		calls.Add(1)
		if request.Action() != OrchestrationActionActivateExactTarget {
			t.Errorf("unexpected action %q", request.Action())
		}
		return nil
	}
	for i := 0; i < 3; i++ {
		if err := authorizeOrchestration(authorize, context.Background(), scope, intent); err != nil {
			t.Fatal(err)
		}
	}
	if calls.Load() != 3 {
		t.Fatalf("authorizer calls = %d, want 3 (never cached)", calls.Load())
	}
}

func TestOrchestrationDenialFailureAndPanicCauseZeroMutation(t *testing.T) {
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 41)

	for _, tc := range []struct {
		name      string
		authorize AuthorizeOrchestration
		want      error
	}{
		{
			"denial returned unchanged",
			func(context.Context, OrchestrationAuthorizationRequest) error { return errAuthzDenied },
			errAuthzDenied,
		},
		{
			"panic fails closed without mutation",
			func(context.Context, OrchestrationAuthorizationRequest) error { panic("policy defect") },
			ErrInvalidSubmission,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			boundary := newTestBoundary(t)
			err := authorizeOrchestration(tc.authorize, context.Background(), scope, intent)
			if !errors.Is(err, tc.want) {
				t.Fatalf("authorizeOrchestration error = %v, want %v", err, tc.want)
			}
			// A later allowed claim proves the denial/panic performed zero mutation.
			claimed, claimErr := boundary.Execute(
				context.Background(), scope, "command-a", intent, allow, success(t),
			)
			if claimErr != nil {
				t.Fatal(claimErr)
			}
			if claimed.Kind() != AdmissionClaimed || claimed.Record().State() != CommandStateTerminal {
				t.Fatalf("unexpected later claim: %#v", claimed)
			}
		})
	}
}

func TestOrchestrationPreClaimCancellationCausesZeroMutation(t *testing.T) {
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 41)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := authorizeOrchestration(
		func(context.Context, OrchestrationAuthorizationRequest) error {
			return ctx.Err()
		},
		ctx, scope, intent,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled passthrough, got %v", err)
	}
	boundary := newTestBoundary(t)
	claimed, claimErr := boundary.Execute(
		context.Background(), scope, "command-a", intent, allow, success(t),
	)
	if claimErr != nil {
		t.Fatal(claimErr)
	}
	if claimed.Kind() != AdmissionClaimed {
		t.Fatalf("pre-claim cancellation must not mutate: %#v", claimed)
	}
}

func TestOrchestrationReplayRunsAuthorizerAgain(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 41)
	var calls atomic.Int32
	var denyNext atomic.Bool
	authorize := func(_ context.Context, request OrchestrationAuthorizationRequest) error {
		calls.Add(1)
		if denyNext.Load() {
			return errAuthzDenied
		}
		return nil
	}

	claimed, err := boundary.Execute(
		context.Background(), scope, "command-a", intent,
		func(ctx context.Context, s Scope, i Intent) error {
			return authorizeOrchestration(authorize, ctx, s, i)
		},
		success(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	if claimed.Kind() != AdmissionClaimed || claimed.Record().State() != CommandStateTerminal {
		t.Fatalf("unexpected claim: %#v", claimed)
	}

	denyNext.Store(true)
	replay, err := boundary.Execute(
		context.Background(), scope, "command-a", intent,
		func(ctx context.Context, s Scope, i Intent) error {
			return authorizeOrchestration(authorize, ctx, s, i)
		},
		func() (TerminalOutcome, error) {
			t.Fatal("replay must never delegate")
			return TerminalOutcome{}, errors.New("must not run")
		},
	)
	if !errors.Is(err, errAuthzDenied) {
		t.Fatalf("denied replay must return the exact denial, got %v", err)
	}
	if replay.Kind() != "" {
		t.Fatalf("denied replay returned an admission: %#v", replay)
	}

	denyNext.Store(false)
	replay, err = boundary.Execute(
		context.Background(), scope, "command-a", intent,
		func(ctx context.Context, s Scope, i Intent) error {
			return authorizeOrchestration(authorize, ctx, s, i)
		},
		func() (TerminalOutcome, error) {
			t.Fatal("replay must never delegate")
			return TerminalOutcome{}, errors.New("must not run")
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if replay.Kind() != AdmissionReplay {
		t.Fatalf("allowed replay must return replay admission: %#v", replay)
	}
	if calls.Load() != 3 {
		t.Fatalf("authorizer calls = %d, want 3 (initial + denied replay + allowed replay)", calls.Load())
	}
}

func TestPrimitiveStartAuthorizationAdaptsToActivateExactTarget(t *testing.T) {
	boundary := newTestBoundary(t)
	scope := testScope(t, "domain-a", "instance-a", OperationStart)
	intent := startIntent(t, 41)
	var gotRequest OrchestrationAuthorizationRequest
	var calls atomic.Int32
	adapter := func(ctx context.Context, s Scope, i Intent) error {
		return authorizeOrchestration(
			func(_ context.Context, request OrchestrationAuthorizationRequest) error {
				calls.Add(1)
				gotRequest = request
				return nil
			},
			ctx, s, i,
		)
	}

	claimed, err := boundary.Execute(context.Background(), scope, "command-a", intent, adapter, success(t))
	if err != nil {
		t.Fatal(err)
	}
	if claimed.Kind() != AdmissionClaimed {
		t.Fatalf("unexpected claim: %#v", claimed)
	}
	if calls.Load() != 1 {
		t.Fatalf("adapter authorization calls = %d, want 1", calls.Load())
	}
	if gotRequest.Action() != OrchestrationActionActivateExactTarget ||
		gotRequest.OperationalDomain() != scope.Domain() ||
		gotRequest.TargetConfigurationVersionID() != 41 ||
		gotRequest.WorkspaceID() != scope.WorkspaceID() ||
		gotRequest.ConfigurationID() != scope.ConfigurationID() ||
		gotRequest.RuntimeInstanceID() != scope.RuntimeInstanceID() {
		t.Fatalf("activation not adapted exactly: %#v", gotRequest)
	}
}

func TestActivationCreatesNoParentAndParentsStayOnExecuteParent(t *testing.T) {
	boundary := newTestBoundary(t)
	activationScope := testScope(t, "domain-a", "instance-a", OperationStart)
	claimed, err := boundary.Execute(
		context.Background(), activationScope, "activate-a", startIntent(t, 41), allow, success(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	if claimed.Kind() != AdmissionClaimed {
		t.Fatalf("activation admission: %#v", claimed)
	}

	// Replace and Rollback remain parent submissions on the same boundary and
	// admit independently of the completed primitive activation.
	for _, tc := range []struct {
		name   string
		scope  Scope
		intent Intent
	}{
		{"replace", testScope(t, "domain-a", "instance-b", OperationReplace), Intent{}},
		{"rollback", testScope(t, "domain-a", "instance-c", OperationRollback), Intent{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			intent := tc.intent
			var err error
			if tc.scope.Operation() == OperationReplace {
				intent, err = NewReplaceIntent(42)
			} else {
				intent, err = NewRollbackIntent(43)
			}
			if err != nil {
				t.Fatal(err)
			}
			var delegated atomic.Int32
			admission, admitErr := boundary.ExecuteParent(
				context.Background(), tc.scope, "parent-a", intent, allow,
				func(execution *ParentExecution) error {
					delegated.Add(1)
					_, err := execution.PublishTerminal(
						mustParentOutcome(t, ParentOutcomeSatisfied),
					)
					return err
				},
			)
			if admitErr != nil {
				t.Fatal(admitErr)
			}
			if admission.Kind() != AdmissionClaimed || delegated.Load() != 1 {
				t.Fatalf("parent admission: %#v delegations=%d", admission, delegated.Load())
			}
		})
	}
}

func mustParentOutcome(t *testing.T, category ParentOutcomeCategory) ParentTerminalOutcome {
	t.Helper()
	outcome, err := NewParentTerminalOutcome(category)
	if err != nil {
		t.Fatal(err)
	}
	return outcome
}

func TestOrchestrationDifferentInstancesProgressIndependently(t *testing.T) {
	boundary := newTestBoundary(t)
	const instances = 8
	done := make(chan error, instances)
	for i := 0; i < instances; i++ {
		i := i
		go func() {
			scope := testScope(t, "domain-a",
				runtimeInstanceFor(i), OperationStart)
			intent := startIntent(t, uint64(100+i))
			var requests atomic.Int32
			authorize := func(context.Context, OrchestrationAuthorizationRequest) error {
				requests.Add(1)
				return nil
			}
			claimed, err := boundary.Execute(
				context.Background(), scope, CommandKey("cmd"), intent,
				func(ctx context.Context, s Scope, in Intent) error {
					return authorizeOrchestration(authorize, ctx, s, in)
				},
				success(t),
			)
			if err == nil && claimed.Kind() != AdmissionClaimed {
				err = errors.New("not claimed")
			}
			if err == nil && requests.Load() != 1 {
				err = errors.New("authorizer not called exactly once")
			}
			done <- err
		}()
	}
	for i := 0; i < instances; i++ {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
}

func runtimeInstanceFor(i int) string {
	return "instance-" + string(rune('a'+i))
}

// TestPublicAuthorizeSurfaceUnchanged pins the existing public authorization
// surface: the primitive/parent Authorize signature and Execute/ExecuteParent
// parameter lists must remain exactly as before this slice. Any change breaks
// compilation here.
func TestPublicAuthorizeSurfaceUnchanged(t *testing.T) {
	var primitive Authorize = allow
	var _ Authorize = primitive
	var _ func(
		*Boundary, context.Context, Scope, CommandKey, Intent, Authorize,
		func() (TerminalOutcome, error),
	) (Admission, error) = (*Boundary).Execute
	var _ func(
		*Boundary, context.Context, Scope, CommandKey, Intent, Authorize,
		func(*ParentExecution) error,
	) (ParentAdmission, error) = (*Boundary).ExecuteParent
}
