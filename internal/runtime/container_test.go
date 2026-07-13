package runtime

import (
	"strings"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

func TestNewCreatesContainer(t *testing.T) {
	snapshot := validSnapshot()

	container, err := New(snapshot)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var _ Container = container
	got := container.Snapshot()
	if got.ConfigurationID != snapshot.ConfigurationID || got.VersionID != snapshot.VersionID {
		t.Errorf("Snapshot identifiers = (%d, %d), want (%d, %d)", got.ConfigurationID, got.VersionID, snapshot.ConfigurationID, snapshot.VersionID)
	}
	if got.Listener != snapshot.Listener {
		t.Errorf("Listener = %#v, want %#v", got.Listener, snapshot.Listener)
	}
	if !got.Authentication.Enabled || len(got.Authentication.Providers) != 3 {
		t.Errorf("Authentication = %#v", got.Authentication)
	}
}

func TestNewRejectsZeroVersionID(t *testing.T) {
	snapshot := validSnapshot()
	snapshot.VersionID = 0

	_, err := New(snapshot)
	if err == nil || !strings.Contains(err.Error(), "VersionID") {
		t.Fatalf("New() error = %v, want VersionID error", err)
	}
}

func TestNewRejectsZeroConfigurationID(t *testing.T) {
	snapshot := validSnapshot()
	snapshot.ConfigurationID = 0

	_, err := New(snapshot)
	if err == nil || !strings.Contains(err.Error(), "ConfigurationID") {
		t.Fatalf("New() error = %v, want ConfigurationID error", err)
	}
}

func TestNewRejectsMissingListener(t *testing.T) {
	snapshot := validSnapshot()
	snapshot.Listener = runtimeconfig.ListenerSnapshot{}

	_, err := New(snapshot)
	if err == nil || !strings.Contains(err.Error(), "Listener") {
		t.Fatalf("New() error = %v, want Listener error", err)
	}
}

func TestNewAllowsEmptyAuthentication(t *testing.T) {
	snapshot := validSnapshot()
	snapshot.Authentication = runtimeconfig.AuthenticationSnapshot{}

	container, err := New(snapshot)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if got := container.Snapshot().Authentication; got.Enabled || got.Providers != nil {
		t.Errorf("Authentication = %#v, want empty", got)
	}
}

func TestNewDeepCopiesSnapshot(t *testing.T) {
	snapshot := validSnapshot()

	container, err := New(snapshot)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	snapshot.Listener.Host = "changed-host"
	snapshot.Authentication.Providers[0].Name = "changed-provider"
	snapshot.Authentication.Providers[0].JWT.SigningKeys[0].SecretRef = "changed-key"
	snapshot.Authentication.Providers[0].JWT.AllowedAlgorithms[0] = runtimeconfig.ES512
	snapshot.Authentication.Providers[0].JWT.AllowedIssuers[0] = "changed-issuer"
	snapshot.Authentication.Providers[0].JWT.AllowedAudiences[0] = "changed-audience"
	snapshot.Authentication.Providers[0].JWT.RequiredClaims[0].Value = "changed-claim"
	snapshot.Authentication.Providers[1].APIKey.Header = "Changed-Header"
	snapshot.Authentication.Providers[2].Basic.Realm = "Changed Realm"
	snapshot.Authentication.Providers = append(snapshot.Authentication.Providers, runtimeconfig.AuthenticationProviderSnapshot{Name: "new"})

	got := container.Snapshot()
	assertOriginalSnapshot(t, got)
}

func TestSnapshotReturnsCopy(t *testing.T) {
	container, err := New(validSnapshot())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	first := container.Snapshot()
	first.Listener.Host = "changed-host"
	first.Authentication.Providers[0].Name = "changed-provider"
	first.Authentication.Providers[0].JWT.SigningKeys[0].SecretRef = "changed-key"
	first.Authentication.Providers[0].JWT.AllowedAlgorithms[0] = runtimeconfig.ES512
	first.Authentication.Providers[0].JWT.AllowedIssuers[0] = "changed-issuer"
	first.Authentication.Providers[0].JWT.AllowedAudiences[0] = "changed-audience"
	first.Authentication.Providers[0].JWT.RequiredClaims[0].Value = "changed-claim"
	first.Authentication.Providers[1].APIKey.Header = "Changed-Header"
	first.Authentication.Providers[2].Basic.Realm = "Changed Realm"
	first.Authentication.Providers = append(first.Authentication.Providers, runtimeconfig.AuthenticationProviderSnapshot{Name: "new"})

	second := container.Snapshot()
	assertOriginalSnapshot(t, second)
}

func assertOriginalSnapshot(t *testing.T, snapshot runtimeconfig.Snapshot) {
	t.Helper()

	if snapshot.Listener.Host != "127.0.0.1" {
		t.Errorf("Listener Host = %q, want 127.0.0.1", snapshot.Listener.Host)
	}
	if len(snapshot.Authentication.Providers) != 3 {
		t.Fatalf("Providers length = %d, want 3", len(snapshot.Authentication.Providers))
	}
	jwt := snapshot.Authentication.Providers[0]
	if jwt.Name != "jwt" ||
		jwt.JWT.SigningKeys[0].SecretRef != "secrets/jwt/main" ||
		jwt.JWT.AllowedAlgorithms[0] != runtimeconfig.HS256 ||
		jwt.JWT.AllowedIssuers[0] != "issuer" ||
		jwt.JWT.AllowedAudiences[0] != "audience" ||
		jwt.JWT.RequiredClaims[0].Value != "internal" {
		t.Errorf("JWT Provider changed: %#v", jwt)
	}
	if snapshot.Authentication.Providers[1].APIKey.Header != "X-API-Key" {
		t.Errorf("API Key changed: %#v", snapshot.Authentication.Providers[1].APIKey)
	}
	if snapshot.Authentication.Providers[2].Basic.Realm != "Universal WebSocket Platform" {
		t.Errorf("Basic changed: %#v", snapshot.Authentication.Providers[2].Basic)
	}
}

func validSnapshot() runtimeconfig.Snapshot {
	return runtimeconfig.Snapshot{
		ConfigurationID: 11,
		VersionID:       17,
		Listener: runtimeconfig.ListenerSnapshot{
			Host: "127.0.0.1",
			Port: 8080,
			TLS: runtimeconfig.TLSSnapshot{
				Enabled:        true,
				CertificateRef: "secrets/tls/certificate",
				PrivateKeyRef:  "secrets/tls/private-key",
				MinVersion:     "1.3",
			},
			Timeouts: runtimeconfig.TimeoutSnapshot{
				HandshakeSeconds: 10,
				ReadSeconds:      30,
				WriteSeconds:     10,
				IdleSeconds:      60,
			},
		},
		Authentication: runtimeconfig.AuthenticationSnapshot{
			Enabled: true,
			Providers: []runtimeconfig.AuthenticationProviderSnapshot{
				{
					Name:     "jwt",
					Type:     runtimeconfig.AuthenticationProviderJWT,
					Enabled:  true,
					Priority: 10,
					JWT: &runtimeconfig.JWTSnapshot{
						SigningKeys:       []runtimeconfig.JWTSigningKeySnapshot{{Name: "main", SecretRef: "secrets/jwt/main"}},
						AllowedAlgorithms: []runtimeconfig.JWTAlgorithm{runtimeconfig.HS256},
						AllowedIssuers:    []string{"issuer"},
						AllowedAudiences:  []string{"audience"},
						RequiredClaims:    []runtimeconfig.JWTRequiredClaimSnapshot{{Name: "tenant", Value: "internal"}},
						ClockSkewSeconds:  60,
					},
				},
				{
					Name:     "api-key",
					Type:     runtimeconfig.AuthenticationProviderAPIKey,
					Enabled:  true,
					Priority: 20,
					APIKey:   &runtimeconfig.APIKeySnapshot{Header: "X-API-Key", SecretRef: "secrets/api-key/main"},
				},
				{
					Name:     "basic",
					Type:     runtimeconfig.AuthenticationProviderBasic,
					Enabled:  false,
					Priority: 30,
					Basic: &runtimeconfig.BasicSnapshot{
						Realm:     "Universal WebSocket Platform",
						SecretRef: "secrets/basic/main",
					},
				},
			},
		},
	}
}
