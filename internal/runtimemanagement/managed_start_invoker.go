package runtimemanagement

import (
	"context"
	"errors"

	"github.com/dsdred/universal-websocket-platform/internal/runtimelaunchflow"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

var (
	// ErrInvalidManagedStartInvoker reports invalid immutable invoker construction.
	ErrInvalidManagedStartInvoker = errors.New("invalid Runtime managed start invoker")
	// ErrInvalidManagedStartInvocation reports an invalid managed Start invocation.
	ErrInvalidManagedStartInvocation = errors.New("invalid Runtime managed start invocation")
)

// ManagedStartInvoker validates one exact management scope before delegating
// synchronously to its preconstructed managed Flow.
type ManagedStartInvoker struct {
	operationalDomain string
	target            Target
	flow              *runtimelaunchflow.ManagedFlow
}

// NewManagedStartInvoker constructs one immutable exact-scope managed Start
// invoker around a borrowed preconstructed managed Flow.
func NewManagedStartInvoker(
	operationalDomain string,
	target Target,
	flow *runtimelaunchflow.ManagedFlow,
) (*ManagedStartInvoker, error) {
	if operationalDomain == "" || !target.valid() || flow == nil {
		return nil, ErrInvalidManagedStartInvoker
	}
	return &ManagedStartInvoker{
		operationalDomain: operationalDomain,
		target:            target,
		flow:              flow,
	}, nil
}

// InvokeManagedStart validates one exact request and structural execution
// binding, then returns the stored managed Flow's exact outcome and error.
func (i *ManagedStartInvoker) InvokeManagedStart(
	ctx context.Context,
	request runtimelifecycle.StartRequest,
	binding runtimeorchestrationbinding.StartExecutionBinding,
) (runtimelifecycle.StartOutcome, error) {
	if !i.valid() || ctx == nil || request.WorkspaceID() == 0 ||
		request.ConfigurationID() == 0 || request.ConfigurationVersionID() == 0 ||
		!binding.Valid() {
		return runtimelifecycle.StartOutcome{}, ErrInvalidManagedStartInvocation
	}

	authorization := binding.Authorization()
	if authorization.OperationalDomain() != i.operationalDomain ||
		authorization.WorkspaceID() != i.target.WorkspaceID() ||
		authorization.ConfigurationID() != i.target.ConfigurationID() ||
		authorization.RuntimeInstanceID() != i.target.RuntimeInstanceID() ||
		request.WorkspaceID() != authorization.WorkspaceID() ||
		request.ConfigurationID() != authorization.ConfigurationID() ||
		request.ConfigurationVersionID() != authorization.TargetConfigurationVersionID() {
		return runtimelifecycle.StartOutcome{}, ErrInvalidManagedStartInvocation
	}

	return i.flow.StartManaged(ctx, request, binding)
}

func (i *ManagedStartInvoker) valid() bool {
	return i != nil && i.operationalDomain != "" && i.target.valid() && i.flow != nil
}
