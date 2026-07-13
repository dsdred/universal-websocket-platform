package configurationversion

import (
	"errors"
	"strings"
	"testing"
)

func TestDefaultAuthenticationValidatorCrossProviderRules(t *testing.T) {
	validator := DefaultAuthenticationValidator{}

	tests := []struct {
		name     string
		settings AuthenticationSettings
		field    string
	}{
		{name: "disabled with Providers", settings: AuthenticationSettings{Enabled: false, Providers: []AuthenticationProvider{validAPIKeyProvider("key", 10, false)}}},
		{name: "enabled with no Providers", settings: AuthenticationSettings{Enabled: true}, field: "providers"},
		{name: "enabled with no enabled Providers", settings: AuthenticationSettings{Enabled: true, Providers: []AuthenticationProvider{validAPIKeyProvider("key", 10, false)}}, field: "providers.enabled"},
		{name: "duplicate Provider name", settings: AuthenticationSettings{Enabled: true, Providers: []AuthenticationProvider{validAPIKeyProvider("same", 10, true), validBasicProvider(" same ", 20, true)}}, field: "providers.name"},
		{name: "duplicate Priority", settings: AuthenticationSettings{Enabled: true, Providers: []AuthenticationProvider{validAPIKeyProvider("key", 10, true), validBasicProvider("basic", 10, true)}}, field: "providers.priority"},
		{name: "multiple JWT Providers", settings: AuthenticationSettings{Enabled: true, Providers: []AuthenticationProvider{validJWTProvider("jwt-a", 10, true), validJWTProvider("jwt-b", 20, true)}}},
		{name: "multiple API Key Providers", settings: AuthenticationSettings{Enabled: true, Providers: []AuthenticationProvider{validAPIKeyProvider("key-a", 10, true), validAPIKeyProvider("key-b", 20, true)}}},
		{name: "mixed Providers", settings: AuthenticationSettings{Enabled: true, Providers: []AuthenticationProvider{validJWTProvider("jwt", 10, true), validAPIKeyProvider("key", 20, true), validBasicProvider("basic", 30, true)}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(cloneAuthenticationSettings(tt.settings))
			if tt.field == "" {
				if err != nil {
					t.Errorf("Validate() error = %v", err)
				}
				return
			}
			assertValidationField(t, err, tt.field)
		})
	}
}

