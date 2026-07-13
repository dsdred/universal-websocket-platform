package connection

import (
	"errors"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/authentication"
)

var (
	// ErrNilAuthenticationService indicates that AuthenticationDispatcher has no Service.
	ErrNilAuthenticationService = errors.New("authentication service is nil")
	// ErrNilNextDispatcher indicates that AuthenticationDispatcher has no downstream Dispatcher.
	ErrNilNextDispatcher = errors.New("authenticated dispatcher is nil")
	// ErrInvalidAuthenticationResult indicates an inconsistent successful Authentication result.
	ErrInvalidAuthenticationResult = errors.New("invalid authentication result")
)

// AuthenticatedDispatcher receives an authenticated connection and its Principal.
type AuthenticatedDispatcher interface {
	DispatchAuthenticated(AuthenticatedContext) error
}

// AuthenticatedContext combines transport references with an authenticated Principal.
type AuthenticatedContext struct {
	connectionContext ConnectionContext
	principal         authentication.Principal
}

// ConnectionContext returns the original transport-only context.
func (authenticatedContext AuthenticatedContext) ConnectionContext() ConnectionContext {
	return authenticatedContext.connectionContext
}

// Principal returns an independent copy of the authenticated Principal.
func (authenticatedContext AuthenticatedContext) Principal() authentication.Principal {
	return clonePrincipal(authenticatedContext.principal)
}

// AuthenticationDispatcher authenticates upgraded connections before downstream dispatch.
type AuthenticationDispatcher struct {
	service authentication.Service
	next    AuthenticatedDispatcher
}

// NewAuthenticationDispatcher creates an immutable Authentication Dispatcher pipeline stage.
func NewAuthenticationDispatcher(
	service authentication.Service,
	next AuthenticatedDispatcher,
) (*AuthenticationDispatcher, error) {
	if service == nil {
		return nil, ErrNilAuthenticationService
	}
	if next == nil {
		return nil, ErrNilNextDispatcher
	}
	return &AuthenticationDispatcher{service: service, next: next}, nil
}

// Dispatch authenticates a transport context and transfers successful connections downstream.
func (dispatcher *AuthenticationDispatcher) Dispatch(connectionContext ConnectionContext) error {
	ctx := connectionContext.Context()
	if err := ctx.Err(); err != nil {
		closeOwnedConnection(connectionContext, websocket.StatusInternalError)
		return authenticationDispatchError{cause: err}
	}

	request := authenticationRequest(connectionContext)
	result, err := dispatcher.service.Authenticate(ctx, request)
	if err != nil {
		closeOwnedConnection(connectionContext, websocket.StatusInternalError)
		return authenticationDispatchError{cause: err}
	}
	if !result.Success {
		closeOwnedConnection(connectionContext, websocket.StatusPolicyViolation)
		return nil
	}
	if result.Principal == nil || !result.Principal.Authenticated || result.Principal.Anonymous {
		closeOwnedConnection(connectionContext, websocket.StatusInternalError)
		return ErrInvalidAuthenticationResult
	}

	authenticatedContext := AuthenticatedContext{
		connectionContext: connectionContext,
		principal:         clonePrincipal(*result.Principal),
	}
	return dispatcher.next.DispatchAuthenticated(authenticatedContext)
}

func authenticationRequest(connectionContext ConnectionContext) authentication.AuthenticationRequest {
	request := connectionContext.Request()
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

func clonePrincipal(principal authentication.Principal) authentication.Principal {
	principal.Claims = cloneStrings(principal.Claims)
	principal.Roles = append([]string(nil), principal.Roles...)
	principal.Attributes = cloneStrings(principal.Attributes)
	principal.Metadata = cloneStrings(principal.Metadata)
	return principal
}

func cloneStrings(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	result := make(map[string]string, len(values))
	for name, value := range values {
		result[name] = value
	}
	return result
}

func closeOwnedConnection(connectionContext ConnectionContext, status websocket.StatusCode) {
	connection := connectionContext.Connection()
	if connection == nil {
		return
	}
	defer connection.CloseNow()
	_ = connection.Close(status, "")
}

type authenticationDispatchError struct {
	cause error
}

func (authenticationError authenticationDispatchError) Error() string {
	return "authentication dispatch failed"
}

func (authenticationError authenticationDispatchError) Unwrap() error {
	return authenticationError.cause
}
