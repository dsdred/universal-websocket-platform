package executionbinding

import (
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestCreateBinding(t *testing.T) {
	binding := New()
	if binding == nil {
		t.Fatal("New() returned a nil Binding")
	}
}

func TestBindingInitialState(t *testing.T) {
	binding := New()

	if binding.state.outcome != 0 {
		t.Fatalf("initial outcome = %d, want unpublished zero value", binding.state.outcome)
	}
	if binding.state.waitClaimed.Load() {
		t.Fatal("dormant path is claimed before Wait")
	}
	select {
	case <-binding.state.ready:
		t.Fatal("Binding is published immediately after construction")
	default:
	}
}

func TestBindingCopySharesOnePublicationAndWaiter(t *testing.T) {
	original := New()
	copy := *original

	waitResult := make(chan Outcome, 1)
	waitError := make(chan error, 1)
	go func() {
		outcome, err := original.Wait()
		if err != nil {
			waitError <- err
			return
		}
		waitResult <- outcome
	}()
	waitForClaim(t, original.state)

	if _, err := copy.Wait(); !errors.Is(err, ErrWaitAlreadyClaimed) {
		t.Fatalf("Wait() through copy error = %v, want ErrWaitAlreadyClaimed", err)
	}

	published, err := copy.Publish(OutcomeCommitted)
	if err != nil {
		t.Fatalf("Publish() through copy error = %v", err)
	}
	if published != OutcomeCommitted {
		t.Fatalf("Publish() through copy = %d, want OutcomeCommitted", published)
	}

	repeated, err := original.Publish(OutcomeNotCommitted)
	if err != nil {
		t.Fatalf("Publish() through original error = %v", err)
	}
	if repeated != OutcomeCommitted {
		t.Fatalf("Publish() through original = %d, want immutable OutcomeCommitted", repeated)
	}

	select {
	case outcome := <-waitResult:
		if outcome != OutcomeCommitted {
			t.Fatalf("Wait() through original = %d, want OutcomeCommitted", outcome)
		}
	case err := <-waitError:
		t.Fatalf("Wait() through original error = %v", err)
	case <-time.After(time.Second):
		t.Fatal("Wait() through original did not return after publication through copy")
	}
}

func TestBindingSingleSuccessfulPublication(t *testing.T) {
	binding := New()

	got, err := binding.Publish(OutcomeCommitted)
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if got != OutcomeCommitted {
		t.Fatalf("Publish() = %d, want OutcomeCommitted", got)
	}
}

func TestBindingRepeatedPublicationIsIdempotent(t *testing.T) {
	binding := New()

	first, err := binding.Publish(OutcomeCommitted)
	if err != nil {
		t.Fatalf("first Publish() error = %v", err)
	}
	second, err := binding.Publish(OutcomeNotCommitted)
	if err != nil {
		t.Fatalf("repeated Publish() error = %v", err)
	}
	if second != first {
		t.Fatalf("repeated Publish() = %d, want immutable outcome %d", second, first)
	}
}

func TestBindingCommittedOutcome(t *testing.T) {
	testPublishedOutcome(t, OutcomeCommitted)
}

func TestBindingNotCommittedOutcome(t *testing.T) {
	testPublishedOutcome(t, OutcomeNotCommitted)
}

func TestBindingPublicationIsImmutable(t *testing.T) {
	binding := New()

	if _, err := binding.Publish(OutcomeNotCommitted); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if _, err := binding.Publish(OutcomeCommitted); err != nil {
		t.Fatalf("repeated Publish() error = %v", err)
	}

	got, err := binding.Wait()
	if err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	if got != OutcomeNotCommitted {
		t.Fatalf("Wait() = %d, want first outcome OutcomeNotCommitted", got)
	}
}

func TestBindingConcurrentPublicationIsExactlyOnce(t *testing.T) {
	const publishers = 128

	binding := New()
	start := make(chan struct{})
	results := make(chan Outcome, publishers)
	errorsChannel := make(chan error, publishers)
	var waitGroup sync.WaitGroup
	for index := 0; index < publishers; index++ {
		outcome := OutcomeCommitted
		if index%2 == 1 {
			outcome = OutcomeNotCommitted
		}
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			got, err := binding.Publish(outcome)
			if err != nil {
				errorsChannel <- err
				return
			}
			results <- got
		}()
	}

	close(start)
	waitForGroup(t, &waitGroup, "concurrent publication")
	close(results)
	close(errorsChannel)

	for err := range errorsChannel {
		t.Errorf("Publish() error = %v", err)
	}

	var published Outcome
	for got := range results {
		if published == 0 {
			published = got
		}
		if got != published {
			t.Fatalf("concurrent Publish() returned outcomes %d and %d", published, got)
		}
	}
	if !published.valid() {
		t.Fatalf("published outcome = %d, want a supported outcome", published)
	}

	got, err := binding.Wait()
	if err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	if got != published {
		t.Fatalf("Wait() = %d, want published outcome %d", got, published)
	}
}

