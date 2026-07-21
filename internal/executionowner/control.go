package executionowner

import (
	"context"
	"errors"
	"fmt"
)

var (
	errInvalidRuntimeObservation   = errors.New("invalid Runtime cancellation observation")
	errObservationAlreadyInstalled = errors.New("Runtime cancellation observation already installed")
	errControlCleanupUnconfirmed   = errors.New("execution owner control cleanup is not confirmed")
)

// terminationCause is the bounded identity of one termination source. It is
// owner-private because Task 6 publishes no terminal result.
type terminationCause uint8

const (
	terminationNone terminationCause = iota
	terminationExplicitStop
	terminationRuntimeCanceled
	terminationNaturalCompletion
	terminationExecutionFailure
	terminationRecoveredPanic
)

type terminationSet uint8

func (set terminationSet) contains(cause terminationCause) bool {
	if cause <= terminationNone || cause > terminationRecoveredPanic {
		return false
	}
	return set&(1<<uint(cause-1)) != 0
}

func (set terminationSet) add(cause terminationCause) terminationSet {
	if cause <= terminationNone || cause > terminationRecoveredPanic {
		return set
	}
	return set | 1<<uint(cause-1)
}

type callbackAnomaly uint8

const callbackAnomalyNone callbackAnomaly = 0

const (
	callbackAnomalyInvocationPanic callbackAnomaly = 1 << iota
	callbackAnomalyInstallFailure
	callbackAnomalyInstallPanic
	callbackAnomalyUnregisterFailure
	callbackAnomalyUnregisterPanic
)

type callbackCleanupStatus uint8

const (
	callbackCleanupUnconfirmed callbackCleanupStatus = iota + 1
	callbackCleanupConfirmed
)

// callbackCleanupResult is a detached immutable result. It deliberately
// contains only bounded scalar state.
type callbackCleanupResult struct {
	status  callbackCleanupStatus
	anomaly callbackAnomaly
}

type callbackRegistration struct {
	unregister func() error
}

// runtimeCancellationObservation is the read-only root Runtime observation
// dependency. It provides no root cancellation authority.
type runtimeCancellationObservation struct {
	root     context.Context
	register func(func()) (callbackRegistration, error)
}

func newRuntimeCancellationObservation(root context.Context) runtimeCancellationObservation {
	return runtimeCancellationObservation{
		root: root,
		register: func(callback func()) (callbackRegistration, error) {
			stop := context.AfterFunc(root, callback)
			return callbackRegistration{
				unregister: func() error {
					stop()
					return nil
				},
			}, nil
		},
	}
}

type controlCell struct {
	primary   terminationCause
	secondary terminationSet
	anomalies callbackAnomaly

	executionCancel func()

	admissionOpen   bool
	outstanding     uint64
	outstandingDone chan struct{}

	installationAttempted  bool
	installationInProgress bool
	installationDone       chan struct{}
	rootObservation        context.Context
	unregister             func() error
	registrationUncertain  bool

	drainStarted bool
	drainDone    chan struct{}
	drainResult  callbackCleanupResult
	sealed       bool
}

func newControlCell() controlCell {
	return controlCell{admissionOpen: true}
}

func (owner *Owner) requestStop() bool {
	if owner == nil || owner.state == nil {
		return false
	}

	state := owner.state
	state.mu.Lock()
	control := &state.control
	if !control.admissionOpen || control.sealed ||
		state.current == StateTerminalizing || state.current == StateTerminal {
		state.mu.Unlock()
		return false
	}

	control.outstanding++
	won := control.recordTermination(terminationExplicitStop)
	cancel := control.executionCancel
	state.mu.Unlock()

	defer owner.leaveControlCall()
	if won && invokeCancellation(cancel) {
		owner.recordCallbackAnomaly(callbackAnomalyInvocationPanic)
		owner.recordTermination(terminationExecutionFailure)
	}
	return won
}

