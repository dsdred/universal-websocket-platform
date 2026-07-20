package runtime

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/router"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

func TestRuntimeStartupConstructsMessageRouterExactlyOnce(t *testing.T) {
	snapshot := routingRuntimeSnapshot(t, nil)
	var constructions atomic.Int32
	capabilities := &handshakeCapabilities{
		canAccept:      func() bool { return true },
		runtimeContext: context.Background,
	}

	runtimeListener, err := composeRuntimeWithRouterFactory(
		snapshot,
		emptyResolver(t),
		message.NewEchoHandler(),
		capabilities,
		nil,
		func(routing *runtimeconfig.RoutingSnapshot, handler message.Handler) (message.Handler, error) {
			constructions.Add(1)
			if routing != nil {
				t.Fatal("absent Routing reached factory as present")
			}
			return router.NewCompatibility(handler), nil
		},
	)
	if err != nil {
		t.Fatalf("composeRuntimeWithRouterFactory() error = %v", err)
	}
	if runtimeListener == nil {
		t.Fatal("composeRuntimeWithRouterFactory() returned nil Listener")
	}
	if constructions.Load() != 1 {
		t.Fatalf("Router constructions = %d, want 1", constructions.Load())
	}
}

func TestRuntimeCompatibilityRouterPreservesLegacyHandler(t *testing.T) {
	handler := &routingIntegrationHandler{echo: true}
	snapshot := routingRuntimeSnapshot(t, nil)
	host := startRoutingRuntime(t, snapshot, handler, nil)
	connection := dialRoutingRuntime(t, snapshot)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	payload := []byte("compatibility")
	if err := connection.Write(ctx, websocket.MessageText, payload); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	messageType, got, err := connection.Read(ctx)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if messageType != websocket.MessageText || string(got) != string(payload) {
		t.Fatalf("Read() = (%v, %q), want text %q", messageType, got, payload)
	}
	if handler.calls.Load() != 1 {
		t.Fatalf("Handler calls = %d, want 1", handler.calls.Load())
	}
	closeRoutingRuntime(t, host, connection)
}

func TestRuntimeCompiledRouterSelectsRouteAndPreservesNoMatch(t *testing.T) {
	observed := make(chan message.Type, 2)
	handler := &routingIntegrationHandler{echo: true, observed: observed}
	snapshot := routingRuntimeSnapshot(t, textRouteRouting())
	host := startRoutingRuntime(t, snapshot, handler, nil)
	connection := dialRoutingRuntime(t, snapshot)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := connection.Write(ctx, websocket.MessageBinary, []byte("no-match")); err != nil {
		t.Fatalf("Write(binary) error = %v", err)
	}
	if err := connection.Write(ctx, websocket.MessageText, []byte("selected")); err != nil {
		t.Fatalf("Write(text) error = %v", err)
	}
	messageType, payload, err := connection.Read(ctx)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if messageType != websocket.MessageText || string(payload) != "selected" {
		t.Fatalf("Read() = (%v, %q), want selected text Message", messageType, payload)
	}
	if handler.calls.Load() != 1 {
		t.Fatalf("Handler calls = %d, want 1", handler.calls.Load())
	}
	select {
	case got := <-observed:
		if got != message.TypeText {
			t.Fatalf("selected Message type = %q, want text", got)
		}
	default:
		t.Fatal("selected Handler invocation was not observed")
	}
	closeRoutingRuntime(t, host, connection)
}

func TestRuntimeCompiledRouterDispatchesDefaultHandler(t *testing.T) {
	handler := &routingIntegrationHandler{echo: true}
	snapshot := routingRuntimeSnapshot(t, &configurationversion.RoutingSettings{DefaultHandlerRef: legacyHandlerReference})
	host := startRoutingRuntime(t, snapshot, handler, nil)
	connection := dialRoutingRuntime(t, snapshot)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	payload := []byte{0x00, 0x01, 0xfe, 0xff}
	if err := connection.Write(ctx, websocket.MessageBinary, payload); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	messageType, got, err := connection.Read(ctx)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if messageType != websocket.MessageBinary || string(got) != string(payload) {
		t.Fatalf("Read() = (%v, %v), want binary %v", messageType, got, payload)
	}
	if handler.calls.Load() != 1 {
		t.Fatalf("Handler calls = %d, want 1", handler.calls.Load())
	}
	closeRoutingRuntime(t, host, connection)
}

