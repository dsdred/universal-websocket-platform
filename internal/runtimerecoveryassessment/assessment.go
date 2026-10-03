// Package runtimerecoveryassessment implements the repository-private DP-017
// read-only recovery assessment for one exact Runtime Instance.
//
// It creates no recovery claim, permit, admission state, lifecycle work,
// command publication, or reconciliation mutation.
package runtimerecoveryassessment

import (
	"context"
	"errors"
	"sort"

	"github.com/dsdred/universal-websocket-platform/internal/runtimecommandidempotency"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecontainment"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeexecutionevidence"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeidentity"
)

var errInvalidTarget = errors.New("invalid runtime recovery assessment target")

// Classification is the closed read-only DP-017 assessment result.
type Classification uint8

const (
	// Unknown reports that exact coherent truth could not be established.
	Unknown Classification = iota
	// Clean reports a coherent set with no active attempt and no non-terminal
	// primitive, parent, or phase command.
	Clean
	// CommandOnly reports non-terminal command truth while exact aggregate facts
	// prove that no Launch Attempt has ever been claimed.
	CommandOnly
	// UnboundAttempt reports one exact active attempt whose execution-generation
	// binding is definitively absent.
	UnboundAttempt
	// ExecutionTerminated reports exact authoritative termination evidence for
	// the bound active execution generation.
	ExecutionTerminated
	// ResourceAbsence reports exact authoritative covered-resource absence for
	// a Stop-in-progress attempt without shutdown-completion proof.
	ResourceAbsence
	// ShutdownCompleted reports exact authoritative Host shutdown-completion
	// evidence for the selected attempt and generation.
	ShutdownCompleted
)

// Target is the exact operational and containment scope assessed by one call.
type Target struct {
	operationalDomain string
	containmentDomain runtimecontainment.Domain
	workspaceID       uint64
	configurationID   uint64
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID
}

// NewTarget validates and constructs one exact assessment target. The
// operational domain must be the canonical containment-domain representation.
func NewTarget(
	operationalDomain string,
	workspaceID uint64,
	configurationID uint64,
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID,
) (Target, error) {
	domain, err := runtimecontainment.ParseDomain(operationalDomain)
	if err != nil || workspaceID == 0 || configurationID == 0 || runtimeInstanceID == "" {
		return Target{}, errInvalidTarget
	}
	return Target{
		operationalDomain: operationalDomain,
		containmentDomain: domain,
		workspaceID:       workspaceID,
		configurationID:   configurationID,
		runtimeInstanceID: runtimeInstanceID,
	}, nil
}

// OperationalDomain returns the canonical operational domain.
func (t Target) OperationalDomain() string { return t.operationalDomain }

// WorkspaceID returns the exact Workspace identity.
func (t Target) WorkspaceID() uint64 { return t.workspaceID }

// ConfigurationID returns the exact Configuration identity.
func (t Target) ConfigurationID() uint64 { return t.configurationID }

// RuntimeInstanceID returns the exact Runtime Instance identity.
func (t Target) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID {
	return t.runtimeInstanceID
}

func (t Target) valid() bool {
	parsed, err := runtimecontainment.ParseDomain(t.operationalDomain)
	return err == nil && parsed == t.containmentDomain && t.workspaceID != 0 &&
		t.configurationID != 0 && t.runtimeInstanceID != ""
}

// IdentityReader is the read-only DP-014 aggregate/history source required by
// the assessment.
type IdentityReader interface {
	ReadRuntimeInstance(runtimeconfigload.RuntimeInstanceID) (runtimeidentity.RuntimeInstanceView, error)
	ReadLaunchAttemptHistory(runtimeconfigload.RuntimeInstanceID) ([]runtimeidentity.LaunchAttemptRecord, error)
}

// CommandSnapshotReader is the read-only DP-015 complete per-Instance source
// required by the assessment.
type CommandSnapshotReader interface {
	ReadAssessmentSnapshot(string, uint64, uint64, runtimeconfigload.RuntimeInstanceID) (runtimecommandidempotency.AssessmentSnapshot, error)
}

