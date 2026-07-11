package main

import (
	"log/slog"
	"net"
	"os"

	"github.com/dsdred/universal-websocket-platform/internal/config"
	httpserver "github.com/dsdred/universal-websocket-platform/internal/http"
	applog "github.com/dsdred/universal-websocket-platform/internal/log"
)

func main() {
	logger := applog.New(os.Stdout)
	cfg := config.Load()
	address := net.JoinHostPort("", cfg.Port)
	server := httpserver.New(address)

	logger.Info("starting control service", slog.String("address", address))

	if err := server.ListenAndServe(); err != nil {
		logger.Error("control service stopped", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
