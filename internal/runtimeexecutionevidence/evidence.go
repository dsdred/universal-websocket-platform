// Package runtimeexecutionevidence composes the repository-private DP-022
// full-tuple execution evidence boundary.
//
// It binds one containment authority to coherent DP-014 exact-attempt reads.
// The package does not wire production admission, mutate durable truth, or
// implement recovery.
package runtimeexecutionevidence

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecontainment"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeidentity"
)

// Question selects the one DP-022 fact requested by an invocation.
type Question uint8

const (
	// GenerationStatus asks whether the exact generation is current or
	// terminated.
	GenerationStatus Question = iota + 1
	// CoveredResourceAbsence asks whether containment-covered resources of an
	// exact terminated generation are absent.
	CoveredResourceAbsence
	// ShutdownCompletion asks whether the exact terminated generation also has
	// DP-014 Owner shutdown-completion provenance.
	ShutdownCompletion
)

func (q Question) valid() bool {
	return q == GenerationStatus || q == CoveredResourceAbsence || q == ShutdownCompletion
}

// OutcomeKind is the closed result kind returned by EvidenceHandle.Consume.
type OutcomeKind uint8

const (
	// Unknown proves no positive DP-022 fact and carries a closed Reason.
	Unknown OutcomeKind = iota + 1
	// GenerationLive proves that the exact queried generation is current.
	GenerationLive
	// GenerationTerminated proves that the exact queried prior generation is
	// terminated.
	GenerationTerminated
	// CoveredResourcesAbsent proves absence of resources covered by the
	// containment guarantee for the exact terminated generation.
	CoveredResourcesAbsent
	// HostShutdownCompleted proves exact Owner shutdown provenance for the exact
	// attempt in the exact terminated generation.
	HostShutdownCompleted
	// LiveUnownedExecution is reserved by DP-022 and is unreachable for the
	// current ProcessContainment topology.
	LiveUnownedExecution
)

// UnknownReason is the closed reason carried by an Unknown outcome.
type UnknownReason string

const (
	// Absent reports that the requested positive fact does not follow from the
	// otherwise coherent tuple.
	Absent UnknownReason = "Absent"
	// Unavailable reports unavailable authority or source state.
	Unavailable UnknownReason = "Unavailable"
	// Stale reports a changed revision or a consumed/replayed handle.
	Stale UnknownReason = "Stale"
	// ScopeMismatch reports an incomplete, foreign, or unbound exact tuple.
	ScopeMismatch UnknownReason = "ScopeMismatch"
	// Contradictory reports inconsistent authoritative facts.
	Contradictory UnknownReason = "Contradictory"
	// Indeterminate reports an observation with no definite conclusion.
	Indeterminate UnknownReason = "Indeterminate"
	// Cancelled reports caller cancellation.
	Cancelled UnknownReason = "Cancelled"
	// UnsupportedTopology reports that the active topology cannot provide the
	// requested containment guarantee.
	UnsupportedTopology UnknownReason = "UnsupportedTopology"
	// GuaranteeNotDeclared reports absence of an active declared guarantee.
	GuaranteeNotDeclared UnknownReason = "GuaranteeNotDeclared"
)

// Outcome is one closed DP-022 evidence result. It exposes no identity payload
// or mutable authority.
type Outcome struct {
	kind   OutcomeKind
	reason UnknownReason
}

// Kind returns the closed outcome kind.
func (r Outcome) Kind() OutcomeKind {
	if r.kind == 0 {
		return Unknown
	}
	return r.kind
}

// Reason returns the closed reason only for Unknown.
func (r Outcome) Reason() UnknownReason {
	if r.Kind() != Unknown {
		return ""
	}
	if r.reason == "" {
		return Indeterminate
	}
	return r.reason
}

func unknown(reason UnknownReason) Outcome { return Outcome{kind: Unknown, reason: reason} }

