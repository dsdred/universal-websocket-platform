package runtimeconfig

import (
	"reflect"
	"sync"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

func TestBuilderBuildsCompleteCanonicalDetachedSnapshot(t *testing.T) {
	version := validConfigurationVersion()
	input := detachedResult(version)

	snapshot, diagnostics := NewBuilder().Build(input)
	if len(diagnostics) != 0 {
		t.Fatalf("Build() diagnostics = %#v", diagnostics)
	}
	if got, want := snapshot.Provenance(), (Provenance{
		WorkspaceID: 11, ConfigurationID: 22, ConfigurationVersionID: 33,
		ConfigurationVersionNumber: 4, SchemaIdentity: "uwp.configuration",
		SchemaVersion: 1, RuntimeInstanceID: "runtime-1", LaunchAttemptID: "attempt-1",
	}); got != want {
		t.Fatalf("Provenance() = %#v, want %#v", got, want)
	}
	if got, want := snapshot.Listener(), (ListenerSnapshot{
		Host: "Example.COM.",
		Port: 8080,
		TLS: TLSSnapshot{
			Enabled: false, CertificateRef: "secrets/tls/cert",
			PrivateKeyRef: "secrets/tls/key", MinVersion: "1.2",
		},
		Timeouts: TimeoutSnapshot{
			HandshakeSeconds: 10, ReadSeconds: 30, WriteSeconds: 20, IdleSeconds: 60,
		},
	}); got != want {
		t.Fatalf("Listener() = %#v, want %#v", got, want)
	}
	wantAuthentication := AuthenticationSnapshot{
		Enabled: true,
		Providers: []AuthenticationProviderSnapshot{
			{
				Name: "jwt-main", Type: AuthenticationProviderJWT, Enabled: true, Priority: 1,
				JWT: &JWTSnapshot{
					SigningKeys:       []JWTSigningKeySnapshot{{Name: "primary", SecretRef: "secrets/jwt/key"}},
					AllowedAlgorithms: []JWTAlgorithm{RS256},
					AllowedIssuers:    []string{"issuer"},
					AllowedAudiences:  []string{"audience"},
					RequiredClaims:    []JWTRequiredClaimSnapshot{{Name: "tenant", Value: "internal"}},
					ClockSkewSeconds:  30,
				},
			},
			{
				Name: "api-key", Type: AuthenticationProviderAPIKey, Enabled: true, Priority: 2,
				APIKey: &APIKeySnapshot{Header: "X-API-Key", SecretRef: "secrets/api-key"},
			},
			{
				Name: "basic", Type: AuthenticationProviderBasic, Enabled: false, Priority: 3,
				Basic: &BasicSnapshot{Realm: "Operations", SecretRef: "secrets/basic"},
			},
		},
	}
	authentication := snapshot.Authentication()
	if !reflect.DeepEqual(authentication, wantAuthentication) {
		t.Fatalf("Authentication() = %#v, want %#v", authentication, wantAuthentication)
	}
	routing, present := snapshot.Routing()
	if !present {
		t.Fatal("Routing() is absent")
	}
	defaultRef, defaultPresent := routing.DefaultHandlerRef()
	if !defaultPresent || defaultRef != "legacy" {
		t.Fatalf("DefaultHandlerRef() = (%q, %t)", defaultRef, defaultPresent)
	}
	routes := routing.Routes()
	wantRoutes := []RouteSnapshot{{
		id: "route", enabled: true, priority: 1, handlerRef: "legacy",
		matchers: []MatcherSnapshot{
			{matcherType: MatcherTypeMessageType, value: "text"},
			{matcherType: MatcherTypePrincipalKind, value: "authenticated"},
		},
	}}
	if !reflect.DeepEqual(routes, wantRoutes) {
		t.Fatalf("Routes() = %#v, want %#v", routes, wantRoutes)
	}

	second, secondDiagnostics := NewBuilder().Build(detachedResult(validConfigurationVersion()))
	if len(secondDiagnostics) != 0 {
		t.Fatalf("second Build diagnostics = %#v", secondDiagnostics)
	}
	version.Listener.Host = "changed"
	version.Authentication.Providers[0].JWT.AllowedIssuers[0] = "changed"
	version.Routing.Routes[0].Matchers[0].Value = "changed"
	authentication.Enabled = false
	authentication.Providers[0].Name = "reader mutation"
	authentication.Providers[0].Enabled = false
	authentication.Providers[0].Priority = 99
	authentication.Providers[0].JWT.SigningKeys[0].Name = "reader mutation"
	authentication.Providers[0].JWT.SigningKeys[0].SecretRef = "reader/mutation"
	authentication.Providers[0].JWT.AllowedAlgorithms[0] = HS256
	authentication.Providers[0].JWT.AllowedIssuers[0] = "reader mutation"
	authentication.Providers[0].JWT.AllowedAudiences[0] = "reader mutation"
	authentication.Providers[0].JWT.RequiredClaims[0].Name = "reader mutation"
	authentication.Providers[0].JWT.RequiredClaims[0].Value = "reader mutation"
	authentication.Providers[0].JWT.ClockSkewSeconds = 99
	authentication.Providers[1].APIKey.Header = "reader mutation"
	authentication.Providers[1].APIKey.SecretRef = "reader/mutation"
	authentication.Providers[2].Basic.Realm = "reader mutation"
	authentication.Providers[2].Basic.SecretRef = "reader/mutation"
	routes[0].id = "reader mutation"
	routes[0].enabled = false
	routes[0].priority = 99
	routes[0].handlerRef = "reader mutation"
	routes[0].matchers[0].matcherType = MatcherTypePrincipalKind
	routes[0].matchers[0].value = "reader mutation"
	if !reflect.DeepEqual(snapshot.Authentication(), wantAuthentication) ||
		!reflect.DeepEqual(second.Authentication(), wantAuthentication) ||
		!reflect.DeepEqual(mustRouting(t, snapshot).Routes(), wantRoutes) ||
		!reflect.DeepEqual(mustRouting(t, second).Routes(), wantRoutes) {
		t.Fatal("Snapshot aliases input or reader-owned memory")
	}
}

func TestBuilderReturnsCompleteDiagnosticsAndNoSnapshot(t *testing.T) {
	version := validConfigurationVersion()
	version.ConfigurationID = 0
	version.Listener.Host = " "
	version.Listener.Port = 0
	version.Authentication.Enabled = true
	version.Authentication.Providers = nil
	version.Routing = &configurationversion.RoutingSettings{
		DefaultHandlerRef: " ",
		Routes: []configurationversion.Route{{
			ID: " ", Enabled: true, Priority: 0, HandlerRef: " ", Matchers: nil,
		}},
	}
	request := runtimeconfigload.NewLoadRequest(0, 22, 33, "", "")
	input := runtimeconfigload.NewDetachedLoadResult(request, version, 9, false, "uwp.configuration", 1)

	snapshot, diagnostics := NewBuilder().Build(input)
	if !reflect.DeepEqual(snapshot, Snapshot{}) {
		t.Fatalf("failure returned partial Snapshot: %#v", snapshot)
	}
	if _, present := snapshot.Routing(); present {
		t.Fatal("failure returned an optional Routing aggregate")
	}
	if len(snapshot.Authentication().Providers) != 0 {
		t.Fatal("failure returned Authentication provider storage")
	}
	got := diagnosticCodes(diagnostics)
	want := []string{
		"snapshot.authentication.providers.required",
		"snapshot.configuration_version.configuration.required",
		"snapshot.configuration_version.number.inconsistent",
		"snapshot.handoff.not_published",
		"snapshot.provenance.launch_attempt.required",
		"snapshot.provenance.runtime_instance.required",
		"snapshot.provenance.workspace.required",
		"snapshot.routing.default_handler.required",
		"snapshot.routing.route.handler.required",
		"snapshot.routing.route.id.required",
		"snapshot.routing.route.matchers.required",
		"snapshot.routing.route.priority.out_of_range",
		"snapshot.listener.host.required",
		"snapshot.listener.port.out_of_range",
	}
	for _, code := range want {
		if !containsString(got, code) {
			t.Errorf("Diagnostics missing %s; got %v", code, got)
		}
	}
	for index := 1; index < len(diagnostics); index++ {
		if compareLocations(diagnostics[index-1].Location(), diagnostics[index].Location()) > 0 {
			t.Fatalf("Diagnostics are not location ordered: %#v", diagnostics)
		}
	}
}

func TestBuilderSuppressesCascadesAndAnchorsDuplicates(t *testing.T) {
	version := validConfigurationVersion()
	version.Authentication.Providers = []configurationversion.AuthenticationProvider{
		{Name: " same ", Type: "", Priority: 7, JWT: &configurationversion.JWTSettings{}},
		{Name: "same", Type: "future", Priority: 7, JWT: &configurationversion.JWTSettings{}},
	}
	version.Routing = &configurationversion.RoutingSettings{Routes: []configurationversion.Route{
		{ID: "route", Enabled: true, Priority: 1, HandlerRef: "legacy", Matchers: []configurationversion.Matcher{{Type: "message-type", Value: "text"}}},
		{ID: "route", Enabled: true, Priority: 1, HandlerRef: "legacy", Matchers: []configurationversion.Matcher{{Type: "message-type", Value: "text"}}},
	}}
	_, diagnostics := NewBuilder().Build(detachedResult(version))
	assertDiagnostic(t, diagnostics, "snapshot.authentication.provider.name.duplicate", "$.authentication.providers[1].name")
	assertDiagnostic(t, diagnostics, "snapshot.authentication.provider.priority.duplicate", "$.authentication.providers[1].priority")
	assertDiagnostic(t, diagnostics, "snapshot.authentication.provider.type.required", "$.authentication.providers[0].type")
	assertDiagnostic(t, diagnostics, "snapshot.authentication.provider.type.unsupported", "$.authentication.providers[1].type")
	assertNoDiagnosticAt(t, diagnostics, "snapshot.authentication.provider.settings.forbidden", "$.authentication.providers[0].jwt")
	assertDiagnostic(t, diagnostics, "snapshot.routing.route.id.duplicate", "$.routing.routes[1].id")
	assertDiagnostic(t, diagnostics, "snapshot.routing.route.priority.duplicate", "$.routing.routes[1].priority")
	assertDiagnostic(t, diagnostics, "snapshot.routing.matcher_set.duplicate", "$.routing.routes[1].matchers")
}

func TestBuilderReportsMissingExpectedAndEveryPresentWrongSettingsBranch(t *testing.T) {
	version := validConfigurationVersion()
	version.Authentication.Providers = []configurationversion.AuthenticationProvider{{
		Name: "provider", Type: configurationversion.AuthenticationProviderAPIKey, Enabled: true,
		Basic: &configurationversion.BasicSettings{},
		JWT:   &configurationversion.JWTSettings{},
	}}
	_, diagnostics := NewBuilder().Build(detachedResult(version))
	assertDiagnostic(t, diagnostics, "snapshot.authentication.provider.settings.required", "$.authentication.providers[0].apiKey")
	assertDiagnostic(t, diagnostics, "snapshot.authentication.provider.settings.forbidden", "$.authentication.providers[0].basic")
	assertDiagnostic(t, diagnostics, "snapshot.authentication.provider.settings.forbidden", "$.authentication.providers[0].jwt")
	assertNoDiagnosticAt(t, diagnostics, "snapshot.authentication.basic.realm.required", "$.authentication.providers[0].basic.realm")
	assertNoDiagnosticAt(t, diagnostics, "snapshot.authentication.jwt.signing_keys.required", "$.authentication.providers[0].jwt.signingKeys")
}

func TestBuilderSettingsRequiredUsesTypeSelectedBranch(t *testing.T) {
	tests := []struct {
		providerType configurationversion.AuthenticationProviderType
		branch       string
	}{
		{providerType: configurationversion.AuthenticationProviderAPIKey, branch: "apiKey"},
		{providerType: configurationversion.AuthenticationProviderBasic, branch: "basic"},
		{providerType: configurationversion.AuthenticationProviderJWT, branch: "jwt"},
	}
	for _, test := range tests {
		t.Run(string(test.providerType), func(t *testing.T) {
			version := validConfigurationVersion()
			version.Authentication.Providers = []configurationversion.AuthenticationProvider{{
				Name: "provider", Type: test.providerType, Enabled: true, Priority: 1,
			}}
			_, diagnostics := NewBuilder().Build(detachedResult(version))
			assertDiagnostic(
				t,
				diagnostics,
				"snapshot.authentication.provider.settings.required",
				"$.authentication.providers[0]."+test.branch,
			)
		})
	}
}

func TestBuilderRepeatedAndConcurrentBuildsAreIndependent(t *testing.T) {
	input := detachedResult(validConfigurationVersion())
	const count = 32
	results := make(chan Snapshot, count)
	var group sync.WaitGroup
	for index := 0; index < count; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			snapshot, diagnostics := NewBuilder().Build(input)
			if len(diagnostics) != 0 {
				t.Errorf("Build() diagnostics = %#v", diagnostics)
				return
			}
			results <- snapshot
		}()
	}
	group.Wait()
	close(results)
	for snapshot := range results {
		authentication := snapshot.Authentication()
		authentication.Providers[0].JWT.AllowedIssuers[0] = "mutated"
		if snapshot.Authentication().Providers[0].JWT.AllowedIssuers[0] != "issuer" {
			t.Fatal("Build result aliases reader memory")
		}
	}
}

