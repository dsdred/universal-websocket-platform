// Package runtimeorchestrationcontinuation binds an authentic Owner claim to
// DP-014 identity persistence through command-owned managed gates.
package runtimeorchestrationcontinuation

import (
	"context"
	"errors"

	"github.com/dsdred/universal-websocket-platform/internal/runtimecommandidempotency"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeidentity"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelaunchflow"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

var (
	// ErrBindingFailed reports coherent proof that the exact binding was not committed.
	ErrBindingFailed = errors.New("runtime identity binding definitively failed")
	// ErrContinuationBlocked reports an indeterminate continuation result.
	ErrContinuationBlocked = errors.New("runtime identity binding outcome is indeterminate")
	// ErrAggregateScopeMismatch reports a foreign immutable aggregate binding.
	ErrAggregateScopeMismatch = errors.New("runtime identity aggregate scope mismatch")
	// ErrInspectionIncoherent reports changed A/history/B evidence.
	ErrInspectionIncoherent = errors.New("runtime identity inspection is incoherent")
)

// IdentityStore is the exact four-operation DP-014 seam required by the
// Owner-claim continuation.
type IdentityStore interface {
	ConditionalClaimLaunchAttempt(runtimeconfigload.RuntimeInstanceID, runtimeidentity.Revision, runtimeconfigload.LaunchAttemptID, uint64) (runtimeidentity.ClaimResult, error)
	ConditionalBindExecutionGeneration(runtimeconfigload.RuntimeInstanceID, runtimeidentity.Revision, runtimeconfigload.LaunchAttemptID, runtimeidentity.ExecutionGeneration) (runtimeidentity.PublishResult, error)
	ReadRuntimeInstance(runtimeconfigload.RuntimeInstanceID) (runtimeidentity.RuntimeInstanceView, error)
	ReadLaunchAttemptHistory(runtimeconfigload.RuntimeInstanceID) ([]runtimeidentity.LaunchAttemptRecord, error)
}

// Continuation is immutable and stateless; all mutable rendezvous and identity
// facts remain owned by their respective stores.
type Continuation struct {
	boundary *runtimecommandidempotency.Boundary
	identity IdentityStore
}

// New constructs one immutable stateless Owner-claim continuation over the
// exact command Boundary and four-operation identity store.
func New(
	boundary *runtimecommandidempotency.Boundary,
	identity IdentityStore,
) (*Continuation, error) {
	if boundary == nil || identity == nil {
		return nil, runtimelaunchflow.ErrInvalidContinuation
	}
	return &Continuation{boundary: boundary, identity: identity}, nil
}

// StartNoClaim closes the exact live managed rendezvous with one definitive
// neutral pre-Owner cause. It performs no identity write or lifecycle work.
func (c *Continuation) StartNoClaim(
	_ context.Context,
	binding runtimeorchestrationbinding.StartExecutionBinding,
	cause runtimeorchestrationbinding.StartNoClaimCause,
) error {
	if c == nil || c.boundary == nil || c.identity == nil {
		return ErrContinuationBlocked
	}
	return c.boundary.SignalManagedStartNoClaim(binding, cause)
}

// AfterOwnerClaim orders the exact Owner claim against Stop, validates the
// aggregate preflight, conditionally claims and binds identity, classifies any
// indeterminate write through a fresh A/history/B inspection, and seals the
// final Continue or BindingFailed disposition. It never loads configuration,
// publishes terminal identity facts, or owns lifecycle convergence.
func (c *Continuation) AfterOwnerClaim(
	_ context.Context,
	binding runtimeorchestrationbinding.StartExecutionBinding,
	view runtimelaunchflow.OwnerClaimView,
) (runtimelaunchflow.StartClaimOutcome, error) {
	if c == nil || c.boundary == nil || c.identity == nil || !binding.Valid() {
		return runtimelaunchflow.StartClaimBlocked, ErrContinuationBlocked
	}
	authorization := binding.Authorization()
	if view.WorkspaceID() == 0 || view.ConfigurationID() == 0 ||
		view.RuntimeInstanceID() == "" || view.LaunchAttemptID() == "" ||
		view.TargetConfigurationVersionID() == 0 ||
		authorization.WorkspaceID() != view.WorkspaceID() ||
		authorization.ConfigurationID() != view.ConfigurationID() ||
		authorization.RuntimeInstanceID() != view.RuntimeInstanceID() ||
		authorization.TargetConfigurationVersionID() != view.TargetConfigurationVersionID() {
		return runtimelaunchflow.StartClaimBlocked, ErrContinuationBlocked
	}
	early, err := c.boundary.ResolveManagedStartEarly(binding, view.LaunchAttemptID())
	if err != nil {
		return runtimelaunchflow.StartClaimBlocked, err
	}
	if early == runtimecommandidempotency.GateStopConverged {
		return runtimelaunchflow.StartClaimStopConverged, nil
	}
	if early != runtimecommandidempotency.GateClear {
		return runtimelaunchflow.StartClaimBlocked, ErrContinuationBlocked
	}

	expected := runtimeidentity.Revision(binding.ExpectedAggregateRevision())
	if expected == 0 || runtimeorchestrationbinding.AggregateRevision(expected) != binding.ExpectedAggregateRevision() {
		return c.blockedFinal(binding, ErrContinuationBlocked)
	}
	generation := runtimeidentity.ExecutionGeneration(binding.ExecutionGeneration())
	if generation == "" || runtimeorchestrationbinding.ExecutionGeneration(generation) != binding.ExecutionGeneration() {
		return c.blockedFinal(binding, ErrContinuationBlocked)
	}
	instanceID := view.RuntimeInstanceID()
	attemptID := view.LaunchAttemptID()
	preflight, preflightErr := c.identity.ReadRuntimeInstance(instanceID)
	if preflightErr != nil {
		return c.blockedFinal(binding, preflightErr)
	}
	if !aggregateScopeMatches(preflight, authorization) {
		return c.blockedFinal(binding, ErrAggregateScopeMismatch)
	}
	if preflight.Revision() != expected {
		return c.blockedFinal(binding, runtimeidentity.ErrStaleRevision)
	}
	claim, claimErr := c.identity.ConditionalClaimLaunchAttempt(
		instanceID, expected, attemptID, view.TargetConfigurationVersionID(),
	)
	var bindRevision runtimeidentity.Revision
	if claimErr == nil && claim.Committed() && claim.Revision() != 0 {
		bindRevision = claim.Revision()
	} else {
		proof, proofErr := c.inspect(authorization, attemptID, view.TargetConfigurationVersionID())
		if proofErr != nil {
			return c.blockedFinal(binding, proofErr)
		}
		if proof.exactClaimed {
			if proof.generation == generation {
				return c.continueFinal(binding)
			}
			// An indeterminate claim never supplies the committed ClaimResult
			// revision required by the bind operation. An empty or conflicting
			// generation therefore remains blocked rather than being retried.
			return c.blockedFinal(binding, firstError(claimErr, ErrContinuationBlocked))
		} else if proof.absent && proof.revision == expected {
			return c.bindingFailedFinal(binding, claimErr)
		} else {
			return c.blockedFinal(binding, firstError(claimErr, ErrContinuationBlocked))
		}
	}

	bound, bindErr := c.identity.ConditionalBindExecutionGeneration(
		instanceID, bindRevision, attemptID, generation,
	)
	if bindErr == nil && bound.Committed() {
		return c.continueFinal(binding)
	}
	proof, proofErr := c.inspect(authorization, attemptID, view.TargetConfigurationVersionID())
	if proofErr != nil {
		return c.blockedFinal(binding, proofErr)
	}
	if proof.exactClaimed && proof.generation == generation {
		return c.continueFinal(binding)
	}
	if proof.exactClaimed && proof.generation == "" && proof.revision == bindRevision {
		return c.bindingFailedFinal(binding, bindErr)
	}
	return c.blockedFinal(binding, firstError(bindErr, ErrContinuationBlocked))
}

type identityProof struct {
	revision     runtimeidentity.Revision
	exactClaimed bool
	absent       bool
	generation   runtimeidentity.ExecutionGeneration
}

func (c *Continuation) inspect(
	authorization runtimeorchestrationbinding.OrchestrationAuthorizationRequest,
	attemptID runtimeconfigload.LaunchAttemptID,
	versionID uint64,
) (identityProof, error) {
	instanceID := authorization.RuntimeInstanceID()
	a, err := c.identity.ReadRuntimeInstance(instanceID)
	if err != nil {
		return identityProof{}, err
	}
	history, err := c.identity.ReadLaunchAttemptHistory(instanceID)
	if err != nil {
		return identityProof{}, err
	}
	b, err := c.identity.ReadRuntimeInstance(instanceID)
	if err != nil {
		return identityProof{}, err
	}
	if !aggregateScopeMatches(a, authorization) || !aggregateScopeMatches(b, authorization) {
		return identityProof{}, ErrAggregateScopeMismatch
	}
	if a != b {
		return identityProof{}, ErrInspectionIncoherent
	}
	active, hasActive := a.ActiveAttempt()
	proof := identityProof{revision: a.Revision(), absent: !hasActive}
	for _, record := range history {
		if record.LaunchAttemptID() != attemptID {
			continue
		}
		if record.RuntimeInstanceID() != instanceID {
			return identityProof{}, ErrAggregateScopeMismatch
		}
		proof.absent = false
		if hasActive && active == attemptID && record.RuntimeInstanceID() == instanceID &&
			record.ConfigurationVersionID() == versionID && record.Phase() == runtimeidentity.AttemptPhaseClaimed {
			proof.exactClaimed = true
			proof.generation = record.ExecutionGeneration()
		}
		break
	}
	return proof, nil
}

func aggregateScopeMatches(
	view runtimeidentity.RuntimeInstanceView,
	authorization runtimeorchestrationbinding.OrchestrationAuthorizationRequest,
) bool {
	return view.RuntimeInstanceID() == authorization.RuntimeInstanceID() &&
		view.WorkspaceID() == authorization.WorkspaceID() &&
		view.ConfigurationID() == authorization.ConfigurationID()
}

func (c *Continuation) continueFinal(
	binding runtimeorchestrationbinding.StartExecutionBinding,
) (runtimelaunchflow.StartClaimOutcome, error) {
	gate, err := c.boundary.ResolveManagedStartFinal(binding, runtimecommandidempotency.FinalContinue)
	if err != nil {
		return runtimelaunchflow.StartClaimBlocked, err
	}
	if gate == runtimecommandidempotency.GateStopConverged {
		return runtimelaunchflow.StartClaimStopConverged, nil
	}
	if gate != runtimecommandidempotency.GateClear {
		return runtimelaunchflow.StartClaimBlocked, ErrContinuationBlocked
	}
	return runtimelaunchflow.StartClaimContinue, nil
}

func (c *Continuation) bindingFailedFinal(
	binding runtimeorchestrationbinding.StartExecutionBinding,
	cause error,
) (runtimelaunchflow.StartClaimOutcome, error) {
	gate, err := c.boundary.ResolveManagedStartFinal(binding, runtimecommandidempotency.FinalBindingFailed)
	if err != nil {
		return runtimelaunchflow.StartClaimBlocked, err
	}
	if gate == runtimecommandidempotency.GateStopConverged {
		return runtimelaunchflow.StartClaimStopConverged, nil
	}
	return runtimelaunchflow.StartClaimBindingFailed, firstError(cause, ErrBindingFailed)
}

func (c *Continuation) blockedFinal(
	binding runtimeorchestrationbinding.StartExecutionBinding,
	cause error,
) (runtimelaunchflow.StartClaimOutcome, error) {
	// An indeterminate identity observation is deliberately not converted into
	// a definitive BindingFailed disposition. Returning on the original stack
	// expires the command permit, which blocks and wakes any pending Stop.
	_ = binding
	return runtimelaunchflow.StartClaimBlocked, firstError(cause, ErrContinuationBlocked)
}

func firstError(primary, fallback error) error {
	if primary != nil {
		return primary
	}
	return fallback
}
