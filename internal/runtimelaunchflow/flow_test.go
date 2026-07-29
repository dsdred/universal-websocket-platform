package runtimelaunchflow

import (
	"context"
	"errors"
	"fmt"
	"go/build"
	"net"
	"reflect"
	"slices"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/configuration"
	"github.com/dsdred/universal-websocket-platform/internal/configurationloader"
	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/runtime"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

const (
	workspaceID     uint64 = 11
	configurationID uint64 = 22
	versionID       uint64 = 33
)

type sourceFunc func(uint64, uint64, uint64) (configurationloader.SourceObservation, error)

func (f sourceFunc) LoadExact(
	workspaceID uint64,
	configurationID uint64,
	configurationVersionID uint64,
) (configurationloader.SourceObservation, error) {
	return f(workspaceID, configurationID, configurationVersionID)
}

func TestFlowStartsRuntimeThroughExactPipeline(t *testing.T) {
	var sourceCalls atomic.Int32
	port := availablePort(t)
	source := sourceFunc(func(workspace, configuration, version uint64) (configurationloader.SourceObservation, error) {
		sourceCalls.Add(1)
		assertSourceIdentities(t, workspace, configuration, version)
		return validObservation(port), nil
	})
	resolver, err := secretresolver.NewMemory(nil)
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	owner := mustOwner(t, "runtime-success", "attempt-success", &runtime.DependencyBindings{
		SecretResolver: resolver,
	})
	flow := mustFlow(t, owner, configurationloader.New(source))

	outcome, err := flow.Start(
		context.Background(),
		runtimelifecycle.NewStartRequest(workspaceID, configurationID, versionID),
	)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if outcome.Kind() != runtimelifecycle.StartRunning {
		t.Fatalf("Start() kind = %q, want %q", outcome.Kind(), runtimelifecycle.StartRunning)
	}
	assertAttempt(t, outcome.Attempt(), "runtime-success", "attempt-success")
	if sourceCalls.Load() != 1 {
		t.Fatalf("LoadExact() calls = %d, want 1", sourceCalls.Load())
	}

	stop, err := owner.Stop(context.Background())
	if err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if stop.Kind() != runtimelifecycle.StopStopped {
		t.Fatalf("Stop() kind = %q, want %q", stop.Kind(), runtimelifecycle.StopStopped)
	}
}

func TestLoaderFailureIsPreservedAndSkipsBuild(t *testing.T) {
	var buildCalls atomic.Int32
	owner := mustOwner(t, "runtime-load-failure", "attempt-load-failure", &runtime.DependencyBindings{})
	loader := configurationloader.New(sourceFunc(func(uint64, uint64, uint64) (configurationloader.SourceObservation, error) {
		return configurationloader.SourceObservation{}, configurationloader.ErrSourceNotFound
	}))
	flow, err := newFlow(
		owner,
		loader,
		func(runtimeconfigload.DetachedLoadResult) (runtimeconfig.Snapshot, []runtimeconfig.Diagnostic) {
			buildCalls.Add(1)
			return runtimeconfig.Snapshot{}, nil
		},
	)
	if err != nil {
		t.Fatalf("newFlow() error = %v", err)
	}

	outcome, err := flow.Start(
		context.Background(),
		runtimelifecycle.NewStartRequest(workspaceID, configurationID, versionID),
	)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if outcome.Kind() != runtimelifecycle.StartPreparationFailed {
		t.Fatalf("Start() kind = %q, want %q", outcome.Kind(), runtimelifecycle.StartPreparationFailed)
	}
	failure, ok := outcome.PreparationFailure()
	if !ok || failure != configurationloader.ErrSourceNotFound {
		t.Fatalf("PreparationFailure() = %v/%t, want exact ErrSourceNotFound", failure, ok)
	}
	if buildCalls.Load() != 0 {
		t.Fatalf("Build() calls = %d, want 0", buildCalls.Load())
	}
}

func TestBuilderDiagnosticsBecomeImmutableBuildFailure(t *testing.T) {
	observation := validObservation(8080)
	observation.ConfigurationVersion.Listener.Host = ""
	observation.ConfigurationVersion.Listener.Port = 0
	owner := mustOwner(t, "runtime-build-failure", "attempt-build-failure", &runtime.DependencyBindings{})
	flow := mustFlow(
		t,
		owner,
		configurationloader.New(sourceFunc(func(uint64, uint64, uint64) (configurationloader.SourceObservation, error) {
			return observation, nil
		})),
	)

	outcome, err := flow.Start(
		context.Background(),
		runtimelifecycle.NewStartRequest(workspaceID, configurationID, versionID),
	)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if outcome.Kind() != runtimelifecycle.StartPreparationFailed {
		t.Fatalf("Start() kind = %q, want %q", outcome.Kind(), runtimelifecycle.StartPreparationFailed)
	}
	failure, ok := outcome.PreparationFailure()
	if !ok {
		t.Fatal("PreparationFailure() missing")
	}
	var buildFailure *BuildFailure
	if !errors.As(failure, &buildFailure) {
		t.Fatalf("PreparationFailure() = %T, want *BuildFailure", failure)
	}
	if buildFailure.Error() != buildFailureDescription {
		t.Fatalf("Error() = %q, want %q", buildFailure.Error(), buildFailureDescription)
	}
	first := buildFailure.Diagnostics()
	if len(first) < 2 {
		t.Fatalf("Diagnostics() count = %d, want at least 2", len(first))
	}
	expected := append([]runtimeconfig.Diagnostic(nil), first...)
	first[0] = runtimeconfig.Diagnostic{}
	if got := buildFailure.Diagnostics(); !reflect.DeepEqual(got, expected) {
		t.Fatalf("Diagnostics() mutated through returned slice: got %#v want %#v", got, expected)
	}
}

func TestValidationAndCallerCancellationGateDoNotClaim(t *testing.T) {
	loaderCalls := atomic.Int32{}
	loader := configurationloader.New(sourceFunc(func(uint64, uint64, uint64) (configurationloader.SourceObservation, error) {
		loaderCalls.Add(1)
		return validObservation(8080), nil
	}))
	ownerCalls := atomic.Int32{}
	owner, err := runtimelifecycle.NewOwner(
		workspaceID,
		configurationID,
		"runtime-validation",
		func() (runtimeconfigload.LaunchAttemptID, error) {
			ownerCalls.Add(1)
			return "attempt-validation", nil
		},
		&runtime.DependencyBindings{},
	)
	if err != nil {
		t.Fatalf("NewOwner() error = %v", err)
	}
	flow := mustFlow(t, owner, loader)
	request := runtimelifecycle.NewStartRequest(workspaceID, configurationID, versionID)

	if _, err := flow.Start(nil, request); !errors.Is(err, ErrInvalidStartContext) {
		t.Fatalf("Start(nil) error = %v, want ErrInvalidStartContext", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := flow.Start(cancelled, request); !errors.Is(err, context.Canceled) {
		t.Fatalf("Start(cancelled) error = %v, want context.Canceled", err)
	}
	if ownerCalls.Load() != 0 || loaderCalls.Load() != 0 {
		t.Fatalf("validation mutated lifecycle: ID calls=%d Loader calls=%d", ownerCalls.Load(), loaderCalls.Load())
	}

	if _, err := New(nil, loader); !errors.Is(err, ErrInvalidFlow) {
		t.Fatalf("New(nil, loader) error = %v, want ErrInvalidFlow", err)
	}
	if _, err := New(owner, nil); !errors.Is(err, ErrInvalidFlow) {
		t.Fatalf("New(owner, nil) error = %v, want ErrInvalidFlow", err)
	}
	var nilFlow *Flow
	if _, err := nilFlow.Start(context.Background(), request); !errors.Is(err, ErrInvalidFlow) {
		t.Fatalf("nil Flow Start() error = %v, want ErrInvalidFlow", err)
	}
}

func TestCancellationAfterGateDoesNotInterruptOperation(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	loader := configurationloader.New(sourceFunc(func(uint64, uint64, uint64) (configurationloader.SourceObservation, error) {
		close(entered)
		<-release
		return validObservation(8080), nil
	}))
	owner := mustOwner(t, "runtime-gate", "attempt-gate", &runtime.DependencyBindings{})
	flow := mustFlow(t, owner, loader)
	ctx, cancel := context.WithCancel(context.Background())
	result := startAsync(flow, ctx)

	<-entered
	cancel()
	close(release)
	completed := <-result
	if completed.err != nil {
		t.Fatalf("Start() error = %v, want Owner outcome after Gate", completed.err)
	}
	if completed.outcome.Kind() != runtimelifecycle.StartLaunchFailed {
		t.Fatalf("Start() kind = %q, want %q", completed.outcome.Kind(), runtimelifecycle.StartLaunchFailed)
	}
}

func TestStopDuringLoadConvergesWithoutBuildOrLaunch(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var buildCalls atomic.Int32
	owner := mustOwner(t, "runtime-stop-load", "attempt-stop-load", &runtime.DependencyBindings{})
	loader := configurationloader.New(sourceFunc(func(uint64, uint64, uint64) (configurationloader.SourceObservation, error) {
		close(entered)
		<-release
		return validObservation(8080), nil
	}))
	flow, err := newFlow(
		owner,
		loader,
		func(runtimeconfigload.DetachedLoadResult) (runtimeconfig.Snapshot, []runtimeconfig.Diagnostic) {
			buildCalls.Add(1)
			return runtimeconfig.Snapshot{}, nil
		},
	)
	if err != nil {
		t.Fatalf("newFlow() error = %v", err)
	}
	result := startAsync(flow, context.Background())

	<-entered
	stop, err := owner.Stop(context.Background())
	if err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if stop.Kind() != runtimelifecycle.StopStopped {
		t.Fatalf("Stop() kind = %q, want %q", stop.Kind(), runtimelifecycle.StopStopped)
	}
	close(release)
	completed := <-result
	if completed.err != nil {
		t.Fatalf("Start() error = %v", completed.err)
	}
	if completed.outcome.Kind() != runtimelifecycle.StartStoppedBeforeRunning {
		t.Fatalf(
			"Start() kind = %q, want %q",
			completed.outcome.Kind(),
			runtimelifecycle.StartStoppedBeforeRunning,
		)
	}
	if buildCalls.Load() != 0 {
		t.Fatalf("Build() calls = %d, want 0", buildCalls.Load())
	}
}

func TestStopDuringBuildConvergesWithoutLaunch(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var buildCalls atomic.Int32
	owner := mustOwner(t, "runtime-stop-build", "attempt-stop-build", &runtime.DependencyBindings{})
	loader := configurationloader.New(sourceFunc(func(uint64, uint64, uint64) (configurationloader.SourceObservation, error) {
		return validObservation(8080), nil
	}))
	builder := runtimeconfig.NewBuilder()
	flow, err := newFlow(
		owner,
		loader,
		func(input runtimeconfigload.DetachedLoadResult) (runtimeconfig.Snapshot, []runtimeconfig.Diagnostic) {
			buildCalls.Add(1)
			close(entered)
			<-release
			return builder.Build(input)
		},
	)
	if err != nil {
		t.Fatalf("newFlow() error = %v", err)
	}
	result := startAsync(flow, context.Background())

	<-entered
	stop, err := owner.Stop(context.Background())
	if err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if stop.Kind() != runtimelifecycle.StopStopped {
		t.Fatalf("Stop() kind = %q, want %q", stop.Kind(), runtimelifecycle.StopStopped)
	}
	close(release)
	completed := <-result
	if completed.err != nil {
		t.Fatalf("Start() error = %v", completed.err)
	}
	if completed.outcome.Kind() != runtimelifecycle.StartStoppedBeforeRunning {
		t.Fatalf(
			"Start() kind = %q, want %q",
			completed.outcome.Kind(),
			runtimelifecycle.StartStoppedBeforeRunning,
		)
	}
	if buildCalls.Load() != 1 {
		t.Fatalf("Build() calls = %d, want 1", buildCalls.Load())
	}
}

func TestConcurrentStartCreatesOneOperationAndOwnersRemainIndependent(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var sourceCalls atomic.Int32
	owner := mustSequentialOwner(t, "runtime-concurrent", &runtime.DependencyBindings{})
	flow := mustFlow(
		t,
		owner,
		configurationloader.New(sourceFunc(func(uint64, uint64, uint64) (configurationloader.SourceObservation, error) {
			sourceCalls.Add(1)
			close(entered)
			<-release
			return validObservation(8080), nil
		})),
	)
	first := startAsync(flow, context.Background())
	<-entered
	if _, err := flow.Start(
		context.Background(),
		runtimelifecycle.NewStartRequest(workspaceID, configurationID, versionID),
	); !errors.Is(err, runtimelifecycle.ErrStartConflict) {
		t.Fatalf("concurrent Start() error = %v, want ErrStartConflict", err)
	}
	close(release)
	if completed := <-first; completed.err != nil ||
		completed.outcome.Kind() != runtimelifecycle.StartLaunchFailed {
		t.Fatalf("first Start() = %q/%v, want StartLaunchFailed/nil", completed.outcome.Kind(), completed.err)
	}
	if sourceCalls.Load() != 1 {
		t.Fatalf("LoadExact() calls = %d, want 1", sourceCalls.Load())
	}

	firstOwner := mustOwner(t, "runtime-independent-1", "attempt-independent-1", &runtime.DependencyBindings{})
	secondOwner := mustOwner(t, "runtime-independent-2", "attempt-independent-2", &runtime.DependencyBindings{})
	firstFlow := mustFlow(t, firstOwner, configurationloader.New(staticSource(8080)))
	secondFlow := mustFlow(t, secondOwner, configurationloader.New(staticSource(8081)))
	firstResult := startAsync(firstFlow, context.Background())
	secondResult := startAsync(secondFlow, context.Background())
	for index, result := range []startResult{<-firstResult, <-secondResult} {
		if result.err != nil || result.outcome.Kind() != runtimelifecycle.StartLaunchFailed {
			t.Fatalf("independent Start[%d] = %q/%v, want StartLaunchFailed/nil", index, result.outcome.Kind(), result.err)
		}
	}
}

func TestProductionFlowHasNoDirectRuntimeDependency(t *testing.T) {
	pkg, err := build.Default.ImportDir(".", 0)
	if err != nil {
		t.Fatalf("ImportDir() error = %v", err)
	}
	got := append([]string(nil), pkg.Imports...)
	slices.Sort(got)
	want := []string{
		"context",
		"errors",
		"github.com/dsdred/universal-websocket-platform/internal/configurationloader",
		"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig",
		"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload",
		"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle",
	}
	slices.Sort(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("production imports = %q, want %q", got, want)
	}
}

type startResult struct {
	outcome runtimelifecycle.StartOutcome
	err     error
}

func startAsync(flow *Flow, ctx context.Context) <-chan startResult {
	result := make(chan startResult, 1)
	go func() {
		outcome, err := flow.Start(
			ctx,
			runtimelifecycle.NewStartRequest(workspaceID, configurationID, versionID),
		)
		result <- startResult{outcome: outcome, err: err}
	}()
	return result
}

func mustFlow(
	t *testing.T,
	owner *runtimelifecycle.Owner,
	loader *configurationloader.Loader,
) *Flow {
	t.Helper()
	flow, err := New(owner, loader)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return flow
}

func mustOwner(
	t *testing.T,
	instanceID runtimeconfigload.RuntimeInstanceID,
	attemptID runtimeconfigload.LaunchAttemptID,
	dependencies *runtime.DependencyBindings,
) *runtimelifecycle.Owner {
	t.Helper()
	owner, err := runtimelifecycle.NewOwner(
		workspaceID,
		configurationID,
		instanceID,
		func() (runtimeconfigload.LaunchAttemptID, error) {
			return attemptID, nil
		},
		dependencies,
	)
	if err != nil {
		t.Fatalf("NewOwner() error = %v", err)
	}
	return owner
}

func mustSequentialOwner(
	t *testing.T,
	instanceID runtimeconfigload.RuntimeInstanceID,
	dependencies *runtime.DependencyBindings,
) *runtimelifecycle.Owner {
	t.Helper()
	var next atomic.Int32
	owner, err := runtimelifecycle.NewOwner(
		workspaceID,
		configurationID,
		instanceID,
		func() (runtimeconfigload.LaunchAttemptID, error) {
			return runtimeconfigload.LaunchAttemptID(
				fmt.Sprintf("attempt-%d", next.Add(1)),
			), nil
		},
		dependencies,
	)
	if err != nil {
		t.Fatalf("NewOwner() error = %v", err)
	}
	return owner
}

func staticSource(port uint16) sourceFunc {
	return func(uint64, uint64, uint64) (configurationloader.SourceObservation, error) {
		return validObservation(port), nil
	}
}

func validObservation(port uint16) configurationloader.SourceObservation {
	return configurationloader.SourceObservation{
		WorkspaceID: workspaceID,
		Configuration: configuration.Configuration{
			ID:          configurationID,
			WorkspaceID: workspaceID,
			Name:        "runtime-launch-flow",
		},
		ConfigurationVersion: configurationversion.ConfigurationVersion{
			ID:              versionID,
			ConfigurationID: configurationID,
			Number:          1,
			State:           configurationversion.Published,
			Listener: configurationversion.ListenerSettings{
				Host: "127.0.0.1",
				Port: port,
				TLS: configurationversion.TLSSettings{
					MinVersion: "1.2",
				},
				Timeouts: configurationversion.TimeoutSettings{
					HandshakeSeconds: 10,
					ReadSeconds:      30,
					WriteSeconds:     20,
					IdleSeconds:      60,
				},
			},
		},
		SchemaIdentity:         "uwp.configuration",
		SchemaVersion:          1,
		RepresentationComplete: true,
	}
}

func assertSourceIdentities(
	t *testing.T,
	workspace uint64,
	configuration uint64,
	version uint64,
) {
	t.Helper()
	if workspace != workspaceID || configuration != configurationID || version != versionID {
		t.Fatalf(
			"LoadExact() identities = (%d, %d, %d), want (%d, %d, %d)",
			workspace,
			configuration,
			version,
			workspaceID,
			configurationID,
			versionID,
		)
	}
}

func assertAttempt(
	t *testing.T,
	attempt runtimelifecycle.AttemptFact,
	instanceID runtimeconfigload.RuntimeInstanceID,
	attemptID runtimeconfigload.LaunchAttemptID,
) {
	t.Helper()
	if attempt.WorkspaceID() != workspaceID ||
		attempt.ConfigurationID() != configurationID ||
		attempt.ConfigurationVersionID() != versionID ||
		attempt.RuntimeInstanceID() != instanceID ||
		attempt.LaunchAttemptID() != attemptID {
		t.Fatalf("Attempt() = %#v, want complete exact identities", attempt)
	}
}

func availablePort(t *testing.T) uint16 {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve listener address: %v", err)
	}
	_, portText, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		listener.Close()
		t.Fatalf("SplitHostPort(%q) error = %v", listener.Addr(), err)
	}
	if err := listener.Close(); err != nil {
		t.Fatalf("close listener reservation: %v", err)
	}
	port, err := strconv.ParseUint(portText, 10, 16)
	if err != nil {
		t.Fatalf("ParseUint(%q) error = %v", portText, err)
	}
	return uint16(port)
}
