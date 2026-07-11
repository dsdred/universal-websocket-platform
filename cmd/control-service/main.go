package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/config"
	httpserver "github.com/dsdred/universal-websocket-platform/internal/http"
	applog "github.com/dsdred/universal-websocket-platform/internal/log"
)

const shutdownTimeout = 10 * time.Second

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.Load()
	if err != nil {
		applog.New(os.Stdout, slog.LevelInfo).Error("invalid configuration", slog.String("error", err.Error()))
		return 1
	}

	logger := applog.New(os.Stdout, cfg.Log.Level)
	address := net.JoinHostPort(cfg.HTTP.Host, strconv.Itoa(cfg.HTTP.Port))
	server := httpserver.New(address)
	logLevel := strings.ToLower(cfg.Log.Level.String())

	logger.Info(
		"starting service",
		slog.String("service", "control-service"),
		slog.String("address", address),
		slog.String("log_level", logLevel),
	)

	signalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return 0
		}

		logger.Error("control service stopped", slog.String("error", err.Error()))
		return 1
	case <-signalContext.Done():
		logger.Info("shutdown signal received", slog.String("service", "control-service"))
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		return 1
	}

	if err := <-serverErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("control service stopped during shutdown", slog.String("error", err.Error()))
		return 1
	}

	logger.Info("control service stopped gracefully", slog.String("service", "control-service"))
	return 0
}
