package authentication

import (
	"context"
	"fmt"
)

// Service authenticates transport-neutral requests through configured Providers.
type Service interface {
	Authenticate(ctx context.Context, request AuthenticationRequest) (AuthenticationResult, error)
}

// DefaultService evaluates an immutable, ordered set of Providers.
type DefaultService struct {
	providers []Provider
}

// NewService creates a Service with an independent copy of providers.
func NewService(providers []Provider) (*DefaultService, error) {
	providerCopy := make([]Provider, len(providers))
	for index, provider := range providers {
		if provider == nil {
			return nil, fmt.Errorf("%w: Provider at index %d is nil", ErrInvalidProvider, index)
		}
		providerCopy[index] = provider
	}

	return &DefaultService{providers: providerCopy}, nil
}

// Authenticate returns the first successful result or stops at the first Provider error.
func (service *DefaultService) Authenticate(
	ctx context.Context,
	request AuthenticationRequest,
) (AuthenticationResult, error) {
	for _, provider := range service.providers {
		result, err := provider.Authenticate(ctx, request)
		if err != nil {
			return result, err
		}
		if result.Success {
			return result, nil
		}
	}

	return AuthenticationResult{}, nil
}
