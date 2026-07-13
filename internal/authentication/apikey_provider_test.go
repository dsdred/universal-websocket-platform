package authentication

import (
	"context"
	"errors"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

type countingResolver struct {
	calls int
	value []byte
	err   error
}

func (resolver *countingResolver) Resolve(context.Context, string) (secretresolver.Secret, error) {
	resolver.calls++
	if resolver.err != nil {
		return secretresolver.Secret{}, resolver.err
	}
	value := append([]byte(nil), resolver.value...)
	return secretresolver.Secret{Value: value}, nil
}

func TestNewAPIKeyProvider(t *testing.T) {
	resolver := &countingResolver{value: []byte("test-key")}
	cfg := validAPIKeyProviderConfig()
	cfg.Name = "  internal-api-key  "
	cfg.Header = "  X-API-Key  "
	cfg.SecretRef = "  secrets/api-keys/internal  "

	provider, err := NewAPIKeyProvider(cfg, resolver)
	if err != nil {
		t.Fatalf("NewAPIKeyProvider() error = %v", err)
	}
	if provider.Name() != "internal-api-key" || provider.Type() != runtimeconfig.AuthenticationProviderAPIKey {
		t.Errorf("Provider = (%q, %q)", provider.Name(), provider.Type())
	}
	if provider.header != "X-API-Key" || provider.secretRef != "secrets/api-keys/internal" {
		t.Errorf("Provider metadata = (%q, %q)", provider.header, provider.secretRef)
	}
	if resolver.calls != 0 {
		t.Errorf("Resolve calls during construction = %d, want 0", resolver.calls)
	}

	var _ Provider = provider
}

func TestNewAPIKeyProviderRejectsNilResolver(t *testing.T) {
	provider, err := NewAPIKeyProvider(validAPIKeyProviderConfig(), nil)
	if provider != nil || !errors.Is(err, ErrNilSecretResolver) {
		t.Fatalf("NewAPIKeyProvider() = (%v, %v), want nil and ErrNilSecretResolver", provider, err)
	}
}

func TestNewAPIKeyProviderRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name                 string
		mutate               func(*APIKeyProviderConfig)
		wantInvalidReference bool
	}{
		{name: "empty Name", mutate: func(cfg *APIKeyProviderConfig) {
			cfg.Name = "   "
		}},
		{name: "empty Header", mutate: func(cfg *APIKeyProviderConfig) {
			cfg.Header = "   "
		}},
		{name: "empty SecretRef", mutate: func(cfg *APIKeyProviderConfig) {
			cfg.SecretRef = "   "
		}},
		{name: "invalid SecretRef", wantInvalidReference: true, mutate: func(cfg *APIKeyProviderConfig) {
			cfg.SecretRef = "https://example.com/secret"
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := validAPIKeyProviderConfig()
			test.mutate(&cfg)
			provider, err := NewAPIKeyProvider(cfg, &countingResolver{})
			if provider != nil || !errors.Is(err, ErrInvalidProvider) {
				t.Errorf("NewAPIKeyProvider() = (%v, %v), want nil and ErrInvalidProvider", provider, err)
			}
			if test.wantInvalidReference && !errors.Is(err, secretresolver.ErrInvalidSecretReference) {
				t.Errorf("NewAPIKeyProvider() error = %v, want ErrInvalidSecretReference", err)
			}
		})
	}
}

func TestAPIKeyProviderAuthenticateSuccess(t *testing.T) {
	provider := mustAPIKeyProvider(t, map[string][]byte{"secrets/api-keys/internal": []byte("expected-key")})

	result, err := provider.Authenticate(context.Background(), AuthenticationRequest{
		Headers: map[string][]string{"X-API-Key": []string{"expected-key"}},
	})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if !result.Success || result.Principal == nil || !result.Principal.Authenticated {
		t.Fatalf("Authenticate() result = %#v", result)
	}
	if result.ProviderName != "internal-api-key" || result.ProviderType != runtimeconfig.AuthenticationProviderAPIKey {
		t.Errorf("result Provider = (%q, %q)", result.ProviderName, result.ProviderType)
	}
	if result.Principal.Name != "internal-api-key" ||
		result.Principal.AuthenticationProvider != "internal-api-key" ||
		result.Principal.AuthenticationMethod != runtimeconfig.AuthenticationProviderAPIKey ||
		result.Principal.AuthenticationType != runtimeconfig.AuthenticationProviderAPIKey ||
		result.Principal.Claims == nil || len(result.Principal.Claims) != 0 {
		t.Errorf("Principal = %#v", result.Principal)
	}
}

func TestAPIKeyProviderRejectsWrongKeyWithConstantTimeComparison(t *testing.T) {
	provider := mustAPIKeyProvider(t, map[string][]byte{"secrets/api-keys/internal": []byte("expected-key")})
	tests := []struct {
		name string
		key  string
	}{
		{name: "same length", key: "incorrect-ke"},
		{name: "different length", key: "wrong"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := provider.Authenticate(context.Background(), AuthenticationRequest{
				Headers: map[string][]string{"X-API-Key": []string{test.key}},
			})
			if err != nil {
				t.Fatalf("Authenticate() error = %v", err)
			}
			assertRejectedAuthentication(t, result)
		})
	}
}