func TestBuilderSuccessfulResultsShareNoPackageLocalMutableStorage(t *testing.T) {
	input := detachedResult(validConfigurationVersion())
	first, firstDiagnostics := NewBuilder().Build(input)
	second, secondDiagnostics := NewBuilder().Build(input)
	if len(firstDiagnostics) != 0 || len(secondDiagnostics) != 0 {
		t.Fatalf("Build diagnostics: first=%#v second=%#v", firstDiagnostics, secondDiagnostics)
	}
	wantAuthentication := second.Authentication()
	wantRouting := mustRouting(t, second)

	first.authentication.Providers[0].Name = "package-local mutation"
	first.authentication.Providers[0].JWT.SigningKeys[0].Name = "package-local mutation"
	first.authentication.Providers[0].JWT.AllowedAlgorithms[0] = HS256
	first.authentication.Providers[0].JWT.AllowedIssuers[0] = "package-local mutation"
	first.authentication.Providers[0].JWT.AllowedAudiences[0] = "package-local mutation"
	first.authentication.Providers[0].JWT.RequiredClaims[0].Value = "package-local mutation"
	first.authentication.Providers[1].APIKey.Header = "package-local mutation"
	first.authentication.Providers[2].Basic.Realm = "package-local mutation"
	first.routing.routes[0].id = "package-local mutation"
	first.routing.routes[0].matchers[0].matcherType = MatcherTypeAuthenticationType
	first.routing.routes[0].matchers[0].value = "jwt"

	if got := second.Authentication(); !reflect.DeepEqual(got, wantAuthentication) {
		t.Fatalf("second Authentication changed through first Build storage: got=%#v want=%#v", got, wantAuthentication)
	}
	if got := mustRouting(t, second); !reflect.DeepEqual(got, wantRouting) {
		t.Fatalf("second Routing changed through first Build storage: got=%#v want=%#v", got, wantRouting)
	}
}

