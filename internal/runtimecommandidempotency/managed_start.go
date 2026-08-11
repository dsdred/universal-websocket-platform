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
	if b == nil || b.storage == nil || ctx == nil || !scope.validPrimitive() ||
		scope.operation != OperationStart || key == "" || !intent.validFor(scope) ||
		expectedAggregateRevision == 0 || executionGeneration == "" ||
		authorize == nil || invoke == nil {
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

	if existing, exists := ledger.records[identity]; exists {
		if existing.intent != intent {
			ledger.mu.Unlock()
			b.storage.clientMu.RUnlock()
			return Admission{}, ErrCommandKeyConflict
		}
		view := existing.view()
		kind := AdmissionInProgress
		if existing.state == CommandStateTerminal {
			kind = AdmissionReplay
		}
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

	rendezvous, err := runtimeorchestrationbinding.NewStartRendezvous(
		strconv.FormatUint(b.generation, 36) + ":" +
			strconv.FormatUint(b.rendezvousSeq.Add(1), 36),
	)
	if err != nil {
		ledger.mu.Unlock()
		b.storage.clientMu.RUnlock()
		return Admission{}, ErrInvalidSubmission
	}
	binding, err := runtimeorchestrationbinding.NewPrimitiveStartExecutionBinding(
		authorization, expectedAggregateRevision, executionGeneration, rendezvous,
	)
	if err != nil {
		ledger.mu.Unlock()
		b.storage.clientMu.RUnlock()
		return Admission{}, ErrInvalidSubmission
	}

	record := &commandRecord{
		identity: identity, intent: intent, state: CommandStateClaimed, revision: 1,
	}
	state := &permitState{generation: b.generation, revision: record.revision}
	ledger.records[identity] = record
	ledger.live[identity] = state
	ledger.managedStart[rendezvous] = managedStartRendezvous{
		identity: identity, generation: b.generation,
	}
	permit := &executionPermit{
		boundary: b, ledger: ledger, identity: identity, state: state, managed: rendezvous,
	}
	claimed := Admission{kind: AdmissionClaimed, record: record.view()}
	ledger.mu.Unlock()
	b.storage.clientMu.RUnlock()

	terminal, executeErr := permit.execute(ctx, func() (TerminalOutcome, error) {
		return invoke(binding)
	})
	if executeErr != nil {
		if terminal.state == CommandStateTerminal {
			claimed.record = terminal
		}
		return claimed, executeErr
	}
	claimed.record = terminal
	return claimed, nil
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
	live := ledger.live[entry.identity]
	return exists && entry.generation == b.generation &&
		entry.identity == (commandIdentity{scope: scope, key: key}) &&
		live != nil && live.generation == b.generation
}
