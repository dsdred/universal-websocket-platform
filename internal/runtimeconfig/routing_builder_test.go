package runtimeconfig

import (
	"errors"
	"reflect"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
)

func TestBuilderRoutingPresenceSemantics(t *testing.T) {
	tests := []struct {
		name        string
		routing     *configurationversion.RoutingSettings
		wantPresent bool
	}{
		{name: "absent", routing: nil, wantPresent: false},
		{name: "present empty", routing: &configurationversion.RoutingSettings{}, wantPresent: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version := fullPublishedVersion()
			version.Routing = test.routing
			snapshot, err := NewBuilder().Build(version)
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}
			if (snapshot.Routing != nil) != test.wantPresent {
				t.Fatalf("Routing present = %t, want %t", snapshot.Routing != nil, test.wantPresent)
			}
			if test.wantPresent && (snapshot.Routing.Routes() != nil || snapshot.Routing.DefaultHandlerRef() != "") {
				t.Fatalf("present-empty Routing = %#v", snapshot.Routing)
			}
		})
	}
}

func TestBuilderBuildsNormalizedImmutableRoutingSnapshot(t *testing.T) {
	version := fullPublishedVersion()
	version.Routing = &configurationversion.RoutingSettings{
		DefaultHandlerRef: "\u2003legacy\u2003",
		Routes: []configurationversion.Route{
			{
				ID:         "\u2003second\u2003",
				Enabled:    true,
				Priority:   20,
				HandlerRef: " legacy ",
				Matchers: []configurationversion.Matcher{
					{Type: configurationversion.MatcherTypeMessageType, Value: " text "},
					{Type: configurationversion.MatcherTypePrincipalKind, Value: " authenticated "},
				},
			},
			{
				ID:         "first",
				Enabled:    false,
				Priority:   10,
				HandlerRef: "future",
				Matchers:   []configurationversion.Matcher{},
			},
		},
	}
	before := cloneRoutingForTest(version.Routing)

	snapshot, err := NewBuilder().Build(version)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !reflect.DeepEqual(version.Routing, before) {
		t.Fatalf("Build() mutated ConfigurationVersion Routing:\n got  %#v\n want %#v", version.Routing, before)
	}
	if snapshot.Routing == nil || snapshot.Routing.DefaultHandlerRef() != "legacy" {
		t.Fatalf("Routing = %#v", snapshot.Routing)
	}
	routes := snapshot.Routing.Routes()
	if len(routes) != 2 || routes[0].ID() != "second" || routes[1].ID() != "first" {
		t.Fatalf("Route order = [%v %v], want [second first]", routes[0].ID(), routes[1].ID())
	}
	if routes[0].Priority() != 20 || !routes[0].Enabled() || routes[0].HandlerRef() != "legacy" {
		t.Errorf("first Route = %#v", routes[0])
	}
	matchers := routes[0].Matchers()
	if len(matchers) != 2 ||
		matchers[0].Type() != MatcherTypeMessageType || matchers[0].Value() != "text" ||
		matchers[1].Type() != MatcherTypePrincipalKind || matchers[1].Value() != "authenticated" {
		t.Fatalf("Matchers = %#v", matchers)
	}
	if routes[1].Enabled() || len(routes[1].Matchers()) != 0 {
		t.Errorf("disabled Route = %#v", routes[1])
	}

	version.Routing.DefaultHandlerRef = "changed"
	version.Routing.Routes[0].ID = "changed"
	version.Routing.Routes[0].Matchers[0].Value = "binary"
	if snapshot.Routing.DefaultHandlerRef() != "legacy" || snapshot.Routing.Routes()[0].ID() != "second" || snapshot.Routing.Routes()[0].Matchers()[0].Value() != "text" {
		t.Fatal("Snapshot changed after ConfigurationVersion mutation")
	}
}

func TestRoutingSnapshotAccessorsReturnDetachedSlices(t *testing.T) {
	version := fullPublishedVersion()
	version.Routing = validRoutingForBuilder()
	snapshot, err := NewBuilder().Build(version)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	routes := snapshot.Routing.Routes()
	routes[0].id = "changed"
	routes[0].matchers[0].value = "changed"
	matchers := snapshot.Routing.Routes()[0].Matchers()
	matchers[0].value = "also-changed"

	freshRoute := snapshot.Routing.Routes()[0]
	if freshRoute.ID() != "route" || freshRoute.Matchers()[0].Value() != "text" {
		t.Fatalf("Snapshot aliases accessor results: %#v", freshRoute)
	}
}

