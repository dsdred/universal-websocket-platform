package router

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

func TestNewCompilesImmutableRoutingTable(t *testing.T) {
	handler := &handlerStub{}
	snapshot := buildRoutingSnapshot(t, &configurationversion.RoutingSettings{
		DefaultHandlerRef: "legacy",
		Routes: []configurationversion.Route{
			{
				ID:         "second",
				Enabled:    true,
				Priority:   20,
				HandlerRef: "legacy",
				Matchers: []configurationversion.Matcher{
					{Type: configurationversion.MatcherTypePrincipalKind, Value: "authenticated"},
					{Type: configurationversion.MatcherTypeMessageType, Value: "text"},
				},
			},
			{
				ID:         "first",
				Enabled:    true,
				Priority:   10,
				HandlerRef: "legacy",
				Matchers:   []configurationversion.Matcher{{Type: configurationversion.MatcherTypeMessageType, Value: "binary"}},
			},
		},
	})

	compiled, err := New(snapshot, map[string]message.Handler{legacyHandlerRef: handler})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if len(compiled.routes) != 2 || compiled.routes[0].id != "first" || compiled.routes[1].id != "second" {
		t.Fatalf("compiled Route order = %#v", compiled.routes)
	}
	if compiled.routes[0].priority != 10 || compiled.routes[0].handlerRef != legacyHandlerRef || compiled.routes[0].handler != handler {
		t.Errorf("compiled first Route = %#v", compiled.routes[0])
	}
	if got := compiled.routes[1].matchers; len(got) != 2 ||
		got[0] != (compiledMatcher{matcherType: runtimeconfig.MatcherTypeMessageType, value: "text"}) ||
		got[1] != (compiledMatcher{matcherType: runtimeconfig.MatcherTypePrincipalKind, value: "authenticated"}) {
		t.Fatalf("compiled Matchers = %#v", got)
	}
	if compiled.defaultHandler == nil || compiled.defaultHandler.reference != legacyHandlerRef || compiled.defaultHandler.handler != handler {
		t.Fatalf("compiled default Handler = %#v", compiled.defaultHandler)
	}
}

func TestNewOrdersEnabledRoutesOnlyByAscendingPriority(t *testing.T) {
	declarations := [][]configurationversion.Route{
		{
			enabledRoute("high", 30, "text"),
			{ID: "disabled", Priority: 5, HandlerRef: "future", Matchers: []configurationversion.Matcher{}},
			enabledRoute("low", 10, "binary"),
			enabledRouteWithPrincipal("middle", 20, "text", "anonymous"),
		},
		{
			enabledRouteWithPrincipal("middle", 20, "text", "anonymous"),
			enabledRoute("high", 30, "text"),
			enabledRoute("low", 10, "binary"),
			{ID: "disabled", Priority: 5, HandlerRef: "future", Matchers: []configurationversion.Matcher{}},
		},
	}

	for declarationIndex, routes := range declarations {
		snapshot := buildRoutingSnapshot(t, &configurationversion.RoutingSettings{Routes: routes})
		for attempt := 0; attempt < 5; attempt++ {
			compiled, err := New(snapshot, map[string]message.Handler{legacyHandlerRef: &handlerStub{}})
			if err != nil {
				t.Fatalf("New(declaration %d, attempt %d) error = %v", declarationIndex, attempt, err)
			}
			if len(compiled.routes) != 3 ||
				compiled.routes[0].id != "low" || compiled.routes[0].priority != 10 ||
				compiled.routes[1].id != "middle" || compiled.routes[1].priority != 20 ||
				compiled.routes[2].id != "high" || compiled.routes[2].priority != 30 {
				t.Fatalf("compiled order for declaration %d, attempt %d = %#v", declarationIndex, attempt, compiled.routes)
			}
		}
	}
}

func TestNewAllowsSharedLegacyHandlerInstance(t *testing.T) {
	handler := &handlerStub{}
	snapshot := buildRoutingSnapshot(t, &configurationversion.RoutingSettings{
		DefaultHandlerRef: legacyHandlerRef,
		Routes: []configurationversion.Route{
			enabledRoute("text", 1, "text"),
			enabledRoute("binary", 2, "binary"),
		},
	})

	compiled, err := New(snapshot, map[string]message.Handler{legacyHandlerRef: handler})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if compiled.routes[0].handler != handler || compiled.routes[1].handler != handler || compiled.defaultHandler.handler != handler {
		t.Fatal("compiled references do not share the supplied Handler instance")
	}
}