func TestAPIKeyProviderMissingHeader(t *testing.T) {
	resolver := &countingResolver{value: []byte("expected-key")}
	provider, err := NewAPIKeyProvider(validAPIKeyProviderConfig(), resolver)
	if err != nil {
		t.Fatalf("NewAPIKeyProvider() error = %v", err)
	}

	result, err := provider.Authenticate(context.Background(), AuthenticationRequest{})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	assertRejectedAuthentication(t, result)
	if resolver.calls != 0 {
		t.Errorf("Resolve calls = %d, want 0 when header is missing", resolver.calls)
	}
}

func TestAPIKeyProviderMissingSecret(t *testing.T) {
	provider := mustAPIKeyProvider(t, nil)

	result, err := provider.Authenticate(context.Background(), AuthenticationRequest{
		Headers: map[string][]string{"X-API-Key": []string{"provided-key"}},
	})
	if !errors.Is(err, secretresolver.ErrSecretNotFound) {
		t.Fatalf("Authenticate() error = %v, want ErrSecretNotFound", err)
	}
	if result.Success || result.ProviderName != "internal-api-key" {
		t.Errorf("Authenticate() result = %#v", result)
	}
}

func TestAPIKeyProviderCanceledContext(t *testing.T) {
	resolver := &countingResolver{value: []byte("expected-key")}
	provider, err := NewAPIKeyProvider(validAPIKeyProviderConfig(), resolver)
	if err != nil {
		t.Fatalf("NewAPIKeyProvider() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = provider.Authenticate(ctx, AuthenticationRequest{
		Headers: map[string][]string{"X-API-Key": []string{"expected-key"}},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Authenticate() error = %v, want context.Canceled", err)
	}
	if resolver.calls != 0 {
		t.Errorf("Resolve calls = %d, want 0 for canceled context", resolver.calls)
	}
}

func TestAPIKeyProviderHeaderLookupIsCaseInsensitive(t *testing.T) {
	provider := mustAPIKeyProvider(t, map[string][]byte{"secrets/api-keys/internal": []byte("expected-key")})

	result, err := provider.Authenticate(context.Background(), AuthenticationRequest{
		Headers: map[string][]string{"x-api-key": []string{"expected-key"}},
	})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if !result.Success {
		t.Errorf("Authenticate() result = %#v", result)
	}
}

func TestAPIKeyProviderResolveResultMutationDoesNotAffectAuthentication(t *testing.T) {
	resolver, err := secretresolver.NewMemory(map[string][]byte{"secrets/api-keys/internal": []byte("expected-key")})
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	provider, err := NewAPIKeyProvider(validAPIKeyProviderConfig(), resolver)
	if err != nil {
		t.Fatalf("NewAPIKeyProvider() error = %v", err)
	}

	resolved, err := resolver.Resolve(context.Background(), "secrets/api-keys/internal")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	resolved.Value[0] = 'X'

	result, err := provider.Authenticate(context.Background(), AuthenticationRequest{
		Headers: map[string][]string{"X-API-Key": []string{"expected-key"}},
	})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if !result.Success {
		t.Errorf("Authenticate() result = %#v", result)
	}
}

func TestAPIKeyProviderSmokeScenario(t *testing.T) {
	provider := mustAPIKeyProvider(t, map[string][]byte{"secrets/api-keys/internal": []byte("smoke-key")})

	result, err := provider.Authenticate(context.Background(), AuthenticationRequest{
		Headers: map[string][]string{"X-API-Key": []string{"smoke-key"}},
	})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if !result.Success || result.Principal == nil {
		t.Fatalf("Authenticate() result = %#v", result)
	}
	t.Logf("authenticated=%t provider=%s method=%s", result.Principal.Authenticated, result.Principal.AuthenticationProvider, result.Principal.AuthenticationMethod)
}

func validAPIKeyProviderConfig() APIKeyProviderConfig {
	return APIKeyProviderConfig{
		Name:      "internal-api-key",
		Header:    "X-API-Key",
		SecretRef: "secrets/api-keys/internal",
	}
}

func mustAPIKeyProvider(t *testing.T, initial map[string][]byte) *APIKeyProvider {
	t.Helper()
	resolver, err := secretresolver.NewMemory(initial)
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	provider, err := NewAPIKeyProvider(validAPIKeyProviderConfig(), resolver)
	if err != nil {
		t.Fatalf("NewAPIKeyProvider() error = %v", err)
	}
	return provider
}

func assertRejectedAuthentication(t *testing.T, result AuthenticationResult) {
	t.Helper()
	if result.Success || result.Principal != nil || result.ProviderName != "internal-api-key" || result.ProviderType != runtimeconfig.AuthenticationProviderAPIKey {
		t.Errorf("Authentication result = %#v, want rejected API Key result", result)
	}
}
