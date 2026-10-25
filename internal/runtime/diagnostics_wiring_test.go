package runtime

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/listener"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnostics"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

// W01 exercises the public production boundary, rather than the old zero-value
// factory seam. Reassignments occur after Bootstrap finishes borrowing the input.
func TestBootstrapDiagnosticsBindingSnapshot(t *testing.T) {
	p1, r1 := runtimediagnostics.New()
	p2, r2 := runtimediagnostics.New()
	request := &BootstrapRequest{
		Snapshot: apiKeySnapshot(t), StartupContext: context.Background(),
		Dependencies: &DependencyBindings{SecretResolver: emptyResolver(t), DiagnosticsProducer: *p1},
	}
	outcome := Bootstrap(request)
	active, ok := outcome.Success()
	if !ok {
		t.Fatalf("Bootstrap failed: %#v", outcome)
	}
	t.Cleanup(func() { _ = active.Stop(context.Background()) })
	host := active.(*DefaultHost)
	request.Dependencies.DiagnosticsProducer = *p2
	*p1 = runtimediagnostics.Producer{}
	assertDiagnosticsEmpty(t, r1)
	assertDiagnosticsEmpty(t, r2)
	assertComposedDiagnostics(t, host.runtimeListener, host.diagnostics)
	diagnosticsRequest(t, request.Snapshot, http.StatusServiceUnavailable)
	assertDiagnosticsEvent(t, r1, runtimediagnostics.AuthenticationProviderFailed, 1)
	assertDiagnosticsEmpty(t, r1)
	assertDiagnosticsEmpty(t, r2)
	if host.diagnostics == *p2 {
		t.Fatal("caller reassignment redirected captured Host capability")
	}
}

// W02 checks both actual component fields without unsafe access or new test APIs,
// then exercises the real HTTP -> Handshake -> API Key Resolver error path.
// Listener emission itself is covered by its package's existing source proofs.
func TestHostDiagnosticsSharedProducerIsolation(t *testing.T) {
	var hosts []*DefaultHost
	var readers []*runtimediagnostics.Reader
	for range 2 {
		producer, reader := runtimediagnostics.New()
		host := bootstrapDiagnosticsHost(t, producer, emptyResolver(t), nil)
		assertComposedDiagnostics(t, host.runtimeListener, *producer)
		hosts = append(hosts, host)
		readers = append(readers, reader)
	}
	for i, host := range hosts {
		diagnosticsRequest(t, host.Snapshot(), http.StatusServiceUnavailable)
		assertDiagnosticsEvent(t, readers[i], runtimediagnostics.AuthenticationProviderFailed, 1)
		for _, reader := range readers {
			assertDiagnosticsEmpty(t, reader)
		}
	}
}

// W03 injects failure only at the existing private composer seam. Production
// composition and real Listener acquisition/rollback still run on each attempt.
func TestHostDiagnosticsRollbackRetryLifetime(t *testing.T) {
	producer, reader := runtimediagnostics.New()
	snapshot := apiKeySnapshot(t)
	host, err := newHostWithDiagnostics(snapshot, emptyResolver(t), nil, nil, nil, *producer)
	if err != nil {
		t.Fatal(err)
	}
	if err := host.Build(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = host.Stop(context.Background()) })
	startCause, rollbackCause := errors.New("start fixture"), errors.New("rollback fixture")
	compose := host.compose
	attempts, rollbacks := 0, 0
	host.compose = func(ctx context.Context, s runtimeconfig.Snapshot, r secretresolver.Resolver, h message.Handler) (runtimeComposition, error) {
		attempts++
		components, err := compose(ctx, s, r, h)
		if err != nil {
			return components, err
		}
		assertComposedDiagnostics(t, components.listener, *producer)
		if attempts == 1 {
			components.listener = diagnosticsRollbackListener{Listener: components.listener,
				startCause: startCause, rollbackCause: rollbackCause, rollbacks: &rollbacks,
				beforeFailure: func() {
					// Real pre-commit request must fail before Authentication.
					diagnosticsRequest(t, snapshot, http.StatusServiceUnavailable)
					assertDiagnosticsEmpty(t, reader)
				},
			}
		}
		return components, nil
	}
	if err := host.Start(context.Background()); !errors.Is(err, startCause) || !errors.Is(err, rollbackCause) {
		t.Fatalf("startup/rollback causes lost: %v", err)
	}
	if rollbacks != 1 || host.Ready() || host.CanAccept() || host.RuntimeContext() != nil {
		t.Fatal("rollback changed lifecycle publication")
	}
	assertDiagnosticsEmpty(t, reader)
	// Seed a historical source fact explicitly: no claim that static failure emits.
	// Its retention/sequence across retry is the mailbox-lifetime proof.
	if !producer.Offer(runtimediagnostics.ListenerServeFailed) {
		t.Fatal("failed to seed retained history")
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if attempts != 2 || !host.Ready() || !host.CanAccept() {
		t.Fatal("retry did not commit exactly one new graph")
	}
	diagnosticsRequest(t, snapshot, http.StatusServiceUnavailable)
	if err := host.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertDiagnosticsEvent(t, reader, runtimediagnostics.ListenerServeFailed, 1)
	assertDiagnosticsEvent(t, reader, runtimediagnostics.AuthenticationProviderFailed, 2)
	assertDiagnosticsEmpty(t, reader)
}

type diagnosticsRollbackListener struct {
	listener.Listener
	startCause, rollbackCause error
	rollbacks                 *int
	beforeFailure             func()
}

func (l diagnosticsRollbackListener) Start(ctx context.Context) error {
	if err := l.Listener.Start(ctx); err != nil {
		return err
	}
	l.beforeFailure()
	return l.startCause
}