// Evidence is the immutable closed observation consumed from one exact
// invocation-scoped containment evidence handle.
type Evidence interface {
	Kind() runtimeexecutionevidence.OutcomeKind
	Reason() runtimeexecutionevidence.UnknownReason
}

// EvidenceQuery reads and consumes one fresh exact-tuple evidence handle. A
// containment composition is adapted by querying it and returning Consume's
// immutable outcome from this function.
type EvidenceQuery func(
	context.Context,
	runtimecontainment.Domain,
	runtimeconfigload.RuntimeInstanceID,
	runtimeconfigload.LaunchAttemptID,
	runtimeidentity.ExecutionGeneration,
	runtimeexecutionevidence.Question,
) Evidence

// Result is one detached assessment observation. It contains only immutable
// facts and exposes no recovery, admission, lifecycle, or command authority.
type Result struct {
	classification    Classification
	aggregateRevision runtimeidentity.Revision
	commands          runtimecommandidempotency.AssessmentSnapshot
	attemptID         runtimeconfigload.LaunchAttemptID
	generation        runtimeidentity.ExecutionGeneration
	evidenceKind      runtimeexecutionevidence.OutcomeKind
	evidenceReason    runtimeexecutionevidence.UnknownReason
	hasEvidence       bool
}

// Classification returns the closed assessment class.
func (r Result) Classification() Classification { return r.classification }

// AggregateRevision returns the stable DP-014 revision observed by the
// assessment, or zero when no coherent result was established.
func (r Result) AggregateRevision() runtimeidentity.Revision { return r.aggregateRevision }

// CommandSnapshot returns the complete detached DP-015 snapshot bound to the
// stable assessment, or the zero snapshot for Unknown.
func (r Result) CommandSnapshot() runtimecommandidempotency.AssessmentSnapshot {
	return r.commands
}

// Attempt returns the selected exact Launch Attempt and execution generation,
// if this classification required one.
func (r Result) Attempt() (
	runtimeconfigload.LaunchAttemptID,
	runtimeidentity.ExecutionGeneration,
	bool,
) {
	return r.attemptID, r.generation, r.attemptID != ""
}

// Evidence returns the exact consumed evidence kind and closed Unknown reason
// when evidence was required for the classification.
func (r Result) Evidence() (
	runtimeexecutionevidence.OutcomeKind,
	runtimeexecutionevidence.UnknownReason,
	bool,
) {
	return r.evidenceKind, r.evidenceReason, r.hasEvidence
}

// Assess performs one mutation-free stable read sandwich over DP-014 identity,
// DP-015 command truth, and, when required, exact attempt/generation evidence.
// Missing, stale, contradictory, foreign, cancelled, or indeterminate input
// returns Unknown and no partial positive result.
func Assess(
	ctx context.Context,
	target Target,
	identities IdentityReader,
	commands CommandSnapshotReader,
	query EvidenceQuery,
) Result {
	if ctx == nil || !target.valid() || identities == nil || commands == nil || query == nil {
		return Result{classification: Unknown}
	}

	identityBefore, ok := readIdentity(ctx, target, identities)
	if !ok {
		return Result{classification: Unknown}
	}
	commandBefore, err := commands.ReadAssessmentSnapshot(
		target.operationalDomain,
		target.workspaceID,
		target.configurationID,
		target.runtimeInstanceID,
	)
	if err != nil || ctx.Err() != nil || !validCommandSnapshot(target, identityBefore, commandBefore) {
		return Result{classification: Unknown}
	}

	provisional := classify(ctx, target, identityBefore, commandBefore, query)

	identityAfter, ok := readIdentity(ctx, target, identities)
	if !ok || !identityObservationsEqual(identityBefore, identityAfter) {
		return Result{classification: Unknown}
	}
	commandAfter, err := commands.ReadAssessmentSnapshot(
		target.operationalDomain,
		target.workspaceID,
		target.configurationID,
		target.runtimeInstanceID,
	)
	if err != nil || ctx.Err() != nil || !validCommandSnapshot(target, identityAfter, commandAfter) ||
		!commandSnapshotsEqual(commandBefore, commandAfter) {
		return Result{classification: Unknown}
	}
	if provisional.classification == Unknown {
		return provisional
	}
	provisional.aggregateRevision = identityBefore.view.Revision()
	provisional.commands = commandBefore
	return provisional
}

