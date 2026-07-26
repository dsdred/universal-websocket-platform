package runtimeconfig

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

func TestEveryApprovedDiagnosticCodeIsReachable(t *testing.T) {
	approved := loadApprovedDiagnosticRegistry(t)
	inputs := []runtimeconfigload.DetachedLoadResult{{}}

	valid := validConfigurationVersion()
	inputs = append(inputs,
		detachedWithFacts(valid, 11, 22, 33, 4, true, "future", 2, "runtime", "attempt"),
	)
	mismatch := valid
	mismatch.ConfigurationID, mismatch.ID, mismatch.Number = 99, 98, 97
	inputs = append(inputs, detachedWithFacts(mismatch, 11, 22, 33, 4, true, "uwp.configuration", 1, "runtime", "attempt"))

	listenerRequired := valid
	listenerRequired.Listener = configurationversion.ListenerSettings{
		TLS:      configurationversion.TLSSettings{Enabled: true},
		Timeouts: configurationversion.TimeoutSettings{ReadSeconds: 86401, IdleSeconds: 86401},
	}
	inputs = append(inputs, detachedResult(listenerRequired))
	listenerInvalid := valid
	listenerInvalid.Listener.Host = strings.Repeat("x", 256) + "!"
	listenerInvalid.Listener.TLS = configurationversion.TLSSettings{
		CertificateRef: strings.Repeat("x", 256) + "/",
		PrivateKeyRef:  strings.Repeat("x", 256) + "/",
		MinVersion:     "1.1",
	}
	listenerInvalid.Listener.Timeouts.HandshakeSeconds = 301
	listenerInvalid.Listener.Timeouts.WriteSeconds = 301
	inputs = append(inputs, detachedResult(listenerInvalid))

	emptyAuthentication := valid
	emptyAuthentication.Authentication = configurationversion.AuthenticationSettings{Enabled: true}
	inputs = append(inputs, detachedResult(emptyAuthentication))
	disabledAuthentication := valid
	disabledAuthentication.Authentication = configurationversion.AuthenticationSettings{
		Enabled: true,
		Providers: []configurationversion.AuthenticationProvider{{
			Name: "disabled", Type: configurationversion.AuthenticationProviderAPIKey,
			Priority: 1, APIKey: &configurationversion.APIKeySettings{Header: "X-Key", SecretRef: "secrets/key"},
		}},
	}
	inputs = append(inputs, detachedResult(disabledAuthentication))

	authentication := valid
	tooLong := strings.Repeat("x", 256) + "/"
	authentication.Authentication = configurationversion.AuthenticationSettings{
		Providers: []configurationversion.AuthenticationProvider{
			{Name: "", Type: "", Priority: 1},
			{Name: strings.Repeat("n", 256), Type: "future", Priority: 1},
			{Name: "duplicate", Type: configurationversion.AuthenticationProviderAPIKey, Priority: 2, Basic: &configurationversion.BasicSettings{}, JWT: &configurationversion.JWTSettings{}},
			{Name: "duplicate", Type: configurationversion.AuthenticationProviderAPIKey, Priority: 3, APIKey: &configurationversion.APIKeySettings{}},
			{Name: "api", Type: configurationversion.AuthenticationProviderAPIKey, Priority: 4, APIKey: &configurationversion.APIKeySettings{Header: strings.Repeat("h", 256) + ":", SecretRef: tooLong}},
			{Name: "basic-required", Type: configurationversion.AuthenticationProviderBasic, Priority: 5, Basic: &configurationversion.BasicSettings{}},
			{Name: "basic-invalid", Type: configurationversion.AuthenticationProviderBasic, Priority: 6, Basic: &configurationversion.BasicSettings{Realm: strings.Repeat("r", 256), SecretRef: tooLong}},
			{Name: "jwt-required", Type: configurationversion.AuthenticationProviderJWT, Priority: 7, JWT: &configurationversion.JWTSettings{}},
			{Name: "jwt-invalid", Type: configurationversion.AuthenticationProviderJWT, Priority: 8, JWT: &configurationversion.JWTSettings{
				SigningKeys: []configurationversion.JWTSigningKey{
					{}, {Name: "key", SecretRef: tooLong}, {Name: "key", SecretRef: "secrets/key"},
				},
				AllowedAlgorithms: []configurationversion.JWTAlgorithm{"future", configurationversion.RS256, configurationversion.RS256},
				AllowedIssuers:    []string{" ", "issuer", "issuer"},
				AllowedAudiences:  []string{" ", "audience", "audience"},
				RequiredClaims: []configurationversion.JWTRequiredClaim{
					{}, {Name: "claim", Value: "value"}, {Name: "claim", Value: " "},
				},
				ClockSkewSeconds: 301,
			}},
		},
	}
	inputs = append(inputs, detachedResult(authentication))

	routing := valid
	routing.Routing = &configurationversion.RoutingSettings{
		DefaultHandlerRef: " ",
		Routes: []configurationversion.Route{
			{Enabled: true},
			{ID: strings.Repeat("r", 129) + "!", Priority: 1, HandlerRef: strings.Repeat("h", 129) + "!", Matchers: []configurationversion.Matcher{
				{}, {Type: "future", Value: "value"}, {Type: configurationversion.MatcherTypeMessageType, Value: "future"},
				{Type: configurationversion.MatcherTypeMessageType, Value: "text"}, {Type: configurationversion.MatcherTypePrincipalKind, Value: "anonymous"},
			}},
			{ID: "route", Enabled: true, Priority: 2, HandlerRef: "legacy", Matchers: []configurationversion.Matcher{{Type: configurationversion.MatcherTypeMessageType, Value: "text"}}},
			{ID: "route", Enabled: true, Priority: 2, HandlerRef: "legacy", Matchers: []configurationversion.Matcher{{Type: configurationversion.MatcherTypeMessageType, Value: "text"}}},
		},
	}
	inputs = append(inputs, detachedResult(routing))
	defaultInvalid := valid
	defaultInvalid.Routing.DefaultHandlerRef = strings.Repeat("h", 129) + "!"
	inputs = append(inputs, detachedResult(defaultInvalid))
	tooMany := valid
	tooMany.Routing.Routes = make([]configurationversion.Route, 257)
	inputs = append(inputs, detachedResult(tooMany))

	reached := make(map[string]struct{})
	for _, input := range inputs {
		_, diagnostics := NewBuilder().Build(input)
		for _, diagnostic := range diagnostics {
			if diagnostic.Severity() != SeverityError || diagnostic.Message() == "" || diagnostic.Location() == "" {
				t.Fatalf("invalid Diagnostic tuple: %#v", diagnostic)
			}
			assertApprovedDiagnosticTuple(t, approved, diagnostic)
			reached[diagnostic.Code()] = struct{}{}
		}
	}
	want := approvedDiagnosticCodes()
	var missing []string
	for _, code := range want {
		if _, ok := reached[code]; !ok {
			missing = append(missing, code)
		}
	}
	var unexpected []string
	for code := range reached {
		if !containsString(want, code) {
			unexpected = append(unexpected, code)
		}
	}
	sort.Strings(unexpected)
	if len(want) != 93 || len(approved) != 93 || len(missing) != 0 || len(unexpected) != 0 {
		t.Fatalf("Diagnostic registry reachability: want=%d reached=%d missing=%v unexpected=%v", len(want), len(reached), missing, unexpected)
	}
}

