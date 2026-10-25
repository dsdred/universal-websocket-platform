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
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnostics"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
	"github.com/dsdred/universal-websocket-platform/internal/session"
	"github.com/dsdred/universal-websocket-platform/internal/sessionmanager"
)

const legacyHandlerReference = "legacy"

type messageRouterFactory func(*runtimeconfig.RoutingSnapshot, message.Handler) (message.Handler, error)

type sessionDispatcherFactory func(*sessionmanager.Manager, session.RuntimeContextProvider, message.Handler, executionowner.TerminalObserver, time.Duration, time.Duration, time.Duration) (*session.TransactionalDispatcher, error)

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
	return composeRuntimeForStartup(context.Background(), snapshot, resolver, handler, capabilities, reportError)
}

func composeRuntimeForStartup(ctx context.Context, snapshot runtimeconfig.Snapshot, resolver secretresolver.Resolver,
	handler message.Handler, capabilities *handshakeCapabilities, reportError func(error),
) (runtimeComposition, error) {
	return composeRuntimeForStartupWithDiagnostics(ctx, snapshot, resolver, handler, capabilities, reportError,
		runtimediagnostics.Producer{})
}

func composeRuntimeForStartupWithDiagnostics(ctx context.Context, snapshot runtimeconfig.Snapshot, resolver secretresolver.Resolver,
	handler message.Handler, capabilities *handshakeCapabilities, reportError func(error), producer runtimediagnostics.Producer,
) (runtimeComposition, error) {
	return composeRuntimeWithDiagnostics(ctx, snapshot, resolver, handler, capabilities, reportError, producer,
		buildMessageRouter, session.NewTransactionalDispatcherWithIdleTimeout)
}

func composeRuntimeWithRouterFactory(
	snapshot runtimeconfig.Snapshot,
	resolver secretresolver.Resolver,
	handler message.Handler,
	capabilities *handshakeCapabilities,
	reportError func(error),
	newMessageRouter messageRouterFactory,
) (runtimeComposition, error) {
	return composeRuntimeWithFactories(snapshot, resolver, handler, capabilities, reportError,
		newMessageRouter, session.NewTransactionalDispatcherWithIdleTimeout)
}

func composeRuntimeWithFactories(
	snapshot runtimeconfig.Snapshot,
	resolver secretresolver.Resolver,
	handler message.Handler,
	capabilities *handshakeCapabilities,
	reportError func(error),
	newMessageRouter messageRouterFactory,
	newSessionDispatcher sessionDispatcherFactory,
) (runtimeComposition, error) {
	return composeRuntimeWithFactoriesContext(context.Background(), snapshot, resolver, handler, capabilities,
		reportError, newMessageRouter, newSessionDispatcher)
}

func composeRuntimeWithFactoriesContext(
	ctx context.Context,
	snapshot runtimeconfig.Snapshot,
	resolver secretresolver.Resolver,
	handler message.Handler,
	capabilities *handshakeCapabilities,
	reportError func(error),
	newMessageRouter messageRouterFactory,
	newSessionDispatcher sessionDispatcherFactory,
) (runtimeComposition, error) {
	return composeRuntimeWithDiagnostics(ctx, snapshot, resolver, handler, capabilities, reportError,
		runtimediagnostics.Producer{}, newMessageRouter, newSessionDispatcher)
}

func composeRuntimeWithDiagnostics(
	ctx context.Context,
	snapshot runtimeconfig.Snapshot,
	resolver secretresolver.Resolver,
	handler message.Handler,
	capabilities *handshakeCapabilities,
	reportError func(error),
	producer runtimediagnostics.Producer,
	newMessageRouter messageRouterFactory,
	newSessionDispatcher sessionDispatcherFactory,
) (runtimeComposition, error) {
	if err := validateExecutableSnapshot(snapshot); err != nil {
		return runtimeComposition{}, fmt.Errorf("validate Runtime Snapshot: %w", err)
	}
	listenerSnapshot := snapshot.Listener()
	preparedTLS, err := listener.PrepareTLS(ctx, listenerSnapshot, resolver)
	if err != nil {
		return runtimeComposition{}, fmt.Errorf("%w: prepare TLS: %w", ErrInvalidRuntimeConfiguration, err)
	}

	routing, routingPresent := snapshot.Routing()
	var routingInput *runtimeconfig.RoutingSnapshot
	if routingPresent {
		routingInput = &routing
	}
	runtimeRouter, err := newMessageRouter(routingInput, handler)
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
	authenticationService, err := authenticationBootstrap.Build(snapshot.Authentication())
	if err != nil {
		return runtimeComposition{}, fmt.Errorf("build Authentication: %w", err)
	}

	sessionManager := sessionmanager.New()
	sessionDispatcher, err := newSessionDispatcher(
		sessionManager,
		capabilities,
		runtimeRouter,
		terminalObserver{},
		time.Duration(listenerSnapshot.Timeouts.ReadSeconds)*time.Second,
		time.Duration(listenerSnapshot.Timeouts.WriteSeconds)*time.Second,
		time.Duration(listenerSnapshot.Timeouts.IdleSeconds)*time.Second,
	)
	if err != nil {
		return runtimeComposition{}, fmt.Errorf("create transactional Session Dispatcher: %w", err)
	}
	handshakeHandler, err := handshake.NewHandlerWithDiagnostics(
		capabilities,
		capabilities,
		authenticationService,
		sessionDispatcher,
		reportError,
		producer,
	)
	if err != nil {
		return runtimeComposition{}, fmt.Errorf("create Handshake: %w", err)
	}
	timedHandshakeHandler := handshakeTimeoutHandler{
		next:    handshakeHandler,
		timeout: time.Duration(listenerSnapshot.Timeouts.HandshakeSeconds) * time.Second,
	}

	if listenerSnapshot.TLS.Enabled {
		if err := ctx.Err(); err != nil {
			return runtimeComposition{}, err
		}
	}
	runtimeListener, err := listener.NewBootstrapWithTLSAndHandshake(
		timedHandshakeHandler,
		reportError,
		preparedTLS,
	).WithDiagnosticsProducer(producer).Build(listenerSnapshot)
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
