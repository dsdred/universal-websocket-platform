package handshake

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/connection"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnostics"
)

func diagnosticsHandler(t *testing.T, producer runtimediagnostics.Producer, service authentication.Service, handoff connection.AuthenticatedDispatcher, reporter func(error)) *Handler {
	t.Helper()
	handler, err := NewHandlerWithDiagnostics(openAdmission(), contextProvider{context.Background()}, service, handoff, reporter, producer)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func assertProviderEvent(t *testing.T, reader *runtimediagnostics.Reader, sequence uint64) {
	t.Helper()
	event, ok := reader.TryRead()
	if !ok || event != (runtimediagnostics.Event{SchemaVersion: 1, Code: runtimediagnostics.AuthenticationProviderFailed, Sequence: sequence}) {
		t.Fatalf("event = (%+v, %t), want Provider sequence %d", event, ok, sequence)
	}
}

func assertHandshakeMailboxEmpty(t *testing.T, reader *runtimediagnostics.Reader) {
	t.Helper()
	if event, ok := reader.TryRead(); ok || event != (runtimediagnostics.Event{}) {
		t.Fatalf("unexpected event (%+v, %t)", event, ok)
	}
}

func assertDiagnostics503(t *testing.T, response *httptest.ResponseRecorder, handoff *handoffStub) {
	t.Helper()
	if response.Code != http.StatusServiceUnavailable || response.Body.String() != "Service Unavailable\n" || response.Header().Get("Upgrade") != "" || handoff.calls.Load() != 0 {
		t.Fatalf("response/handoff changed: status=%d body=%q handoffs=%d", response.Code, response.Body.String(), handoff.calls.Load())
	}
}

// I04: error shape is not cancellation evidence when the actual context is active.
func TestHandshakeDiagnosticsOperationalFailure(t *testing.T) {
	for index, cause := range []error{errors.New("private dependency failure"), context.Canceled, context.DeadlineExceeded,
		fmt.Errorf("private wrapper: %w", context.Canceled), fmt.Errorf("private wrapper: %w", context.DeadlineExceeded)} {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			producer, reader := runtimediagnostics.New()
			handoff := &handoffStub{}
			handler := diagnosticsHandler(t, *producer, &serviceStub{err: cause}, handoff, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/?credential=private", nil))
			assertDiagnostics503(t, response, handoff)
			assertProviderEvent(t, reader, 1)
			assertHandshakeMailboxEmpty(t, reader)
		})
	}
	t.Run("ordered Providers one boundary", func(t *testing.T) {
		producer, reader := runtimediagnostics.New()
		var calls [3]int
		providers := []authentication.Provider{
			diagnosticProvider{authenticate: func(context.Context, authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
				calls[0]++
				return authentication.AuthenticationResult{}, nil
			}},
			diagnosticProvider{authenticate: func(context.Context, authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
				calls[1]++
				return authentication.AuthenticationResult{}, errors.New("private dependency")
			}},
			diagnosticProvider{authenticate: func(context.Context, authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
				calls[2]++
				return allowedResult(), nil
			}},
		}
		service, err := authentication.NewService(providers)
		if err != nil {
			t.Fatal(err)
		}
		handoff := &handoffStub{}
		response := httptest.NewRecorder()
		diagnosticsHandler(t, *producer, service, handoff, nil).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
		assertDiagnostics503(t, response, handoff)
		if calls != [3]int{1, 1, 0} {
			t.Fatalf("Provider evaluation order changed: %v", calls)
		}
		assertProviderEvent(t, reader, 1)
		assertHandshakeMailboxEmpty(t, reader)
	})
}

type diagnosticProvider struct {
	authenticate func(context.Context, authentication.AuthenticationRequest) (authentication.AuthenticationResult, error)
}

func (diagnosticProvider) Name() string { panic("Provider metadata inspected") }
func (diagnosticProvider) Type() runtimeconfig.AuthenticationProviderType {
	panic("Provider metadata inspected")
}
func (provider diagnosticProvider) Authenticate(ctx context.Context, request authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
	return provider.authenticate(ctx, request)
}

