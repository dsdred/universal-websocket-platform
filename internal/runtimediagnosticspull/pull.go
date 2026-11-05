// Package runtimediagnosticspull provides caller-scheduled, bounded, lossy
// diagnostic batches. It owns no Runtime lifecycle, workers, timers or I/O.
// Poll must not run inside Runtime callbacks or locks. Presentation and retention
// of returned values belong to the caller and are never joined by Runtime.
package runtimediagnosticspull

import (
	"context"
	"sync"

	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnostics"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnosticsconsumer"
)

// PollResult is a detached fixed64 batch and its local C1 receipt.
// Count equals Receipt.SinkSucceeded and is at most 64; unused Events are zero.
// Pre-sink discarded events are absent. Capture success is not acknowledgment,
// external observation or persistence. Producer drops, physical removals and
// remaining queue length are unknown; Sequence cannot measure loss or identify
// a Host. NoItem does not prove empty, drained, flushed or complete.
type PollResult struct {
	Receipt runtimediagnosticsconsumer.DrainResult
	Count   uint8
	Events  [64]runtimediagnostics.Event
}

// View is an opaque copyable capability. Copies share private capture state and
// an outer guard; separately constructed Views do not share that guard. One
// logical Reader owner per mailbox is required for FIFO consumption: extra
// readers partition records destructively, never replay or fan out. Zero View
// is disabled. View provides no lifecycle, flush, reset or shutdown join.
type View struct {
	state *viewState
}

type viewState struct {
	guard    sync.Mutex
	consumer runtimediagnosticsconsumer.Consumer
	events   [64]runtimediagnostics.Event
	count    uint8
}

// New inertly captures reader by value with a fixed private capture sink.
// A zero Reader returns a disabled View. Construction performs no read, offer,
// caller callback, I/O, worker, timer or lifecycle action.
func New(reader runtimediagnostics.Reader) View {
	if reader == (runtimediagnostics.Reader{}) {
		return View{}
	}
	s := &viewState{}
	s.consumer = runtimediagnosticsconsumer.NewConsumer(reader, func(event runtimediagnostics.Event) error {
		// The single bounded Drain permits at most 64 successful captures.
		s.events[s.count] = event
		s.count++
		return nil
	})
	return View{state: s}
}

// Poll makes one outer TryLock attempt and, if admitted, exactly one C1 Drain
// with budget 1..64. Disabled precedes InvalidInput, which precedes Busy; rejected
// turns do not sample context, read events or touch an admitted turn's storage.
// The guard covers clearing, Drain, detached copying and final clearing.
// Returning/unwinding paths clear private storage before releasing the guard.
// Context must be well-formed, prompt and nonpanicking; typed-nil or hostile
// contexts are outside contract. This is not a hard wall-clock latency bound.
// Cancellation preserves C1 sampling, including successful capture before a
// post-sink cancel sample. There is no retry, reinsert, cross-turn history,
// durable/lossless delivery or hard termination guarantee on process failure.
func (view View) Poll(ctx context.Context, budget uint8) PollResult {
	if view.state == nil {
		return PollResult{Receipt: runtimediagnosticsconsumer.DrainResult{Stop: runtimediagnosticsconsumer.Disabled}}
	}
	if ctx == nil || budget == 0 || budget > 64 {
		return PollResult{Receipt: runtimediagnosticsconsumer.DrainResult{Stop: runtimediagnosticsconsumer.InvalidInput}}
	}
	s := view.state
	if !s.guard.TryLock() {
		return PollResult{Receipt: runtimediagnosticsconsumer.DrainResult{Stop: runtimediagnosticsconsumer.Busy}}
	}
	defer func() {
		s.events = [64]runtimediagnostics.Event{}
		s.count = 0
		s.guard.Unlock()
	}()
	s.events = [64]runtimediagnostics.Event{}
	s.count = 0
	receipt := s.consumer.Drain(ctx, budget)
	return PollResult{Receipt: receipt, Count: s.count, Events: s.events}
}
