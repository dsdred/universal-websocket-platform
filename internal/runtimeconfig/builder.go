package runtimeconfig

import (
	"fmt"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
)

// Builder creates immutable runtime Configuration Snapshots.
type Builder struct{}

// NewBuilder creates a Configuration Snapshot Builder.
func NewBuilder() Builder {
	return Builder{}
}

// Build copies a Published Configuration Version into a runtime Snapshot.
func (Builder) Build(version configurationversion.ConfigurationVersion) (Snapshot, error) {
	if version.State != configurationversion.Published {
		return Snapshot{}, fmt.Errorf("build runtime configuration snapshot: version must be Published")
	}

	routing, err := buildRouting(version.Routing)
	if err != nil {
		return Snapshot{}, fmt.Errorf("build runtime configuration snapshot: routing: %w", err)
	}

	return Snapshot{
		ConfigurationID: version.ConfigurationID,
		VersionID:       version.ID,
		Listener:        buildListener(version.Listener),
		Authentication:  buildAuthentication(version.Authentication),
		Routing:         routing,
	}, nil
}

func buildRouting(routing *configurationversion.RoutingSettings) (*RoutingSnapshot, error) {
	normalized, err := (configurationversion.DefaultRoutingValidator{}).Validate(routing)
	if err != nil {
		return nil, err
	}
	if normalized == nil {
		return nil, nil
	}

	return &RoutingSnapshot{
		routes: copySlice(normalized.Routes, func(route configurationversion.Route) RouteSnapshot {
			return RouteSnapshot{
				id:       route.ID,
				enabled:  route.Enabled,
				priority: route.Priority,
				matchers: copySlice(route.Matchers, func(matcher configurationversion.Matcher) MatcherSnapshot {
					return MatcherSnapshot{
						matcherType: MatcherType(matcher.Type),
						value:       matcher.Value,
					}
				}),
				handlerRef: route.HandlerRef,
			}
		}),
		defaultHandlerRef: normalized.DefaultHandlerRef,
	}, nil
}

func cloneRouteSnapshot(route RouteSnapshot) RouteSnapshot {
	route.matchers = cloneSlice(route.matchers)
	return route
}

func buildListener(listener configurationversion.ListenerSettings) ListenerSnapshot {
	return ListenerSnapshot{
		Host: listener.Host,
		Port: listener.Port,
		TLS: TLSSnapshot{
			Enabled:        listener.TLS.Enabled,
			CertificateRef: listener.TLS.CertificateRef,
			PrivateKeyRef:  listener.TLS.PrivateKeyRef,
			MinVersion:     listener.TLS.MinVersion,
		},
		Timeouts: TimeoutSnapshot{
			HandshakeSeconds: listener.Timeouts.HandshakeSeconds,
			ReadSeconds:      listener.Timeouts.ReadSeconds,
			WriteSeconds:     listener.Timeouts.WriteSeconds,
			IdleSeconds:      listener.Timeouts.IdleSeconds,
		},
	}
}

func buildAuthentication(authentication configurationversion.AuthenticationSettings) AuthenticationSnapshot {
	providers := copySlice(authentication.Providers, buildAuthenticationProvider)

	return AuthenticationSnapshot{
		Enabled:   authentication.Enabled,
		Providers: providers,
	}
}

func buildAuthenticationProvider(provider configurationversion.AuthenticationProvider) AuthenticationProviderSnapshot {
	return AuthenticationProviderSnapshot{
		Name:     provider.Name,
		Type:     AuthenticationProviderType(provider.Type),
		Enabled:  provider.Enabled,
		Priority: provider.Priority,
		APIKey:   buildAPIKey(provider.APIKey),
		JWT:      buildJWT(provider.JWT),
		Basic:    buildBasic(provider.Basic),
	}
}

func buildAPIKey(settings *configurationversion.APIKeySettings) *APIKeySnapshot {
	if settings == nil {
		return nil
	}

	return &APIKeySnapshot{
		Header:    settings.Header,
		SecretRef: settings.SecretRef,
	}
}

func buildBasic(settings *configurationversion.BasicSettings) *BasicSnapshot {
	if settings == nil {
		return nil
	}

	return &BasicSnapshot{
		Realm:     settings.Realm,
		SecretRef: settings.SecretRef,
	}
}

func buildJWT(settings *configurationversion.JWTSettings) *JWTSnapshot {
	if settings == nil {
		return nil
	}

	return &JWTSnapshot{
		SigningKeys: copySlice(settings.SigningKeys, func(key configurationversion.JWTSigningKey) JWTSigningKeySnapshot {
			return JWTSigningKeySnapshot{Name: key.Name, SecretRef: key.SecretRef}
		}),
		AllowedAlgorithms: copySlice(settings.AllowedAlgorithms, func(algorithm configurationversion.JWTAlgorithm) JWTAlgorithm {
			return JWTAlgorithm(algorithm)
		}),
		AllowedIssuers:   cloneSlice(settings.AllowedIssuers),
		AllowedAudiences: cloneSlice(settings.AllowedAudiences),
		RequiredClaims: copySlice(settings.RequiredClaims, func(claim configurationversion.JWTRequiredClaim) JWTRequiredClaimSnapshot {
			return JWTRequiredClaimSnapshot{Name: claim.Name, Value: claim.Value}
		}),
		ClockSkewSeconds: settings.ClockSkewSeconds,
	}
}

func cloneSlice[T any](source []T) []T {
	if source == nil {
		return nil
	}

	result := make([]T, len(source))
	copy(result, source)
	return result
}

func copySlice[S, T any](source []S, convert func(S) T) []T {
	if source == nil {
		return nil
	}

	result := make([]T, len(source))
	for index, value := range source {
		result[index] = convert(value)
	}
	return result
}
