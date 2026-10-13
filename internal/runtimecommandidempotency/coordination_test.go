package runtimecommandidempotency

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecoordination"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

func coordinatedFixture(t *testing.T, instance string) (*MemoryStorage, *Boundary, Scope, runtimecoordination.Scope) {
	t.Helper()
	s, err := NewCoordinatedMemoryStorage(runtimecoordination.NewAuthorityFence())
	if err != nil {
		t.Fatal(err)
	}
	scope := testScope(t, "domain-a", instance, OperationStart)
	coordScope := coordinationScope(scope.instanceScope())
	p, _ := runtimecoordination.NewParticipant(coordScope)
	if err := s.AttachInstance(coordScope, p); err != nil {
		t.Fatal(err)
	}
	b, err := NewBoundary(s)
	if err != nil {
		t.Fatal(err)
	}
	return s, b, scope, coordScope
}

// seedCommand is a package-local successful-state fixture. It proves observation
// mechanics only and never successful coordinated startup or admission.
func seedCommand(s *MemoryStorage, b *Boundary, scope Scope, key CommandKey, intent Intent, terminal bool) *commandLedger {
	l := s.ledger(scope.instanceScope())
	id := commandIdentity{scope, key}
	l.lock()
	r := &commandRecord{identity: id, intent: intent, state: CommandStateClaimed, revision: 1}
	if terminal {
		r.state = CommandStateTerminal
		r.revision = 2
		r.outcome, _ = NewTerminalOutcome(OutcomeSatisfied, "attempt-fixture")
		r.hasOutcome = true
	}
	l.records[id] = r
	l.unlock()
	return l
}

func seedParent(t *testing.T, s *MemoryStorage, scope Scope, key CommandKey, intent Intent, terminal bool) *commandLedger {
	t.Helper()
	l := s.ledger(scope.instanceScope())
	id := commandIdentity{scope, key}
	l.lock()
	r := &parentRecord{identity: id, intent: intent, state: CommandStateClaimed, revision: 1}
	if terminal {
		r.state = CommandStateTerminal
		r.revision = 2
		r.outcome = parentOutcome(t, ParentOutcomeSatisfied)
		r.hasOutcome = true
	}
	l.parents[id] = r
	l.unlock()
	return l
}

func TestCoordinationAttachmentRejectsZeroForeignDuplicateAndCopies(t *testing.T) {
	if _, err := NewCoordinatedMemoryStorage(nil); err == nil {
		t.Fatal("zero authority accepted")
	}
	s, _, scope, cs := coordinatedFixture(t, "a")
	other, _ := runtimecoordination.NewParticipant(cs)
	if err := s.AttachInstance(cs, other); !errors.Is(err, runtimecoordination.ErrAlreadyAttached) {
		t.Fatal(err)
	}
	copy := &MemoryStorage{coord: s.coord, generation: s.generation, generationDone: s.generationDone, ledgers: s.ledgers}
	if _, err := NewBoundary(copy); !errors.Is(err, ErrInvalidSubmission) {
		t.Fatal("copied owner replaced client", err)
	}
	if _, err := copy.CaptureCoordinationSnapshot(cs); err == nil {
		t.Fatal("copied owner disclosed facts")
	}
	foreign := coordinationScope(testScope(t, "domain-a", "b", OperationStart).instanceScope())
	if err := s.AttachInstance(foreign, other); !errors.Is(err, runtimecoordination.ErrForeignScope) {
		t.Fatal(err)
	}
	if len(s.ledgers) != 0 || scope.instanceScope().runtimeInstanceID != "a" {
		t.Fatal("attachment created ledger")
	}
}

func TestCoordinationUnattachedReadsDoNotDiscloseOrPoisonRegistration(t *testing.T) {
	s, _ := NewCoordinatedMemoryStorage(runtimecoordination.NewAuthorityFence())
	b, _ := NewBoundary(s)
	scope := testScope(t, "domain-a", "a", OperationStart)
	cs := coordinationScope(scope.instanceScope())
	if _, err := s.ReadAssessmentSnapshot(scope.domain, scope.workspaceID, scope.configurationID, scope.runtimeInstanceID); err == nil {
		t.Fatal("unattached absence disclosed")
	}
	if _, err := b.Execute(context.Background(), scope, "a", startIntent(t, 1), allow, success(t)); err == nil {
		t.Fatal("unattached command admitted")
	}
	if len(s.ledgers) != 0 {
		t.Fatal("rejected operation allocated ledger")
	}
	p, _ := runtimecoordination.NewParticipant(cs)
	if err := s.AttachInstance(cs, p); err != nil {
		t.Fatal("rejected observation poisoned attachment", err)
	}
	if snap, err := s.CaptureCoordinationSnapshot(cs); err != nil || snap.Present() {
		t.Fatal("attached absence unavailable", err)
	}
}

