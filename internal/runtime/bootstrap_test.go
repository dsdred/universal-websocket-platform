package runtime

import (
	"context"
	"errors"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
	"github.com/golang-jwt/jwt/v5"
)

func TestBootstrapRequestShapeAndCompleteProvenance(t *testing.T) {
	requestType := reflect.TypeOf(BootstrapRequest{})
	if requestType.NumField() != 3 {
		t.Fatalf("BootstrapRequest fields = %d, want exactly 3", requestType.NumField())
	}
	for index, want := range []struct {
		name string
		typ  reflect.Type
	}{
		{name: "Snapshot", typ: reflect.TypeOf(runtimeconfig.Snapshot{})},
		{name: "StartupContext", typ: reflect.TypeOf((*context.Context)(nil)).Elem()},
		{name: "Dependencies", typ: reflect.TypeOf((*DependencyBindings)(nil))},
	} {
		field := requestType.Field(index)
		if field.Name != want.name || field.Type != want.typ {
			t.Fatalf("BootstrapRequest field %d = %s %v, want %s %v",
				index, field.Name, field.Type, want.name, want.typ)
		}
	}

	complete := runtimeconfig.Provenance{
		WorkspaceID:                1,
		ConfigurationID:            2,
		ConfigurationVersionID:     3,
		ConfigurationVersionNumber: 4,
		SchemaIdentity:             "uwp.configuration",
		SchemaVersion:              5,
		RuntimeInstanceID:          runtimeconfigload.RuntimeInstanceID("runtime"),
		LaunchAttemptID:            runtimeconfigload.LaunchAttemptID("attempt"),
	}
	if !completeBootstrapProvenance(complete) {
		t.Fatal("complete provenance was rejected")
	}
	tests := []struct {
		name   string
		mutate func(*runtimeconfig.Provenance)
	}{
		{name: "Workspace ID", mutate: func(value *runtimeconfig.Provenance) { value.WorkspaceID = 0 }},
		{name: "Configuration ID", mutate: func(value *runtimeconfig.Provenance) { value.ConfigurationID = 0 }},
		{name: "ConfigurationVersion ID", mutate: func(value *runtimeconfig.Provenance) { value.ConfigurationVersionID = 0 }},
		{name: "ConfigurationVersion number", mutate: func(value *runtimeconfig.Provenance) { value.ConfigurationVersionNumber = 0 }},
		{name: "schema identity", mutate: func(value *runtimeconfig.Provenance) { value.SchemaIdentity = "" }},
		{name: "schema version", mutate: func(value *runtimeconfig.Provenance) { value.SchemaVersion = 0 }},
		{name: "Runtime Instance ID", mutate: func(value *runtimeconfig.Provenance) { value.RuntimeInstanceID = "" }},
		{name: "Launch Attempt ID", mutate: func(value *runtimeconfig.Provenance) { value.LaunchAttemptID = "" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			incomplete := complete
			test.mutate(&incomplete)
			if completeBootstrapProvenance(incomplete) {
				t.Fatalf("provenance without %s was accepted", test.name)
			}
		})
	}
}

