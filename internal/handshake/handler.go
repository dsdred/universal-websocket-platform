// Package handshake owns pre-Upgrade Authentication and the WebSocket ownership boundary.
package handshake

import (
	"context"
	"errors"
	"net/http"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/connection"
)

var (
	ErrNilAdmissionCapability    = errors.New("Handshake admission capability is nil")
	ErrNilRuntimeContextProvider = errors.New("Handshake Runtime context provider is nil")
	ErrNilAuthenticationService  = errors.New("Handshake Authentication Service is nil")
	ErrNilSessionHandoff         = errors.New("Handshake Session handoff is nil")
	// ErrHandshakeTimeout identifies expiration of the configured pre-Upgrade Handshake deadline.
	ErrHandshakeTimeout = errors.New("Handshake timeout")
)

// AdmissionCapability provides a live, read-only view of Host admission state.
type AdmissionCapability interface {
	CanAccept() bool
}

// RuntimeContextProvider provides the active Host-owned Runtime context without cancellation access.
type RuntimeContextProvider interface {
	RuntimeContext() context.Context
}

// Handler performs Authentication before committing a WebSocket Upgrade.
type Handler struct {
	admission      AdmissionCapability
	runtimeContext RuntimeContextProvider
	authentication authentication.Service
	handoff        connection.AuthenticatedDispatcher
	reportError    func(error)
}

// NewHandler creates the pre-Upgrade Handshake boundary from narrow dependencies.
func NewHandler(
	admission AdmissionCapability,
	runtimeContext RuntimeContextProvider,
	authenticationService authentication.Service,
	handoff connection.AuthenticatedDispatcher,
) (*Handler, error) {
	return NewHandlerWithTerminalErrorReporter(
		admission,
		runtimeContext,
		authenticationService,
		handoff,
		nil,
	)
}

// NewHandlerWithTerminalErrorReporter creates the Handshake boundary with a local terminal error callback.
// The callback is synchronous and must return promptly; a callback panic is isolated from Handshake cleanup.
func NewHandlerWithTerminalErrorReporter(
	admission AdmissionCapability,
	runtimeContext RuntimeContextProvider,
	authenticationService authentication.Service,
	handoff connection.AuthenticatedDispatcher,
	reportError func(error),
) (*Handler, error) {
	if admission == nil {
		return nil, ErrNilAdmissionCapability
	}
	if runtimeContext == nil {
		return nil, ErrNilRuntimeContextProvider
	}
	if authenticationService == nil {
		return nil, ErrNilAuthenticationService
	}
	if handoff == nil {
		return nil, ErrNilSessionHandoff
	}
	return &Handler{
		admission:      admission,
		runtimeContext: runtimeContext,
		authentication: authenticationService,
		handoff:        handoff,
		reportError:    reportError,
	}, nil
}

// ServeHTTP authenticates one request and upgrades only a valid one-use Allow decision.
func (handler *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if !handler.admission.CanAccept() {
		http.Error(response, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	runtimeContext := handler.runtimeContext.RuntimeContext()
	if runtimeContext == nil || runtimeContext.Err() != nil {
		http.Error(response, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	handshakeContext, cancelHandshake := context.WithCancel(request.Context())
	stopRuntimeCancellation := context.AfterFunc(runtimeContext, cancelHandshake)
	defer func() {
		stopRuntimeCancellation()
		cancelHandshake()
	}()

	result, err := handler.authentication.Authenticate(handshakeContext, authenticationRequest(request))
	if err != nil {
		handler.rejectCanceledHandshake(response, handshakeContext)
		return
	}
	if handshakeContext.Err() != nil {
		handler.rejectCanceledHandshake(response, handshakeContext)
		return
	}
	decision, ok := allowDecision(result)
	if !ok {
		http.Error(response, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// This live check is intentionally the last admission operation before websocket.Accept.
	if !handler.admission.CanAccept() {
		http.Error(response, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	// A Decision loses validity if request, Runtime, or configured timeout cancellation
	// wins before the Upgrade commit begins.
	if handshakeContext.Err() != nil {
		handler.rejectCanceledHandshake(response, handshakeContext)
		return
	}
	websocketConnection, err := websocket.Accept(response, request, nil)
	if err != nil {
		return
	}

	connectionContext := connection.NewRuntimeContext(
		runtimeContext,
		websocketConnection,
		request,
	)
	authenticatedContext := connection.NewAuthenticatedContext(connectionContext, decision.principal)
	accepted, handoffErr := handler.handoff.DispatchAuthenticated(authenticatedContext)
	if handoffErr != nil && !errors.Is(handoffErr, context.Canceled) {
		reportTerminalError(handler.reportError, sessionHandoffError{cause: handoffErr})
	}
	if !accepted {
		connectionContext.Cancel()
		_ = websocketConnection.Close(websocket.StatusInternalError, "")
		websocketConnection.CloseNow()
	}
}

func (handler *Handler) rejectCanceledHandshake(response http.ResponseWriter, ctx context.Context) {
	cause := context.Cause(ctx)
	if errors.Is(cause, ErrHandshakeTimeout) {
		reportTerminalError(handler.reportError, ErrHandshakeTimeout)
	}
	http.Error(response, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
}

type sessionHandoffError struct {
	cause error
}

func (err sessionHandoffError) Error() string {
	return "Session handoff failed"
}

func (err sessionHandoffError) Unwrap() error {
	return err.cause
}

func reportTerminalError(reporter func(error), err error) {
	if reporter == nil || err == nil {
		return
	}
	// Reporting is observer-only: a faulty consumer cannot change Handshake ownership or cleanup.
	func() {
		defer func() {
			_ = recover()
		}()
		reporter(err)
	}()
}

type decision struct {
	principal authentication.Principal
}

func allowDecision(result authentication.AuthenticationResult) (decision, bool) {
	if !result.Success || result.Principal == nil {
		return decision{}, false
	}
	principal := *result.Principal
	validAuthenticated := principal.Authenticated && !principal.Anonymous
	validAnonymous := principal.Anonymous && !principal.Authenticated && principal.ID == "anonymous"
	if !validAuthenticated && !validAnonymous {
		return decision{}, false
	}
	return decision{principal: principal}, true
}

func authenticationRequest(request *http.Request) authentication.AuthenticationRequest {
	return authentication.AuthenticationRequest{
		Headers:       cloneValues(request.Header),
		Query:         cloneValues(request.URL.Query()),
		RemoteAddress: request.RemoteAddr,
		Transport:     "websocket",
	}
}

func cloneValues(values map[string][]string) map[string][]string {
	if values == nil {
		return nil
	}
	result := make(map[string][]string, len(values))
	for name, items := range values {
		result[name] = append([]string(nil), items...)
	}
	return result
}
