package configurationversion

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestDefaultRoutingValidatorPresenceSemantics(t *testing.T) {
	validator := DefaultRoutingValidator{}

	absent, err := validator.Validate(nil)
	if err != nil || absent != nil {
		t.Fatalf("Validate(nil) = (%#v, %v), want (nil, nil)", absent, err)
	}

	present, err := validator.Validate(&RoutingSettings{})
	if err != nil {
		t.Fatalf("Validate(empty) error = %v", err)
	}
	if present == nil || present.Routes != nil || present.DefaultHandlerRef != "" {
		t.Fatalf("Validate(empty) = %#v, want present empty Routing", present)
	}
}

func TestDefaultRoutingValidatorNormalizesDetachedCopy(t *testing.T) {
	input := &RoutingSettings{
		DefaultHandlerRef: "\u2003legacy\u2003",
		Routes: []Route{{
			ID:         "\u2003route.one\u2003",
			Enabled:    true,
			Priority:   1,
			HandlerRef: "\u2003legacy\u2003",
			Matchers: []Matcher{
				{Type: MatcherType("\u2003principal-kind\u2003"), Value: "\u2003authenticated\u2003"},
				{Type: MatcherType(" authentication-provider "), Value: "\u2003Provider  One\u2003"},
				{Type: MatcherType(" message-type "), Value: " text "},
			},
		}},
	}
	wantInput := cloneRoutingSettings(input)

	got, err := (DefaultRoutingValidator{}).Validate(input)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !reflect.DeepEqual(input, wantInput) {
		t.Fatalf("Validate() mutated input:\n got  %#v\n want %#v", input, wantInput)
	}
	if got.DefaultHandlerRef != "legacy" || got.Routes[0].ID != "route.one" || got.Routes[0].HandlerRef != "legacy" {
		t.Errorf("normalized identifiers = %#v", got)
	}
	wantMatchers := []Matcher{
		{Type: MatcherTypeAuthenticationProvider, Value: "Provider  One"},
		{Type: MatcherTypeMessageType, Value: "text"},
		{Type: MatcherTypePrincipalKind, Value: "authenticated"},
	}
	if !reflect.DeepEqual(got.Routes[0].Matchers, wantMatchers) {
		t.Errorf("Matchers = %#v, want %#v", got.Routes[0].Matchers, wantMatchers)
	}

	got.Routes[0].Matchers[0].Value = "changed"
	got.Routes[0].ID = "changed"
	if !reflect.DeepEqual(input, wantInput) {
		t.Fatal("mutating normalized result changed caller-owned Routing")
	}
}

