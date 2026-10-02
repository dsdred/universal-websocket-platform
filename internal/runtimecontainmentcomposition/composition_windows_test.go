//go:build windows

package runtimecontainmentcomposition

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	runtimeplatform "github.com/dsdred/universal-websocket-platform/internal/runtime"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeactivation"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecommandidempotency"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecontainment"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeexecutionevidence"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeidentity"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle"
	"github.com/dsdred/universal-websocket-platform/internal/runtimemanagement"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/sys/windows"
)

const compositionProofHelper = "UWP_TASK074_COMPOSITION_PROOF_HELPER"

func TestConcreteCompositionGate(t *testing.T) {
	if root := os.Getenv(compositionProofHelper); root != "" {
		exerciseConcreteCompositionGate(t, root)
		return
	}
	root, err := os.MkdirTemp("", "uwp-task074-composition-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	command := exec.Command(os.Args[0], "-test.run=^TestConcreteCompositionGate$", "-test.v")
	command.Env = append(os.Environ(), compositionProofHelper+"="+root)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("composition proof helper: %v\n%s", err, output)
	}
}

func exerciseConcreteCompositionGate(t *testing.T, root string) {
	t.Run("construction rejects foreign and fatal authority", func(t *testing.T) {
		fixture := newConcreteFixture(t, filepath.Join(root, "construction"), nil)
		foreignText, _ := randomCompositionIdentity(t)
		composition, err := New(foreignText, fixture.domain, fixture.containment,
			fixture.target, fixture.versions, fixture.identities, fixture.owner,
			fixture.commands, fixture.invoker, fixture.authorize)
		if composition != nil || err == nil {
			t.Fatalf("New(foreign domain) = %#v/%v, want nil/error", composition, err)
		}

		fatalText, fatalBytes := randomCompositionIdentity(t)
		fatalDomain, err := runtimecontainment.ParseDomain(fatalText)
		if err != nil {
			t.Fatal(err)
		}
		fatalRoot := filepath.Join(root, "fatal")
		config := provisionCompositionLedger(t, fatalRoot, fatalText, fatalBytes)
		if err := os.Remove(filepath.Join(fatalRoot, "runtime-containment", fatalText+".db")); err != nil {
			t.Fatal(err)
		}
		fatal := runtimecontainment.AcquireProcessContainment(fatalDomain, config)
		if !fatal.IsFatalFenced() {
			t.Fatalf("corrupt acquisition = unavailable:%v fatal:%v", fatal.IsUnavailable(), fatal.IsFatalFenced())
		}
		composition, err = New(fatalText, fatalDomain, fatal, fixture.target,
			fixture.versions, fixture.identities, fixture.owner, fixture.commands,
			fixture.invoker, fixture.authorize)
		if composition != nil || err == nil {
			t.Fatalf("New(fatal-fenced) = %#v/%v, want nil/error", composition, err)
		}
	})

	t.Run("entry loss and cancellation precede command admission", func(t *testing.T) {
		for name, invalidate := range map[string]func(*concreteFixture) context.Context{
			"fenced": func(f *concreteFixture) context.Context {
				signalCompositionCapability(t, f.domainText)
				return context.Background()
			},
			"cancelled": func(*concreteFixture) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
		} {
			t.Run(name, func(t *testing.T) {
				fixture := newConcreteFixture(t, filepath.Join(root, "entry-"+name), nil)
				before := fixture.snapshot()
				result, err := fixture.composition.ActivateExact(invalidate(fixture), mustCompositionRequest(t, "entry-"+name))
				if result != (runtimeactivation.Result{}) || err == nil {
					t.Fatalf("ActivateExact(%s) = %#v/%v, want zero/error", name, result, err)
				}
				fixture.assertNoDownstreamWork(t, before)
				assertCompositionCommandAbsent(t, fixture, "entry-"+name)
			})
		}
	})

	t.Run("loss inside policy is caught before inspection", func(t *testing.T) {
		var fixture *concreteFixture
		fixture = newConcreteFixture(t, filepath.Join(root, "policy-loss"), func(context.Context, runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error {
			signalCompositionCapability(t, fixture.domainText)
			return nil
		})
		before := fixture.snapshot()
		result, err := fixture.composition.ActivateExact(context.Background(), mustCompositionRequest(t, "policy-loss"))
		if result != (runtimeactivation.Result{}) || err == nil {
			t.Fatalf("ActivateExact(policy loss) = %#v/%v, want zero/error", result, err)
		}
		before.authorizations++
		fixture.assertNoDownstreamWork(t, before)
		assertCompositionCommandAbsent(t, fixture, "policy-loss")
	})

	t.Run("late loss preserves winning unresolved claim", func(t *testing.T) {
		fixture := newConcreteFixture(t, filepath.Join(root, "provider-loss"), nil)
		fixture.versions.hook = func() { signalCompositionCapability(t, fixture.domainText) }
		result, err := fixture.composition.ActivateExact(context.Background(), mustCompositionRequest(t, "provider-loss"))
		if result.Category() != runtimeactivation.ResultUnresolved || !errors.Is(err, runtimecommandidempotency.ErrIndeterminateExecution) {
			t.Fatalf("ActivateExact(provider loss) = %#v/%v", result, err)
		}
		if fixture.invoker.count() != 0 {
			t.Fatal("provider loss reached lifecycle invoker")
		}
		view, readErr := fixture.identities.Store.ReadRuntimeInstance(fixture.target.RuntimeInstanceID())
		history, historyErr := fixture.identities.Store.ReadLaunchAttemptHistory(fixture.target.RuntimeInstanceID())
		if readErr != nil || historyErr != nil || view.Revision() != 1 || len(history) != 0 {
			t.Fatalf("provider loss mutated DP-014: view=%#v history=%d errors=%v/%v", view, len(history), readErr, historyErr)
		}
		if disposition := inspectCompositionCommand(t, fixture, "provider-loss"); disposition != runtimecommandidempotency.ReplayFirstAdmitted {
			t.Fatalf("provider-loss command disposition = %v, want admitted unresolved", disposition)
		}
	})

	t.Run("late cancellation preserves winning unresolved claim", func(t *testing.T) {
		fixture := newConcreteFixture(t, filepath.Join(root, "provider-cancel"), nil)
		ctx := newCancelAfterErrContext(9)
		result, err := fixture.composition.ActivateExact(ctx, mustCompositionRequest(t, "provider-cancel"))
		if result.Category() != runtimeactivation.ResultUnresolved || !errors.Is(err, runtimecommandidempotency.ErrIndeterminateExecution) {
			t.Fatalf("ActivateExact(provider cancellation) = %#v/%v", result, err)
		}
		if ctx.calls.Load() < 9 || fixture.invoker.count() != 0 {
			t.Fatalf("provider cancellation calls=%d lifecycle=%d", ctx.calls.Load(), fixture.invoker.count())
		}
		view, readErr := fixture.identities.Store.ReadRuntimeInstance(fixture.target.RuntimeInstanceID())
		history, historyErr := fixture.identities.Store.ReadLaunchAttemptHistory(fixture.target.RuntimeInstanceID())
		if readErr != nil || historyErr != nil || view.Revision() != 1 || len(history) != 0 {
			t.Fatalf("provider cancellation mutated DP-014: view=%#v history=%d errors=%v/%v", view, len(history), readErr, historyErr)
		}
		if disposition := inspectCompositionCommand(t, fixture, "provider-cancel"); disposition != runtimecommandidempotency.ReplayFirstAdmitted {
			t.Fatalf("provider-cancel command disposition = %v, want admitted unresolved", disposition)
		}
	})

	t.Run("definitive no-claim does not request a generation", func(t *testing.T) {
		fixture := newConcreteFixture(t, filepath.Join(root, "no-claim"), nil)
		preparation, err := fixture.owner.Owner.PrepareStart(runtimelifecycle.NewStartRequest(
			compositionWorkspace, compositionConfiguration, compositionVersion))
		if err != nil {
			t.Fatal(err)
		}
		attempt := preparation.LoadRequest().LaunchAttemptID()
		claim, err := fixture.identities.Store.ConditionalClaimLaunchAttempt(
			fixture.target.RuntimeInstanceID(), 1, attempt, compositionVersion)
		if err != nil || !claim.Committed() {
			t.Fatalf("identity claim = %#v/%v", claim, err)
		}
		fixture.versions.hook = func() { signalCompositionCapability(t, fixture.domainText) }
		result, err := fixture.composition.ActivateExact(context.Background(), mustCompositionRequest(t, "no-claim"))
		if err != nil || result.Category() != runtimeactivation.ResultInProgress || fixture.invoker.count() != 0 {
			t.Fatalf("ActivateExact(no claim) = %#v/%v lifecycle=%d", result, err, fixture.invoker.count())
		}
		assertCompositionCommandAbsent(t, fixture, "no-claim")
	})

	t.Run("primitive and linked winners bind exact current generation", func(t *testing.T) {
		for name, submit := range map[string]func(*Composition, context.Context, runtimeactivation.Request) (runtimeactivation.Result, error){
			"primitive": (*Composition).ActivateExact,
			"linked":    (*Composition).ReplaceExact,
		} {
			t.Run(name, func(t *testing.T) {
				fixture := newConcreteFixture(t, filepath.Join(root, "winner-"+name), nil)
				result, err := submit(fixture.composition, context.Background(), mustCompositionRequest(t, "winner-"+name))
				if err != nil || result.Category() != runtimeactivation.ResultFailed {
					t.Fatalf("submission = %#v/%v, want failed/nil", result, err)
				}
				bindings := fixture.invoker.allBindings()
				if len(bindings) != 1 || bindings[0].ExecutionGeneration() != runtimeorchestrationbinding.ExecutionGeneration(fixture.authority.CurrentGeneration()) {
					t.Fatalf("bindings = %#v, want one exact current generation", bindings)
				}
				_, linked := bindings[0].LinkedExecutionIdentity()
				if linked != (name == "linked") {
					t.Fatalf("linked binding = %v, want %v", linked, name == "linked")
				}
				versionsBefore := fixture.versions.calls.Load()
				replayed, replayErr := submit(fixture.composition, context.Background(), mustCompositionRequest(t, "winner-"+name))
				if replayErr != nil || replayed.Category() != runtimeactivation.ResultFailed || fixture.invoker.count() != 1 || fixture.versions.calls.Load() != versionsBefore {
					t.Fatalf("replay = %#v/%v invocations=%d versions=%d/%d", replayed, replayErr, fixture.invoker.count(), fixture.versions.calls.Load(), versionsBefore)
				}
			})
		}
	})

	t.Run("same authority serves evidence and fencing removes positive results", func(t *testing.T) {
		fixture := newConcreteFixture(t, filepath.Join(root, "evidence"), nil)
		generation := runtimeidentity.ExecutionGeneration(fixture.authority.CurrentGeneration())
		attempt := runtimeconfigload.LaunchAttemptID("attempt-evidence")
		view, err := fixture.identities.Store.ReadRuntimeInstance(fixture.target.RuntimeInstanceID())
		if err != nil {
			t.Fatal(err)
		}
		claim, err := fixture.identities.Store.ConditionalClaimLaunchAttempt(fixture.target.RuntimeInstanceID(), view.Revision(), attempt, compositionVersion)
		if err != nil || !claim.Committed() {
			t.Fatalf("claim = %#v/%v", claim, err)
		}
		bound, err := fixture.identities.Store.ConditionalBindExecutionGeneration(fixture.target.RuntimeInstanceID(), claim.Revision(), attempt, generation)
		if err != nil || !bound.Committed() {
			t.Fatalf("bind = %#v/%v", bound, err)
		}
		outcome := fixture.composition.QueryEvidence(context.Background(), fixture.domain,
			fixture.target.RuntimeInstanceID(), attempt, generation, runtimeexecutionevidence.GenerationStatus).Consume(context.Background())
		if outcome.Kind() != runtimeexecutionevidence.GenerationLive {
			t.Fatalf("same-authority evidence = %v/%v", outcome.Kind(), outcome.Reason())
		}
		foreignText, _ := randomCompositionIdentity(t)
		foreign, err := runtimecontainment.ParseDomain(foreignText)
		if err != nil {
			t.Fatal(err)
		}
		outcome = fixture.composition.QueryEvidence(context.Background(), foreign,
			fixture.target.RuntimeInstanceID(), attempt, generation, runtimeexecutionevidence.GenerationStatus).Consume(context.Background())
		if outcome.Kind() != runtimeexecutionevidence.Unknown || outcome.Reason() != runtimeexecutionevidence.ScopeMismatch {
			t.Fatalf("cross-domain evidence = %v/%v", outcome.Kind(), outcome.Reason())
		}
		signalCompositionCapability(t, fixture.domainText)
		outcome = fixture.composition.QueryEvidence(context.Background(), fixture.domain,
			fixture.target.RuntimeInstanceID(), attempt, generation, runtimeexecutionevidence.GenerationStatus).Consume(context.Background())
		if outcome.Kind() != runtimeexecutionevidence.Unknown || outcome.Reason() != runtimeexecutionevidence.Unavailable || fixture.authority.IsAuthoritative() {
			t.Fatalf("fenced evidence = %v/%v authoritative=%v", outcome.Kind(), outcome.Reason(), fixture.authority.IsAuthoritative())
		}
	})

	t.Run("concurrent submissions preserve one winner and fence blocks later work", func(t *testing.T) {
		fixture := newConcreteFixture(t, filepath.Join(root, "concurrent"), nil)
		const submissions = 16
		errorsOut := make(chan error, submissions)
		var winners atomic.Int32
		var blocked atomic.Int32
		var wait sync.WaitGroup
		for index := 0; index < submissions; index++ {
			wait.Add(1)
			go func(index int) {
				defer wait.Done()
				result, err := fixture.composition.ActivateExact(context.Background(), mustCompositionRequest(t, fmt.Sprintf("concurrent-%d", index)))
				switch {
				case err == nil && result.Category() == runtimeactivation.ResultFailed:
					winners.Add(1)
				case result == (runtimeactivation.Result{}) && errors.Is(err, runtimecommandidempotency.ErrInstanceBlocked):
					blocked.Add(1)
				default:
					errorsOut <- fmt.Errorf("submission %d = %#v/%w", index, result, err)
				}
			}(index)
		}
		wait.Wait()
		close(errorsOut)
		for err := range errorsOut {
			t.Error(err)
		}
		bindings := fixture.invoker.allBindings()
		if winners.Load() != 1 || blocked.Load() != submissions-1 || len(bindings) != 1 {
			t.Fatalf("concurrent admission winners=%d blocked=%d invocations=%d", winners.Load(), blocked.Load(), len(bindings))
		}
		wantGeneration := runtimeorchestrationbinding.ExecutionGeneration(fixture.authority.CurrentGeneration())
		for _, binding := range bindings {
			if binding.ExecutionGeneration() != wantGeneration {
				t.Fatalf("generation = %q, want %q", binding.ExecutionGeneration(), wantGeneration)
			}
		}
		signalCompositionCapability(t, fixture.domainText)
		before := fixture.invoker.count()
		if result, err := fixture.composition.ActivateExact(context.Background(), mustCompositionRequest(t, "after-fence")); result != (runtimeactivation.Result{}) || err == nil || fixture.invoker.count() != before {
			t.Fatalf("post-fence submission = %#v/%v invocations=%d/%d", result, err, fixture.invoker.count(), before)
		}
	})
}

const (
	compositionWorkspace     = uint64(71)
	compositionConfiguration = uint64(72)
	compositionVersion       = uint64(73)
)

type compositionVersionStore struct {
	calls atomic.Int32
	hook  func()
}

func (s *compositionVersionStore) Get(id uint64) (configurationversion.ConfigurationVersion, error) {
	s.calls.Add(1)
	if s.hook != nil {
		s.hook()
	}
	if id != compositionVersion {
		return configurationversion.ConfigurationVersion{}, errors.New("version absent")
	}
	return configurationversion.ConfigurationVersion{ID: id, ConfigurationID: compositionConfiguration, Number: uint32(id), State: configurationversion.Published}, nil
}

type observedCompositionIdentity struct {
	*runtimeidentity.Store
	reads atomic.Int32
}

func (s *observedCompositionIdentity) ReadRuntimeInstance(id runtimeconfigload.RuntimeInstanceID) (runtimeidentity.RuntimeInstanceView, error) {
	s.reads.Add(1)
	return s.Store.ReadRuntimeInstance(id)
}

func (s *observedCompositionIdentity) ReadLaunchAttemptHistory(id runtimeconfigload.RuntimeInstanceID) ([]runtimeidentity.LaunchAttemptRecord, error) {
	s.reads.Add(1)
	return s.Store.ReadLaunchAttemptHistory(id)
}

type observedCompositionOwner struct {
	*runtimelifecycle.Owner
	observes atomic.Int32
	stops    atomic.Int32
}

func (o *observedCompositionOwner) Observe() runtimelifecycle.Observation {
	o.observes.Add(1)
	return o.Owner.Observe()
}

func (o *observedCompositionOwner) StopExpectedAttempt(ctx context.Context, attempt runtimeconfigload.LaunchAttemptID) (runtimelifecycle.StopOutcome, error) {
	o.stops.Add(1)
	return o.Owner.StopExpectedAttempt(ctx, attempt)
}

type recordingCompositionInvoker struct {
	mu       sync.Mutex
	bindings []runtimeorchestrationbinding.StartExecutionBinding
	boundary *runtimecommandidempotency.Boundary
}

func (i *recordingCompositionInvoker) InvokeManagedStart(_ context.Context, _ runtimelifecycle.StartRequest, binding runtimeorchestrationbinding.StartExecutionBinding) (runtimelifecycle.StartOutcome, error) {
	i.mu.Lock()
	i.bindings = append(i.bindings, binding)
	i.mu.Unlock()
	if err := i.boundary.SignalManagedStartNoClaim(binding, runtimeorchestrationbinding.StartNoClaimFailed); err != nil {
		return runtimelifecycle.StartOutcome{}, err
	}
	return runtimelifecycle.StartOutcome{}, runtimelifecycle.ErrAttemptIDSourceFailed
}

func (i *recordingCompositionInvoker) count() int { return len(i.allBindings()) }

func (i *recordingCompositionInvoker) allBindings() []runtimeorchestrationbinding.StartExecutionBinding {
	i.mu.Lock()
	defer i.mu.Unlock()
	return append([]runtimeorchestrationbinding.StartExecutionBinding(nil), i.bindings...)
}

type concreteFixture struct {
	domainText     string
	domain         runtimecontainment.Domain
	containment    runtimecontainment.Result
	authority      *runtimecontainment.ActiveAuthority
	target         runtimemanagement.Target
	versions       *compositionVersionStore
	identities     *observedCompositionIdentity
	owner          *observedCompositionOwner
	commands       *runtimecommandidempotency.Boundary
	invoker        *recordingCompositionInvoker
	authorize      runtimeorchestrationbinding.AuthorizeOrchestration
	authorizations atomic.Int32
	composition    *Composition
}

type fixtureSnapshot struct {
	versions, identityReads, ownerObserves, ownerStops, authorizations int32
	invocations                                                        int
	revision                                                           runtimeidentity.Revision
	history                                                            int
}

type cancelAfterErrContext struct {
	calls    atomic.Int32
	cancelAt int32
	done     chan struct{}
	once     sync.Once
}

func newCancelAfterErrContext(cancelAt int32) *cancelAfterErrContext {
	return &cancelAfterErrContext{cancelAt: cancelAt, done: make(chan struct{})}
}

func (*cancelAfterErrContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *cancelAfterErrContext) Done() <-chan struct{}     { return c.done }
func (c *cancelAfterErrContext) Value(any) any             { return nil }
func (c *cancelAfterErrContext) Err() error {
	if c.calls.Add(1) < c.cancelAt {
		return nil
	}
	c.once.Do(func() { close(c.done) })
	return context.Canceled
}

func newConcreteFixture(t *testing.T, root string, policy runtimeorchestrationbinding.AuthorizeOrchestration) *concreteFixture {
	t.Helper()
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	domainText, domainBytes := randomCompositionIdentity(t)
	domain, err := runtimecontainment.ParseDomain(domainText)
	if err != nil {
		t.Fatal(err)
	}
	containment := runtimecontainment.AcquireProcessContainment(domain, provisionCompositionLedger(t, root, domainText, domainBytes))
	authority, ok := containment.Authority()
	if !ok {
		t.Fatalf("containment acquisition = unavailable:%v fatal:%v", containment.IsUnavailable(), containment.IsFatalFenced())
	}
	instance := runtimeconfigload.RuntimeInstanceID("runtime-" + filepath.Base(root))
	target, err := runtimemanagement.NewTarget(compositionWorkspace, compositionConfiguration, instance)
	if err != nil {
		t.Fatal(err)
	}
	identities := &observedCompositionIdentity{Store: runtimeidentity.NewStore()}
	if err := identities.CreateRuntimeInstance(compositionWorkspace, compositionConfiguration, instance); err != nil {
		t.Fatal(err)
	}
	var attempts atomic.Int32
	resolver, err := secretresolver.NewMemory(nil)
	if err != nil {
		t.Fatal(err)
	}
	ownerValue, err := runtimelifecycle.NewOwner(compositionWorkspace, compositionConfiguration, instance,
		func() (runtimeconfigload.LaunchAttemptID, error) {
			return runtimeconfigload.LaunchAttemptID(fmt.Sprintf("attempt-%d", attempts.Add(1))), nil
		}, &runtimeplatform.DependencyBindings{SecretResolver: resolver})
	if err != nil {
		t.Fatal(err)
	}
	owner := &observedCompositionOwner{Owner: ownerValue}
	commands, err := runtimecommandidempotency.NewBoundary(runtimecommandidempotency.NewMemoryStorage())
	if err != nil {
		t.Fatal(err)
	}
	invoker := &recordingCompositionInvoker{boundary: commands}
	fixture := &concreteFixture{
		domainText: domainText, domain: domain, containment: containment, authority: authority,
		target: target, versions: &compositionVersionStore{}, identities: identities, owner: owner,
		commands: commands, invoker: invoker,
	}
	fixture.authorize = func(ctx context.Context, request runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error {
		fixture.authorizations.Add(1)
		if policy != nil {
			return policy(ctx, request)
		}
		return nil
	}
	fixture.composition, err = New(domainText, domain, containment, target, fixture.versions,
		identities, owner, commands, fixture.invoker, fixture.authorize)
	if err != nil {
		t.Fatal(err)
	}
	return fixture
}

func (f *concreteFixture) snapshot() fixtureSnapshot {
	view, _ := f.identities.Store.ReadRuntimeInstance(f.target.RuntimeInstanceID())
	history, _ := f.identities.Store.ReadLaunchAttemptHistory(f.target.RuntimeInstanceID())
	return fixtureSnapshot{
		versions: f.versions.calls.Load(), identityReads: f.identities.reads.Load(),
		ownerObserves: f.owner.observes.Load(), ownerStops: f.owner.stops.Load(),
		authorizations: f.authorizations.Load(), invocations: f.invoker.count(),
		revision: view.Revision(), history: len(history),
	}
}

func (f *concreteFixture) assertNoDownstreamWork(t *testing.T, before fixtureSnapshot) {
	t.Helper()
	after := f.snapshot()
	if after != before {
		t.Fatalf("downstream work changed: before=%+v after=%+v", before, after)
	}
}

func mustCompositionRequest(t *testing.T, key string) runtimeactivation.Request {
	t.Helper()
	request, err := runtimeactivation.NewRequest(runtimecommandidempotency.CommandKey(key), compositionVersion)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func assertCompositionCommandAbsent(t *testing.T, fixture *concreteFixture, key string) {
	t.Helper()
	if disposition := inspectCompositionCommand(t, fixture, key); disposition != runtimecommandidempotency.ReplayFirstNoClaim {
		t.Fatalf("command %q disposition = %v, want no-claim absence", key, disposition)
	}
}

func inspectCompositionCommand(t *testing.T, fixture *concreteFixture, key string) runtimecommandidempotency.ReplayFirstDisposition {
	t.Helper()
	scope, err := runtimecommandidempotency.NewScope(fixture.domainText, compositionWorkspace, compositionConfiguration,
		fixture.target.RuntimeInstanceID(), runtimecommandidempotency.OperationStart)
	if err != nil {
		t.Fatal(err)
	}
	intent, err := runtimecommandidempotency.NewStartIntent(compositionVersion)
	if err != nil {
		t.Fatal(err)
	}
	_, disposition, err := fixture.commands.ExecuteReplayFirstManagedStart(context.Background(), scope,
		runtimecommandidempotency.CommandKey(key), intent,
		func(context.Context, runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error { return nil },
		func(context.Context) (runtimecommandidempotency.AbsentCandidate, error) {
			return runtimecommandidempotency.NewNoClaimCandidate(), nil
		},
		func(context.Context, runtimecommandidempotency.AbsentCandidate) (runtimecommandidempotency.CandidateRevalidation, error) {
			return runtimecommandidempotency.CandidateUnresolved, nil
		},
		func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
			panic("generation provider called by read-only probe")
		},
		func(runtimeorchestrationbinding.StartExecutionBinding) (runtimecommandidempotency.TerminalOutcome, error) {
			panic("lifecycle called by read-only probe")
		})
	if err != nil {
		t.Fatal(err)
	}
	return disposition
}

func provisionCompositionLedger(t *testing.T, root, domainText string, domainBytes []byte) runtimecontainment.LocalConfig {
	t.Helper()
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	canonicalRoot, physicalRoot := inspectCompositionWindowsPath(t, root)
	directory := filepath.Join(root, "runtime-containment")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, domainText+".db")
	db, err := bolt.Open(path, 0o600, &bolt.Options{NoSync: false, NoGrowSync: false})
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	_, ledgerIdentity := inspectCompositionWindowsFile(t, file)
	_ = file.Close()
	storageAuthorityText, storageAuthority := randomCompositionIdentity(t)
	err = db.Update(func(tx *bolt.Tx) error {
		anchor, err := tx.CreateBucket([]byte("anchor-v1"))
		if err != nil {
			return err
		}
		if _, err = tx.CreateBucket([]byte("generations-v1")); err != nil {
			return err
		}
		if _, err = tx.CreateBucket([]byte("state-v1")); err != nil {
			return err
		}
		values := map[string][]byte{
			"domain": domainBytes, "canonical-root": []byte(canonicalRoot),
			"physical-root": []byte(physicalRoot), "storage-authority": storageAuthority,
			"ledger-file-identity": []byte(ledgerIdentity),
		}
		for key, value := range values {
			if err := anchor.Put([]byte(key), value); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	return runtimecontainment.LocalConfig{
		ExpectedDomain: domainText, RootDir: root, ExpectedCanonicalRoot: canonicalRoot,
		ExpectedPhysicalRootIdentity: physicalRoot, StorageAuthorityID: storageAuthorityText,
	}
}

func randomCompositionIdentity(t *testing.T) (string, []byte) {
	t.Helper()
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		t.Fatal(err)
	}
	value[0] |= 1
	return hex.EncodeToString(value), value
}

func signalCompositionCapability(t *testing.T, domainText string) {
	t.Helper()
	name, err := windows.UTF16PtrFromString(`Global\uwp-runtime-containment-` + domainText)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := windows.OpenEvent(windows.EVENT_MODIFY_STATE, false, name)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(handle)
	if err := windows.SetEvent(handle); err != nil {
		t.Fatal(err)
	}
}

func inspectCompositionWindowsPath(t *testing.T, path string) (string, string) {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := windows.CreateFile(windows.StringToUTF16Ptr(abs), windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(handle)
	return inspectCompositionWindowsHandle(t, handle)
}

func inspectCompositionWindowsFile(t *testing.T, file *os.File) (string, string) {
	t.Helper()
	return inspectCompositionWindowsHandle(t, windows.Handle(file.Fd()))
}

func inspectCompositionWindowsHandle(t *testing.T, handle windows.Handle) (string, string) {
	t.Helper()
	buffer := make([]uint16, windows.MAX_PATH+1)
	n, err := windows.GetFinalPathNameByHandle(handle, &buffer[0], uint32(len(buffer)), 0)
	if err != nil {
		t.Fatal(err)
	}
	if n >= uint32(len(buffer)) {
		buffer = make([]uint16, n+1)
		n, err = windows.GetFinalPathNameByHandle(handle, &buffer[0], uint32(len(buffer)), 0)
		if err != nil {
			t.Fatal(err)
		}
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		t.Fatal(err)
	}
	if info.VolumeSerialNumber == 0 || (info.FileIndexHigh == 0 && info.FileIndexLow == 0) {
		t.Fatal("Windows path identity is zero")
	}
	canonical := strings.ToLower(windows.UTF16ToString(buffer[:n]))
	identity := strings.ToLower(fmt.Sprintf("%08x:%08x%08x", info.VolumeSerialNumber, info.FileIndexHigh, info.FileIndexLow))
	return canonical, identity
}
