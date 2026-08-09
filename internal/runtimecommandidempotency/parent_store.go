package runtimecommandidempotency

import (
	"context"
	"sync"
)

type parentRecord struct {
	identity   commandIdentity
	intent     Intent
	state      CommandState
	revision   Revision
	outcome    ParentTerminalOutcome
	hasOutcome bool
}

func (r parentRecord) view() ParentRecordView {
	return ParentRecordView{
		scope: r.identity.scope, key: r.identity.key, intent: r.intent,
		state: r.state, revision: r.revision, outcome: r.outcome,
		hasOutcome: r.hasOutcome,
	}
}

type phaseIdentity struct {
	parent  commandIdentity
	kind    PhaseKind
	ordinal uint8
}

func newPhaseIdentity(parent commandIdentity, kind PhaseKind) (phaseIdentity, bool) {
	switch kind {
	case PhaseStopOld:
		return phaseIdentity{parent: parent, kind: kind, ordinal: 0}, true
	case PhaseStartTarget:
		return phaseIdentity{parent: parent, kind: kind, ordinal: 1}, true
	default:
		return phaseIdentity{}, false
	}
}

type phaseRecord struct {
	identity   phaseIdentity
	state      CommandState
	revision   Revision
	outcome    TerminalOutcome
	hasOutcome bool
}

func (r phaseRecord) view() PhaseRecordView {
	return PhaseRecordView{
		parentScope: r.identity.parent.scope, parentKey: r.identity.parent.key,
		kind: r.identity.kind, ordinal: r.identity.ordinal, state: r.state,
		revision: r.revision, outcome: r.outcome, hasOutcome: r.hasOutcome,
	}
}

// ExecuteParent validates and authorizes one exact replacement or rollback
// submission, then synchronously delegates a newly committed parent claim to
// one callback-scoped ParentExecution. The callback may execute the finite
// StopOld then gated StartTarget sequence; replay receives no capability.
func (b *Boundary) ExecuteParent(
	ctx context.Context,
	scope Scope,
	key CommandKey,
	intent Intent,
	authorize Authorize,
	invoke func(*ParentExecution) error,
) (ParentAdmission, error) {
	if b == nil || b.storage == nil || ctx == nil || !scope.validParent() ||
		key == "" || !intent.validParentFor(scope) || authorize == nil || invoke == nil {
		return ParentAdmission{}, ErrInvalidSubmission
	}
	if err := authorizeParentSafely(authorize, ctx, scope, intent); err != nil {
		return ParentAdmission{}, err
	}
	if err := ctx.Err(); err != nil {
		return ParentAdmission{}, err
	}

	b.storage.clientMu.RLock()
	if b.storage.generation != b.generation {
		b.storage.clientMu.RUnlock()
		return ParentAdmission{}, ErrBoundaryExpired
	}
	ledger := b.storage.ledger(scope.instanceScope())
	identity := commandIdentity{scope: scope, key: key}
	ledger.mu.Lock()
	if existing, ok := ledger.parents[identity]; ok {
		if existing.intent != intent {
			ledger.mu.Unlock()
			b.storage.clientMu.RUnlock()
			return ParentAdmission{}, ErrCommandKeyConflict
		}
		view := existing.view()
		kind := AdmissionInProgress
		if existing.state == CommandStateTerminal {
			kind = AdmissionReplay
		}
		ledger.mu.Unlock()
		b.storage.clientMu.RUnlock()
		return ParentAdmission{kind: kind, record: view}, nil
	}
	if ledger.hasAnyNonterminalLocked() {
		ledger.mu.Unlock()
		b.storage.clientMu.RUnlock()
		return ParentAdmission{}, ErrInstanceBlocked
	}
	record := &parentRecord{
		identity: identity, intent: intent, state: CommandStateClaimed, revision: 1,
	}
	state := &permitState{generation: b.generation, revision: record.revision}
	ledger.parents[identity] = record
	ledger.liveParents[identity] = state
	ledger.rendezvous[identity] = newStartRendezvous(b.generation)
	execution := &ParentExecution{
		boundary: b, ledger: ledger, identity: identity, state: state,
		usage: &parentUsage{},
	}
	execution.usage.cond = sync.NewCond(&execution.usage.mu)
	claimed := ParentAdmission{kind: AdmissionClaimed, record: record.view()}
	ledger.mu.Unlock()
	b.storage.clientMu.RUnlock()

	defer execution.closeAndExpire()
	callbackReturned := invokeParentSafely(invoke, execution)
	execution.closeAndWait()

	b.storage.clientMu.RLock()
	ledger.mu.Lock()
	current := ledger.parents[identity]
	if current != nil && current.state == CommandStateTerminal {
		claimed.record = current.view()
		ledger.mu.Unlock()
		b.storage.clientMu.RUnlock()
		return claimed, nil
	}
	ledger.mu.Unlock()
	b.storage.clientMu.RUnlock()
	if !callbackReturned {
		return claimed, ErrIndeterminateExecution
	}
	return claimed, ErrIndeterminateExecution
}