func TestBindingConcurrentWaitReleasesOneDormantPath(t *testing.T) {
	const waiters = 64

	binding := New()
	start := make(chan struct{})
	results := make(chan Outcome, waiters)
	errorsChannel := make(chan error, waiters)
	var entered sync.WaitGroup
	var finished sync.WaitGroup
	for range waiters {
		entered.Add(1)
		finished.Add(1)
		go func() {
			defer finished.Done()
			entered.Done()
			<-start
			outcome, err := binding.Wait()
			if err != nil {
				errorsChannel <- err
				return
			}
			results <- outcome
		}()
	}

	entered.Wait()
	close(start)
	if _, err := binding.Publish(OutcomeCommitted); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	waitForGroup(t, &finished, "concurrent wait")
	close(results)
	close(errorsChannel)

	successes := 0
	for outcome := range results {
		successes++
		if outcome != OutcomeCommitted {
			t.Errorf("Wait() = %d, want OutcomeCommitted", outcome)
		}
	}
	if successes != 1 {
		t.Fatalf("successful Wait() calls = %d, want 1", successes)
	}

	alreadyClaimed := 0
	for err := range errorsChannel {
		if !errors.Is(err, ErrWaitAlreadyClaimed) {
			t.Errorf("Wait() error = %v, want ErrWaitAlreadyClaimed", err)
			continue
		}
		alreadyClaimed++
	}
	if alreadyClaimed != waiters-1 {
		t.Fatalf("ErrWaitAlreadyClaimed count = %d, want %d", alreadyClaimed, waiters-1)
	}
}

func TestBindingTwoWaitsBeforePublication(t *testing.T) {
	binding := New()
	firstResult := make(chan Outcome, 1)
	firstError := make(chan error, 1)
	go func() {
		outcome, err := binding.Wait()
		if err != nil {
			firstError <- err
			return
		}
		firstResult <- outcome
	}()
	waitForClaim(t, binding.state)

	if _, err := binding.Wait(); !errors.Is(err, ErrWaitAlreadyClaimed) {
		t.Fatalf("second Wait() error = %v, want ErrWaitAlreadyClaimed", err)
	}
	select {
	case outcome := <-firstResult:
		t.Fatalf("first Wait() returned %d before publication", outcome)
	case err := <-firstError:
		t.Fatalf("first Wait() error before publication = %v", err)
	default:
	}

	if _, err := binding.Publish(OutcomeNotCommitted); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	select {
	case outcome := <-firstResult:
		if outcome != OutcomeNotCommitted {
			t.Fatalf("first Wait() = %d, want OutcomeNotCommitted", outcome)
		}
	case err := <-firstError:
		t.Fatalf("first Wait() error = %v", err)
	case <-time.After(time.Second):
		t.Fatal("first Wait() did not return after publication")
	}
}

func TestBindingTwoWaitsAfterPublication(t *testing.T) {
	binding := New()
	if _, err := binding.Publish(OutcomeCommitted); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	first, err := binding.Wait()
	if err != nil {
		t.Fatalf("first Wait() error = %v", err)
	}
	if first != OutcomeCommitted {
		t.Fatalf("first Wait() = %d, want OutcomeCommitted", first)
	}
	if _, err := binding.Wait(); !errors.Is(err, ErrWaitAlreadyClaimed) {
		t.Fatalf("second Wait() error = %v, want ErrWaitAlreadyClaimed", err)
	}
}

