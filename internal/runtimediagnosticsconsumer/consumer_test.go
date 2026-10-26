package runtimediagnosticsconsumer_test

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	d "github.com/dsdred/universal-websocket-platform/internal/runtimediagnostics"
	c "github.com/dsdred/universal-websocket-platform/internal/runtimediagnosticsconsumer"
)

func offer(t *testing.T, p *d.Producer, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		if !p.Offer(d.ListenerServeFailed) {
			t.Fatalf("offer %d lost in uncontended fixture", i)
		}
	}
}

func receipt(t *testing.T, got, want c.DrainResult) {
	t.Helper()
	if got != want {
		t.Fatalf("receipt %+v, want %+v", got, want)
	}
	if got.Observed != got.SinkSucceeded+got.SinkFailed+got.SinkPanicked+got.Discarded ||
		got.ReadAttempts < got.Observed || got.ReadAttempts > 64 {
		t.Fatalf("receipt equation/bound violated: %+v", got)
	}
}

func TestC01InertDisabledConstruction(t *testing.T) {
	p, r := d.New()
	offer(t, p, 1)
	calls := 0
	sink := func(d.Event) error { calls++; return nil }
	for _, consumer := range []c.Consumer{{}, c.NewConsumer(d.Reader{}, sink), c.NewConsumer(*r, nil)} {
		receipt(t, consumer.Drain(nil, 0), c.DrainResult{Stop: c.Disabled})
	}
	consumer := c.NewConsumer(*r, sink)
	if calls != 0 {
		t.Fatal("construction invoked sink")
	}
	receipt(t, consumer.Drain(context.Background(), 1), c.DrainResult{ReadAttempts: 1, Observed: 1, SinkSucceeded: 1, Stop: c.Budget})
	if calls != 1 {
		t.Fatal("construction removed event")
	}
}

func TestC02C07CaptureCopiesFIFOAndPartition(t *testing.T) {
	p, r := d.New()
	p2, r2 := d.New()
	reader := *r
	var seen []uint64
	sink := c.Sink(func(e d.Event) error {
		seen = append(seen, e.Sequence)
		e.Code, e.SchemaVersion, e.Sequence = 255, 255, 0
		return nil
	})
	consumer := c.NewConsumer(reader, sink)
	copyConsumer := consumer
	reader = *r2
	sink = func(d.Event) error { t.Fatal("reassigned sink invoked"); return nil }
	_ = reader
	_ = sink
	offer(t, p, 4)
	offer(t, p2, 1)
	receipt(t, consumer.Drain(context.Background(), 1), c.DrainResult{ReadAttempts: 1, Observed: 1, SinkSucceeded: 1, Stop: c.Budget})
	// An extra Reader partitions, never fans out; no global sink order is claimed.
	if e, ok := r.TryRead(); !ok || e.Sequence != 2 {
		t.Fatal("extra reader did not partition sequence2")
	}
	receipt(t, copyConsumer.Drain(context.Background(), 4), c.DrainResult{ReadAttempts: 3, Observed: 2, SinkSucceeded: 2, Stop: c.NoItem})
	if !reflect.DeepEqual(seen, []uint64{1, 3, 4}) {
		t.Fatalf("FIFO/partition = %v", seen)
	}
	if e, ok := r2.TryRead(); !ok || e.Sequence != 1 || e.Code != d.ListenerServeFailed {
		t.Fatal("independent mailbox or detached Event mutated")
	}
}

func TestC03InputPrecedenceConsumesNothing(t *testing.T) {
	p, r := d.New()
	offer(t, p, 1)
	consumer := c.NewConsumer(*r, func(d.Event) error { t.Fatal("invalid call reached sink"); return nil })
	for _, input := range []struct {
		ctx    context.Context
		budget uint8
	}{{nil, 1}, {context.Background(), 0}, {context.Background(), 65}, {context.Background(), 255}} {
		receipt(t, consumer.Drain(input.ctx, input.budget), c.DrainResult{Stop: c.InvalidInput})
	}
	if e, ok := r.TryRead(); !ok || e.Sequence != 1 {
		t.Fatal("invalid input consumed event")
	}
}

