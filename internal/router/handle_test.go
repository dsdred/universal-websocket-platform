package router

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

var _ message.Handler = (*Router)(nil)

func TestHandleSelectsFirstCompiledRoute(t *testing.T) {
	first := &countingHandler{}
	second := &countingHandler{}
	compiled := &Router{routes: []compiledRoute{
		testCompiledRoute(10, first, testMatcher(runtimeconfig.MatcherTypeMessageType, "text")),
		testCompiledRoute(20, second, testMatcher(runtimeconfig.MatcherTypeMessageType, "text")),
	}}

	if err := compiled.Handle(context.Background(), authenticatedContext(t, message.TypeText)); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if first.calls.Load() != 1 || second.calls.Load() != 0 {
		t.Fatalf("Handler calls = (%d, %d), want (1, 0)", first.calls.Load(), second.calls.Load())
	}
}

func TestHandleLowestNumericPriorityWins(t *testing.T) {
	snapshot := buildRoutingSnapshot(t, &configurationversion.RoutingSettings{Routes: []configurationversion.Route{
		enabledRoute("later", 20, "text"),
		enabledRouteWithPrincipal("winner", 10, "text", "authenticated"),
	}})
	handler := &countingHandler{}
	compiled, err := New(snapshot, map[string]message.Handler{legacyHandlerRef: handler})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if compiled.routes[0].id != "winner" {
		t.Fatalf("first compiled Route = %q, want winner", compiled.routes[0].id)
	}
	if err := compiled.Handle(context.Background(), authenticatedContext(t, message.TypeText)); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if handler.calls.Load() != 1 {
		t.Fatalf("Handler calls = %d, want 1", handler.calls.Load())
	}
}

func TestHandleApprovedMatchers(t *testing.T) {
	tests := []struct {
		name           string
		matcher        compiledMatcher
		runtimeContext message.Context
	}{
		{name: "message type", matcher: testMatcher(runtimeconfig.MatcherTypeMessageType, "binary"), runtimeContext: authenticatedContext(t, message.TypeBinary)},
		{name: "authenticated principal", matcher: testMatcher(runtimeconfig.MatcherTypePrincipalKind, "authenticated"), runtimeContext: authenticatedContext(t, message.TypeText)},
		{name: "anonymous principal", matcher: testMatcher(runtimeconfig.MatcherTypePrincipalKind, "anonymous"), runtimeContext: anonymousContext(t, message.TypeText)},
		{name: "authentication type", matcher: testMatcher(runtimeconfig.MatcherTypeAuthenticationType, "token"), runtimeContext: authenticatedContext(t, message.TypeText)},
		{name: "authentication provider", matcher: testMatcher(runtimeconfig.MatcherTypeAuthenticationProvider, "provider"), runtimeContext: authenticatedContext(t, message.TypeText)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := &countingHandler{}
			compiled := &Router{routes: []compiledRoute{testCompiledRoute(1, handler, test.matcher)}}
			if err := compiled.Handle(context.Background(), test.runtimeContext); err != nil {
				t.Fatalf("Handle() error = %v", err)
			}
			if handler.calls.Load() != 1 {
				t.Fatalf("Handler calls = %d, want 1", handler.calls.Load())
			}
		})
	}
}

func TestHandleMatchersUseANDSemantics(t *testing.T) {
	selected := &countingHandler{}
	fallback := &countingHandler{}
	compiled := &Router{
		routes: []compiledRoute{testCompiledRoute(1, selected,
			testMatcher(runtimeconfig.MatcherTypeMessageType, "text"),
			testMatcher(runtimeconfig.MatcherTypePrincipalKind, "anonymous"),
		)},
		defaultHandler: &compiledHandler{reference: legacyHandlerRef, handler: fallback},
	}

	if err := compiled.Handle(context.Background(), authenticatedContext(t, message.TypeText)); err != nil {
		t.Fatalf("Handle(mismatch) error = %v", err)
	}
	if selected.calls.Load() != 0 || fallback.calls.Load() != 1 {
		t.Fatalf("mismatch calls = (%d, %d), want (0, 1)", selected.calls.Load(), fallback.calls.Load())
	}
	if err := compiled.Handle(context.Background(), anonymousContext(t, message.TypeText)); err != nil {
		t.Fatalf("Handle(match) error = %v", err)
	}
	if selected.calls.Load() != 1 || fallback.calls.Load() != 1 {
		t.Fatalf("match calls = (%d, %d), want (1, 1)", selected.calls.Load(), fallback.calls.Load())
	}
}

