package listener

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestListenerServesNotImplementedForNonWebSocketPaths(t *testing.T) {
	listener := mustListener(t)
	if err := listener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	response, err := testHTTPClient().Get("http://" + listener.Address() + "/")
	if err != nil {
		t.Fatalf("GET / error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotImplemented {
		t.Fatalf("GET / status = %d, want %d", response.StatusCode, http.StatusNotImplemented)
	}

	otherResponse, err := testHTTPClient().Get("http://" + listener.Address() + "/anything")
	if err != nil {
		t.Fatalf("GET /anything error = %v", err)
	}
	defer otherResponse.Body.Close()
	if otherResponse.StatusCode != http.StatusNotImplemented {
		t.Fatalf("GET /anything status = %d, want %d", otherResponse.StatusCode, http.StatusNotImplemented)
	}
}

func TestListenerStopReleasesTCPPort(t *testing.T) {
	listener := mustListener(t)
	if err := listener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	address := listener.Address()
	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	assertTCPPortAvailable(t, address)
}

func assertTCPPortAvailable(t *testing.T, address string) {
	t.Helper()
	replacement, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatalf("port was not released: %v", err)
	}
	if err := replacement.Close(); err != nil {
		t.Fatalf("replacement Close() error = %v", err)
	}
}

func TestListenerShutdownStopsHTTPServer(t *testing.T) {
	listener := mustListener(t)
	if err := listener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	response, err := testHTTPClient().Get("http://" + listener.Address() + "/")
	if err != nil {
		t.Fatalf("GET / error = %v", err)
	}
	if err := response.Body.Close(); err != nil {
		t.Fatalf("response Body.Close() error = %v", err)
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := listener.Stop(shutdownContext); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	if _, err := testHTTPClient().Get("http://" + listener.Address() + "/"); err == nil {
		t.Fatal("GET / after Stop succeeded, want connection error")
	}
}

func TestListenerHTTPSmokeScenario(t *testing.T) {
	listener := mustListener(t)
	if err := listener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	response, err := testHTTPClient().Get("http://" + listener.Address() + "/")
	if err != nil {
		t.Fatalf("GET / error = %v", err)
	}
	if err := response.Body.Close(); err != nil {
		t.Fatalf("response Body.Close() error = %v", err)
	}
	if response.StatusCode != http.StatusNotImplemented {
		t.Fatalf("GET / status = %d, want %d", response.StatusCode, http.StatusNotImplemented)
	}

	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	t.Log("Listener -> HTTP GET / -> 501 Not Implemented -> Stop")
}

func testHTTPClient() *http.Client {
	return &http.Client{Timeout: time.Second}
}
