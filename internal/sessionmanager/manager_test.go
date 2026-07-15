package sessionmanager

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

const (
	concurrencyIterations = 100
	concurrentCalls       = 32
)

func TestManagerInitialState(t *testing.T) {
	manager := New()

	if got := manager.State(); got != StateOpen {
		t.Fatalf("State() = %v, want StateOpen", got)
	}
}

func TestManagerBeginShutdown(t *testing.T) {
	manager := New()

	manager.BeginShutdown()

	if got := manager.State(); got != StateClosing {
		t.Fatalf("State() = %v, want StateClosing", got)
	}
}

func TestManagerRepeatedBeginShutdown(t *testing.T) {
	manager := New()
	manager.BeginShutdown()

	for range 10 {
		manager.BeginShutdown()
	}

	if got := manager.State(); got != StateClosing {
		t.Fatalf("State() = %v, want StateClosing", got)
	}
}

func TestManagerWaitInClosingAndClosed(t *testing.T) {
	manager := New()
	manager.BeginShutdown()

	if err := manager.Wait(context.Background()); err != nil {
		t.Fatalf("Wait() in Closing error = %v", err)
	}
	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() = %v, want StateClosed", got)
	}
	if err := manager.Wait(context.Background()); err != nil {
		t.Fatalf("Wait() in Closed error = %v", err)
	}
}

func TestManagerWaitInOpen(t *testing.T) {
	manager := New()

	err := manager.Wait(context.Background())

	if !errors.Is(err, ErrShutdownNotStarted) {
		t.Fatalf("Wait() error = %v, want ErrShutdownNotStarted", err)
	}
	if got := manager.State(); got != StateOpen {
		t.Fatalf("State() = %v, want StateOpen", got)
	}
}

func TestManagerConcurrentBeginShutdown(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		start := make(chan struct{})
		done := make(chan struct{})
		var waitGroup sync.WaitGroup

		for range concurrentCalls {
			waitGroup.Add(1)
			go func() {
				defer waitGroup.Done()
				<-start
				manager.BeginShutdown()
			}()
		}

		go func() {
			waitGroup.Wait()
			close(done)
		}()
		close(start)
		waitForCompletion(t, done, "concurrent BeginShutdown")

		if got := manager.State(); got != StateClosing {
			t.Fatalf("iteration %d: State() = %v, want StateClosing", iteration, got)
		}
	}
}

func TestManagerConcurrentWait(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		manager.BeginShutdown()
		start := make(chan struct{})
		results := make(chan error, concurrentCalls)

		for range concurrentCalls {
			go func() {
				<-start
				results <- manager.Wait(context.Background())
			}()
		}

		close(start)
		for call := range concurrentCalls {
			select {
			case err := <-results:
				if err != nil {
					t.Fatalf("iteration %d, call %d: Wait() error = %v", iteration, call, err)
				}
			case <-time.After(time.Second):
				t.Fatalf("iteration %d: concurrent Wait deadlocked", iteration)
			}
		}
		if got := manager.State(); got != StateClosed {
			t.Fatalf("iteration %d: State() = %v, want StateClosed", iteration, got)
		}
	}
}

func TestManagerBeginShutdownWaitRace(t *testing.T) {
	for iteration := range concurrencyIterations {
		manager := New()
		start := make(chan struct{})
		beginDone := make(chan struct{})
		waitResult := make(chan error, 1)

		go func() {
			<-start
			manager.BeginShutdown()
			close(beginDone)
		}()
		go func() {
			<-start
			waitResult <- manager.Wait(context.Background())
		}()

		close(start)
		waitForCompletion(t, beginDone, "BeginShutdown race participant")
		var raceErr error
		select {
		case raceErr = <-waitResult:
		case <-time.After(time.Second):
			t.Fatalf("iteration %d: Wait race participant deadlocked", iteration)
		}
		if raceErr != nil && !errors.Is(raceErr, ErrShutdownNotStarted) {
			t.Fatalf("iteration %d: Wait() error = %v", iteration, raceErr)
		}

		manager.BeginShutdown()
		if err := manager.Wait(context.Background()); err != nil {
			t.Fatalf("iteration %d: final Wait() error = %v", iteration, err)
		}
		if got := manager.State(); got != StateClosed {
			t.Fatalf("iteration %d: State() = %v, want StateClosed", iteration, got)
		}
	}
}

func TestManagerRestartProhibited(t *testing.T) {
	manager := New()
	manager.BeginShutdown()
	if err := manager.Wait(context.Background()); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}

	manager.BeginShutdown()

	if got := manager.State(); got != StateClosed {
		t.Fatalf("State() after repeated BeginShutdown = %v, want StateClosed", got)
	}
	if err := manager.Wait(context.Background()); err != nil {
		t.Fatalf("Wait() after repeated BeginShutdown error = %v", err)
	}
}

func waitForCompletion(t *testing.T, done <-chan struct{}, operation string) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("%s did not complete", operation)
	}
}
