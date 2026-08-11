// Package runtimeorchestrationbinding owns immutable dependency-leaf values
// passed across Runtime orchestration, command, identity, management, and
// launch-flow boundaries.
package runtimeorchestrationbinding

import (
	"context"
	"errors"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

// ErrInvalidBinding reports an incomplete or internally inconsistent value.
var ErrInvalidBinding = errors.New("invalid Runtime orchestration binding")

// OrchestrationAction identifies one exact caller authorization action.
type OrchestrationAction string

const (
	// OrchestrationActionActivateExactTarget authorizes primitive exact Start.
	OrchestrationActionActivateExactTarget OrchestrationAction = "activate-exact-target"
	// OrchestrationActionReplaceExactTarget authorizes exact replacement.
	OrchestrationActionReplaceExactTarget OrchestrationAction = "replace-with-exact-target"
	// OrchestrationActionRollbackExactTarget authorizes exact rollback.
	OrchestrationActionRollbackExactTarget OrchestrationAction = "rollback-to-exact-target"
)

// Valid reports whether the action is one of the closed orchestration set.
func (a OrchestrationAction) Valid() bool {
	return a == OrchestrationActionActivateExactTarget ||
		a == OrchestrationActionReplaceExactTarget ||
		a == OrchestrationActionRollbackExactTarget
}

// OrchestrationAuthorizationRequest is the exact immutable six-field policy
// input. It contains no Principal, credential, mutable observation, or cached
// authority.
type OrchestrationAuthorizationRequest struct {
	operationalDomain            string
	workspaceID                  uint64
	configurationID              uint64
	runtimeInstanceID            runtimeconfigload.RuntimeInstanceID
	action                       OrchestrationAction
	targetConfigurationVersionID uint64
}

// NewOrchestrationAuthorizationRequest validates one exact authorization tuple.
func NewOrchestrationAuthorizationRequest(
	operationalDomain string,
	workspaceID uint64,
	configurationID uint64,
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID,
	action OrchestrationAction,
	targetConfigurationVersionID uint64,
) (OrchestrationAuthorizationRequest, error) {
	request := OrchestrationAuthorizationRequest{
		operationalDomain:            operationalDomain,
		workspaceID:                  workspaceID,
		configurationID:              configurationID,
		runtimeInstanceID:            runtimeInstanceID,
		action:                       action,
		targetConfigurationVersionID: targetConfigurationVersionID,
	}
	if !request.Valid() {
		return OrchestrationAuthorizationRequest{}, ErrInvalidBinding
	}
	return request, nil
}

// OperationalDomain returns the exact opaque management domain.
func (r OrchestrationAuthorizationRequest) OperationalDomain() string {
	return r.operationalDomain
}

// WorkspaceID returns the exact Workspace identity.
func (r OrchestrationAuthorizationRequest) WorkspaceID() uint64 { return r.workspaceID }

// ConfigurationID returns the exact Configuration identity.
func (r OrchestrationAuthorizationRequest) ConfigurationID() uint64 {
	return r.configurationID
}

// RuntimeInstanceID returns the exact Runtime Instance identity.
func (r OrchestrationAuthorizationRequest) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID {
	return r.runtimeInstanceID
}

// Action returns the exact caller authorization action.
func (r OrchestrationAuthorizationRequest) Action() OrchestrationAction { return r.action }

// TargetConfigurationVersionID returns the exact target version identity.
func (r OrchestrationAuthorizationRequest) TargetConfigurationVersionID() uint64 {
	return r.targetConfigurationVersionID
}

// Valid reports whether all six exact authorization fields are valid.
func (r OrchestrationAuthorizationRequest) Valid() bool {
	return r.operationalDomain != "" && r.workspaceID != 0 &&
		r.configurationID != 0 && r.runtimeInstanceID != "" &&
		r.action.Valid() && r.targetConfigurationVersionID != 0
}

// AuthorizeOrchestration checks current caller authority for one exact request.
type AuthorizeOrchestration func(context.Context, OrchestrationAuthorizationRequest) error

// AggregateRevision is a lossless neutral aggregate revision proof.
type AggregateRevision uint64

// ExecutionGeneration is an opaque composition-owned execution generation.
type ExecutionGeneration string

// StartRendezvous is an opaque comparable identity for command-owned mutable
// rendezvous state. It exposes no signaling or waiting capability.
type StartRendezvous struct{ identity string }

// NewStartRendezvous constructs a non-zero opaque rendezvous identity.
func NewStartRendezvous(identity string) (StartRendezvous, error) {
	rendezvous := StartRendezvous{identity: identity}
	if !rendezvous.valid() {
		return StartRendezvous{}, ErrInvalidBinding
	}
	return rendezvous, nil
}

func (r StartRendezvous) valid() bool { return r.identity != "" }

// ParentCommandIdentity is one exact replacement or rollback parent identity.
type ParentCommandIdentity struct {
	authorization OrchestrationAuthorizationRequest
	commandKey    string
}

// NewParentCommandIdentity constructs one exact parent identity.
func NewParentCommandIdentity(
	authorization OrchestrationAuthorizationRequest,
	commandKey string,
) (ParentCommandIdentity, error) {
	parent := ParentCommandIdentity{authorization: authorization, commandKey: commandKey}
	if !parent.valid() {
		return ParentCommandIdentity{}, ErrInvalidBinding
	}
	return parent, nil
}

func (p ParentCommandIdentity) valid() bool {
	return p.authorization.Valid() && p.commandKey != "" &&
		(p.authorization.Action() == OrchestrationActionReplaceExactTarget ||
			p.authorization.Action() == OrchestrationActionRollbackExactTarget)
}

// Authorization returns the exact immutable parent authorization tuple.
func (p ParentCommandIdentity) Authorization() OrchestrationAuthorizationRequest {
	return p.authorization
}

// CommandKey returns the opaque parent command key.
func (p ParentCommandIdentity) CommandKey() string { return p.commandKey }

// StartTargetPhaseIdentity is the command-boundary-derived ordinal-one phase.
type StartTargetPhaseIdentity struct {
	parent  ParentCommandIdentity
	kind    string
	ordinal uint8
}

// DeriveStartTargetPhaseIdentity derives the only legal linked Start phase.
func DeriveStartTargetPhaseIdentity(parent ParentCommandIdentity) (StartTargetPhaseIdentity, error) {
	phase := StartTargetPhaseIdentity{parent: parent, kind: "start-target", ordinal: 1}
	if !phase.valid() {
		return StartTargetPhaseIdentity{}, ErrInvalidBinding
	}
	return phase, nil
}

func (p StartTargetPhaseIdentity) valid() bool {
	return p.parent.valid() && p.kind == "start-target" && p.ordinal == 1
}

// Parent returns the exact parent identity from which this phase was derived.
func (p StartTargetPhaseIdentity) Parent() ParentCommandIdentity { return p.parent }

// Ordinal returns the fixed StartTarget ordinal one.
func (p StartTargetPhaseIdentity) Ordinal() uint8 { return p.ordinal }

// LinkedExecutionIdentity is an all-or-none parent plus derived phase identity.
type LinkedExecutionIdentity struct {
	parent ParentCommandIdentity
	phase  StartTargetPhaseIdentity
}

// NewLinkedExecutionIdentity validates an exact parent/phase pair.
func NewLinkedExecutionIdentity(
	parent ParentCommandIdentity,
	phase StartTargetPhaseIdentity,
) (LinkedExecutionIdentity, error) {
	linked := LinkedExecutionIdentity{parent: parent, phase: phase}
	if !linked.valid() {
		return LinkedExecutionIdentity{}, ErrInvalidBinding
	}
	return linked, nil
}

func (l LinkedExecutionIdentity) valid() bool {
	return l.parent.valid() && l.phase.valid() && l.phase.parent == l.parent
}

// Parent returns the exact linked parent identity.
func (l LinkedExecutionIdentity) Parent() ParentCommandIdentity { return l.parent }

// Phase returns the exact derived StartTarget phase identity.
func (l LinkedExecutionIdentity) Phase() StartTargetPhaseIdentity { return l.phase }

// StartExecutionBinding is one immutable per-invocation managed Start binding.
type StartExecutionBinding struct {
	authorization    OrchestrationAuthorizationRequest
	expectedRevision AggregateRevision
	generation       ExecutionGeneration
	linked           LinkedExecutionIdentity
	hasLinked        bool
	rendezvous       StartRendezvous
}

// NewPrimitiveStartExecutionBinding constructs a primitive Activate binding.
func NewPrimitiveStartExecutionBinding(
	authorization OrchestrationAuthorizationRequest,
	expectedRevision AggregateRevision,
	generation ExecutionGeneration,
	rendezvous StartRendezvous,
) (StartExecutionBinding, error) {
	binding := StartExecutionBinding{
		authorization: authorization, expectedRevision: expectedRevision,
		generation: generation, rendezvous: rendezvous,
	}
	if !binding.Valid() {
		return StartExecutionBinding{}, ErrInvalidBinding
	}
	return binding, nil
}

// NewLinkedStartExecutionBinding constructs a Replace/Rollback StartTarget binding.
func NewLinkedStartExecutionBinding(
	authorization OrchestrationAuthorizationRequest,
	expectedRevision AggregateRevision,
	generation ExecutionGeneration,
	linked LinkedExecutionIdentity,
	rendezvous StartRendezvous,
) (StartExecutionBinding, error) {
	binding := StartExecutionBinding{
		authorization: authorization, expectedRevision: expectedRevision,
		generation: generation, linked: linked, hasLinked: true, rendezvous: rendezvous,
	}
	if !binding.Valid() {
		return StartExecutionBinding{}, ErrInvalidBinding
	}
	return binding, nil
}

// Authorization returns the exact six-field authorization tuple.
func (b StartExecutionBinding) Authorization() OrchestrationAuthorizationRequest {
	return b.authorization
}

// ExpectedAggregateRevision returns the lossless expected aggregate revision.
func (b StartExecutionBinding) ExpectedAggregateRevision() AggregateRevision {
	return b.expectedRevision
}

// ExecutionGeneration returns the exact composition generation.
func (b StartExecutionBinding) ExecutionGeneration() ExecutionGeneration {
	return b.generation
}

// LinkedExecutionIdentity returns the linked identity when this is StartTarget.
func (b StartExecutionBinding) LinkedExecutionIdentity() (LinkedExecutionIdentity, bool) {
	return b.linked, b.hasLinked
}

// Rendezvous returns the opaque command-owned rendezvous identity.
func (b StartExecutionBinding) Rendezvous() StartRendezvous { return b.rendezvous }

// Valid reports whether the complete binding is internally consistent.
func (b StartExecutionBinding) Valid() bool {
	if !b.authorization.Valid() || b.expectedRevision == 0 || b.generation == "" ||
		!b.rendezvous.valid() {
		return false
	}
	if b.hasLinked {
		return b.linked.valid() &&
			b.linked.parent.authorization == b.authorization &&
			(b.authorization.Action() == OrchestrationActionReplaceExactTarget ||
				b.authorization.Action() == OrchestrationActionRollbackExactTarget)
	}
	return b.linked == (LinkedExecutionIdentity{}) &&
		b.authorization.Action() == OrchestrationActionActivateExactTarget
}
