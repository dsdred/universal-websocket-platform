package runtimecommandidempotency

import (
	"context"
	"strconv"
	"sync"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

// ManagedParentExecution is the sole callback-scoped source of linked managed
// StartTarget execution. It wraps, but does not widen, the accepted parent
// permit.
type ManagedParentExecution struct {
	parent        *ParentExecution
	authorization runtimeorchestrationbinding.OrchestrationAuthorizationRequest
}

// ExecuteManagedParent authorizes every submission and supplies the newly
// claimed Replace/Rollback parent with managed phase authority only.
func (b *Boundary) ExecuteManagedParent(
	ctx context.Context,
	scope Scope,
	key CommandKey,
	intent Intent,
	authorize AuthorizeOrchestration,
	invoke func(*ManagedParentExecution) error,
) (ParentAdmission, error) {
	if b == nil || b.storage == nil || ctx == nil || !scope.validParent() || key == "" ||
		!intent.validParentFor(scope) || authorize == nil || invoke == nil {
		return ParentAdmission{}, ErrInvalidSubmission
	}
	authorization, ok := authorizeCommandMaps(scope, intent)
	if !ok {
		return ParentAdmission{}, ErrInvalidSubmission
	}
	if err := authorizeOrchestration(authorize, ctx, scope, intent); err != nil {
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
	if existing := ledger.parents[identity]; existing != nil {
		if existing.intent != intent {
			ledger.mu.Unlock()
			b.storage.clientMu.RUnlock()
			return ParentAdmission{}, ErrCommandKeyConflict
		}
		kind := AdmissionInProgress
		if existing.state == CommandStateTerminal {
			kind = AdmissionReplay
		}
		view := existing.view()
		ledger.mu.Unlock()
		b.storage.clientMu.RUnlock()
		return ParentAdmission{kind: kind, record: view}, nil
	}
	if ledger.hasAnyNonterminalLocked() {
		ledger.mu.Unlock()
		b.storage.clientMu.RUnlock()
		return ParentAdmission{}, ErrInstanceBlocked
	}
	record := &parentRecord{identity: identity, intent: intent, state: CommandStateClaimed, revision: 1}
	state := &permitState{generation: b.generation, revision: record.revision}
	ledger.parents[identity] = record
	ledger.liveParents[identity] = state
	ledger.rendezvous[identity] = newStartRendezvous(b.generation)
	parent := &ParentExecution{boundary: b, ledger: ledger, identity: identity, state: state, usage: &parentUsage{}}
	parent.usage.cond = sync.NewCond(&parent.usage.mu)
	managed := &ManagedParentExecution{parent: parent, authorization: authorization}
	claimed := ParentAdmission{kind: AdmissionClaimed, record: record.view()}
	ledger.mu.Unlock()
	b.storage.clientMu.RUnlock()

	defer parent.closeAndExpire()
	definitive := invokeManagedParentSafely(invoke, managed)
	parent.closeAndWait()
	b.storage.clientMu.RLock()
	ledger.mu.Lock()
	if current := ledger.parents[identity]; current != nil && current.state == CommandStateTerminal {
		claimed.record = current.view()
		ledger.mu.Unlock()
		b.storage.clientMu.RUnlock()
		return claimed, nil
	}
	ledger.mu.Unlock()
	b.storage.clientMu.RUnlock()
	if !definitive {
		return claimed, ErrIndeterminateExecution
	}
	return claimed, ErrIndeterminateExecution
}

func invokeManagedParentSafely(
	invoke func(*ManagedParentExecution) error,
	execution *ManagedParentExecution,
) (definitive bool) {
	defer func() {
		if recover() != nil {
			definitive = false
		}
	}()
	return invoke(execution) == nil
}

// InspectOrExecuteStopOld preserves the accepted optional ordinal-zero phase.
func (p *ManagedParentExecution) InspectOrExecuteStopOld(
	invoke func() (TerminalOutcome, error),
) (PhaseAdmission, error) {
	if p == nil || p.parent == nil {
		return PhaseAdmission{}, ErrInvalidSubmission
	}
	return p.parent.InspectOrExecuteStopOld(invoke)
}

// PublishTerminal preserves parent publication without exposing the legacy
// StartTarget capability.
func (p *ManagedParentExecution) PublishTerminal(
	outcome ParentTerminalOutcome,
) (ParentRecordView, error) {
	if p == nil || p.parent == nil {
		return ParentRecordView{}, ErrInvalidSubmission
	}
	return p.parent.PublishTerminal(outcome)
}

// ContinueOrExecuteManagedStartTarget claims the derived ordinal-one phase and
// creates its complete linked binding. Existing phases are observations only.
func (p *ManagedParentExecution) ContinueOrExecuteManagedStartTarget(
	ctx context.Context,
	expectedAggregateRevision runtimeorchestrationbinding.AggregateRevision,
	executionGeneration runtimeorchestrationbinding.ExecutionGeneration,
	invoke func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error),
) (PhaseAdmission, bool, error) {
	if p == nil || p.parent == nil || ctx == nil || expectedAggregateRevision == 0 ||
		executionGeneration == "" || invoke == nil {
		return PhaseAdmission{}, false, ErrInvalidSubmission
	}
	parent := p.parent
	if !parent.beginUse() {
		return PhaseAdmission{}, false, ErrBoundaryExpired
	}
	defer parent.endUse()
	identity, _ := newPhaseIdentity(parent.identity, PhaseStartTarget)
	parent.boundary.storage.clientMu.RLock()
	parent.ledger.mu.Lock()
	if !parent.liveLocked() {
		parent.ledger.mu.Unlock()
		parent.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{}, false, ErrBoundaryExpired
	}
	if existing := parent.ledger.phases[identity]; existing != nil {
		kind := AdmissionInProgress
		if existing.state == CommandStateTerminal {
			kind = AdmissionReplay
		}
		view := existing.view()
		parent.ledger.mu.Unlock()
		parent.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{kind: kind, record: view}, false, nil
	}
	if !parent.phaseOrderAllowsLocked(PhaseStartTarget) {
		parent.ledger.mu.Unlock()
		parent.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{}, false, ErrIllegalPhaseOrder
	}
	legacy := parent.ledger.rendezvous[parent.identity]
	if legacy == nil || legacy.generation != parent.boundary.generation || legacy.signal == startSignalBlocked {
		parent.ledger.mu.Unlock()
		parent.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{}, false, ErrBoundaryExpired
	}
	if legacy.continueCancelled || legacy.stopCancelledBeforePhase {
		parent.ledger.mu.Unlock()
		parent.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{}, true, context.Canceled
	}
	if err := ctx.Err(); err != nil {
		legacy.continueCancelled = true
		wait := legacy.stopState != nil
		if wait {
			legacy.signal = startSignalNoClaim
			legacy.notifyLocked()
		}
		parent.ledger.mu.Unlock()
		parent.boundary.storage.clientMu.RUnlock()
		if wait {
			_, _ = parent.waitForStopFirst(legacy)
		}
		return PhaseAdmission{}, true, err
	}
	if legacy.stopState != nil {
		legacy.signal = startSignalNoClaim
		legacy.notifyLocked()
		parent.ledger.mu.Unlock()
		parent.boundary.storage.clientMu.RUnlock()
		prevented, err := parent.waitForStopFirst(legacy)
		return PhaseAdmission{}, prevented, err
	}

	record := &phaseRecord{identity: identity, state: CommandStateClaimed, revision: 1}
	state := &permitState{generation: parent.boundary.generation, revision: record.revision}
	parent.ledger.phases[identity] = record
	parent.ledger.livePhases[identity] = state
	legacy.startPhaseClaimed = true
	legacy.signal = startSignalBlocked // later Stops use only the managed rendezvous.
	rendezvous, err := runtimeorchestrationbinding.NewStartRendezvous(
		strconv.FormatUint(parent.boundary.generation, 36) + ":" +
			strconv.FormatUint(parent.boundary.rendezvousSeq.Add(1), 36),
	)
	if err != nil {
		delete(parent.ledger.phases, identity)
		delete(parent.ledger.livePhases, identity)
		parent.ledger.mu.Unlock()
		parent.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{}, false, ErrInvalidSubmission
	}
	parentIdentity, err := runtimeorchestrationbinding.NewParentCommandIdentity(p.authorization, string(parent.identity.key))
	if err != nil {
		parent.ledger.mu.Unlock()
		parent.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{}, false, ErrInvalidSubmission
	}
	phaseIdentityValue, _ := runtimeorchestrationbinding.DeriveStartTargetPhaseIdentity(parentIdentity)
	linked, _ := runtimeorchestrationbinding.NewLinkedExecutionIdentity(parentIdentity, phaseIdentityValue)
	binding, err := runtimeorchestrationbinding.NewLinkedStartExecutionBinding(
		p.authorization, expectedAggregateRevision, executionGeneration, linked, rendezvous,
	)
	if err != nil {
		parent.ledger.mu.Unlock()
		parent.boundary.storage.clientMu.RUnlock()
		return PhaseAdmission{}, false, ErrInvalidSubmission
	}
	managed := &managedStartRendezvous{
		binding: binding, phase: identity, state: state, generation: parent.boundary.generation,
		bridge: newStartRendezvous(parent.boundary.generation), stage: managedStagePreOwner,
	}
	parent.ledger.managedStart[rendezvous] = managed
	permit := &managedPhasePermit{parent: parent, identity: identity, state: state, rendezvous: rendezvous}
	claimed := PhaseAdmission{kind: AdmissionClaimed, record: record.view()}
	parent.ledger.mu.Unlock()
	parent.boundary.storage.clientMu.RUnlock()

	terminal, executeErr := permit.execute(invoke, binding)
	if executeErr != nil {
		return claimed, false, executeErr
	}
	claimed.record = terminal
	return claimed, false, nil
}

type managedPhasePermit struct {
	parent     *ParentExecution
	identity   phaseIdentity
	state      *permitState
	rendezvous runtimeorchestrationbinding.StartRendezvous
}

func (p *managedPhasePermit) execute(
	invoke func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error),
	binding runtimeorchestrationbinding.StartExecutionBinding,
) (PhaseRecordView, error) {
	defer p.expire()
	p.parent.boundary.storage.clientMu.RLock()
	p.parent.ledger.mu.Lock()
	if !p.parent.liveLocked() || p.parent.ledger.livePhases[p.identity] != p.state {
		p.parent.ledger.mu.Unlock()
		p.parent.boundary.storage.clientMu.RUnlock()
		return PhaseRecordView{}, ErrBoundaryExpired
	}
	p.parent.ledger.mu.Unlock()
	p.parent.boundary.storage.clientMu.RUnlock()
	outcome, definitive := invokeManagedStartSafely(invoke, binding)
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
	if record == nil || record.state != CommandStateClaimed || record.revision != p.state.revision {
		return PhaseRecordView{}, ErrBoundaryExpired
	}
	managed := p.parent.ledger.managedStart[p.rendezvous]
	if managed == nil || !managed.terminalCompatible(outcome) {
		return PhaseRecordView{}, ErrIndeterminateExecution
	}
	record.state = CommandStateTerminal
	record.revision++
	record.outcome = outcome
	record.hasOutcome = true
	delete(p.parent.ledger.livePhases, p.identity)
	return record.view(), nil
}

func invokeManagedStartSafely(
	invoke func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error),
	binding runtimeorchestrationbinding.StartExecutionBinding,
) (outcome TerminalOutcome, definitive bool) {
	defer func() {
		if recover() != nil {
			outcome = TerminalOutcome{}
			definitive = false
		}
	}()
	outcome, err := invoke(binding)
	return outcome, err == nil
}

func (p *managedPhasePermit) expire() {
	if p == nil || p.parent == nil {
		return
	}
	p.parent.ledger.mu.Lock()
	if managed := p.parent.ledger.managedStart[p.rendezvous]; managed != nil {
		managed.blockLocked()
	}
	delete(p.parent.ledger.managedStart, p.rendezvous)
	if p.parent.ledger.livePhases[p.identity] == p.state {
		delete(p.parent.ledger.livePhases, p.identity)
	}
	p.parent.ledger.mu.Unlock()
}
