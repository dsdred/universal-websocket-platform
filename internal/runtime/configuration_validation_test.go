package runtime

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

func TestRuntimeSnapshotFieldSupportMatrixIsExhaustive(t *testing.T) {
	decisions := map[string]string{
		"provenance.WorkspaceID":                                 "ARCH-005 provenance: declarative owner",
		"provenance.ConfigurationID":                             "Runtime identity: validated before composition",
		"provenance.ConfigurationVersionID":                      "Runtime identity: validated before composition",
		"provenance.ConfigurationVersionNumber":                  "ARCH-005 provenance: exact declarative version",
		"provenance.SchemaIdentity":                              "Builder-supported Configuration schema",
		"provenance.SchemaVersion":                               "Builder-supported Configuration schema version",
		"provenance.RuntimeInstanceID":                           "ARCH-004 operational identity",
		"provenance.LaunchAttemptID":                             "ARCH-004 execution identity",
		"listener.Host":                                          "Listener: executed by net.Listen",
		"listener.Port":                                          "Listener: executed by net.Listen",
		"listener.TLS.Enabled":                                   "Runtime composition: enabled TLS rejected before Listener",
		"listener.TLS.CertificateRef":                            "TLS: inactive while TLS is disabled",
		"listener.TLS.PrivateKeyRef":                             "TLS: inactive while TLS is disabled",
		"listener.TLS.MinVersion":                                "TLS: inactive while TLS is disabled",
		"listener.Timeouts.HandshakeSeconds":                     "Handshake: executed as pre-Upgrade deadline",
		"listener.Timeouts.ReadSeconds":                          "Listener settings: configured but inactive until Gate 9",
		"listener.Timeouts.WriteSeconds":                         "Listener settings: configured but inactive until Gate 9",
		"listener.Timeouts.IdleSeconds":                          "Listener settings: configured but inactive until Gate 9",
		"authentication.Enabled":                                 "Authentication Bootstrap: service or anonymous identity",
		"authentication.Providers[].Name":                        "Authentication Factory: provider identity",
		"authentication.Providers[].Type":                        "Authentication Registry: executed or explicitly rejected",
		"authentication.Providers[].Enabled":                     "Authentication Bootstrap: active provider selection",
		"authentication.Providers[].Priority":                    "Authentication Bootstrap: evaluation order",
		"authentication.Providers[].APIKey.Header":               "API Key Provider: request header",
		"authentication.Providers[].APIKey.SecretRef":            "API Key Provider: request-time secret resolution",
		"authentication.Providers[].JWT.SigningKeys[].Name":      "JWT Provider: signing key identity",
		"authentication.Providers[].JWT.SigningKeys[].SecretRef": "JWT Provider: request-time secret resolution",
		"authentication.Providers[].JWT.AllowedAlgorithms[]":     "JWT Provider: algorithm policy",
		"authentication.Providers[].JWT.AllowedIssuers[]":        "JWT Provider: issuer policy",
		"authentication.Providers[].JWT.AllowedAudiences[]":      "JWT Provider: audience policy",
		"authentication.Providers[].JWT.RequiredClaims[].Name":   "JWT Provider: required claim name",
		"authentication.Providers[].JWT.RequiredClaims[].Value":  "JWT Provider: required claim value",
		"authentication.Providers[].JWT.ClockSkewSeconds":        "JWT Provider: clock skew policy",
		"authentication.Providers[].Basic.Realm":                 "Authentication Registry: Basic rejected in this build",
		"authentication.Providers[].Basic.SecretRef":             "Authentication Registry: Basic rejected in this build",
		"routing.routes[].id":                                    "DP-005 Router compiler: immutable Route identity",
		"routing.routes[].enabled":                               "DP-005 Router compiler: disabled Route exclusion",
		"routing.routes[].priority":                              "DP-005 Router compiler: ascending selection order",
		"routing.routes[].matchers[].matcherType":                "DP-005 Router compiler: supported Matcher selection",
		"routing.routes[].matchers[].value":                      "DP-005 Router compiler: exact Matcher value",
		"routing.routes[].handlerRef":                            "Runtime composition: active Handler resolution",
		"routing.defaultHandlerRef":                              "Runtime composition: optional Default Handler resolution",
		"routing.defaultHandlerRefPresent":                       "DP-008: Default Handler presence is observable",
	}

	fields := snapshotLeafFields(reflect.TypeOf(runtimeconfig.Snapshot{}), "", nil)
	slices.Sort(fields)
	if !reflect.DeepEqual(fields, mapKeys(decisions)) {
		t.Fatalf("Snapshot fields = %v, support decisions = %v", fields, mapKeys(decisions))
	}
}