func authorizeParentSafely(
	authorize Authorize,
	ctx context.Context,
	scope Scope,
	intent Intent,
) (err error) {
	defer func() {
		if recover() != nil {
			err = ErrInvalidSubmission
		}
	}()
	return authorize(ctx, scope, intent)
}

func invokeParentSafely(invoke func(*ParentExecution) error, execution *ParentExecution) (definitive bool) {
	defer func() {
		if recover() != nil {
			definitive = false
		}
	}()
	return invoke(execution) == nil
}

func (l *commandLedger) hasNonterminalParentOrPhaseLocked() bool {
	for _, record := range l.parents {
		if record.state != CommandStateTerminal {
			return true
		}
	}
	for _, record := range l.phases {
		if record.state != CommandStateTerminal {
			return true
		}
	}
	return false
}

func (l *commandLedger) hasAnyNonterminalLocked() bool {
	for _, record := range l.records {
		if record.state != CommandStateTerminal {
			return true
		}
	}
	return l.hasNonterminalParentOrPhaseLocked()
}

// ParentExecution is a callback-scoped, generation-bound capability for one
// newly committed parent. Retaining it after ExecuteParent returns does not
// restore authority.
type ParentExecution struct {
	boundary *Boundary
	ledger   *commandLedger
	identity commandIdentity
	state    *permitState
	usage    *parentUsage
}

type parentUsage struct {
	mu       sync.Mutex
	cond     *sync.Cond
	closing  bool
	inFlight uint64
}

// InspectOrExecuteStopOld inspects or claims the optional ordinal-zero phase.
// Only a new phase claim invokes lifecycle work, synchronously and at most once.
func (p *ParentExecution) InspectOrExecuteStopOld(
	invoke func() (TerminalOutcome, error),
) (PhaseAdmission, error) {
	return p.inspectOrExecutePhase(PhaseStopOld, invoke)
}