func TestCoordinationAllDirectAdmissionsRejectBeforeAllocationAndCallbacks(t *testing.T) {
	s, b, scope, _ := coordinatedFixture(t, "a")
	parentScope := scope
	parentScope.operation = OperationReplace
	var callbacks atomic.Int32
	primitive := func() (TerminalOutcome, error) { callbacks.Add(1); return TerminalOutcome{}, nil }
	parent := func(*ParentExecution) error { callbacks.Add(1); return nil }
	managed := func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
		callbacks.Add(1)
		return TerminalOutcome{}, nil
	}
	managedParent := func(*ManagedParentExecution) error { callbacks.Add(1); return nil }
	tracked := func(*TrackedStartManagedParentExecution) error { callbacks.Add(1); return nil }
	intent := startIntent(t, 1)
	parentIntent := replaceIntent(t, 1)
	_, err := b.Execute(context.Background(), scope, "a", intent, allow, primitive)
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	_, err = b.ExecuteParent(context.Background(), parentScope, "b", parentIntent, allow, parent)
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	_, err = b.ExecuteManagedStart(context.Background(), scope, "c", intent, 1, "g", allowOrchestration, managed)
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	_, err = b.ExecuteManagedParent(context.Background(), parentScope, "d", parentIntent, allowOrchestration, managedParent)
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	_, err = b.ExecuteManagedParentFromTrackedStart(context.Background(), parentScope, "e", parentIntent, allowOrchestration, tracked)
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	if callbacks.Load() != 0 || len(s.ledgers) != 0 {
		t.Fatal("closed admission created facts or delegated")
	}
}

func TestCoordinationAuthorizedPrimitiveAndParentReplayZeroLifecycle(t *testing.T) {
	s, b, scope, cs := coordinatedFixture(t, "a")
	intent := startIntent(t, 1)
	seedCommand(s, b, scope, "a", intent, true)
	parentScope := scope
	parentScope.operation = OperationReplace
	pi := replaceIntent(t, 1)
	seedParent(t, s, parentScope, "b", pi, true)
	var authorizations, callbacks atomic.Int32
	authorize := func(context.Context, Scope, Intent) error { authorizations.Add(1); return nil }
	invoke := func() (TerminalOutcome, error) { callbacks.Add(1); return TerminalOutcome{}, nil }
	a, err := b.Execute(context.Background(), scope, "a", intent, authorize, invoke)
	if err != nil || a.Kind() != AdmissionReplay {
		t.Fatal(a, err)
	}
	p, err := b.ExecuteParent(context.Background(), parentScope, "b", pi, authorize, func(*ParentExecution) error { callbacks.Add(1); return nil })
	if err != nil || p.Kind() != AdmissionReplay {
		t.Fatal(p, err)
	}
	if authorizations.Load() != 2 || callbacks.Load() != 0 {
		t.Fatal("authorization or replay callback mismatch")
	}
	if _, err := s.CaptureCoordinationSnapshot(cs); err != nil {
		t.Fatal("terminal fixture incoherent", err)
	}
	if _, err := b.Execute(context.Background(), scope, "foreign", intent, authorize, invoke); !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
}

func TestCoordinationReplayFirstAbsentRejectsDecisionAndGenerationCallbacks(t *testing.T) {
	s, b, scope, _ := coordinatedFixture(t, "a")
	var callbacks atomic.Int32
	decide := func(context.Context) (AbsentCandidate, error) { callbacks.Add(1); return NewNoClaimCandidate(), nil }
	revalidate := func(context.Context, AbsentCandidate) (CandidateRevalidation, error) {
		callbacks.Add(1)
		return CandidateRevalidated, nil
	}
	provide := func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
		callbacks.Add(1)
		return "g", nil
	}
	_, _, err := b.ExecuteReplayFirstManagedStart(context.Background(), scope, "a", startIntent(t, 1), allowOrchestration, decide, revalidate, provide, func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
		callbacks.Add(1)
		return TerminalOutcome{}, nil
	})
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	scope.operation = OperationReplace
	_, _, err = b.ExecuteReplayFirstManagedParent(context.Background(), scope, "b", replaceIntent(t, 1), allowOrchestration, decide, revalidate, provide, func(*ReplayFirstManagedParentExecution) error { callbacks.Add(1); return nil })
	if !errors.Is(err, runtimecoordination.ErrClosed) || callbacks.Load() != 0 || len(s.ledgers) != 0 {
		t.Fatal("closed replay-first created/delegated", err, callbacks.Load())
	}
}

func TestCoordinationSnapshotAbsenceCreationForeignPartialAndEpochStaleness(t *testing.T) {
	s, b, scope, cs := coordinatedFixture(t, "a")
	before, err := s.CaptureCoordinationSnapshot(cs)
	if err != nil || before.Present() {
		t.Fatal(before, err)
	}
	if err := s.RevalidateCoordinationSnapshot(before); err != nil {
		t.Fatal(err)
	}
	seedCommand(s, b, scope, "a", startIntent(t, 1), false)
	if err := s.RevalidateCoordinationSnapshot(before); err == nil {
		t.Fatal("create after absence retained stamp")
	}
	current, err := s.CaptureCoordinationSnapshot(cs)
	if err != nil || !current.Present() || current.PresenceRevision() != 1 {
		t.Fatal(current, err)
	}
	partial := current
	partial.complete = false
	if err := s.RevalidateCoordinationSnapshot(partial); err == nil {
		t.Fatal("partial stamp accepted")
	}
	other, _, _, _ := coordinatedFixture(t, "a")
	if err := other.RevalidateCoordinationSnapshot(current); err == nil {
		t.Fatal("foreign owner stamp accepted")
	}
	if _, err := NewBoundary(s); err != nil {
		t.Fatal(err)
	}
	if err := s.RevalidateCoordinationSnapshot(current); err == nil {
		t.Fatal("replacement retained snapshot")
	}
}

