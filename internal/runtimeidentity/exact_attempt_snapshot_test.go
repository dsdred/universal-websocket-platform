package runtimeidentity

import (
	"errors"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

type scriptedExactAttemptReader struct {
	views      []RuntimeInstanceView
	history    []LaunchAttemptRecord
	viewCalls  int
	historyErr error
}

func (r *scriptedExactAttemptReader) ReadRuntimeInstance(runtimeconfigload.RuntimeInstanceID) (RuntimeInstanceView, error) {
	if len(r.views) == 0 {
		return RuntimeInstanceView{}, ErrInstanceNotFound
	}
	index := r.viewCalls
	if index >= len(r.views) {
		index = len(r.views) - 1
	}
	r.viewCalls++
	return r.views[index], nil
}

func (r *scriptedExactAttemptReader) ReadLaunchAttemptHistory(runtimeconfigload.RuntimeInstanceID) ([]LaunchAttemptRecord, error) {
	if r.historyErr != nil {
		return nil, r.historyErr
	}
	return append([]LaunchAttemptRecord(nil), r.history...), nil
}

func TestReadExactAttemptSnapshot_OwnerShutdownAndFreshRevalidation(t *testing.T) {
	s := newStoreWithInstance(t, "ri-snapshot", 11, 22)
	claim := mustClaim(t, s, "ri-snapshot", "la-snapshot", 33)
	bound, err := s.ConditionalBindExecutionGeneration("ri-snapshot", claim.Revision(), "la-snapshot", "gen-snapshot")
	if err != nil {
		t.Fatal(err)
	}
	running, _ := s.ConditionalPublishRunning("ri-snapshot", bound.Revision(), "la-snapshot")
	stopping, _ := s.ConditionalClaimStop("ri-snapshot", running.Revision(), "la-snapshot")
	terminal, err := s.OwnerTerminalPublisher().ConditionalPublishShutdownCompleted("ri-snapshot", stopping.Revision(), "la-snapshot")
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := ReadExactAttemptSnapshot(s, "ri-snapshot", "la-snapshot")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.WorkspaceID() != 11 || snapshot.ConfigurationID() != 22 ||
		snapshot.RuntimeInstanceID() != "ri-snapshot" || snapshot.LaunchAttemptID() != "la-snapshot" ||
		snapshot.ConfigurationVersionID() != 33 || snapshot.ExecutionGeneration() != "gen-snapshot" ||
		snapshot.Phase() != AttemptPhaseStopped ||
		snapshot.TerminalCompletionBasis() != TerminalCompletionOwnerShutdownCompleted ||
		snapshot.Revision() != terminal.Revision() {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	if err := RevalidateExactAttemptSnapshot(s, snapshot); err != nil {
		t.Fatalf("fresh revalidation = %v", err)
	}

	if _, err := s.ConditionalClaimLaunchAttempt("ri-snapshot", snapshot.Revision(), "la-later", 34); err != nil {
		t.Fatal(err)
	}
	if err := RevalidateExactAttemptSnapshot(s, snapshot); !errors.Is(err, ErrStaleRevision) {
		t.Fatalf("revalidation after revision change = %v, want ErrStaleRevision", err)
	}
}

func TestReadExactAttemptSnapshot_ConcurrentRevisionChangeFailsStale(t *testing.T) {
	before := exactSnapshotView("ri-race", 7)
	after := before
	after.revision++
	reader := &scriptedExactAttemptReader{
		views:   []RuntimeInstanceView{before, after},
		history: []LaunchAttemptRecord{exactSnapshotAttempt("ri-race", "la-race")},
	}
	if _, err := ReadExactAttemptSnapshot(reader, "ri-race", "la-race"); !errors.Is(err, ErrStaleRevision) {
		t.Fatalf("snapshot across concurrent revision = %v, want ErrStaleRevision", err)
	}
}

func TestReadExactAttemptSnapshot_RejectsMalformedExactFacts(t *testing.T) {
	view := exactSnapshotView("ri-malformed", 9)
	valid := exactSnapshotAttempt("ri-malformed", "la-exact")
	duplicate := valid
	foreign := valid
	foreign.runtimeInstanceID = "ri-foreign"
	foreignUnrelated := valid
	foreignUnrelated.runtimeInstanceID = "ri-foreign"
	foreignUnrelated.launchAttemptID = "la-foreign"
	partial := valid
	partial.executionGeneration = ""
	missingBasis := valid
	missingBasis.terminalCompletionBasis = ""
	nonterminalBasis := valid
	nonterminalBasis.phase = AttemptPhaseRunning
	unknownPhase := valid
	unknownPhase.phase = AttemptPhase("unknown")
	unknownPhase.terminalCompletionBasis = ""

	tests := []struct {
		name    string
		history []LaunchAttemptRecord
		want    error
	}{
		{name: "missing", history: nil, want: ErrAttemptNotFound},
		{name: "duplicate", history: []LaunchAttemptRecord{valid, duplicate}, want: ErrIncoherentAttemptSnapshot},
		{name: "foreign-parent", history: []LaunchAttemptRecord{foreign}, want: ErrIncoherentAttemptSnapshot},
		{name: "foreign-unrelated-history", history: []LaunchAttemptRecord{valid, foreignUnrelated}, want: ErrIncoherentAttemptSnapshot},
		{name: "partial-binding", history: []LaunchAttemptRecord{partial}, want: ErrIncoherentAttemptSnapshot},
		{name: "terminal-without-basis", history: []LaunchAttemptRecord{missingBasis}, want: ErrIncoherentAttemptSnapshot},
		{name: "nonterminal-with-basis", history: []LaunchAttemptRecord{nonterminalBasis}, want: ErrIncoherentAttemptSnapshot},
		{name: "unknown-phase", history: []LaunchAttemptRecord{unknownPhase}, want: ErrIncoherentAttemptSnapshot},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := &scriptedExactAttemptReader{views: []RuntimeInstanceView{view, view}, history: test.history}
			if _, err := ReadExactAttemptSnapshot(reader, "ri-malformed", "la-exact"); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func exactSnapshotView(id runtimeconfigload.RuntimeInstanceID, revision Revision) RuntimeInstanceView {
	return RuntimeInstanceView{
		workspaceID:       11,
		configurationID:   22,
		runtimeInstanceID: id,
		revision:          revision,
		desired:           DesiredStateStopped,
		actual:            ActualStateStopped,
	}
}

func exactSnapshotAttempt(instanceID runtimeconfigload.RuntimeInstanceID, attemptID runtimeconfigload.LaunchAttemptID) LaunchAttemptRecord {
	return LaunchAttemptRecord{
		runtimeInstanceID:       instanceID,
		launchAttemptID:         attemptID,
		configurationVersionID:  33,
		phase:                   AttemptPhaseStopped,
		executionGeneration:     "gen-exact",
		terminalCompletionBasis: TerminalCompletionOwnerShutdownCompleted,
	}
}
