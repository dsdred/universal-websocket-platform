package runtime

import (
	"context"
	"testing"
	"time"
)

func TestHostReadinessDuringStarting(t *testing.T) {
	runtimeListener := newControlledListener(nil, true)
	host := newTestHost(t, fixedComposer(runtimeListener))
	if err := host.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	startResult := make(chan error, 1)
	go func() { startResult <- host.Start(context.Background()) }()
	waitForSignal(t, runtimeListener.startEntered, "Listener.Start entry")

	if got := currentHostState(host); got != hostStarting {
		t.Fatalf("Host state = %v, want hostStarting", got)
	}
	if host.Ready() {
		t.Fatal("Ready() = true while Listener.Start is blocked")
	}
	if host.RuntimeContext() != nil {
		t.Fatal("RuntimeContext() was published while Listener.Start is blocked")
	}
	assertConcurrentReadiness(t, host, false)

	close(runtimeListener.releaseStart)
	waitForResult(t, startResult, "Start")
	if !host.Running() || !host.Ready() {
		t.Fatal("Host is not Running and Ready after successful Start")
	}
	if host.RuntimeContext() == nil {
		t.Fatal("RuntimeContext() is nil after successful Start")
	}
	assertConcurrentReadiness(t, host, true)

	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func assertConcurrentReadiness(t *testing.T, host *DefaultHost, want bool) {
	t.Helper()
	const readers = 32
	begin := make(chan struct{})
	results := make(chan bool, readers)
	for range readers {
		go func() {
			<-begin
			results <- host.Ready()
		}()
	}
	close(begin)
	for reader := range readers {
		select {
		case got := <-results:
			if got != want {
				t.Fatalf("reader %d: Ready() = %t, want %t", reader, got, want)
			}
		case <-time.After(time.Second):
			t.Fatal("concurrent Ready() readers deadlocked")
		}
	}
}
