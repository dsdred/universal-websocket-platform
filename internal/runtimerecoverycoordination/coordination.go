// Package runtimerecoverycoordination models the isolated DP-024 local ordering
// foundation. Concrete owner models are test ports, not DP-014/DP-015 stores.
// Nothing here proves durable absence, process loss, claim truth or authority.
package runtimerecoverycoordination

import (
	"context"
	"errors"
	"slices"
	"sync"

	"github.com/dsdred/universal-websocket-platform/internal/runtimecontainment"
)

var errRejected = errors.New("local recovery coordination rejected")

const maxMembers = 64

type scope struct {
	operational              string
	containment              runtimecontainment.Domain
	workspace, configuration uint64
	instance                 string
}

func newScope(domain string, workspace, configuration uint64, instance string) (scope, error) {
	parsed, err := runtimecontainment.ParseDomain(domain)
	if err != nil || workspace == 0 || configuration == 0 || instance == "" {
		return scope{}, errRejected
	}
	return scope{domain, parsed, workspace, configuration, instance}, nil
}

func (s scope) valid() bool {
	parsed, err := runtimecontainment.ParseDomain(s.operational)
	return err == nil && parsed == s.containment && s.workspace != 0 && s.configuration != 0 && s.instance != ""
}

type fact uint8

const (
	unknown fact = iota
	clean
	managedLive
	recoveryRequired
)

type memberKind uint8

const (
	primitive memberKind = iota + 1
	parent
	phase
)

// member captures complete local membership, including terminal facts/absence.
type member struct {
	kind           memberKind
	key, parentKey string
	revision       uint64
	terminal       bool
}

type aggregateModel struct {
	mu          sync.Mutex
	identity    *modelIdentity
	participant *participant
	scope       scope
	revision    uint64
	fact        fact
}

type commandModel struct {
	mu                sync.Mutex
	identity          *modelIdentity
	participant       *participant
	scope             scope
	revision          uint64
	complete, noClaim bool
	members           []member
}

type participant struct {
	scope               scope
	outer               chan struct{}
	state               sync.Mutex
	epoch               uint64
	open, sticky, fatal bool
	aggregate           *aggregateModel
	commands            *commandModel
}

// Non-zero opaque identity keeps detached observations free of owner pointers.
type modelIdentity struct{ marker byte }

func newParticipant(s scope) (*participant, error) {
	if !s.valid() {
		return nil, errRejected
	}
	p := &participant{scope: s, outer: make(chan struct{}, 1), epoch: 1}
	p.outer <- struct{}{}
	return p, nil
}

// attachModels binds one pair once to an explicitly supplied participant.
// Initial unknown/incomplete model facts cannot open startup admission.
func attachModels(p *participant) (*aggregateModel, *commandModel, error) {
	if p == nil || p.outer == nil {
		return nil, nil, errRejected
	}
	p.state.Lock()
	defer p.state.Unlock()
	if p.aggregate != nil || p.commands != nil {
		return nil, nil, errRejected
	}
	a := &aggregateModel{identity: &modelIdentity{}, participant: p, scope: p.scope, revision: 1}
	c := &commandModel{identity: &modelIdentity{}, participant: p, scope: p.scope, revision: 1}
	p.aggregate, p.commands = a, c
	return a, c, nil
}

func (p *participant) acquire(ctx context.Context) error {
	if p == nil || p.outer == nil || ctx == nil || ctx.Err() != nil {
		return errRejected
	}
	select {
	case <-ctx.Done():
		return errRejected
	case <-p.outer:
		if ctx.Err() != nil {
			p.release()
			return errRejected
		}
		return nil
	}
}

func (p *participant) release() { p.outer <- struct{}{} }

