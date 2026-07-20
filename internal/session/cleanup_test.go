package session

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
	"github.com/dsdred/universal-websocket-platform/internal/message"
)

func TestCleanupFromCreatedStopsBeforeConfirmedCancellation(t *testing.T) {
	connection := newCleanupTestConnection()
	cancellation := newCleanupTestCancellation()
	cancellation.stopReturned = &connection.closeNowReturned
	var handled atomic.Int32
	prepared := newPreparedCleanupSession(t, connection, cancellation.dependency(), messageHandlerFunc(
		func(context.Context, message.Context) error {
			handled.Add(1)
			return nil
		},
	))

	result := prepared.cleanup.run(context.Background())

	assertCleanupResult(t, result, transportCleanupSucceeded, cancellationConfirmed, cleanupPanicNone)
	if connection.closeCalls.Load() != 1 || connection.closeNowCalls.Load() != 1 {
		t.Fatalf("transport close calls = (%d, %d), want (1, 1)", connection.closeCalls.Load(), connection.closeNowCalls.Load())
	}
	if cancellation.calls.Load() != 1 {
		t.Fatalf("cancellation calls = %d, want 1", cancellation.calls.Load())
	}
	if !connection.closeNowReturned.Load() || cancellation.beforeStopReturned.Load() {
		t.Fatal("cancellation ran before Session Stop completed")
	}
	if prepared.session.state != stateStopped {
		t.Fatalf("Session state = %v, want Stopped", prepared.session.state)
	}
	if handled.Load() != 0 {
		t.Fatalf("Handler calls = %d, want 0", handled.Load())
	}
	assertPreparedOwnerDormant(t, prepared)
}

func TestCleanupAfterRunWaitsForReadLoopBeforeCancellation(t *testing.T) {
	connection := newCleanupTestConnection()
	connection.blockRead = true
	connection.releaseReadOnCloseNow = false
	cancellation := newCleanupTestCancellation()
	prepared := newPreparedCleanupSession(t, connection, cancellation.dependency(), nil)
	if err := prepared.session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	runResult := make(chan error, 1)
	var goroutines sync.WaitGroup
	goroutines.Add(1)
	go func() {
		defer goroutines.Done()
		runResult <- prepared.session.Run(context.Background())
	}()
	defer func() {
		connection.releaseAll()
		waitCleanupGroup(t, &goroutines, "Run and Cleanup goroutines")
	}()
	waitCleanupSignal(t, connection.readEntered, "read-loop entry")

	cleanupResult := make(chan cleanupAcknowledgement, 1)
	goroutines.Add(1)
	go func() {
		defer goroutines.Done()
		cleanupResult <- prepared.cleanup.run(context.Background())
	}()
	waitCleanupSignal(t, connection.closeEntered, "transport Close")
	if cancellation.calls.Load() != 0 {
		t.Fatalf("cancellation calls while Stop is in progress = %d, want 0", cancellation.calls.Load())
	}
	connection.releaseClose()
	connection.releaseRead()

	result := waitCleanupResult(t, cleanupResult, "Cleanup after Run")
	assertCleanupResult(t, result, transportCleanupSucceeded, cancellationConfirmed, cleanupPanicNone)
	if cancellation.calls.Load() != 1 {
		t.Fatalf("cancellation calls = %d, want 1", cancellation.calls.Load())
	}
	select {
	case <-runResult:
	case <-time.After(time.Second):
		t.Fatal("Run() did not return after Cleanup")
	}
}

func TestCleanupCategorizesTransportFailureAndStillConfirmsCancellation(t *testing.T) {
	wantErr := errors.New("close failed")
	connection := newCleanupTestConnection()
	connection.closeErr = wantErr
	cancellation := newCleanupTestCancellation()
	prepared := newPreparedCleanupSession(t, connection, cancellation.dependency(), nil)

	result := prepared.cleanup.run(context.Background())

	assertCleanupResult(t, result, transportCleanupFailed, cancellationConfirmed, cleanupPanicNone)
	if cancellation.calls.Load() != 1 {
		t.Fatalf("cancellation calls = %d, want 1", cancellation.calls.Load())
	}
	if reflect.TypeOf(result).Field(0).Type.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		t.Fatal("Cleanup acknowledgement retained a raw error")
	}
}

