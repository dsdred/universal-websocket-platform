package runtimerecoveryassessment

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimecommandidempotency"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecontainment"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeexecutionevidence"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeidentity"
)

const (
	testWorkspace     = uint64(1)
	testConfiguration = uint64(2)
	testInstance      = runtimeconfigload.RuntimeInstanceID("instance")
	testAttempt       = runtimeconfigload.LaunchAttemptID("attempt")
	testGeneration    = runtimeidentity.ExecutionGeneration("generation")
)

var errTestIndeterminate = errors.New("test indeterminate command")

func TestAssessCleanCommandOnlyAndUnboundWithoutMutation(t *testing.T) {
	target := mustTarget(t, testDomain('1'), testInstance)

	t.Run("clean", func(t *testing.T) {
		identities := newIdentityStore(t, testInstance)
		commands := runtimecommandidempotency.NewMemoryStorage()
		before, _ := identities.ReadRuntimeInstance(testInstance)
		result := Assess(context.Background(), target, identities, commands, panicEvidenceQuery(t))
		assertClassification(t, result, Clean)
		if result.AggregateRevision() != before.Revision() {
			t.Fatalf("revision = %d, want %d", result.AggregateRevision(), before.Revision())
		}
		if _, _, ok := result.Attempt(); ok {
			t.Fatal("clean result exposed an attempt")
		}
		if _, _, ok := result.Evidence(); ok {
			t.Fatal("clean result consumed evidence")
		}
		after, _ := identities.ReadRuntimeInstance(testInstance)
		if before != after {
			t.Fatal("clean assessment mutated identity")
		}
		if snapshot := result.CommandSnapshot(); len(snapshot.PrimitiveRecords()) != 0 {
			t.Fatal("clean result contains commands")
		}
	})

	t.Run("command only", func(t *testing.T) {
		identities := newIdentityStore(t, testInstance)
		commands := unresolvedStart(t, target)
		result := Assess(context.Background(), target, identities, commands, panicEvidenceQuery(t))
		assertClassification(t, result, CommandOnly)
		if got := result.CommandSnapshot().PrimitiveRecords(); len(got) != 1 ||
			got[0].State() != runtimecommandidempotency.CommandStateClaimed {
			t.Fatalf("unexpected command-only snapshot: %#v", got)
		}
	})

	t.Run("unbound attempt", func(t *testing.T) {
		identities := newIdentityStore(t, testInstance)
		claimAttempt(t, identities, testInstance, testAttempt)
		commands := runtimecommandidempotency.NewMemoryStorage()
		result := Assess(context.Background(), target, identities, commands, panicEvidenceQuery(t))
		assertClassification(t, result, UnboundAttempt)
		attemptID, generation, ok := result.Attempt()
		if !ok || attemptID != testAttempt || generation != "" {
			t.Fatalf("unbound attempt = (%q, %q, %v)", attemptID, generation, ok)
		}
	})
}

func TestAssessExactEvidenceDistinctions(t *testing.T) {
	tests := []struct {
		name          string
		stopping      bool
		answers       map[runtimeexecutionevidence.Question]evidenceValue
		want          Classification
		wantQuestions []runtimeexecutionevidence.Question
		wantEvidence  runtimeexecutionevidence.OutcomeKind
		wantReason    runtimeexecutionevidence.UnknownReason
	}{
		{
			name: "execution terminated",
			answers: map[runtimeexecutionevidence.Question]evidenceValue{
				runtimeexecutionevidence.GenerationStatus: {kind: runtimeexecutionevidence.GenerationTerminated},
			},
			want: ExecutionTerminated, wantQuestions: []runtimeexecutionevidence.Question{runtimeexecutionevidence.GenerationStatus},
			wantEvidence: runtimeexecutionevidence.GenerationTerminated,
		},
		{
			name:     "resource absence",
			stopping: true,
			answers: map[runtimeexecutionevidence.Question]evidenceValue{
				runtimeexecutionevidence.ShutdownCompletion:     {kind: runtimeexecutionevidence.Unknown, reason: runtimeexecutionevidence.Absent},
				runtimeexecutionevidence.CoveredResourceAbsence: {kind: runtimeexecutionevidence.CoveredResourcesAbsent},
			},
			want: ResourceAbsence,
			wantQuestions: []runtimeexecutionevidence.Question{
				runtimeexecutionevidence.ShutdownCompletion,
				runtimeexecutionevidence.CoveredResourceAbsence,
			},
			wantEvidence: runtimeexecutionevidence.CoveredResourcesAbsent,
		},
		{
			name:     "shutdown complete",
			stopping: true,
			answers: map[runtimeexecutionevidence.Question]evidenceValue{
				runtimeexecutionevidence.ShutdownCompletion: {kind: runtimeexecutionevidence.HostShutdownCompleted},
			},
			want: ShutdownCompleted, wantQuestions: []runtimeexecutionevidence.Question{runtimeexecutionevidence.ShutdownCompletion},
			wantEvidence: runtimeexecutionevidence.HostShutdownCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := mustTarget(t, testDomain('2'), testInstance)
			identities := boundIdentityStore(t, testInstance, testAttempt, testGeneration, tt.stopping)
			commands := runtimecommandidempotency.NewMemoryStorage()
			var questions []runtimeexecutionevidence.Question
			query := func(
				_ context.Context,
				domain runtimecontainment.Domain,
				instanceID runtimeconfigload.RuntimeInstanceID,
				attemptID runtimeconfigload.LaunchAttemptID,
				generation runtimeidentity.ExecutionGeneration,
				question runtimeexecutionevidence.Question,
			) Evidence {
				if domain != target.containmentDomain || instanceID != testInstance ||
					attemptID != testAttempt || generation != testGeneration {
					t.Fatalf("foreign evidence tuple: domain=%v instance=%q attempt=%q generation=%q", domain, instanceID, attemptID, generation)
				}
				questions = append(questions, question)
				answer, ok := tt.answers[question]
				if !ok {
					t.Fatalf("unexpected evidence question %v", question)
				}
				return answer
			}
			result := Assess(context.Background(), target, identities, commands, query)
			assertClassification(t, result, tt.want)
			if !sameQuestions(questions, tt.wantQuestions) {
				t.Fatalf("questions = %v, want %v", questions, tt.wantQuestions)
			}
			kind, reason, ok := result.Evidence()
			if !ok || kind != tt.wantEvidence || reason != tt.wantReason {
				t.Fatalf("evidence = (%v, %v, %v), want (%v, %v, true)", kind, reason, ok, tt.wantEvidence, tt.wantReason)
			}
			attemptID, generation, ok := result.Attempt()
			if !ok || attemptID != testAttempt || generation != testGeneration {
				t.Fatalf("attempt = (%q, %q, %v)", attemptID, generation, ok)
			}
		})
	}
}

