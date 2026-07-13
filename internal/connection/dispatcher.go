package connection

import "github.com/coder/websocket"

// Dispatcher receives ownership of an upgraded WebSocket connection.
type Dispatcher interface {
	Dispatch(ConnectionContext) error
}

// DefaultDispatcher closes accepted connections until downstream components exist.
type DefaultDispatcher struct{}

// Dispatch closes the WebSocket with a normal closure and returns no error.
func (DefaultDispatcher) Dispatch(connectionContext ConnectionContext) error {
	websocketConnection := connectionContext.Connection()
	defer websocketConnection.CloseNow()

	_ = websocketConnection.Close(websocket.StatusNormalClosure, "")
	return nil
}
