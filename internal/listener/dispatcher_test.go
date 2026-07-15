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
)

func TestListenerDispatchesConnectionContext(t *testing.T) {
	dispatcher := &recordingDispatcher{
		received: make(chan platformconnection.ConnectionContext, 1),
		delegate: platformconnection.DefaultDispatcher{},
	}
	listener := startedListenerWithDispatcher(t, dispatcher)

	websocketConnection, _ := dialWebSocket(t, listener, websocketPath)
	defer websocketConnection.CloseNow()

	select {
	case connectionContext := <-dispatcher.received:
		if connectionContext.Context() == nil {
			t.Fatal("ConnectionContext.Context() = nil")
		}
		if connectionContext.Connection() == nil {
			t.Fatal("ConnectionContext.Connection() = nil")
		}
		if connectionContext.Request() == nil {
			t.Fatal("ConnectionContext.Request() = nil")
		}
		if connectionContext.Request().URL.Path != websocketPath {
			t.Fatalf("ConnectionContext request path = %q, want %q", connectionContext.Request().URL.Path, websocketPath)
		}
	case <-time.After(time.Second):
		t.Fatal("Dispatcher did not receive ConnectionContext")
	}

	readContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, _, readErr := websocketConnection.Read(readContext)
	if status := websocket.CloseStatus(readErr); status != websocket.StatusNormalClosure {
		t.Fatalf("close status = %d, want %d", status, websocket.StatusNormalClosure)
	}

	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	t.Log("TCP -> HTTP -> Upgrade -> Dispatcher -> Normal Closure -> Stop")
}

func TestListenerDispatchesConcurrentWebSocketConnections(t *testing.T) {
	listener := startedListener(t)
	const connections = 16

	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, connections)
	for range connections {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			websocketConnection, response, err := websocket.Dial(ctx, websocketURL(listener, websocketPath), nil)
			if err != nil {
				errorsChannel <- fmt.Errorf("Dial(): %w", err)
				return
			}
			defer websocketConnection.CloseNow()
			if response == nil || response.StatusCode != http.StatusSwitchingProtocols {
				errorsChannel <- fmt.Errorf("handshake response = %v", response)
				return
			}

			_, _, readErr := websocketConnection.Read(ctx)
			if status := websocket.CloseStatus(readErr); status != websocket.StatusNormalClosure {
				errorsChannel <- fmt.Errorf("close status = %d: %w", status, readErr)
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		waitGroup.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("concurrent WebSocket connections did not complete")
	}
	close(errorsChannel)
	for err := range errorsChannel {
		t.Errorf("concurrent connection error = %v", err)
	}

	response, err := testHTTPClient().Get("http://" + listener.Address() + "/")
	if err != nil {
		t.Fatalf("GET / after concurrent connections error = %v", err)
	}
	if err := response.Body.Close(); err != nil {
		t.Fatalf("response Body.Close() error = %v", err)
	}
	if response.StatusCode != http.StatusNotImplemented {
		t.Fatalf("GET / status = %d, want %d", response.StatusCode, http.StatusNotImplemented)
	}

	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

type recordingDispatcher struct {
	received chan platformconnection.ConnectionContext
	delegate platformconnection.Dispatcher
}

func (dispatcher *recordingDispatcher) Dispatch(connectionContext platformconnection.ConnectionContext) error {
	dispatcher.received <- connectionContext
	return dispatcher.delegate.Dispatch(connectionContext)
}

func startedListenerWithDispatcher(t *testing.T, dispatcher platformconnection.Dispatcher) Listener {
	t.Helper()
	snapshot := validListenerSnapshot()
	snapshot.Port = availableTCPPort(t)
	listener, err := NewBootstrapWithHandshake(testWebSocketHandler{dispatcher: dispatcher}).Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	t.Cleanup(func() {
		if err := listener.Stop(context.Background()); err != nil {
			t.Errorf("cleanup Stop() error = %v", err)
		}
	})
	if err := listener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	return listener
}

type testWebSocketHandler struct {
	dispatcher platformconnection.Dispatcher
}

func (handler testWebSocketHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	websocketConnection, err := websocket.Accept(response, request, nil)
	if err != nil {
		return
	}
	connectionContext := platformconnection.NewContext(request.Context(), websocketConnection, request)
	_ = handler.dispatcher.Dispatch(connectionContext)
}
