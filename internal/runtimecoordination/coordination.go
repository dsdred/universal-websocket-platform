// Package runtimecoordination provides the closed local DP-024 participation
// seam. It owns no lifecycle, command, containment or recovery authority.
package runtimecoordination

import (
	"errors"
	"sync"
)

// Coordination errors distinguish invalid zero scope or handles (ErrInvalidScope),
// foreign exact scope (ErrForeignScope), duplicate binding (ErrAlreadyAttached),
// and the permanently closed transition boundary (ErrClosed).
var (
	ErrInvalidScope    = errors.New("invalid coordination scope")
	ErrForeignScope    = errors.New("foreign coordination scope")
	ErrAlreadyAttached = errors.New("aggregate participant already attached")
	ErrClosed          = errors.New("coordination transition closed")
)

// Scope is the validated immutable domain/Workspace/Configuration/Instance tuple.
// Its zero value is invalid. Strings are opaque identities, not locators.
type Scope struct {
	domain                   string
	workspace, configuration uint64
	instance                 string
}

// NewScope validates the exact tuple supplied by composition.
func NewScope(domain string, workspace, configuration uint64, instance string) (Scope, error) {
	if domain == "" || workspace == 0 || configuration == 0 || instance == "" {
		return Scope{}, ErrInvalidScope
	}
	return Scope{domain, workspace, configuration, instance}, nil
}

// Valid reports whether every required identity is nonzero.
func (s Scope) Valid() bool {
	return s.domain != "" && s.workspace != 0 && s.configuration != 0 && s.instance != ""
}

// Domain returns the immutable domain identity.
func (s Scope) Domain() string { return s.domain }

// WorkspaceID returns the immutable Workspace identity.
func (s Scope) WorkspaceID() uint64 { return s.workspace }

// ConfigurationID returns the immutable Configuration identity.
func (s Scope) ConfigurationID() uint64 { return s.configuration }

// InstanceID returns the immutable Instance identity.
func (s Scope) InstanceID() string { return s.instance }

// Participant is one exact Instance's local serialization boundary. Construction
// is permanently closed in this slice: no command/authority/no-claim qualification
// exists. Copies preserve the same opaque state and attachment identity.
type Participant struct{ data *participantState }

type participantState struct {
	outer         sync.Mutex
	state         sync.Mutex
	scope         Scope
	attached      bool
	command       bool
	commandMarker *commandMarker
	roots         *rootBinding
	source        *authorityState
	epoch         uint64
	terminal      bool
}

// NewParticipant constructs a closed participant for a validated scope.
func NewParticipant(scope Scope) (*Participant, error) {
	if !scope.Valid() {
		return nil, ErrInvalidScope
	}
	return &Participant{data: &participantState{scope: scope}}, nil
}

// AggregateAttachment is the non-rebinding identity of the one conforming
// aggregate owner. Its zero value is invalid; copying it does not attach again.
type AggregateAttachment struct{ participant *participantState }

// AttachAggregate binds exactly one aggregate owner before any allocation or
// observation of absence. It cannot assert command or authority coverage.
func (p *Participant) AttachAggregate(scope Scope) (AggregateAttachment, error) {
	if p == nil || p.data == nil || !scope.Valid() || !p.data.scope.Valid() {
		return AggregateAttachment{}, ErrInvalidScope
	}
	if scope != p.data.scope {
		return AggregateAttachment{}, ErrForeignScope
	}
	if p.data.source != nil {
		p.data.source.gate.RLock()
		defer p.data.source.gate.RUnlock()
	}
	p.data.state.Lock()
	defer p.data.state.Unlock()
	if p.data.terminal {
		return AggregateAttachment{}, ErrClosed
	}
	if p.data.attached {
		return AggregateAttachment{}, ErrAlreadyAttached
	}
	p.data.attached = true
	return AggregateAttachment{p.data}, nil
}

// ObservationRegion is a short owner observation region. Close it before any
// external work or wait. It exposes no raw lock or callback execution surface.
type ObservationRegion struct{ release *regionRelease }
type regionRelease struct {
	participant                   *participantState
	once                          sync.Once
	roots                         *rootBinding
	clientEpoch, participantEpoch uint64
	stateHeld                     bool
	authority                     *authorityState
}

// Observe enters Instance outer before owner-local locks. Foreign or zero scope
// fails before the owner may disclose facts. Observation grants no authority.
func (a AggregateAttachment) Observe(scope Scope) (*ObservationRegion, error) {
	p := a.participant
	if p == nil || !scope.Valid() || !p.scope.Valid() {
		return nil, ErrInvalidScope
	}
	if scope != p.scope {
		return nil, ErrForeignScope
	}
	p.outer.Lock()
	p.state.Lock()
	roots := p.roots
	authority := p.source
	if authority == nil && roots != nil {
		authority = roots.authority
	}
	p.state.Unlock()
	if authority != nil {
		authority.gate.RLock()
	}
	if roots != nil {
		roots.client.gate.RLock()
	}
	p.state.Lock()
	// Registration can publish the immutable binding between the initial read
	// and this acquisition. Retry before exposing an incompletely ordered region.
	if roots != p.roots {
		p.state.Unlock()
		if roots != nil {
			roots.client.gate.RUnlock()
		}
		if authority != nil {
			authority.gate.RUnlock()
		}
		p.outer.Unlock()
		return a.Observe(scope)
	}
	if p.terminal {
		p.state.Unlock()
		if roots != nil {
			roots.client.gate.RUnlock()
		}
		if authority != nil {
			authority.gate.RUnlock()
		}
		p.outer.Unlock()
		return nil, ErrClosed
	}
	release := &regionRelease{participant: p, roots: roots, authority: authority, stateHeld: true}
	if roots != nil {
		release.clientEpoch = roots.client.epoch
		release.participantEpoch = p.epoch
	}
	return &ObservationRegion{release: release}, nil
}

// Close releases the short observation region exactly once.
func (r *ObservationRegion) Close() {
	if r != nil && r.release != nil {
		r.release.once.Do(func() {
			if r.release.stateHeld {
				r.release.participant.state.Unlock()
			}
			if r.release.roots != nil {
				r.release.roots.client.gate.RUnlock()
			}
			if r.release.authority != nil {
				r.release.authority.gate.RUnlock()
			}
			r.release.participant.outer.Unlock()
		})
	}
}

// Transition enters the same ordered local region, then rejects permanently.
// Missing command attachment, current authority and qualified no-claim cannot
// be replaced by a caller assertion or by an observation.
func (a AggregateAttachment) Transition(scope Scope) error {
	r, err := a.Observe(scope)
	if err != nil {
		return err
	}
	defer r.Close()
	return ErrClosed
}
