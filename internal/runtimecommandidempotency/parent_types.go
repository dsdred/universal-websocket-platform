package runtimecommandidempotency

// PhaseKind identifies one privately derived child of a replacement or
// rollback parent. Callers cannot construct a phase identity.
type PhaseKind string

const (
	// PhaseStopOld is the optional ordinal-zero stop phase.
	PhaseStopOld PhaseKind = "stop-old"
	// PhaseStartTarget is the ordinal-one exact target start phase.
	PhaseStartTarget PhaseKind = "start-target"
)

// ParentOutcomeCategory is a bounded semantic parent result. It is separate
// from primitive lifecycle outcomes so parent semantics cannot broaden them.
type ParentOutcomeCategory string

const (
	// ParentOutcomeSucceeded means all required linked work completed.
	ParentOutcomeSucceeded ParentOutcomeCategory = "succeeded"
	// ParentOutcomeSatisfied means the exact target already required no phases.
	ParentOutcomeSatisfied ParentOutcomeCategory = "satisfied"
	// ParentOutcomeStopped means the parent definitively left the Instance stopped.
	ParentOutcomeStopped ParentOutcomeCategory = "stopped"
	// ParentOutcomeCancelled means cancellation won at a definitive boundary.
	ParentOutcomeCancelled ParentOutcomeCategory = "cancelled"
	// ParentOutcomeRejected means a definitive policy or precondition rejection.
	ParentOutcomeRejected ParentOutcomeCategory = "rejected"
	// ParentOutcomeFailed means the parent completed with a definitive failure.
	ParentOutcomeFailed ParentOutcomeCategory = "failed"
)

// ParentTerminalOutcome is an immutable redacted parent result.
type ParentTerminalOutcome struct {
	category ParentOutcomeCategory
}

// NewParentTerminalOutcome validates and constructs a replay-safe parent
// result. Exact phase facts remain owned by their linked records.
func NewParentTerminalOutcome(category ParentOutcomeCategory) (ParentTerminalOutcome, error) {
	outcome := ParentTerminalOutcome{category: category}
	if !outcome.valid() {
		return ParentTerminalOutcome{}, ErrInvalidSubmission
	}
	return outcome, nil
}

// Category returns the stable parent result category.
func (o ParentTerminalOutcome) Category() ParentOutcomeCategory { return o.category }

func (o ParentTerminalOutcome) valid() bool {
	switch o.category {
	case ParentOutcomeSucceeded, ParentOutcomeSatisfied, ParentOutcomeStopped,
		ParentOutcomeCancelled, ParentOutcomeRejected, ParentOutcomeFailed:
		return true
	default:
		return false
	}
}

// ParentRecordView is a detached coherent parent observation.
type ParentRecordView struct {
	scope      Scope
	key        CommandKey
	intent     Intent
	state      CommandState
	revision   Revision
	outcome    ParentTerminalOutcome
	hasOutcome bool
}

// Scope returns the exact parent command scope.
func (v ParentRecordView) Scope() Scope { return v.scope }

// Key returns the opaque parent command key.
func (v ParentRecordView) Key() CommandKey { return v.key }

// Intent returns the immutable replacement or rollback intent.
func (v ParentRecordView) Intent() Intent { return v.intent }

// State returns the durable parent command state.
func (v ParentRecordView) State() CommandState { return v.state }

// Revision returns the durable parent revision.
func (v ParentRecordView) Revision() Revision { return v.revision }

// Outcome returns the parent terminal outcome and whether it exists.
func (v ParentRecordView) Outcome() (ParentTerminalOutcome, bool) { return v.outcome, v.hasOutcome }

// ParentAdmission is one authorized parent inspect-or-claim result.
type ParentAdmission struct {
	kind   AdmissionKind
	record ParentRecordView
}

// Kind returns whether the parent was claimed, observed, or replayed.
func (a ParentAdmission) Kind() AdmissionKind { return a.kind }

// Record returns the coherent detached parent observation.
func (a ParentAdmission) Record() ParentRecordView { return a.record }

// PhaseRecordView is a detached coherent linked-phase observation.
type PhaseRecordView struct {
	parentScope Scope
	parentKey   CommandKey
	kind        PhaseKind
	ordinal     uint8
	state       CommandState
	revision    Revision
	outcome     TerminalOutcome
	hasOutcome  bool
}

// ParentScope returns the exact parent scope.
func (v PhaseRecordView) ParentScope() Scope { return v.parentScope }

// ParentKey returns the opaque parent key.
func (v PhaseRecordView) ParentKey() CommandKey { return v.parentKey }

// Kind returns the privately derived phase kind.
func (v PhaseRecordView) Kind() PhaseKind { return v.kind }

// Ordinal returns the fixed phase ordinal.
func (v PhaseRecordView) Ordinal() uint8 { return v.ordinal }

// State returns the durable linked-phase state.
func (v PhaseRecordView) State() CommandState { return v.state }

// Revision returns the durable linked-phase revision.
func (v PhaseRecordView) Revision() Revision { return v.revision }

// Outcome returns the phase terminal outcome and whether it exists.
func (v PhaseRecordView) Outcome() (TerminalOutcome, bool) { return v.outcome, v.hasOutcome }

// PhaseAdmission is one linked-phase inspect-or-claim result.
type PhaseAdmission struct {
	kind   AdmissionKind
	record PhaseRecordView
}

// Kind returns whether the phase was claimed, observed, or replayed.
func (a PhaseAdmission) Kind() AdmissionKind { return a.kind }

// Record returns the coherent detached phase observation.
func (a PhaseAdmission) Record() PhaseRecordView { return a.record }
