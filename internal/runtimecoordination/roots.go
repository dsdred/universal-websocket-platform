package runtimecoordination

import (
	"math"
	"sync"
)

// AuthorityFence is an inert local ordering root. It conveys no containment
// authority and exposes neither opening nor external callback operations.
type AuthorityFence struct{ data *authorityState }
type authorityState struct {
	issuer       *issuerMarker
	generation   string
	gate         sync.RWMutex
	registry     sync.Mutex
	domain       string
	terminal     bool
	participants map[*participantState]struct{}
	clients      map[*clientState]struct{}
}

// NewAuthorityFence constructs an inert root explicitly supplied by composition.
func NewAuthorityFence() *AuthorityFence { return &AuthorityFence{data: &authorityState{}} }

// ClientEpoch is a storage-owned opaque root. Copies preserve the same gates
// and epoch. Replacement never takes Instance outer.
type ClientEpoch struct{ data *clientState }
type clientState struct {
	gate         sync.RWMutex
	authority    *authorityState
	participants map[*participantState]struct{}
	epoch        uint64
	fenced       bool
}

// NewClientEpoch creates one private client root for one storage construction.
func NewClientEpoch(authority *AuthorityFence) (*ClientEpoch, error) {
	if authority == nil || authority.data == nil {
		return nil, ErrInvalidScope
	}
	a := authority.data
	a.gate.RLock()
	defer a.gate.RUnlock()
	if a.terminal {
		return nil, ErrClosed
	}
	c := &clientState{authority: a, participants: make(map[*participantState]struct{})}
	if a.domain != "" {
		a.registry.Lock()
		a.clients[c] = struct{}{}
		a.registry.Unlock()
	}
	return &ClientEpoch{data: c}, nil
}

type rootBinding struct {
	authority *authorityState
	client    *clientState
}

// CommandAttachment is the exact non-rebinding command owner identity.
// Copies share the same participant and do not register another owner.
type CommandAttachment struct {
	participant *participantState
	marker      *commandMarker
}
type commandMarker struct{ value byte }

// RegistrationRegion holds only bounded registration gates. Owner registration
// completes before Close; no callback or ledger operation belongs in this region.
type RegistrationRegion struct{ release *rootRelease }
type rootRelease struct {
	client      *clientState
	participant *participantState
	once        sync.Once
	committed   bool
	disposition sync.Mutex
	closed      bool
	marker      *commandMarker
}

// RegisterCommand stages a previously unattached command owner under authority,
// client and participant gates; Commit finishes binding. The caller finishes short catalog registration
// while retaining the returned registration region.
func (handle *ClientEpoch) RegisterCommand(p *Participant, scope Scope) (CommandAttachment, *RegistrationRegion, error) {
	if handle == nil || handle.data == nil || p == nil || p.data == nil || !scope.Valid() {
		return CommandAttachment{}, nil, ErrInvalidScope
	}
	c := handle.data
	if p.data.scope != scope {
		return CommandAttachment{}, nil, ErrForeignScope
	}
	c.authority.gate.RLock()
	if c.authority.terminal || (c.authority.domain != "" && p.data.source != c.authority) || (p.data.source != nil && p.data.source != c.authority) {
		terminal := c.authority.terminal
		c.authority.gate.RUnlock()
		if terminal {
			return CommandAttachment{}, nil, ErrClosed
		}
		return CommandAttachment{}, nil, ErrForeignScope
	}
	c.gate.Lock()
	p.data.state.Lock()
	if p.data.command || p.data.roots != nil || c.fenced {
		fenced := c.fenced
		p.data.state.Unlock()
		c.gate.Unlock()
		c.authority.gate.RUnlock()
		if fenced {
			return CommandAttachment{}, nil, ErrClosed
		}
		return CommandAttachment{}, nil, ErrAlreadyAttached
	}
	marker := &commandMarker{}
	return CommandAttachment{p.data, marker}, &RegistrationRegion{release: &rootRelease{client: c, participant: p.data, marker: marker}}, nil
}