// I05: rejected credentials and invalid success do not become operational events.
func TestHandshakeDiagnosticsCredentialRejection(t *testing.T) {
	for index, result := range []authentication.AuthenticationResult{
		{}, {Success: true}, {Success: true, Principal: &authentication.Principal{ID: "invalid"}},
		{Success: true, Principal: &authentication.Principal{ID: "invalid", Authenticated: true, Anonymous: true}},
	} {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			producer, reader := runtimediagnostics.New()
			handoff := &handoffStub{}
			handler := diagnosticsHandler(t, *producer, &serviceStub{result: result}, handoff, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
			if response.Code != http.StatusUnauthorized || handoff.calls.Load() != 0 {
				t.Fatalf("rejection changed: %d, handoffs %d", response.Code, handoff.calls.Load())
			}
			assertHandshakeMailboxEmpty(t, reader)
		})
	}
}

// I06: cancellation is established before Authenticate returns, not by a timing sleep.
// CancelCause(ErrHandshakeTimeout) deterministically models the configured timeout cause;
// the existing configured-deadline regression separately covers the actual timer.
func TestHandshakeDiagnosticsCancellationPrecedence(t *testing.T) {
	for _, timeout := range []bool{false, true} {
		for _, lateError := range []bool{false, true} {
			t.Run(fmt.Sprintf("timeout=%t/error=%t", timeout, lateError), func(t *testing.T) {
				producer, reader := runtimediagnostics.New()
				ctx, cancel := context.WithCancelCause(context.Background())
				defer cancel(nil)
				entered, release := make(chan struct{}), make(chan struct{})
				service := &serviceStub{authenticate: func(ctx context.Context, _ authentication.AuthenticationRequest) (authentication.AuthenticationResult, error) {
					close(entered)
					<-release
					if ctx.Err() == nil {
						t.Error("test phase did not establish cancellation")
					}
					if lateError {
						return authentication.AuthenticationResult{}, errors.New("independent late dependency failure")
					}
					return allowedResult(), nil
				}}
				handoff := &handoffStub{}
				var reports []error
				handler := diagnosticsHandler(t, *producer, service, handoff, func(err error) { reports = append(reports, err) })
				response := httptest.NewRecorder()
				completed := make(chan struct{})
				go func() {
					handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx))
					close(completed)
				}()
				receive(t, entered, "Authentication entry")
				cause := context.Canceled
				if timeout {
					cause = ErrHandshakeTimeout
				}
				cancel(cause)
				close(release)
				receive(t, completed, "cancelled response")
				assertDiagnostics503(t, response, handoff)
				assertHandshakeMailboxEmpty(t, reader)
				if timeout {
					if len(reports) != 1 || !errors.Is(reports[0], ErrHandshakeTimeout) {
						t.Fatalf("timeout reporting changed: %v", reports)
					}
				} else if len(reports) != 0 {
					t.Fatalf("unexpected legacy reports: %v", reports)
				}
			})
		}
	}
	t.Run("later cancellation does not retract", func(t *testing.T) {
		producer, reader := runtimediagnostics.New()
		ctx, cancel := context.WithCancelCause(context.Background())
		defer cancel(nil)
		handoff := &handoffStub{}
		handler := diagnosticsHandler(t, *producer, &serviceStub{err: errors.New("operational failure")}, handoff, nil)
		response := &diagnosticResponsePhase{ResponseRecorder: httptest.NewRecorder(), onHeader: func() { cancel(ErrHandshakeTimeout) }}
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx))
		if ctx.Err() == nil {
			t.Fatal("after-sample response phase not reached")
		}
		assertDiagnostics503(t, response.ResponseRecorder, handoff)
		assertProviderEvent(t, reader, 1)
		assertHandshakeMailboxEmpty(t, reader)
		// Header occurs after the existing helper reads context.Cause; this fixture
		// does not claim to force the earlier active-sample/late-timeout helper window.
	})
}

type diagnosticResponsePhase struct {
	*httptest.ResponseRecorder
	onHeader func()
}

func (response *diagnosticResponsePhase) Header() http.Header {
	if response.onHeader != nil {
		once := response.onHeader
		response.onHeader = nil
		once()
	}
	return response.ResponseRecorder.Header()
}

type hostileDiagnosticError struct{}

func (hostileDiagnosticError) Error() string { panic("Error inspected") }
func (hostileDiagnosticError) Unwrap() error { panic("Unwrap inspected") }
func (hostileDiagnosticError) Is(error) bool { panic("Is inspected") }