type identityObservation struct {
	view    runtimeidentity.RuntimeInstanceView
	history []runtimeidentity.LaunchAttemptRecord
}

func readIdentity(ctx context.Context, target Target, reader IdentityReader) (identityObservation, bool) {
	if ctx.Err() != nil {
		return identityObservation{}, false
	}
	before, err := reader.ReadRuntimeInstance(target.runtimeInstanceID)
	if err != nil || ctx.Err() != nil {
		return identityObservation{}, false
	}
	history, err := reader.ReadLaunchAttemptHistory(target.runtimeInstanceID)
	if err != nil || ctx.Err() != nil {
		return identityObservation{}, false
	}
	after, err := reader.ReadRuntimeInstance(target.runtimeInstanceID)
	if err != nil || ctx.Err() != nil || before != after {
		return identityObservation{}, false
	}
	observation := identityObservation{view: before, history: append([]runtimeidentity.LaunchAttemptRecord(nil), history...)}
	if !validIdentityObservation(target, observation) {
		return identityObservation{}, false
	}
	sort.Slice(observation.history, func(i, j int) bool {
		return observation.history[i].LaunchAttemptID() < observation.history[j].LaunchAttemptID()
	})
	return observation, true
}

func validIdentityObservation(target Target, observation identityObservation) bool {
	view := observation.view
	if view.WorkspaceID() != target.workspaceID || view.ConfigurationID() != target.configurationID ||
		view.RuntimeInstanceID() != target.runtimeInstanceID || view.Revision() == 0 ||
		!validDesired(view.DesiredState()) || !validActual(view.ActualState()) {
		return false
	}
	seen := make(map[runtimeconfigload.LaunchAttemptID]struct{}, len(observation.history))
	nonterminal := 0
	var nonterminalID runtimeconfigload.LaunchAttemptID
	for _, attempt := range observation.history {
		if attempt.RuntimeInstanceID() != target.runtimeInstanceID || attempt.LaunchAttemptID() == "" ||
			attempt.ConfigurationVersionID() == 0 || !validAttempt(attempt) {
			return false
		}
		if _, exists := seen[attempt.LaunchAttemptID()]; exists {
			return false
		}
		seen[attempt.LaunchAttemptID()] = struct{}{}
		if !attemptTerminal(attempt.Phase()) {
			nonterminal++
			nonterminalID = attempt.LaunchAttemptID()
		}
	}
	activeID, active := view.ActiveAttempt()
	if active {
		return activeID != "" && nonterminal == 1 && nonterminalID == activeID
	}
	return activeID == "" && nonterminal == 0
}

func validAttempt(attempt runtimeidentity.LaunchAttemptRecord) bool {
	phase := attempt.Phase()
	basis := attempt.TerminalCompletionBasis()
	if !validPhase(phase) {
		return false
	}
	if !attemptTerminal(phase) {
		if basis != "" {
			return false
		}
		return nonterminalPhaseGenerationCoherent(phase, attempt.ExecutionGeneration())
	}
	switch basis {
	case runtimeidentity.TerminalCompletionOwnerShutdownCompleted:
		return phase == runtimeidentity.AttemptPhaseStopped && attempt.ExecutionGeneration() != ""
	case runtimeidentity.TerminalCompletionNoHostProduced:
		return true
	case runtimeidentity.TerminalCompletionRecoveryReconciled:
		return phase == runtimeidentity.AttemptPhaseFailed
	default:
		return false
	}
}

func nonterminalPhaseGenerationCoherent(
	phase runtimeidentity.AttemptPhase,
	generation runtimeidentity.ExecutionGeneration,
) bool {
	return generation != "" ||
		(phase != runtimeidentity.AttemptPhaseLaunching && phase != runtimeidentity.AttemptPhaseRunning)
}

func validDesired(state runtimeidentity.DesiredState) bool {
	return state == runtimeidentity.DesiredStateStopped || state == runtimeidentity.DesiredStateStarted
}

