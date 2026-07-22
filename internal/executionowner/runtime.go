package executionowner

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/dsdred/universal-websocket-platform/internal/lifetimelease"
)

var (
	ErrNilSessionLifecycle = errors.New("execution Session lifecycle is nil")
	ErrNilCompletion       = errors.New("execution Completion Adapter is nil")
	ErrNilTerminalObserver = errors.New("execution Terminal Observer is nil")
	ErrNilLifetimeLease    = errors.New("execution Owner Lifetime Lease is nil")
	ErrSessionPanic        = errors.New("execution Session lifecycle panicked")
)

// SessionLifecycle is the complete Session lifecycle contract used by an Owner.
type SessionLifecycle interface {
	Start(context.Context) error
	Run(context.Context) error
	Cleanup(context.Context) (CleanupCategory, CancellationCategory)
}

// Execute runs the single post-Commit Session execution and terminal path in
// the calling goroutine. It is the exclusive post-Commit lifecycle writer.
func (owner *Owner) Execute(
	ctx context.Context,
	session SessionLifecycle,
	completion CompletionAdapter,
	observer TerminalObserver,
	lease lifetimelease.Lease,
) error {
	if owner == nil || owner.state == nil {
		return ErrUninitializedOwner
	}
	if isNilContract(session) {
		return ErrNilSessionLifecycle
	}
	if isNilContract(completion) {
		return ErrNilCompletion
	}
	if isNilContract(observer) {
		return ErrNilTerminalObserver
	}
	if isNilContract(lease) {
		return ErrNilLifetimeLease
	}

	if err := owner.claimExecution(); err != nil {
		return err
	}
	return owner.executeClaimed(ctx, nil, session, completion, observer, lease, nil)
}

// ExecuteInEnvironment runs the single committed path using the prepared
// derived execution context and installs observation of only the root Runtime
// context before Session Start.
func (owner *Owner) ExecuteInEnvironment(
	environment ExecutionEnvironment,
	session SessionLifecycle,
	completion CompletionAdapter,
	observer TerminalObserver,
	lease lifetimelease.Lease,
) error {
	if !environment.valid() {
		return ErrInvalidExecutionEnvironment
	}
	if owner == nil || owner.state == nil {
		return ErrUninitializedOwner
	}
	if isNilContract(session) {
		return ErrNilSessionLifecycle
	}
	if isNilContract(completion) {
		return ErrNilCompletion
	}
	if isNilContract(observer) {
		return ErrNilTerminalObserver
	}
	if isNilContract(lease) {
		return ErrNilLifetimeLease
	}
	if err := owner.claimExecution(); err != nil {
		return err
	}

	var preparationErr error
	if err := owner.bindExecutionCancellation(environment.cancel); err != nil {
		owner.recordTermination(terminationExecutionFailure)
		preparationErr = err
	} else if err := owner.installRuntimeCancellation(
		newRuntimeCancellationObservation(environment.root),
	); err != nil {
		preparationErr = err
	}

	return owner.executeClaimed(
		environment.execution,
		environment.root,
		session,
		completion,
		observer,
		lease,
		preparationErr,
	)
}

func (owner *Owner) claimExecution() error {
	if owner == nil || owner.state == nil {
		return ErrUninitializedOwner
	}
	state := owner.state
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.current != StateCommitted || state.executionClaimed {
		return invalidTransitionError(state.current, StateCommitted, StateStarting)
	}
	state.executionClaimed = true
	return nil
}

func (owner *Owner) executeClaimed(
	ctx context.Context,
	rootObservation context.Context,
	session SessionLifecycle,
	completion CompletionAdapter,
	observer TerminalObserver,
	lease lifetimelease.Lease,
	preparationErr error,
) error {
	outcomes, executionErr := owner.executeSession(ctx, session)
	// A root cancellation can make the derived execution return before the
	// registered callback is scheduled. Observe the same root source once more
	// before closing callback admission so that cancellation cannot be lost.
	if rootObservation != nil && rootObservation.Err() != nil {
		owner.runtimeCancellationCallback()
	}
	executionErr = errors.Join(preparationErr, executionErr)

	return owner.runTerminalLifecycle(
		session,
		completion,
		observer,
		lease,
		outcomes,
		executionErr,
	)
}

