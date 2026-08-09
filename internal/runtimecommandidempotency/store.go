package runtimecommandidempotency

import (
	"context"
	"sync"
)

type commandIdentity struct {
	scope Scope
	key   CommandKey
}

type commandRecord struct {
	identity   commandIdentity
	intent     Intent
	state      CommandState
	revision   Revision
	outcome    TerminalOutcome
	hasOutcome bool
}

func (r commandRecord) view() RecordView {
	return RecordView{
		scope:      r.identity.scope,
		key:        r.identity.key,
		intent:     r.intent,
		state:      r.state,
		revision:   r.revision,
		outcome:    r.outcome,
		hasOutcome: r.hasOutcome,
	}
}

type permitState struct {
	generation uint64
	revision   Revision
}

type commandLedger struct {
	mu           sync.Mutex
	records      map[commandIdentity]*commandRecord
	live         map[commandIdentity]*permitState
	stopForStart map[commandIdentity]commandIdentity
	parents      map[commandIdentity]*parentRecord
	phases       map[phaseIdentity]*phaseRecord
	liveParents  map[commandIdentity]*permitState
	livePhases   map[phaseIdentity]*permitState
	rendezvous   map[commandIdentity]*startRendezvous
}

// MemoryStorage owns process-lifetime command facts independently from one
// Boundary client. Constructing a new Boundary preserves records but expires
// all permits issued by the previous client generation.
type MemoryStorage struct {
	clientMu       sync.RWMutex
	generation     uint64
	generationDone chan struct{}
	mu             sync.Mutex
	ledgers        map[instanceScope]*commandLedger
}

// NewMemoryStorage constructs empty process-local command storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		generationDone: make(chan struct{}),
		ledgers:        make(map[instanceScope]*commandLedger),
	}
}

func (s *MemoryStorage) nextGeneration() (uint64, <-chan struct{}) {
	s.clientMu.Lock()
	defer s.clientMu.Unlock()
	close(s.generationDone)
	s.generationDone = make(chan struct{})
	s.generation++
	return s.generation, s.generationDone
}

func (s *MemoryStorage) ledger(scope instanceScope) *commandLedger {
	s.mu.Lock()
	defer s.mu.Unlock()
	ledger, ok := s.ledgers[scope]
	if !ok {
		ledger = &commandLedger{
			records:      make(map[commandIdentity]*commandRecord),
			live:         make(map[commandIdentity]*permitState),
			stopForStart: make(map[commandIdentity]commandIdentity),
			parents:      make(map[commandIdentity]*parentRecord),
			phases:       make(map[phaseIdentity]*phaseRecord),
			liveParents:  make(map[commandIdentity]*permitState),
			livePhases:   make(map[phaseIdentity]*permitState),
			rendezvous:   make(map[commandIdentity]*startRendezvous),
		}
		s.ledgers[scope] = ledger
	}
	return ledger
}

// Boundary is one active storage client and process-local permit issuer.
// Reconstructing it over the same MemoryStorage preserves durable facts while
// making every earlier live permit unusable.
type Boundary struct {
	storage        *MemoryStorage
	generation     uint64
	generationDone <-chan struct{}
}

// NewBoundary constructs the only active client generation for storage.
func NewBoundary(storage *MemoryStorage) (*Boundary, error) {
	if storage == nil {
		return nil, ErrInvalidSubmission
	}
	generation, generationDone := storage.nextGeneration()
	return &Boundary{storage: storage, generation: generation, generationDone: generationDone}, nil
}

// Execute validates and authorizes one exact primitive submission, then
// atomically inspects or claims its command identity. A new claim keeps one
// private permit on this synchronous call stack. Ordinary claims invoke the
// callback immediately; a Stop admitted against gated StartTarget waits here
// for OwnerClaimed or StartNoClaim. Cancellation before a phase claim is an
// irreversible Stop-first cancellation; cancellation after the phase claim
// removes the active wait but still consumes the Stop slot. Neither invokes
// the callback. Authorization runs on replay.
func (b *Boundary) Execute(
	ctx context.Context,
	scope Scope,
	key CommandKey,
	intent Intent,
	authorize Authorize,
	invoke func() (TerminalOutcome, error),
) (Admission, error) {
	if b == nil || b.storage == nil || ctx == nil || !scope.validPrimitive() || key == "" ||
		!intent.validFor(scope) || authorize == nil || invoke == nil {
		return Admission{}, ErrInvalidSubmission
	}
	if err := authorize(ctx, scope, intent); err != nil {
		return Admission{}, err
	}
	if err := ctx.Err(); err != nil {
		return Admission{}, err
	}

	// The read lock makes the active-generation check and command admission one
	// linearization region against NewBoundary's generation transition. A stale
	// client therefore cannot insert a Claim after reconstruction.
	b.storage.clientMu.RLock()
	if b.storage.generation != b.generation {
		b.storage.clientMu.RUnlock()
		return Admission{}, ErrBoundaryExpired
	}
	ledger := b.storage.ledger(scope.instanceScope())
	identity := commandIdentity{scope: scope, key: key}
	ledger.mu.Lock()

	if existing, ok := ledger.records[identity]; ok {
		if existing.intent != intent {
			ledger.mu.Unlock()
			b.storage.clientMu.RUnlock()
			return Admission{}, ErrCommandKeyConflict
		}
		view := existing.view()
		if existing.state == CommandStateTerminal {
			ledger.mu.Unlock()
			b.storage.clientMu.RUnlock()
			return Admission{kind: AdmissionReplay, record: view}, nil
		}
		ledger.mu.Unlock()
		b.storage.clientMu.RUnlock()
		return Admission{kind: AdmissionInProgress, record: view}, nil
	}

	allowed, trackedStart, pending := b.mayClaimLocked(ledger, scope.operation)
	if !allowed {
		ledger.mu.Unlock()
		b.storage.clientMu.RUnlock()
		return Admission{}, ErrInstanceBlocked
	}

	record := &commandRecord{
		identity: identity,
		intent:   intent,
		state:    CommandStateClaimed,
		revision: 1,
	}
	state := &permitState{generation: b.generation, revision: record.revision}
	ledger.records[identity] = record
	ledger.live[identity] = state
	if trackedStart != nil {
		ledger.stopForStart[*trackedStart] = identity
	}
	permit := &executionPermit{
		boundary: b,
		ledger:   ledger,
		identity: identity,
		state:    state,
		pending:  pending,
	}
	if pending != nil {
		pending.stop = identity
		pending.stopState = state
		pending.stopConsumed = true
		pending.notifyLocked()
	}
	claimed := Admission{kind: AdmissionClaimed, record: record.view()}
	ledger.mu.Unlock()
	b.storage.clientMu.RUnlock()

	terminal, err := permit.execute(ctx, invoke)
	if err != nil {
		if terminal.state == CommandStateTerminal {
			claimed.record = terminal
		}
		return claimed, err
	}
	claimed.record = terminal
	return claimed, nil
}