func TestBuilderRoutingNormalizationMatchesControlPlaneValidator(t *testing.T) {
	raw := &configurationversion.RoutingSettings{Routes: []configurationversion.Route{{
		ID:         " route ",
		Enabled:    true,
		Priority:   1,
		HandlerRef: " legacy ",
		Matchers: []configurationversion.Matcher{
			{Type: configurationversion.MatcherType(" principal-kind "), Value: " anonymous "},
			{Type: configurationversion.MatcherType(" message-type "), Value: " binary "},
		},
	}}}
	normalized, err := (configurationversion.DefaultRoutingValidator{}).Validate(raw)
	if err != nil {
		t.Fatalf("Control Plane Validate() error = %v", err)
	}
	version := fullPublishedVersion()
	version.Routing = raw
	snapshot, err := NewBuilder().Build(version)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	route := snapshot.Routing.Routes()[0]
	if route.ID() != normalized.Routes[0].ID || route.HandlerRef() != normalized.Routes[0].HandlerRef {
		t.Fatalf("normalized Route = %#v, want %#v", route, normalized.Routes[0])
	}
	for index, matcher := range route.Matchers() {
		if matcher.Type() != MatcherType(normalized.Routes[0].Matchers[index].Type) || matcher.Value() != normalized.Routes[0].Matchers[index].Value {
			t.Fatalf("Matcher[%d] = %#v, want %#v", index, matcher, normalized.Routes[0].Matchers[index])
		}
	}
}

func TestBuilderRejectsMalformedRouting(t *testing.T) {
	tests := []struct {
		name    string
		routing *configurationversion.RoutingSettings
	}{
		{name: "enabled empty Route", routing: &configurationversion.RoutingSettings{Routes: []configurationversion.Route{{ID: "route", Enabled: true, Priority: 1, HandlerRef: "legacy"}}}},
		{name: "unsupported disabled Matcher", routing: &configurationversion.RoutingSettings{Routes: []configurationversion.Route{{ID: "route", Priority: 1, HandlerRef: "legacy", Matchers: []configurationversion.Matcher{{Type: "future", Value: "value"}}}}}},
		{name: "duplicate normalized sets", routing: &configurationversion.RoutingSettings{Routes: []configurationversion.Route{
			{ID: "first", Enabled: true, Priority: 1, HandlerRef: "legacy", Matchers: []configurationversion.Matcher{{Type: configurationversion.MatcherTypeMessageType, Value: "text"}}},
			{ID: "second", Enabled: true, Priority: 2, HandlerRef: "legacy", Matchers: []configurationversion.Matcher{{Type: configurationversion.MatcherType(" message-type "), Value: " text "}}},
		}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version := fullPublishedVersion()
			version.Routing = test.routing
			snapshot, err := NewBuilder().Build(version)
			var validationError *configurationversion.ValidationError
			if err == nil || !errors.As(err, &validationError) || !reflect.DeepEqual(snapshot, Snapshot{}) {
				t.Fatalf("Build() = (%#v, %v), want zero Snapshot and ValidationError", snapshot, err)
			}
		})
	}
}

func TestRoutingSnapshotDoesNotExposeSlices(t *testing.T) {
	for _, typ := range []reflect.Type{
		reflect.TypeOf(RoutingSnapshot{}),
		reflect.TypeOf(RouteSnapshot{}),
		reflect.TypeOf(MatcherSnapshot{}),
	} {
		for index := 0; index < typ.NumField(); index++ {
			field := typ.Field(index)
			if field.IsExported() && field.Type.Kind() == reflect.Slice {
				t.Fatalf("%s exposes mutable slice field %s", typ.Name(), field.Name)
			}
		}
	}
}

func TestZeroRoutingSnapshotSemantics(t *testing.T) {
	var snapshot Snapshot
	if snapshot.Routing != nil {
		t.Fatalf("zero Snapshot Routing = %#v, want nil", snapshot.Routing)
	}
	var routing RoutingSnapshot
	if routing.Routes() != nil || routing.DefaultHandlerRef() != "" {
		t.Fatalf("zero RoutingSnapshot = %#v", routing)
	}
	var route RouteSnapshot
	if route.ID() != "" || route.Enabled() || route.Priority() != 0 || route.Matchers() != nil || route.HandlerRef() != "" {
		t.Fatalf("zero RouteSnapshot = %#v", route)
	}
	var matcher MatcherSnapshot
	if matcher.Type() != "" || matcher.Value() != "" {
		t.Fatalf("zero MatcherSnapshot = %#v", matcher)
	}
}

func validRoutingForBuilder() *configurationversion.RoutingSettings {
	return &configurationversion.RoutingSettings{Routes: []configurationversion.Route{{
		ID:         "route",
		Enabled:    true,
		Priority:   1,
		HandlerRef: "legacy",
		Matchers:   []configurationversion.Matcher{{Type: configurationversion.MatcherTypeMessageType, Value: "text"}},
	}}}
}

func cloneRoutingForTest(source *configurationversion.RoutingSettings) *configurationversion.RoutingSettings {
	if source == nil {
		return nil
	}
	result := &configurationversion.RoutingSettings{DefaultHandlerRef: source.DefaultHandlerRef}
	if source.Routes != nil {
		result.Routes = make([]configurationversion.Route, len(source.Routes))
		for index, route := range source.Routes {
			result.Routes[index] = route
			if route.Matchers != nil {
				result.Routes[index].Matchers = make([]configurationversion.Matcher, len(route.Matchers))
				copy(result.Routes[index].Matchers, route.Matchers)
			}
		}
	}
	return result
}
