package handshake

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/connection"
)

func TestHandlerRejectsNilDependencies(t *testing.T) {
	admission := &admissionStub{}
	runtimeContext := contextProvider{ctx: context.Background()}
	service := &serviceStub{}
	handoff := &handoffStub{}

	tests := []struct {
		name string
		new  func() (*Handler, error)
		want error
	}{
		{name: "admission", new: func() (*Handler, error) { return NewHandler(nil, runtimeContext, service, handoff) }, want: ErrNilAdmissionCapability},
		{name: "runtime context", new: func() (*Handler, error) { return NewHandler(admission, nil, service, handoff) }, want: ErrNilRuntimeContextProvider},
		{name: "authentication", new: func() (*Handler, error) { return NewHandler(admission, runtimeContext, nil, handoff) }, want: ErrNilAuthenticationService},
		{name: "handoff", new: func() (*Handler, error) { return NewHandler(admission, runtimeContext, service, nil) }, want: ErrNilSessionHandoff},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			created, err := test.new()
			if created != nil || !errors.Is(err, test.want) {
				t.Fatalf("NewHandler() = (%v, %v), want nil and %v", created, err, test.want)
			}
		})
	}
}

func TestHandlerClosedAdmissionSkipsAuthenticationAndUpgrade(t *testing.T) {
	admission := &admissionStub{}
	service := &serviceStub{result: allowedResult()}
	handoff := &handoffStub{}
	server := newHandshakeServer(t, admission, service, handoff)

	response := rejectedWebSocketDial(t, server.URL)
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusServiceUnavailable)
	}
	if service.calls.Load() != 0 || handoff.calls.Load() != 0 {
		t.Fatalf("calls = (Authentication %d, handoff %d), want zero", service.calls.Load(), handoff.calls.Load())
	}
	assertSafeResponseBody(t, response, "credential-that-must-not-leak")
}

func TestHandlerAuthenticationRejectSkipsUpgradeAndSession(t *testing.T) {
	admission := openAdmission()
	service := &serviceStub{}
	handoff := &handoffStub{}
	server := newHandshakeServer(t, admission, service, handoff)

	response := rejectedWebSocketDialWithHeader(t, server.URL, "X-API-Key", "rejected-credential")
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusUnauthorized)
	}
	if service.calls.Load() != 1 || handoff.calls.Load() != 0 {
		t.Fatalf("calls = (Authentication %d, handoff %d)", service.calls.Load(), handoff.calls.Load())
	}
	assertSafeResponseBody(t, response, "rejected-credential")
}

func TestHandlerRejectsInvalidSuccessfulAuthenticationResult(t *testing.T) {
	for _, test := range []struct {
		name      string
		principal *authentication.Principal
	}{
		{name: "missing Principal"},
		{name: "neither authenticated nor anonymous", principal: &authentication.Principal{ID: "invalid"}},
		{name: "anonymous with wrong ID", principal: &authentication.Principal{ID: "guest", Anonymous: true}},
		{name: "conflicting flags", principal: &authentication.Principal{ID: "invalid", Authenticated: true, Anonymous: true}},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := &serviceStub{result: authentication.AuthenticationResult{Success: true, Principal: test.principal}}
			handoff := &handoffStub{}
			server := newHandshakeServer(t, openAdmission(), service, handoff)

			response := rejectedWebSocketDial(t, server.URL)
			if response.StatusCode != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusUnauthorized)
			}
			if service.calls.Load() != 1 || handoff.calls.Load() != 0 {
				t.Fatalf("calls = (Authentication %d, handoff %d), want (1, 0)", service.calls.Load(), handoff.calls.Load())
			}
		})
	}
}

func TestHandlerAuthenticationErrorSkipsUpgradeAndSession(t *testing.T) {
	admission := openAdmission()
	wantErr := errors.New("provider unavailable")
	service := &serviceStub{err: wantErr}
	handoff := &handoffStub{}
	server := newHandshakeServer(t, admission, service, handoff)

	response := rejectedWebSocketDialWithHeader(t, server.URL, "Authorization", "sensitive-credential")
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusServiceUnavailable)
	}
	if service.calls.Load() != 1 || handoff.calls.Load() != 0 {
		t.Fatalf("calls = (Authentication %d, handoff %d)", service.calls.Load(), handoff.calls.Load())
	}
	assertSafeResponseBody(t, response, "sensitive-credential")
}

