package sessionmanager

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestManagerShutdownSnapshotEmptyManager(t *testing.T) {
	manager := New()

	snapshot := manager.BeginShutdown()

	assertSnapshotCount(t, snapshot, 0)
	if got := manager.State(); got != StateClosing {
		t.Fatalf("State() = %v, want StateClosing", got)
	}
	if err := manager.Wait(context.Background()); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
}

func TestManagerShutdownSnapshotContainsCommittedRegistration(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))

	snapshot := manager.BeginShutdown()

	assertSnapshotRegistration(t, snapshot, "session-1", registrationID)
	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() = false, want true")
	}
}

func TestShutdownRegistrationExposesIdentityAndStopOnly(t *testing.T) {
	registrationType := reflect.TypeOf(ShutdownRegistration{})
	if got := registrationType.NumField(); got != 3 {
		t.Fatalf("ShutdownRegistration field count = %d, want 3", got)
	}
	wantFieldTypes := []reflect.Type{
		reflect.TypeOf(SessionID("")),
		reflect.TypeOf(RegistrationID{}),
		reflect.TypeOf((*StopPublicationBinding)(nil)).Elem(),
	}
	for fieldIndex, wantType := range wantFieldTypes {
		field := registrationType.Field(fieldIndex)
		if field.IsExported() {
			t.Fatalf("ShutdownRegistration field %q is exported", field.Name)
		}
		if field.Type != wantType {
			t.Fatalf("ShutdownRegistration field %q type = %v, want %v", field.Name, field.Type, wantType)
		}
	}

	wantMethods := map[string]struct{}{
		"RegistrationID": {},
		"RequestStop":    {},
		"SessionID":      {},
	}
	if got := registrationType.NumMethod(); got != len(wantMethods) {
		t.Fatalf("ShutdownRegistration exported method count = %d, want %d", got, len(wantMethods))
	}
	for methodIndex := range registrationType.NumMethod() {
		method := registrationType.Method(methodIndex)
		if _, allowed := wantMethods[method.Name]; !allowed {
			t.Fatalf("ShutdownRegistration exposes unexpected method %q", method.Name)
		}
	}
}

func TestManagerShutdownSnapshotExcludesReservation(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")

	snapshot := manager.BeginShutdown()

	assertSnapshotCount(t, snapshot, 0)
	if registrationID, err := commitTestReservation(handle); registrationID != (CommitResult{}) || !errors.Is(err, ErrManagerNotOpen) {
		t.Fatalf("Commit() = (%+v, %v), want zero ID and ErrManagerNotOpen", registrationID, err)
	}
	handle.AbortUnlessCommitted()
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() = %v, want StateClosed", got)
	}
}

func TestManagerCommitBeforeBeginShutdownAppearsInSnapshot(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))

	snapshot := manager.BeginShutdown()

	assertSnapshotRegistration(t, snapshot, "session-1", registrationID)
	manager.Complete(registrationID)
}

func TestManagerCommitAfterBeginShutdownDoesNotAppearInSnapshot(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	snapshot := manager.BeginShutdown()

	registrationID, err := commitTestReservation(handle)

	if registrationID != (CommitResult{}) || !errors.Is(err, ErrManagerNotOpen) {
		t.Fatalf("Commit() = (%+v, %v), want zero ID and ErrManagerNotOpen", registrationID, err)
	}
	assertSnapshotCount(t, snapshot, 0)
	assertSnapshotCount(t, manager.BeginShutdown(), 0)
	handle.AbortUnlessCommitted()
}

func TestManagerShutdownSnapshotIsImmutableDetachedValue(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
	snapshot := manager.BeginShutdown()
	registrations := snapshot.Registrations()

	registrations[0] = ShutdownRegistration{}
	snapshot.registrations[0] = ShutdownRegistration{}

	fresh := manager.BeginShutdown()
	assertSnapshotRegistration(t, fresh, "session-1", registrationID)
	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() = false, want true")
	}
}

