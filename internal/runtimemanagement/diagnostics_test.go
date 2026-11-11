package runtimemanagement

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/configurationloader"
	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/runtime"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnostics"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnosticsconsumer"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnosticspull"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

type diagnosticsFixture struct {
	a      DiagnosticsAdapter
	h      []DiagnosticsPrincipal
	target Target
	owner  *runtimelifecycle.Owner
	now    atomic.Int64
}

func diagnosticsFixtureNew(t *testing.T, principals uint8) *diagnosticsFixture {
	t.Helper()
	f := &diagnosticsFixture{target: mustTarget(t, 11, 22, "runtime-core")}
	resolver, err := secretresolver.NewMemory(nil)
	if err != nil {
		t.Fatal(err)
	}
	f.owner, err = runtimelifecycle.NewOwner(11, 22, "runtime-core", func() (runtimeconfigload.LaunchAttemptID, error) { return "attempt-core", nil }, &runtime.DependencyBindings{SecretResolver: resolver})
	if err != nil {
		t.Fatal(err)
	}
	d := mustDirectory(t, func(context.Context, Action, Target, uint64) error {
		t.Error("management authorizer bypass")
		return nil
	}, mustBinding(t, f.target, f.owner, configurationloader.New(&unavailableSource{})))
	f.a, f.h, err = NewDiagnosticsAdapter(d, principals, func(_ context.Context, a DiagnosticsAuthorization) (DiagnosticsSnapshot, error) {
		if a.Action != DiagnosticsConsume {
			t.Error("wrong action")
		}
		return DiagnosticsSnapshot{[]DiagnosticsGrant{{a.Principal, DiagnosticsConsume, a.Target, a.ExpectedAttempt, true}}}, nil
	}, func() time.Duration { return time.Duration(f.now.Load()) })
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *diagnosticsFixture) request(p int) DiagnosticsRequest {
	return DiagnosticsRequest{f.h[p], f.target, "attempt-core", 16}
}
func (f *diagnosticsFixture) run(r DiagnosticsRequest, output func(context.Context, DiagnosticsResponse) error) DiagnosticsStatus {
	if output == nil {
		output = func(context.Context, DiagnosticsResponse) error { return nil }
	}
	return f.a.Execute(context.Background(), func(context.Context) (DiagnosticsRequest, error) { return r, nil }, output)
}

func (f *diagnosticsFixture) start(t *testing.T) {
	t.Helper()
	reservation, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := uint16(reservation.Addr().(*net.TCPAddr).Port)
	if err = reservation.Close(); err != nil {
		t.Fatal(err)
	}
	prep, err := f.owner.PrepareStart(runtimelifecycle.NewStartRequest(11, 22, 33))
	if err != nil {
		t.Fatal(err)
	}
	version := configurationversion.ConfigurationVersion{ID: 33, ConfigurationID: 22, Number: 1, State: configurationversion.Published, Listener: configurationversion.ListenerSettings{Host: "127.0.0.1", Port: port, TLS: configurationversion.TLSSettings{MinVersion: "1.2"}, Timeouts: configurationversion.TimeoutSettings{HandshakeSeconds: 10, ReadSeconds: 30, WriteSeconds: 20, IdleSeconds: 60}}}
	input := runtimeconfigload.NewDetachedLoadResult(prep.LoadRequest(), version, 1, true, "uwp.configuration", 1)
	snapshot, diagnostics := runtimeconfig.NewBuilder().Build(input)
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	outcome, err := f.owner.Start(context.Background(), prep, runtimelifecycle.PreparedSnapshot(snapshot))
	if err != nil || outcome.Kind() != runtimelifecycle.StartRunning {
		t.Fatalf("actual public Owner launch: %v %#v", err, outcome)
	}
	t.Cleanup(func() { _, _ = f.owner.Stop(context.Background()) })
}

