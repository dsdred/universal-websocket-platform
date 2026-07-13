package authentication

import (
	"slices"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

// JWTFactory creates JWT Providers from runtime Authentication Snapshots.
type JWTFactory struct{}

// Type returns the Provider type supported by the Factory.
func (JWTFactory) Type() runtimeconfig.AuthenticationProviderType {
	return runtimeconfig.AuthenticationProviderJWT
}

// Create converts JWT Snapshot metadata into Provider-local runtime configuration.
func (JWTFactory) Create(
	provider runtimeconfig.AuthenticationProviderSnapshot,
	resolver secretresolver.Resolver,
) (Provider, error) {
	if resolver == nil {
		return nil, ErrNilSecretResolver
	}
	if provider.Type != runtimeconfig.AuthenticationProviderJWT {
		return nil, ErrUnsupportedProviderType
	}
	if !provider.Enabled || provider.JWT == nil {
		return nil, ErrInvalidProvider
	}

	created, err := NewJWTProvider(jwtProviderConfigFromSnapshot(provider), resolver)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func jwtProviderConfigFromSnapshot(provider runtimeconfig.AuthenticationProviderSnapshot) JWTProviderConfig {
	config := JWTProviderConfig{
		Name:              provider.Name,
		AllowedAlgorithms: slices.Clone(provider.JWT.AllowedAlgorithms),
		AllowedIssuers:    slices.Clone(provider.JWT.AllowedIssuers),
		AllowedAudiences:  slices.Clone(provider.JWT.AllowedAudiences),
		ClockSkewSeconds:  provider.JWT.ClockSkewSeconds,
		SigningKeys:       make([]JWTSigningKeyConfig, len(provider.JWT.SigningKeys)),
		RequiredClaims:    make([]JWTRequiredClaimConfig, len(provider.JWT.RequiredClaims)),
	}
	for index, key := range provider.JWT.SigningKeys {
		config.SigningKeys[index] = JWTSigningKeyConfig{
			Name:      key.Name,
			SecretRef: key.SecretRef,
		}
	}
	for index, claim := range provider.JWT.RequiredClaims {
		config.RequiredClaims[index] = JWTRequiredClaimConfig{
			Name:  claim.Name,
			Value: claim.Value,
		}
	}
	return config
}
