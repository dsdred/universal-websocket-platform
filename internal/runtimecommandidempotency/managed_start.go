package runtimecommandidempotency

import (
	"context"
	"strconv"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

// ExecuteManagedStart authorizes and admits one exact primitive Start, then
// supplies the sole newly committed permit with one complete immutable managed
// binding. Replay and in-progress observations never receive execution authority.
func (b *Boundary) ExecuteManagedStart(
	ctx context.Context,
	scope Scope,
	key CommandKey,
	intent Intent,
	expectedAggregateRevision runtimeorchestrationbinding.AggregateRevision,
	executionGeneration runtimeorchestrationbinding.ExecutionGeneration,
	authorize AuthorizeOrchestration,
	invoke func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error),
) (Admission, error) {
	return b.executeManagedStart(ctx, scope, key, intent, expectedAggregateRevision,
		executionGeneration, authorize, nil, invoke)
}

// executeLateManagedStart claims before requesting the composition-owned
// execution generation. It is used only by replay-first orchestration.
func (b *Boundary) executeLateManagedStart(
	ctx context.Context,
	scope Scope,
	key CommandKey,
	intent Intent,
	expectedAggregateRevision runtimeorchestrationbinding.AggregateRevision,
	authorize AuthorizeOrchestration,
	provideGeneration ProvideExecutionGeneration,
	invoke func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error),
) (Admission, error) {
	return b.executeManagedStart(ctx, scope, key, intent, expectedAggregateRevision,
		"", authorize, provideGeneration, invoke)
}

func (b *Boundary) executeManagedStart(
	ctx context.Context, scope Scope, key CommandKey, intent Intent,
	expectedRevision runtimeorchestrationbinding.AggregateRevision,
	generation runtimeorchestrationbinding.ExecutionGeneration,
	authorize AuthorizeOrchestration,
	provide ProvideExecutionGeneration,
	invoke func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error),
) (Admission, error) {
	late := provide != nil
	if b == nil || b.storage == nil || ctx == nil || scope.operation != OperationStart ||
		!scope.validPrimitive() || key == "" || !intent.validFor(scope) ||
		expectedRevision == 0 || authorize == nil || invoke == nil || late == (generation != "") {
		return Admission{}, ErrInvalidSubmission
	}
	authorization, ok := authorizeCommandMaps(scope, intent)
	if !ok {
		return Admission{}, ErrInvalidSubmission
	}
	if err := authorizeOrchestration(authorize, ctx, scope, intent); err != nil {
		return Admission{}, err
	}
	if err := ctx.Err(); err != nil {
		return Admission{}, err
	}

	b.storage.clientMu.RLock()
	if b.storage.generation != b.generation {
		b.storage.clientMu.RUnlock()
		return Admission{}, ErrBoundaryExpired
	}
	ledger := b.storage.ledger(scope.instanceScope())
	identity := commandIdentity{scope: scope, key: key}
	ledger.mu.Lock()
	if existing := ledger.records[identity]; existing != nil {
		if existing.intent != intent {
			ledger.mu.Unlock()
			b.storage.clientMu.RUnlock()
			return Admission{}, ErrCommandKeyConflict
		}
		kind := AdmissionInProgress
		if existing.state == CommandStateTerminal {
			kind = AdmissionReplay
		}
		view := existing.view()
		ledger.mu.Unlock()
		b.storage.clientMu.RUnlock()
		return Admission{kind: kind, record: view}, nil
	}
	allowed, _, pending := b.mayClaimLocked(ledger, scope.operation)
	if !allowed || pending != nil {
		ledger.mu.Unlock()
		b.storage.clientMu.RUnlock()
		return Admission{}, ErrInstanceBlocked
	}
	var binding runtimeorchestrationbinding.StartExecutionBinding
	rendezvous, rendezvousErr := b.newStartRendezvous()
	if rendezvousErr != nil {
		ledger.mu.Unlock()
		b.storage.clientMu.RUnlock()
		return Admission{}, ErrInvalidSubmission
	}
	if !late {
		var err error
		binding, err = runtimeorchestrationbinding.NewPrimitiveStartExecutionBinding(
			authorization, expectedRevision, generation, rendezvous)
		if err != nil {
			ledger.mu.Unlock()
			b.storage.clientMu.RUnlock()
			return Admission{}, ErrInvalidSubmission
		}
	}
	record := &commandRecord{identity: identity, intent: intent, state: CommandStateClaimed, revision: 1}
	state := &permitState{generation: b.generation, revision: record.revision}
	ledger.records[identity] = record
	ledger.live[identity] = state
	permit := &executionPermit{boundary: b, ledger: ledger, identity: identity, state: state, managed: rendezvous}
	ledger.managedStart[rendezvous] = &managedStartRendezvous{
		binding: binding, identity: identity, state: state, generation: b.generation,
		bridge: newStartRendezvous(b.generation), stage: managedStagePreOwner,
	}
	claimed := Admission{kind: AdmissionClaimed, record: record.view()}
	ledger.mu.Unlock()
	b.storage.clientMu.RUnlock()

	var terminal RecordView
	var err error
	if late {
		terminal, err = permit.executeAfterGeneration(ctx, authorization, expectedRevision, provide, invoke)
	} else {
		terminal, err = permit.execute(ctx, func() (TerminalOutcome, error) { return invoke(binding) })
	}
	if err != nil {
		if terminal.state == CommandStateTerminal {
			claimed.record = terminal
		}
		return claimed, err
	}
	claimed.record = terminal
	return claimed, nil
}

