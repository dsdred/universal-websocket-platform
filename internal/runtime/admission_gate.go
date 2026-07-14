package runtime

import "sync/atomic"

type admissionGate struct {
	open atomic.Bool
}

func (gate *admissionGate) allow() {
	gate.open.Store(true)
}

func (gate *admissionGate) close() {
	gate.open.Store(false)
}

func (gate *admissionGate) canAccept() bool {
	return gate.open.Load()
}