func TestDiagnosticsE01E02ExactGrantsAndNoDeniedLookup(t *testing.T) {
	for _, kind := range []string{"empty", "disabled", "wrong-action", "admin", "wildcard", "replay", "foreign", "wrong-principal", "wrong-target", "wrong-attempt", "duplicate", "error", "too-many", "cancel"} {
		t.Run(kind, func(t *testing.T) {
			f := diagnosticsFixtureNew(t, 2)
			r := f.request(0)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			f.a.state.policy = func(_ context.Context, a DiagnosticsAuthorization) (DiagnosticsSnapshot, error) {
				calls++
				if a.DiagnosticsRequest != r || a.Action != DiagnosticsConsume {
					t.Fatal("authorization input changed")
				}
				g := DiagnosticsGrant{r.Principal, DiagnosticsConsume, r.Target, r.ExpectedAttempt, true}
				snapshot := DiagnosticsSnapshot{[]DiagnosticsGrant{g}}
				switch kind {
				case "empty":
					snapshot.Grants = nil
				case "disabled":
					snapshot.Grants[0].Enabled = false
				case "wrong-action":
					snapshot.Grants[0].Action = "observe"
				case "admin":
					snapshot.Grants[0].Action = "admin"
				case "wildcard":
					snapshot.Grants[0].Action = "*"
				case "replay":
					snapshot.Grants[0].Action = "runtime.command.replay"
				case "foreign":
					snapshot.Grants[0].Principal = diagnosticsFixtureNew(t, 1).h[0]
				case "wrong-principal":
					snapshot.Grants[0].Principal = f.h[1]
				case "wrong-target":
					snapshot.Grants[0].Target = mustTarget(t, 12, 22, "runtime-core")
				case "wrong-attempt":
					snapshot.Grants[0].ExpectedAttempt = "old"
				case "duplicate":
					snapshot.Grants = append(snapshot.Grants, g)
				case "too-many":
					snapshot.Grants = make([]DiagnosticsGrant, 2049)
				case "cancel":
					cancel()
				}
				if kind == "error" {
					return snapshot, errors.New("SECRET raw cause")
				}
				return snapshot, nil
			}
			// Poison before Execute: preauthorization lookup/valid/Observe must not run.
			// This fixture is sequential; no concurrent Directory mutation occurs.
			f.a.state.directory = nil
			got := f.a.Execute(ctx, func(context.Context) (DiagnosticsRequest, error) { return r, nil }, func(_ context.Context, response DiagnosticsResponse) error {
				if response.Status != DiagnosticsUnavailable || string(response.Body) != `{"error":"diagnostics_unavailable"}` {
					t.Fatalf("unsafe denial: %v %s", response.Status, response.Body)
				}
				return nil
			})
			if got != DiagnosticsUnavailable || calls != 1 {
				t.Fatalf("got %v, policy calls %d", got, calls)
			}
		})
	}
}

func TestDiagnosticsE02E05InvalidInputsAndOpaqueHandles(t *testing.T) {
	for _, kind := range []string{"zero", "foreign", "budget0", "budget65", "empty-attempt", "invalid-target", "canceled"} {
		t.Run(kind, func(t *testing.T) {
			f := diagnosticsFixtureNew(t, 1)
			r := f.request(0)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			f.a.state.policy = func(context.Context, DiagnosticsAuthorization) (DiagnosticsSnapshot, error) {
				calls++
				panic("must not evaluate")
			}
			want := DiagnosticsInvalid
			switch kind {
			case "zero":
				r.Principal = DiagnosticsPrincipal{}
				want = DiagnosticsUnavailable
			case "foreign":
				r.Principal = diagnosticsFixtureNew(t, 1).h[0]
				want = DiagnosticsUnavailable
			case "budget0":
				r.Budget = 0
			case "budget65":
				r.Budget = 65
			case "empty-attempt":
				r.ExpectedAttempt = ""
			case "invalid-target":
				r.Target = Target{}
			case "canceled":
				cancel()
				want = DiagnosticsUnavailable
			}
			// Poison before Execute: preauthorization lookup/valid/Observe must not run.
			// This fixture is sequential; no concurrent Directory mutation occurs.
			f.a.state.directory = nil
			got := f.a.Execute(ctx, func(context.Context) (DiagnosticsRequest, error) { return r, nil }, func(context.Context, DiagnosticsResponse) error { return nil })
			if got != want || calls != 0 {
				t.Fatalf("got %v want %v calls %d", got, want, calls)
			}
		})
	}
}