func TestCoordinationCompleteSnapshotCopiesAllTerminalAndNonterminalMembers(t *testing.T) {
	s, b, scope, cs := coordinatedFixture(t, "a")
	l := seedCommand(s, b, scope, "a", startIntent(t, 1), false)
	seedCommand(s, b, scope, "terminal", startIntent(t, 1), true)
	ps := scope
	ps.operation = OperationReplace
	seedParent(t, s, ps, "p", replaceIntent(t, 1), false)
	pid := commandIdentity{ps, "p"}
	phase, _ := newPhaseIdentity(pid, PhaseStopOld)
	l.lock()
	l.phases[phase] = &phaseRecord{identity: phase, state: CommandStateClaimed, revision: 1}
	l.unlock()
	snap, err := s.CaptureCoordinationSnapshot(cs)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.facts.records) != 2 || len(snap.facts.parents) != 1 || len(snap.facts.phases) != 1 {
		t.Fatal("incomplete membership")
	}
	partial := snap
	partial.facts = detachedFacts(l.facts())
	delete(partial.facts.records, commandIdentity{scope, "terminal"})
	if err := s.RevalidateCoordinationSnapshot(partial); err == nil {
		t.Fatal("partial member set accepted")
	}
	if err := s.RevalidateCoordinationSnapshot(snap); err != nil {
		t.Fatal("snapshot shares mutable maps", err)
	}
}

func TestCoordinationMembershipCountsEveryLocalTransitionClass(t *testing.T) {
	s, b, scope, cs := coordinatedFixture(t, "a")
	l := seedCommand(s, b, scope, "a", startIntent(t, 1), false)
	id := commandIdentity{scope, "a"}
	ps := scope
	ps.operation = OperationReplace
	seedParent(t, s, ps, "p", replaceIntent(t, 1), false)
	pid := commandIdentity{ps, "p"}
	phase, _ := newPhaseIdentity(pid, PhaseStopOld)
	// These are fixture-only coherent publications; closed production writes
	// never use this route. Each actual accounting region must invalidate stamps.
	changes := []struct {
		name   string
		change func()
	}{
		{"primitive revision", func() { l.records[id].revision++ }},
		{"parent revision", func() { l.parents[pid].revision++ }},
		{"phase membership", func() { l.phases[phase] = &phaseRecord{identity: phase, state: CommandStateClaimed, revision: 1} }},
		{"live issuance", func() { l.live[id] = &permitState{generation: b.generation, revision: l.records[id].revision} }},
		{"equal value live identity replacement", func() { v := *l.live[id]; l.live[id] = &v }},
		{"live expiry", func() { delete(l.live, id) }},
		{"rendezvous membership", func() { l.rendezvous[pid] = newStartRendezvous(b.generation) }},
		{"rendezvous field", func() { l.rendezvous[pid].continueCancelled = true }},
		{"rendezvous notification", func() { l.rendezvous[pid].notifyLocked() }},
		{"phase live issuance", func() { l.livePhases[phase] = &permitState{generation: b.generation, revision: 1} }},
		{"phase live expiry", func() { delete(l.livePhases, phase) }},
		{"parent live issuance", func() { l.liveParents[pid] = &permitState{generation: b.generation, revision: l.parents[pid].revision} }},
		{"parent live expiry", func() { delete(l.liveParents, pid) }},
		{"stop slot", func() { v := phase; l.stopForStart[id] = stopExceptionOccupant{phase: &v} }},
		{"stop slot expiry", func() { delete(l.stopForStart, id) }},
	}
	for _, test := range changes {
		t.Run(test.name, func(t *testing.T) {
			before, err := s.CaptureCoordinationSnapshot(cs)
			if err != nil {
				t.Fatal(err)
			}
			l.lock()
			test.change()
			l.unlock()
			after, err := s.CaptureCoordinationSnapshot(cs)
			if err != nil {
				t.Fatal(err)
			}
			if after.membership <= before.membership {
				t.Fatal("transition not counted")
			}
			if err := s.RevalidateCoordinationSnapshot(before); err == nil {
				t.Fatal("transition retained stamp")
			}
		})
	}
}

