package runtime

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

func TestDefaultBootstrapImplementsBootstrap(t *testing.T) {
	var _ Bootstrap = (*DefaultBootstrap)(nil)
}

func TestNewBootstrapRejectsNilResolver(t *testing.T) {
	bootstrap, err := NewBootstrap(nil, nil)
	if bootstrap != nil || !errors.Is(err, ErrNilSecretResolver) {
		t.Fatalf("NewBootstrap() = (%v, %v), want nil and ErrNilSecretResolver", bootstrap, err)
	}
}

func TestBootstrapBuildUsesRuntimeHost(t *testing.T) {
	resolver := apiKeyResolver(t)
	bootstrap, err := NewBootstrap(resolver, nil)
	if err != nil {
		t.Fatalf("NewBootstrap() error = %v", err)
	}

	built, err := bootstrap.Build(apiKeySnapshot(t))
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	host, ok := built.(*DefaultHost)
	if !ok {
		t.Fatalf("Build() Host type = %T, want *DefaultHost", built)
	}
	if host.state != hostBuilt || host.runtimeListener != nil {
		t.Fatalf("Build() Host = %#v, want Built Host without published Listener", host)
	}
	if built.Ready() {
		t.Fatal("Build() returned Ready Host")
	}
	if built.CanAccept() {
		t.Fatal("Build() returned Host accepting connections")
	}
}

func TestBootstrapHostStartPreservesAuthenticationBuildErrors(t *testing.T) {
	bootstrap, err := NewBootstrap(emptyResolver(t), nil)
	if err != nil {
		t.Fatalf("NewBootstrap() error = %v", err)
	}
	snapshot := apiKeySnapshot(t)
	snapshot.Authentication.Providers = []runtimeconfig.AuthenticationProviderSnapshot{
		{
			Name:    "basic",
			Type:    runtimeconfig.AuthenticationProviderBasic,
			Enabled: true,
			Basic: &runtimeconfig.BasicSnapshot{
				Realm:     "Universal WebSocket Platform",
				SecretRef: "secrets/basic/main",
			},
		},
	}

	built, err := bootstrap.Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	err = built.Start(context.Background())
	if !errors.Is(err, authentication.ErrFactoryNotFound) {
		t.Fatalf("Start() error = %v, want ErrFactoryNotFound", err)
	}
	host := built.(*DefaultHost)
	if got := currentHostState(host); got != hostBuilt {
		t.Fatalf("Host state = %v, want hostBuilt", got)
	}
	if host.RuntimeContext() != nil || host.runtimeListener != nil {
		t.Fatal("failed dependency acquisition published Runtime resources")
	}
	if built.Ready() {
		t.Fatal("dependency acquisition error left Host Ready")
	}
	if built.CanAccept() {
		t.Fatal("dependency acquisition error left admission open")
	}
}

func TestBootstrapHostPreservesRuntimeVertical(t *testing.T) {
	bootstrap, err := NewBootstrap(apiKeyResolver(t), message.NewEchoHandler())
	if err != nil {
		t.Fatalf("NewBootstrap() error = %v", err)
	}
	snapshot := apiKeySnapshot(t)
	built, err := bootstrap.Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if err := built.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !built.Ready() {
		t.Fatal("Host Ready() = false after successful Start")
	}
	if !built.CanAccept() {
		t.Fatal("Host CanAccept() = false after successful Start")
	}
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
		"ws://127.0.0.1:"+portString(snapshot.Listener.Port)+"/ws",
		&websocket.DialOptions{HTTPHeader: header},
	)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer connection.CloseNow()
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("Dial() status = %d, want 101", response.StatusCode)
	}

	want := []byte("runtime host")
	if err := connection.Write(ctx, websocket.MessageText, want); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	messageType, got, err := connection.Read(ctx)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if messageType != websocket.MessageText || string(got) != string(want) {
		t.Fatalf("Read() = (%v, %q), want text %q", messageType, got, want)
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
}

func apiKeySnapshot(t *testing.T) runtimeconfig.Snapshot {
	t.Helper()
	return runtimeconfig.Snapshot{
		ConfigurationID: 1,
		VersionID:       1,
		Listener: runtimeconfig.ListenerSnapshot{
			Host: "127.0.0.1",
			Port: availablePort(t),
			TLS:  runtimeconfig.TLSSnapshot{MinVersion: "1.2"},
		},
		Authentication: runtimeconfig.AuthenticationSnapshot{
			Enabled: true,
			Providers: []runtimeconfig.AuthenticationProviderSnapshot{
				{
					Name:     "api-key",
					Type:     runtimeconfig.AuthenticationProviderAPIKey,
					Enabled:  true,
					Priority: 10,
					APIKey: &runtimeconfig.APIKeySnapshot{
						Header:    "X-API-Key",
						SecretRef: "secrets/api-key/runtime",
					},
				},
			},
		},
	}
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