func TestBuilderDiagnosticsAreIndependentAcrossBuilds(t *testing.T) {
	version := validConfigurationVersion()
	version.Listener.Host = ""
	input := detachedResult(version)

	_, first := NewBuilder().Build(input)
	_, second := NewBuilder().Build(input)
	if len(first) == 0 || !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated invalid Builds differ: first=%#v second=%#v", first, second)
	}
	first[0] = Diagnostic{}
	_, third := NewBuilder().Build(input)
	if !reflect.DeepEqual(second, third) {
		t.Fatalf("caller mutation changed later Diagnostics: second=%#v third=%#v", second, third)
	}
}

func TestBuilderSchemaFailureSuppressesOnlySectionRules(t *testing.T) {
	version := validConfigurationVersion()
	version.Listener.Host = ""
	request := runtimeconfigload.NewLoadRequest(0, 22, 33, "", "attempt-1")
	input := runtimeconfigload.NewDetachedLoadResult(request, version, version.Number, false, "", 0)

	_, diagnostics := NewBuilder().Build(input)
	assertDiagnostic(t, diagnostics, "snapshot.schema.identity.required", "$.provenance.schemaIdentity")
	assertDiagnostic(t, diagnostics, "snapshot.schema.version.required", "$.provenance.schemaVersion")
	assertDiagnostic(t, diagnostics, "snapshot.provenance.workspace.required", "$.provenance.workspaceId")
	assertDiagnostic(t, diagnostics, "snapshot.provenance.runtime_instance.required", "$.provenance.runtimeInstanceId")
	assertDiagnostic(t, diagnostics, "snapshot.handoff.not_published", "$.handoff.published")
	assertNoDiagnosticAt(t, diagnostics, "snapshot.listener.host.required", "$.listener.host")
}

