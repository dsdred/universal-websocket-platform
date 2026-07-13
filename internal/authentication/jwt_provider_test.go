package authentication

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

const (
	jwtTestSecret    = "runtime-jwt-test-secret"
	jwtTestSecretRef = "secrets/jwt/runtime"
)

func TestNewJWTProviderAcceptsValidHS256Config(t *testing.T) {
	cfg := validJWTProviderConfig()
	cfg.Name = "  internal-jwt  "
	cfg.SigningKeys[0].Name = "  primary  "
	cfg.SigningKeys[0].SecretRef = "  secrets/jwt/runtime  "
	resolver := mustJWTMemoryResolver(t, jwtTestSecret)

	provider, err := NewJWTProvider(cfg, resolver)
	if err != nil {
		t.Fatalf("NewJWTProvider() error = %v", err)
	}
	if provider.Name() != "internal-jwt" || provider.Type() != runtimeconfig.AuthenticationProviderJWT {
		t.Fatalf("Provider = (%q, %q)", provider.Name(), provider.Type())
	}
	if provider.config.SigningKeys[0] != (JWTSigningKeyConfig{Name: "primary", SecretRef: jwtTestSecretRef}) {
		t.Fatalf("SigningKey = %+v", provider.config.SigningKeys[0])
	}
	var _ Provider = provider
}

func TestNewJWTProviderRejectsNilResolver(t *testing.T) {
	provider, err := NewJWTProvider(validJWTProviderConfig(), nil)
	if provider != nil || !errors.Is(err, ErrNilSecretResolver) {
		t.Fatalf("NewJWTProvider() = (%v, %v), want nil and ErrNilSecretResolver", provider, err)
	}
}

func TestNewJWTProviderRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name                 string
		mutate               func(*JWTProviderConfig)
		wantInvalidReference bool
	}{
		{name: "empty name", mutate: func(cfg *JWTProviderConfig) { cfg.Name = "   " }},
		{name: "no signing keys", mutate: func(cfg *JWTProviderConfig) { cfg.SigningKeys = nil }},
		{name: "no algorithms", mutate: func(cfg *JWTProviderConfig) { cfg.AllowedAlgorithms = nil }},
		{name: "empty signing key name", mutate: func(cfg *JWTProviderConfig) { cfg.SigningKeys[0].Name = " " }},
		{name: "invalid SecretRef", wantInvalidReference: true, mutate: func(cfg *JWTProviderConfig) {
			cfg.SigningKeys[0].SecretRef = "https://example.com/secret"
		}},
		{name: "duplicate key name", mutate: func(cfg *JWTProviderConfig) {
			cfg.SigningKeys = append(cfg.SigningKeys, cfg.SigningKeys[0])
		}},
		{name: "duplicate algorithm", mutate: func(cfg *JWTProviderConfig) {
			cfg.AllowedAlgorithms = append(cfg.AllowedAlgorithms, runtimeconfig.HS256)
		}},
		{name: "unknown algorithm", mutate: func(cfg *JWTProviderConfig) {
			cfg.AllowedAlgorithms[0] = runtimeconfig.JWTAlgorithm("UNKNOWN")
		}},
		{name: "empty issuer", mutate: func(cfg *JWTProviderConfig) { cfg.AllowedIssuers = []string{" "} }},
		{name: "duplicate issuer", mutate: func(cfg *JWTProviderConfig) { cfg.AllowedIssuers = []string{"issuer", "issuer"} }},
		{name: "empty audience", mutate: func(cfg *JWTProviderConfig) { cfg.AllowedAudiences = []string{" "} }},
		{name: "duplicate audience", mutate: func(cfg *JWTProviderConfig) { cfg.AllowedAudiences = []string{"aud", "aud"} }},
		{name: "empty required claim", mutate: func(cfg *JWTProviderConfig) {
			cfg.RequiredClaims = []JWTRequiredClaimConfig{{Name: "tenant"}}
		}},
		{name: "duplicate required claim", mutate: func(cfg *JWTProviderConfig) {
			cfg.RequiredClaims = []JWTRequiredClaimConfig{{Name: "tenant", Value: "one"}, {Name: "tenant", Value: "two"}}
		}},
		{name: "invalid clock skew", mutate: func(cfg *JWTProviderConfig) { cfg.ClockSkewSeconds = 301 }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := validJWTProviderConfig()
			test.mutate(&cfg)
			provider, err := NewJWTProvider(cfg, mustJWTMemoryResolver(t, jwtTestSecret))
			if provider != nil || !errors.Is(err, ErrInvalidJWTProviderConfig) {
				t.Fatalf("NewJWTProvider() = (%v, %v), want nil and ErrInvalidJWTProviderConfig", provider, err)
			}
			if test.wantInvalidReference && !errors.Is(err, secretresolver.ErrInvalidSecretReference) {
				t.Fatalf("NewJWTProvider() error = %v, want ErrInvalidSecretReference", err)
			}
		})
	}
}