func TestBootstrapFailurePrecedenceAndRegistry(t *testing.T) {
	snapshot := apiKeySnapshot(t)
	resolver := emptyResolver(t)
	constructionCause := errors.New("construction cause")
	buildCause := errors.New("build cause")

	tests := []struct {
		name        string
		request     *BootstrapRequest
		factory     bootstrapHostFactory
		wantStage   BootstrapFailureStage
		wantCode    BootstrapFailureCode
		wantError   string
		wantCause   error
		wantFactory int
		wantBuild   int
		wantStart   int
	}{
		{
			name:        "missing request envelope",
			wantStage:   BootstrapStageInputValidation,
			wantCode:    BootstrapCodeInvalidStartupContext,
			wantError:   invalidStartupContextDescription,
			wantFactory: 0,
		},
		{
			name: "typed nil startup context wins over invalid snapshot and bindings",
			request: &BootstrapRequest{
				StartupContext: (*bootstrapProofContext)(nil),
			},
			wantStage:   BootstrapStageInputValidation,
			wantCode:    BootstrapCodeInvalidStartupContext,
			wantError:   invalidStartupContextDescription,
			wantFactory: 0,
		},
		{
			name: "invalid snapshot wins over missing bindings",
			request: &BootstrapRequest{
				StartupContext: context.Background(),
			},
			wantStage:   BootstrapStageInputValidation,
			wantCode:    BootstrapCodeInvalidSnapshot,
			wantError:   invalidSnapshotDescription,
			wantFactory: 0,
		},
		{
			name: "missing secret resolver",
			request: &BootstrapRequest{
				Snapshot:       snapshot,
				StartupContext: context.Background(),
				Dependencies:   &DependencyBindings{},
			},
			wantStage:   BootstrapStageDependencyBinding,
			wantCode:    BootstrapCodeMissingSecretResolver,
			wantError:   missingSecretResolverDescription,
			wantFactory: 0,
		},
		{
			name: "typed nil secret resolver",
			request: &BootstrapRequest{
				Snapshot:       snapshot,
				StartupContext: context.Background(),
				Dependencies: &DependencyBindings{
					SecretResolver: (*bootstrapProofResolver)(nil),
				},
			},
			wantStage:   BootstrapStageDependencyBinding,
			wantCode:    BootstrapCodeMissingSecretResolver,
			wantError:   missingSecretResolverDescription,
			wantFactory: 0,
		},
		{
			name: "host construction",
			request: &BootstrapRequest{
				Snapshot:       snapshot,
				StartupContext: context.Background(),
				Dependencies:   &DependencyBindings{SecretResolver: resolver},
			},
			factory: func(runtimeconfig.Snapshot, secretresolver.Resolver, message.Handler, func(error)) (Host, error) {
				return nil, constructionCause
			},
			wantStage:   BootstrapStageHostConstruction,
			wantCode:    BootstrapCodeHostConstructionFailed,
			wantError:   hostConstructionFailedDescription,
			wantCause:   constructionCause,
			wantFactory: 1,
		},
		{
			name: "host preparation",
			request: &BootstrapRequest{
				Snapshot:       snapshot,
				StartupContext: context.Background(),
				Dependencies:   &DependencyBindings{SecretResolver: resolver},
			},
			factory:     proofBootstrapFactory(&bootstrapProofHost{buildErr: buildCause}, nil),
			wantStage:   BootstrapStageHostPreparation,
			wantCode:    BootstrapCodeHostBuildFailed,
			wantError:   hostBuildFailedDescription,
			wantCause:   buildCause,
			wantFactory: 1,
			wantBuild:   1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var factoryCalls int
			factory := test.factory
			var proofHost *bootstrapProofHost
			if factory == nil {
				proofHost = &bootstrapProofHost{}
				factory = proofBootstrapFactory(proofHost, &factoryCalls)
			} else {
				delegate := factory
				factory = func(
					snapshot runtimeconfig.Snapshot,
					resolver secretresolver.Resolver,
					handler message.Handler,
					reporter func(error),
				) (Host, error) {
					factoryCalls++
					host, err := delegate(snapshot, resolver, handler, reporter)
					if candidate, ok := host.(*bootstrapProofHost); ok {
						proofHost = candidate
					}
					return host, err
				}
			}

			outcome := bootstrap(test.request, factory)
			failure, ok := outcome.BootstrapFailure()
			if !ok {
				t.Fatalf("bootstrap() = %#v, want BootstrapFailure", outcome)
			}
			if failure.Stage() != test.wantStage || failure.Code() != test.wantCode {
				t.Fatalf("failure identity = (%q, %q), want (%q, %q)",
					failure.Stage(), failure.Code(), test.wantStage, test.wantCode)
			}
			if failure.Error() != test.wantError {
				t.Fatalf("failure Error() = %q, want %q", failure.Error(), test.wantError)
			}
			if test.wantCause != nil && !errors.Is(failure, test.wantCause) {
				t.Fatalf("errors.Is(failure, cause) = false, cause %v", test.wantCause)
			}
			if factoryCalls != test.wantFactory {
				t.Fatalf("factory calls = %d, want %d", factoryCalls, test.wantFactory)
			}
			if proofHost != nil {
				if proofHost.buildCalls != test.wantBuild || proofHost.startCalls != test.wantStart {
					t.Fatalf("Host calls Build/Start = %d/%d, want %d/%d",
						proofHost.buildCalls, proofHost.startCalls, test.wantBuild, test.wantStart)
				}
				if proofHost.stopCalls != 0 {
					t.Fatalf("Host Stop calls = %d, want 0", proofHost.stopCalls)
				}
			}
			assertExclusiveBootstrapOutcome(t, outcome, bootstrapOutcomeBootstrapFailure)
		})
	}
}