func TestRuntimeCompatibilityRouterWithNilHandlerPreservesDiscard(t *testing.T) {
	snapshot := routingRuntimeSnapshot(t, nil)
	host := startRoutingRuntime(t, snapshot, nil, nil)
	connection := dialRoutingRuntime(t, snapshot)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for _, messageType := range []websocket.MessageType{websocket.MessageText, websocket.MessageBinary} {
		if err := connection.Write(ctx, messageType, []byte("discarded")); err != nil {
			t.Fatalf("Write(%v) error = %v", messageType, err)
		}
	}
	closeRoutingRuntime(t, host, connection)
}

func TestRuntimeRouterHandlerErrorPreservesTerminalPath(t *testing.T) {
	wantErr := errors.New("routing Handler failure")
	handler := &routingIntegrationHandler{err: wantErr}
	reportedErrors := make(chan error, 1)
	snapshot := routingRuntimeSnapshot(t, textRouteRouting())
	host := startRoutingRuntime(t, snapshot, handler, func(err error) { reportedErrors <- err })
	connection := dialRoutingRuntime(t, snapshot)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := connection.Write(ctx, websocket.MessageText, []byte("failure")); err != nil {
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
			t.Fatalf("reported error = %v, want Handler cause", reported)
		}
	case <-ctx.Done():
		_ = connection.CloseNow()
		<-readResult
		t.Fatalf("Handler error was not reported: %v (Handler calls: %d)", ctx.Err(), handler.calls.Load())
	}
	select {
	case <-readResult:
	case <-ctx.Done():
		_ = connection.CloseNow()
		<-readResult
		t.Fatalf("Session connection did not close: %v", ctx.Err())
	}
	if handler.calls.Load() != 1 {
		t.Fatalf("Handler calls = %d, want 1", handler.calls.Load())
	}
	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestRuntimeStartupFailsWhenRouterConstructionFails(t *testing.T) {
	snapshot := routingRuntimeSnapshot(t, &configurationversion.RoutingSettings{Routes: []configurationversion.Route{{
		ID:         "unresolved",
		Enabled:    true,
		Priority:   1,
		HandlerRef: "future",
		Matchers: []configurationversion.Matcher{{
			Type:  configurationversion.MatcherTypeMessageType,
			Value: "text",
		}},
	}}})
	bootstrap, err := NewBootstrap(emptyResolver(t), message.NewEchoHandler())
	if err != nil {
		t.Fatalf("NewBootstrap() error = %v", err)
	}
	host, err := bootstrap.Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	err = host.Start(context.Background())
	if !errors.Is(err, router.ErrUnresolvedHandlerRef) {
		t.Fatalf("Start() error = %v, want ErrUnresolvedHandlerRef", err)
	}
	if host.Running() || host.Ready() || host.CanAccept() || host.RuntimeContext() != nil {
		t.Fatal("Router construction failure published active Runtime state")
	}
	assertPortAvailable(t, snapshot.Listener.Port)
}

