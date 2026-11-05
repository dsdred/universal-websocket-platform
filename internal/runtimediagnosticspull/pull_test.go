package runtimediagnosticspull

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	d "github.com/dsdred/universal-websocket-platform/internal/runtimediagnostics"
	c "github.com/dsdred/universal-websocket-platform/internal/runtimediagnosticsconsumer"
)

func offer(t *testing.T, producer *d.Producer, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		if !producer.Offer(d.ListenerServeFailed) {
			t.Fatalf("uncontended fixture offer %d failed", i)
		}
	}
}

func check(t *testing.T, got PollResult, want c.DrainResult) {
	t.Helper()
	if got.Receipt != want {
		t.Fatalf("receipt %+v, want %+v", got.Receipt, want)
	}
	r := got.Receipt
	if got.Count != r.SinkSucceeded || got.Count > 64 ||
		r.Observed != r.SinkSucceeded+r.SinkFailed+r.SinkPanicked+r.Discarded ||
		r.ReadAttempts < r.Observed || r.ReadAttempts > 64 {
		t.Fatalf("count/receipt equation violated: %+v", got)
	}
	for i := int(got.Count); i < len(got.Events); i++ {
		if got.Events[i] != (d.Event{}) {
			t.Fatalf("unused slot %d not zero", i)
		}
	}
}

func success(n uint8) c.DrainResult {
	return c.DrainResult{ReadAttempts: n, Observed: n, SinkSucceeded: n, Stop: c.Budget}
}

func clean(t *testing.T, view View) {
	t.Helper()
	if !view.state.guard.TryLock() {
		t.Fatal("outer guard retained after return")
	}
	defer view.state.guard.Unlock()
	if view.state.count != 0 || view.state.events != ([64]d.Event{}) {
		t.Fatal("private capture retained after return")
	}
}

// sampleContext forwards real cancellation and allows deterministic observations
// at C1's documented sample points. It never replaces Reader or the capture sink.
type sampleContext struct {
	context.Context
	samples  int
	onSample func(int)
}

func (ctx *sampleContext) Err() error {
	ctx.samples++
	if ctx.onSample != nil {
		ctx.onSample(ctx.samples)
	}
	return ctx.Context.Err()
}

func TestO01O02ConstructionAndReaderCapture(t *testing.T) {
	p, reader := d.New()
	p2, reader2 := d.New()
	offer(t, p, 2)
	offer(t, p2, 1)
	for _, view := range []View{{}, New(d.Reader{})} {
		ctx := &sampleContext{Context: context.Background(), onSample: func(int) { t.Fatal("disabled inspected context") }}
		check(t, view.Poll(ctx, 255), c.DrainResult{Stop: c.Disabled})
		check(t, view.Poll(nil, 0), c.DrainResult{Stop: c.Disabled})
	}
	captured := *reader
	view := New(captured)
	copyView := view
	captured = *reader2
	_ = captured
	check(t, view.Poll(context.Background(), 1), success(1))
	got := copyView.Poll(context.Background(), 1)
	check(t, got, success(1))
	if got.Events[0].Sequence != 2 {
		t.Fatal("construction read or View copy lost captured Reader")
	}
	if e, ok := reader2.TryRead(); !ok || e.Sequence != 1 {
		t.Fatal("captured Reader was redirected or other mailbox consumed")
	}
	clean(t, view)
}

func TestO03O04O08BusyInvalidPreserveAdmittedStorage(t *testing.T) {
	p, reader := d.New()
	offer(t, p, 1)
	view := New(*reader)
	copyView := view
	view.state.guard.Lock()
	view.state.count = 1
	view.state.events[0] = d.Event{SchemaVersion: 1, Code: d.ListenerServeFailed, Sequence: 99}
	wantStorage := view.state.events
	// Only new-package private state is held. All Poll calls must return without
	// entering C1; the outer guard protects the admitted turn's capture storage.
	var unlock sync.Once
	defer unlock.Do(view.state.guard.Unlock)
	for _, input := range []struct {
		ctx    context.Context
		budget uint8
	}{
		{nil, 1}, {context.Background(), 0}, {context.Background(), 65}, {context.Background(), 255},
	} {
		check(t, copyView.Poll(input.ctx, input.budget), c.DrainResult{Stop: c.InvalidInput})
	}
	ctx := &sampleContext{Context: context.Background(), onSample: func(int) { t.Error("Busy sampled context") }}
	done := make(chan struct{})
	var result PollResult
	go func() {
		defer close(done)
		result = copyView.Poll(ctx, 1)
	}()
	select {
	case <-done:
		check(t, result, c.DrainResult{Stop: c.Busy})
	case <-time.After(3 * time.Second):
		unlock.Do(view.state.guard.Unlock)
		<-done // Join fixture before failure; timeout is not a latency guarantee.
		t.Fatal("Busy waited for outer guard")
	}
	if view.state.events != wantStorage || view.state.count != 1 {
		t.Fatal("invalid/Busy touched admitted storage")
	}
	if e, ok := reader.TryRead(); !ok || e.Sequence != 1 {
		t.Fatal("invalid/Busy consumed event")
	}
	unlock.Do(view.state.guard.Unlock)
	// A real admitted turn must remove even deliberately seeded old storage.
	check(t, view.Poll(context.Background(), 1), c.DrainResult{ReadAttempts: 1, Stop: c.NoItem})
	clean(t, view)
}

