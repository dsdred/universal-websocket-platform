package authentication

import (
	"errors"
	"fmt"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

var ErrNilRegistry = errors.New("authentication registry is nil")

// Bootstrap assembles an Authentication Service from an immutable runtime Snapshot.
type Bootstrap interface {
	Build(snapshot runtimeconfig.AuthenticationSnapshot) (Service, error)
}

// DefaultBootstrap owns the dependencies required to assemble Authentication.
type DefaultBootstrap struct {
	registry Registry
	resolver secretresolver.Resolver
}

// NewBootstrap creates an Authentication composition root with immutable dependencies.
func NewBootstrap(
	registry Registry,
	resolver secretresolver.Resolver,
) (*DefaultBootstrap, error) {
	if registry == nil {
		return nil, ErrNilRegistry
	}
	if resolver == nil {
		return nil, ErrNilSecretResolver
	}

	return &DefaultBootstrap{
		registry: registry,
		resolver: resolver,
	}, nil
}

// Build creates an ordered Service without performing Authentication.
func (bootstrap *DefaultBootstrap) Build(
	snapshot runtimeconfig.AuthenticationSnapshot,
) (Service, error) {
	if !snapshot.Enabled {
		return NewService(nil)
	}

	providers := make([]Provider, 0, len(snapshot.Providers))
	for _, providerSnapshot := range snapshot.Providers {
		provider, err := bootstrap.registry.Create(providerSnapshot, bootstrap.resolver)
		if err != nil {
			return nil, fmt.Errorf("create Authentication Provider %q: %w", providerSnapshot.Name, err)
		}
		providers = append(providers, provider)
	}

	service, err := NewService(providers)
	if err != nil {
		return nil, fmt.Errorf("create Authentication Service: %w", err)
	}
	return service, nil
}
