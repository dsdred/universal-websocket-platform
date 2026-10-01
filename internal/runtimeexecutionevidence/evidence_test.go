package runtimeexecutionevidence

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecontainment"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeidentity"
)

const (
	testInstance   runtimeconfigload.RuntimeInstanceID = "ri-evidence"
	testAttempt    runtimeconfigload.LaunchAttemptID   = "la-evidence"
	testGeneration runtimeidentity.ExecutionGeneration = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

func TestFullTupleQuestionProjection(t *testing.T) {
	domain := testDomain(t, '1')

	t.Run("current", func(t *testing.T) {
		store := boundStore(t)
		composer := scriptedComposer(domain, store, generationRead{kind: runtimecontainment.EvidenceGenerationLive})
		assertQuery(t, composer, domain, GenerationStatus, GenerationLive, "")
		assertQuery(t, composer, domain, CoveredResourceAbsence, Unknown, Absent)
		assertQuery(t, composer, domain, ShutdownCompletion, Unknown, Absent)
	})

	t.Run("terminated-owner-shutdown", func(t *testing.T) {
		store := ownerShutdownStore(t)
		composer := scriptedComposer(domain, store, generationRead{kind: runtimecontainment.EvidenceGenerationTerminated})
		assertQuery(t, composer, domain, GenerationStatus, GenerationTerminated, "")
		assertQuery(t, composer, domain, CoveredResourceAbsence, CoveredResourcesAbsent, "")
		assertQuery(t, composer, domain, ShutdownCompletion, HostShutdownCompleted, "")
	})

	t.Run("no-host-and-recovery-never-upgrade", func(t *testing.T) {
		for name, store := range map[string]*runtimeidentity.Store{
			"no-host":  noHostStore(t),
			"recovery": recoveryStore(t),
		} {
			t.Run(name, func(t *testing.T) {
				composer := scriptedComposer(domain, store, generationRead{kind: runtimecontainment.EvidenceGenerationTerminated})
				assertQuery(t, composer, domain, ShutdownCompletion, Unknown, Absent)
			})
		}
	})
}

func TestQueryFailsClosedForTupleAndBindingFaults(t *testing.T) {
	domain := testDomain(t, '2')
	foreignDomain := testDomain(t, '3')
	positive := generationRead{kind: runtimecontainment.EvidenceGenerationLive}

	tests := []struct {
		name       string
		store      *runtimeidentity.Store
		domain     runtimecontainment.Domain
		instance   runtimeconfigload.RuntimeInstanceID
		attempt    runtimeconfigload.LaunchAttemptID
		generation runtimeidentity.ExecutionGeneration
		question   Question
		want       UnknownReason
	}{
		{name: "zero-domain", store: boundStore(t), instance: testInstance, attempt: testAttempt, generation: testGeneration, question: GenerationStatus, want: ScopeMismatch},
		{name: "foreign-domain", store: boundStore(t), domain: foreignDomain, instance: testInstance, attempt: testAttempt, generation: testGeneration, question: GenerationStatus, want: ScopeMismatch},
		{name: "missing-instance", store: boundStore(t), domain: domain, instance: "missing", attempt: testAttempt, generation: testGeneration, question: GenerationStatus, want: ScopeMismatch},
		{name: "missing-attempt", store: boundStore(t), domain: domain, instance: testInstance, attempt: "missing", generation: testGeneration, question: GenerationStatus, want: ScopeMismatch},
		{name: "missing-binding", store: unboundStore(t), domain: domain, instance: testInstance, attempt: testAttempt, generation: testGeneration, question: GenerationStatus, want: ScopeMismatch},
		{name: "foreign-generation", store: boundStore(t), domain: domain, instance: testInstance, attempt: testAttempt, generation: "other-generation", question: GenerationStatus, want: ScopeMismatch},
		{name: "unknown-question", store: boundStore(t), domain: domain, instance: testInstance, attempt: testAttempt, generation: testGeneration, question: Question(99), want: ScopeMismatch},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			composer := scriptedComposer(domain, test.store, positive)
			got := composer.Query(context.Background(), test.domain, test.instance, test.attempt, test.generation, test.question).Consume(context.Background())
			assertResult(t, got, Unknown, test.want)
		})
	}

	t.Run("nil-composer", func(t *testing.T) {
		var composer *Composer
		got := composer.Query(context.Background(), domain, testInstance, testAttempt, testGeneration, GenerationStatus).Consume(context.Background())
		assertResult(t, got, Unknown, ScopeMismatch)
	})
	t.Run("nil-identity-reader", func(t *testing.T) {
		composer := scriptedComposer(domain, nil, positive)
		got := composer.Query(context.Background(), domain, testInstance, testAttempt, testGeneration, GenerationStatus).Consume(context.Background())
		assertResult(t, got, Unknown, ScopeMismatch)
	})
}

