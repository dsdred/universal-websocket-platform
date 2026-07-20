// Package router compiles immutable Runtime Message routing metadata.
package router

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

const (
	legacyHandlerRef       = "legacy"
	maximumRoutes          = 256
	maximumMatchers        = 4
	maximumIdentifierBytes = 128
)

var (
	// ErrInvalidRoutingSnapshot identifies malformed Router construction input.
	ErrInvalidRoutingSnapshot = errors.New("invalid routing snapshot")
	// ErrUnresolvedHandlerRef identifies an active Handler reference without a supported binding.
	ErrUnresolvedHandlerRef = errors.New("unresolved handler reference")
	// ErrImpossibleCompiledState identifies an internal compiler invariant violation.
	ErrImpossibleCompiledState = errors.New("impossible compiled router state")
)

// Router owns an immutable compiled routing table and implements message.Handler.
type Router struct {
	routes         []compiledRoute
	defaultHandler *compiledHandler
}

var _ message.Handler = (*Router)(nil)

// Handle selects the first matching compiled Route and synchronously invokes exactly one Handler.
func (r *Router) Handle(ctx context.Context, runtimeContext message.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !validRuntimeContext(runtimeContext) {
		return message.ErrInvalidContext
	}

	for index := range r.routes {
		route := &r.routes[index]
		if route.matches(runtimeContext) {
			return route.handler.Handle(ctx, runtimeContext)
		}
	}
	if r.defaultHandler != nil {
		return r.defaultHandler.handler.Handle(ctx, runtimeContext)
	}
	return nil
}

func (r *compiledRoute) matches(runtimeContext message.Context) bool {
	for index := range r.matchers {
		if !r.matchers[index].matches(runtimeContext) {
			return false
		}
	}
	return true
}

func (m *compiledMatcher) matches(runtimeContext message.Context) bool {
	switch m.matcherType {
	case runtimeconfig.MatcherTypeMessageType:
		return string(runtimeContext.MessageType()) == m.value
	case runtimeconfig.MatcherTypePrincipalKind:
		switch m.value {
		case "authenticated":
			return runtimeContext.Authenticated()
		case "anonymous":
			return runtimeContext.Anonymous()
		default:
			return false
		}
	case runtimeconfig.MatcherTypeAuthenticationType:
		return runtimeContext.AuthenticationType() == m.value
	case runtimeconfig.MatcherTypeAuthenticationProvider:
		return runtimeContext.AuthenticationProvider() == m.value
	default:
		return false
	}
}

func validRuntimeContext(runtimeContext message.Context) bool {
	messageType := runtimeContext.MessageType()
	if messageType != message.TypeText && messageType != message.TypeBinary {
		return false
	}
	if runtimeContext.Sender() == nil || runtimeContext.SessionID() == "" || runtimeContext.Authenticated() == runtimeContext.Anonymous() {
		return false
	}
	if runtimeContext.Authenticated() {
		return runtimeContext.AuthenticationType() != "" && runtimeContext.AuthenticationProvider() != ""
	}
	return runtimeContext.AuthenticationType() == "" && runtimeContext.AuthenticationProvider() == ""
}

type compiledRoute struct {
	id         string
	priority   uint32
	matchers   []compiledMatcher
	handlerRef string
	handler    message.Handler
}

type compiledMatcher struct {
	matcherType runtimeconfig.MatcherType
	value       string
}

type compiledHandler struct {
	reference string
	handler   message.Handler
}

type normalizedMatcherSet struct {
	count    uint8
	matchers [maximumMatchers]compiledMatcher
}

