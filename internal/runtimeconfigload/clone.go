package runtimeconfigload

import "github.com/dsdred/universal-websocket-platform/internal/configurationversion"

func cloneConfigurationVersion(source configurationversion.ConfigurationVersion) configurationversion.ConfigurationVersion {
	result := source
	result.Authentication.Providers = copySlice(source.Authentication.Providers, cloneAuthenticationProvider)
	if source.Routing != nil {
		routing := *source.Routing
		routing.Routes = copySlice(source.Routing.Routes, cloneRoute)
		result.Routing = &routing
	}
	return result
}

func cloneAuthenticationProvider(source configurationversion.AuthenticationProvider) configurationversion.AuthenticationProvider {
	result := source
	if source.APIKey != nil {
		settings := *source.APIKey
		result.APIKey = &settings
	}
	if source.Basic != nil {
		settings := *source.Basic
		result.Basic = &settings
	}
	if source.JWT != nil {
		settings := *source.JWT
		settings.SigningKeys = cloneSlice(source.JWT.SigningKeys)
		settings.AllowedAlgorithms = cloneSlice(source.JWT.AllowedAlgorithms)
		settings.AllowedIssuers = cloneSlice(source.JWT.AllowedIssuers)
		settings.AllowedAudiences = cloneSlice(source.JWT.AllowedAudiences)
		settings.RequiredClaims = cloneSlice(source.JWT.RequiredClaims)
		result.JWT = &settings
	}
	return result
}

func cloneRoute(source configurationversion.Route) configurationversion.Route {
	result := source
	result.Matchers = cloneSlice(source.Matchers)
	return result
}

func cloneSlice[T any](source []T) []T {
	if source == nil {
		return nil
	}
	result := make([]T, len(source))
	copy(result, source)
	return result
}

func copySlice[S, T any](source []S, clone func(S) T) []T {
	if source == nil {
		return nil
	}
	result := make([]T, len(source))
	for index, value := range source {
		result[index] = clone(value)
	}
	return result
}