func TestBootstrapBindsCapabilitiesAndStartsExactlyOnce(t *testing.T) {
	snapshot := apiKeySnapshot(t)
	startupContext, cancel := context.WithCancel(context.Background())
	cancel()
	resolver := &bootstrapProofResolver{}
	var typedNilHandler *bootstrapProofHandler
	var typedNilReporter func(error)
	host := &bootstrapProofHost{snapshot: snapshot}
	var capturedSnapshot runtimeconfig.Snapshot
	var capturedResolver secretresolver.Resolver
	var capturedHandler message.Handler
	var capturedReporter func(error)
	factoryCalls := 0

	outcome := bootstrap(&BootstrapRequest{
		Snapshot:       snapshot,
		StartupContext: startupContext,
		Dependencies: &DependencyBindings{
			SecretResolver:        resolver,
			LegacyMessageHandler:  typedNilHandler,
			TerminalErrorReporter: typedNilReporter,
		},
	}, func(
		gotSnapshot runtimeconfig.Snapshot,
		gotResolver secretresolver.Resolver,
		gotHandler message.Handler,
		gotReporter func(error),
	) (Host, error) {
		factoryCalls++
		capturedSnapshot = gotSnapshot
		capturedResolver = gotResolver
		capturedHandler = gotHandler
		capturedReporter = gotReporter
		return host, nil
	})

	active, ok := outcome.Success()
	if !ok || active != host {
		t.Fatalf("bootstrap() Success = (%T, %t), want proof Host", active, ok)
	}
	if factoryCalls != 1 || host.buildCalls != 1 || host.startCalls != 1 || host.stopCalls != 0 {
		t.Fatalf("calls factory/Build/Start/Stop = %d/%d/%d/%d, want 1/1/1/0",
			factoryCalls, host.buildCalls, host.startCalls, host.stopCalls)
	}
	if capturedSnapshot.Provenance() != snapshot.Provenance() {
		t.Fatal("factory did not receive Snapshot by value")
	}
	if capturedResolver != resolver || capturedHandler != nil || capturedReporter != nil {
		t.Fatal("factory received altered required binding or non-normalized optional typed nil")
	}
	if host.startContext != startupContext {
		t.Fatal("Host.Start did not receive the original startup context")
	}
	if resolver.resolveCalls.Load() != 0 {
		t.Fatal("Bootstrap invoked a bound capability")
	}
	assertExclusiveBootstrapOutcome(t, outcome, bootstrapOutcomeSuccess)
}

func TestBootstrapPassesOptionalCapabilitiesWithoutCallingThem(t *testing.T) {
	request := validBootstrapProofRequest(t)
	handler := &bootstrapProofHandler{}
	var reporterCalls atomic.Int32
	reporter := func(error) { reporterCalls.Add(1) }
	request.Dependencies.LegacyMessageHandler = handler
	request.Dependencies.TerminalErrorReporter = reporter
	host := &bootstrapProofHost{}

	outcome := bootstrap(request, func(
		_ runtimeconfig.Snapshot,
		_ secretresolver.Resolver,
		gotHandler message.Handler,
		gotReporter func(error),
	) (Host, error) {
		if gotHandler != handler || reflect.ValueOf(gotReporter).Pointer() != reflect.ValueOf(reporter).Pointer() {
			t.Fatal("factory did not receive the original optional capabilities")
		}
		return host, nil
	})

	if _, ok := outcome.Success(); !ok {
		t.Fatalf("bootstrap() = %#v, want Success", outcome)
	}
	if handler.handleCalls.Load() != 0 || reporterCalls.Load() != 0 {
		t.Fatal("Bootstrap invoked an optional capability")
	}
}

func TestBootstrapStartupFailurePreservesCauseAndPerformsNoCleanup(t *testing.T) {
	first := errors.New("first startup cause")
	second := &bootstrapProofError{label: "second startup cause"}
	joined := errors.Join(first, second)
	host := &bootstrapProofHost{startErr: joined}

	outcome := bootstrap(validBootstrapProofRequest(t), proofBootstrapFactory(host, nil))
	failure, ok := outcome.StartupFailure()
	if !ok {
		t.Fatalf("bootstrap() = %#v, want StartupFailure", outcome)
	}
	if failure.Error() != hostStartupFailedDescription {
		t.Fatalf("StartupFailure.Error() = %q", failure.Error())
	}
	if errors.Unwrap(failure) != joined || !errors.Is(failure, first) {
		t.Fatal("StartupFailure did not directly preserve the joined cause")
	}
	var found *bootstrapProofError
	if !errors.As(failure, &found) || found != second {
		t.Fatal("StartupFailure did not preserve errors.As through joined cause")
	}
	if host.buildCalls != 1 || host.startCalls != 1 || host.stopCalls != 0 {
		t.Fatalf("Host calls Build/Start/Stop = %d/%d/%d, want 1/1/0",
			host.buildCalls, host.startCalls, host.stopCalls)
	}
	assertExclusiveBootstrapOutcome(t, outcome, bootstrapOutcomeStartupFailure)
}

