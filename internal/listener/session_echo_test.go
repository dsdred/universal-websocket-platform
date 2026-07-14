package listener

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/dsdred/universal-websocket-platform/internal/message"
	platformsession "github.com/dsdred/universal-websocket-platform/internal/session"
)

func TestEchoHandlerProductionPipeline(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher(message.NewEchoHandler()))
	address := listener.Address()
	websocketConnection, response := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("handshake status = %d, want %d", response.StatusCode, http.StatusSwitchingProtocols)
	}

	for _, expected := range []struct {
		messageType websocket.MessageType
		payload     []byte
	}{
		{messageType: websocket.MessageText, payload: []byte("first text")},
		{messageType: websocket.MessageBinary, payload: []byte{0x00, 0x01, 0xff}},
		{messageType: websocket.MessageText, payload: []byte("second text")},
	} {
		writeSessionMessage(t, websocketConnection, expected.messageType, expected.payload)
		actualType, actualPayload := readEchoMessage(t, websocketConnection)
		if actualType != expected.messageType || !bytes.Equal(actualPayload, expected.payload) {
			t.Fatalf("echo = (%d, %v), want (%d, %v)", actualType, actualPayload, expected.messageType, expected.payload)
		}
	}
	if err := websocketConnection.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("client Close() error = %v", err)
	}
	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Listener Stop() error = %v", err)
	}
	assertTCPPortAvailable(t, address)
}

func TestEchoHandlerConcurrentClients(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher(message.NewEchoHandler()))
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
			for _, messageType := range []websocket.MessageType{websocket.MessageText, websocket.MessageBinary} {
				payload := []byte(fmt.Sprintf("client-%d-type-%d", index, messageType))
				if err := websocketConnection.Write(ctx, messageType, payload); err != nil {
					errorsChannel <- fmt.Errorf("client %d Write(): %w", index, err)
					return
				}
				actualType, actualPayload, err := websocketConnection.Read(ctx)
				if err != nil {
					errorsChannel <- fmt.Errorf("client %d Read(): %w", index, err)
					return
				}
				if actualType != messageType || !bytes.Equal(actualPayload, payload) {
					errorsChannel <- fmt.Errorf("client %d echo = (%d, %v), want (%d, %v)", index, actualType, actualPayload, messageType, payload)
					return
				}
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

func TestEchoHandlerListenerStopTerminatesAllSessions(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher(message.NewEchoHandler()))
	address := listener.Address()
	const clients = 8
	connections := make([]*websocket.Conn, 0, clients)
	for index := range clients {
		websocketConnection, _ := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
		connections = append(connections, websocketConnection)
		payload := []byte(fmt.Sprintf("active-%d", index))
		writeSessionMessage(t, websocketConnection, websocket.MessageText, payload)
		actualType, actualPayload := readEchoMessage(t, websocketConnection)
		if actualType != websocket.MessageText || !bytes.Equal(actualPayload, payload) {
			t.Fatalf("echo = (%d, %v), want text %v", actualType, actualPayload, payload)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := listener.Stop(ctx); err != nil {
		t.Fatalf("Listener Stop() error = %v", err)
	}
	for _, websocketConnection := range connections {
		if _, _, err := websocketConnection.Read(ctx); err == nil {
			t.Fatal("WebSocket remained active after Listener Stop")
		}
		websocketConnection.CloseNow()
	}
	assertTCPPortAvailable(t, address)
}

func TestEchoHandlerSmokeScenario(t *testing.T) {
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, platformsession.NewDispatcher(message.NewEchoHandler()))
	address := listener.Address()
	websocketConnection, response := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	writeSessionMessage(t, websocketConnection, websocket.MessageText, []byte("echo"))
	messageType, payload := readEchoMessage(t, websocketConnection)
	if messageType != websocket.MessageText || string(payload) != "echo" {
		t.Fatalf("echo = (%d, %q)", messageType, payload)
	}
	if err := websocketConnection.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("client Close() error = %v", err)
	}
	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Listener Stop() error = %v", err)
	}
	assertTCPPortAvailable(t, address)
	t.Logf("Upgrade %d -> Authentication -> Session.Run -> EchoHandler -> Session.Send -> client echo -> port released", response.StatusCode)
}

func readEchoMessage(t *testing.T, connection *websocket.Conn) (websocket.MessageType, []byte) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	messageType, payload, err := connection.Read(ctx)
	if err != nil {
		t.Fatalf("client Read() error = %v", err)
	}
	return messageType, payload
}
