package connection

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

func TestNewAuthenticationDispatcher(t *testing.T) {
	service := &authenticationServiceStub{}
	next := &authenticatedDispatcherStub{}
	dispatcher, err := NewAuthenticationDispatcher(service, next)
	if err != nil {
		t.Fatalf("NewAuthenticationDispatcher() error = %v", err)
	}
	if dispatcher == nil {
		t.Fatal("NewAuthenticationDispatcher() = nil")
	}
	var _ Dispatcher = dispatcher
}

func TestNewAuthenticationDispatcherRejectsNilDependencies(t *testing.T) {
	service := &authenticationServiceStub{}
	next := &authenticatedDispatcherStub{}

	if dispatcher, err := NewAuthenticationDispatcher(nil, next); dispatcher != nil || !errors.Is(err, ErrNilAuthenticationService) {
		t.Fatalf("NewAuthenticationDispatcher(nil service) = (%v, %v)", dispatcher, err)
	}
	if dispatcher, err := NewAuthenticationDispatcher(service, nil); dispatcher != nil || !errors.Is(err, ErrNilNextDispatcher) {
		t.Fatalf("NewAuthenticationDispatcher(nil next) = (%v, %v)", dispatcher, err)
	}
}

func TestAuthenticationDispatcherRejectsCanceledContext(t *testing.T) {
	service := &authenticationServiceStub{}
	next := &authenticatedDispatcherStub{}
	dispatcher := mustAuthenticationDispatcher(t, service, next)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := httptest.NewRequest("GET", "http://example.test/ws", nil)

	err := dispatcher.Dispatch(NewContext(ctx, nil, request))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Dispatch() error = %v, want context.Canceled", err)
	}
	if service.calls.Load() != 0 || next.callCount() != 0 {
		t.Fatalf("calls = (service %d, next %d), want zero", service.calls.Load(), next.callCount())
	}
}

func TestAuthenticationDispatcherCopiesHandshakeData(t *testing.T) {
	request := httptest.NewRequest("GET", "http://example.test/ws?tenant=alpha", nil)
	request.Header["X-API-Key"] = []string{"credential"}
	request.RemoteAddr = "192.0.2.1:4321"
	service := &authenticationServiceStub{}
	service.authenticate = func(_ context.Context, authenticationRequest authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
		request.Header["X-API-Key"][0] = "changed"
		request.URL.RawQuery = "tenant=changed"
		if authenticationRequest.Headers["X-API-Key"][0] != "credential" {
			return authentication.AuthenticationResult{}, errors.New("Headers were not copied")
		}
		if authenticationRequest.Query["tenant"][0] != "alpha" {
			return authentication.AuthenticationResult{}, errors.New("Query was not copied")
		}
		if authenticationRequest.RemoteAddress != "192.0.2.1:4321" || authenticationRequest.Transport != "websocket" {
			return authentication.AuthenticationResult{}, errors.New("transport metadata mismatch")
		}
		return successfulAuthentication("api-key"), nil
	}
	next := &authenticatedDispatcherStub{}
	dispatcher := mustAuthenticationDispatcher(t, service, next)

	if err := dispatcher.Dispatch(NewContext(request.Context(), nil, request)); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	if service.context.Load() != request.Context() {
		t.Fatal("Authenticate() did not receive the ConnectionContext Context")
	}
	if next.callCount() != 1 {
		t.Fatalf("next calls = %d, want 1", next.callCount())
	}
}

func TestAuthenticationDispatcherWithProductionAPIKeyProvider(t *testing.T) {
	service := newAPIKeyService(t, secretResolverFromValue(t, "env/UWP_API_KEY", "correct-key"))
	next := &authenticatedDispatcherStub{}
	dispatcher := mustAuthenticationDispatcher(t, service, next)
	request := httptest.NewRequest("GET", "http://example.test/ws", nil)
	request.Header.Set("X-API-Key", "correct-key")

	if err := dispatcher.Dispatch(NewContext(request.Context(), nil, request)); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	if next.callCount() != 1 {
		t.Fatalf("next calls = %d, want 1", next.callCount())
	}
	principal := next.lastContext().Principal()
	if !principal.Authenticated || principal.AuthenticationProvider != "runtime-api-key" {
		t.Fatalf("Principal = %+v", principal)
	}
}

