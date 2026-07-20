package session

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/coder/websocket"

	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
	"github.com/dsdred/universal-websocket-platform/internal/message"
)

func TestPrepareProvisionalSessionCreatesCompletePreCommitUnit(t *testing.T) {
	var handled atomic.Int32
	handler := messageHandlerFunc(func(context.Context, message.Context) error {
		handled.Add(1)
		return nil
	})
	core, err := newSessionCore(
		validPrincipal(),
		"192.0.2.1:4321",
		fixedSessionID("prepared-session"),
		nil,
		handler,
	)
	if err != nil {
		t.Fatalf("newSessionCore() error = %v", err)
	}
	connection := &provisionalTestConnection{}

	prepared, err := prepareProvisionalSession(core, connection)
	if err != nil {
		t.Fatalf("prepareProvisionalSession() error = %v", err)
	}
	if prepared == nil || prepared.core == nil || prepared.session == nil || prepared.owner == nil {
		t.Fatalf("prepared unit = %#v, want complete unit", prepared)
	}
	if prepared.core != core || prepared.session.core != core {
		t.Fatal("prepared Session does not retain the exact supplied Core")
	}
	if prepared.session.ID() != "prepared-session" {
		t.Fatalf("prepared Session ID = %q, want prepared-session", prepared.session.ID())
	}
	if prepared.session.state != stateCreated || prepared.session.Running() {
		t.Fatalf("prepared Session state = %v, Running = %v, want Created", prepared.session.state, prepared.session.Running())
	}
	if prepared.owner.State() != executionowner.StatePreCommit {
		t.Fatalf("prepared Owner state = %v, want PreCommit", prepared.owner.State())
	}
	if prepared.owner.StopRequested() {
		t.Fatal("prepared Owner has an unexpected Stop intent")
	}
	if handled.Load() != 0 {
		t.Fatalf("Handler calls = %d, want 0", handled.Load())
	}
	connection.assertUnused(t)
}

func TestPrepareProvisionalSessionRejectsInvalidDependenciesWithoutSideEffects(t *testing.T) {
	core, err := newSessionCore(validPrincipal(), "", fixedSessionID("session-id"), nil, nil)
	if err != nil {
		t.Fatalf("newSessionCore() error = %v", err)
	}

	t.Run("nil Core", func(t *testing.T) {
		connection := &provisionalTestConnection{}
		prepared, err := prepareProvisionalSession(nil, connection)
		if prepared != nil || !errors.Is(err, errNilSessionCore) {
			t.Fatalf("prepareProvisionalSession() = (%#v, %v), want nil and %v", prepared, err, errNilSessionCore)
		}
		connection.assertUnused(t)
	})

	t.Run("nil transport", func(t *testing.T) {
		prepared, err := prepareProvisionalSession(core, nil)
		if prepared != nil || !errors.Is(err, ErrNilConnection) {
			t.Fatalf("prepareProvisionalSession() = (%#v, %v), want nil and %v", prepared, err, ErrNilConnection)
		}
	})
}

func TestPrepareProvisionalSessionFailureCannotCancelCallerContext(t *testing.T) {
	callerContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	prepared, err := prepareProvisionalSession(nil, &provisionalTestConnection{})
	if prepared != nil || !errors.Is(err, errNilSessionCore) {
		t.Fatalf("prepareProvisionalSession() = (%#v, %v), want nil and %v", prepared, err, errNilSessionCore)
	}
	select {
	case <-callerContext.Done():
		t.Fatal("preparation failure canceled caller-owned context")
	default:
	}
}

func TestSeparateProvisionalSessionsDoNotSharePreparedState(t *testing.T) {
	source := validPrincipal()
	firstCore, err := newSessionCore(source, "first", fixedSessionID("first"), nil, nil)
	if err != nil {
		t.Fatalf("first newSessionCore() error = %v", err)
	}
	secondCore, err := newSessionCore(source, "second", fixedSessionID("second"), nil, nil)
	if err != nil {
		t.Fatalf("second newSessionCore() error = %v", err)
	}

	first, err := prepareProvisionalSession(firstCore, &provisionalTestConnection{})
	if err != nil {
		t.Fatalf("first prepareProvisionalSession() error = %v", err)
	}
	second, err := prepareProvisionalSession(secondCore, &provisionalTestConnection{})
	if err != nil {
		t.Fatalf("second prepareProvisionalSession() error = %v", err)
	}

	if first == second || first.core == second.core || first.session == second.session || first.owner == second.owner {
		t.Fatal("separate preparations share identity-bearing objects")
	}
	first.core.principal.Claims["tenant"] = "first-mutated"
	first.core.principal.Roles[0] = "first-mutated"
	if second.core.principal.Claims["tenant"] != "alpha" || second.core.principal.Roles[0] != "admin" {
		t.Fatalf("second prepared Principal shares mutable state: %#v", second.core.principal)
	}
	if first.owner.State() != executionowner.StatePreCommit || second.owner.State() != executionowner.StatePreCommit {
		t.Fatalf("Owner states = (%v, %v), want PreCommit", first.owner.State(), second.owner.State())
	}
}

func TestExistingSessionConstructionRemainsCreatedAndInactive(t *testing.T) {
	connection := &provisionalTestConnection{}
	runtimeSession, err := newWithConnectionDependencies(
		connection,
		validPrincipal(),
		"192.0.2.1:4321",
		fixedSessionID("legacy-construction"),
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("newWithConnectionDependencies() error = %v", err)
	}
	if runtimeSession.state != stateCreated || runtimeSession.Running() {
		t.Fatalf("Session state = %v, Running = %v, want Created", runtimeSession.state, runtimeSession.Running())
	}
	connection.assertUnused(t)
}

type provisionalTestConnection struct {
	reads     atomic.Int32
	writes    atomic.Int32
	closes    atomic.Int32
	closeNows atomic.Int32
}

func (connection *provisionalTestConnection) Read(context.Context) (websocket.MessageType, []byte, error) {
	connection.reads.Add(1)
	return 0, nil, errors.New("unexpected Read")
}

func (connection *provisionalTestConnection) Write(context.Context, websocket.MessageType, []byte) error {
	connection.writes.Add(1)
	return errors.New("unexpected Write")
}

func (connection *provisionalTestConnection) Close(websocket.StatusCode, string) error {
	connection.closes.Add(1)
	return errors.New("unexpected Close")
}

func (connection *provisionalTestConnection) CloseNow() error {
	connection.closeNows.Add(1)
	return errors.New("unexpected CloseNow")
}

func (connection *provisionalTestConnection) assertUnused(t *testing.T) {
	t.Helper()
	if reads := connection.reads.Load(); reads != 0 {
		t.Fatalf("transport Read calls = %d, want 0", reads)
	}
	if writes := connection.writes.Load(); writes != 0 {
		t.Fatalf("transport Write calls = %d, want 0", writes)
	}
	if closes := connection.closes.Load(); closes != 0 {
		t.Fatalf("transport Close calls = %d, want 0", closes)
	}
	if closeNows := connection.closeNows.Load(); closeNows != 0 {
		t.Fatalf("transport CloseNow calls = %d, want 0", closeNows)
	}
}
