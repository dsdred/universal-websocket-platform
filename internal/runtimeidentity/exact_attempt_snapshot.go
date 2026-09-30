package runtimeidentity

import "github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"

// exactAttemptSnapshotReader is the existing DP-014 aggregate/history read
// capability. It is repository-private and grants no mutation authority.
type exactAttemptSnapshotReader interface {
	ReadRuntimeInstance(runtimeconfigload.RuntimeInstanceID) (RuntimeInstanceView, error)
	ReadLaunchAttemptHistory(runtimeconfigload.RuntimeInstanceID) ([]LaunchAttemptRecord, error)
}

// ExactAttemptSnapshot is one detached, revision-coherent observation of an
// exact Launch Attempt and its immutable containment correlation. It is not an
// evidence record and grants no lifecycle, command, or recovery authority.
type ExactAttemptSnapshot struct {
	workspaceID             uint64
	configurationID         uint64
	runtimeInstanceID       runtimeconfigload.RuntimeInstanceID
	launchAttemptID         runtimeconfigload.LaunchAttemptID
	configurationVersionID  uint64
	executionGeneration     ExecutionGeneration
	phase                   AttemptPhase
	terminalCompletionBasis TerminalCompletionBasis
	revision                Revision
}

func (s ExactAttemptSnapshot) WorkspaceID() uint64     { return s.workspaceID }
func (s ExactAttemptSnapshot) ConfigurationID() uint64 { return s.configurationID }
func (s ExactAttemptSnapshot) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID {
	return s.runtimeInstanceID
}
func (s ExactAttemptSnapshot) LaunchAttemptID() runtimeconfigload.LaunchAttemptID {
	return s.launchAttemptID
}
func (s ExactAttemptSnapshot) ConfigurationVersionID() uint64 {
	return s.configurationVersionID
}
func (s ExactAttemptSnapshot) ExecutionGeneration() ExecutionGeneration {
	return s.executionGeneration
}
func (s ExactAttemptSnapshot) Phase() AttemptPhase { return s.phase }
func (s ExactAttemptSnapshot) TerminalCompletionBasis() TerminalCompletionBasis {
	return s.terminalCompletionBasis
}
func (s ExactAttemptSnapshot) Revision() Revision { return s.revision }

// ReadExactAttemptSnapshot composes the existing aggregate and history reads
// as aggregate/history/aggregate. A concurrent aggregate revision change is
// stale; malformed, partial, foreign, or contradictory facts are incoherent.
func ReadExactAttemptSnapshot(
	reader exactAttemptSnapshotReader,
	instanceID runtimeconfigload.RuntimeInstanceID,
	attemptID runtimeconfigload.LaunchAttemptID,
) (ExactAttemptSnapshot, error) {
	if reader == nil || instanceID == "" || attemptID == "" {
		return ExactAttemptSnapshot{}, ErrInvalidIdentity
	}

	before, err := reader.ReadRuntimeInstance(instanceID)
	if err != nil {
		return ExactAttemptSnapshot{}, err
	}
	history, err := reader.ReadLaunchAttemptHistory(instanceID)
	if err != nil {
		return ExactAttemptSnapshot{}, err
	}
	after, err := reader.ReadRuntimeInstance(instanceID)
	if err != nil {
		return ExactAttemptSnapshot{}, err
	}
	if before != after {
		return ExactAttemptSnapshot{}, ErrStaleRevision
	}
	if before.workspaceID == 0 || before.configurationID == 0 ||
		before.runtimeInstanceID != instanceID || before.revision == 0 {
		return ExactAttemptSnapshot{}, ErrIncoherentAttemptSnapshot
	}

	var exact LaunchAttemptRecord
	matches := 0
	seen := make(map[runtimeconfigload.LaunchAttemptID]struct{}, len(history))
	nonterminal := 0
	var nonterminalID runtimeconfigload.LaunchAttemptID
	for _, attempt := range history {
		if attempt.runtimeInstanceID != instanceID || attempt.launchAttemptID == "" ||
			attempt.configurationVersionID == 0 || !coherentAttemptTerminalFacts(attempt) {
			return ExactAttemptSnapshot{}, ErrIncoherentAttemptSnapshot
		}
		if _, duplicate := seen[attempt.launchAttemptID]; duplicate {
			return ExactAttemptSnapshot{}, ErrIncoherentAttemptSnapshot
		}
		seen[attempt.launchAttemptID] = struct{}{}
		if !attempt.phase.isTerminal() {
			nonterminal++
			nonterminalID = attempt.launchAttemptID
		}
		if attempt.launchAttemptID == attemptID {
			exact = attempt
			matches++
		}
	}
	if matches == 0 {
		return ExactAttemptSnapshot{}, ErrAttemptNotFound
	}
	if matches != 1 || exact.executionGeneration == "" {
		return ExactAttemptSnapshot{}, ErrIncoherentAttemptSnapshot
	}

	activeID, active := before.ActiveAttempt()
	if (active && (nonterminal != 1 || nonterminalID != activeID)) || (!active && nonterminal != 0) {
		return ExactAttemptSnapshot{}, ErrIncoherentAttemptSnapshot
	}
	if exact.phase.isTerminal() && active && activeID == attemptID {
		return ExactAttemptSnapshot{}, ErrIncoherentAttemptSnapshot
	}
	if !exact.phase.isTerminal() && (!active || activeID != attemptID) {
		return ExactAttemptSnapshot{}, ErrIncoherentAttemptSnapshot
	}

	return ExactAttemptSnapshot{
		workspaceID:             before.workspaceID,
		configurationID:         before.configurationID,
		runtimeInstanceID:       instanceID,
		launchAttemptID:         attemptID,
		configurationVersionID:  exact.configurationVersionID,
		executionGeneration:     exact.executionGeneration,
		phase:                   exact.phase,
		terminalCompletionBasis: exact.terminalCompletionBasis,
		revision:                before.revision,
	}, nil
}

// RevalidateExactAttemptSnapshot performs the fresh exact aggregate-revision
// check required before a caller uses the detached observation. It writes no
// fact and returns ErrStaleRevision on every aggregate change.
func RevalidateExactAttemptSnapshot(
	reader exactAttemptSnapshotReader,
	snapshot ExactAttemptSnapshot,
) error {
	if reader == nil || snapshot.runtimeInstanceID == "" || snapshot.launchAttemptID == "" ||
		snapshot.workspaceID == 0 || snapshot.configurationID == 0 || snapshot.revision == 0 {
		return ErrInvalidIdentity
	}
	current, err := reader.ReadRuntimeInstance(snapshot.runtimeInstanceID)
	if err != nil {
		return err
	}
	if current.revision != snapshot.revision {
		return ErrStaleRevision
	}
	if current.workspaceID != snapshot.workspaceID ||
		current.configurationID != snapshot.configurationID ||
		current.runtimeInstanceID != snapshot.runtimeInstanceID {
		return ErrIncoherentAttemptSnapshot
	}
	return nil
}

func coherentAttemptTerminalFacts(attempt LaunchAttemptRecord) bool {
	if !attempt.phase.valid() {
		return false
	}
	if attempt.phase.isTerminal() {
		if !attempt.terminalCompletionBasis.valid() {
			return false
		}
		switch attempt.terminalCompletionBasis {
		case TerminalCompletionOwnerShutdownCompleted:
			return attempt.phase == AttemptPhaseStopped && attempt.executionGeneration != ""
		case TerminalCompletionRecoveryReconciled:
			return attempt.phase == AttemptPhaseFailed
		default:
			return true
		}
	}
	return attempt.terminalCompletionBasis == ""
}
