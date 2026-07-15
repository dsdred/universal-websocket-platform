package connection

import "github.com/dsdred/universal-websocket-platform/internal/authentication"

// AuthenticatedDispatcher receives an authenticated connection and reports explicit ownership acceptance.
type AuthenticatedDispatcher interface {
	DispatchAuthenticated(AuthenticatedContext) (accepted bool, err error)
}

// AuthenticatedContext combines transport references with an authenticated Principal.
type AuthenticatedContext struct {
	connectionContext ConnectionContext
	principal         authentication.Principal
}

// NewAuthenticatedContext combines an upgraded connection with an independently copied Principal.
func NewAuthenticatedContext(
	connectionContext ConnectionContext,
	principal authentication.Principal,
) AuthenticatedContext {
	return AuthenticatedContext{
		connectionContext: connectionContext,
		principal:         clonePrincipal(principal),
	}
}

// ConnectionContext returns the original transport-only context.
func (authenticatedContext AuthenticatedContext) ConnectionContext() ConnectionContext {
	return authenticatedContext.connectionContext
}

// Principal returns an independent copy of the authenticated Principal.
func (authenticatedContext AuthenticatedContext) Principal() authentication.Principal {
	return clonePrincipal(authenticatedContext.principal)
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
