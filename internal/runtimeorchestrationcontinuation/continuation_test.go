package runtimeorchestrationcontinuation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimecommandidempotency"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeidentity"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelaunchflow"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

type traceIdentityStore struct {
	store              *runtimeidentity.Store
	calls              []string
	claimErr           error
	bindErr            error
	readErrAt          int
	readErr            error
	reads              int
	mutateOnHistory    bool
	claimRevision      runtimeidentity.Revision
	bindRevision       runtimeidentity.Revision
	claimCommitThenErr bool
	bindCommitThenErr  bool
	historyErr         error
	bStore             *runtimeidentity.Store
}

func (s *traceIdentityStore) ConditionalClaimLaunchAttempt(
	id runtimeconfigload.RuntimeInstanceID, revision runtimeidentity.Revision,
	attempt runtimeconfigload.LaunchAttemptID, version uint64,
) (runtimeidentity.ClaimResult, error) {
	s.calls = append(s.calls, "claim")
	s.claimRevision = revision
	if s.claimErr != nil {
		if s.claimCommitThenErr {
			_, _ = s.store.ConditionalClaimLaunchAttempt(id, revision, attempt, version)
		}
		return runtimeidentity.ClaimResult{}, s.claimErr
	}
	return s.store.ConditionalClaimLaunchAttempt(id, revision, attempt, version)
}

func (s *traceIdentityStore) ConditionalBindExecutionGeneration(
	id runtimeconfigload.RuntimeInstanceID, revision runtimeidentity.Revision,
	attempt runtimeconfigload.LaunchAttemptID, generation runtimeidentity.ExecutionGeneration,
) (runtimeidentity.PublishResult, error) {
	s.calls = append(s.calls, "bind")
	s.bindRevision = revision
	if s.bindErr != nil {
		if s.bindCommitThenErr {
			_, _ = s.store.ConditionalBindExecutionGeneration(id, revision, attempt, generation)
		}
		return runtimeidentity.PublishResult{}, s.bindErr
	}
	return s.store.ConditionalBindExecutionGeneration(id, revision, attempt, generation)
}

func (s *traceIdentityStore) ReadRuntimeInstance(
	id runtimeconfigload.RuntimeInstanceID,
) (runtimeidentity.RuntimeInstanceView, error) {
	s.calls = append(s.calls, "read")
	s.reads++
	if s.readErrAt == s.reads {
		return runtimeidentity.RuntimeInstanceView{}, s.readErr
	}
	if s.reads == 3 && s.bStore != nil {
		return s.bStore.ReadRuntimeInstance(id)
	}
	return s.store.ReadRuntimeInstance(id)
}

func (s *traceIdentityStore) ReadLaunchAttemptHistory(
	id runtimeconfigload.RuntimeInstanceID,
) ([]runtimeidentity.LaunchAttemptRecord, error) {
	s.calls = append(s.calls, "history")
	if s.historyErr != nil {
		return nil, s.historyErr
	}
	history, err := s.store.ReadLaunchAttemptHistory(id)
	if err == nil && s.mutateOnHistory {
		view, _ := s.store.ReadRuntimeInstance(id)
		_, _ = s.store.ConditionalClaimLaunchAttempt(id, view.Revision(), "other-attempt", 33)
	}
	return history, err
}