type approvedDiagnostic struct {
	location string
	message  string
}

var diagnosticIndexPattern = regexp.MustCompile(`\[\d+\]`)

func loadApprovedDiagnosticRegistry(t *testing.T) map[string]approvedDiagnostic {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate registry test source")
	}
	documentPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "docs", "en", "design", "DP-008-snapshot-builder-contract.md")
	content, err := os.ReadFile(documentPath)
	if err != nil {
		t.Fatalf("read approved DP-008 Diagnostic registry: %v", err)
	}
	result := make(map[string]approvedDiagnostic)
	for _, line := range strings.Split(string(content), "\n") {
		columns := strings.Split(line, "|")
		if len(columns) < 5 {
			continue
		}
		code := strings.Trim(strings.TrimSpace(columns[1]), "`")
		if !strings.HasPrefix(code, "snapshot.") {
			continue
		}
		location := strings.Trim(strings.TrimSpace(columns[2]), "`")
		message := strings.Trim(strings.TrimSpace(columns[3]), "`")
		if _, duplicate := result[code]; duplicate {
			t.Fatalf("approved Diagnostic code %q occurs more than once", code)
		}
		result[code] = approvedDiagnostic{location: location, message: message}
	}
	if len(result) != 93 {
		t.Fatalf("approved DP-008 Diagnostic registry contains %d entries, want 93", len(result))
	}
	return result
}

func assertApprovedDiagnosticTuple(t *testing.T, approved map[string]approvedDiagnostic, actual Diagnostic) {
	t.Helper()
	expected, ok := approved[actual.Code()]
	if !ok {
		t.Errorf("Diagnostic code %q is absent from approved DP-008 registry", actual.Code())
		return
	}
	if actual.Message() != expected.message {
		t.Errorf("Diagnostic %q message = %q, want %q", actual.Code(), actual.Message(), expected.message)
	}
	if actual.Code() == "snapshot.authentication.provider.settings.required" ||
		actual.Code() == "snapshot.authentication.provider.settings.forbidden" {
		matched, err := regexp.MatchString(
			`^\$\.authentication\.providers\[\d+\]\.(apiKey|basic|jwt)$`,
			actual.Location(),
		)
		if err != nil || !matched {
			t.Errorf("Diagnostic %q location = %q, want approved settings branch", actual.Code(), actual.Location())
		}
		return
	}
	index := 0
	normalized := diagnosticIndexPattern.ReplaceAllStringFunc(actual.Location(), func(string) string {
		index++
		if index == 1 {
			return "[i]"
		}
		return "[j]"
	})
	if normalized != expected.location {
		t.Errorf("Diagnostic %q location = %q (pattern %q), want %q", actual.Code(), actual.Location(), normalized, expected.location)
	}
}