func TestC04C11BlockedSinkSharedGuardAndIndependentProgress(t *testing.T) {
	p, r := d.New()
	offer(t, p, 1)
	entered, release := make(chan struct{}), make(chan struct{})
	done := make(chan c.DrainResult, 1)
	var closeOnce sync.Once
	t.Cleanup(func() {
		closeOnce.Do(func() { close(release) })
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("fixture failed to join after sink release")
		}
	})
	consumer := c.NewConsumer(*r, func(d.Event) error { close(entered); <-release; return nil })
	copyConsumer := consumer
	go func() { done <- consumer.Drain(context.Background(), 1) }()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("sink did not enter")
	}
	// Validation precedes Busy, which precedes cancellation sampling.
	receipt(t, copyConsumer.Drain(nil, 0), c.DrainResult{Stop: c.InvalidInput})
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	receipt(t, copyConsumer.Drain(canceled, 1), c.DrainResult{Stop: c.Busy})
	offer(t, p, 64)
	if p.Offer(d.ListenerServeFailed) {
		t.Fatal("mailbox capacity expanded")
	}
	independent := c.NewConsumer(*r, func(d.Event) error { return nil })
	receipt(t, independent.Drain(context.Background(), 1), c.DrainResult{ReadAttempts: 1, Observed: 1, SinkSucceeded: 1, Stop: c.Budget})
	if e, ok := r.TryRead(); !ok || e.Sequence != 3 {
		t.Fatal("blocked sink held mailbox lock")
	}
	closeOnce.Do(func() { close(release) })
	select {
	case result := <-done:
		receipt(t, result, c.DrainResult{ReadAttempts: 1, Observed: 1, SinkSucceeded: 1, Stop: c.Budget})
		done <- result // Cleanup joins the already completed fixture too.
	case <-time.After(3 * time.Second):
		t.Fatal("sink did not return after release")
	}
}

func TestC04ReentrantReturnsBusy(t *testing.T) {
	p, r := d.New()
	offer(t, p, 2)
	var consumer c.Consumer
	consumer = c.NewConsumer(*r, func(d.Event) error {
		receipt(t, consumer.Drain(context.Background(), 64), c.DrainResult{Stop: c.Busy})
		return nil
	})
	receipt(t, consumer.Drain(context.Background(), 2), c.DrainResult{ReadAttempts: 2, Observed: 2, SinkSucceeded: 2, Stop: c.Budget})
}

func TestC05ReadAttemptBudgetUnderContinuousRefill(t *testing.T) {
	for _, budget := range []uint8{1, 64} {
		t.Run(fmt.Sprint(budget), func(t *testing.T) {
			p, r := d.New()
			offer(t, p, 1)
			consumer := c.NewConsumer(*r, func(d.Event) error { offer(t, p, 1); return nil })
			receipt(t, consumer.Drain(context.Background(), budget), c.DrainResult{ReadAttempts: budget, Observed: budget, SinkSucceeded: budget, Stop: c.Budget})
			if e, ok := r.TryRead(); !ok || e.Sequence != uint64(budget)+1 {
				t.Fatal("budget failed to leave refilled item")
			}
		})
	}
}

func TestC06FalseIsOneAttemptNoItem(t *testing.T) {
	_, r := d.New()
	consumer := c.NewConsumer(*r, func(d.Event) error { t.Fatal("false read reached sink"); return nil })
	receipt(t, consumer.Drain(context.Background(), 64), c.DrainResult{ReadAttempts: 1, Stop: c.NoItem})
	// Busy/invalid-record false are compositional with the mailbox white-box tests;
	// the concrete Reader is intentionally neither forged nor replaced with a fake.
}

type cancelOnSample struct {
	context.Context
	cancel  context.CancelFunc
	at      int32
	samples atomic.Int32
}

func (ctx *cancelOnSample) Err() error {
	if ctx.samples.Add(1) == ctx.at {
		ctx.cancel()
	}
	return ctx.Context.Err()
}

func TestC08CancellationAroundRead(t *testing.T) {
	for _, sample := range []int{1, 2} {
		t.Run(fmt.Sprint(sample), func(t *testing.T) {
			p, r := d.New()
			offer(t, p, 2)
			base, cancel := context.WithCancel(context.Background())
			defer cancel()
			ctx := &cancelOnSample{Context: base, cancel: cancel, at: int32(sample)}
			consumer := c.NewConsumer(*r, func(d.Event) error { t.Fatal("canceled before sink"); return nil })
			want := c.DrainResult{Stop: c.Canceled}
			if sample == 2 {
				want.ReadAttempts, want.Observed, want.Discarded = 1, 1, 1
			}
			receipt(t, consumer.Drain(ctx, 64), want)
			select {
			case <-ctx.Done():
			default:
				t.Fatal("forwarding context violated cancellation contract")
			}
			if e, ok := r.TryRead(); !ok || e.Sequence != uint64(sample) {
				t.Fatal("cancellation read/discard boundary incorrect")
			}
		})
	}
}

