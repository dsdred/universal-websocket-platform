package runtime

import (
	"errors"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

// Container exposes dependencies available to Runtime components.
type Container interface {
	Snapshot() runtimeconfig.Snapshot
}

// DefaultContainer stores the immutable Runtime Configuration Snapshot.
type DefaultContainer struct {
	snapshot runtimeconfig.Snapshot
}

// New creates a Runtime dependency Container with its own Snapshot copy.
func New(snapshot runtimeconfig.Snapshot) (*DefaultContainer, error) {
	if snapshot.VersionID == 0 {
		return nil, errors.New("create runtime container: VersionID must not be zero")
	}
	if snapshot.ConfigurationID == 0 {
		return nil, errors.New("create runtime container: ConfigurationID must not be zero")
	}
	if snapshot.Listener.Host == "" || snapshot.Listener.Port == 0 {
		return nil, errors.New("create runtime container: Listener must contain Host and Port")
	}

	return &DefaultContainer{snapshot: cloneSnapshot(snapshot)}, nil
}

// Snapshot returns an independent copy of the Runtime Configuration Snapshot.
func (container *DefaultContainer) Snapshot() runtimeconfig.Snapshot {
	return cloneSnapshot(container.snapshot)
}

func cloneSnapshot(snapshot runtimeconfig.Snapshot) runtimeconfig.Snapshot {
	result := snapshot
	result.Authentication.Providers = cloneProviders(snapshot.Authentication.Providers)
	return result
}

func cloneProviders(providers []runtimeconfig.AuthenticationProviderSnapshot) []runtimeconfig.AuthenticationProviderSnapshot {
	if providers == nil {
		return nil
	}

	result := make([]runtimeconfig.AuthenticationProviderSnapshot, len(providers))
	for index, provider := range providers {
		result[index] = provider
		result[index].APIKey = cloneAPIKey(provider.APIKey)
		result[index].JWT = cloneJWT(provider.JWT)
		result[index].Basic = cloneBasic(provider.Basic)
	}
	return result
}

func cloneAPIKey(settings *runtimeconfig.APIKeySnapshot) *runtimeconfig.APIKeySnapshot {
	if settings == nil {
		return nil
	}

	result := *settings
	return &result
}

func cloneBasic(settings *runtimeconfig.BasicSnapshot) *runtimeconfig.BasicSnapshot {
	if settings == nil {
		return nil
	}

	result := *settings
	return &result
}

func cloneJWT(settings *runtimeconfig.JWTSnapshot) *runtimeconfig.JWTSnapshot {
	if settings == nil {
		return nil
	}

	result := *settings
	result.SigningKeys = cloneSlice(settings.SigningKeys)
	result.AllowedAlgorithms = cloneSlice(settings.AllowedAlgorithms)
	result.AllowedIssuers = cloneSlice(settings.AllowedIssuers)
	result.AllowedAudiences = cloneSlice(settings.AllowedAudiences)
	result.RequiredClaims = cloneSlice(settings.RequiredClaims)
	return &result
}

func cloneSlice[T any](source []T) []T {
	if source == nil {
		return nil
	}

	result := make([]T, len(source))
	copy(result, source)
	return result
}