// New compiles one immutable Routing Snapshot using the caller's finite Handler registry.
// The initial compiler recognizes only the legacy Handler reference and does not retain registry.
func New(snapshot *runtimeconfig.RoutingSnapshot, registry map[string]message.Handler) (*Router, error) {
	if snapshot == nil {
		return nil, invalidSnapshot("Routing is absent")
	}

	routes := snapshot.Routes()
	if len(routes) > maximumRoutes {
		return nil, invalidSnapshot("too many Routes")
	}
	if reference := snapshot.DefaultHandlerRef(); reference != "" && !validIdentifier(reference) {
		return nil, invalidSnapshot("invalid DefaultHandlerRef")
	}

	routeIDs := make(map[string]struct{}, len(routes))
	priorities := make(map[uint32]struct{}, len(routes))
	enabledMatcherSets := make(map[normalizedMatcherSet]struct{}, len(routes))
	compiledRoutes := make([]compiledRoute, 0, len(routes))

	for _, route := range routes {
		if err := validateRoute(route, routeIDs, priorities); err != nil {
			return nil, err
		}

		matchers := route.Matchers()
		set, err := validateMatchers(matchers, route.Enabled())
		if err != nil {
			return nil, fmt.Errorf("%w: Route %s", err, route.ID())
		}
		if !route.Enabled() {
			continue
		}
		if _, duplicate := enabledMatcherSets[set]; duplicate {
			return nil, invalidSnapshot("duplicate enabled Matcher set")
		}
		enabledMatcherSets[set] = struct{}{}

		handler, ok := resolveHandler(registry, route.HandlerRef())
		if !ok {
			return nil, unresolvedHandler(route.HandlerRef())
		}
		compiledRoutes = append(compiledRoutes, compiledRoute{
			id:         route.ID(),
			priority:   route.Priority(),
			matchers:   compileMatchers(matchers),
			handlerRef: route.HandlerRef(),
			handler:    handler,
		})
	}
	slices.SortFunc(compiledRoutes, func(first, second compiledRoute) int {
		return cmp.Compare(first.priority, second.priority)
	})

	compiled := &Router{routes: compiledRoutes}
	if reference := snapshot.DefaultHandlerRef(); reference != "" {
		handler, ok := resolveHandler(registry, reference)
		if !ok {
			return nil, unresolvedHandler(reference)
		}
		compiled.defaultHandler = &compiledHandler{reference: reference, handler: handler}
	}
	if err := validateCompiled(compiled); err != nil {
		return nil, err
	}
	return compiled, nil
}

func validateRoute(route runtimeconfig.RouteSnapshot, routeIDs map[string]struct{}, priorities map[uint32]struct{}) error {
	if !validIdentifier(route.ID()) {
		return invalidSnapshot("invalid Route ID")
	}
	if _, duplicate := routeIDs[route.ID()]; duplicate {
		return invalidSnapshot("duplicate Route ID")
	}
	routeIDs[route.ID()] = struct{}{}

	if route.Priority() == 0 {
		return invalidSnapshot("non-positive Route Priority")
	}
	if _, duplicate := priorities[route.Priority()]; duplicate {
		return invalidSnapshot("duplicate Route Priority")
	}
	priorities[route.Priority()] = struct{}{}

	if !validIdentifier(route.HandlerRef()) {
		return invalidSnapshot("invalid HandlerRef")
	}
	return nil
}

func validateMatchers(matchers []runtimeconfig.MatcherSnapshot, enabled bool) (normalizedMatcherSet, error) {
	if len(matchers) > maximumMatchers {
		return normalizedMatcherSet{}, invalidSnapshot("too many Matchers")
	}
	if enabled && len(matchers) == 0 {
		return normalizedMatcherSet{}, invalidSnapshot("enabled Route has no Matchers")
	}

	result := normalizedMatcherSet{count: uint8(len(matchers))}
	seen := make(map[runtimeconfig.MatcherType]struct{}, len(matchers))
	var previous compiledMatcher
	for index, matcher := range matchers {
		compiled := compiledMatcher{matcherType: matcher.Type(), value: matcher.Value()}
		if err := validateMatcher(compiled); err != nil {
			return normalizedMatcherSet{}, err
		}
		if _, duplicate := seen[compiled.matcherType]; duplicate {
			return normalizedMatcherSet{}, invalidSnapshot("duplicate Matcher type")
		}
		seen[compiled.matcherType] = struct{}{}
		if index > 0 && compareMatchers(previous, compiled) >= 0 {
			return normalizedMatcherSet{}, invalidSnapshot("Matchers are not in canonical order")
		}
		result.matchers[index] = compiled
		previous = compiled
	}
	return result, nil
}

