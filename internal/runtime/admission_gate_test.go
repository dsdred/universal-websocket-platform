package runtime

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestHostAdmissionGateLifecycle(t *testing.T) {
	runtimeListener := newControlledListener(nil, false)
	runtimeListener.stopEntered = make(chan struct{})
	runtimeListener.releaseStop = make(chan struct{})
	host := newTestHost(t, fixedComposer(runtimeListener))
	var gateClosedAtListenerStop atomic.Bool
	runtimeListener.stopObserver = func() {
		gateClosedAtListenerStop.Store(!host.CanAccept())
	}

	if host.CanAccept() {
		t.Fatal("CanAccept() = true after NewHost")
	}
	if err := host.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if host.CanAccept() {
		t.Fatal("CanAccept() = true after Build")
	}

	runtimeListener.startObserver = func() {
		if !runtimeListener.Running() {
			t.Error("Listener is not Running before startup commit")
		}
		if host.CanAccept() {
			t.Error("CanAccept() = true before startup commit")
		}
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !host.Running() || !host.Ready() || !host.CanAccept() {
		t.Fatal("Host is not Running, Ready, and accepting after startup commit")
	}
	assertConcurrentAdmission(t, host, true)

	stopResult := make(chan error, 1)
	go func() { stopResult <- host.Stop(context.Background()) }()
	waitForSignal(t, runtimeListener.stopEntered, "Listener.Stop entry")
	if !gateClosedAtListenerStop.Load() {
		t.Fatal("admission Gate was open when Listener.Stop began")
	}
	if host.CanAccept() {
		t.Fatal("CanAccept() = true after shutdown began")
	}
	assertConcurrentAdmission(t, host, false)
	close(runtimeListener.releaseStop)
	waitForResult(t, stopResult, "Stop")

	if host.CanAccept() {
		t.Fatal("CanAccept() = true after Stop")
	}
	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("repeated Stop() error = %v", err)
	}
	if host.CanAccept() {
		t.Fatal("CanAccept() = true after repeated Stop")
	}
}

func TestHostAdmissionGateDuringStarting(t *testing.T) {
	runtimeListener := newControlledListener(nil, true)
	host := newTestHost(t, fixedComposer(runtimeListener))
	if err := host.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	startResult := make(chan error, 1)
	go func() { startResult <- host.Start(context.Background()) }()
	waitForSignal(t, runtimeListener.startEntered, "Listener.Start entry")
	if host.CanAccept() {
		t.Fatal("CanAccept() = true during Starting")
	}
	assertConcurrentAdmission(t, host, false)

	close(runtimeListener.releaseStart)
	waitForResult(t, startResult, "Start")
	if !host.CanAccept() {
		t.Fatal("CanAccept() = false after startup commit")
	}
	assertConcurrentAdmission(t, host, true)

	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func assertConcurrentAdmission(t *testing.T, host *DefaultHost, want bool) {
	t.Helper()
	const readers = 32

	begin := make(chan struct{})
	results := make(chan bool, readers)
	var ready sync.WaitGroup
	ready.Add(readers)
	for range readers {
		go func() {
			ready.Done()
			<-begin
			results <- host.CanAccept()
		}()
	}
	ready.Wait()
	close(begin)

	for reader := range readers {
		select {
		case got := <-results:
			if got != want {
				t.Fatalf("reader %d: CanAccept() = %v, want %v", reader, got, want)
			}
		case <-time.After(time.Second):
			t.Fatalf("reader %d: CanAccept() deadlocked", reader)
		}
	}
}