func TestAfterOwnerClaimClaimsBindsAndClearsFinalGate(t *testing.T) {
	identity := runtimeidentity.NewStore()
	if err := identity.CreateRuntimeInstance(11, 22, "instance-a"); err != nil {
		t.Fatal(err)
	}
	boundary, err := runtimecommandidempotency.NewBoundary(runtimecommandidempotency.NewMemoryStorage())
	if err != nil {
		t.Fatal(err)
	}
	continuation, err := New(boundary, identity)
	if err != nil {
		t.Fatal(err)
	}
	scope, _ := runtimecommandidempotency.NewScope(
		"domain-a", 11, 22, "instance-a", runtimecommandidempotency.OperationStart,
	)
	intent, _ := runtimecommandidempotency.NewStartIntent(33)
	authorize := func(context.Context, runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error { return nil }
	admission, err := boundary.ExecuteManagedStart(
		context.Background(), scope, "start-a", intent, 1, "generation-a", authorize,
		func(binding runtimeorchestrationbinding.StartExecutionBinding) (runtimecommandidempotency.TerminalOutcome, error) {
			view, viewErr := runtimelaunchflow.NewOwnerClaimView(runtimeconfigload.NewLoadRequest(
				11, 22, 33, "instance-a", "attempt-a",
			))
			if viewErr != nil {
				return runtimecommandidempotency.TerminalOutcome{}, viewErr
			}
			outcome, continuationErr := continuation.AfterOwnerClaim(context.Background(), binding, view)
			if continuationErr != nil || outcome != runtimelaunchflow.StartClaimContinue {
				t.Fatalf("continuation = %v, %v", outcome, continuationErr)
			}
			return runtimecommandidempotency.NewTerminalOutcome(
				runtimecommandidempotency.OutcomeSucceeded, "attempt-a",
			)
		},
	)
	if err != nil || admission.Kind() != runtimecommandidempotency.AdmissionClaimed {
		t.Fatalf("managed admission = %v, %v", admission.Kind(), err)
	}
	view, err := identity.ReadRuntimeInstance("instance-a")
	if err != nil || view.Revision() != 3 {
		t.Fatalf("identity revision = %v, %v", view.Revision(), err)
	}
	history, _ := identity.ReadLaunchAttemptHistory("instance-a")
	if len(history) != 1 || history[0].ExecutionGeneration() != "generation-a" {
		t.Fatalf("unexpected history: %#v", history)
	}
}

func TestAfterOwnerClaimConvergesExactExistingClaimAndGeneration(t *testing.T) {
	identity := runtimeidentity.NewStore()
	if err := identity.CreateRuntimeInstance(11, 22, "instance-a"); err != nil {
		t.Fatal(err)
	}
	claim, err := identity.ConditionalClaimLaunchAttempt("instance-a", 1, "attempt-a", 33)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := identity.ConditionalBindExecutionGeneration(
		"instance-a", claim.Revision(), "attempt-a", "generation-a",
	); err != nil {
		t.Fatal(err)
	}
	boundary, _ := runtimecommandidempotency.NewBoundary(runtimecommandidempotency.NewMemoryStorage())
	continuation, _ := New(boundary, identity)
	scope, _ := runtimecommandidempotency.NewScope(
		"domain-a", 11, 22, "instance-a", runtimecommandidempotency.OperationStart,
	)
	intent, _ := runtimecommandidempotency.NewStartIntent(33)
	_, err = boundary.ExecuteManagedStart(
		context.Background(), scope, "start-a", intent, 3, "generation-a",
		func(context.Context, runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error { return nil },
		func(binding runtimeorchestrationbinding.StartExecutionBinding) (runtimecommandidempotency.TerminalOutcome, error) {
			view, _ := runtimelaunchflow.NewOwnerClaimView(runtimeconfigload.NewLoadRequest(
				11, 22, 33, "instance-a", "attempt-a",
			))
			outcome, continuationErr := continuation.AfterOwnerClaim(context.Background(), binding, view)
			if continuationErr != nil || outcome != runtimelaunchflow.StartClaimContinue {
				t.Fatalf("convergence = %v, %v", outcome, continuationErr)
			}
			return runtimecommandidempotency.NewTerminalOutcome(
				runtimecommandidempotency.OutcomeSucceeded, "attempt-a",
			)
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAfterOwnerClaimPreflightRejectsScopeAndStaleRevisionBeforeWrites(t *testing.T) {
	for _, tc := range []struct {
		name      string
		workspace uint64
		revision  runtimeorchestrationbinding.AggregateRevision
		want      error
	}{
		{name: "foreign-scope", workspace: 99, revision: 1, want: ErrAggregateScopeMismatch},
		{name: "stale-revision", workspace: 11, revision: 2, want: runtimeidentity.ErrStaleRevision},
	} {
		t.Run(tc.name, func(t *testing.T) {
			identity := runtimeidentity.NewStore()
			if err := identity.CreateRuntimeInstance(tc.workspace, 22, "instance-a"); err != nil {
				t.Fatal(err)
			}
			trace := &traceIdentityStore{store: identity}
			outcome, gotErr := executeContinuation(t, trace, tc.revision)
			if outcome != runtimelaunchflow.StartClaimBlocked || !errors.Is(gotErr, tc.want) {
				t.Fatalf("result = %v, %v", outcome, gotErr)
			}
			if !reflect.DeepEqual(trace.calls, []string{"read"}) {
				t.Fatalf("calls = %v, want preflight only", trace.calls)
			}
			history, _ := identity.ReadLaunchAttemptHistory("instance-a")
			if len(history) != 0 {
				t.Fatalf("preflight failure mutated identity: %#v", history)
			}
		})
	}
}

func TestAfterOwnerClaimThreadsOnlyCommittedClaimRevisionIntoBind(t *testing.T) {
	identity := runtimeidentity.NewStore()
	if err := identity.CreateRuntimeInstance(11, 22, "instance-a"); err != nil {
		t.Fatal(err)
	}
	trace := &traceIdentityStore{store: identity}
	outcome, gotErr := executeContinuation(t, trace, 1)
	if gotErr != nil || outcome != runtimelaunchflow.StartClaimContinue {
		t.Fatalf("result = %v, %v", outcome, gotErr)
	}
	if !reflect.DeepEqual(trace.calls, []string{"read", "claim", "bind"}) ||
		trace.claimRevision != 1 || trace.bindRevision != 2 {
		t.Fatalf("calls=%v claim-rev=%d bind-rev=%d", trace.calls, trace.claimRevision, trace.bindRevision)
	}
}

func TestAfterOwnerClaimFailureSandwichUsesFreshAHistoryB(t *testing.T) {
	for _, stage := range []string{"claim", "bind"} {
		t.Run(stage, func(t *testing.T) {
			writeErr := errors.New(stage + " rejected")
			identity := runtimeidentity.NewStore()
			if err := identity.CreateRuntimeInstance(11, 22, "instance-a"); err != nil {
				t.Fatal(err)
			}
			trace := &traceIdentityStore{store: identity}
			want := []string{"read", "claim", "read", "history", "read"}
			if stage == "claim" {
				trace.claimErr = writeErr
			} else {
				trace.bindErr = writeErr
				want = []string{"read", "claim", "bind", "read", "history", "read"}
			}
			outcome, gotErr := executeContinuation(t, trace, 1)
			if outcome != runtimelaunchflow.StartClaimBindingFailed || gotErr != writeErr {
				t.Fatalf("result = %v, %v", outcome, gotErr)
			}
			if !reflect.DeepEqual(trace.calls, want) {
				t.Fatalf("calls = %v, want %v", trace.calls, want)
			}
		})
	}
}

func TestAfterOwnerClaimInspectionErrorAndIncoherenceSupersedeWriteError(t *testing.T) {
	writeErr := errors.New("claim rejected")
	readErr := errors.New("read unavailable")
	for _, tc := range []struct {
		name  string
		trace func(*runtimeidentity.Store) *traceIdentityStore
		want  error
	}{
		{
			name: "read-error",
			trace: func(store *runtimeidentity.Store) *traceIdentityStore {
				return &traceIdentityStore{store: store, claimErr: writeErr, readErrAt: 2, readErr: readErr}
			},
			want: readErr,
		},
		{
			name: "changed-sandwich",
			trace: func(store *runtimeidentity.Store) *traceIdentityStore {
				return &traceIdentityStore{store: store, claimErr: writeErr, mutateOnHistory: true}
			},
			want: ErrInspectionIncoherent,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			identity := runtimeidentity.NewStore()
			if err := identity.CreateRuntimeInstance(11, 22, "instance-a"); err != nil {
				t.Fatal(err)
			}
			outcome, gotErr := executeContinuation(t, tc.trace(identity), 1)
			if outcome != runtimelaunchflow.StartClaimBlocked {
				t.Fatalf("outcome = %v", outcome)
			}
			if tc.name == "read-error" {
				if gotErr != tc.want {
					t.Fatalf("error = %v", gotErr)
				}
			} else if !errors.Is(gotErr, tc.want) {
				t.Fatalf("error = %v, want %v", gotErr, tc.want)
			}
		})
	}
}

func executeContinuation(
	t *testing.T,
	identity IdentityStore,
	revision runtimeorchestrationbinding.AggregateRevision,
) (runtimelaunchflow.StartClaimOutcome, error) {
	t.Helper()
	boundary, _ := runtimecommandidempotency.NewBoundary(runtimecommandidempotency.NewMemoryStorage())
	continuation, _ := New(boundary, identity)
	scope, _ := runtimecommandidempotency.NewScope(
		"domain-a", 11, 22, "instance-a", runtimecommandidempotency.OperationStart,
	)
	intent, _ := runtimecommandidempotency.NewStartIntent(33)
	var gotOutcome runtimelaunchflow.StartClaimOutcome
	var gotErr error
	_, _ = boundary.ExecuteManagedStart(
		context.Background(), scope, "start-a", intent, revision, "generation-a",
		func(context.Context, runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error { return nil },
		func(binding runtimeorchestrationbinding.StartExecutionBinding) (runtimecommandidempotency.TerminalOutcome, error) {
			view, _ := runtimelaunchflow.NewOwnerClaimView(runtimeconfigload.NewLoadRequest(
				11, 22, 33, "instance-a", "attempt-a",
			))
			gotOutcome, gotErr = continuation.AfterOwnerClaim(context.Background(), binding, view)
			if gotErr != nil {
				return runtimecommandidempotency.TerminalOutcome{}, gotErr
			}
			return runtimecommandidempotency.NewTerminalOutcome(
				runtimecommandidempotency.OutcomeSucceeded, "attempt-a",
			)
		},
	)
	return gotOutcome, gotErr
}

func TestAfterOwnerClaimIndeterminateClassifierExactRows(t *testing.T) {
	writeErr := errors.New("indeterminate")
	t.Run("claim-exact-empty-generation-blocked", func(t *testing.T) {
		s := runtimeidentity.NewStore()
		_ = s.CreateRuntimeInstance(11, 22, "instance-a")
		trace := &traceIdentityStore{store: s, claimErr: writeErr, claimCommitThenErr: true}
		out, _ := executeContinuation(t, trace, 1)
		if out != runtimelaunchflow.StartClaimBlocked || !reflect.DeepEqual(trace.calls, []string{"read", "claim", "read", "history", "read"}) {
			t.Fatalf("out=%v calls=%v", out, trace.calls)
		}
	})
	t.Run("claim-different-generation-blocked", func(t *testing.T) {
		s := runtimeidentity.NewStore()
		_ = s.CreateRuntimeInstance(11, 22, "instance-a")
		claim, _ := s.ConditionalClaimLaunchAttempt("instance-a", 1, "attempt-a", 33)
		_, _ = s.ConditionalBindExecutionGeneration("instance-a", claim.Revision(), "attempt-a", "other")
		trace := &traceIdentityStore{store: s, claimErr: writeErr}
		out, _ := executeContinuation(t, trace, 3)
		if out != runtimelaunchflow.StartClaimBlocked || len(trace.calls) != 5 {
			t.Fatalf("out=%v calls=%v", out, trace.calls)
		}
	})
	t.Run("bind-error-same-generation-continues", func(t *testing.T) {
		s := runtimeidentity.NewStore()
		_ = s.CreateRuntimeInstance(11, 22, "instance-a")
		trace := &traceIdentityStore{store: s, bindErr: writeErr, bindCommitThenErr: true}
		out, e := executeContinuation(t, trace, 1)
		if e != nil || out != runtimelaunchflow.StartClaimContinue || !reflect.DeepEqual(trace.calls, []string{"read", "claim", "bind", "read", "history", "read"}) {
			t.Fatalf("out=%v err=%v calls=%v", out, e, trace.calls)
		}
	})
	t.Run("history-read-error-exact", func(t *testing.T) {
		s := runtimeidentity.NewStore()
		_ = s.CreateRuntimeInstance(11, 22, "instance-a")
		historyErr := errors.New("history unavailable")
		trace := &traceIdentityStore{store: s, claimErr: writeErr, historyErr: historyErr}
		out, e := executeContinuation(t, trace, 1)
		if out != runtimelaunchflow.StartClaimBlocked || e != historyErr {
			t.Fatalf("out=%v err=%v", out, e)
		}
	})
	t.Run("b-read-error-exact", func(t *testing.T) {
		s := runtimeidentity.NewStore()
		_ = s.CreateRuntimeInstance(11, 22, "instance-a")
		readErr := errors.New("B unavailable")
		trace := &traceIdentityStore{store: s, claimErr: writeErr, readErrAt: 3, readErr: readErr}
		out, e := executeContinuation(t, trace, 1)
		if out != runtimelaunchflow.StartClaimBlocked || e != readErr {
			t.Fatalf("out=%v err=%v", out, e)
		}
	})
}

func TestAfterOwnerClaimConflictingAttemptFactsBlockWithoutRetry(t *testing.T) {
	for _, mode := range []string{"wrong-attempt", "wrong-version", "non-claimed", "terminal-reuse"} {
		t.Run(mode, func(t *testing.T) {
			s := runtimeidentity.NewStore()
			_ = s.CreateRuntimeInstance(11, 22, "instance-a")
			attempt := runtimeconfigload.LaunchAttemptID("attempt-a")
			version := uint64(33)
			if mode == "wrong-attempt" {
				attempt = "other"
			}
			if mode == "wrong-version" {
				version = 44
			}
			claim, _ := s.ConditionalClaimLaunchAttempt("instance-a", 1, attempt, version)
			revision := claim.Revision()
			if mode == "non-claimed" {
				bind, _ := s.ConditionalBindExecutionGeneration("instance-a", revision, attempt, "generation-a")
				running, _ := s.ConditionalPublishRunning("instance-a", bind.Revision(), attempt)
				revision = running.Revision()
			}
			if mode == "terminal-reuse" {
				terminal, _ := s.ConditionalPublishTerminal("instance-a", revision, attempt, true)
				revision = terminal.Revision()
			}
			trace := &traceIdentityStore{store: s}
			out, _ := executeContinuation(t, trace, runtimeorchestrationbinding.AggregateRevision(revision))
			if out != runtimelaunchflow.StartClaimBlocked || len(trace.calls) != 5 {
				t.Fatalf("out=%v calls=%v", out, trace.calls)
			}
		})
	}
}

func TestAfterOwnerClaimEqualRevisionChangedActiveFactsIsIncoherent(t *testing.T) {
	a := runtimeidentity.NewStore()
	b := runtimeidentity.NewStore()
	_ = a.CreateRuntimeInstance(11, 22, "instance-a")
	_ = b.CreateRuntimeInstance(11, 22, "instance-a")
	_, _ = a.ConditionalClaimLaunchAttempt("instance-a", 1, "attempt-a", 33)
	_, _ = b.ConditionalClaimLaunchAttempt("instance-a", 1, "other", 33)
	trace := &traceIdentityStore{store: a, bStore: b, claimErr: errors.New("indeterminate")}
	outcome, err := executeContinuation(t, trace, 2)
	if outcome != runtimelaunchflow.StartClaimBlocked || !errors.Is(err, ErrInspectionIncoherent) {
		t.Fatalf("outcome=%v error=%v", outcome, err)
	}
}