func TestDefaultAuthenticationValidatorProviderRules(t *testing.T) {
	validAPIKey := &APIKeySettings{Header: "X-API-Key", SecretRef: "secrets/api-keys/main"}
	validBasic := &BasicSettings{Realm: "Realm", SecretRef: "secrets/basic/main"}

	tests := []struct {
		name     string
		provider AuthenticationProvider
		field    string
	}{
		{name: "empty Provider name", provider: AuthenticationProvider{Type: AuthenticationProviderAPIKey, Enabled: true, APIKey: validAPIKey}, field: "providers.name"},
		{name: "long Provider name", provider: AuthenticationProvider{Name: strings.Repeat("я", 256), Type: AuthenticationProviderAPIKey, Enabled: true, APIKey: validAPIKey}, field: "providers.name"},
		{name: "unknown Provider type", provider: AuthenticationProvider{Name: "unknown", Type: "unknown", Enabled: true}, field: "providers.type"},
		{name: "API Key settings missing", provider: AuthenticationProvider{Name: "key", Type: AuthenticationProviderAPIKey, Enabled: true}, field: "providers.apiKey"},
		{name: "API Key empty Header", provider: AuthenticationProvider{Name: "key", Type: AuthenticationProviderAPIKey, Enabled: true, APIKey: &APIKeySettings{SecretRef: validAPIKey.SecretRef}}, field: "providers.apiKey.header"},
		{name: "API Key invalid Header", provider: AuthenticationProvider{Name: "key", Type: AuthenticationProviderAPIKey, Enabled: true, APIKey: &APIKeySettings{Header: "X Key", SecretRef: validAPIKey.SecretRef}}, field: "providers.apiKey.header"},
		{name: "API Key invalid SecretRef", provider: AuthenticationProvider{Name: "key", Type: AuthenticationProviderAPIKey, Enabled: true, APIKey: &APIKeySettings{Header: validAPIKey.Header, SecretRef: "https://example.com/secret"}}, field: "providers.apiKey.secretRef"},
		{name: "API Key with JWT", provider: AuthenticationProvider{Name: "key", Type: AuthenticationProviderAPIKey, Enabled: true, APIKey: validAPIKey, JWT: validJWTSettings()}, field: "providers.jwt"},
		{name: "JWT settings missing", provider: AuthenticationProvider{Name: "jwt", Type: AuthenticationProviderJWT, Enabled: true}, field: "providers.jwt"},
		{name: "JWT with API Key", provider: AuthenticationProvider{Name: "jwt", Type: AuthenticationProviderJWT, Enabled: true, JWT: validJWTSettings(), APIKey: validAPIKey}, field: "providers.apiKey"},
		{name: "JWT no Signing Keys", provider: AuthenticationProvider{Name: "jwt", Type: AuthenticationProviderJWT, Enabled: true, JWT: &JWTSettings{AllowedAlgorithms: []JWTAlgorithm{HS256}}}, field: "providers.jwt.signingKeys"},
		{name: "JWT duplicate Signing Key", provider: AuthenticationProvider{Name: "jwt", Type: AuthenticationProviderJWT, Enabled: true, JWT: &JWTSettings{SigningKeys: []JWTSigningKey{{Name: "main", SecretRef: "secrets/jwt/a"}, {Name: " main ", SecretRef: "secrets/jwt/b"}}, AllowedAlgorithms: []JWTAlgorithm{HS256}}}, field: "providers.jwt.signingKeys.name"},
		{name: "JWT invalid SecretRef", provider: AuthenticationProvider{Name: "jwt", Type: AuthenticationProviderJWT, Enabled: true, JWT: &JWTSettings{SigningKeys: []JWTSigningKey{{Name: "main", SecretRef: `C:\secrets\key`}}, AllowedAlgorithms: []JWTAlgorithm{HS256}}}, field: "providers.jwt.signingKeys.secretRef"},
		{name: "JWT no Algorithms", provider: AuthenticationProvider{Name: "jwt", Type: AuthenticationProviderJWT, Enabled: true, JWT: &JWTSettings{SigningKeys: []JWTSigningKey{{Name: "main", SecretRef: "secrets/jwt/main"}}}}, field: "providers.jwt.allowedAlgorithms"},
		{name: "JWT invalid Algorithm", provider: AuthenticationProvider{Name: "jwt", Type: AuthenticationProviderJWT, Enabled: true, JWT: &JWTSettings{SigningKeys: []JWTSigningKey{{Name: "main", SecretRef: "secrets/jwt/main"}}, AllowedAlgorithms: []JWTAlgorithm{"none"}}}, field: "providers.jwt.allowedAlgorithms"},
		{name: "JWT duplicate Algorithm", provider: AuthenticationProvider{Name: "jwt", Type: AuthenticationProviderJWT, Enabled: true, JWT: &JWTSettings{SigningKeys: []JWTSigningKey{{Name: "main", SecretRef: "secrets/jwt/main"}}, AllowedAlgorithms: []JWTAlgorithm{HS256, HS256}}}, field: "providers.jwt.allowedAlgorithms"},
		{name: "JWT empty Issuer", provider: AuthenticationProvider{Name: "jwt", Type: AuthenticationProviderJWT, Enabled: true, JWT: jwtWithStrings([]string{" "}, nil, nil)}, field: "providers.jwt.allowedIssuers"},
		{name: "JWT duplicate Issuer", provider: AuthenticationProvider{Name: "jwt", Type: AuthenticationProviderJWT, Enabled: true, JWT: jwtWithStrings([]string{"issuer", " issuer "}, nil, nil)}, field: "providers.jwt.allowedIssuers"},
		{name: "JWT duplicate Audience", provider: AuthenticationProvider{Name: "jwt", Type: AuthenticationProviderJWT, Enabled: true, JWT: jwtWithStrings(nil, []string{"aud", " aud "}, nil)}, field: "providers.jwt.allowedAudiences"},
		{name: "JWT empty Required Claim value", provider: AuthenticationProvider{Name: "jwt", Type: AuthenticationProviderJWT, Enabled: true, JWT: jwtWithStrings(nil, nil, []JWTRequiredClaim{{Name: "tenant", Value: " "}})}, field: "providers.jwt.requiredClaims.value"},
		{name: "JWT duplicate Required Claim", provider: AuthenticationProvider{Name: "jwt", Type: AuthenticationProviderJWT, Enabled: true, JWT: jwtWithStrings(nil, nil, []JWTRequiredClaim{{Name: "tenant", Value: "a"}, {Name: " tenant ", Value: "b"}})}, field: "providers.jwt.requiredClaims.name"},
		{name: "JWT ClockSkew", provider: AuthenticationProvider{Name: "jwt", Type: AuthenticationProviderJWT, Enabled: true, JWT: &JWTSettings{SigningKeys: []JWTSigningKey{{Name: "main", SecretRef: "secrets/jwt/main"}}, AllowedAlgorithms: []JWTAlgorithm{HS256}, ClockSkewSeconds: 301}}, field: "providers.jwt.clockSkewSeconds"},
		{name: "Basic settings missing", provider: AuthenticationProvider{Name: "basic", Type: AuthenticationProviderBasic, Enabled: true}, field: "providers.basic"},
		{name: "Basic empty Realm", provider: AuthenticationProvider{Name: "basic", Type: AuthenticationProviderBasic, Enabled: true, Basic: &BasicSettings{SecretRef: validBasic.SecretRef}}, field: "providers.basic.realm"},
		{name: "Basic invalid SecretRef", provider: AuthenticationProvider{Name: "basic", Type: AuthenticationProviderBasic, Enabled: true, Basic: &BasicSettings{Realm: validBasic.Realm, SecretRef: "-----BEGIN PRIVATE KEY-----"}}, field: "providers.basic.secretRef"},
		{name: "Basic with JWT", provider: AuthenticationProvider{Name: "basic", Type: AuthenticationProviderBasic, Enabled: true, Basic: validBasic, JWT: validJWTSettings()}, field: "providers.jwt"},
	}

	validator := DefaultAuthenticationValidator{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(AuthenticationSettings{Enabled: true, Providers: []AuthenticationProvider{tt.provider}})
			assertValidationField(t, err, tt.field)
		})
	}
}