// I07: no projection of raw error, credential, Principal, label or Provider metadata.
func TestHandshakeDiagnosticsDoNotInspectProviderError(t *testing.T) {
	producer, reader := runtimediagnostics.New()
	service := &serviceStub{err: hostileDiagnosticError{}, result: authentication.AuthenticationResult{
		ProviderName: "private Provider", FailureReason: "private credential", Metadata: map[string]string{"private": "value"},
	}}
	handoff := &handoffStub{}
	request := httptest.NewRequest(http.MethodGet, "/?credential=private", nil)
	request.Header.Set("Authorization", "private credential")
	response := httptest.NewRecorder()
	diagnosticsHandler(t, *producer, service, handoff, nil).ServeHTTP(response, request)
	assertDiagnostics503(t, response, handoff)
	assertProviderEvent(t, reader, 1)
	assertHandshakeMailboxEmpty(t, reader)
	shape := reflect.TypeOf(runtimediagnostics.Event{})
	if shape.NumField() != 3 || shape.Field(0).Name != "SchemaVersion" || shape.Field(1).Name != "Code" || shape.Field(2).Name != "Sequence" {
		t.Fatal("closed three-scalar schema changed")
	}
}

// I08: full loss and zero Producer preserve the generic rejection path.
func TestHandshakeDiagnosticsSaturationPreservesSourceBehavior(t *testing.T) {
	producer, reader := runtimediagnostics.New()
	for i := 0; i < 64; i++ {
		if !producer.Offer(runtimediagnostics.ListenerServeFailed) {
			t.Fatal("prefill failed")
		}
	}
	handoff := &handoffStub{}
	service := &serviceStub{err: errors.New("private failure")}
	handler := diagnosticsHandler(t, *producer, service, handoff, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	assertDiagnostics503(t, response, handoff)
	for i := uint64(1); i <= 64; i++ {
		event, ok := reader.TryRead()
		if !ok || event.Code != runtimediagnostics.ListenerServeFailed || event.Sequence != i {
			t.Fatalf("prefill FIFO changed: %+v, %t", event, ok)
		}
	}
	assertHandshakeMailboxEmpty(t, reader)
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	assertProviderEvent(t, reader, 65)
	zero := diagnosticsHandler(t, runtimediagnostics.Producer{}, service, handoff, nil)
	response = httptest.NewRecorder()
	zero.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	assertDiagnostics503(t, response, handoff)
	assertHandshakeMailboxEmpty(t, reader)
	if service.calls.Load() != 3 {
		t.Fatal("loss/disabled changed Authentication occurrence count")
	}
}

// I09: caller reassignment after construction cannot redirect an existing Handler.
func TestHandshakeDiagnosticsCaptureProducerValue(t *testing.T) {
	for _, zero := range []bool{false, true} {
		t.Run(fmt.Sprint(zero), func(t *testing.T) {
			first, reader1 := runtimediagnostics.New()
			second, reader2 := runtimediagnostics.New()
			caller := *first
			handler := diagnosticsHandler(t, caller, &serviceStub{err: errors.New("first")}, &handoffStub{}, nil)
			caller = *second
			if zero {
				caller = runtimediagnostics.Producer{}
			}
			handler2 := diagnosticsHandler(t, caller, &serviceStub{err: errors.New("second")}, &handoffStub{}, nil)
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
			assertProviderEvent(t, reader1, 1)
			assertHandshakeMailboxEmpty(t, reader2)
			handler2.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
			assertHandshakeMailboxEmpty(t, reader1)
			if zero {
				assertHandshakeMailboxEmpty(t, reader2)
			} else {
				assertProviderEvent(t, reader2, 1)
			}
		})
	}
}

// I10: one Handler safely shares only the Producer value among concurrent requests.
func TestHandshakeDiagnosticsConcurrentSourcesAndReaders(t *testing.T) {
	producer, reader := runtimediagnostics.New()
	_, isolated := runtimediagnostics.New()
	service := &serviceStub{err: errors.New("operational failure")}
	handoff := &handoffStub{}
	handler := diagnosticsHandler(t, *producer, service, handoff, nil)
	const occurrences = 128
	var done atomic.Bool
	var readers sync.WaitGroup
	observed := make(chan []runtimediagnostics.Event, 2)
	for range 2 {
		readers.Add(1)
		go func() {
			defer readers.Done()
			var events []runtimediagnostics.Event
			for {
				if event, ok := reader.TryRead(); ok {
					events = append(events, event)
				} else if done.Load() {
					observed <- events
					return
				} else {
					runtime.Gosched()
				}
			}
		}()
	}
	var requests sync.WaitGroup
	for range occurrences {
		requests.Add(1)
		go func() {
			defer requests.Done()
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
			if response.Code != http.StatusServiceUnavailable || response.Body.String() != "Service Unavailable\n" {
				t.Error("concurrent response changed")
			}
		}()
	}
	requests.Wait()
	done.Store(true)
	readers.Wait()
	close(observed)
	seen := make(map[uint64]bool)
	check := func(event runtimediagnostics.Event) {
		if event.SchemaVersion != 1 || event.Code != runtimediagnostics.AuthenticationProviderFailed || event.Sequence == 0 || seen[event.Sequence] {
			t.Fatalf("invalid or duplicate accepted event: %+v", event)
		}
		seen[event.Sequence] = true
	}
	for events := range observed {
		var previous uint64
		for _, event := range events {
			if event.Sequence <= previous {
				t.Fatal("per-reader FIFO changed")
			}
			previous = event.Sequence
			check(event)
		}
	}
	for {
		event, ok := reader.TryRead()
		if !ok {
			break
		}
		check(event)
	}
	if len(seen) > occurrences || service.calls.Load() != occurrences || handoff.calls.Load() != 0 {
		t.Fatal("source acceptance/call bound changed")
	}
	assertHandshakeMailboxEmpty(t, isolated)
}

// I11: valid Authentication, Upgrade and downstream handoff outcomes add no codes.
func TestHandshakeDiagnosticsDoNotAddOutcomeKinds(t *testing.T) {
	for _, failHandoff := range []bool{false, true} {
		t.Run(fmt.Sprint(failHandoff), func(t *testing.T) {
			producer, reader := runtimediagnostics.New()
			cause := errors.New("handoff failure")
			reports := &terminalErrorRecorder{}
			handoff := closingHandoff(websocket.StatusNormalClosure)
			if failHandoff {
				handoff = &handoffStub{dispatch: func(connection.AuthenticatedContext) (bool, error) { return false, cause }}
			}
			handler := diagnosticsHandler(t, *producer, &serviceStub{result: allowedResult()}, handoff, reports.Report)
			// A valid result with an invalid Upgrade request still adds no diagnostics.
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
			assertHandshakeMailboxEmpty(t, reader)
			server := httptest.NewServer(handler)
			defer server.Close()
			client, _, err := websocket.Dial(context.Background(), websocketURL(server.URL), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer client.CloseNow()
			_, _, _ = client.Read(context.Background())
			if handoff.calls.Load() != 1 {
				t.Fatal("handoff occurrence changed")
			}
			if failHandoff && !errors.Is(reports.Single(t), cause) {
				t.Fatal("legacy handoff cause lost")
			}
			assertHandshakeMailboxEmpty(t, reader)
		})
	}
}

func TestHandshakeDiagnosticsConstructorValidationOrder(t *testing.T) {
	producer, reader := runtimediagnostics.New()
	admission := openAdmission()
	provider := contextProvider{context.Background()}
	service := &serviceStub{}
	handoff := &handoffStub{}
	for index, want := range []error{ErrNilAdmissionCapability, ErrNilRuntimeContextProvider, ErrNilAuthenticationService, ErrNilSessionHandoff} {
		var a AdmissionCapability
		var r RuntimeContextProvider
		var s authentication.Service
		if index >= 1 {
			a = admission
		}
		if index >= 2 {
			r = provider
		}
		if index >= 3 {
			s = service
		}
		created, err := NewHandlerWithDiagnostics(a, r, s, nil, nil, *producer)
		if created != nil || !errors.Is(err, want) {
			t.Fatalf("validation precedence changed at %d: %v", index, err)
		}
	}
	created, err := NewHandlerWithDiagnostics(admission, provider, service, handoff, nil, *producer)
	if err != nil || created == nil {
		t.Fatalf("valid construction failed: %v", err)
	}
	assertHandshakeMailboxEmpty(t, reader)
}
