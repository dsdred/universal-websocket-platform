package configurationloader

import (
	"errors"
	"fmt"
	"go/build"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/configuration"
	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

const (
	testWorkspaceID            uint64            = 11
	testConfigurationID        uint64            = 22
	testConfigurationVersionID uint64            = 33
	testRuntimeInstanceID      RuntimeInstanceID = "runtime-instance-44"
	testLaunchAttemptID        LaunchAttemptID   = "launch-attempt-55"
)

type sourceFunc func(uint64, uint64, uint64) (SourceObservation, error)

func (f sourceFunc) LoadExact(workspaceID, configurationID, configurationVersionID uint64) (SourceObservation, error) {
	return f(workspaceID, configurationID, configurationVersionID)
}

func TestLoadExactPublishedConfiguration(t *testing.T) {
	request := mustLoadRequest(t)
	observation := completeObservation()
	var calls atomic.Int32
	loader := New(sourceFunc(func(workspaceID, configurationID, versionID uint64) (SourceObservation, error) {
		calls.Add(1)
		if workspaceID != testWorkspaceID ||
			configurationID != testConfigurationID ||
			versionID != testConfigurationVersionID {
			t.Fatalf("LoadExact() identities = (%d, %d, %d)", workspaceID, configurationID, versionID)
		}
		return observation, nil
	}))

	result, err := loader.Load(request)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("LoadExact() calls = %d, want 1", calls.Load())
	}
	assertResultFacts(t, result)
	if got := result.ConfigurationVersion(); !reflect.DeepEqual(got, observation.ConfigurationVersion) {
		t.Fatalf("ConfigurationVersion() = %#v, want %#v", got, observation.ConfigurationVersion)
	}
}

func TestLoadRequestValidation(t *testing.T) {
	tests := []struct {
		name                  string
		workspaceID           uint64
		configurationID       uint64
		versionID             uint64
		runtimeInstanceID     RuntimeInstanceID
		launchAttemptIdentity LaunchAttemptID
	}{
		{"missing Workspace", 0, testConfigurationID, testConfigurationVersionID, testRuntimeInstanceID, testLaunchAttemptID},
		{"missing Configuration", testWorkspaceID, 0, testConfigurationVersionID, testRuntimeInstanceID, testLaunchAttemptID},
		{"missing ConfigurationVersion", testWorkspaceID, testConfigurationID, 0, testRuntimeInstanceID, testLaunchAttemptID},
		{"missing Runtime Instance", testWorkspaceID, testConfigurationID, testConfigurationVersionID, "", testLaunchAttemptID},
		{"missing Launch Attempt", testWorkspaceID, testConfigurationID, testConfigurationVersionID, testRuntimeInstanceID, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, err := NewLoadRequest(
				test.workspaceID,
				test.configurationID,
				test.versionID,
				test.runtimeInstanceID,
				test.launchAttemptIdentity,
			)
			if !errors.Is(err, ErrInvalidLoadRequest) {
				t.Fatalf("NewLoadRequest() error = %v, want ErrInvalidLoadRequest", err)
			}
			if request != (LoadRequest{}) {
				t.Fatalf("NewLoadRequest() request = %#v, want zero value", request)
			}
		})
	}
}

func TestLoadRejectsInvalidRequestBeforeSourceAccess(t *testing.T) {
	var calls atomic.Int32
	loader := New(sourceFunc(func(uint64, uint64, uint64) (SourceObservation, error) {
		calls.Add(1)
		return completeObservation(), nil
	}))

	result, err := loader.Load(LoadRequest{})
	if !errors.Is(err, ErrInvalidLoadRequest) {
		t.Fatalf("Load() error = %v, want ErrInvalidLoadRequest", err)
	}
	assertNoResult(t, result)
	if calls.Load() != 0 {
		t.Fatalf("LoadExact() calls = %d, want 0", calls.Load())
	}
}

