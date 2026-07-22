package session

import (
	"context"
	"errors"
	"net"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/dsdred/universal-websocket-platform/internal/connection"
	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
	"github.com/dsdred/universal-websocket-platform/internal/sessionmanager"
)

func TestTransactionalDispatcherCommitsBeforeDormantExecutionStarts(t *testing.T) {
	manager := sessionmanager.New()
	root := context.Background()
	transport := newTransactionalTestConnection()
	observer := &transactionalTestObserver{observed: make(chan executionowner.TerminalResult, 1)}
	var prepared *provisionalSession
	dispatcher, err := newTransactionalDispatcher(
		manager,
		transactionalRuntimeContextProvider{ctx: root},
		nil,
		observer,
		fixedSessionID("transactional-session"),
		func(core *sessionCore, _ websocketConnection, cancellation cancellationDependency) (*provisionalSession, error) {
			result, prepareErr := prepareProvisionalSession(core, transport, cancellation)
			prepared = result
			return result, prepareErr
		},
	)
	if err != nil {
		t.Fatalf("newTransactionalDispatcher() error = %v", err)
	}
	authenticated, cancelConnection := transactionalAuthenticatedContext(t, root)
	defer cancelConnection()

	accepted, err := dispatcher.DispatchAuthenticated(authenticated)
	if err != nil || !accepted {
		t.Fatalf("DispatchAuthenticated() = (%t, %v), want true and nil", accepted, err)
	}
	if prepared == nil {
		t.Fatal("provisional Session was not prepared")
	}
	awaitSignal(t, transport.readStarted, "Session Run did not start after Commit")
	if prepared.owner.State() != executionowner.StateRunning {
		t.Fatalf("Owner state = %v, want Running", prepared.owner.State())
	}

	snapshot := manager.BeginShutdown()
	registrations := snapshot.Registrations()
	if len(registrations) != 1 || registrations[0].SessionID() != "transactional-session" {
		t.Fatalf("Shutdown registrations = %+v, want committed transactional-session", registrations)
	}
	if !registrations[0].RequestStop() {
		t.Fatal("first committed Stop request was not accepted")
	}

	result := awaitTerminalResult(t, observer.observed)
	if result.PrimaryCause() != executionowner.TerminationCauseExplicitStop {
		t.Fatalf("PrimaryCause() = %v, want ExplicitStop", result.PrimaryCause())
	}
	if transport.reads.Load() != 1 || transport.closes.Load() != 1 || transport.closeNows.Load() != 1 {
		t.Fatalf("transport calls = (Read %d, Close %d, CloseNow %d), want (1, 1, 1)", transport.reads.Load(), transport.closes.Load(), transport.closeNows.Load())
	}

	waitContext, cancelWait := context.WithTimeout(context.Background(), time.Second)
	defer cancelWait()
	if err := manager.Wait(waitContext); err != nil {
		t.Fatalf("Manager.Wait() error = %v", err)
	}
	if prepared.owner.State() != executionowner.StateTerminal {
		t.Fatalf("Owner state after Manager.Wait = %v, want Terminal", prepared.owner.State())
	}
	select {
	case extra := <-observer.observed:
		t.Fatalf("Observer received a second Terminal Result: %+v", extra)
	default:
	}
}

