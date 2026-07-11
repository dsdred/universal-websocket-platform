package http

import (
	stdhttp "net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// New creates the Control Service HTTP server.
func New(address string) *stdhttp.Server {
	router := chi.NewRouter()
	router.Get("/health", health)

	return &stdhttp.Server{
		Addr:              address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func health(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(stdhttp.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
