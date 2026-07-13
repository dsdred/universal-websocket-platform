package configurationversion

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/net/http/httpguts"
)

const (
	maxAuthenticationProviderNameLength = 255
	maxAPIKeyHeaderLength               = 255
	maxSecretReferenceLength            = 255
)

// AuthenticationValidator validates and normalizes Authentication metadata.
type AuthenticationValidator interface {
	Validate(AuthenticationSettings) error
}

// DefaultAuthenticationValidator applies the default Authentication business rules.
type DefaultAuthenticationValidator struct{}

// Validate validates Authentication settings and normalizes their mutable copy in place.
func (DefaultAuthenticationValidator) Validate(authentication AuthenticationSettings) error {
	if authentication.Enabled && len(authentication.Providers) == 0 {
		return &ValidationError{Field: "providers", Message: "must contain at least one provider when Authentication is enabled"}
	}

	names := make(map[string]struct{}, len(authentication.Providers))
	priorities := make(map[uint32]struct{}, len(authentication.Providers))
	enabledProviders := 0
	for index, provider := range authentication.Providers {
		provider.Name = strings.TrimSpace(provider.Name)
		if provider.Name == "" {
			return &ValidationError{Field: "providers.name", Message: "must not be empty"}
		}
		if utf8.RuneCountInString(provider.Name) > maxAuthenticationProviderNameLength {
			return &ValidationError{Field: "providers.name", Message: "must not exceed 255 characters"}
		}
		if _, exists := names[provider.Name]; exists {
			return &ValidationError{Field: "providers.name", Message: "must be unique"}
		}
		names[provider.Name] = struct{}{}
		if _, exists := priorities[provider.Priority]; exists {
			return &ValidationError{Field: "providers.priority", Message: "must be unique"}
		}
		priorities[provider.Priority] = struct{}{}
		if provider.Enabled {
			enabledProviders++
		}

		if err := validateAuthenticationProvider(&provider); err != nil {
			return err
		}
		authentication.Providers[index] = provider
	}

	if authentication.Enabled && enabledProviders == 0 {
		return &ValidationError{Field: "providers.enabled", Message: "must contain at least one enabled Provider when Authentication is enabled"}
	}
	return nil
}

func validateAuthenticationProvider(provider *AuthenticationProvider) error {
	switch provider.Type {
	case AuthenticationProviderAPIKey:
		if provider.Basic != nil {
			return &ValidationError{Field: "providers.basic", Message: "must be omitted for non-basic Provider"}
		}
		if provider.JWT != nil {
			return &ValidationError{Field: "providers.jwt", Message: "must be omitted for non-jwt Provider"}
		}
		if provider.APIKey == nil {
			return &ValidationError{Field: "providers.apiKey", Message: "must be provided for api-key Provider"}
		}
		return validateAPIKey(provider.APIKey)
	case AuthenticationProviderJWT:
		if provider.Basic != nil {
			return &ValidationError{Field: "providers.basic", Message: "must be omitted for non-basic Provider"}
		}
		if provider.APIKey != nil {
			return &ValidationError{Field: "providers.apiKey", Message: "must be omitted for non-api-key Provider"}
		}
		if provider.JWT == nil {
			return &ValidationError{Field: "providers.jwt", Message: "must be provided for jwt Provider"}
		}
		return validateJWT(provider.JWT)
	case AuthenticationProviderBasic:
		if provider.APIKey != nil {
			return &ValidationError{Field: "providers.apiKey", Message: "must be omitted for non-api-key Provider"}
		}
		if provider.JWT != nil {
			return &ValidationError{Field: "providers.jwt", Message: "must be omitted for non-jwt Provider"}
		}
		if provider.Basic == nil {
			return &ValidationError{Field: "providers.basic", Message: "must be provided for basic Provider"}
		}
		return validateBasic(provider.Basic)
	default:
		return &ValidationError{Field: "providers.type", Message: "must be one of jwt, api-key, or basic"}
	}
}

func validateAPIKey(apiKey *APIKeySettings) error {
	apiKey.Header = strings.TrimSpace(apiKey.Header)
	if apiKey.Header == "" {
		return &ValidationError{Field: "providers.apiKey.header", Message: "must not be empty"}
	}
	if utf8.RuneCountInString(apiKey.Header) > maxAPIKeyHeaderLength {
		return &ValidationError{Field: "providers.apiKey.header", Message: "must not exceed 255 characters"}
	}
	if !httpguts.ValidHeaderFieldName(apiKey.Header) {
		return &ValidationError{Field: "providers.apiKey.header", Message: "must be a valid HTTP header field name"}
	}
	apiKey.SecretRef = strings.TrimSpace(apiKey.SecretRef)
	return validateSecretReference(apiKey.SecretRef, "providers.apiKey.secretRef")
}

func validateBasic(basic *BasicSettings) error {
	basic.Realm = strings.TrimSpace(basic.Realm)
	if basic.Realm == "" {
		return &ValidationError{Field: "providers.basic.realm", Message: "must not be empty"}
	}
	if utf8.RuneCountInString(basic.Realm) > 255 {
		return &ValidationError{Field: "providers.basic.realm", Message: "must not exceed 255 characters"}
	}
	basic.SecretRef = strings.TrimSpace(basic.SecretRef)
	return validateSecretReference(basic.SecretRef, "providers.basic.secretRef")
}

