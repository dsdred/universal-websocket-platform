// Package configurationloadsource adapts the in-memory configuration
// repositories to the exact Configuration Loader source boundary.
package configurationloadsource

import (
	"errors"

	"github.com/dsdred/universal-websocket-platform/internal/configuration"
	"github.com/dsdred/universal-websocket-platform/internal/configurationloader"
	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
)

type configurationGetter func(uint64) (configuration.Configuration, error)
type configurationVersionGetter func(uint64) (configurationversion.ConfigurationVersion, error)

// MemorySource loads one exact Published ConfigurationVersion and its parent
// from a borrowed pair of in-memory repositories.
type MemorySource struct {
	getConfiguration        configurationGetter
	getConfigurationVersion configurationVersionGetter
}

var _ configurationloader.Source = (*MemorySource)(nil)

// NewMemorySource constructs a source over borrowed in-memory repositories.
func NewMemorySource(
	configurations *configuration.MemoryConfigurationRepository,
	versions *configurationversion.MemoryConfigurationVersionRepository,
) *MemorySource {
	source := &MemorySource{}
	if configurations != nil {
		source.getConfiguration = configurations.Get
	}
	if versions != nil {
		source.getConfigurationVersion = versions.Get
	}
	return source
}

// LoadExact loads and detaches the exact Published ConfigurationVersion and
// verifies its requested Configuration and Workspace ownership chain.
func (s *MemorySource) LoadExact(
	workspaceID uint64,
	configurationID uint64,
	configurationVersionID uint64,
) (configurationloader.SourceObservation, error) {
	if s == nil || s.getConfiguration == nil || s.getConfigurationVersion == nil {
		return configurationloader.SourceObservation{}, configurationloader.ErrSourceUnavailable
	}

	version, err := s.getConfigurationVersion(configurationVersionID)
	if err != nil {
		return configurationloader.SourceObservation{}, normalizeVersionError(err)
	}
	if version.ID != configurationVersionID || version.ConfigurationID != configurationID {
		return configurationloader.SourceObservation{}, configurationloader.ErrIdentityMismatch
	}
	switch version.State {
	case configurationversion.Draft, configurationversion.Validated, configurationversion.Archived:
		return configurationloader.SourceObservation{}, configurationloader.ErrVersionNotPublished
	case configurationversion.Published:
	default:
		return configurationloader.SourceObservation{}, configurationloader.ErrSourceIntegrity
	}
	if version.Number == 0 {
		return configurationloader.SourceObservation{}, configurationloader.ErrSourceIntegrity
	}
	version = cloneConfigurationVersion(version)

	parent, err := s.getConfiguration(configurationID)
	if err != nil {
		return configurationloader.SourceObservation{}, normalizeConfigurationError(err)
	}
	if parent.ID != configurationID || parent.WorkspaceID != workspaceID {
		return configurationloader.SourceObservation{}, configurationloader.ErrIdentityMismatch
	}

	return configurationloader.SourceObservation{
		WorkspaceID:            workspaceID,
		Configuration:          parent,
		ConfigurationVersion:   version,
		SchemaIdentity:         "uwp.configuration",
		SchemaVersion:          1,
		RepresentationComplete: true,
	}, nil
}

func normalizeVersionError(err error) error {
	if errors.Is(err, configurationversion.ErrConfigurationVersionNotFound) {
		return configurationloader.ErrSourceNotFound
	}
	return configurationloader.ErrSourceUnavailable
}

func normalizeConfigurationError(err error) error {
	if errors.Is(err, configuration.ErrConfigurationNotFound) {
		return configurationloader.ErrSourceNotFound
	}
	return configurationloader.ErrSourceUnavailable
}

func cloneConfigurationVersion(
	version configurationversion.ConfigurationVersion,
) configurationversion.ConfigurationVersion {
	version.Authentication = cloneAuthentication(version.Authentication)
	version.Routing = cloneRouting(version.Routing)
	return version
}

func cloneAuthentication(
	authentication configurationversion.AuthenticationSettings,
) configurationversion.AuthenticationSettings {
	if authentication.Providers == nil {
		return authentication
	}
	authentication.Providers = cloneSlice(authentication.Providers)
	for index := range authentication.Providers {
		provider := &authentication.Providers[index]
		if provider.APIKey != nil {
			value := *provider.APIKey
			provider.APIKey = &value
		}
		if provider.Basic != nil {
			value := *provider.Basic
			provider.Basic = &value
		}
		if provider.JWT != nil {
			value := *provider.JWT
			value.SigningKeys = cloneSlice(value.SigningKeys)
			value.AllowedAlgorithms = cloneSlice(value.AllowedAlgorithms)
			value.AllowedIssuers = cloneSlice(value.AllowedIssuers)
			value.AllowedAudiences = cloneSlice(value.AllowedAudiences)
			value.RequiredClaims = cloneSlice(value.RequiredClaims)
			provider.JWT = &value
		}
	}
	return authentication
}

func cloneRouting(
	routing *configurationversion.RoutingSettings,
) *configurationversion.RoutingSettings {
	if routing == nil {
		return nil
	}
	cloned := *routing
	if routing.Routes == nil {
		return &cloned
	}
	cloned.Routes = cloneSlice(routing.Routes)
	for index := range cloned.Routes {
		cloned.Routes[index].Matchers = cloneSlice(cloned.Routes[index].Matchers)
	}
	return &cloned
}

func cloneSlice[T any](values []T) []T {
	if values == nil {
		return nil
	}
	cloned := make([]T, len(values))
	copy(cloned, values)
	return cloned
}
