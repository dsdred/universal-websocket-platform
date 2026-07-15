package listener

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"

	platformconnection "github.com/dsdred/universal-websocket-platform/internal/connection"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	platformsession "github.com/dsdred/universal-websocket-platform/internal/session"
)

func TestSessionWriterProductionPipeline(t *testing.T) {
	dispatcher := newExposingSessionDispatcher()
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, dispatcher)
	address := listener.Address()
	websocketConnection, response := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("handshake status = %d, want %d", response.StatusCode, http.StatusSwitchingProtocols)
	}
	runtimeSession := dispatcher.nextSession(t)

	textMessage := newOutboundMessage(t, message.TypeText, []byte("server text"))
	if err := runtimeSession.Send(context.Background(), textMessage); err != nil {
		t.Fatalf("text Send() error = %v", err)
	}
	assertClientMessage(t, websocketConnection, websocket.MessageText, []byte("server text"))

	binaryPayload := []byte{0x00, 0x01, 0xff}
	binaryMessage := newOutboundMessage(t, message.TypeBinary, binaryPayload)
	if err := runtimeSession.Send(context.Background(), binaryMessage); err != nil {
		t.Fatalf("binary Send() error = %v", err)
	}
	assertClientMessage(t, websocketConnection, websocket.MessageBinary, binaryPayload)

	if err := websocketConnection.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("client Close() error = %v", err)
	}
	dispatcher.wait(t)
	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Listener Stop() error = %v", err)
	}
	assertTCPPortAvailable(t, address)
}

func TestSessionWriterConcurrentSendIntegration(t *testing.T) {
	dispatcher := newExposingSessionDispatcher()
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, dispatcher)
	websocketConnection, _ := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	runtimeSession := dispatcher.nextSession(t)
	const messages = 32

	errorsChannel := make(chan error, messages)
	var waitGroup sync.WaitGroup
	for index := range messages {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			runtimeMessage, err := message.New(message.TypeText, []byte(fmt.Sprintf("outbound-%d", index)))
			if err != nil {
				errorsChannel <- err
				return
			}
			if err := runtimeSession.Send(context.Background(), runtimeMessage); err != nil {
				errorsChannel <- err
			}
		}()
	}
	received := make(map[string]struct{}, messages)
	for range messages {
		messageType, payload := readOutboundClientMessage(t, websocketConnection)
		if messageType != websocket.MessageText {
			t.Fatalf("received type = %d, want text", messageType)
		}
		received[string(payload)] = struct{}{}
	}
	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		t.Errorf("concurrent Send() error = %v", err)
	}
	if len(received) != messages {
		t.Fatalf("unique received messages = %d, want %d", len(received), messages)
	}
	if err := websocketConnection.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("client Close() error = %v", err)
	}
	dispatcher.wait(t)
}

func TestSessionWriterSmokeScenario(t *testing.T) {
	dispatcher := newExposingSessionDispatcher()
	service := listenerAPIKeyService(t, listenerMemoryResolver(t, "correct-key"))
	listener := startedListenerWithAuthentication(t, service, dispatcher)
	address := listener.Address()
	websocketConnection, response := dialWebSocketWithHeader(t, listener, "X-API-Key", "correct-key")
	runtimeSession := dispatcher.nextSession(t)
	if err := runtimeSession.Send(context.Background(), newOutboundMessage(t, message.TypeText, []byte("text"))); err != nil {
		t.Fatalf("text Send() error = %v", err)
	}
	assertClientMessage(t, websocketConnection, websocket.MessageText, []byte("text"))
	if err := runtimeSession.Send(context.Background(), newOutboundMessage(t, message.TypeBinary, []byte{0x01})); err != nil {
		t.Fatalf("binary Send() error = %v", err)
	}
	assertClientMessage(t, websocketConnection, websocket.MessageBinary, []byte{0x01})
	if err := websocketConnection.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("client Close() error = %v", err)
	}
	dispatcher.wait(t)
	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Listener Stop() error = %v", err)
	}
	assertTCPPortAvailable(t, address)
	t.Logf("Upgrade %d -> Authentication -> Session.Run + Send(text/binary) -> client close -> port released", response.StatusCode)
}

type exposingSessionDispatcher struct {
	sessions chan platformsession.Session
	done     chan error
}

func newExposingSessionDispatcher() *exposingSessionDispatcher {
	return &exposingSessionDispatcher{
		sessions: make(chan platformsession.Session, 1),
		done:     make(chan error, 1),
	}
}

func (dispatcher *exposingSessionDispatcher) DispatchAuthenticated(authenticatedContext platformconnection.AuthenticatedContext) (bool, error) {
	connectionContext := authenticatedContext.ConnectionContext()
	runtimeSession, err := platformsession.New(
		connectionContext.Connection(),
		authenticatedContext.Principal(),
		connectionContext.Request().RemoteAddr,
	)
	if err != nil {
		dispatcher.done <- err
		return false, err
	}
	if err := runtimeSession.Start(connectionContext.Context()); err != nil {
		_ = runtimeSession.Stop(context.Background())
		dispatcher.done <- err
		return false, err
	}
	dispatcher.sessions <- runtimeSession
	runErr := runtimeSession.Run(connectionContext.Context())
	stopErr := runtimeSession.Stop(context.Background())
	if runErr != nil {
		dispatcher.done <- runErr
		return true, runErr
	}
	dispatcher.done <- stopErr
	return true, stopErr
}

func (dispatcher *exposingSessionDispatcher) nextSession(t *testing.T) platformsession.Session {
	t.Helper()
	select {
	case runtimeSession := <-dispatcher.sessions:
		return runtimeSession
	case <-time.After(time.Second):
		t.Fatal("Session Dispatcher did not create Session")
		return nil
	}
}

func (dispatcher *exposingSessionDispatcher) wait(t *testing.T) {
	t.Helper()
	select {
	case err := <-dispatcher.done:
		if err != nil {
			t.Fatalf("Session Dispatcher error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Session Dispatcher did not finish")
	}
}

func newOutboundMessage(t *testing.T, messageType message.Type, payload []byte) message.Message {
	t.Helper()
	runtimeMessage, err := message.New(messageType, payload)
	if err != nil {
		t.Fatalf("message.New() error = %v", err)
	}
	return runtimeMessage
}

func assertClientMessage(
	t *testing.T,
	connection *websocket.Conn,
	expectedType websocket.MessageType,
	expectedPayload []byte,
) {
	t.Helper()
	messageType, payload := readOutboundClientMessage(t, connection)
	if messageType != expectedType || string(payload) != string(expectedPayload) {
		t.Fatalf("client message = (%d, %v), want (%d, %v)", messageType, payload, expectedType, expectedPayload)
	}
}

func readOutboundClientMessage(t *testing.T, connection *websocket.Conn) (websocket.MessageType, []byte) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	messageType, payload, err := connection.Read(ctx)
	if err != nil {
		t.Fatalf("client Read() error = %v", err)
	}
	return messageType, payload
}