func TestQueryAndConsumeFreshness(t *testing.T) {
	domain := testDomain(t, '4')

	t.Run("aggregate-change", func(t *testing.T) {
		store := boundStore(t)
		composer := scriptedComposer(domain, store, generationRead{kind: runtimecontainment.EvidenceGenerationLive})
		handle := composer.Query(context.Background(), domain, testInstance, testAttempt, testGeneration, GenerationStatus)
		view, err := store.ReadRuntimeInstance(testInstance)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.ConditionalPublishRunning(testInstance, view.Revision(), testAttempt); err != nil {
			t.Fatal(err)
		}
		assertResult(t, handle.Consume(context.Background()), Unknown, Stale)
		assertResult(t, handle.Consume(context.Background()), Unknown, Stale)
	})

	t.Run("aggregate-change-during-query", func(t *testing.T) {
		store := boundStore(t)
		composer := &Composer{
			domain:     domain,
			identities: store,
			readGeneration: func(context.Context, runtimecontainment.Domain, string) generationRead {
				view, err := store.ReadRuntimeInstance(testInstance)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := store.ConditionalPublishRunning(testInstance, view.Revision(), testAttempt); err != nil {
					t.Fatal(err)
				}
				return generationRead{kind: runtimecontainment.EvidenceGenerationLive}
			},
		}
		handle := composer.Query(context.Background(), domain, testInstance, testAttempt, testGeneration, GenerationStatus)
		assertResult(t, handle.Consume(context.Background()), Unknown, Stale)
	})

	t.Run("two-valid-reads-disagree", func(t *testing.T) {
		store := boundStore(t)
		composer := sequenceComposer(domain, store,
			generationRead{kind: runtimecontainment.EvidenceGenerationLive},
			generationRead{kind: runtimecontainment.EvidenceGenerationTerminated},
		)
		handle := composer.Query(context.Background(), domain, testInstance, testAttempt, testGeneration, GenerationStatus)
		assertResult(t, handle.Consume(context.Background()), Unknown, Contradictory)
	})

	t.Run("fresh-authority-fault-retains-reason", func(t *testing.T) {
		store := boundStore(t)
		composer := sequenceComposer(domain, store,
			generationRead{kind: runtimecontainment.EvidenceGenerationLive},
			generationRead{kind: runtimecontainment.EvidenceUnknown, reason: runtimecontainment.UnknownUnavailable},
		)
		handle := composer.Query(context.Background(), domain, testInstance, testAttempt, testGeneration, GenerationStatus)
		assertResult(t, handle.Consume(context.Background()), Unknown, Unavailable)
	})

	t.Run("fatal-fault-precedes-cancellation", func(t *testing.T) {
		store := boundStore(t)
		composer := sequenceComposer(domain, store,
			generationRead{kind: runtimecontainment.EvidenceGenerationLive},
			generationRead{kind: runtimecontainment.EvidenceUnknown, reason: runtimecontainment.UnknownUnavailable},
		)
		handle := composer.Query(context.Background(), domain, testInstance, testAttempt, testGeneration, GenerationStatus)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		assertResult(t, handle.Consume(ctx), Unknown, Unavailable)
	})
}

func TestGenerationUnknownReasonsArePreservedEndToEnd(t *testing.T) {
	domain := testDomain(t, '9')
	store := boundStore(t)
	tests := []struct {
		name string
		from runtimecontainment.UnknownReason
		want UnknownReason
	}{
		{name: "unavailable", from: runtimecontainment.UnknownUnavailable, want: Unavailable},
		{name: "scope-mismatch", from: runtimecontainment.UnknownScopeMismatch, want: ScopeMismatch},
		{name: "contradictory", from: runtimecontainment.UnknownContradictory, want: Contradictory},
		{name: "indeterminate", from: runtimecontainment.UnknownIndeterminate, want: Indeterminate},
		{name: "cancelled", from: runtimecontainment.UnknownCancelled, want: Cancelled},
		{name: "unsupported-topology", from: runtimecontainment.UnknownUnsupportedTopology, want: UnsupportedTopology},
		{name: "guarantee-not-declared", from: runtimecontainment.UnknownGuaranteeNotDeclared, want: GuaranteeNotDeclared},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			composer := scriptedComposer(domain, store, generationRead{kind: runtimecontainment.EvidenceUnknown, reason: test.from})
			handle := composer.Query(context.Background(), domain, testInstance, testAttempt, testGeneration, GenerationStatus)
			assertResult(t, handle.Consume(context.Background()), Unknown, test.want)
			assertResult(t, handle.Consume(context.Background()), Unknown, Stale)
		})
	}
}