// ContinueOrExecuteStartTarget atomically orders one StartTarget phase against
// an independent Stop on the same Runtime Instance. A new phase invokes the
// callback synchronously with callback-scoped OwnerClaimed and StartNoClaim
// signals. A true bool with nil error means Stop won before phase creation and
// terminalized satisfied without lifecycle work. A true bool with a context
// error means caller cancellation or a pre-phase Stop cancellation won. Both
// return a zero PhaseAdmission. A false bool with an error means fail-closed
// indeterminacy or expiry; after a phase claim its Admission remains returned.
// A Stop admitted after the phase claim keeps its permit on its original
// Boundary.Execute stack and rendezvous there.
func (p *ParentExecution) ContinueOrExecuteStartTarget(
	ctx context.Context,
	invoke func(*StartTargetExecution) (TerminalOutcome, error),
) (PhaseAdmission, bool, error) {
	if p == nil || p.boundary == nil || p.ledger == nil || p.state == nil ||
		ctx == nil || invoke == nil {
		return PhaseAdmission{}, false, ErrInvalidSubmission
	}
	if !p.beginUse() {
		return PhaseAdmission{}, false, ErrBoundaryExpired
	}
	defer p.endUse()
	identity, _ := newPhaseIdentity(p.identity, PhaseStartTarget)
	p.boundary.storage.clientMu.RLock()
	p.ledger.mu.Lock()
	if !p.liveLocked() {
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{}, false, ErrBoundaryExpired
	}
	if existing := p.ledger.phases[identity]; existing != nil {
		view := existing.view()
		kind := AdmissionInProgress
		if existing.state == CommandStateTerminal {
			kind = AdmissionReplay
		}
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{kind: kind, record: view}, false, nil
	}
	if !p.phaseOrderAllowsLocked(PhaseStartTarget) {
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{}, false, ErrIllegalPhaseOrder
	}
	rendezvous := p.ledger.rendezvous[p.identity]
	if rendezvous == nil || rendezvous.generation != p.boundary.generation ||
		rendezvous.signal == startSignalBlocked {
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{}, false, ErrBoundaryExpired
	}
	if rendezvous.continueCancelled {
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{}, true, context.Canceled
	}
	if rendezvous.stopCancelledBeforePhase {
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{}, true, context.Canceled
	}
	if err := ctx.Err(); err != nil {
		rendezvous.continueCancelled = true
		waitForStop := rendezvous.stopState != nil
		if waitForStop {
			rendezvous.signal = startSignalNoClaim
			rendezvous.notifyLocked()
		}
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		if waitForStop {
			_, _ = p.waitForStopFirst(rendezvous)
		}
		return PhaseAdmission{}, true, err
	}
	if rendezvous.stopState != nil {
		if rendezvous.signal != startSignalNone {
			p.ledger.mu.Unlock()
			p.boundary.storage.clientMu.RUnlock()
			return PhaseAdmission{}, false, ErrIndeterminateExecution
		}
		rendezvous.signal = startSignalNoClaim
		rendezvous.notifyLocked()
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		prevented, err := p.waitForStopFirst(rendezvous)
		return PhaseAdmission{}, prevented, err
	}
	record := &phaseRecord{identity: identity, state: CommandStateClaimed, revision: 1}
	state := &permitState{generation: p.boundary.generation, revision: record.revision}
	p.ledger.phases[identity] = record
	p.ledger.livePhases[identity] = state
	rendezvous.startPhaseClaimed = true
	permit := &phasePermit{parent: p, identity: identity, state: state}
	claimed := PhaseAdmission{kind: AdmissionClaimed, record: record.view()}
	p.ledger.mu.Unlock()
	p.boundary.storage.clientMu.RUnlock()

	terminal, err := permit.executeStartTarget(invoke)
	if err != nil {
		return claimed, false, err
	}
	claimed.record = terminal
	return claimed, false, nil
}

func (p *ParentExecution) inspectOrExecutePhase(
	kind PhaseKind,
	invoke func() (TerminalOutcome, error),
) (PhaseAdmission, error) {
	if p == nil || p.boundary == nil || p.ledger == nil || p.state == nil || invoke == nil {
		return PhaseAdmission{}, ErrInvalidSubmission
	}
	if !p.beginUse() {
		return PhaseAdmission{}, ErrBoundaryExpired
	}
	defer p.endUse()
	identity, ok := newPhaseIdentity(p.identity, kind)
	if !ok {
		return PhaseAdmission{}, ErrInvalidSubmission
	}
	p.boundary.storage.clientMu.RLock()
	p.ledger.mu.Lock()
	if !p.liveLocked() {
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{}, ErrBoundaryExpired
	}
	if existing, exists := p.ledger.phases[identity]; exists {
		view := existing.view()
		admissionKind := AdmissionInProgress
		if existing.state == CommandStateTerminal {
			admissionKind = AdmissionReplay
		}
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{kind: admissionKind, record: view}, nil
	}
	if !p.phaseOrderAllowsLocked(kind) {
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{}, ErrIllegalPhaseOrder
	}
	record := &phaseRecord{identity: identity, state: CommandStateClaimed, revision: 1}
	state := &permitState{generation: p.boundary.generation, revision: record.revision}
	p.ledger.phases[identity] = record
	p.ledger.livePhases[identity] = state
	permit := &phasePermit{parent: p, identity: identity, state: state}
	claimed := PhaseAdmission{kind: AdmissionClaimed, record: record.view()}
	p.ledger.mu.Unlock()
	p.boundary.storage.clientMu.RUnlock()

	terminal, err := permit.execute(invoke)
	if err != nil {
		return claimed, err
	}
	claimed.record = terminal
	return claimed, nil
}

