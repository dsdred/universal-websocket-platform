package runtimeorchestrationbinding

import (
	"errors"
	"reflect"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

func TestAuthorizationRequestCarriesExactSixFields(t *testing.T) {
	request := testAuthorization(t, OrchestrationActionActivateExactTarget)
	if request.OperationalDomain() != "domain-a" || request.WorkspaceID() != 1 ||
		request.ConfigurationID() != 2 || request.RuntimeInstanceID() != "instance-a" ||
		request.Action() != OrchestrationActionActivateExactTarget ||
		request.TargetConfigurationVersionID() != 3 {
		t.Fatalf("unexpected authorization: %#v", request)
	}

	cases := []struct {
		name      string
		domain    string
		workspace uint64
		config    uint64
		instance  runtimeconfigload.RuntimeInstanceID
		action    OrchestrationAction
		target    uint64
	}{
		{"domain", "", 1, 2, "instance-a", OrchestrationActionActivateExactTarget, 3},
		{"workspace", "domain-a", 0, 2, "instance-a", OrchestrationActionActivateExactTarget, 3},
		{"configuration", "domain-a", 1, 0, "instance-a", OrchestrationActionActivateExactTarget, 3},
		{"instance", "domain-a", 1, 2, "", OrchestrationActionActivateExactTarget, 3},
		{"action", "domain-a", 1, 2, "instance-a", "unknown", 3},
		{"target", "domain-a", 1, 2, "instance-a", OrchestrationActionActivateExactTarget, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewOrchestrationAuthorizationRequest(
				tc.domain, tc.workspace, tc.config, tc.instance, tc.action, tc.target,
			)
			if !errors.Is(err, ErrInvalidBinding) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestPrimitiveBindingIsLosslessAndHasNoLinkedIdentity(t *testing.T) {
	authorization := testAuthorization(t, OrchestrationActionActivateExactTarget)
	rendezvous := testRendezvous(t, "rendezvous-a")
	binding, err := NewPrimitiveStartExecutionBinding(
		authorization, AggregateRevision(^uint64(0)), ExecutionGeneration("generation-a"), rendezvous,
	)
	if err != nil {
		t.Fatal(err)
	}
	if binding.Authorization() != authorization ||
		binding.ExpectedAggregateRevision() != AggregateRevision(^uint64(0)) ||
		binding.ExecutionGeneration() != "generation-a" || binding.Rendezvous() != rendezvous {
		t.Fatalf("unexpected binding: %#v", binding)
	}
	if _, linked := binding.LinkedExecutionIdentity(); linked {
		t.Fatal("primitive binding unexpectedly carries linked identity")
	}

	for _, tc := range []struct {
		name       string
		auth       OrchestrationAuthorizationRequest
		revision   AggregateRevision
		generation ExecutionGeneration
		rendezvous StartRendezvous
	}{
		{"authorization", OrchestrationAuthorizationRequest{}, 1, "generation-a", rendezvous},
		{"revision", authorization, 0, "generation-a", rendezvous},
		{"generation", authorization, 1, "", rendezvous},
		{"rendezvous", authorization, 1, "generation-a", StartRendezvous{}},
		{"linked action", testAuthorization(t, OrchestrationActionReplaceExactTarget), 1, "generation-a", rendezvous},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewPrimitiveStartExecutionBinding(
				tc.auth, tc.revision, tc.generation, tc.rendezvous,
			); !errors.Is(err, ErrInvalidBinding) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestLinkedIdentityRequiresExactDerivedStartTarget(t *testing.T) {
	for _, action := range []OrchestrationAction{
		OrchestrationActionReplaceExactTarget, OrchestrationActionRollbackExactTarget,
	} {
		authorization := testAuthorization(t, action)
		parent, err := NewParentCommandIdentity(authorization, "parent-a")
		if err != nil {
			t.Fatal(err)
		}
		phase, err := DeriveStartTargetPhaseIdentity(parent)
		if err != nil {
			t.Fatal(err)
		}
		if phase.Ordinal() != 1 || phase.Parent() != parent {
			t.Fatalf("unexpected phase: %#v", phase)
		}
		linked, err := NewLinkedExecutionIdentity(parent, phase)
		if err != nil {
			t.Fatal(err)
		}
		binding, err := NewLinkedStartExecutionBinding(
			authorization, 1, "generation-a", linked, testRendezvous(t, string(action)),
		)
		if err != nil || !binding.Valid() {
			t.Fatalf("binding = %#v, error = %v", binding, err)
		}
		got, ok := binding.LinkedExecutionIdentity()
		if !ok || got != linked {
			t.Fatalf("linked = %#v/%v", got, ok)
		}
	}

	activate := testAuthorization(t, OrchestrationActionActivateExactTarget)
	if _, err := NewParentCommandIdentity(activate, "parent-a"); !errors.Is(err, ErrInvalidBinding) {
		t.Fatalf("activate parent error = %v", err)
	}
	replace := testAuthorization(t, OrchestrationActionReplaceExactTarget)
	parent, _ := NewParentCommandIdentity(replace, "parent-a")
	foreign, _ := NewParentCommandIdentity(replace, "parent-b")
	foreignPhase, _ := DeriveStartTargetPhaseIdentity(foreign)
	if _, err := NewLinkedExecutionIdentity(parent, foreignPhase); !errors.Is(err, ErrInvalidBinding) {
		t.Fatalf("foreign phase error = %v", err)
	}
	forged := StartTargetPhaseIdentity{parent: parent, kind: "stop-old", ordinal: 0}
	if _, err := NewLinkedExecutionIdentity(parent, forged); !errors.Is(err, ErrInvalidBinding) {
		t.Fatalf("forged phase error = %v", err)
	}
}

func TestRendezvousIsOpaqueComparableIdentity(t *testing.T) {
	first := testRendezvous(t, "first")
	second := testRendezvous(t, "second")
	if first == second || first == (StartRendezvous{}) {
		t.Fatalf("invalid identities: %#v %#v", first, second)
	}
	if _, err := NewStartRendezvous(""); !errors.Is(err, ErrInvalidBinding) {
		t.Fatalf("zero identity error = %v", err)
	}
	typeOf := reflect.TypeOf(first)
	for index := 0; index < typeOf.NumField(); index++ {
		kind := typeOf.Field(index).Type.Kind()
		if kind == reflect.Pointer || kind == reflect.Chan || kind == reflect.Func ||
			kind == reflect.Interface || kind == reflect.Map || kind == reflect.Slice {
			t.Fatalf("rendezvous contains capability-like %v", kind)
		}
	}
}

func testAuthorization(t *testing.T, action OrchestrationAction) OrchestrationAuthorizationRequest {
	t.Helper()
	request, err := NewOrchestrationAuthorizationRequest(
		"domain-a", 1, 2, "instance-a", action, 3,
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func testRendezvous(t *testing.T, identity string) StartRendezvous {
	t.Helper()
	rendezvous, err := NewStartRendezvous(identity)
	if err != nil {
		t.Fatal(err)
	}
	return rendezvous
}
