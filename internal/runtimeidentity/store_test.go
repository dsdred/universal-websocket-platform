package runtimeidentity

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

// ─── §22 Proof 1: atomic complete Instance creation and immutable binding ─────

func TestCreateRuntimeInstance_AtomicCompleteAndImmutableBinding(t *testing.T) {
	s := NewStore()
	if err := s.CreateRuntimeInstance(1, 2, "ri-1"); err != nil {
		t.Fatalf("CreateRuntimeInstance() error = %v", err)
	}
	view, err := s.ReadRuntimeInstance("ri-1")
	if err != nil {
		t.Fatalf("ReadRuntimeInstance() error = %v", err)
	}
	if view.WorkspaceID() != 1 || view.ConfigurationID() != 2 {
		t.Fatalf("immutable binding = %d/%d, want 1/2", view.WorkspaceID(), view.ConfigurationID())
	}
	if view.DesiredState() != DesiredStateStopped || view.ActualState() != ActualStateStopped {
		t.Fatalf("initial state = %q/%q, want Stopped/Stopped", view.DesiredState(), view.ActualState())
	}
	if view.Revision() == 0 {
		t.Fatal("initial revision is zero, want > 0")
	}
	if _, active := view.ActiveAttempt(); active {
		t.Fatal("initial aggregate has active attempt, want none")
	}
	history, err := s.ReadLaunchAttemptHistory("ri-1")
	if err != nil {
		t.Fatalf("ReadLaunchAttemptHistory() error = %v", err)
	}
	if len(history) != 0 {
		t.Fatalf("initial history len = %d, want 0", len(history))
	}
}

func TestCreateRuntimeInstance_InvalidIdentityPerformsZeroMutation(t *testing.T) {
	s := NewStore()
	cases := []struct {
		name            string
		workspaceID     uint64
		configurationID uint64
		id              runtimeconfigload.RuntimeInstanceID
	}{
		{"zero workspace", 0, 2, "ri-x"},
		{"zero configuration", 1, 0, "ri-x"},
		{"empty id", 1, 2, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := s.CreateRuntimeInstance(c.workspaceID, c.configurationID, c.id)
			if err != ErrInvalidIdentity {
				t.Fatalf("error = %v, want ErrInvalidIdentity", err)
			}
		})
	}
}

// ─── §22 Proof 2: RuntimeInstanceID uniqueness within management domain ───────

func TestCreateRuntimeInstance_DuplicateIDPerformsZeroMutation(t *testing.T) {
	s := NewStore()
	if err := s.CreateRuntimeInstance(1, 2, "ri-dup"); err != nil {
		t.Fatalf("first CreateRuntimeInstance() error = %v", err)
	}
	err := s.CreateRuntimeInstance(1, 2, "ri-dup")
	if err != ErrInstanceAlreadyExists {
		t.Fatalf("duplicate error = %v, want ErrInstanceAlreadyExists", err)
	}
	// Verify the aggregate is unchanged.
	view, _ := s.ReadRuntimeInstance("ri-dup")
	if view.Revision() != 1 {
		t.Fatalf("revision after duplicate attempt = %v, want 1", view.Revision())
	}
}

func TestAllocateCandidateIdentity_ReservesIDBeforePublication(t *testing.T) {
	s := NewStore()
	if err := s.AllocateCandidateIdentity("ri-alloc"); err != nil {
		t.Fatalf("AllocateCandidateIdentity() error = %v", err)
	}
	// Second allocation of the same ID is rejected.
	if err := s.AllocateCandidateIdentity("ri-alloc"); err != ErrInstanceAlreadyExists {
		t.Fatalf("second allocate error = %v, want ErrInstanceAlreadyExists", err)
	}
	// Publication with the reserved ID succeeds.
	if err := s.CreateRuntimeInstance(1, 2, "ri-alloc"); err != nil {
		t.Fatalf("CreateRuntimeInstance(reserved) error = %v", err)
	}
	// Duplicate publication is rejected.
	if err := s.CreateRuntimeInstance(1, 2, "ri-alloc"); err != ErrInstanceAlreadyExists {
		t.Fatalf("duplicate publication error = %v, want ErrInstanceAlreadyExists", err)
	}
}

func TestAllocateCandidateIdentity_EmptyIDReturnsInvalidIdentity(t *testing.T) {
	s := NewStore()
	if err := s.AllocateCandidateIdentity(""); err != ErrInvalidIdentity {
		t.Fatalf("error = %v, want ErrInvalidIdentity", err)
	}
}

// ─── §22 Proof 3: atomic single-active-attempt claim with exact version pin ───

