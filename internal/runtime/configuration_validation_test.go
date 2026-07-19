package runtime

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

func TestRuntimeSnapshotFieldSupportMatrixIsExhaustive(t *testing.T) {
	decisions := map[string]string{
		"ConfigurationID":                                        "Runtime identity: validated before composition",
		"VersionID":                                              "Runtime identity: validated before composition",
		"Listener.Host":                                          "Listener: executed by net.Listen",
		"Listener.Port":                                          "Listener: executed by net.Listen",
		"Listener.TLS.Enabled":                                   "Runtime composition: enabled TLS rejected before Listener",
		"Listener.TLS.CertificateRef":                            "TLS: inactive while TLS is disabled",
		"Listener.TLS.PrivateKeyRef":                             "TLS: inactive while TLS is disabled",
		"Listener.TLS.MinVersion":                                "TLS: inactive while TLS is disabled",
		"Listener.Timeouts.HandshakeSeconds":                     "Handshake: executed as pre-Upgrade deadline",
		"Listener.Timeouts.ReadSeconds":                          "Listener settings: configured but inactive until Gate 9",
		"Listener.Timeouts.WriteSeconds":                         "Listener settings: configured but inactive until Gate 9",
		"Listener.Timeouts.IdleSeconds":                          "Listener settings: configured but inactive until Gate 9",
		"Authentication.Enabled":                                 "Authentication Bootstrap: service or anonymous identity",
		"Authentication.Providers[].Name":                        "Authentication Factory: provider identity",
		"Authentication.Providers[].Type":                        "Authentication Registry: executed or explicitly rejected",
		"Authentication.Providers[].Enabled":                     "Authentication Bootstrap: active provider selection",
		"Authentication.Providers[].Priority":                    "Authentication Bootstrap: evaluation order",
		"Authentication.Providers[].APIKey.Header":               "API Key Provider: request header",
		"Authentication.Providers[].APIKey.SecretRef":            "API Key Provider: request-time secret resolution",
		"Authentication.Providers[].JWT.SigningKeys[].Name":      "JWT Provider: signing key identity",
		"Authentication.Providers[].JWT.SigningKeys[].SecretRef": "JWT Provider: request-time secret resolution",
		"Authentication.Providers[].JWT.AllowedAlgorithms[]":     "JWT Provider: executed HMAC or explicitly rejected",
		"Authentication.Providers[].JWT.AllowedIssuers[]":        "JWT Provider: issuer policy",
		"Authentication.Providers[].JWT.AllowedAudiences[]":      "JWT Provider: audience policy",
		"Authentication.Providers[].JWT.RequiredClaims[].Name":   "JWT Provider: required claim name",
		"Authentication.Providers[].JWT.RequiredClaims[].Value":  "JWT Provider: required claim value",
		"Authentication.Providers[].JWT.ClockSkewSeconds":        "JWT Provider: clock skew policy",
		"Authentication.Providers[].Basic.Realm":                 "Authentication Registry: Basic rejected in this build",
		"Authentication.Providers[].Basic.SecretRef":             "Authentication Registry: Basic rejected in this build",
		"Routing.routes[].id":                                    "Routing: configured but inactive until Router implementation",
		"Routing.routes[].enabled":                               "Routing: configured but inactive until Router implementation",
		"Routing.routes[].priority":                              "Routing: configured but inactive until Router implementation",
		"Routing.routes[].matchers[].matcherType":                "Routing: configured but inactive until Router implementation",
		"Routing.routes[].matchers[].value":                      "Routing: configured but inactive until Router implementation",
		"Routing.routes[].handlerRef":                            "Routing: configured but inactive until Router implementation",
		"Routing.defaultHandlerRef":                              "Routing: configured but inactive until Router implementation",
	}

	fields := snapshotLeafFields(reflect.TypeOf(runtimeconfig.Snapshot{}), "", nil)
	slices.Sort(fields)
	if !reflect.DeepEqual(fields, mapKeys(decisions)) {
		t.Fatalf("Snapshot fields = %v, support decisions = %v", fields, mapKeys(decisions))
	}
}

