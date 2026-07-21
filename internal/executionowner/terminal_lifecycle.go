package executionowner

import (
	"context"
	"errors"
	"fmt"

	"github.com/dsdred/universal-websocket-platform/internal/lifetimelease"
)

func (owner *Owner) runTerminalLifecycle(
	session SessionLifecycle,
	completion CompletionAdapter,
	observer TerminalObserver,
	lease lifetimelease.Lease,
	execution executionOutcomes,
	executionErr error,
) error {
	cleanupCategory, cancellationCategory := invokeSessionCleanup(session)
	completionCategory := invokeBoundCompletion(completion)
	primaryCause, secondaryCauses, observationAnomalies := owner.terminalCauseSnapshot()

	result, resultErr := NewTerminalResult(
		execution.start,
		execution.run,
		cleanupCategory,
		cancellationCategory,
		completionCategory,
		execution.panicPhase,
		primaryCause,
		secondaryCauses,
		observationAnomalies,
	)
	if resultErr != nil {
		// Result validation failure is an internal invariant violation. Callback
		// admission still closes, but no fallback result, Observer invocation,
		// Terminal transition, or lease release is permitted.
		owner.unregisterAndDrain()
		return errors.Join(executionErr, fmt.Errorf("construct Terminal Result: %w", resultErr))
	}

	invokeTerminalObserver(observer, result)
	if drain := owner.unregisterAndDrain(); drain.status != callbackCleanupConfirmed {
		return executionErr
	}
	if err := owner.sealControl(); err != nil {
		return errors.Join(executionErr, fmt.Errorf("seal execution control: %w", err))
	}
	if err := owner.transitionLifecycle(StateTerminalizing, StateTerminal); err != nil {
		return errors.Join(executionErr, err)
	}
	if cancellationCategory != CancellationCategoryConfirmed {
		return executionErr
	}

	invokeLifetimeLeaseRelease(lease)
	return executionErr
}

func invokeSessionCleanup(
	session SessionLifecycle,
) (cleanupCategory CleanupCategory, cancellationCategory CancellationCategory) {
	cleanupCategory = CleanupCategoryPanicked
	cancellationCategory = CancellationCategoryAnomaly
	defer func() {
		if recover() != nil {
			cleanupCategory = CleanupCategoryPanicked
			cancellationCategory = CancellationCategoryAnomaly
		}
	}()
	return session.Cleanup(context.Background())
}

func invokeBoundCompletion(completion CompletionAdapter) (
	category CompletionCategory,
) {
	category = CompletionCategoryPanicked
	defer func() {
		if recover() != nil {
			category = CompletionCategoryPanicked
		}
	}()
	switch completion.CompleteBoundRegistration() {
	case CompleteOutcomeCompleted:
		return CompletionCategoryCompleted
	case CompleteOutcomeAccountingAnomaly:
		return CompletionCategoryAccountingAnomaly
	default:
		return CompletionCategoryAccountingAnomaly
	}
}

func invokeTerminalObserver(observer TerminalObserver, result TerminalResult) {
	defer func() {
		recover()
	}()
	observer.Observe(result)
}

func invokeLifetimeLeaseRelease(lease lifetimelease.Lease) (
	outcome lifetimelease.ReleaseOutcome,
) {
	outcome = lifetimelease.ReleaseOutcomeAccountingAnomaly
	defer func() {
		if recover() != nil {
			outcome = lifetimelease.ReleaseOutcomeAccountingAnomaly
		}
	}()
	return lease.Release()
}

func (owner *Owner) terminalCauseSnapshot() (
	TerminationCause,
	SecondaryCauses,
	RuntimeObservationAnomalies,
) {
	state := owner.state
	state.mu.RLock()
	defer state.mu.RUnlock()

	return publicTerminationCause(state.control.primary),
		publicSecondaryCauses(state.control.secondary),
		publicObservationAnomalies(state.control.anomalies)
}

func publicTerminationCause(cause terminationCause) TerminationCause {
	switch cause {
	case terminationExplicitStop:
		return TerminationCauseExplicitStop
	case terminationRuntimeCanceled:
		return TerminationCauseRuntimeCanceled
	case terminationNaturalCompletion:
		return TerminationCauseNaturalCompletion
	case terminationExecutionFailure:
		return TerminationCauseExecutionFailure
	case terminationRecoveredPanic:
		return TerminationCauseRecoveredPanic
	default:
		return 0
	}
}

func publicSecondaryCauses(causes terminationSet) SecondaryCauses {
	var result SecondaryCauses
	for _, cause := range []terminationCause{
		terminationExplicitStop,
		terminationRuntimeCanceled,
		terminationNaturalCompletion,
		terminationExecutionFailure,
		terminationRecoveredPanic,
	} {
		if causes.contains(cause) {
			result |= secondaryCauseFor(cause)
		}
	}
	return result
}

func secondaryCauseFor(cause terminationCause) SecondaryCauses {
	switch cause {
	case terminationExplicitStop:
		return SecondaryCauseExplicitStop
	case terminationRuntimeCanceled:
		return SecondaryCauseRuntimeCanceled
	case terminationNaturalCompletion:
		return SecondaryCauseNaturalCompletion
	case terminationExecutionFailure:
		return SecondaryCauseExecutionFailure
	case terminationRecoveredPanic:
		return SecondaryCauseRecoveredPanic
	default:
		return 0
	}
}

func publicObservationAnomalies(anomalies callbackAnomaly) RuntimeObservationAnomalies {
	var result RuntimeObservationAnomalies
	if anomalies&callbackAnomalyInstallFailure != 0 {
		result |= RuntimeObservationAnomalyInstallFailure
	}
	if anomalies&callbackAnomalyInstallPanic != 0 {
		result |= RuntimeObservationAnomalyInstallPanic
	}
	if anomalies&callbackAnomalyInvocationPanic != 0 {
		result |= RuntimeObservationAnomalyInvocationPanic
	}
	return result
}
