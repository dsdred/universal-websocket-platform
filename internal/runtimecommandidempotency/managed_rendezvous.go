package runtimecommandidempotency

import (
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

// ManagedStartGate is the closed command-owned result of managed Start
// rendezvous resolution.
type ManagedStartGate uint8

const (
	// GateClear permits the continuation to advance to the next protocol step.
	GateClear ManagedStartGate = iota + 1
	// GateStopConverged reports exact convergence by the admitted Stop.
	GateStopConverged
	// GateBlocked reports stale, foreign, lost, or otherwise unproven authority.
	GateBlocked
)

// ManagedStartFinalDisposition is the only disposition a continuation may
// submit after the identity revision sandwich.
type ManagedStartFinalDisposition uint8

const (
	// FinalContinue seals successful exact identity binding.
	FinalContinue ManagedStartFinalDisposition = iota + 1
	// FinalBindingFailed seals a definitive coherent binding failure.
	FinalBindingFailed
)

type managedStartStage uint8

const (
	managedStagePreOwner managedStartStage = iota
	managedStageBinding
	managedStageContinue
	managedStageBindingFailed
	managedStageNoClaim
	managedStageBlocked
)

type managedStartRendezvous struct {
	binding    runtimeorchestrationbinding.StartExecutionBinding
	identity   commandIdentity
	phase      phaseIdentity
	state      *permitState
	generation uint64
	bridge     *startRendezvous
	stage      managedStartStage
}

func (r *managedStartRendezvous) blockLocked() {
	if r == nil {
		return
	}
	// Continue is already sealed. A later Stop is ordinary tracked work and
	// retains its own permit on Boundary.Execute even if the Start callback
	// returns while that Stop callback is still running.
	if r.stage == managedStageContinue {
		return
	}
	r.stage = managedStageBlocked
	if r.bridge != nil && r.bridge.resolution == stopResolutionNone {
		r.bridge.signal = startSignalBlocked
		r.bridge.notifyLocked()
	}
}

func (b *Boundary) pendingManagedStopRendezvousLocked(
	ledger *commandLedger,
	operation Operation,
) (*startRendezvous, bool) {
	if operation != OperationStop {
		return nil, false
	}
	for _, managed := range ledger.managedStart {
		if managed == nil || managed.generation != b.generation || !managed.liveLocked(b, ledger) {
			continue
		}
		switch managed.stage {
		case managedStagePreOwner, managedStageBinding:
			if managed.bridge == nil || managed.bridge.stopConsumed {
				return nil, true
			}
			return managed.bridge, false
		case managedStageContinue:
			if managed.bridge == nil || managed.bridge.stopConsumed {
				return nil, true
			}
			return managed.bridge, false
		default:
			return nil, true
		}
	}
	return nil, false
}

func (r *managedStartRendezvous) liveLocked(b *Boundary, ledger *commandLedger) bool {
	if r == nil || b == nil || ledger == nil || r.state == nil ||
		b.storage.generation != b.generation || r.generation != b.generation {
		return false
	}
	if r.phase != (phaseIdentity{}) {
		record := ledger.phases[r.phase]
		return ledger.livePhases[r.phase] == r.state && record != nil &&
			record.state == CommandStateClaimed && record.revision == r.state.revision
	}
	record := ledger.records[r.identity]
	return ledger.live[r.identity] == r.state && record != nil &&
		record.state == CommandStateClaimed && record.revision == r.state.revision
}

func (r *managedStartRendezvous) terminalCompatible(outcome TerminalOutcome) bool {
	if r == nil || r.bridge == nil || !outcome.valid() {
		return false
	}
	switch r.stage {
	case managedStageContinue:
		return true
	case managedStageBindingFailed:
		return outcome.category == OutcomeFailed
	case managedStageNoClaim:
		if outcome.launchAttemptID != "" {
			return false
		}
		if r.bridge.startNoClaimCause == StartNoClaimFailed {
			return outcome.category == OutcomeFailed
		}
		return (r.bridge.startNoClaimCause == StartNoClaimCancelled ||
			r.bridge.startNoClaimCause == StartNoClaimRejected) && outcome.category == OutcomeRejected
	case managedStageBinding:
		return r.bridge.resolution == stopResolutionConverged &&
			outcome.category == OutcomeSucceeded && outcome.launchAttemptID == r.bridge.attempt
	default:
		return false
	}
}

func (b *Boundary) lockManagedRendezvous(
	binding runtimeorchestrationbinding.StartExecutionBinding,
) (*commandLedger, *managedStartRendezvous, error) {
	if b == nil || b.storage == nil || !binding.Valid() {
		return nil, nil, ErrInvalidSubmission
	}
	b.storage.clientMu.RLock()
	if b.storage.generation != b.generation {
		b.storage.clientMu.RUnlock()
		return nil, nil, ErrBoundaryExpired
	}
	authorization := binding.Authorization()
	ledger := b.storage.existingLedger(instanceScope{
		domain: authorization.OperationalDomain(), workspaceID: authorization.WorkspaceID(),
		configurationID: authorization.ConfigurationID(), runtimeInstanceID: authorization.RuntimeInstanceID(),
	})
	if ledger == nil {
		b.storage.clientMu.RUnlock()
		return nil, nil, ErrIndeterminateExecution
	}
	ledger.mu.Lock()
	rendezvous := ledger.managedStart[binding.Rendezvous()]
	if rendezvous == nil || rendezvous.binding != binding || !rendezvous.liveLocked(b, ledger) {
		ledger.mu.Unlock()
		b.storage.clientMu.RUnlock()
		return nil, nil, ErrIndeterminateExecution
	}
	return ledger, rendezvous, nil
}

func (b *Boundary) unlockManagedRendezvous(ledger *commandLedger) {
	ledger.mu.Unlock()
	b.storage.clientMu.RUnlock()
}

// ResolveManagedStartEarly linearizes the exact Owner claim against a pending
// command Stop and advances the live rendezvous into the binding gate.
func (b *Boundary) ResolveManagedStartEarly(
	binding runtimeorchestrationbinding.StartExecutionBinding,
	attempt runtimeconfigload.LaunchAttemptID,
) (ManagedStartGate, error) {
	if attempt == "" {
		return GateBlocked, ErrInvalidSubmission
	}
	ledger, rendezvous, err := b.lockManagedRendezvous(binding)
	if err != nil {
		return GateBlocked, err
	}
	if rendezvous.stage != managedStagePreOwner || rendezvous.bridge == nil {
		b.unlockManagedRendezvous(ledger)
		return GateBlocked, ErrIndeterminateExecution
	}
	rendezvous.stage = managedStageBinding
	rendezvous.bridge.attempt = attempt
	if rendezvous.bridge.stopState != nil {
		rendezvous.bridge.signal = startSignalOwnerClaimed
		rendezvous.bridge.notifyLocked()
	}
	b.unlockManagedRendezvous(ledger)
	return b.waitManagedStop(rendezvous)
}

// SignalManagedStartNoClaim definitively closes a pre-Owner rendezvous.
func (b *Boundary) SignalManagedStartNoClaim(
	binding runtimeorchestrationbinding.StartExecutionBinding,
	cause runtimeorchestrationbinding.StartNoClaimCause,
) error {
	if !cause.Valid() {
		return ErrInvalidSubmission
	}
	ledger, rendezvous, err := b.lockManagedRendezvous(binding)
	if err != nil {
		return err
	}
	if rendezvous.stage != managedStagePreOwner || rendezvous.bridge == nil {
		b.unlockManagedRendezvous(ledger)
		return ErrIndeterminateExecution
	}
	rendezvous.stage = managedStageNoClaim
	rendezvous.bridge.signal = startSignalNoClaim
	rendezvous.bridge.startNoClaimCause = cause
	rendezvous.bridge.notifyLocked()
	b.unlockManagedRendezvous(ledger)
	gate, waitErr := b.waitManagedStop(rendezvous)
	if waitErr != nil || gate == GateBlocked {
		return waitErr
	}
	return nil
}

// ResolveManagedStartFinal seals the identity-binding disposition against any
// Stop admitted while the binding sequence was in progress.
func (b *Boundary) ResolveManagedStartFinal(
	binding runtimeorchestrationbinding.StartExecutionBinding,
	disposition ManagedStartFinalDisposition,
) (ManagedStartGate, error) {
	if disposition != FinalContinue && disposition != FinalBindingFailed {
		return GateBlocked, ErrInvalidSubmission
	}
	ledger, rendezvous, err := b.lockManagedRendezvous(binding)
	if err != nil {
		return GateBlocked, err
	}
	if rendezvous.stage != managedStageBinding || rendezvous.bridge == nil {
		b.unlockManagedRendezvous(ledger)
		return GateBlocked, ErrIndeterminateExecution
	}
	if rendezvous.bridge.stopState != nil && rendezvous.bridge.signal == startSignalNone {
		rendezvous.bridge.signal = startSignalOwnerClaimed
		rendezvous.bridge.notifyLocked()
	}
	b.unlockManagedRendezvous(ledger)
	for {
		gate, waitErr := b.waitManagedStop(rendezvous)
		if waitErr != nil || gate == GateStopConverged {
			return gate, waitErr
		}
		lockedLedger, lockedRendezvous, lockErr := b.lockManagedRendezvous(binding)
		if lockErr != nil {
			return GateBlocked, lockErr
		}
		if lockedRendezvous != rendezvous || rendezvous.stage != managedStageBinding {
			b.unlockManagedRendezvous(lockedLedger)
			return GateBlocked, ErrIndeterminateExecution
		}
		// A Stop may have been admitted after the preceding unlocked absence
		// observation. Signal it before sealing; otherwise it could be stranded
		// behind a disposition that incorrectly won the race.
		if rendezvous.bridge.stopState != nil && rendezvous.bridge.signal == startSignalNone {
			rendezvous.bridge.signal = startSignalOwnerClaimed
			rendezvous.bridge.notifyLocked()
			b.unlockManagedRendezvous(lockedLedger)
			continue
		}
		if disposition == FinalContinue {
			rendezvous.stage = managedStageContinue
			rendezvous.bridge.signal = startSignalOwnerClaimed
			rendezvous.bridge.notifyLocked()
		} else {
			rendezvous.stage = managedStageBindingFailed
		}
		b.unlockManagedRendezvous(lockedLedger)
		return GateClear, nil
	}
}

func (b *Boundary) waitManagedStop(rendezvous *managedStartRendezvous) (ManagedStartGate, error) {
	for {
		if rendezvous == nil || rendezvous.bridge == nil {
			return GateBlocked, ErrIndeterminateExecution
		}
		// The rendezvous belongs to exactly one ledger; locate it through the
		// immutable authorization without retaining a store lock while waiting.
		authorization := rendezvous.binding.Authorization()
		ledger := b.storage.existingLedger(instanceScope{
			domain: authorization.OperationalDomain(), workspaceID: authorization.WorkspaceID(),
			configurationID: authorization.ConfigurationID(), runtimeInstanceID: authorization.RuntimeInstanceID(),
		})
		if ledger == nil {
			return GateBlocked, ErrIndeterminateExecution
		}
		ledger.mu.Lock()
		resolution := rendezvous.bridge.resolution
		hasStop := rendezvous.bridge.stopState != nil
		notify := rendezvous.bridge.notify
		stage := rendezvous.stage
		ledger.mu.Unlock()
		switch resolution {
		case stopResolutionConverged:
			return GateStopConverged, nil
		case stopResolutionSatisfiedNoClaim:
			return GateClear, nil
		case stopResolutionBlocked:
			return GateBlocked, ErrIndeterminateExecution
		}
		if stage == managedStageBlocked {
			return GateBlocked, ErrIndeterminateExecution
		}
		if !hasStop {
			return GateClear, nil
		}
		select {
		case <-b.generationDone:
			return GateBlocked, ErrBoundaryExpired
		case <-notify:
		}
	}
}
