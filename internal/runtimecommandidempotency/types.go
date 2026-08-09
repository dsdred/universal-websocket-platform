// Package runtimecommandidempotency implements the isolated primitive command
// boundary defined by Approved DP-015 and the partial parent/phase sequential
// core and Continue/pending-Stop command rendezvous defined by Approved DP-019.
// It does not provide managed Flow continuation, lifecycle integration,
// recovery, or persistence across process restart.
package runtimecommandidempotency

import (
	"context"
	"errors"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

var (
	// ErrInvalidSubmission reports an incomplete context, scope, key, intent,
	// authorization function, or terminal outcome.
	ErrInvalidSubmission = errors.New("invalid runtime command submission")
	// ErrCommandKeyConflict reports reuse of one command identity for a
	// different immutable intent.
	ErrCommandKeyConflict = errors.New("runtime command key conflicts with immutable intent")
	// ErrInstanceBlocked reports a non-terminal command that does not permit the
	// requested tracked-Start Stop exception.
	ErrInstanceBlocked = errors.New("runtime instance has an unresolved command")
	// ErrIllegalPhaseOrder reports a linked phase request outside the finite
	// optional StopOld followed by StartTarget sequence.
	ErrIllegalPhaseOrder = errors.New("runtime command parent phase order is invalid")
	// ErrBoundaryExpired reports an operation attempted through a Boundary that
	// is no longer the active storage-client generation.
	ErrBoundaryExpired = errors.New("runtime command boundary client expired")
	// ErrIndeterminateExecution reports a lifecycle callback that did not return
	// a definitive semantic outcome. The durable command remains Claimed.
	ErrIndeterminateExecution = errors.New("runtime command execution outcome is indeterminate")
)

// CommandKey is an opaque caller-supplied identity within one exact Scope.
type CommandKey string

// Operation identifies the state-changing management operation.
type Operation string

const (
	// OperationStart identifies exact-version Start.
	OperationStart Operation = "start"
	// OperationStop identifies Stop for one exact Runtime Instance target.
	OperationStop Operation = "stop"
	// OperationReplace identifies replacement with one exact target version.
	OperationReplace Operation = "replace"
	// OperationRollback identifies rollback to one exact target version.
	OperationRollback Operation = "rollback"
)

// Scope is the exact authorization and command-identity namespace.
type Scope struct {
	domain            string
	workspaceID       uint64
	configurationID   uint64
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID
	operation         Operation
}

// NewScope constructs one exact command scope.
func NewScope(
	domain string,
	workspaceID uint64,
	configurationID uint64,
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID,
	operation Operation,
) (Scope, error) {
	scope := Scope{
		domain:            domain,
		workspaceID:       workspaceID,
		configurationID:   configurationID,
		runtimeInstanceID: runtimeInstanceID,
		operation:         operation,
	}
	if !scope.valid() {
		return Scope{}, ErrInvalidSubmission
	}
	return scope, nil
}

// Domain returns the operational management domain.
func (s Scope) Domain() string { return s.domain }

// WorkspaceID returns the exact Workspace identity.
func (s Scope) WorkspaceID() uint64 { return s.workspaceID }

// ConfigurationID returns the exact Configuration identity.
func (s Scope) ConfigurationID() uint64 { return s.configurationID }

// RuntimeInstanceID returns the exact Runtime Instance identity.
func (s Scope) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID {
	return s.runtimeInstanceID
}

// Operation returns the state-changing operation identity.
func (s Scope) Operation() Operation { return s.operation }

func (s Scope) valid() bool {
	return s.domain != "" && s.workspaceID != 0 && s.configurationID != 0 &&
		s.runtimeInstanceID != "" &&
		(s.operation == OperationStart || s.operation == OperationStop ||
			s.operation == OperationReplace || s.operation == OperationRollback)
}

func (s Scope) validPrimitive() bool {
	return s.valid() && (s.operation == OperationStart || s.operation == OperationStop)
}

func (s Scope) validParent() bool {
	return s.valid() && (s.operation == OperationReplace || s.operation == OperationRollback)
}

type instanceScope struct {
	domain            string
	workspaceID       uint64
	configurationID   uint64
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID
}

func (s Scope) instanceScope() instanceScope {
	return instanceScope{
		domain:            s.domain,
		workspaceID:       s.workspaceID,
		configurationID:   s.configurationID,
		runtimeInstanceID: s.runtimeInstanceID,
	}
}

// Intent is the immutable normalized semantic input bound to a command key.
// Start, Replace, and Rollback carry one exact Configuration Version identity;
// Stop carries no inferred version or mutable observation. Publication,
// Configuration membership, and rollback eligibility are upstream checks.
type Intent struct {
	operation              Operation
	configurationVersionID uint64
}

// NewStartIntent constructs an exact-version Start intent.
func NewStartIntent(configurationVersionID uint64) (Intent, error) {
	if configurationVersionID == 0 {
		return Intent{}, ErrInvalidSubmission
	}
	return Intent{
		operation:              OperationStart,
		configurationVersionID: configurationVersionID,
	}, nil
}

// NewStopIntent constructs a Stop intent for the exact target in Scope.
func NewStopIntent() Intent { return Intent{operation: OperationStop} }

// NewReplaceIntent constructs a replacement intent for one exact target
// Configuration Version. Publication and Configuration membership are
// upstream preconditions and are not inferred by this package.
func NewReplaceIntent(configurationVersionID uint64) (Intent, error) {
	return newParentIntent(OperationReplace, configurationVersionID)
}

// NewRollbackIntent constructs a rollback intent for one exact target
// Configuration Version. Historical-target policy is an upstream precondition.
func NewRollbackIntent(configurationVersionID uint64) (Intent, error) {
	return newParentIntent(OperationRollback, configurationVersionID)
}

func newParentIntent(operation Operation, configurationVersionID uint64) (Intent, error) {
	if configurationVersionID == 0 ||
		(operation != OperationReplace && operation != OperationRollback) {
		return Intent{}, ErrInvalidSubmission
	}
	return Intent{operation: operation, configurationVersionID: configurationVersionID}, nil
}

// Operation returns the intent operation.
func (i Intent) Operation() Operation { return i.operation }

// ConfigurationVersionID returns the exact Start, Replace, or Rollback target
// version identity, or zero for Stop.
func (i Intent) ConfigurationVersionID() uint64 { return i.configurationVersionID }

func (i Intent) validFor(scope Scope) bool {
	if i.operation != scope.operation {
		return false
	}
	if i.operation == OperationStart {
		return i.configurationVersionID != 0
	}
	return i.operation == OperationStop && i.configurationVersionID == 0
}

func (i Intent) validParentFor(scope Scope) bool {
	return scope.validParent() && i.operation == scope.operation &&
		i.configurationVersionID != 0
}

// Authorize checks the current caller for the exact action, Target, and intent.
// It is invoked on every submission, including observation and replay.
type Authorize func(context.Context, Scope, Intent) error

// CommandState is the durable monotonic command state.
type CommandState string

const (
	// CommandStateClaimed means terminal replay facts are not durable yet.
	CommandStateClaimed CommandState = "claimed"
	// CommandStateTerminal means one immutable semantic outcome is replayable.
	CommandStateTerminal CommandState = "terminal"
)

// OutcomeCategory is a bounded semantic result category. It is intentionally
// not a raw error or transport response.
type OutcomeCategory string

const (
	// OutcomeSucceeded reports definitive operation success.
	OutcomeSucceeded OutcomeCategory = "succeeded"
	// OutcomeRejected reports a definitive no-mutation lifecycle rejection.
	OutcomeRejected OutcomeCategory = "rejected"
	// OutcomeFailed reports a definitive lifecycle failure.
	OutcomeFailed OutcomeCategory = "failed"
)

// TerminalOutcome is the immutable redacted semantic result stored for replay.
type TerminalOutcome struct {
	category        OutcomeCategory
	launchAttemptID runtimeconfigload.LaunchAttemptID
}

// NewTerminalOutcome validates and constructs a replay-safe outcome. The
// optional LaunchAttemptID is an observed identity fact, not a new allocation.
func NewTerminalOutcome(
	category OutcomeCategory,
	launchAttemptID runtimeconfigload.LaunchAttemptID,
) (TerminalOutcome, error) {
	if category != OutcomeSucceeded && category != OutcomeRejected && category != OutcomeFailed {
		return TerminalOutcome{}, ErrInvalidSubmission
	}
	return TerminalOutcome{category: category, launchAttemptID: launchAttemptID}, nil
}

// Category returns the stable semantic category.
func (o TerminalOutcome) Category() OutcomeCategory { return o.category }

// LaunchAttemptID returns an observed Launch Attempt identity, if one exists.
func (o TerminalOutcome) LaunchAttemptID() runtimeconfigload.LaunchAttemptID {
	return o.launchAttemptID
}

func (o TerminalOutcome) valid() bool {
	return o.category == OutcomeSucceeded || o.category == OutcomeRejected || o.category == OutcomeFailed
}

// AdmissionKind classifies one authorized submission. No Admission exposes a
// live execution permit; a newly committed claim is delegated synchronously by
// Boundary.Execute through its private permit before that call returns.
type AdmissionKind string

const (
	// AdmissionClaimed reports the path that newly committed a claim. Its Record
	// is Terminal after definitive publication or Claimed with an accompanying
	// execution error after an indeterminate outcome. It exposes no permit.
	AdmissionClaimed AdmissionKind = "claimed"
	// AdmissionInProgress reports a matching non-terminal record without permit.
	AdmissionInProgress AdmissionKind = "in-progress"
	// AdmissionReplay reports a matching terminal record without permit.
	AdmissionReplay AdmissionKind = "replay"
)

// Revision is a monotonically advancing command concurrency token.
type Revision uint64

// RecordView is a detached coherent command observation.
type RecordView struct {
	scope      Scope
	key        CommandKey
	intent     Intent
	state      CommandState
	revision   Revision
	outcome    TerminalOutcome
	hasOutcome bool
}

// Scope returns the exact command scope.
func (v RecordView) Scope() Scope { return v.scope }

// Key returns the opaque command key.
func (v RecordView) Key() CommandKey { return v.key }

// Intent returns the immutable semantic intent.
func (v RecordView) Intent() Intent { return v.intent }

// State returns the durable command state.
func (v RecordView) State() CommandState { return v.state }

// Revision returns the durable command revision.
func (v RecordView) Revision() Revision { return v.revision }

// Outcome returns the terminal outcome and whether it exists.
func (v RecordView) Outcome() (TerminalOutcome, bool) { return v.outcome, v.hasOutcome }

// Admission is the result of one authorized inspect-or-claim submission.
type Admission struct {
	kind   AdmissionKind
	record RecordView
}

// Kind returns whether the submission claimed, observed, or replayed.
func (a Admission) Kind() AdmissionKind { return a.kind }

// Record returns the coherent durable command observation.
func (a Admission) Record() RecordView { return a.record }