func TestO04ReentrantContextSharesOuterGuard(t *testing.T) {
	p, reader := d.New()
	offer(t, p, 2)
	view := New(*reader)
	copyView := view
	ctx := &sampleContext{Context: context.Background(), onSample: func(int) {
		check(t, copyView.Poll(context.Background(), 1), c.DrainResult{Stop: c.Busy})
	}}
	check(t, view.Poll(ctx, 2), success(2))
	if ctx.samples != 6 {
		t.Fatalf("C1 sample count %d, want 6", ctx.samples)
	}
	clean(t, view)
}

func TestO05BoundedTurnDuringRefill(t *testing.T) {
	for _, budget := range []uint8{1, 64} {
		t.Run(fmt.Sprint(budget), func(t *testing.T) {
			p, reader := d.New()
			view := New(*reader)
			ctx := &sampleContext{Context: context.Background(), onSample: func(sample int) {
				if sample%3 == 1 {
					offer(t, p, 1)
				}
			}}
			got := view.Poll(ctx, budget)
			check(t, got, success(budget))
			if ctx.samples != 3*int(budget) {
				t.Fatal("turn exceeded C1 sampling budget")
			}
			for i := 0; i < int(got.Count); i++ {
				if got.Events[i].Sequence != uint64(i+1) {
					t.Fatal("refill FIFO changed")
				}
			}
			clean(t, view)
		})
	}
}

func TestO06O10CancellationSamples(t *testing.T) {
	for _, sample := range []int{1, 2, 3} {
		t.Run(fmt.Sprint(sample), func(t *testing.T) {
			p, reader := d.New()
			offer(t, p, 2)
			view := New(*reader)
			base, cancel := context.WithCancel(context.Background())
			defer cancel()
			ctx := &sampleContext{Context: base, onSample: func(n int) {
				if n == sample {
					cancel()
				}
			}}
			want := c.DrainResult{Stop: c.Canceled}
			if sample >= 2 {
				want.ReadAttempts, want.Observed = 1, 1
			}
			if sample == 2 {
				want.Discarded = 1
			}
			if sample == 3 {
				want.SinkSucceeded = 1
			}
			got := view.Poll(ctx, 64)
			check(t, got, want)
			if ctx.samples != sample {
				t.Fatal("cancellation sampled again")
			}
			if sample == 3 && got.Events[0].Sequence != 1 {
				t.Fatal("post-capture cancellation lost captured success")
			}
			select {
			case <-ctx.Done():
			default:
				t.Fatal("context cancellation not forwarded")
			}
			clean(t, view)
			next := view.Poll(context.Background(), 1)
			check(t, next, success(1))
			wantSequence := uint64(2)
			if sample == 1 {
				wantSequence = 1
			}
			if next.Events[0].Sequence != wantSequence {
				t.Fatal("discard reinserted or following event removed")
			}
		})
	}
}

func TestO07O08DetachedFIFOAndClearing(t *testing.T) {
	p, reader := d.New()
	view := New(*reader)
	// Initial clearing is inspected at the first forwarded Context sample while
	// this same goroutine holds the admitted outer guard.
	view.state.count = 64
	for i := range view.state.events {
		view.state.events[i] = d.Event{Sequence: 99}
	}
	ctx := &sampleContext{Context: context.Background(), onSample: func(n int) {
		if n == 1 && (view.state.count != 0 || view.state.events != ([64]d.Event{})) {
			t.Fatal("storage not cleared before Drain")
		}
	}}
	offer(t, p, 64)
	batch := view.Poll(ctx, 64)
	check(t, batch, success(64))
	for i := range batch.Events {
		if batch.Events[i].Sequence != uint64(i+1) {
			t.Fatal("batch FIFO changed")
		}
	}
	clean(t, view)
	batch.Events[0] = d.Event{Sequence: 999}
	batch.Count = 0
	batch.Receipt = c.DrainResult{}
	offer(t, p, 1)
	got := view.Poll(context.Background(), 1)
	check(t, got, success(1))
	if got.Events[0].Sequence != 65 {
		t.Fatal("returned mutation affected later Poll")
	}
	clean(t, view)
	check(t, view.Poll(context.Background(), 64), c.DrainResult{ReadAttempts: 1, Stop: c.NoItem})
	if batch.Events[0].Sequence != 999 {
		t.Fatal("later Poll mutated returned value")
	}
}