func TestNewRejectsUnresolvedActiveHandlerRef(t *testing.T) {
	snapshot := buildRoutingSnapshot(t, &configurationversion.RoutingSettings{Routes: []configurationversion.Route{{
		ID:         "route",
		Enabled:    true,
		Priority:   1,
		HandlerRef: "future",
		Matchers:   []configurationversion.Matcher{{Type: configurationversion.MatcherTypeMessageType, Value: "text"}},
	}}})

	compiled, err := New(snapshot, map[string]message.Handler{
		legacyHandlerRef: &handlerStub{},
		"future":         &handlerStub{},
	})
	if compiled != nil || !errors.Is(err, ErrUnresolvedHandlerRef) {
		t.Fatalf("New() = (%#v, %v), want ErrUnresolvedHandlerRef", compiled, err)
	}
}

func TestNewRejectsUnresolvedDefaultHandlerRef(t *testing.T) {
	snapshot := buildRoutingSnapshot(t, &configurationversion.RoutingSettings{DefaultHandlerRef: "future"})

	compiled, err := New(snapshot, map[string]message.Handler{"future": &handlerStub{}})
	if compiled != nil || !errors.Is(err, ErrUnresolvedHandlerRef) {
		t.Fatalf("New() = (%#v, %v), want ErrUnresolvedHandlerRef", compiled, err)
	}
}