func validActual(state runtimeidentity.ActualState) bool {
	switch state {
	case runtimeidentity.ActualStateStopped, runtimeidentity.ActualStateClaimed,
		runtimeidentity.ActualStateRunning, runtimeidentity.ActualStateStopping,
		runtimeidentity.ActualStateFailed:
		return true
	default:
		return false
	}
}

func validPhase(phase runtimeidentity.AttemptPhase) bool {
	switch phase {
	case runtimeidentity.AttemptPhaseClaimed, runtimeidentity.AttemptPhaseLaunching,
		runtimeidentity.AttemptPhaseRunning, runtimeidentity.AttemptPhaseStopping,
		runtimeidentity.AttemptPhaseStopped, runtimeidentity.AttemptPhaseFailed:
		return true
	default:
		return false
	}
}

func attemptTerminal(phase runtimeidentity.AttemptPhase) bool {
	return phase == runtimeidentity.AttemptPhaseStopped || phase == runtimeidentity.AttemptPhaseFailed
}

func classify(
	ctx context.Context,
	target Target,
	identity identityObservation,
	commands runtimecommandidempotency.AssessmentSnapshot,
	query EvidenceQuery,
) Result {
	nonterminal := nonterminalCommandCount(commands)
	activeID, active := identity.view.ActiveAttempt()
	if !active && nonterminal == 0 {
		return Result{classification: Clean}
	}
	if !active && len(identity.history) == 0 && nonterminal > 0 {
		return Result{classification: CommandOnly}
	}

	var attempt runtimeidentity.LaunchAttemptRecord
	found := false
	if active {
		attempt, found = findAttempt(identity.history, activeID)
	} else if reference, ok := parentRecoveryAttempt(commands); ok {
		attempt, found = findAttempt(identity.history, reference)
	}
	if !found {
		return Result{classification: Unknown}
	}
	if attempt.ExecutionGeneration() == "" {
		if active {
			return Result{classification: UnboundAttempt, attemptID: attempt.LaunchAttemptID()}
		}
		return Result{classification: Unknown}
	}

	if attempt.Phase() == runtimeidentity.AttemptPhaseStopping {
		shutdown := consumeEvidence(ctx, target, attempt, runtimeexecutionevidence.ShutdownCompletion, query)
		if shutdown.kind == runtimeexecutionevidence.HostShutdownCompleted {
			return evidenceResult(ShutdownCompleted, attempt, shutdown)
		}
		if shutdown.kind != runtimeexecutionevidence.Unknown ||
			shutdown.reason != runtimeexecutionevidence.Absent {
			return evidenceResult(Unknown, attempt, shutdown)
		}
		absence := consumeEvidence(ctx, target, attempt, runtimeexecutionevidence.CoveredResourceAbsence, query)
		if absence.kind == runtimeexecutionevidence.CoveredResourcesAbsent {
			return evidenceResult(ResourceAbsence, attempt, absence)
		}
		return evidenceResult(Unknown, attempt, absence)
	}

	if attempt.Phase() == runtimeidentity.AttemptPhaseStopped &&
		attempt.TerminalCompletionBasis() == runtimeidentity.TerminalCompletionOwnerShutdownCompleted {
		shutdown := consumeEvidence(ctx, target, attempt, runtimeexecutionevidence.ShutdownCompletion, query)
		if shutdown.kind == runtimeexecutionevidence.HostShutdownCompleted {
			return evidenceResult(ShutdownCompleted, attempt, shutdown)
		}
		return evidenceResult(Unknown, attempt, shutdown)
	}

	generation := consumeEvidence(ctx, target, attempt, runtimeexecutionevidence.GenerationStatus, query)
	if generation.kind == runtimeexecutionevidence.GenerationTerminated {
		return evidenceResult(ExecutionTerminated, attempt, generation)
	}
	return evidenceResult(Unknown, attempt, generation)
}

