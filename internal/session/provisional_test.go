package session

import (
	"context"
	"errors"
	"reflect"
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
	cancellation := newProvisionalTestCancellation()

	prepared, err := prepareProvisionalSession(core, connection, cancellation.dependency())
	if err != nil {
		t.Fatalf("prepareProvisionalSession() error = %v", err)
	}
	if prepared == nil || prepared.core == nil || prepared.session == nil || prepared.owner == nil ||
		prepared.cleanup == nil || prepared.lifecycle == nil {
		t.Fatalf("prepared unit = %#v, want complete unit", prepared)
	}
	if prepared.core != core || prepared.session.core != core {
		t.Fatal("prepared Session does not retain the exact supplied Core")
	}
	if prepared.session.connection != connection {
		t.Fatal("prepared Session is not bound to the supplied transport")
	}
	if prepared.cleanup.session != prepared.session {
		t.Fatal("prepared Cleanup is not bound to the Session in the same aggregate")
	}
	if prepared.lifecycle.session != prepared.session || prepared.lifecycle.cleanup != prepared.cleanup {
		t.Fatal("prepared execution lifecycle is not bound to the Session and Cleanup in the same aggregate")
	}
	if prepared.cleanup.cancellation == nil || prepared.cleanup.cancellation.done != cancellation.done {
		t.Fatal("prepared Cleanup is not bound to the supplied cancellation observation")
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
	if prepared.cleanup.started || prepared.cleanup.cancellation.invoked.Load() {
		t.Fatal("prepared Cleanup or cancellation is unexpectedly active")
	}
	if handled.Load() != 0 {
		t.Fatalf("Handler calls = %d, want 0", handled.Load())
	}
	if cancellation.calls.Load() != 0 {
		t.Fatalf("cancellation calls = %d, want 0", cancellation.calls.Load())
	}
	connection.assertUnused(t)
}

func TestProvisionalSessionMembershipIsStructurallyStable(t *testing.T) {
	core, err := newSessionCore(validPrincipal(), "", fixedSessionID("stable"), nil, nil)
	if err != nil {
		t.Fatalf("newSessionCore() error = %v", err)
	}
	connection := &provisionalTestConnection{}
	cancellation := newProvisionalTestCancellation()
	prepared, err := prepareProvisionalSession(core, connection, cancellation.dependency())
	if err != nil {
		t.Fatalf("prepareProvisionalSession() error = %v", err)
	}

	wantCore := prepared.core
	wantSession := prepared.session
	wantOwner := prepared.owner
	wantCleanup := prepared.cleanup
	wantLifecycle := prepared.lifecycle
	wantCancellation := prepared.cleanup.cancellation
	for range 10 {
		if prepared.core != wantCore || prepared.session != wantSession || prepared.owner != wantOwner ||
			prepared.cleanup != wantCleanup || prepared.lifecycle != wantLifecycle ||
			prepared.cleanup.cancellation != wantCancellation {
			t.Fatal("provisional Session component identity changed after construction")
		}
	}

	bundleType := reflect.TypeOf(provisionalSession{})
	for index := 0; index < bundleType.NumField(); index++ {
		field := bundleType.Field(index)
		if field.IsExported() {
			t.Fatalf("provisional Session field %q is exported", field.Name)
		}
		if field.Type.Kind() == reflect.Map || field.Type.Kind() == reflect.Slice {
			t.Fatalf("provisional Session field %q is a mutable component registry", field.Name)
		}
	}
	if bundleType.NumMethod() != 0 {
		t.Fatalf("provisional Session exported method count = %d, want 0", bundleType.NumMethod())
	}
	if cancellation.calls.Load() != 0 {
		t.Fatalf("cancellation calls = %d, want 0", cancellation.calls.Load())
	}
	connection.assertUnused(t)
}

func TestPrepareProvisionalSessionRejectsInvalidDependenciesWithoutSideEffects(t *testing.T) {
	var handled atomic.Int32
	core, err := newSessionCore(
		validPrincipal(),
		"",
		fixedSessionID("session-id"),
		nil,
		messageHandlerFunc(func(context.Context, message.Context) error {
			handled.Add(1)
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("newSessionCore() error = %v", err)
	}

	t.Run("nil Core", func(t *testing.T) {
		connection := &provisionalTestConnection{}
		cancellation := newProvisionalTestCancellation()
		prepared, err := prepareProvisionalSession(nil, connection, cancellation.dependency())
		if prepared != nil || !errors.Is(err, errNilSessionCore) {
			t.Fatalf("prepareProvisionalSession() = (%#v, %v), want nil and %v", prepared, err, errNilSessionCore)
		}
		if cancellation.calls.Load() != 0 {
			t.Fatalf("cancellation calls = %d, want 0", cancellation.calls.Load())
		}
		connection.assertUnused(t)
	})

	t.Run("nil transport", func(t *testing.T) {
		cancellation := newProvisionalTestCancellation()
		prepared, err := prepareProvisionalSession(core, nil, cancellation.dependency())
		if prepared != nil || !errors.Is(err, ErrNilConnection) {
			t.Fatalf("prepareProvisionalSession() = (%#v, %v), want nil and %v", prepared, err, ErrNilConnection)
		}
		if cancellation.calls.Load() != 0 {
			t.Fatalf("cancellation calls = %d, want 0", cancellation.calls.Load())
		}
	})

	t.Run("nil cancellation observation", func(t *testing.T) {
		var calls atomic.Int32
		connection := &provisionalTestConnection{}
		prepared, err := prepareProvisionalSession(core, connection, cancellationDependency{
			cancel: func() { calls.Add(1) },
		})
		if prepared != nil || !errors.Is(err, errNilCancellationObservation) {
			t.Fatalf("prepareProvisionalSession() = (%#v, %v), want nil and %v", prepared, err, errNilCancellationObservation)
		}
		if calls.Load() != 0 {
			t.Fatalf("cancellation calls = %d, want 0", calls.Load())
		}
		connection.assertUnused(t)
	})

	t.Run("nil cancellation operation", func(t *testing.T) {
		connection := &provisionalTestConnection{}
		prepared, err := prepareProvisionalSession(core, connection, cancellationDependency{
			done: make(chan struct{}),
		})
		if prepared != nil || !errors.Is(err, errNilCancellationOperation) {
			t.Fatalf("prepareProvisionalSession() = (%#v, %v), want nil and %v", prepared, err, errNilCancellationOperation)
		}
		connection.assertUnused(t)
	})

	if handled.Load() != 0 {
		t.Fatalf("Handler calls after failed preparations = %d, want 0", handled.Load())
	}
}

func TestPrepareProvisionalSessionFailureCannotCancelCallerContext(t *testing.T) {
	callerContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	var cancellationCalls atomic.Int32
	prepared, err := prepareProvisionalSession(nil, &provisionalTestConnection{}, cancellationDependency{
		done: callerContext.Done(),
		cancel: func() {
			cancellationCalls.Add(1)
			cancel()
		},
	})
	if prepared != nil || !errors.Is(err, errNilSessionCore) {
		t.Fatalf("prepareProvisionalSession() = (%#v, %v), want nil and %v", prepared, err, errNilSessionCore)
	}
	select {
	case <-callerContext.Done():
		t.Fatal("preparation failure canceled caller-owned context")
	default:
	}
	if cancellationCalls.Load() != 0 {
		t.Fatalf("cancellation calls = %d, want 0", cancellationCalls.Load())
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

	firstConnection := &provisionalTestConnection{}
	firstCancellation := newProvisionalTestCancellation()
	first, err := prepareProvisionalSession(firstCore, firstConnection, firstCancellation.dependency())
	if err != nil {
		t.Fatalf("first prepareProvisionalSession() error = %v", err)
	}
	secondConnection := &provisionalTestConnection{}
	secondCancellation := newProvisionalTestCancellation()
	second, err := prepareProvisionalSession(secondCore, secondConnection, secondCancellation.dependency())
	if err != nil {
		t.Fatalf("second prepareProvisionalSession() error = %v", err)
	}

	if first == second || first.core == second.core || first.session == second.session || first.owner == second.owner ||
		first.cleanup == second.cleanup || first.lifecycle == second.lifecycle ||
		first.cleanup.cancellation == second.cleanup.cancellation {
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
	if !first.owner.RequestStop() || second.owner.StopRequested() {
		t.Fatal("Stop intent in first prepared Owner affected the second Owner")
	}

	first.cleanup.run(context.Background())
	if firstCancellation.calls.Load() != 1 || secondCancellation.calls.Load() != 0 {
		t.Fatalf("cancellation calls = (%d, %d), want (1, 0)", firstCancellation.calls.Load(), secondCancellation.calls.Load())
	}
	if first.session.state != stateStopped || second.session.state != stateCreated {
		t.Fatalf("Session states = (%v, %v), want (Stopped, Created)", first.session.state, second.session.state)
	}
	secondConnection.assertUnused(t)
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

type provisionalTestCancellation struct {
	done      chan struct{}
	calls     atomic.Int32
	cancelled atomic.Bool
}

func newProvisionalTestCancellation() *provisionalTestCancellation {
	return &provisionalTestCancellation{done: make(chan struct{})}
}

func (cancellation *provisionalTestCancellation) dependency() cancellationDependency {
	return cancellationDependency{
		done: cancellation.done,
		cancel: func() {
			cancellation.calls.Add(1)
			if cancellation.cancelled.CompareAndSwap(false, true) {
				close(cancellation.done)
			}
		},
	}
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
