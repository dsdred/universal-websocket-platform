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

// DefaultBootstrap supplies explicit production dependencies to DefaultHost.
type DefaultBootstrap struct {
	resolver    secretresolver.Resolver
	handler     message.Handler
	reportError func(error)
}

// NewBootstrap creates the Runtime entry composition boundary.
func NewBootstrap(
	resolver secretresolver.Resolver,
	handler message.Handler,
) (*DefaultBootstrap, error) {
	return NewBootstrapWithTerminalErrorReporter(resolver, handler, nil)
}

// NewBootstrapWithTerminalErrorReporter creates Runtime composition with an explicit terminal error consumer.
// The callback is synchronous and must return promptly; subsystem boundaries isolate callback panics.
func NewBootstrapWithTerminalErrorReporter(
	resolver secretresolver.Resolver,
	handler message.Handler,
	reportError func(error),
) (*DefaultBootstrap, error) {
	if resolver == nil {
		return nil, ErrNilSecretResolver
	}
	return &DefaultBootstrap{resolver: resolver, handler: handler, reportError: reportError}, nil
}

// Build creates a Built Host; component acquisition remains deferred until Start.
func (bootstrap *DefaultBootstrap) Build(snapshot runtimeconfig.Snapshot) (Host, error) {
	host, err := newHostWithTerminalErrorReporter(
		snapshot,
		bootstrap.resolver,
		bootstrap.handler,
		nil,
		bootstrap.reportError,
	)
	if err != nil {
		return nil, fmt.Errorf("create Runtime Host: %w", err)
	}
	if err := host.Build(); err != nil {
		return nil, fmt.Errorf("build Runtime Host: %w", err)
	}
	return host, nil
}