// Enter only after outer acquisition, including session calls. Every internal
// transition uses state -> aggregate -> commands; external work occurs outside.
func (p *participant) lockOwners(a *aggregateModel, c *commandModel) bool {
	p.state.Lock()
	if a == nil || c == nil || a != p.aggregate || c != p.commands ||
		a.participant != p || c.participant != p || a.scope != p.scope || c.scope != p.scope {
		p.state.Unlock()
		return false
	}
	a.mu.Lock()
	c.mu.Lock()
	return true
}

func (p *participant) unlockOwners(a *aggregateModel, c *commandModel) {
	c.mu.Unlock()
	a.mu.Unlock()
	p.state.Unlock()
}

type snapshot struct {
	scope                                     scope
	aggregate                                 *modelIdentity
	commands                                  *modelIdentity
	epoch, aggregateRevision, commandRevision uint64
	fact                                      fact
	complete, noClaim                         bool
	members                                   []member
}

func (p *participant) capture(a *aggregateModel, c *commandModel) snapshot {
	return snapshot{p.scope, a.identity, c.identity, p.epoch, a.revision, c.revision, a.fact,
		c.complete, c.noClaim, slices.Clone(c.members)}
}

func validFacts(s snapshot) bool {
	if !s.scope.valid() || s.aggregateRevision == 0 || s.commandRevision == 0 || s.epoch == 0 ||
		!s.complete || len(s.members) > maxMembers || s.fact > recoveryRequired {
		return false
	}
	seen := make(map[string]memberKind, len(s.members))
	for _, m := range s.members {
		if m.key == "" || m.revision == 0 || m.kind < primitive || m.kind > phase {
			return false
		}
		if _, duplicate := seen[m.key]; duplicate {
			return false
		}
		if (m.kind == phase) != (m.parentKey != "") {
			return false
		}
		seen[m.key] = m.kind
	}
	for _, m := range s.members {
		if m.kind == phase && seen[m.parentKey] != parent {
			return false
		}
		if s.fact == clean && !m.terminal {
			return false
		}
	}
	return true
}

func (p *participant) matches(s snapshot, a *aggregateModel, c *commandModel) bool {
	return validFacts(s) && s.scope == p.scope && s.aggregate == a.identity && s.commands == c.identity &&
		s.epoch == p.epoch && s.aggregateRevision == a.revision && s.commandRevision == c.revision &&
		s.fact == a.fact && s.complete == c.complete && s.noClaim == c.noClaim && slices.Equal(s.members, c.members)
}

func (p *participant) read(ctx context.Context, a *aggregateModel, c *commandModel) (snapshot, error) {
	if err := p.acquire(ctx); err != nil {
		return snapshot{}, err
	}
	defer p.release()
	if !p.lockOwners(a, c) {
		return snapshot{}, errRejected
	}
	defer p.unlockOwners(a, c)
	return p.capture(a, c), nil
}

func (p *participant) closeFence() { p.open, p.sticky = false, true }

// Clean opening is restricted to startup; sticky recovery/unknown closure has
// no release operation. These local facts do not qualify durable no-claim.
func (p *participant) observe(ctx context.Context, a *aggregateModel, c *commandModel, s snapshot) error {
	if err := p.acquire(ctx); err != nil {
		return err
	}
	defer p.release()
	if !p.lockOwners(a, c) {
		return errRejected
	}
	defer p.unlockOwners(a, c)
	if s.fact == unknown || s.fact == recoveryRequired || !s.complete || !s.noClaim ||
		a.fact == unknown || a.fact == recoveryRequired || !c.complete || !c.noClaim ||
		!validFacts(s) || !validFacts(p.capture(a, c)) {
		p.closeFence()
	}
	if !p.matches(s, a, c) || p.fatal {
		return errRejected
	}
	if s.fact == clean && s.noClaim && !p.sticky {
		p.open = true
	}
	return nil
}

// modelUpdate is a closed test-owner publication, never an arbitrary callback.
type modelUpdate struct {
	fact              fact
	complete, noClaim bool
	members           []member
}