func TestNewDoesNotResolveDisabledRoute(t *testing.T) {
	snapshot := buildRoutingSnapshot(t, &configurationversion.RoutingSettings{Routes: []configurationversion.Route{{
		ID:         "disabled",
		Enabled:    false,
		Priority:   1,
		HandlerRef: "future",
		Matchers:   []configurationversion.Matcher{},
	}}})

	compiled, err := New(snapshot, nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if len(compiled.routes) != 0 || compiled.defaultHandler != nil {
		t.Fatalf("compiled Router = %#v, want empty table", compiled)
	}
}

func TestNewDoesNotRetainRegistry(t *testing.T) {
	original := &handlerStub{}
	replacement := &handlerStub{}
	registry := map[string]message.Handler{legacyHandlerRef: original}
	snapshot := buildRoutingSnapshot(t, &configurationversion.RoutingSettings{Routes: []configurationversion.Route{
		enabledRoute("route", 1, "text"),
	}})

	compiled, err := New(snapshot, registry)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	registry[legacyHandlerRef] = replacement
	delete(registry, legacyHandlerRef)
	if compiled.routes[0].handler != original {
		t.Fatal("compiled Router changed after registry mutation")
	}

	typ := reflect.TypeOf(*compiled)
	for index := 0; index < typ.NumField(); index++ {
		if typ.Field(index).Type.Kind() == reflect.Map {
			t.Fatalf("Router retains registry-like map field %s", typ.Field(index).Name)
		}
	}
}

func TestNewOwnsIndependentCompiledMemory(t *testing.T) {
	snapshot := buildRoutingSnapshot(t, &configurationversion.RoutingSettings{Routes: []configurationversion.Route{
		enabledRoute("text", 1, "text"),
		enabledRoute("binary", 2, "binary"),
	}})
	registry := map[string]message.Handler{legacyHandlerRef: &handlerStub{}}
	first, err := New(snapshot, registry)
	if err != nil {
		t.Fatalf("New(first) error = %v", err)
	}
	second, err := New(snapshot, registry)
	if err != nil {
		t.Fatalf("New(second) error = %v", err)
	}

	first.routes[0].matchers[0].value = "changed"
	first.routes[1].matchers[0].value = "also-changed"
	if second.routes[0].matchers[0].value != "text" || second.routes[1].matchers[0].value != "binary" {
		t.Fatal("separate Router compilations share Matcher memory")
	}
}

func TestNewIsolatedFromDetachedSourceViews(t *testing.T) {
	snapshot := buildRoutingSnapshot(t, &configurationversion.RoutingSettings{Routes: []configurationversion.Route{
		enabledRoute("first", 1, "text"),
		enabledRoute("second", 2, "binary"),
	}})
	sourceRoutes := snapshot.Routes()
	sourceMatchers := sourceRoutes[0].Matchers()
	compiled, err := New(snapshot, map[string]message.Handler{legacyHandlerRef: &handlerStub{}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	sourceRoutes[0], sourceRoutes[1] = sourceRoutes[1], sourceRoutes[0]
	sourceMatchers[0] = sourceRoutes[0].Matchers()[0]
	if compiled.routes[0].id != "first" || compiled.routes[0].matchers[0].value != "text" {
		t.Fatal("compiled Router aliases detached Snapshot accessor results")
	}
}

func TestNewConstructorValidation(t *testing.T) {
	compiled, err := New(nil, nil)
	if compiled != nil || !errors.Is(err, ErrInvalidRoutingSnapshot) {
		t.Fatalf("New(nil) = (%#v, %v), want ErrInvalidRoutingSnapshot", compiled, err)
	}

	empty := buildRoutingSnapshot(t, &configurationversion.RoutingSettings{})
	compiled, err = New(empty, nil)
	if err != nil || compiled == nil || len(compiled.routes) != 0 {
		t.Fatalf("New(empty) = (%#v, %v), want empty Router", compiled, err)
	}

	active := buildRoutingSnapshot(t, &configurationversion.RoutingSettings{Routes: []configurationversion.Route{
		enabledRoute("route", 1, "text"),
	}})
	var typedNil *handlerStub
	compiled, err = New(active, map[string]message.Handler{legacyHandlerRef: typedNil})
	if compiled != nil || !errors.Is(err, ErrUnresolvedHandlerRef) {
		t.Fatalf("New(typed nil Handler) = (%#v, %v), want ErrUnresolvedHandlerRef", compiled, err)
	}
}

func TestCompiledInvariantValidation(t *testing.T) {
	if err := validateCompiled(nil); !errors.Is(err, ErrImpossibleCompiledState) {
		t.Fatalf("validateCompiled(nil) error = %v, want ErrImpossibleCompiledState", err)
	}
	invalid := &Router{routes: []compiledRoute{{id: "route"}}}
	if err := validateCompiled(invalid); !errors.Is(err, ErrImpossibleCompiledState) {
		t.Fatalf("validateCompiled(invalid) error = %v, want ErrImpossibleCompiledState", err)
	}
}

func TestDefensiveSnapshotValidationRejectsMalformedValues(t *testing.T) {
	if err := validateRoute(runtimeconfig.RouteSnapshot{}, map[string]struct{}{}, map[uint32]struct{}{}); !errors.Is(err, ErrInvalidRoutingSnapshot) {
		t.Fatalf("validateRoute(zero) error = %v, want ErrInvalidRoutingSnapshot", err)
	}
	if _, err := validateMatchers(nil, true); !errors.Is(err, ErrInvalidRoutingSnapshot) {
		t.Fatalf("validateMatchers(empty enabled) error = %v, want ErrInvalidRoutingSnapshot", err)
	}
	if err := validateMatcher(compiledMatcher{}); !errors.Is(err, ErrInvalidRoutingSnapshot) {
		t.Fatalf("validateMatcher(zero) error = %v, want ErrInvalidRoutingSnapshot", err)
	}
}

func TestCompiledRouterDoesNotExposeMutableCollectionsOrExecution(t *testing.T) {
	for _, typ := range []reflect.Type{
		reflect.TypeOf(Router{}),
		reflect.TypeOf(compiledRoute{}),
		reflect.TypeOf(compiledMatcher{}),
		reflect.TypeOf(compiledHandler{}),
	} {
		for index := 0; index < typ.NumField(); index++ {
			field := typ.Field(index)
			if field.IsExported() && (field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Map) {
				t.Fatalf("%s exposes mutable collection field %s", typ.Name(), field.Name)
			}
		}
	}
	if reflect.TypeOf((*Router)(nil)).NumMethod() != 0 {
		t.Fatal("Router exposes behavior before the message-routing step")
	}
}

func buildRoutingSnapshot(t *testing.T, routing *configurationversion.RoutingSettings) *runtimeconfig.RoutingSnapshot {
	t.Helper()
	snapshot, err := runtimeconfig.NewBuilder().Build(configurationversion.ConfigurationVersion{
		ID:              1,
		ConfigurationID: 1,
		State:           configurationversion.Published,
		Routing:         routing,
	})
	if err != nil {
		t.Fatalf("runtimeconfig.Build() error = %v", err)
	}
	return snapshot.Routing
}

func enabledRoute(id string, priority uint32, messageType string) configurationversion.Route {
	return configurationversion.Route{
		ID:         id,
		Enabled:    true,
		Priority:   priority,
		HandlerRef: legacyHandlerRef,
		Matchers: []configurationversion.Matcher{{
			Type:  configurationversion.MatcherTypeMessageType,
			Value: messageType,
		}},
	}
}

func enabledRouteWithPrincipal(id string, priority uint32, messageType, principalKind string) configurationversion.Route {
	route := enabledRoute(id, priority, messageType)
	route.Matchers = append(route.Matchers, configurationversion.Matcher{
		Type:  configurationversion.MatcherTypePrincipalKind,
		Value: principalKind,
	})
	return route
}

type handlerStub struct{}

func (*handlerStub) Handle(context.Context, message.Context) error { return nil }