func TestDefaultRoutingValidatorAcceptsSupportedMatchers(t *testing.T) {
	routing := routes(validRoute("all", 1,
		Matcher{Type: MatcherTypeMessageType, Value: "binary"},
		Matcher{Type: MatcherTypePrincipalKind, Value: "anonymous"},
		Matcher{Type: MatcherTypeAuthenticationType, Value: "api-key"},
		Matcher{Type: MatcherTypeAuthenticationProvider, Value: "Provider"},
	))
	if _, err := (DefaultRoutingValidator{}).Validate(routing); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestDefaultRoutingValidatorAcceptsRouteLimit(t *testing.T) {
	routing := &RoutingSettings{Routes: make([]Route, maximumRoutes)}
	for index := range routing.Routes {
		routing.Routes[index] = disabledRoute(fmt.Sprintf("route.%d", index), uint32(index+1))
	}
	if _, err := (DefaultRoutingValidator{}).Validate(routing); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestDefaultRoutingValidatorRejectsInvalidRouting(t *testing.T) {
	tooManyRoutes := make([]Route, maximumRoutes+1)
	fiveMatchers := []Matcher{
		{Type: MatcherTypeMessageType, Value: "text"},
		{Type: MatcherTypePrincipalKind, Value: "anonymous"},
		{Type: MatcherTypeAuthenticationType, Value: "jwt"},
		{Type: MatcherTypeAuthenticationProvider, Value: "provider"},
		{Type: MatcherType("future"), Value: "value"},
	}
	longID := "a" + strings.Repeat("b", maximumRoutingIDLength)

	tests := []struct {
		name    string
		routing *RoutingSettings
	}{
		{name: "route limit", routing: &RoutingSettings{Routes: tooManyRoutes}},
		{name: "empty route id", routing: routes(validRoute("", 1, Matcher{Type: MatcherTypeMessageType, Value: "text"}))},
		{name: "invalid route id start", routing: routes(validRoute("1route", 1, Matcher{Type: MatcherTypeMessageType, Value: "text"}))},
		{name: "invalid route id character", routing: routes(validRoute("route one", 1, Matcher{Type: MatcherTypeMessageType, Value: "text"}))},
		{name: "non ascii route id", routing: routes(validRoute("routeЖ", 1, Matcher{Type: MatcherTypeMessageType, Value: "text"}))},
		{name: "long route id", routing: routes(validRoute(longID, 1, Matcher{Type: MatcherTypeMessageType, Value: "text"}))},
		{name: "duplicate normalized ids", routing: routes(
			validRoute("route", 1, Matcher{Type: MatcherTypeMessageType, Value: "text"}),
			validRoute(" route ", 2, Matcher{Type: MatcherTypeMessageType, Value: "binary"}),
		)},
		{name: "zero priority", routing: routes(validRoute("route", 0, Matcher{Type: MatcherTypeMessageType, Value: "text"}))},
		{name: "duplicate disabled priority", routing: routes(
			validRoute("first", 1, Matcher{Type: MatcherTypeMessageType, Value: "text"}),
			disabledRoute("second", 1),
		)},
		{name: "empty handler reference", routing: routes(validRouteWithHandler("route", 1, "", Matcher{Type: MatcherTypeMessageType, Value: "text"}))},
		{name: "invalid disabled handler reference", routing: routes(Route{ID: "disabled", Priority: 1, HandlerRef: "bad ref"})},
		{name: "invalid default handler reference", routing: &RoutingSettings{DefaultHandlerRef: "bad ref"}},
		{name: "empty normalized default handler reference", routing: &RoutingSettings{DefaultHandlerRef: "\u2003"}},
		{name: "matcher limit", routing: routes(validRoute("route", 1, fiveMatchers...))},
		{name: "enabled empty matchers", routing: routes(validRoute("route", 1))},
		{name: "duplicate matcher type", routing: routes(validRoute("route", 1,
			Matcher{Type: MatcherTypeMessageType, Value: "text"},
			Matcher{Type: MatcherType(" message-type "), Value: "binary"},
		))},
		{name: "missing matcher type", routing: routes(validRoute("route", 1, Matcher{Value: "text"}))},
		{name: "missing matcher value", routing: routes(validRoute("route", 1, Matcher{Type: MatcherTypeMessageType}))},
		{name: "unsupported disabled matcher", routing: routes(Route{ID: "disabled", Priority: 1, HandlerRef: "legacy", Matchers: []Matcher{{Type: "future", Value: "value"}}})},
		{name: "message type case sensitive", routing: routes(validRoute("route", 1, Matcher{Type: MatcherTypeMessageType, Value: "Text"}))},
		{name: "principal kind case sensitive", routing: routes(validRoute("route", 1, Matcher{Type: MatcherTypePrincipalKind, Value: "Authenticated"}))},
		{name: "authentication type case sensitive", routing: routes(validRoute("route", 1, Matcher{Type: MatcherTypeAuthenticationType, Value: "JWT"}))},
		{name: "matcher type case sensitive", routing: routes(validRoute("route", 1, Matcher{Type: "Message-Type", Value: "text"}))},
	}

	validator := DefaultRoutingValidator{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got, err := validator.Validate(test.routing); err == nil || got != nil {
				t.Fatalf("Validate() = (%#v, %v), want (nil, validation error)", got, err)
			}
		})
	}
}

func TestDefaultRoutingValidatorDuplicateNormalizedMatcherSets(t *testing.T) {
	first := validRoute("first", 1,
		Matcher{Type: MatcherTypeMessageType, Value: "text"},
		Matcher{Type: MatcherTypePrincipalKind, Value: "authenticated"},
	)
	second := validRoute("second", 2,
		Matcher{Type: MatcherType(" principal-kind "), Value: " authenticated "},
		Matcher{Type: MatcherType(" message-type "), Value: " text "},
	)

	if got, err := (DefaultRoutingValidator{}).Validate(routes(first, second)); err == nil || got != nil {
		t.Fatalf("Validate(duplicate enabled sets) = (%#v, %v), want error", got, err)
	}

	second.Enabled = false
	if _, err := (DefaultRoutingValidator{}).Validate(routes(first, second)); err != nil {
		t.Fatalf("Validate(disabled duplicate set) error = %v", err)
	}
	first.Matchers[1].Value = "Provider"
	first.Matchers[1].Type = MatcherTypeAuthenticationProvider
	second.Enabled = true
	second.Matchers = []Matcher{{Type: MatcherTypeAuthenticationProvider, Value: "provider"}}
	if _, err := (DefaultRoutingValidator{}).Validate(routes(first, second)); err != nil {
		t.Fatalf("Validate(case-distinct provider values) error = %v", err)
	}
}

func TestDefaultRoutingValidatorAllowsStructurallyValidDisabledEmptyRoute(t *testing.T) {
	if _, err := (DefaultRoutingValidator{}).Validate(routes(disabledRoute("disabled", 1))); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func validRoute(id string, priority uint32, matchers ...Matcher) Route {
	return validRouteWithHandler(id, priority, "legacy", matchers...)
}

func validRouteWithHandler(id string, priority uint32, handlerRef string, matchers ...Matcher) Route {
	return Route{ID: id, Enabled: true, Priority: priority, Matchers: matchers, HandlerRef: handlerRef}
}

func disabledRoute(id string, priority uint32) Route {
	return Route{ID: id, Priority: priority, HandlerRef: "legacy"}
}

func routes(values ...Route) *RoutingSettings {
	return &RoutingSettings{Routes: values}
}
