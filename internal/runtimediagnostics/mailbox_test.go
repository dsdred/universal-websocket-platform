package runtimediagnostics

import (
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestClosedSchema(t *testing.T) {
	typ := reflect.TypeFor[Event]()
	fields := []struct {
		name string
		typ  reflect.Type
	}{{"SchemaVersion", reflect.TypeFor[uint8]()}, {"Code", reflect.TypeFor[Code]()}, {"Sequence", reflect.TypeFor[uint64]()}}
	if typ.NumField() != len(fields) {
		t.Fatalf("event fields = %d, want exactly three scalars", typ.NumField())
	}
	for i, want := range fields {
		got := typ.Field(i)
		if got.Name != want.name || got.Type != want.typ {
			t.Fatalf("field %d = %v, want %v", i, got, want)
		}
	}
	p, r := New()
	accepted := 0
	for value := 0; value < 256; value++ {
		code := Code(value)
		want := code == ListenerServeFailed || code == AuthenticationProviderFailed
		if got := p.Offer(code); got != want {
			t.Fatalf("Offer(%d) = %v, want %v", value, got, want)
		}
		if want {
			accepted++
			e, ok := r.TryRead()
			if !ok || e.SchemaVersion != 1 || e.Code != code || e.Sequence != uint64(accepted) {
				t.Fatalf("closed event = %+v, %v", e, ok)
			}
		}
	}
	if accepted != 2 || p.mailbox.sequence != 2 {
		t.Fatalf("accepted = %d, sequence = %d", accepted, p.mailbox.sequence)
	}
}

func TestDisabledCapabilities(t *testing.T) {
	for _, p := range []*Producer{nil, {}} {
		if p.Offer(ListenerServeFailed) {
			t.Fatal("disabled producer accepted event")
		}
	}
	for _, r := range []*Reader{nil, {}} {
		if e, ok := r.TryRead(); ok || e != (Event{}) {
			t.Fatalf("disabled reader = %+v, %v", e, ok)
		}
	}
}

func TestCapacityFIFOAndRingReuse(t *testing.T) {
	p, r := New()
	if e, ok := r.TryRead(); ok || e != (Event{}) {
		t.Fatal("new mailbox not empty")
	}
	// Repeat beyond one ring wrap; saturation must never replace an older event.
	for round := 0; round < 5; round++ {
		for i := 0; i < 64; i++ {
			code := ListenerServeFailed
			if i%2 != 0 {
				code = AuthenticationProviderFailed
			}
			if !p.Offer(code) {
				t.Fatalf("rejected slot %d in round %d", i, round)
			}
		}
		for i := 0; i < 10; i++ {
			if p.Offer(ListenerServeFailed) {
				t.Fatal("accepted beyond fixed64")
			}
		}
		for i := 0; i < 64; i++ {
			wantCode := ListenerServeFailed
			if i%2 != 0 {
				wantCode = AuthenticationProviderFailed
			}
			e, ok := r.TryRead()
			if !ok || e.Sequence != uint64(round*64+i+1) || e.Code != wantCode {
				t.Fatalf("FIFO event = %+v, %v at round %d offset %d", e, ok, round, i)
			}
		}
		if e, ok := r.TryRead(); ok || e != (Event{}) {
			t.Fatal("extra item after drain")
		}
		if p.mailbox.slots != ([64]Event{}) {
			t.Fatal("removed slots retain data")
		}
	}
	// Advance head partially, then fill across the array boundary while occupied.
	for i := 0; i < 40; i++ {
		p.Offer(ListenerServeFailed)
	}
	for i := 0; i < 31; i++ {
		r.TryRead()
	}
	for i := 0; i < 55; i++ {
		if !p.Offer(AuthenticationProviderFailed) {
			t.Fatal("wrapped ring rejected available slot")
		}
	}
	for seq := uint64(352); seq <= 415; seq++ {
		e, ok := r.TryRead()
		if !ok || e.Sequence != seq {
			t.Fatalf("partial-wrap FIFO = %+v, %v, want %d", e, ok, seq)
		}
	}
}

func TestBusyOperationsLoseWithoutWaiting(t *testing.T) {
	p, r := New()
	p.Offer(ListenerServeFailed)
	m := p.mailbox
	m.mu.Lock()
	defer m.mu.Unlock()
	offered := make(chan bool, 1)
	read := make(chan bool, 1)
	go func() { offered <- p.Offer(AuthenticationProviderFailed) }()
	go func() {
		e, ok := r.TryRead()
		read <- ok || e != (Event{})
	}()
	for _, result := range []<-chan bool{offered, read} {
		select {
		case success := <-result:
			if success {
				t.Fatal("busy operation returned success/data")
			}
		case <-time.After(2 * time.Second):
			t.Fatal("busy operation waited for lock")
		}
	}
	if m.count != 1 || m.sequence != 1 || m.slots[0].Code != ListenerServeFailed {
		t.Fatal("busy loss changed existing queue/sequence")
	}
}

func TestSequenceExhaustionIsPermanent(t *testing.T) {
	p, r := New()
	p.mailbox.sequence = ^uint64(0) - 1
	if p.Offer(Code(0)) || !p.Offer(ListenerServeFailed) {
		t.Fatal("invalid input consumed last sequence or final insertion rejected")
	}
	if p.Offer(AuthenticationProviderFailed) {
		t.Fatal("sequence wrapped before read")
	}
	e, ok := r.TryRead()
	if !ok || e.Sequence != ^uint64(0) {
		t.Fatalf("last event = %+v, %v", e, ok)
	}
	for i := 0; i < 100; i++ {
		if p.Offer(ListenerServeFailed) {
			t.Fatal("draining reset exhausted sequence")
		}
	}
	if p.mailbox.sequence != ^uint64(0) || p.mailbox.count != 0 {
		t.Fatal("exhaustion state changed")
	}
}

func TestInvalidRecordRemovedWithoutFallbackOrScan(t *testing.T) {
	for _, bad := range []Event{
		{}, {SchemaVersion: 0, Code: ListenerServeFailed, Sequence: 1},
		{SchemaVersion: 2, Code: ListenerServeFailed, Sequence: 1},
		{SchemaVersion: 1, Code: Code(255), Sequence: 1},
		{SchemaVersion: 1, Code: ListenerServeFailed, Sequence: 0},
	} {
		p, r := New()
		p.Offer(ListenerServeFailed)
		p.Offer(AuthenticationProviderFailed)
		p.mailbox.slots[0] = bad // White-box fault injection, no public Event ingress.
		if e, ok := r.TryRead(); ok || e != (Event{}) {
			t.Fatalf("invalid event escaped: %+v, %v", e, ok)
		}
		if p.mailbox.count != 1 || p.mailbox.slots[0] != (Event{}) {
			t.Fatal("invalid removal did not clear exactly one slot")
		}
		if e, ok := r.TryRead(); !ok || e.Code != AuthenticationProviderFailed || e.Sequence != 2 {
			t.Fatalf("next valid event changed: %+v, %v", e, ok)
		}
	}
}

func TestOperationsDoNotAllocate(t *testing.T) {
	p, r := New()
	allocs := testing.AllocsPerRun(1000, func() {
		if !p.Offer(ListenerServeFailed) {
			panic("unexpected loss without contention")
		}
		if _, ok := r.TryRead(); !ok {
			panic("missing offered event")
		}
	})
	if allocs != 0 {
		t.Fatalf("operation allocations = %v, want zero", allocs)
	}
}

func TestConcurrentProducersReaders(t *testing.T) {
	for _, readers := range []int{1, 4} {
		t.Run(string(rune('0'+readers))+" readers", func(t *testing.T) {
			const producers, attempts = 8, 2000
			p, r := New()
			// Guarantee at least one record even if contention loses every later offer.
			p.Offer(ListenerServeFailed)
			var accepted atomic.Uint64
			accepted.Store(1)
			var finished atomic.Bool
			var writers, consumers sync.WaitGroup
			results := make(chan Event, producers*attempts+1)
			start := make(chan struct{})
			for i := 0; i < readers; i++ {
				consumers.Go(func() {
					<-start
					var last uint64
					for {
						e, ok := r.TryRead()
						if ok {
							if e.Sequence <= last {
								t.Error("one reader observed non-increasing FIFO sequence")
							}
							last = e.Sequence
							results <- e
						} else if finished.Load() {
							return
						}
					}
				})
			}
			for i := 0; i < producers; i++ {
				writers.Go(func() {
					<-start
					for n := 0; n < attempts; n++ {
						code := ListenerServeFailed
						if n%2 != 0 {
							code = AuthenticationProviderFailed
						}
						if p.Offer(code) {
							accepted.Add(1)
						}
					}
				})
			}
			close(start)
			writers.Wait()
			finished.Store(true)
			consumers.Wait()
			// A busy reader may have stopped after producer completion: drain leftovers.
			for {
				e, ok := r.TryRead()
				if !ok {
					break
				}
				results <- e
			}
			close(results)
			var seqs []uint64
			for e := range results {
				if !e.valid() {
					t.Errorf("invalid concurrent event: %+v", e)
				}
				seqs = append(seqs, e.Sequence)
			}
			if uint64(len(seqs)) != accepted.Load() {
				t.Fatalf("removed %d, accepted %d", len(seqs), accepted.Load())
			}
			sort.Slice(seqs, func(i, j int) bool { return seqs[i] < seqs[j] })
			for i, seq := range seqs {
				if seq != uint64(i+1) {
					t.Fatalf("duplicate/missing sequence at %d: %d", i, seq)
				}
			}
			if p.mailbox.count != 0 || p.mailbox.slots != ([64]Event{}) {
				t.Fatal("concurrent removal retained records")
			}
		})
	}
}