func TestLoadSourceFailureContract(t *testing.T) {
	sourceFailure := errors.New("storage unavailable")
	tests := []struct {
		name string
		err  error
		want error
	}{
		{"Source Unavailable", ErrSourceUnavailable, ErrSourceUnavailable},
		{"Source Not Found", ErrSourceNotFound, ErrSourceNotFound},
		{"Identity Mismatch", ErrIdentityMismatch, ErrIdentityMismatch},
		{"Version Not Published", ErrVersionNotPublished, ErrVersionNotPublished},
		{"Inconsistent Source Observation", ErrInconsistentSourceObservation, ErrInconsistentSourceObservation},
		{"Source Integrity Failure", ErrSourceIntegrity, ErrSourceIntegrity},
		{"source-specific failure", sourceFailure, ErrSourceUnavailable},
		{"wrapped Source Not Found", fmt.Errorf("adapter: %w", ErrSourceNotFound), ErrSourceNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			loader := New(sourceFunc(func(uint64, uint64, uint64) (SourceObservation, error) {
				return SourceObservation{}, test.err
			}))
			result, err := loader.Load(mustLoadRequest(t))
			if !errors.Is(err, test.want) {
				t.Fatalf("Load() error = %v, want %v", err, test.want)
			}
			assertNoResult(t, result)
		})
	}
}

func TestLoadWithoutSourceReturnsSourceUnavailable(t *testing.T) {
	loaders := []*Loader{nil, New(nil)}
	for index, loader := range loaders {
		result, err := loader.Load(mustLoadRequest(t))
		if !errors.Is(err, ErrSourceUnavailable) {
			t.Fatalf("loaders[%d].Load() error = %v, want ErrSourceUnavailable", index, err)
		}
		assertNoResult(t, result)
	}
}

func TestLoadRejectsIdentityMismatch(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*SourceObservation)
	}{
		{"observation Workspace", func(value *SourceObservation) { value.WorkspaceID++ }},
		{"Configuration", func(value *SourceObservation) { value.Configuration.ID++ }},
		{"Configuration Workspace", func(value *SourceObservation) { value.Configuration.WorkspaceID++ }},
		{"ConfigurationVersion", func(value *SourceObservation) { value.ConfigurationVersion.ID++ }},
		{"ConfigurationVersion parent", func(value *SourceObservation) { value.ConfigurationVersion.ConfigurationID++ }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			observation := completeObservation()
			test.mutate(&observation)
			result, err := New(staticSource(observation)).Load(mustLoadRequest(t))
			if !errors.Is(err, ErrIdentityMismatch) {
				t.Fatalf("Load() error = %v, want ErrIdentityMismatch", err)
			}
			assertNoResult(t, result)
		})
	}
}

func TestLoadRejectsNonPublishedPinnedVersion(t *testing.T) {
	for _, state := range []configurationversion.VersionState{
		configurationversion.Draft,
		configurationversion.Validated,
		configurationversion.Archived,
	} {
		t.Run(string(state), func(t *testing.T) {
			observation := completeObservation()
			observation.ConfigurationVersion.State = state
			result, err := New(staticSource(observation)).Load(mustLoadRequest(t))
			if !errors.Is(err, ErrVersionNotPublished) {
				t.Fatalf("Load() error = %v, want ErrVersionNotPublished", err)
			}
			assertNoResult(t, result)
		})
	}
}

func TestLoadRejectsRepresentationIntegrityFailures(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*SourceObservation)
	}{
		{"incomplete representation", func(value *SourceObservation) { value.RepresentationComplete = false }},
		{"missing schema identity", func(value *SourceObservation) { value.SchemaIdentity = "" }},
		{"blank schema identity", func(value *SourceObservation) { value.SchemaIdentity = " \t " }},
		{"missing schema version", func(value *SourceObservation) { value.SchemaVersion = 0 }},
		{"missing version number", func(value *SourceObservation) { value.ConfigurationVersion.Number = 0 }},
		{"invalid lifecycle state", func(value *SourceObservation) { value.ConfigurationVersion.State = "Unknown" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			observation := completeObservation()
			test.mutate(&observation)
			result, err := New(staticSource(observation)).Load(mustLoadRequest(t))
			if !errors.Is(err, ErrSourceIntegrity) {
				t.Fatalf("Load() error = %v, want ErrSourceIntegrity", err)
			}
			assertNoResult(t, result)
		})
	}
}

func TestLoadNeverReselectsVersion(t *testing.T) {
	var calls atomic.Int32
	var requestedVersion atomic.Uint64
	loader := New(sourceFunc(func(_, _ uint64, versionID uint64) (SourceObservation, error) {
		calls.Add(1)
		requestedVersion.Store(versionID)
		return completeObservation(), nil
	}))

	result, err := loader.Load(mustLoadRequest(t))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("LoadExact() calls = %d, want 1", calls.Load())
	}
	if requestedVersion.Load() != testConfigurationVersionID {
		t.Fatalf("LoadExact() version = %d, want pinned %d", requestedVersion.Load(), testConfigurationVersionID)
	}
	if result.ConfigurationVersionID() != testConfigurationVersionID {
		t.Fatalf("result version = %d, want pinned %d", result.ConfigurationVersionID(), testConfigurationVersionID)
	}
}