func TestHandlerConfiguredDeadlinePreventsUpgrade(t *testing.T) {
	entered := make(chan struct{})
	deadlineObserved := make(chan time.Time, 1)
	service := &serviceStub{authenticate: func(ctx context.Context, _ authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
		deadline, ok := ctx.Deadline()
		if !ok {
			return authentication.AuthenticationResult{}, errors.New("configured deadline missing")
		}
		deadlineObserved <- deadline
		close(entered)
		<-ctx.Done()
		// A context-aware dependency normally returns ctx.Err(). Returning Allow here
		// proves that the Handshake itself invalidates a Decision after its deadline.
		return allowedResult(), nil
	}}
	handoff := &handoffStub{}
	reporter := &terminalErrorRecorder{}
	handler, err := NewHandlerWithTerminalErrorReporter(
		openAdmission(),
		contextProvider{ctx: context.Background()},
		service,
		handoff,
		reporter.Report,
	)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	timedHandler := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		ctx, cancel := context.WithTimeoutCause(request.Context(), 25*time.Millisecond, ErrHandshakeTimeout)
		defer cancel()
		handler.ServeHTTP(response, request.WithContext(ctx))
	})
	server := httptest.NewServer(timedHandler)
	defer server.Close()

	type dialResult struct {
		connection *websocket.Conn
		response   *http.Response
		err        error
	}
	dialResultChannel := make(chan dialResult, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		connection, response, dialErr := websocket.Dial(ctx, websocketURL(server.URL), nil)
		dialResultChannel <- dialResult{connection: connection, response: response, err: dialErr}
	}()
	receive(t, entered, "Authentication entry")
	if deadline := receive(t, deadlineObserved, "configured Authentication deadline"); !deadline.After(time.Now().Add(-time.Second)) {
		t.Fatalf("Authentication deadline = %v, want configured future deadline", deadline)
	}
	dial := receive(t, dialResultChannel, "Handshake timeout response")
	if dial.connection != nil {
		dial.connection.CloseNow()
		t.Fatal("Handshake timeout unexpectedly upgraded WebSocket")
	}
	if dial.err == nil || dial.response == nil {
		t.Fatalf("websocket.Dial() = (%v, %v), want HTTP rejection", dial.response, dial.err)
	}
	response := dial.response
	defer response.Body.Close()
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusServiceUnavailable)
	}
	if handoff.calls.Load() != 0 {
		t.Fatalf("handoff calls = %d, want 0", handoff.calls.Load())
	}
	if reported := reporter.Single(t); !errors.Is(reported, ErrHandshakeTimeout) {
		t.Fatalf("reported error = %v, want ErrHandshakeTimeout", reported)
	} else if strings.Contains(reported.Error(), "credential-that-must-not-leak") {
		t.Fatal("timeout error exposed credentials")
	}
	assertSafeResponseBody(t, response, "credential-that-must-not-leak")
}

func TestHandlerFinalAdmissionValidationPreventsUpgrade(t *testing.T) {
	admission := openAdmission()
	entered := make(chan struct{})
	release := make(chan struct{})
	service := &serviceStub{authenticate: func(context.Context, authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
		close(entered)
		<-release
		return allowedResult(), nil
	}}
	handoff := &handoffStub{}
	server := newHandshakeServer(t, admission, service, handoff)

	responseChannel := make(chan *http.Response, 1)
	go func() {
		responseChannel <- rejectedWebSocketDial(t, server.URL)
	}()
	receive(t, entered, "Authentication entry")
	admission.open.Store(false)
	close(release)
	response := receive(t, responseChannel, "final admission response")
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusServiceUnavailable)
	}
	if service.calls.Load() != 1 {
		t.Fatalf("Authentication calls = %d, want 1", service.calls.Load())
	}
	if handoff.calls.Load() != 0 {
		t.Fatalf("handoff calls = %d, want 0", handoff.calls.Load())
	}
}

func TestHandlerUnavailableRuntimeContextFailsClosed(t *testing.T) {
	for _, test := range []struct {
		name string
		ctx  context.Context
	}{
		{name: "nil"},
		{name: "canceled", ctx: canceledContext()},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := &serviceStub{result: allowedResult()}
			handoff := &handoffStub{}
			handler, err := NewHandler(openAdmission(), contextProvider{ctx: test.ctx}, service, handoff)
			if err != nil {
				t.Fatalf("NewHandler() error = %v", err)
			}
			server := httptest.NewServer(handler)
			defer server.Close()

			response := rejectedWebSocketDial(t, server.URL)
			if response.StatusCode != http.StatusServiceUnavailable {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusServiceUnavailable)
			}
			if service.calls.Load() != 0 || handoff.calls.Load() != 0 {
				t.Fatalf("calls = (Authentication %d, handoff %d), want zero", service.calls.Load(), handoff.calls.Load())
			}
		})
	}
}