func validateJWT(jwt *JWTSettings) error {
	if len(jwt.SigningKeys) == 0 {
		return &ValidationError{Field: "providers.jwt.signingKeys", Message: "must contain at least one signing key"}
	}
	keyNames := make(map[string]struct{}, len(jwt.SigningKeys))
	for index, key := range jwt.SigningKeys {
		key.Name = strings.TrimSpace(key.Name)
		if key.Name == "" {
			return &ValidationError{Field: "providers.jwt.signingKeys.name", Message: "must not be empty"}
		}
		if _, exists := keyNames[key.Name]; exists {
			return &ValidationError{Field: "providers.jwt.signingKeys.name", Message: "must be unique"}
		}
		keyNames[key.Name] = struct{}{}
		key.SecretRef = strings.TrimSpace(key.SecretRef)
		if err := validateSecretReference(key.SecretRef, "providers.jwt.signingKeys.secretRef"); err != nil {
			return err
		}
		jwt.SigningKeys[index] = key
	}

	if len(jwt.AllowedAlgorithms) == 0 {
		return &ValidationError{Field: "providers.jwt.allowedAlgorithms", Message: "must contain at least one algorithm"}
	}
	seenAlgorithms := make(map[JWTAlgorithm]struct{}, len(jwt.AllowedAlgorithms))
	for _, algorithm := range jwt.AllowedAlgorithms {
		if !validJWTAlgorithm(algorithm) {
			return &ValidationError{Field: "providers.jwt.allowedAlgorithms", Message: "contains an unsupported algorithm"}
		}
		if _, exists := seenAlgorithms[algorithm]; exists {
			return &ValidationError{Field: "providers.jwt.allowedAlgorithms", Message: "must not contain duplicates"}
		}
		seenAlgorithms[algorithm] = struct{}{}
	}

	var err error
	jwt.AllowedIssuers, err = normalizeUniqueStrings(jwt.AllowedIssuers, "providers.jwt.allowedIssuers")
	if err != nil {
		return err
	}
	jwt.AllowedAudiences, err = normalizeUniqueStrings(jwt.AllowedAudiences, "providers.jwt.allowedAudiences")
	if err != nil {
		return err
	}

	claimNames := make(map[string]struct{}, len(jwt.RequiredClaims))
	for index, claim := range jwt.RequiredClaims {
		claim.Name = strings.TrimSpace(claim.Name)
		claim.Value = strings.TrimSpace(claim.Value)
		if claim.Name == "" {
			return &ValidationError{Field: "providers.jwt.requiredClaims.name", Message: "must not be empty"}
		}
		if claim.Value == "" {
			return &ValidationError{Field: "providers.jwt.requiredClaims.value", Message: "must not be empty"}
		}
		if _, exists := claimNames[claim.Name]; exists {
			return &ValidationError{Field: "providers.jwt.requiredClaims.name", Message: "must be unique"}
		}
		claimNames[claim.Name] = struct{}{}
		jwt.RequiredClaims[index] = claim
	}
	if jwt.ClockSkewSeconds > 300 {
		return &ValidationError{Field: "providers.jwt.clockSkewSeconds", Message: "must be between 0 and 300"}
	}
	return nil
}

func validateSecretReference(reference, field string) error {
	if reference == "" {
		return &ValidationError{Field: field, Message: "must not be empty"}
	}
	if utf8.RuneCountInString(reference) > maxSecretReferenceLength || !validAuthenticationSecretReference(reference) {
		return &ValidationError{Field: field, Message: "must be a valid Secret Reference"}
	}
	return nil
}

func validAuthenticationSecretReference(reference string) bool {
	if strings.HasPrefix(reference, "/") || strings.HasSuffix(reference, "/") ||
		strings.Contains(reference, "//") || strings.Contains(reference, "://") ||
		strings.Contains(reference, "-----BEGIN") {
		return false
	}
	for _, character := range reference {
		if !((character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			character == '/' || character == '-' || character == '_' || character == '.') {
			return false
		}
	}
	return true
}

func validJWTAlgorithm(algorithm JWTAlgorithm) bool {
	switch algorithm {
	case HS256, HS384, HS512, RS256, RS384, RS512, ES256, ES384, ES512, PS256, PS384, PS512:
		return true
	default:
		return false
	}
}

func normalizeUniqueStrings(values []string, field string) ([]string, error) {
	normalized := make([]string, len(values))
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, &ValidationError{Field: field, Message: "must not contain empty values"}
		}
		if _, exists := seen[value]; exists {
			return nil, &ValidationError{Field: field, Message: "must not contain duplicates"}
		}
		seen[value] = struct{}{}
		normalized[index] = value
	}
	return normalized, nil
}

func cloneAuthenticationSettings(authentication AuthenticationSettings) AuthenticationSettings {
	clone := AuthenticationSettings{Enabled: authentication.Enabled, Providers: make([]AuthenticationProvider, len(authentication.Providers))}
	for index, provider := range authentication.Providers {
		if provider.APIKey != nil {
			apiKey := *provider.APIKey
			provider.APIKey = &apiKey
		}
		if provider.Basic != nil {
			basic := *provider.Basic
			provider.Basic = &basic
		}
		if provider.JWT != nil {
			jwt := *provider.JWT
			jwt.SigningKeys = append([]JWTSigningKey(nil), jwt.SigningKeys...)
			jwt.AllowedAlgorithms = append([]JWTAlgorithm(nil), jwt.AllowedAlgorithms...)
			jwt.AllowedIssuers = append([]string(nil), jwt.AllowedIssuers...)
			jwt.AllowedAudiences = append([]string(nil), jwt.AllowedAudiences...)
			jwt.RequiredClaims = append([]JWTRequiredClaim(nil), jwt.RequiredClaims...)
			provider.JWT = &jwt
		}
		clone.Providers[index] = provider
	}
	return clone
}
