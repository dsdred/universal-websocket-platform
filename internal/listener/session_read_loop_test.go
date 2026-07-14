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

func TestSessionReadLoopProductionPipeline(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher())
	address := listener.Address()

	websocketConnection, response := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("handshake status = %d, want %d", response.StatusCode, http.StatusSwitchingProtocols)
	}
	writeSessionMessage(t, websocketConnection, websocket.MessageText, []byte("text payload"))
	writeSessionMessage(t, websocketConnection, websocket.MessageBinary, []byte{0x00, 0x01, 0x02})
	writeSessionMessage(t, websocketConnection, websocket.MessageText, []byte("still open"))
	if err := websocketConnection.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("client Close() error = %v", err)
	}

	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Listener Stop() error = %v", err)
	}
	assertTCPPortAvailable(t, address)
}

func TestSessionReadLoopMissingCredentialsDoesNotCreateSessionOrStopListener(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher())

	rejectedConnection, _ := dialWebSocketWithHeader(t, listener, "X-API-Key", "")
	if status := readWebSocketClose(t, rejectedConnection); status != websocket.StatusPolicyViolation {
		t.Fatalf("rejected close status = %d, want %d", status, websocket.StatusPolicyViolation)
	}
	rejectedConnection.CloseNow()

	acceptedConnection, _ := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	writeSessionMessage(t, acceptedConnection, websocket.MessageText, []byte("accepted"))
	if err := acceptedConnection.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("accepted client Close() error = %v", err)
	}
}

func TestSessionReadLoopConcurrentClients(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher())
	const clients = 16

	errorsChannel := make(chan error, clients)
	var waitGroup sync.WaitGroup
	for index := range clients {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			websocketConnection, response, err := websocket.Dial(ctx, websocketURL(listener, websocketPath), &websocket.DialOptions{
				HTTPHeader: http.Header{"X-API-Key": []string{"correct-key"}},
			})
			if err != nil {
				errorsChannel <- fmt.Errorf("client %d Dial(): %w", index, err)
				return
			}
			defer websocketConnection.CloseNow()
			if response.StatusCode != http.StatusSwitchingProtocols {
				errorsChannel <- fmt.Errorf("client %d handshake status = %d", index, response.StatusCode)
				return
			}
			if err := websocketConnection.Write(ctx, websocket.MessageText, []byte("text")); err != nil {
				errorsChannel <- fmt.Errorf("client %d text Write(): %w", index, err)
				return
			}
			if err := websocketConnection.Write(ctx, websocket.MessageBinary, []byte{byte(index)}); err != nil {
				errorsChannel <- fmt.Errorf("client %d binary Write(): %w", index, err)
				return
			}
			if err := websocketConnection.Close(websocket.StatusNormalClosure, ""); err != nil {
				errorsChannel <- fmt.Errorf("client %d Close(): %w", index, err)
			}
		}()
	}
	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		t.Error(err)
	}
}

func TestListenerStopTerminatesActiveSessionReadLoop(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher())
	address := listener.Address()
	websocketConnection, _ := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	defer websocketConnection.CloseNow()
	writeSessionMessage(t, websocketConnection, websocket.MessageText, []byte("active"))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := listener.Stop(ctx); err != nil {
		t.Fatalf("Listener Stop() error = %v", err)
	}
	if _, _, err := websocketConnection.Read(ctx); err == nil {
		t.Fatal("active WebSocket remained readable after Listener Stop")
	}
	assertTCPPortAvailable(t, address)
}

func TestSessionReadLoopSmokeScenario(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher())
	address := listener.Address()
	websocketConnection, response := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	writeSessionMessage(t, websocketConnection, websocket.MessageText, []byte("text"))
	writeSessionMessage(t, websocketConnection, websocket.MessageBinary, []byte{0x01})
	if err := websocketConnection.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("client Close() error = %v", err)
	}
	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Listener Stop() error = %v", err)
	}
	assertTCPPortAvailable(t, address)
	t.Logf("Upgrade %d -> Authentication -> Session.Run -> text/binary read -> client close -> port released", response.StatusCode)
}

func writeSessionMessage(t *testing.T, connection *websocket.Conn, messageType websocket.MessageType, payload []byte) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := connection.Write(ctx, messageType, payload); err != nil {
		t.Fatalf("Write(%d) error = %v", messageType, err)
	}
}