func consumeEvidence(
	ctx context.Context,
	target Target,
	attempt runtimeidentity.LaunchAttemptRecord,
	question runtimeexecutionevidence.Question,
	query EvidenceQuery,
) (outcome evidenceObservation) {
	defer func() {
		if recover() != nil {
			outcome = evidenceObservation{
				kind: runtimeexecutionevidence.Unknown, reason: runtimeexecutionevidence.Indeterminate,
			}
		}
	}()
	evidence := query(
		ctx,
		target.containmentDomain,
		target.runtimeInstanceID,
		attempt.LaunchAttemptID(),
		attempt.ExecutionGeneration(),
		question,
	)
	if evidence == nil {
		return evidenceObservation{
			kind: runtimeexecutionevidence.Unknown, reason: runtimeexecutionevidence.Indeterminate,
		}
	}
	outcome = evidenceObservation{kind: evidence.Kind(), reason: evidence.Reason()}
	if !outcome.valid() {
		return evidenceObservation{
			kind: runtimeexecutionevidence.Unknown, reason: runtimeexecutionevidence.Contradictory,
		}
	}
	return outcome
}

type evidenceObservation struct {
	kind   runtimeexecutionevidence.OutcomeKind
	reason runtimeexecutionevidence.UnknownReason
}

func (o evidenceObservation) valid() bool {
	if o.kind == runtimeexecutionevidence.Unknown {
		switch o.reason {
		case runtimeexecutionevidence.Absent, runtimeexecutionevidence.Unavailable,
			runtimeexecutionevidence.Stale, runtimeexecutionevidence.ScopeMismatch,
			runtimeexecutionevidence.Contradictory, runtimeexecutionevidence.Indeterminate,
			runtimeexecutionevidence.Cancelled, runtimeexecutionevidence.UnsupportedTopology,
			runtimeexecutionevidence.GuaranteeNotDeclared:
			return true
		default:
			return false
		}
	}
	if o.reason != "" {
		return false
	}
	switch o.kind {
	case runtimeexecutionevidence.GenerationLive,
		runtimeexecutionevidence.GenerationTerminated,
		runtimeexecutionevidence.CoveredResourcesAbsent,
		runtimeexecutionevidence.HostShutdownCompleted,
		runtimeexecutionevidence.LiveUnownedExecution:
		return true
	default:
		return false
	}
}

func evidenceResult(
	classification Classification,
	attempt runtimeidentity.LaunchAttemptRecord,
	evidence evidenceObservation,
) Result {
	return Result{
		classification: classification,
		attemptID:      attempt.LaunchAttemptID(),
		generation:     attempt.ExecutionGeneration(),
		evidenceKind:   evidence.kind,
		evidenceReason: evidence.reason,
		hasEvidence:    true,
	}
}

func findAttempt(
	history []runtimeidentity.LaunchAttemptRecord,
	id runtimeconfigload.LaunchAttemptID,
) (runtimeidentity.LaunchAttemptRecord, bool) {
	for _, attempt := range history {
		if attempt.LaunchAttemptID() == id {
			return attempt, true
		}
	}
	return runtimeidentity.LaunchAttemptRecord{}, false
}