func validateMatcher(matcher compiledMatcher) error {
	if matcher.matcherType == "" || strings.TrimSpace(string(matcher.matcherType)) != string(matcher.matcherType) {
		return invalidSnapshot("invalid Matcher type")
	}
	if matcher.value == "" || strings.TrimSpace(matcher.value) != matcher.value {
		return invalidSnapshot("invalid Matcher value")
	}

	switch matcher.matcherType {
	case runtimeconfig.MatcherTypeMessageType:
		if matcher.value != "text" && matcher.value != "binary" {
			return invalidSnapshot("invalid message-type value")
		}
	case runtimeconfig.MatcherTypePrincipalKind:
		if matcher.value != "authenticated" && matcher.value != "anonymous" {
			return invalidSnapshot("invalid principal-kind value")
		}
	case runtimeconfig.MatcherTypeAuthenticationType:
		if matcher.value != "jwt" && matcher.value != "api-key" && matcher.value != "basic" {
			return invalidSnapshot("invalid authentication-type value")
		}
	case runtimeconfig.MatcherTypeAuthenticationProvider:
		// Provider spelling and case are preserved after canonical trimming.
	default:
		return invalidSnapshot("unsupported Matcher type")
	}
	return nil
}

func compileMatchers(matchers []runtimeconfig.MatcherSnapshot) []compiledMatcher {
	compiled := make([]compiledMatcher, len(matchers))
	for index, matcher := range matchers {
		compiled[index] = compiledMatcher{matcherType: matcher.Type(), value: matcher.Value()}
	}
	return compiled
}

func compareMatchers(first, second compiledMatcher) int {
	if first.matcherType < second.matcherType {
		return -1
	}
	if first.matcherType > second.matcherType {
		return 1
	}
	if first.value < second.value {
		return -1
	}
	if first.value > second.value {
		return 1
	}
	return 0
}

func resolveHandler(registry map[string]message.Handler, reference string) (message.Handler, bool) {
	if reference != legacyHandlerRef {
		return nil, false
	}
	handler, exists := registry[legacyHandlerRef]
	return handler, exists && !nilHandler(handler)
}

func nilHandler(handler message.Handler) bool {
	if handler == nil {
		return true
	}
	value := reflect.ValueOf(handler)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func validateCompiled(compiled *Router) error {
	if compiled == nil {
		return impossibleCompiledState("Router is nil")
	}
	for _, route := range compiled.routes {
		if route.id == "" || route.handlerRef == "" || nilHandler(route.handler) || len(route.matchers) == 0 {
			return impossibleCompiledState("incomplete compiled Route")
		}
	}
	if compiled.defaultHandler != nil && (compiled.defaultHandler.reference == "" || nilHandler(compiled.defaultHandler.handler)) {
		return impossibleCompiledState("incomplete compiled default Handler")
	}
	return nil
}

func validIdentifier(identifier string) bool {
	if len(identifier) == 0 || len(identifier) > maximumIdentifierBytes || strings.TrimSpace(identifier) != identifier || !asciiLetter(identifier[0]) {
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

func invalidSnapshot(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalidRoutingSnapshot, reason)
}

func unresolvedHandler(reference string) error {
	return fmt.Errorf("%w: %s", ErrUnresolvedHandlerRef, reference)
}

func impossibleCompiledState(reason string) error {
	return fmt.Errorf("%w: %s", ErrImpossibleCompiledState, reason)
}