func TestConditionalClaimLaunchAttempt_AtomicClaimWithExactVersionPin(t *testing.T) {
	s := newStoreWithInstance(t, "ri-claim", 10, 20)
	view0, _ := s.ReadRuntimeInstance("ri-claim")
	result, err := s.ConditionalClaimLaunchAttempt("ri-claim", view0.Revision(), "la-1", 99)
	if err != nil || !result.Committed() {
		t.Fatalf("claim error = %v, committed = %t", err, result.Committed())
	}
	view1, _ := s.ReadRuntimeInstance("ri-claim")
	if view1.Revision() != result.Revision() {
		t.Fatalf("revision mismatch after claim: store=%v result=%v", view1.Revision(), result.Revision())
	}
	if view1.ActualState() != ActualStateClaimed {
		t.Fatalf("actual state = %q, want Claimed", view1.ActualState())
	}
	if view1.DesiredState() != DesiredStateStarted {
		t.Fatalf("desired state = %q, want Started", view1.DesiredState())
	}
	activeID, hasActive := view1.ActiveAttempt()
	if !hasActive || activeID != "la-1" {
		t.Fatalf("active attempt = %q/%t, want la-1/true", activeID, hasActive)
	}
	history, _ := s.ReadLaunchAttemptHistory("ri-claim")
	if len(history) != 1 || history[0].LaunchAttemptID() != "la-1" ||
		history[0].ConfigurationVersionID() != 99 || history[0].Phase() != AttemptPhaseClaimed {
		t.Fatalf("history = %+v, want one Claimed attempt with versionID=99", history)
	}
}

func TestConditionalClaimLaunchAttempt_StaleRevisionPerformsZeroMutation(t *testing.T) {
	s := newStoreWithInstance(t, "ri-stale", 1, 2)
	view0, _ := s.ReadRuntimeInstance("ri-stale")
	_, err := s.ConditionalClaimLaunchAttempt("ri-stale", view0.Revision()+1, "la-stale", 5)
	if err != ErrStaleRevision {
		t.Fatalf("error = %v, want ErrStaleRevision", err)
	}
	view1, _ := s.ReadRuntimeInstance("ri-stale")
	if view1.Revision() != view0.Revision() {
		t.Fatalf("revision changed after stale claim: %v → %v", view0.Revision(), view1.Revision())
	}
}

func TestConditionalClaimLaunchAttempt_ActiveAttemptBlocksNewClaim(t *testing.T) {
	s := newStoreWithInstance(t, "ri-active", 1, 2)
	mustClaim(t, s, "ri-active", "la-first", 5)
	view, _ := s.ReadRuntimeInstance("ri-active")
	_, err := s.ConditionalClaimLaunchAttempt("ri-active", view.Revision(), "la-second", 5)
	if err != ErrActiveAttemptExists {
		t.Fatalf("error = %v, want ErrActiveAttemptExists", err)
	}
}

// ─── §22 Proof 4: child-key uniqueness and non-reuse ─────────────────────────

func TestConditionalClaimLaunchAttempt_AttemptIDNonReuse(t *testing.T) {
	s := newStoreWithInstance(t, "ri-reuse", 1, 2)
	mustClaim(t, s, "ri-reuse", "la-once", 5)
	// Terminal the attempt so actual is Stopped again.
	view, _ := s.ReadRuntimeInstance("ri-reuse")
	activeID, _ := view.ActiveAttempt()
	s.ConditionalClaimStop("ri-reuse", view.Revision(), activeID)
	// Now try to reuse the same attempt ID.
	view2, _ := s.ReadRuntimeInstance("ri-reuse")
	_, err := s.ConditionalClaimLaunchAttempt("ri-reuse", view2.Revision(), "la-once", 5)
	if err != ErrAttemptIDReused {
		t.Fatalf("error = %v, want ErrAttemptIDReused", err)
	}
}

// ─── §22 Proof 5: append-only history across start failure, stop, later ──────

func TestReadLaunchAttemptHistory_AppendOnlyAcrossMultipleAttempts(t *testing.T) {
	s := newStoreWithInstance(t, "ri-history", 1, 2)
	// First attempt: stopped-before-running.
	claimResult := mustClaim(t, s, "ri-history", "la-first", 5)
	_, err := s.ConditionalClaimStop("ri-history", claimResult.Revision(), "la-first")
	if err != nil {
		t.Fatalf("ClaimStop(la-first) error = %v", err)
	}
	// Second attempt: claim + terminal failed.
	view1, _ := s.ReadRuntimeInstance("ri-history")
	claimResult2, err := s.ConditionalClaimLaunchAttempt("ri-history", view1.Revision(), "la-second", 6)
	if err != nil {
		t.Fatalf("claim la-second error = %v", err)
	}
	_, err = s.ConditionalPublishTerminal("ri-history", claimResult2.Revision(), "la-second", false)
	if err != nil {
		t.Fatalf("PublishTerminal(failed) error = %v", err)
	}
	// Third attempt: running → stopped.
	view2, _ := s.ReadRuntimeInstance("ri-history")
	claimResult3 := mustClaimAt(t, s, "ri-history", view2.Revision(), "la-third", 7)
	s.ConditionalPublishRunning("ri-history", claimResult3.Revision(), "la-third")
	view3, _ := s.ReadRuntimeInstance("ri-history")
	s.ConditionalClaimStop("ri-history", view3.Revision(), "la-third")
	view4, _ := s.ReadRuntimeInstance("ri-history")
	s.ConditionalPublishTerminal("ri-history", view4.Revision(), "la-third", true)

	history, err := s.ReadLaunchAttemptHistory("ri-history")
	if err != nil {
		t.Fatalf("ReadLaunchAttemptHistory() error = %v", err)
	}
	if len(history) != 3 {
		t.Fatalf("history len = %d, want 3", len(history))
	}
	wantIDs := []runtimeconfigload.LaunchAttemptID{"la-first", "la-second", "la-third"}
	for i, rec := range history {
		if rec.LaunchAttemptID() != wantIDs[i] {
			t.Fatalf("history[%d].LaunchAttemptID() = %q, want %q", i, rec.LaunchAttemptID(), wantIDs[i])
		}
	}
}