func (l diagnosticsRollbackListener) Stop(ctx context.Context) error {
	*l.rollbacks++
	return errors.Join(l.Listener.Stop(ctx), l.rollbackCause)
}

func TestHostDiagnosticsStopIndependence(t *testing.T) {
	for _, withReporter := range []bool{false, true} {
		t.Run(map[bool]string{false: "nil legacy", true: "prompt legacy"}[withReporter], func(t *testing.T) {
			producer, reader := runtimediagnostics.New()
			var reporter func(error)
			if withReporter {
				reporter = func(error) { t.Error("Provider failure should not invoke legacy reporter") }
			}
			host := bootstrapDiagnosticsHost(t, producer, emptyResolver(t), reporter)
			producer.Offer(runtimediagnostics.ListenerServeFailed)
			entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
			go func() {
				defer close(done)
				_, _ = reader.TryRead()
				close(entered)
				<-release // An external sink blocks only after removing a copy.
			}()
			<-entered
			defer func() { close(release); <-done }()
			for range 64 {
				if !producer.Offer(runtimediagnostics.ListenerServeFailed) {
					t.Fatal("prefill failed")
				}
			}
			diagnosticsRequest(t, host.Snapshot(), http.StatusServiceUnavailable)
			stopped := make(chan error, 1)
			go func() { stopped <- host.Stop(context.Background()) }()
			select {
			case err := <-stopped:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("Stop waited for external consumer or full mailbox")
			}
			if host.Ready() || host.CanAccept() || host.RuntimeContext().Err() == nil {
				t.Fatal("Stop lifecycle changed")
			}
			for sequence := uint64(2); sequence <= 65; sequence++ {
				assertDiagnosticsEvent(t, reader, runtimediagnostics.ListenerServeFailed, sequence)
			}
			assertDiagnosticsEmpty(t, reader)
		})
	}
}

func TestHostDiagnosticsCancellationPrecedence(t *testing.T) {
	producer, reader := runtimediagnostics.New()
	entered, canceled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	resolver := runtimeTLSResolverFunc(func(ctx context.Context, _ string) (secretresolver.Secret, error) {
		close(entered)
		<-ctx.Done()
		close(canceled)
		<-release
		return secretresolver.Secret{}, errors.New("late secret dependency failure")
	})
	host := bootstrapDiagnosticsHost(t, producer, resolver, nil)
	requestDone := make(chan struct{})
	go func() {
		defer close(requestDone)
		diagnosticsRequest(t, host.Snapshot(), http.StatusServiceUnavailable)
	}()
	waitForSignal(t, entered, "Provider entry")
	stopDone := make(chan error, 1)
	go func() { stopDone <- host.Stop(context.Background()) }()
	waitForSignal(t, canceled, "Provider boundary cancellation")
	close(release)
	waitForSignal(t, requestDone, "canceled request")
	waitForResult(t, stopDone, "Stop after Provider return")
	assertDiagnosticsEmpty(t, reader)
}

func bootstrapDiagnosticsHost(t *testing.T, producer *runtimediagnostics.Producer, resolver secretresolver.Resolver, reporter func(error)) *DefaultHost {
	t.Helper()
	outcome := Bootstrap(&BootstrapRequest{Snapshot: apiKeySnapshot(t), StartupContext: context.Background(),
		Dependencies: &DependencyBindings{SecretResolver: resolver, TerminalErrorReporter: reporter, DiagnosticsProducer: *producer}})
	host, ok := outcome.Success()
	if !ok {
		t.Fatalf("Bootstrap failed: %#v", outcome)
	}
	t.Cleanup(func() { _ = host.Stop(context.Background()) })
	return host.(*DefaultHost)
}

func diagnosticsRequest(t *testing.T, snapshot runtimeconfig.Snapshot, status int) {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:"+portString(snapshot.Listener().Port)+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("X-API-Key", "private-credential")
	client := &http.Client{Timeout: 3 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		t.Error(err)
		return
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil || response.StatusCode != status || string(body) != "Service Unavailable\n" {
		t.Errorf("response = %d %q, err %v; want generic %d", response.StatusCode, body, err, status)
	}
}

func assertComposedDiagnostics(t *testing.T, composed listener.Listener, producer runtimediagnostics.Producer) {
	t.Helper()
	// Equal reads comparable opaque values; no Interface/unsafe/private mutation.
	l := reflect.ValueOf(composed).Elem()
	if !l.FieldByName("diagnostics").Equal(reflect.ValueOf(producer)) {
		t.Fatal("Listener does not hold Host's Producer value")
	}
	timed := l.FieldByName("handshakeHandler").Elem()
	h := timed.FieldByName("next").Elem().Elem()
	if !h.FieldByName("diagnostics").Equal(reflect.ValueOf(producer)) {
		t.Fatal("Handshake does not hold Host's Producer value")
	}
}

func assertDiagnosticsEmpty(t *testing.T, reader *runtimediagnostics.Reader) {
	t.Helper()
	if event, ok := reader.TryRead(); ok {
		t.Fatalf("unexpected diagnostic: %#v", event)
	}
}

func assertDiagnosticsEvent(t *testing.T, reader *runtimediagnostics.Reader, code runtimediagnostics.Code, sequence uint64) {
	t.Helper()
	event, ok := reader.TryRead()
	if !ok || event != (runtimediagnostics.Event{SchemaVersion: 1, Code: code, Sequence: sequence}) {
		t.Fatalf("diagnostic = %#v, %t; want code %d sequence %d", event, ok, code, sequence)
	}
}
