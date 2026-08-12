package runtimecommandidempotency

import (
	"context"
	"sync"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

type startSignal uint8

const (
	startSignalNone startSignal = iota
	startSignalOwnerClaimed
	startSignalNoClaim
	startSignalBlocked
)

type stopResolution uint8

const (
	stopResolutionNone stopResolution = iota
	stopResolutionConverged
	stopResolutionSatisfiedNoClaim
	stopResolutionBlocked
)

// StartNoClaimCause retains the historical command-package name while the
// authoritative neutral value belongs to runtimeorchestrationbinding.
type StartNoClaimCause = runtimeorchestrationbinding.StartNoClaimCause

const (
	// StartNoClaimCancelled records caller cancellation before Owner claim.
	StartNoClaimCancelled = runtimeorchestrationbinding.StartNoClaimCancelled
	// StartNoClaimRejected records a definitive no-mutation rejection.
	StartNoClaimRejected = runtimeorchestrationbinding.StartNoClaimRejected
	// StartNoClaimFailed records a definitive pre-claim failure.
	StartNoClaimFailed = runtimeorchestrationbinding.StartNoClaimFailed
)

// startRendezvous is process-local coordination state. Durable ownership stays
// in the existing parent, phase, and primitive command records.
type startRendezvous struct {
	generation uint64
	notify     chan struct{}

	signal  startSignal
	attempt runtimeconfigload.LaunchAttemptID

	stop                     commandIdentity
	stopState                *permitState
	stopConsumed             bool
	stopConverged            bool
	stopFirstWon             bool
	stopCancelledBeforePhase bool
	startPhaseClaimed        bool
	startNoClaimCause        StartNoClaimCause
	resolution               stopResolution
	continueCancelled        bool
}

func newStartRendezvous(generation uint64) *startRendezvous {
	return &startRendezvous{
		generation: generation, notify: make(chan struct{}),
	}
}

func (r *startRendezvous) notifyLocked() {
	close(r.notify)
	r.notify = make(chan struct{})
}

func (b *Boundary) pendingStopRendezvousLocked(
	ledger *commandLedger,
	operation Operation,
) *startRendezvous {
	if operation != OperationStop {
		return nil
	}
	for _, record := range ledger.records {
		if record.state != CommandStateTerminal {
			return nil
		}
	}
	for identity, parent := range ledger.parents {
		if parent.state == CommandStateTerminal {
			continue
		}
		live := ledger.liveParents[identity]
		if live == nil || live.generation != b.generation {
			return nil
		}
		rendezvous := ledger.rendezvous[identity]
		if rendezvous == nil || rendezvous.generation != b.generation ||
			rendezvous.stopConsumed || rendezvous.signal == startSignalBlocked ||
			rendezvous.signal == startSignalNoClaim || rendezvous.continueCancelled {
			return nil
		}
		stopID, _ := newPhaseIdentity(identity, PhaseStopOld)
		if stop := ledger.phases[stopID]; stop != nil && stop.state != CommandStateTerminal {
			return nil
		}
		startID, _ := newPhaseIdentity(identity, PhaseStartTarget)
		if start := ledger.phases[startID]; start != nil && start.state != CommandStateTerminal {
			phaseLive := ledger.livePhases[startID]
			if phaseLive == nil || phaseLive.generation != b.generation {
				return nil
			}
		}
		return rendezvous
	}
	return nil
}

func (p *executionPermit) executePendingStop(
	ctx context.Context,
	invoke func() (TerminalOutcome, error),
) (view RecordView, err error) {
	completed := false
	defer func() {
		if !completed {
			p.blockPendingStop()
		}
	}()

	for {
		p.boundary.storage.clientMu.RLock()
		p.ledger.mu.Lock()
		if !p.pendingStopLiveLocked() {
			p.ledger.mu.Unlock()
			p.boundary.storage.clientMu.RUnlock()
			return RecordView{}, ErrBoundaryExpired
		}
		signal := p.pending.signal
		notify := p.pending.notify
		p.ledger.mu.Unlock()
		p.boundary.storage.clientMu.RUnlock()

		switch signal {
		case startSignalNoClaim:
			view, err = p.publishPendingStop(
				TerminalOutcome{category: OutcomeSucceeded}, stopResolutionSatisfiedNoClaim,
			)
			completed = err == nil
			return view, err
		case startSignalOwnerClaimed:
			outcome, definitive := invokeSafely(invoke)
			if !definitive || !outcome.valid() {
				return RecordView{}, ErrIndeterminateExecution
			}
			resolution := stopResolutionBlocked
			if outcome.category == OutcomeSucceeded &&
				outcome.launchAttemptID == p.pending.attempt {
				resolution = stopResolutionConverged
			}
			view, err = p.publishPendingStop(outcome, resolution)
			completed = err == nil
			return view, err
		case startSignalBlocked:
			return RecordView{}, ErrIndeterminateExecution
		}

		select {
		case <-ctx.Done():
			view, cancelled, cancelErr := p.cancelPendingStop(ctx.Err())
			if cancelErr != nil {
				return view, cancelErr
			}
			if cancelled {
				completed = true
				return view, ctx.Err()
			}
		case <-p.boundary.generationDone:
			return RecordView{}, ErrBoundaryExpired
		case <-notify:
		}
	}
}

func (p *executionPermit) cancelPendingStop(cause error) (RecordView, bool, error) {
	p.boundary.storage.clientMu.RLock()
	defer p.boundary.storage.clientMu.RUnlock()
	p.ledger.mu.Lock()
	defer p.ledger.mu.Unlock()
	if !p.pendingStopLiveLocked() {
		return RecordView{}, false, ErrBoundaryExpired
	}
	if p.pending.signal != startSignalNone {
		return RecordView{}, false, nil
	}
	record := p.ledger.records[p.identity]
	if record == nil || record.state != CommandStateClaimed || cause == nil {
		return RecordView{}, false, ErrIndeterminateExecution
	}
	record.state = CommandStateTerminal
	record.revision++
	record.outcome = TerminalOutcome{category: OutcomeRejected}
	record.hasOutcome = true
	delete(p.ledger.live, p.identity)
	if !p.pending.startPhaseClaimed {
		p.pending.stopCancelledBeforePhase = true
	}
	p.pending.stopState = nil
	p.pending.resolution = stopResolutionNone
	p.pending.notifyLocked()
	return record.view(), true, nil
}

func (p *executionPermit) pendingStopLiveLocked() bool {
	return p.pending != nil && p.boundary.storage.generation == p.boundary.generation &&
		p.pending.generation == p.boundary.generation &&
		p.pending.stop == p.identity && p.pending.stopState == p.state &&
		p.pending.resolution == stopResolutionNone &&
		p.ledger.live[p.identity] == p.state
}

func (p *executionPermit) publishPendingStop(
	outcome TerminalOutcome,
	resolution stopResolution,
) (RecordView, error) {
	p.boundary.storage.clientMu.RLock()
	defer p.boundary.storage.clientMu.RUnlock()
	p.ledger.mu.Lock()
	defer p.ledger.mu.Unlock()
	if !p.pendingStopLiveLocked() {
		return RecordView{}, ErrBoundaryExpired
	}
	record := p.ledger.records[p.identity]
	if record == nil || record.state != CommandStateClaimed || !outcome.valid() {
		return RecordView{}, ErrIndeterminateExecution
	}
	record.state = CommandStateTerminal
	record.revision++
	record.outcome = outcome
	record.hasOutcome = true
	delete(p.ledger.live, p.identity)
	p.pending.resolution = resolution
	if resolution == stopResolutionConverged {
		p.pending.stopConverged = true
	} else if resolution == stopResolutionSatisfiedNoClaim &&
		!p.pending.startPhaseClaimed && !p.pending.continueCancelled {
		p.pending.stopFirstWon = true
	}
	p.pending.notifyLocked()
	return record.view(), nil
}

func (p *executionPermit) blockPendingStop() {
	if p == nil || p.pending == nil || p.ledger == nil {
		return
	}
	p.ledger.mu.Lock()
	defer p.ledger.mu.Unlock()
	if p.pending.stop == p.identity && p.pending.stopState == p.state &&
		p.pending.resolution == stopResolutionNone {
		p.pending.resolution = stopResolutionBlocked
		p.pending.notifyLocked()
	}
}

// StartTargetExecution is a callback-scoped signal capability for one newly
// claimed StartTarget phase. It carries no phase or Stop permit and expires
// when the ContinueOrExecuteStartTarget callback returns.
type StartTargetExecution struct {
	permit   *phasePermit
	mu       sync.Mutex
	active   bool
	signaled bool
}

// OwnerClaimed signals the exact Owner-issued Launch Attempt. It waits for an
// already admitted pending Stop to converge on its original Execute stack.
// The bool is true only when that Stop converged and Start work must not
// continue. A false nil result means no active Stop was pending; a previously
// cancelled Stop still consumes the parent's one distinct-Stop slot.
func (e *StartTargetExecution) OwnerClaimed(
	attempt runtimeconfigload.LaunchAttemptID,
) (bool, error) {
	if e == nil || attempt == "" {
		return false, ErrInvalidSubmission
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.permit == nil || !e.active || e.signaled {
		return false, ErrInvalidSubmission
	}
	e.signaled = true
	return e.permit.signalOwnerClaimed(attempt)
}

// StartNoClaim explicitly and immutably records why the callback is ending
// before any Owner claim. It lets an already admitted Stop terminalize
// satisfied without lifecycle work; that satisfaction is not proof of Stop
// convergence. The later callback outcome must exactly match cause and contain
// no Launch Attempt. Later context cancellation cannot change the recorded
// cause.
func (e *StartTargetExecution) StartNoClaim(cause StartNoClaimCause) error {
	if e == nil || !cause.Valid() {
		return ErrInvalidSubmission
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.permit == nil || !e.active || e.signaled {
		return ErrInvalidSubmission
	}
	e.signaled = true
	return e.permit.signalStartNoClaim(cause)
}

func (e *StartTargetExecution) close() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.active = false
	return e.signaled
}

func (p *phasePermit) signalOwnerClaimed(
	attempt runtimeconfigload.LaunchAttemptID,
) (bool, error) {
	p.parent.boundary.storage.clientMu.RLock()
	p.parent.ledger.mu.Lock()
	if !p.parent.liveLocked() || p.parent.ledger.livePhases[p.identity] != p.state {
		p.parent.ledger.mu.Unlock()
		p.parent.boundary.storage.clientMu.RUnlock()
		return false, ErrBoundaryExpired
	}
	rendezvous := p.parent.ledger.rendezvous[p.identity.parent]
	if rendezvous == nil || rendezvous.signal != startSignalNone {
		p.parent.ledger.mu.Unlock()
		p.parent.boundary.storage.clientMu.RUnlock()
		return false, ErrInvalidSubmission
	}
	rendezvous.signal = startSignalOwnerClaimed
	rendezvous.attempt = attempt
	rendezvous.notifyLocked()
	p.parent.ledger.mu.Unlock()
	p.parent.boundary.storage.clientMu.RUnlock()
	return p.waitForPendingStop(rendezvous)
}

func (p *phasePermit) signalStartNoClaim(cause StartNoClaimCause) error {
	p.parent.boundary.storage.clientMu.RLock()
	p.parent.ledger.mu.Lock()
	if !p.parent.liveLocked() || p.parent.ledger.livePhases[p.identity] != p.state {
		p.parent.ledger.mu.Unlock()
		p.parent.boundary.storage.clientMu.RUnlock()
		return ErrBoundaryExpired
	}
	rendezvous := p.parent.ledger.rendezvous[p.identity.parent]
	if rendezvous == nil || rendezvous.signal != startSignalNone {
		p.parent.ledger.mu.Unlock()
		p.parent.boundary.storage.clientMu.RUnlock()
		return ErrInvalidSubmission
	}
	rendezvous.signal = startSignalNoClaim
	rendezvous.startNoClaimCause = cause
	rendezvous.notifyLocked()
	p.parent.ledger.mu.Unlock()
	p.parent.boundary.storage.clientMu.RUnlock()
	_, err := p.waitForPendingStop(rendezvous)
	return err
}

func (p *phasePermit) waitForPendingStop(rendezvous *startRendezvous) (bool, error) {
	for {
		p.parent.ledger.mu.Lock()
		resolution := rendezvous.resolution
		hasStop := rendezvous.stopState != nil
		notify := rendezvous.notify
		p.parent.ledger.mu.Unlock()
		switch resolution {
		case stopResolutionConverged:
			return true, nil
		case stopResolutionSatisfiedNoClaim:
			return true, nil
		case stopResolutionBlocked:
			return false, ErrIndeterminateExecution
		}
		if !hasStop {
			return false, nil
		}
		select {
		case <-p.parent.boundary.generationDone:
			return false, ErrBoundaryExpired
		case <-notify:
		}
	}
}

func (p *phasePermit) blockStartRendezvous() {
	if p == nil || p.parent == nil || p.parent.ledger == nil {
		return
	}
	p.parent.ledger.mu.Lock()
	defer p.parent.ledger.mu.Unlock()
	rendezvous := p.parent.ledger.rendezvous[p.identity.parent]
	if rendezvous != nil && rendezvous.signal == startSignalNone {
		rendezvous.signal = startSignalBlocked
		rendezvous.notifyLocked()
	}
}

func (p *phasePermit) executeStartTarget(
	invoke func(*StartTargetExecution) (TerminalOutcome, error),
) (view PhaseRecordView, err error) {
	if invoke == nil {
		return PhaseRecordView{}, ErrInvalidSubmission
	}
	completed := false
	defer func() {
		if !completed {
			p.blockStartRendezvous()
		}
		p.expire()
	}()

	p.parent.boundary.storage.clientMu.RLock()
	p.parent.ledger.mu.Lock()
	if !p.parent.liveLocked() || p.parent.ledger.livePhases[p.identity] != p.state {
		p.parent.ledger.mu.Unlock()
		p.parent.boundary.storage.clientMu.RUnlock()
		return PhaseRecordView{}, ErrBoundaryExpired
	}
	p.parent.ledger.mu.Unlock()
	p.parent.boundary.storage.clientMu.RUnlock()

	execution := &StartTargetExecution{permit: p, active: true}
	outcome, definitive := invokeStartTargetSafely(invoke, execution)
	signaled := execution.close()
	if !definitive || !outcome.valid() {
		return PhaseRecordView{}, ErrIndeterminateExecution
	}
	if !signaled {
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
	rendezvous := p.parent.ledger.rendezvous[p.identity.parent]
	if rendezvous == nil {
		return PhaseRecordView{}, ErrIndeterminateExecution
	}
	if rendezvous.signal == startSignalNoClaim {
		compatible := outcome.launchAttemptID == ""
		switch rendezvous.startNoClaimCause {
		case StartNoClaimCancelled, StartNoClaimRejected:
			compatible = compatible && outcome.category == OutcomeRejected
		case StartNoClaimFailed:
			compatible = compatible && outcome.category == OutcomeFailed
		default:
			compatible = false
		}
		if !compatible {
			return PhaseRecordView{}, ErrIndeterminateExecution
		}
	}
	record.state = CommandStateTerminal
	record.revision++
	record.outcome = outcome
	record.hasOutcome = true
	delete(p.parent.ledger.livePhases, p.identity)
	completed = true
	return record.view(), nil
}

func invokeStartTargetSafely(
	invoke func(*StartTargetExecution) (TerminalOutcome, error),
	execution *StartTargetExecution,
) (outcome TerminalOutcome, definitive bool) {
	defer func() {
		execution.close()
		if recover() != nil {
			outcome = TerminalOutcome{}
			definitive = false
		}
	}()
	outcome, err := invoke(execution)
	return outcome, err == nil
}

func (p *ParentExecution) waitForStopFirst(rendezvous *startRendezvous) (bool, error) {
	for {
		p.ledger.mu.Lock()
		resolution := rendezvous.resolution
		notify := rendezvous.notify
		p.ledger.mu.Unlock()
		switch resolution {
		case stopResolutionConverged:
			return true, nil
		case stopResolutionSatisfiedNoClaim:
			return true, nil
		case stopResolutionBlocked:
			return false, ErrIndeterminateExecution
		}
		select {
		case <-p.boundary.generationDone:
			return false, ErrBoundaryExpired
		case <-notify:
		}
	}
}