func (owner *Owner) bindExecutionCancellation(cancel func()) error {
	if owner == nil || owner.state == nil || cancel == nil {
		return ErrUninitializedOwner
	}

	state := owner.state
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.control.executionCancel != nil || state.control.sealed {
		return fmt.Errorf("%w: execution cancellation is already bound", errInvalidRuntimeObservation)
	}
	state.control.executionCancel = cancel
	return nil
}

func (owner *Owner) installRuntimeCancellation(
	observation runtimeCancellationObservation,
) error {
	if owner == nil || owner.state == nil {
		return ErrUninitializedOwner
	}
	if observation.root == nil || observation.register == nil {
		return errInvalidRuntimeObservation
	}

	state := owner.state
	state.mu.Lock()
	control := &state.control
	if state.current != StateCommitted || control.drainStarted || control.sealed {
		state.mu.Unlock()
		return fmt.Errorf("%w: observation requires an undrained Committed Owner", errInvalidRuntimeObservation)
	}
	if control.installationAttempted {
		state.mu.Unlock()
		return errObservationAlreadyInstalled
	}
	control.installationAttempted = true
	control.installationInProgress = true
	control.installationDone = make(chan struct{})
	control.rootObservation = observation.root
	state.mu.Unlock()

	registration, registerErr, registerPanicked := invokeObservationRegister(
		observation.register,
		owner.runtimeCancellationCallback,
	)

	state.mu.Lock()
	control = &state.control
	control.unregister = registration.unregister
	control.installationInProgress = false
	if registerPanicked {
		control.anomalies |= callbackAnomalyInstallPanic
		control.registrationUncertain = registration.unregister == nil
	} else if registerErr != nil || registration.unregister == nil {
		control.anomalies |= callbackAnomalyInstallFailure
		control.registrationUncertain = registration.unregister == nil
	}
	close(control.installationDone)
	state.mu.Unlock()

	if registerPanicked {
		owner.recordTermination(terminationExecutionFailure)
		return fmt.Errorf("%w: callback registration panicked", errInvalidRuntimeObservation)
	}
	if registerErr != nil {
		owner.recordTermination(terminationExecutionFailure)
		return fmt.Errorf("%w: %w", errInvalidRuntimeObservation, registerErr)
	}
	if registration.unregister == nil {
		owner.recordTermination(terminationExecutionFailure)
		return fmt.Errorf("%w: callback registration returned no unregister operation", errInvalidRuntimeObservation)
	}

	// Registration happens before this synchronous check. The registered
	// callback and this path deliberately use the same admitted mutation.
	if observation.root.Err() != nil {
		owner.runtimeCancellationCallback()
	}
	return nil
}

func (owner *Owner) runtimeCancellationCallback() {
	if owner == nil || owner.state == nil {
		return
	}

	state := owner.state
	state.mu.Lock()
	control := &state.control
	if !control.admissionOpen || control.sealed {
		state.mu.Unlock()
		return
	}
	control.outstanding++
	won := control.recordTermination(terminationRuntimeCanceled)
	cancel := control.executionCancel
	state.mu.Unlock()

	defer owner.leaveControlCall()
	defer func() {
		if recover() != nil {
			owner.recordCallbackAnomaly(callbackAnomalyInvocationPanic)
			owner.recordTermination(terminationExecutionFailure)
		}
	}()
	if won && cancel != nil {
		cancel()
	}
}

func (owner *Owner) recordTermination(cause terminationCause) bool {
	if owner == nil || owner.state == nil {
		return false
	}
	state := owner.state
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.control.sealed {
		return false
	}
	return state.control.recordTermination(cause)
}

func (control *controlCell) recordTermination(cause terminationCause) bool {
	if cause <= terminationNone || cause > terminationRecoveredPanic {
		return false
	}
	if control.primary == terminationNone {
		control.primary = cause
		return true
	}
	if control.primary != cause {
		control.secondary = control.secondary.add(cause)
	}
	return false
}

func (owner *Owner) recordCallbackAnomaly(anomaly callbackAnomaly) {
	if owner == nil || owner.state == nil {
		return
	}
	state := owner.state
	state.mu.Lock()
	state.control.anomalies |= anomaly
	state.mu.Unlock()
}