func TestCleanupRecoversStopPanicAndCannotSkipCancellation(t *testing.T) {
	connection := newCleanupTestConnection()
	connection.panicClose = true
	cancellation := newCleanupTestCancellation()
	prepared := newPreparedCleanupSession(t, connection, cancellation.dependency(), nil)

	result := prepared.cleanup.run(context.Background())

	assertCleanupResult(t, result, transportCleanupPanicked, cancellationConfirmed, cleanupPanicStop)
	if cancellation.calls.Load() != 1 {
		t.Fatalf("cancellation calls = %d, want 1", cancellation.calls.Load())
	}
}

func TestConcurrentCleanupWaiterReturnsAfterPrimaryStopPanic(t *testing.T) {
	connection := newCleanupTestConnection()
	connection.blockClose = true
	connection.panicClose = true
	cancellation := newCleanupTestCancellation()
	prepared := newPreparedCleanupSession(t, connection, cancellation.dependency(), nil)

	results := make(chan cleanupAcknowledgement, 2)
	var goroutines sync.WaitGroup
	goroutines.Add(1)
	go func() {
		defer goroutines.Done()
		results <- prepared.cleanup.run(context.Background())
	}()
	defer func() {
		connection.releaseAll()
		waitCleanupGroup(t, &goroutines, "panic Cleanup callers")
	}()
	waitCleanupSignal(t, connection.closeEntered, "primary Cleanup Stop")

	waiterStarted := make(chan struct{})
	goroutines.Add(1)
	go func() {
		defer goroutines.Done()
		close(waiterStarted)
		results <- prepared.cleanup.run(context.Background())
	}()
	waitCleanupSignal(t, waiterStarted, "concurrent Cleanup waiter")
	select {
	case result := <-results:
		t.Fatalf("Cleanup returned before blocked primary panic: %#v", result)
	default:
	}

	connection.releaseClose()
	waitCleanupGroup(t, &goroutines, "panic Cleanup callers")
	close(results)
	want := cleanupAcknowledgement{
		transport:    transportCleanupPanicked,
		cancellation: cancellationConfirmed,
		panics:       cleanupPanicStop,
	}
	for result := range results {
		if result != want {
			t.Fatalf("Cleanup result = %#v, want %#v", result, want)
		}
	}
	if connection.closeCalls.Load() != 1 {
		t.Fatalf("transport Close calls = %d, want 1", connection.closeCalls.Load())
	}
	if cancellation.calls.Load() != 1 {
		t.Fatalf("cancellation calls = %d, want 1", cancellation.calls.Load())
	}
}

func TestCleanupCancellationPanicReportsObservedState(t *testing.T) {
	tests := []struct {
		name          string
		effective     bool
		wantOutcome   cancellationOutcome
		alreadyClosed bool
	}{
		{name: "panic before cancellation", wantOutcome: cancellationAnomaly},
		{name: "effective before panic", effective: true, wantOutcome: cancellationConfirmed},
		{name: "already canceled", wantOutcome: cancellationConfirmed, alreadyClosed: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			connection := newCleanupTestConnection()
			cancellation := newCleanupTestCancellation()
			if test.alreadyClosed {
				cancellation.closeDone()
			}
			cancellation.panicOnCancel = true
			cancellation.effectiveBeforePanic = test.effective
			prepared := newPreparedCleanupSession(t, connection, cancellation.dependency(), nil)

			result := prepared.cleanup.run(context.Background())

			assertCleanupResult(
				t,
				result,
				transportCleanupSucceeded,
				test.wantOutcome,
				cleanupPanicCancellation,
			)
			if cancellation.calls.Load() != 1 {
				t.Fatalf("cancellation calls = %d, want 1", cancellation.calls.Load())
			}
		})
	}
}

func TestCleanupCancellationWithoutObservableEffectIsAnomaly(t *testing.T) {
	connection := newCleanupTestConnection()
	cancellation := newCleanupTestCancellation()
	cancellation.noEffect = true
	prepared := newPreparedCleanupSession(t, connection, cancellation.dependency(), nil)

	result := prepared.cleanup.run(context.Background())

	assertCleanupResult(t, result, transportCleanupSucceeded, cancellationAnomaly, cleanupPanicNone)
	if cancellation.calls.Load() != 1 {
		t.Fatalf("cancellation calls = %d, want 1", cancellation.calls.Load())
	}
}