func TestCoordinationSnapshotsRejectContradictoryAndForeignLinks(t *testing.T) {
	s, b, scope, cs := coordinatedFixture(t, "a")
	l := seedCommand(s, b, scope, "a", startIntent(t, 1), false)
	id := commandIdentity{scope, "a"}
	mutations := []func(){func() { l.records[id] = nil }, func() { l.records[id].revision = 0 }, func() { l.records[id].hasOutcome = true }, func() { l.live[id] = &permitState{generation: b.generation, revision: 99} }, func() { l.stopForStart[id] = stopExceptionOccupant{} }, func() {
		p := commandIdentity{scope, "missing"}
		phase, _ := newPhaseIdentity(p, PhaseStopOld)
		l.phases[phase] = &phaseRecord{identity: phase, state: CommandStateClaimed, revision: 1}
	}}
	for i, change := range mutations {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			l.mu.Lock()
			original := l.records[id]
			copy := *original
			change()
			l.mu.Unlock()
			if _, err := s.CaptureCoordinationSnapshot(cs); err == nil {
				t.Fatal("contradictory snapshot returned")
			}
			l.mu.Lock()
			l.records[id] = &copy
			l.live = make(map[commandIdentity]*permitState)
			l.stopForStart = make(map[commandIdentity]stopExceptionOccupant)
			l.phases = make(map[phaseIdentity]*phaseRecord)
			l.mu.Unlock()
		})
	}
}

func TestCoordinationMembershipExhaustionReadsDoNotMutateAndWritesFence(t *testing.T) {
	s, b, scope, cs := coordinatedFixture(t, "a")
	l := seedCommand(s, b, scope, "a", startIntent(t, 1), false)
	l.membership = ^uint64(0)
	l.fenced = false
	before := l.facts()
	snap, err := s.CaptureCoordinationSnapshot(cs)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.RevalidateCoordinationSnapshot(snap); err != nil {
		t.Fatal(err)
	}
	if l.fenced || l.membership != ^uint64(0) || !sameLedgerFacts(before, l.facts()) {
		t.Fatal("exhausted read mutated owner")
	}
	l.lock()
	if l.writableLocked() {
		t.Fatal("exhausted writer allowed")
	}
	l.unlock()
	if !l.fenced || l.membership != ^uint64(0) {
		t.Fatal("writer did not permanently fence")
	}
}

func TestCoordinationClientExhaustionNeverPublishesAndRemainsFenced(t *testing.T) {
	s, b, scope, _ := coordinatedFixture(t, "a")
	s.generation = ^uint64(0)
	if next, err := NewBoundary(s); err == nil || next != nil {
		t.Fatal("overflow published client")
	}
	if s.generation != ^uint64(0) || !s.clientFenced {
		t.Fatal("generation wrapped or unfenced")
	}
	s.generation = b.generation
	if _, err := NewBoundary(s); err == nil {
		t.Fatal("client fence reopened")
	}
	if _, err := b.Execute(context.Background(), scope, "a", startIntent(t, 1), allow, success(t)); !errors.Is(err, ErrBoundaryExpired) {
		t.Fatal(err)
	}
}

func TestCoordinationRendezvousCounterExhaustionDoesNotWrap(t *testing.T) {
	b := newTestBoundary(t)
	b.rendezvousSeq.Store(^uint64(0))
	if _, err := b.newStartRendezvous(); !errors.Is(err, ErrBoundaryExpired) {
		t.Fatal(err)
	}
	if b.rendezvousSeq.Load() != ^uint64(0) {
		t.Fatal("rendezvous counter wrapped")
	}
}

func TestCoordinationAuthorizationRunsOutsideAllShortGates(t *testing.T) {
	s, b, scope, cs := coordinatedFixture(t, "a")
	done := make(chan error, 1)
	go func() {
		_, err := b.Execute(context.Background(), scope, "a", startIntent(t, 1), func(context.Context, Scope, Intent) error {
			if _, err := s.CaptureCoordinationSnapshot(cs); err != nil {
				return err
			}
			_, err := NewBoundary(s)
			return err
		}, success(t))
		done <- err
	}()
	select {
	case err := <-done:
		if !errors.Is(err, ErrBoundaryExpired) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("authorization ran under gates")
	}
}

func TestCoordinationTwoInstancesAndReplacementRegistrationRaces(t *testing.T) {
	s, b, scope, cs := coordinatedFixture(t, "a")
	before, _ := s.CaptureCoordinationSnapshot(cs)
	var wg sync.WaitGroup
	var failures atomic.Int32
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			other := scope
			other.runtimeInstanceID = runtimeInstanceForTest(i)
			os := coordinationScope(other.instanceScope())
			p, _ := runtimecoordination.NewParticipant(os)
			if err := s.AttachInstance(os, p); err != nil {
				failures.Add(1)
			}
		}(i)
		go func() {
			defer wg.Done()
			if _, err := NewBoundary(s); err != nil {
				failures.Add(1)
			}
		}()
	}
	wg.Wait()
	if failures.Load() != 0 {
		t.Fatal("race rejected valid registration/replacement")
	}
	if err := s.RevalidateCoordinationSnapshot(before); err == nil {
		t.Fatal("old Instance missed fanout")
	}
	for i := 0; i < 20; i++ {
		other := scope
		other.runtimeInstanceID = runtimeInstanceForTest(i)
		snap, err := s.CaptureCoordinationSnapshot(coordinationScope(other.instanceScope()))
		if err != nil {
			t.Fatal(err)
		}
		c, p := snap.Epochs()
		if c != p || snap.ClientGeneration() != b.generation+20 {
			t.Fatal("attachment escaped replacement")
		}
	}
}

