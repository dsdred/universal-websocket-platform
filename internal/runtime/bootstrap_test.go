package runtime

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
	"github.com/golang-jwt/jwt/v5"
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
	reportedErrors := make(chan error, 4)
	bootstrap, err := NewBootstrapWithTerminalErrorReporter(
		apiKeyResolver(t),
		message.NewEchoHandler(),
		func(err error) { reportedErrors <- err },
	)
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
	runtimeContext := built.RuntimeContext()
	if runtimeContext == nil {
		t.Fatal("RuntimeContext() = nil after successful Start")
	}
	address := net.JoinHostPort(snapshot.Listener.Host, portString(snapshot.Listener.Port))
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
	bootstrap, err := NewBootstrap(apiKeyResolver(t), message.NewEchoHandler())
	if err != nil {
		t.Fatalf("NewBootstrap() error = %v", err)
	}
	snapshot := apiKeySnapshot(t)
	host, err := bootstrap.Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() { _ = host.Stop(context.Background()) })

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	header := make(http.Header)
	header.Set("X-API-Key", "wrong-secret")
	connection, response, err := websocket.Dial(
		ctx,
		"ws://127.0.0.1:"+portString(snapshot.Listener.Port)+"/ws",
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
	bootstrap, err := NewBootstrap(emptyResolver(t), nil)
	if err != nil {
		t.Fatalf("NewBootstrap() error = %v", err)
	}
	snapshot := apiKeySnapshot(t)
	host, err := bootstrap.Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() { _ = host.Stop(context.Background()) })

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	header := make(http.Header)
	header.Set("X-API-Key", "credential")
	connection, response, err := websocket.Dial(
		ctx,
		"ws://127.0.0.1:"+portString(snapshot.Listener.Port)+"/ws",
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

func TestBootstrapReportsSessionTerminalErrorOnce(t *testing.T) {
	wantErr := errors.New("handler failed with credential-that-must-not-leak")
	reportedErrors := make(chan error, 2)
	bootstrap, err := NewBootstrapWithTerminalErrorReporter(
		apiKeyResolver(t),
		runtimeErrorHandler{err: wantErr},
		func(err error) { reportedErrors <- err },
	)
	if err != nil {
		t.Fatalf("NewBootstrapWithTerminalErrorReporter() error = %v", err)
	}
	snapshot := apiKeySnapshot(t)
	host, err := bootstrap.Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() { _ = host.Stop(context.Background()) })

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	header := make(http.Header)
	header.Set("X-API-Key", "runtime-secret")
	connection, _, err := websocket.Dial(
		ctx,
		"ws://127.0.0.1:"+portString(snapshot.Listener.Port)+"/ws",
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
	case reported := <-reportedErrors:
		if !errors.Is(reported, wantErr) {
			t.Fatalf("reported error = %v, want Handler sentinel", reported)
		}
		if reported.Error() != "Session handoff failed" {
			t.Fatalf("reported error text = %q, want redacted terminal category", reported.Error())
		}
	case <-ctx.Done():
		t.Fatalf("terminal Session error was not reported: %v", ctx.Err())
	}
	select {
	case <-readResult:
	case <-ctx.Done():
		t.Fatalf("Session connection did not close: %v", ctx.Err())
	}
	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	select {
	case duplicate := <-reportedErrors:
		t.Fatalf("terminal error reported more than once: %v", duplicate)
	default:
	}
}

func TestBootstrapHostDisabledAuthenticationUsesAnonymousSession(t *testing.T) {
	bootstrap, err := NewBootstrap(emptyResolver(t), message.NewEchoHandler())
	if err != nil {
		t.Fatalf("NewBootstrap() error = %v", err)
	}
	snapshot := apiKeySnapshot(t)
	snapshot.Authentication = runtimeconfig.AuthenticationSnapshot{}
	host, err := bootstrap.Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() { _ = host.Stop(context.Background()) })

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	connection, response, err := websocket.Dial(
		ctx,
		"ws://127.0.0.1:"+portString(snapshot.Listener.Port)+"/ws",
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
	bootstrap, err := NewBootstrap(resolver, message.NewEchoHandler())
	if err != nil {
		t.Fatalf("NewBootstrap() error = %v", err)
	}
	snapshot := apiKeySnapshot(t)
	snapshot.Authentication.Providers = []runtimeconfig.AuthenticationProviderSnapshot{
		{
			Name:    "runtime-jwt",
			Type:    runtimeconfig.AuthenticationProviderJWT,
			Enabled: true,
			JWT: &runtimeconfig.JWTSnapshot{
				SigningKeys:       []runtimeconfig.JWTSigningKeySnapshot{{Name: "primary", SecretRef: secretRef}},
				AllowedAlgorithms: []runtimeconfig.JWTAlgorithm{runtimeconfig.HS256},
			},
		},
	}
	host, err := bootstrap.Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
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
		"ws://127.0.0.1:"+portString(snapshot.Listener.Port)+"/ws",
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

type runtimeErrorHandler struct {
	err error
}

func (handler runtimeErrorHandler) Handle(context.Context, message.Sender, message.Message) error {
	return handler.err
}
