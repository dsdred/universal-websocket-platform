package configurationloadsource

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/configuration"
	"github.com/dsdred/universal-websocket-platform/internal/configurationloader"
	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	runtimeplatform "github.com/dsdred/universal-websocket-platform/internal/runtime"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelaunchflow"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle"
)

const (
	testWorkspaceID     = uint64(11)
	testConfigurationID = uint64(22)
	testVersionID       = uint64(33)
)

func TestMemorySourceNilBindingsReturnBareUnavailable(t *testing.T) {
	t.Parallel()

	configurations := configuration.NewMemoryConfigurationRepository()
	versions := configurationversion.NewMemoryConfigurationVersionRepository()
	cases := map[string]*MemorySource{
		"nil receiver":       nil,
		"nil repositories":   NewMemorySource(nil, nil),
		"nil configurations": NewMemorySource(nil, versions),
		"nil versions":       NewMemorySource(configurations, nil),
		"nil private getter": {getConfiguration: configurations.Get},
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			observation, err := source.LoadExact(testWorkspaceID, testConfigurationID, testVersionID)
			if err != configurationloader.ErrSourceUnavailable {
				t.Fatalf("LoadExact() error = %v, want bare ErrSourceUnavailable", err)
			}
			if !reflect.DeepEqual(observation, configurationloader.SourceObservation{}) {
				t.Fatalf("LoadExact() observation = %#v, want zero", observation)
			}
		})
	}
}

func TestNewMemorySourceHasNoRepositorySideEffects(t *testing.T) {
	t.Parallel()

	configurations := configuration.NewMemoryConfigurationRepository()
	versions := configurationversion.NewMemoryConfigurationVersionRepository()
	source := NewMemorySource(configurations, versions)
	if source == nil {
		t.Fatal("NewMemorySource() returned nil")
	}

	parent, err := configurations.Create(configuration.Configuration{WorkspaceID: testWorkspaceID})
	if err != nil {
		t.Fatalf("configuration Create() error = %v", err)
	}
	version, err := versions.Create(configurationversion.ConfigurationVersion{
		ConfigurationID: parent.ID,
		Number:          1,
		State:           configurationversion.Published,
	})
	if err != nil {
		t.Fatalf("version Create() error = %v", err)
	}
	if parent.ID != 1 || version.ID != 1 {
		t.Fatalf("constructor changed repository state: parent=%d version=%d", parent.ID, version.ID)
	}
}

func TestMemorySourceSuccessUsesVersionFirstExactlyOnce(t *testing.T) {
	t.Parallel()

	var calls []string
	source := &MemorySource{
		getConfigurationVersion: func(id uint64) (configurationversion.ConfigurationVersion, error) {
			calls = append(calls, "version")
			if id != testVersionID {
				t.Fatalf("version Get(%d), want %d", id, testVersionID)
			}
			return completeVersion(), nil
		},
		getConfiguration: func(id uint64) (configuration.Configuration, error) {
			calls = append(calls, "configuration")
			if id != testConfigurationID {
				t.Fatalf("configuration Get(%d), want %d", id, testConfigurationID)
			}
			return completeConfiguration(), nil
		},
	}

	observation, err := source.LoadExact(testWorkspaceID, testConfigurationID, testVersionID)
	if err != nil {
		t.Fatalf("LoadExact() error = %v", err)
	}
	if !reflect.DeepEqual(calls, []string{"version", "configuration"}) {
		t.Fatalf("call order = %v, want [version configuration]", calls)
	}
	if observation.WorkspaceID != testWorkspaceID ||
		observation.Configuration != completeConfiguration() ||
		!reflect.DeepEqual(observation.ConfigurationVersion, completeVersion()) ||
		observation.SchemaIdentity != "uwp.configuration" ||
		observation.SchemaVersion != 1 ||
		!observation.RepresentationComplete {
		t.Fatalf("LoadExact() observation = %#v, want exact complete observation", observation)
	}
}

