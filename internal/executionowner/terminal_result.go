package executionowner

import "fmt"

type terminalResultError string

const (
	// ErrInvalidTerminalResult indicates that construction received an unknown
	// category or an inconsistent combination of bounded outcomes.
	ErrInvalidTerminalResult terminalResultError = "invalid execution owner Terminal Result"
)

func (err terminalResultError) Error() string {
	return string(err)
}

// StartCategory is the bounded outcome of Session Start.
type StartCategory uint8

const (
	StartCategoryNotAttempted StartCategory = iota + 1
	StartCategorySucceeded
	StartCategoryFailed
	StartCategoryPanicked
)

// RunCategory is the bounded outcome of Session Run.
type RunCategory uint8

const (
	RunCategoryNotStarted RunCategory = iota + 1
	RunCategoryReturned
	RunCategoryFailed
	RunCategoryPanicked
)

// CleanupCategory is the bounded outcome of Session Cleanup.
type CleanupCategory uint8

const (
	CleanupCategoryNotRequired CleanupCategory = iota + 1
	CleanupCategorySucceeded
	CleanupCategoryFailed
	CleanupCategoryPanicked
	// CleanupCategoryBlocked is an internal pre-result category. A published
	// Terminal Result cannot be constructed with this value.
	CleanupCategoryBlocked
)

// CancellationCategory records whether effective connection cancellation was
// confirmed by Session Cleanup.
type CancellationCategory uint8

const (
	CancellationCategoryConfirmed CancellationCategory = iota + 1
	CancellationCategoryAnomaly
)

// CompletionCategory is the bounded outcome of the bound Registration
// completion invocation.
type CompletionCategory uint8

const (
	CompletionCategoryCompleted CompletionCategory = iota + 1
	CompletionCategoryAccountingAnomaly
	CompletionCategoryPanicked
)

// RecoveredPanicPhase identifies an execution phase without retaining a panic
// payload.
type RecoveredPanicPhase uint8

const (
	RecoveredPanicPhaseNone RecoveredPanicPhase = iota
	RecoveredPanicPhaseStart
	RecoveredPanicPhaseRun
)

// TerminationCause is the bounded primary termination identity.
type TerminationCause uint8

const (
	TerminationCauseExplicitStop TerminationCause = iota + 1
	TerminationCauseRuntimeCanceled
	TerminationCauseNaturalCompletion
	TerminationCauseExecutionFailure
	TerminationCauseRecoveredPanic
)

// SecondaryCauses is a bounded set of non-primary termination causes.
type SecondaryCauses uint8

const (
	SecondaryCauseExplicitStop SecondaryCauses = 1 << iota
	SecondaryCauseRuntimeCanceled
	SecondaryCauseNaturalCompletion
	SecondaryCauseExecutionFailure
	SecondaryCauseRecoveredPanic
)

const allSecondaryCauses = SecondaryCauseExplicitStop |
	SecondaryCauseRuntimeCanceled |
	SecondaryCauseNaturalCompletion |
	SecondaryCauseExecutionFailure |
	SecondaryCauseRecoveredPanic

// Contains reports whether the bounded set contains cause.
func (causes SecondaryCauses) Contains(cause TerminationCause) bool {
	flag, ok := secondaryFlag(cause)
	return ok && causes&flag != 0
}

// RuntimeObservationAnomalies is a bounded set of Runtime-cancellation
// observation anomalies known before Terminal Result construction.
type RuntimeObservationAnomalies uint8

const (
	RuntimeObservationAnomalyInstallFailure RuntimeObservationAnomalies = 1 << iota
	RuntimeObservationAnomalyInstallPanic
	RuntimeObservationAnomalyInvocationPanic
)

const allRuntimeObservationAnomalies = RuntimeObservationAnomalyInstallFailure |
	RuntimeObservationAnomalyInstallPanic |
	RuntimeObservationAnomalyInvocationPanic