func TestDefaultAuthenticationValidatorNormalizesMetadata(t *testing.T) {
	settings := AuthenticationSettings{Enabled: true, Providers: []AuthenticationProvider{{
		Name: " jwt ", Type: AuthenticationProviderJWT, Enabled: true, JWT: &JWTSettings{
			SigningKeys:       []JWTSigningKey{{Name: " main ", SecretRef: " secrets/jwt/main "}},
			AllowedAlgorithms: []JWTAlgorithm{HS256},
			AllowedIssuers:    []string{" issuer "},
			AllowedAudiences:  []string{" audience "},
			RequiredClaims:    []JWTRequiredClaim{{Name: " tenant ", Value: " internal "}},
		},
	}}}

	if err := (DefaultAuthenticationValidator{}).Validate(settings); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	provider := settings.Providers[0]
	if provider.Name != "jwt" || provider.JWT.SigningKeys[0].Name != "main" || provider.JWT.SigningKeys[0].SecretRef != "secrets/jwt/main" {
		t.Errorf("normalized Provider = %#v", provider)
	}
	if provider.JWT.AllowedIssuers[0] != "issuer" || provider.JWT.AllowedAudiences[0] != "audience" || provider.JWT.RequiredClaims[0] != (JWTRequiredClaim{Name: "tenant", Value: "internal"}) {
		t.Errorf("normalized JWT = %#v", provider.JWT)
	}
}

func validJWTProvider(name string, priority uint32, enabled bool) AuthenticationProvider {
	return AuthenticationProvider{Name: name, Type: AuthenticationProviderJWT, Enabled: enabled, Priority: priority, JWT: validJWTSettings()}
}

func validAPIKeyProvider(name string, priority uint32, enabled bool) AuthenticationProvider {
	return AuthenticationProvider{Name: name, Type: AuthenticationProviderAPIKey, Enabled: enabled, Priority: priority, APIKey: &APIKeySettings{Header: "X-API-Key", SecretRef: "secrets/api-keys/" + name}}
}

func validBasicProvider(name string, priority uint32, enabled bool) AuthenticationProvider {
	return AuthenticationProvider{Name: name, Type: AuthenticationProviderBasic, Enabled: enabled, Priority: priority, Basic: &BasicSettings{Realm: "Realm", SecretRef: "secrets/basic/" + name}}
}

func jwtWithStrings(issuers, audiences []string, claims []JWTRequiredClaim) *JWTSettings {
	return &JWTSettings{
		SigningKeys:       []JWTSigningKey{{Name: "main", SecretRef: "secrets/jwt/main"}},
		AllowedAlgorithms: []JWTAlgorithm{HS256},
		AllowedIssuers:    issuers,
		AllowedAudiences:  audiences,
		RequiredClaims:    claims,
	}
}

func assertValidationField(t *testing.T, err error, field string) {
	t.Helper()
	var validationError *ValidationError
	if !errors.As(err, &validationError) || validationError.Field != field {
		t.Errorf("Validate() error = %v, want ValidationError for %s", err, field)
	}
}