func TestTransactionalDispatcherCommitFailureRetiresDormantPathWithoutExecution(t *testing.T) {
	wantErr := errors.New("Commit failed")
	reservation := &transactionalFailingReservation{commitErr: wantErr}
	manager := &transactionalReservationManager{reservation: reservation}
	root := context.Background()
	transport := newTransactionalTestConnection()
	observer := &transactionalTestObserver{observed: make(chan executionowner.TerminalResult, 1)}
	var prepared *provisionalSession
	dispatcher, err := newTransactionalDispatcher(
		manager,
		transactionalRuntimeContextProvider{ctx: root},
		nil,
		observer,
		fixedSessionID("not-committed-session"),
		func(core *sessionCore, _ websocketConnection, cancellation cancellationDependency) (*provisionalSession, error) {
			result, prepareErr := prepareProvisionalSession(core, transport, cancellation)
			prepared = result
			return result, prepareErr
		},
	)
	if err != nil {
		t.Fatalf("newTransactionalDispatcher() error = %v", err)
	}
	authenticated, cancelConnection := transactionalAuthenticatedContext(t, root)
	defer cancelConnection()

	accepted, err := dispatcher.DispatchAuthenticated(authenticated)
	if accepted || !errors.Is(err, wantErr) {
		t.Fatalf("DispatchAuthenticated() = (%t, %v), want false and Commit error", accepted, err)
	}
	if reservation.commitCalls.Load() != 1 || reservation.abortCalls.Load() != 1 {
		t.Fatalf("Reservation calls = (Commit %d, Abort %d), want (1, 1)", reservation.commitCalls.Load(), reservation.abortCalls.Load())
	}
	if prepared == nil {
		t.Fatal("provisional Session was not prepared")
	}
	if prepared.owner.State() != executionowner.StatePreCommit {
		t.Fatalf("prepared Owner state = %v, want PreCommit", prepared.owner.State())
	}
	transport.assertUnused(t)
	select {
	case result := <-observer.observed:
		t.Fatalf("Observer received Terminal Result before Commit: %+v", result)
	default:
	}
}

func TestTransactionalDispatcherRecoversPreCommitPanicAndJoinsDormantPath(t *testing.T) {
	reservation := &transactionalFailingReservation{panicOnCommit: true}
	manager := &transactionalReservationManager{reservation: reservation}
	transport := newTransactionalTestConnection()
	observer := &transactionalTestObserver{observed: make(chan executionowner.TerminalResult, 1)}
	var prepared *provisionalSession
	dispatcher, err := newTransactionalDispatcher(
		manager,
		transactionalRuntimeContextProvider{ctx: context.Background()},
		nil,
		observer,
		fixedSessionID("panic-before-commit"),
		func(core *sessionCore, _ websocketConnection, cancellation cancellationDependency) (*provisionalSession, error) {
			result, prepareErr := prepareProvisionalSession(core, transport, cancellation)
			prepared = result
			return result, prepareErr
		},
	)
	if err != nil {
		t.Fatalf("newTransactionalDispatcher() error = %v", err)
	}
	authenticated, cancelConnection := transactionalAuthenticatedContext(t, context.Background())
	defer cancelConnection()

	accepted, err := dispatcher.DispatchAuthenticated(authenticated)
	if accepted || !errors.Is(err, ErrTransactionalDispatchPanic) {
		t.Fatalf("DispatchAuthenticated() = (%t, %v), want false and sanitized panic", accepted, err)
	}
	if prepared == nil || prepared.owner.State() != executionowner.StatePreCommit {
		t.Fatal("recovered pre-Commit panic changed dormant Owner lifecycle")
	}
	if reservation.abortCalls.Load() != 1 {
		t.Fatalf("Abort calls = %d, want 1", reservation.abortCalls.Load())
	}
	transport.assertUnused(t)
}

func TestTransactionalDispatcherCommitLosingToBeginShutdownPublishesNoExecution(t *testing.T) {
	manager := sessionmanager.New()
	shutdownManager := &transactionalShutdownAfterReserveManager{manager: manager}
	root := context.Background()
	transport := newTransactionalTestConnection()
	observer := &transactionalTestObserver{observed: make(chan executionowner.TerminalResult, 1)}
	var prepared *provisionalSession
	dispatcher, err := newTransactionalDispatcher(
		shutdownManager,
		transactionalRuntimeContextProvider{ctx: root},
		nil,
		observer,
		fixedSessionID("shutdown-loser"),
		func(core *sessionCore, _ websocketConnection, cancellation cancellationDependency) (*provisionalSession, error) {
			result, prepareErr := prepareProvisionalSession(core, transport, cancellation)
			prepared = result
			return result, prepareErr
		},
	)
	if err != nil {
		t.Fatalf("newTransactionalDispatcher() error = %v", err)
	}
	authenticated, cancelConnection := transactionalAuthenticatedContext(t, root)
	defer cancelConnection()

	accepted, err := dispatcher.DispatchAuthenticated(authenticated)
	if accepted || !errors.Is(err, sessionmanager.ErrManagerNotOpen) {
		t.Fatalf("DispatchAuthenticated() = (%t, %v), want false and ErrManagerNotOpen", accepted, err)
	}
	if prepared == nil || prepared.owner.State() != executionowner.StatePreCommit {
		t.Fatal("Commit loser did not remain a dormant PreCommit Owner")
	}
	if registrations := shutdownManager.snapshot.Registrations(); len(registrations) != 0 {
		t.Fatalf("Shutdown Snapshot registrations = %d, want 0", len(registrations))
	}
	transport.assertUnused(t)
	if manager.State() != sessionmanager.StateClosed {
		t.Fatalf("Manager state = %v, want Closed after Reservation Abort", manager.State())
	}
	if err := manager.Wait(context.Background()); err != nil {
		t.Fatalf("Manager.Wait() error = %v", err)
	}
	select {
	case result := <-observer.observed:
		t.Fatalf("Observer received Terminal Result for non-committed execution: %+v", result)
	default:
	}
}