func TestAssessEvidenceFailuresRemainUnknown(t *testing.T) {
	tests := []struct {
		name     string
		evidence evidenceValue
	}{
		{name: "live", evidence: evidenceValue{kind: runtimeexecutionevidence.GenerationLive}},
		{name: "unavailable", evidence: evidenceValue{kind: runtimeexecutionevidence.Unknown, reason: runtimeexecutionevidence.Unavailable}},
		{name: "stale", evidence: evidenceValue{kind: runtimeexecutionevidence.Unknown, reason: runtimeexecutionevidence.Stale}},
		{name: "contradictory", evidence: evidenceValue{kind: runtimeexecutionevidence.Unknown, reason: runtimeexecutionevidence.Contradictory}},
		{name: "cancelled", evidence: evidenceValue{kind: runtimeexecutionevidence.Unknown, reason: runtimeexecutionevidence.Cancelled}},
		{name: "cross-domain", evidence: evidenceValue{kind: runtimeexecutionevidence.Unknown, reason: runtimeexecutionevidence.ScopeMismatch}},
		{name: "live orphan", evidence: evidenceValue{kind: runtimeexecutionevidence.LiveUnownedExecution}},
		{name: "malformed positive", evidence: evidenceValue{kind: runtimeexecutionevidence.GenerationTerminated, reason: runtimeexecutionevidence.Absent}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := mustTarget(t, testDomain('3'), testInstance)
			identities := boundIdentityStore(t, testInstance, testAttempt, testGeneration, false)
			commands := runtimecommandidempotency.NewMemoryStorage()
			result := Assess(context.Background(), target, identities, commands, fixedEvidenceQuery(tt.evidence))
			assertClassification(t, result, Unknown)
		})
	}

	target := mustTarget(t, testDomain('4'), testInstance)
	identities := boundIdentityStore(t, testInstance, testAttempt, testGeneration, false)
	foreignCommands := runtimecommandidempotency.NewMemoryStorage()
	reader := commandReaderFunc(func(string, uint64, uint64, runtimeconfigload.RuntimeInstanceID) (runtimecommandidempotency.AssessmentSnapshot, error) {
		return foreignCommands.ReadAssessmentSnapshot(testDomain('5'), testWorkspace, testConfiguration, testInstance)
	})
	result := Assess(context.Background(), target, identities, reader, panicEvidenceQuery(t))
	assertClassification(t, result, Unknown)
}