func TestMemorySourceRejectsVersionBeforeParentLookup(t *testing.T) {
	t.Parallel()

	rawFailure := errors.New("raw version failure")
	cases := []struct {
		name    string
		version configurationversion.ConfigurationVersion
		err     error
		want    error
	}{
		{name: "not found", err: configurationversion.ErrConfigurationVersionNotFound, want: configurationloader.ErrSourceNotFound},
		{name: "wrapped not found", err: errors.Join(rawFailure, configurationversion.ErrConfigurationVersionNotFound), want: configurationloader.ErrSourceNotFound},
		{name: "unavailable", err: rawFailure, want: configurationloader.ErrSourceUnavailable},
		{name: "version identity", version: withVersion(func(v *configurationversion.ConfigurationVersion) { v.ID++ }), want: configurationloader.ErrIdentityMismatch},
		{name: "parent identity", version: withVersion(func(v *configurationversion.ConfigurationVersion) { v.ConfigurationID++ }), want: configurationloader.ErrIdentityMismatch},
		{name: "draft", version: withVersion(func(v *configurationversion.ConfigurationVersion) { v.State = configurationversion.Draft; v.Number = 0 }), want: configurationloader.ErrVersionNotPublished},
		{name: "validated", version: withVersion(func(v *configurationversion.ConfigurationVersion) { v.State = configurationversion.Validated }), want: configurationloader.ErrVersionNotPublished},
		{name: "archived", version: withVersion(func(v *configurationversion.ConfigurationVersion) { v.State = configurationversion.Archived }), want: configurationloader.ErrVersionNotPublished},
		{name: "unknown state", version: withVersion(func(v *configurationversion.ConfigurationVersion) { v.State = "future" }), want: configurationloader.ErrSourceIntegrity},
		{name: "zero number", version: withVersion(func(v *configurationversion.ConfigurationVersion) { v.Number = 0 }), want: configurationloader.ErrSourceIntegrity},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			parentCalls := 0
			versionCalls := 0
			source := &MemorySource{
				getConfigurationVersion: func(uint64) (configurationversion.ConfigurationVersion, error) {
					versionCalls++
					return test.version, test.err
				},
				getConfiguration: func(uint64) (configuration.Configuration, error) {
					parentCalls++
					return completeConfiguration(), nil
				},
			}
			_, err := source.LoadExact(testWorkspaceID, testConfigurationID, testVersionID)
			if err != test.want {
				t.Fatalf("LoadExact() error = %v, want bare %v", err, test.want)
			}
			if versionCalls != 1 || parentCalls != 0 {
				t.Fatalf("calls = version:%d parent:%d, want 1:0", versionCalls, parentCalls)
			}
			if errors.Is(err, rawFailure) {
				t.Fatalf("LoadExact() leaked raw failure through %v", err)
			}
		})
	}
}

func TestMemorySourceMapsParentFailuresWithoutRetry(t *testing.T) {
	t.Parallel()

	rawFailure := errors.New("raw parent failure")
	cases := []struct {
		name   string
		parent configuration.Configuration
		err    error
		want   error
	}{
		{name: "not found", err: configuration.ErrConfigurationNotFound, want: configurationloader.ErrSourceNotFound},
		{name: "wrapped not found", err: errors.Join(rawFailure, configuration.ErrConfigurationNotFound), want: configurationloader.ErrSourceNotFound},
		{name: "unavailable", err: rawFailure, want: configurationloader.ErrSourceUnavailable},
		{name: "configuration identity", parent: withConfiguration(func(c *configuration.Configuration) { c.ID++ }), want: configurationloader.ErrIdentityMismatch},
		{name: "workspace identity", parent: withConfiguration(func(c *configuration.Configuration) { c.WorkspaceID++ }), want: configurationloader.ErrIdentityMismatch},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			versionCalls := 0
			parentCalls := 0
			source := &MemorySource{
				getConfigurationVersion: func(uint64) (configurationversion.ConfigurationVersion, error) {
					versionCalls++
					return completeVersion(), nil
				},
				getConfiguration: func(uint64) (configuration.Configuration, error) {
					parentCalls++
					return test.parent, test.err
				},
			}
			_, err := source.LoadExact(testWorkspaceID, testConfigurationID, testVersionID)
			if err != test.want {
				t.Fatalf("LoadExact() error = %v, want bare %v", err, test.want)
			}
			if versionCalls != 1 || parentCalls != 1 {
				t.Fatalf("calls = version:%d parent:%d, want 1:1", versionCalls, parentCalls)
			}
			if errors.Is(err, rawFailure) {
				t.Fatalf("LoadExact() leaked raw failure through %v", err)
			}
		})
	}
}