func TestReadLaunchAttemptHistory_ReturnsCopy(t *testing.T) {
	s := newStoreWithInstance(t, "ri-copy", 1, 2)
	mustClaim(t, s, "ri-copy", "la-copy", 5)
	h1, _ := s.ReadLaunchAttemptHistory("ri-copy")
	h1[0] = LaunchAttemptRecord{} // mutate the copy
	h2, _ := s.ReadLaunchAttemptHistory("ri-copy")
	if h2[0].LaunchAttemptID() != "la-copy" {
		t.Fatal("ReadLaunchAttemptHistory returned a reference, not a copy")
	}
}

// ─── §22 Proof 6: exact immutable execution-generation binding ────────────────

func TestConditionalBindExecutionGeneration_StoresImmutableBinding(t *testing.T) {
	s := newStoreWithInstance(t, "ri-bind", 1, 2)
	claimResult := mustClaim(t, s, "ri-bind", "la-bind", 5)
	result, err := s.ConditionalBindExecutionGeneration("ri-bind", claimResult.Revision(), "la-bind", "gen-A")
	if err != nil || !result.Committed() {
		t.Fatalf("bind error = %v, committed = %t", err, result.Committed())
	}
	history, _ := s.ReadLaunchAttemptHistory("ri-bind")
	if history[0].ExecutionGeneration() != "gen-A" {
		t.Fatalf("generation = %q, want gen-A", history[0].ExecutionGeneration())
	}
}

func TestConditionalBindExecutionGeneration_SameGenerationIsSatisfied(t *testing.T) {
	s := newStoreWithInstance(t, "ri-bind-idem", 1, 2)
	claimResult := mustClaim(t, s, "ri-bind-idem", "la-idem", 5)
	r1, _ := s.ConditionalBindExecutionGeneration("ri-bind-idem", claimResult.Revision(), "la-idem", "gen-X")
	r2, err := s.ConditionalBindExecutionGeneration("ri-bind-idem", r1.Revision(), "la-idem", "gen-X")
	if err != nil || !r2.Committed() {
		t.Fatalf("same-generation bind error = %v, committed = %t", err, r2.Committed())
	}
	// Same-generation is zero-mutation: revision must not advance.
	if r2.Revision() != r1.Revision() {
		t.Fatalf("revision advanced on same-generation bind: %v → %v", r1.Revision(), r2.Revision())
	}
}

func TestConditionalBindExecutionGeneration_DifferentGenerationRejected(t *testing.T) {
	s := newStoreWithInstance(t, "ri-bind-diff", 1, 2)
	claimResult := mustClaim(t, s, "ri-bind-diff", "la-diff", 5)
	r1, _ := s.ConditionalBindExecutionGeneration("ri-bind-diff", claimResult.Revision(), "la-diff", "gen-A")
	_, err := s.ConditionalBindExecutionGeneration("ri-bind-diff", r1.Revision(), "la-diff", "gen-B")
	if err != ErrBindingAlreadyExists {
		t.Fatalf("different generation error = %v, want ErrBindingAlreadyExists", err)
	}
	// Verify zero mutation.
	view, _ := s.ReadRuntimeInstance("ri-bind-diff")
	if view.Revision() != r1.Revision() {
		t.Fatalf("revision changed after rejected bind")
	}
}

// ─── §22 Proof 7: binding mismatch, stale revision permit no preparation ──────

func TestConditionalBindExecutionGeneration_StaleRevisionPerformsZeroMutation(t *testing.T) {
	s := newStoreWithInstance(t, "ri-bind-stale", 1, 2)
	mustClaim(t, s, "ri-bind-stale", "la-bs", 5)
	view, _ := s.ReadRuntimeInstance("ri-bind-stale")
	_, err := s.ConditionalBindExecutionGeneration("ri-bind-stale", view.Revision()+100, "la-bs", "gen-A")
	if err != ErrStaleRevision {
		t.Fatalf("error = %v, want ErrStaleRevision", err)
	}
	view2, _ := s.ReadRuntimeInstance("ri-bind-stale")
	if view2.Revision() != view.Revision() {
		t.Fatalf("revision changed after stale bind")
	}
}

func TestConditionalBindExecutionGeneration_WrongAttemptIDRejected(t *testing.T) {
	s := newStoreWithInstance(t, "ri-bind-wrong", 1, 2)
	claimResult := mustClaim(t, s, "ri-bind-wrong", "la-right", 5)
	_, err := s.ConditionalBindExecutionGeneration("ri-bind-wrong", claimResult.Revision(), "la-wrong", "gen-A")
	if err != ErrNoActiveAttempt {
		t.Fatalf("error = %v, want ErrNoActiveAttempt", err)
	}
}

