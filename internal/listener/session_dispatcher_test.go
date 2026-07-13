package listener

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"

	platformsession "github.com/dsdred/universal-websocket-platform/internal/session"
)

func TestSessionDispatcherListenerProductionPipeline(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher())
	address := listener.Address()

	websocketConnection, response := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	defer websocketConnection.CloseNow()
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("handshake status = %d, want %d", response.StatusCode, http.StatusSwitchingProtocols)
	}
	if status := readWebSocketClose(t, websocketConnection); status != websocket.StatusNormalClosure {
		t.Fatalf("close status = %d, want %d", status, websocket.StatusNormalClosure)
	}

	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	assertTCPPortAvailable(t, address)
}

func TestSessionDispatcherListenerRejectsMissingCredentialsAndContinues(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher())

	rejectedConnection, _ := dialWebSocketWithHeader(t, listener, "X-API-Key", "")
	defer rejectedConnection.CloseNow()
	if status := readWebSocketClose(t, rejectedConnection); status != websocket.StatusPolicyViolation {
		t.Fatalf("rejected close status = %d, want %d", status, websocket.StatusPolicyViolation)
	}

	acceptedConnection, _ := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	defer acceptedConnection.CloseNow()
	if status := readWebSocketClose(t, acceptedConnection); status != websocket.StatusNormalClosure {
		t.Fatalf("accepted close status = %d, want %d", status, websocket.StatusNormalClosure)
	}
}

func TestSessionDispatcherListenerConcurrentConnections(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher())
	const connections = 16

	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, connections)
	for range connections {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			websocketConnection, response, err := websocket.Dial(ctx, websocketURL(listener, websocketPath), &websocket.DialOptions{
				HTTPHeader: http.Header{"X-API-Key": []string{"correct-key"}},
			})
			if err != nil {
				errorsChannel <- fmt.Errorf("Dial(): %w", err)
				return
			}
			defer websocketConnection.CloseNow()
			if response.StatusCode != http.StatusSwitchingProtocols {
				errorsChannel <- fmt.Errorf("handshake status = %d", response.StatusCode)
				return
			}
			_, _, readErr := websocketConnection.Read(ctx)
			if status := websocket.CloseStatus(readErr); status != websocket.StatusNormalClosure {
				errorsChannel <- fmt.Errorf("close status = %d: %w", status, readErr)
			}
		}()
	}
	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		t.Errorf("concurrent connection error = %v", err)
	}
}

func TestSessionDispatcherListenerSmokeScenario(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher())
	address := listener.Address()

	websocketConnection, response := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	status := readWebSocketClose(t, websocketConnection)
	websocketConnection.CloseNow()
	if response.StatusCode != http.StatusSwitchingProtocols || status != websocket.StatusNormalClosure {
		t.Fatalf("pipeline statuses = (%d, %d), want (%d, %d)", response.StatusCode, status, http.StatusSwitchingProtocols, websocket.StatusNormalClosure)
	}
	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	assertTCPPortAvailable(t, address)
	t.Logf("Upgrade %d -> Authentication -> Session -> close %d -> Stop -> port released", response.StatusCode, status)
}