func TestTransactionalDispatcherCommittedOwnerObservesOnlyRootRuntimeCancellation(t *testing.T) {
	manager := sessionmanager.New()
	root, cancelRoot := context.WithCancel(context.Background())
	defer cancelRoot()
	transport := newTransactionalTestConnection()
	observer := &transactionalTestObserver{observed: make(chan executionowner.TerminalResult, 1)}
	dispatcher, err := newTransactionalDispatcher(
		manager,
		transactionalRuntimeContextProvider{ctx: root},
		nil,
		observer,
		fixedSessionID("runtime-canceled"),
		func(core *sessionCore, _ websocketConnection, cancellation cancellationDependency) (*provisionalSession, error) {
			return prepareProvisionalSession(core, transport, cancellation)
		},
	)
	if err != nil {
		t.Fatalf("newTransactionalDispatcher() error = %v", err)
	}
	authenticated, cancelConnection := transactionalAuthenticatedContext(t, root)
	defer cancelConnection()

	accepted, err := dispatcher.DispatchAuthenticated(authenticated)
	if err != nil || !accepted {
		t.Fatalf("DispatchAuthenticated() = (%t, %v), want true and nil", accepted, err)
	}
	awaitSignal(t, transport.readStarted, "Session Run did not start after Commit")
	manager.BeginShutdown()
	cancelRoot()

	result := awaitTerminalResult(t, observer.observed)
	if result.PrimaryCause() != executionowner.TerminationCauseRuntimeCanceled &&
		!result.SecondaryCauses().Contains(executionowner.TerminationCauseRuntimeCanceled) {
		t.Fatalf("Terminal causes = (%v, %v), want RuntimeCanceled observation", result.PrimaryCause(), result.SecondaryCauses())
	}
	waitContext, cancelWait := context.WithTimeout(context.Background(), time.Second)
	defer cancelWait()
	if err := manager.Wait(waitContext); err != nil {
		t.Fatalf("Manager.Wait() error = %v", err)
	}
}

