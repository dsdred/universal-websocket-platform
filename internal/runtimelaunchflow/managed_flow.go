package runtimelaunchflow

import (
	"context"
	"errors"

	"github.com/dsdred/universal-websocket-platform/internal/configurationloader"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeidentity"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle"
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

// ManagedStartBinding is the immutable per-invocation evidence that one exact
// authorized orchestration submission and one exact Owner-issued Launch
// Attempt claim are being managed through one composition-owned execution
// generation and one opaque rendezvous. It carries no primitive, parent,
// phase, or Stop permit, no preparation token, no Host or Snapshot, no
// context cancellation authority, and no mutable Owner state. It is validated
// before any Owner mutation, is retained only on the synchronous call stack
// that invokes it, is invoked at most once, and is invalidated on return.
//
// The zero value is never valid. Construct it with NewManagedStartBinding.
type ManagedStartBinding struct {
	expectedRevision runtimeidentity.Revision
	generation       runtimeidentity.ExecutionGeneration
	rendezvous       runtimeconfigload.StartRendezvous
}

// NewManagedStartBinding validates and constructs one immutable binding. The
// expected revision and generation must be non-zero and the rendezvous must
// be non-zero. The composition-owned execution generation is never allocated
// or derived here; it is supplied by composition.
func NewManagedStartBinding(
	expectedRevision runtimeidentity.Revision,
	generation runtimeidentity.ExecutionGeneration,
	rendezvous runtimeconfigload.StartRendezvous,
) (ManagedStartBinding, error) {
	binding := ManagedStartBinding{
		expectedRevision: expectedRevision,
		generation:       generation,
		rendezvous:       rendezvous,
	}
	if !binding.valid() {
		return ManagedStartBinding{}, ErrInvalidManagedBinding
	}
	return binding, nil
}

// ExpectedRevision returns the expected aggregate revision proof.
func (b ManagedStartBinding) ExpectedRevision() runtimeidentity.Revision {
	return b.expectedRevision
}

// ExecutionGeneration returns the composition-owned execution generation.
func (b ManagedStartBinding) ExecutionGeneration() runtimeidentity.ExecutionGeneration {
	return b.generation
}

// Rendezvous returns the opaque Start rendezvous handle.
func (b ManagedStartBinding) Rendezvous() runtimeconfigload.StartRendezvous {
	return b.rendezvous
}

func (b ManagedStartBinding) valid() bool {
	return b.expectedRevision != 0 &&
		b.generation != "" &&
		b.rendezvous != (runtimeconfigload.StartRendezvous{})
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

// StartClaimContinuation is the stateless private synchronous service that a
// managed Flow binds exactly once at construction. Its per-invocation
// contract is one exact AfterOwnerClaim decision called on the original
// synchronous call stack immediately after Owner.PrepareStart and before any
// Load, Build, or Launcher work. It never receives a permit and never
// publishes a lifecycle, phase, or parent terminal outcome. The DP-020
// OwnerClaim-to-DP-014 binding sequence is the next slice.
type StartClaimContinuation interface {
	AfterOwnerClaim(context.Context, ManagedStartBinding, OwnerClaimView) error
}

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

// StartManaged performs one synchronous PrepareStart, Load, Build, and
// Owner.Start operation under one per-invocation ManagedStartBinding. The
// binding is validated before any Owner mutation and is never retained by
// the Flow. Caller cancellation is observed exactly once before PrepareStart.
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
		return runtimelifecycle.StartOutcome{}, err
	}
	if !binding.valid() {
		return runtimelifecycle.StartOutcome{}, ErrInvalidManagedBinding
	}

	preparation, err := m.flow.owner.PrepareStart(request)
	if err != nil {
		return runtimelifecycle.StartOutcome{}, err
	}
	if preparation.Context().Err() != nil {
		return convergeStoppedPreparation(m.flow.owner, preparation)
	}
	claimView, err := NewOwnerClaimView(preparation.LoadRequest())
	if err != nil {
		return convergeStoppedPreparation(m.flow.owner, preparation)
	}
	if err := m.continuation.AfterOwnerClaim(ctx, binding, claimView); err != nil {
		// Converge the claimed attempt through the authentic Owner preparation:
		// the failure reaches Owner.Start with its preparation token so that
		// no claimed attempt is dropped and no lifecycle state is leaked. The
		// exact continuation error is returned unchanged on success; only a
		// failure of the Owner-side convergence itself replaces it.
		outcome, ownerErr := m.flow.owner.Start(
			context.Background(), preparation, runtimelifecycle.FailedPreparation(err),
		)
		if ownerErr != nil {
			return outcome, ownerErr
		}
		return runtimelifecycle.StartOutcome{}, err
	}
	return m.flow.Start(ctx, request)
}
