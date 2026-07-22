package runtime

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
	"github.com/dsdred/universal-websocket-platform/internal/handshake"
	"github.com/dsdred/universal-websocket-platform/internal/listener"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/router"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
	"github.com/dsdred/universal-websocket-platform/internal/session"
	"github.com/dsdred/universal-websocket-platform/internal/sessionmanager"
)

const legacyHandlerReference = "legacy"

type messageRouterFactory func(*runtimeconfig.RoutingSnapshot, message.Handler) (message.Handler, error)

type runtimeComposition struct {
	listener       listener.Listener
	sessionManager *sessionmanager.Manager
}

type terminalObserver struct{}

func (terminalObserver) Observe(executionowner.TerminalResult) {}

var (
	_ executionowner.TerminalObserver = terminalObserver{}
	_ session.RuntimeContextProvider  = (*handshakeCapabilities)(nil)
)

type handshakeTimeoutHandler struct {
	next    http.Handler
	timeout time.Duration
}

func (handler handshakeTimeoutHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeoutCause(request.Context(), handler.timeout, handshake.ErrHandshakeTimeout)
	defer cancel()
	handler.next.ServeHTTP(response, request.WithContext(ctx))
}

func composeRuntime(
	snapshot runtimeconfig.Snapshot,
	resolver secretresolver.Resolver,
	handler message.Handler,
	capabilities *handshakeCapabilities,
	reportError func(error),
) (runtimeComposition, error) {
	return composeRuntimeWithRouterFactory(
		snapshot,
		resolver,
		handler,
		capabilities,
		reportError,
		buildMessageRouter,
	)
}

func composeRuntimeWithRouterFactory(
	snapshot runtimeconfig.Snapshot,
	resolver secretresolver.Resolver,
	handler message.Handler,
	capabilities *handshakeCapabilities,
	reportError func(error),
	newMessageRouter messageRouterFactory,
) (runtimeComposition, error) {
	if err := validateExecutableSnapshot(snapshot); err != nil {
		return runtimeComposition{}, fmt.Errorf("validate Runtime Snapshot: %w", err)
	}

	runtimeRouter, err := newMessageRouter(snapshot.Routing, handler)
	if err != nil {
		return runtimeComposition{}, fmt.Errorf("build Runtime Message Router: %w", err)
	}

	registry := authentication.NewRegistry()
	if err := registry.Register(authentication.APIKeyFactory{}); err != nil {
		return runtimeComposition{}, fmt.Errorf("register API Key Factory: %w", err)
	}
	if err := registry.Register(authentication.JWTFactory{}); err != nil {
		return runtimeComposition{}, fmt.Errorf("register JWT Factory: %w", err)
	}

	authenticationBootstrap, err := authentication.NewBootstrap(registry, resolver)
	if err != nil {
		return runtimeComposition{}, fmt.Errorf("create Authentication Bootstrap: %w", err)
	}
	authenticationService, err := authenticationBootstrap.Build(snapshot.Authentication)
	if err != nil {
		return runtimeComposition{}, fmt.Errorf("build Authentication: %w", err)
	}

	sessionManager := sessionmanager.New()
	sessionDispatcher, err := session.NewTransactionalDispatcher(
		sessionManager,
		capabilities,
		runtimeRouter,
		terminalObserver{},
	)
	if err != nil {
		return runtimeComposition{}, fmt.Errorf("create transactional Session Dispatcher: %w", err)
	}
	handshakeHandler, err := handshake.NewHandlerWithTerminalErrorReporter(
		capabilities,
		capabilities,
		authenticationService,
		sessionDispatcher,
		reportError,
	)
	if err != nil {
		return runtimeComposition{}, fmt.Errorf("create Handshake: %w", err)
	}
	timedHandshakeHandler := handshakeTimeoutHandler{
		next:    handshakeHandler,
		timeout: time.Duration(snapshot.Listener.Timeouts.HandshakeSeconds) * time.Second,
	}

	runtimeListener, err := listener.NewBootstrapWithHandshakeAndTerminalErrorReporter(
		timedHandshakeHandler,
		reportError,
	).Build(snapshot.Listener)
	if err != nil {
		return runtimeComposition{}, fmt.Errorf("build Listener: %w", err)
	}
	return runtimeComposition{
		listener:       runtimeListener,
		sessionManager: sessionManager,
	}, nil
}

func buildMessageRouter(
	routing *runtimeconfig.RoutingSnapshot,
	legacyHandler message.Handler,
) (message.Handler, error) {
	if routing == nil {
		return router.NewCompatibility(legacyHandler), nil
	}
	return router.New(routing, map[string]message.Handler{
		legacyHandlerReference: legacyHandler,
	})
}
