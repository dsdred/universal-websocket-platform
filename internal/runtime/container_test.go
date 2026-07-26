package runtime

import (
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

func TestContainerStoresAndReturnsImmutableSnapshot(t *testing.T) {
	snapshot := validSnapshot()
	container, err := New(snapshot)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	got := container.Snapshot()
	if got.Provenance() != snapshot.Provenance() || got.Listener() != snapshot.Listener() {
		t.Fatalf("Snapshot() = %#v, want provenance/listener %#v/%#v", got, snapshot.Provenance(), snapshot.Listener())
	}
	authentication := got.Authentication()
	authentication.Providers[0].JWT.AllowedIssuers[0] = "changed"
	if container.Snapshot().Authentication().Providers[0].JWT.AllowedIssuers[0] != "issuer" {
		t.Fatal("Container Snapshot aliases reader memory")
	}
}

func TestNewRejectsZeroSnapshot(t *testing.T) {
	if _, err := New(runtimeconfig.Snapshot{}); err == nil {
		t.Fatal("New(zero Snapshot) error = nil")
	}
}

func validSnapshot() runtimeconfig.Snapshot {
	version := validSnapshotVersion()
	request := runtimeconfigload.NewLoadRequest(1, version.ConfigurationID, version.ID, "runtime", "attempt")
	input := runtimeconfigload.NewDetachedLoadResult(request, version, version.Number, true, "uwp.configuration", 1)
	snapshot, diagnostics := runtimeconfig.NewBuilder().Build(input)
	if len(diagnostics) != 0 {
		panic(diagnostics)
	}
	return snapshot
}

func buildSnapshotForTest(mutate func(*configurationversion.ConfigurationVersion)) (runtimeconfig.Snapshot, []runtimeconfig.Diagnostic) {
	version := validSnapshotVersion()
	if mutate != nil {
		mutate(&version)
	}
	request := runtimeconfigload.NewLoadRequest(1, version.ConfigurationID, version.ID, "runtime", "attempt")
	input := runtimeconfigload.NewDetachedLoadResult(request, version, version.Number, true, "uwp.configuration", 1)
	return runtimeconfig.NewBuilder().Build(input)
}

func validSnapshotVersion() configurationversion.ConfigurationVersion {
	return configurationversion.ConfigurationVersion{
		ID: 2, ConfigurationID: 1, Number: 1, State: configurationversion.Published,
		Listener: configurationversion.ListenerSettings{
			Host: "127.0.0.1", Port: 8080,
			TLS:      configurationversion.TLSSettings{MinVersion: "1.2"},
			Timeouts: configurationversion.TimeoutSettings{HandshakeSeconds: 10, ReadSeconds: 30, WriteSeconds: 10, IdleSeconds: 60},
		},
		Authentication: configurationversion.AuthenticationSettings{
			Enabled: false,
			Providers: []configurationversion.AuthenticationProvider{{
				Name: "jwt", Type: configurationversion.AuthenticationProviderJWT, Enabled: true, Priority: 1,
				JWT: &configurationversion.JWTSettings{
					SigningKeys:       []configurationversion.JWTSigningKey{{Name: "primary", SecretRef: "secrets/jwt/key"}},
					AllowedAlgorithms: []configurationversion.JWTAlgorithm{configurationversion.RS256},
					AllowedIssuers:    []string{"issuer"}, AllowedAudiences: []string{"audience"},
				},
			}},
		},
	}
}
