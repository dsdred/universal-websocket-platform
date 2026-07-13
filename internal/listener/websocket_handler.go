package listener

import (
	"net/http"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/connection"
)

const websocketPath = "/ws"

type websocketHandler struct {
	dispatcher connection.Dispatcher
}

func newHTTPHandler(dispatcher connection.Dispatcher) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(websocketPath, websocketHandler{dispatcher: dispatcher})
	mux.HandleFunc("/", notImplementedHandler)
	return mux
}

func (handler websocketHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		response.Header().Set("Allow", http.MethodGet)
		http.Error(response, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	websocketConnection, err := websocket.Accept(response, request, nil)
	if err != nil {
		return
	}

	connectionContext := connection.NewContext(request.Context(), websocketConnection, request)
	_ = handler.dispatcher.Dispatch(connectionContext)
}
