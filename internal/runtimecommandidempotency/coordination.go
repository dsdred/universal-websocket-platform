package runtimecommandidempotency

import (
	"reflect"
	"sync"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecoordination"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

type commandCoordination struct {
	self        *MemoryStorage
	owner       *coordinationOwner
	presence    map[instanceScope]uint64
	client      *runtimecoordination.ClientEpoch
	attachments map[instanceScope]runtimecoordination.CommandAttachment
}

// NewCoordinatedMemoryStorage constructs permanently closed process-local
// command storage against an explicitly supplied inert authority ordering root.
// It grants no containment, recovery or ordinary execution authority.
func NewCoordinatedMemoryStorage(authority *runtimecoordination.AuthorityFence) (*MemoryStorage, error) {
	client, err := runtimecoordination.NewClientEpoch(authority)
	if err != nil {
		return nil, err
	}
	s := NewMemoryStorage()
	s.coord = &commandCoordination{self: s, owner: &coordinationOwner{}, presence: make(map[instanceScope]uint64), client: client, attachments: make(map[instanceScope]runtimecoordination.CommandAttachment)}
	return s, nil
}

func coordinationScope(scope instanceScope) runtimecoordination.Scope {
	s, _ := runtimecoordination.NewScope(scope.domain, scope.workspaceID, scope.configurationID, string(scope.runtimeInstanceID))
	return s
}

// AttachInstance explicitly binds the exact shared participant before any
// ledger allocation or observation of absence. Duplicate/rebinding and foreign
// scope reject without changing owner facts.
func (s *MemoryStorage) AttachInstance(scope runtimecoordination.Scope, p *runtimecoordination.Participant) error {
	if s == nil || s.coord == nil || s.coord.self != s || !scope.Valid() {
		return runtimecoordination.ErrInvalidScope
	}
	key := instanceScope{domain: scope.Domain(), workspaceID: scope.WorkspaceID(), configurationID: scope.ConfigurationID(), runtimeInstanceID: runtimeconfigload.RuntimeInstanceID(scope.InstanceID())}
	// Catalog precheck is released before root acquisition. Registration then
	// retains participant state while completing the short owner catalog.
	s.mu.Lock()
	used := s.ledgers[key] != nil
	_, duplicate := s.coord.attachments[key]
	s.mu.Unlock()
	if used || duplicate {
		return runtimecoordination.ErrAlreadyAttached
	}
	a, r, err := s.coord.client.RegisterCommand(p, scope)
	if err != nil {
		return err
	}
	defer r.Close()
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.coord.attachments[key]; exists || s.ledgers[key] != nil {
		return runtimecoordination.ErrAlreadyAttached
	}
	s.coord.attachments[key] = a
	r.Commit()
	return nil
}

type shortRegion struct {
	storage                       *MemoryStorage
	owner                         *runtimecoordination.ObservationRegion
	closed                        bool
	invalid                       bool
	clientEpoch, participantEpoch uint64
	once                          sync.Once
}

func (s *MemoryStorage) lockRegion(scope instanceScope) *shortRegion {
	r := &shortRegion{storage: s}
	if s.coord != nil {
		if s.coord.self != s {
			r.closed = true
			r.invalid = true
			s.clientMu.RLock()
			return r
		}
		s.mu.Lock()
		a, ok := s.coord.attachments[scope]

		s.mu.Unlock()
		if ok && s.coord.self == s {
			r.owner, _ = a.Observe(coordinationScope(scope))
			r.clientEpoch, r.participantEpoch = r.owner.Epochs()
		}
		r.closed = true
		r.invalid = r.owner == nil
	}
	s.clientMu.RLock()
	return r
}

func (r *shortRegion) Close() {
	if r != nil {
		r.once.Do(func() {
			r.storage.clientMu.RUnlock()
			if r.owner != nil {
				r.owner.Close()
			}
		})
	}
}

// ledgerState is a private exact before/after accounting copy. Permit/channel
// identities are retained only inside the ledger lock and never in observations.
type ledgerState struct {
	records           map[commandIdentity]commandRecord
	parents           map[commandIdentity]parentRecord
	phases            map[phaseIdentity]phaseRecord
	live, liveParents map[commandIdentity]liveFact
	livePhases        map[phaseIdentity]liveFact
	slots             map[commandIdentity]slotFact
	rendezvous        map[commandIdentity]rendezvousFact
	managed           map[runtimeorchestrationbinding.StartRendezvous]managedFact
}
type liveFact struct {
	identity *permitState
	value    permitState
}
type slotFact struct {
	command              commandIdentity
	phase                phaseIdentity
	hasCommand, hasPhase bool
}
type rendezvousFact struct {
	value    startRendezvous
	state    permitState
	hasState bool
}
type managedFact struct {
	value               managedStartRendezvous
	state               permitState
	bridge              rendezvousFact
	hasState, hasBridge bool
}

func copyRendezvous(r *startRendezvous) rendezvousFact {
	f := rendezvousFact{value: *r}
	if r.stopState != nil {
		f.state = *r.stopState
		f.hasState = true
	}
	return f
}
func (l *commandLedger) facts() ledgerState {
	f := ledgerState{records: make(map[commandIdentity]commandRecord), parents: make(map[commandIdentity]parentRecord), phases: make(map[phaseIdentity]phaseRecord), live: make(map[commandIdentity]liveFact), liveParents: make(map[commandIdentity]liveFact), livePhases: make(map[phaseIdentity]liveFact), slots: make(map[commandIdentity]slotFact), rendezvous: make(map[commandIdentity]rendezvousFact), managed: make(map[runtimeorchestrationbinding.StartRendezvous]managedFact)}
	for id, r := range l.records {
		if r != nil {
			f.records[id] = *r
		}
	}
	for id, r := range l.parents {
		if r != nil {
			f.parents[id] = *r
		}
	}
	for id, r := range l.phases {
		if r != nil {
			f.phases[id] = *r
		}
	}
	for id, r := range l.live {
		if r != nil {
			f.live[id] = liveFact{r, *r}
		}
	}
	for id, r := range l.liveParents {
		if r != nil {
			f.liveParents[id] = liveFact{r, *r}
		}
	}
	for id, r := range l.livePhases {
		if r != nil {
			f.livePhases[id] = liveFact{r, *r}
		}
	}
	for id, o := range l.stopForStart {
		v := slotFact{}
		if o.command != nil {
			v.command = *o.command
			v.hasCommand = true
		}
		if o.phase != nil {
			v.phase = *o.phase
			v.hasPhase = true
		}
		f.slots[id] = v
	}
	for id, r := range l.rendezvous {
		if r != nil {
			f.rendezvous[id] = copyRendezvous(r)
		}
	}
	for id, r := range l.managedStart {
		if r != nil {
			v := managedFact{value: *r}
			if r.state != nil {
				v.state = *r.state
				v.hasState = true
			}
			if r.bridge != nil {
				v.bridge = copyRendezvous(r.bridge)
				v.hasBridge = true
			}
			f.managed[id] = v
		}
	}
	return f
}
func (l *commandLedger) lock() {
	l.mu.Lock()
	l.before = l.facts()
}
func (l *commandLedger) unlock() {
	if !sameLedgerFacts(l.before, l.facts()) && !l.fenced {
		l.membership++
		if l.membership == ^uint64(0) {
			l.fenced = true
		}
	}
	l.before = ledgerState{}
	l.mu.Unlock()
}

type coordinationOwner struct{ marker byte }

// CoordinationSnapshot is a complete detached, authority-free observation of
// one attached Instance, including absence, all membership/links and local epochs.
// It contains no participant, owner, lock, channel or permit links.
type CoordinationSnapshot struct {
	clientFenced, ledgerFenced                                      bool
	scope                                                           runtimecoordination.Scope
	owner                                                           *coordinationOwner
	present, complete                                               bool
	presence, membership, generation, clientEpoch, participantEpoch uint64
	facts                                                           ledgerState
}

// Scope returns the exact immutable Instance scope.
func (s CoordinationSnapshot) Scope() runtimecoordination.Scope { return s.scope }

// Present reports whether the complete command ledger exists.
func (s CoordinationSnapshot) Present() bool { return s.present }

// PresenceRevision returns the ledger allocation/absence concurrency stamp.
func (s CoordinationSnapshot) PresenceRevision() uint64 { return s.presence }

// MembershipRevision returns the complete owner membership concurrency stamp.
func (s CoordinationSnapshot) MembershipRevision() uint64 { return s.membership }

// ClientGeneration returns the observed storage client generation.
func (s CoordinationSnapshot) ClientGeneration() uint64 { return s.generation }

// Epochs returns detached client and participant invalidation stamps.
func (s CoordinationSnapshot) Epochs() (uint64, uint64) { return s.clientEpoch, s.participantEpoch }

func detachedFacts(f ledgerState) ledgerState {
	for id, v := range f.live {
		v.identity = nil
		f.live[id] = v
	}
	for id, v := range f.liveParents {
		v.identity = nil
		f.liveParents[id] = v
	}
	for id, v := range f.livePhases {
		v.identity = nil
		f.livePhases[id] = v
	}
	for id, v := range f.rendezvous {
		v.value.notify = nil
		v.value.stopState = nil
		f.rendezvous[id] = v
	}
	for id, v := range f.managed {
		v.value.state = nil
		v.value.bridge = nil
		v.bridge.value.notify = nil
		v.bridge.value.stopState = nil
		f.managed[id] = v
	}
	return f
}

// CaptureCoordinationSnapshot copies complete coherent owner facts under one
// ordered short region. Unattached, copied or contradictory owners disclose no
// partial result. Absence remains an explicit complete observation, no authority.
func (s *MemoryStorage) CaptureCoordinationSnapshot(scope runtimecoordination.Scope) (CoordinationSnapshot, error) {
	if s == nil || s.coord == nil || s.coord.self != s || !scope.Valid() {
		return CoordinationSnapshot{}, runtimecoordination.ErrInvalidScope
	}
	key := instanceScope{domain: scope.Domain(), workspaceID: scope.WorkspaceID(), configurationID: scope.ConfigurationID(), runtimeInstanceID: runtimeconfigload.RuntimeInstanceID(scope.InstanceID())}
	r := s.lockRegion(key)
	defer r.Close()
	if r.invalid {
		return CoordinationSnapshot{}, runtimecoordination.ErrForeignScope
	}
	s.mu.Lock()
	l := s.ledgers[key]
	presence := s.coord.presence[key]
	s.mu.Unlock()
	snapshot := CoordinationSnapshot{scope: scope, owner: s.coord.owner, complete: true, present: l != nil, presence: presence, generation: s.generation, clientFenced: s.clientFenced, clientEpoch: r.clientEpoch, participantEpoch: r.participantEpoch}
	if l == nil {
		return snapshot, nil
	}
	l.lock()
	defer l.unlock()
	if !l.coherent(key, s.generation) {
		return CoordinationSnapshot{}, ErrIndeterminateExecution
	}
	snapshot.membership = l.membership
	snapshot.ledgerFenced = l.fenced
	snapshot.facts = detachedFacts(l.facts())
	return snapshot, nil
}

// RevalidateCoordinationSnapshot compares the entire detached observation.
// Foreign, stale, incomplete or contradictory snapshots reject without mutation,
// callback, participant disclosure or usable execution/recovery authority.
func (s *MemoryStorage) RevalidateCoordinationSnapshot(snapshot CoordinationSnapshot) error {
	if s == nil || s.coord == nil || s.coord.self != s || snapshot.owner != s.coord.owner || !snapshot.complete || !snapshot.scope.Valid() {
		return runtimecoordination.ErrInvalidScope
	}
	current, err := s.CaptureCoordinationSnapshot(snapshot.scope)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(current, snapshot) {
		return ErrIndeterminateExecution
	}
	return nil
}

func coherentState(state CommandState, revision Revision, hasOutcome bool, validOutcome bool) bool {
	return revision != 0 && ((state == CommandStateClaimed && !hasOutcome) || (state == CommandStateTerminal && hasOutcome && validOutcome))
}

func (l *commandLedger) coherent(scope instanceScope, generation uint64) bool {
	for id, r := range l.records {
		if r == nil || r.identity != id || id.scope.instanceScope() != scope || id.key == "" || !id.scope.validPrimitive() || !r.intent.validFor(id.scope) || (r.state == CommandStateClaimed && r.outcome != (TerminalOutcome{})) || !coherentState(r.state, r.revision, r.hasOutcome, r.outcome.valid()) {
			return false
		}
	}
	for id, r := range l.parents {
		if r == nil || r.identity != id || id.scope.instanceScope() != scope || id.key == "" || !id.scope.validParent() || !r.intent.validParentFor(id.scope) || (r.state == CommandStateClaimed && r.outcome != (ParentTerminalOutcome{})) || !coherentState(r.state, r.revision, r.hasOutcome, r.outcome.valid()) {
			return false
		}
	}
	for id, r := range l.phases {
		expected, valid := newPhaseIdentity(id.parent, id.kind)
		parent := l.parents[id.parent]
		if r == nil || r.identity != id || !valid || id != expected || parent == nil || (r.state == CommandStateClaimed && r.outcome != (TerminalOutcome{})) || !coherentState(r.state, r.revision, r.hasOutcome, r.outcome.valid()) || (parent.state == CommandStateTerminal && r.state != CommandStateTerminal) {
			return false
		}
	}
	for id, parent := range l.parents {
		stopID, _ := newPhaseIdentity(id, PhaseStopOld)
		startID, _ := newPhaseIdentity(id, PhaseStartTarget)
		stop, start := l.phases[stopID], l.phases[startID]
		if start != nil && stop != nil && stop.state != CommandStateTerminal {
			return false
		}
		if parent.state == CommandStateTerminal && (stop != nil || start != nil) {
			if (stop != nil && stop.state != CommandStateTerminal) || (start == nil && !mayOmitStartTarget(l.rendezvous[id], stop, parent.outcome)) || (start != nil && start.state != CommandStateTerminal) {
				return false
			}
		}
	}
	for id, state := range l.live {
		r := l.records[id]
		if state == nil || state.generation == 0 || state.generation > generation || r == nil || r.state != CommandStateClaimed || r.revision != state.revision {
			return false
		}
	}
	for id, state := range l.liveParents {
		r := l.parents[id]
		if state == nil || state.generation == 0 || state.generation > generation || r == nil || r.state != CommandStateClaimed || r.revision != state.revision {
			return false
		}
	}
	for id, state := range l.livePhases {
		r := l.phases[id]
		if state == nil || state.generation == 0 || state.generation > generation || r == nil || r.state != CommandStateClaimed || r.revision != state.revision {
			return false
		}
	}
	for id, slot := range l.stopForStart {
		start := l.records[id]
		if start == nil || id.scope.operation != OperationStart || start.state != CommandStateClaimed || (slot.command == nil) == (slot.phase == nil) {
			return false
		}
		if slot.command != nil {
			stop := l.records[*slot.command]
			if stop == nil || slot.command.scope.operation != OperationStop || slot.command.scope.instanceScope() != scope {
				return false
			}
		}
		if slot.phase != nil {
			if slot.phase.kind != PhaseStopOld || l.phases[*slot.phase] == nil {
				return false
			}
		}
	}
	for id, r := range l.rendezvous {
		if l.parents[id] == nil || !l.coherentRendezvous(r, generation) {
			return false
		}
		if r.startPhaseClaimed {
			phase, _ := newPhaseIdentity(id, PhaseStartTarget)
			if l.phases[phase] == nil {
				return false
			}
		}
	}
	for id, r := range l.managedStart {
		if r == nil || r.state == nil || r.generation == 0 || r.generation > generation || !l.coherentRendezvous(r.bridge, generation) || r.stage > managedStageBlocked {
			return false
		}
		var ownerScope Scope
		var intent Intent
		if r.phase != (phaseIdentity{}) {
			phase := l.phases[r.phase]
			parent := l.parents[r.phase.parent]
			if r.phase.kind != PhaseStartTarget || r.identity != r.phase.parent || parent == nil || phase == nil || l.livePhases[r.phase] != r.state {
				return false
			}
			ownerScope, intent = parent.identity.scope, parent.intent
		} else {
			record := l.records[r.identity]
			if record == nil || r.identity.scope.operation != OperationStart || l.live[r.identity] != r.state {
				return false
			}
			ownerScope, intent = record.identity.scope, record.intent
		}
		if r.binding != (runtimeorchestrationbinding.StartExecutionBinding{}) {
			expected, valid := authorizeCommandMaps(ownerScope, intent)
			if !valid || !r.binding.Valid() || r.binding.Rendezvous() != id || r.binding.Authorization() != expected {
				return false
			}
			linked, isLinked := r.binding.LinkedExecutionIdentity()
			if r.phase != (phaseIdentity{}) {
				if !isLinked || linked.Parent().Authorization() != expected || linked.Parent().CommandKey() != string(r.identity.key) || linked.Phase().Ordinal() != r.phase.ordinal {
					return false
				}
			} else if isLinked {
				return false
			}
		}

	}
	return true
}

func (l *commandLedger) coherentRendezvous(r *startRendezvous, generation uint64) bool {
	if r == nil || r.generation == 0 || r.generation > generation || r.notify == nil || r.signal > startSignalBlocked || r.resolution > stopResolutionBlocked {
		return false
	}
	if r.stopConsumed {
		stop := l.records[r.stop]
		if stop == nil || r.stop.scope.operation != OperationStop {
			return false
		}
	}
	if r.stopState != nil {
		if !r.stopConsumed {
			return false
		}
		stop := l.records[r.stop]
		if stop == nil || r.stopState.generation != r.generation {
			return false
		}
		switch r.resolution {
		case stopResolutionNone:
			if r.generation == generation && l.live[r.stop] != r.stopState {
				return false
			}
			if stop.state != CommandStateClaimed || stop.revision != r.stopState.revision {
				return false
			}
		case stopResolutionConverged, stopResolutionSatisfiedNoClaim:
			if stop.state != CommandStateTerminal || !stop.hasOutcome || r.stopState.revision == ^Revision(0) || stop.revision != r.stopState.revision+1 {
				return false
			}
			if r.resolution == stopResolutionSatisfiedNoClaim {
				if stop.outcome.category != OutcomeSucceeded || stop.outcome.launchAttemptID != "" {
					return false
				}
			} else if !r.stopConverged || stop.outcome.category != OutcomeSucceeded || stop.outcome.launchAttemptID != r.attempt {
				return false
			}
		case stopResolutionBlocked:
			if stop.state == CommandStateClaimed {
				if stop.revision != r.stopState.revision {
					return false
				}
			} else if r.stopState.revision == ^Revision(0) || stop.revision != r.stopState.revision+1 {
				return false
			}
		}
	}
	if r.signal == startSignalOwnerClaimed && r.attempt == "" {
		return false
	}
	if r.signal == startSignalNoClaim && r.startPhaseClaimed && !r.startNoClaimCause.Valid() {
		return false
	}
	return true
}

func sameLedgerFacts(a, b ledgerState) bool {
	if !reflect.DeepEqual(a, b) {
		return false
	}
	for id, v := range a.live {
		if v.identity != b.live[id].identity {
			return false
		}
	}
	for id, v := range a.liveParents {
		if v.identity != b.liveParents[id].identity {
			return false
		}
	}
	for id, v := range a.livePhases {
		if v.identity != b.livePhases[id].identity {
			return false
		}
	}
	for id, v := range a.rendezvous {
		if v.value.stopState != b.rendezvous[id].value.stopState {
			return false
		}
	}
	for id, v := range a.managed {
		w := b.managed[id]
		if v.value.state != w.value.state || v.value.bridge != w.value.bridge || v.bridge.value.stopState != w.bridge.value.stopState {
			return false
		}
	}
	return true
}
func (l *commandLedger) writableLocked() bool {
	if l.membership == ^uint64(0) {
		l.fenced = true
	}
	return !l.fenced
}

func (s *MemoryStorage) fenceClient() {
	s.clientMu.Lock()
	defer s.clientMu.Unlock()
	s.fenceClientLocked()
}
func (s *MemoryStorage) fenceClientLocked() {
	if !s.clientFenced {
		s.clientFenced = true
		close(s.generationDone)
	}
}

func (s *MemoryStorage) checkDelegation(scope instanceScope) error {
	r := s.lockRegion(scope)
	defer r.Close()
	if r.closed {
		return runtimecoordination.ErrClosed
	}
	return nil
}

func (l *commandLedger) primitiveDelegable(id commandIdentity, state *permitState) bool {
	r := l.records[id]
	return state != nil && r != nil && r.state == CommandStateClaimed && r.revision != ^Revision(0) && r.revision == state.revision && l.live[id] == state
}
func (l *commandLedger) phaseDelegable(id phaseIdentity, state *permitState) bool {
	r := l.phases[id]
	return state != nil && r != nil && r.state == CommandStateClaimed && r.revision != ^Revision(0) && r.revision == state.revision && l.livePhases[id] == state
}
