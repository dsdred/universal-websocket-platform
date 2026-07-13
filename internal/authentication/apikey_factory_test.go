package authentication

import (
	"errors"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

func TestAPIKeyFactoryImplementsFactory(t *testing.T) {
	var _ Factory = APIKeyFactory{}
}

func TestAPIKeyFactoryCreatesProvider(t *testing.T) {
	resolver := &countingResolver{value: []byte("secret")}
	provider, err := (APIKeyFactory{}).Create(validAPIKeyProviderSnapshot(), resolver)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if provider == nil {
		t.Fatal("Create() returned nil Provider")
	}
	if provider.Name() != "internal-api-key" || provider.Type() != runtimeconfig.AuthenticationProviderAPIKey {
		t.Fatalf("Provider = (%q, %q), want (internal-api-key, api-key)", provider.Name(), provider.Type())
	}
	if resolver.calls != 0 {
		t.Fatalf("Resolve calls during Create() = %d, want 0", resolver.calls)
	}
}

func TestAPIKeyFactoryConvertsSnapshotToProviderConfig(t *testing.T) {
	snapshot := validAPIKeyProviderSnapshot()
	snapshot.Name = "  configured-api-key  "
	snapshot.APIKey.Header = "  X-Internal-Key  "
	snapshot.APIKey.SecretRef = "  secrets/api-key/configured  "

	created, err := (APIKeyFactory{}).Create(snapshot, &countingResolver{})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	provider, ok := created.(*APIKeyProvider)
	if !ok {
		t.Fatalf("Create() Provider type = %T, want *APIKeyProvider", created)
	}
	if provider.name != "configured-api-key" ||
		provider.header != "X-Internal-Key" ||
		provider.secretRef != "secrets/api-key/configured" {
		t.Fatalf("Provider config = (%q, %q, %q)", provider.name, provider.header, provider.secretRef)
	}
}

func TestAPIKeyFactoryRejectsWrongType(t *testing.T) {
	snapshot := validAPIKeyProviderSnapshot()
	snapshot.Type = runtimeconfig.AuthenticationProviderJWT

	provider, err := (APIKeyFactory{}).Create(snapshot, &countingResolver{})
	if provider != nil || !errors.Is(err, ErrUnsupportedProviderType) {
		t.Fatalf("Create() = (%v, %v), want nil and ErrUnsupportedProviderType", provider, err)
	}
}

func TestAPIKeyFactoryRejectsDisabledProvider(t *testing.T) {
	snapshot := validAPIKeyProviderSnapshot()
	snapshot.Enabled = false

	provider, err := (APIKeyFactory{}).Create(snapshot, &countingResolver{})
	if provider != nil || !errors.Is(err, ErrInvalidProvider) {
		t.Fatalf("Create() = (%v, %v), want nil and ErrInvalidProvider", provider, err)
	}
}

func TestAPIKeyFactoryRejectsNilMetadata(t *testing.T) {
	snapshot := validAPIKeyProviderSnapshot()
	snapshot.APIKey = nil

	provider, err := (APIKeyFactory{}).Create(snapshot, &countingResolver{})
	if provider != nil || !errors.Is(err, ErrInvalidProvider) {
		t.Fatalf("Create() = (%v, %v), want nil and ErrInvalidProvider", provider, err)
	}
}

func TestAPIKeyFactoryRejectsNilResolver(t *testing.T) {
	provider, err := (APIKeyFactory{}).Create(validAPIKeyProviderSnapshot(), nil)
	if provider != nil || !errors.Is(err, ErrNilSecretResolver) {
		t.Fatalf("Create() = (%v, %v), want nil and ErrNilSecretResolver", provider, err)
	}
}

func TestAPIKeyFactoryPropagatesProviderConstructorErrors(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*runtimeconfig.AuthenticationProviderSnapshot)
	}{
		{name: "empty Name", mutate: func(snapshot *runtimeconfig.AuthenticationProviderSnapshot) {
			snapshot.Name = "   "
		}},
		{name: "empty Header", mutate: func(snapshot *runtimeconfig.AuthenticationProviderSnapshot) {
			snapshot.APIKey.Header = "   "
		}},
		{name: "empty SecretRef", mutate: func(snapshot *runtimeconfig.AuthenticationProviderSnapshot) {
			snapshot.APIKey.SecretRef = "   "
		}},
		{name: "invalid SecretRef", mutate: func(snapshot *runtimeconfig.AuthenticationProviderSnapshot) {
			snapshot.APIKey.SecretRef = "https://example.com/secret"
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := validAPIKeyProviderSnapshot()
			test.mutate(&snapshot)
			provider, err := (APIKeyFactory{}).Create(snapshot, &countingResolver{})
			if provider != nil || !errors.Is(err, ErrInvalidProvider) {
				t.Fatalf("Create() = (%v, %v), want nil and ErrInvalidProvider", provider, err)
			}
			if test.name == "invalid SecretRef" && !errors.Is(err, secretresolver.ErrInvalidSecretReference) {
				t.Fatalf("Create() error = %v, want ErrInvalidSecretReference", err)
			}
		})
	}
}

func validAPIKeyProviderSnapshot() runtimeconfig.AuthenticationProviderSnapshot {
	return runtimeconfig.AuthenticationProviderSnapshot{
		Name:     "internal-api-key",
		Type:     runtimeconfig.AuthenticationProviderAPIKey,
		Enabled:  true,
		Priority: 10,
		APIKey: &runtimeconfig.APIKeySnapshot{
			Header:    "X-API-Key",
			SecretRef: "secrets/api-key/internal",
		},
	}
}