func TestConcurrentPublicationCannotRedirectPinnedLoad(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var latestPublished atomic.Uint64
	var observedVersion atomic.Uint64
	latestPublished.Store(testConfigurationVersionID)

	loader := New(sourceFunc(func(_, _ uint64, versionID uint64) (SourceObservation, error) {
		close(entered)
		<-release
		observedVersion.Store(versionID)
		return completeObservation(), nil
	}))

	type loadOutcome struct {
		result DetachedLoadResult
		err    error
	}
	outcome := make(chan loadOutcome, 1)
	request := mustLoadRequest(t)
	go func() {
		result, err := loader.Load(request)
		outcome <- loadOutcome{result: result, err: err}
	}()

	<-entered
	latestPublished.Store(testConfigurationVersionID + 1)
	close(release)
	got := <-outcome
	if got.err != nil {
		t.Fatalf("Load() error = %v", got.err)
	}
	if observedVersion.Load() != testConfigurationVersionID {
		t.Fatalf("LoadExact() version = %d, want %d", observedVersion.Load(), testConfigurationVersionID)
	}
	if got.result.ConfigurationVersionID() != testConfigurationVersionID {
		t.Fatalf("result version = %d, want pinned %d", got.result.ConfigurationVersionID(), testConfigurationVersionID)
	}
	if latestPublished.Load() == got.result.ConfigurationVersionID() {
		t.Fatal("test did not establish a distinct concurrently published version")
	}
}

func TestDetachedLoadResultOwnsConfigurationMaterial(t *testing.T) {
	observation := completeObservation()
	loader := New(sourceFunc(func(uint64, uint64, uint64) (SourceObservation, error) {
		return observation, nil
	}))
	result, err := loader.Load(mustLoadRequest(t))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := result.ConfigurationVersion()

	callerCopy := result.ConfigurationVersion()
	mutateVersion(&callerCopy)
	if !reflect.DeepEqual(observation.ConfigurationVersion, want) {
		t.Fatal("caller mutation changed source-owned material")
	}
	if got := result.ConfigurationVersion(); !reflect.DeepEqual(got, want) {
		t.Fatalf("result changed after accessor result mutation:\n got  %#v\n want %#v", got, want)
	}

	mutateVersion(&observation.ConfigurationVersion)
	if got := result.ConfigurationVersion(); !reflect.DeepEqual(got, want) {
		t.Fatalf("result changed after source mutation:\n got  %#v\n want %#v", got, want)
	}
}

func TestEquivalentSourcesProduceEquivalentDetachedResults(t *testing.T) {
	firstObservation := completeObservation()
	secondObservation := completeObservation()
	first, err := New(staticSource(firstObservation)).Load(mustLoadRequest(t))
	if err != nil {
		t.Fatalf("first Load() error = %v", err)
	}
	second, err := New(staticSource(secondObservation)).Load(mustLoadRequest(t))
	if err != nil {
		t.Fatalf("second Load() error = %v", err)
	}

	assertSemanticallyEquivalent(t, first, second)
	firstSnapshot, err := runtimeconfig.NewBuilder().Build(first.ConfigurationVersion())
	if err != nil {
		t.Fatalf("Build(first) error = %v", err)
	}
	secondSnapshot, err := runtimeconfig.NewBuilder().Build(second.ConfigurationVersion())
	if err != nil {
		t.Fatalf("Build(second) error = %v", err)
	}
	if !reflect.DeepEqual(firstSnapshot, secondSnapshot) {
		t.Fatalf("Builder snapshots differ:\n first  %#v\n second %#v", firstSnapshot, secondSnapshot)
	}
}

func TestRepeatedAndConcurrentLoadsAreEquivalent(t *testing.T) {
	const loadCount = 64
	loader := New(staticSource(completeObservation()))
	request := mustLoadRequest(t)
	results := make([]DetachedLoadResult, loadCount)
	errs := make([]error, loadCount)
	var wait sync.WaitGroup

	for index := range results {
		wait.Add(1)
		go func() {
			defer wait.Done()
			results[index], errs[index] = loader.Load(request)
		}()
	}
	wait.Wait()

	for index := range results {
		if errs[index] != nil {
			t.Fatalf("Load()[%d] error = %v", index, errs[index])
		}
		assertSemanticallyEquivalent(t, results[0], results[index])
	}
}