func TestNewJWTProviderRejectsUnsupportedRS256(t *testing.T) {
	cfg := validJWTProviderConfig()
	cfg.AllowedAlgorithms = []runtimeconfig.JWTAlgorithm{runtimeconfig.RS256}

	provider, err := NewJWTProvider(cfg, mustJWTMemoryResolver(t, jwtTestSecret))
	if provider != nil || !errors.Is(err, ErrUnsupportedJWTAlgorithm) {
		t.Fatalf("NewJWTProvider() = (%v, %v), want nil and ErrUnsupportedJWTAlgorithm", provider, err)
	}
}

func TestNewJWTProviderDeepCopiesConfig(t *testing.T) {
	cfg := validJWTProviderConfig()
	cfg.AllowedIssuers = []string{"issuer"}
	cfg.AllowedAudiences = []string{"audience"}
	cfg.RequiredClaims = []JWTRequiredClaimConfig{{Name: "tenant", Value: "internal"}}
	provider, err := NewJWTProvider(cfg, mustJWTMemoryResolver(t, jwtTestSecret))
	if err != nil {
		t.Fatalf("NewJWTProvider() error = %v", err)
	}

	cfg.SigningKeys[0].Name = "changed"
	cfg.AllowedAlgorithms[0] = runtimeconfig.HS512
	cfg.AllowedIssuers[0] = "changed"
	cfg.AllowedAudiences[0] = "changed"
	cfg.RequiredClaims[0].Value = "changed"

	if provider.config.SigningKeys[0].Name != "primary" ||
		provider.config.AllowedAlgorithms[0] != runtimeconfig.HS256 ||
		provider.config.AllowedIssuers[0] != "issuer" ||
		provider.config.AllowedAudiences[0] != "audience" ||
		provider.config.RequiredClaims[0].Value != "internal" {
		t.Fatalf("Provider config changed with constructor input: %+v", provider.config)
	}
}

func TestJWTProviderAuthenticateSuccess(t *testing.T) {
	provider := mustJWTProvider(t, validJWTProviderConfig(), mustJWTMemoryResolver(t, jwtTestSecret))
	token := signedJWT(t, jwt.SigningMethodHS256, jwtTestSecret, jwt.MapClaims{
		"sub":    "alice",
		"active": true,
		"level":  7,
		"nested": map[string]any{"ignored": true},
		"exp":    time.Now().Add(time.Minute).Unix(),
	})

	result, err := provider.Authenticate(context.Background(), bearerRequest("Authorization", token))
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if !result.Success || result.Principal == nil || !result.Principal.Authenticated {
		t.Fatalf("Authenticate() = %+v, want authenticated Principal", result)
	}
	if result.ProviderName != "internal-jwt" || result.ProviderType != runtimeconfig.AuthenticationProviderJWT {
		t.Fatalf("Result Provider = (%q, %q)", result.ProviderName, result.ProviderType)
	}
	principal := result.Principal
	if principal.Name != "alice" || principal.AuthenticationProvider != "internal-jwt" ||
		principal.AuthenticationMethod != runtimeconfig.AuthenticationProviderJWT {
		t.Fatalf("Principal = %+v", principal)
	}
	if principal.Claims["active"] != "true" || principal.Claims["level"] != "7" {
		t.Fatalf("Principal Claims = %+v", principal.Claims)
	}
	if _, exists := principal.Claims["nested"]; exists {
		t.Fatal("non-scalar Claim was copied to Principal")
	}
}

