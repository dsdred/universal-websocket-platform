package http

import (
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	server := New("127.0.0.1:8080", nil)
	request := httptest.NewRequest(stdhttp.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != stdhttp.StatusOK {
		t.Errorf("status code = %d, want %d", response.Code, stdhttp.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want it to contain %q", contentType, "application/json")
	}
	if response.Body.String() != `{"status":"ok"}` {
		t.Errorf("body = %q, want %q", response.Body.String(), `{"status":"ok"}`)
	}
}

func TestServerTimeouts(t *testing.T) {
	server := New("127.0.0.1:8080", nil)

	if server.ReadHeaderTimeout != readHeaderTimeout {
		t.Errorf("ReadHeaderTimeout = %s, want %s", server.ReadHeaderTimeout, readHeaderTimeout)
	}
	if server.ReadTimeout != readTimeout {
		t.Errorf("ReadTimeout = %s, want %s", server.ReadTimeout, readTimeout)
	}
	if server.WriteTimeout != writeTimeout {
		t.Errorf("WriteTimeout = %s, want %s", server.WriteTimeout, writeTimeout)
	}
	if server.IdleTimeout != idleTimeout {
		t.Errorf("IdleTimeout = %s, want %s", server.IdleTimeout, idleTimeout)
	}
}