func runtimeInstanceForTest(i int) runtimeconfigload.RuntimeInstanceID {
	return runtimeconfigload.RuntimeInstanceID(fmt.Sprintf("instance-%d", i))
}

func TestCoordinationSnapshotContainsNoOwnerOrPermitLinks(t *testing.T) {
	s, b, scope, cs := coordinatedFixture(t, "a")
	l := seedCommand(s, b, scope, "a", startIntent(t, 1), false)
	id := commandIdentity{scope, "a"}
	l.lock()
	l.live[id] = &permitState{generation: b.generation, revision: 1}
	l.unlock()
	snap, err := s.CaptureCoordinationSnapshot(cs)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range snap.facts.live {
		if v.identity != nil {
			t.Fatal("permit link escaped")
		}
	}
	if reflect.TypeOf(snap.owner).Elem() != reflect.TypeOf(coordinationOwner{}) {
		t.Fatal("opaque owner marker linked owner")
	}
	if len(snap.facts.live) != 1 {
		t.Fatal("live membership missing")
	}
}

func TestCoordinationAllCallbackTokensRejectWithoutMutationOrDelegation(t *testing.T) {
	s, b, scope, cs := coordinatedFixture(t, "a")
	l := seedCommand(s, b, scope, "start", startIntent(t, 1), false)
	ps := scope
	ps.operation = OperationReplace
	seedParent(t, s, ps, "p", replaceIntent(t, 1), false)
	id := commandIdentity{scope, "start"}
	pid := commandIdentity{ps, "p"}
	phase, _ := newPhaseIdentity(pid, PhaseStartTarget)
	state := &permitState{generation: b.generation, revision: 1}
	parentState := &permitState{generation: b.generation, revision: 1}
	rendezvous, _ := runtimeorchestrationbinding.NewStartRendezvous("fixture")
	authorization, _ := authorizeCommandMaps(scope, startIntent(t, 1))
	binding, _ := runtimeorchestrationbinding.NewPrimitiveStartExecutionBinding(authorization, 1, "fixture-g", rendezvous)
	l.lock()
	l.live[id] = state
	l.liveParents[pid] = parentState
	l.phases[phase] = &phaseRecord{identity: phase, state: CommandStateClaimed, revision: 1}
	l.livePhases[phase] = state
	l.rendezvous[pid] = newStartRendezvous(b.generation)
	l.managedStart[rendezvous] = &managedStartRendezvous{binding: binding, identity: id, state: state, generation: b.generation, bridge: newStartRendezvous(b.generation)}
	l.unlock()
	p := &ParentExecution{boundary: b, ledger: l, identity: pid, state: parentState, usage: &parentUsage{}}
	p.usage.cond = sync.NewCond(&p.usage.mu)
	permit := &phasePermit{parent: p, identity: phase, state: state}
	managed := &ManagedParentExecution{parent: p, authorization: authorization}
	primitive := &executionPermit{boundary: b, ledger: l, identity: id, state: state, managed: rendezvous}
	var callbacks atomic.Int32
	invoke := func() (TerminalOutcome, error) { callbacks.Add(1); return TerminalOutcome{}, nil }
	managedInvoke := func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
		callbacks.Add(1)
		return TerminalOutcome{}, nil
	}
	provider := func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
		callbacks.Add(1)
		return "g", nil
	}
	before, err := s.CaptureCoordinationSnapshot(cs)
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.InspectOrExecuteStopOld(invoke)
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	_, _, err = p.ContinueOrExecuteStartTarget(context.Background(), func(*StartTargetExecution) (TerminalOutcome, error) { callbacks.Add(1); return TerminalOutcome{}, nil })
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	_, err = p.PublishTerminal(parentOutcome(t, ParentOutcomeSatisfied))
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	_, _, err = managed.ContinueOrExecuteManagedStartTarget(context.Background(), 1, "g", managedInvoke)
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	tracked := &TrackedStartManagedParentExecution{managed: managed, stopOld: permit}
	_, err = tracked.ExecutePreclaimedStopOld(invoke)
	if !errors.Is(err, runtimecoordination.ErrClosed) || tracked.stopConsumed {
		t.Fatal("closed preclaim consumed", err)
	}
	_, err = primitive.execute(context.Background(), invoke)
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	_, err = primitive.executeAfterGeneration(context.Background(), authorization, 1, provider, managedInvoke)
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	phaseManaged := &managedPhasePermit{parent: p, identity: phase, state: state, rendezvous: rendezvous}
	_, err = phaseManaged.executeAfterGeneration(context.Background(), authorization, 1, provider, managedInvoke)
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	_, err = permit.executeStartTarget(func(*StartTargetExecution) (TerminalOutcome, error) { callbacks.Add(1); return TerminalOutcome{}, nil })
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	e := &StartTargetExecution{permit: permit, active: true}
	_, err = e.OwnerClaimed("attempt")
	if !errors.Is(err, runtimecoordination.ErrClosed) || e.signaled {
		t.Fatal("closed signal consumed", err)
	}
	if err = e.StartNoClaim(StartNoClaimRejected); !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	if e.close() || !e.active {
		t.Fatal("closed callback cleanup changed token")
	}
	if err = p.beginUseError(); !errors.Is(err, runtimecoordination.ErrClosed) || p.usage.inFlight != 0 {
		t.Fatal("closed use gate consumed lifetime", err)
	}
	_, err = b.ResolveManagedStartEarly(binding, "attempt")
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	_, err = b.ResolveManagedStartFinal(binding, FinalContinue)
	if !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	if err = b.SignalManagedStartNoClaim(binding, StartNoClaimRejected); !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	primitive.expire()
	p.expire()
	permit.expire()
	phaseManaged.expire()
	permit.blockStartRendezvous()
	if callbacks.Load() != 0 {
		t.Fatal("closed token delegated callbacks")
	}
	if err = s.RevalidateCoordinationSnapshot(before); err != nil {
		t.Fatal("closed callback changed membership", err)
	}
}