func TestHandleComparisonsAreExactAndCaseSensitive(t *testing.T) {
	tests := []compiledMatcher{
		testMatcher(runtimeconfig.MatcherTypeMessageType, "Text"),
		testMatcher(runtimeconfig.MatcherTypePrincipalKind, "Authenticated"),
		testMatcher(runtimeconfig.MatcherTypeAuthenticationType, "Token"),
		testMatcher(runtimeconfig.MatcherTypeAuthenticationProvider, "Provider"),
	}
	for _, matcher := range tests {
		handler := &countingHandler{}
		compiled := &Router{routes: []compiledRoute{testCompiledRoute(1, handler, matcher)}}
		if err := compiled.Handle(context.Background(), authenticatedContext(t, message.TypeText)); err != nil {
			t.Fatalf("Handle(%s) error = %v", matcher.matcherType, err)
		}
		if handler.calls.Load() != 0 {
			t.Fatalf("case-variant %s matched", matcher.matcherType)
		}
	}
}

func TestHandleAuthenticationMetadataDoesNotMatchAnonymousContext(t *testing.T) {
	for _, matcher := range []compiledMatcher{
		testMatcher(runtimeconfig.MatcherTypeAuthenticationType, "token"),
		testMatcher(runtimeconfig.MatcherTypeAuthenticationProvider, "provider"),
	} {
		handler := &countingHandler{}
		compiled := &Router{routes: []compiledRoute{testCompiledRoute(1, handler, matcher)}}
		if err := compiled.Handle(context.Background(), anonymousContext(t, message.TypeText)); err != nil {
			t.Fatalf("Handle(%s) error = %v", matcher.matcherType, err)
		}
		if handler.calls.Load() != 0 {
			t.Fatalf("anonymous Context matched %s", matcher.matcherType)
		}
	}
}

func TestHandleUsesDefaultAndNoMatchReturnsNil(t *testing.T) {
	defaultHandler := &countingHandler{}
	withDefault := &Router{defaultHandler: &compiledHandler{reference: legacyHandlerRef, handler: defaultHandler}}
	if err := withDefault.Handle(context.Background(), authenticatedContext(t, message.TypeText)); err != nil {
		t.Fatalf("Handle(default) error = %v", err)
	}
	if defaultHandler.calls.Load() != 1 {
		t.Fatalf("Default Handler calls = %d, want 1", defaultHandler.calls.Load())
	}

	withoutDefault := &Router{}
	if err := withoutDefault.Handle(context.Background(), authenticatedContext(t, message.TypeText)); err != nil {
		t.Fatalf("Handle(no match) error = %v, want nil", err)
	}
}

func TestHandleShortCircuitsAndReturnsSelectedErrorUnchanged(t *testing.T) {
	selectedError := errors.New("selected failure")
	selected := &countingHandler{err: selectedError}
	notSelected := &countingHandler{err: errors.New("irrelevant failure")}
	compiled := &Router{routes: []compiledRoute{
		testCompiledRoute(1, selected, testMatcher(runtimeconfig.MatcherTypeMessageType, "text")),
		testCompiledRoute(2, notSelected, testMatcher(runtimeconfig.MatcherTypeMessageType, "text")),
	}}

	err := compiled.Handle(context.Background(), authenticatedContext(t, message.TypeText))
	if err != selectedError {
		t.Fatalf("Handle() error = %v, want unchanged selected error", err)
	}
	if selected.calls.Load() != 1 || notSelected.calls.Load() != 0 {
		t.Fatalf("Handler calls = (%d, %d), want (1, 0)", selected.calls.Load(), notSelected.calls.Load())
	}
}

func TestHandleDisabledRoutesNeverExecute(t *testing.T) {
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
	if err := compiled.Handle(context.Background(), authenticatedContext(t, message.TypeText)); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
}

