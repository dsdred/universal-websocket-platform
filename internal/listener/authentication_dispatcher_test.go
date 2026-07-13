package listener

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	platformconnection "github.com/dsdred/universal-websocket-platform/internal/connection"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

func TestAuthenticationDispatcherListenerIntegrationWithValidAPIKey(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	next := newClosingAuthenticatedDispatcher(websocket.StatusNormalClosure, nil)
	listener := startedListenerWithAuthentication(t, service, next)

	websocketConnection, response := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	defer websocketConnection.CloseNow()
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("handshake status = %d, want %d", response.StatusCode, http.StatusSwitchingProtocols)
	}
	if status := readWebSocketClose(t, websocketConnection); status != websocket.StatusNormalClosure {
		t.Fatalf("close status = %d, want %d", status, websocket.StatusNormalClosure)
	}

	authenticatedContext := next.receivedContext(t)
	transportContext := authenticatedContext.ConnectionContext()
	if transportContext.Connection() == nil || transportContext.Request() == nil || transportContext.Context() == nil {
		t.Fatal("next received incomplete transport context")
	}
	if transportContext.Request().URL.Path != websocketPath {
		t.Fatalf("request path = %q, want %q", transportContext.Request().URL.Path, websocketPath)
	}
	if transportContext.Request().Header.Get("X-API-Key") != "correct-key" {
		t.Fatal("next did not receive the original HTTP request")
	}
	principal := authenticatedContext.Principal()
	if !principal.Authenticated || principal.AuthenticationProvider != "listener-api-key" {
		t.Fatalf("Principal = %+v", principal)
	}
	if next.calls.Load() != 1 {
		t.Fatalf("next calls = %d, want 1", next.calls.Load())
	}
}

func TestAuthenticationDispatcherListenerRejectsCredentialsAndContinues(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	next := newClosingAuthenticatedDispatcher(websocket.StatusNormalClosure, nil)
	listener := startedListenerWithAuthentication(t, service, next)

	for _, test := range []struct {
		name  string
		value string
	}{
		{name: "missing"},
		{name: "invalid", value: "wrong-key"},
	} {
		t.Run(test.name, func(t *testing.T) {
			websocketConnection, _ := dialWebSocketWithHeader(t, listener, "X-API-Key", test.value)
			defer websocketConnection.CloseNow()
			if status := readWebSocketClose(t, websocketConnection); status != websocket.StatusPolicyViolation {
				t.Fatalf("close status = %d, want %d", status, websocket.StatusPolicyViolation)
			}
		})
	}
	if next.calls.Load() != 0 {
		t.Fatalf("next calls after rejected credentials = %d, want 0", next.calls.Load())
	}

	response, err := testHTTPClient().Get("http://" + listener.Address() + "/")
	if err != nil {
		t.Fatalf("GET / after rejected credentials error = %v", err)
	}
	if err := response.Body.Close(); err != nil {
		t.Fatalf("response Body.Close() error = %v", err)
	}
	if response.StatusCode != http.StatusNotImplemented {
		t.Fatalf("GET / status = %d, want %d", response.StatusCode, http.StatusNotImplemented)
	}

	validConnection, _ := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	defer validConnection.CloseNow()
	if status := readWebSocketClose(t, validConnection); status != websocket.StatusNormalClosure {
		t.Fatalf("valid connection close status = %d, want %d", status, websocket.StatusNormalClosure)
	}
	if next.calls.Load() != 1 {
		t.Fatalf("next calls = %d, want 1", next.calls.Load())
	}
}

func TestAuthenticationDispatcherListenerClosesOnResolverError(t *testing.T) {
	wantErr := errors.New("resolver unavailable")
	service := listenerAPIKeyService(t, failingResolver{err: wantErr})
	next := newClosingAuthenticatedDispatcher(websocket.StatusNormalClosure, nil)
	listener := startedListenerWithAuthentication(t, service, next)

	websocketConnection, _ := dialWebSocketWithHeader(t, listener, "X-API-Key", "credential")
	defer websocketConnection.CloseNow()
	if status := readWebSocketClose(t, websocketConnection); status != websocket.StatusInternalError {
		t.Fatalf("close status = %d, want %d", status, websocket.StatusInternalError)
	}
	if next.calls.Load() != 0 {
		t.Fatalf("next calls = %d, want 0", next.calls.Load())
	}
}

func TestAuthenticationDispatcherDoesNotCloseAfterNextError(t *testing.T) {
	wantErr := errors.New("downstream failed")
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	next := newClosingAuthenticatedDispatcher(websocket.StatusGoingAway, wantErr)
	listener := startedListenerWithAuthentication(t, service, next)

	websocketConnection, _ := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	defer websocketConnection.CloseNow()
	if status := readWebSocketClose(t, websocketConnection); status != websocket.StatusGoingAway {
		t.Fatalf("close status = %d, want downstream status %d", status, websocket.StatusGoingAway)
	}
	if next.calls.Load() != 1 {
		t.Fatalf("next calls = %d, want 1", next.calls.Load())
	}
}

