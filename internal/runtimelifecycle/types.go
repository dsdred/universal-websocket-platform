// Package runtimelifecycle owns the in-process lifecycle of one Runtime
// Instance.
package runtimelifecycle

import (
	"context"
	"errors"

	"github.com/dsdred/universal-websocket-platform/internal/runtime"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

// ErrInvalidExpectedAttempt reports an empty expected Launch Attempt identity.
var ErrInvalidExpectedAttempt = errors.New("invalid expected Launch Attempt identity")

var (
	ErrInvalidOwner             = errors.New("invalid Runtime Lifecycle Owner")
	ErrInvalidStartRequest      = errors.New("invalid Runtime start request")
	ErrAttemptIDSourceFailed    = errors.New("Launch Attempt ID source failed")
	ErrAttemptIDReused          = errors.New("Launch Attempt ID was reused")
	ErrStartConflict            = errors.New("Runtime start conflicts with active lifecycle work")
	ErrPreparationNotOwned      = errors.New("launch preparation is not owned by this Owner")
	ErrInvalidPreparationResult = errors.New("invalid launch preparation result")
)

type DesiredState string

const (
	DesiredStopped DesiredState = "stopped"
	DesiredRunning DesiredState = "running"
)

type ActualState string

const (
	ActualStopped  ActualState = "stopped"
	ActualStarting ActualState = "starting"
	ActualRunning  ActualState = "running"
	ActualStopping ActualState = "stopping"
	ActualFailed   ActualState = "failed"
)

type LaunchAttemptIDSource func() (runtimeconfigload.LaunchAttemptID, error)

type StartRequest struct {
	workspaceID            uint64
	configurationID        uint64
	configurationVersionID uint64
}

func NewStartRequest(workspaceID, configurationID, configurationVersionID uint64) StartRequest {
	return StartRequest{
		workspaceID:            workspaceID,
		configurationID:        configurationID,
		configurationVersionID: configurationVersionID,
	}
}

func (r StartRequest) WorkspaceID() uint64            { return r.workspaceID }
func (r StartRequest) ConfigurationID() uint64        { return r.configurationID }
func (r StartRequest) ConfigurationVersionID() uint64 { return r.configurationVersionID }

type LaunchPreparation struct {
	owner       *Owner
	attempt     *attemptState
	loadRequest runtimeconfigload.LoadRequest
	context     context.Context
}

func (p LaunchPreparation) LoadRequest() runtimeconfigload.LoadRequest { return p.loadRequest }
func (p LaunchPreparation) Context() context.Context                   { return p.context }

type PreparationResultKind string

const (
	PreparationSnapshot PreparationResultKind = "snapshot"
	PreparationFailure  PreparationResultKind = "failure"
)

type PreparationResult struct {
	kind        PreparationResultKind
	snapshot    runtimeconfig.Snapshot
	failure     error
	hasSnapshot bool
	hasFailure  bool
}

func PreparedSnapshot(snapshot runtimeconfig.Snapshot) PreparationResult {
	return PreparationResult{
		kind:        PreparationSnapshot,
		snapshot:    snapshot,
		hasSnapshot: true,
	}
}

func FailedPreparation(cause error) PreparationResult {
	return PreparationResult{
		kind:       PreparationFailure,
		failure:    cause,
		hasFailure: true,
	}
}

func (r PreparationResult) Kind() PreparationResultKind { return r.kind }

func (r PreparationResult) Snapshot() (runtimeconfig.Snapshot, bool) {
	return r.snapshot, r.kind == PreparationSnapshot && r.hasSnapshot && !r.hasFailure
}

func (r PreparationResult) Failure() (error, bool) {
	return r.failure, r.kind == PreparationFailure && r.hasFailure && !r.hasSnapshot && r.failure != nil
}

type StartOutcomeKind string

const (
	StartRunning              StartOutcomeKind = "running"
	StartPreparationFailed    StartOutcomeKind = "preparation-failed"
	StartLaunchFailed         StartOutcomeKind = "launch-failed"
	StartStoppedBeforeRunning StartOutcomeKind = "stopped-before-running"
)

type StartOutcome struct {
	kind               StartOutcomeKind
	attempt            AttemptFact
	preparationFailure error
	launchOutcome      runtime.BootstrapOutcome
}

func (r StartOutcome) Kind() StartOutcomeKind { return r.kind }
func (r StartOutcome) Attempt() AttemptFact   { return r.attempt }

func (r StartOutcome) PreparationFailure() (error, bool) {
	return r.preparationFailure, r.kind == StartPreparationFailed
}

func (r StartOutcome) LaunchOutcome() (runtime.BootstrapOutcome, bool) {
	return r.launchOutcome, r.kind == StartLaunchFailed
}

type StopOutcomeKind string

const (
	StopStopped StopOutcomeKind = "stopped"
	StopFailed  StopOutcomeKind = "stop-failed"
)

// StopAttemptMismatch reports that the relevant attempt does not match the expected identity.
const StopAttemptMismatch StopOutcomeKind = "attempt-mismatch"

type StopOutcome struct {
	kind       StopOutcomeKind
	attempt    AttemptFact
	hasAttempt bool
	failure    error
}

func (r StopOutcome) Kind() StopOutcomeKind { return r.kind }

func (r StopOutcome) Attempt() (AttemptFact, bool) {
	return r.attempt, r.hasAttempt
}

func (r StopOutcome) Failure() (error, bool) {
	return r.failure, r.kind == StopFailed
}

type AttemptPhase string

const (
	AttemptPreparing  AttemptPhase = "preparing"
	AttemptLaunching  AttemptPhase = "launching"
	AttemptRunning    AttemptPhase = "running"
	AttemptStopping   AttemptPhase = "stopping"
	AttemptHistorical AttemptPhase = "historical"
)

type StopOrigin string

const (
	StopNotClaimed    StopOrigin = ""
	StopBeforeRunning StopOrigin = "before-running"
	StopAfterRunning  StopOrigin = "after-running"
)

type AttemptTerminalKind string

const (
	AttemptNotTerminal          AttemptTerminalKind = ""
	AttemptPreparationFailed    AttemptTerminalKind = "preparation-failed"
	AttemptLaunchFailed         AttemptTerminalKind = "launch-failed"
	AttemptStoppedBeforeRunning AttemptTerminalKind = "stopped-before-running"
	AttemptStopped              AttemptTerminalKind = "stopped"
	AttemptStopFailed           AttemptTerminalKind = "stop-failed"
)

type AttemptFact struct {
	workspaceID            uint64
	configurationID        uint64
	configurationVersionID uint64
	runtimeInstanceID      runtimeconfigload.RuntimeInstanceID
	launchAttemptID        runtimeconfigload.LaunchAttemptID
	phase                  AttemptPhase
	stopOrigin             StopOrigin
	runningPublished       bool
	terminalKind           AttemptTerminalKind
}

func (f AttemptFact) WorkspaceID() uint64            { return f.workspaceID }
func (f AttemptFact) ConfigurationID() uint64        { return f.configurationID }
func (f AttemptFact) ConfigurationVersionID() uint64 { return f.configurationVersionID }
func (f AttemptFact) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID {
	return f.runtimeInstanceID
}
func (f AttemptFact) LaunchAttemptID() runtimeconfigload.LaunchAttemptID {
	return f.launchAttemptID
}
func (f AttemptFact) Phase() AttemptPhase               { return f.phase }
func (f AttemptFact) StopOrigin() StopOrigin            { return f.stopOrigin }
func (f AttemptFact) RunningPublished() bool            { return f.runningPublished }
func (f AttemptFact) TerminalKind() AttemptTerminalKind { return f.terminalKind }

type Observation struct {
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID
	workspaceID       uint64
	configurationID   uint64
	desired           DesiredState
	actual            ActualState
	active            AttemptFact
	hasActive         bool
	last              AttemptFact
	hasLast           bool
}

func (s Observation) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID {
	return s.runtimeInstanceID
}
func (s Observation) WorkspaceID() uint64        { return s.workspaceID }
func (s Observation) ConfigurationID() uint64    { return s.configurationID }
func (s Observation) DesiredState() DesiredState { return s.desired }
func (s Observation) ActualState() ActualState   { return s.actual }
func (s Observation) ActiveAttempt() (AttemptFact, bool) {
	return s.active, s.hasActive
}
func (s Observation) LastAttempt() (AttemptFact, bool) {
	return s.last, s.hasLast
}