func TestTransactionalDispatcherDormantExecutionWaitsForCommitPublication(t *testing.T) {
	manager := sessionmanager.New()
	commitEntered := make(chan struct{})
	releaseCommit := make(chan struct{})
	blockingManager := &transactionalBlockingManager{
		manager:       manager,
		commitEntered: commitEntered,
		releaseCommit: releaseCommit,
	}
	transport := newTransactionalTestConnection()
	observer := &transactionalTestObserver{observed: make(chan executionowner.TerminalResult, 1)}
	var prepared *provisionalSession
	dispatcher, err := newTransactionalDispatcher(
		blockingManager,
		transactionalRuntimeContextProvider{ctx: context.Background()},
		nil,
		observer,
		fixedSessionID("blocked-commit"),
		func(core *sessionCore, _ websocketConnection, cancellation cancellationDependency) (*provisionalSession, error) {
			result, prepareErr := prepareProvisionalSession(core, transport, cancellation)
			prepared = result
			return result, prepareErr
		},
	)
	if err != nil {
		t.Fatalf("newTransactionalDispatcher() error = %v", err)
	}
	authenticated, cancelConnection := transactionalAuthenticatedContext(t, context.Background())
	defer cancelConnection()
	dispatchResult := make(chan transactionalDispatchResult, 1)
	go func() {
		accepted, dispatchErr := dispatcher.DispatchAuthenticated(authenticated)
		dispatchResult <- transactionalDispatchResult{accepted: accepted, err: dispatchErr}
	}()

	awaitSignal(t, commitEntered, "Commit was not entered")
	if prepared == nil || prepared.owner.State() != executionowner.StatePreCommit {
		t.Fatal("Owner left PreCommit while Commit publication was blocked")
	}
	transport.assertUnused(t)
	close(releaseCommit)
	result := <-dispatchResult
	if result.err != nil || !result.accepted {
		t.Fatalf("DispatchAuthenticated() = (%t, %v), want true and nil", result.accepted, result.err)
	}
	awaitSignal(t, transport.readStarted, "dormant execution did not start after Commit")
	snapshot := manager.BeginShutdown()
	if registrations := snapshot.Registrations(); len(registrations) != 1 || !registrations[0].RequestStop() {
		t.Fatal("committed execution was not available to shutdown Snapshot")
	}
	awaitTerminalResult(t, observer.observed)
	awaitManagerWait(t, manager)
}

func TestTransactionalDispatcherRootCancellationBeforeFinalEligibilityDoesNotCommit(t *testing.T) {
	root := newTransactionalEligibilityContext()
	manager := sessionmanager.New()
	transport := newTransactionalTestConnection()
	observer := &transactionalTestObserver{observed: make(chan executionowner.TerminalResult, 1)}
	var prepared *provisionalSession
	dispatcher, err := newTransactionalDispatcher(
		manager,
		transactionalRuntimeContextProvider{ctx: root},
		nil,
		observer,
		fixedSessionID("canceled-before-commit"),
		func(core *sessionCore, _ websocketConnection, cancellation cancellationDependency) (*provisionalSession, error) {
			result, prepareErr := prepareProvisionalSession(core, transport, cancellation)
			prepared = result
			return result, prepareErr
		},
	)
	if err != nil {
		t.Fatalf("newTransactionalDispatcher() error = %v", err)
	}
	authenticated, cancelConnection := transactionalAuthenticatedContext(t, root)
	defer cancelConnection()
	dispatchResult := make(chan transactionalDispatchResult, 1)
	go func() {
		accepted, dispatchErr := dispatcher.DispatchAuthenticated(authenticated)
		dispatchResult <- transactionalDispatchResult{accepted: accepted, err: dispatchErr}
	}()

	awaitSignal(t, root.finalCheckEntered, "final Commit eligibility check was not entered")
	root.cancel()
	close(root.releaseFinalCheck)
	result := <-dispatchResult
	if result.accepted || !errors.Is(result.err, context.Canceled) {
		t.Fatalf("DispatchAuthenticated() = (%t, %v), want false and context.Canceled", result.accepted, result.err)
	}
	if prepared == nil || prepared.owner.State() != executionowner.StatePreCommit {
		t.Fatal("pre-Commit cancellation changed dormant Owner lifecycle")
	}
	transport.assertUnused(t)
	if _, found := manager.Lookup("canceled-before-commit"); found {
		t.Fatal("pre-Commit cancellation published Registration")
	}
}

