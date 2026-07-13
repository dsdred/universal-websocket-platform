package authentication

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

func TestJWTFactoryImplementsFactory(t *testing.T) {
	var _ Factory = JWTFactory{}
}

func TestJWTFactoryCreatesProvider(t *testing.T) {
	resolver := &countingResolver{value: []byte(jwtTestSecret)}
	created, err := (JWTFactory{}).Create(validJWTProviderSnapshot(), resolver)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	provider, ok := created.(*JWTProvider)
	if !ok {
		t.Fatalf("Create() Provider type = %T, want *JWTProvider", created)
	}
	if provider.Name() != "internal-jwt" || provider.Type() != runtimeconfig.AuthenticationProviderJWT {
		t.Fatalf("Provider = (%q, %q), want (internal-jwt, jwt)", provider.Name(), provider.Type())
	}
	if resolver.calls != 0 {
		t.Fatalf("Resolve calls during Create() = %d, want 0", resolver.calls)
	}
}

func TestJWTFactoryRejectsWrongType(t *testing.T) {
	snapshot := validJWTProviderSnapshot()
	snapshot.Type = runtimeconfig.AuthenticationProviderAPIKey

	provider, err := (JWTFactory{}).Create(snapshot, &countingResolver{})
	if provider != nil || !errors.Is(err, ErrUnsupportedProviderType) {
		t.Fatalf("Create() = (%v, %v), want nil and ErrUnsupportedProviderType", provider, err)
	}
}

func TestJWTFactoryRejectsDisabledProvider(t *testing.T) {
	snapshot := validJWTProviderSnapshot()
	snapshot.Enabled = false

	provider, err := (JWTFactory{}).Create(snapshot, &countingResolver{})
	if provider != nil || !errors.Is(err, ErrInvalidProvider) {
		t.Fatalf("Create() = (%v, %v), want nil and ErrInvalidProvider", provider, err)
	}
}

func TestJWTFactoryRejectsNilMetadata(t *testing.T) {
	snapshot := validJWTProviderSnapshot()
	snapshot.JWT = nil

	provider, err := (JWTFactory{}).Create(snapshot, &countingResolver{})
	if provider != nil || !errors.Is(err, ErrInvalidProvider) {
		t.Fatalf("Create() = (%v, %v), want nil and ErrInvalidProvider", provider, err)
	}
}

func TestJWTFactoryRejectsNilResolver(t *testing.T) {
	provider, err := (JWTFactory{}).Create(validJWTProviderSnapshot(), nil)
	if provider != nil || !errors.Is(err, ErrNilSecretResolver) {
		t.Fatalf("Create() = (%v, %v), want nil and ErrNilSecretResolver", provider, err)
	}
}

func TestJWTFactoryDeepCopiesSnapshotCollections(t *testing.T) {
	snapshot := validJWTProviderSnapshot()
	want := JWTProviderConfig{
		Name: "internal-jwt",
		SigningKeys: []JWTSigningKeyConfig{
			{Name: "primary", SecretRef: jwtTestSecretRef},
			{Name: "rotation", SecretRef: "secrets/jwt/rotation"},
		},
		AllowedAlgorithms: []runtimeconfig.JWTAlgorithm{runtimeconfig.HS256, runtimeconfig.HS512},
		AllowedIssuers:    []string{"issuer-a", "issuer-b"},
		AllowedAudiences:  []string{"audience-a", "audience-b"},
		RequiredClaims: []JWTRequiredClaimConfig{
			{Name: "tenant", Value: "internal"},
			{Name: "active", Value: "true"},
		},
		ClockSkewSeconds: 60,
	}
	created, err := (JWTFactory{}).Create(snapshot, &countingResolver{})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	provider := created.(*JWTProvider)

	snapshot.JWT.SigningKeys[0].Name = "changed"
	snapshot.JWT.AllowedAlgorithms[0] = runtimeconfig.HS384
	snapshot.JWT.AllowedIssuers[0] = "changed"
	snapshot.JWT.AllowedAudiences[0] = "changed"
	snapshot.JWT.RequiredClaims[0].Value = "changed"

	if !reflect.DeepEqual(provider.config, want) {
		t.Fatalf("Provider config = %+v, want %+v", provider.config, want)
	}
}

