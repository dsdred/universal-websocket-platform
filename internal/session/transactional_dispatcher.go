package session

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/dsdred/universal-websocket-platform/internal/connection"
	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/sessionmanager"
)

var (
	ErrNilSessionManager          = errors.New("Session Manager is nil")
	ErrNilRuntimeContext          = errors.New("root Runtime context observation input is nil or inactive")
	ErrNilTransactionalObserver   = errors.New("transactional Dispatcher Terminal Observer is nil")
	ErrTransactionalDispatchPanic = errors.New("transactional Session dispatch panicked")
)

type reservationManager interface {
	Reserve(sessionmanager.SessionID) (sessionmanager.ReservationHandle, error)
}

type provisionalPreparer func(
	*sessionCore,
	websocketConnection,
	cancellationDependency,
) (*provisionalSession, error)

// RuntimeContextProvider exposes the active Host-owned root Runtime context
// without exposing its cancellation authority.
type RuntimeContextProvider interface {
	RuntimeContext() context.Context
}

// TransactionalDispatcher owns the complete pre-Commit transaction and
// returns immediately after successful irreversible ownership transfer. It is
// the sole production Session handoff selected by Runtime composition.
type TransactionalDispatcher struct {
	manager  reservationManager
	root     RuntimeContextProvider
	handler  message.Handler
	observer executionowner.TerminalObserver
	generate idGenerator
	prepare  provisionalPreparer
}

// NewTransactionalDispatcher creates the production transaction-capable
// Dispatcher required by the Runtime Foundation.
func NewTransactionalDispatcher(
	manager *sessionmanager.Manager,
	root RuntimeContextProvider,
	handler message.Handler,
	observer executionowner.TerminalObserver,
) (*TransactionalDispatcher, error) {
	return newTransactionalDispatcher(
		manager,
		root,
		handler,
		observer,
		generateID,
		prepareProvisionalSession,
	)
}

func newTransactionalDispatcher(
	manager reservationManager,
	root RuntimeContextProvider,
	handler message.Handler,
	observer executionowner.TerminalObserver,
	generate idGenerator,
	prepare provisionalPreparer,
) (*TransactionalDispatcher, error) {
	if isNilDependency(manager) {
		return nil, ErrNilSessionManager
	}
	if isNilDependency(root) {
		return nil, ErrNilRuntimeContext
	}
	if isNilDependency(observer) {
		return nil, ErrNilTransactionalObserver
	}
	if generate == nil || prepare == nil {
		return nil, errNilIDGenerator
	}
	return &TransactionalDispatcher{
		manager:  manager,
		root:     root,
		handler:  handler,
		observer: observer,
		generate: generate,
		prepare:  prepare,
	}, nil
}

// DispatchAuthenticated prepares one dormant execution path and transfers
// ownership only through successful Session Manager Commit.
func (dispatcher *TransactionalDispatcher) DispatchAuthenticated(
	authenticatedContext connection.AuthenticatedContext,
) (accepted bool, err error) {
	transaction := dormantTransaction{}
	defer func() {
		if recovered := recover(); recovered != nil {
			accepted = false
			err = errors.Join(err, ErrTransactionalDispatchPanic)
		}
		if !accepted {
			err = errors.Join(err, transaction.retire())
		}
	}()

	connectionContext := authenticatedContext.ConnectionContext()
	executionContext := connectionContext.Context()
	if executionContext == nil {
		return false, executionowner.ErrInvalidExecutionEnvironment
	}
	if err := executionContext.Err(); err != nil {
		return false, err
	}
	rootContext := dispatcher.root.RuntimeContext()
	if rootContext == nil {
		return false, ErrNilRuntimeContext
	}
	if err := rootContext.Err(); err != nil {
		return false, err
	}
	request := connectionContext.Request()
	if request == nil {
		return false, errors.New("WebSocket upgrade request is nil")
	}

	sessionID, err := dispatcher.generate()
	if err != nil {
		return false, fmt.Errorf("generate Session ID: %w", err)
	}
	if sessionID == "" {
		return false, errInvalidGeneratedSessionID
	}
	transaction.reservation, err = dispatcher.manager.Reserve(
		sessionmanager.SessionID(sessionID),
	)
	if err != nil {
		return false, err
	}

	core, err := newSessionCore(
		authenticatedContext.Principal(),
		request.RemoteAddr,
		func() (string, error) { return sessionID, nil },
		nil,
		dispatcher.handler,
	)
	if err != nil {
		return false, err
	}
	prepared, err := dispatcher.prepare(
		core,
		connectionContext.Connection(),
		cancellationDependency{
			done:   executionContext.Done(),
			cancel: connectionContext.Cancel,
		},
	)
	if err != nil {
		return false, err
	}
	environment, err := executionowner.NewExecutionEnvironment(
		rootContext,
		executionContext,
		connectionContext.Cancel,
	)
	if err != nil {
		return false, err
	}

	handoff := sessionmanager.NewCommitHandoff()
	transaction.notCommitted = handoff.NotCommittedPublisher()
	transaction.dormantDone = make(chan error, 1)
	go runDormantExecution(
		handoff.Waiter(),
		prepared,
		environment,
		dispatcher.observer,
		transaction.dormantDone,
	)

	if err := executionContext.Err(); err != nil {
		return false, err
	}
	if err := rootContext.Err(); err != nil {
		return false, err
	}
	input, err := sessionmanager.NewCommitInput(
		prepared.owner,
		handoff.CommitPublisher(),
	)
	if err != nil {
		return false, err
	}
	if _, err := transaction.reservation.Commit(input); err != nil {
		return false, err
	}

	transaction.committed = true
	return true, nil
}

type dormantTransaction struct {
	reservation  sessionmanager.ReservationHandle
	notCommitted sessionmanager.NotCommittedPublisher
	dormantDone  chan error
	committed    bool
}

func (transaction *dormantTransaction) retire() error {
	if transaction == nil || transaction.committed {
		return nil
	}
	var err error
	if transaction.notCommitted != (sessionmanager.NotCommittedPublisher{}) {
		err = transaction.notCommitted.Publish()
	}
	if transaction.dormantDone != nil {
		err = errors.Join(err, <-transaction.dormantDone)
	}
	if transaction.reservation != nil {
		transaction.reservation.AbortUnlessCommitted()
	}
	return err
}

func runDormantExecution(
	waiter sessionmanager.CommitHandoffWaiter,
	prepared *provisionalSession,
	environment executionowner.ExecutionEnvironment,
	observer executionowner.TerminalObserver,
	done chan<- error,
) {
	var result error
	defer func() {
		if recover() != nil {
			result = errors.Join(result, ErrTransactionalDispatchPanic)
		}
		done <- result
	}()
	result = func() error {
		outcome, err := waiter.Wait()
		if err != nil {
			return err
		}
		commitResult, committed := outcome.CommitResult()
		if !committed {
			return nil
		}
		if err := prepared.owner.Transition(
			executionowner.StatePreCommit,
			executionowner.StateCommitted,
		); err != nil {
			return err
		}
		return prepared.owner.ExecuteInEnvironment(
			environment,
			prepared.lifecycle,
			commitResult.CompletionAdapter(),
			observer,
			commitResult.LifetimeLease(),
		)
	}()
}

func isNilDependency(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

var _ connection.AuthenticatedDispatcher = (*TransactionalDispatcher)(nil)