func TestValidateExecutableSnapshot(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*runtimeconfig.Snapshot)
		want   error
	}{
		{name: "supported defaults"},
		{name: "zero configuration identity", mutate: func(snapshot *runtimeconfig.Snapshot) { snapshot.ConfigurationID = 0 }, want: ErrInvalidRuntimeConfiguration},
		{name: "zero version identity", mutate: func(snapshot *runtimeconfig.Snapshot) { snapshot.VersionID = 0 }, want: ErrInvalidRuntimeConfiguration},
		{name: "empty listener host", mutate: func(snapshot *runtimeconfig.Snapshot) { snapshot.Listener.Host = " " }, want: ErrInvalidRuntimeConfiguration},
		{name: "zero listener port", mutate: func(snapshot *runtimeconfig.Snapshot) { snapshot.Listener.Port = 0 }, want: ErrInvalidRuntimeConfiguration},
		{name: "invalid TLS minimum version", mutate: func(snapshot *runtimeconfig.Snapshot) { snapshot.Listener.TLS.MinVersion = "1.1" }, want: ErrInvalidRuntimeConfiguration},
		{name: "handshake timeout disabled", mutate: func(snapshot *runtimeconfig.Snapshot) { snapshot.Listener.Timeouts.HandshakeSeconds = 0 }, want: ErrInvalidRuntimeConfiguration},
		{name: "handshake timeout above control plane range", mutate: func(snapshot *runtimeconfig.Snapshot) { snapshot.Listener.Timeouts.HandshakeSeconds = 301 }, want: ErrInvalidRuntimeConfiguration},
		{name: "TLS enabled", mutate: func(snapshot *runtimeconfig.Snapshot) {
			snapshot.Listener.TLS.Enabled = true
			snapshot.Listener.TLS.CertificateRef = "safe-cert-ref"
			snapshot.Listener.TLS.PrivateKeyRef = "credential-that-must-not-leak"
		}, want: ErrUnsupportedRuntimeCapability},
		{name: "configured runtime timeouts remain inactive", mutate: func(snapshot *runtimeconfig.Snapshot) {
			snapshot.Listener.Timeouts.ReadSeconds = 123
			snapshot.Listener.Timeouts.WriteSeconds = 234
			snapshot.Listener.Timeouts.IdleSeconds = 345
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := executableSnapshot(t)
			if test.mutate != nil {
				test.mutate(&snapshot)
			}
			err := validateExecutableSnapshot(snapshot)
			if !errors.Is(err, test.want) {
				t.Fatalf("validateExecutableSnapshot() error = %v, want %v", err, test.want)
			}
			if err != nil && strings.Contains(err.Error(), "credential-that-must-not-leak") {
				t.Fatal("validation error exposed a Secret Reference")
			}
		})
	}
}

func TestHandshakeTimeoutHandlerUsesConfiguredDuration(t *testing.T) {
	for _, test := range []struct {
		name    string
		timeout time.Duration
		minimum time.Duration
		maximum time.Duration
	}{
		{name: "short", timeout: 40 * time.Millisecond, minimum: 20 * time.Millisecond, maximum: 100 * time.Millisecond},
		{name: "long", timeout: 400 * time.Millisecond, minimum: 300 * time.Millisecond, maximum: time.Second},
	} {
		t.Run(test.name, func(t *testing.T) {
			var remaining time.Duration
			next := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				deadline, ok := request.Context().Deadline()
				if !ok {
					t.Fatal("Handshake request context has no deadline")
				}
				remaining = time.Until(deadline)
				response.WriteHeader(http.StatusNoContent)
			})
			handler := handshakeTimeoutHandler{next: next, timeout: test.timeout}
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ws", nil))
			if remaining < test.minimum || remaining > test.maximum {
				t.Fatalf("deadline remaining = %v, want [%v, %v] for configured %v", remaining, test.minimum, test.maximum, test.timeout)
			}
		})
	}
}

func TestTLSValidationFailsBeforeSocketAndPreservesSnapshot(t *testing.T) {
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
	version, err = service.UpdateTLS(context.Background(), 1, 1, version.ID, configurationversion.TLSSettings{
		Enabled:        true,
		CertificateRef: "certificates/runtime",
		PrivateKeyRef:  "credential-that-must-not-leak",
		MinVersion:     "1.3",
	})
	if err != nil {
		t.Fatalf("UpdateTLS() error = %v", err)
	}
	version, err = service.Publish(context.Background(), 1, 1, version.ID)
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	snapshot, err := runtimeconfig.NewBuilder().Build(version)
	if err != nil {
		t.Fatalf("runtimeconfig.Build() error = %v", err)
	}
	wantSnapshot := cloneSnapshot(snapshot)
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
	assertPortAvailable(t, snapshot.Listener.Port)
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
	snapshot, err := runtimeconfig.NewBuilder().Build(version)
	if err != nil {
		t.Fatalf("runtimeconfig.Build() error = %v", err)
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

func executableSnapshot(t *testing.T) runtimeconfig.Snapshot {
	t.Helper()
	return runtimeconfig.Snapshot{
		ConfigurationID: 1,
		VersionID:       1,
		Listener: runtimeconfig.ListenerSnapshot{
			Host: "127.0.0.1",
			Port: availablePort(t),
			TLS:  runtimeconfig.TLSSnapshot{MinVersion: "1.2"},
			Timeouts: runtimeconfig.TimeoutSnapshot{
				HandshakeSeconds: 10,
				WriteSeconds:     10,
				IdleSeconds:      60,
			},
		},
		Authentication: runtimeconfig.AuthenticationSnapshot{Providers: []runtimeconfig.AuthenticationProviderSnapshot{}},
	}
}

type configurationExists struct{}

func (configurationExists) Exists(context.Context, uint64, uint64) (bool, error) { return true, nil }

func assertPortAvailable(t *testing.T, port uint16) {
	t.Helper()
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", portString(port)))
	if err != nil {
		t.Fatalf("port %d remains occupied: %v", port, err)
	}
	_ = listener.Close()
}

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
