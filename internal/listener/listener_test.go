package listener

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestListenerStartAndStop(t *testing.T) {
	listener := mustListener(t)
	if err := listener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !listener.Running() {
		t.Fatal("Running() = false after Start")
	}
	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if listener.Running() {
		t.Fatal("Running() = true after Stop")
	}
	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("second Stop() error = %v", err)
	}
}

func TestListenerDoubleStart(t *testing.T) {
	listener := mustListener(t)
	if err := listener.Start(context.Background()); err != nil {
		t.Fatalf("first Start() error = %v", err)
	}
	if err := listener.Start(context.Background()); !errors.Is(err, ErrListenerAlreadyRunning) {
		t.Fatalf("second Start() error = %v, want ErrListenerAlreadyRunning", err)
	}
}

func TestListenerDoesNotSupportRestart(t *testing.T) {
	listener := mustListener(t)
	if err := listener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if err := listener.Start(context.Background()); !errors.Is(err, ErrListenerAlreadyRunning) {
		t.Fatalf("Start() after Stop error = %v, want ErrListenerAlreadyRunning", err)
	}
}

func TestListenerStopBeforeStartIsNoOp(t *testing.T) {
	listener := mustListener(t)
	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if listener.Running() {
		t.Fatal("Running() = true after no-op Stop")
	}
	if err := listener.Start(context.Background()); err != nil {
		t.Fatalf("Start() after no-op Stop error = %v", err)
	}
}

func TestListenerConcurrentStart(t *testing.T) {
	listener := mustListener(t)
	const goroutines = 64
	var successes atomic.Int64
	var alreadyRunning atomic.Int64
	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, goroutines)
	for range goroutines {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := listener.Start(context.Background())
			switch {
			case err == nil:
				successes.Add(1)
			case errors.Is(err, ErrListenerAlreadyRunning):
				alreadyRunning.Add(1)
			default:
				errorsChannel <- err
			}
		}()
	}
	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		t.Errorf("Start() unexpected error = %v", err)
	}
	if successes.Load() != 1 || alreadyRunning.Load() != goroutines-1 {
		t.Fatalf("Start outcomes = (%d success, %d already running)", successes.Load(), alreadyRunning.Load())
	}
}

func TestListenerConcurrentStop(t *testing.T) {
	listener := mustListener(t)
	if err := listener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	const goroutines = 64
	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, goroutines)
	for range goroutines {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			if err := listener.Stop(context.Background()); err != nil {
				errorsChannel <- err
			}
		}()
	}
	done := make(chan struct{})
	go func() {
		waitGroup.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("concurrent Stop() calls did not complete")
	}
	close(errorsChannel)
	for err := range errorsChannel {
		t.Errorf("Stop() error = %v", err)
	}
	if listener.Running() {
		t.Fatal("Running() = true after concurrent Stop")
	}
}

func TestListenerSmokeScenario(t *testing.T) {
	listener := mustListener(t)
	if err := listener.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Logf("Snapshot -> Bootstrap -> Listener -> Start: Running=%t", listener.Running())
	if err := listener.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	t.Logf("Listener -> Stop: Running=%t", listener.Running())
	if listener.Running() {
		t.Fatal("Running() = true after Stop")
	}
}

func mustListener(t *testing.T) Listener {
	t.Helper()
	snapshot := validListenerSnapshot()
	snapshot.Port = availableTCPPort(t)
	listener, err := (DefaultBootstrap{}).Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	t.Cleanup(func() {
		if err := listener.Stop(context.Background()); err != nil {
			t.Errorf("cleanup Stop() error = %v", err)
		}
	})
	return listener
}

func availableTCPPort(t *testing.T) uint16 {
	t.Helper()
	tcpListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve TCP port: %v", err)
	}
	port := tcpListener.Addr().(*net.TCPAddr).Port
	if err := tcpListener.Close(); err != nil {
		t.Fatalf("release reserved TCP port: %v", err)
	}
	return uint16(port)
}