func TestRuntimeCompiledRouterHandlesConcurrentSessions(t *testing.T) {
	handler := &routingIntegrationHandler{echo: true}
	snapshot := routingRuntimeSnapshot(t, textRouteRouting())
	host := startRoutingRuntime(t, snapshot, handler, nil)

	const clients = 16
	start := make(chan struct{})
	results := make(chan error, clients)
	var waitGroup sync.WaitGroup
	waitGroup.Add(clients)
	for client := range clients {
		go func() {
			defer waitGroup.Done()
			<-start
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			connection, response, err := websocket.Dial(ctx, routingRuntimeURL(snapshot), nil)
			if err != nil {
				results <- fmt.Errorf("Dial() error: %w", err)
				return
			}
			defer connection.CloseNow()
			if response == nil || response.StatusCode != http.StatusSwitchingProtocols {
				results <- fmt.Errorf("Dial() status = %v", response)
				return
			}
			payload := []byte(fmt.Sprintf("client-%d", client))
			if err := connection.Write(ctx, websocket.MessageText, payload); err != nil {
				results <- fmt.Errorf("Write() error: %w", err)
				return
			}
			messageType, got, err := connection.Read(ctx)
			if err != nil {
				results <- fmt.Errorf("Read() error: %w", err)
				return
			}
			if messageType != websocket.MessageText || string(got) != string(payload) {
				results <- fmt.Errorf("Read() = (%v, %q), want text %q", messageType, got, payload)
				return
			}
			if err := connection.Close(websocket.StatusNormalClosure, ""); err != nil {
				results <- fmt.Errorf("Close() error: %w", err)
				return
			}
			results <- nil
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Error(err)
		}
	}
	if handler.calls.Load() != clients {
		t.Fatalf("Handler calls = %d, want %d", handler.calls.Load(), clients)
	}
	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func routingRuntimeSnapshot(t *testing.T, routing *configurationversion.RoutingSettings) runtimeconfig.Snapshot {
	t.Helper()
	snapshot, err := runtimeconfig.NewBuilder().Build(configurationversion.ConfigurationVersion{
		ID:              1,
		ConfigurationID: 1,
		State:           configurationversion.Published,
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
		Routing: routing,
	})
	if err != nil {
		t.Fatalf("runtimeconfig.Build() error = %v", err)
	}
	return snapshot
}

func textRouteRouting() *configurationversion.RoutingSettings {
	return &configurationversion.RoutingSettings{Routes: []configurationversion.Route{{
		ID:         "text",
		Enabled:    true,
		Priority:   1,
		HandlerRef: legacyHandlerReference,
		Matchers: []configurationversion.Matcher{{
			Type:  configurationversion.MatcherTypeMessageType,
			Value: "text",
		}},
	}}}
}

func startRoutingRuntime(
	t *testing.T,
	snapshot runtimeconfig.Snapshot,
	handler message.Handler,
	reportError func(error),
) Host {
	t.Helper()
	bootstrap, err := NewBootstrapWithTerminalErrorReporter(emptyResolver(t), handler, reportError)
	if err != nil {
		t.Fatalf("NewBootstrapWithTerminalErrorReporter() error = %v", err)
	}
	host, err := bootstrap.Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() { _ = host.Stop(context.Background()) })
	return host
}

func dialRoutingRuntime(t *testing.T, snapshot runtimeconfig.Snapshot) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	connection, response, err := websocket.Dial(ctx, routingRuntimeURL(snapshot), nil)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	if response == nil || response.StatusCode != http.StatusSwitchingProtocols {
		connection.CloseNow()
		t.Fatalf("Dial() status = %v, want 101", response)
	}
	t.Cleanup(func() { _ = connection.CloseNow() })
	return connection
}

func closeRoutingRuntime(t *testing.T, host Host, connection *websocket.Conn) {
	t.Helper()
	if err := connection.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func routingRuntimeURL(snapshot runtimeconfig.Snapshot) string {
	return "ws://127.0.0.1:" + portString(snapshot.Listener.Port) + "/ws"
}

type routingIntegrationHandler struct {
	calls    atomic.Int64
	echo     bool
	err      error
	observed chan<- message.Type
}

func (handler *routingIntegrationHandler) Handle(ctx context.Context, runtimeContext message.Context) error {
	handler.calls.Add(1)
	if handler.observed != nil {
		handler.observed <- runtimeContext.MessageType()
	}
	if handler.err != nil {
		return handler.err
	}
	if handler.echo {
		return runtimeContext.Sender().Send(ctx, runtimeContext.Message())
	}
	return nil
}