func TestTransactionalDispatcherStartFailureCompletesManagerAccounting(t *testing.T) {
	manager := sessionmanager.New()
	transport := newTransactionalTestConnection()
	observer := &transactionalTestObserver{observed: make(chan executionowner.TerminalResult, 1)}
	dispatcher, err := newTransactionalDispatcher(
		manager,
		transactionalRuntimeContextProvider{ctx: context.Background()},
		nil,
		observer,
		fixedSessionID("start-failure"),
		func(core *sessionCore, _ websocketConnection, cancellation cancellationDependency) (*provisionalSession, error) {
			prepared, prepareErr := prepareProvisionalSession(core, transport, cancellation)
			if prepareErr == nil {
				prepared.session.state = stateStopped
			}
			return prepared, prepareErr
		},
	)
	if err != nil {
		t.Fatalf("newTransactionalDispatcher() error = %v", err)
	}
	authenticated, cancelConnection := transactionalAuthenticatedContext(t, context.Background())
	defer cancelConnection()
	accepted, err := dispatcher.DispatchAuthenticated(authenticated)
	if err != nil || !accepted {
		t.Fatalf("DispatchAuthenticated() = (%t, %v), want true and nil", accepted, err)
	}
	result := awaitTerminalResult(t, observer.observed)
	if result.StartCategory() != executionowner.StartCategoryFailed ||
		result.RunCategory() != executionowner.RunCategoryNotStarted {
		t.Fatalf("execution categories = (%v, %v), want StartFailed and RunNotStarted", result.StartCategory(), result.RunCategory())
	}
	manager.BeginShutdown()
	awaitManagerWait(t, manager)
	if _, found := manager.Lookup("start-failure"); found {
		t.Fatal("Start failure left Registration published")
	}
}

func TestTransactionalDispatcherRunFailureCompletesManagerAccounting(t *testing.T) {
	manager := sessionmanager.New()
	transport := &transactionalImmediateFailureConnection{}
	observer := &transactionalTestObserver{observed: make(chan executionowner.TerminalResult, 1)}
	dispatcher := newTransactionalFailureDispatcher(t, manager, transport, observer, false)
	authenticated, cancelConnection := transactionalAuthenticatedContext(t, context.Background())
	defer cancelConnection()
	accepted, err := dispatcher.DispatchAuthenticated(authenticated)
	if err != nil || !accepted {
		t.Fatalf("DispatchAuthenticated() = (%t, %v), want true and nil", accepted, err)
	}
	result := awaitTerminalResult(t, observer.observed)
	if result.StartCategory() != executionowner.StartCategorySucceeded ||
		result.RunCategory() != executionowner.RunCategoryFailed {
		t.Fatalf("execution categories = (%v, %v), want StartSucceeded and RunFailed", result.StartCategory(), result.RunCategory())
	}
	manager.BeginShutdown()
	awaitManagerWait(t, manager)
}

func TestTransactionalDispatcherCleanupPanicIsReportedAndAccountingCompletes(t *testing.T) {
	manager := sessionmanager.New()
	transport := &transactionalImmediateFailureConnection{}
	observer := &transactionalTestObserver{observed: make(chan executionowner.TerminalResult, 1)}
	dispatcher := newTransactionalFailureDispatcher(t, manager, transport, observer, true)
	authenticated, cancelConnection := transactionalAuthenticatedContext(t, context.Background())
	defer cancelConnection()
	accepted, err := dispatcher.DispatchAuthenticated(authenticated)
	if err != nil || !accepted {
		t.Fatalf("DispatchAuthenticated() = (%t, %v), want true and nil", accepted, err)
	}
	result := awaitTerminalResult(t, observer.observed)
	if result.CleanupCategory() != executionowner.CleanupCategoryPanicked ||
		result.CancellationCategory() != executionowner.CancellationCategoryConfirmed {
		t.Fatalf("cleanup categories = (%v, %v), want Panicked and Confirmed", result.CleanupCategory(), result.CancellationCategory())
	}
	manager.BeginShutdown()
	awaitManagerWait(t, manager)
}

func TestTransactionalDispatcherIsNotSelectedByLegacyDispatcher(t *testing.T) {
	legacy := NewDispatcher(nil)
	if _, ok := any(legacy).(*TransactionalDispatcher); ok {
		t.Fatal("production-compatible legacy constructor selected TransactionalDispatcher")
	}
}

type transactionalTestObserver struct {
	observed chan executionowner.TerminalResult
}

type transactionalDispatchResult struct {
	accepted bool
	err      error
}

type transactionalBlockingManager struct {
	manager       *sessionmanager.Manager
	commitEntered chan struct{}
	releaseCommit chan struct{}
}

func (manager *transactionalBlockingManager) Reserve(
	sessionID sessionmanager.SessionID,
) (sessionmanager.ReservationHandle, error) {
	handle, err := manager.manager.Reserve(sessionID)
	if err != nil {
		return nil, err
	}
	return &transactionalBlockingReservation{
		ReservationHandle: handle,
		commitEntered:     manager.commitEntered,
		releaseCommit:     manager.releaseCommit,
	}, nil
}