func TestJWTProviderAuthenticateSupportedAlgorithms(t *testing.T) {
	tests := []struct {
		name      string
		algorithm runtimeconfig.JWTAlgorithm
		method    jwt.SigningMethod
	}{
		{name: "HS256", algorithm: runtimeconfig.HS256, method: jwt.SigningMethodHS256},
		{name: "HS384", algorithm: runtimeconfig.HS384, method: jwt.SigningMethodHS384},
		{name: "HS512", algorithm: runtimeconfig.HS512, method: jwt.SigningMethodHS512},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := validJWTProviderConfig()
			cfg.AllowedAlgorithms = []runtimeconfig.JWTAlgorithm{test.algorithm}
			provider := mustJWTProvider(t, cfg, mustJWTMemoryResolver(t, jwtTestSecret))
			token := signedJWT(t, test.method, jwtTestSecret, jwt.MapClaims{"sub": "alice"})

			result, err := provider.Authenticate(context.Background(), bearerRequest("Authorization", token))
			if err != nil || !result.Success {
				t.Fatalf("Authenticate() = (%+v, %v), want success", result, err)
			}
		})
	}
}

func TestJWTProviderAuthenticateTriesSigningKeysInOrderUntilMatch(t *testing.T) {
	cfg := multipleJWTSigningKeysConfig()
	resolver := &orderedJWTResolver{values: map[string][]byte{
		"secrets/jwt/first":  []byte("wrong-secret"),
		"secrets/jwt/second": []byte("matching-secret"),
	}}
	provider := mustJWTProvider(t, cfg, resolver)
	token := validSignedJWT(t, "matching-secret")

	result, err := provider.Authenticate(context.Background(), bearerRequest("Authorization", token))
	if err != nil || !result.Success {
		t.Fatalf("Authenticate() = (%+v, %v), want success from second Signing Key", result, err)
	}
	wantOrder := []string{"secrets/jwt/first", "secrets/jwt/second"}
	if !slices.Equal(resolver.calls, wantOrder) {
		t.Fatalf("Resolve order = %v, want %v", resolver.calls, wantOrder)
	}
}

func TestJWTProviderAuthenticateRejectsWhenNoSigningKeyMatches(t *testing.T) {
	cfg := multipleJWTSigningKeysConfig()
	resolver := &orderedJWTResolver{values: map[string][]byte{
		"secrets/jwt/first":  []byte("first-wrong-secret"),
		"secrets/jwt/second": []byte("second-wrong-secret"),
	}}
	provider := mustJWTProvider(t, cfg, resolver)
	token := validSignedJWT(t, "different-secret")

	result, err := provider.Authenticate(context.Background(), bearerRequest("Authorization", token))
	if err != nil || result.Success {
		t.Fatalf("Authenticate() = (%+v, %v), want rejected credentials without error", result, err)
	}
	wantOrder := []string{"secrets/jwt/first", "secrets/jwt/second"}
	if !slices.Equal(resolver.calls, wantOrder) {
		t.Fatalf("Resolve order = %v, want %v", resolver.calls, wantOrder)
	}
}

