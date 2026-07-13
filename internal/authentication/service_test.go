package authentication

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

func TestDefaultServiceImplementsService(t *testing.T) {
	var _ Service = (*DefaultService)(nil)
}

func TestServiceAuthenticateWithoutProviders(t *testing.T) {
	service, err := NewService(nil)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	result, err := service.Authenticate(context.Background(), AuthenticationRequest{})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if result.Success {
		t.Fatal("Authenticate() Success = true, want false")
	}
}

func TestServiceAuthenticateWithOneProvider(t *testing.T) {
	provider := newServiceProvider("first", true, nil)
	service, err := NewService([]Provider{provider})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	result, err := service.Authenticate(context.Background(), AuthenticationRequest{})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if !result.Success || result.ProviderName != "first" {
		t.Fatalf("Authenticate() = %+v, want success from first", result)
	}
	if calls := provider.calls.Load(); calls != 1 {
		t.Fatalf("Provider calls = %d, want 1", calls)
	}
}

func TestServiceAuthenticateStopsAfterFirstSuccess(t *testing.T) {
	first := newServiceProvider("first", true, nil)
	second := newServiceProvider("second", true, nil)
	service, err := NewService([]Provider{first, second})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	result, err := service.Authenticate(context.Background(), AuthenticationRequest{})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if result.ProviderName != "first" {
		t.Fatalf("ProviderName = %q, want first", result.ProviderName)
	}
	if calls := second.calls.Load(); calls != 0 {
		t.Fatalf("Second Provider calls = %d, want 0", calls)
	}
}

func TestServiceAuthenticateUsesSecondSuccessfulProvider(t *testing.T) {
	first := newServiceProvider("first", false, nil)
	second := newServiceProvider("second", true, nil)
	third := newServiceProvider("third", true, nil)
	service, err := NewService([]Provider{first, second, third})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	result, err := service.Authenticate(context.Background(), AuthenticationRequest{})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if result.ProviderName != "second" {
		t.Fatalf("ProviderName = %q, want second", result.ProviderName)
	}
	if first.calls.Load() != 1 || second.calls.Load() != 1 || third.calls.Load() != 0 {
		t.Fatalf("Provider calls = (%d, %d, %d), want (1, 1, 0)", first.calls.Load(), second.calls.Load(), third.calls.Load())
	}
}

func TestServiceAuthenticateReturnsProviderError(t *testing.T) {
	wantErr := errors.New("Provider failed")
	first := newServiceProvider("first", false, wantErr)
	second := newServiceProvider("second", true, nil)
	service, err := NewService([]Provider{first, second})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.Authenticate(context.Background(), AuthenticationRequest{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Authenticate() error = %v, want errors.Is(_, %v)", err, wantErr)
	}
	if calls := second.calls.Load(); calls != 0 {
		t.Fatalf("Second Provider calls = %d, want 0", calls)
	}
}

func TestNewServiceCopiesProviderSlice(t *testing.T) {
	original := newServiceProvider("original", true, nil)
	replacement := newServiceProvider("replacement", true, nil)
	providers := []Provider{original}
	service, err := NewService(providers)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	providers[0] = replacement

	result, err := service.Authenticate(context.Background(), AuthenticationRequest{})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if result.ProviderName != "original" {
		t.Fatalf("ProviderName = %q, want original", result.ProviderName)
	}
}

func TestNewServiceRejectsNilProvider(t *testing.T) {
	_, err := NewService([]Provider{nil})
	if !errors.Is(err, ErrInvalidProvider) {
		t.Fatalf("NewService() error = %v, want errors.Is(_, ErrInvalidProvider)", err)
	}
}

func TestServiceAuthenticateConcurrently(t *testing.T) {
	provider := newServiceProvider("concurrent", true, nil)
	service, err := NewService([]Provider{provider})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	const goroutines = 64
	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, goroutines)
	for range goroutines {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			result, authenticateErr := service.Authenticate(context.Background(), AuthenticationRequest{})
			if authenticateErr != nil {
				errorsChannel <- authenticateErr
				return
			}
			if !result.Success {
				errorsChannel <- errors.New("Authentication was not successful")
			}
		}()
	}
	waitGroup.Wait()
	close(errorsChannel)

	for authenticateErr := range errorsChannel {
		t.Errorf("Authenticate() error = %v", authenticateErr)
	}
	if calls := provider.calls.Load(); calls != goroutines {
		t.Fatalf("Provider calls = %d, want %d", calls, goroutines)
	}
}

func TestServiceWithAPIKeyProvider(t *testing.T) {
	resolver, err := secretresolver.NewMemory(map[string][]byte{
		"env/UWP_API_KEY": []byte("test-secret"),
	})
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	provider, err := NewAPIKeyProvider(APIKeyProviderConfig{
		Name:      "internal-api-key",
		Header:    "X-API-Key",
		SecretRef: "env/UWP_API_KEY",
	}, resolver)
	if err != nil {
		t.Fatalf("NewAPIKeyProvider() error = %v", err)
	}
	service, err := NewService([]Provider{provider})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	result, err := service.Authenticate(context.Background(), AuthenticationRequest{
		Headers: map[string][]string{"X-API-Key": []string{"test-secret"}},
	})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if !result.Success || result.Principal == nil || !result.Principal.Authenticated {
		t.Fatalf("Authenticate() = %+v, want authenticated Principal", result)
	}
	t.Logf("authenticated Principal %q through Provider %q", result.Principal.Name, result.ProviderName)
}

type serviceProvider struct {
	name   string
	result AuthenticationResult
	err    error
	calls  atomic.Uint64
}

func newServiceProvider(name string, success bool, err error) *serviceProvider {
	return &serviceProvider{
		name: name,
		result: AuthenticationResult{
			Success:      success,
			ProviderName: name,
			ProviderType: runtimeconfig.AuthenticationProviderAPIKey,
		},
		err: err,
	}
}

func (provider *serviceProvider) Name() string {
	return provider.name
}

func (provider *serviceProvider) Type() runtimeconfig.AuthenticationProviderType {
	return runtimeconfig.AuthenticationProviderAPIKey
}

func (provider *serviceProvider) Authenticate(context.Context, AuthenticationRequest) (AuthenticationResult, error) {
	provider.calls.Add(1)
	return provider.result, provider.err
}