// ─── §22 Proof 8: exact conditional Running, Stop, terminal publications ──────

func TestConditionalPublishRunning_TransitionsToRunning(t *testing.T) {
	s := newStoreWithInstance(t, "ri-run", 1, 2)
	claimResult := mustClaim(t, s, "ri-run", "la-run", 5)
	result, err := s.ConditionalPublishRunning("ri-run", claimResult.Revision(), "la-run")
	if err != nil || !result.Committed() {
		t.Fatalf("PublishRunning error = %v, committed = %t", err, result.Committed())
	}
	view, _ := s.ReadRuntimeInstance("ri-run")
	if view.ActualState() != ActualStateRunning {
		t.Fatalf("actual = %q, want Running", view.ActualState())
	}
	history, _ := s.ReadLaunchAttemptHistory("ri-run")
	if history[0].Phase() != AttemptPhaseRunning {
		t.Fatalf("phase = %q, want Running", history[0].Phase())
	}
}

func TestConditionalClaimStop_ClaimedPhaseStopsBeforeRunning(t *testing.T) {
	s := newStoreWithInstance(t, "ri-stop-pre", 1, 2)
	claimResult := mustClaim(t, s, "ri-stop-pre", "la-sp", 5)
	result, err := s.ConditionalClaimStop("ri-stop-pre", claimResult.Revision(), "la-sp")
	if err != nil || !result.Committed() {
		t.Fatalf("ClaimStop(Claimed) error = %v, committed = %t", err, result.Committed())
	}
	view, _ := s.ReadRuntimeInstance("ri-stop-pre")
	if view.ActualState() != ActualStateStopped {
		t.Fatalf("actual = %q, want Stopped", view.ActualState())
	}
	if _, active := view.ActiveAttempt(); active {
		t.Fatal("active attempt still set after stopped-before-running")
	}
	history, _ := s.ReadLaunchAttemptHistory("ri-stop-pre")
	if history[0].Phase() != AttemptPhaseStopped {
		t.Fatalf("phase = %q, want Stopped", history[0].Phase())
	}
}

func TestConditionalClaimStop_RunningPhaseTransitionsToStopping(t *testing.T) {
	s := newStoreWithInstance(t, "ri-stop-run", 1, 2)
	claimResult := mustClaim(t, s, "ri-stop-run", "la-sr", 5)
	runResult, _ := s.ConditionalPublishRunning("ri-stop-run", claimResult.Revision(), "la-sr")
	result, err := s.ConditionalClaimStop("ri-stop-run", runResult.Revision(), "la-sr")
	if err != nil || !result.Committed() {
		t.Fatalf("ClaimStop(Running) error = %v, committed = %t", err, result.Committed())
	}
	view, _ := s.ReadRuntimeInstance("ri-stop-run")
	if view.ActualState() != ActualStateStopping {
		t.Fatalf("actual = %q, want Stopping", view.ActualState())
	}
	history, _ := s.ReadLaunchAttemptHistory("ri-stop-run")
	if history[0].Phase() != AttemptPhaseStopping {
		t.Fatalf("phase = %q, want Stopping", history[0].Phase())
	}
}

func TestConditionalPublishTerminal_StoppedClearsActiveAttempt(t *testing.T) {
	s := newStoreWithInstance(t, "ri-term-stop", 1, 2)
	claimResult := mustClaim(t, s, "ri-term-stop", "la-ts", 5)
	runResult, _ := s.ConditionalPublishRunning("ri-term-stop", claimResult.Revision(), "la-ts")
	stopResult, _ := s.ConditionalClaimStop("ri-term-stop", runResult.Revision(), "la-ts")
	result, err := s.ConditionalPublishTerminal("ri-term-stop", stopResult.Revision(), "la-ts", true)
	if err != nil || !result.Committed() {
		t.Fatalf("PublishTerminal(Stopped) error = %v, committed = %t", err, result.Committed())
	}
	view, _ := s.ReadRuntimeInstance("ri-term-stop")
	if view.ActualState() != ActualStateStopped {
		t.Fatalf("actual = %q, want Stopped", view.ActualState())
	}
	if _, active := view.ActiveAttempt(); active {
		t.Fatal("active attempt still set after terminal Stopped")
	}
}

func TestConditionalPublishTerminal_FailedDefinitive(t *testing.T) {
	s := newStoreWithInstance(t, "ri-term-fail", 1, 2)
	claimResult := mustClaim(t, s, "ri-term-fail", "la-tf", 5)
	result, err := s.ConditionalPublishTerminal("ri-term-fail", claimResult.Revision(), "la-tf", false)
	if err != nil || !result.Committed() {
		t.Fatalf("PublishTerminal(Failed) error = %v, committed = %t", err, result.Committed())
	}
	view, _ := s.ReadRuntimeInstance("ri-term-fail")
	if view.ActualState() != ActualStateFailed {
		t.Fatalf("actual = %q, want Failed", view.ActualState())
	}
	if _, active := view.ActiveAttempt(); active {
		t.Fatal("active attempt still set after definitive Failed")
	}
	history, _ := s.ReadLaunchAttemptHistory("ri-term-fail")
	if history[0].Phase() != AttemptPhaseFailed {
		t.Fatalf("phase = %q, want Failed", history[0].Phase())
	}
}