func TestHandlerAllowUpgradesAndCreatesSessionHandoff(t *testing.T) {
	admission := openAdmission()
	service := &serviceStub{result: allowedResult()}
	handoff := closingHandoff(websocket.StatusNormalClosure)
	reporter := &terminalErrorRecorder{}
	server := newHandshakeServerWithReporter(t, admission, service, handoff, reporter.Report)

	client, response := successfulWebSocketDial(t, server.URL)
	defer client.CloseNow()
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusSwitchingProtocols)
	}
	if status := readClose(t, client); status != websocket.StatusNormalClosure {
		t.Fatalf("close status = %d, want %d", status, websocket.StatusNormalClosure)
	}
	if service.calls.Load() != 1 || handoff.calls.Load() != 1 {
		t.Fatalf("calls = (Authentication %d, handoff %d)", service.calls.Load(), handoff.calls.Load())
	}
	if reporter.Count() != 0 {
		t.Fatalf("reported errors = %d, want 0", reporter.Count())
	}
}

func TestHandlerAcceptFailureDoesNotCreateSession(t *testing.T) {
	admission := openAdmission()
	service := &serviceStub{result: allowedResult()}
	handoff := &handoffStub{}
	reporter := &terminalErrorRecorder{}
	server := newHandshakeServerWithReporter(t, admission, service, handoff, reporter.Report)

	response, err := server.Client().Get(server.URL)
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	response.Body.Close()
	if response.StatusCode == http.StatusSwitchingProtocols {
		t.Fatal("plain HTTP request unexpectedly upgraded")
	}
	if service.calls.Load() != 1 || handoff.calls.Load() != 0 {
		t.Fatalf("calls = (Authentication %d, handoff %d)", service.calls.Load(), handoff.calls.Load())
	}
	if reporter.Count() != 0 {
		t.Fatalf("reported errors = %d, want 0 for protocol rejection", reporter.Count())
	}
}

func TestHandlerHandoffErrorClosesOwnedWebSocket(t *testing.T) {
	admission := openAdmission()
	service := &serviceStub{result: allowedResult()}
	wantErr := errors.New("create Session with credential-that-must-not-leak")
	reporter := &terminalErrorRecorder{}
	receivedContext := make(chan connection.ConnectionContext, 1)
	handoff := &handoffStub{dispatch: func(authenticated connection.AuthenticatedContext) (bool, error) {
		receivedContext <- authenticated.ConnectionContext()
		return false, wantErr
	}}
	server := newHandshakeServerWithReporter(t, admission, service, handoff, reporter.Report)

	client, _ := successfulWebSocketDial(t, server.URL)
	defer client.CloseNow()
	if status := readClose(t, client); status != websocket.StatusInternalError {
		t.Fatalf("close status = %d, want %d", status, websocket.StatusInternalError)
	}
	if handoff.calls.Load() != 1 {
		t.Fatalf("handoff calls = %d, want 1", handoff.calls.Load())
	}
	connectionContext := receive(t, receivedContext, "failed Session handoff")
	waitDone(t, connectionContext.Context().Done(), "failed handoff context cancellation")
	reported := reporter.Single(t)
	if !errors.Is(reported, wantErr) {
		t.Fatalf("reported error does not preserve handoff sentinel: %v", reported)
	}
	if strings.Contains(reported.Error(), "credential-that-must-not-leak") {
		t.Fatal("reported error exposes credentials")
	}
}

func TestHandlerAcceptedHandoffReportsDownstreamErrorWithoutReclaimingOwnership(t *testing.T) {
	wantErr := errors.New("Session lifecycle failed")
	reporter := &terminalErrorRecorder{}
	handoff := &handoffStub{dispatch: func(authenticated connection.AuthenticatedContext) (bool, error) {
		connectionContext := authenticated.ConnectionContext()
		defer connectionContext.Cancel()
		defer connectionContext.Connection().CloseNow()
		_ = connectionContext.Connection().Close(websocket.StatusGoingAway, "")
		return true, wantErr
	}}
	server := newHandshakeServerWithReporter(
		t,
		openAdmission(),
		&serviceStub{result: allowedResult()},
		handoff,
		reporter.Report,
	)

	client, _ := successfulWebSocketDial(t, server.URL)
	defer client.CloseNow()
	if status := readClose(t, client); status != websocket.StatusGoingAway {
		t.Fatalf("close status = %d, want Session-owned %d", status, websocket.StatusGoingAway)
	}
	if handoff.calls.Load() != 1 {
		t.Fatalf("handoff calls = %d, want 1", handoff.calls.Load())
	}
	if reported := reporter.Single(t); !errors.Is(reported, wantErr) {
		t.Fatalf("reported error = %v, want downstream sentinel", reported)
	}
}