func TestMemorySourceDeeplyDetachesInBothDirections(t *testing.T) {
	t.Parallel()

	repositoryValue := completeVersion()
	untouchedBaseline := completeVersion()
	source := &MemorySource{
		getConfigurationVersion: func(uint64) (configurationversion.ConfigurationVersion, error) {
			return repositoryValue, nil
		},
		getConfiguration: func(uint64) (configuration.Configuration, error) {
			return completeConfiguration(), nil
		},
	}

	first, err := source.LoadExact(testWorkspaceID, testConfigurationID, testVersionID)
	if err != nil {
		t.Fatalf("first LoadExact() error = %v", err)
	}
	mutateVersionMutableContent(&repositoryValue, "changed-input", configurationversion.HS512)
	if !reflect.DeepEqual(first.ConfigurationVersion, untouchedBaseline) {
		t.Fatalf(
			"repository-side mutation changed observation:\ngot  %#v\nwant %#v",
			first.ConfigurationVersion,
			untouchedBaseline,
		)
	}

	expectedRepositoryBaseline := completeVersion()
	mutateVersionMutableContent(
		&expectedRepositoryBaseline,
		"changed-input",
		configurationversion.HS512,
	)
	mutateVersionMutableContent(
		&first.ConfigurationVersion,
		"changed-output",
		configurationversion.PS512,
	)
	if !reflect.DeepEqual(repositoryValue, expectedRepositoryBaseline) {
		t.Fatalf(
			"caller-side mutation changed repository value:\ngot  %#v\nwant %#v",
			repositoryValue,
			expectedRepositoryBaseline,
		)
	}

	second, err := source.LoadExact(testWorkspaceID, testConfigurationID, testVersionID)
	if err != nil {
		t.Fatalf("second LoadExact() error = %v", err)
	}
	if !reflect.DeepEqual(second.ConfigurationVersion, expectedRepositoryBaseline) {
		t.Fatalf(
			"caller mutation affected later observation:\ngot  %#v\nwant %#v",
			second.ConfigurationVersion,
			expectedRepositoryBaseline,
		)
	}
}

func TestMemorySourcePreservesNilNestedCollections(t *testing.T) {
	t.Parallel()

	version := completeVersion()
	version.Authentication.Providers = nil
	version.Routing = &configurationversion.RoutingSettings{Routes: nil}
	source := sourceFor(version, completeConfiguration())
	observation, err := source.LoadExact(testWorkspaceID, testConfigurationID, testVersionID)
	if err != nil {
		t.Fatalf("LoadExact() error = %v", err)
	}
	if observation.ConfigurationVersion.Authentication.Providers != nil {
		t.Fatal("nil Providers became non-nil")
	}
	if observation.ConfigurationVersion.Routing == nil ||
		observation.ConfigurationVersion.Routing.Routes != nil {
		t.Fatal("nil Routes was not preserved")
	}

	version = completeVersion()
	version.Authentication.Providers[2].JWT.AllowedIssuers = nil
	version.Routing.Routes[0].Matchers = nil
	observation, err = sourceFor(version, completeConfiguration()).LoadExact(testWorkspaceID, testConfigurationID, testVersionID)
	if err != nil {
		t.Fatalf("LoadExact() error = %v", err)
	}
	if observation.ConfigurationVersion.Authentication.Providers[2].JWT.AllowedIssuers != nil ||
		observation.ConfigurationVersion.Routing.Routes[0].Matchers != nil {
		t.Fatal("nested nil slice was not preserved")
	}
}