func TestO09NoItemIsLocalOneAttempt(t *testing.T) {
	_, reader := d.New()
	view := New(*reader)
	check(t, view.Poll(context.Background(), 64), c.DrainResult{ReadAttempts: 1, Stop: c.NoItem})
	// Busy/invalid false cases are COMPOSITIONAL with unchanged mailbox white-box
	// tests and C1 false handling. No Reader implementation is forged here.
	clean(t, view)
}

func TestO11PresentationDoesNotOwnPollResources(t *testing.T) {
	p, reader := d.New()
	offer(t, p, 1)
	view := New(*reader)
	batch := view.Poll(context.Background(), 1)
	check(t, batch, success(1))
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var preserved bool
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }); <-done })
	go func() {
		defer close(done)
		close(entered)
		<-release // Caller-owned failed presentation; no retry or acknowledgment.
		preserved = batch.Events[0].Sequence == 1
	}()
	<-entered
	offer(t, p, 3)
	if e, ok := reader.TryRead(); !ok || e.Sequence != 2 {
		t.Fatal("presentation held mailbox")
	}
	got := view.Poll(context.Background(), 1)
	check(t, got, success(1))
	if got.Events[0].Sequence != 3 {
		t.Fatal("presentation held View guard")
	}
	clean(t, view)
	once.Do(func() { close(release) })
	<-done
	if !preserved {
		t.Fatal("later work mutated presentation batch")
	}
}

func TestO12SeparateViewsPartitionWithoutGlobalGuard(t *testing.T) {
	p, reader := d.New()
	p2, reader2 := d.New()
	offer(t, p, 4)
	offer(t, p2, 1)
	view, separate, other := New(*reader), New(*reader), New(*reader2)
	view.state.guard.Lock()
	got := separate.Poll(context.Background(), 1)
	view.state.guard.Unlock()
	check(t, got, success(1))
	if got.Events[0].Sequence != 1 {
		t.Fatal("separate View shared global guard")
	}
	if e, ok := reader.TryRead(); !ok || e.Sequence != 2 {
		t.Fatal("extra Reader did not partition")
	}
	got = view.Poll(context.Background(), 2)
	check(t, got, success(2))
	if got.Events[0].Sequence != 3 || got.Events[1].Sequence != 4 {
		t.Fatal("partitioned records replayed or reordered")
	}
	got = other.Poll(context.Background(), 1)
	check(t, got, success(1))
	if got.Events[0].Sequence != 1 {
		t.Fatal("distinct mailbox not isolated")
	}
}

func TestO14SafeFixedResultSchema(t *testing.T) {
	typ := reflect.TypeOf(PollResult{})
	want := []struct {
		name string
		typ  reflect.Type
	}{
		{"Receipt", reflect.TypeOf(c.DrainResult{})},
		{"Count", reflect.TypeOf(uint8(0))},
		{"Events", reflect.TypeOf([64]d.Event{})},
	}
	if typ.NumField() != len(want) {
		t.Fatal("result surface expanded")
	}
	for i, expected := range want {
		field := typ.Field(i)
		if field.Name != expected.name || field.Type != expected.typ {
			t.Fatal("result retained unapproved state")
		}
	}
	// Saturation loses offers without consuming sequence; no gaps is not proof
	// of no loss, and no cumulative/drop metric is returned by Poll.
	p, reader := d.New()
	offer(t, p, 64)
	if p.Offer(d.AuthenticationProviderFailed) {
		t.Fatal("capacity expanded")
	}
	view := New(*reader)
	check(t, view.Poll(context.Background(), 64), success(64))
	offer(t, p, 1)
	got := view.Poll(context.Background(), 1)
	check(t, got, success(1))
	if got.Events[0].Sequence != 65 {
		t.Fatal("lost offer consumed sequence")
	}
}
