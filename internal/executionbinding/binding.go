// Package executionbinding provides one-shot execution publication.
package executionbinding

import (
	"errors"
	"sync"
	"sync/atomic"
)

type bindingError string

const (
	// ErrInvalidOutcome indicates that publication received an unsupported outcome.
	ErrInvalidOutcome bindingError = "invalid execution publication outcome"
	// ErrWaitAlreadyClaimed indicates that the dormant path was already claimed.
	ErrWaitAlreadyClaimed bindingError = "execution publication wait already claimed"
	// ErrInvalidCommitPublisher indicates that a committed-publication facet is
	// uninitialized or its Binding already has an outcome.
	ErrInvalidCommitPublisher bindingError = "invalid execution commit publisher"
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

// CommitPublisher is the narrow Manager-facing facet that can publish only a
// committed outcome. It exposes neither Wait nor non-committed publication.
type CommitPublisher struct {
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

// CommitPublisher returns the committed-publication facet for this Binding.
func (binding *Binding) CommitPublisher() CommitPublisher {
	if binding == nil {
		return CommitPublisher{}
	}
	return CommitPublisher{state: binding.state}
}

// ValidateFresh verifies that this publisher belongs to an unpublished
// Binding. Publication ownership must remain serialized after validation.
func (publisher CommitPublisher) ValidateFresh() error {
	if publisher.state == nil || publisher.state.ready == nil {
		return ErrInvalidCommitPublisher
	}
	select {
	case <-publisher.state.ready:
		return ErrInvalidCommitPublisher
	default:
		return nil
	}
}

// PublishCommitted fixes and returns the committed outcome. For a valid fresh
// publisher this operation is nonblocking and panic-free.
func (publisher CommitPublisher) PublishCommitted() (Outcome, error) {
	if err := publisher.ValidateFresh(); err != nil {
		return 0, err
	}
	state := publisher.state
	state.publishOnce.Do(func() {
		state.outcome = OutcomeCommitted
		close(state.ready)
	})
	<-state.ready
	if state.outcome != OutcomeCommitted {
		return state.outcome, errors.Join(ErrInvalidCommitPublisher, ErrInvalidOutcome)
	}
	return state.outcome, nil
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
