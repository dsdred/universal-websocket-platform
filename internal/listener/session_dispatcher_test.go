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
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher(nil))
	address := listener.Address()

	websocketConnection, response := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	defer websocketConnection.CloseNow()
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("handshake status = %d, want %d", response.StatusCode, http.StatusSwitchingProtocols)
	}
	if err := websocketConnection.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("client Close() error = %v", err)
	}

	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	assertTCPPortAvailable(t, address)
}

func TestSessionDispatcherListenerRejectsMissingCredentialsAndContinues(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher(nil))

	rejectedConnection, _ := dialWebSocketWithHeader(t, listener, "X-API-Key", "")
	defer rejectedConnection.CloseNow()
	if status := readWebSocketClose(t, rejectedConnection); status != websocket.StatusPolicyViolation {
		t.Fatalf("rejected close status = %d, want %d", status, websocket.StatusPolicyViolation)
	}

	acceptedConnection, _ := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	defer acceptedConnection.CloseNow()
	if err := acceptedConnection.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("accepted client Close() error = %v", err)
	}
}

func TestSessionDispatcherListenerConcurrentConnections(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher(nil))
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
			if err := websocketConnection.Close(websocket.StatusNormalClosure, ""); err != nil {
				errorsChannel <- fmt.Errorf("client Close(): %w", err)
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
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher(nil))
	address := listener.Address()

	websocketConnection, response := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	if err := websocketConnection.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("client Close() error = %v", err)
	}
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("pipeline handshake status = %d, want %d", response.StatusCode, http.StatusSwitchingProtocols)
	}
	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	assertTCPPortAvailable(t, address)
	t.Logf("Upgrade %d -> Authentication -> Session.Run -> client normal closure -> Stop -> port released", response.StatusCode)
}