func TestValidateExecutableSnapshotRejectsZeroSnapshot(t *testing.T) {
	if err := validateExecutableSnapshot(runtimeconfig.Snapshot{}); !errors.Is(err, ErrInvalidRuntimeConfiguration) {
		t.Fatalf("validateExecutableSnapshot(zero) error = %v", err)
	}
}

func TestValidateExecutableSnapshotAcceptsBuilderSnapshot(t *testing.T) {
	if err := validateExecutableSnapshot(validSnapshot()); err != nil {
		t.Fatalf("validateExecutableSnapshot() error = %v", err)
	}
}

func TestValidateExecutableSnapshotRejectsUnsupportedTLSWithoutLeakingReferences(t *testing.T) {
	snapshot, diagnostics := buildSnapshotForTest(func(version *configurationversion.ConfigurationVersion) {
		version.Listener.TLS = configurationversion.TLSSettings{
			Enabled:        true,
			CertificateRef: "safe-cert-ref",
			PrivateKeyRef:  "credential-that-must-not-leak",
			MinVersion:     "1.2",
		}
	})
	if len(diagnostics) != 0 {
		t.Fatalf("Builder diagnostics = %#v", diagnostics)
	}
	err := validateExecutableSnapshot(snapshot)
	if !errors.Is(err, ErrInvalidRuntimeConfiguration) || !errors.Is(err, ErrUnsupportedRuntimeCapability) {
		t.Fatalf("validateExecutableSnapshot() error = %v", err)
	}
	if strings.Contains(err.Error(), "credential-that-must-not-leak") {
		t.Fatalf("error leaked secret reference: %v", err)
	}
}

func TestHandshakeTimeoutHandlerUsesConfiguredDuration(t *testing.T) {
	for _, seconds := range []uint32{1, 2} {
		t.Run((time.Duration(seconds) * time.Second).String(), func(t *testing.T) {
			snapshot, diagnostics := buildSnapshotForTest(func(version *configurationversion.ConfigurationVersion) {
				version.Listener.Timeouts.HandshakeSeconds = seconds
			})
			if len(diagnostics) != 0 {
				t.Fatalf("Builder diagnostics = %#v", diagnostics)
			}
			configured := time.Duration(snapshot.Listener().Timeouts.HandshakeSeconds) * time.Second
			var remaining time.Duration
			next := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				deadline, ok := request.Context().Deadline()
				if !ok {
					t.Fatal("Handshake request context has no deadline")
				}
				remaining = time.Until(deadline)
				response.WriteHeader(http.StatusNoContent)
			})
			handler := handshakeTimeoutHandler{next: next, timeout: configured}
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ws", nil))
			if remaining <= configured-250*time.Millisecond || remaining > configured {
				t.Fatalf("deadline remaining = %v, want configured duration near %v", remaining, configured)
			}
		})
	}
}