func TestRepeatedAndConcurrentCleanupShareOneExecutionAndResult(t *testing.T) {
	connection := newCleanupTestConnection()
	connection.blockClose = true
	cancellation := newCleanupTestCancellation()
	prepared := newPreparedCleanupSession(t, connection, cancellation.dependency(), nil)

	const callers = 32
	results := make(chan cleanupAcknowledgement, callers)
	var callersDone sync.WaitGroup
	callersDone.Add(1)
	go func() {
		defer callersDone.Done()
		results <- prepared.cleanup.run(context.Background())
	}()
	waitCleanupSignal(t, connection.closeEntered, "primary Cleanup Stop")

	if !prepared.cleanup.mu.TryLock() {
		t.Fatal("Cleanup state lock is held while Session Stop is blocked")
	}
	prepared.cleanup.mu.Unlock()

	begin := make(chan struct{})
	for range callers - 1 {
		callersDone.Add(1)
		go func() {
			defer callersDone.Done()
			<-begin
			results <- prepared.cleanup.run(context.Background())
		}()
	}
	defer func() {
		connection.releaseAll()
		waitCleanupGroup(t, &callersDone, "concurrent Cleanup callers")
	}()
	close(begin)
	if cancellation.calls.Load() != 0 {
		t.Fatalf("cancellation calls while Stop is blocked = %d, want 0", cancellation.calls.Load())
	}
	select {
	case result := <-results:
		t.Fatalf("Cleanup returned a partial acknowledgement while Stop was blocked: %#v", result)
	default:
	}

	connection.releaseClose()
	waitCleanupGroup(t, &callersDone, "concurrent Cleanup callers")
	close(results)
	want := cleanupAcknowledgement{
		transport:    transportCleanupSucceeded,
		cancellation: cancellationConfirmed,
	}
	for result := range results {
		if result != want {
			t.Fatalf("concurrent Cleanup result = %#v, want %#v", result, want)
		}
	}
	if connection.closeCalls.Load() != 1 || connection.closeNowCalls.Load() != 1 {
		t.Fatalf("transport close calls = (%d, %d), want (1, 1)", connection.closeCalls.Load(), connection.closeNowCalls.Load())
	}
	if cancellation.calls.Load() != 1 {
		t.Fatalf("cancellation calls = %d, want 1", cancellation.calls.Load())
	}
	if repeated := prepared.cleanup.run(context.Background()); repeated != want {
		t.Fatalf("repeated Cleanup result = %#v, want %#v", repeated, want)
	}
	if connection.closeCalls.Load() != 1 || cancellation.calls.Load() != 1 {
		t.Fatalf("repeated Cleanup changed effective calls: Close=%d Cancel=%d", connection.closeCalls.Load(), cancellation.calls.Load())
	}
}

func TestCleanupWithCanceledCooperativeContextDoesNotClaimEarlyCompletion(t *testing.T) {
	connection := newCleanupTestConnection()
	connection.blockClose = true
	cancellation := newCleanupTestCancellation()
	prepared := newPreparedCleanupSession(t, connection, cancellation.dependency(), nil)
	cleanupContext, cancelCleanup := context.WithCancel(context.Background())
	cancelCleanup()

	resultChannel := make(chan cleanupAcknowledgement, 1)
	var goroutines sync.WaitGroup
	goroutines.Add(1)
	go func() {
		defer goroutines.Done()
		resultChannel <- prepared.cleanup.run(cleanupContext)
	}()
	defer func() {
		connection.releaseAll()
		waitCleanupGroup(t, &goroutines, "canceled-context Cleanup")
	}()
	waitCleanupSignal(t, connection.closeEntered, "Stop with canceled cleanup context")
	if cancellation.calls.Load() != 0 {
		t.Fatalf("cancellation calls while Stop is blocked = %d, want 0", cancellation.calls.Load())
	}
	select {
	case result := <-resultChannel:
		t.Fatalf("Cleanup returned before Stop completed: %#v", result)
	default:
	}
	connection.releaseClose()
	result := waitCleanupResult(t, resultChannel, "Cleanup with canceled context")
	assertCleanupResult(t, result, transportCleanupSucceeded, cancellationConfirmed, cleanupPanicNone)
}

