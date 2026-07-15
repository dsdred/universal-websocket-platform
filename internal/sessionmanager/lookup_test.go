package sessionmanager

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestManagerLookupExistingCommittedRegistration(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))

	view, found := manager.Lookup("session-1")

	assertLookupResult(t, view, found, "session-1", registrationID)
}

func TestRegistrationViewContainsRegisteredState(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))

	view, found := manager.Lookup("session-1")

	assertLookupResult(t, view, found, "session-1", registrationID)
	if got := view.State(); got != StateRegistered {
		t.Fatalf("RegistrationView.State() = %v, want StateRegistered", got)
	}
}

func TestManagerLookupUnknownSessionID(t *testing.T) {
	manager := New()
	mustCommit(t, mustReserve(t, manager, "session-1"))

	view, found := manager.Lookup("unknown")

	if found || view != (RegistrationView{}) {
		t.Fatalf("Lookup(unknown) = (%+v, %t), want zero View and false", view, found)
	}
}

func TestManagerLookupDoesNotExposeReservation(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")

	view, found := manager.Lookup("session-1")

	if found || view != (RegistrationView{}) {
		t.Fatalf("Lookup(reserved) = (%+v, %t), want zero View and false", view, found)
	}
	handle.Abort()
}

func TestManagerLookupDoesNotExposeCompletedRegistration(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() = false, want true")
	}

	view, found := manager.Lookup("session-1")

	if found || view != (RegistrationView{}) {
		t.Fatalf("Lookup(completed) = (%+v, %t), want zero View and false", view, found)
	}
}

func TestManagerLookupDoesNotExposeAbortedReservation(t *testing.T) {
	manager := New()
	handle := mustReserve(t, manager, "session-1")
	handle.Abort()

	view, found := manager.Lookup("session-1")

	if found || view != (RegistrationView{}) {
		t.Fatalf("Lookup(aborted) = (%+v, %t), want zero View and false", view, found)
	}
}

func TestRegistrationViewIsDetachedValue(t *testing.T) {
	viewType := reflect.TypeOf(RegistrationView{})
	for fieldIndex := range viewType.NumField() {
		if field := viewType.Field(fieldIndex); field.IsExported() {
			t.Fatalf("RegistrationView field %q is exported", field.Name)
		}
	}

	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
	view, found := manager.Lookup("session-1")
	assertLookupResult(t, view, found, "session-1", registrationID)

	view.sessionID = "changed"
	view.registrationID = RegistrationID{}
	view.state = 0

	fresh, freshFound := manager.Lookup("session-1")
	assertLookupResult(t, fresh, freshFound, "session-1", registrationID)
	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() after detached View mutation = false, want true")
	}
}

func TestManagerConcurrentLookup(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
	start := make(chan struct{})
	results := make(chan lookupResult, concurrentCalls)
	var waitGroup sync.WaitGroup

	for range concurrentCalls {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			view, found := manager.Lookup("session-1")
			results <- lookupResult{view: view, found: found}
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)

	for result := range results {
		assertLookupResult(t, result.view, result.found, "session-1", registrationID)
	}
}

func TestManagerConcurrentCommitAndLookup(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		handle := mustReserve(t, manager, "session-1")
		registrationID := reservationFromHandle(t, handle).registrationID
		start := make(chan struct{})
		commitResults := make(chan commitResult, 1)
		lookupResults := make(chan lookupResult, 1)

		go func() {
			<-start
			committedID, err := handle.Commit()
			commitResults <- commitResult{registrationID: committedID, err: err}
		}()
		go func() {
			<-start
			view, found := manager.Lookup("session-1")
			lookupResults <- lookupResult{view: view, found: found}
		}()
		close(start)

		commitResult := <-commitResults
		lookupResult := <-lookupResults
		if commitResult.err != nil || commitResult.registrationID != registrationID {
			t.Fatalf("iteration %d: Commit() = (%+v, %v)", iteration, commitResult.registrationID, commitResult.err)
		}
		if lookupResult.found {
			assertLookupResult(t, lookupResult.view, true, "session-1", registrationID)
		} else if lookupResult.view != (RegistrationView{}) {
			t.Fatalf("iteration %d: absent Lookup View = %+v, want zero", iteration, lookupResult.view)
		}
		finalView, finalFound := manager.Lookup("session-1")
		assertLookupResult(t, finalView, finalFound, "session-1", registrationID)
	}
}