func TestJWTProviderAuthenticateReturnsErrorWhenSigningKeyResolutionFails(t *testing.T) {
	wantErr := errors.New("second key unavailable")
	cfg := multipleJWTSigningKeysConfig()
	resolver := &orderedJWTResolver{
		values: map[string][]byte{"secrets/jwt/first": []byte("first-secret")},
		errors: map[string]error{"secrets/jwt/second": wantErr},
	}
	provider := mustJWTProvider(t, cfg, resolver)

	result, err := provider.Authenticate(
		context.Background(),
		bearerRequest("Authorization", validSignedJWT(t, "first-secret")),
	)
	if result.Success {
		t.Fatalf("Authenticate() result = %+v, want failure", result)
	}
	if !errors.Is(err, ErrJWTProviderUnavailable) || !errors.Is(err, wantErr) {
		t.Fatalf("Authenticate() error = %v, want ErrJWTProviderUnavailable and Resolver error", err)
	}
	wantOrder := []string{"secrets/jwt/first", "secrets/jwt/second"}
	if !slices.Equal(resolver.calls, wantOrder) {
		t.Fatalf("Resolve order = %v, want %v", resolver.calls, wantOrder)
	}
}

func TestJWTProviderRejectsMissingOrInvalidAuthorization(t *testing.T) {
	provider := mustJWTProvider(t, validJWTProviderConfig(), mustJWTMemoryResolver(t, jwtTestSecret))
	tests := []AuthenticationRequest{
		{},
		{Headers: map[string][]string{"Authorization": {"Basic credentials"}}},
		{Headers: map[string][]string{"Authorization": {"Bearer"}}},
	}
	for _, request := range tests {
		result, err := provider.Authenticate(context.Background(), request)
		if err != nil || result.Success {
			t.Errorf("Authenticate() = (%+v, %v), want rejected credentials", result, err)
		}
	}
}

func TestJWTProviderUsesCaseInsensitiveAuthorizationHeader(t *testing.T) {
	provider := mustJWTProvider(t, validJWTProviderConfig(), mustJWTMemoryResolver(t, jwtTestSecret))
	token := validSignedJWT(t, jwtTestSecret)

	result, err := provider.Authenticate(context.Background(), bearerRequest("authorization", token))
	if err != nil || !result.Success {
		t.Fatalf("Authenticate() = (%+v, %v), want success", result, err)
	}
}

func TestJWTProviderRejectsInvalidTokens(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name   string
		config JWTProviderConfig
		token  func(*testing.T) string
	}{
		{name: "wrong signature", config: validJWTProviderConfig(), token: func(t *testing.T) string {
			return validSignedJWT(t, "wrong-secret")
		}},
		{name: "expired", config: validJWTProviderConfig(), token: func(t *testing.T) string {
			return signedJWT(t, jwt.SigningMethodHS256, jwtTestSecret, jwt.MapClaims{"sub": "alice", "exp": now.Add(-time.Minute).Unix()})
		}},
		{name: "not before", config: validJWTProviderConfig(), token: func(t *testing.T) string {
			return signedJWT(t, jwt.SigningMethodHS256, jwtTestSecret, jwt.MapClaims{"sub": "alice", "nbf": now.Add(time.Minute).Unix()})
		}},
		{name: "invalid issuer", config: configWithIssuer("trusted"), token: func(t *testing.T) string {
			return signedJWT(t, jwt.SigningMethodHS256, jwtTestSecret, jwt.MapClaims{"sub": "alice", "iss": "untrusted"})
		}},
		{name: "invalid audience", config: configWithAudience("expected"), token: func(t *testing.T) string {
			return signedJWT(t, jwt.SigningMethodHS256, jwtTestSecret, jwt.MapClaims{"sub": "alice", "aud": []string{"other"}})
		}},
		{name: "missing required claim", config: configWithRequiredClaim("tenant", "internal"), token: func(t *testing.T) string {
			return signedJWT(t, jwt.SigningMethodHS256, jwtTestSecret, jwt.MapClaims{"sub": "alice"})
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := mustJWTProvider(t, test.config, mustJWTMemoryResolver(t, jwtTestSecret))
			result, err := provider.Authenticate(context.Background(), bearerRequest("Authorization", test.token(t)))
			if err != nil || result.Success {
				t.Fatalf("Authenticate() = (%+v, %v), want rejected credentials", result, err)
			}
		})
	}
}