func TestMemorySourcePreservesNonNilEmptyNestedCollections(t *testing.T) {
	t.Parallel()

	t.Run("providers and routes", func(t *testing.T) {
		version := completeVersion()
		version.Authentication.Providers = make([]configurationversion.AuthenticationProvider, 0)
		version.Routing.Routes = make([]configurationversion.Route, 0)

		observation, err := sourceFor(version, completeConfiguration()).LoadExact(
			testWorkspaceID,
			testConfigurationID,
			testVersionID,
		)
		if err != nil {
			t.Fatalf("LoadExact() error = %v", err)
		}
		if observation.ConfigurationVersion.Authentication.Providers == nil {
			t.Fatal("non-nil empty Providers became nil")
		}
		if observation.ConfigurationVersion.Routing == nil ||
			observation.ConfigurationVersion.Routing.Routes == nil {
			t.Fatal("non-nil empty Routes became nil")
		}
	})

	t.Run("JWT slices and matchers", func(t *testing.T) {
		version := completeVersion()
		jwt := version.Authentication.Providers[2].JWT
		jwt.SigningKeys = make([]configurationversion.JWTSigningKey, 0)
		jwt.AllowedAlgorithms = make([]configurationversion.JWTAlgorithm, 0)
		jwt.AllowedIssuers = make([]string, 0)
		jwt.AllowedAudiences = make([]string, 0)
		jwt.RequiredClaims = make([]configurationversion.JWTRequiredClaim, 0)
		version.Routing.Routes[0].Matchers = make([]configurationversion.Matcher, 0)

		observation, err := sourceFor(version, completeConfiguration()).LoadExact(
			testWorkspaceID,
			testConfigurationID,
			testVersionID,
		)
		if err != nil {
			t.Fatalf("LoadExact() error = %v", err)
		}
		gotJWT := observation.ConfigurationVersion.Authentication.Providers[2].JWT
		if gotJWT == nil {
			t.Fatal("JWT pointer became nil")
		}
		if gotJWT.SigningKeys == nil ||
			gotJWT.AllowedAlgorithms == nil ||
			gotJWT.AllowedIssuers == nil ||
			gotJWT.AllowedAudiences == nil ||
			gotJWT.RequiredClaims == nil {
			t.Fatalf("non-nil empty JWT slice became nil: %#v", gotJWT)
		}
		if observation.ConfigurationVersion.Routing.Routes[0].Matchers == nil {
			t.Fatal("non-nil empty Matchers became nil")
		}
	})
}

func TestMemorySourceRepeatedAndConcurrentLoadsAreEquivalent(t *testing.T) {
	configurations, versions, parent, version := populatedRepositories(t)
	source := NewMemorySource(configurations, versions)

	want, err := source.LoadExact(parent.WorkspaceID, parent.ID, version.ID)
	if err != nil {
		t.Fatalf("initial LoadExact() error = %v", err)
	}
	const workers = 32
	results := make(chan configurationloader.SourceObservation, workers)
	failures := make(chan error, workers)
	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			result, loadErr := source.LoadExact(parent.WorkspaceID, parent.ID, version.ID)
			if loadErr != nil {
				failures <- loadErr
				return
			}
			results <- result
		}()
	}
	group.Wait()
	close(results)
	close(failures)
	for failure := range failures {
		t.Errorf("concurrent LoadExact() error = %v", failure)
	}
	for result := range results {
		if !reflect.DeepEqual(result, want) {
			t.Errorf("concurrent result differs:\ngot  %#v\nwant %#v", result, want)
		}
	}
}