func (p *participant) update(ctx context.Context, a *aggregateModel, c *commandModel, before snapshot, next modelUpdate) error {
	if err := p.acquire(ctx); err != nil {
		return err
	}
	defer p.release()
	if !p.lockOwners(a, c) {
		return errRejected
	}
	defer p.unlockOwners(a, c)
	current := p.capture(a, c)
	if current.fact == unknown || current.fact == recoveryRequired || !current.complete || !current.noClaim || !validFacts(current) {
		p.closeFence()
		return errRejected
	}
	if !p.open || p.fatal || !p.matches(before, a, c) || a.revision == ^uint64(0) || c.revision == ^uint64(0) {
		return errRejected
	}
	check := current
	// Incomplete but structurally valid model facts may be published only with
	// immediate closed admission; malformed identities/links remain rejected.
	check.fact, check.complete, check.noClaim, check.members = next.fact, true, next.noClaim, next.members
	if !validFacts(check) {
		return errRejected
	}
	a.revision++
	c.revision++
	a.fact = next.fact
	c.complete, c.noClaim, c.members = next.complete, next.noClaim, slices.Clone(next.members)
	if a.fact == unknown || a.fact == recoveryRequired || !c.complete || !c.noClaim {
		p.closeFence()
	}
	return nil
}

// Invalidation takes only state, never waits for an outer-retaining session's
// simulated I/O, and proves neither process loss nor an absent publication.
func (p *participant) invalidate(fatal bool) error {
	if p == nil || p.outer == nil {
		return errRejected
	}
	p.state.Lock()
	defer p.state.Unlock()
	p.closeFence()
	p.fatal = p.fatal || fatal || p.epoch == ^uint64(0)
	if p.epoch != ^uint64(0) {
		p.epoch++
	}
	return nil
}

func (p *participant) admissionOpen() bool {
	if p == nil || p.outer == nil {
		return false
	}
	p.state.Lock()
	defer p.state.Unlock()
	return p.open && !p.fatal && !p.sticky
}

// reservation is only exact captured local data, never a claim or permit.
type reservation struct {
	participant *participant
	aggregate   *aggregateModel
	commands    *commandModel
	facts       snapshot
}

// session retains only outer between reserve and close. Its mutex protects
// disposal; no owner mutex is retained, and no external callback is exposed.
type session struct {
	mu              sync.Mutex
	participant     *participant
	reservation     *reservation
	closed, checked bool
}

func (p *participant) reserve(ctx context.Context, a *aggregateModel, c *commandModel, s snapshot) (*session, error) {
	if err := p.acquire(ctx); err != nil {
		return nil, err
	}
	if !p.lockOwners(a, c) {
		p.release()
		return nil, errRejected
	}
	// Even a stale recovery reservation leaves admission closed.
	p.closeFence()
	if p.fatal || !p.matches(s, a, c) || s.fact == unknown || s.fact == managedLive || (s.fact == clean && s.noClaim) {
		p.unlockOwners(a, c)
		p.release()
		return nil, errRejected
	}
	r := &reservation{participant: p, aggregate: a, commands: c, facts: p.capture(a, c)}
	p.unlockOwners(a, c)
	return &session{participant: p, reservation: r}, nil
}

// One local validation disposition only: success delivers no live authority.
func (s *session) revalidate(r *reservation) error {
	if s == nil {
		return errRejected
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.checked || r == nil || r != s.reservation || s.participant == nil || r.participant != s.participant {
		return errRejected
	}
	s.checked = true
	p := s.participant
	a, c := r.aggregate, r.commands
	if !p.lockOwners(a, c) {
		return errRejected
	}
	defer p.unlockOwners(a, c)
	if p.fatal || !p.matches(r.facts, a, c) {
		return errRejected
	}
	return nil
}

func (s *session) close() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.participant == nil {
		return
	}
	s.closed = true
	s.participant.release()
}