func TestHandlerAcceptedHandoffContextCancellationIsNotReported(t *testing.T) {
	reporter := &terminalErrorRecorder{}
	received := make(chan connection.ConnectionContext, 1)
	handoff := &handoffStub{dispatch: func(authenticated connection.AuthenticatedContext) (bool, error) {
		connectionContext := authenticated.ConnectionContext()
		received <- connectionContext
		defer connectionContext.Cancel()
		defer connectionContext.Connection().CloseNow()
		_ = connectionContext.Connection().Close(websocket.StatusNormalClosure, "")
		return true, context.Canceled
	}}
	server := newHandshakeServerWithReporter(
		t,
		openAdmission(),
		&serviceStub{result: allowedResult()},
		handoff,
		reporter.Report,
	)

	client, _ := successfulWebSocketDial(t, server.URL)
	defer client.CloseNow()
	if status := readClose(t, client); status != websocket.StatusNormalClosure {
		t.Fatalf("close status = %d, want %d", status, websocket.StatusNormalClosure)
	}
	connectionContext := receive(t, received, "accepted canceled handoff")
	waitDone(t, connectionContext.Request().Context().Done(), "Handshake completion")
	if reporter.Count() != 0 {
		t.Fatalf("reported errors = %d, want 0 for lifecycle cancellation", reporter.Count())
	}
}

func TestHandlerReporterPanicDoesNotChangeHandoffCleanup(t *testing.T) {
	handoff := &handoffStub{dispatch: func(connection.AuthenticatedContext) (bool, error) {
		return false, errors.New("handoff failed")
	}}
	server := newHandshakeServerWithReporter(
		t,
		openAdmission(),
		&serviceStub{result: allowedResult()},
		handoff,
		func(error) { panic("reporter failed") },
	)

	client, _ := successfulWebSocketDial(t, server.URL)
	defer client.CloseNow()
	if status := readClose(t, client); status != websocket.StatusInternalError {
		t.Fatalf("close status = %d, want %d", status, websocket.StatusInternalError)
	}
}

func TestHandlerSessionContextOutlivesHTTPRequestContext(t *testing.T) {
	admission := openAdmission()
	service := &serviceStub{result: allowedResult()}
	received := make(chan connection.ConnectionContext, 1)
	handoff := &handoffStub{dispatch: func(authenticated connection.AuthenticatedContext) (bool, error) {
		received <- authenticated.ConnectionContext()
		return true, nil
	}}
	runtimeContext, cancelRuntime := context.WithCancel(context.Background())
	defer cancelRuntime()
	handler, err := NewHandler(admission, contextProvider{ctx: runtimeContext}, service, handoff)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	client, _ := successfulWebSocketDial(t, server.URL)
	defer client.CloseNow()
	connectionContext := receive(t, received, "Session handoff")
	requestContext := connectionContext.Request().Context()
	waitDone(t, requestContext.Done(), "HTTP request context completion")
	select {
	case <-connectionContext.Context().Done():
		t.Fatalf("Session context ended with request context: %v", connectionContext.Context().Err())
	default:
	}
	cancelRuntime()
	waitDone(t, connectionContext.Context().Done(), "Runtime-derived Session context cancellation")
	if !errors.Is(connectionContext.Context().Err(), context.Canceled) {
		t.Fatalf("Session context error = %v, want context.Canceled", connectionContext.Context().Err())
	}
	connectionContext.Connection().CloseNow()
}