func TestMemorySourceLifecycleBoundariesAtLAndC(t *testing.T) {
	t.Parallel()

	t.Run("archive before L", func(t *testing.T) {
		version := completeVersion()
		version.State = configurationversion.Archived
		_, err := sourceFor(version, completeConfiguration()).LoadExact(testWorkspaceID, testConfigurationID, testVersionID)
		if err != configurationloader.ErrVersionNotPublished {
			t.Fatalf("LoadExact() error = %v, want ErrVersionNotPublished", err)
		}
	})
	t.Run("archive after L", func(t *testing.T) {
		version := completeVersion()
		source := &MemorySource{
			getConfigurationVersion: func(uint64) (configurationversion.ConfigurationVersion, error) {
				return version, nil
			},
			getConfiguration: func(uint64) (configuration.Configuration, error) {
				version.State = configurationversion.Archived
				return completeConfiguration(), nil
			},
		}
		observation, err := source.LoadExact(testWorkspaceID, testConfigurationID, testVersionID)
		if err != nil {
			t.Fatalf("LoadExact() error = %v", err)
		}
		if observation.ConfigurationVersion.State != configurationversion.Published {
			t.Fatalf("observed state = %s, want detached Published at L", observation.ConfigurationVersion.State)
		}
	})
	t.Run("delete before C", func(t *testing.T) {
		source := &MemorySource{
			getConfigurationVersion: func(uint64) (configurationversion.ConfigurationVersion, error) {
				return completeVersion(), nil
			},
			getConfiguration: func(uint64) (configuration.Configuration, error) {
				return configuration.Configuration{}, configuration.ErrConfigurationNotFound
			},
		}
		_, err := source.LoadExact(testWorkspaceID, testConfigurationID, testVersionID)
		if err != configurationloader.ErrSourceNotFound {
			t.Fatalf("LoadExact() error = %v, want ErrSourceNotFound", err)
		}
	})
	t.Run("delete after C", func(t *testing.T) {
		deleted := false
		source := &MemorySource{
			getConfigurationVersion: func(uint64) (configurationversion.ConfigurationVersion, error) {
				return completeVersion(), nil
			},
			getConfiguration: func(uint64) (configuration.Configuration, error) {
				parent := completeConfiguration()
				deleted = true
				return parent, nil
			},
		}
		_, err := source.LoadExact(testWorkspaceID, testConfigurationID, testVersionID)
		if err != nil || !deleted {
			t.Fatalf("LoadExact() error = %v, deleted=%v; want success after C", err, deleted)
		}
	})
}

func TestMemorySourceIntegratesWithLoader(t *testing.T) {
	t.Parallel()

	configurations, versions, parent, version := populatedRepositories(t)
	loader := configurationloader.New(NewMemorySource(configurations, versions))
	request, err := configurationloader.NewLoadRequest(
		parent.WorkspaceID,
		parent.ID,
		version.ID,
		runtimeconfigload.RuntimeInstanceID("runtime-1"),
		runtimeconfigload.LaunchAttemptID("attempt-1"),
	)
	if err != nil {
		t.Fatalf("NewLoadRequest() error = %v", err)
	}
	result, err := loader.Load(request)
	if err != nil {
		t.Fatalf("Loader.Load() error = %v", err)
	}
	if result.WorkspaceID() != parent.WorkspaceID ||
		result.ConfigurationID() != parent.ID ||
		result.ConfigurationVersionID() != version.ID ||
		result.ConfigurationVersionNumber() != version.Number ||
		!result.Published() ||
		result.SchemaIdentity() != "uwp.configuration" ||
		result.SchemaVersion() != 1 {
		t.Fatalf("Loader.Load() result does not preserve exact source facts: %#v", result)
	}
}

func TestMemorySourceLoaderOwnerFlowConstructWithoutActivation(t *testing.T) {
	t.Parallel()

	configurations, versions, parent, _ := populatedRepositories(t)
	source := NewMemorySource(configurations, versions)
	loader := configurationloader.New(source)
	owner, err := runtimelifecycle.NewOwner(
		parent.WorkspaceID,
		parent.ID,
		runtimeconfigload.RuntimeInstanceID("runtime-1"),
		func() (runtimeconfigload.LaunchAttemptID, error) {
			return runtimeconfigload.LaunchAttemptID("attempt-1"), nil
		},
		&runtimeplatform.DependencyBindings{},
	)
	if err != nil {
		t.Fatalf("NewOwner() error = %v", err)
	}
	flow, err := runtimelaunchflow.New(owner, loader)
	if err != nil || flow == nil {
		t.Fatalf("Flow New() = (%v, %v), want non-nil without Start/Host", flow, err)
	}
	if observation := owner.Observe(); observation.ActualState() != runtimelifecycle.ActualStopped {
		t.Fatalf("construction activated Owner: state = %s", observation.ActualState())
	}
}