func TestContradictoryAttemptHistoryFailsClosed(t *testing.T) {
	domain := testDomain(t, 'b')
	store := boundStore(t)
	reader := duplicateHistoryReader{Store: store}
	composer := scriptedComposer(domain, reader, generationRead{kind: runtimecontainment.EvidenceGenerationLive})
	handle := composer.Query(context.Background(), domain, testInstance, testAttempt, testGeneration, GenerationStatus)
	assertResult(t, handle.Consume(context.Background()), Unknown, Contradictory)
}

func TestUnknownIdentityReaderErrorIsIndeterminate(t *testing.T) {
	domain := testDomain(t, 'c')
	reader := errorIdentityReader{err: errors.New("unclassified identity read failure")}
	composer := scriptedComposer(domain, reader, generationRead{kind: runtimecontainment.EvidenceGenerationLive})
	handle := composer.Query(context.Background(), domain, testInstance, testAttempt, testGeneration, GenerationStatus)
	assertResult(t, handle.Consume(context.Background()), Unknown, Indeterminate)
}

func TestUnknownReadIsTerminalAndNotUpgraded(t *testing.T) {
	domain := testDomain(t, '5')
	store := boundStore(t)
	var calls atomic.Int32
	composer := &Composer{
		domain:     domain,
		identities: store,
		readGeneration: func(context.Context, runtimecontainment.Domain, string) generationRead {
			if calls.Add(1) == 1 {
				return generationRead{kind: runtimecontainment.EvidenceUnknown, reason: runtimecontainment.UnknownCancelled}
			}
			return generationRead{kind: runtimecontainment.EvidenceGenerationLive}
		},
	}
	handle := composer.Query(context.Background(), domain, testInstance, testAttempt, testGeneration, GenerationStatus)
	assertResult(t, handle.Consume(context.Background()), Unknown, Cancelled)
	if got := calls.Load(); got != 1 {
		t.Fatalf("generation reads = %d, want 1; Unknown must not be upgraded", got)
	}
}

func TestCopiedAndConcurrentConsumeHasExactlyOneWinner(t *testing.T) {
	domain := testDomain(t, '6')
	store := ownerShutdownStore(t)
	composer := scriptedComposer(domain, store, generationRead{kind: runtimecontainment.EvidenceGenerationTerminated})
	handle := composer.Query(context.Background(), domain, testInstance, testAttempt, testGeneration, ShutdownCompletion)
	copied := handle

	const consumers = 32
	results := make(chan Outcome, consumers)
	var wg sync.WaitGroup
	for i := 0; i < consumers; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			if index%2 == 0 {
				results <- handle.Consume(context.Background())
				return
			}
			results <- copied.Consume(context.Background())
		}(i)
	}
	wg.Wait()
	close(results)

	winners, stale := 0, 0
	for result := range results {
		switch {
		case result.Kind() == HostShutdownCompleted:
			winners++
		case result.Kind() == Unknown && result.Reason() == Stale:
			stale++
		default:
			t.Fatalf("unexpected result: kind=%v reason=%v", result.Kind(), result.Reason())
		}
	}
	if winners != 1 || stale != consumers-1 {
		t.Fatalf("winners/stale = %d/%d, want 1/%d", winners, stale, consumers-1)
	}
}

