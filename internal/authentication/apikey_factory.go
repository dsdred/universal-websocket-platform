package authentication

import (
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

// APIKeyFactory creates API Key Providers from runtime Authentication Snapshots.
type APIKeyFactory struct{}

// Type returns the Provider type supported by the Factory.
func (APIKeyFactory) Type() runtimeconfig.AuthenticationProviderType {
	return runtimeconfig.AuthenticationProviderAPIKey
}

// Create converts API Key Snapshot metadata into Provider-local runtime configuration.
func (APIKeyFactory) Create(
	provider runtimeconfig.AuthenticationProviderSnapshot,
	resolver secretresolver.Resolver,
) (Provider, error) {
	if resolver == nil {
		return nil, ErrNilSecretResolver
	}
	if provider.Type != runtimeconfig.AuthenticationProviderAPIKey {
		return nil, ErrUnsupportedProviderType
	}
	if !provider.Enabled || provider.APIKey == nil {
		return nil, ErrInvalidProvider
	}

	created, err := NewAPIKeyProvider(APIKeyProviderConfig{
		Name:      provider.Name,
		Header:    provider.APIKey.Header,
		SecretRef: provider.APIKey.SecretRef,
	}, resolver)
	if err != nil {
		return nil, err
	}
	return created, nil
}