func TestBootstrapFailurePreservesCauseAndPerformsNoCleanup(t *testing.T) {
	first := errors.New("first construction cause")
	second := &bootstrapProofError{label: "second construction cause"}
	joined := errors.Join(first, second)

	outcome := bootstrap(validBootstrapProofRequest(t), func(
		runtimeconfig.Snapshot,
		secretresolver.Resolver,
		message.Handler,
		func(error),
	) (Host, error) {
		return nil, joined
	})
	failure, ok := outcome.BootstrapFailure()
	if !ok {
		t.Fatalf("bootstrap() = %#v, want BootstrapFailure", outcome)
	}
	if failure.Error() != hostConstructionFailedDescription {
		t.Fatalf("BootstrapFailure.Error() = %q", failure.Error())
	}
	if errors.Unwrap(failure) != joined || !errors.Is(failure, first) {
		t.Fatal("BootstrapFailure did not directly preserve the joined cause")
	}
	var found *bootstrapProofError
	if !errors.As(failure, &found) || found != second {
		t.Fatal("BootstrapFailure did not preserve errors.As through joined cause")
	}
	if strings.Contains(failure.Error(), first.Error()) ||
		strings.Contains(failure.Error(), second.Error()) {
		t.Fatal("BootstrapFailure description exposed cause text")
	}
	assertExclusiveBootstrapOutcome(t, outcome, bootstrapOutcomeBootstrapFailure)
}

func TestBootstrapConcurrentInvocationsAreIndependent(t *testing.T) {
	const invocations = 32
	request := validBootstrapProofRequest(t)
	var factoryCalls atomic.Int32
	hosts := make(chan *bootstrapProofHost, invocations)
	factory := func(
		runtimeconfig.Snapshot,
		secretresolver.Resolver,
		message.Handler,
		func(error),
	) (Host, error) {
		factoryCalls.Add(1)
		host := &bootstrapProofHost{}
		hosts <- host
		return host, nil
	}

	var wait sync.WaitGroup
	outcomes := make(chan BootstrapOutcome, invocations)
	for range invocations {
		wait.Add(1)
		go func() {
			defer wait.Done()
			outcomes <- bootstrap(request, factory)
		}()
	}
	wait.Wait()
	close(outcomes)
	close(hosts)

	for outcome := range outcomes {
		if host, ok := outcome.Success(); !ok || host == nil {
			t.Fatalf("concurrent bootstrap() = %#v, want Success", outcome)
		}
	}
	if factoryCalls.Load() != invocations {
		t.Fatalf("factory calls = %d, want %d", factoryCalls.Load(), invocations)
	}
	seen := map[*bootstrapProofHost]bool{}
	for host := range hosts {
		if seen[host] {
			t.Fatal("concurrent invocation reused a Host")
		}
		seen[host] = true
		if host.buildCalls != 1 || host.startCalls != 1 {
			t.Fatalf("Host calls Build/Start = %d/%d, want 1/1", host.buildCalls, host.startCalls)
		}
	}
}

func TestBootstrapUsesRuntimeHost(t *testing.T) {
	resolver := apiKeyResolver(t)
	outcome := Bootstrap(&BootstrapRequest{
		Snapshot:       apiKeySnapshot(t),
		StartupContext: context.Background(),
		Dependencies:   &DependencyBindings{SecretResolver: resolver},
	})
	active, ok := outcome.Success()
	if !ok {
		t.Fatalf("Bootstrap() = %#v, want Success", outcome)
	}
	host, ok := active.(*DefaultHost)
	if !ok {
		t.Fatalf("Bootstrap() Host type = %T, want *DefaultHost", active)
	}
	t.Cleanup(func() { _ = active.Stop(context.Background()) })
	if host.state != hostRunning || host.runtimeListener == nil {
		t.Fatalf("Bootstrap() Host = %#v, want Running Host with published Listener", host)
	}
	if !active.Ready() {
		t.Fatal("Bootstrap() returned Host that is not Ready")
	}
	if !active.CanAccept() {
		t.Fatal("Bootstrap() returned Host with closed admission")
	}
}

