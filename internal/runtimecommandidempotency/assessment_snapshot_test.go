package runtimecommandidempotency

import (
	"sync"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

func TestReadAssessmentSnapshotEmptyDoesNotCreateLedger(t *testing.T) {
	storage := NewMemoryStorage()

	first, err := storage.ReadAssessmentSnapshot("domain", 1, 2, "instance")
	if err != nil {
		t.Fatalf("read empty snapshot: %v", err)
	}
	if got := len(storage.ledgers); got != 0 {
		t.Fatalf("empty snapshot created %d ledgers", got)
	}
	if first.Domain() != "domain" || first.WorkspaceID() != 1 ||
		first.ConfigurationID() != 2 || first.RuntimeInstanceID() != "instance" {
		t.Fatalf("unexpected snapshot scope: %#v", first)
	}
	if len(first.PrimitiveRecords()) != 0 || len(first.ParentRecords()) != 0 ||
		len(first.PhaseRecords()) != 0 {
		t.Fatalf("empty snapshot contains records")
	}

	second, err := storage.ReadAssessmentSnapshot("domain", 1, 2, "instance")
	if err != nil {
		t.Fatalf("repeat empty snapshot: %v", err)
	}
	if got := len(storage.ledgers); got != 0 {
		t.Fatalf("repeat empty snapshot created %d ledgers", got)
	}
	if !assessmentSnapshotsEqual(first, second) {
		t.Fatal("repeat empty snapshot changed")
	}
}

func TestReadAssessmentSnapshotIsCompleteDetachedAndDeterministic(t *testing.T) {
	storage := NewMemoryStorage()
	instance := runtimeconfigload.RuntimeInstanceID("instance")
	primitiveStartScope := mustAssessmentScope(t, "domain", 1, 2, instance, OperationStart)
	primitiveStopScope := mustAssessmentScope(t, "domain", 1, 2, instance, OperationStop)
	parentScope := mustAssessmentScope(t, "domain", 1, 2, instance, OperationReplace)
	startIntent, _ := NewStartIntent(11)
	parentIntent, _ := NewReplaceIntent(12)
	primitiveStartID := commandIdentity{scope: primitiveStartScope, key: "z-start"}
	primitiveStopID := commandIdentity{scope: primitiveStopScope, key: "a-stop"}
	parentID := commandIdentity{scope: parentScope, key: "parent"}
	stopPhaseID, _ := newPhaseIdentity(parentID, PhaseStopOld)
	startPhaseID, _ := newPhaseIdentity(parentID, PhaseStartTarget)
	terminal, _ := NewTerminalOutcome(OutcomeSucceeded, "attempt")
	parentTerminal, _ := NewParentTerminalOutcome(ParentOutcomeSucceeded)
	ledger := storage.ledger(primitiveStartScope.instanceScope())
	ledger.mu.Lock()
	ledger.records[primitiveStartID] = &commandRecord{
		identity: primitiveStartID, intent: startIntent, state: CommandStateClaimed, revision: 1,
	}
	ledger.records[primitiveStopID] = &commandRecord{
		identity: primitiveStopID, intent: NewStopIntent(), state: CommandStateTerminal,
		revision: 2, outcome: terminal, hasOutcome: true,
	}
	ledger.parents[parentID] = &parentRecord{
		identity: parentID, intent: parentIntent, state: CommandStateTerminal,
		revision: 3, outcome: parentTerminal, hasOutcome: true,
	}
	ledger.phases[startPhaseID] = &phaseRecord{
		identity: startPhaseID, state: CommandStateTerminal, revision: 4,
		outcome: terminal, hasOutcome: true,
	}
	ledger.phases[stopPhaseID] = &phaseRecord{
		identity: stopPhaseID, state: CommandStateTerminal, revision: 5,
		outcome: terminal, hasOutcome: true,
	}
	ledger.mu.Unlock()

	first, err := storage.ReadAssessmentSnapshot("domain", 1, 2, instance)
	if err != nil {
		t.Fatalf("read populated snapshot: %v", err)
	}
	second, err := storage.ReadAssessmentSnapshot("domain", 1, 2, instance)
	if err != nil {
		t.Fatalf("repeat populated snapshot: %v", err)
	}
	if !assessmentSnapshotsEqual(first, second) {
		t.Fatal("snapshot ordering is not deterministic")
	}
	if got := first.PrimitiveRecords(); len(got) != 2 ||
		got[0].Scope().Operation() != OperationStart || got[0].Key() != "z-start" ||
		got[1].Scope().Operation() != OperationStop || got[1].Key() != "a-stop" {
		t.Fatalf("unexpected primitive order: %#v", got)
	}
	if got := first.ParentRecords(); len(got) != 1 || got[0].Key() != "parent" {
		t.Fatalf("unexpected parents: %#v", got)
	}
	if got := first.PhaseRecords(); len(got) != 2 ||
		got[0].Kind() != PhaseStopOld || got[1].Kind() != PhaseStartTarget {
		t.Fatalf("unexpected phase order: %#v", got)
	}

	primitives := first.PrimitiveRecords()
	parents := first.ParentRecords()
	phases := first.PhaseRecords()
	primitives[0] = RecordView{}
	parents[0] = ParentRecordView{}
	phases[0] = PhaseRecordView{}
	if !assessmentSnapshotsEqual(first, second) {
		t.Fatal("caller slice mutation changed detached snapshot")
	}

	ledger.mu.Lock()
	ledger.records[primitiveStartID].revision = 9
	ledger.parents[parentID].revision = 9
	ledger.phases[startPhaseID].revision = 9
	ledger.mu.Unlock()
	if got := first.PrimitiveRecords()[0].Revision(); got != 1 {
		t.Fatalf("primitive snapshot aliased ledger: revision %d", got)
	}
	if got := first.ParentRecords()[0].Revision(); got != 3 {
		t.Fatalf("parent snapshot aliased ledger: revision %d", got)
	}
	if got := first.PhaseRecords()[1].Revision(); got != 4 {
		t.Fatalf("phase snapshot aliased ledger: revision %d", got)
	}
}

func TestReadAssessmentSnapshotSharesOneLedgerCoherenceBoundary(t *testing.T) {
	storage := NewMemoryStorage()
	primitiveScope := mustAssessmentScope(t, "domain", 1, 2, "instance", OperationStart)
	parentScope := mustAssessmentScope(t, "domain", 1, 2, "instance", OperationReplace)
	primitiveID := commandIdentity{scope: primitiveScope, key: "primitive"}
	parentID := commandIdentity{scope: parentScope, key: "parent"}
	phaseID, _ := newPhaseIdentity(parentID, PhaseStartTarget)
	primitiveIntent, _ := NewStartIntent(10)
	parentIntent, _ := NewReplaceIntent(11)
	ledger := storage.ledger(primitiveScope.instanceScope())
	ledger.mu.Lock()
	ledger.records[primitiveID] = &commandRecord{
		identity: primitiveID, intent: primitiveIntent, state: CommandStateClaimed, revision: 1,
	}
	ledger.parents[parentID] = &parentRecord{
		identity: parentID, intent: parentIntent, state: CommandStateClaimed, revision: 1,
	}
	ledger.phases[phaseID] = &phaseRecord{
		identity: phaseID, state: CommandStateClaimed, revision: 1,
	}
	ledger.mu.Unlock()

	const (
		readers        = 8
		readsPerReader = 256
		finalRevision  = Revision(2048)
	)
	type observation struct {
		err                      error
		primitive, parent, phase Revision
		complete                 bool
	}
	start := make(chan struct{})
	results := make(chan observation, readers*readsPerReader)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for revision := Revision(2); revision <= finalRevision; revision++ {
			ledger.mu.Lock()
			ledger.records[primitiveID].revision = revision
			ledger.parents[parentID].revision = revision
			ledger.phases[phaseID].revision = revision
			ledger.mu.Unlock()
		}
	}()
	for range readers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range readsPerReader {
				snapshot, err := storage.ReadAssessmentSnapshot("domain", 1, 2, "instance")
				primitive := snapshot.PrimitiveRecords()
				parents := snapshot.ParentRecords()
				phases := snapshot.PhaseRecords()
				result := observation{err: err, complete: len(primitive) == 1 && len(parents) == 1 && len(phases) == 1}
				if result.complete {
					result.primitive = primitive[0].Revision()
					result.parent = parents[0].Revision()
					result.phase = phases[0].Revision()
				}
				results <- result
			}
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	for result := range results {
		if result.err != nil || !result.complete || result.primitive != result.parent || result.parent != result.phase {
			t.Fatalf("mixed concurrent snapshot: %#v", result)
		}
	}
	final, err := storage.ReadAssessmentSnapshot("domain", 1, 2, "instance")
	if err != nil || final.PrimitiveRecords()[0].Revision() != finalRevision ||
		final.ParentRecords()[0].Revision() != finalRevision ||
		final.PhaseRecords()[0].Revision() != finalRevision {
		t.Fatalf("writer did not publish complete final set: snapshot=%#v err=%v", final, err)
	}
}

func mustAssessmentScope(
	t *testing.T,
	domain string,
	workspaceID uint64,
	configurationID uint64,
	instanceID runtimeconfigload.RuntimeInstanceID,
	operation Operation,
) Scope {
	t.Helper()
	scope, err := NewScope(domain, workspaceID, configurationID, instanceID, operation)
	if err != nil {
		t.Fatalf("new scope: %v", err)
	}
	return scope
}

func assessmentSnapshotsEqual(left, right AssessmentSnapshot) bool {
	if left.domain != right.domain || left.workspaceID != right.workspaceID ||
		left.configurationID != right.configurationID ||
		left.runtimeInstanceID != right.runtimeInstanceID ||
		len(left.primitive) != len(right.primitive) ||
		len(left.parents) != len(right.parents) || len(left.phases) != len(right.phases) {
		return false
	}
	for i := range left.primitive {
		if left.primitive[i] != right.primitive[i] {
			return false
		}
	}
	for i := range left.parents {
		if left.parents[i] != right.parents[i] {
			return false
		}
	}
	for i := range left.phases {
		if left.phases[i] != right.phases[i] {
			return false
		}
	}
	return true
}
