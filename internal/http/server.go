package http

import (
	stdhttp "net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 60 * time.Second
)

// New creates the Control Service HTTP server.
func New(address string, registerRoutes func(chi.Router)) *stdhttp.Server {
	router := chi.NewRouter()
	router.Get("/health", health)
	if registerRoutes != nil {
		registerRoutes(router)
	}

	return &stdhttp.Server{
		Addr:              address,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

func health(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(stdhttp.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
