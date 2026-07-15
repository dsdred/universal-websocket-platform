package listener

import (
	"net/http"
)

const websocketPath = "/ws"

func newHTTPHandlerWithHandshake(handshakeHandler http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(websocketPath, websocketMethodHandler{next: handshakeHandler})
	mux.HandleFunc("/", notImplementedHandler)
	return mux
}

type websocketMethodHandler struct {
	next http.Handler
}

func (handler websocketMethodHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		response.Header().Set("Allow", http.MethodGet)
		http.Error(response, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	handler.next.ServeHTTP(response, request)
}