func TestHandleRepeatedCallsAreDeterministic(t *testing.T) {
	selected := &countingHandler{}
	notSelected := &countingHandler{}
	compiled := &Router{routes: []compiledRoute{
		testCompiledRoute(1, selected, testMatcher(runtimeconfig.MatcherTypeMessageType, "text")),
		testCompiledRoute(2, notSelected, testMatcher(runtimeconfig.MatcherTypeMessageType, "text")),
	}}
	runtimeContext := authenticatedContext(t, message.TypeText)

	const attempts = 100
	for range attempts {
		if err := compiled.Handle(context.Background(), runtimeContext); err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
	}
	if selected.calls.Load() != attempts || notSelected.calls.Load() != 0 {
		t.Fatalf("Handler calls = (%d, %d), want (%d, 0)", selected.calls.Load(), notSelected.calls.Load(), attempts)
	}
}

func TestHandleConcurrentReadsDoNotMutateRouter(t *testing.T) {
	selected := &countingHandler{}
	compiled := &Router{routes: []compiledRoute{testCompiledRoute(1, selected,
		testMatcher(runtimeconfig.MatcherTypeMessageType, "text"),
		testMatcher(runtimeconfig.MatcherTypeAuthenticationProvider, "provider"),
	)}}
	before := cloneCompiledRoutes(compiled.routes)
	runtimeContext := authenticatedContext(t, message.TypeText)
	start := make(chan struct{})
	const callers = 64
	errorsByCall := make(chan error, callers)
	var waitGroup sync.WaitGroup
	waitGroup.Add(callers)
	for range callers {
		go func() {
			defer waitGroup.Done()
			<-start
			errorsByCall <- compiled.Handle(context.Background(), runtimeContext)
		}()
	}
	close(start)
	waitGroup.Wait()
	close(errorsByCall)
	for err := range errorsByCall {
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
	}
	if selected.calls.Load() != callers {
		t.Fatalf("Handler calls = %d, want %d", selected.calls.Load(), callers)
	}
	if !reflect.DeepEqual(compiled.routes, before) {
		t.Fatal("Handle mutated compiled Router state")
	}
}

func TestHandleDoesNotMutateContext(t *testing.T) {
	runtimeContext := authenticatedContext(t, message.TypeBinary)
	beforeMessage := runtimeContext.Message()
	before := contextScalars(runtimeContext)
	compiled := &Router{routes: []compiledRoute{testCompiledRoute(1, &countingHandler{},
		testMatcher(runtimeconfig.MatcherTypeMessageType, "binary"),
	)}}

	if err := compiled.Handle(context.Background(), runtimeContext); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if !reflect.DeepEqual(runtimeContext.Message(), beforeMessage) || contextScalars(runtimeContext) != before {
		t.Fatal("Handle mutated Runtime Message Context")
	}
}

func TestHandleCanceledBeforeSelection(t *testing.T) {
	handler := &countingHandler{}
	compiled := &Router{routes: []compiledRoute{testCompiledRoute(1, handler,
		testMatcher(runtimeconfig.MatcherTypeMessageType, "text"),
	)}}
	callContext, cancel := context.WithCancel(context.Background())
	cancel()

	if err := compiled.Handle(callContext, authenticatedContext(t, message.TypeText)); !errors.Is(err, context.Canceled) {
		t.Fatalf("Handle() error = %v, want context.Canceled", err)
	}
	if handler.calls.Load() != 0 {
		t.Fatalf("Handler calls = %d, want 0", handler.calls.Load())
	}
}

func TestHandleReturnsSelectedHandlerResultWhenCancellationRacesExecution(t *testing.T) {
	handlerError := errors.New("handler result")
	callContext, cancel := context.WithCancel(context.Background())
	selected := handlerFunc(func(context.Context, message.Context) error {
		cancel()
		return handlerError
	})
	compiled := &Router{routes: []compiledRoute{testCompiledRoute(1, selected,
		testMatcher(runtimeconfig.MatcherTypeMessageType, "text"),
	)}}

	if err := compiled.Handle(callContext, authenticatedContext(t, message.TypeText)); err != handlerError {
		t.Fatalf("Handle() error = %v, want unchanged Handler result", err)
	}
	if !errors.Is(callContext.Err(), context.Canceled) {
		t.Fatalf("call context error = %v, want context.Canceled", callContext.Err())
	}
}