func TestBootstrapHostStartPreservesAuthenticationBuildErrors(t *testing.T) {
	snapshot := snapshotWithAuthentication(t, configurationversion.AuthenticationSettings{
		Enabled: true,
		Providers: []configurationversion.AuthenticationProvider{{
			Name:    "basic",
			Type:    configurationversion.AuthenticationProviderBasic,
			Enabled: true,
			Basic: &configurationversion.BasicSettings{
				Realm:     "Universal WebSocket Platform",
				SecretRef: "secrets/basic/main",
			},
		}},
	})

	outcome := Bootstrap(&BootstrapRequest{
		Snapshot:       snapshot,
		StartupContext: context.Background(),
		Dependencies:   &DependencyBindings{SecretResolver: emptyResolver(t)},
	})
	failure, ok := outcome.StartupFailure()
	if !ok || !errors.Is(failure, authentication.ErrFactoryNotFound) {
		t.Fatalf("Bootstrap() StartupFailure = (%v, %t), want ErrFactoryNotFound", failure, ok)
	}
	if host, ok := outcome.Success(); ok || host != nil {
		t.Fatal("startup failure published a partial Host")
	}
}

func TestBootstrapHostPreservesRuntimeVertical(t *testing.T) {
	reportedErrors := make(chan error, 4)
	snapshot := apiKeySnapshot(t)
	outcome := Bootstrap(&BootstrapRequest{
		Snapshot:       snapshot,
		StartupContext: context.Background(),
		Dependencies: &DependencyBindings{
			SecretResolver:        apiKeyResolver(t),
			LegacyMessageHandler:  message.NewEchoHandler(),
			TerminalErrorReporter: func(err error) { reportedErrors <- err },
		},
	})
	built, ok := outcome.Success()
	if !ok {
		t.Fatalf("Bootstrap() = %#v, want Success", outcome)
	}
	if !built.Ready() {
		t.Fatal("Host Ready() = false after successful Start")
	}
	if !built.CanAccept() {
		t.Fatal("Host CanAccept() = false after successful Start")
	}
	runtimeContext := built.RuntimeContext()
	if runtimeContext == nil {
		t.Fatal("RuntimeContext() = nil after successful Start")
	}
	address := net.JoinHostPort(snapshot.Listener().Host, portString(snapshot.Listener().Port))
	t.Cleanup(func() {
		if err := built.Stop(context.Background()); err != nil {
			t.Errorf("cleanup Stop() error = %v", err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	header := make(http.Header)
	header.Set("X-API-Key", "runtime-secret")
	connection, response, err := websocket.Dial(
		ctx,
		"ws://127.0.0.1:"+portString(snapshot.Listener().Port)+"/ws",
		&websocket.DialOptions{HTTPHeader: header},
	)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer connection.CloseNow()
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("Dial() status = %d, want 101", response.StatusCode)
	}

	for _, test := range []struct {
		messageType websocket.MessageType
		payload     []byte
	}{
		{messageType: websocket.MessageText, payload: []byte("runtime host text")},
		{messageType: websocket.MessageBinary, payload: []byte{0x00, 0x01, 0xfe, 0xff}},
	} {
		if err := connection.Write(ctx, test.messageType, test.payload); err != nil {
			t.Fatalf("Write(%v) error = %v", test.messageType, err)
		}
		messageType, got, err := connection.Read(ctx)
		if err != nil {
			t.Fatalf("Read(%v) error = %v", test.messageType, err)
		}
		if messageType != test.messageType || string(got) != string(test.payload) {
			t.Fatalf("Read() = (%v, %v), want (%v, %v)", messageType, got, test.messageType, test.payload)
		}
	}

	if err := connection.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := built.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if built.Ready() {
		t.Fatal("Host Ready() = true after Stop")
	}
	if built.CanAccept() {
		t.Fatal("Host CanAccept() = true after Stop")
	}
	assertContextCanceled(t, runtimeContext, "Runtime context after production vertical Stop")
	port, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatalf("TCP port %s was not released: %v", address, err)
	}
	if err := port.Close(); err != nil {
		t.Fatalf("close verification Listener: %v", err)
	}
	if len(reportedErrors) != 0 {
		reported := <-reportedErrors
		t.Fatalf("terminal error reported during clean production vertical: %v (cause: %v)", reported, errors.Unwrap(reported))
	}
}

func TestBootstrapHostRejectsAuthenticationBeforeUpgrade(t *testing.T) {
	snapshot := apiKeySnapshot(t)
	outcome := Bootstrap(&BootstrapRequest{
		Snapshot:       snapshot,
		StartupContext: context.Background(),
		Dependencies: &DependencyBindings{
			SecretResolver:       apiKeyResolver(t),
			LegacyMessageHandler: message.NewEchoHandler(),
		},
	})
	host, ok := outcome.Success()
	if !ok {
		t.Fatalf("Bootstrap() = %#v, want Success", outcome)
	}
	t.Cleanup(func() { _ = host.Stop(context.Background()) })

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	header := make(http.Header)
	header.Set("X-API-Key", "wrong-secret")
	connection, response, err := websocket.Dial(
		ctx,
		"ws://127.0.0.1:"+portString(snapshot.Listener().Port)+"/ws",
		&websocket.DialOptions{HTTPHeader: header},
	)
	if connection != nil {
		connection.CloseNow()
		t.Fatal("rejected Authentication unexpectedly created a WebSocket")
	}
	if err == nil || response == nil {
		t.Fatalf("Dial() = (%v, %v), want HTTP Authentication rejection", response, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusUnauthorized)
	}
}

func TestBootstrapHostAuthenticationErrorPreventsUpgrade(t *testing.T) {
	snapshot := apiKeySnapshot(t)
	outcome := Bootstrap(&BootstrapRequest{
		Snapshot:       snapshot,
		StartupContext: context.Background(),
		Dependencies:   &DependencyBindings{SecretResolver: emptyResolver(t)},
	})
	host, ok := outcome.Success()
	if !ok {
		t.Fatalf("Bootstrap() = %#v, want Success", outcome)
	}
	t.Cleanup(func() { _ = host.Stop(context.Background()) })

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	header := make(http.Header)
	header.Set("X-API-Key", "credential")
	connection, response, err := websocket.Dial(
		ctx,
		"ws://127.0.0.1:"+portString(snapshot.Listener().Port)+"/ws",
		&websocket.DialOptions{HTTPHeader: header},
	)
	if connection != nil {
		connection.CloseNow()
		t.Fatal("Authentication error unexpectedly created a WebSocket")
	}
	if err == nil || response == nil {
		t.Fatalf("Dial() = (%v, %v), want HTTP operational rejection", response, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusServiceUnavailable)
	}
}

func TestBootstrapTerminalObserverConsumesSessionError(t *testing.T) {
	wantErr := errors.New("handler failed with credential-that-must-not-leak")
	reportedErrors := make(chan error, 2)
	snapshot := apiKeySnapshot(t)
	outcome := Bootstrap(&BootstrapRequest{
		Snapshot:       snapshot,
		StartupContext: context.Background(),
		Dependencies: &DependencyBindings{
			SecretResolver:        apiKeyResolver(t),
			LegacyMessageHandler:  runtimeErrorHandler{err: wantErr},
			TerminalErrorReporter: func(err error) { reportedErrors <- err },
		},
	})
	host, ok := outcome.Success()
	if !ok {
		t.Fatalf("Bootstrap() = %#v, want Success", outcome)
	}
	t.Cleanup(func() { _ = host.Stop(context.Background()) })

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	header := make(http.Header)
	header.Set("X-API-Key", "runtime-secret")
	connection, _, err := websocket.Dial(
		ctx,
		"ws://127.0.0.1:"+portString(snapshot.Listener().Port)+"/ws",
		&websocket.DialOptions{HTTPHeader: header},
	)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer connection.CloseNow()
	if err := connection.Write(ctx, websocket.MessageText, []byte("message")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	readResult := make(chan error, 1)
	go func() {
		_, _, readErr := connection.Read(ctx)
		readResult <- readErr
	}()

	select {
	case <-readResult:
	case <-ctx.Done():
		t.Fatalf("Session connection did not close: %v", ctx.Err())
	}
	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	select {
	case reported := <-reportedErrors:
		t.Fatalf("post-Commit Session error escaped the Terminal Observer: %v", reported)
	default:
	}
}

func TestBootstrapHostDisabledAuthenticationUsesAnonymousSession(t *testing.T) {
	snapshot := snapshotWithAuthentication(t, configurationversion.AuthenticationSettings{})
	outcome := Bootstrap(&BootstrapRequest{
		Snapshot:       snapshot,
		StartupContext: context.Background(),
		Dependencies: &DependencyBindings{
			SecretResolver:       emptyResolver(t),
			LegacyMessageHandler: message.NewEchoHandler(),
		},
	})
	host, ok := outcome.Success()
	if !ok {
		t.Fatalf("Bootstrap() = %#v, want Success", outcome)
	}
	t.Cleanup(func() { _ = host.Stop(context.Background()) })

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	connection, response, err := websocket.Dial(
		ctx,
		"ws://127.0.0.1:"+portString(snapshot.Listener().Port)+"/ws",
		nil,
	)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer connection.CloseNow()
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusSwitchingProtocols)
	}
	want := []byte("anonymous echo")
	if err := connection.Write(ctx, websocket.MessageText, want); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	_, got, err := connection.Read(ctx)
	if err != nil || string(got) != string(want) {
		t.Fatalf("Read() = (%q, %v), want %q", got, err, want)
	}
}

func TestBootstrapHostJWTAuthenticationBeforeUpgrade(t *testing.T) {
	const (
		secret    = "runtime-jwt-secret"
		secretRef = "secrets/jwt/runtime"
	)
	resolver, err := secretresolver.NewMemory(map[string][]byte{secretRef: []byte(secret)})
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	snapshot := snapshotWithAuthentication(t, configurationversion.AuthenticationSettings{
		Enabled: true,
		Providers: []configurationversion.AuthenticationProvider{{
			Name:    "runtime-jwt",
			Type:    configurationversion.AuthenticationProviderJWT,
			Enabled: true,
			JWT: &configurationversion.JWTSettings{
				SigningKeys:       []configurationversion.JWTSigningKey{{Name: "primary", SecretRef: secretRef}},
				AllowedAlgorithms: []configurationversion.JWTAlgorithm{configurationversion.HS256},
			},
		}},
	})
	outcome := Bootstrap(&BootstrapRequest{
		Snapshot:       snapshot,
		StartupContext: context.Background(),
		Dependencies: &DependencyBindings{
			SecretResolver:       resolver,
			LegacyMessageHandler: message.NewEchoHandler(),
		},
	})
	host, ok := outcome.Success()
	if !ok {
		t.Fatalf("Bootstrap() = %#v, want Success", outcome)
	}
	t.Cleanup(func() { _ = host.Stop(context.Background()) })

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "runtime-jwt-client",
		"exp": time.Now().Add(time.Minute).Unix(),
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	header := make(http.Header)
	header.Set("Authorization", "Bearer "+token)
	connection, response, err := websocket.Dial(
		ctx,
		"ws://127.0.0.1:"+portString(snapshot.Listener().Port)+"/ws",
		&websocket.DialOptions{HTTPHeader: header},
	)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer connection.CloseNow()
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusSwitchingProtocols)
	}
	if err := connection.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func apiKeySnapshot(t *testing.T) runtimeconfig.Snapshot {
	t.Helper()
	return snapshotWithAuthentication(t, configurationversion.AuthenticationSettings{
		Enabled: true,
		Providers: []configurationversion.AuthenticationProvider{{
			Name:     "api-key",
			Type:     configurationversion.AuthenticationProviderAPIKey,
			Enabled:  true,
			Priority: 10,
			APIKey: &configurationversion.APIKeySettings{
				Header:    "X-API-Key",
				SecretRef: "secrets/api-key/runtime",
			},
		}},
	})
}