func (b *Boundary) newStartRendezvous() (runtimeorchestrationbinding.StartRendezvous, error) {
	return runtimeorchestrationbinding.NewStartRendezvous(
		strconv.FormatUint(b.generation, 36) + ":" + strconv.FormatUint(b.rendezvousSeq.Add(1), 36))
}

func (p *executionPermit) executeAfterGeneration(
	ctx context.Context,
	authorization OrchestrationAuthorizationRequest,
	expectedAggregateRevision runtimeorchestrationbinding.AggregateRevision,
	provideGeneration ProvideExecutionGeneration,
	invoke func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error),
) (RecordView, error) {
	defer p.expire()
	generation, definitive := provideGenerationSafely(provideGeneration, ctx)
	if !definitive || generation == "" || ctx.Err() != nil {
		return RecordView{}, ErrIndeterminateExecution
	}
	p.boundary.storage.clientMu.RLock()
	p.ledger.mu.Lock()
	if p.boundary.storage.generation != p.boundary.generation ||
		p.ledger.live[p.identity] != p.state {
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		return RecordView{}, ErrBoundaryExpired
	}
	managed := p.ledger.managedStart[p.managed]
	if managed == nil || !managed.liveLocked(p.boundary, p.ledger) ||
		managed.binding != (runtimeorchestrationbinding.StartExecutionBinding{}) ||
		managed.stage != managedStagePreOwner || managed.bridge == nil {
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		return RecordView{}, ErrIndeterminateExecution
	}
	binding, err := runtimeorchestrationbinding.NewPrimitiveStartExecutionBinding(
		authorization, expectedAggregateRevision, generation, p.managed,
	)
	if err != nil {
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		return RecordView{}, ErrIndeterminateExecution
	}
	managed.binding = binding
	p.ledger.mu.Unlock()
	p.boundary.storage.clientMu.RUnlock()
	return p.execute(ctx, func() (TerminalOutcome, error) { return invoke(binding) })
}

func provideGenerationSafely(
	provide ProvideExecutionGeneration,
	ctx context.Context,
) (generation runtimeorchestrationbinding.ExecutionGeneration, definitive bool) {
	defer func() {
		if recover() != nil {
			generation = ""
			definitive = false
		}
	}()
	generation, err := provide(ctx)
	return generation, err == nil
}

// managedStartRendezvousLive is the private Slice-2R lookup proof used by the
// command-owned continuation implementation. Missing, foreign, stale, reused,
// or post-callback identities all fail closed.
func (b *Boundary) managedStartRendezvousLive(
	scope Scope,
	key CommandKey,
	rendezvous runtimeorchestrationbinding.StartRendezvous,
) bool {
	if b == nil || b.storage == nil || !scope.validPrimitive() || key == "" ||
		rendezvous == (runtimeorchestrationbinding.StartRendezvous{}) {
		return false
	}
	b.storage.clientMu.RLock()
	defer b.storage.clientMu.RUnlock()
	if b.storage.generation != b.generation {
		return false
	}
	ledger := b.storage.existingLedger(scope.instanceScope())
	if ledger == nil {
		return false
	}
	ledger.mu.Lock()
	defer ledger.mu.Unlock()
	entry, exists := ledger.managedStart[rendezvous]
	if !exists || entry == nil {
		return false
	}
	live := ledger.live[entry.identity]
	return entry.generation == b.generation &&
		entry.identity == (commandIdentity{scope: scope, key: key}) &&
		live != nil && live.generation == b.generation
}
