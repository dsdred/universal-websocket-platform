package authentication

import (
	"context"
	"crypto/subtle"
	"fmt"
	"strings"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

// APIKeyProvider authenticates a client-supplied header against a resolved Secret.
type APIKeyProvider struct {
	name      string
	header    string
	secretRef string
	resolver  secretresolver.Resolver
}

// APIKeyProviderConfig contains runtime configuration required by APIKeyProvider.
type APIKeyProviderConfig struct {
	Name      string
	Header    string
	SecretRef string
}

// NewAPIKeyProvider creates an API Key Provider without resolving its Secret.
func NewAPIKeyProvider(
	cfg APIKeyProviderConfig,
	resolver secretresolver.Resolver,
) (*APIKeyProvider, error) {
	if resolver == nil {
		return nil, ErrNilSecretResolver
	}

	name := strings.TrimSpace(cfg.Name)
	header := strings.TrimSpace(cfg.Header)
	secretRef := strings.TrimSpace(cfg.SecretRef)
	if name == "" || header == "" || secretRef == "" {
		return nil, ErrInvalidProvider
	}
	if err := secretresolver.ValidateReference(secretRef); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidProvider, err)
	}

	return &APIKeyProvider{
		name:      name,
		header:    header,
		secretRef: secretRef,
		resolver:  resolver,
	}, nil
}

// Name returns the configured Provider name.
func (provider *APIKeyProvider) Name() string {
	return provider.name
}

// Type returns the API Key Provider type.
func (provider *APIKeyProvider) Type() runtimeconfig.AuthenticationProviderType {
	return runtimeconfig.AuthenticationProviderAPIKey
}

// Authenticate compares the configured header with the Secret resolved for this attempt.
func (provider *APIKeyProvider) Authenticate(
	ctx context.Context,
	request AuthenticationRequest,
) (AuthenticationResult, error) {
	if err := ctx.Err(); err != nil {
		return AuthenticationResult{}, err
	}

	headerValue, exists := request.Header(provider.header)
	if !exists {
		return provider.result(false, nil), nil
	}

	secret, err := provider.resolver.Resolve(ctx, provider.secretRef)
	if err != nil {
		return provider.result(false, nil), fmt.Errorf("resolve API Key secret: %w", err)
	}
	defer clear(secret.Value)

	if subtle.ConstantTimeCompare([]byte(headerValue), secret.Value) != 1 {
		return provider.result(false, nil), nil
	}

	principal := &Principal{
		Name:                   provider.name,
		AuthenticationType:     runtimeconfig.AuthenticationProviderAPIKey,
		AuthenticationProvider: provider.name,
		AuthenticationMethod:   runtimeconfig.AuthenticationProviderAPIKey,
		Claims:                 make(map[string]string),
		Authenticated:          true,
	}
	return provider.result(true, principal), nil
}

func (provider *APIKeyProvider) result(success bool, principal *Principal) AuthenticationResult {
	return AuthenticationResult{
		Success:      success,
		Principal:    principal,
		ProviderName: provider.name,
		ProviderType: runtimeconfig.AuthenticationProviderAPIKey,
	}
}