func TestDiagnosticsE03AntiEnumerationAndAuthorityFirst(t *testing.T) {
	var bodies [][]byte
	for _, kind := range []string{"absent", "mismatch", "retired", "attempt", "target-busy"} {
		t.Run(kind, func(t *testing.T) {
			f := diagnosticsFixtureNew(t, 1)
			r := f.request(0)
			called := false
			switch kind {
			case "absent":
				r.Target = mustTarget(t, 11, 22, "missing")
			case "mismatch":
				r.Target = mustTarget(t, 99, 22, "runtime-core")
			case "attempt":
				r.ExpectedAttempt = "old-attempt"
			case "target-busy":
				index := f.a.state.targets[r.Target]
				f.a.state.targetActive[index].Store(true)
			}
			policy := f.a.state.policy
			directory := f.a.state.directory
			f.a.state.directory = nil
			f.a.state.policy = func(ctx context.Context, a DiagnosticsAuthorization) (DiagnosticsSnapshot, error) {
				called = true
				// Only dedicated authorization may make the routing capability usable.
				f.a.state.directory = directory
				return policy(ctx, a)
			}
			if got := f.run(r, func(_ context.Context, response DiagnosticsResponse) error {
				bodies = append(bodies, bytes.Clone(response.Body))
				return nil
			}); got != DiagnosticsUnavailable || !called {
				t.Fatalf("got %v policy %v", got, called)
			}
		})
	}
	for _, body := range bodies {
		if string(body) != `{"error":"diagnostics_unavailable"}` {
			t.Fatalf("enumeration: %s", body)
		}
	}
}

func TestDiagnosticsE06ProcessCapBeforePreparationAndCopies(t *testing.T) {
	f := diagnosticsFixtureNew(t, 1)
	copyAdapter := f.a
	entered := make(chan struct{}, 8)
	release := make(chan struct{})
	done := make(chan DiagnosticsStatus, 8)
	for range 8 {
		go func() {
			done <- copyAdapter.Execute(context.Background(), func(ctx context.Context) (DiagnosticsRequest, error) {
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) > 2*time.Second {
					t.Error("missing request deadline")
				}
				entered <- struct{}{}
				<-release
				return DiagnosticsRequest{}, errors.New("invalid")
			}, func(context.Context, DiagnosticsResponse) error { return nil })
		}()
	}
	for range 8 {
		<-entered
	}
	prepared := false
	if got := f.a.Execute(context.Background(), func(context.Context) (DiagnosticsRequest, error) { prepared = true; return DiagnosticsRequest{}, nil }, func(context.Context, DiagnosticsResponse) error { return nil }); got != DiagnosticsLimited || prepared {
		t.Fatalf("cap bypass %v prepared %v", got, prepared)
	}
	close(release)
	for range 8 {
		if got := <-done; got != DiagnosticsInvalid {
			t.Fatal(got)
		}
	}
	for i := range f.a.state.calls {
		if f.a.state.calls[i].Load() != nil {
			t.Fatal("process slot retained")
		}
	}
}