func TestSingleVersionServiceProtectsPublishedState(t *testing.T) {
	configurations := configuration.NewMemoryConfigurationRepository()
	parent, err := configurations.Create(configuration.Configuration{WorkspaceID: testWorkspaceID})
	if err != nil {
		t.Fatalf("configuration Create() error = %v", err)
	}
	versions := configurationversion.NewMemoryConfigurationVersionRepository()
	now := func() time.Time { return time.Unix(100, 0) }
	service := configurationversion.NewService(versions, configurations, now)
	version, err := service.Create(context.Background(), parent.WorkspaceID, parent.ID)
	if err != nil {
		t.Fatalf("Service.Create() error = %v", err)
	}

	start := make(chan struct{})
	var publishErr, updateErr error
	var group sync.WaitGroup
	group.Add(2)
	go func() {
		defer group.Done()
		<-start
		_, publishErr = service.Publish(context.Background(), parent.WorkspaceID, parent.ID, version.ID)
	}()
	go func() {
		defer group.Done()
		<-start
		_, updateErr = service.UpdateAuthentication(
			context.Background(),
			parent.WorkspaceID,
			parent.ID,
			version.ID,
			configurationversion.AuthenticationSettings{Providers: []configurationversion.AuthenticationProvider{}},
		)
	}()
	close(start)
	group.Wait()
	if publishErr != nil {
		t.Fatalf("Publish() error = %v", publishErr)
	}
	if updateErr != nil && !errors.Is(updateErr, configurationversion.ErrVersionNotEditable) {
		t.Fatalf("UpdateAuthentication() error = %v, want nil or ErrVersionNotEditable", updateErr)
	}
	stored, err := versions.Get(version.ID)
	if err != nil {
		t.Fatalf("version Get() error = %v", err)
	}
	if stored.State != configurationversion.Published {
		t.Fatalf("stale update overwrote Published state: %s", stored.State)
	}
	_, err = service.UpdateAuthentication(
		context.Background(),
		parent.WorkspaceID,
		parent.ID,
		version.ID,
		configurationversion.AuthenticationSettings{},
	)
	if !errors.Is(err, configurationversion.ErrVersionNotEditable) {
		t.Fatalf("post-publish update error = %v, want ErrVersionNotEditable", err)
	}
}

func TestConfigurationServicePreservesIdentityAndCannotResurrectDeletedParent(t *testing.T) {
	t.Parallel()

	repository := configuration.NewMemoryConfigurationRepository()
	service := configuration.NewService(repository, existingWorkspace{}, func() time.Time {
		return time.Unix(100, 0)
	})
	parent, err := service.Create(
		context.Background(),
		testWorkspaceID,
		configuration.CreateConfiguration{Name: "original"},
	)
	if err != nil {
		t.Fatalf("Service.Create() error = %v", err)
	}
	updated, err := service.Update(
		context.Background(),
		parent.WorkspaceID,
		parent.ID,
		configuration.UpdateConfiguration{Name: "updated"},
	)
	if err != nil {
		t.Fatalf("Service.Update() error = %v", err)
	}
	if updated.ID != parent.ID || updated.WorkspaceID != parent.WorkspaceID {
		t.Fatalf("Update changed identity: got (%d,%d), want (%d,%d)",
			updated.ID, updated.WorkspaceID, parent.ID, parent.WorkspaceID)
	}
	if err := service.Delete(context.Background(), parent.WorkspaceID, parent.ID); err != nil {
		t.Fatalf("Service.Delete() error = %v", err)
	}
	if _, err := repository.Update(updated); !errors.Is(err, configuration.ErrConfigurationNotFound) {
		t.Fatalf("stale repository Update() error = %v, want ErrConfigurationNotFound", err)
	}
	if _, err := repository.Get(parent.ID); !errors.Is(err, configuration.ErrConfigurationNotFound) {
		t.Fatalf("deleted parent was resurrected: Get() error = %v", err)
	}
}