func TestBuilderPayloadIdentityPrecedence(t *testing.T) {
	tests := []struct {
		name             string
		carried          uint64
		payload          uint64
		wantRequired     bool
		wantInconsistent bool
	}{
		{name: "both zero", wantRequired: true},
		{name: "carried only", carried: 33, wantRequired: true},
		{name: "payload only", payload: 33},
		{name: "unequal", carried: 33, payload: 44, wantInconsistent: true},
		{name: "equal", carried: 33, payload: 33},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version := validConfigurationVersion()
			version.ID = test.payload
			request := runtimeconfigload.NewLoadRequest(11, 22, test.carried, "runtime-1", "attempt-1")
			input := runtimeconfigload.NewDetachedLoadResult(
				request, version, version.Number, true, "uwp.configuration", 1,
			)
			_, diagnostics := NewBuilder().Build(input)
			hasRequired := hasDiagnostic(diagnostics, "snapshot.configuration_version.identity.required")
			hasCarriedRequired := hasDiagnostic(diagnostics, "snapshot.provenance.configuration_version.required")
			hasInconsistent := hasDiagnostic(diagnostics, "snapshot.configuration_version.identity.inconsistent")
			if hasRequired != test.wantRequired {
				t.Errorf("payload required = %t, want %t; diagnostics=%#v", hasRequired, test.wantRequired, diagnostics)
			}
			if hasCarriedRequired != (test.carried == 0) {
				t.Errorf("carried required = %t, want %t; diagnostics=%#v", hasCarriedRequired, test.carried == 0, diagnostics)
			}
			if hasInconsistent != test.wantInconsistent {
				t.Errorf("inconsistent = %t, want %t; diagnostics=%#v", hasInconsistent, test.wantInconsistent, diagnostics)
			}
		})
	}
}

