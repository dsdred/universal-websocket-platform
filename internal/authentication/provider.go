package authentication

import (
	"context"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

// AuthenticationRequest contains normalized, transport-neutral client input.
// The context passed to Provider.Authenticate carries cancellation and deadlines.
type AuthenticationRequest struct {
	Headers       map[string][]string
	Query         map[string][]string
	RemoteAddress string
	Transport     string
	Attributes    map[string]string
	Body          []byte
}

// Principal represents an authenticated or explicitly anonymous identity.
type Principal struct {
	ID                 string
	Name               string
	AuthenticationType runtimeconfig.AuthenticationProviderType
	Claims             map[string]string
	Roles              []string
	Attributes         map[string]string
	Anonymous          bool
	Authenticated      bool
	Metadata           map[string]string
}

// AuthenticationResult contains the transport-neutral outcome of Authentication.
type AuthenticationResult struct {
	Success       bool
	Principal     *Principal
	FailureReason string
	ProviderName  string
	ProviderType  runtimeconfig.AuthenticationProviderType
	Metadata      map[string]string
}

// Provider authenticates normalized input without depending on Runtime or transport details.
type Provider interface {
	Name() string
	Type() runtimeconfig.AuthenticationProviderType
	Authenticate(ctx context.Context, request AuthenticationRequest) (AuthenticationResult, error)
}
