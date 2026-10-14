package runtimecontainment

import "github.com/dsdred/universal-websocket-platform/internal/runtimecoordination"

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