func TestHandleRejectsInvalidRuntimeMessageContext(t *testing.T) {
	handler := &countingHandler{}
	compiled := &Router{defaultHandler: &compiledHandler{reference: legacyHandlerRef, handler: handler}}

	if err := compiled.Handle(context.Background(), message.Context{}); !errors.Is(err, message.ErrInvalidContext) {
		t.Fatalf("Handle() error = %v, want message.ErrInvalidContext", err)
	}
	if handler.calls.Load() != 0 {
		t.Fatalf("Handler calls = %d, want 0", handler.calls.Load())
	}
}

func TestHandleAllocations(t *testing.T) {
	compiled := &Router{routes: []compiledRoute{testCompiledRoute(1, &handlerStub{},
		testMatcher(runtimeconfig.MatcherTypeMessageType, "text"),
		testMatcher(runtimeconfig.MatcherTypePrincipalKind, "authenticated"),
		testMatcher(runtimeconfig.MatcherTypeAuthenticationType, "token"),
		testMatcher(runtimeconfig.MatcherTypeAuthenticationProvider, "provider"),
	)}}
	runtimeContext := authenticatedContext(t, message.TypeText)
	callContext := context.Background()
	var handleError error
	allocations := testing.AllocsPerRun(1000, func() {
		handleError = compiled.Handle(callContext, runtimeContext)
	})
	if handleError != nil {
		t.Fatalf("Handle() error = %v", handleError)
	}
	if allocations != 0 {
		t.Fatalf("Handle() allocations = %v, want 0", allocations)
	}
}

func testCompiledRoute(priority uint32, handler message.Handler, matchers ...compiledMatcher) compiledRoute {
	return compiledRoute{
		id:         "route",
		priority:   priority,
		handlerRef: legacyHandlerRef,
		matchers:   matchers,
		handler:    handler,
	}
}

func testMatcher(matcherType runtimeconfig.MatcherType, value string) compiledMatcher {
	return compiledMatcher{matcherType: matcherType, value: value}
}

func authenticatedContext(t *testing.T, messageType message.Type) message.Context {
	t.Helper()
	return newRuntimeMessageContext(t, messageType, true, false, "token", "provider")
}

func anonymousContext(t *testing.T, messageType message.Type) message.Context {
	t.Helper()
	return newRuntimeMessageContext(t, messageType, false, true, "", "")
}

func newRuntimeMessageContext(
	t *testing.T,
	messageType message.Type,
	authenticated bool,
	anonymous bool,
	authenticationType string,
	authenticationProvider string,
) message.Context {
	t.Helper()
	runtimeMessage, err := message.New(messageType, []byte("payload"))
	if err != nil {
		t.Fatalf("message.New() error = %v", err)
	}
	runtimeContext, err := message.NewContext(
		&runtimeMessage,
		testSender{},
		"session-id",
		authenticated,
		anonymous,
		authenticationType,
		authenticationProvider,
	)
	if err != nil {
		t.Fatalf("message.NewContext() error = %v", err)
	}
	return runtimeContext
}

func cloneCompiledRoutes(routes []compiledRoute) []compiledRoute {
	cloned := make([]compiledRoute, len(routes))
	copy(cloned, routes)
	for index := range cloned {
		cloned[index].matchers = append([]compiledMatcher(nil), routes[index].matchers...)
	}
	return cloned
}

func contextScalars(runtimeContext message.Context) [6]string {
	return [6]string{
		string(runtimeContext.MessageType()),
		runtimeContext.SessionID(),
		runtimeContext.AuthenticationType(),
		runtimeContext.AuthenticationProvider(),
		boolString(runtimeContext.Authenticated()),
		boolString(runtimeContext.Anonymous()),
	}
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

type countingHandler struct {
	calls atomic.Int64
	err   error
}

func (handler *countingHandler) Handle(context.Context, message.Context) error {
	handler.calls.Add(1)
	return handler.err
}

type handlerFunc func(context.Context, message.Context) error

func (handler handlerFunc) Handle(ctx context.Context, runtimeContext message.Context) error {
	return handler(ctx, runtimeContext)
}

type testSender struct{}

func (testSender) Send(context.Context, message.Message) error { return nil }