func validCommandSnapshot(
	target Target,
	identity identityObservation,
	snapshot runtimecommandidempotency.AssessmentSnapshot,
) bool {
	if snapshot.Domain() != target.operationalDomain || snapshot.WorkspaceID() != target.workspaceID ||
		snapshot.ConfigurationID() != target.configurationID ||
		snapshot.RuntimeInstanceID() != target.runtimeInstanceID {
		return false
	}

	attempts := make(map[runtimeconfigload.LaunchAttemptID]runtimeidentity.LaunchAttemptRecord, len(identity.history))
	for _, attempt := range identity.history {
		attempts[attempt.LaunchAttemptID()] = attempt
	}
	activeAttempt, hasActiveAttempt := readActiveAttempt(identity, attempts)
	parents := make(map[commandIdentity]runtimecommandidempotency.ParentRecordView)
	primitiveSeen := make(map[commandIdentity]struct{})
	for _, record := range snapshot.PrimitiveRecords() {
		outcome := primitiveOutcome(record)
		if !scopeMatches(target, record.Scope()) || record.Key() == "" ||
			(record.Scope().Operation() != runtimecommandidempotency.OperationStart &&
				record.Scope().Operation() != runtimecommandidempotency.OperationStop) ||
			!intentMatches(record.Scope(), record.Intent()) ||
			!validRecordState(record.State(), record.Revision(), outcome) ||
			!primitiveAttemptLinkMatches(record, outcome, attempts) ||
			!activePrimitiveLinkMatches(record, activeAttempt, hasActiveAttempt) {
			return false
		}
		identity := commandIdentity{operation: record.Scope().Operation(), key: record.Key()}
		if _, duplicate := primitiveSeen[identity]; duplicate {
			return false
		}
		primitiveSeen[identity] = struct{}{}
	}
	for _, record := range snapshot.ParentRecords() {
		if !scopeMatches(target, record.Scope()) || record.Key() == "" ||
			(record.Scope().Operation() != runtimecommandidempotency.OperationReplace &&
				record.Scope().Operation() != runtimecommandidempotency.OperationRollback) ||
			!intentMatches(record.Scope(), record.Intent()) ||
			!validParentRecordState(record) {
			return false
		}
		identity := commandIdentity{operation: record.Scope().Operation(), key: record.Key()}
		if _, duplicate := parents[identity]; duplicate {
			return false
		}
		parents[identity] = record
	}
	phaseSeen := make(map[phaseIdentity]struct{})
	for _, record := range snapshot.PhaseRecords() {
		outcome := phaseOutcome(record)
		if !scopeMatches(target, record.ParentScope()) || record.ParentKey() == "" ||
			(record.ParentScope().Operation() != runtimecommandidempotency.OperationReplace &&
				record.ParentScope().Operation() != runtimecommandidempotency.OperationRollback) ||
			!validPhaseIdentity(record.Kind(), record.Ordinal()) ||
			!validRecordState(record.State(), record.Revision(), outcome) {
			return false
		}
		parent := commandIdentity{operation: record.ParentScope().Operation(), key: record.ParentKey()}
		parentRecord, exists := parents[parent]
		if !exists || !phaseAttemptLinkMatches(record, parentRecord, outcome, attempts) ||
			!activePhaseLinkMatches(record, parentRecord, activeAttempt, hasActiveAttempt) {
			return false
		}
		identity := phaseIdentity{parent: parent, kind: record.Kind(), ordinal: record.Ordinal()}
		if _, duplicate := phaseSeen[identity]; duplicate {
			return false
		}
		phaseSeen[identity] = struct{}{}
	}
	return true
}

func readActiveAttempt(
	identity identityObservation,
	attempts map[runtimeconfigload.LaunchAttemptID]runtimeidentity.LaunchAttemptRecord,
) (runtimeidentity.LaunchAttemptRecord, bool) {
	id, present := identity.view.ActiveAttempt()
	if !present {
		return runtimeidentity.LaunchAttemptRecord{}, false
	}
	attempt, exists := attempts[id]
	return attempt, exists
}

type outcomeState struct {
	present bool
	valid   bool
	attempt runtimeconfigload.LaunchAttemptID
}

func primitiveOutcome(record runtimecommandidempotency.RecordView) outcomeState {
	outcome, present := record.Outcome()
	return outcomeState{present: present, valid: validTerminalOutcome(outcome), attempt: outcome.LaunchAttemptID()}
}

func phaseOutcome(record runtimecommandidempotency.PhaseRecordView) outcomeState {
	outcome, present := record.Outcome()
	return outcomeState{present: present, valid: validTerminalOutcome(outcome), attempt: outcome.LaunchAttemptID()}
}

func primitiveAttemptLinkMatches(
	record runtimecommandidempotency.RecordView,
	outcome outcomeState,
	attempts map[runtimeconfigload.LaunchAttemptID]runtimeidentity.LaunchAttemptRecord,
) bool {
	if !outcome.present || outcome.attempt == "" {
		return true
	}
	attempt, exists := attempts[outcome.attempt]
	if !exists {
		return false
	}
	return record.Scope().Operation() != runtimecommandidempotency.OperationStart ||
		attempt.ConfigurationVersionID() == record.Intent().ConfigurationVersionID()
}

