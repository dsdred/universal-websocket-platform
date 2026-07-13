package authentication

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

type FakeFactory struct {
	providerType runtimeconfig.AuthenticationProviderType
	provider     Provider
}

func (factory FakeFactory) Type() runtimeconfig.AuthenticationProviderType {
	return factory.providerType
}

func (factory FakeFactory) Create(
	configuration runtimeconfig.AuthenticationProviderSnapshot,
	resolver secretresolver.Resolver,
) (Provider, error) {
	if resolver == nil {
		return nil, ErrNilSecretResolver
	}
	if strings.TrimSpace(configuration.Name) == "" || configuration.Type == "" {
		return nil, ErrInvalidProvider
	}
	if configuration.Type != factory.providerType || factory.provider == nil {
		return nil, ErrUnsupportedProviderType
	}
	return factory.provider, nil
}

type fakeProvider struct {
	name         string
	providerType runtimeconfig.AuthenticationProviderType
}

func (provider fakeProvider) Name() string {
	return provider.name
}

func (provider fakeProvider) Type() runtimeconfig.AuthenticationProviderType {
	return provider.providerType
}

func (provider fakeProvider) Authenticate(
	context.Context,
	AuthenticationRequest,
) (AuthenticationResult, error) {
	return AuthenticationResult{
		Success: true,
		Principal: &Principal{
			ID:                 "test-principal",
			AuthenticationType: provider.providerType,
			Authenticated:      true,
		},
		ProviderName: provider.name,
		ProviderType: provider.providerType,
	}, nil
}

type resolverStub struct{}

func (resolverStub) Resolve(context.Context, string) (secretresolver.Secret, error) {
	return secretresolver.Secret{}, secretresolver.ErrSecretNotFound
}

func TestFactoryInterfaceImplementation(t *testing.T) {
	provider := fakeProvider{name: "fake-jwt", providerType: runtimeconfig.AuthenticationProviderJWT}
	factory := FakeFactory{providerType: runtimeconfig.AuthenticationProviderJWT, provider: provider}

	var _ Factory = factory
	var _ Provider = provider
}

func TestFakeFactoryCreatesProvider(t *testing.T) {
	want := fakeProvider{name: "fake-jwt", providerType: runtimeconfig.AuthenticationProviderJWT}
	factory := FakeFactory{providerType: runtimeconfig.AuthenticationProviderJWT, provider: want}
	configuration := runtimeconfig.AuthenticationProviderSnapshot{
		Name: "configured-jwt",
		Type: runtimeconfig.AuthenticationProviderJWT,
		JWT:  &runtimeconfig.JWTSnapshot{},
	}

	created, err := factory.Create(configuration, resolverStub{})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Name() != want.Name() || created.Type() != want.Type() {
		t.Errorf("Create() Provider = (%q, %q), want (%q, %q)", created.Name(), created.Type(), want.Name(), want.Type())
	}

	result, err := created.Authenticate(context.Background(), AuthenticationRequest{})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if !result.Success || result.Principal == nil || !result.Principal.Authenticated {
		t.Errorf("Authenticate() result = %#v", result)
	}
}

func TestFakeFactoryRejectsNilResolver(t *testing.T) {
	factory := FakeFactory{
		providerType: runtimeconfig.AuthenticationProviderJWT,
		provider:     fakeProvider{name: "fake-jwt", providerType: runtimeconfig.AuthenticationProviderJWT},
	}
	configuration := runtimeconfig.AuthenticationProviderSnapshot{Name: "jwt", Type: runtimeconfig.AuthenticationProviderJWT}

	_, err := factory.Create(configuration, nil)
	if !errors.Is(err, ErrNilSecretResolver) {
		t.Fatalf("Create() error = %v, want ErrNilSecretResolver", err)
	}
}

func TestFakeFactoryRejectsUnsupportedProviderType(t *testing.T) {
	factory := FakeFactory{
		providerType: runtimeconfig.AuthenticationProviderJWT,
		provider:     fakeProvider{name: "fake-jwt", providerType: runtimeconfig.AuthenticationProviderJWT},
	}
	configuration := runtimeconfig.AuthenticationProviderSnapshot{Name: "basic", Type: runtimeconfig.AuthenticationProviderBasic}

	_, err := factory.Create(configuration, resolverStub{})
	if !errors.Is(err, ErrUnsupportedProviderType) {
		t.Fatalf("Create() error = %v, want ErrUnsupportedProviderType", err)
	}
}

func TestFakeFactoryRejectsInvalidProvider(t *testing.T) {
	factory := FakeFactory{
		providerType: runtimeconfig.AuthenticationProviderJWT,
		provider:     fakeProvider{name: "fake-jwt", providerType: runtimeconfig.AuthenticationProviderJWT},
	}
	tests := []struct {
		name          string
		configuration runtimeconfig.AuthenticationProviderSnapshot
	}{
		{name: "empty name", configuration: runtimeconfig.AuthenticationProviderSnapshot{Type: runtimeconfig.AuthenticationProviderJWT}},
		{name: "whitespace name", configuration: runtimeconfig.AuthenticationProviderSnapshot{Name: "   ", Type: runtimeconfig.AuthenticationProviderJWT}},
		{name: "empty type", configuration: runtimeconfig.AuthenticationProviderSnapshot{Name: "jwt"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := factory.Create(test.configuration, resolverStub{})
			if !errors.Is(err, ErrInvalidProvider) {
				t.Errorf("Create() error = %v, want ErrInvalidProvider", err)
			}
		})
	}
}

func TestFakeFactoryWithoutProviderReturnsUnsupported(t *testing.T) {
	factory := FakeFactory{providerType: runtimeconfig.AuthenticationProviderJWT}
	configuration := runtimeconfig.AuthenticationProviderSnapshot{Name: "jwt", Type: runtimeconfig.AuthenticationProviderJWT}

	_, err := factory.Create(configuration, resolverStub{})
	if !errors.Is(err, ErrUnsupportedProviderType) {
		t.Fatalf("Create() error = %v, want ErrUnsupportedProviderType", err)
	}
}