func detachedResult(version configurationversion.ConfigurationVersion) runtimeconfigload.DetachedLoadResult {
	request := runtimeconfigload.NewLoadRequest(11, 22, 33, "runtime-1", "attempt-1")
	return runtimeconfigload.NewDetachedLoadResult(request, version, version.Number, true, "uwp.configuration", 1)
}

func validConfigurationVersion() configurationversion.ConfigurationVersion {
	return configurationversion.ConfigurationVersion{
		ID: 33, ConfigurationID: 22, Number: 4, State: configurationversion.Published,
		Listener: configurationversion.ListenerSettings{
			Host: " Example.COM. ", Port: 8080,
			TLS:      configurationversion.TLSSettings{CertificateRef: " secrets/tls/cert ", PrivateKeyRef: " secrets/tls/key ", MinVersion: " 1.2 "},
			Timeouts: configurationversion.TimeoutSettings{HandshakeSeconds: 10, ReadSeconds: 30, WriteSeconds: 20, IdleSeconds: 60},
		},
		Authentication: configurationversion.AuthenticationSettings{
			Enabled: true,
			Providers: []configurationversion.AuthenticationProvider{
				{Name: " jwt-main ", Type: configurationversion.AuthenticationProviderJWT, Enabled: true, Priority: 1, JWT: &configurationversion.JWTSettings{
					SigningKeys:       []configurationversion.JWTSigningKey{{Name: " primary ", SecretRef: " secrets/jwt/key "}},
					AllowedAlgorithms: []configurationversion.JWTAlgorithm{configurationversion.RS256},
					AllowedIssuers:    []string{" issuer "}, AllowedAudiences: []string{" audience "},
					RequiredClaims: []configurationversion.JWTRequiredClaim{{Name: " tenant ", Value: " internal "}}, ClockSkewSeconds: 30,
				}},
				{Name: "api-key", Type: configurationversion.AuthenticationProviderAPIKey, Enabled: true, Priority: 2, APIKey: &configurationversion.APIKeySettings{Header: " X-API-Key ", SecretRef: " secrets/api-key "}},
				{Name: "basic", Type: configurationversion.AuthenticationProviderBasic, Priority: 3, Basic: &configurationversion.BasicSettings{Realm: " Operations ", SecretRef: " secrets/basic "}},
			},
		},
		Routing: &configurationversion.RoutingSettings{
			DefaultHandlerRef: " legacy ",
			Routes: []configurationversion.Route{{ID: " route ", Enabled: true, Priority: 1, HandlerRef: " legacy ", Matchers: []configurationversion.Matcher{
				{Type: configurationversion.MatcherTypePrincipalKind, Value: " authenticated "},
				{Type: configurationversion.MatcherTypeMessageType, Value: " text "},
			}}},
		},
	}
}