func (owner *Owner) leaveControlCall() {
	state := owner.state
	state.mu.Lock()
	control := &state.control
	if control.outstanding > 0 {
		control.outstanding--
	}
	if !control.admissionOpen && control.outstanding == 0 && control.outstandingDone != nil {
		close(control.outstandingDone)
		control.outstandingDone = nil
	}
	state.mu.Unlock()
}

func (owner *Owner) unregisterAndDrain() callbackCleanupResult {
	if owner == nil || owner.state == nil {
		return callbackCleanupResult{status: callbackCleanupUnconfirmed}
	}

	state := owner.state
	state.mu.Lock()
	control := &state.control
	if control.drainStarted {
		done := control.drainDone
		state.mu.Unlock()
		<-done
		state.mu.RLock()
		result := state.control.drainResult
		state.mu.RUnlock()
		return result
	}
	control.drainStarted = true
	control.admissionOpen = false
	control.drainDone = make(chan struct{})
	installationDone := control.installationDone
	installationInProgress := control.installationInProgress
	state.mu.Unlock()

	if installationInProgress {
		<-installationDone
	}

	state.mu.RLock()
	unregister := state.control.unregister
	registrationUncertain := state.control.registrationUncertain
	state.mu.RUnlock()

	unregisterErr, unregisterPanicked := invokeUnregister(unregister)
	if unregisterPanicked {
		owner.recordCallbackAnomaly(callbackAnomalyUnregisterPanic)
		return owner.publishDrainResult(callbackCleanupResult{
			status:  callbackCleanupUnconfirmed,
			anomaly: callbackAnomalyUnregisterPanic,
		})
	}
	if unregisterErr != nil || registrationUncertain {
		owner.recordCallbackAnomaly(callbackAnomalyUnregisterFailure)
		return owner.publishDrainResult(callbackCleanupResult{
			status:  callbackCleanupUnconfirmed,
			anomaly: callbackAnomalyUnregisterFailure,
		})
	}

	state.mu.Lock()
	control = &state.control
	var outstandingDone <-chan struct{}
	if control.outstanding > 0 {
		if control.outstandingDone == nil {
			control.outstandingDone = make(chan struct{})
		}
		outstandingDone = control.outstandingDone
	}
	state.mu.Unlock()
	if outstandingDone != nil {
		<-outstandingDone
	}

	return owner.publishDrainResult(callbackCleanupResult{status: callbackCleanupConfirmed})
}

func (owner *Owner) publishDrainResult(result callbackCleanupResult) callbackCleanupResult {
	state := owner.state
	state.mu.Lock()
	state.control.drainResult = result
	close(state.control.drainDone)
	state.mu.Unlock()
	return result
}

func (owner *Owner) sealControl() error {
	if owner == nil || owner.state == nil {
		return ErrUninitializedOwner
	}

	state := owner.state
	state.mu.Lock()
	defer state.mu.Unlock()
	control := &state.control
	if control.sealed {
		return nil
	}
	if !control.drainStarted || control.drainResult.status != callbackCleanupConfirmed {
		return errControlCleanupUnconfirmed
	}
	control.sealed = true
	control.admissionOpen = false
	control.executionCancel = nil
	control.rootObservation = nil
	control.unregister = nil
	return nil
}

func invokeObservationRegister(
	register func(func()) (callbackRegistration, error),
	callback func(),
) (registration callbackRegistration, err error, panicked bool) {
	defer func() {
		if recover() != nil {
			registration = callbackRegistration{}
			err = nil
			panicked = true
		}
	}()
	registration, err = register(callback)
	return registration, err, false
}

func invokeUnregister(unregister func() error) (err error, panicked bool) {
	if unregister == nil {
		return nil, false
	}
	defer func() {
		if recover() != nil {
			err = nil
			panicked = true
		}
	}()
	return unregister(), false
}

func invokeCancellation(cancel func()) (panicked bool) {
	if cancel == nil {
		return false
	}
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	cancel()
	return false
}
