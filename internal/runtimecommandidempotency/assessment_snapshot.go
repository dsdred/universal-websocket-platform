package runtimecommandidempotency

import (
	"sort"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

// AssessmentSnapshot is one complete detached read of the primitive, parent,
// and linked-phase command facts for an exact Runtime Instance scope. It
// carries no permit, callback, rendezvous, admission, or mutation capability.
type AssessmentSnapshot struct {
	domain            string
	workspaceID       uint64
	configurationID   uint64
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID
	primitive         []RecordView
	parents           []ParentRecordView
	phases            []PhaseRecordView
}

// Domain returns the exact operational management domain.
func (s AssessmentSnapshot) Domain() string { return s.domain }

// WorkspaceID returns the exact Workspace identity.
func (s AssessmentSnapshot) WorkspaceID() uint64 { return s.workspaceID }

// ConfigurationID returns the exact Configuration identity.
func (s AssessmentSnapshot) ConfigurationID() uint64 { return s.configurationID }

// RuntimeInstanceID returns the exact Runtime Instance identity.
func (s AssessmentSnapshot) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID {
	return s.runtimeInstanceID
}

// PrimitiveRecords returns a detached deterministically ordered copy of all
// primitive command records in the snapshot.
func (s AssessmentSnapshot) PrimitiveRecords() []RecordView {
	return append([]RecordView(nil), s.primitive...)
}

// ParentRecords returns a detached deterministically ordered copy of all
// parent command records in the snapshot.
func (s AssessmentSnapshot) ParentRecords() []ParentRecordView {
	return append([]ParentRecordView(nil), s.parents...)
}

// PhaseRecords returns a detached deterministically ordered copy of all linked
// phase records in the snapshot.
func (s AssessmentSnapshot) PhaseRecords() []PhaseRecordView {
	return append([]PhaseRecordView(nil), s.phases...)
}

// ReadAssessmentSnapshot returns one complete per-Instance command snapshot.
// A valid scope with no ledger returns an empty complete snapshot and does not
// create a ledger. Existing records are copied while holding the ledger's one
// per-Instance lock, then sorted after the lock is released.
func (s *MemoryStorage) ReadAssessmentSnapshot(
	domain string,
	workspaceID uint64,
	configurationID uint64,
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID,
) (AssessmentSnapshot, error) {
	if s == nil || domain == "" || workspaceID == 0 || configurationID == 0 || runtimeInstanceID == "" {
		return AssessmentSnapshot{}, ErrInvalidSubmission
	}

	snapshot := AssessmentSnapshot{
		domain:            domain,
		workspaceID:       workspaceID,
		configurationID:   configurationID,
		runtimeInstanceID: runtimeInstanceID,
		primitive:         []RecordView{},
		parents:           []ParentRecordView{},
		phases:            []PhaseRecordView{},
	}
	ledger := s.existingLedger(instanceScope{
		domain:            domain,
		workspaceID:       workspaceID,
		configurationID:   configurationID,
		runtimeInstanceID: runtimeInstanceID,
	})
	if ledger == nil {
		return snapshot, nil
	}

	ledger.mu.Lock()
	snapshot.primitive = make([]RecordView, 0, len(ledger.records))
	for _, record := range ledger.records {
		snapshot.primitive = append(snapshot.primitive, record.view())
	}
	snapshot.parents = make([]ParentRecordView, 0, len(ledger.parents))
	for _, record := range ledger.parents {
		snapshot.parents = append(snapshot.parents, record.view())
	}
	snapshot.phases = make([]PhaseRecordView, 0, len(ledger.phases))
	for _, record := range ledger.phases {
		snapshot.phases = append(snapshot.phases, record.view())
	}
	ledger.mu.Unlock()

	sort.Slice(snapshot.primitive, func(i, j int) bool {
		return recordViewLess(snapshot.primitive[i], snapshot.primitive[j])
	})
	sort.Slice(snapshot.parents, func(i, j int) bool {
		return parentRecordViewLess(snapshot.parents[i], snapshot.parents[j])
	})
	sort.Slice(snapshot.phases, func(i, j int) bool {
		return phaseRecordViewLess(snapshot.phases[i], snapshot.phases[j])
	})
	return snapshot, nil
}

func recordViewLess(left, right RecordView) bool {
	if left.scope.operation != right.scope.operation {
		return left.scope.operation < right.scope.operation
	}
	return left.key < right.key
}

func parentRecordViewLess(left, right ParentRecordView) bool {
	if left.scope.operation != right.scope.operation {
		return left.scope.operation < right.scope.operation
	}
	return left.key < right.key
}

func phaseRecordViewLess(left, right PhaseRecordView) bool {
	if left.parentScope.operation != right.parentScope.operation {
		return left.parentScope.operation < right.parentScope.operation
	}
	if left.parentKey != right.parentKey {
		return left.parentKey < right.parentKey
	}
	if left.ordinal != right.ordinal {
		return left.ordinal < right.ordinal
	}
	return left.kind < right.kind
}
