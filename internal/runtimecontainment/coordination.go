package runtimecontainment

import (
	"context"

	"github.com/dsdred/universal-websocket-platform/internal/runtimecoordination"
)

func (k *keeper) publishFatal() {
	region := k.root.BeginFatalPublication()
	k.fenced.Store(true)
	region.Invalidate()
	region.Close()
}

func (k *keeper) joinFatal() { k.root.JoinFatalPublication() }

// CoordinationRoot supplies only this keeper's immutable same-domain ordering
// root after external validation. The root grants no current-authority proof.
// Legacy/no-source authorities cannot fabricate or adopt a source association.
func (a *ActiveAuthority) CoordinationRoot(domain Domain) (*runtimecoordination.AuthorityFence, error) {
	if a == nil || a.keeper == nil || zeroIdentity(domain.value) || a.keeper.domain != domain {
		return nil, runtimecoordination.ErrForeignScope
	}
	if !a.IsAuthoritative() || a.keeper.root == nil {
		return nil, runtimecoordination.ErrClosed
	}
	return a.keeper.root, nil
}

// InspectCoordination inspects actual current generation and containment authority
// outside all coordination gates, then returns a detached local source binding.
// The result grants no continuing authority or recovery qualification.
func (a *ActiveAuthority) InspectCoordination(ctx context.Context, domain Domain) (runtimecoordination.CurrentObservation, error) {
	var zero runtimecoordination.CurrentObservation
	if a == nil || a.keeper == nil || ctx == nil || zeroIdentity(domain.value) {
		return zero, runtimecoordination.ErrInvalidScope
	}
	k := a.keeper
	if domain != k.domain {
		return zero, runtimecoordination.ErrForeignScope
	}
	if k.root == nil || k.issuer == nil {
		return zero, runtimecoordination.ErrClosed
	}
	if zeroIdentity(a.generation.value) || k.committedGeneration != a.generation {
		k.publishFatal()
		return zero, runtimecoordination.ErrClosed
	}
	named := a.CurrentGeneration()
	fact := a.ReadGeneration(ctx, domain, named)
	// Always finish actual authority inspection: detected faults outrank cancellation.
	if !a.IsAuthoritative() {
		return zero, runtimecoordination.ErrClosed
	}
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	if fact.Kind() != EvidenceGenerationLive || named != a.CurrentGeneration() {
		return zero, runtimecoordination.ErrClosed
	}
	observation, err := k.issuer.Observe()
	if err == nil {
		err = k.root.RevalidateCurrentObservation(observation)
	}
	if err != nil {
		k.publishFatal()
		return zero, runtimecoordination.ErrClosed
	}
	return observation, nil
}