func TestAssessCrossValidatesCommandAttemptLinks(t *testing.T) {
	t.Run("ghost primitive attempt", func(t *testing.T) {
		target := mustTarget(t, testDomain('a'), testInstance)
		result := Assess(
			context.Background(), target, newIdentityStore(t, testInstance),
			terminalStart(t, target, 10, "ghost"), panicEvidenceQuery(t),
		)
		assertClassification(t, result, Unknown)
	})

	t.Run("ghost start target attempt", func(t *testing.T) {
		target := mustTarget(t, testDomain('a'), testInstance)
		result := Assess(
			context.Background(), target, newIdentityStore(t, testInstance),
			terminalStartTarget(t, target, 10, "ghost"), panicEvidenceQuery(t),
		)
		assertClassification(t, result, Unknown)
	})

	t.Run("foreign history attempt", func(t *testing.T) {
		target := mustTarget(t, testDomain('b'), testInstance)
		local := newIdentityStore(t, testInstance)
		foreignID := runtimeconfigload.RuntimeInstanceID("foreign")
		foreign := newIdentityStore(t, foreignID)
		claimAttempt(t, foreign, foreignID, testAttempt)
		reader := identityReaderFuncs{
			readInstance: local.ReadRuntimeInstance,
			readHistory:  foreign.ReadLaunchAttemptHistory,
		}
		result := Assess(
			context.Background(), target, reader,
			terminalStart(t, target, 10, testAttempt), panicEvidenceQuery(t),
		)
		assertClassification(t, result, Unknown)
	})

	t.Run("missing active history", func(t *testing.T) {
		target := mustTarget(t, testDomain('c'), testInstance)
		store := newIdentityStore(t, testInstance)
		claimAttempt(t, store, testInstance, testAttempt)
		reader := identityReaderFuncs{
			readInstance: store.ReadRuntimeInstance,
			readHistory: func(runtimeconfigload.RuntimeInstanceID) ([]runtimeidentity.LaunchAttemptRecord, error) {
				return nil, nil
			},
		}
		result := Assess(
			context.Background(), target, reader,
			runtimecommandidempotency.NewMemoryStorage(), panicEvidenceQuery(t),
		)
		assertClassification(t, result, Unknown)
	})

	t.Run("contradictory primitive attempt reference", func(t *testing.T) {
		target := mustTarget(t, testDomain('d'), testInstance)
		identities := boundIdentityStore(t, testInstance, testAttempt, testGeneration, false)
		result := Assess(
			context.Background(), target, identities,
			terminalStart(t, target, 10, "different-attempt"), panicEvidenceQuery(t),
		)
		assertClassification(t, result, Unknown)
	})

	t.Run("primitive version mismatch", func(t *testing.T) {
		target := mustTarget(t, testDomain('e'), testInstance)
		identities := boundIdentityStore(t, testInstance, testAttempt, testGeneration, false)
		result := Assess(
			context.Background(), target, identities,
			terminalStart(t, target, 11, testAttempt), panicEvidenceQuery(t),
		)
		assertClassification(t, result, Unknown)
	})

	t.Run("start target version mismatch", func(t *testing.T) {
		target := mustTarget(t, testDomain('f'), testInstance)
		identities := boundIdentityStore(t, testInstance, testAttempt, testGeneration, false)
		result := Assess(
			context.Background(), target, identities,
			terminalStartTarget(t, target, 11, testAttempt), panicEvidenceQuery(t),
		)
		assertClassification(t, result, Unknown)
	})

	t.Run("matching start link remains assessable", func(t *testing.T) {
		target := mustTarget(t, testDomain('1'), testInstance)
		identities := boundIdentityStore(t, testInstance, testAttempt, testGeneration, false)
		result := Assess(
			context.Background(), target, identities,
			terminalStart(t, target, 10, testAttempt),
			fixedEvidenceQuery(evidenceValue{kind: runtimeexecutionevidence.GenerationTerminated}),
		)
		assertClassification(t, result, ExecutionTerminated)
	})
}