type transactionalBlockingReservation struct {
	sessionmanager.ReservationHandle
	commitEntered chan struct{}
	releaseCommit chan struct{}
	once          sync.Once
}

func (reservation *transactionalBlockingReservation) Commit(
	input sessionmanager.CommitInput,
) (sessionmanager.CommitResult, error) {
	reservation.once.Do(func() { close(reservation.commitEntered) })
	<-reservation.releaseCommit
	return reservation.ReservationHandle.Commit(input)
}

type transactionalEligibilityContext struct {
	done              chan struct{}
	finalCheckEntered chan struct{}
	releaseFinalCheck chan struct{}
	cancelOnce        sync.Once
	finalOnce         sync.Once
	canceled          atomic.Bool
	errCalls          atomic.Int32
}

func newTransactionalEligibilityContext() *transactionalEligibilityContext {
	return &transactionalEligibilityContext{
		done:              make(chan struct{}),
		finalCheckEntered: make(chan struct{}),
		releaseFinalCheck: make(chan struct{}),
	}
}

func (*transactionalEligibilityContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (ctx *transactionalEligibilityContext) Done() <-chan struct{}   { return ctx.done }
func (*transactionalEligibilityContext) Value(any) any               { return nil }

func (ctx *transactionalEligibilityContext) Err() error {
	if ctx.errCalls.Add(1) == 2 {
		ctx.finalOnce.Do(func() { close(ctx.finalCheckEntered) })
		<-ctx.releaseFinalCheck
	}
	if ctx.canceled.Load() {
		return context.Canceled
	}
	return nil
}

func (ctx *transactionalEligibilityContext) cancel() {
	ctx.canceled.Store(true)
	ctx.cancelOnce.Do(func() { close(ctx.done) })
}

type transactionalImmediateFailureConnection struct{}

func (*transactionalImmediateFailureConnection) Read(context.Context) (websocket.MessageType, []byte, error) {
	return 0, nil, errors.New("test transactional Run failure")
}

func (*transactionalImmediateFailureConnection) Write(context.Context, websocket.MessageType, []byte) error {
	return nil
}

func (*transactionalImmediateFailureConnection) Close(websocket.StatusCode, string) error { return nil }
func (*transactionalImmediateFailureConnection) CloseNow() error                          { return nil }

func newTransactionalFailureDispatcher(
	t *testing.T,
	manager *sessionmanager.Manager,
	transport websocketConnection,
	observer executionowner.TerminalObserver,
	cancellationPanic bool,
) *TransactionalDispatcher {
	t.Helper()
	dispatcher, err := newTransactionalDispatcher(
		manager,
		transactionalRuntimeContextProvider{ctx: context.Background()},
		nil,
		observer,
		fixedSessionID("transactional-failure"),
		func(core *sessionCore, _ websocketConnection, cancellation cancellationDependency) (*provisionalSession, error) {
			if cancellationPanic {
				originalCancel := cancellation.cancel
				cancellation.cancel = func() {
					originalCancel()
					panic("test cancellation cleanup panic")
				}
			}
			return prepareProvisionalSession(core, transport, cancellation)
		},
	)
	if err != nil {
		t.Fatalf("newTransactionalDispatcher() error = %v", err)
	}
	return dispatcher
}

type transactionalRuntimeContextProvider struct {
	ctx context.Context
}

func (provider transactionalRuntimeContextProvider) RuntimeContext() context.Context {
	return provider.ctx
}

func (observer *transactionalTestObserver) Observe(result executionowner.TerminalResult) {
	observer.observed <- result
}

type transactionalTestConnection struct {
	readStarted chan struct{}
	closed      chan struct{}
	readOnce    sync.Once
	closeOnce   sync.Once
	reads       atomic.Int32
	writes      atomic.Int32
	closes      atomic.Int32
	closeNows   atomic.Int32
}

func newTransactionalTestConnection() *transactionalTestConnection {
	return &transactionalTestConnection{
		readStarted: make(chan struct{}),
		closed:      make(chan struct{}),
	}
}

func (connection *transactionalTestConnection) Read(ctx context.Context) (websocket.MessageType, []byte, error) {
	connection.reads.Add(1)
	connection.readOnce.Do(func() { close(connection.readStarted) })
	select {
	case <-ctx.Done():
		return 0, nil, ctx.Err()
	case <-connection.closed:
		return 0, nil, net.ErrClosed
	}
}

func (connection *transactionalTestConnection) Write(context.Context, websocket.MessageType, []byte) error {
	connection.writes.Add(1)
	return nil
}

func (connection *transactionalTestConnection) Close(websocket.StatusCode, string) error {
	connection.closes.Add(1)
	connection.closeOnce.Do(func() { close(connection.closed) })
	return nil
}

func (connection *transactionalTestConnection) CloseNow() error {
	connection.closeNows.Add(1)
	connection.closeOnce.Do(func() { close(connection.closed) })
	return nil
}

func (connection *transactionalTestConnection) assertUnused(t *testing.T) {
	t.Helper()
	if connection.reads.Load() != 0 || connection.writes.Load() != 0 ||
		connection.closes.Load() != 0 || connection.closeNows.Load() != 0 {
		t.Fatalf("transport calls = (Read %d, Write %d, Close %d, CloseNow %d), want zero", connection.reads.Load(), connection.writes.Load(), connection.closes.Load(), connection.closeNows.Load())
	}
}

type transactionalReservationManager struct {
	reservation sessionmanager.ReservationHandle
}

type transactionalShutdownAfterReserveManager struct {
	manager  *sessionmanager.Manager
	snapshot sessionmanager.ShutdownSnapshot
}

func (manager *transactionalShutdownAfterReserveManager) Reserve(
	sessionID sessionmanager.SessionID,
) (sessionmanager.ReservationHandle, error) {
	handle, err := manager.manager.Reserve(sessionID)
	if err != nil {
		return nil, err
	}
	manager.snapshot = manager.manager.BeginShutdown()
	return handle, nil
}

func (manager *transactionalReservationManager) Reserve(sessionmanager.SessionID) (sessionmanager.ReservationHandle, error) {
	return manager.reservation, nil
}

type transactionalFailingReservation struct {
	commitErr     error
	panicOnCommit bool
	commitCalls   atomic.Int32
	abortCalls    atomic.Int32
}

func (reservation *transactionalFailingReservation) Commit(sessionmanager.CommitInput) (sessionmanager.CommitResult, error) {
	reservation.commitCalls.Add(1)
	if reservation.panicOnCommit {
		panic("test Commit panic")
	}
	return sessionmanager.CommitResult{}, reservation.commitErr
}

func (reservation *transactionalFailingReservation) Abort() {
	reservation.abortCalls.Add(1)
}

func (reservation *transactionalFailingReservation) AbortUnlessCommitted() {
	reservation.Abort()
}

func transactionalAuthenticatedContext(
	t *testing.T,
	root context.Context,
) (connection.AuthenticatedContext, context.CancelFunc) {
	t.Helper()
	request := httptest.NewRequest("GET", "http://example.test/ws", nil)
	request.RemoteAddr = "192.0.2.1:4321"
	connectionContext := connection.NewRuntimeContext(root, nil, request)
	return connection.NewAuthenticatedContext(connectionContext, validPrincipal()), connectionContext.Cancel
}

func awaitSignal(t *testing.T, signal <-chan struct{}, failure string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(time.Second):
		t.Fatal(failure)
	}
}

func awaitTerminalResult(
	t *testing.T,
	results <-chan executionowner.TerminalResult,
) executionowner.TerminalResult {
	t.Helper()
	select {
	case result := <-results:
		return result
	case <-time.After(time.Second):
		t.Fatal("Terminal Observer was not invoked")
		return executionowner.TerminalResult{}
	}
}

func awaitManagerWait(t *testing.T, manager *sessionmanager.Manager) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := manager.Wait(ctx); err != nil {
		t.Fatalf("Manager.Wait() error = %v", err)
	}
}
