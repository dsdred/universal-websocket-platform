package authentication

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	var _ Registry = registry
}

func TestRegistryRegistersFactory(t *testing.T) {
	registry := NewRegistry()
	factory := newFakeFactory(runtimeconfig.AuthenticationProviderJWT, "fake-jwt")

	if err := registry.Register(factory); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
}

func TestRegistryRejectsDuplicateFactory(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register(newFakeFactory(runtimeconfig.AuthenticationProviderJWT, "first")); err != nil {
		t.Fatalf("Register(first) error = %v", err)
	}

	err := registry.Register(newFakeFactory(runtimeconfig.AuthenticationProviderJWT, "second"))
	if !errors.Is(err, ErrFactoryAlreadyRegistered) {
		t.Fatalf("Register(second) error = %v, want ErrFactoryAlreadyRegistered", err)
	}
}

func TestRegistryRejectsNilFactoryAndEmptyFactoryType(t *testing.T) {
	registry := NewRegistry()
	tests := []struct {
		name    string
		factory Factory
	}{
		{name: "nil Factory", factory: nil},
		{name: "empty Factory type", factory: FakeFactory{provider: fakeProvider{name: "fake"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := registry.Register(test.factory)
			if !errors.Is(err, ErrNilFactory) {
				t.Errorf("Register() error = %v, want ErrNilFactory", err)
			}
		})
	}
}

func TestRegistryFactoryNotFound(t *testing.T) {
	registry := NewRegistry()
	provider := runtimeconfig.AuthenticationProviderSnapshot{
		Name: "missing",
		Type: runtimeconfig.AuthenticationProviderType("not-registered"),
	}

	_, err := registry.Create(provider, resolverStub{})
	if !errors.Is(err, ErrFactoryNotFound) {
		t.Fatalf("Create() error = %v, want ErrFactoryNotFound", err)
	}
}

func TestRegistryCreatesProviderThroughFactory(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register(newFakeFactory(runtimeconfig.AuthenticationProviderJWT, "fake-jwt")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	configuration := runtimeconfig.AuthenticationProviderSnapshot{
		Name: "configured-jwt",
		Type: runtimeconfig.AuthenticationProviderJWT,
	}

	provider, err := registry.Create(configuration, resolverStub{})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if provider.Name() != "fake-jwt" || provider.Type() != runtimeconfig.AuthenticationProviderJWT {
		t.Errorf("Create() Provider = (%q, %q)", provider.Name(), provider.Type())
	}
}

func TestRegistryRejectsNilResolver(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register(newFakeFactory(runtimeconfig.AuthenticationProviderJWT, "fake-jwt")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	configuration := runtimeconfig.AuthenticationProviderSnapshot{
		Name: "configured-jwt",
		Type: runtimeconfig.AuthenticationProviderJWT,
	}

	_, err := registry.Create(configuration, nil)
	if !errors.Is(err, ErrNilSecretResolver) {
		t.Fatalf("Create() error = %v, want ErrNilSecretResolver", err)
	}
}

func TestRegistryConcurrentRegisterAndCreate(t *testing.T) {
	const operations = 64
	registry := NewRegistry()
	baseType := runtimeconfig.AuthenticationProviderType("concurrent-base")
	if err := registry.Register(newFakeFactory(baseType, "base-provider")); err != nil {
		t.Fatalf("Register(base) error = %v", err)
	}

	errorsChannel := make(chan error, operations*2)
	var waitGroup sync.WaitGroup
	for index := 0; index < operations; index++ {
		index := index
		waitGroup.Add(2)
		go func() {
			defer waitGroup.Done()
			providerType := runtimeconfig.AuthenticationProviderType(fmt.Sprintf("concurrent-%d", index))
			errorsChannel <- registry.Register(newFakeFactory(providerType, fmt.Sprintf("provider-%d", index)))
		}()
		go func() {
			defer waitGroup.Done()
			_, err := registry.Create(runtimeconfig.AuthenticationProviderSnapshot{
				Name: "base",
				Type: baseType,
			}, resolverStub{})
			errorsChannel <- err
		}()
	}

	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		if err != nil {
			t.Errorf("concurrent operation error = %v", err)
		}
	}

	for index := 0; index < operations; index++ {
		providerType := runtimeconfig.AuthenticationProviderType(fmt.Sprintf("concurrent-%d", index))
		if _, err := registry.Create(runtimeconfig.AuthenticationProviderSnapshot{
			Name: fmt.Sprintf("configured-%d", index),
			Type: providerType,
		}, resolverStub{}); err != nil {
			t.Errorf("Create(%q) after concurrent Register error = %v", providerType, err)
		}
	}
}

func TestRegistryConcurrentDuplicateRegistration(t *testing.T) {
	const operations = 64
	registry := NewRegistry()
	providerType := runtimeconfig.AuthenticationProviderType("duplicate")

	var successful atomic.Int32
	errorsChannel := make(chan error, operations)
	var waitGroup sync.WaitGroup
	for index := 0; index < operations; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := registry.Register(newFakeFactory(providerType, "fake"))
			if err == nil {
				successful.Add(1)
				return
			}
			errorsChannel <- err
		}()
	}

	waitGroup.Wait()
	close(errorsChannel)
	if successful.Load() != 1 {
		t.Errorf("successful registrations = %d, want 1", successful.Load())
	}
	for err := range errorsChannel {
		if !errors.Is(err, ErrFactoryAlreadyRegistered) {
			t.Errorf("concurrent Register() error = %v, want ErrFactoryAlreadyRegistered", err)
		}
	}
}

func TestRegistrySmokeScenario(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register(newFakeFactory(runtimeconfig.AuthenticationProviderJWT, "smoke-provider")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	provider, err := registry.Create(runtimeconfig.AuthenticationProviderSnapshot{
		Name: "configured-jwt",
		Type: runtimeconfig.AuthenticationProviderJWT,
	}, resolverStub{})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	t.Logf("provider created: name=%s", provider.Name())
}

func newFakeFactory(providerType runtimeconfig.AuthenticationProviderType, providerName string) FakeFactory {
	return FakeFactory{
		providerType: providerType,
		provider: fakeProvider{
			name:         providerName,
			providerType: providerType,
		},
	}
}
