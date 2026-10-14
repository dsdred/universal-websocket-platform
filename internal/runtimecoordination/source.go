package runtimecoordination

import (
	"math"
	"sync"
)

// NewSourceAuthorityFence creates an immutable domain-associated ordering root.
// It proves no containment authority; only the containment keeper supplies its
// own exact root to composition. Inert legacy roots remain unqualified.
func NewSourceAuthorityFence(domain string) (*AuthorityFence, error) {
	if domain == "" {
		return nil, ErrInvalidScope
	}
	return &AuthorityFence{data: &authorityState{domain: domain, participants: make(map[*participantState]struct{}), clients: make(map[*clientState]struct{})}}, nil
}

// NewSourceParticipant registers before any owner attachment. Registration is
// permanent, including aborted or partial construction, and cannot rebind.
func NewSourceParticipant(root *AuthorityFence, scope Scope) (*Participant, error) {
	if root == nil || root.data == nil || !scope.Valid() {
		return nil, ErrInvalidScope
	}
	a := root.data
	if a.domain == "" || a.domain != scope.Domain() {
		return nil, ErrForeignScope
	}
	a.gate.RLock()
	defer a.gate.RUnlock()
	if a.terminal {
		return nil, ErrClosed
	}
	p, _ := NewParticipant(scope)
	p.data.source = a
	a.registry.Lock()
	a.participants[p.data] = struct{}{}
	a.registry.Unlock()
	return p, nil
}

// FatalPublication is an opaque bounded source publication region. Copies share
// one disposition. No owner/outer lock, callback or external work belongs here.
type FatalPublication struct{ release *fatalRelease }
type fatalRelease struct {
	authority           *authorityState
	mu                  sync.Mutex
	closed, invalidated bool
}

// BeginFatalPublication excludes source registration and short observations.
// A zero root returns a safe zero region for unqualified legacy fixtures.
func (root *AuthorityFence) BeginFatalPublication() *FatalPublication {
	if root == nil || root.data == nil {
		return &FatalPublication{}
	}
	root.data.gate.Lock()
	return &FatalPublication{release: &fatalRelease{authority: root.data}}
}

func (r *fatalRelease) invalidate() {
	if r.invalidated {
		return
	}
	a := r.authority
	a.terminal = true
	for c := range a.clients {
		c.gate.Lock()
		c.fenced = true
		c.epoch = math.MaxUint64
		for p := range c.participants {
			p.state.Lock()
			p.terminal, p.epoch = true, math.MaxUint64
			p.state.Unlock()
		}
		c.gate.Unlock()
	}
	for p := range a.participants {
		p.state.Lock()
		p.terminal, p.epoch = true, math.MaxUint64
		p.state.Unlock()
	}
	r.invalidated = true
}

// Invalidate permanently terminals every associated client and participant.
func (region *FatalPublication) Invalidate() {
	if region == nil || region.release == nil {
		return
	}
	r := region.release
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.closed {
		r.invalidate()
	}
}

// Close completes invalidation and releases publication exactly once.
func (region *FatalPublication) Close() {
	if region == nil || region.release == nil {
		return
	}
	r := region.release
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	r.invalidate()
	r.closed = true
	r.authority.gate.Unlock()
}

// JoinFatalPublication waits only for bounded same-root publication completion.
// Call outside every owner/outer gate, after external source inspection.
func (root *AuthorityFence) JoinFatalPublication() {
	if root == nil || root.data == nil {
		return
	}
	root.data.gate.RLock()
	root.data.gate.RUnlock()
}
