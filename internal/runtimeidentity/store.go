package runtimeidentity

import (
	"sync"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

// runtimeInstanceAggregate is the mutable in-memory aggregate for one Runtime
// Instance. It is protected by its own mutex so that different Instances may
// progress independently (DP-014 §16, §22 proof 11).
type runtimeInstanceAggregate struct {
	mu sync.Mutex

	// --- Immutable identity binding (DP-014 §9, §15) ---

	workspaceID       uint64
	configurationID   uint64
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID

	// --- Conditional revision (DP-014 §16) ---

	revision Revision

	// --- Last Owner-confirmed desired and actual facts (DP-014 §7, §13) ---

	desired DesiredState
	actual  ActualState

	// --- Launch Attempt tracking (DP-014 §10, §15) ---

	// activeAttemptID is the identity of the currently active attempt, or
	// empty when no attempt is active.
	activeAttemptID  runtimeconfigload.LaunchAttemptID
	hasActiveAttempt bool

	// attempts is the ordered append-only history of all Launch Attempts ever
	// claimed for this Instance, including terminal ones.
	attempts []LaunchAttemptRecord

	// attemptIndex is a map from LaunchAttemptID to its index in attempts for
	// O(1) lookup and history-presence checking.
	attemptIndex map[runtimeconfigload.LaunchAttemptID]int
}

// advanceRevision increments the revision by one and returns the new value.
// Must be called under the aggregate mutex.
func (a *runtimeInstanceAggregate) advanceRevision() Revision {
	a.revision++
	return a.revision
}

// activeAttemptPtr returns a pointer to the active attempt record, or nil.
// Must be called under the aggregate mutex.
func (a *runtimeInstanceAggregate) activeAttemptPtr() *LaunchAttemptRecord {
	if !a.hasActiveAttempt || a.activeAttemptID == "" {
		return nil
	}
	idx, ok := a.attemptIndex[a.activeAttemptID]
	if !ok {
		return nil
	}
	return &a.attempts[idx]
}

// Store is the in-memory Runtime Instance aggregate store. It implements all
// nine conceptual operations from DP-014 §21 and enforces all acceptance proofs
// from DP-014 §22.
//
// Different Runtime Instances may progress independently (per-aggregate mutex).
// Same-Instance operations are serialized through that Instance's mutex.
// Store-level operations (AllocateCandidateIdentity, CreateRuntimeInstance)
// are protected by the Store-wide mutex only for the minimal identity
// registration step; they do not hold the store mutex while performing
// per-aggregate work.
type Store struct {
	mu sync.Mutex

	// allocated tracks every RuntimeInstanceID that has been allocated through
	// AllocateCandidateIdentity or used in CreateRuntimeInstance. An ID in this
	// set cannot be reused even if no aggregate was committed.
	allocated map[runtimeconfigload.RuntimeInstanceID]struct{}

	// instances maps RuntimeInstanceID to the committed aggregate.
	instances map[runtimeconfigload.RuntimeInstanceID]*runtimeInstanceAggregate
}

// NewStore constructs an empty in-memory Store ready to accept operations.
func NewStore() *Store {
	return &Store{
		allocated: make(map[runtimeconfigload.RuntimeInstanceID]struct{}),
		instances: make(map[runtimeconfigload.RuntimeInstanceID]*runtimeInstanceAggregate),
	}
}

// ─── §21 Operation 1: AllocateCandidateIdentity ──────────────────────────────

// AllocateCandidateIdentity registers a candidate RuntimeInstanceID within the
// management domain without publishing an aggregate. The ID is reserved so that
// a subsequent CreateRuntimeInstance can use it, and so that it cannot be
// reused by any later allocation.
//
// Returns ErrInvalidIdentity if id is empty.
// Returns ErrInstanceAlreadyExists if the ID is already allocated or committed.
//
// Allocation alone does not prove that an aggregate exists (DP-014 §9).
func (s *Store) AllocateCandidateIdentity(
	id runtimeconfigload.RuntimeInstanceID,
) error {
	if id == "" {
		return ErrInvalidIdentity
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.allocated[id]; exists {
		return ErrInstanceAlreadyExists
	}
	s.allocated[id] = struct{}{}
	return nil
}

// ─── §21 Operation 2: CreateRuntimeInstance ──────────────────────────────────

// CreateRuntimeInstance atomically publishes a new Runtime Instance aggregate
// with the given immutable identity binding, initial desired Stopped and actual
// Stopped facts, no active attempt, and empty history.
//
// The RuntimeInstanceID must have been allocated via AllocateCandidateIdentity,
// or it is allocated implicitly by this call. Either way the resulting aggregate
// is committed or nothing is committed (DP-014 §9, §14, §22 proof 1).
//
// Returns ErrInvalidIdentity if any identity field is empty or zero.
// Returns ErrInstanceAlreadyExists if the ID already identifies a committed
// aggregate (DP-014 §9, §22 proof 2).
func (s *Store) CreateRuntimeInstance(
	workspaceID uint64,
	configurationID uint64,
	id runtimeconfigload.RuntimeInstanceID,
) error {
	if workspaceID == 0 || configurationID == 0 || id == "" {
		return ErrInvalidIdentity
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// Reject if a committed aggregate already exists.
	if _, exists := s.instances[id]; exists {
		return ErrInstanceAlreadyExists
	}
	// Reserve identity (idempotent if AllocateCandidateIdentity was called
	// first).
	s.allocated[id] = struct{}{}
	// Atomically publish the complete initial aggregate.
	s.instances[id] = &runtimeInstanceAggregate{
		workspaceID:       workspaceID,
		configurationID:   configurationID,
		runtimeInstanceID: id,
		revision:          1,
		desired:           DesiredStateStopped,
		actual:            ActualStateStopped,
		attemptIndex:      make(map[runtimeconfigload.LaunchAttemptID]int),
	}
	return nil
}

// ─── §21 Operation 3: ReadRuntimeInstance ─────────────────────────────────────

// ReadRuntimeInstance returns one coherent view of the Runtime Instance
// aggregate at the current revision. The view captures immutable binding,
// desired and actual facts, revision, and the active attempt identity if any.
//
// Reads are observation only and do not advance revision (DP-014 §18,
// §22 proof 12).
//
// Returns ErrInstanceNotFound if no committed aggregate exists for id.
func (s *Store) ReadRuntimeInstance(
	id runtimeconfigload.RuntimeInstanceID,
) (RuntimeInstanceView, error) {
	if id == "" {
		return RuntimeInstanceView{}, ErrInvalidIdentity
	}
	agg, err := s.getAggregate(id)
	if err != nil {
		return RuntimeInstanceView{}, err
	}
	agg.mu.Lock()
	defer agg.mu.Unlock()
	view := RuntimeInstanceView{
		workspaceID:       agg.workspaceID,
		configurationID:   agg.configurationID,
		runtimeInstanceID: agg.runtimeInstanceID,
		revision:          agg.revision,
		desired:           agg.desired,
		actual:            agg.actual,
		activeAttemptID:   agg.activeAttemptID,
		hasActiveAttempt:  agg.hasActiveAttempt,
	}
	return view, nil
}

// ─── §21 Operation 4: ReadLaunchAttemptHistory ────────────────────────────────

// ReadLaunchAttemptHistory returns a complete detached copy of the append-only
// Launch Attempt history for the Runtime Instance. Every committed attempt and
// its exact version pin are included.
//
// Reads are observation only and do not advance revision (DP-014 §18).
// The returned slice is a deep copy; mutations by the caller do not affect the
// store.
//
// Returns ErrInstanceNotFound if no committed aggregate exists for id.
func (s *Store) ReadLaunchAttemptHistory(
	id runtimeconfigload.RuntimeInstanceID,
) ([]LaunchAttemptRecord, error) {
	if id == "" {
		return nil, ErrInvalidIdentity
	}
	agg, err := s.getAggregate(id)
	if err != nil {
		return nil, err
	}
	agg.mu.Lock()
	defer agg.mu.Unlock()
	if len(agg.attempts) == 0 {
		return []LaunchAttemptRecord{}, nil
	}
	history := make([]LaunchAttemptRecord, len(agg.attempts))
	copy(history, agg.attempts)
	return history, nil
}

// ─── §21 Operation 5: ConditionalClaimLaunchAttempt ──────────────────────────

// ConditionalClaimLaunchAttempt atomically validates and claims a new Launch
// Attempt for the Runtime Instance.
//
// The operation:
//  1. validates the expected aggregate revision (stale → ErrStaleRevision,
//     zero mutation);
//  2. validates that actual state permits Start (must be ActualStateStopped;
//     any active attempt → ErrActiveAttemptExists);
//  3. validates that attemptID is absent from the complete history
//     (→ ErrAttemptIDReused);
//  4. atomically appends the attempt, sets it as active, publishes desired
//     Started and actual Claimed facts, and advances revision.
//
// A rejected claim performs zero mutation (DP-014 §16, §22 proofs 3, 4, 9).
//
// Returns ClaimResult with Committed() true and the new Revision on success.
// Returns a definitive error and ClaimResult with Committed() false on rejection.
func (s *Store) ConditionalClaimLaunchAttempt(
	id runtimeconfigload.RuntimeInstanceID,
	expectedRevision Revision,
	attemptID runtimeconfigload.LaunchAttemptID,
	configurationVersionID uint64,
) (ClaimResult, error) {
	if id == "" || attemptID == "" || configurationVersionID == 0 {
		return ClaimResult{}, ErrInvalidIdentity
	}
	agg, err := s.getAggregate(id)
	if err != nil {
		return ClaimResult{}, err
	}
	agg.mu.Lock()
	defer agg.mu.Unlock()

	// 1. Revision check.
	if agg.revision != expectedRevision {
		return ClaimResult{}, ErrStaleRevision
	}
	// 2. Lifecycle state check.
	// A new claim is permitted only when no active attempt is present.
	// Actual state may be Stopped (initial or after clean stop) or Failed
	// (after a definitive failure with no Host resources). Both are valid
	// starting conditions for a new attempt; only an active attempt blocks.
	if agg.hasActiveAttempt {
		return ClaimResult{}, ErrActiveAttemptExists
	}
	// If actual state is Stopping, a prior stop claim is still unresolved —
	// block until the active attempt reference is cleared by a terminal
	// publication.
	if agg.actual == ActualStateStopping || agg.actual == ActualStateClaimed ||
		agg.actual == ActualStateRunning {
		return ClaimResult{}, ErrActiveAttemptExists
	}
	// 3. History uniqueness check.
	if _, reused := agg.attemptIndex[attemptID]; reused {
		return ClaimResult{}, ErrAttemptIDReused
	}

	// 4. Atomic append + activate + publish.
	record := LaunchAttemptRecord{
		runtimeInstanceID:      id,
		launchAttemptID:        attemptID,
		configurationVersionID: configurationVersionID,
		phase:                  AttemptPhaseClaimed,
	}
	idx := len(agg.attempts)
	agg.attempts = append(agg.attempts, record)
	agg.attemptIndex[attemptID] = idx

	agg.activeAttemptID = attemptID
	agg.hasActiveAttempt = true
	agg.desired = DesiredStateStarted
	agg.actual = ActualStateClaimed

	newRevision := agg.advanceRevision()
	return ClaimResult{revision: newRevision, committed: true}, nil
}

// ─── §21 Operation 6: ConditionalBindExecutionGeneration ─────────────────────

// ConditionalBindExecutionGeneration conditionally stores an immutable
// execution-generation binding for the exact active non-terminal attempt.
//
// The operation:
//   - validates expected revision (stale → ErrStaleRevision, zero mutation);
//   - validates that an active attempt exists (→ ErrNoActiveAttempt);
//   - validates the active attempt is in Claimed phase (→ ErrInvalidAttemptPhase);
//   - if the same generation is already bound, returns success with the current
//     revision (zero-mutation satisfied observation, DP-014 §10);
//   - if a different generation is already bound, returns ErrBindingAlreadyExists
//     (zero mutation);
//   - otherwise stores the binding and advances revision.
//
// A rejected bind performs zero mutation (DP-014 §10, §22 proofs 6, 7, 9).
func (s *Store) ConditionalBindExecutionGeneration(
	id runtimeconfigload.RuntimeInstanceID,
	expectedRevision Revision,
	attemptID runtimeconfigload.LaunchAttemptID,
	generation ExecutionGeneration,
) (PublishResult, error) {
	if id == "" || attemptID == "" || generation == "" {
		return PublishResult{}, ErrInvalidIdentity
	}
	agg, err := s.getAggregate(id)
	if err != nil {
		return PublishResult{}, err
	}
	agg.mu.Lock()
	defer agg.mu.Unlock()

	// Revision check.
	if agg.revision != expectedRevision {
		return PublishResult{}, ErrStaleRevision
	}
	// Active attempt check.
	active := agg.activeAttemptPtr()
	if active == nil {
		return PublishResult{}, ErrNoActiveAttempt
	}
	if active.launchAttemptID != attemptID {
		return PublishResult{}, ErrNoActiveAttempt
	}
	// Phase check: binding is only allowed in Claimed phase.
	if active.phase != AttemptPhaseClaimed {
		return PublishResult{}, ErrInvalidAttemptPhase
	}
	// Same-generation idempotent satisfied observation.
	if active.executionGeneration == generation {
		return PublishResult{revision: agg.revision, committed: true}, nil
	}
	// Different generation already bound.
	if active.executionGeneration != "" {
		return PublishResult{}, ErrBindingAlreadyExists
	}
	// Store immutable binding and advance revision.
	active.executionGeneration = generation
	newRevision := agg.advanceRevision()
	return PublishResult{revision: newRevision, committed: true}, nil
}

// ─── §21 Operation 7: ConditionalPublishRunning ───────────────────────────────

// ConditionalPublishRunning atomically validates and publishes the Running fact
// for the exact active attempt after the Owner confirms Host startup and
// readiness.
//
// The operation:
//   - validates expected revision (stale → ErrStaleRevision, zero mutation);
//   - validates that an active attempt in Claimed or Launching phase exists;
//   - publishes actual Running, advances attempt phase to Launching, and
//     advances aggregate revision.
//
// A rejected publication performs zero mutation (DP-014 §11, §22 proofs 8, 9).
func (s *Store) ConditionalPublishRunning(
	id runtimeconfigload.RuntimeInstanceID,
	expectedRevision Revision,
	attemptID runtimeconfigload.LaunchAttemptID,
) (PublishResult, error) {
	if id == "" || attemptID == "" {
		return PublishResult{}, ErrInvalidIdentity
	}
	agg, err := s.getAggregate(id)
	if err != nil {
		return PublishResult{}, err
	}
	agg.mu.Lock()
	defer agg.mu.Unlock()

	// Revision check.
	if agg.revision != expectedRevision {
		return PublishResult{}, ErrStaleRevision
	}
	// Active attempt check.
	active := agg.activeAttemptPtr()
	if active == nil {
		return PublishResult{}, ErrNoActiveAttempt
	}
	if active.launchAttemptID != attemptID {
		return PublishResult{}, ErrNoActiveAttempt
	}
	// Phase check: must be Claimed or Launching.
	if active.phase != AttemptPhaseClaimed && active.phase != AttemptPhaseLaunching {
		return PublishResult{}, ErrInvalidAttemptPhase
	}
	// Publish Running.
	active.phase = AttemptPhaseRunning
	agg.actual = ActualStateRunning
	newRevision := agg.advanceRevision()
	return PublishResult{revision: newRevision, committed: true}, nil
}

// ─── §21 Operation 8: ConditionalClaimStop ────────────────────────────────────

// ConditionalClaimStop atomically claims Stop for the exact active attempt.
//
// Phase-sensitive rules (DP-014 §12):
//   - If the active attempt is in Claimed phase (no Host resources), this
//     atomically publishes desired Stopped, actual Stopped, phase Stopped, and
//     clears the active attempt reference (stopped-before-running).
//   - If the active attempt is in Launching or Running phase, this publishes
//     desired Stopped, actual Stopping, and phase Stopping.
//
// A rejected claim performs zero mutation (DP-014 §12, §22 proofs 8, 9).
func (s *Store) ConditionalClaimStop(
	id runtimeconfigload.RuntimeInstanceID,
	expectedRevision Revision,
	attemptID runtimeconfigload.LaunchAttemptID,
) (PublishResult, error) {
	if id == "" || attemptID == "" {
		return PublishResult{}, ErrInvalidIdentity
	}
	agg, err := s.getAggregate(id)
	if err != nil {
		return PublishResult{}, err
	}
	agg.mu.Lock()
	defer agg.mu.Unlock()

	// Revision check.
	if agg.revision != expectedRevision {
		return PublishResult{}, ErrStaleRevision
	}
	// Active attempt check.
	active := agg.activeAttemptPtr()
	if active == nil {
		return PublishResult{}, ErrNoActiveAttempt
	}
	if active.launchAttemptID != attemptID {
		return PublishResult{}, ErrNoActiveAttempt
	}
	// Phase-sensitive stop claim.
	switch active.phase {
	case AttemptPhaseClaimed:
		// Stopped-before-running: no Host resources were ever owned.
		active.phase = AttemptPhaseStopped
		agg.desired = DesiredStateStopped
		agg.actual = ActualStateStopped
		agg.hasActiveAttempt = false
		agg.activeAttemptID = ""
	case AttemptPhaseLaunching, AttemptPhaseRunning:
		// Transfer Stop responsibility; actual becomes Stopping.
		active.phase = AttemptPhaseStopping
		agg.desired = DesiredStateStopped
		agg.actual = ActualStateStopping
	default:
		return PublishResult{}, ErrInvalidAttemptPhase
	}
	newRevision := agg.advanceRevision()
	return PublishResult{revision: newRevision, committed: true}, nil
}

// ─── §21 Operation 9: ConditionalPublishTerminal ──────────────────────────────

// ConditionalPublishTerminal atomically validates and publishes a terminal
// outcome for the exact active attempt.
//
// Phase-sensitive rules (DP-014 §12):
//   - AttemptPhaseStopped: attempt is already in terminal Stopped; re-entrant
//     publication of the same terminal fact is treated as a definitive rejection
//     because the active reference has already been cleared.
//   - AttemptPhaseStopping: the attempt may receive a terminal Stopped or Failed
//     outcome. Stopped clears the active reference (Host resources proven
//     absent). Failed retains the active reference (AttemptStopping with
//     stop-failure / cleanup-unproven).
//   - Other non-terminal phases: the attempt may receive Failed (definitive
//     failure with no Host resources).
//
// terminalStopped=true publishes Stopped; terminalStopped=false publishes Failed.
//
// A rejected publication performs zero mutation (DP-014 §12, §14, §22 proofs
// 8, 9, 13).
func (s *Store) ConditionalPublishTerminal(
	id runtimeconfigload.RuntimeInstanceID,
	expectedRevision Revision,
	attemptID runtimeconfigload.LaunchAttemptID,
	terminalStopped bool,
) (PublishResult, error) {
	if id == "" || attemptID == "" {
		return PublishResult{}, ErrInvalidIdentity
	}
	agg, err := s.getAggregate(id)
	if err != nil {
		return PublishResult{}, err
	}
	agg.mu.Lock()
	defer agg.mu.Unlock()

	// Revision check.
	if agg.revision != expectedRevision {
		return PublishResult{}, ErrStaleRevision
	}
	// Active attempt check.
	active := agg.activeAttemptPtr()
	if active == nil {
		return PublishResult{}, ErrNoActiveAttempt
	}
	if active.launchAttemptID != attemptID {
		return PublishResult{}, ErrNoActiveAttempt
	}

	if terminalStopped {
		// Terminal Stopped: Host resources are proven absent.
		// Allowed from Stopping (after Stop claim) or Claimed (stopped-before-running
		// was not yet published via ConditionalClaimStop).
		switch active.phase {
		case AttemptPhaseStopping, AttemptPhaseClaimed, AttemptPhaseLaunching, AttemptPhaseRunning:
		default:
			return PublishResult{}, ErrInvalidAttemptPhase
		}
		active.phase = AttemptPhaseStopped
		agg.actual = ActualStateStopped
		agg.desired = DesiredStateStopped
		agg.hasActiveAttempt = false
		agg.activeAttemptID = ""
	} else {
		// Terminal Failed: definitive failure; proven no Host resources remain.
		// Allowed from any non-terminal phase.
		if active.phase.isTerminal() {
			return PublishResult{}, ErrInvalidAttemptPhase
		}
		switch active.phase {
		case AttemptPhaseStopping:
			// Stop-failure / cleanup-unproven: retain active association per
			// DP-014 §12 (AttemptStopping with stop-failure fact).
			active.phase = AttemptPhaseStopping
			agg.actual = ActualStateFailed
		default:
			// Definitive failure with no Host resources.
			active.phase = AttemptPhaseFailed
			agg.actual = ActualStateFailed
			agg.desired = DesiredStateStopped
			agg.hasActiveAttempt = false
			agg.activeAttemptID = ""
		}
	}
	newRevision := agg.advanceRevision()
	return PublishResult{revision: newRevision, committed: true}, nil
}

// ─── internal helpers ─────────────────────────────────────────────────────────

// getAggregate looks up a committed aggregate. It is safe to call without the
// store mutex only when used in a read pattern that does not require the result
// to be stable under concurrent CreateRuntimeInstance calls — for that reason
// the store mutex is held only during map access.
func (s *Store) getAggregate(
	id runtimeconfigload.RuntimeInstanceID,
) (*runtimeInstanceAggregate, error) {
	s.mu.Lock()
	agg, exists := s.instances[id]
	s.mu.Unlock()
	if !exists {
		return nil, ErrInstanceNotFound
	}
	return agg, nil
}
