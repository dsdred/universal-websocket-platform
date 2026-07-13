package authentication

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

var (
	ErrInvalidJWTProviderConfig = errors.New("invalid JWT Provider configuration")
	ErrUnsupportedJWTAlgorithm  = errors.New("unsupported JWT runtime algorithm")
	ErrJWTProviderUnavailable   = errors.New("JWT Provider unavailable")
)

// JWTProviderConfig contains runtime configuration required by JWTProvider.
type JWTProviderConfig struct {
	Name              string
	SigningKeys       []JWTSigningKeyConfig
	AllowedAlgorithms []runtimeconfig.JWTAlgorithm
	AllowedIssuers    []string
	AllowedAudiences  []string
	RequiredClaims    []JWTRequiredClaimConfig
	ClockSkewSeconds  uint32
}

// JWTSigningKeyConfig identifies a symmetric signing key by Secret Reference.
type JWTSigningKeyConfig struct {
	Name      string
	SecretRef string
}

// JWTRequiredClaimConfig describes one required scalar Claim value.
type JWTRequiredClaimConfig struct {
	Name  string
	Value string
}

// JWTProvider verifies signed JWT credentials using Secrets resolved per request.
type JWTProvider struct {
	config   JWTProviderConfig
	resolver secretresolver.Resolver
}

// NewJWTProvider validates and copies Provider-local runtime configuration.
func NewJWTProvider(cfg JWTProviderConfig, resolver secretresolver.Resolver) (*JWTProvider, error) {
	if resolver == nil {
		return nil, ErrNilSecretResolver
	}

	validated, err := validateJWTProviderConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &JWTProvider{config: validated, resolver: resolver}, nil
}

// Name returns the configured Provider name.
func (provider *JWTProvider) Name() string {
	return provider.config.Name
}

// Type returns the JWT Provider type.
func (*JWTProvider) Type() runtimeconfig.AuthenticationProviderType {
	return runtimeconfig.AuthenticationProviderJWT
}

// Authenticate verifies a Bearer JWT without retaining token or Secret values.
func (provider *JWTProvider) Authenticate(
	ctx context.Context,
	request AuthenticationRequest,
) (AuthenticationResult, error) {
	if err := ctx.Err(); err != nil {
		return AuthenticationResult{}, err
	}

	tokenString, exists := bearerToken(request)
	if !exists {
		return provider.result(false, nil), nil
	}

	verificationKeys, secrets, err := provider.resolveSigningKeys(ctx)
	if err != nil {
		return provider.result(false, nil), err
	}
	defer clearResolvedSecrets(secrets)

	claims := jwt.MapClaims{}
	parsed, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(*jwt.Token) (any, error) {
			return jwt.VerificationKeySet{Keys: verificationKeys}, nil
		},
		jwt.WithValidMethods(provider.algorithmNames()),
		jwt.WithLeeway(time.Duration(provider.config.ClockSkewSeconds)*time.Second),
		jwt.WithJSONNumber(),
	)
	if err != nil || parsed == nil || !parsed.Valid {
		return provider.result(false, nil), nil
	}
	if !provider.claimsAllowed(claims) {
		return provider.result(false, nil), nil
	}

	normalizedClaims := scalarClaims(claims)
	principal := &Principal{
		Name:                   normalizedClaims["sub"],
		AuthenticationType:     runtimeconfig.AuthenticationProviderJWT,
		AuthenticationProvider: provider.config.Name,
		AuthenticationMethod:   runtimeconfig.AuthenticationProviderJWT,
		Claims:                 normalizedClaims,
		Roles:                  make([]string, 0),
		Authenticated:          true,
	}
	return provider.result(true, principal), nil
}

func validateJWTProviderConfig(cfg JWTProviderConfig) (JWTProviderConfig, error) {
	result := JWTProviderConfig{
		Name:             strings.TrimSpace(cfg.Name),
		ClockSkewSeconds: cfg.ClockSkewSeconds,
	}
	if result.Name == "" || len(cfg.SigningKeys) == 0 || len(cfg.AllowedAlgorithms) == 0 || cfg.ClockSkewSeconds > 300 {
		return JWTProviderConfig{}, ErrInvalidJWTProviderConfig
	}

	keyNames := make(map[string]struct{}, len(cfg.SigningKeys))
	result.SigningKeys = make([]JWTSigningKeyConfig, len(cfg.SigningKeys))
	for index, key := range cfg.SigningKeys {
		name := strings.TrimSpace(key.Name)
		secretRef := strings.TrimSpace(key.SecretRef)
		if name == "" {
			return JWTProviderConfig{}, ErrInvalidJWTProviderConfig
		}
		if _, duplicate := keyNames[name]; duplicate {
			return JWTProviderConfig{}, ErrInvalidJWTProviderConfig
		}
		if err := secretresolver.ValidateReference(secretRef); err != nil {
			return JWTProviderConfig{}, fmt.Errorf("%w: %w", ErrInvalidJWTProviderConfig, err)
		}
		keyNames[name] = struct{}{}
		result.SigningKeys[index] = JWTSigningKeyConfig{Name: name, SecretRef: secretRef}
	}

	algorithms := make(map[runtimeconfig.JWTAlgorithm]struct{}, len(cfg.AllowedAlgorithms))
	result.AllowedAlgorithms = make([]runtimeconfig.JWTAlgorithm, len(cfg.AllowedAlgorithms))
	for index, algorithm := range cfg.AllowedAlgorithms {
		if _, duplicate := algorithms[algorithm]; duplicate {
			return JWTProviderConfig{}, ErrInvalidJWTProviderConfig
		}
		switch algorithm {
		case runtimeconfig.HS256, runtimeconfig.HS384, runtimeconfig.HS512:
		case runtimeconfig.RS256, runtimeconfig.RS384, runtimeconfig.RS512,
			runtimeconfig.ES256, runtimeconfig.ES384, runtimeconfig.ES512,
			runtimeconfig.PS256, runtimeconfig.PS384, runtimeconfig.PS512:
			return JWTProviderConfig{}, fmt.Errorf("%w: %s", ErrUnsupportedJWTAlgorithm, algorithm)
		default:
			return JWTProviderConfig{}, ErrInvalidJWTProviderConfig
		}
		algorithms[algorithm] = struct{}{}
		result.AllowedAlgorithms[index] = algorithm
	}

	var err error
	result.AllowedIssuers, err = normalizedUniqueStrings(cfg.AllowedIssuers)
	if err != nil {
		return JWTProviderConfig{}, err
	}
	result.AllowedAudiences, err = normalizedUniqueStrings(cfg.AllowedAudiences)
	if err != nil {
		return JWTProviderConfig{}, err
	}

	claimNames := make(map[string]struct{}, len(cfg.RequiredClaims))
	result.RequiredClaims = make([]JWTRequiredClaimConfig, len(cfg.RequiredClaims))
	for index, claim := range cfg.RequiredClaims {
		name := strings.TrimSpace(claim.Name)
		value := strings.TrimSpace(claim.Value)
		if name == "" || value == "" {
			return JWTProviderConfig{}, ErrInvalidJWTProviderConfig
		}
		if _, duplicate := claimNames[name]; duplicate {
			return JWTProviderConfig{}, ErrInvalidJWTProviderConfig
		}
		claimNames[name] = struct{}{}
		result.RequiredClaims[index] = JWTRequiredClaimConfig{Name: name, Value: value}
	}

	return result, nil
}

