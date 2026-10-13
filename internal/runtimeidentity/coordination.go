package runtimeidentity

import (
	"slices"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecoordination"
)

type coordinationOwner struct{ marker byte }

type storeCoordination struct {
	store      *Store
	owner      *coordinationOwner
	scope      runtimecoordination.Scope
	attachment runtimecoordination.AggregateAttachment
}

// NewCoordinatedStore binds one exact Instance before allocation. Every writer
// remains closed until later complete command/authority/no-claim qualification;
// this slice exposes no opening operation. NewStore retains legacy behavior.
func NewCoordinatedStore(scope runtimecoordination.Scope, participant *runtimecoordination.Participant) (*Store, error) {
	attachment, err := participant.AttachAggregate(scope)
	if err != nil {
		return nil, err
	}
	s := NewStore()
	s.coordination = &storeCoordination{store: s, owner: &coordinationOwner{}, scope: scope, attachment: attachment}
	return s, nil
}

func (s *Store) validateScope(id runtimeconfigload.RuntimeInstanceID) error {
	if s == nil || id == "" {
		return ErrInvalidIdentity
	}
	if s.coordination != nil && s.coordination.store != s {
		return ErrInvalidIdentity
	}
	if s.coordination != nil && string(id) != s.coordination.scope.InstanceID() {
		return runtimecoordination.ErrForeignScope
	}
	return nil
}

func (s *Store) enterTransition(id runtimeconfigload.RuntimeInstanceID) error {
	if err := s.validateScope(id); err != nil {
		return err
	}
	if s.coordination == nil {
		return nil
	}
	return s.coordination.attachment.Transition(s.coordination.scope)
}

func (s *Store) enterObservation(id runtimeconfigload.RuntimeInstanceID) (*runtimecoordination.ObservationRegion, error) {
	if err := s.validateScope(id); err != nil {
		return nil, err
	}
	if s.coordination == nil {
		return nil, nil
	}
	return s.coordination.attachment.Observe(s.coordination.scope)
}

// CoordinationSnapshot is the complete detached DP-014 observation, including
// explicit allocated/absent facts. The opaque owner identity prevents another
// attachment with the same tuple from revalidating it. It conveys no authority.
type CoordinationSnapshot struct {
	scope                        runtimecoordination.Scope
	owner                        *coordinationOwner
	presenceRevision             uint64
	allocated, present, complete bool
	view                         RuntimeInstanceView
	history                      []LaunchAttemptRecord
}

// Scope returns the exact immutable scope of this observation.
func (s CoordinationSnapshot) Scope() runtimecoordination.Scope { return s.scope }

// PresenceRevision returns the allocation/creation concurrency stamp.
func (s CoordinationSnapshot) PresenceRevision() uint64 { return s.presenceRevision }

// Allocated reports whether the candidate identity is registered.
func (s CoordinationSnapshot) Allocated() bool { return s.allocated }

// Present reports whether an aggregate exists; allocation alone is not presence.
func (s CoordinationSnapshot) Present() bool { return s.present }

// View returns the complete aggregate observation, or zero when absent.
func (s CoordinationSnapshot) View() RuntimeInstanceView { return s.view }

// History returns a detached copy of every terminal and nonterminal attempt.
func (s CoordinationSnapshot) History() []LaunchAttemptRecord { return slices.Clone(s.history) }

// CaptureCoordinationSnapshot copies presence, aggregate and full history under
// one ordered short region. It supports coordinated owners only and releases
// all local gates before returning to callers performing external work.
func (s *Store) CaptureCoordinationSnapshot(scope runtimecoordination.Scope) (CoordinationSnapshot, error) {
	if s == nil || s.coordination == nil || s.coordination.store != s || !scope.Valid() {
		return CoordinationSnapshot{}, ErrInvalidIdentity
	}
	if scope != s.coordination.scope {
		return CoordinationSnapshot{}, runtimecoordination.ErrForeignScope
	}
	region, err := s.enterObservation(runtimeconfigload.RuntimeInstanceID(scope.InstanceID()))
	if err != nil {
		return CoordinationSnapshot{}, err
	}
	defer region.Close()
	return s.captureCoordinationLocked()
}