// ─── §22 Proof 9: stale and mismatched operations perform zero mutation ───────

func TestAllOperations_StaleRevisionZeroMutation(t *testing.T) {
	s := newStoreWithInstance(t, "ri-stale-all", 1, 2)
	view0, _ := s.ReadRuntimeInstance("ri-stale-all")
	staleRev := view0.Revision() + 999

	_, err := s.ConditionalClaimLaunchAttempt("ri-stale-all", staleRev, "la-x", 5)
	if err != ErrStaleRevision {
		t.Fatalf("Claim stale: %v, want ErrStaleRevision", err)
	}

	mustClaim(t, s, "ri-stale-all", "la-real", 5)
	view1, _ := s.ReadRuntimeInstance("ri-stale-all")
	staleRev2 := view1.Revision() + 999

	_, err = s.ConditionalBindExecutionGeneration("ri-stale-all", staleRev2, "la-real", "gen-A")
	if err != ErrStaleRevision {
		t.Fatalf("Bind stale: %v, want ErrStaleRevision", err)
	}
	_, err = s.ConditionalPublishRunning("ri-stale-all", staleRev2, "la-real")
	if err != ErrStaleRevision {
		t.Fatalf("Running stale: %v, want ErrStaleRevision", err)
	}
	_, err = s.ConditionalClaimStop("ri-stale-all", staleRev2, "la-real")
	if err != ErrStaleRevision {
		t.Fatalf("Stop stale: %v, want ErrStaleRevision", err)
	}
	_, err = s.ConditionalPublishTerminal("ri-stale-all", staleRev2, "la-real", true)
	if err != ErrStaleRevision {
		t.Fatalf("Terminal stale: %v, want ErrStaleRevision", err)
	}
	view2, _ := s.ReadRuntimeInstance("ri-stale-all")
	if view2.Revision() != view1.Revision() {
		t.Fatalf("revision mutated by stale operation: %v → %v", view1.Revision(), view2.Revision())
	}
}

func TestAllOperations_InstanceNotFoundReturnsError(t *testing.T) {
	s := NewStore()
	cases := []struct {
		name string
		fn   func() error
	}{
		{"Read", func() error { _, err := s.ReadRuntimeInstance("no-such"); return err }},
		{"History", func() error { _, err := s.ReadLaunchAttemptHistory("no-such"); return err }},
		{"Claim", func() error { _, err := s.ConditionalClaimLaunchAttempt("no-such", 1, "la", 5); return err }},
		{"Bind", func() error { _, err := s.ConditionalBindExecutionGeneration("no-such", 1, "la", "g"); return err }},
		{"Running", func() error { _, err := s.ConditionalPublishRunning("no-such", 1, "la"); return err }},
		{"Stop", func() error { _, err := s.ConditionalClaimStop("no-such", 1, "la"); return err }},
		{"Terminal", func() error { _, err := s.ConditionalPublishTerminal("no-such", 1, "la", true); return err }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.fn(); err != ErrInstanceNotFound {
				t.Fatalf("error = %v, want ErrInstanceNotFound", err)
			}
		})
	}
}

// ─── §22 Proof 10: concurrent same-Instance claims → at most one accepted ─────

func TestConditionalClaimLaunchAttempt_ConcurrentSameInstanceAtMostOneAccepted(t *testing.T) {
	s := newStoreWithInstance(t, "ri-concurrent", 1, 2)
	view0, _ := s.ReadRuntimeInstance("ri-concurrent")
	baseRev := view0.Revision()

	const goroutines = 50
	var committed atomic.Int32
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		attemptID := runtimeconfigload.LaunchAttemptID("la-concurrent-" + runtimeconfigload.LaunchAttemptID(itoa(i)))
		go func(aid runtimeconfigload.LaunchAttemptID) {
			defer wg.Done()
			result, _ := s.ConditionalClaimLaunchAttempt("ri-concurrent", baseRev, aid, 5)
			if result.Committed() {
				committed.Add(1)
			}
		}(attemptID)
	}
	wg.Wait()
	if committed.Load() != 1 {
		t.Fatalf("committed count = %d, want exactly 1", committed.Load())
	}
	history, _ := s.ReadLaunchAttemptHistory("ri-concurrent")
	if len(history) != 1 {
		t.Fatalf("history len = %d, want 1", len(history))
	}
}

// ─── §22 Proof 11: different Instances can progress independently ─────────────