func TestMemorySourceSurfaceAndImportsRemainConfined(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}
	sourcePath := filepath.Join(filepath.Dir(currentFile), "source.go")
	file, err := parser.ParseFile(token.NewFileSet(), sourcePath, nil, parser.AllErrors)
	if err != nil {
		t.Fatalf("ParseFile(source.go) error = %v", err)
	}

	var exported []string
	var selectors []string
	ast.Inspect(file, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.TypeSpec:
			if value.Name.IsExported() {
				exported = append(exported, value.Name.Name)
			}
		case *ast.FuncDecl:
			if value.Name.IsExported() {
				exported = append(exported, value.Name.Name)
			}
		case *ast.SelectorExpr:
			selectors = append(selectors, value.Sel.Name)
		}
		return true
	})
	sort.Strings(exported)
	if want := []string{"LoadExact", "MemorySource", "NewMemorySource"}; !reflect.DeepEqual(exported, want) {
		t.Fatalf("exported surface = %v, want %v", exported, want)
	}

	var imports []string
	for _, spec := range file.Imports {
		imports = append(imports, spec.Path.Value)
	}
	sort.Strings(imports)
	wantImports := []string{
		`"errors"`,
		`"github.com/dsdred/universal-websocket-platform/internal/configuration"`,
		`"github.com/dsdred/universal-websocket-platform/internal/configurationloader"`,
		`"github.com/dsdred/universal-websocket-platform/internal/configurationversion"`,
	}
	sort.Strings(wantImports)
	if !reflect.DeepEqual(imports, wantImports) {
		t.Fatalf("production imports = %v, want %v", imports, wantImports)
	}
	for _, forbidden := range []string{
		"Create", "Update", "UpdateBatch", "Delete", "GetPublished",
		"ListByConfiguration", "Lock", "RLock", "Go",
	} {
		if contains(selectors, forbidden) {
			t.Fatalf("source.go uses forbidden selector %q", forbidden)
		}
	}
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("ReadFile(source.go) error = %v", err)
	}
	for _, forbidden := range []string{
		"ErrInconsistentSourceObservation",
		"sync.", "context.", "runtimeconfig.", "go func", "interface {",
	} {
		if bytesContain(content, forbidden) {
			t.Fatalf("source.go contains forbidden production surface/behavior %q", forbidden)
		}
	}
}

type existingWorkspace struct{}

func (existingWorkspace) Exists(context.Context, uint64) (bool, error) {
	return true, nil
}

func populatedRepositories(
	t *testing.T,
) (
	*configuration.MemoryConfigurationRepository,
	*configurationversion.MemoryConfigurationVersionRepository,
	configuration.Configuration,
	configurationversion.ConfigurationVersion,
) {
	t.Helper()
	configurations := configuration.NewMemoryConfigurationRepository()
	parent, err := configurations.Create(configuration.Configuration{
		WorkspaceID: testWorkspaceID,
		Name:        "configuration",
	})
	if err != nil {
		t.Fatalf("configuration Create() error = %v", err)
	}
	versions := configurationversion.NewMemoryConfigurationVersionRepository()
	version := completeVersion()
	version.ID = 0
	version.ConfigurationID = parent.ID
	version, err = versions.Create(version)
	if err != nil {
		t.Fatalf("version Create() error = %v", err)
	}
	return configurations, versions, parent, version
}