func (s *Store) captureCoordinationLocked() (CoordinationSnapshot, error) {
	c := s.coordination
	id := runtimeconfigload.RuntimeInstanceID(c.scope.InstanceID())
	s.mu.Lock()
	_, allocated := s.allocated[id]
	agg, present := s.instances[id]
	snapshot := CoordinationSnapshot{scope: c.scope, owner: c.owner, presenceRevision: s.presenceRevision,
		allocated: allocated, present: present, complete: true}
	if present && agg == nil {
		s.mu.Unlock()
		return CoordinationSnapshot{}, ErrIncoherentAttemptSnapshot
	}
	if present {
		agg.mu.Lock()
	}
	s.mu.Unlock()
	if !present {
		return snapshot, nil
	}
	defer agg.mu.Unlock()
	snapshot.view = RuntimeInstanceView{owner: c.owner, workspaceID: agg.workspaceID,
		configurationID: agg.configurationID, runtimeInstanceID: agg.runtimeInstanceID,
		revision: agg.revision, desired: agg.desired, actual: agg.actual,
		activeAttemptID: agg.activeAttemptID, hasActiveAttempt: agg.hasActiveAttempt}
	snapshot.history = slices.Clone(agg.attempts)
	if !allocated || agg.workspaceID != c.scope.WorkspaceID() || agg.configurationID != c.scope.ConfigurationID() || agg.runtimeInstanceID != id || agg.revision == 0 {
		return CoordinationSnapshot{}, ErrIncoherentAttemptSnapshot
	}
	if len(agg.attemptIndex) != len(snapshot.history) {
		return CoordinationSnapshot{}, ErrIncoherentAttemptSnapshot
	}
	seen := make(map[runtimeconfigload.LaunchAttemptID]bool, len(snapshot.history))
	nonterminal := 0
	for index, attempt := range snapshot.history {
		if storedIndex, exists := agg.attemptIndex[attempt.launchAttemptID]; !exists || storedIndex != index {
			return CoordinationSnapshot{}, ErrIncoherentAttemptSnapshot
		}
		if attempt.runtimeInstanceID != id || attempt.launchAttemptID == "" || attempt.configurationVersionID == 0 || seen[attempt.launchAttemptID] || !coherentAttemptTerminalFacts(attempt) {
			return CoordinationSnapshot{}, ErrIncoherentAttemptSnapshot
		}
		seen[attempt.launchAttemptID] = true
		if !attempt.phase.isTerminal() {
			nonterminal++
			if !coherentActiveFacts(snapshot.view, attempt) {
				return CoordinationSnapshot{}, ErrIncoherentAttemptSnapshot
			}
			if !agg.hasActiveAttempt || agg.activeAttemptID != attempt.launchAttemptID {
				return CoordinationSnapshot{}, ErrIncoherentAttemptSnapshot
			}
		}
	}
	if (agg.hasActiveAttempt && nonterminal != 1) || (!agg.hasActiveAttempt && (nonterminal != 0 || agg.activeAttemptID != "")) {
		return CoordinationSnapshot{}, ErrIncoherentAttemptSnapshot
	}
	if !agg.hasActiveAttempt {
		actual := ActualStateStopped
		if len(snapshot.history) > 0 && snapshot.history[len(snapshot.history)-1].phase == AttemptPhaseFailed {
			actual = ActualStateFailed
		}
		if agg.desired != DesiredStateStopped || agg.actual != actual {
			return CoordinationSnapshot{}, ErrIncoherentAttemptSnapshot
		}
	}
	return snapshot, nil
}

func coherentActiveFacts(view RuntimeInstanceView, attempt LaunchAttemptRecord) bool {
	switch attempt.phase {
	case AttemptPhaseClaimed, AttemptPhaseLaunching:
		return view.desired == DesiredStateStarted && view.actual == ActualStateClaimed
	case AttemptPhaseRunning:
		return view.desired == DesiredStateStarted && view.actual == ActualStateRunning
	case AttemptPhaseStopping:
		return view.desired == DesiredStateStopped && (view.actual == ActualStateStopping || view.actual == ActualStateFailed) &&
			(attempt.stopClaimedFrom == AttemptPhaseClaimed || attempt.stopClaimedFrom == AttemptPhaseLaunching || attempt.stopClaimedFrom == AttemptPhaseRunning)
	default:
		return false
	}
}

// RevalidateCoordinationSnapshot performs a fresh complete local comparison.
// Stale, foreign, zero and incomplete stamps reject without mutation or permit.
func (s *Store) RevalidateCoordinationSnapshot(snapshot CoordinationSnapshot) error {
	if s == nil || s.coordination == nil || s.coordination.store != s || !snapshot.complete || snapshot.owner == nil || !snapshot.scope.Valid() {
		return ErrInvalidIdentity
	}
	if snapshot.owner != s.coordination.owner || snapshot.scope != s.coordination.scope {
		return runtimecoordination.ErrForeignScope
	}
	current, err := s.CaptureCoordinationSnapshot(snapshot.scope)
	if err != nil {
		return err
	}
	if current.presenceRevision != snapshot.presenceRevision || current.allocated != snapshot.allocated || current.present != snapshot.present || current.view != snapshot.view || !slices.Equal(current.history, snapshot.history) {
		return ErrStaleRevision
	}
	return nil
}

func (s *Store) observationOwner() *coordinationOwner {
	if s.coordination == nil {
		return nil
	}
	return s.coordination.owner
}
