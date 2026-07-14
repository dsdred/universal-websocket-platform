package session

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/connection"
)

func TestDispatcherImplementsAuthenticatedDispatcher(t *testing.T) {
	var _ connection.AuthenticatedDispatcher = NewDispatcher()
}

func TestDispatcherCreatesStartsAndStopsSession(t *testing.T) {
	runtimeSession := &sessionStub{}
	factory := &sessionFactoryStub{session: runtimeSession}
	dispatcher := newDispatcher(factory)
	request := httptest.NewRequest("GET", "http://example.test/ws?credential=not-retained", nil)
	request.RemoteAddr = "192.0.2.1:4321"
	request.Header.Set("Authorization", "not-retained")

	err := dispatchToSession(t, context.Background(), request, dispatcher, validPrincipal(), nil)
	if err != nil {
		t.Fatalf("DispatchAuthenticated() error = %v", err)
	}
	if factory.calls.Load() != 1 || runtimeSession.startCalls.Load() != 1 || runtimeSession.runCalls.Load() != 1 || runtimeSession.stopCalls.Load() != 1 {
		t.Fatalf("calls = (Create %d, Start %d, Run %d, Stop %d)", factory.calls.Load(), runtimeSession.startCalls.Load(), runtimeSession.runCalls.Load(), runtimeSession.stopCalls.Load())
	}
	if factory.remoteAddress != "192.0.2.1:4321" {
		t.Fatalf("remote address = %q", factory.remoteAddress)
	}
	if factory.principal.ID != validPrincipal().ID || !factory.principal.Authenticated {
		t.Fatalf("Principal = %+v", factory.principal)
	}
}

func TestDispatcherReturnsFactoryError(t *testing.T) {
	wantErr := errors.New("create Session")
	factory := &sessionFactoryStub{err: wantErr}
	dispatcher := newDispatcher(factory)
	serverConnection, _ := testWebSocketPair(t)
	request := httptest.NewRequest("GET", "http://example.test/ws", nil)

	err := dispatchToSession(t, context.Background(), request, dispatcher, validPrincipal(), serverConnection)
	if !errors.Is(err, wantErr) {
		t.Fatalf("DispatchAuthenticated() error = %v, want Factory error", err)
	}
}

func TestDispatcherReturnsStartErrorAndStopsSession(t *testing.T) {
	wantErr := errors.New("start Session")
	runtimeSession := &sessionStub{startErr: wantErr}
	dispatcher := newDispatcher(&sessionFactoryStub{session: runtimeSession})
	request := httptest.NewRequest("GET", "http://example.test/ws", nil)

	err := dispatchToSession(t, context.Background(), request, dispatcher, validPrincipal(), nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("DispatchAuthenticated() error = %v, want Start error", err)
	}
	if runtimeSession.stopCalls.Load() != 1 {
		t.Fatalf("Stop() calls = %d, want 1", runtimeSession.stopCalls.Load())
	}
}

func TestDispatcherReturnsStopError(t *testing.T) {
	wantErr := errors.New("stop Session")
	runtimeSession := &sessionStub{stopErr: wantErr}
	dispatcher := newDispatcher(&sessionFactoryStub{session: runtimeSession})
	request := httptest.NewRequest("GET", "http://example.test/ws", nil)

	err := dispatchToSession(t, context.Background(), request, dispatcher, validPrincipal(), nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("DispatchAuthenticated() error = %v, want Stop error", err)
	}
}

func TestDispatcherReturnsRunErrorAndStillStopsSession(t *testing.T) {
	wantErr := errors.New("read failed")
	runtimeSession := &sessionStub{runErr: wantErr}
	dispatcher := newDispatcher(&sessionFactoryStub{session: runtimeSession})
	request := httptest.NewRequest("GET", "http://example.test/ws", nil)

	err := dispatchToSession(t, context.Background(), request, dispatcher, validPrincipal(), nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("DispatchAuthenticated() error = %v, want Run error", err)
	}
	if runtimeSession.stopCalls.Load() != 1 {
		t.Fatalf("Stop() calls = %d, want 1", runtimeSession.stopCalls.Load())
	}
}

func TestDispatcherRejectsCanceledContextBeforeCreatingSession(t *testing.T) {
	serverConnection, _ := testWebSocketPair(t)
	factory := &sessionFactoryStub{session: &sessionStub{}}
	dispatcher := newDispatcher(factory)
	ctx, cancel := context.WithCancel(context.Background())
	service := authenticationServiceFunc(func(context.Context, authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
		cancel()
		return successfulResult(validPrincipal()), nil
	})
	authenticationDispatcher, err := connection.NewAuthenticationDispatcher(service, dispatcher)
	if err != nil {
		t.Fatalf("NewAuthenticationDispatcher() error = %v", err)
	}
	request := httptest.NewRequest("GET", "http://example.test/ws", nil).WithContext(ctx)

	err = authenticationDispatcher.Dispatch(connection.NewContext(ctx, serverConnection, request))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Dispatch() error = %v, want context.Canceled", err)
	}
	if factory.calls.Load() != 0 {
		t.Fatalf("Factory calls = %d, want 0", factory.calls.Load())
	}
}