func TestAuthenticationDispatcherConcurrentAuthenticatedConnections(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	next := newClosingAuthenticatedDispatcher(websocket.StatusNormalClosure, nil)
	listener := startedListenerWithAuthentication(t, service, next)
	const connections = 16

	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, connections)
	for range connections {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			websocketConnection, _, err := websocket.Dial(ctx, websocketURL(listener, websocketPath), &websocket.DialOptions{
				HTTPHeader: http.Header{"X-API-Key": []string{"correct-key"}},
			})
			if err != nil {
				errorsChannel <- fmt.Errorf("Dial(): %w", err)
				return
			}
			defer websocketConnection.CloseNow()
			_, _, readErr := websocketConnection.Read(ctx)
			if status := websocket.CloseStatus(readErr); status != websocket.StatusNormalClosure {
				errorsChannel <- fmt.Errorf("close status = %d: %w", status, readErr)
			}
		}()
	}
	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		t.Errorf("concurrent connection error = %v", err)
	}
	if next.calls.Load() != connections {
		t.Fatalf("next calls = %d, want %d", next.calls.Load(), connections)
	}
}

func TestAuthenticationDispatcherListenerSmokeScenario(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	next := newClosingAuthenticatedDispatcher(websocket.StatusNormalClosure, nil)
	listener := startedListenerWithAuthentication(t, service, next)
	address := listener.Address()

	websocketConnection, response := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	if status := readWebSocketClose(t, websocketConnection); status != websocket.StatusNormalClosure {
		t.Fatalf("close status = %d, want %d", status, websocket.StatusNormalClosure)
	}
	websocketConnection.CloseNow()
	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	assertTCPPortAvailable(t, address)
	t.Logf("Upgrade %d -> Authentication success -> downstream -> normal closure -> Stop -> port released", response.StatusCode)
}

type closingAuthenticatedDispatcher struct {
	status   websocket.StatusCode
	err      error
	calls    atomic.Uint64
	received chan platformconnection.AuthenticatedContext
}

func newClosingAuthenticatedDispatcher(status websocket.StatusCode, err error) *closingAuthenticatedDispatcher {
	return &closingAuthenticatedDispatcher{
		status:   status,
		err:      err,
		received: make(chan platformconnection.AuthenticatedContext, 64),
	}
}

func (dispatcher *closingAuthenticatedDispatcher) DispatchAuthenticated(
	authenticatedContext platformconnection.AuthenticatedContext,
) error {
	dispatcher.calls.Add(1)
	dispatcher.received <- authenticatedContext
	websocketConnection := authenticatedContext.ConnectionContext().Connection()
	defer websocketConnection.CloseNow()
	_ = websocketConnection.Close(dispatcher.status, "")
	return dispatcher.err
}

func (dispatcher *closingAuthenticatedDispatcher) receivedContext(t *testing.T) platformconnection.AuthenticatedContext {
	t.Helper()
	select {
	case authenticatedContext := <-dispatcher.received:
		return authenticatedContext
	case <-time.After(time.Second):
		t.Fatal("next Dispatcher did not receive AuthenticatedContext")
		return platformconnection.AuthenticatedContext{}
	}
}

func startedListenerWithAuthentication(
	t *testing.T,
	service authentication.Service,
	next platformconnection.AuthenticatedDispatcher,
) Listener {
	t.Helper()
	dispatcher, err := platformconnection.NewAuthenticationDispatcher(service, next)
	if err != nil {
		t.Fatalf("NewAuthenticationDispatcher() error = %v", err)
	}
	return startedListenerWithDispatcher(t, dispatcher)
}

func listenerAPIKeyService(t *testing.T, resolver secretresolver.Resolver) authentication.Service {
	t.Helper()
	provider, err := authentication.NewAPIKeyProvider(authentication.APIKeyProviderConfig{
		Name:      "listener-api-key",
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

func listenerMemoryResolver(t *testing.T, value string) secretresolver.Resolver {
	t.Helper()
	resolver, err := secretresolver.NewMemory(map[string][]byte{"env/UWP_API_KEY": []byte(value)})
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	return resolver
}

type failingResolver struct {
	err error
}

func (resolver failingResolver) Resolve(context.Context, string) (secretresolver.Secret, error) {
	return secretresolver.Secret{}, resolver.err
}

func dialWebSocketWithHeader(
	t *testing.T,
	listener Listener,
	headerName string,
	headerValue string,
) (*websocket.Conn, *http.Response) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	header := make(http.Header)
	if headerValue != "" {
		header.Set(headerName, headerValue)
	}
	websocketConnection, response, err := websocket.Dial(ctx, websocketURL(listener, websocketPath), &websocket.DialOptions{
		HTTPHeader: header,
	})
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	return websocketConnection, response
}

func readWebSocketClose(t *testing.T, websocketConnection *websocket.Conn) websocket.StatusCode {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, _, err := websocketConnection.Read(ctx)
	return websocket.CloseStatus(err)
}
