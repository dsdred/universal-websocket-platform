package runtime

import (
	"fmt"

	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

// Bootstrap creates a built Runtime Host from explicit production dependencies.
type Bootstrap interface {
	Build(snapshot runtimeconfig.Snapshot) (Host, error)
}

// DefaultBootstrap delegates the complete Runtime composition to DefaultHost.
type DefaultBootstrap struct {
	resolver secretresolver.Resolver
	handler  message.Handler
}

// NewBootstrap creates the Runtime entry composition boundary.
func NewBootstrap(
	resolver secretresolver.Resolver,
	handler message.Handler,
) (*DefaultBootstrap, error) {
	if resolver == nil {
		return nil, ErrNilSecretResolver
	}
	return &DefaultBootstrap{resolver: resolver, handler: handler}, nil
}

// Build creates a Host and delegates component assembly to it.
func (bootstrap *DefaultBootstrap) Build(snapshot runtimeconfig.Snapshot) (Host, error) {
	host, err := NewHost(snapshot, bootstrap.resolver, bootstrap.handler)
	if err != nil {
		return nil, fmt.Errorf("create Runtime Host: %w", err)
	}
	if err := host.Build(); err != nil {
		return nil, fmt.Errorf("build Runtime Host: %w", err)
	}
	return host, nil
}