func (p *ParentExecution) liveLocked() bool {
	if p.boundary.storage.generation != p.boundary.generation ||
		p.state.generation != p.boundary.generation ||
		p.ledger.liveParents[p.identity] != p.state {
		return false
	}
	record := p.ledger.parents[p.identity]
	return record != nil && record.state == CommandStateClaimed &&
		record.revision == p.state.revision
}

func (p *ParentExecution) phaseOrderAllowsLocked(kind PhaseKind) bool {
	stopID, _ := newPhaseIdentity(p.identity, PhaseStopOld)
	startID, _ := newPhaseIdentity(p.identity, PhaseStartTarget)
	switch kind {
	case PhaseStopOld:
		return p.ledger.phases[startID] == nil
	case PhaseStartTarget:
		stop := p.ledger.phases[stopID]
		return stop == nil || stop.state == CommandStateTerminal
	default:
		return false
	}
}

// PublishTerminal publishes one definitive parent result. It cannot supply or
// replace child facts: if any phase exists, StartTarget and every existing
// phase must already be terminal. A cancellation winner permits only Cancelled;
// an exact converged rendezvous Stop or pre-phase Stop-first winner permits only
// Stopped. Post-phase StartNoClaim maps its immutable explicit
// Cancelled/Rejected/Failed cause without treating the satisfied Stop as
// converged.
func (p *ParentExecution) PublishTerminal(
	outcome ParentTerminalOutcome,
) (ParentRecordView, error) {
	if p == nil || !outcome.valid() {
		return ParentRecordView{}, ErrInvalidSubmission
	}
	if !p.beginUse() {
		return ParentRecordView{}, ErrBoundaryExpired
	}
	defer p.endUse()
	p.boundary.storage.clientMu.RLock()
	p.ledger.mu.Lock()
	defer p.ledger.mu.Unlock()
	defer p.boundary.storage.clientMu.RUnlock()
	if !p.liveLocked() {
		return ParentRecordView{}, ErrBoundaryExpired
	}
	stopID, _ := newPhaseIdentity(p.identity, PhaseStopOld)
	startID, _ := newPhaseIdentity(p.identity, PhaseStartTarget)
	stop := p.ledger.phases[stopID]
	start := p.ledger.phases[startID]
	rendezvous := p.ledger.rendezvous[p.identity]
	if rendezvous != nil {
		if (rendezvous.continueCancelled || rendezvous.stopCancelledBeforePhase) &&
			outcome.category != ParentOutcomeCancelled {
			return ParentRecordView{}, ErrInstanceBlocked
		} else if !rendezvous.continueCancelled && !rendezvous.stopCancelledBeforePhase &&
			(rendezvous.stopConverged || rendezvous.stopFirstWon) &&
			outcome.category != ParentOutcomeStopped {
			return ParentRecordView{}, ErrInstanceBlocked
		} else if start != nil && start.state == CommandStateTerminal &&
			rendezvous.signal == startSignalNoClaim {
			want := ParentOutcomeRejected
			switch rendezvous.startNoClaimCause {
			case StartNoClaimCancelled:
				want = ParentOutcomeCancelled
			case StartNoClaimFailed:
				want = ParentOutcomeFailed
			}
			if outcome.category != want {
				return ParentRecordView{}, ErrInstanceBlocked
			}
		}
	}
	if rendezvous != nil && rendezvous.stopState != nil &&
		rendezvous.resolution != stopResolutionConverged &&
		rendezvous.resolution != stopResolutionSatisfiedNoClaim {
		return ParentRecordView{}, ErrInstanceBlocked
	}
	if stop != nil || start != nil {
		if start == nil || start.state != CommandStateTerminal ||
			(stop != nil && stop.state != CommandStateTerminal) {
			return ParentRecordView{}, ErrInstanceBlocked
		}
	}
	record := p.ledger.parents[p.identity]
	record.state = CommandStateTerminal
	record.revision++
	record.outcome = outcome
	record.hasOutcome = true
	delete(p.ledger.liveParents, p.identity)
	return record.view(), nil
}