func snapshotWithAuthentication(t *testing.T, authentication configurationversion.AuthenticationSettings) runtimeconfig.Snapshot {
	t.Helper()
	version := configurationversion.ConfigurationVersion{
		ID: 1, ConfigurationID: 1, Number: 1, State: configurationversion.Published,
		Listener: configurationversion.ListenerSettings{
			Host: "127.0.0.1",
			Port: availablePort(t),
			TLS:  configurationversion.TLSSettings{MinVersion: "1.2"},
			Timeouts: configurationversion.TimeoutSettings{
				HandshakeSeconds: 10,
				WriteSeconds:     10,
				IdleSeconds:      60,
			},
		},
		Authentication: authentication,
	}
	request := runtimeconfigload.NewLoadRequest(1, 1, 1, "runtime", "attempt")
	input := runtimeconfigload.NewDetachedLoadResult(request, version, 1, true, "uwp.configuration", 1)
	snapshot, diagnostics := runtimeconfig.NewBuilder().Build(input)
	if len(diagnostics) != 0 {
		t.Fatalf("runtimeconfig.Build() diagnostics = %#v", diagnostics)
	}
	return snapshot
}

func apiKeyResolver(t *testing.T) secretresolver.Resolver {
	t.Helper()
	resolver, err := secretresolver.NewMemory(map[string][]byte{
		"secrets/api-key/runtime": []byte("runtime-secret"),
	})
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	return resolver
}