func TestAssessValidatesActiveAttemptCommandAndPhaseCoherence(t *testing.T) {
	t.Run("matching claimed Start preserves unbound classification", func(t *testing.T) {
		target := mustTarget(t, testDomain('2'), testInstance)
		identities := newIdentityStore(t, testInstance)
		claimAttempt(t, identities, testInstance, testAttempt)
		result := Assess(
			context.Background(), target, identities,
			unresolvedStartVersion(t, target, 10), panicEvidenceQuery(t),
		)
		assertClassification(t, result, UnboundAttempt)
	})

	t.Run("claimed Start version mismatch", func(t *testing.T) {
		target := mustTarget(t, testDomain('3'), testInstance)
		identities := newIdentityStore(t, testInstance)
		claimAttempt(t, identities, testInstance, testAttempt)
		result := Assess(
			context.Background(), target, identities,
			unresolvedStartVersion(t, target, 11), panicEvidenceQuery(t),
		)
		assertClassification(t, result, Unknown)
	})

	t.Run("bound attempt still rejects claimed Start version mismatch", func(t *testing.T) {
		target := mustTarget(t, testDomain('4'), testInstance)
		identities := boundIdentityStore(t, testInstance, testAttempt, testGeneration, false)
		result := Assess(
			context.Background(), target, identities,
			unresolvedStartVersion(t, target, 11), panicEvidenceQuery(t),
		)
		assertClassification(t, result, Unknown)
	})

	t.Run("matching claimed StartTarget preserves unbound classification", func(t *testing.T) {
		target := mustTarget(t, testDomain('5'), testInstance)
		identities := newIdentityStore(t, testInstance)
		claimAttempt(t, identities, testInstance, testAttempt)
		result := Assess(
			context.Background(), target, identities,
			unresolvedStartTarget(t, target, 10), panicEvidenceQuery(t),
		)
		assertClassification(t, result, UnboundAttempt)
	})

	t.Run("claimed StartTarget version mismatch", func(t *testing.T) {
		target := mustTarget(t, testDomain('6'), testInstance)
		identities := newIdentityStore(t, testInstance)
		claimAttempt(t, identities, testInstance, testAttempt)
		result := Assess(
			context.Background(), target, identities,
			unresolvedStartTarget(t, target, 11), panicEvidenceQuery(t),
		)
		assertClassification(t, result, Unknown)
	})

	t.Run("unbound Running is contradictory", func(t *testing.T) {
		target := mustTarget(t, testDomain('7'), testInstance)
		identities := newIdentityStore(t, testInstance)
		revision := claimAttempt(t, identities, testInstance, testAttempt)
		published, err := identities.ConditionalPublishRunning(testInstance, revision, testAttempt)
		if err != nil || !published.Committed() {
			t.Fatalf("publish unbound Running: committed=%v err=%v", published.Committed(), err)
		}
		result := Assess(
			context.Background(), target, identities,
			runtimecommandidempotency.NewMemoryStorage(), panicEvidenceQuery(t),
		)
		assertClassification(t, result, Unknown)
	})

	t.Run("unbound Stop from Claimed remains pre-binding", func(t *testing.T) {
		target := mustTarget(t, testDomain('8'), testInstance)
		identities := newIdentityStore(t, testInstance)
		revision := claimAttempt(t, identities, testInstance, testAttempt)
		stopping, err := identities.ConditionalClaimStop(testInstance, revision, testAttempt)
		if err != nil || !stopping.Committed() {
			t.Fatalf("claim pre-binding Stop: committed=%v err=%v", stopping.Committed(), err)
		}
		result := Assess(
			context.Background(), target, identities,
			runtimecommandidempotency.NewMemoryStorage(), panicEvidenceQuery(t),
		)
		assertClassification(t, result, UnboundAttempt)
	})

	t.Run("bound Running remains evidence-assessable", func(t *testing.T) {
		target := mustTarget(t, testDomain('9'), testInstance)
		identities := boundIdentityStore(t, testInstance, testAttempt, testGeneration, false)
		view, _ := identities.ReadRuntimeInstance(testInstance)
		published, err := identities.ConditionalPublishRunning(testInstance, view.Revision(), testAttempt)
		if err != nil || !published.Committed() {
			t.Fatalf("publish bound Running: committed=%v err=%v", published.Committed(), err)
		}
		result := Assess(
			context.Background(), target, identities,
			runtimecommandidempotency.NewMemoryStorage(),
			fixedEvidenceQuery(evidenceValue{kind: runtimeexecutionevidence.GenerationTerminated}),
		)
		assertClassification(t, result, ExecutionTerminated)
	})
}