func TestLoaderProductionDependenciesExcludePipelineAndInfrastructure(t *testing.T) {
	pkg, err := build.Default.ImportDir(".", 0)
	if err != nil {
		t.Fatalf("ImportDir() error = %v", err)
	}
	forbidden := map[string]bool{
		"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig": true,
		"github.com/dsdred/universal-websocket-platform/internal/runtime":       true,
		"net/http": true,
	}
	for _, imported := range pkg.Imports {
		if forbidden[imported] {
			t.Errorf("production package imports forbidden dependency %q", imported)
		}
	}
}

func TestHandoffContractIsNeutral(t *testing.T) {
	const contractPackage = "github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	if got := reflect.TypeFor[LoadRequest]().PkgPath(); got != contractPackage {
		t.Fatalf("LoadRequest package = %q, want %q", got, contractPackage)
	}
	if got := reflect.TypeFor[DetachedLoadResult]().PkgPath(); got != contractPackage {
		t.Fatalf("DetachedLoadResult package = %q, want %q", got, contractPackage)
	}

	pkg, err := build.Default.ImportDir("../runtimeconfigload", 0)
	if err != nil {
		t.Fatalf("ImportDir(runtimeconfigload) error = %v", err)
	}
	const loaderPackage = "github.com/dsdred/universal-websocket-platform/internal/configurationloader"
	for _, imported := range pkg.Imports {
		if imported == loaderPackage {
			t.Fatalf("neutral contract imports concrete Loader package %q", imported)
		}
	}

	var _ runtimeconfigload.LoadRequest = mustLoadRequest(t)
	var _ runtimeconfigload.DetachedLoadResult
}

func TestDetachedLoadResultExposesNoMutableFields(t *testing.T) {
	resultType := reflect.TypeFor[DetachedLoadResult]()
	for index := 0; index < resultType.NumField(); index++ {
		field := resultType.Field(index)
		if field.IsExported() {
			t.Errorf("DetachedLoadResult field %q is exported", field.Name)
		}
	}
}

func mustLoadRequest(t *testing.T) LoadRequest {
	t.Helper()
	request, err := NewLoadRequest(
		testWorkspaceID,
		testConfigurationID,
		testConfigurationVersionID,
		testRuntimeInstanceID,
		testLaunchAttemptID,
	)
	if err != nil {
		t.Fatalf("NewLoadRequest() error = %v", err)
	}
	return request
}

func completeObservation() SourceObservation {
	return SourceObservation{
		WorkspaceID: testWorkspaceID,
		Configuration: configuration.Configuration{
			ID:          testConfigurationID,
			WorkspaceID: testWorkspaceID,
			Name:        "production",
		},
		ConfigurationVersion: configurationversion.ConfigurationVersion{
			ID:              testConfigurationVersionID,
			ConfigurationID: testConfigurationID,
			Number:          7,
			State:           configurationversion.Published,
			Listener: configurationversion.ListenerSettings{
				Host: "127.0.0.1",
				Port: 8080,
				TLS: configurationversion.TLSSettings{
					Enabled:        true,
					CertificateRef: "secret://tls/certificate",
					PrivateKeyRef:  "secret://tls/private-key",
					MinVersion:     "1.3",
				},
				Timeouts: configurationversion.TimeoutSettings{
					HandshakeSeconds: 10,
					ReadSeconds:      20,
					WriteSeconds:     30,
					IdleSeconds:      40,
				},
			},
			Authentication: configurationversion.AuthenticationSettings{
				Enabled: true,
				Providers: []configurationversion.AuthenticationProvider{
					{
						Name:     "jwt",
						Type:     configurationversion.AuthenticationProviderJWT,
						Enabled:  true,
						Priority: 1,
						JWT: &configurationversion.JWTSettings{
							SigningKeys: []configurationversion.JWTSigningKey{
								{Name: "primary", SecretRef: "secret://jwt/key"},
							},
							AllowedAlgorithms: []configurationversion.JWTAlgorithm{configurationversion.RS256},
							AllowedIssuers:    []string{"issuer"},
							AllowedAudiences:  []string{"audience"},
							RequiredClaims: []configurationversion.JWTRequiredClaim{
								{Name: "tenant", Value: "one"},
							},
							ClockSkewSeconds: 5,
						},
					},
					{
						Name:     "api",
						Type:     configurationversion.AuthenticationProviderAPIKey,
						Enabled:  true,
						Priority: 2,
						APIKey: &configurationversion.APIKeySettings{
							Header:    "X-API-Key",
							SecretRef: "secret://api/key",
						},
					},
					{
						Name:     "basic",
						Type:     configurationversion.AuthenticationProviderBasic,
						Enabled:  true,
						Priority: 3,
						Basic: &configurationversion.BasicSettings{
							Realm:     "runtime",
							SecretRef: "secret://basic/users",
						},
					},
				},
			},
			Routing: &configurationversion.RoutingSettings{
				Routes: []configurationversion.Route{
					{
						ID:       "authenticated-text",
						Enabled:  true,
						Priority: 1,
						Matchers: []configurationversion.Matcher{
							{Type: configurationversion.MatcherTypeMessageType, Value: "text"},
							{Type: configurationversion.MatcherTypePrincipalKind, Value: "authenticated"},
						},
						HandlerRef: "legacy",
					},
				},
				DefaultHandlerRef: "legacy",
			},
			CreatedAt: time.Unix(100, 0).UTC(),
			UpdatedAt: time.Unix(200, 0).UTC(),
		},
		SchemaIdentity:         "uwp.configuration",
		SchemaVersion:          1,
		RepresentationComplete: true,
	}
}