func TestHandlerGateClosureDuringConcurrentEvaluationPreventsNewAdmissions(t *testing.T) {
	const count = 16
	admission := openAdmission()
	entered := make(chan struct{}, count)
	release := make(chan struct{})
	service := &serviceStub{authenticate: func(context.Context, authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
		entered <- struct{}{}
		<-release
		return allowedResult(), nil
	}}
	handoff := &handoffStub{}
	server := newHandshakeServer(t, admission, service, handoff)

	responses := make(chan *http.Response, count)
	for range count {
		go func() { responses <- rejectedWebSocketDial(t, server.URL) }()
	}
	for range count {
		receive(t, entered, "concurrent Authentication entry")
	}
	admission.open.Store(false)

	closedGateResponse := rejectedWebSocketDial(t, server.URL)
	if closedGateResponse.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("closed Gate status = %d, want %d", closedGateResponse.StatusCode, http.StatusServiceUnavailable)
	}
	if service.calls.Load() != count {
		t.Fatalf("Authentication calls after closed-Gate request = %d, want %d", service.calls.Load(), count)
	}
	close(release)
	for range count {
		response := receive(t, responses, "concurrent final Gate response")
		if response.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusServiceUnavailable)
		}
	}
	if handoff.calls.Load() != 0 {
		t.Fatalf("handoff calls = %d, want 0", handoff.calls.Load())
	}
}

func TestHandlerConcurrentHandshakesCreateExactlyOneSessionEach(t *testing.T) {
	const count = 16
	admission := openAdmission()
	entered := make(chan struct{}, count)
	release := make(chan struct{})
	service := &serviceStub{authenticate: func(context.Context, authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
		entered <- struct{}{}
		<-release
		return allowedResult(), nil
	}}
	handoff := closingHandoff(websocket.StatusNormalClosure)
	server := newHandshakeServer(t, admission, service, handoff)

	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, count)
	for range count {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			client, _, err := websocket.Dial(ctx, websocketURL(server.URL), nil)
			if err != nil {
				errorsChannel <- err
				return
			}
			defer client.CloseNow()
			_, _, readErr := client.Read(ctx)
			if websocket.CloseStatus(readErr) != websocket.StatusNormalClosure {
				errorsChannel <- fmt.Errorf("close status %d: %w", websocket.CloseStatus(readErr), readErr)
			}
		}()
	}
	for range count {
		receive(t, entered, "Authentication barrier")
	}
	close(release)
	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		t.Errorf("concurrent Handshake error = %v", err)
	}
	if service.calls.Load() != count || handoff.calls.Load() != count {
		t.Fatalf("calls = (Authentication %d, handoff %d), want %d each", service.calls.Load(), handoff.calls.Load(), count)
	}
}

func TestAuthenticationRequestCopiesTransportMetadata(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"http://example.test/ws?tenant=first&tenant=second",
		nil,
	)
	request.RemoteAddr = "192.0.2.10:4321"
	request.Header["X-Test"] = []string{"first", "second"}

	normalized := authenticationRequest(request)
	request.Header["X-Test"][0] = "mutated"
	request.Header.Set("X-Other", "added")

	if normalized.Transport != "websocket" || normalized.RemoteAddress != "192.0.2.10:4321" {
		t.Fatalf("transport metadata = (%q, %q)", normalized.Transport, normalized.RemoteAddress)
	}
	if got := normalized.Headers["X-Test"]; len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("normalized Headers = %v, want independent copy", normalized.Headers)
	}
	if _, exists := normalized.Headers["X-Other"]; exists {
		t.Fatal("normalized Headers observed later request mutation")
	}
	if got := normalized.Query["tenant"]; len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("normalized Query = %v", normalized.Query)
	}
	normalized.Query["tenant"][0] = "mutated"
	if request.URL.Query().Get("tenant") != "first" {
		t.Fatal("normalized Query mutation changed the HTTP request")
	}
}

type admissionStub struct {
	open atomic.Bool
}

func openAdmission() *admissionStub {
	admission := &admissionStub{}
	admission.open.Store(true)
	return admission
}

func (admission *admissionStub) CanAccept() bool { return admission.open.Load() }

type contextProvider struct{ ctx context.Context }

func (provider contextProvider) RuntimeContext() context.Context { return provider.ctx }

type serviceStub struct {
	result       authentication.AuthenticationResult
	err          error
	authenticate func(context.Context, authentication.AuthenticationRequest) (authentication.AuthenticationResult, error)
	calls        atomic.Int32
}

func (service *serviceStub) Authenticate(ctx context.Context, request authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
	service.calls.Add(1)
	if service.authenticate != nil {
		return service.authenticate(ctx, request)
	}
	return service.result, service.err
}

type handoffStub struct {
	dispatch func(connection.AuthenticatedContext) (bool, error)
	calls    atomic.Int32
}

func (handoff *handoffStub) DispatchAuthenticated(authenticated connection.AuthenticatedContext) (bool, error) {
	handoff.calls.Add(1)
	if handoff.dispatch != nil {
		return handoff.dispatch(authenticated)
	}
	return true, nil
}