func activePrimitiveLinkMatches(
	record runtimecommandidempotency.RecordView,
	active runtimeidentity.LaunchAttemptRecord,
	hasActive bool,
) bool {
	return !hasActive || record.State() != runtimecommandidempotency.CommandStateClaimed ||
		record.Scope().Operation() != runtimecommandidempotency.OperationStart ||
		record.Intent().ConfigurationVersionID() == active.ConfigurationVersionID()
}

func phaseAttemptLinkMatches(
	record runtimecommandidempotency.PhaseRecordView,
	parent runtimecommandidempotency.ParentRecordView,
	outcome outcomeState,
	attempts map[runtimeconfigload.LaunchAttemptID]runtimeidentity.LaunchAttemptRecord,
) bool {
	if !outcome.present || outcome.attempt == "" {
		return true
	}
	attempt, exists := attempts[outcome.attempt]
	if !exists {
		return false
	}
	return record.Kind() != runtimecommandidempotency.PhaseStartTarget ||
		attempt.ConfigurationVersionID() == parent.Intent().ConfigurationVersionID()
}

func activePhaseLinkMatches(
	record runtimecommandidempotency.PhaseRecordView,
	parent runtimecommandidempotency.ParentRecordView,
	active runtimeidentity.LaunchAttemptRecord,
	hasActive bool,
) bool {
	return !hasActive || record.State() != runtimecommandidempotency.CommandStateClaimed ||
		record.Kind() != runtimecommandidempotency.PhaseStartTarget ||
		parent.Intent().ConfigurationVersionID() == active.ConfigurationVersionID()
}

func validRecordState(
	state runtimecommandidempotency.CommandState,
	revision runtimecommandidempotency.Revision,
	outcome outcomeState,
) bool {
	if revision == 0 {
		return false
	}
	switch state {
	case runtimecommandidempotency.CommandStateClaimed:
		return !outcome.present
	case runtimecommandidempotency.CommandStateTerminal:
		return outcome.present && outcome.valid
	default:
		return false
	}
}

func validParentRecordState(record runtimecommandidempotency.ParentRecordView) bool {
	if record.Revision() == 0 {
		return false
	}
	outcome, present := record.Outcome()
	switch record.State() {
	case runtimecommandidempotency.CommandStateClaimed:
		return !present
	case runtimecommandidempotency.CommandStateTerminal:
		return present && validParentOutcome(outcome.Category())
	default:
		return false
	}
}

func validTerminalOutcome(outcome runtimecommandidempotency.TerminalOutcome) bool {
	switch outcome.Category() {
	case runtimecommandidempotency.OutcomeSucceeded, runtimecommandidempotency.OutcomeSatisfied,
		runtimecommandidempotency.OutcomeRejected, runtimecommandidempotency.OutcomeFailed:
		return true
	default:
		return false
	}
}

func validParentOutcome(category runtimecommandidempotency.ParentOutcomeCategory) bool {
	switch category {
	case runtimecommandidempotency.ParentOutcomeSucceeded,
		runtimecommandidempotency.ParentOutcomeSatisfied,
		runtimecommandidempotency.ParentOutcomeStopped,
		runtimecommandidempotency.ParentOutcomeCancelled,
		runtimecommandidempotency.ParentOutcomeRejected,
		runtimecommandidempotency.ParentOutcomeFailed:
		return true
	default:
		return false
	}
}

func validPhaseIdentity(kind runtimecommandidempotency.PhaseKind, ordinal uint8) bool {
	return (kind == runtimecommandidempotency.PhaseStopOld && ordinal == 0) ||
		(kind == runtimecommandidempotency.PhaseStartTarget && ordinal == 1)
}

func scopeMatches(target Target, scope runtimecommandidempotency.Scope) bool {
	return scope.Domain() == target.operationalDomain && scope.WorkspaceID() == target.workspaceID &&
		scope.ConfigurationID() == target.configurationID &&
		scope.RuntimeInstanceID() == target.runtimeInstanceID
}