func TestAuthenticationDispatcherWithProductionJWTProvider(t *testing.T) {
	const signingSecret = "jwt-signing-secret"
	resolver := secretResolverFromValue(t, "env/UWP_JWT_KEY", signingSecret)
	provider, err := authentication.NewJWTProvider(authentication.JWTProviderConfig{
		Name: "runtime-jwt",
		SigningKeys: []authentication.JWTSigningKeyConfig{
			{Name: "primary", SecretRef: "env/UWP_JWT_KEY"},
		},
		AllowedAlgorithms: []runtimeconfig.JWTAlgorithm{runtimeconfig.HS256},
		ClockSkewSeconds:  60,
	}, resolver)
	if err != nil {
		t.Fatalf("NewJWTProvider() error = %v", err)
	}
	service, err := authentication.NewService([]authentication.Provider{provider})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "administrator"})
	tokenString, err := token.SignedString([]byte(signingSecret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	request := httptest.NewRequest("GET", "http://example.test/ws", nil)
	request.Header.Set("Authorization", "Bearer "+tokenString)
	next := &authenticatedDispatcherStub{}
	dispatcher := mustAuthenticationDispatcher(t, service, next)

	if err := dispatcher.Dispatch(NewContext(request.Context(), nil, request)); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	principal := next.lastContext().Principal()
	if next.callCount() != 1 || principal.Name != "administrator" {
		t.Fatalf("next calls = %d, Principal = %+v", next.callCount(), principal)
	}
}

func TestAuthenticationDispatcherRejectsMissingAndInvalidCredentials(t *testing.T) {
	service := newAPIKeyService(t, secretResolverFromValue(t, "env/UWP_API_KEY", "correct-key"))
	for _, test := range []struct {
		name  string
		value string
	}{
		{name: "missing"},
		{name: "invalid", value: "wrong-key"},
	} {
		t.Run(test.name, func(t *testing.T) {
			next := &authenticatedDispatcherStub{}
			dispatcher := mustAuthenticationDispatcher(t, service, next)
			request := httptest.NewRequest("GET", "http://example.test/ws", nil)
			if test.value != "" {
				request.Header.Set("X-API-Key", test.value)
			}

			if err := dispatcher.Dispatch(NewContext(request.Context(), nil, request)); err != nil {
				t.Fatalf("Dispatch() error = %v", err)
			}
			if next.callCount() != 0 {
				t.Fatalf("next calls = %d, want 0", next.callCount())
			}
		})
	}
}

func TestAuthenticationDispatcherReturnsSafeServiceError(t *testing.T) {
	const credential = "do-not-expose-this-api-key"
	wantErr := errors.New("resolver unavailable: " + credential)
	service := &authenticationServiceStub{
		authenticate: func(context.Context, authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
			return authentication.AuthenticationResult{}, wantErr
		},
	}
	next := &authenticatedDispatcherStub{}
	dispatcher := mustAuthenticationDispatcher(t, service, next)
	request := httptest.NewRequest("GET", "http://example.test/ws", nil)
	request.Header.Set("X-API-Key", credential)

	err := dispatcher.Dispatch(NewContext(request.Context(), nil, request))
	if !errors.Is(err, wantErr) {
		t.Fatalf("Dispatch() error = %v, want errors.Is(_, service error)", err)
	}
	if strings.Contains(err.Error(), credential) {
		t.Fatalf("Dispatch() error exposes credential: %q", err)
	}
	if next.callCount() != 0 {
		t.Fatalf("next calls = %d, want 0", next.callCount())
	}
}

func TestAuthenticationDispatcherRejectsInvalidSuccessfulResult(t *testing.T) {
	service := &authenticationServiceStub{
		authenticate: func(context.Context, authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
			return authentication.AuthenticationResult{Success: true}, nil
		},
	}
	next := &authenticatedDispatcherStub{}
	dispatcher := mustAuthenticationDispatcher(t, service, next)
	request := httptest.NewRequest("GET", "http://example.test/ws", nil)

	err := dispatcher.Dispatch(NewContext(request.Context(), nil, request))
	if !errors.Is(err, ErrInvalidAuthenticationResult) {
		t.Fatalf("Dispatch() error = %v, want ErrInvalidAuthenticationResult", err)
	}
	if next.callCount() != 0 {
		t.Fatalf("next calls = %d, want 0", next.callCount())
	}
}

func TestAuthenticationDispatcherTransfersOwnershipAndReturnsNextError(t *testing.T) {
	wantErr := errors.New("downstream failed")
	service := &authenticationServiceStub{
		authenticate: func(context.Context, authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
			return successfulAuthentication("api-key"), nil
		},
	}
	next := &authenticatedDispatcherStub{err: wantErr}
	dispatcher := mustAuthenticationDispatcher(t, service, next)
	request := httptest.NewRequest("GET", "http://example.test/ws", nil)

	err := dispatcher.Dispatch(NewContext(request.Context(), nil, request))
	if !errors.Is(err, wantErr) {
		t.Fatalf("Dispatch() error = %v, want next error", err)
	}
	if next.callCount() != 1 {
		t.Fatalf("next calls = %d, want 1", next.callCount())
	}
}

func TestAuthenticatedContextReturnsIndependentPrincipal(t *testing.T) {
	principal := authentication.Principal{
		Authenticated: true,
		Claims:        map[string]string{"tenant": "alpha"},
		Roles:         []string{"admin"},
		Attributes:    map[string]string{"region": "eu"},
		Metadata:      map[string]string{"provider": "test"},
	}
	authenticatedContext := AuthenticatedContext{principal: clonePrincipal(principal)}
	copy := authenticatedContext.Principal()
	copy.Claims["tenant"] = "changed"
	copy.Roles[0] = "changed"
	copy.Attributes["region"] = "changed"
	copy.Metadata["provider"] = "changed"

	unchanged := authenticatedContext.Principal()
	if unchanged.Claims["tenant"] != "alpha" || unchanged.Roles[0] != "admin" ||
		unchanged.Attributes["region"] != "eu" || unchanged.Metadata["provider"] != "test" {
		t.Fatalf("stored Principal was mutated: %+v", unchanged)
	}
}

type authenticationServiceStub struct {
	authenticate func(context.Context, authentication.AuthenticationRequest) (authentication.AuthenticationResult, error)
	calls        atomic.Uint64
	context      atomic.Value
}

func (service *authenticationServiceStub) Authenticate(
	ctx context.Context,
	request authentication.AuthenticationRequest,
) (authentication.AuthenticationResult, error) {
	service.calls.Add(1)
	service.context.Store(ctx)
	if service.authenticate == nil {
		return successfulAuthentication("stub"), nil
	}
	return service.authenticate(ctx, request)
}

type authenticatedDispatcherStub struct {
	mu       sync.Mutex
	contexts []AuthenticatedContext
	err      error
}

func (dispatcher *authenticatedDispatcherStub) DispatchAuthenticated(authenticatedContext AuthenticatedContext) error {
	dispatcher.mu.Lock()
	defer dispatcher.mu.Unlock()
	dispatcher.contexts = append(dispatcher.contexts, authenticatedContext)
	return dispatcher.err
}

func (dispatcher *authenticatedDispatcherStub) callCount() int {
	dispatcher.mu.Lock()
	defer dispatcher.mu.Unlock()
	return len(dispatcher.contexts)
}

func (dispatcher *authenticatedDispatcherStub) lastContext() AuthenticatedContext {
	dispatcher.mu.Lock()
	defer dispatcher.mu.Unlock()
	return dispatcher.contexts[len(dispatcher.contexts)-1]
}

func mustAuthenticationDispatcher(
	t *testing.T,
	service authentication.Service,
	next AuthenticatedDispatcher,
) *AuthenticationDispatcher {
	t.Helper()
	dispatcher, err := NewAuthenticationDispatcher(service, next)
	if err != nil {
		t.Fatalf("NewAuthenticationDispatcher() error = %v", err)
	}
	return dispatcher
}

func successfulAuthentication(providerName string) authentication.AuthenticationResult {
	return authentication.AuthenticationResult{
		Success: true,
		Principal: &authentication.Principal{
			Name:                   providerName,
			AuthenticationProvider: providerName,
			Authenticated:          true,
			Claims:                 map[string]string{"source": "test"},
		},
	}
}

func newAPIKeyService(t *testing.T, resolver secretresolver.Resolver) authentication.Service {
	t.Helper()
	provider, err := authentication.NewAPIKeyProvider(authentication.APIKeyProviderConfig{
		Name:      "runtime-api-key",
		Header:    "X-API-Key",
		SecretRef: "env/UWP_API_KEY",
	}, resolver)
	if err != nil {
		t.Fatalf("NewAPIKeyProvider() error = %v", err)
	}
	service, err := authentication.NewService([]authentication.Provider{provider})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return service
}

func secretResolverFromValue(t *testing.T, reference string, value string) secretresolver.Resolver {
	t.Helper()
	resolver, err := secretresolver.NewMemory(map[string][]byte{reference: []byte(value)})
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	return resolver
}