func TestCoordinationRetainedResolvedAndBlockedRendezvousAreDiagnostic(t *testing.T) {
	for _, resolution := range []stopResolution{stopResolutionConverged, stopResolutionSatisfiedNoClaim, stopResolutionBlocked} {
		t.Run(fmt.Sprint(resolution), func(t *testing.T) {
			s, b, scope, cs := coordinatedFixture(t, "a")
			ps := scope
			ps.operation = OperationReplace
			l := seedParent(t, s, ps, "p", replaceIntent(t, 1), false)
			pid := commandIdentity{ps, "p"}
			stopScope := scope
			stopScope.operation = OperationStop
			seedCommand(s, b, stopScope, "stop", NewStopIntent(), false)
			stopID := commandIdentity{stopScope, "stop"}
			// Fixture exactly reproduces retained post-publication diagnostic links.
			state := &permitState{generation: b.generation, revision: 1}
			r := newStartRendezvous(b.generation)
			r.stop = stopID
			r.stopState = state
			r.stopConsumed = true
			r.resolution = resolution
			l.lock()
			l.rendezvous[pid] = r
			if resolution == stopResolutionBlocked {
				r.signal = startSignalBlocked
			} else {
				stop := l.records[stopID]
				stop.state = CommandStateTerminal
				stop.revision = 2
				stop.hasOutcome = true
				stop.outcome = TerminalOutcome{category: OutcomeSucceeded}
				if resolution == stopResolutionConverged {
					r.signal = startSignalOwnerClaimed
					r.attempt = "attempt"
					r.stopConverged = true
					stop.outcome.launchAttemptID = "attempt"
				} else {
					r.signal = startSignalNoClaim
				}
			}
			l.unlock()
			if _, err := s.CaptureCoordinationSnapshot(cs); err != nil {
				t.Fatal("valid retained diagnostic relation rejected", err)
			}
			state.revision = 8
			if _, err := s.CaptureCoordinationSnapshot(cs); err == nil {
				t.Fatal("contradictory diagnostic lineage accepted")
			}
		})
	}
}

func TestCoordinationActualLegacyStopFirstBridgeAndMembershipRemainCoherent(t *testing.T) {
	b := newTestBoundary(t)
	ps := testScope(t, "domain-a", "legacy", OperationReplace)
	ss := ps
	ss.operation = OperationStop
	entered := make(chan struct{})
	continueNow := make(chan struct{})
	parentDone := make(chan error, 1)
	go func() {
		_, err := b.ExecuteParent(context.Background(), ps, "p", replaceIntent(t, 1), allow, func(p *ParentExecution) error {
			close(entered)
			<-continueNow
			_, prevented, err := p.ContinueOrExecuteStartTarget(context.Background(), func(*StartTargetExecution) (TerminalOutcome, error) {
				return TerminalOutcome{}, errors.New("unexpected lifecycle")
			})
			if err != nil || !prevented {
				return fmt.Errorf("continue %v/%v", prevented, err)
			}
			_, err = p.PublishTerminal(parentOutcome(t, ParentOutcomeStopped))
			return err
		})
		parentDone <- err
	}()
	<-entered
	stopDone := make(chan error, 1)
	go func() {
		_, err := b.Execute(context.Background(), ss, "stop", NewStopIntent(), allow, func() (TerminalOutcome, error) { return TerminalOutcome{}, errors.New("unexpected Stop lifecycle") })
		stopDone <- err
	}()
	waitForCommandRecord(t, b, ss, "stop")
	close(continueNow)
	if err := <-stopDone; err != nil {
		t.Fatal(err)
	}
	if err := <-parentDone; err != nil {
		t.Fatal(err)
	}
	l := b.storage.existingLedger(ps.instanceScope())
	l.lock()
	defer l.unlock()
	if !l.coherent(ps.instanceScope(), b.generation) {
		t.Fatal("actual legacy Stop-first retained bridge is incoherent")
	}
	if l.membership < 4 {
		t.Fatal("actual legacy transitions not accounted")
	}
}