func TestSeparatePreparedUnitsDoNotShareCleanupState(t *testing.T) {
	firstConnection := newCleanupTestConnection()
	firstCancellation := newCleanupTestCancellation()
	first := newPreparedCleanupSession(t, firstConnection, firstCancellation.dependency(), nil)
	secondConnection := newCleanupTestConnection()
	secondCancellation := newCleanupTestCancellation()
	second := newPreparedCleanupSession(t, secondConnection, secondCancellation.dependency(), nil)

	firstResult := first.cleanup.run(context.Background())
	assertCleanupResult(t, firstResult, transportCleanupSucceeded, cancellationConfirmed, cleanupPanicNone)
	if firstCancellation.calls.Load() != 1 || secondCancellation.calls.Load() != 0 {
		t.Fatalf("cancellation calls after first Cleanup = (%d, %d), want (1, 0)", firstCancellation.calls.Load(), secondCancellation.calls.Load())
	}
	if firstConnection.closeCalls.Load() != 1 || secondConnection.closeCalls.Load() != 0 {
		t.Fatalf("Close calls after first Cleanup = (%d, %d), want (1, 0)", firstConnection.closeCalls.Load(), secondConnection.closeCalls.Load())
	}
	if second.cleanup.started || second.session.state != stateCreated {
		t.Fatalf("second prepared unit changed before its Cleanup: started=%v state=%v", second.cleanup.started, second.session.state)
	}
	assertPreparedOwnerDormant(t, first)
	assertPreparedOwnerDormant(t, second)

	secondResult := second.cleanup.run(context.Background())
	assertCleanupResult(t, secondResult, transportCleanupSucceeded, cancellationConfirmed, cleanupPanicNone)
	if firstCancellation.calls.Load() != 1 || secondCancellation.calls.Load() != 1 {
		t.Fatalf("final cancellation calls = (%d, %d), want (1, 1)", firstCancellation.calls.Load(), secondCancellation.calls.Load())
	}
}

func TestCleanupAcknowledgementIsPrivateDetachedScalarState(t *testing.T) {
	resultType := reflect.TypeOf(cleanupAcknowledgement{})
	for index := 0; index < resultType.NumField(); index++ {
		field := resultType.Field(index)
		if field.IsExported() {
			t.Fatalf("Cleanup acknowledgement field %q is exported", field.Name)
		}
		if field.Type.Kind() != reflect.Uint8 {
			t.Fatalf("Cleanup acknowledgement field %q has mutable or non-scalar type %s", field.Name, field.Type)
		}
	}

	preparedType := reflect.TypeOf(provisionalSession{})
	for index := 0; index < preparedType.NumField(); index++ {
		if field := preparedType.Field(index); field.IsExported() {
			t.Fatalf("provisional Session field %q is exported", field.Name)
		}
	}
}

func newPreparedCleanupSession(
	t *testing.T,
	connection websocketConnection,
	cancellation cancellationDependency,
	handler message.Handler,
) *provisionalSession {
	t.Helper()
	core, err := newSessionCore(validPrincipal(), "", fixedSessionID("cleanup-session"), nil, handler)
	if err != nil {
		t.Fatalf("newSessionCore() error = %v", err)
	}
	prepared, err := prepareProvisionalSession(core, connection, cancellation)
	if err != nil {
		t.Fatalf("prepareProvisionalSession() error = %v", err)
	}
	return prepared
}

func assertCleanupResult(
	t *testing.T,
	result cleanupAcknowledgement,
	wantTransport transportCleanupOutcome,
	wantCancellation cancellationOutcome,
	wantPanics cleanupPanicCategory,
) {
	t.Helper()
	if result.transport != wantTransport || result.cancellation != wantCancellation || result.panics != wantPanics {
		t.Fatalf(
			"Cleanup result = (%d, %d, %d), want (%d, %d, %d)",
			result.transport,
			result.cancellation,
			result.panics,
			wantTransport,
			wantCancellation,
			wantPanics,
		)
	}
}

func assertPreparedOwnerDormant(t *testing.T, prepared *provisionalSession) {
	t.Helper()
	if prepared.owner.State() != executionowner.StatePreCommit {
		t.Fatalf("Owner state = %v, want PreCommit", prepared.owner.State())
	}
	if prepared.owner.StopRequested() {
		t.Fatal("Cleanup published unexpected Owner Stop intent")
	}
}