func snapshotRoutingMatcherValue(t *testing.T, snapshot Snapshot, route, matcher int) string {
	t.Helper()
	routing, present := snapshot.Routing()
	if !present {
		t.Fatal("Routing absent")
	}
	return routing.Routes()[route].Matchers()[matcher].Value()
}

func mustRouting(t *testing.T, snapshot Snapshot) RoutingSnapshot {
	t.Helper()
	routing, present := snapshot.Routing()
	if !present {
		t.Fatal("Routing absent")
	}
	return routing
}

func diagnosticCodes(values []Diagnostic) []string {
	result := make([]string, len(values))
	for index := range values {
		result[index] = values[index].Code()
	}
	return result
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func hasDiagnostic(values []Diagnostic, code string) bool {
	for _, value := range values {
		if value.Code() == code {
			return true
		}
	}
	return false
}

func assertDiagnostic(t *testing.T, diagnostics []Diagnostic, code, location string) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Code() == code && diagnostic.Location() == location {
			return
		}
	}
	t.Errorf("missing Diagnostic (%s, %s): %#v", code, location, diagnostics)
}

func assertNoDiagnosticAt(t *testing.T, diagnostics []Diagnostic, code, location string) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Code() == code && diagnostic.Location() == location {
			t.Errorf("unexpected Diagnostic (%s, %s)", code, location)
		}
	}
}
