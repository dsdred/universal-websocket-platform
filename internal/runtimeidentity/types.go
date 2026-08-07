// Package runtimeidentity implements the in-memory Runtime Instance aggregate
// store for DP-014 Runtime Operational Identity Persistence.
//
// The store enforces all conceptual operations from DP-014 §21 and satisfies
// all acceptance proofs from DP-014 §22. It is isolated, in-process, and
// durable only within process lifetime. No external storage, HTTP API,
// production wiring, or second lifecycle owner is introduced.
package runtimeidentity

import (
	"errors"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

// Sentinel errors returned by Store operations. Each error is definitive: the
// caller may inspect the category without re-reading aggregate state.
var (
	// ErrInstanceNotFound is returned when the requested RuntimeInstanceID does
	// not identify any aggregate in the management domain.
	ErrInstanceNotFound = errors.New("runtime instance not found")

	// ErrInstanceAlreadyExists is returned when CreateRuntimeInstance is called
	// with an ID that already identifies a committed aggregate.
	ErrInstanceAlreadyExists = errors.New("runtime instance already exists")

	// ErrStaleRevision is returned when the supplied expected revision does not
	// match the current aggregate revision. Zero mutation is performed.
	ErrStaleRevision = errors.New("stale aggregate revision")

	// ErrActiveAttemptExists is returned when ConditionalClaimLaunchAttempt is
	// called while an active attempt is already present.
	ErrActiveAttemptExists = errors.New("active launch attempt already exists")

	// ErrNoActiveAttempt is returned when an operation requires an active
	// attempt but none is present.
	ErrNoActiveAttempt = errors.New("no active launch attempt")

	// ErrAttemptIDReused is returned when a LaunchAttemptID is already present
	// in the complete history of the Runtime Instance.
	ErrAttemptIDReused = errors.New("launch attempt ID reused within instance history")

	// ErrInvalidAttemptPhase is returned when the current attempt phase does
	// not allow the requested operation.
	ErrInvalidAttemptPhase = errors.New("invalid launch attempt phase for operation")

	// ErrBindingAlreadyExists is returned when ConditionalBindExecutionGeneration
	// is called and a different execution generation is already bound.
	ErrBindingAlreadyExists = errors.New("execution generation binding already exists")

	// ErrInvalidIdentity is returned when a required identity field is empty or
	// otherwise invalid.
	ErrInvalidIdentity = errors.New("invalid identity")
)

// Revision is an opaque, monotonically advancing concurrency token. It
// identifies one committed aggregate state and is not a timestamp or business
// identity. Callers compare it for equality only.
type Revision uint64

// AttemptPhase records the current lifecycle phase of one Launch Attempt.
type AttemptPhase string

const (
	// AttemptPhaseClaimed is set when the attempt is first atomically claimed.
	// No Host work has begun.
	AttemptPhaseClaimed AttemptPhase = "claimed"

	// AttemptPhaseLaunching is set when Host startup has been initiated but not
	// confirmed Running.
	AttemptPhaseLaunching AttemptPhase = "launching"

	// AttemptPhaseRunning is set after the Owner confirms Host startup and
	// readiness.
	AttemptPhaseRunning AttemptPhase = "running"

	// AttemptPhaseStopping is set when Stop has been claimed for an attempt
	// that reached Launching or Running.
	AttemptPhaseStopping AttemptPhase = "stopping"

	// AttemptPhaseStopped is a terminal phase: the attempt completed a clean
	// stop with no remaining Host resources.
	AttemptPhaseStopped AttemptPhase = "stopped"

	// AttemptPhaseFailed is a terminal phase: the attempt ended in failure with
	// no remaining Host resources.
	AttemptPhaseFailed AttemptPhase = "failed"
)

// isTerminal reports whether the phase represents a terminal historical state.
func (p AttemptPhase) isTerminal() bool {
	return p == AttemptPhaseStopped || p == AttemptPhaseFailed
}

// DesiredState records the last accepted management intent.
type DesiredState string

const (
	// DesiredStateStopped is the initial desired state.
	DesiredStateStopped DesiredState = "stopped"

	// DesiredStateStarted is set when a Start intent has been accepted.
	DesiredStateStarted DesiredState = "started"
)

// ActualState records the last lifecycle fact confirmed by the Owner.
type ActualState string

const (
	// ActualStateStopped is the initial actual state.
	ActualStateStopped ActualState = "stopped"

	// ActualStateClaimed is set when a Launch Attempt has been atomically
	// claimed.
	ActualStateClaimed ActualState = "claimed"

	// ActualStateRunning is set when the Owner confirms Host startup.
	ActualStateRunning ActualState = "running"

	// ActualStateStopping is set when Stop has been claimed.
	ActualStateStopping ActualState = "stopping"

	// ActualStateFailed is set after a definitive failure that proves no Host
	// resources remain.
	ActualStateFailed ActualState = "failed"
)

// ExecutionGeneration is an opaque correlation identity allocated by the
// Control Service composition for its process-containment boundary. It is not
// allocated by persistence.
type ExecutionGeneration string

// LaunchAttemptRecord is one immutable-identity owned child of a Runtime
// Instance aggregate. Its identity and version pin are immutable after
// creation; its phase and terminal outcome may advance conditionally.
type LaunchAttemptRecord struct {
	// runtimeInstanceID is the immutable parent identity.
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID

	// launchAttemptID is the immutable child identity.
	launchAttemptID runtimeconfigload.LaunchAttemptID

	// configurationVersionID is the exact immutable Published
	// ConfigurationVersion pin.
	configurationVersionID uint64

	// phase is the current lifecycle phase.
	phase AttemptPhase

	// executionGeneration is the optional immutable opaque binding allocated by
	// the Control Service. Empty means no binding has been stored.
	executionGeneration ExecutionGeneration
}

// RuntimeInstanceID returns the immutable parent identity.
func (r LaunchAttemptRecord) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID {
	return r.runtimeInstanceID
}

// LaunchAttemptID returns the immutable child identity.
func (r LaunchAttemptRecord) LaunchAttemptID() runtimeconfigload.LaunchAttemptID {
	return r.launchAttemptID
}

// ConfigurationVersionID returns the exact immutable Published
// ConfigurationVersion pin.
func (r LaunchAttemptRecord) ConfigurationVersionID() uint64 {
	return r.configurationVersionID
}

// Phase returns the current lifecycle phase.
func (r LaunchAttemptRecord) Phase() AttemptPhase {
	return r.phase
}

// ExecutionGeneration returns the bound execution generation, or empty if none
// has been stored.
func (r LaunchAttemptRecord) ExecutionGeneration() ExecutionGeneration {
	return r.executionGeneration
}

// RuntimeInstanceView is one coherent read of a Runtime Instance aggregate. All
// fields correspond to a single committed revision. Reads are observation only
// and do not advance revision.
type RuntimeInstanceView struct {
	// workspaceID is the immutable Workspace binding.
	workspaceID uint64

	// configurationID is the immutable Configuration binding.
	configurationID uint64

	// runtimeInstanceID is the immutable aggregate identity.
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID

	// revision is the aggregate revision at the time of the read.
	revision Revision

	// desired is the last accepted management intent.
	desired DesiredState

	// actual is the last lifecycle fact confirmed by the Owner.
	actual ActualState

	// activeAttemptID is the identity of the active attempt, if any.
	activeAttemptID runtimeconfigload.LaunchAttemptID

	// hasActiveAttempt reports whether an active attempt is present.
	hasActiveAttempt bool
}

// WorkspaceID returns the immutable Workspace binding.
func (v RuntimeInstanceView) WorkspaceID() uint64 { return v.workspaceID }

// ConfigurationID returns the immutable Configuration binding.
func (v RuntimeInstanceView) ConfigurationID() uint64 { return v.configurationID }

// RuntimeInstanceID returns the immutable aggregate identity.
func (v RuntimeInstanceView) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID {
	return v.runtimeInstanceID
}

// Revision returns the aggregate revision at the time of the read.
func (v RuntimeInstanceView) Revision() Revision { return v.revision }

// DesiredState returns the last accepted management intent.
func (v RuntimeInstanceView) DesiredState() DesiredState { return v.desired }

// ActualState returns the last lifecycle fact confirmed by the Owner.
func (v RuntimeInstanceView) ActualState() ActualState { return v.actual }

// ActiveAttempt returns the active attempt identity and whether one exists.
func (v RuntimeInstanceView) ActiveAttempt() (runtimeconfigload.LaunchAttemptID, bool) {
	return v.activeAttemptID, v.hasActiveAttempt
}

// ClaimResult is returned by ConditionalClaimLaunchAttempt. It carries the new
// revision on success and a definitive error on rejection.
type ClaimResult struct {
	revision  Revision
	committed bool
}

// Committed reports whether the claim was atomically committed.
func (r ClaimResult) Committed() bool { return r.committed }

// Revision returns the new aggregate revision. Only meaningful when Committed.
func (r ClaimResult) Revision() Revision { return r.revision }

// PublishResult is returned by conditional publication operations. It carries
// the new revision on commit and a definitive error on rejection.
type PublishResult struct {
	revision  Revision
	committed bool
}

// Committed reports whether the publication was atomically committed.
func (r PublishResult) Committed() bool { return r.committed }

// Revision returns the new aggregate revision. Only meaningful when Committed.
func (r PublishResult) Revision() Revision { return r.revision }