// Has reports whether the bounded observation anomaly is present.
func (anomalies RuntimeObservationAnomalies) Has(
	anomaly RuntimeObservationAnomalies,
) bool {
	return anomaly != 0 &&
		anomaly&^allRuntimeObservationAnomalies == 0 &&
		anomalies&anomaly == anomaly
}

// TerminalResult is the immutable, structurally comparable value describing
// execution outcomes known after Completion and before Observer invocation.
type TerminalResult struct {
	start                StartCategory
	run                  RunCategory
	cleanup              CleanupCategory
	cancellation         CancellationCategory
	completion           CompletionCategory
	recoveredPanicPhase  RecoveredPanicPhase
	primaryCause         TerminationCause
	secondaryCauses      SecondaryCauses
	observationAnomalies RuntimeObservationAnomalies
}

// NewTerminalResult validates and constructs one immutable Terminal Result.
func NewTerminalResult(
	start StartCategory,
	run RunCategory,
	cleanup CleanupCategory,
	cancellation CancellationCategory,
	completion CompletionCategory,
	recoveredPanicPhase RecoveredPanicPhase,
	primaryCause TerminationCause,
	secondaryCauses SecondaryCauses,
	observationAnomalies RuntimeObservationAnomalies,
) (TerminalResult, error) {
	if err := validateTerminalResult(
		start,
		run,
		cleanup,
		cancellation,
		completion,
		recoveredPanicPhase,
		primaryCause,
		secondaryCauses,
		observationAnomalies,
	); err != nil {
		return TerminalResult{}, err
	}

	return TerminalResult{
		start:                start,
		run:                  run,
		cleanup:              cleanup,
		cancellation:         cancellation,
		completion:           completion,
		recoveredPanicPhase:  recoveredPanicPhase,
		primaryCause:         primaryCause,
		secondaryCauses:      secondaryCauses,
		observationAnomalies: observationAnomalies,
	}, nil
}

// StartCategory returns the bounded Session Start outcome.
func (result TerminalResult) StartCategory() StartCategory {
	return result.start
}

// RunCategory returns the bounded Session Run outcome.
func (result TerminalResult) RunCategory() RunCategory {
	return result.run
}

// CleanupCategory returns the bounded Session Cleanup outcome.
func (result TerminalResult) CleanupCategory() CleanupCategory {
	return result.cleanup
}

// CancellationCategory returns the effective-cancellation acknowledgement.
func (result TerminalResult) CancellationCategory() CancellationCategory {
	return result.cancellation
}

// CompletionCategory returns the bound Registration completion outcome.
func (result TerminalResult) CompletionCategory() CompletionCategory {
	return result.completion
}

// RecoveredPanicPhase returns the bounded execution-panic phase.
func (result TerminalResult) RecoveredPanicPhase() RecoveredPanicPhase {
	return result.recoveredPanicPhase
}

// PrimaryCause returns the immutable primary termination cause.
func (result TerminalResult) PrimaryCause() TerminationCause {
	return result.primaryCause
}

// SecondaryCauses returns the bounded secondary termination set.
func (result TerminalResult) SecondaryCauses() SecondaryCauses {
	return result.secondaryCauses
}

// RuntimeObservationAnomalies returns the bounded observation anomalies known
// at construction.
func (result TerminalResult) RuntimeObservationAnomalies() RuntimeObservationAnomalies {
	return result.observationAnomalies
}

