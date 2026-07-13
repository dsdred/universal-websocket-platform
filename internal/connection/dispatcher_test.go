package connection

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestDefaultDispatcherImplementsDispatcher(t *testing.T) {
	var _ Dispatcher = DefaultDispatcher{}
}

func TestDefaultDispatcherReceivesContextAndClosesConnection(t *testing.T) {
	received := make(chan ConnectionContext, 1)
	dispatched := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		websocketConnection, err := websocket.Accept(response, request, nil)
		if err != nil {
			dispatched <- fmt.Errorf("Accept(): %w", err)
			return
		}

		connectionContext := NewContext(request.Context(), websocketConnection, request)
		received <- connectionContext
		dispatched <- (DefaultDispatcher{}).Dispatch(connectionContext)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	websocketConnection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer websocketConnection.CloseNow()

	select {
	case connectionContext := <-received:
		if connectionContext.Context() == nil {
			t.Fatal("ConnectionContext.Context() = nil")
		}
		if connectionContext.Connection() == nil {
			t.Fatal("ConnectionContext.Connection() = nil")
		}
		if connectionContext.Request() == nil {
			t.Fatal("ConnectionContext.Request() = nil")
		}
	case <-ctx.Done():
		t.Fatalf("Dispatcher did not receive ConnectionContext: %v", ctx.Err())
	}

	_, _, readErr := websocketConnection.Read(ctx)
	if status := websocket.CloseStatus(readErr); status != websocket.StatusNormalClosure {
		t.Fatalf("close status = %d, want %d (error %v)", status, websocket.StatusNormalClosure, readErr)
	}

	select {
	case dispatchErr := <-dispatched:
		if dispatchErr != nil {
			t.Fatalf("Dispatch() error = %v", dispatchErr)
		}
	case <-ctx.Done():
		t.Fatalf("Dispatch() did not return: %v", ctx.Err())
	}
}