func detachedWithFacts(
	version configurationversion.ConfigurationVersion,
	workspace, configuration, versionID uint64,
	number uint32,
	published bool,
	schema string,
	schemaVersion uint32,
	runtimeID runtimeconfigload.RuntimeInstanceID,
	attemptID runtimeconfigload.LaunchAttemptID,
) runtimeconfigload.DetachedLoadResult {
	request := runtimeconfigload.NewLoadRequest(workspace, configuration, versionID, runtimeID, attemptID)
	return runtimeconfigload.NewDetachedLoadResult(request, version, number, published, schema, schemaVersion)
}

func approvedDiagnosticCodes() []string {
	const values = `
snapshot.provenance.workspace.required snapshot.provenance.configuration.required snapshot.provenance.configuration_version.required
snapshot.provenance.configuration_number.required snapshot.provenance.runtime_instance.required snapshot.provenance.launch_attempt.required
snapshot.schema.identity.required snapshot.schema.identity.unsupported snapshot.schema.version.required snapshot.schema.version.unsupported
snapshot.handoff.not_published snapshot.configuration_version.state.not_published snapshot.configuration_version.configuration.required
snapshot.configuration_version.identity.required snapshot.configuration_version.number.required snapshot.configuration_version.configuration.inconsistent
snapshot.configuration_version.identity.inconsistent snapshot.configuration_version.number.inconsistent snapshot.listener.host.required
snapshot.listener.host.too_long snapshot.listener.host.invalid snapshot.listener.port.out_of_range snapshot.listener.tls.min_version.required
snapshot.listener.tls.min_version.unsupported snapshot.listener.tls.certificate.required snapshot.listener.tls.certificate.too_long
snapshot.listener.tls.certificate.invalid snapshot.listener.tls.private_key.required snapshot.listener.tls.private_key.too_long
snapshot.listener.tls.private_key.invalid snapshot.listener.timeout.handshake.out_of_range snapshot.listener.timeout.read.out_of_range
snapshot.listener.timeout.write.out_of_range snapshot.listener.timeout.idle.out_of_range snapshot.authentication.providers.required
snapshot.authentication.enabled_provider.required snapshot.authentication.provider.name.required snapshot.authentication.provider.name.too_long
snapshot.authentication.provider.name.duplicate snapshot.authentication.provider.priority.duplicate snapshot.authentication.provider.type.required
snapshot.authentication.provider.type.unsupported snapshot.authentication.provider.settings.required snapshot.authentication.provider.settings.forbidden
snapshot.authentication.api_key.header.required snapshot.authentication.api_key.header.too_long snapshot.authentication.api_key.header.invalid
snapshot.authentication.api_key.secret_ref.required snapshot.authentication.api_key.secret_ref.too_long snapshot.authentication.api_key.secret_ref.invalid
snapshot.authentication.basic.realm.required snapshot.authentication.basic.realm.too_long snapshot.authentication.basic.secret_ref.required
snapshot.authentication.basic.secret_ref.too_long snapshot.authentication.basic.secret_ref.invalid snapshot.authentication.jwt.signing_keys.required
snapshot.authentication.jwt.signing_key.name.required snapshot.authentication.jwt.signing_key.name.duplicate
snapshot.authentication.jwt.signing_key.secret_ref.required snapshot.authentication.jwt.signing_key.secret_ref.too_long
snapshot.authentication.jwt.signing_key.secret_ref.invalid snapshot.authentication.jwt.algorithms.required
snapshot.authentication.jwt.algorithm.unsupported snapshot.authentication.jwt.algorithm.duplicate snapshot.authentication.jwt.issuer.required
snapshot.authentication.jwt.issuer.duplicate snapshot.authentication.jwt.audience.required snapshot.authentication.jwt.audience.duplicate
snapshot.authentication.jwt.claim.name.required snapshot.authentication.jwt.claim.name.duplicate snapshot.authentication.jwt.claim.value.required
snapshot.authentication.jwt.clock_skew.out_of_range snapshot.routing.routes.too_many snapshot.routing.default_handler.required
snapshot.routing.default_handler.too_long snapshot.routing.default_handler.invalid snapshot.routing.route.id.required
snapshot.routing.route.id.too_long snapshot.routing.route.id.invalid snapshot.routing.route.id.duplicate
snapshot.routing.route.priority.out_of_range snapshot.routing.route.priority.duplicate snapshot.routing.route.handler.required
snapshot.routing.route.handler.too_long snapshot.routing.route.handler.invalid snapshot.routing.route.matchers.too_many
snapshot.routing.route.matchers.required snapshot.routing.matcher.type.required snapshot.routing.matcher.type.unsupported
snapshot.routing.matcher.type.duplicate snapshot.routing.matcher.value.required snapshot.routing.matcher.value.unsupported
snapshot.routing.matcher_set.duplicate`
	return strings.Fields(values)
}
