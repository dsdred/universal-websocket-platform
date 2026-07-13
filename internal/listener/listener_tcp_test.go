package listener

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

func TestListenerStartOpensTCPPortAndClosesAcceptedConnection(t *testing.T) {
	listener := mustListener(t)
	if err := listener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	connection, err := net.DialTimeout("tcp", listener.Address(), time.Second)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer connection.Close()
	if err := connection.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline() error = %v", err)
	}

	buffer := make([]byte, 1)
	_, err = connection.Read(buffer)
	if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
		t.Fatalf("Read() error = %v, want closed connection", err)
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

	replacement, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatalf("port was not released: %v", err)
	}
	if err := replacement.Close(); err != nil {
		t.Fatalf("replacement Close() error = %v", err)
	}
}

func TestListenerTCPSmokeScenario(t *testing.T) {
	listener := mustListener(t)
	if err := listener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	connection, err := net.DialTimeout("tcp", listener.Address(), time.Second)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	if err := connection.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline() error = %v", err)
	}
	_, readErr := connection.Read(make([]byte, 1))
	_ = connection.Close()
	if !errors.Is(readErr, io.EOF) && !errors.Is(readErr, net.ErrClosed) {
		t.Fatalf("accepted connection was not closed: %v", readErr)
	}

	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if listener.Running() {
		t.Fatal("Running() = true after Stop")
	}
	t.Log("Listener -> Start -> net.Dial -> accepted -> closed -> Stop")
}