func TestTLSValidationFailsBeforeSocketAndPreservesSnapshot(t *testing.T) {
	snapshot, diagnostics := buildSnapshotForTest(func(version *configurationversion.ConfigurationVersion) {
		version.Listener.Port = availablePort(t)
		version.Listener.TLS = configurationversion.TLSSettings{
			Enabled:        true,
			CertificateRef: "certificates/runtime",
			PrivateKeyRef:  "credential-that-must-not-leak",
			MinVersion:     "1.3",
		}
	})
	if len(diagnostics) != 0 {
		t.Fatalf("Builder diagnostics = %#v", diagnostics)
	}
	wantSnapshot := snapshot
	bootstrap, err := NewBootstrapWithTerminalErrorReporter(emptyResolver(t), nil, nil)
	if err != nil {
		t.Fatalf("NewBootstrap() error = %v", err)
	}
	built, err := bootstrap.Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	err = built.Start(context.Background())
	if !errors.Is(err, ErrUnsupportedRuntimeCapability) || !errors.Is(err, ErrInvalidRuntimeConfiguration) {
		t.Fatalf("Start() error = %v, want unsupported invalid Runtime configuration", err)
	}
	if strings.Contains(err.Error(), "credential-that-must-not-leak") {
		t.Fatal("Start() error exposed a Secret Reference")
	}
	if built.Running() || built.Ready() || built.CanAccept() || built.RuntimeContext() != nil {
		t.Fatal("failed Host published active Runtime state")
	}
	if built.(*DefaultHost).runtimeListener != nil {
		t.Fatal("failed Host published a Listener")
	}
	if !reflect.DeepEqual(built.Snapshot(), wantSnapshot) {
		t.Fatal("validation mutated Snapshot")
	}
	assertPortAvailable(t, snapshot.Listener().Port)
}

func TestDefaultPublishedConfigurationStarts(t *testing.T) {
	repository := configurationversion.NewMemoryConfigurationVersionRepository()
	service := configurationversion.NewService(repository, configurationExists{}, time.Now)
	version, err := service.Create(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	version, err = service.UpdateListener(context.Background(), 1, 1, version.ID, configurationversion.ListenerSettings{
		Host: "127.0.0.1",
		Port: availablePort(t),
	})
	if err != nil {
		t.Fatalf("UpdateListener() error = %v", err)
	}
	version, err = service.Publish(context.Background(), 1, 1, version.ID)
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	request := runtimeconfigload.NewLoadRequest(1, 1, version.ID, "runtime-default", "attempt-default")
	input := runtimeconfigload.NewDetachedLoadResult(
		request, version, version.Number, true, "uwp.configuration", 1,
	)
	snapshot, diagnostics := runtimeconfig.NewBuilder().Build(input)
	if len(diagnostics) != 0 {
		t.Fatalf("runtimeconfig.Build() diagnostics = %#v", diagnostics)
	}
	bootstrap, err := NewBootstrap(emptyResolver(t), nil)
	if err != nil {
		t.Fatalf("NewBootstrap() error = %v", err)
	}
	host, err := bootstrap.Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !host.Running() || !host.Ready() || !host.CanAccept() || host.RuntimeContext() == nil {
		t.Fatal("default Published Configuration did not become Ready")
	}
	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

type configurationExists struct{}

func (configurationExists) Exists(context.Context, uint64, uint64) (bool, error) { return true, nil }

func snapshotLeafFields(typ reflect.Type, prefix string, seen map[reflect.Type]bool) []string {
	if seen == nil {
		seen = make(map[reflect.Type]bool)
	}
	for typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice {
		if typ.Kind() == reflect.Slice {
			prefix += "[]"
		}
		typ = typ.Elem()
	}
	if typ.PkgPath() != reflect.TypeOf(runtimeconfig.Snapshot{}).PkgPath() || typ.Kind() != reflect.Struct {
		return []string{strings.TrimPrefix(prefix, ".")}
	}
	if seen[typ] {
		return nil
	}
	seen[typ] = true
	defer delete(seen, typ)
	var result []string
	for index := 0; index < typ.NumField(); index++ {
		field := typ.Field(index)
		result = append(result, snapshotLeafFields(field.Type, prefix+"."+field.Name, seen)...)
	}
	return result
}

func mapKeys(values map[string]string) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	slices.Sort(result)
	return result
}
