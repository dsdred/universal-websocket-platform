package runtime

import (
	"fmt"

	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/connection"
	"github.com/dsdred/universal-websocket-platform/internal/listener"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
	"github.com/dsdred/universal-websocket-platform/internal/session"
)

func composeRuntime(
	snapshot runtimeconfig.Snapshot,
	resolver secretresolver.Resolver,
	handler message.Handler,
) (listener.Listener, error) {
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
	authenticationDispatcher, err := connection.NewAuthenticationDispatcher(
		authenticationService,
		sessionDispatcher,
	)
	if err != nil {
		return nil, fmt.Errorf("create Authentication Dispatcher: %w", err)
	}

	runtimeListener, err := listener.NewBootstrap(authenticationDispatcher).Build(snapshot.Listener)
	if err != nil {
		return nil, fmt.Errorf("build Listener: %w", err)
	}
	return runtimeListener, nil
}