type executionOutcomes struct {
	start      StartCategory
	run        RunCategory
	panicPhase RecoveredPanicPhase
}

func (owner *Owner) executeSession(
	ctx context.Context,
	session SessionLifecycle,
) (executionOutcomes, error) {
	outcomes := executionOutcomes{
		start: StartCategoryNotAttempted,
		run:   RunCategoryNotStarted,
	}

	state := owner.state
	state.mu.Lock()
	if state.current != StateCommitted {
		err := invalidTransitionError(state.current, StateCommitted, StateStarting)
		state.mu.Unlock()
		return outcomes, err
	}
	if state.control.primary != terminationNone || ctx.Err() != nil {
		if state.control.primary == terminationNone {
			state.control.recordTermination(terminationExecutionFailure)
		}
		state.current = StateTerminalizing
		err := ctx.Err()
		state.mu.Unlock()
		return outcomes, err
	}
	state.current = StateStarting
	state.mu.Unlock()

	startErr, startPanicked := invokeSessionStart(session, ctx)
	switch {
	case startPanicked:
		outcomes.start = StartCategoryPanicked
		outcomes.panicPhase = RecoveredPanicPhaseStart
		return outcomes,
			owner.finishExecution(StateStarting, terminationRecoveredPanic, ErrSessionPanic)
	case startErr != nil:
		outcomes.start = StartCategoryFailed
		return outcomes,
			owner.finishExecution(StateStarting, terminationExecutionFailure, startErr)
	default:
		outcomes.start = StartCategorySucceeded
	}

	state.mu.Lock()
	if state.current != StateStarting {
		err := invalidTransitionError(state.current, StateStarting, StateRunning)
		state.mu.Unlock()
		return outcomes, err
	}
	if state.control.primary != terminationNone || ctx.Err() != nil {
		if state.control.primary == terminationNone {
			state.control.recordTermination(terminationExecutionFailure)
		}
		state.current = StateTerminalizing
		err := ctx.Err()
		state.mu.Unlock()
		return outcomes, err
	}
	state.current = StateRunning
	state.mu.Unlock()

	runErr, runPanicked := invokeSessionRun(session, ctx)
	switch {
	case runPanicked:
		outcomes.run = RunCategoryPanicked
		outcomes.panicPhase = RecoveredPanicPhaseRun
		return outcomes,
			owner.finishExecution(StateRunning, terminationRecoveredPanic, ErrSessionPanic)
	case runErr != nil:
		outcomes.run = RunCategoryFailed
		return outcomes,
			owner.finishExecution(StateRunning, terminationExecutionFailure, runErr)
	default:
		outcomes.run = RunCategoryReturned
		return outcomes,
			owner.finishExecution(StateRunning, terminationNaturalCompletion, nil)
	}
}

func (owner *Owner) finishExecution(
	from State,
	cause terminationCause,
	executionErr error,
) error {
	state := owner.state
	state.mu.Lock()
	defer state.mu.Unlock()

	if state.current != from {
		return invalidTransitionError(state.current, from, StateTerminalizing)
	}
	state.control.recordTermination(cause)
	state.current = StateTerminalizing
	return executionErr
}

func invokeSessionStart(
	session SessionLifecycle,
	ctx context.Context,
) (err error, panicked bool) {
	defer func() {
		if recover() != nil {
			err = nil
			panicked = true
		}
	}()
	return session.Start(ctx), false
}

func invokeSessionRun(
	session SessionLifecycle,
	ctx context.Context,
) (err error, panicked bool) {
	defer func() {
		if recover() != nil {
			err = nil
			panicked = true
		}
	}()
	return session.Run(ctx), false
}

func isNilContract(contract any) bool {
	if contract == nil {
		return true
	}
	value := reflect.ValueOf(contract)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func invalidTransitionError(current, from, to State) error {
	return fmt.Errorf(
		"%w: current=%s expected=%s requested=%s: current state does not match expected source",
		ErrInvalidTransition,
		stateName(current),
		stateName(from),
		stateName(to),
	)
}