func intentMatches(scope runtimecommandidempotency.Scope, intent runtimecommandidempotency.Intent) bool {
	if intent.Operation() != scope.Operation() {
		return false
	}
	switch scope.Operation() {
	case runtimecommandidempotency.OperationStop:
		return intent.ConfigurationVersionID() == 0
	case runtimecommandidempotency.OperationStart,
		runtimecommandidempotency.OperationReplace,
		runtimecommandidempotency.OperationRollback:
		return intent.ConfigurationVersionID() != 0
	default:
		return false
	}
}

type commandIdentity struct {
	operation runtimecommandidempotency.Operation
	key       runtimecommandidempotency.CommandKey
}

type phaseIdentity struct {
	parent  commandIdentity
	kind    runtimecommandidempotency.PhaseKind
	ordinal uint8
}

func nonterminalCommandCount(snapshot runtimecommandidempotency.AssessmentSnapshot) int {
	count := 0
	for _, record := range snapshot.PrimitiveRecords() {
		if record.State() != runtimecommandidempotency.CommandStateTerminal {
			count++
		}
	}
	for _, record := range snapshot.ParentRecords() {
		if record.State() != runtimecommandidempotency.CommandStateTerminal {
			count++
		}
	}
	for _, record := range snapshot.PhaseRecords() {
		if record.State() != runtimecommandidempotency.CommandStateTerminal {
			count++
		}
	}
	return count
}

func parentRecoveryAttempt(snapshot runtimecommandidempotency.AssessmentSnapshot) (
	runtimeconfigload.LaunchAttemptID,
	bool,
) {
	var parent *runtimecommandidempotency.ParentRecordView
	for _, record := range snapshot.PrimitiveRecords() {
		if record.State() != runtimecommandidempotency.CommandStateTerminal {
			return "", false
		}
	}
	parents := snapshot.ParentRecords()
	for i := range parents {
		if parents[i].State() != runtimecommandidempotency.CommandStateTerminal {
			if parent != nil {
				return "", false
			}
			parent = &parents[i]
		}
	}
	if parent == nil {
		return "", false
	}

	var reference runtimeconfigload.LaunchAttemptID
	phaseCount := 0
	for _, phase := range snapshot.PhaseRecords() {
		if phase.ParentScope() != parent.Scope() || phase.ParentKey() != parent.Key() {
			continue
		}
		phaseCount++
		if phase.State() != runtimecommandidempotency.CommandStateTerminal {
			return "", false
		}
		outcome, present := phase.Outcome()
		if !present || outcome.LaunchAttemptID() == "" {
			return "", false
		}
		if reference == "" {
			reference = outcome.LaunchAttemptID()
		} else if reference != outcome.LaunchAttemptID() {
			return "", false
		}
	}
	return reference, phaseCount > 0 && reference != ""
}

func identityObservationsEqual(left, right identityObservation) bool {
	if left.view != right.view || len(left.history) != len(right.history) {
		return false
	}
	for i := range left.history {
		if left.history[i] != right.history[i] {
			return false
		}
	}
	return true
}

func commandSnapshotsEqual(
	left runtimecommandidempotency.AssessmentSnapshot,
	right runtimecommandidempotency.AssessmentSnapshot,
) bool {
	if left.Domain() != right.Domain() || left.WorkspaceID() != right.WorkspaceID() ||
		left.ConfigurationID() != right.ConfigurationID() ||
		left.RuntimeInstanceID() != right.RuntimeInstanceID() {
		return false
	}
	leftPrimitive, rightPrimitive := left.PrimitiveRecords(), right.PrimitiveRecords()
	leftParents, rightParents := left.ParentRecords(), right.ParentRecords()
	leftPhases, rightPhases := left.PhaseRecords(), right.PhaseRecords()
	if len(leftPrimitive) != len(rightPrimitive) || len(leftParents) != len(rightParents) ||
		len(leftPhases) != len(rightPhases) {
		return false
	}
	for i := range leftPrimitive {
		if leftPrimitive[i] != rightPrimitive[i] {
			return false
		}
	}
	for i := range leftParents {
		if leftParents[i] != rightParents[i] {
			return false
		}
	}
	for i := range leftPhases {
		if leftPhases[i] != rightPhases[i] {
			return false
		}
	}
	return true
}