func TestIndependentConcurrentQueriesAndNoSourceMutation(t *testing.T) {
	domain := testDomain(t, '7')
	store := ownerShutdownStore(t)
	before, err := store.ReadRuntimeInstance(testInstance)
	if err != nil {
		t.Fatal(err)
	}
	composer := scriptedComposer(domain, store, generationRead{kind: runtimecontainment.EvidenceGenerationTerminated})

	const queries = 24
	results := make(chan Outcome, queries)
	var wg sync.WaitGroup
	for i := 0; i < queries; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handle := composer.Query(context.Background(), domain, testInstance, testAttempt, testGeneration, CoveredResourceAbsence)
			results <- handle.Consume(context.Background())
		}()
	}
	wg.Wait()
	close(results)
	for result := range results {
		assertResult(t, result, CoveredResourcesAbsent, "")
	}
	after, err := store.ReadRuntimeInstance(testInstance)
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatalf("source aggregate mutated: before=%#v after=%#v", before, after)
	}
}

func TestClosedReasonMappings(t *testing.T) {
	identityCases := map[error]UnknownReason{
		runtimeidentity.ErrInvalidIdentity:             ScopeMismatch,
		runtimeidentity.ErrInstanceNotFound:            ScopeMismatch,
		runtimeidentity.ErrAttemptNotFound:             ScopeMismatch,
		runtimeidentity.ErrExecutionGenerationNotBound: ScopeMismatch,
		runtimeidentity.ErrStaleRevision:               Stale,
		runtimeidentity.ErrIncoherentAttemptSnapshot:   Contradictory,
		errors.New("unknown identity reader error"):    Indeterminate,
	}
	for err, want := range identityCases {
		if got := mapIdentityError(err); got != want {
			t.Fatalf("mapIdentityError(%v) = %v, want %v", err, got, want)
		}
	}

	containmentCases := map[runtimecontainment.UnknownReason]UnknownReason{
		runtimecontainment.UnknownUnavailable:          Unavailable,
		runtimecontainment.UnknownScopeMismatch:        ScopeMismatch,
		runtimecontainment.UnknownContradictory:        Contradictory,
		runtimecontainment.UnknownIndeterminate:        Indeterminate,
		runtimecontainment.UnknownCancelled:            Cancelled,
		runtimecontainment.UnknownUnsupportedTopology:  UnsupportedTopology,
		runtimecontainment.UnknownGuaranteeNotDeclared: GuaranteeNotDeclared,
	}
	for reason, want := range containmentCases {
		if got := mapContainmentReason(reason); got != want {
			t.Fatalf("mapContainmentReason(%v) = %v, want %v", reason, got, want)
		}
	}
	if got := mapContainmentReason(runtimecontainment.UnknownReason("future")); got != Indeterminate {
		t.Fatalf("unknown containment reason = %v, want %v", got, Indeterminate)
	}
}

func TestHandleIsOpaqueNonReconstructableAndZeroFailsStale(t *testing.T) {
	typeOfHandle := reflect.TypeOf(EvidenceHandle{})
	for i := 0; i < typeOfHandle.NumField(); i++ {
		if typeOfHandle.Field(i).IsExported() {
			t.Fatalf("Handle field %q is exported", typeOfHandle.Field(i).Name)
		}
	}
	encoded, err := json.Marshal(EvidenceHandle{})
	if err != nil {
		t.Fatal(err)
	}
	var reconstructed EvidenceHandle
	if err := json.Unmarshal(encoded, &reconstructed); err != nil {
		t.Fatal(err)
	}
	assertResult(t, reconstructed.Consume(context.Background()), Unknown, Stale)
	assertResult(t, EvidenceHandle{}.Consume(context.Background()), Unknown, Stale)
}

func TestConcreteConstructorFailsClosedWithoutDeclaredAuthority(t *testing.T) {
	domain := testDomain(t, '8')
	composer := NewComposer(domain, nil, boundStore(t))
	handle := composer.Query(context.Background(), domain, testInstance, testAttempt, testGeneration, GenerationStatus)
	want := GuaranteeNotDeclared
	if runtime.GOOS != "windows" {
		want = UnsupportedTopology
	}
	assertResult(t, handle.Consume(context.Background()), Unknown, want)
}

type duplicateHistoryReader struct{ *runtimeidentity.Store }

func (r duplicateHistoryReader) ReadLaunchAttemptHistory(id runtimeconfigload.RuntimeInstanceID) ([]runtimeidentity.LaunchAttemptRecord, error) {
	history, err := r.Store.ReadLaunchAttemptHistory(id)
	if err != nil || len(history) == 0 {
		return history, err
	}
	return append(history, history[0]), nil
}

type errorIdentityReader struct{ err error }

func (r errorIdentityReader) ReadRuntimeInstance(runtimeconfigload.RuntimeInstanceID) (runtimeidentity.RuntimeInstanceView, error) {
	return runtimeidentity.RuntimeInstanceView{}, r.err
}