// Close releases registration gates exactly once.
func (r *RegistrationRegion) Close() {
	if r != nil && r.release != nil {
		r.release.disposition.Lock()
		defer r.release.disposition.Unlock()
		r.release.once.Do(func() {
			r.release.closed = true
			r.release.participant.state.Unlock()
			r.release.client.gate.Unlock()
			r.release.client.authority.gate.RUnlock()
		})
	}
}

// Observe enters one complete ordered short owner region.
func (a CommandAttachment) Observe(scope Scope) (*ObservationRegion, error) {
	if a.participant == nil || a.marker == nil {
		return nil, ErrInvalidScope
	}
	a.participant.state.Lock()
	committed := a.participant.commandMarker == a.marker
	a.participant.state.Unlock()
	if !committed {
		return nil, ErrInvalidScope
	}
	return (AggregateAttachment{participant: a.participant}).Observe(scope)
}

// Epochs returns detached epochs captured on entry, safe even after Close.
func (r *ObservationRegion) Epochs() (uint64, uint64) {
	if r == nil || r.release == nil || r.release.roots == nil {
		return 0, 0
	}
	return r.release.clientEpoch, r.release.participantEpoch
}

// ReplacementRegion excludes registration/short reads while invalidating all
// attached participants. Owner client generation publication occurs before Close.
type ReplacementRegion struct{ release *rootRelease }

// BeginReplacement invalidates every registered participant without taking any
// outer or owner lock. Exhaustion permanently fences and publishes no new epoch.
func (handle *ClientEpoch) BeginReplacement() (*ReplacementRegion, error) {
	if handle == nil || handle.data == nil {
		return nil, ErrInvalidScope
	}
	c := handle.data
	c.gate.Lock()
	for p := range c.participants {
		p.state.Lock()
		exhausted := p.epoch == math.MaxUint64
		p.state.Unlock()
		if exhausted {
			c.fenced = true
		}
	}
	if c.fenced || c.epoch == math.MaxUint64 {
		c.fenced = true
		for p := range c.participants {
			p.state.Lock()
			p.epoch = math.MaxUint64
			p.state.Unlock()
		}
		c.gate.Unlock()
		return nil, ErrClosed
	}
	c.epoch++
	for p := range c.participants {
		p.state.Lock()
		p.epoch = c.epoch
		p.state.Unlock()
	}
	return &ReplacementRegion{release: &rootRelease{client: c}}, nil
}

// Close releases replacement exclusivity after generation publication.
func (r *ReplacementRegion) Close() {
	if r != nil && r.release != nil {
		r.release.disposition.Lock()
		defer r.release.disposition.Unlock()
		r.release.once.Do(func() { r.release.closed = true; r.release.client.gate.Unlock() })
	}
}

// Commit completes the staged exact command registration after the conforming
// owner has checked and registered its short catalog. Close aborts an uncommitted
// registration; no owner facts or attachment are changed by an aborted attempt.
func (r *RegistrationRegion) Commit() {
	if r == nil || r.release == nil {
		return
	}
	release := r.release
	release.disposition.Lock()
	defer release.disposition.Unlock()
	if release.closed || release.committed {
		return
	}
	release.participant.commandMarker = release.marker
	release.participant.command = true
	release.participant.roots = &rootBinding{release.client.authority, release.client}
	release.participant.epoch = release.client.epoch
	release.client.participants[release.participant] = struct{}{}
	release.committed = true
}

// Fence permanently exhausts this root and attached epochs while replacement
// exclusivity is retained. It cannot open admission or publish a generation.
func (r *ReplacementRegion) Fence() {
	if r == nil || r.release == nil {
		return
	}
	r.release.disposition.Lock()
	defer r.release.disposition.Unlock()
	if r.release.closed {
		return
	}
	c := r.release.client
	c.fenced = true
	for p := range c.participants {
		p.state.Lock()
		p.epoch = math.MaxUint64
		p.state.Unlock()
	}
}