func TestCoordinationManagedMembershipAndNestedBridgeChangesInvalidate(t *testing.T) {
	s, b, scope, cs := coordinatedFixture(t, "a")
	l := seedCommand(s, b, scope, "a", startIntent(t, 1), false)
	id := commandIdentity{scope, "a"}
	authorization, _ := authorizeCommandMaps(scope, startIntent(t, 1))
	rendezvous, _ := runtimeorchestrationbinding.NewStartRendezvous("fixture-managed")
	binding, _ := runtimeorchestrationbinding.NewPrimitiveStartExecutionBinding(authorization, 1, "g", rendezvous)
	state := &permitState{generation: b.generation, revision: 1}
	l.lock()
	l.live[id] = state
	l.unlock()
	changes := []func(){
		func() {
			l.managedStart[rendezvous] = &managedStartRendezvous{binding: binding, identity: id, state: state, generation: b.generation, bridge: newStartRendezvous(b.generation)}
		},
		func() { l.managedStart[rendezvous].stage = managedStageBindingFailed },
		func() { l.managedStart[rendezvous].bridge.continueCancelled = true },
		func() { l.managedStart[rendezvous].bridge.notifyLocked() },
		func() {
			value := *l.live[id]
			state = &value
			l.live[id] = state
			l.managedStart[rendezvous].state = state
		},
		func() { delete(l.managedStart, rendezvous) },
	}
	for i, change := range changes {
		before, err := s.CaptureCoordinationSnapshot(cs)
		if err != nil {
			t.Fatal(i, err)
		}
		l.lock()
		change()
		l.unlock()
		after, err := s.CaptureCoordinationSnapshot(cs)
		if err != nil {
			t.Fatal(i, err)
		}
		if after.membership <= before.membership || s.RevalidateCoordinationSnapshot(before) == nil {
			t.Fatal("managed transition retained stamp", i)
		}
	}
}

func TestCoordinationPresenceExhaustionDoesNotAllocateOrWrap(t *testing.T) {
	s, _, scope, cs := coordinatedFixture(t, "a")
	s.coord.presence[scope.instanceScope()] = ^uint64(0)
	before, err := s.CaptureCoordinationSnapshot(cs)
	if err != nil {
		t.Fatal(err)
	}
	if ledger := s.ledger(scope.instanceScope()); ledger != nil {
		t.Fatal("overflow allocated ledger")
	}
	if s.coord.presence[scope.instanceScope()] != ^uint64(0) || len(s.ledgers) != 0 {
		t.Fatal("presence wrapped or allocated")
	}
	if err := s.RevalidateCoordinationSnapshot(before); err != nil {
		t.Fatal("overflow changed absence", err)
	}
}

func TestCoordinationLegacyCounterExhaustionRejectsBeforeLifecycle(t *testing.T) {
	b := newTestBoundary(t)
	scope := testScope(t, "domain-a", "legacy-overflow", OperationStart)
	l := seedCommand(b.storage, b, scope, "a", startIntent(t, 1), false)
	id := commandIdentity{scope, "a"}
	l.mu.Lock()
	l.records[id].revision = ^Revision(0)
	state := &permitState{generation: b.generation, revision: ^Revision(0)}
	l.live[id] = state
	l.mu.Unlock()
	p := &executionPermit{boundary: b, ledger: l, identity: id, state: state}
	var callbacks atomic.Int32
	if _, err := p.execute(context.Background(), func() (TerminalOutcome, error) { callbacks.Add(1); return TerminalOutcome{}, nil }); !errors.Is(err, ErrBoundaryExpired) {
		t.Fatal(err)
	}
	if callbacks.Load() != 0 || l.records[id].state != CommandStateClaimed || l.records[id].revision != ^Revision(0) {
		t.Fatal("overflow delegated or partially terminalized")
	}
	state = &permitState{generation: b.generation, revision: l.records[id].revision}
	l.mu.Lock()
	l.live[id] = state
	l.membership = ^uint64(0)
	l.fenced = false
	l.mu.Unlock()
	p.state = state
	p.expire()
	if !l.fenced || l.live[id] != state || l.membership != ^uint64(0) {
		t.Fatal("expiry mutated exhausted membership")
	}
}

