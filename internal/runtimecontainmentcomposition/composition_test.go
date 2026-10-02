package runtimecontainmentcomposition

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeactivation"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecontainment"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeexecutionevidence"
	"github.com/dsdred/universal-websocket-platform/internal/runtimemanagement"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

func TestCompositionSurfaceExposesOnlyGatedOperations(t *testing.T) {
	typeOfComposition := reflect.TypeOf(&Composition{})
	want := map[string]bool{
		"ActivateExact": true,
		"QueryEvidence": true,
		"ReplaceExact":  true,
		"RollbackExact": true,
	}
	if typeOfComposition.NumMethod() != len(want) {
		t.Fatalf("exported method count = %d, want %d", typeOfComposition.NumMethod(), len(want))
	}
	for index := 0; index < typeOfComposition.NumMethod(); index++ {
		method := typeOfComposition.Method(index)
		if !want[method.Name] {
			t.Fatalf("unexpected exported method %q", method.Name)
		}
	}
	for index := 0; index < typeOfComposition.Elem().NumField(); index++ {
		if field := typeOfComposition.Elem().Field(index); field.IsExported() {
			t.Fatalf("Composition field %q is exported", field.Name)
		}
	}
}

func TestUnavailableContainmentConstructsNoCapability(t *testing.T) {
	domain := parseTestDomain(t, '1')
	composition, err := New(strings.Repeat("1", 64), domain, runtimecontainment.Result{},
		runtimemanagement.Target{}, nil, nil, nil, nil, nil,
		func(context.Context, runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error { return nil })
	if composition != nil || err == nil {
		t.Fatalf("New(unavailable) = %#v/%v, want nil/error", composition, err)
	}
}

func TestNilCompositionFailsClosed(t *testing.T) {
	var composition *Composition
	if result, err := composition.ActivateExact(context.Background(), runtimeactivation.Request{}); result != (runtimeactivation.Result{}) || err == nil {
		t.Fatalf("ActivateExact(nil) = %#v/%v, want zero/error", result, err)
	}
	outcome := composition.QueryEvidence(context.Background(), parseTestDomain(t, '2'), "instance", "attempt", "generation", runtimeexecutionevidence.GenerationStatus).Consume(context.Background())
	if outcome.Kind() != runtimeexecutionevidence.Unknown || outcome.Reason() != runtimeexecutionevidence.ScopeMismatch {
		t.Fatalf("QueryEvidence(nil) = %v/%v, want Unknown/ScopeMismatch", outcome.Kind(), outcome.Reason())
	}
}

func parseTestDomain(t *testing.T, digit byte) runtimecontainment.Domain {
	t.Helper()
	domain, err := runtimecontainment.ParseDomain(strings.Repeat(string(digit), 64))
	if err != nil {
		t.Fatal(err)
	}
	return domain
}
