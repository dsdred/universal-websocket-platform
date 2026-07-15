package connection

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/authentication"
)

func TestAuthenticatedContextPrincipalIsDeepCopied(t *testing.T) {
	principal := authentication.Principal{
		ID:            "principal",
		Authenticated: true,
		Claims:        map[string]string{"claim": "original"},
		Roles:         []string{"reader"},
		Attributes:    map[string]string{"attribute": "original"},
		Metadata:      map[string]string{"metadata": "original"},
	}
	connectionContext := NewRuntimeContext(
		context.Background(),
		nil,
		httptest.NewRequest("GET", "http://example.test/ws", nil),
	)
	defer connectionContext.Cancel()

	authenticated := NewAuthenticatedContext(connectionContext, principal)
	principal.Claims["claim"] = "mutated"
	principal.Roles[0] = "mutated"
	principal.Attributes["attribute"] = "mutated"
	principal.Metadata["metadata"] = "mutated"

	first := authenticated.Principal()
	if first.Claims["claim"] != "original" || first.Roles[0] != "reader" ||
		first.Attributes["attribute"] != "original" || first.Metadata["metadata"] != "original" {
		t.Fatalf("stored Principal was mutated: %+v", first)
	}
	first.Claims["claim"] = "returned mutation"
	first.Roles[0] = "returned mutation"
	first.Attributes["attribute"] = "returned mutation"
	first.Metadata["metadata"] = "returned mutation"

	second := authenticated.Principal()
	if second.Claims["claim"] != "original" || second.Roles[0] != "reader" ||
		second.Attributes["attribute"] != "original" || second.Metadata["metadata"] != "original" {
		t.Fatalf("Principal() exposed internal mutable data: %+v", second)
	}
}
