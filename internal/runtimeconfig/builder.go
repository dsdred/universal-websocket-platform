package runtimeconfig

import (
	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

const (
	supportedSchemaIdentity = "uwp.configuration"
	supportedSchemaVersion  = uint32(1)
)

// Builder constructs complete immutable Runtime Snapshots.
type Builder struct{}

func NewBuilder() Builder { return Builder{} }

// Build returns exactly one complete Snapshot or one non-empty Diagnostics
// collection. It retains neither input nor output.
func (Builder) Build(input runtimeconfigload.DetachedLoadResult) (Snapshot, []Diagnostic) {
	collector := newDiagnosticCollector()
	version := input.ConfigurationVersion()
	validateHandoff(input, version, collector)

	schemaSupported := input.SchemaIdentity() == supportedSchemaIdentity &&
		input.SchemaVersion() == supportedSchemaVersion
	if schemaSupported {
		validateConfiguration(version, collector)
	}

	diagnostics := collector.diagnostics()
	if len(diagnostics) != 0 {
		return Snapshot{}, diagnostics
	}

	return Snapshot{
		provenance: Provenance{
			WorkspaceID:                input.WorkspaceID(),
			ConfigurationID:            input.ConfigurationID(),
			ConfigurationVersionID:     input.ConfigurationVersionID(),
			ConfigurationVersionNumber: input.ConfigurationVersionNumber(),
			SchemaIdentity:             input.SchemaIdentity(),
			SchemaVersion:              input.SchemaVersion(),
			RuntimeInstanceID:          input.RuntimeInstanceID(),
			LaunchAttemptID:            input.LaunchAttemptID(),
		},
		listener:       buildListener(version.Listener),
		authentication: buildAuthentication(version.Authentication),
		routing:        buildRouting(version.Routing),
	}, nil
}

func buildListener(source configurationversion.ListenerSettings) ListenerSnapshot {
	return ListenerSnapshot{
		Host: stringsTrim(source.Host),
		Port: source.Port,
		TLS: TLSSnapshot{
			Enabled:        source.TLS.Enabled,
			CertificateRef: stringsTrim(source.TLS.CertificateRef),
			PrivateKeyRef:  stringsTrim(source.TLS.PrivateKeyRef),
			MinVersion:     stringsTrim(source.TLS.MinVersion),
		},
		Timeouts: TimeoutSnapshot{
			HandshakeSeconds: source.Timeouts.HandshakeSeconds,
			ReadSeconds:      source.Timeouts.ReadSeconds,
			WriteSeconds:     source.Timeouts.WriteSeconds,
			IdleSeconds:      source.Timeouts.IdleSeconds,
		},
	}
}

func buildAuthentication(source configurationversion.AuthenticationSettings) AuthenticationSnapshot {
	return AuthenticationSnapshot{
		Enabled: source.Enabled,
		Providers: copySlice(source.Providers, func(provider configurationversion.AuthenticationProvider) AuthenticationProviderSnapshot {
			result := AuthenticationProviderSnapshot{
				Name:     stringsTrim(provider.Name),
				Type:     AuthenticationProviderType(provider.Type),
				Enabled:  provider.Enabled,
				Priority: provider.Priority,
			}
			if provider.APIKey != nil {
				result.APIKey = &APIKeySnapshot{
					Header:    stringsTrim(provider.APIKey.Header),
					SecretRef: stringsTrim(provider.APIKey.SecretRef),
				}
			}
			if provider.Basic != nil {
				result.Basic = &BasicSnapshot{
					Realm:     stringsTrim(provider.Basic.Realm),
					SecretRef: stringsTrim(provider.Basic.SecretRef),
				}
			}
			if provider.JWT != nil {
				result.JWT = &JWTSnapshot{
					SigningKeys: copySlice(provider.JWT.SigningKeys, func(key configurationversion.JWTSigningKey) JWTSigningKeySnapshot {
						return JWTSigningKeySnapshot{Name: stringsTrim(key.Name), SecretRef: stringsTrim(key.SecretRef)}
					}),
					AllowedAlgorithms: copySlice(provider.JWT.AllowedAlgorithms, func(algorithm configurationversion.JWTAlgorithm) JWTAlgorithm {
						return JWTAlgorithm(algorithm)
					}),
					AllowedIssuers:   copySlice(provider.JWT.AllowedIssuers, stringsTrim),
					AllowedAudiences: copySlice(provider.JWT.AllowedAudiences, stringsTrim),
					RequiredClaims: copySlice(provider.JWT.RequiredClaims, func(claim configurationversion.JWTRequiredClaim) JWTRequiredClaimSnapshot {
						return JWTRequiredClaimSnapshot{Name: stringsTrim(claim.Name), Value: stringsTrim(claim.Value)}
					}),
					ClockSkewSeconds: provider.JWT.ClockSkewSeconds,
				}
			}
			return result
		}),
	}
}

func buildRouting(source *configurationversion.RoutingSettings) *RoutingSnapshot {
	if source == nil {
		return nil
	}
	defaultPresent := source.DefaultHandlerRef != ""
	result := &RoutingSnapshot{
		routes: copySlice(source.Routes, func(route configurationversion.Route) RouteSnapshot {
			matchers := copySlice(route.Matchers, func(matcher configurationversion.Matcher) MatcherSnapshot {
				return MatcherSnapshot{
					matcherType: MatcherType(stringsTrim(string(matcher.Type))),
					value:       stringsTrim(matcher.Value),
				}
			})
			sortMatchers(matchers)
			return RouteSnapshot{
				id:         stringsTrim(route.ID),
				enabled:    route.Enabled,
				priority:   route.Priority,
				matchers:   matchers,
				handlerRef: stringsTrim(route.HandlerRef),
			}
		}),
		defaultHandlerRef:        stringsTrim(source.DefaultHandlerRef),
		defaultHandlerRefPresent: defaultPresent,
	}
	return result
}
