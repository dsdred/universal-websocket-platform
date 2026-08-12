package runtimelaunchflow

import (
	"context"
	"errors"

	"github.com/dsdred/universal-websocket-platform/internal/configurationloader"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

var (
	// ErrInvalidManagedBinding indicates one incomplete, foreign, or reused
	// managed Start binding.
	ErrInvalidManagedBinding = errors.New(
		"invalid Runtime managed start binding",
	)
	// ErrInvalidOwnerClaimView indicates one incomplete Owner-issued claim
	// view.
	ErrInvalidOwnerClaimView = errors.New(
		"invalid Runtime Owner claim view",
	)
	// ErrInvalidContinuation indicates a nil or failing StartClaimContinuation
	// at construction.
	ErrInvalidContinuation = errors.New(
		"invalid Runtime Start claim continuation",
	)
)

// ManagedStartBinding retains the historical internal name during migration.
// The authoritative value belongs to runtimeorchestrationbinding.
type ManagedStartBinding = runtimeorchestrationbinding.StartExecutionBinding

// NewManagedStartBinding validates and constructs one primitive immutable
// binding. Authorization must be exact ActivateExactTarget; expected revision,
// generation, and rendezvous must be non-zero. The composition-owned execution
// generation is never allocated or derived here.
func NewManagedStartBinding(
	authorization runtimeorchestrationbinding.OrchestrationAuthorizationRequest,
	expectedRevision runtimeorchestrationbinding.AggregateRevision,
	generation runtimeorchestrationbinding.ExecutionGeneration,
	rendezvous runtimeorchestrationbinding.StartRendezvous,
) (ManagedStartBinding, error) {
	binding, err := runtimeorchestrationbinding.NewPrimitiveStartExecutionBinding(
		authorization, expectedRevision, generation, rendezvous,
	)
	if err != nil {
		return ManagedStartBinding{}, ErrInvalidManagedBinding
	}
	return binding, nil
}

// OwnerClaimView is one immutable five-identity claim view extracted from
// the Owner-issued LoadRequest immediately after a successful authentic
// Owner.PrepareStart and before any Load, Build, or Launcher work. It
// carries no preparation token, Host, Snapshot, context cancellation
// authority, parent/phase or Stop permit, or mutable Owner state.
type OwnerClaimView struct {
	workspaceID                  uint64
	configurationID              uint64
	runtimeInstanceID            runtimeconfigload.RuntimeInstanceID
	launchAttemptID              runtimeconfigload.LaunchAttemptID
	targetConfigurationVersionID uint64
}

// NewOwnerClaimView validates and constructs one immutable claim view from
// one exact Owner-issued LoadRequest. The target version identity is the
// LoadRequest's exact ConfigurationVersion identity.
func NewOwnerClaimView(
	request runtimeconfigload.LoadRequest,
) (OwnerClaimView, error) {
	view := OwnerClaimView{
		workspaceID:                  request.WorkspaceID(),
		configurationID:              request.ConfigurationID(),
		runtimeInstanceID:            request.RuntimeInstanceID(),
		launchAttemptID:              request.LaunchAttemptID(),
		targetConfigurationVersionID: request.ConfigurationVersionID(),
	}
	if !view.valid() {
		return OwnerClaimView{}, ErrInvalidOwnerClaimView
	}
	return view, nil
}

// WorkspaceID returns the exact Workspace identity.
func (v OwnerClaimView) WorkspaceID() uint64 { return v.workspaceID }

// ConfigurationID returns the exact Configuration identity.
func (v OwnerClaimView) ConfigurationID() uint64 { return v.configurationID }

// RuntimeInstanceID returns the exact Runtime Instance identity.
func (v OwnerClaimView) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID {
	return v.runtimeInstanceID
}

// LaunchAttemptID returns the Owner-issued Launch Attempt identity.
func (v OwnerClaimView) LaunchAttemptID() runtimeconfigload.LaunchAttemptID {
	return v.launchAttemptID
}

// TargetConfigurationVersionID returns the exact target Configuration
// Version identity.
func (v OwnerClaimView) TargetConfigurationVersionID() uint64 {
	return v.targetConfigurationVersionID
}

func (v OwnerClaimView) valid() bool {
	return v.workspaceID != 0 && v.configurationID != 0 &&
		v.runtimeInstanceID != "" && v.launchAttemptID != "" &&
		v.targetConfigurationVersionID != 0
}

// StartClaimContinuation is the stateless private synchronous Slice-3 service
// that a managed Flow binds exactly once at construction. StartNoClaim closes
// the exact binding when cancellation, rejection, or failure occurs before an
// Owner claim. AfterOwnerClaim resolves the exact Owner-issued claim to one
// closed Continue, StopConverged, BindingFailed, or Blocked decision on the
// original synchronous stack before any Load, Build, or Launcher work. The
// service receives no permit and publishes no lifecycle, phase, or parent
// terminal outcome.
type StartClaimContinuation interface {
	StartNoClaim(context.Context, ManagedStartBinding, runtimeorchestrationbinding.StartNoClaimCause) error
	AfterOwnerClaim(context.Context, ManagedStartBinding, OwnerClaimView) (StartClaimOutcome, error)
}

// StartClaimOutcome is the closed Flow-owned continuation decision.
type StartClaimOutcome uint8

const (
	// StartClaimContinue permits Load, Build, and Owner.Start.
	StartClaimContinue StartClaimOutcome = iota + 1
	// StartClaimStopConverged reports exact Owner convergence by Stop.
	StartClaimStopConverged
	// StartClaimBindingFailed reports a definitive no-binding identity outcome.
	StartClaimBindingFailed
	// StartClaimBlocked reports indeterminate continuation evidence.
	StartClaimBlocked
)

// ManagedFlow is one immutable Runtime launch Flow extension that validates
// one per-invocation ManagedStartBinding before any Owner mutation, calls
// one bound StartClaimContinuation exactly once after an authentic
// Owner.PrepareStart, and invokes no other managed behavior. The existing
// unmanaged New construction and Start semantics remain unchanged; only this
// construction carries the managed seam.
type ManagedFlow struct {
	flow         *Flow
	continuation StartClaimContinuation
}

// NewManaged constructs one immutable managed Runtime launch Flow bound to
// one Owner, one Loader, and exactly one StartClaimContinuation. It creates
// no registry entry, mutable current-operation slot, goroutine, or detached
// callback.
func NewManaged(
	owner *runtimelifecycle.Owner,
	loader *configurationloader.Loader,
	continuation StartClaimContinuation,
) (*ManagedFlow, error) {
	if continuation == nil {
		return nil, ErrInvalidContinuation
	}
	flow, err := New(owner, loader)
	if err != nil {
		return nil, err
	}
	return &ManagedFlow{flow: flow, continuation: continuation}, nil
}

// StartManaged validates one per-invocation ManagedStartBinding, prepares the
// Owner claim, and resolves the closed continuation decision synchronously.
// Only Continue enters Load, Build, and prepared Owner.Start. StopConverged,
// BindingFailed, and Blocked perform zero Load and converge the authentic Owner
// preparation through their exact non-Continue mapping. The binding is never
// retained. Caller cancellation is consumed only before PrepareStart; after a
// successful claim, continuation and Owner convergence preserve context values
// while ignoring caller cancellation and deadlines.
func (m *ManagedFlow) StartManaged(
	ctx context.Context,
	request runtimelifecycle.StartRequest,
	binding ManagedStartBinding,
) (runtimelifecycle.StartOutcome, error) {
	if m == nil || m.flow == nil || m.continuation == nil {
		return runtimelifecycle.StartOutcome{}, ErrInvalidFlow
	}
	if ctx == nil {
		return runtimelifecycle.StartOutcome{}, ErrInvalidStartContext
	}
	if err := ctx.Err(); err != nil {
		if binding.Valid() {
			if signalErr := invokeStartNoClaimSafely(m.continuation,
				context.WithoutCancel(ctx), binding, runtimeorchestrationbinding.StartNoClaimCancelled,
			); signalErr != nil {
				return runtimelifecycle.StartOutcome{}, signalErr
			}
		}
		return runtimelifecycle.StartOutcome{}, err
	}
	if !managedBindingMatchesRequest(binding, request) {
		if binding.Valid() {
			if signalErr := invokeStartNoClaimSafely(m.continuation,
				context.WithoutCancel(ctx), binding, runtimeorchestrationbinding.StartNoClaimRejected,
			); signalErr != nil {
				return runtimelifecycle.StartOutcome{}, signalErr
			}
		}
		return runtimelifecycle.StartOutcome{}, ErrInvalidManagedBinding
	}

	preparation, err := m.flow.owner.PrepareStart(request)
	if err != nil {
		cause := runtimeorchestrationbinding.StartNoClaimRejected
		if errors.Is(err, runtimelifecycle.ErrAttemptIDSourceFailed) {
			cause = runtimeorchestrationbinding.StartNoClaimFailed
		}
		if signalErr := invokeStartNoClaimSafely(m.continuation, context.WithoutCancel(ctx), binding, cause); signalErr != nil {
			return runtimelifecycle.StartOutcome{}, signalErr
		}
		return runtimelifecycle.StartOutcome{}, err
	}
	postClaimCtx := context.WithoutCancel(ctx)
	claimView, err := NewOwnerClaimView(preparation.LoadRequest())
	if err != nil {
		return m.convergeBlocked(postClaimCtx, preparation, ErrInvalidOwnerClaimView)
	}
	if claimView.RuntimeInstanceID() != binding.Authorization().RuntimeInstanceID() {
		return m.convergeBlocked(postClaimCtx, preparation, ErrInvalidManagedBinding)
	}
	outcome, continuationErr := invokeContinuationSafely(
		m.continuation, postClaimCtx, binding, claimView,
	)
	if !validContinuationResult(outcome, continuationErr) {
		outcome = StartClaimBlocked
		continuationErr = ErrInvalidContinuation
	}
	switch outcome {
	case StartClaimContinue:
		return m.flow.startPreparedWithContext(postClaimCtx, preparation)
	case StartClaimStopConverged:
		return m.flow.owner.Start(postClaimCtx, preparation, runtimelifecycle.PreparationResult{})
	case StartClaimBindingFailed:
		return m.flow.owner.Start(
			postClaimCtx, preparation, runtimelifecycle.FailedPreparation(continuationErr),
		)
	case StartClaimBlocked:
		return m.convergeBlocked(postClaimCtx, preparation, continuationErr)
	}
	return m.convergeBlocked(postClaimCtx, preparation, ErrInvalidContinuation)
}

func invokeStartNoClaimSafely(
	continuation StartClaimContinuation,
	ctx context.Context,
	binding ManagedStartBinding,
	cause runtimeorchestrationbinding.StartNoClaimCause,
) (err error) {
	defer func() {
		if recover() != nil {
			err = ErrInvalidContinuation
		}
	}()
	return continuation.StartNoClaim(ctx, binding, cause)
}

func invokeContinuationSafely(
	continuation StartClaimContinuation,
	ctx context.Context,
	binding ManagedStartBinding,
	view OwnerClaimView,
) (outcome StartClaimOutcome, err error) {
	defer func() {
		if recover() != nil {
			outcome = StartClaimBlocked
			err = ErrInvalidContinuation
		}
	}()
	return continuation.AfterOwnerClaim(ctx, binding, view)
}

func validContinuationResult(outcome StartClaimOutcome, err error) bool {
	switch outcome {
	case StartClaimContinue, StartClaimStopConverged:
		return err == nil
	case StartClaimBindingFailed, StartClaimBlocked:
		return err != nil
	default:
		return false
	}
}

func (m *ManagedFlow) convergeBlocked(
	ctx context.Context,
	preparation runtimelifecycle.LaunchPreparation,
	cause error,
) (runtimelifecycle.StartOutcome, error) {
	if cause == nil {
		cause = ErrInvalidContinuation
	}
	outcome, ownerErr := m.flow.owner.Start(
		ctx, preparation, runtimelifecycle.FailedPreparation(cause),
	)
	if ownerErr != nil {
		return outcome, ownerErr
	}
	return runtimelifecycle.StartOutcome{}, cause
}

func managedBindingMatchesRequest(
	binding runtimeorchestrationbinding.StartExecutionBinding,
	request runtimelifecycle.StartRequest,
) bool {
	if !binding.Valid() {
		return false
	}
	authorization := binding.Authorization()
	_, linked := binding.LinkedExecutionIdentity()
	actionValid := authorization.Action() == runtimeorchestrationbinding.OrchestrationActionActivateExactTarget
	if linked {
		actionValid = authorization.Action() == runtimeorchestrationbinding.OrchestrationActionReplaceExactTarget ||
			authorization.Action() == runtimeorchestrationbinding.OrchestrationActionRollbackExactTarget
	}
	return actionValid &&
		authorization.WorkspaceID() == request.WorkspaceID() &&
		authorization.ConfigurationID() == request.ConfigurationID() &&
		authorization.TargetConfigurationVersionID() == request.ConfigurationVersionID()
}
