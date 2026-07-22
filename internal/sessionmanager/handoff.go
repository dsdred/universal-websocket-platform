package sessionmanager

import (
	"errors"
	"sync"

	"github.com/dsdred/universal-websocket-platform/internal/executionbinding"
)

var (
	// ErrInvalidCommitHandoff indicates an uninitialized or inconsistent
	// Commit-to-dormant handoff capability.
	ErrInvalidCommitHandoff = errors.New("invalid Commit handoff")
	// ErrCommitHandoffAlreadyPublished indicates that a committed publication
	// was attempted after the one-shot outcome had already been fixed.
	ErrCommitHandoffAlreadyPublished = errors.New("Commit handoff already published")
)

// CommitHandoff owns one domain-specific Commit-to-dormant publication. It
// exposes behavior only through restricted publisher and waiter facets.
type CommitHandoff struct {
	state *commitHandoffState
}

// NotCommittedPublisher is the Dispatcher-facing facet. It can publish only
// the terminal pre-Commit outcome and exposes no committed capabilities.
type NotCommittedPublisher struct {
	state *commitHandoffState
}

// CommitHandoffPublisher is the opaque Manager-facing identity facet. Only
// Session Manager internals can validate it or publish through it.
type CommitHandoffPublisher struct {
	state *commitHandoffState
}

// CommitHandoffWaiter is the dormant-path facet. Exactly one waiter may claim
// the prepared execution path.
type CommitHandoffWaiter struct {
	state *commitHandoffState
}

// CommitHandoffOutcome is the immutable result observed by the dormant path.
// A non-committed outcome never exposes a CommitResult.
type CommitHandoffOutcome struct {
	publication executionbinding.Outcome
	result      CommitResult
}

type commitHandoffState struct {
	mu      sync.Mutex
	binding *executionbinding.Binding
	result  CommitResult
}

// NewCommitHandoff creates one unpublished Commit-to-dormant handoff.
func NewCommitHandoff() *CommitHandoff {
	return &CommitHandoff{
		state: &commitHandoffState{binding: executionbinding.New()},
	}
}

// NotCommittedPublisher returns the Dispatcher-facing publication facet.
func (handoff *CommitHandoff) NotCommittedPublisher() NotCommittedPublisher {
	if handoff == nil {
		return NotCommittedPublisher{}
	}
	return NotCommittedPublisher{state: handoff.state}
}

// CommitPublisher returns the Session Manager-facing publication facet.
func (handoff *CommitHandoff) CommitPublisher() CommitHandoffPublisher {
	if handoff == nil {
		return CommitHandoffPublisher{}
	}
	return CommitHandoffPublisher{state: handoff.state}
}

// Waiter returns the single dormant-path observation facet.
func (handoff *CommitHandoff) Waiter() CommitHandoffWaiter {
	if handoff == nil {
		return CommitHandoffWaiter{}
	}
	return CommitHandoffWaiter{state: handoff.state}
}

// Publish fixes NotCommitted. Repeated NotCommitted publication is
// idempotent; it never overwrites a committed result.
func (publisher NotCommittedPublisher) Publish() error {
	if publisher.state == nil || publisher.state.binding == nil {
		return ErrInvalidCommitHandoff
	}

	state := publisher.state
	state.mu.Lock()
	defer state.mu.Unlock()
	outcome, err := state.binding.Publish(executionbinding.OutcomeNotCommitted)
	if err != nil {
		return errors.Join(ErrInvalidCommitHandoff, err)
	}
	if outcome != executionbinding.OutcomeNotCommitted {
		return ErrCommitHandoffAlreadyPublished
	}
	return nil
}

// validateFresh verifies that this facet belongs to an unpublished handoff.
// It remains package-private so only Session Manager can exercise the
// committed-publication authority carried by the opaque exported value.
func (publisher CommitHandoffPublisher) validateFresh() error {
	if publisher.state == nil || publisher.state.binding == nil {
		return ErrInvalidCommitHandoff
	}
	publisher.state.mu.Lock()
	defer publisher.state.mu.Unlock()
	if err := publisher.state.binding.CommitPublisher().ValidateFresh(); err != nil {
		return errors.Join(ErrInvalidCommitHandoff, err)
	}
	return nil
}

func (publisher CommitHandoffPublisher) lockFresh() (*commitHandoffState, error) {
	if publisher.state == nil || publisher.state.binding == nil {
		return nil, ErrInvalidCommitHandoff
	}
	state := publisher.state
	state.mu.Lock()
	if err := state.binding.CommitPublisher().ValidateFresh(); err != nil {
		state.mu.Unlock()
		return nil, errors.Join(ErrCommitHandoffAlreadyPublished, err)
	}
	return state, nil
}

func (state *commitHandoffState) publishCommittedLocked(result CommitResult) {
	// lockFresh and CommitResult validation are held continuously through this
	// call. The underlying committed publication is therefore the documented
	// panic-free, non-failing operation for a fresh serialized publisher.
	state.result = result
	_, _ = state.binding.CommitPublisher().PublishCommitted()
}

func (publisher CommitHandoffPublisher) sameIdentity(other CommitHandoffPublisher) bool {
	return publisher.state != nil && publisher.state == other.state
}

// Wait claims and observes the single immutable handoff outcome.
func (waiter CommitHandoffWaiter) Wait() (CommitHandoffOutcome, error) {
	if waiter.state == nil || waiter.state.binding == nil {
		return CommitHandoffOutcome{}, ErrInvalidCommitHandoff
	}
	outcome, err := waiter.state.binding.Wait()
	if err != nil {
		return CommitHandoffOutcome{}, errors.Join(ErrInvalidCommitHandoff, err)
	}
	if outcome == executionbinding.OutcomeNotCommitted {
		return CommitHandoffOutcome{publication: outcome}, nil
	}

	waiter.state.mu.Lock()
	result := waiter.state.result
	waiter.state.mu.Unlock()
	if outcome != executionbinding.OutcomeCommitted || !result.valid() {
		return CommitHandoffOutcome{}, ErrInvalidCommitHandoff
	}
	return CommitHandoffOutcome{publication: outcome, result: result}, nil
}

// Committed reports whether this outcome carries the complete CommitResult.
func (outcome CommitHandoffOutcome) Committed() bool {
	return outcome.publication == executionbinding.OutcomeCommitted && outcome.result.valid()
}

// CommitResult returns the committed result only for a valid Committed
// outcome. NotCommitted returns the zero result and false.
func (outcome CommitHandoffOutcome) CommitResult() (CommitResult, bool) {
	if !outcome.Committed() {
		return CommitResult{}, false
	}
	return outcome.result, true
}