func staticSource(observation SourceObservation) Source {
	return sourceFunc(func(uint64, uint64, uint64) (SourceObservation, error) {
		return observation, nil
	})
}

func assertResultFacts(t *testing.T, result DetachedLoadResult) {
	t.Helper()
	if result.WorkspaceID() != testWorkspaceID ||
		result.ConfigurationID() != testConfigurationID ||
		result.ConfigurationVersionID() != testConfigurationVersionID ||
		result.ConfigurationVersionNumber() != 7 ||
		!result.Published() ||
		result.SchemaIdentity() != "uwp.configuration" ||
		result.SchemaVersion() != 1 ||
		result.RuntimeInstanceID() != testRuntimeInstanceID ||
		result.LaunchAttemptID() != testLaunchAttemptID {
		t.Fatalf("unexpected result facts: %#v", result)
	}
}

func assertNoResult(t *testing.T, result DetachedLoadResult) {
	t.Helper()
	if !reflect.DeepEqual(result, DetachedLoadResult{}) {
		t.Fatalf("failure returned partial result %#v", result)
	}
}

func assertSemanticallyEquivalent(t *testing.T, first, second DetachedLoadResult) {
	t.Helper()
	if first.WorkspaceID() != second.WorkspaceID() ||
		first.ConfigurationID() != second.ConfigurationID() ||
		first.ConfigurationVersionID() != second.ConfigurationVersionID() ||
		first.ConfigurationVersionNumber() != second.ConfigurationVersionNumber() ||
		first.Published() != second.Published() ||
		first.SchemaIdentity() != second.SchemaIdentity() ||
		first.SchemaVersion() != second.SchemaVersion() ||
		first.RuntimeInstanceID() != second.RuntimeInstanceID() ||
		first.LaunchAttemptID() != second.LaunchAttemptID() ||
		!reflect.DeepEqual(first.ConfigurationVersion(), second.ConfigurationVersion()) {
		t.Fatalf("results are not semantically equivalent:\n first  %#v\n second %#v", first, second)
	}
}

func mutateVersion(version *configurationversion.ConfigurationVersion) {
	version.Listener.Host = "mutated"
	version.Authentication.Providers[0].Name = "mutated"
	version.Authentication.Providers[0].JWT.SigningKeys[0].Name = "mutated"
	version.Authentication.Providers[0].JWT.AllowedAlgorithms[0] = configurationversion.HS256
	version.Authentication.Providers[0].JWT.AllowedIssuers[0] = "mutated"
	version.Authentication.Providers[0].JWT.AllowedAudiences[0] = "mutated"
	version.Authentication.Providers[0].JWT.RequiredClaims[0].Value = "mutated"
	version.Authentication.Providers[1].APIKey.Header = "Mutated"
	version.Authentication.Providers[2].Basic.Realm = "mutated"
	version.Routing.Routes[0].ID = "mutated"
	version.Routing.Routes[0].Matchers[0].Value = "binary"
}