func TestBindingRepeatedWaitAfterSuccessfulReturn(t *testing.T) {
	binding := New()
	if _, err := binding.Publish(OutcomeNotCommitted); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if _, err := binding.Wait(); err != nil {
		t.Fatalf("first Wait() error = %v", err)
	}
	if _, err := binding.Wait(); !errors.Is(err, ErrWaitAlreadyClaimed) {
		t.Fatalf("repeated Wait() error = %v, want ErrWaitAlreadyClaimed", err)
	}
}

func TestBindingConcurrentInvalidAndValidPublication(t *testing.T) {
	binding := New()
	start := make(chan struct{})
	validResult := make(chan Outcome, 1)
	validError := make(chan error, 1)
	invalidError := make(chan error, 1)
	var ready sync.WaitGroup
	ready.Add(2)

	go func() {
		ready.Done()
		<-start
		outcome, err := binding.Publish(OutcomeCommitted)
		if err != nil {
			validError <- err
			return
		}
		validResult <- outcome
	}()
	go func() {
		ready.Done()
		<-start
		_, err := binding.Publish(Outcome(255))
		invalidError <- err
	}()

	ready.Wait()
	close(start)

	select {
	case outcome := <-validResult:
		if outcome != OutcomeCommitted {
			t.Fatalf("valid Publish() = %d, want OutcomeCommitted", outcome)
		}
	case err := <-validError:
		t.Fatalf("valid Publish() error = %v", err)
	case <-time.After(time.Second):
		t.Fatal("valid Publish() did not return")
	}
	select {
	case err := <-invalidError:
		if !errors.Is(err, ErrInvalidOutcome) {
			t.Fatalf("invalid Publish() error = %v, want ErrInvalidOutcome", err)
		}
	case <-time.After(time.Second):
		t.Fatal("invalid Publish() did not return")
	}

	outcome, err := binding.Wait()
	if err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	if outcome != OutcomeCommitted {
		t.Fatalf("Wait() = %d, want OutcomeCommitted", outcome)
	}
}

func TestBindingRejectsInvalidOutcome(t *testing.T) {
	binding := New()

	_, err := binding.Publish(Outcome(255))
	if !errors.Is(err, ErrInvalidOutcome) {
		t.Fatalf("Publish() error = %v, want ErrInvalidOutcome", err)
	}

	got, err := binding.Publish(OutcomeCommitted)
	if err != nil {
		t.Fatalf("valid Publish() after rejection error = %v", err)
	}
	if got != OutcomeCommitted {
		t.Fatalf("Publish() = %d, want OutcomeCommitted", got)
	}
}

func testPublishedOutcome(t *testing.T, want Outcome) {
	t.Helper()

	binding := New()
	result := make(chan Outcome, 1)
	errResult := make(chan error, 1)
	go func() {
		got, err := binding.Wait()
		if err != nil {
			errResult <- err
			return
		}
		result <- got
	}()

	select {
	case got := <-result:
		t.Fatalf("Wait() returned %d before publication", got)
	case err := <-errResult:
		t.Fatalf("Wait() error before publication = %v", err)
	default:
	}

	if _, err := binding.Publish(want); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	select {
	case got := <-result:
		if got != want {
			t.Fatalf("Wait() = %d, want %d", got, want)
		}
	case err := <-errResult:
		t.Fatalf("Wait() error = %v", err)
	case <-time.After(time.Second):
		t.Fatal("Wait() did not return after publication")
	}
}

func waitForGroup(t *testing.T, waitGroup *sync.WaitGroup, operation string) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		waitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for %s", operation)
	}
}

func waitForClaim(t *testing.T, state *bindingState) {
	t.Helper()

	deadline := time.After(time.Second)
	for !state.waitClaimed.Load() {
		select {
		case <-deadline:
			t.Fatal("timeout waiting for dormant path claim")
		default:
			runtime.Gosched()
		}
	}
}