func TestCoordinationSnapshotsRejectManagedOperationPhaseAndBindingCorruption(t *testing.T) {
	for _, kind := range []string{"primitive Stop", "linked StopOld", "target mismatch", "parent key mismatch", "phase order"} {
		t.Run(kind, func(t *testing.T) {
			s, b, scope, cs := coordinatedFixture(t, "a")
			l := seedCommand(s, b, scope, "a", startIntent(t, 1), false)
			id := commandIdentity{scope, "a"}
			rendezvous, _ := runtimeorchestrationbinding.NewStartRendezvous("fixture-corrupt")
			authorization, _ := authorizeCommandMaps(scope, startIntent(t, 1))
			binding, _ := runtimeorchestrationbinding.NewPrimitiveStartExecutionBinding(authorization, 1, "g", rendezvous)
			state := &permitState{generation: b.generation, revision: 1}
			l.lock()
			l.live[id] = state
			l.managedStart[rendezvous] = &managedStartRendezvous{binding: binding, identity: id, state: state, generation: b.generation, bridge: newStartRendezvous(b.generation)}
			l.unlock()
			if _, err := s.CaptureCoordinationSnapshot(cs); err != nil {
				t.Fatal("valid managed fixture", err)
			}
			switch kind {
			case "primitive Stop":
				stop := scope
				stop.operation = OperationStop
				seedCommand(s, b, stop, "stop", NewStopIntent(), false)
				stopID := commandIdentity{stop, "stop"}
				l.mu.Lock()
				l.live[stopID] = state
				l.managedStart[rendezvous].identity = stopID
				l.mu.Unlock()
			case "target mismatch":
				wrong, _ := authorizeCommandMaps(scope, startIntent(t, 2))
				wrongBinding, _ := runtimeorchestrationbinding.NewPrimitiveStartExecutionBinding(wrong, 1, "g", rendezvous)
				l.mu.Lock()
				l.managedStart[rendezvous].binding = wrongBinding
				l.mu.Unlock()
			default:
				ps := scope
				ps.operation = OperationReplace
				seedParent(t, s, ps, "p", replaceIntent(t, 1), false)
				pid := commandIdentity{ps, "p"}
				start, _ := newPhaseIdentity(pid, PhaseStartTarget)
				stop, _ := newPhaseIdentity(pid, PhaseStopOld)
				pa, _ := authorizeCommandMaps(ps, replaceIntent(t, 1))
				pk := "p"
				if kind == "parent key mismatch" {
					pk = "wrong"
				}
				parentIdentity, _ := runtimeorchestrationbinding.NewParentCommandIdentity(pa, pk)
				phaseIdentity, _ := runtimeorchestrationbinding.DeriveStartTargetPhaseIdentity(parentIdentity)
				linked, _ := runtimeorchestrationbinding.NewLinkedExecutionIdentity(parentIdentity, phaseIdentity)
				linkedBinding, _ := runtimeorchestrationbinding.NewLinkedStartExecutionBinding(pa, 1, "g", linked, rendezvous)
				l.mu.Lock()
				l.phases[start] = &phaseRecord{identity: start, state: CommandStateClaimed, revision: 1}
				l.livePhases[start] = state
				managed := l.managedStart[rendezvous]
				managed.identity = pid
				managed.phase = start
				managed.binding = linkedBinding
				if kind == "linked StopOld" || kind == "phase order" {
					l.phases[stop] = &phaseRecord{identity: stop, state: CommandStateClaimed, revision: 1}
					if kind == "linked StopOld" {
						delete(l.phases, start)
						delete(l.livePhases, start)
						managed.phase = stop
						l.livePhases[stop] = state
					}
				}
				l.mu.Unlock()
			}
			if _, err := s.CaptureCoordinationSnapshot(cs); err == nil {
				t.Fatal("contradictory managed links disclosed")
			}
		})
	}
}

func TestCoordinationManagedAndReplayFirstTerminalReplaySkipsAllCallbacks(t *testing.T) {
	s, b, scope, _ := coordinatedFixture(t, "a")
	intent := startIntent(t, 1)
	seedCommand(s, b, scope, "a", intent, true)
	ps := scope
	ps.operation = OperationReplace
	pi := replaceIntent(t, 1)
	seedParent(t, s, ps, "p", pi, true)
	var callbacks atomic.Int32
	invoke := func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error) {
		callbacks.Add(1)
		return TerminalOutcome{}, nil
	}
	decide := func(context.Context) (AbsentCandidate, error) { callbacks.Add(1); return NewNoClaimCandidate(), nil }
	revalidate := func(context.Context, AbsentCandidate) (CandidateRevalidation, error) {
		callbacks.Add(1)
		return CandidateRevalidated, nil
	}
	provide := func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
		callbacks.Add(1)
		return "g", nil
	}
	a, err := b.ExecuteManagedStart(context.Background(), scope, "a", intent, 1, "g", allowOrchestration, invoke)
	if err != nil || a.Kind() != AdmissionReplay {
		t.Fatal(a, err)
	}
	p, err := b.ExecuteManagedParent(context.Background(), ps, "p", pi, allowOrchestration, func(*ManagedParentExecution) error { callbacks.Add(1); return nil })
	if err != nil || p.Kind() != AdmissionReplay {
		t.Fatal(p, err)
	}
	p, err = b.ExecuteManagedParentFromTrackedStart(context.Background(), ps, "p", pi, allowOrchestration, func(*TrackedStartManagedParentExecution) error { callbacks.Add(1); return nil })
	if err != nil || p.Kind() != AdmissionReplay {
		t.Fatal(p, err)
	}
	a, _, err = b.ExecuteReplayFirstManagedStart(context.Background(), scope, "a", intent, allowOrchestration, decide, revalidate, provide, invoke)
	if err != nil || a.Kind() != AdmissionReplay {
		t.Fatal(a, err)
	}
	p, _, err = b.ExecuteReplayFirstManagedParent(context.Background(), ps, "p", pi, allowOrchestration, decide, revalidate, provide, func(*ReplayFirstManagedParentExecution) error { callbacks.Add(1); return nil })
	if err != nil || p.Kind() != AdmissionReplay || callbacks.Load() != 0 {
		t.Fatal(p, err, callbacks.Load())
	}
}