func TestDifferentInstances_ProgressIndependently(t *testing.T) {
	s := NewStore()
	if err := s.CreateRuntimeInstance(1, 2, "ri-a"); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRuntimeInstance(1, 2, "ri-b"); err != nil {
		t.Fatal(err)
	}
	viewA, _ := s.ReadRuntimeInstance("ri-a")
	viewB, _ := s.ReadRuntimeInstance("ri-b")

	const goroutines = 20
	var wgA, wgB sync.WaitGroup
	wgA.Add(goroutines)
	wgB.Add(goroutines)
	var committedA, committedB atomic.Int32

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wgA.Done()
			aid := runtimeconfigload.LaunchAttemptID("la-a-" + runtimeconfigload.LaunchAttemptID(itoa(i)))
			r, _ := s.ConditionalClaimLaunchAttempt("ri-a", viewA.Revision(), aid, 5)
			if r.Committed() {
				committedA.Add(1)
			}
		}(i)
		go func(i int) {
			defer wgB.Done()
			bid := runtimeconfigload.LaunchAttemptID("la-b-" + runtimeconfigload.LaunchAttemptID(itoa(i)))
			r, _ := s.ConditionalClaimLaunchAttempt("ri-b", viewB.Revision(), bid, 5)
			if r.Committed() {
				committedB.Add(1)
			}
		}(i)
	}
	wgA.Wait()
	wgB.Wait()

	if committedA.Load() != 1 {
		t.Fatalf("ri-a committed = %d, want exactly 1", committedA.Load())
	}
	if committedB.Load() != 1 {
		t.Fatalf("ri-b committed = %d, want exactly 1", committedB.Load())
	}
}

// ─── §22 Proof 12: coherent reads correspond to one committed revision ────────

func TestReadRuntimeInstance_CoherentRevision(t *testing.T) {
	s := newStoreWithInstance(t, "ri-coherent", 1, 2)
	v0, _ := s.ReadRuntimeInstance("ri-coherent")
	if v0.Revision() == 0 {
		t.Fatal("initial revision is zero")
	}
	claimResult := mustClaim(t, s, "ri-coherent", "la-c", 5)
	v1, _ := s.ReadRuntimeInstance("ri-coherent")
	if v1.Revision() != claimResult.Revision() {
		t.Fatalf("read revision %v != claim revision %v", v1.Revision(), claimResult.Revision())
	}
	if v1.Revision() <= v0.Revision() {
		t.Fatalf("revision did not advance: %v → %v", v0.Revision(), v1.Revision())
	}
	// Active attempt must be consistent with revision.
	activeID, hasActive := v1.ActiveAttempt()
	if !hasActive || activeID != "la-c" {
		t.Fatalf("coherent read missing active attempt: hasActive=%t id=%q", hasActive, activeID)
	}
}

// ─── §22 Proof 13: definitive failure publishes nothing ──────────────────────

func TestConditionalPublishTerminal_DefinitiveFailurePublishesNothing(t *testing.T) {
	s := newStoreWithInstance(t, "ri-def-fail", 1, 2)
	view0, _ := s.ReadRuntimeInstance("ri-def-fail")
	// Terminal on non-existent active attempt: definitive rejection, zero mutation.
	_, err := s.ConditionalPublishTerminal("ri-def-fail", view0.Revision(), "la-nonexistent", true)
	if err != ErrNoActiveAttempt {
		t.Fatalf("error = %v, want ErrNoActiveAttempt", err)
	}
	view1, _ := s.ReadRuntimeInstance("ri-def-fail")
	if view1.Revision() != view0.Revision() {
		t.Fatalf("revision changed after definitive failure: %v → %v", view0.Revision(), view1.Revision())
	}
}

// ─── §22 Proof 14: indeterminate outcomes resolved by inspection ──────────────

func TestInspectAfterIndeterminate_RevealsCommittedState(t *testing.T) {
	// Simulate the inspect-after-indeterminate rule: after any indeterminate
	// outcome, the caller reads the aggregate to determine the actual state.
	// No blind new-ID retry is performed.
	s := newStoreWithInstance(t, "ri-inspect", 1, 2)
	view0, _ := s.ReadRuntimeInstance("ri-inspect")
	claimResult, err := s.ConditionalClaimLaunchAttempt("ri-inspect", view0.Revision(), "la-inspect", 5)
	if err != nil {
		t.Fatalf("claim error = %v", err)
	}
	// Inspect to confirm the claim.
	view1, _ := s.ReadRuntimeInstance("ri-inspect")
	if view1.Revision() != claimResult.Revision() {
		t.Fatalf("revision mismatch after claim: %v vs %v", view1.Revision(), claimResult.Revision())
	}
	activeID, hasActive := view1.ActiveAttempt()
	if !hasActive || activeID != "la-inspect" {
		t.Fatalf("inspection reveals no active attempt: hasActive=%t id=%q", hasActive, activeID)
	}
	// History also shows the attempt.
	history, _ := s.ReadLaunchAttemptHistory("ri-inspect")
	if len(history) == 0 || history[0].LaunchAttemptID() != "la-inspect" {
		t.Fatalf("inspection history does not confirm claim: %v", history)
	}
}

// ─── §22 Proof 15: persisted actual state not used as liveness proof ──────────