func TestDiagnosticsE06E13SlotsThroughOutputAndPanicCleanup(t *testing.T) {
	f := diagnosticsFixtureNew(t, 2)
	f.start(t)
	entered := make(chan struct{})
	release := make(chan struct{})
	done := make(chan DiagnosticsStatus, 1)
	var borrowed []byte
	go func() {
		done <- f.run(f.request(0), func(_ context.Context, response DiagnosticsResponse) error {
			if response.Status != DiagnosticsOK {
				t.Errorf("first poll %v", response.Status)
			}
			borrowed = response.Body
			close(entered)
			<-release
			return errors.New("SECRET write failure")
		})
	}()
	<-entered
	if got := f.run(f.request(0), nil); got != DiagnosticsLimited {
		t.Fatalf("principal slot released early: %v", got)
	}
	if got := f.run(f.request(1), nil); got != DiagnosticsUnavailable {
		t.Fatalf("Target slot released early: %v", got)
	}
	// Real Owner Stop must complete while an admitted output callback is blocked.
	if outcome, err := f.owner.Stop(context.Background()); err != nil || outcome.Kind() != runtimelifecycle.StopStopped {
		t.Fatalf("Stop joined output: %v %#v", err, outcome)
	}
	close(release)
	if got := <-done; got != DiagnosticsUnavailable {
		t.Fatal(got)
	}
	for _, b := range borrowed {
		if b != 0 {
			t.Fatal("borrowed buffer retained")
		}
	}
	if f.a.state.principalActive[0].Load() || f.a.state.targetActive[0].Load() {
		t.Fatal("slots retained")
	}
	for _, where := range []string{"prepare", "policy", "output"} {
		t.Run(where, func(t *testing.T) {
			f := diagnosticsFixtureNew(t, 1)
			r := f.request(0)
			if where == "policy" {
				f.a.state.policy = func(context.Context, DiagnosticsAuthorization) (DiagnosticsSnapshot, error) { panic("policy") }
			}
			func() {
				defer func() {
					if recover() == nil {
						t.Error("panic missing")
					}
				}()
				f.a.Execute(context.Background(), func(context.Context) (DiagnosticsRequest, error) {
					if where == "prepare" {
						panic("prepare")
					}
					return r, nil
				}, func(context.Context, DiagnosticsResponse) error { panic("output") })
			}()
			if f.a.state.principalActive[0].Load() || f.a.state.targetActive[0].Load() {
				t.Fatal("panic retained slots")
			}
			for i := range f.a.state.calls {
				if f.a.state.calls[i].Load() != nil {
					t.Fatal("panic retained process")
				}
			}
		})
	}
}

func TestDiagnosticsE07RefillRemainderSaturationBackwardOverflow(t *testing.T) {
	b := diagnosticsBucket{0, 0}
	for _, step := range []struct {
		now    time.Duration
		tokens uint8
		at     time.Duration
	}{{124 * time.Millisecond, 0, 0}, {126 * time.Millisecond, 1, 125 * time.Millisecond}, {249 * time.Millisecond, 1, 125 * time.Millisecond}, {250 * time.Millisecond, 2, 250 * time.Millisecond}, {10 * time.Second, 8, 10 * time.Second}, {10001 * time.Millisecond, 8, 10 * time.Second}} {
		var ok bool
		b, ok = diagnosticsRefill(b, step.now, 125*time.Millisecond, 8)
		if !ok || b.tokens != step.tokens || b.at != step.at {
			t.Fatalf("refill %v: %+v", step.now, b)
		}
	}
	f := diagnosticsFixtureNew(t, 1)
	f.now.Store(int64(100 * time.Millisecond))
	if !f.a.state.charge(0, false) {
		t.Fatal("initial charge")
	}
	before := f.a.state.process
	principal := f.a.state.principal[0]
	sample := f.a.state.sampled
	for _, now := range []int64{int64(50 * time.Millisecond), -1, math.MinInt64} {
		f.now.Store(now)
		if f.a.state.charge(0, false) || f.a.state.process != before || f.a.state.principal[0] != principal || f.a.state.sampled != sample {
			t.Fatal("backward/overflow modified state")
		}
	}
	f.now.Store(math.MaxInt64)
	if !f.a.state.charge(0, false) {
		t.Fatal("max nonoverflow timestamp")
	}
}

func TestDiagnosticsE07E08DeniedChargingNoRefundNoAttemptReset(t *testing.T) {
	f := diagnosticsFixtureNew(t, 2)
	f.a.state.policy = func(context.Context, DiagnosticsAuthorization) (DiagnosticsSnapshot, error) {
		return DiagnosticsSnapshot{}, nil
	}
	if f.run(f.request(0), nil) != DiagnosticsUnavailable || f.run(f.request(0), nil) != DiagnosticsUnavailable || f.run(f.request(0), nil) != DiagnosticsLimited {
		t.Fatal("principal burst not charged on deny")
	}
	f.now.Store(int64(499 * time.Millisecond))
	if f.run(f.request(0), nil) != DiagnosticsLimited {
		t.Fatal("early refill")
	}
	f.now.Store(int64(500 * time.Millisecond))
	r := f.request(0)
	r.ExpectedAttempt = "later"
	if f.run(r, nil) != DiagnosticsUnavailable || f.run(r, nil) != DiagnosticsLimited {
		t.Fatal("grant/attempt resets quota")
	}
	if len(f.a.state.targets) != 1 || len(f.h) != 2 {
		t.Fatal("dynamic state growth")
	}
	for i := 0; i < 100; i++ {
		r.Target = mustTarget(t, 11, 22, runtimeconfigload.RuntimeInstanceID(strconv.Itoa(i)))
		f.run(r, nil)
	}
	if len(f.a.state.targets) != 1 {
		t.Fatal("request-created Target slot")
	}
	f.a.state.quota.Lock()
	if f.a.state.charge(0, false) {
		t.Fatal("quota lock contention admitted")
	}
	f.a.state.quota.Unlock()
}