type hostilePayload struct{}

func (hostilePayload) Error() string          { panic("Error inspected") }
func (hostilePayload) String() string         { panic("String inspected") }
func (hostilePayload) Unwrap() error          { panic("Unwrap inspected") }
func (hostilePayload) Is(error) bool          { panic("Is inspected") }
func (hostilePayload) As(any) bool            { panic("As inspected") }
func (hostilePayload) Format(fmt.State, rune) { panic("Format inspected") }

func TestC09C10SinkOutcomeWinsLateCancellationWithoutPayloadInspection(t *testing.T) {
	for _, mode := range []string{"nil", "error", "panic", "panic-nil"} {
		t.Run(mode, func(t *testing.T) {
			p, r := d.New()
			offer(t, p, 2)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			consumer := c.NewConsumer(*r, func(d.Event) error {
				calls++
				cancel() // Already entered: never retract or preempt this invocation.
				switch mode {
				case "error":
					return hostilePayload{}
				case "panic":
					panic(hostilePayload{})
				case "panic-nil":
					panic(nil)
				default:
					return nil
				}
			})
			want := c.DrainResult{ReadAttempts: 1, Observed: 1}
			switch mode {
			case "nil":
				want.SinkSucceeded, want.Stop = 1, c.Canceled
			case "error":
				want.SinkFailed, want.Stop = 1, c.SinkFailed
			default:
				want.SinkPanicked, want.Stop = 1, c.SinkPanicked
			}
			receipt(t, consumer.Drain(ctx, 64), want)
			if calls != 1 {
				t.Fatal("sink retried")
			}
			if e, ok := r.TryRead(); !ok || e.Sequence != 2 {
				t.Fatal("failed event reinserted or following event consumed")
			}
			// Every failure/cancellation returning path must release the shared guard.
			receipt(t, consumer.Drain(context.Background(), 1), c.DrainResult{ReadAttempts: 1, Stop: c.NoItem})
		})
	}
}

func TestC10LegacyNilPanicStillClassified(t *testing.T) {
	t.Setenv("GODEBUG", "panicnil=1")
	p, r := d.New()
	offer(t, p, 1)
	consumer := c.NewConsumer(*r, func(d.Event) error { panic(nil) })
	receipt(t, consumer.Drain(context.Background(), 1), c.DrainResult{ReadAttempts: 1, Observed: 1, SinkPanicked: 1, Stop: c.SinkPanicked})
}

func TestC12C13SaturationLossLocalReceiptAndRetainedCapability(t *testing.T) {
	p, r := d.New()
	var sequences []uint64
	consumer := c.NewConsumer(*r, func(e d.Event) error { sequences = append(sequences, e.Sequence); return nil })
	offer(t, p, 64)
	for i := 0; i < 20; i++ {
		if p.Offer(d.AuthenticationProviderFailed) {
			t.Fatal("abandoned consumer expanded bound")
		}
	}
	receipt(t, consumer.Drain(context.Background(), 64), c.DrainResult{ReadAttempts: 64, Observed: 64, SinkSucceeded: 64, Stop: c.Budget})
	for i, sequence := range sequences {
		if sequence != uint64(i+1) {
			t.Fatal("full loss changed FIFO")
		}
	}
	// Drops did not consume sequence numbers; absence of gaps is not loss evidence.
	offer(t, p, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	receipt(t, consumer.Drain(ctx, 64), c.DrainResult{Stop: c.Canceled})
	copyConsumer := consumer
	receipt(t, copyConsumer.Drain(context.Background(), 1), c.DrainResult{ReadAttempts: 1, Observed: 1, SinkSucceeded: 1, Stop: c.Budget})
	if sequences[len(sequences)-1] != 65 {
		t.Fatal("retained capability reset sequence")
	}
	receipt(t, copyConsumer.Drain(context.Background(), 64), c.DrainResult{ReadAttempts: 1, Stop: c.NoItem})
}

func TestC13C16ReceiptHasOnlyBoundedScalarFields(t *testing.T) {
	typ := reflect.TypeOf(c.DrainResult{})
	want := []string{"ReadAttempts", "Observed", "SinkSucceeded", "SinkFailed", "SinkPanicked", "Discarded", "Stop"}
	if typ.NumField() != len(want) {
		t.Fatal("receipt schema expanded")
	}
	for i, name := range want {
		field := typ.Field(i)
		if field.Name != name || field.Type.Kind() != reflect.Uint8 {
			t.Fatal("receipt retains non-scalar state")
		}
	}
}