func closingHandoff(status websocket.StatusCode) *handoffStub {
	return &handoffStub{dispatch: func(authenticated connection.AuthenticatedContext) (bool, error) {
		connectionContext := authenticated.ConnectionContext()
		defer connectionContext.Cancel()
		defer connectionContext.Connection().CloseNow()
		_ = connectionContext.Connection().Close(status, "")
		return true, nil
	}}
}

func allowedResult() authentication.AuthenticationResult {
	return authentication.AuthenticationResult{
		Success: true,
		Principal: &authentication.Principal{
			ID:            "authenticated-client",
			Name:          "authenticated-client",
			Authenticated: true,
		},
	}
}

func newHandshakeServer(t *testing.T, admission *admissionStub, service authentication.Service, handoff connection.AuthenticatedDispatcher) *httptest.Server {
	return newHandshakeServerWithReporter(t, admission, service, handoff, nil)
}

func newHandshakeServerWithReporter(
	t *testing.T,
	admission *admissionStub,
	service authentication.Service,
	handoff connection.AuthenticatedDispatcher,
	reportError func(error),
) *httptest.Server {
	t.Helper()
	runtimeContext, cancelRuntime := context.WithCancel(context.Background())
	t.Cleanup(cancelRuntime)
	handler, err := NewHandlerWithTerminalErrorReporter(
		admission,
		contextProvider{ctx: runtimeContext},
		service,
		handoff,
		reportError,
	)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

type terminalErrorRecorder struct {
	mu     sync.Mutex
	errors []error
	notify chan struct{}
}

func (recorder *terminalErrorRecorder) Report(err error) {
	recorder.mu.Lock()
	recorder.errors = append(recorder.errors, err)
	if recorder.notify != nil {
		close(recorder.notify)
		recorder.notify = nil
	}
	recorder.mu.Unlock()
}

func (recorder *terminalErrorRecorder) Count() int {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return len(recorder.errors)
}

func (recorder *terminalErrorRecorder) Single(t *testing.T) error {
	t.Helper()
	recorder.mu.Lock()
	if len(recorder.errors) == 0 {
		recorder.notify = make(chan struct{})
		notify := recorder.notify
		recorder.mu.Unlock()
		waitDone(t, notify, "terminal error report")
		recorder.mu.Lock()
	}
	defer recorder.mu.Unlock()
	if len(recorder.errors) != 1 {
		t.Fatalf("reported errors = %d, want 1", len(recorder.errors))
	}
	return recorder.errors[0]
}

func successfulWebSocketDial(t *testing.T, serverURL string) (*websocket.Conn, *http.Response) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client, response, err := websocket.Dial(ctx, websocketURL(serverURL), nil)
	if err != nil {
		t.Fatalf("websocket.Dial() error = %v", err)
	}
	return client, response
}

func rejectedWebSocketDial(t *testing.T, serverURL string) *http.Response {
	return rejectedWebSocketDialWithHeader(t, serverURL, "", "")
}

func rejectedWebSocketDialWithHeader(t *testing.T, serverURL string, name string, value string) *http.Response {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	header := make(http.Header)
	if name != "" {
		header.Set(name, value)
	}
	client, response, err := websocket.Dial(ctx, websocketURL(serverURL), &websocket.DialOptions{HTTPHeader: header})
	if client != nil {
		client.CloseNow()
		t.Fatal("websocket.Dial() unexpectedly succeeded")
	}
	if err == nil || response == nil {
		t.Fatalf("websocket.Dial() = (%v, %v), want HTTP rejection", response, err)
	}
	t.Cleanup(func() { response.Body.Close() })
	return response
}

func assertSafeResponseBody(t *testing.T, response *http.Response, sensitive string) {
	t.Helper()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if strings.Contains(string(body), sensitive) {
		t.Fatal("response body contains sensitive input")
	}
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func websocketURL(serverURL string) string {
	return "ws" + strings.TrimPrefix(serverURL, "http")
}

func readClose(t *testing.T, client *websocket.Conn) websocket.StatusCode {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _, err := client.Read(ctx)
	return websocket.CloseStatus(err)
}

func receive[T any](t *testing.T, channel <-chan T, operation string) T {
	t.Helper()
	select {
	case value := <-channel:
		return value
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for %s", operation)
		var zero T
		return zero
	}
}

func waitDone(t *testing.T, done <-chan struct{}, operation string) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for %s", operation)
	}
}
