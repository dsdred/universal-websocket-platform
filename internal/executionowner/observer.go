package executionowner

// TerminalObserver synchronously receives one immutable Terminal Result. The
// interface provides no execution, completion, Stop, or lease capability.
type TerminalObserver interface {
	Observe(TerminalResult)
}