func TestJWTProviderAcceptsIssuerAudienceAndRequiredClaim(t *testing.T) {
	cfg := validJWTProviderConfig()
	cfg.AllowedIssuers = []string{"trusted", "partner"}
	cfg.AllowedAudiences = []string{"service-a", "service-b"}
	cfg.RequiredClaims = []JWTRequiredClaimConfig{{Name: "tenant", Value: "internal"}}
	provider := mustJWTProvider(t, cfg, mustJWTMemoryResolver(t, jwtTestSecret))
	token := signedJWT(t, jwt.SigningMethodHS256, jwtTestSecret, jwt.MapClaims{
		"sub": "alice", "iss": "partner", "aud": []string{"service-b"}, "tenant": "internal",
	})

	result, err := provider.Authenticate(context.Background(), bearerRequest("Authorization", token))
	if err != nil || !result.Success {
		t.Fatalf("Authenticate() = (%+v, %v), want success", result, err)
	}
}

func TestJWTProviderReturnsCanceledContext(t *testing.T) {
	provider := mustJWTProvider(t, validJWTProviderConfig(), mustJWTMemoryResolver(t, jwtTestSecret))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := provider.Authenticate(ctx, bearerRequest("Authorization", validSignedJWT(t, jwtTestSecret)))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Authenticate() error = %v, want context.Canceled", err)
	}
}

func TestJWTProviderReturnsResolverError(t *testing.T) {
	wantErr := errors.New("resolver unavailable")
	provider := mustJWTProvider(t, validJWTProviderConfig(), jwtErrorResolver{err: wantErr})

	_, err := provider.Authenticate(context.Background(), bearerRequest("Authorization", validSignedJWT(t, jwtTestSecret)))
	if !errors.Is(err, ErrJWTProviderUnavailable) || !errors.Is(err, wantErr) {
		t.Fatalf("Authenticate() error = %v, want ErrJWTProviderUnavailable and resolver error", err)
	}
}