func validateTerminalResult(
	start StartCategory,
	run RunCategory,
	cleanup CleanupCategory,
	cancellation CancellationCategory,
	completion CompletionCategory,
	recoveredPanicPhase RecoveredPanicPhase,
	primaryCause TerminationCause,
	secondaryCauses SecondaryCauses,
	observationAnomalies RuntimeObservationAnomalies,
) error {
	if !start.valid() || !run.valid() || !cleanup.publishable() ||
		!cancellation.valid() || !completion.valid() ||
		!recoveredPanicPhase.valid() || !primaryCause.valid() ||
		secondaryCauses&^allSecondaryCauses != 0 ||
		observationAnomalies&^allRuntimeObservationAnomalies != 0 {
		return ErrInvalidTerminalResult
	}
	if start != StartCategorySucceeded && run != RunCategoryNotStarted {
		return fmt.Errorf("%w: Run requires successful Start", ErrInvalidTerminalResult)
	}
	if secondaryCauses.Contains(primaryCause) {
		return fmt.Errorf("%w: primary cause is duplicated in secondary causes", ErrInvalidTerminalResult)
	}

	allCauses := secondaryCauses
	primaryFlag, _ := secondaryFlag(primaryCause)
	allCauses |= primaryFlag
	if (start == StartCategoryFailed || run == RunCategoryFailed) &&
		!allCauses.Contains(TerminationCauseExecutionFailure) {
		return fmt.Errorf("%w: failed execution requires ExecutionFailure cause", ErrInvalidTerminalResult)
	}
	if run == RunCategoryReturned &&
		!allCauses.Contains(TerminationCauseNaturalCompletion) {
		return fmt.Errorf("%w: returned Run requires NaturalCompletion cause", ErrInvalidTerminalResult)
	}
	if allCauses.Contains(TerminationCauseNaturalCompletion) && run != RunCategoryReturned {
		return fmt.Errorf("%w: NaturalCompletion requires returned Run", ErrInvalidTerminalResult)
	}

	panicCause := allCauses.Contains(TerminationCauseRecoveredPanic)
	switch recoveredPanicPhase {
	case RecoveredPanicPhaseNone:
		if panicCause || start == StartCategoryPanicked || run == RunCategoryPanicked {
			return fmt.Errorf("%w: recovered panic phase is missing", ErrInvalidTerminalResult)
		}
	case RecoveredPanicPhaseStart:
		if !panicCause || start != StartCategoryPanicked || run != RunCategoryNotStarted {
			return fmt.Errorf("%w: invalid recovered Start panic", ErrInvalidTerminalResult)
		}
	case RecoveredPanicPhaseRun:
		if !panicCause || start != StartCategorySucceeded || run != RunCategoryPanicked {
			return fmt.Errorf("%w: invalid recovered Run panic", ErrInvalidTerminalResult)
		}
	}
	return nil
}

func (category StartCategory) valid() bool {
	return category >= StartCategoryNotAttempted && category <= StartCategoryPanicked
}

func (category RunCategory) valid() bool {
	return category >= RunCategoryNotStarted && category <= RunCategoryPanicked
}

func (category CleanupCategory) publishable() bool {
	return category >= CleanupCategoryNotRequired && category < CleanupCategoryBlocked
}

func (category CancellationCategory) valid() bool {
	return category == CancellationCategoryConfirmed || category == CancellationCategoryAnomaly
}

func (category CompletionCategory) valid() bool {
	return category >= CompletionCategoryCompleted && category <= CompletionCategoryPanicked
}

func (phase RecoveredPanicPhase) valid() bool {
	return phase <= RecoveredPanicPhaseRun
}

func (cause TerminationCause) valid() bool {
	return cause >= TerminationCauseExplicitStop && cause <= TerminationCauseRecoveredPanic
}

func secondaryFlag(cause TerminationCause) (SecondaryCauses, bool) {
	switch cause {
	case TerminationCauseExplicitStop:
		return SecondaryCauseExplicitStop, true
	case TerminationCauseRuntimeCanceled:
		return SecondaryCauseRuntimeCanceled, true
	case TerminationCauseNaturalCompletion:
		return SecondaryCauseNaturalCompletion, true
	case TerminationCauseExecutionFailure:
		return SecondaryCauseExecutionFailure, true
	case TerminationCauseRecoveredPanic:
		return SecondaryCauseRecoveredPanic, true
	default:
		return 0, false
	}
}
