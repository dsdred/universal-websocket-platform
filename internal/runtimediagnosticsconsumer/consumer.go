// Package runtimediagnosticsconsumer provides caller-scheduled, bounded, lossy
// consumption of safe diagnostic events. It owns no Runtime lifecycle, workers,
// timers or I/O. Callers must not invoke it inside Runtime callbacks or locks.
package runtimediagnosticsconsumer

import (
	"context"
	"sync"

	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnostics"
)

// Sink synchronously receives a detached allowlisted Event. Nil return means
// local success only, not acknowledgment or persistence. Sink-owned resources
// and effects are external; blocked/nonreturning sinks have no termination bound.
type Sink func(runtimediagnostics.Event) error

// DrainStop describes a local turn outcome, never a raw cause or wire category.
type DrainStop uint8

const (
	// Disabled means the Consumer has no usable construction inputs.
	Disabled DrainStop = iota
	// InvalidInput means nil context or a budget outside 1..64.
	InvalidInput
	// Busy means the single consumer guard attempt failed; nothing was read.
	Busy
	// NoItem means TryRead returned false; it does not prove empty or drained.
	NoItem
	// Budget means all permitted read attempts completed with local sink success.
	Budget
	// Canceled means cancellation was sampled outside a failed/panicking sink.
	Canceled
	// SinkFailed means one sink invocation returned a nonnil error.
	SinkFailed
	// SinkPanicked means one sink invocation panicked recoverably.
	SinkPanicked
)

// DrainResult is a bounded per-turn scalar receipt. Observed counts valid reads,
// not physical removals. Observed equals SinkSucceeded + SinkFailed +
// SinkPanicked + Discarded; counts are at most 64. Producer drops, remaining
// queue length and durable delivery are unknown; sequences cannot measure loss.
type DrainResult struct {
	ReadAttempts  uint8
	Observed      uint8
	SinkSucceeded uint8
	SinkFailed    uint8
	SinkPanicked  uint8
	Discarded     uint8
	Stop          DrainStop
}

// Consumer is an opaque copyable capability. Copies share state and a guard;
// separately constructed Consumers do not. One logical Reader owner per mailbox
// is required for FIFO sink invocation order; extra readers partition records.
// Zero Consumer is disabled. There is no Close, flush, reset or shutdown join.
type Consumer struct {
	state *consumerState
}

type consumerState struct {
	reader runtimediagnostics.Reader
	sink   Sink
	guard  sync.Mutex
}

// NewConsumer inertly captures Reader and Sink by value. A zero Reader or nil
// Sink returns a disabled Consumer; later variable reassignment cannot redirect
// it. Construction performs no read, offer, callback, worker or lifecycle action.
func NewConsumer(reader runtimediagnostics.Reader, sink Sink) Consumer {
	if reader == (runtimediagnostics.Reader{}) || sink == nil {
		return Consumer{}
	}
	return Consumer{state: &consumerState{reader: reader, sink: sink}}
}

// Drain attempts at most budget TryRead calls (1..64), retaining one Event copy
// at a time. Disabled precedes input validation, which precedes Busy. Context
// must be well-formed and nonnil; typed-nil/hostile contexts are outside contract.
// A failed guard or read is never retried. A valid removed event is delivered
// once or discarded on sampled pre-sink cancellation, never reinserted.
// Sink error/panic wins over later cancellation; nil return is counted first.
// A synchronous blocked sink delays this caller only. No receipt or hard stop
// is promised for nonreturning sinks, Goexit, fatal runtime or process failure.
func (consumer Consumer) Drain(ctx context.Context, budget uint8) DrainResult {
	if consumer.state == nil {
		return DrainResult{Stop: Disabled}
	}
	if ctx == nil || budget == 0 || budget > 64 {
		return DrainResult{Stop: InvalidInput}
	}
	s := consumer.state
	if !s.guard.TryLock() {
		return DrainResult{Stop: Busy}
	}
	defer s.guard.Unlock()
	var result DrainResult
	for result.ReadAttempts < budget {
		if ctx.Err() != nil {
			result.Stop = Canceled
			return result
		}
		result.ReadAttempts++
		event, ok := s.reader.TryRead()
		if !ok {
			result.Stop = NoItem
			return result
		}
		result.Observed++
		if ctx.Err() != nil {
			result.Discarded++
			result.Stop = Canceled
			return result
		}
		failed, panicked := callSink(s.sink, event)
		if panicked {
			result.SinkPanicked++
			result.Stop = SinkPanicked
			return result
		}
		if failed {
			result.SinkFailed++
			result.Stop = SinkFailed
			return result
		}
		result.SinkSucceeded++
		if ctx.Err() != nil {
			result.Stop = Canceled
			return result
		}
	}
	result.Stop = Budget
	return result
}

func callSink(sink Sink, event runtimediagnostics.Event) (failed, panicked bool) {
	normalReturn := false
	defer func() {
		if !normalReturn {
			_ = recover() // Deliberately discard payload, including nil.
			panicked = true
		}
	}()
	failed = sink(event) != nil
	normalReturn = true
	return
}