func TestJWTProviderUsesRotatedSecretOnEveryAuthenticate(t *testing.T) {
	resolver := mustJWTMemoryResolver(t, "first-secret")
	provider := mustJWTProvider(t, validJWTProviderConfig(), resolver)
	firstToken := validSignedJWT(t, "first-secret")
	secondToken := validSignedJWT(t, "second-secret")

	first, err := provider.Authenticate(context.Background(), bearerRequest("Authorization", firstToken))
	if err != nil || !first.Success {
		t.Fatalf("Authenticate(first) = (%+v, %v), want success", first, err)
	}
	if err := resolver.Set(jwtTestSecretRef, []byte("second-secret")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	oldResult, err := provider.Authenticate(context.Background(), bearerRequest("Authorization", firstToken))
	if err != nil || oldResult.Success {
		t.Fatalf("Authenticate(old after rotation) = (%+v, %v), want rejection", oldResult, err)
	}
	second, err := provider.Authenticate(context.Background(), bearerRequest("Authorization", secondToken))
	if err != nil || !second.Success {
		t.Fatalf("Authenticate(second) = (%+v, %v), want success", second, err)
	}
}

func TestJWTProviderDoesNotExposeTokenOrSecretInErrors(t *testing.T) {
	token := validSignedJWT(t, jwtTestSecret)
	provider := mustJWTProvider(t, validJWTProviderConfig(), jwtErrorResolver{err: errors.New("storage unavailable")})

	_, err := provider.Authenticate(context.Background(), bearerRequest("Authorization", token))
	if err == nil {
		t.Fatal("Authenticate() error = nil, want Resolver error")
	}
	if strings.Contains(err.Error(), token) || strings.Contains(err.Error(), jwtTestSecret) {
		t.Fatalf("Authenticate() error exposes token or Secret: %v", err)
	}
}

func TestJWTProviderSmokeScenario(t *testing.T) {
	provider := mustJWTProvider(t, configWithRequiredClaim("tenant", "internal"), mustJWTMemoryResolver(t, jwtTestSecret))
	token := signedJWT(t, jwt.SigningMethodHS256, jwtTestSecret, jwt.MapClaims{
		"sub": "smoke-user", "tenant": "internal", "exp": time.Now().Add(time.Minute).Unix(),
	})

	result, err := provider.Authenticate(context.Background(), bearerRequest("Authorization", token))
	if err != nil || !result.Success || result.Principal == nil {
		t.Fatalf("Authenticate() = (%+v, %v), want Principal", result, err)
	}
	t.Logf("Bearer JWT -> HS256 verification -> Principal %q", result.Principal.Name)
}

func validJWTProviderConfig() JWTProviderConfig {
	return JWTProviderConfig{
		Name: "internal-jwt",
		SigningKeys: []JWTSigningKeyConfig{
			{Name: "primary", SecretRef: jwtTestSecretRef},
		},
		AllowedAlgorithms: []runtimeconfig.JWTAlgorithm{runtimeconfig.HS256},
		ClockSkewSeconds:  0,
	}
}

func configWithIssuer(issuer string) JWTProviderConfig {
	cfg := validJWTProviderConfig()
	cfg.AllowedIssuers = []string{issuer}
	return cfg
}

func configWithAudience(audience string) JWTProviderConfig {
	cfg := validJWTProviderConfig()
	cfg.AllowedAudiences = []string{audience}
	return cfg
}

func configWithRequiredClaim(name, value string) JWTProviderConfig {
	cfg := validJWTProviderConfig()
	cfg.RequiredClaims = []JWTRequiredClaimConfig{{Name: name, Value: value}}
	return cfg
}

func multipleJWTSigningKeysConfig() JWTProviderConfig {
	cfg := validJWTProviderConfig()
	cfg.SigningKeys = []JWTSigningKeyConfig{
		{Name: "first", SecretRef: "secrets/jwt/first"},
		{Name: "second", SecretRef: "secrets/jwt/second"},
	}
	return cfg
}

func mustJWTProvider(t *testing.T, cfg JWTProviderConfig, resolver secretresolver.Resolver) *JWTProvider {
	t.Helper()
	provider, err := NewJWTProvider(cfg, resolver)
	if err != nil {
		t.Fatalf("NewJWTProvider() error = %v", err)
	}
	return provider
}

func mustJWTMemoryResolver(t *testing.T, secret string) *secretresolver.MemoryResolver {
	t.Helper()
	resolver, err := secretresolver.NewMemory(map[string][]byte{jwtTestSecretRef: []byte(secret)})
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	return resolver
}

func signedJWT(t *testing.T, method jwt.SigningMethod, secret string, claims jwt.MapClaims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return token
}

func validSignedJWT(t *testing.T, secret string) string {
	t.Helper()
	return signedJWT(t, jwt.SigningMethodHS256, secret, jwt.MapClaims{
		"sub": "alice", "exp": time.Now().Add(time.Minute).Unix(),
	})
}

func bearerRequest(headerName, token string) AuthenticationRequest {
	return AuthenticationRequest{Headers: map[string][]string{headerName: {"Bearer " + token}}}
}

type jwtErrorResolver struct {
	err error
}

func (resolver jwtErrorResolver) Resolve(context.Context, string) (secretresolver.Secret, error) {
	return secretresolver.Secret{}, resolver.err
}

type orderedJWTResolver struct {
	values map[string][]byte
	errors map[string]error
	calls  []string
}

func (resolver *orderedJWTResolver) Resolve(ctx context.Context, ref string) (secretresolver.Secret, error) {
	if err := ctx.Err(); err != nil {
		return secretresolver.Secret{}, err
	}
	resolver.calls = append(resolver.calls, ref)
	if err := resolver.errors[ref]; err != nil {
		return secretresolver.Secret{}, err
	}
	value, exists := resolver.values[ref]
	if !exists {
		return secretresolver.Secret{}, secretresolver.ErrSecretNotFound
	}
	return secretresolver.Secret{Value: slices.Clone(value)}, nil
}
