package runtimediagnostics_test

import (
	"errors"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnostics"
)

func TestCopiesAndIndependentMailboxes(t *testing.T) {
	p, r := runtimediagnostics.New()
	p2, r2 := runtimediagnostics.New()
	copyProducer, copyReader := *p, *r
	if !copyProducer.Offer(runtimediagnostics.ListenerServeFailed) {
		t.Fatal("copied producer lost original binding")
	}
	e, ok := copyReader.TryRead()
	if !ok || e.Sequence != 1 {
		t.Fatalf("copied reader = %+v, %v", e, ok)
	}
	e.SchemaVersion, e.Code, e.Sequence = 255, 255, 999
	if !p.Offer(runtimediagnostics.AuthenticationProviderFailed) {
		t.Fatal("second offer lost")
	}
	next, ok := r.TryRead()
	if !ok || next.SchemaVersion != 1 || next.Code != runtimediagnostics.AuthenticationProviderFailed || next.Sequence != 2 {
		t.Fatalf("caller mutation reached mailbox: %+v, %v", next, ok)
	}
	if item, ok := r2.TryRead(); ok || item != (runtimediagnostics.Event{}) {
		t.Fatal("cross-mailbox event")
	}
	if !p2.Offer(runtimediagnostics.ListenerServeFailed) {
		t.Fatal("independent offer failed")
	}
	if item, ok := r2.TryRead(); !ok || item.Sequence != 1 {
		t.Fatal("independent sequence not local")
	}
}

func TestBlockedConsumerCannotHoldMailbox(t *testing.T) {
	p, r := runtimediagnostics.New()
	p.Offer(runtimediagnostics.ListenerServeFailed)
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	defer close(release)
	go func() {
		defer close(done)
		if _, ok := r.TryRead(); !ok {
			t.Error("consumer could not remove initial event")
			return
		}
		close(entered)
		<-release // Arbitrary blocked external sink, after removal.
	}()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("consumer did not reach sink")
	}
	// A second reader and producer still progress while the sink is blocked.
	for i := 0; i < 64; i++ {
		if !p.Offer(runtimediagnostics.AuthenticationProviderFailed) {
			t.Fatalf("blocked consumer affected producer at slot %d", i)
		}
	}
	if p.Offer(runtimediagnostics.ListenerServeFailed) {
		t.Fatal("unread events exceeded bound")
	}
	for i := 0; i < 64; i++ {
		if e, ok := r.TryRead(); !ok || e.Sequence != uint64(i+2) {
			t.Fatalf("independent reader = %+v, %v", e, ok)
		}
	}
	select {
	case <-done:
		t.Fatal("sink did not stay blocked")
	default:
	}
	// Joining the test's goroutine is test ownership, not a mailbox lifecycle hook.
	t.Cleanup(func() {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("test consumer did not return after release")
		}
	})
}

func TestConsumerFailureDoesNotReinsert(t *testing.T) {
	for _, sink := range []func(runtimediagnostics.Event) error{
		func(runtimediagnostics.Event) error { return errors.New("external sink failure") },
		func(runtimediagnostics.Event) error { panic("external sink panic") },
	} {
		p, r := runtimediagnostics.New()
		p.Offer(runtimediagnostics.ListenerServeFailed)
		e, ok := r.TryRead()
		if !ok {
			t.Fatal("missing event")
		}
		func() {
			defer func() { _ = recover() }()
			_ = sink(e)
		}()
		if item, ok := r.TryRead(); ok || item != (runtimediagnostics.Event{}) {
			t.Fatal("sink failure reinserted event")
		}
		if !p.Offer(runtimediagnostics.AuthenticationProviderFailed) {
			t.Fatal("sink failure held mailbox")
		}
		if item, ok := r.TryRead(); !ok || item.Sequence != 2 {
			t.Fatal("sink failure reset sequence")
		}
	}
}

func TestReaderOutlivesProducerWithoutLifecycle(t *testing.T) {
	makeReader := func() *runtimediagnostics.Reader {
		p, r := runtimediagnostics.New()
		p.Offer(runtimediagnostics.ListenerServeFailed)
		return r // No Start, Stop, context, Close, flush, join or worker needed.
	}
	r := makeReader()
	if e, ok := r.TryRead(); !ok || e.Sequence != 1 {
		t.Fatalf("retained reader = %+v, %v", e, ok)
	}
	if _, ok := r.TryRead(); ok {
		t.Fatal("mailbox replayed event")
	}
	// Unused mailboxes can simply be discarded with queued records.
	p, _ := runtimediagnostics.New()
	p.Offer(runtimediagnostics.ListenerServeFailed)
}