func (r errorIdentityReader) ReadLaunchAttemptHistory(runtimeconfigload.RuntimeInstanceID) ([]runtimeidentity.LaunchAttemptRecord, error) {
	return nil, r.err
}

func assertQuery(t *testing.T, composer *Composer, domain runtimecontainment.Domain, question Question, kind OutcomeKind, reason UnknownReason) {
	t.Helper()
	handle := composer.Query(context.Background(), domain, testInstance, testAttempt, testGeneration, question)
	assertResult(t, handle.Consume(context.Background()), kind, reason)
	assertResult(t, handle.Consume(context.Background()), Unknown, Stale)
}

func assertResult(t *testing.T, got Outcome, kind OutcomeKind, reason UnknownReason) {
	t.Helper()
	if got.Kind() != kind || got.Reason() != reason {
		t.Fatalf("result = (%v, %v), want (%v, %v)", got.Kind(), got.Reason(), kind, reason)
	}
}

func scriptedComposer(domain runtimecontainment.Domain, identities identityReader, read generationRead) *Composer {
	return &Composer{
		domain:     domain,
		identities: identities,
		readGeneration: func(context.Context, runtimecontainment.Domain, string) generationRead {
			return read
		},
	}
}

func sequenceComposer(domain runtimecontainment.Domain, identities identityReader, reads ...generationRead) *Composer {
	var mu sync.Mutex
	next := 0
	return &Composer{
		domain:     domain,
		identities: identities,
		readGeneration: func(context.Context, runtimecontainment.Domain, string) generationRead {
			mu.Lock()
			defer mu.Unlock()
			index := next
			if index >= len(reads) {
				index = len(reads) - 1
			} else {
				next++
			}
			return reads[index]
		},
	}
}

func testDomain(t *testing.T, digit byte) runtimecontainment.Domain {
	t.Helper()
	value := make([]byte, 64)
	for i := range value {
		value[i] = digit
	}
	domain, err := runtimecontainment.ParseDomain(string(value))
	if err != nil {
		t.Fatal(err)
	}
	return domain
}

func unboundStore(t *testing.T) *runtimeidentity.Store {
	t.Helper()
	store := runtimeidentity.NewStore()
	if err := store.CreateRuntimeInstance(11, 22, testInstance); err != nil {
		t.Fatal(err)
	}
	view, err := store.ReadRuntimeInstance(testInstance)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConditionalClaimLaunchAttempt(testInstance, view.Revision(), testAttempt, 33); err != nil {
		t.Fatal(err)
	}
	return store
}

func boundStore(t *testing.T) *runtimeidentity.Store {
	t.Helper()
	store := unboundStore(t)
	view, err := store.ReadRuntimeInstance(testInstance)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConditionalBindExecutionGeneration(testInstance, view.Revision(), testAttempt, testGeneration); err != nil {
		t.Fatal(err)
	}
	return store
}

func ownerShutdownStore(t *testing.T) *runtimeidentity.Store {
	t.Helper()
	store := boundStore(t)
	view, _ := store.ReadRuntimeInstance(testInstance)
	running, err := store.ConditionalPublishRunning(testInstance, view.Revision(), testAttempt)
	if err != nil {
		t.Fatal(err)
	}
	stopping, err := store.ConditionalClaimStop(testInstance, running.Revision(), testAttempt)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.OwnerTerminalPublisher().ConditionalPublishShutdownCompleted(testInstance, stopping.Revision(), testAttempt); err != nil {
		t.Fatal(err)
	}
	return store
}

func noHostStore(t *testing.T) *runtimeidentity.Store {
	t.Helper()
	store := boundStore(t)
	view, _ := store.ReadRuntimeInstance(testInstance)
	if _, err := store.OwnerTerminalPublisher().ConditionalPublishNoHostProduced(testInstance, view.Revision(), testAttempt, true); err != nil {
		t.Fatal(err)
	}
	return store
}

func recoveryStore(t *testing.T) *runtimeidentity.Store {
	t.Helper()
	store := boundStore(t)
	view, _ := store.ReadRuntimeInstance(testInstance)
	if _, err := store.RecoveryTerminalPublisher().ConditionalPublishReconciled(testInstance, view.Revision(), testAttempt); err != nil {
		t.Fatal(err)
	}
	return store
}
