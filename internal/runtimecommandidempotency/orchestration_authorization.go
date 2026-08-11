package runtimecommandidempotency

import (
	"context"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

// OrchestrationAction is one exact caller authorization action for Runtime
// activation orchestration. Its values are fixed; no other action exists and
// no action falls back to another.
type OrchestrationAction string

const (
	// OrchestrationActionActivateExactTarget authorizes the existing primitive
	// exact-version Start submission as one activation.
	OrchestrationActionActivateExactTarget OrchestrationAction = "activate-exact-target"
	// OrchestrationActionReplaceExactTarget authorizes one replacement parent
	// submission with one exact target version.
	OrchestrationActionReplaceExactTarget OrchestrationAction = "replace-with-exact-target"
	// OrchestrationActionRollbackExactTarget authorizes one rollback parent
	// submission to one exact historical target version.
	OrchestrationActionRollbackExactTarget OrchestrationAction = "rollback-to-exact-target"
)

func (a OrchestrationAction) valid() bool {
	switch a {
	case OrchestrationActionActivateExactTarget,
		OrchestrationActionReplaceExactTarget,
		OrchestrationActionRollbackExactTarget:
		return true
	default:
		return false
	}
}

// OrchestrationAuthorizationRequest is one immutable validated authorization
// input for one exact orchestration intent. Per the DP-020 single-node
// baseline it carries exactly the five fields below; no operational domain,
// Principal, credential, or aggregate observation is a durable intent field.
type OrchestrationAuthorizationRequest struct {
	workspaceID                  uint64
	configurationID              uint64
	runtimeInstanceID            runtimeconfigload.RuntimeInstanceID
	action                       OrchestrationAction
	targetConfigurationVersionID uint64
}

// NewOrchestrationAuthorizationRequest validates and constructs one exact
// authorization request. Every identity must be non-zero and exact, the
// action must be one of the defined values, and the target version identity
// must be non-zero. A Published-state and Configuration-membership check of
// the target version is an upstream caller precondition, not performed here.
func NewOrchestrationAuthorizationRequest(
	workspaceID uint64,
	configurationID uint64,
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID,
	action OrchestrationAction,
	targetConfigurationVersionID uint64,
) (OrchestrationAuthorizationRequest, error) {
	request := OrchestrationAuthorizationRequest{
		workspaceID:                  workspaceID,
		configurationID:              configurationID,
		runtimeInstanceID:            runtimeInstanceID,
		action:                       action,
		targetConfigurationVersionID: targetConfigurationVersionID,
	}
	if !request.valid() {
		return OrchestrationAuthorizationRequest{}, ErrInvalidSubmission
	}
	return request, nil
}

// WorkspaceID returns the exact Workspace identity.
func (r OrchestrationAuthorizationRequest) WorkspaceID() uint64 {
	return r.workspaceID
}

// ConfigurationID returns the exact Configuration identity.
func (r OrchestrationAuthorizationRequest) ConfigurationID() uint64 {
	return r.configurationID
}

// RuntimeInstanceID returns the exact Runtime Instance identity.
func (r OrchestrationAuthorizationRequest) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID {
	return r.runtimeInstanceID
}

// Action returns the exact caller authorization action.
func (r OrchestrationAuthorizationRequest) Action() OrchestrationAction {
	return r.action
}

// TargetConfigurationVersionID returns the exact target Configuration Version
// identity.
func (r OrchestrationAuthorizationRequest) TargetConfigurationVersionID() uint64 {
	return r.targetConfigurationVersionID
}

func (r OrchestrationAuthorizationRequest) valid() bool {
	return r.workspaceID != 0 && r.configurationID != 0 &&
		r.runtimeInstanceID != "" && r.action.valid() &&
		r.targetConfigurationVersionID != 0
}

// AuthorizeOrchestration checks the current caller for one exact orchestration
// authorization request. It is invoked on every submission, including replay,
// and its result is never cached as durable authority. A nil function is
// rejected as invalid before any submission state is touched. A non-nil error
// is returned unchanged and unwrapped; a panic is a policy defect that the
// caller boundary converts to a failed submission without mutation.
type AuthorizeOrchestration func(
	context.Context,
	OrchestrationAuthorizationRequest,
) error

// authorizeCommandMaps one already validated exact command Scope and Intent to
// its immutable orchestration authorization request without fallback between
// actions. The mapping is fixed: Start authorizes as
// OrchestrationActionActivateExactTarget, OperationReplace as
// OrchestrationActionReplaceExactTarget, and OperationRollback as
// OrchestrationActionRollbackExactTarget. Stop has no exact target version
// intent and is not an orchestration action; it returns false. An invalid
// pairing returns false without mutating anything.
func authorizeCommandMaps(
	scope Scope,
	intent Intent,
) (OrchestrationAuthorizationRequest, bool) {
	if !scope.valid() || !intent.validFor(scope) && !intent.validParentFor(scope) {
		return OrchestrationAuthorizationRequest{}, false
	}
	var action OrchestrationAction
	switch intent.operation {
	case OperationStart:
		action = OrchestrationActionActivateExactTarget
	case OperationReplace:
		action = OrchestrationActionReplaceExactTarget
	case OperationRollback:
		action = OrchestrationActionRollbackExactTarget
	default:
		return OrchestrationAuthorizationRequest{}, false
	}
	if intent.configurationVersionID == 0 {
		return OrchestrationAuthorizationRequest{}, false
	}
	request, err := NewOrchestrationAuthorizationRequest(
		scope.workspaceID,
		scope.configurationID,
		scope.runtimeInstanceID,
		action,
		intent.configurationVersionID,
	)
	if err != nil {
		return OrchestrationAuthorizationRequest{}, false
	}
	return request, true
}

// authorizeOrchestration invokes the policy-neutral authorizer exactly once
// for one already validated exact command Scope and Intent. A nil or panicking
// authorizer is a submission defect and returns an error without any command,
// aggregate, or lifecycle mutation; a deny or failure error is returned
// unchanged.
func authorizeOrchestration(
	authorize AuthorizeOrchestration,
	ctx context.Context,
	scope Scope,
	intent Intent,
) (err error) {
	if authorize == nil || ctx == nil {
		return ErrInvalidSubmission
	}
	request, ok := authorizeCommandMaps(scope, intent)
	if !ok {
		return ErrInvalidSubmission
	}
	defer func() {
		if recover() != nil {
			err = ErrInvalidSubmission
		}
	}()
	return authorize(ctx, request)
}
