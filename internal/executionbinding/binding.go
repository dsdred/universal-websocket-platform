// Package executionbinding provides one-shot execution publication.
package executionbinding

import (
	"sync"
	"sync/atomic"
)

type bindingError string

const (
	// ErrInvalidOutcome indicates that publication received an unsupported outcome.
	ErrInvalidOutcome bindingError = "invalid execution publication outcome"
	// ErrWaitAlreadyClaimed indicates that the dormant path was already claimed.
	ErrWaitAlreadyClaimed bindingError = "execution publication wait already claimed"
)

func (err bindingError) Error() string {
	return string(err)
}

// Outcome is the immutable result of one execution publication.
type Outcome uint8

const (
	// OutcomeCommitted permits the dormant execution path to continue.
	OutcomeCommitted Outcome = iota + 1
	// OutcomeNotCommitted terminates the dormant execution path without execution.
	OutcomeNotCommitted
)

// Binding publishes one immutable outcome to one pre-existing dormant path.
// A Binding must be created with New.
type Binding struct {
	state *bindingState
}

type bindingState struct {
	publishOnce sync.Once
	ready       chan struct{}
	outcome     Outcome
	waitClaimed atomic.Bool
}

// New creates an unpublished Binding.
func New() *Binding {
	return &Binding{
		state: &bindingState{ready: make(chan struct{})},
	}
}

// Publish fixes and returns the publication outcome.
// Repeated valid calls return the outcome fixed by the first call.
func (binding *Binding) Publish(outcome Outcome) (Outcome, error) {
	if !outcome.valid() {
		return 0, ErrInvalidOutcome
	}

	state := binding.state
	state.publishOnce.Do(func() {
		state.outcome = outcome
		close(state.ready)
	})
	<-state.ready
	return state.outcome, nil
}

// Wait claims the single dormant execution path and waits for publication.
func (binding *Binding) Wait() (Outcome, error) {
	state := binding.state
	if !state.waitClaimed.CompareAndSwap(false, true) {
		return 0, ErrWaitAlreadyClaimed
	}

	<-state.ready
	return state.outcome, nil
}

func (outcome Outcome) valid() bool {
	return outcome == OutcomeCommitted || outcome == OutcomeNotCommitted
}