func sourceFor(
	version configurationversion.ConfigurationVersion,
	parent configuration.Configuration,
) *MemorySource {
	return &MemorySource{
		getConfigurationVersion: func(uint64) (configurationversion.ConfigurationVersion, error) {
			return version, nil
		},
		getConfiguration: func(uint64) (configuration.Configuration, error) {
			return parent, nil
		},
	}
}

func completeConfiguration() configuration.Configuration {
	return configuration.Configuration{
		ID:          testConfigurationID,
		WorkspaceID: testWorkspaceID,
		Name:        "configuration",
		Description: "description",
		CreatedAt:   time.Unix(1, 0),
		UpdatedAt:   time.Unix(2, 0),
	}
}

func withConfiguration(change func(*configuration.Configuration)) configuration.Configuration {
	parent := completeConfiguration()
	change(&parent)
	return parent
}

func completeVersion() configurationversion.ConfigurationVersion {
	return configurationversion.ConfigurationVersion{
		ID:              testVersionID,
		ConfigurationID: testConfigurationID,
		Number:          7,
		State:           configurationversion.Published,
		Authentication: configurationversion.AuthenticationSettings{
			Enabled: true,
			Providers: []configurationversion.AuthenticationProvider{
				{
					Name:   "key",
					Type:   configurationversion.AuthenticationProviderAPIKey,
					APIKey: &configurationversion.APIKeySettings{Header: "X-Key", SecretRef: "key/ref"},
				},
				{
					Name:  "basic",
					Type:  configurationversion.AuthenticationProviderBasic,
					Basic: &configurationversion.BasicSettings{Realm: "realm", SecretRef: "basic/ref"},
				},
				{
					Name: "jwt",
					Type: configurationversion.AuthenticationProviderJWT,
					JWT: &configurationversion.JWTSettings{
						SigningKeys:       []configurationversion.JWTSigningKey{{Name: "key", SecretRef: "jwt/ref"}},
						AllowedAlgorithms: []configurationversion.JWTAlgorithm{configurationversion.RS256},
						AllowedIssuers:    []string{"issuer"},
						AllowedAudiences:  []string{"audience"},
						RequiredClaims:    []configurationversion.JWTRequiredClaim{{Name: "role", Value: "admin"}},
					},
				},
			},
		},
		Routing: &configurationversion.RoutingSettings{
			DefaultHandlerRef: "default",
			Routes: []configurationversion.Route{{
				ID:         "route",
				Enabled:    true,
				HandlerRef: "handler",
				Matchers: []configurationversion.Matcher{{
					Type:  configurationversion.MatcherTypeMessageType,
					Value: "event",
				}},
			}},
		},
		CreatedAt: time.Unix(1, 0),
		UpdatedAt: time.Unix(2, 0),
	}
}

func withVersion(
	change func(*configurationversion.ConfigurationVersion),
) configurationversion.ConfigurationVersion {
	version := completeVersion()
	change(&version)
	return version
}

func mutateVersionMutableContent(
	version *configurationversion.ConfigurationVersion,
	value string,
	algorithm configurationversion.JWTAlgorithm,
) {
	version.Authentication.Providers[0].Name = value
	version.Authentication.Providers[0].APIKey.Header = value
	version.Authentication.Providers[1].Basic.Realm = value
	version.Authentication.Providers[2].JWT.ClockSkewSeconds++
	version.Authentication.Providers[2].JWT.SigningKeys[0].Name = value
	version.Authentication.Providers[2].JWT.AllowedAlgorithms[0] = algorithm
	version.Authentication.Providers[2].JWT.AllowedIssuers[0] = value
	version.Authentication.Providers[2].JWT.AllowedAudiences[0] = value
	version.Authentication.Providers[2].JWT.RequiredClaims[0].Value = value
	version.Routing.DefaultHandlerRef = value
	version.Routing.Routes[0].HandlerRef = value
	version.Routing.Routes[0].Matchers[0].Value = value
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func bytesContain(content []byte, value string) bool {
	for index := 0; index+len(value) <= len(content); index++ {
		if string(content[index:index+len(value)]) == value {
			return true
		}
	}
	return false
}
