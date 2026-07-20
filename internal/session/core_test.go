package session

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/coder/websocket"

	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/message"
)

func TestSessionCoreExistsBeforeTransportAttachment(t *testing.T) {
	generated := 0
	handler := message.NewEchoHandler()
	core, err := newSessionCore(
		validPrincipal(),
		" 192.0.2.1:4321 ",
		func() (string, error) {
			generated++
			return "core-session-id", nil
		},
		nil,
		handler,
	)
	if err != nil {
		t.Fatalf("newSessionCore() error = %v", err)
	}
	if core.id != "core-session-id" || core.id == "" {
		t.Fatalf("Core ID = %q, want core-session-id", core.id)
	}
	if core.remoteAddress != "192.0.2.1:4321" {
		t.Fatalf("Core RemoteAddress = %q", core.remoteAddress)
	}
	if core.createdAt.IsZero() {
		t.Fatal("Core CreatedAt is zero")
	}
	if core.handler != handler {
		t.Fatal("Core did not retain the configured Handler")
	}
	if generated != 1 {
		t.Fatalf("ID generator calls = %d, want 1", generated)
	}

	connection := coreTestConnection{}
	runtimeSession, err := newSessionFromCore(core, connection)
	if err != nil {
		t.Fatalf("newSessionFromCore() error = %v", err)
	}
	if runtimeSession.ID() != core.id {
		t.Fatalf("Session ID = %q, want Core ID %q", runtimeSession.ID(), core.id)
	}
	if runtimeSession.core != core {
		t.Fatal("Session does not retain the prepared Core")
	}
	if generated != 1 {
		t.Fatalf("ID generator calls after attachment = %d, want 1", generated)
	}
	if got := runtimeSession.Principal(); !reflect.DeepEqual(got, core.principal) {
		t.Fatalf("Session Principal = %#v, want Core Principal %#v", got, core.principal)
	}
}

func TestSessionCoresOwnIndependentPrincipalState(t *testing.T) {
	source := validPrincipal()
	first, err := newSessionCore(source, "", fixedSessionID("first"), nil, nil)
	if err != nil {
		t.Fatalf("first newSessionCore() error = %v", err)
	}
	second, err := newSessionCore(source, "", fixedSessionID("second"), nil, nil)
	if err != nil {
		t.Fatalf("second newSessionCore() error = %v", err)
	}

	source.Claims["tenant"] = "source-mutated"
	source.Roles[0] = "source-mutated"
	first.principal.Claims["tenant"] = "first-mutated"
	first.principal.Roles[0] = "first-mutated"

	if second.principal.Claims["tenant"] != "alpha" || second.principal.Roles[0] != "admin" {
		t.Fatalf("second Core Principal shares mutable state: %#v", second.principal)
	}
	if first.id == second.id || first.id == "" || second.id == "" {
		t.Fatalf("Core IDs = (%q, %q), want separate non-empty identities", first.id, second.id)
	}
}

func TestSessionCoreContainsNoTransportOwnership(t *testing.T) {
	coreType := reflect.TypeOf(sessionCore{})
	connectionType := reflect.TypeOf((*websocketConnection)(nil)).Elem()
	websocketType := reflect.TypeOf((*websocket.Conn)(nil))

	for index := 0; index < coreType.NumField(); index++ {
		field := coreType.Field(index)
		fieldName := strings.ToLower(field.Name)
		if field.Type == connectionType || field.Type == websocketType ||
			field.Type.Implements(connectionType) || reflect.PointerTo(field.Type).Implements(connectionType) ||
			strings.Contains(fieldName, "connection") || strings.Contains(fieldName, "websocket") ||
			strings.Contains(fieldName, "transport") || strings.Contains(fieldName, "context") ||
			strings.Contains(fieldName, "cancel") {
			t.Fatalf("sessionCore contains transport field %q (%s)", field.Name, field.Type)
		}
	}
}

func TestSessionCoreRejectsInvalidConstruction(t *testing.T) {
	tests := []struct {
		name      string
		principal authentication.Principal
		generate  idGenerator
		wantErr   error
	}{
		{name: "invalid Principal", principal: authentication.Principal{}, generate: fixedSessionID("id"), wantErr: ErrInvalidPrincipal},
		{name: "nil ID generator", principal: validPrincipal(), wantErr: errNilIDGenerator},
		{name: "empty generated ID", principal: validPrincipal(), generate: fixedSessionID(""), wantErr: errInvalidGeneratedSessionID},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			core, err := newSessionCore(test.principal, "", test.generate, nil, nil)
			if core != nil || !errors.Is(err, test.wantErr) {
				t.Fatalf("newSessionCore() = (%v, %v), want nil and %v", core, err, test.wantErr)
			}
		})
	}
}

func TestSessionFormationRejectsInvalidDependencies(t *testing.T) {
	connection := coreTestConnection{}
	core, err := newSessionCore(validPrincipal(), "", fixedSessionID("session-id"), nil, nil)
	if err != nil {
		t.Fatalf("newSessionCore() error = %v", err)
	}

	if runtimeSession, err := newSessionFromCore(nil, connection); runtimeSession != nil || !errors.Is(err, errNilSessionCore) {
		t.Fatalf("newSessionFromCore(nil Core) = (%v, %v)", runtimeSession, err)
	}
	if runtimeSession, err := newSessionFromCore(core, nil); runtimeSession != nil || !errors.Is(err, ErrNilConnection) {
		t.Fatalf("newSessionFromCore(nil connection) = (%v, %v)", runtimeSession, err)
	}
}

func fixedSessionID(id string) idGenerator {
	return func() (string, error) { return id, nil }
}

type coreTestConnection struct{}

func (coreTestConnection) Read(context.Context) (websocket.MessageType, []byte, error) {
	return 0, nil, errors.New("unexpected Read")
}

func (coreTestConnection) Write(context.Context, websocket.MessageType, []byte) error {
	return errors.New("unexpected Write")
}

func (coreTestConnection) Close(websocket.StatusCode, string) error {
	return errors.New("unexpected Close")
}

func (coreTestConnection) CloseNow() error {
	return errors.New("unexpected CloseNow")
}