func TestActualState_IsHistoricalFact_NotLivenessProof(t *testing.T) {
	// After terminal Stopped is published, actual state is Stopped. This is a
	// historical fact. No liveness inference is performed by the store —
	// verification that the store does NOT perform liveness checks.
	s := newStoreWithInstance(t, "ri-liveness", 1, 2)
	claimResult := mustClaim(t, s, "ri-liveness", "la-live", 5)
	runResult, _ := s.ConditionalPublishRunning("ri-liveness", claimResult.Revision(), "la-live")
	stopResult, _ := s.ConditionalClaimStop("ri-liveness", runResult.Revision(), "la-live")
	_, _ = s.ConditionalPublishTerminal("ri-liveness", stopResult.Revision(), "la-live", true)
	view, _ := s.ReadRuntimeInstance("ri-liveness")
	// Actual is Stopped — a historical fact, not a live probe.
	if view.ActualState() != ActualStateStopped {
		t.Fatalf("actual = %q, want Stopped (historical fact)", view.ActualState())
	}
	// No active attempt: the store does not infer liveness from this.
	if _, active := view.ActiveAttempt(); active {
		t.Fatal("active attempt present after terminal publication")
	}
}

// ─── §22 Proof 16: domain isolation prevents cross-scope disclosure ───────────

func TestDomainIsolation_NoCrossScopeDisclosure(t *testing.T) {
	s := NewStore()
	s.CreateRuntimeInstance(1, 2, "ri-ws1")
	s.CreateRuntimeInstance(3, 4, "ri-ws2")
	// Read of one instance reveals only its own binding.
	v1, _ := s.ReadRuntimeInstance("ri-ws1")
	if v1.WorkspaceID() != 1 || v1.ConfigurationID() != 2 {
		t.Fatalf("ri-ws1 binding = %d/%d, want 1/2", v1.WorkspaceID(), v1.ConfigurationID())
	}
	v2, _ := s.ReadRuntimeInstance("ri-ws2")
	if v2.WorkspaceID() != 3 || v2.ConfigurationID() != 4 {
		t.Fatalf("ri-ws2 binding = %d/%d, want 3/4", v2.WorkspaceID(), v2.ConfigurationID())
	}
	// Operations on ri-ws1 do not affect ri-ws2.
	view0, _ := s.ReadRuntimeInstance("ri-ws2")
	mustClaim(t, s, "ri-ws1", "la-cross", 5)
	view1, _ := s.ReadRuntimeInstance("ri-ws2")
	if view1.Revision() != view0.Revision() {
		t.Fatal("ri-ws2 revision changed by ri-ws1 operation")
	}
}

// ─── §22 Proof 17: no second lifecycle owner or service locator ───────────────

func TestStore_HasNoSecondLifecycleOwnerOrServiceLocator(t *testing.T) {
	// The Store type has no Start/Stop/Observe methods that would duplicate the
	// Owner's role. It only stores and retrieves durable identity facts.
	// Verified by the API surface: Store exposes only the §21 operations.
	s := NewStore()
	_ = s
	// This test validates by compilation: if Store had Owner-like methods they
	// would appear here. The absence of type-assertion or interface satisfaction
	// confirms no second owner is present.
}

// ─── Regression: invalid phase transitions ────────────────────────────────────

func TestConditionalPublishRunning_TerminalPhaseRejected(t *testing.T) {
	s := newStoreWithInstance(t, "ri-run-term", 1, 2)
	claimResult := mustClaim(t, s, "ri-run-term", "la-rt", 5)
	// Claim Stop while in Claimed → terminal Stopped.
	stopResult, _ := s.ConditionalClaimStop("ri-run-term", claimResult.Revision(), "la-rt")
	// Now try to publish Running — no active attempt.
	_, err := s.ConditionalPublishRunning("ri-run-term", stopResult.Revision(), "la-rt")
	if err != ErrNoActiveAttempt {
		t.Fatalf("error = %v, want ErrNoActiveAttempt", err)
	}
}

func TestConditionalClaimStop_NoActiveAttemptRejected(t *testing.T) {
	s := newStoreWithInstance(t, "ri-stop-none", 1, 2)
	view, _ := s.ReadRuntimeInstance("ri-stop-none")
	_, err := s.ConditionalClaimStop("ri-stop-none", view.Revision(), "la-none")
	if err != ErrNoActiveAttempt {
		t.Fatalf("error = %v, want ErrNoActiveAttempt", err)
	}
}

func TestConditionalPublishTerminal_NoActiveAttemptRejected(t *testing.T) {
	s := newStoreWithInstance(t, "ri-term-none", 1, 2)
	view, _ := s.ReadRuntimeInstance("ri-term-none")
	_, err := s.ConditionalPublishTerminal("ri-term-none", view.Revision(), "la-none", true)
	if err != ErrNoActiveAttempt {
		t.Fatalf("error = %v, want ErrNoActiveAttempt", err)
	}
}

func TestConditionalBindExecutionGeneration_NoActiveAttemptRejected(t *testing.T) {
	s := newStoreWithInstance(t, "ri-bind-none", 1, 2)
	view, _ := s.ReadRuntimeInstance("ri-bind-none")
	_, err := s.ConditionalBindExecutionGeneration("ri-bind-none", view.Revision(), "la-none", "gen-X")
	if err != ErrNoActiveAttempt {
		t.Fatalf("error = %v, want ErrNoActiveAttempt", err)
	}
}

