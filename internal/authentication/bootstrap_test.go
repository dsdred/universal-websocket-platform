package authentication

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

func TestDefaultBootstrapImplementsBootstrap(t *testing.T) {
	var _ Bootstrap = (*DefaultBootstrap)(nil)
}

func TestNewBootstrapValidatesDependencies(t *testing.T) {
	registry := &bootstrapRegistry{}
	resolver := bootstrapResolver{}

	if _, err := NewBootstrap(nil, resolver); !errors.Is(err, ErrNilRegistry) {
		t.Fatalf("NewBootstrap(nil registry) error = %v, want ErrNilRegistry", err)
	}
	if _, err := NewBootstrap(registry, nil); !errors.Is(err, ErrNilSecretResolver) {
		t.Fatalf("NewBootstrap(nil resolver) error = %v, want ErrNilSecretResolver", err)
	}
}

func TestBootstrapBuildDisabledAuthentication(t *testing.T) {
	registry := &bootstrapRegistry{
		create: func(runtimeconfig.AuthenticationProviderSnapshot, secretresolver.Resolver) (Provider, error) {
			t.Fatal("Registry.Create() called for disabled Authentication")
			return nil, nil
		},
	}
	bootstrap := mustBootstrap(t, registry, bootstrapResolver{})

	service, err := bootstrap.Build(runtimeconfig.AuthenticationSnapshot{
		Enabled: false,
		Providers: []runtimeconfig.AuthenticationProviderSnapshot{
			{Name: "ignored", Type: runtimeconfig.AuthenticationProviderAPIKey},
		},
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	result, err := service.Authenticate(context.Background(), AuthenticationRequest{})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if result.Success {
		t.Fatal("Authenticate() Success = true, want false")
	}
}

func TestBootstrapBuildEnabledAuthenticationWithoutProviders(t *testing.T) {
	registry := &bootstrapRegistry{
		create: func(runtimeconfig.AuthenticationProviderSnapshot, secretresolver.Resolver) (Provider, error) {
			t.Fatal("Registry.Create() called without Provider Snapshots")
			return nil, nil
		},
	}
	bootstrap := mustBootstrap(t, registry, bootstrapResolver{})

	service, err := bootstrap.Build(runtimeconfig.AuthenticationSnapshot{Enabled: true})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	result, err := service.Authenticate(context.Background(), AuthenticationRequest{})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if result.Success {
		t.Fatal("Authenticate() Success = true, want false")
	}
}

func TestBootstrapBuildEnabledAuthenticationWithOneProvider(t *testing.T) {
	provider := &bootstrapProvider{name: "only", success: true}
	registry := &bootstrapRegistry{
		create: func(snapshot runtimeconfig.AuthenticationProviderSnapshot, _ secretresolver.Resolver) (Provider, error) {
			if snapshot.Name != "only" {
				t.Fatalf("Provider name = %q, want only", snapshot.Name)
			}
			return provider, nil
		},
	}
	bootstrap := mustBootstrap(t, registry, bootstrapResolver{})

	service, err := bootstrap.Build(runtimeconfig.AuthenticationSnapshot{
		Enabled: true,
		Providers: []runtimeconfig.AuthenticationProviderSnapshot{
			{Name: "only", Type: runtimeconfig.AuthenticationProviderAPIKey},
		},
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	result, err := service.Authenticate(context.Background(), AuthenticationRequest{})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if !result.Success || result.ProviderName != "only" {
		t.Fatalf("Authenticate() = %+v, want success from only", result)
	}
}

func TestBootstrapBuildPreservesProviderOrder(t *testing.T) {
	var mutex sync.Mutex
	authenticationOrder := make([]string, 0, 2)
	providers := map[string]*bootstrapProvider{
		"first":  {name: "first", success: false, calls: &authenticationOrder, mutex: &mutex},
		"second": {name: "second", success: true, calls: &authenticationOrder, mutex: &mutex},
	}
	creationOrder := make([]string, 0, 2)
	registry := &bootstrapRegistry{
		create: func(snapshot runtimeconfig.AuthenticationProviderSnapshot, _ secretresolver.Resolver) (Provider, error) {
			creationOrder = append(creationOrder, snapshot.Name)
			return providers[snapshot.Name], nil
		},
	}
	bootstrap := mustBootstrap(t, registry, bootstrapResolver{})

	service, err := bootstrap.Build(runtimeconfig.AuthenticationSnapshot{
		Enabled: true,
		Providers: []runtimeconfig.AuthenticationProviderSnapshot{
			{Name: "first", Type: runtimeconfig.AuthenticationProviderJWT},
			{Name: "second", Type: runtimeconfig.AuthenticationProviderAPIKey},
		},
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	result, err := service.Authenticate(context.Background(), AuthenticationRequest{})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if result.ProviderName != "second" {
		t.Fatalf("ProviderName = %q, want second", result.ProviderName)
	}
	if !reflect.DeepEqual(creationOrder, []string{"first", "second"}) {
		t.Fatalf("creation order = %v, want [first second]", creationOrder)
	}
	if !reflect.DeepEqual(authenticationOrder, []string{"first", "second"}) {
		t.Fatalf("Authentication order = %v, want [first second]", authenticationOrder)
	}
}

func TestBootstrapBuildReturnsRegistryError(t *testing.T) {
	wantErr := errors.New("Registry failed")
	registry := &bootstrapRegistry{
		create: func(runtimeconfig.AuthenticationProviderSnapshot, secretresolver.Resolver) (Provider, error) {
			return nil, wantErr
		},
	}
	bootstrap := mustBootstrap(t, registry, bootstrapResolver{})

	_, err := bootstrap.Build(enabledAuthenticationSnapshot())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Build() error = %v, want errors.Is(_, %v)", err, wantErr)
	}
}

func TestBootstrapBuildReturnsFactoryError(t *testing.T) {
	wantErr := errors.New("Factory failed")
	registry := NewRegistry()
	if err := registry.Register(bootstrapFactory{
		providerType: runtimeconfig.AuthenticationProviderAPIKey,
		create: func(runtimeconfig.AuthenticationProviderSnapshot, secretresolver.Resolver) (Provider, error) {
			return nil, wantErr
		},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	bootstrap := mustBootstrap(t, registry, bootstrapResolver{})

	_, err := bootstrap.Build(enabledAuthenticationSnapshot())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Build() error = %v, want errors.Is(_, %v)", err, wantErr)
	}
}

func TestBootstrapBuildReturnsServiceError(t *testing.T) {
	registry := &bootstrapRegistry{
		create: func(runtimeconfig.AuthenticationProviderSnapshot, secretresolver.Resolver) (Provider, error) {
			return nil, nil
		},
	}
	bootstrap := mustBootstrap(t, registry, bootstrapResolver{})

	_, err := bootstrap.Build(enabledAuthenticationSnapshot())
	if !errors.Is(err, ErrInvalidProvider) {
		t.Fatalf("Build() error = %v, want errors.Is(_, ErrInvalidProvider)", err)
	}
}

func TestBootstrapBuildDoesNotModifySnapshot(t *testing.T) {
	snapshot := runtimeconfig.AuthenticationSnapshot{
		Enabled: true,
		Providers: []runtimeconfig.AuthenticationProviderSnapshot{
			{
				Name:     "api-key",
				Type:     runtimeconfig.AuthenticationProviderAPIKey,
				Enabled:  true,
				Priority: 10,
				APIKey:   &runtimeconfig.APIKeySnapshot{Header: "X-API-Key", SecretRef: "secrets/api-key"},
			},
		},
	}
	want := runtimeconfig.AuthenticationSnapshot{
		Enabled: true,
		Providers: []runtimeconfig.AuthenticationProviderSnapshot{
			{
				Name:     "api-key",
				Type:     runtimeconfig.AuthenticationProviderAPIKey,
				Enabled:  true,
				Priority: 10,
				APIKey:   &runtimeconfig.APIKeySnapshot{Header: "X-API-Key", SecretRef: "secrets/api-key"},
			},
		},
	}
	registry := &bootstrapRegistry{
		create: func(runtimeconfig.AuthenticationProviderSnapshot, secretresolver.Resolver) (Provider, error) {
			return &bootstrapProvider{name: "api-key", success: true}, nil
		},
	}
	bootstrap := mustBootstrap(t, registry, bootstrapResolver{})

	if _, err := bootstrap.Build(snapshot); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !reflect.DeepEqual(snapshot, want) {
		t.Fatalf("Snapshot changed: got %+v, want %+v", snapshot, want)
	}
}

func TestBootstrapAPIKeySmokeScenario(t *testing.T) {
	resolver, err := secretresolver.NewMemory(map[string][]byte{
		"secrets/api-key/internal": []byte("smoke-secret"),
	})
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	registry := NewRegistry()
	if err := registry.Register(bootstrapFactory{
		providerType: runtimeconfig.AuthenticationProviderAPIKey,
		create: func(snapshot runtimeconfig.AuthenticationProviderSnapshot, resolver secretresolver.Resolver) (Provider, error) {
			return NewAPIKeyProvider(APIKeyProviderConfig{
				Name:      snapshot.Name,
				Header:    snapshot.APIKey.Header,
				SecretRef: snapshot.APIKey.SecretRef,
			}, resolver)
		},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	bootstrap := mustBootstrap(t, registry, resolver)
	service, err := bootstrap.Build(runtimeconfig.AuthenticationSnapshot{
		Enabled: true,
		Providers: []runtimeconfig.AuthenticationProviderSnapshot{
			{
				Name:    "internal-api-key",
				Type:    runtimeconfig.AuthenticationProviderAPIKey,
				Enabled: true,
				APIKey: &runtimeconfig.APIKeySnapshot{
					Header:    "X-API-Key",
					SecretRef: "secrets/api-key/internal",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	result, err := service.Authenticate(context.Background(), AuthenticationRequest{
		Headers: map[string][]string{"X-API-Key": {"smoke-secret"}},
	})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if !result.Success || result.Principal == nil || !result.Principal.Authenticated {
		t.Fatalf("Authenticate() = %+v, want authenticated Principal", result)
	}
	t.Logf("Snapshot -> Bootstrap -> Service -> Principal %q", result.Principal.Name)
}

type bootstrapRegistry struct {
	create func(runtimeconfig.AuthenticationProviderSnapshot, secretresolver.Resolver) (Provider, error)
}

func (*bootstrapRegistry) Register(Factory) error {
	return nil
}

func (registry *bootstrapRegistry) Create(
	snapshot runtimeconfig.AuthenticationProviderSnapshot,
	resolver secretresolver.Resolver,
) (Provider, error) {
	return registry.create(snapshot, resolver)
}

type bootstrapFactory struct {
	providerType runtimeconfig.AuthenticationProviderType
	create       func(runtimeconfig.AuthenticationProviderSnapshot, secretresolver.Resolver) (Provider, error)
}

func (factory bootstrapFactory) Type() runtimeconfig.AuthenticationProviderType {
	return factory.providerType
}

func (factory bootstrapFactory) Create(
	snapshot runtimeconfig.AuthenticationProviderSnapshot,
	resolver secretresolver.Resolver,
) (Provider, error) {
	return factory.create(snapshot, resolver)
}

type bootstrapProvider struct {
	name    string
	success bool
	calls   *[]string
	mutex   *sync.Mutex
}

func (provider *bootstrapProvider) Name() string {
	return provider.name
}

func (*bootstrapProvider) Type() runtimeconfig.AuthenticationProviderType {
	return runtimeconfig.AuthenticationProviderAPIKey
}

func (provider *bootstrapProvider) Authenticate(
	context.Context,
	AuthenticationRequest,
) (AuthenticationResult, error) {
	if provider.calls != nil {
		provider.mutex.Lock()
		*provider.calls = append(*provider.calls, provider.name)
		provider.mutex.Unlock()
	}
	return AuthenticationResult{
		Success:      provider.success,
		ProviderName: provider.name,
		ProviderType: runtimeconfig.AuthenticationProviderAPIKey,
	}, nil
}

type bootstrapResolver struct{}

func (bootstrapResolver) Resolve(context.Context, string) (secretresolver.Secret, error) {
	return secretresolver.Secret{}, secretresolver.ErrSecretNotFound
}

func mustBootstrap(t *testing.T, registry Registry, resolver secretresolver.Resolver) *DefaultBootstrap {
	t.Helper()
	bootstrap, err := NewBootstrap(registry, resolver)
	if err != nil {
		t.Fatalf("NewBootstrap() error = %v", err)
	}
	return bootstrap
}

func enabledAuthenticationSnapshot() runtimeconfig.AuthenticationSnapshot {
	return runtimeconfig.AuthenticationSnapshot{
		Enabled: true,
		Providers: []runtimeconfig.AuthenticationProviderSnapshot{
			{Name: "api-key", Type: runtimeconfig.AuthenticationProviderAPIKey},
		},
	}
}