func TestDiagnosticsE11ProjectionInvariantsAndPrecision(t *testing.T) {
	valid := runtimediagnosticspull.PollResult{Count: 1, Receipt: runtimediagnosticsconsumer.DrainResult{ReadAttempts: 1, Observed: 1, SinkSucceeded: 1, Stop: runtimediagnosticsconsumer.Budget}}
	valid.Events[0] = runtimediagnostics.Event{SchemaVersion: 1, Code: runtimediagnostics.ListenerServeFailed, Sequence: math.MaxUint64}
	body := diagnosticsProjection(valid, 1)
	if !bytes.Contains(body, []byte(`"sequence":"18446744073709551615"`)) || len(body) > 16384 {
		t.Fatalf("unsafe precision %s", body)
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(body, &fields) != nil || len(fields) != 4 {
		t.Fatal("unexpected output fields")
	}
	for _, mutate := range []func(*runtimediagnosticspull.PollResult){func(r *runtimediagnosticspull.PollResult) { r.Count = 65 }, func(r *runtimediagnosticspull.PollResult) { r.Receipt.ReadAttempts = 2 }, func(r *runtimediagnosticspull.PollResult) { r.Receipt.Observed = 2 }, func(r *runtimediagnosticspull.PollResult) { r.Receipt.SinkFailed = 1 }, func(r *runtimediagnosticspull.PollResult) { r.Receipt.Discarded = 1 }, func(r *runtimediagnosticspull.PollResult) { r.Receipt.Stop = runtimediagnosticsconsumer.Canceled }, func(r *runtimediagnosticspull.PollResult) { r.Events[0].Sequence = 0 }, func(r *runtimediagnosticspull.PollResult) { r.Events[0].Code = 3 }, func(r *runtimediagnosticspull.PollResult) { r.Events[0].SchemaVersion = 2 }, func(r *runtimediagnosticspull.PollResult) { r.Events[1] = r.Events[0] }} {
		r := valid
		mutate(&r)
		if diagnosticsProjection(r, 1) != nil {
			t.Fatalf("accepted malformed receipt/event %+v", r)
		}
	}
	noItem := runtimediagnosticspull.PollResult{Receipt: runtimediagnosticsconsumer.DrainResult{ReadAttempts: 1, Stop: runtimediagnosticsconsumer.NoItem}}
	if diagnosticsProjection(noItem, 64) == nil {
		t.Fatal("zero-count NoItem rejected")
	}
	noItem.Receipt.ReadAttempts = 0
	if diagnosticsProjection(noItem, 64) != nil {
		t.Fatal("phantom NoItem success")
	}
	for _, budget := range []uint8{0, 65} {
		if diagnosticsProjection(valid, budget) != nil {
			t.Fatal("invalid budget")
		}
	}
}

func TestDiagnosticsE04E14OwnerRetirementAndShutdown(t *testing.T) {
	f := diagnosticsFixtureNew(t, 1)
	f.start(t)
	entered := make(chan struct{})
	release := make(chan struct{})
	done := make(chan DiagnosticsStatus, 1)
	policy := f.a.state.policy
	f.a.state.policy = func(ctx context.Context, a DiagnosticsAuthorization) (DiagnosticsSnapshot, error) {
		snapshot, err := policy(ctx, a)
		close(entered)
		<-release
		return snapshot, err
	}
	go func() { done <- f.run(f.request(0), nil) }()
	<-entered
	if _, err := f.owner.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	close(release)
	if got := <-done; got != DiagnosticsUnavailable {
		t.Fatalf("retired lifetime read: %v", got)
	}
	f = diagnosticsFixtureNew(t, 1)
	entered = make(chan struct{})
	release = make(chan struct{})
	done = make(chan DiagnosticsStatus, 1)
	var activeCtx context.Context
	go func() {
		done <- f.a.Execute(context.Background(), func(ctx context.Context) (DiagnosticsRequest, error) {
			activeCtx = ctx
			close(entered)
			<-release
			return f.request(0), nil
		}, func(context.Context, DiagnosticsResponse) error { return nil })
	}()
	<-entered
	f.a.Shutdown()
	if activeCtx.Err() == nil {
		t.Fatal("shutdown did not cancel admitted context")
	}
	prepared := false
	if got := f.a.Execute(context.Background(), func(context.Context) (DiagnosticsRequest, error) { prepared = true; return f.request(0), nil }, func(context.Context, DiagnosticsResponse) error { return nil }); got != DiagnosticsClosed || prepared {
		t.Fatal("shutdown admission bypass")
	}
	close(release)
	if got := <-done; got != DiagnosticsUnavailable {
		t.Fatal(got)
	}
}

func TestDiagnosticsE08ConstructorBoundsAndSharedCopy(t *testing.T) {
	f := diagnosticsFixtureNew(t, 1)
	clock := func() time.Duration { return 0 }
	policy := f.a.state.policy
	for _, n := range []uint8{0, 33, 255} {
		if _, _, err := NewDiagnosticsAdapter(f.a.state.directory, n, policy, clock); err != ErrDiagnosticsConfiguration {
			t.Fatal("invalid principal bound")
		}
	}
	for _, construct := range []func() (DiagnosticsAdapter, []DiagnosticsPrincipal, error){func() (DiagnosticsAdapter, []DiagnosticsPrincipal, error) {
		return NewDiagnosticsAdapter(nil, 1, policy, clock)
	}, func() (DiagnosticsAdapter, []DiagnosticsPrincipal, error) {
		return NewDiagnosticsAdapter(f.a.state.directory, 1, nil, clock)
	}, func() (DiagnosticsAdapter, []DiagnosticsPrincipal, error) {
		return NewDiagnosticsAdapter(f.a.state.directory, 1, policy, nil)
	}, func() (DiagnosticsAdapter, []DiagnosticsPrincipal, error) {
		return NewDiagnosticsAdapter(f.a.state.directory, 1, policy, func() time.Duration { return -1 })
	}} {
		if _, _, err := construct(); err != ErrDiagnosticsConfiguration {
			t.Fatal("invalid construction allowed")
		}
	}
	copyAdapter := f.a
	if copyAdapter.state != f.a.state {
		t.Fatal("copy budgets independent")
	}
	var evaluations atomic.Int32
	evaluator := f.a.state.policy
	f.a.state.policy = func(ctx context.Context, a DiagnosticsAuthorization) (DiagnosticsSnapshot, error) {
		evaluations.Add(1)
		return evaluator(ctx, a)
	}
	var wg sync.WaitGroup
	wg.Add(16)
	for range 16 {
		go func() {
			defer wg.Done()
			copyAdapter.Execute(context.Background(), func(context.Context) (DiagnosticsRequest, error) { return f.request(0), nil }, func(context.Context, DiagnosticsResponse) error { return nil })
		}()
	}
	wg.Wait()
	if evaluations.Load() > 2 {
		t.Fatal("concurrent copy manufactured principal credit")
	}
	// TryLock refusal need not exhaust credit or promise progress. With contention
	// gone, drain any remaining burst and prove no third policy evaluation.
	for range 3 {
		f.run(f.request(0), nil)
	}
	if evaluations.Load() > 2 || f.a.state.principal[0].tokens != 0 {
		t.Fatal("shared burst boundary changed")
	}
	// Charging precedes principal CAS: contention can spend a token without
	// reaching policy, so exactly two evaluations would incorrectly require progress.
	before := evaluations.Load()
	if got := f.run(f.request(0), nil); got != DiagnosticsLimited || evaluations.Load() != before {
		t.Fatal("exhausted shared burst reached policy")
	}
}

func TestDiagnosticsE04CurrentGrantRevokeIndependentOfLifetime(t *testing.T) {
	f := diagnosticsFixtureNew(t, 1)
	f.start(t)
	var enabled atomic.Bool
	enabled.Store(true)
	captured := make(chan struct{})
	release := make(chan struct{})
	done := make(chan DiagnosticsStatus, 1)
	f.a.state.policy = func(_ context.Context, a DiagnosticsAuthorization) (DiagnosticsSnapshot, error) {
		current := enabled.Load()
		snapshot := DiagnosticsSnapshot{[]DiagnosticsGrant{{a.Principal, DiagnosticsConsume, a.Target, a.ExpectedAttempt, current}}}
		if current {
			close(captured)
			<-release
		}
		return snapshot, nil
	}
	go func() { done <- f.run(f.request(0), nil) }()
	<-captured
	enabled.Store(false)
	close(release)
	if got := <-done; got != DiagnosticsOK {
		t.Fatalf("current evaluation snapshot revoked retroactively: %v", got)
	}
	f.now.Store(int64(time.Second))
	if got := f.run(f.request(0), nil); got != DiagnosticsUnavailable {
		t.Fatalf("subsequent request ignored revoke: %v", got)
	}
	// Grant denial did not retire the Host or steal lifecycle authority.
	if f.owner.Observe().ActualState() != runtimelifecycle.ActualRunning {
		t.Fatal("security revoke changed lifecycle")
	}
}

func TestDiagnosticsE08FixedConstructionTargetsAndDuplicateOwner(t *testing.T) {
	f := diagnosticsFixtureNew(t, 1)
	bindings := make([]Binding, 0, 65)
	for i := 0; i < 65; i++ {
		target := mustTarget(t, 11, 22, runtimeconfigload.RuntimeInstanceID("runtime-"+strconv.Itoa(i)))
		owner := mustOwner(t, 11, 22, target.RuntimeInstanceID())
		bindings = append(bindings, mustBinding(t, target, owner, configurationloader.New(&unavailableSource{})))
	}
	d := mustDirectory(t, func(context.Context, Action, Target, uint64) error { return nil }, bindings...)
	if _, _, err := NewDiagnosticsAdapter(d, 32, f.a.state.policy, func() time.Duration { return 0 }); err != ErrDiagnosticsConfiguration {
		t.Fatal("65 Targets accepted")
	}
	delete(d.scopes, bindings[64].target.RuntimeInstanceID())
	a, h, err := NewDiagnosticsAdapter(d, 32, f.a.state.policy, func() time.Duration { return 0 })
	if err != nil || len(h) != 32 || len(a.state.targets) != 64 {
		t.Fatal("max construction rejected", err)
	}
	first := d.scopes[bindings[0].target.RuntimeInstanceID()]
	second := d.scopes[bindings[1].target.RuntimeInstanceID()]
	second.owner = first.owner
	d.scopes[second.target.RuntimeInstanceID()] = second
	if _, _, err := NewDiagnosticsAdapter(d, 1, f.a.state.policy, func() time.Duration { return 0 }); err != ErrDiagnosticsConfiguration {
		t.Fatal("reused Owner accepted")
	}
}

func TestDiagnosticsE11MaxBatchClosedNestedAllowlist(t *testing.T) {
	result := runtimediagnosticspull.PollResult{Count: 64, Receipt: runtimediagnosticsconsumer.DrainResult{ReadAttempts: 64, Observed: 64, SinkSucceeded: 64, Stop: runtimediagnosticsconsumer.Budget}}
	for i := range result.Events {
		result.Events[i] = runtimediagnostics.Event{SchemaVersion: 1, Code: runtimediagnostics.AuthenticationProviderFailed, Sequence: math.MaxUint64 - uint64(i)}
	}
	body := diagnosticsProjection(result, 64)
	if len(body) == 0 || len(body) > 16384 {
		t.Fatal("max batch bound", len(body))
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	var events []map[string]json.RawMessage
	if err := json.Unmarshal(decoded["events"], &events); err != nil || len(events) != 64 {
		t.Fatal("events shape")
	}
	for _, event := range events {
		if len(event) != 3 || event["schemaVersion"] == nil || event["code"] == nil || event["sequence"] == nil {
			t.Fatal("extra event fields")
		}
	}
	var receipt map[string]json.RawMessage
	if err := json.Unmarshal(decoded["receipt"], &receipt); err != nil || len(receipt) != 7 {
		t.Fatal("receipt shape")
	}
	for _, key := range []string{"readAttempts", "observed", "sinkSucceeded", "sinkFailed", "sinkPanicked", "discarded", "stop"} {
		if receipt[key] == nil {
			t.Fatal("missing receipt field", key)
		}
	}
}

func TestDiagnosticsE14ClosedRoutingGateRefusesWithoutLookup(t *testing.T) {
	f := diagnosticsFixtureNew(t, 1)
	f.a.state.routing.Lock()
	f.a.state.directory = nil
	if got := f.run(f.request(0), nil); got != DiagnosticsUnavailable {
		t.Fatalf("routing contention admitted: %v", got)
	}
	f.a.state.routing.Unlock()
	f.a.Shutdown()
	if got := f.run(f.request(0), nil); got != DiagnosticsClosed {
		t.Fatalf("closed gate admitted: %v", got)
	}
}

func TestDiagnosticsE01E08Snapshot2048AndDistinctTargetBounds(t *testing.T) {
	f := diagnosticsFixtureNew(t, 32)
	grants := make([]DiagnosticsGrant, 0, 2048)
	for p := range f.h {
		for target := 0; target < 64; target++ {
			grants = append(grants, DiagnosticsGrant{f.h[p], DiagnosticsConsume, mustTarget(t, 11, 22, runtimeconfigload.RuntimeInstanceID("grant-"+strconv.Itoa(target))), "attempt", true})
		}
	}
	r := DiagnosticsRequest{f.h[31], grants[len(grants)-1].Target, "attempt", 1}
	if !f.a.state.authorized(DiagnosticsSnapshot{grants}, r) {
		t.Fatal("max valid grant snapshot denied")
	}
	grants = append(grants, DiagnosticsGrant{f.h[0], DiagnosticsConsume, mustTarget(t, 11, 22, "sixty-fifth"), "attempt", true})
	if f.a.state.authorized(DiagnosticsSnapshot{grants[:64]}, r) {
		t.Fatal("unmatched snapshot admitted")
	}
	if f.a.state.authorized(DiagnosticsSnapshot{grants}, r) {
		t.Fatal("2049 grant snapshot admitted")
	}
	distinct := append(append([]DiagnosticsGrant(nil), grants[:64]...), grants[len(grants)-1])
	if f.a.state.authorized(DiagnosticsSnapshot{distinct}, DiagnosticsRequest{f.h[0], distinct[0].Target, "attempt", 1}) {
		t.Fatal("65 distinct Targets admitted")
	}
}

func TestDiagnosticsE05ActualOwnerOneReadAtBudgetsOneAnd64(t *testing.T) {
	f := diagnosticsFixtureNew(t, 1)
	f.start(t)
	for _, budget := range []uint8{1, 64} {
		f.now.Add(int64(time.Second))
		r := f.request(0)
		r.Budget = budget
		got := f.run(r, func(_ context.Context, response DiagnosticsResponse) error {
			var body struct {
				Count   uint8
				Events  []json.RawMessage
				Receipt struct {
					ReadAttempts, Observed, SinkSucceeded uint8
					Stop                                  string
				}
			}
			if json.Unmarshal(response.Body, &body) != nil || body.Count != 0 || len(body.Events) != 0 || body.Receipt.ReadAttempts != 1 || body.Receipt.Observed != 0 || body.Receipt.Stop != "no_item" {
				t.Fatalf("unexpected one-read receipt %s", response.Body)
			}
			return nil
		})
		if got != DiagnosticsOK {
			t.Fatal("actual Owner result", got)
		}
	}
}
