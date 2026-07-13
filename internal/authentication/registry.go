package authentication

import (
	"errors"
	"sync"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

var (
	// ErrNilFactory indicates that a Factory or its Provider type is missing.
	ErrNilFactory = errors.New("authentication factory is nil")
	// ErrFactoryAlreadyRegistered indicates that a Factory type is already registered.
	ErrFactoryAlreadyRegistered = errors.New("authentication factory already registered")
	// ErrFactoryNotFound indicates that no Factory is registered for a Provider type.
	ErrFactoryNotFound = errors.New("authentication factory not found")
)

// Registry manages Authentication Provider Factories by Provider type.
type Registry interface {
	Register(factory Factory) error
	Create(
		provider runtimeconfig.AuthenticationProviderSnapshot,
		resolver secretresolver.Resolver,
	) (Provider, error)
}

// DefaultRegistry is a thread-safe Authentication Provider Factory Registry.
type DefaultRegistry struct {
	mu        sync.RWMutex
	factories map[runtimeconfig.AuthenticationProviderType]Factory
}

// NewRegistry creates an empty Authentication Provider Factory Registry.
func NewRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		factories: make(map[runtimeconfig.AuthenticationProviderType]Factory),
	}
}

// Register adds a Factory for its Provider type.
func (registry *DefaultRegistry) Register(factory Factory) error {
	if factory == nil {
		return ErrNilFactory
	}

	providerType := factory.Type()
	if providerType == "" {
		return ErrNilFactory
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()
	if registry.factories == nil {
		registry.factories = make(map[runtimeconfig.AuthenticationProviderType]Factory)
	}
	if _, exists := registry.factories[providerType]; exists {
		return ErrFactoryAlreadyRegistered
	}
	registry.factories[providerType] = factory
	return nil
}

// Create delegates Provider creation to the Factory registered for provider.Type.
func (registry *DefaultRegistry) Create(
	provider runtimeconfig.AuthenticationProviderSnapshot,
	resolver secretresolver.Resolver,
) (Provider, error) {
	if resolver == nil {
		return nil, ErrNilSecretResolver
	}

	registry.mu.RLock()
	factory, exists := registry.factories[provider.Type]
	registry.mu.RUnlock()
	if !exists {
		return nil, ErrFactoryNotFound
	}

	return factory.Create(provider, resolver)
}
