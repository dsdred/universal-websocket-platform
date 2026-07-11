package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

const (
	defaultHTTPHost = "127.0.0.1"
	defaultHTTPPort = 8080
	defaultLogLevel = "info"
)

// Config contains the Control Service startup configuration.
type Config struct {
	HTTP HTTPConfig
	Log  LogConfig
}

// HTTPConfig contains the HTTP Server configuration.
type HTTPConfig struct {
	Host string
	Port int
}

// LogConfig contains the logging configuration.
type LogConfig struct {
	Level slog.Level
}

// Load reads configuration from the environment, applies defaults, and validates it.
func Load() (Config, error) {
	host := environmentValue("UWP_HTTP_HOST", defaultHTTPHost)
	portValue := environmentValue("UWP_HTTP_PORT", strconv.Itoa(defaultHTTPPort))
	logLevelValue := environmentValue("UWP_LOG_LEVEL", defaultLogLevel)

	if strings.TrimSpace(host) == "" {
		return Config{}, fmt.Errorf("UWP_HTTP_HOST must not be empty")
	}

	port, err := strconv.Atoi(portValue)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("UWP_HTTP_PORT must be an integer between 1 and 65535, got %q", portValue)
	}

	logLevel, err := parseLogLevel(logLevelValue)
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTP: HTTPConfig{
			Host: host,
			Port: port,
		},
		Log: LogConfig{Level: logLevel},
	}, nil
}

func environmentValue(name, fallback string) string {
	value, exists := os.LookupEnv(name)
	if !exists {
		return fallback
	}

	return value
}

func parseLogLevel(value string) (slog.Level, error) {
	switch value {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("UWP_LOG_LEVEL must be one of debug, info, warn, error, got %q", value)
	}
}