func portString(port uint16) string {
	return strconv.Itoa(int(port))
}

type runtimeErrorHandler struct {
	err error
}

func (handler runtimeErrorHandler) Handle(context.Context, message.Context) error {
	return handler.err
}

type bootstrapProofHost struct {
	snapshot     runtimeconfig.Snapshot
	buildErr     error
	startErr     error
	startContext context.Context
	buildCalls   int
	startCalls   int
	stopCalls    int
}

func (host *bootstrapProofHost) Snapshot() runtimeconfig.Snapshot { return host.snapshot }
func (host *bootstrapProofHost) RuntimeContext() context.Context  { return nil }
func (host *bootstrapProofHost) Build() error {
	host.buildCalls++
	return host.buildErr
}
func (host *bootstrapProofHost) Start(ctx context.Context) error {
	host.startCalls++
	host.startContext = ctx
	return host.startErr
}
func (host *bootstrapProofHost) Stop(context.Context) error {
	host.stopCalls++
	return nil
}
func (host *bootstrapProofHost) Running() bool   { return host.startCalls == 1 && host.startErr == nil }
func (host *bootstrapProofHost) Ready() bool     { return host.Running() }
func (host *bootstrapProofHost) CanAccept() bool { return host.Running() }

type bootstrapProofResolver struct {
	resolveCalls atomic.Int32
}