func TestManagerCompleteDoesNotChangeShutdownSnapshot(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
	initial := manager.BeginShutdown()

	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() = false, want true")
	}
	repeated := manager.BeginShutdown()

	assertSnapshotRegistration(t, initial, "session-1", registrationID)
	assertSnapshotRegistration(t, repeated, "session-1", registrationID)
	if !reflect.DeepEqual(initial.Registrations(), repeated.Registrations()) {
		t.Fatalf("Snapshot changed after Complete: initial=%+v repeated=%+v", initial, repeated)
	}
}

func TestManagerConcurrentCommitAndBeginShutdownSnapshotLinearization(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		handle := mustReserve(t, manager, "session-1")
		registrationID := reservationFromHandle(t, handle).registrationID
		start := make(chan struct{})
		commitResults := make(chan commitResult, 1)
		snapshots := make(chan ShutdownSnapshot, 1)

		go func() {
			<-start
			committed, err := commitTestReservation(handle)
			commitResults <- commitResult{registrationID: committed.RegistrationID(), err: err}
		}()
		go func() {
			<-start
			snapshots <- manager.BeginShutdown()
		}()
		close(start)

		commitResult := <-commitResults
		snapshot := <-snapshots
		switch {
		case commitResult.err == nil:
			if commitResult.registrationID != registrationID {
				t.Fatalf("iteration %d: Commit ID = %+v, want %+v", iteration, commitResult.registrationID, registrationID)
			}
			assertSnapshotRegistration(t, snapshot, "session-1", registrationID)
			if completed := manager.Complete(registrationID); !completed {
				t.Fatalf("iteration %d: Complete() = false, want true", iteration)
			}
		case errors.Is(commitResult.err, ErrManagerNotOpen):
			if commitResult.registrationID != (RegistrationID{}) {
				t.Fatalf("iteration %d: rejected Commit ID = %+v, want zero", iteration, commitResult.registrationID)
			}
			assertSnapshotCount(t, snapshot, 0)
			handle.AbortUnlessCommitted()
		default:
			t.Fatalf("iteration %d: Commit() error = %v", iteration, commitResult.err)
		}
	}
}

func TestManagerRepeatedBeginShutdownReturnsSameSnapshot(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
	initial := manager.BeginShutdown()

	for call := range concurrentCalls {
		repeated := manager.BeginShutdown()
		if !reflect.DeepEqual(initial.Registrations(), repeated.Registrations()) {
			t.Fatalf("call %d: repeated Snapshot = %+v, want %+v", call, repeated, initial)
		}
	}
	manager.Complete(registrationID)
}

func TestManagerShutdownSnapshotDoesNotChangeAccounting(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
	assertAccountingCount(t, manager, 1)

	snapshot := manager.BeginShutdown()

	assertSnapshotRegistration(t, snapshot, "session-1", registrationID)
	assertAccountingCount(t, manager, 1)
	if got := manager.State(); got != StateClosing {
		t.Fatalf("State() = %v, want StateClosing", got)
	}
	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() = false, want true")
	}
	assertAccountingCount(t, manager, 0)
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() = %v, want StateClosed", got)
	}
}

func assertSnapshotRegistration(
	t *testing.T,
	snapshot ShutdownSnapshot,
	sessionID SessionID,
	registrationID RegistrationID,
) {
	t.Helper()
	registrations := snapshot.Registrations()
	if len(registrations) != 1 {
		t.Fatalf("Snapshot Registration count = %d, want 1", len(registrations))
	}
	registration := registrations[0]
	if got := registration.SessionID(); got != sessionID {
		t.Fatalf("ShutdownRegistration.SessionID() = %q, want %q", got, sessionID)
	}
	if got := registration.RegistrationID(); got != registrationID {
		t.Fatalf("ShutdownRegistration.RegistrationID() = %+v, want %+v", got, registrationID)
	}
}

func assertSnapshotCount(t *testing.T, snapshot ShutdownSnapshot, want int) {
	t.Helper()
	if got := len(snapshot.Registrations()); got != want {
		t.Fatalf("Snapshot Registration count = %d, want %d", got, want)
	}
}