type identityReader interface {
	ReadRuntimeInstance(runtimeconfigload.RuntimeInstanceID) (runtimeidentity.RuntimeInstanceView, error)
	ReadLaunchAttemptHistory(runtimeconfigload.RuntimeInstanceID) ([]runtimeidentity.LaunchAttemptRecord, error)
}

type generationRead struct {
	kind   runtimecontainment.EvidenceKind
	reason runtimecontainment.UnknownReason
}

type generationReader func(context.Context, runtimecontainment.Domain, string) generationRead

// Composer binds one immutable containment domain, one concrete active
// authority, and one read-only DP-014 source. Its state does not change across
// independent queries.
type Composer struct {
	domain         runtimecontainment.Domain
	identities     identityReader
	readGeneration generationReader
}

// NewComposer constructs the repository-private full-tuple composer. A nil or
// invalid dependency is preserved as a fail-closed query outcome; construction
// grants no authority and performs no observation.
func NewComposer(
	domain runtimecontainment.Domain,
	authority *runtimecontainment.ActiveAuthority,
	identities identityReader,
) *Composer {
	return &Composer{
		domain:     domain,
		identities: identities,
		readGeneration: func(ctx context.Context, requested runtimecontainment.Domain, generation string) generationRead {
			evidence := authority.ReadGeneration(ctx, requested, generation)
			return generationRead{kind: evidence.Kind(), reason: evidence.Reason()}
		},
	}
}

// EvidenceHandle is an opaque invocation-scoped, use-once evidence capability. Value
// copies share one private consumed cell and therefore cannot duplicate use.
// Its zero value is invalid and consumes as Unknown(Stale).
type EvidenceHandle struct {
	state *handleState
}

type handleState struct {
	used atomic.Bool

	identities     identityReader
	readGeneration generationReader
	domain         runtimecontainment.Domain
	generation     runtimeidentity.ExecutionGeneration
	question       Question
	snapshot       runtimeidentity.ExactAttemptSnapshot
	initial        generationRead
	terminal       Outcome
	readIdentity   *readIdentity
}

type readIdentity struct{ marker *byte }

// Query performs one fresh exact-attempt and generation observation and binds
// it to a use-once handle. No positive result is exposed before Consume.
func (c *Composer) Query(
	ctx context.Context,
	domain runtimecontainment.Domain,
	instanceID runtimeconfigload.RuntimeInstanceID,
	attemptID runtimeconfigload.LaunchAttemptID,
	generation runtimeidentity.ExecutionGeneration,
	question Question,
) EvidenceHandle {
	state := &handleState{
		terminal:     unknown(ScopeMismatch),
		readIdentity: &readIdentity{marker: new(byte)},
	}
	result := EvidenceHandle{state: state}
	if c == nil || c.identities == nil || c.readGeneration == nil ||
		domain == (runtimecontainment.Domain{}) || domain != c.domain ||
		instanceID == "" || attemptID == "" || generation == "" || !question.valid() {
		return result
	}

	snapshot, err := runtimeidentity.ReadExactAttemptSnapshot(c.identities, instanceID, attemptID)
	if err != nil {
		state.terminal = unknown(mapIdentityError(err))
		return result
	}
	if snapshot.RuntimeInstanceID() != instanceID || snapshot.LaunchAttemptID() != attemptID ||
		snapshot.ExecutionGeneration() != generation {
		return result
	}

	initial := c.readGeneration(ctx, domain, string(generation))
	if initial.kind == runtimecontainment.EvidenceUnknown {
		state.terminal = unknown(mapContainmentReason(initial.reason))
		return result
	}
	if initial.kind != runtimecontainment.EvidenceGenerationLive &&
		initial.kind != runtimecontainment.EvidenceGenerationTerminated {
		state.terminal = unknown(Contradictory)
		return result
	}
	if err := runtimeidentity.RevalidateExactAttemptSnapshot(c.identities, snapshot); err != nil {
		state.terminal = unknown(mapIdentityError(err))
		return result
	}

	state.identities = c.identities
	state.readGeneration = c.readGeneration
	state.domain = domain
	state.generation = generation
	state.question = question
	state.snapshot = snapshot
	state.initial = initial
	state.terminal = Outcome{}
	return result
}

