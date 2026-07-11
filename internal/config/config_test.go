package config

import (
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	unsetEnvironment(t, "UWP_HTTP_HOST", "UWP_HTTP_PORT", "UWP_LOG_LEVEL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTP.Host != "127.0.0.1" {
		t.Errorf("HTTP.Host = %q, want %q", cfg.HTTP.Host, "127.0.0.1")
	}
	if cfg.HTTP.Port != 8080 {
		t.Errorf("HTTP.Port = %d, want %d", cfg.HTTP.Port, 8080)
	}
	if cfg.Log.Level != slog.LevelInfo {
		t.Errorf("Log.Level = %s, want %s", cfg.Log.Level, slog.LevelInfo)
	}
}

func TestLoadEnvironment(t *testing.T) {
	t.Setenv("UWP_HTTP_HOST", "0.0.0.0")
	t.Setenv("UWP_HTTP_PORT", "9090")
	t.Setenv("UWP_LOG_LEVEL", "debug")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTP.Host != "0.0.0.0" {
		t.Errorf("HTTP.Host = %q, want %q", cfg.HTTP.Host, "0.0.0.0")
	}
	if cfg.HTTP.Port != 9090 {
		t.Errorf("HTTP.Port = %d, want %d", cfg.HTTP.Port, 9090)
	}
	if cfg.Log.Level != slog.LevelDebug {
		t.Errorf("Log.Level = %s, want %s", cfg.Log.Level, slog.LevelDebug)
	}
}

func TestLoadValidation(t *testing.T) {
	tests := []struct {
		name      string
		variable  string
		value     string
		wantError string
	}{
		{name: "empty host", variable: "UWP_HTTP_HOST", value: "", wantError: "must not be empty"},
		{name: "whitespace host", variable: "UWP_HTTP_HOST", value: "  ", wantError: "must not be empty"},
		{name: "non-numeric port", variable: "UWP_HTTP_PORT", value: "http", wantError: "between 1 and 65535"},
		{name: "port below range", variable: "UWP_HTTP_PORT", value: "0", wantError: "between 1 and 65535"},
		{name: "port above range", variable: "UWP_HTTP_PORT", value: "65536", wantError: "between 1 and 65535"},
		{name: "unknown log level", variable: "UWP_LOG_LEVEL", value: "trace", wantError: "must be one of"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unsetEnvironment(t, "UWP_HTTP_HOST", "UWP_HTTP_PORT", "UWP_LOG_LEVEL")
			t.Setenv(tt.variable, tt.value)

			_, err := Load()
			if err == nil {
				t.Fatal("Load() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("Load() error = %q, want it to contain %q", err, tt.wantError)
			}
		})
	}
}

func unsetEnvironment(t *testing.T, names ...string) {
	t.Helper()

	for _, name := range names {
		value, exists := os.LookupEnv(name)
		if err := os.Unsetenv(name); err != nil {
			t.Fatalf("os.Unsetenv(%q): %v", name, err)
		}

		t.Cleanup(func() {
			if exists {
				_ = os.Setenv(name, value)
				return
			}
			_ = os.Unsetenv(name)
		})
	}
}
