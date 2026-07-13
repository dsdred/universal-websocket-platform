package authentication

import (
	"errors"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

var (
	// ErrUnsupportedProviderType indicates that a Factory cannot create the configured Provider type.
	ErrUnsupportedProviderType = errors.New("unsupported authentication provider type")
	// ErrNilSecretResolver indicates that Provider creation did not receive a Secret Resolver.
	ErrNilSecretResolver = errors.New("secret resolver is nil")
	// ErrInvalidProvider indicates that Provider runtime Configuration is incomplete or invalid.
	ErrInvalidProvider = errors.New("invalid authentication provider")
)

// Factory creates an Authentication Provider from runtime Configuration and injected dependencies.
type Factory interface {
	Type() runtimeconfig.AuthenticationProviderType
	Create(
		provider runtimeconfig.AuthenticationProviderSnapshot,
		resolver secretresolver.Resolver,
	) (Provider, error)
}
