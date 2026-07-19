package configurationversion

import (
	"cmp"
	"slices"
	"strings"
)

const (
	maximumRoutes           = 256
	maximumMatchersPerRoute = 4
	maximumRoutingIDLength  = 128
)

// RoutingValidator validates Routing metadata and returns a detached canonical copy.
type RoutingValidator interface {
	Validate(*RoutingSettings) (*RoutingSettings, error)
}

// DefaultRoutingValidator applies the declarative DP-005 Routing rules.
type DefaultRoutingValidator struct{}

// Validate preserves absence and returns a detached normalized Routing value when present.
func (DefaultRoutingValidator) Validate(routing *RoutingSettings) (*RoutingSettings, error) {
	if routing == nil {
		return nil, nil
	}
	if len(routing.Routes) > maximumRoutes {
		return nil, routingValidationError("routing.routes", "must not contain more than 256 Routes")
	}

	normalized := cloneRoutingSettings(routing)
	defaultHandlerPresent := normalized.DefaultHandlerRef != ""
	normalized.DefaultHandlerRef = strings.TrimSpace(normalized.DefaultHandlerRef)
	if defaultHandlerPresent && !validRoutingIdentifier(normalized.DefaultHandlerRef) {
		return nil, routingValidationError("routing.defaultHandlerRef", "must be a valid Handler reference")
	}

	routeIDs := make(map[string]struct{}, len(normalized.Routes))
	priorities := make(map[uint32]struct{}, len(normalized.Routes))
	enabledMatcherSets := make(map[normalizedMatcherSet]string, len(normalized.Routes))
	for index := range normalized.Routes {
		route := &normalized.Routes[index]
		route.ID = strings.TrimSpace(route.ID)
		if !validRoutingIdentifier(route.ID) {
			return nil, routingValidationError("routing.routes.id", "must be a valid Route identifier")
		}
		if _, exists := routeIDs[route.ID]; exists {
			return nil, routingValidationError("routing.routes.id", "must be unique")
		}
		routeIDs[route.ID] = struct{}{}

		if route.Priority == 0 {
			return nil, routingValidationError("routing.routes.priority", "must be positive")
		}
		if _, exists := priorities[route.Priority]; exists {
			return nil, routingValidationError("routing.routes.priority", "must be unique")
		}
		priorities[route.Priority] = struct{}{}

		route.HandlerRef = strings.TrimSpace(route.HandlerRef)
		if !validRoutingIdentifier(route.HandlerRef) {
			return nil, routingValidationError("routing.routes.handlerRef", "must be a valid Handler reference")
		}
		if len(route.Matchers) > maximumMatchersPerRoute {
			return nil, routingValidationError("routing.routes.matchers", "must not contain more than four Matchers")
		}
		if route.Enabled && len(route.Matchers) == 0 {
			return nil, routingValidationError("routing.routes.matchers", "must not be empty for an enabled Route")
		}

		matcherTypes := make(map[MatcherType]struct{}, len(route.Matchers))
		for matcherIndex := range route.Matchers {
			matcher := &route.Matchers[matcherIndex]
			matcher.Type = MatcherType(strings.TrimSpace(string(matcher.Type)))
			matcher.Value = strings.TrimSpace(matcher.Value)
			if _, exists := matcherTypes[matcher.Type]; exists {
				return nil, routingValidationError("routing.routes.matchers.type", "must be unique within a Route")
			}
			matcherTypes[matcher.Type] = struct{}{}
			if err := validateMatcher(*matcher); err != nil {
				return nil, err
			}
		}
		slices.SortFunc(route.Matchers, compareMatchers)

		if route.Enabled {
			matcherSet := newNormalizedMatcherSet(route.Matchers)
			if existingRouteID, exists := enabledMatcherSets[matcherSet]; exists {
				return nil, routingValidationError(
					"routing.routes.matchers",
					"must not duplicate enabled Route "+existingRouteID,
				)
			}
			enabledMatcherSets[matcherSet] = route.ID
		}
	}

	return normalized, nil
}

func validateMatcher(matcher Matcher) error {
	if matcher.Type == "" {
		return routingValidationError("routing.routes.matchers.type", "must not be empty")
	}
	if matcher.Value == "" {
		return routingValidationError("routing.routes.matchers.value", "must not be empty")
	}

	switch matcher.Type {
	case MatcherTypeMessageType:
		if matcher.Value != "text" && matcher.Value != "binary" {
			return routingValidationError("routing.routes.matchers.value", "must be text or binary for message-type")
		}
	case MatcherTypePrincipalKind:
		if matcher.Value != "authenticated" && matcher.Value != "anonymous" {
			return routingValidationError("routing.routes.matchers.value", "must be authenticated or anonymous for principal-kind")
		}
	case MatcherTypeAuthenticationType:
		if matcher.Value != "jwt" && matcher.Value != "api-key" && matcher.Value != "basic" {
			return routingValidationError("routing.routes.matchers.value", "must be jwt, api-key, or basic for authentication-type")
		}
	case MatcherTypeAuthenticationProvider:
		// Provider names preserve their normalized spelling and case.
	default:
		return routingValidationError("routing.routes.matchers.type", "is not supported")
	}
	return nil
}

func validRoutingIdentifier(identifier string) bool {
	if len(identifier) == 0 || len(identifier) > maximumRoutingIDLength || !asciiLetter(identifier[0]) {
		return false
	}
	for index := 1; index < len(identifier); index++ {
		character := identifier[index]
		if !asciiLetter(character) && !asciiDigit(character) && character != '.' && character != '_' && character != '-' {
			return false
		}
	}
	return true
}

func asciiLetter(character byte) bool {
	return character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z'
}

func asciiDigit(character byte) bool {
	return character >= '0' && character <= '9'
}

func compareMatchers(first, second Matcher) int {
	if result := cmp.Compare(first.Type, second.Type); result != 0 {
		return result
	}
	return cmp.Compare(first.Value, second.Value)
}

type normalizedMatcherSet struct {
	count    uint8
	matchers [maximumMatchersPerRoute]normalizedMatcher
}

type normalizedMatcher struct {
	typeName MatcherType
	value    string
}

func newNormalizedMatcherSet(matchers []Matcher) normalizedMatcherSet {
	result := normalizedMatcherSet{count: uint8(len(matchers))}
	for index, matcher := range matchers {
		result.matchers[index] = normalizedMatcher{typeName: matcher.Type, value: matcher.Value}
	}
	return result
}

func cloneRoutingSettings(routing *RoutingSettings) *RoutingSettings {
	if routing == nil {
		return nil
	}
	clone := &RoutingSettings{DefaultHandlerRef: routing.DefaultHandlerRef}
	if routing.Routes != nil {
		clone.Routes = make([]Route, len(routing.Routes))
		for index, route := range routing.Routes {
			clone.Routes[index] = route
			clone.Routes[index].Matchers = append([]Matcher(nil), route.Matchers...)
		}
	}
	return clone
}

func routingValidationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