// Consume atomically spends the handle before any external observation, then
// freshly revalidates the exact aggregate revision and containment authority.
// Every later, copied, or concurrent consumption returns Unknown(Stale).
func (h EvidenceHandle) Consume(ctx context.Context) Outcome {
	state := h.state
	if state == nil || !state.used.CompareAndSwap(false, true) {
		return unknown(Stale)
	}
	if state.terminal.Kind() == Unknown && state.terminal.reason != "" {
		return state.terminal
	}
	if state.readIdentity == nil || state.readIdentity.marker == nil ||
		state.identities == nil || state.readGeneration == nil {
		return unknown(Stale)
	}
	if err := runtimeidentity.RevalidateExactAttemptSnapshot(state.identities, state.snapshot); err != nil {
		return unknown(mapIdentityError(err))
	}

	fresh := state.readGeneration(ctx, state.domain, string(state.generation))
	if fresh.kind == runtimecontainment.EvidenceUnknown {
		return unknown(mapContainmentReason(fresh.reason))
	}
	if fresh.kind != runtimecontainment.EvidenceGenerationLive &&
		fresh.kind != runtimecontainment.EvidenceGenerationTerminated {
		return unknown(Contradictory)
	}
	if fresh.kind != state.initial.kind {
		return unknown(Contradictory)
	}
	return project(state.snapshot, state.question, fresh.kind)
}

func project(
	snapshot runtimeidentity.ExactAttemptSnapshot,
	question Question,
	kind runtimecontainment.EvidenceKind,
) Outcome {
	switch kind {
	case runtimecontainment.EvidenceGenerationLive:
		if question == GenerationStatus {
			return Outcome{kind: GenerationLive}
		}
		return unknown(Absent)
	case runtimecontainment.EvidenceGenerationTerminated:
		switch question {
		case GenerationStatus:
			return Outcome{kind: GenerationTerminated}
		case CoveredResourceAbsence:
			return Outcome{kind: CoveredResourcesAbsent}
		case ShutdownCompletion:
			if snapshot.Phase() == runtimeidentity.AttemptPhaseStopped &&
				snapshot.TerminalCompletionBasis() == runtimeidentity.TerminalCompletionOwnerShutdownCompleted {
				return Outcome{kind: HostShutdownCompleted}
			}
			return unknown(Absent)
		}
	}
	return unknown(Contradictory)
}

func mapIdentityError(err error) UnknownReason {
	switch {
	case errors.Is(err, runtimeidentity.ErrInvalidIdentity),
		errors.Is(err, runtimeidentity.ErrInstanceNotFound),
		errors.Is(err, runtimeidentity.ErrAttemptNotFound),
		errors.Is(err, runtimeidentity.ErrExecutionGenerationNotBound):
		return ScopeMismatch
	case errors.Is(err, runtimeidentity.ErrStaleRevision):
		return Stale
	case errors.Is(err, runtimeidentity.ErrIncoherentAttemptSnapshot):
		return Contradictory
	default:
		return Indeterminate
	}
}

func mapContainmentReason(reason runtimecontainment.UnknownReason) UnknownReason {
	switch reason {
	case runtimecontainment.UnknownUnavailable:
		return Unavailable
	case runtimecontainment.UnknownScopeMismatch:
		return ScopeMismatch
	case runtimecontainment.UnknownContradictory:
		return Contradictory
	case runtimecontainment.UnknownIndeterminate:
		return Indeterminate
	case runtimecontainment.UnknownCancelled:
		return Cancelled
	case runtimecontainment.UnknownUnsupportedTopology:
		return UnsupportedTopology
	case runtimecontainment.UnknownGuaranteeNotDeclared:
		return GuaranteeNotDeclared
	default:
		return Indeterminate
	}
}
