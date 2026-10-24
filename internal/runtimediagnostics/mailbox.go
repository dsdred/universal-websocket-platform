// Package runtimediagnostics provides an isolated, lossy diagnostic mailbox.
// It has no Runtime hooks, delivery workers or sink execution. Operations make
// at most one lock attempt; this is not a scheduling latency guarantee.
package runtimediagnostics

import "sync"

const (
	capacity      = 64
	schemaVersion = 1
)

// Code is a closed diagnostic identity. All values except the two below are invalid.
type Code uint8

const (
	// ListenerServeFailed identifies an unexpected Listener Serve return.
	ListenerServeFailed Code = iota + 1
	// AuthenticationProviderFailed identifies an eligible Provider operational failure.
	AuthenticationProviderFailed
)

func (c Code) valid() bool {
	return c == ListenerServeFailed || c == AuthenticationProviderFailed
}

// Event is a detached safe report. Sequence orders accepted insertions only
// within one mailbox lifetime; it is not a global identity or metric label.
type Event struct {
	SchemaVersion uint8
	Code          Code
	Sequence      uint64
}

func (e Event) valid() bool {
	return e.SchemaVersion == schemaVersion && e.Code.valid() && e.Sequence != 0
}

type mailbox struct {
	mu       sync.Mutex
	slots    [capacity]Event
	head     uint8
	count    uint8
	sequence uint64
}

// Producer can only offer closed codes to its original mailbox. A nil or zero
// Producer is disabled. Copying a capability shares, rather than copies, state.
type Producer struct {
	mailbox *mailbox
}

// Reader destructively shares queued events with any other readers of the same
// mailbox. A nil or zero Reader returns no item. It provides no replay or fanout.
type Reader struct {
	mailbox *mailbox
}

// New constructs one inert fixed64 mailbox and its separate capabilities.
// Construction starts no goroutines, timers or I/O and needs no lifecycle cleanup.
func New() (*Producer, *Reader) {
	m := &mailbox{}
	return &Producer{mailbox: m}, &Reader{mailbox: m}
}

// Offer attempts one insertion. False means loss: disabled, invalid, busy, full
// or permanently sequence-exhausted. It never waits, retries or invokes a sink.
// Only accepted insertions consume sequence numbers; full preserves queued FIFO.
func (p *Producer) Offer(code Code) bool {
	if p == nil || p.mailbox == nil || !code.valid() {
		return false
	}
	m := p.mailbox
	if !m.mu.TryLock() {
		return false
	}
	if m.count == capacity || m.sequence == ^uint64(0) {
		m.mu.Unlock()
		return false
	}
	m.sequence++
	tail := (m.head + m.count) % capacity
	m.slots[tail] = Event{SchemaVersion: schemaVersion, Code: code, Sequence: m.sequence}
	m.count++
	m.mu.Unlock()
	return true
}

// TryRead attempts to remove one event, clearing its slot before unlocking and
// returning a value copy. False returns Event{} for disabled, busy, empty or an
// invalid record. Consumers execute their own sinks only after this returns;
// sink blocking, panic or failure cannot reinsert an event or hold this lock.
func (r *Reader) TryRead() (Event, bool) {
	if r == nil || r.mailbox == nil {
		return Event{}, false
	}
	m := r.mailbox
	if !m.mu.TryLock() {
		return Event{}, false
	}
	if m.count == 0 {
		m.mu.Unlock()
		return Event{}, false
	}
	event := m.slots[m.head]
	m.slots[m.head] = Event{}
	m.head = (m.head + 1) % capacity
	m.count--
	m.mu.Unlock()
	if !event.valid() {
		return Event{}, false
	}
	return event, true
}