func (b *Boundary) mayClaimLocked(
	ledger *commandLedger,
	operation Operation,
) (bool, *commandIdentity, *startRendezvous) {
	if pending := b.pendingStopRendezvousLocked(ledger, operation); pending != nil {
		return true, nil, pending
	}
	if ledger.hasNonterminalParentOrPhaseLocked() {
		return false, nil, nil
	}
	var trackedStart *commandIdentity
	for identity, record := range ledger.records {
		if record.state == CommandStateTerminal {
			continue
		}
		live := ledger.live[identity]
		tracked := live != nil && live.generation == b.generation
		if tracked && identity.scope.operation == OperationStart {
			candidate := identity
			trackedStart = &candidate
			continue
		}
		return false, nil, nil
	}
	if trackedStart == nil {
		return true, nil, nil
	}
	if operation != OperationStop {
		return false, nil, nil
	}
	if _, consumed := ledger.stopForStart[*trackedStart]; consumed {
		return false, nil, nil
	}
	return true, trackedStart, nil
}

// executionPermit is a private non-replayable process-local capability for one
// exact newly committed claim. It never leaves Boundary.Execute.
type executionPermit struct {
	boundary *Boundary
	ledger   *commandLedger
	identity commandIdentity
	state    *permitState
	pending  *startRendezvous
}

// execute invokes lifecycle work at most once. A nil callback error requires a
// valid definitive outcome and publishes it atomically as Terminal. Any
// callback error is treated as indeterminate, expires the permit, and leaves
// the durable record Claimed.
func (p *executionPermit) execute(
	ctx context.Context,
	invoke func() (TerminalOutcome, error),
) (RecordView, error) {
	if p == nil || p.boundary == nil || p.ledger == nil || p.state == nil || invoke == nil {
		return RecordView{}, ErrInvalidSubmission
	}
	// Expire on every exit path, including runtime.Goexit in invoke. Goexit runs
	// deferred calls but never returns control to this function, so cleanup that
	// only followed invokeSafely would leave a lost permit falsely tracked.
	// Successful terminal publication removes the live entry first, making this
	// deferred cleanup a no-op.
	defer p.expire()
	if p.pending != nil {
		return p.executePendingStop(ctx, invoke)
	}

	p.boundary.storage.clientMu.RLock()
	p.ledger.mu.Lock()
	if p.boundary.storage.generation != p.boundary.generation ||
		p.state.generation != p.boundary.generation {
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		return RecordView{}, ErrBoundaryExpired
	}
	live := p.ledger.live[p.identity]
	if live != p.state {
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()
		return RecordView{}, ErrBoundaryExpired
	}
	p.ledger.mu.Unlock()
	p.boundary.storage.clientMu.RUnlock()

	outcome, definitive := invokeSafely(invoke)
	if !definitive {
		return RecordView{}, ErrIndeterminateExecution
	}
	if !outcome.valid() {
		return RecordView{}, ErrIndeterminateExecution
	}

	p.boundary.storage.clientMu.RLock()
	p.ledger.mu.Lock()
	defer p.ledger.mu.Unlock()
	defer p.boundary.storage.clientMu.RUnlock()
	if p.boundary.storage.generation != p.boundary.generation {
		delete(p.ledger.live, p.identity)
		return RecordView{}, ErrBoundaryExpired
	}
	record := p.ledger.records[p.identity]
	if record == nil || record.state != CommandStateClaimed ||
		p.ledger.live[p.identity] != p.state {
		return RecordView{}, ErrBoundaryExpired
	}
	record.state = CommandStateTerminal
	record.revision++
	record.outcome = outcome
	record.hasOutcome = true
	delete(p.ledger.live, p.identity)
	if p.identity.scope.operation == OperationStart {
		delete(p.ledger.stopForStart, p.identity)
	}
	return record.view(), nil
}

func invokeSafely(invoke func() (TerminalOutcome, error)) (
	outcome TerminalOutcome,
	definitive bool,
) {
	defer func() {
		if recover() != nil {
			outcome = TerminalOutcome{}
			definitive = false
		}
	}()
	outcome, err := invoke()
	return outcome, err == nil
}

func (p *executionPermit) expire() {
	p.ledger.mu.Lock()
	defer p.ledger.mu.Unlock()
	if p.ledger.live[p.identity] == p.state {
		delete(p.ledger.live, p.identity)
	}
	if p.identity.scope.operation == OperationStart {
		delete(p.ledger.stopForStart, p.identity)
	}
}