func (resolver *bootstrapProofResolver) Resolve(context.Context, string) (secretresolver.Secret, error) {
	resolver.resolveCalls.Add(1)
	return secretresolver.Secret{}, nil
}

type bootstrapProofHandler struct {
	handleCalls atomic.Int32
}

func (handler *bootstrapProofHandler) Handle(context.Context, message.Context) error {
	handler.handleCalls.Add(1)
	return nil
}

type bootstrapProofContext struct{}

func (*bootstrapProofContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (*bootstrapProofContext) Done() <-chan struct{}       { return nil }
func (*bootstrapProofContext) Err() error                  { return nil }
func (*bootstrapProofContext) Value(any) any               { return nil }

type bootstrapProofError struct {
	label string
}

func (failure *bootstrapProofError) Error() string { return failure.label }

func validBootstrapProofRequest(t *testing.T) *BootstrapRequest {
	t.Helper()
	return &BootstrapRequest{
		Snapshot:       apiKeySnapshot(t),
		StartupContext: context.Background(),
		Dependencies:   &DependencyBindings{SecretResolver: &bootstrapProofResolver{}},
	}
}

func proofBootstrapFactory(
	host Host,
	calls *int,
) bootstrapHostFactory {
	return func(
		runtimeconfig.Snapshot,
		secretresolver.Resolver,
		message.Handler,
		func(error),
	) (Host, error) {
		if calls != nil {
			*calls++
		}
		return host, nil
	}
}

func assertExclusiveBootstrapOutcome(
	t *testing.T,
	outcome BootstrapOutcome,
	want bootstrapOutcomeKind,
) {
	t.Helper()
	host, success := outcome.Success()
	bootstrapFailure, failedBeforeStart := outcome.BootstrapFailure()
	startupFailure, failedDuringStart := outcome.StartupFailure()
	present := 0
	for _, value := range []bool{success, failedBeforeStart, failedDuringStart} {
		if value {
			present++
		}
	}
	if present != 1 {
		t.Fatalf("outcome accessors selected %d variants, want exactly one", present)
	}
	switch want {
	case bootstrapOutcomeSuccess:
		if !success || host == nil || bootstrapFailure != nil || startupFailure != nil {
			t.Fatalf("outcome = %#v, want exclusive Success", outcome)
		}
	case bootstrapOutcomeBootstrapFailure:
		if success || host != nil || !failedBeforeStart || bootstrapFailure == nil || startupFailure != nil {
			t.Fatalf("outcome = %#v, want exclusive BootstrapFailure", outcome)
		}
	case bootstrapOutcomeStartupFailure:
		if success || host != nil || bootstrapFailure != nil || !failedDuringStart || startupFailure == nil {
			t.Fatalf("outcome = %#v, want exclusive StartupFailure", outcome)
		}
	default:
		t.Fatalf("unknown expected outcome kind %d", want)
	}
}
