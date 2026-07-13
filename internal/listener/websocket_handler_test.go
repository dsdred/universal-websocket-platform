package listener

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestListenerRootStillReturnsNotImplemented(t *testing.T) {
	listener := startedListener(t)

	response, err := testHTTPClient().Get("http://" + listener.Address() + "/")
	if err != nil {
		t.Fatalf("GET / error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotImplemented {
		t.Fatalf("GET / status = %d, want %d", response.StatusCode, http.StatusNotImplemented)
	}
}

func TestListenerWebSocketPathWithoutUpgradeDoesNotSwitchProtocols(t *testing.T) {
	listener := startedListener(t)

	response, err := testHTTPClient().Get("http://" + listener.Address() + websocketPath)
	if err != nil {
		t.Fatalf("GET %s error = %v", websocketPath, err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusSwitchingProtocols {
		t.Fatalf("GET %s status = %d, want HTTP error", websocketPath, response.StatusCode)
	}
}

func TestListenerRejectsWebSocketUpgradeWithPOSTAndContinuesServing(t *testing.T) {
	listener := startedListener(t)

	request, err := http.NewRequest(http.MethodPost, "http://"+listener.Address()+websocketPath, nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	request.Header.Set("Connection", "Upgrade")
	request.Header.Set("Upgrade", "websocket")
	request.Header.Set("Sec-WebSocket-Version", "13")
	request.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	response, err := testHTTPClient().Do(request)
	if err != nil {
		t.Fatalf("POST %s error = %v", websocketPath, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("POST %s status = %d, want %d", websocketPath, response.StatusCode, http.StatusMethodNotAllowed)
	}
	if response.StatusCode == http.StatusSwitchingProtocols {
		t.Fatalf("POST %s established a WebSocket connection", websocketPath)
	}
	if allow := response.Header.Get("Allow"); allow != http.MethodGet {
		t.Fatalf("POST %s Allow = %q, want %q", websocketPath, allow, http.MethodGet)
	}

	rootResponse, err := testHTTPClient().Get("http://" + listener.Address() + "/")
	if err != nil {
		t.Fatalf("GET / after POST %s error = %v", websocketPath, err)
	}
	if err := rootResponse.Body.Close(); err != nil {
		t.Fatalf("GET / response Body.Close() error = %v", err)
	}
	if rootResponse.StatusCode != http.StatusNotImplemented {
		t.Fatalf("GET / after POST %s status = %d, want %d", websocketPath, rootResponse.StatusCode, http.StatusNotImplemented)
	}

	connection, handshakeResponse := dialWebSocket(t, listener, websocketPath)
	defer connection.CloseNow()
	if handshakeResponse.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("WebSocket handshake status = %d, want %d", handshakeResponse.StatusCode, http.StatusSwitchingProtocols)
	}

	readContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, _, readErr := connection.Read(readContext)
	if status := websocket.CloseStatus(readErr); status != websocket.StatusNormalClosure {
		t.Fatalf("WebSocket close status = %d, want %d", status, websocket.StatusNormalClosure)
	}
}

func TestListenerWebSocketHandshakeAndNormalClosure(t *testing.T) {
	listener := startedListener(t)

	connection, response := dialWebSocket(t, listener, websocketPath)
	defer connection.CloseNow()
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("WebSocket handshake status = %d, want %d", response.StatusCode, http.StatusSwitchingProtocols)
	}

	readContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, _, err := connection.Read(readContext)
	if status := websocket.CloseStatus(err); status != websocket.StatusNormalClosure {
		t.Fatalf("WebSocket close status = %d, want %d (error %v)", status, websocket.StatusNormalClosure, err)
	}
}

func TestListenerRejectsWebSocketUpgradeOnWrongPath(t *testing.T) {
	listener := startedListener(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	connection, response, err := websocket.Dial(ctx, websocketURL(listener, "/not-websocket"), nil)
	if connection != nil {
		connection.CloseNow()
		t.Fatal("WebSocket Dial() on wrong path succeeded")
	}
	if err == nil {
		t.Fatal("WebSocket Dial() on wrong path error = nil")
	}
	if response == nil {
		t.Fatal("WebSocket Dial() on wrong path response = nil")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotImplemented {
		t.Fatalf("wrong path status = %d, want %d", response.StatusCode, http.StatusNotImplemented)
	}
}

func TestListenerContinuesAfterFailedWebSocketUpgrade(t *testing.T) {
	listener := startedListener(t)

	failedResponse, err := testHTTPClient().Get("http://" + listener.Address() + websocketPath)
	if err != nil {
		t.Fatalf("GET %s error = %v", websocketPath, err)
	}
	if err := failedResponse.Body.Close(); err != nil {
		t.Fatalf("failed upgrade response Body.Close() error = %v", err)
	}
	if failedResponse.StatusCode == http.StatusSwitchingProtocols {
		t.Fatalf("GET %s unexpectedly switched protocols", websocketPath)
	}

	response, err := testHTTPClient().Get("http://" + listener.Address() + "/")
	if err != nil {
		t.Fatalf("GET / after failed upgrade error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotImplemented {
		t.Fatalf("GET / after failed upgrade status = %d, want %d", response.StatusCode, http.StatusNotImplemented)
	}
}

func TestListenerStopAfterWebSocketHandshake(t *testing.T) {
	listener := startedListener(t)

	connection, _ := dialWebSocket(t, listener, websocketPath)
	readContext, cancelRead := context.WithTimeout(context.Background(), time.Second)
	_, _, readErr := connection.Read(readContext)
	cancelRead()
	connection.CloseNow()
	if status := websocket.CloseStatus(readErr); status != websocket.StatusNormalClosure {
		t.Fatalf("WebSocket close status = %d, want %d", status, websocket.StatusNormalClosure)
	}

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), time.Second)
	defer cancelShutdown()
	if err := listener.Stop(shutdownContext); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if listener.Running() {
		t.Fatal("Running() = true after Stop")
	}
}

func TestListenerWebSocketSmokeScenario(t *testing.T) {
	listener := startedListener(t)
	address := listener.Address()

	connection, response := dialWebSocket(t, listener, websocketPath)
	readContext, cancel := context.WithTimeout(context.Background(), time.Second)
	_, _, readErr := connection.Read(readContext)
	cancel()
	connection.CloseNow()
	if websocket.CloseStatus(readErr) != websocket.StatusNormalClosure {
		t.Fatalf("WebSocket close error = %v", readErr)
	}

	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	assertTCPPortAvailable(t, address)
	t.Logf("Start -> WebSocket handshake status %d -> normal closure -> Stop -> port released", response.StatusCode)
}

func startedListener(t *testing.T) Listener {
	t.Helper()
	listener := mustListener(t)
	if err := listener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	return listener
}

func dialWebSocket(t *testing.T, listener Listener, path string) (*websocket.Conn, *http.Response) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	connection, response, err := websocket.Dial(ctx, websocketURL(listener, path), nil)
	if err != nil {
		t.Fatalf("WebSocket Dial() error = %v", err)
	}
	if response == nil {
		connection.CloseNow()
		t.Fatal("WebSocket Dial() response = nil")
	}
	return connection, response
}

func websocketURL(listener Listener, path string) string {
	return "ws://" + listener.Address() + "/" + strings.TrimPrefix(path, "/")
}
