package runtimecommandidempotency

import (
	"context"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

// Source-compatible aliases keep the accepted internal command surface while
// the authoritative cross-layer values live in the dependency-leaf package.
type OrchestrationAction = runtimeorchestrationbinding.OrchestrationAction

const (
	OrchestrationActionActivateExactTarget = runtimeorchestrationbinding.OrchestrationActionActivateExactTarget
	OrchestrationActionReplaceExactTarget  = runtimeorchestrationbinding.OrchestrationActionReplaceExactTarget
	OrchestrationActionRollbackExactTarget = runtimeorchestrationbinding.OrchestrationActionRollbackExactTarget
)

type OrchestrationAuthorizationRequest = runtimeorchestrationbinding.OrchestrationAuthorizationRequest
type AuthorizeOrchestration = runtimeorchestrationbinding.AuthorizeOrchestration

// NewOrchestrationAuthorizationRequest validates the exact six-field request.
// The operational domain is never inferred or defaulted.
func NewOrchestrationAuthorizationRequest(
	operationalDomain string,
	workspaceID uint64,
	configurationID uint64,
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID,
	action OrchestrationAction,
	targetConfigurationVersionID uint64,
) (OrchestrationAuthorizationRequest, error) {
	request, err := runtimeorchestrationbinding.NewOrchestrationAuthorizationRequest(
		operationalDomain, workspaceID, configurationID, runtimeInstanceID,
		action, targetConfigurationVersionID,
	)
	if err != nil {
		return OrchestrationAuthorizationRequest{}, ErrInvalidSubmission
	}
	return request, nil
}

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
	request, err := NewOrchestrationAuthorizationRequest(
		scope.domain, scope.workspaceID, scope.configurationID,
		scope.runtimeInstanceID, action, intent.configurationVersionID,
	)
	return request, err == nil
}

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