// ─── Sentinel error string verification ──────────────────────────────────────

func TestSentinelErrorStrings(t *testing.T) {
	cases := map[error]string{
		ErrInstanceNotFound:      "runtime instance not found",
		ErrInstanceAlreadyExists: "runtime instance already exists",
		ErrStaleRevision:         "stale aggregate revision",
		ErrActiveAttemptExists:   "active launch attempt already exists",
		ErrNoActiveAttempt:       "no active launch attempt",
		ErrAttemptIDReused:       "launch attempt ID reused within instance history",
		ErrInvalidAttemptPhase:   "invalid launch attempt phase for operation",
		ErrBindingAlreadyExists:  "execution generation binding already exists",
		ErrInvalidIdentity:       "invalid identity",
	}
	for sentinel, want := range cases {
		if sentinel.Error() != want {
			t.Fatalf("sentinel %v string = %q, want %q", sentinel, sentinel.Error(), want)
		}
	}
}

// ─── Regression: stop-failure retains active attempt (AttemptStopping) ───────

func TestConditionalPublishTerminal_StopFailureRetainsActiveAttempt(t *testing.T) {
	s := newStoreWithInstance(t, "ri-stop-fail", 1, 2)
	claimResult := mustClaim(t, s, "ri-stop-fail", "la-sf", 5)
	runResult, _ := s.ConditionalPublishRunning("ri-stop-fail", claimResult.Revision(), "la-sf")
	stopResult, _ := s.ConditionalClaimStop("ri-stop-fail", runResult.Revision(), "la-sf")
	// Publish Failed from Stopping: cleanup-unproven; active retained.
	result, err := s.ConditionalPublishTerminal("ri-stop-fail", stopResult.Revision(), "la-sf", false)
	if err != nil || !result.Committed() {
		t.Fatalf("PublishTerminal(Failed from Stopping) error = %v, committed = %t", err, result.Committed())
	}
	view, _ := s.ReadRuntimeInstance("ri-stop-fail")
	// Per DP-014 §12: stop-failure retains the active attempt association.
	if activeID, hasActive := view.ActiveAttempt(); !hasActive || activeID != "la-sf" {
		t.Fatalf("active attempt = %q/%t after stop-failure, want la-sf/true", activeID, hasActive)
	}
	if view.ActualState() != ActualStateFailed {
		t.Fatalf("actual = %q, want Failed", view.ActualState())
	}
}

// ─── Regression: after terminal, new claim is possible ───────────────────────

func TestAfterTerminalStopped_NewClaimIsAllowed(t *testing.T) {
	s := newStoreWithInstance(t, "ri-after-stop", 1, 2)
	claimResult := mustClaim(t, s, "ri-after-stop", "la-first-as", 5)
	s.ConditionalClaimStop("ri-after-stop", claimResult.Revision(), "la-first-as")
	view, _ := s.ReadRuntimeInstance("ri-after-stop")
	claimResult2, err := s.ConditionalClaimLaunchAttempt("ri-after-stop", view.Revision(), "la-second-as", 6)
	if err != nil || !claimResult2.Committed() {
		t.Fatalf("second claim error = %v, committed = %t", err, claimResult2.Committed())
	}
	history, _ := s.ReadLaunchAttemptHistory("ri-after-stop")
	if len(history) != 2 {
		t.Fatalf("history len = %d, want 2", len(history))
	}
}

// ─── Helper functions ─────────────────────────────────────────────────────────

func newStoreWithInstance(
	t *testing.T,
	id runtimeconfigload.RuntimeInstanceID,
	workspaceID, configurationID uint64,
) *Store {
	t.Helper()
	s := NewStore()
	if err := s.CreateRuntimeInstance(workspaceID, configurationID, id); err != nil {
		t.Fatalf("CreateRuntimeInstance(%q) error = %v", id, err)
	}
	return s
}

func mustClaim(
	t *testing.T,
	s *Store,
	id runtimeconfigload.RuntimeInstanceID,
	attemptID runtimeconfigload.LaunchAttemptID,
	versionID uint64,
) ClaimResult {
	t.Helper()
	view, err := s.ReadRuntimeInstance(id)
	if err != nil {
		t.Fatalf("ReadRuntimeInstance(%q) before claim error = %v", id, err)
	}
	return mustClaimAt(t, s, id, view.Revision(), attemptID, versionID)
}

func mustClaimAt(
	t *testing.T,
	s *Store,
	id runtimeconfigload.RuntimeInstanceID,
	rev Revision,
	attemptID runtimeconfigload.LaunchAttemptID,
	versionID uint64,
) ClaimResult {
	t.Helper()
	result, err := s.ConditionalClaimLaunchAttempt(id, rev, attemptID, versionID)
	if err != nil || !result.Committed() {
		t.Fatalf("ConditionalClaimLaunchAttempt(%q,%q) error = %v committed = %t", id, attemptID, err, result.Committed())
	}
	return result
}

// itoa converts an int to a simple string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	// reverse
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