func TestNonterminalPhaseGenerationCoherence(t *testing.T) {
	tests := []struct {
		name       string
		phase      runtimeidentity.AttemptPhase
		generation runtimeidentity.ExecutionGeneration
		want       bool
	}{
		{name: "claimed before binding", phase: runtimeidentity.AttemptPhaseClaimed, want: true},
		{name: "stop claimed before binding", phase: runtimeidentity.AttemptPhaseStopping, want: true},
		{name: "launching requires binding", phase: runtimeidentity.AttemptPhaseLaunching, want: false},
		{name: "running requires binding", phase: runtimeidentity.AttemptPhaseRunning, want: false},
		{name: "launching bound", phase: runtimeidentity.AttemptPhaseLaunching, generation: testGeneration, want: true},
		{name: "running bound", phase: runtimeidentity.AttemptPhaseRunning, generation: testGeneration, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nonterminalPhaseGenerationCoherent(tt.phase, tt.generation); got != tt.want {
				t.Fatalf("coherent = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssessRejectsIdentityAndCommandChangesBetweenReads(t *testing.T) {
	t.Run("aggregate revision", func(t *testing.T) {
		target := mustTarget(t, testDomain('6'), testInstance)
		store := newIdentityStore(t, testInstance)
		claimAttempt(t, store, testInstance, testAttempt)
		var reads atomic.Int32
		reader := identityReaderFuncs{
			readInstance: func(id runtimeconfigload.RuntimeInstanceID) (runtimeidentity.RuntimeInstanceView, error) {
				if reads.Add(1) == 3 {
					view, _ := store.ReadRuntimeInstance(id)
					_, err := store.ConditionalBindExecutionGeneration(id, view.Revision(), testAttempt, testGeneration)
					if err != nil {
						t.Fatalf("bind during revalidation: %v", err)
					}
				}
				return store.ReadRuntimeInstance(id)
			},
			readHistory: store.ReadLaunchAttemptHistory,
		}
		result := Assess(context.Background(), target, reader, runtimecommandidempotency.NewMemoryStorage(), panicEvidenceQuery(t))
		assertClassification(t, result, Unknown)
	})

	t.Run("command membership", func(t *testing.T) {
		target := mustTarget(t, testDomain('7'), testInstance)
		identities := newIdentityStore(t, testInstance)
		storage := runtimecommandidempotency.NewMemoryStorage()
		boundary, err := runtimecommandidempotency.NewBoundary(storage)
		if err != nil {
			t.Fatalf("new boundary: %v", err)
		}
		var reads atomic.Int32
		reader := commandReaderFunc(func(domain string, workspaceID, configurationID uint64, instanceID runtimeconfigload.RuntimeInstanceID) (runtimecommandidempotency.AssessmentSnapshot, error) {
			if reads.Add(1) == 2 {
				scope, _ := runtimecommandidempotency.NewScope(domain, workspaceID, configurationID, instanceID, runtimecommandidempotency.OperationStart)
				intent, _ := runtimecommandidempotency.NewStartIntent(10)
				outcome, _ := runtimecommandidempotency.NewTerminalOutcome(runtimecommandidempotency.OutcomeSucceeded, "")
				if _, err := boundary.Execute(context.Background(), scope, "new-command", intent, allowCommand, func() (runtimecommandidempotency.TerminalOutcome, error) {
					return outcome, nil
				}); err != nil {
					t.Fatalf("mutate command between reads: %v", err)
				}
			}
			return storage.ReadAssessmentSnapshot(domain, workspaceID, configurationID, instanceID)
		})
		result := Assess(context.Background(), target, identities, reader, panicEvidenceQuery(t))
		assertClassification(t, result, Unknown)
	})

	t.Run("same command revision", func(t *testing.T) {
		target := mustTarget(t, testDomain('7'), testInstance)
		identities := newIdentityStore(t, testInstance)
		storage := runtimecommandidempotency.NewMemoryStorage()
		boundary, err := runtimecommandidempotency.NewBoundary(storage)
		if err != nil {
			t.Fatalf("new boundary: %v", err)
		}
		scope, _ := runtimecommandidempotency.NewScope(
			target.operationalDomain, target.workspaceID, target.configurationID,
			target.runtimeInstanceID, runtimecommandidempotency.OperationStart,
		)
		intent, _ := runtimecommandidempotency.NewStartIntent(10)
		started := make(chan struct{})
		finish := make(chan struct{})
		done := make(chan error, 1)
		go func() {
			_, executeErr := boundary.Execute(context.Background(), scope, "command", intent, allowCommand, func() (runtimecommandidempotency.TerminalOutcome, error) {
				close(started)
				<-finish
				outcome, _ := runtimecommandidempotency.NewTerminalOutcome(runtimecommandidempotency.OutcomeSucceeded, "")
				return outcome, nil
			})
			done <- executeErr
		}()
		<-started
		var reads atomic.Int32
		reader := commandReaderFunc(func(domain string, workspaceID, configurationID uint64, instanceID runtimeconfigload.RuntimeInstanceID) (runtimecommandidempotency.AssessmentSnapshot, error) {
			if reads.Add(1) == 2 {
				close(finish)
				if executeErr := <-done; executeErr != nil {
					t.Fatalf("complete command between reads: %v", executeErr)
				}
			}
			return storage.ReadAssessmentSnapshot(domain, workspaceID, configurationID, instanceID)
		})
		result := Assess(context.Background(), target, identities, reader, panicEvidenceQuery(t))
		assertClassification(t, result, Unknown)
	})

	t.Run("fresh command read revalidates active version linkage", func(t *testing.T) {
		target := mustTarget(t, testDomain('7'), testInstance)
		identities := newIdentityStore(t, testInstance)
		claimAttempt(t, identities, testInstance, testAttempt)
		matching := unresolvedStartVersion(t, target, 10)
		mismatch := unresolvedStartVersion(t, target, 11)
		var reads atomic.Int32
		reader := commandReaderFunc(func(domain string, workspaceID, configurationID uint64, instanceID runtimeconfigload.RuntimeInstanceID) (runtimecommandidempotency.AssessmentSnapshot, error) {
			if reads.Add(1) == 1 {
				return matching.ReadAssessmentSnapshot(domain, workspaceID, configurationID, instanceID)
			}
			return mismatch.ReadAssessmentSnapshot(domain, workspaceID, configurationID, instanceID)
		})
		result := Assess(context.Background(), target, identities, reader, panicEvidenceQuery(t))
		assertClassification(t, result, Unknown)
	})
}

func TestAssessRejectsDuplicateForeignAndCancelledInputs(t *testing.T) {
	target := mustTarget(t, testDomain('8'), testInstance)
	base := boundIdentityStore(t, testInstance, testAttempt, testGeneration, false)

	duplicate := identityReaderFuncs{
		readInstance: base.ReadRuntimeInstance,
		readHistory: func(id runtimeconfigload.RuntimeInstanceID) ([]runtimeidentity.LaunchAttemptRecord, error) {
			history, err := base.ReadLaunchAttemptHistory(id)
			return append(history, history...), err
		},
	}
	result := Assess(context.Background(), target, duplicate, runtimecommandidempotency.NewMemoryStorage(), panicEvidenceQuery(t))
	assertClassification(t, result, Unknown)

	foreign := newIdentityStoreFor(t, 99, testConfiguration, testInstance)
	result = Assess(context.Background(), target, foreign, runtimecommandidempotency.NewMemoryStorage(), panicEvidenceQuery(t))
	assertClassification(t, result, Unknown)

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	var evidenceCalls atomic.Int32
	result = Assess(cancelled, target, base, runtimecommandidempotency.NewMemoryStorage(), func(
		context.Context,
		runtimecontainment.Domain,
		runtimeconfigload.RuntimeInstanceID,
		runtimeconfigload.LaunchAttemptID,
		runtimeidentity.ExecutionGeneration,
		runtimeexecutionevidence.Question,
	) Evidence {
		evidenceCalls.Add(1)
		return evidenceValue{kind: runtimeexecutionevidence.GenerationTerminated}
	})
	assertClassification(t, result, Unknown)
	if evidenceCalls.Load() != 0 {
		t.Fatalf("cancelled assessment made %d evidence calls", evidenceCalls.Load())
	}
}

func TestConcurrentAssessmentsCreateNoAuthorityAndInstancesRemainIndependent(t *testing.T) {
	domain := testDomain('9')
	firstID := runtimeconfigload.RuntimeInstanceID("first")
	secondID := runtimeconfigload.RuntimeInstanceID("second")
	identities := runtimeidentity.NewStore()
	if err := identities.CreateRuntimeInstance(testWorkspace, testConfiguration, firstID); err != nil {
		t.Fatalf("create first: %v", err)
	}
	if err := identities.CreateRuntimeInstance(testWorkspace, testConfiguration, secondID); err != nil {
		t.Fatalf("create second: %v", err)
	}
	commands := runtimecommandidempotency.NewMemoryStorage()
	firstTarget := mustTarget(t, domain, firstID)
	secondTarget := mustTarget(t, domain, secondID)
	beforeFirst, _ := identities.ReadRuntimeInstance(firstID)
	beforeSecond, _ := identities.ReadRuntimeInstance(secondID)

	const count = 64
	var wg sync.WaitGroup
	errs := make(chan string, count)
	for i := range count {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			target := firstTarget
			if index%2 == 1 {
				target = secondTarget
			}
			if got := Assess(context.Background(), target, identities, commands, panicEvidenceQuery(t)).Classification(); got != Clean {
				errs <- "classification was not Clean"
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	afterFirst, _ := identities.ReadRuntimeInstance(firstID)
	afterSecond, _ := identities.ReadRuntimeInstance(secondID)
	if beforeFirst != afterFirst || beforeSecond != afterSecond {
		t.Fatal("concurrent assessment mutated aggregate state")
	}
	firstCommands, _ := commands.ReadAssessmentSnapshot(domain, testWorkspace, testConfiguration, firstID)
	secondCommands, _ := commands.ReadAssessmentSnapshot(domain, testWorkspace, testConfiguration, secondID)
	if len(firstCommands.PrimitiveRecords()) != 0 || len(firstCommands.ParentRecords()) != 0 || len(firstCommands.PhaseRecords()) != 0 ||
		len(secondCommands.PrimitiveRecords()) != 0 || len(secondCommands.ParentRecords()) != 0 || len(secondCommands.PhaseRecords()) != 0 {
		t.Fatal("concurrent assessment created command authority")
	}
}

type evidenceValue struct {
	kind   runtimeexecutionevidence.OutcomeKind
	reason runtimeexecutionevidence.UnknownReason
}

func (e evidenceValue) Kind() runtimeexecutionevidence.OutcomeKind     { return e.kind }
func (e evidenceValue) Reason() runtimeexecutionevidence.UnknownReason { return e.reason }

type commandReaderFunc func(string, uint64, uint64, runtimeconfigload.RuntimeInstanceID) (runtimecommandidempotency.AssessmentSnapshot, error)

func (f commandReaderFunc) ReadAssessmentSnapshot(
	domain string,
	workspaceID uint64,
	configurationID uint64,
	instanceID runtimeconfigload.RuntimeInstanceID,
) (runtimecommandidempotency.AssessmentSnapshot, error) {
	return f(domain, workspaceID, configurationID, instanceID)
}

type identityReaderFuncs struct {
	readInstance func(runtimeconfigload.RuntimeInstanceID) (runtimeidentity.RuntimeInstanceView, error)
	readHistory  func(runtimeconfigload.RuntimeInstanceID) ([]runtimeidentity.LaunchAttemptRecord, error)
}

func (r identityReaderFuncs) ReadRuntimeInstance(id runtimeconfigload.RuntimeInstanceID) (runtimeidentity.RuntimeInstanceView, error) {
	return r.readInstance(id)
}

func (r identityReaderFuncs) ReadLaunchAttemptHistory(id runtimeconfigload.RuntimeInstanceID) ([]runtimeidentity.LaunchAttemptRecord, error) {
	return r.readHistory(id)
}

func testDomain(digit byte) string { return strings.Repeat(string(digit), 64) }

func mustTarget(t *testing.T, domain string, instanceID runtimeconfigload.RuntimeInstanceID) Target {
	t.Helper()
	target, err := NewTarget(domain, testWorkspace, testConfiguration, instanceID)
	if err != nil {
		t.Fatalf("new target: %v", err)
	}
	return target
}

func newIdentityStore(t *testing.T, instanceID runtimeconfigload.RuntimeInstanceID) *runtimeidentity.Store {
	t.Helper()
	return newIdentityStoreFor(t, testWorkspace, testConfiguration, instanceID)
}

func newIdentityStoreFor(t *testing.T, workspaceID, configurationID uint64, instanceID runtimeconfigload.RuntimeInstanceID) *runtimeidentity.Store {
	t.Helper()
	store := runtimeidentity.NewStore()
	if err := store.CreateRuntimeInstance(workspaceID, configurationID, instanceID); err != nil {
		t.Fatalf("create Runtime Instance: %v", err)
	}
	return store
}

func claimAttempt(t *testing.T, store *runtimeidentity.Store, instanceID runtimeconfigload.RuntimeInstanceID, attemptID runtimeconfigload.LaunchAttemptID) runtimeidentity.Revision {
	t.Helper()
	view, err := store.ReadRuntimeInstance(instanceID)
	if err != nil {
		t.Fatalf("read before claim: %v", err)
	}
	claim, err := store.ConditionalClaimLaunchAttempt(instanceID, view.Revision(), attemptID, 10)
	if err != nil || !claim.Committed() {
		t.Fatalf("claim attempt: committed=%v err=%v", claim.Committed(), err)
	}
	return claim.Revision()
}

func boundIdentityStore(
	t *testing.T,
	instanceID runtimeconfigload.RuntimeInstanceID,
	attemptID runtimeconfigload.LaunchAttemptID,
	generation runtimeidentity.ExecutionGeneration,
	stopping bool,
) *runtimeidentity.Store {
	t.Helper()
	store := newIdentityStore(t, instanceID)
	revision := claimAttempt(t, store, instanceID, attemptID)
	bound, err := store.ConditionalBindExecutionGeneration(instanceID, revision, attemptID, generation)
	if err != nil || !bound.Committed() {
		t.Fatalf("bind generation: committed=%v err=%v", bound.Committed(), err)
	}
	if stopping {
		stopped, err := store.ConditionalClaimStop(instanceID, bound.Revision(), attemptID)
		if err != nil || !stopped.Committed() {
			t.Fatalf("claim Stop: committed=%v err=%v", stopped.Committed(), err)
		}
	}
	return store
}

func unresolvedStart(t *testing.T, target Target) *runtimecommandidempotency.MemoryStorage {
	t.Helper()
	return unresolvedStartVersion(t, target, 10)
}

func unresolvedStartVersion(t *testing.T, target Target, version uint64) *runtimecommandidempotency.MemoryStorage {
	t.Helper()
	storage := runtimecommandidempotency.NewMemoryStorage()
	boundary, err := runtimecommandidempotency.NewBoundary(storage)
	if err != nil {
		t.Fatalf("new boundary: %v", err)
	}
	scope, err := runtimecommandidempotency.NewScope(
		target.operationalDomain,
		target.workspaceID,
		target.configurationID,
		target.runtimeInstanceID,
		runtimecommandidempotency.OperationStart,
	)
	if err != nil {
		t.Fatalf("new command scope: %v", err)
	}
	intent, _ := runtimecommandidempotency.NewStartIntent(version)
	admission, err := boundary.Execute(context.Background(), scope, "command", intent, allowCommand, func() (runtimecommandidempotency.TerminalOutcome, error) {
		return runtimecommandidempotency.TerminalOutcome{}, errTestIndeterminate
	})
	if !errors.Is(err, runtimecommandidempotency.ErrIndeterminateExecution) ||
		admission.Kind() != runtimecommandidempotency.AdmissionClaimed {
		t.Fatalf("create unresolved command: admission=%v err=%v", admission.Kind(), err)
	}
	return storage
}

func unresolvedStartTarget(
	t *testing.T,
	target Target,
	version uint64,
) *runtimecommandidempotency.MemoryStorage {
	t.Helper()
	storage := runtimecommandidempotency.NewMemoryStorage()
	boundary, err := runtimecommandidempotency.NewBoundary(storage)
	if err != nil {
		t.Fatalf("new boundary: %v", err)
	}
	scope, err := runtimecommandidempotency.NewScope(
		target.operationalDomain, target.workspaceID, target.configurationID,
		target.runtimeInstanceID, runtimecommandidempotency.OperationReplace,
	)
	if err != nil {
		t.Fatalf("new unresolved parent scope: %v", err)
	}
	intent, err := runtimecommandidempotency.NewReplaceIntent(version)
	if err != nil {
		t.Fatalf("new unresolved parent intent: %v", err)
	}
	_, executeErr := boundary.ExecuteParent(
		context.Background(), scope, "unresolved-parent", intent, allowCommand,
		func(parent *runtimecommandidempotency.ParentExecution) error {
			_, _, phaseErr := parent.ContinueOrExecuteStartTarget(
				context.Background(),
				func(*runtimecommandidempotency.StartTargetExecution) (runtimecommandidempotency.TerminalOutcome, error) {
					return runtimecommandidempotency.TerminalOutcome{}, errTestIndeterminate
				},
			)
			return phaseErr
		},
	)
	if !errors.Is(executeErr, runtimecommandidempotency.ErrIndeterminateExecution) {
		t.Fatalf("create unresolved StartTarget: %v", executeErr)
	}
	return storage
}

func terminalStart(
	t *testing.T,
	target Target,
	version uint64,
	attemptID runtimeconfigload.LaunchAttemptID,
) *runtimecommandidempotency.MemoryStorage {
	t.Helper()
	storage := runtimecommandidempotency.NewMemoryStorage()
	boundary, err := runtimecommandidempotency.NewBoundary(storage)
	if err != nil {
		t.Fatalf("new boundary: %v", err)
	}
	scope, err := runtimecommandidempotency.NewScope(
		target.operationalDomain, target.workspaceID, target.configurationID,
		target.runtimeInstanceID, runtimecommandidempotency.OperationStart,
	)
	if err != nil {
		t.Fatalf("new terminal Start scope: %v", err)
	}
	intent, err := runtimecommandidempotency.NewStartIntent(version)
	if err != nil {
		t.Fatalf("new terminal Start intent: %v", err)
	}
	outcome, err := runtimecommandidempotency.NewTerminalOutcome(
		runtimecommandidempotency.OutcomeSucceeded, attemptID,
	)
	if err != nil {
		t.Fatalf("new terminal Start outcome: %v", err)
	}
	if _, err := boundary.Execute(context.Background(), scope, "terminal-start", intent, allowCommand, func() (runtimecommandidempotency.TerminalOutcome, error) {
		return outcome, nil
	}); err != nil {
		t.Fatalf("execute terminal Start: %v", err)
	}
	return storage
}

func terminalStartTarget(
	t *testing.T,
	target Target,
	version uint64,
	attemptID runtimeconfigload.LaunchAttemptID,
) *runtimecommandidempotency.MemoryStorage {
	t.Helper()
	storage := runtimecommandidempotency.NewMemoryStorage()
	boundary, err := runtimecommandidempotency.NewBoundary(storage)
	if err != nil {
		t.Fatalf("new boundary: %v", err)
	}
	scope, err := runtimecommandidempotency.NewScope(
		target.operationalDomain, target.workspaceID, target.configurationID,
		target.runtimeInstanceID, runtimecommandidempotency.OperationReplace,
	)
	if err != nil {
		t.Fatalf("new parent scope: %v", err)
	}
	intent, err := runtimecommandidempotency.NewReplaceIntent(version)
	if err != nil {
		t.Fatalf("new parent intent: %v", err)
	}
	outcome, err := runtimecommandidempotency.NewTerminalOutcome(
		runtimecommandidempotency.OutcomeSucceeded, attemptID,
	)
	if err != nil {
		t.Fatalf("new StartTarget outcome: %v", err)
	}
	parentOutcome, err := runtimecommandidempotency.NewParentTerminalOutcome(
		runtimecommandidempotency.ParentOutcomeSucceeded,
	)
	if err != nil {
		t.Fatalf("new parent outcome: %v", err)
	}
	_, err = boundary.ExecuteParent(context.Background(), scope, "parent", intent, allowCommand, func(parent *runtimecommandidempotency.ParentExecution) error {
		_, prevented, phaseErr := parent.ContinueOrExecuteStartTarget(
			context.Background(),
			func(start *runtimecommandidempotency.StartTargetExecution) (runtimecommandidempotency.TerminalOutcome, error) {
				stopped, signalErr := start.OwnerClaimed(attemptID)
				if signalErr != nil {
					return runtimecommandidempotency.TerminalOutcome{}, signalErr
				}
				if stopped {
					return runtimecommandidempotency.TerminalOutcome{}, errTestIndeterminate
				}
				return outcome, nil
			},
		)
		if phaseErr != nil || prevented {
			return phaseErr
		}
		_, publishErr := parent.PublishTerminal(parentOutcome)
		return publishErr
	})
	if err != nil {
		t.Fatalf("execute terminal StartTarget: %v", err)
	}
	return storage
}

func allowCommand(context.Context, runtimecommandidempotency.Scope, runtimecommandidempotency.Intent) error {
	return nil
}

func fixedEvidenceQuery(value evidenceValue) EvidenceQuery {
	return func(
		context.Context,
		runtimecontainment.Domain,
		runtimeconfigload.RuntimeInstanceID,
		runtimeconfigload.LaunchAttemptID,
		runtimeidentity.ExecutionGeneration,
		runtimeexecutionevidence.Question,
	) Evidence {
		return value
	}
}

func panicEvidenceQuery(t *testing.T) EvidenceQuery {
	t.Helper()
	return func(
		context.Context,
		runtimecontainment.Domain,
		runtimeconfigload.RuntimeInstanceID,
		runtimeconfigload.LaunchAttemptID,
		runtimeidentity.ExecutionGeneration,
		runtimeexecutionevidence.Question,
	) Evidence {
		t.Fatal("evidence must not be queried")
		return nil
	}
}

func assertClassification(t *testing.T, result Result, want Classification) {
	t.Helper()
	if got := result.Classification(); got != want {
		t.Fatalf("classification = %v, want %v", got, want)
	}
}

func sameQuestions(left, right []runtimeexecutionevidence.Question) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
