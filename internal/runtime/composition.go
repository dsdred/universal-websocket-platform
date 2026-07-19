package runtime

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/handshake"
	"github.com/dsdred/universal-websocket-platform/internal/listener"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
	"github.com/dsdred/universal-websocket-platform/internal/session"
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
) (listener.Listener, error) {
	if err := validateExecutableSnapshot(snapshot); err != nil {
		return nil, fmt.Errorf("validate Runtime Snapshot: %w", err)
	}

	registry := authentication.NewRegistry()
	if err := registry.Register(authentication.APIKeyFactory{}); err != nil {
		return nil, fmt.Errorf("register API Key Factory: %w", err)
	}
	if err := registry.Register(authentication.JWTFactory{}); err != nil {
		return nil, fmt.Errorf("register JWT Factory: %w", err)
	}

	authenticationBootstrap, err := authentication.NewBootstrap(registry, resolver)
	if err != nil {
		return nil, fmt.Errorf("create Authentication Bootstrap: %w", err)
	}
	authenticationService, err := authenticationBootstrap.Build(snapshot.Authentication)
	if err != nil {
		return nil, fmt.Errorf("build Authentication: %w", err)
	}

	sessionDispatcher := session.NewDispatcher(handler)
	handshakeHandler, err := handshake.NewHandlerWithTerminalErrorReporter(
		capabilities,
		capabilities,
		authenticationService,
		sessionDispatcher,
		reportError,
	)
	if err != nil {
		return nil, fmt.Errorf("create Handshake: %w", err)
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
		return nil, fmt.Errorf("build Listener: %w", err)
	}
	return runtimeListener, nil
}
