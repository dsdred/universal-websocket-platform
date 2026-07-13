package runtimeconfig

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
)

func TestBuilderBuild(t *testing.T) {
	version := fullPublishedVersion()

	snapshot, err := NewBuilder().Build(version)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	want := Snapshot{
		ConfigurationID: 41,
		VersionID:       73,
		Listener: ListenerSnapshot{
			Host: "0.0.0.0",
			Port: 9443,
			TLS: TLSSnapshot{
				Enabled:        true,
				CertificateRef: "secrets/tls/certificate",
				PrivateKeyRef:  "secrets/tls/private-key",
				MinVersion:     "1.3",
			},
			Timeouts: TimeoutSnapshot{
				HandshakeSeconds: 15,
				ReadSeconds:      30,
				WriteSeconds:     20,
				IdleSeconds:      120,
			},
		},
		Authentication: AuthenticationSnapshot{
			Enabled: true,
			Providers: []AuthenticationProviderSnapshot{
				{
					Name:     "internal-jwt",
					Type:     AuthenticationProviderJWT,
					Enabled:  true,
					Priority: 10,
					JWT: &JWTSnapshot{
						SigningKeys:       []JWTSigningKeySnapshot{{Name: "primary", SecretRef: "secrets/jwt/primary"}},
						AllowedAlgorithms: []JWTAlgorithm{HS256, RS256},
						AllowedIssuers:    []string{"issuer-a", "issuer-b"},
						AllowedAudiences:  []string{"audience-a", "audience-b"},
						RequiredClaims:    []JWTRequiredClaimSnapshot{{Name: "tenant", Value: "internal"}},
						ClockSkewSeconds:  60,
					},
				},
				{
					Name:     "partner-key",
					Type:     AuthenticationProviderAPIKey,
					Enabled:  true,
					Priority: 20,
					APIKey:   &APIKeySnapshot{Header: "X-API-Key", SecretRef: "secrets/api-key/partner"},
				},
				{
					Name:     "operations-basic",
					Type:     AuthenticationProviderBasic,
					Enabled:  false,
					Priority: 30,
					Basic:    &BasicSnapshot{Realm: "Operations", SecretRef: "secrets/basic/operations"},
				},
			},
		},
	}

	if !reflect.DeepEqual(snapshot, want) {
		t.Errorf("Build() Snapshot = %#v, want %#v", snapshot, want)
	}
}

func TestBuilderDeepCopiesConfigurationVersion(t *testing.T) {
	version := fullPublishedVersion()

	snapshot, err := NewBuilder().Build(version)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	version.Listener.Host = "127.0.0.1"
	version.Listener.TLS.CertificateRef = "changed-certificate"
	version.Listener.Timeouts.ReadSeconds = 999
	version.Authentication.Providers[0].Name = "changed-provider"
	version.Authentication.Providers[0].JWT.SigningKeys[0].SecretRef = "changed-key"
	version.Authentication.Providers[0].JWT.AllowedAlgorithms[0] = configurationversion.ES512
	version.Authentication.Providers[0].JWT.AllowedIssuers[0] = "changed-issuer"
	version.Authentication.Providers[0].JWT.AllowedAudiences[0] = "changed-audience"
	version.Authentication.Providers[0].JWT.RequiredClaims[0].Value = "changed-claim"
	version.Authentication.Providers[1].APIKey.Header = "Changed-Header"
	version.Authentication.Providers[2].Basic.Realm = "Changed Realm"
	version.Authentication.Providers = append(version.Authentication.Providers, configurationversion.AuthenticationProvider{Name: "new"})

	if snapshot.Listener.Host != "0.0.0.0" || snapshot.Listener.TLS.CertificateRef != "secrets/tls/certificate" || snapshot.Listener.Timeouts.ReadSeconds != 30 {
		t.Errorf("Listener Snapshot changed with source: %#v", snapshot.Listener)
	}
	if len(snapshot.Authentication.Providers) != 3 {
		t.Fatalf("Providers length = %d, want 3", len(snapshot.Authentication.Providers))
	}
	jwt := snapshot.Authentication.Providers[0].JWT
	if snapshot.Authentication.Providers[0].Name != "internal-jwt" ||
		jwt.SigningKeys[0].SecretRef != "secrets/jwt/primary" ||
		jwt.AllowedAlgorithms[0] != HS256 ||
		jwt.AllowedIssuers[0] != "issuer-a" ||
		jwt.AllowedAudiences[0] != "audience-a" ||
		jwt.RequiredClaims[0].Value != "internal" {
		t.Errorf("JWT Snapshot changed with source: %#v", snapshot.Authentication.Providers[0])
	}
	if snapshot.Authentication.Providers[1].APIKey.Header != "X-API-Key" {
		t.Errorf("API Key Snapshot changed with source: %#v", snapshot.Authentication.Providers[1].APIKey)
	}
	if snapshot.Authentication.Providers[2].Basic.Realm != "Operations" {
		t.Errorf("Basic Snapshot changed with source: %#v", snapshot.Authentication.Providers[2].Basic)
	}
}

func TestBuilderRejectsNonPublishedVersion(t *testing.T) {
	version := fullPublishedVersion()
	version.State = configurationversion.Draft

	_, err := NewBuilder().Build(version)
	if err == nil || !strings.Contains(err.Error(), "must be Published") {
		t.Fatalf("Build() error = %v, want Published state error", err)
	}
}

func fullPublishedVersion() configurationversion.ConfigurationVersion {
	return configurationversion.ConfigurationVersion{
		ID:              73,
		ConfigurationID: 41,
		State:           configurationversion.Published,
		Listener: configurationversion.ListenerSettings{
			Host: "0.0.0.0",
			Port: 9443,
			TLS: configurationversion.TLSSettings{
				Enabled:        true,
				CertificateRef: "secrets/tls/certificate",
				PrivateKeyRef:  "secrets/tls/private-key",
				MinVersion:     "1.3",
			},
			Timeouts: configurationversion.TimeoutSettings{
				HandshakeSeconds: 15,
				ReadSeconds:      30,
				WriteSeconds:     20,
				IdleSeconds:      120,
			},
		},
		Authentication: configurationversion.AuthenticationSettings{
			Enabled: true,
			Providers: []configurationversion.AuthenticationProvider{
				{
					Name:     "internal-jwt",
					Type:     configurationversion.AuthenticationProviderJWT,
					Enabled:  true,
					Priority: 10,
					JWT: &configurationversion.JWTSettings{
						SigningKeys:       []configurationversion.JWTSigningKey{{Name: "primary", SecretRef: "secrets/jwt/primary"}},
						AllowedAlgorithms: []configurationversion.JWTAlgorithm{configurationversion.HS256, configurationversion.RS256},
						AllowedIssuers:    []string{"issuer-a", "issuer-b"},
						AllowedAudiences:  []string{"audience-a", "audience-b"},
						RequiredClaims:    []configurationversion.JWTRequiredClaim{{Name: "tenant", Value: "internal"}},
						ClockSkewSeconds:  60,
					},
				},
				{
					Name:     "partner-key",
					Type:     configurationversion.AuthenticationProviderAPIKey,
					Enabled:  true,
					Priority: 20,
					APIKey:   &configurationversion.APIKeySettings{Header: "X-API-Key", SecretRef: "secrets/api-key/partner"},
				},
				{
					Name:     "operations-basic",
					Type:     configurationversion.AuthenticationProviderBasic,
					Enabled:  false,
					Priority: 30,
					Basic:    &configurationversion.BasicSettings{Realm: "Operations", SecretRef: "secrets/basic/operations"},
				},
			},
		},
	}
}