type cleanupTestConnection struct {
	closeCalls            atomic.Int32
	closeNowCalls         atomic.Int32
	closeNowReturned      atomic.Bool
	closeErr              error
	panicClose            bool
	blockClose            bool
	blockRead             bool
	closeEntered          chan struct{}
	closeRelease          chan struct{}
	readEntered           chan struct{}
	readRelease           chan struct{}
	releaseReadOnCloseNow bool
	closeEnteredOnce      sync.Once
	closeReleaseOnce      sync.Once
	readEnteredOnce       sync.Once
	readReleaseOnce       sync.Once
}

func newCleanupTestConnection() *cleanupTestConnection {
	return &cleanupTestConnection{
		closeEntered:          make(chan struct{}),
		closeRelease:          make(chan struct{}),
		readEntered:           make(chan struct{}),
		readRelease:           make(chan struct{}),
		releaseReadOnCloseNow: true,
	}
}

func (connection *cleanupTestConnection) Read(context.Context) (websocket.MessageType, []byte, error) {
	if connection.blockRead {
		connection.readEnteredOnce.Do(func() { close(connection.readEntered) })
		<-connection.readRelease
	}
	return 0, nil, errors.New("read stopped")
}

func (*cleanupTestConnection) Write(context.Context, websocket.MessageType, []byte) error {
	return errors.New("unexpected Write")
}

func (connection *cleanupTestConnection) Close(websocket.StatusCode, string) error {
	connection.closeCalls.Add(1)
	connection.closeEnteredOnce.Do(func() { close(connection.closeEntered) })
	if connection.blockClose {
		<-connection.closeRelease
	}
	if connection.panicClose {
		panic("test transport panic")
	}
	return connection.closeErr
}

func (connection *cleanupTestConnection) CloseNow() error {
	connection.closeNowCalls.Add(1)
	if connection.releaseReadOnCloseNow {
		connection.releaseRead()
	}
	connection.closeNowReturned.Store(true)
	return nil
}

func (connection *cleanupTestConnection) releaseClose() {
	connection.closeReleaseOnce.Do(func() { close(connection.closeRelease) })
}

func (connection *cleanupTestConnection) releaseAll() {
	connection.releaseClose()
	connection.releaseRead()
}

func (connection *cleanupTestConnection) releaseRead() {
	connection.readReleaseOnce.Do(func() { close(connection.readRelease) })
}

type cleanupTestCancellation struct {
	done                 chan struct{}
	calls                atomic.Int32
	closed               atomic.Bool
	beforeStopReturned   atomic.Bool
	stopReturned         *atomic.Bool
	noEffect             bool
	panicOnCancel        bool
	effectiveBeforePanic bool
}

func newCleanupTestCancellation() *cleanupTestCancellation {
	return &cleanupTestCancellation{done: make(chan struct{})}
}

func (cancellation *cleanupTestCancellation) dependency() cancellationDependency {
	return cancellationDependency{
		done: cancellation.done,
		cancel: func() {
			cancellation.calls.Add(1)
			if cancellation.stopReturned != nil && !cancellation.stopReturned.Load() {
				cancellation.beforeStopReturned.Store(true)
			}
			if cancellation.effectiveBeforePanic || (!cancellation.noEffect && !cancellation.panicOnCancel) {
				cancellation.closeDone()
			}
			if cancellation.panicOnCancel {
				panic("test cancellation panic")
			}
		},
	}
}

func (cancellation *cleanupTestCancellation) closeDone() {
	if cancellation.closed.CompareAndSwap(false, true) {
		close(cancellation.done)
	}
}

func waitCleanupSignal(t *testing.T, signal <-chan struct{}, operation string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for %s", operation)
	}
}

func waitCleanupResult(
	t *testing.T,
	result <-chan cleanupAcknowledgement,
	operation string,
) cleanupAcknowledgement {
	t.Helper()
	select {
	case acknowledgement := <-result:
		return acknowledgement
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for %s", operation)
		return cleanupAcknowledgement{}
	}
}

func waitCleanupGroup(t *testing.T, group *sync.WaitGroup, description string) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		group.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Errorf("timeout joining %s", description)
	}
}