func TestJWTFactoryPropagatesJWTProviderErrors(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*runtimeconfig.AuthenticationProviderSnapshot)
		wantErr error
	}{
		{name: "invalid config", mutate: func(snapshot *runtimeconfig.AuthenticationProviderSnapshot) {
			snapshot.Name = "   "
		}, wantErr: ErrInvalidJWTProviderConfig},
		{name: "unsupported algorithm", mutate: func(snapshot *runtimeconfig.AuthenticationProviderSnapshot) {
			snapshot.JWT.AllowedAlgorithms = []runtimeconfig.JWTAlgorithm{runtimeconfig.RS256}
		}, wantErr: ErrUnsupportedJWTAlgorithm},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := validJWTProviderSnapshot()
			test.mutate(&snapshot)
			provider, err := (JWTFactory{}).Create(snapshot, &countingResolver{})
			if provider != nil || !errors.Is(err, test.wantErr) {
				t.Fatalf("Create() = (%v, %v), want nil and errors.Is(_, %v)", provider, err, test.wantErr)
			}
		})
	}
}

func TestJWTFactorySmokeScenario(t *testing.T) {
	resolver := mustJWTMemoryResolver(t, jwtTestSecret)
	registry := NewRegistry()
	if err := registry.Register(JWTFactory{}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	bootstrap, err := NewBootstrap(registry, resolver)
	if err != nil {
		t.Fatalf("NewBootstrap() error = %v", err)
	}
	snapshot := validJWTProviderSnapshot()
	snapshot.JWT.SigningKeys = snapshot.JWT.SigningKeys[:1]
	snapshot.JWT.AllowedAlgorithms = snapshot.JWT.AllowedAlgorithms[:1]
	snapshot.JWT.AllowedIssuers = nil
	snapshot.JWT.AllowedAudiences = nil
	snapshot.JWT.RequiredClaims = nil
	service, err := bootstrap.Build(runtimeconfig.AuthenticationSnapshot{
		Enabled:   true,
		Providers: []runtimeconfig.AuthenticationProviderSnapshot{snapshot},
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	result, err := service.Authenticate(
		context.Background(),
		bearerRequest("Authorization", validSignedJWT(t, jwtTestSecret)),
	)
	if err != nil || !result.Success || result.Principal == nil {
		t.Fatalf("Authenticate() = (%+v, %v), want Principal", result, err)
	}
	t.Logf("Registry -> JWTFactory -> Bootstrap -> Service -> Principal %q", result.Principal.Name)
}

func validJWTProviderSnapshot() runtimeconfig.AuthenticationProviderSnapshot {
	return runtimeconfig.AuthenticationProviderSnapshot{
		Name:     "internal-jwt",
		Type:     runtimeconfig.AuthenticationProviderJWT,
		Enabled:  true,
		Priority: 10,
		JWT: &runtimeconfig.JWTSnapshot{
			SigningKeys: []runtimeconfig.JWTSigningKeySnapshot{
				{Name: "primary", SecretRef: jwtTestSecretRef},
				{Name: "rotation", SecretRef: "secrets/jwt/rotation"},
			},
			AllowedAlgorithms: []runtimeconfig.JWTAlgorithm{runtimeconfig.HS256, runtimeconfig.HS512},
			AllowedIssuers:    []string{"issuer-a", "issuer-b"},
			AllowedAudiences:  []string{"audience-a", "audience-b"},
			RequiredClaims: []runtimeconfig.JWTRequiredClaimSnapshot{
				{Name: "tenant", Value: "internal"},
				{Name: "active", Value: "true"},
			},
			ClockSkewSeconds: 60,
		},
	}
}
