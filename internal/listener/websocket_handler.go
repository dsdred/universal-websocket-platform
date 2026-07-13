package listener

import (
	"net/http"

	"github.com/coder/websocket"
)

const websocketPath = "/ws"

type websocketHandler struct{}

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle(websocketPath, websocketHandler{})
	mux.HandleFunc("/", notImplementedHandler)
	return mux
}

func (websocketHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		response.Header().Set("Allow", http.MethodGet)
		http.Error(response, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	connection, err := websocket.Accept(response, request, nil)
	if err != nil {
		return
	}
	defer connection.CloseNow()

	_ = connection.Close(websocket.StatusNormalClosure, "")
}