func TestManagerConcurrentCompleteAndLookup(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
		start := make(chan struct{})
		completeResults := make(chan bool, 1)
		lookupResults := make(chan lookupResult, 1)

		go func() {
			<-start
			completeResults <- manager.Complete(registrationID)
		}()
		go func() {
			<-start
			view, found := manager.Lookup("session-1")
			lookupResults <- lookupResult{view: view, found: found}
		}()
		close(start)

		if completed := <-completeResults; !completed {
			t.Fatalf("iteration %d: Complete() = false, want true", iteration)
		}
		lookupResult := <-lookupResults
		if lookupResult.found {
			assertLookupResult(t, lookupResult.view, true, "session-1", registrationID)
		} else if lookupResult.view != (RegistrationView{}) {
			t.Fatalf("iteration %d: absent Lookup View = %+v, want zero", iteration, lookupResult.view)
		}
		if finalView, finalFound := manager.Lookup("session-1"); finalFound || finalView != (RegistrationView{}) {
			t.Fatalf("iteration %d: final Lookup() = (%+v, %t), want zero and false", iteration, finalView, finalFound)
		}
	}
}

func TestManagerConcurrentReserveAndLookupKeepsReservationInvisible(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		start := make(chan struct{})
		reserveResults := make(chan reserveResult, 1)
		lookupResults := make(chan lookupResult, 1)

		go func() {
			<-start
			handle, err := manager.Reserve("session-1")
			reserveResults <- reserveResult{handle: handle, err: err}
		}()
		go func() {
			<-start
			view, found := manager.Lookup("session-1")
			lookupResults <- lookupResult{view: view, found: found}
		}()
		close(start)

		reserveResult := <-reserveResults
		if reserveResult.err != nil {
			t.Fatalf("iteration %d: Reserve() error = %v", iteration, reserveResult.err)
		}
		lookupResult := <-lookupResults
		if lookupResult.found || lookupResult.view != (RegistrationView{}) {
			t.Fatalf("iteration %d: Lookup(reserved) = (%+v, %t), want zero and false", iteration, lookupResult.view, lookupResult.found)
		}
		reserveResult.handle.Abort()
	}
}

func TestManagerLookupAfterBeginShutdown(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
	manager.BeginShutdown()

	view, found := manager.Lookup("session-1")

	assertLookupResult(t, view, found, "session-1", registrationID)
	if got := manager.State(); got != StateClosing {
		t.Fatalf("State() = %v, want StateClosing", got)
	}
}

func TestManagerLookupAfterClosedReturnsAbsence(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))
	manager.BeginShutdown()
	if completed := manager.Complete(registrationID); !completed {
		t.Fatal("Complete() = false, want true")
	}

	view, found := manager.Lookup("session-1")

	if found || view != (RegistrationView{}) {
		t.Fatalf("Lookup() after Closed = (%+v, %t), want zero and false", view, found)
	}
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() = %v, want StateClosed", got)
	}
}

func TestManagerLookupDoesNotChangeLifecycle(t *testing.T) {
	manager := New()
	registrationID := mustCommit(t, mustReserve(t, manager, "session-1"))

	for call := range concurrentCalls {
		view, found := manager.Lookup("session-1")
		assertLookupResult(t, view, found, "session-1", registrationID)
		if got := manager.State(); got != StateOpen {
			t.Fatalf("call %d: State() = %v, want StateOpen", call, got)
		}
	}
}

type lookupResult struct {
	view  RegistrationView
	found bool
}

func assertLookupResult(
	t *testing.T,
	view RegistrationView,
	found bool,
	wantSessionID SessionID,
	wantRegistrationID RegistrationID,
) {
	t.Helper()
	if !found {
		t.Fatalf("Lookup(%q) found = false, want true", wantSessionID)
	}
	if got := view.SessionID(); got != wantSessionID {
		t.Fatalf("RegistrationView.SessionID() = %q, want %q", got, wantSessionID)
	}
	if got := view.RegistrationID(); got != wantRegistrationID {
		t.Fatalf("RegistrationView.RegistrationID() = %+v, want %+v", got, wantRegistrationID)
	}
	if got := view.State(); got != StateRegistered {
		t.Fatalf("RegistrationView.State() = %v, want StateRegistered", got)
	}
}

func ExampleManager_Lookup() {
	manager := New()
	reservation, _ := manager.Reserve("session-1")
	registrationID, _ := reservation.Commit()

	view, found := manager.Lookup("session-1")
	fmt.Println(found, view.SessionID(), view.RegistrationID() == registrationID, view.State() == StateRegistered)
	// Output: true session-1 true true
}