func normalizedUniqueStrings(values []string) ([]string, error) {
	result := make([]string, len(values))
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, ErrInvalidJWTProviderConfig
		}
		if _, duplicate := seen[value]; duplicate {
			return nil, ErrInvalidJWTProviderConfig
		}
		seen[value] = struct{}{}
		result[index] = value
	}
	return result, nil
}

func bearerToken(request AuthenticationRequest) (string, bool) {
	header, exists := request.Header("Authorization")
	if !exists {
		return "", false
	}
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func (provider *JWTProvider) resolveSigningKeys(
	ctx context.Context,
) ([]jwt.VerificationKey, []secretresolver.Secret, error) {
	keys := make([]jwt.VerificationKey, 0, len(provider.config.SigningKeys))
	secrets := make([]secretresolver.Secret, 0, len(provider.config.SigningKeys))
	for _, key := range provider.config.SigningKeys {
		secret, err := provider.resolver.Resolve(ctx, key.SecretRef)
		if err != nil {
			clearResolvedSecrets(secrets)
			if contextErr := ctx.Err(); contextErr != nil {
				return nil, nil, contextErr
			}
			return nil, nil, fmt.Errorf("%w: %w", ErrJWTProviderUnavailable, err)
		}
		secrets = append(secrets, secret)
		keys = append(keys, secret.Value)
	}
	return keys, secrets, nil
}

func clearResolvedSecrets(secrets []secretresolver.Secret) {
	for _, secret := range secrets {
		clear(secret.Value)
	}
}

func (provider *JWTProvider) algorithmNames() []string {
	result := make([]string, len(provider.config.AllowedAlgorithms))
	for index, algorithm := range provider.config.AllowedAlgorithms {
		result[index] = string(algorithm)
	}
	return result
}

func (provider *JWTProvider) claimsAllowed(claims jwt.MapClaims) bool {
	if len(provider.config.AllowedIssuers) > 0 {
		issuer, err := claims.GetIssuer()
		if err != nil || !contains(provider.config.AllowedIssuers, issuer) {
			return false
		}
	}
	if len(provider.config.AllowedAudiences) > 0 {
		audiences, err := claims.GetAudience()
		if err != nil || !intersects(provider.config.AllowedAudiences, audiences) {
			return false
		}
	}
	for _, required := range provider.config.RequiredClaims {
		actual, exists := scalarClaim(claims[required.Name])
		if !exists || actual != required.Value {
			return false
		}
	}
	return true
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func intersects(allowed []string, actual []string) bool {
	for _, value := range actual {
		if contains(allowed, value) {
			return true
		}
	}
	return false
}

func scalarClaims(claims jwt.MapClaims) map[string]string {
	result := make(map[string]string)
	for name, value := range claims {
		if normalized, ok := scalarClaim(value); ok {
			result[name] = normalized
		}
	}
	return result
}

func scalarClaim(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, true
	case bool:
		return strconv.FormatBool(typed), true
	case json.Number:
		return typed.String(), true
	case float64:
		return strconv.FormatFloat(typed, 'g', -1, 64), true
	case float32:
		return strconv.FormatFloat(float64(typed), 'g', -1, 32), true
	case int:
		return strconv.Itoa(typed), true
	case int64:
		return strconv.FormatInt(typed, 10), true
	case uint64:
		return strconv.FormatUint(typed, 10), true
	default:
		return "", false
	}
}

func (provider *JWTProvider) result(success bool, principal *Principal) AuthenticationResult {
	return AuthenticationResult{
		Success:      success,
		Principal:    principal,
		ProviderName: provider.config.Name,
		ProviderType: runtimeconfig.AuthenticationProviderJWT,
	}
}
