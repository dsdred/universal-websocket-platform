package authentication

import (
	"context"
	"strings"

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

// Header returns the first value for name using a case-insensitive lookup.
func (request AuthenticationRequest) Header(name string) (string, bool) {
	if values, exists := request.Headers[name]; exists && len(values) > 0 {
		return values[0], true
	}
	for headerName, values := range request.Headers {
		if strings.EqualFold(headerName, name) && len(values) > 0 {
			return values[0], true
		}
	}
	return "", false
}

// Principal represents an authenticated or explicitly anonymous identity.
type Principal struct {
	ID                     string
	Name                   string
	AuthenticationType     runtimeconfig.AuthenticationProviderType
	AuthenticationProvider string
	AuthenticationMethod   runtimeconfig.AuthenticationProviderType
	Claims                 map[string]string
	Roles                  []string
	Attributes             map[string]string
	Anonymous              bool
	Authenticated          bool
	Metadata               map[string]string
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