func TestProductionDispatcherRunsUntilClientClosesNormally(t *testing.T) {
	serverConnection, clientConnection := testWebSocketPair(t)
	dispatcher := NewDispatcher()
	request := httptest.NewRequest("GET", "http://example.test/ws", nil)
	request.RemoteAddr = "192.0.2.1:4321"
	dispatchResult := make(chan error, 1)
	go func() {
		dispatchResult <- dispatchToSession(t, request.Context(), request, dispatcher, validPrincipal(), serverConnection)
	}()

	if err := clientConnection.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("client Close() error = %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case err := <-dispatchResult:
		if err != nil {
			t.Fatalf("DispatchAuthenticated() error = %v", err)
		}
	case <-ctx.Done():
		t.Fatalf("DispatchAuthenticated() did not return: %v", ctx.Err())
	}
}

func TestDispatcherConcurrentAuthenticatedContexts(t *testing.T) {
	const operations = 64
	factory := &sessionFactoryStub{create: func() Session { return &sessionStub{} }}
	dispatcher := newDispatcher(factory)
	request := httptest.NewRequest("GET", "http://example.test/ws", nil)
	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, operations)
	for range operations {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			errorsChannel <- dispatchToSession(t, request.Context(), request, dispatcher, validPrincipal(), nil)
		}()
	}
	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		if err != nil {
			t.Errorf("DispatchAuthenticated() error = %v", err)
		}
	}
	if factory.calls.Load() != operations {
		t.Fatalf("Factory calls = %d, want %d", factory.calls.Load(), operations)
	}
}

func dispatchToSession(
	t *testing.T,
	ctx context.Context,
	request *http.Request,
	dispatcher connection.AuthenticatedDispatcher,
	principal authentication.Principal,
	websocketConnection *websocket.Conn,
) error {
	t.Helper()
	service := authenticationServiceFunc(func(context.Context, authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
		return successfulResult(principal), nil
	})
	authenticationDispatcher, err := connection.NewAuthenticationDispatcher(service, dispatcher)
	if err != nil {
		t.Fatalf("NewAuthenticationDispatcher() error = %v", err)
	}
	return authenticationDispatcher.Dispatch(connection.NewContext(ctx, websocketConnection, request))
}

func successfulResult(principal authentication.Principal) authentication.AuthenticationResult {
	return authentication.AuthenticationResult{Success: true, Principal: &principal}
}

type authenticationServiceFunc func(context.Context, authentication.AuthenticationRequest) (authentication.AuthenticationResult, error)

func (function authenticationServiceFunc) Authenticate(
	ctx context.Context,
	request authentication.AuthenticationRequest,
) (authentication.AuthenticationResult, error) {
	return function(ctx, request)
}

type sessionFactoryStub struct {
	mu            sync.Mutex
	session       Session
	create        func() Session
	err           error
	calls         atomic.Uint64
	connection    *websocket.Conn
	principal     authentication.Principal
	remoteAddress string
}

func (factory *sessionFactoryStub) Create(
	websocketConnection *websocket.Conn,
	principal authentication.Principal,
	remoteAddress string,
) (Session, error) {
	factory.calls.Add(1)
	factory.mu.Lock()
	defer factory.mu.Unlock()
	factory.connection = websocketConnection
	factory.principal = principal
	factory.remoteAddress = remoteAddress
	if factory.err != nil {
		return nil, factory.err
	}
	if factory.create != nil {
		return factory.create(), nil
	}
	return factory.session, nil
}

type sessionStub struct {
	startErr   error
	runErr     error
	stopErr    error
	startCalls atomic.Uint64
	runCalls   atomic.Uint64
	stopCalls  atomic.Uint64
}

func (*sessionStub) ID() string                          { return "test-session" }
func (*sessionStub) Principal() authentication.Principal { return validPrincipal() }
func (*sessionStub) RemoteAddress() string               { return "192.0.2.1:4321" }
func (*sessionStub) CreatedAt() time.Time                { return time.Now().UTC() }
func (*sessionStub) Running() bool                       { return false }
func (session *sessionStub) Start(context.Context) error {
	session.startCalls.Add(1)
	return session.startErr
}
func (session *sessionStub) Run(context.Context) error {
	session.runCalls.Add(1)
	return session.runErr
}
func (session *sessionStub) Stop(context.Context) error {
	session.stopCalls.Add(1)
	return session.stopErr
}