func (p *ParentExecution) expire() {
	if p == nil || p.ledger == nil || p.state == nil {
		return
	}
	p.ledger.mu.Lock()
	defer p.ledger.mu.Unlock()
	if p.ledger.liveParents[p.identity] == p.state {
		delete(p.ledger.liveParents, p.identity)
	}
	if rendezvous := p.ledger.rendezvous[p.identity]; rendezvous != nil &&
		rendezvous.resolution == stopResolutionNone && rendezvous.stopState != nil {
		rendezvous.signal = startSignalBlocked
		rendezvous.notifyLocked()
	}
}

func (p *ParentExecution) beginUse() bool {
	if p == nil || p.usage == nil || p.usage.cond == nil {
		return false
	}
	p.usage.mu.Lock()
	defer p.usage.mu.Unlock()
	if p.usage.closing {
		return false
	}
	p.usage.inFlight++
	return true
}

func (p *ParentExecution) endUse() {
	p.usage.mu.Lock()
	p.usage.inFlight--
	if p.usage.inFlight == 0 {
		p.usage.cond.Broadcast()
	}
	p.usage.mu.Unlock()
}

func (p *ParentExecution) closeAndWait() {
	p.usage.mu.Lock()
	p.usage.closing = true
	for p.usage.inFlight != 0 {
		p.usage.cond.Wait()
	}
	p.usage.mu.Unlock()
}

func (p *ParentExecution) closeAndExpire() {
	p.closeAndWait()
	p.expire()
}

type phasePermit struct {
	parent   *ParentExecution
	identity phaseIdentity
	state    *permitState
}

func (p *phasePermit) execute(
	invoke func() (TerminalOutcome, error),
) (PhaseRecordView, error) {
	if p == nil || p.parent == nil || p.state == nil || invoke == nil {
		return PhaseRecordView{}, ErrInvalidSubmission
	}
	defer p.expire()
	p.parent.boundary.storage.clientMu.RLock()
	p.parent.ledger.mu.Lock()
	if !p.parent.liveLocked() ||
		p.parent.ledger.livePhases[p.identity] != p.state {
		p.parent.ledger.mu.Unlock()
		p.parent.boundary.storage.clientMu.RUnlock()
		return PhaseRecordView{}, ErrBoundaryExpired
	}
	p.parent.ledger.mu.Unlock()
	p.parent.boundary.storage.clientMu.RUnlock()

	outcome, definitive := invokeSafely(invoke)
	if !definitive || !outcome.valid() {
		return PhaseRecordView{}, ErrIndeterminateExecution
	}

	p.parent.boundary.storage.clientMu.RLock()
	p.parent.ledger.mu.Lock()
	defer p.parent.ledger.mu.Unlock()
	defer p.parent.boundary.storage.clientMu.RUnlock()
	if !p.parent.liveLocked() || p.parent.ledger.livePhases[p.identity] != p.state {
		return PhaseRecordView{}, ErrBoundaryExpired
	}
	record := p.parent.ledger.phases[p.identity]
	if record == nil || record.state != CommandStateClaimed ||
		record.revision != p.state.revision {
		return PhaseRecordView{}, ErrBoundaryExpired
	}
	record.state = CommandStateTerminal
	record.revision++
	record.outcome = outcome
	record.hasOutcome = true
	delete(p.parent.ledger.livePhases, p.identity)
	return record.view(), nil
}

func (p *phasePermit) expire() {
	p.parent.ledger.mu.Lock()
	defer p.parent.ledger.mu.Unlock()
	if p.parent.ledger.livePhases[p.identity] == p.state {
		delete(p.parent.ledger.livePhases, p.identity)
	}
}
