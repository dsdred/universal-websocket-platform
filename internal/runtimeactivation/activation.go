// Package runtimeactivation composes exact DP-016 lifecycle orchestration.
package runtimeactivation

import (
	"context"
	"errors"
	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecommandidempotency"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeidentity"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle"
	"github.com/dsdred/universal-websocket-platform/internal/runtimemanagement"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

// ErrInvalidRequest reports an incomplete exact command request.
var ErrInvalidRequest = errors.New("invalid Runtime activation request")

// ErrVersionNotPublished reports a foreign or non-Published exact target.
var ErrVersionNotPublished = errors.New("Runtime activation target is not exact Published version")

var errIncoherentState = errors.New("Runtime activation state is incoherent")

type versionStore interface {
	Get(uint64) (configurationversion.ConfigurationVersion, error)
}
type identityStore interface {
	ReadRuntimeInstance(runtimeconfigload.RuntimeInstanceID) (runtimeidentity.RuntimeInstanceView, error)
	ReadLaunchAttemptHistory(runtimeconfigload.RuntimeInstanceID) ([]runtimeidentity.LaunchAttemptRecord, error)
	ConditionalClaimStop(runtimeconfigload.RuntimeInstanceID, runtimeidentity.Revision, runtimeconfigload.LaunchAttemptID) (runtimeidentity.PublishResult, error)
	ConditionalPublishRunning(runtimeconfigload.RuntimeInstanceID, runtimeidentity.Revision, runtimeconfigload.LaunchAttemptID) (runtimeidentity.PublishResult, error)
	ConditionalPublishTerminal(runtimeconfigload.RuntimeInstanceID, runtimeidentity.Revision, runtimeconfigload.LaunchAttemptID, bool) (runtimeidentity.PublishResult, error)
}
type lifecycleOwner interface {
	Observe() runtimelifecycle.Observation
	StopExpectedAttempt(context.Context, runtimeconfigload.LaunchAttemptID) (runtimelifecycle.StopOutcome, error)
}
type managedStartInvoker interface {
	InvokeManagedStart(context.Context, runtimelifecycle.StartRequest, runtimeorchestrationbinding.StartExecutionBinding) (runtimelifecycle.StartOutcome, error)
}

// Request is one immutable command key, exact target, and optional tracked Start fact.
type Request struct {
	key             runtimecommandidempotency.CommandKey
	versionID       uint64
	trackedStartKey runtimecommandidempotency.CommandKey
	trackedRevision runtimecommandidempotency.Revision
}

// NewRequest constructs an ordinary exact-version command request.
func NewRequest(key runtimecommandidempotency.CommandKey, versionID uint64) (Request, error) {
	if key == "" || versionID == 0 {
		return Request{}, ErrInvalidRequest
	}
	return Request{key: key, versionID: versionID}, nil
}

// NewTrackedStartRequest constructs a parent request over one exact live Start record.
func NewTrackedStartRequest(key runtimecommandidempotency.CommandKey, versionID uint64, trackedKey runtimecommandidempotency.CommandKey, trackedRevision runtimecommandidempotency.Revision) (Request, error) {
	request, err := NewRequest(key, versionID)
	if err != nil || trackedKey == "" || trackedRevision == 0 {
		return Request{}, ErrInvalidRequest
	}
	request.trackedStartKey, request.trackedRevision = trackedKey, trackedRevision
	return request, nil
}

// ResultCategory is the closed caller-visible orchestration result set.
type ResultCategory string

const (
	// ResultSucceeded reports a definitive exact activation or replacement.
	ResultSucceeded ResultCategory = "succeeded"
	// ResultSatisfied reports an exact Running target with zero lifecycle mutation.
	ResultSatisfied ResultCategory = "satisfied"
	// ResultStopped reports a definitive Stop winner before a replacement Start.
	ResultStopped ResultCategory = "stopped"
	// ResultCancelled reports definitive cancellation with no further lifecycle work.
	ResultCancelled ResultCategory = "cancelled"
	// ResultRejected reports a definitive no-mutation rejection.
	ResultRejected ResultCategory = "rejected"
	// ResultFailed reports a definitive lifecycle failure.
	ResultFailed ResultCategory = "failed"
	// ResultInProgress reports a non-mutating conflicting live operation.
	ResultInProgress ResultCategory = "in-progress"
	// ResultUnresolved reports a claimed operation lacking terminal truth.
	ResultUnresolved ResultCategory = "unresolved"
)

// Result is a detached redacted semantic result.
type Result struct {
	category ResultCategory
}

// Category returns the stable result category.
func (r Result) Category() ResultCategory { return r.category }

// Orchestrator is one immutable exact Runtime Instance composition.
type Orchestrator struct {
	domain     string
	target     runtimemanagement.Target
	versions   versionStore
	identity   identityStore
	owner      lifecycleOwner
	commands   *runtimecommandidempotency.Boundary
	start      managedStartInvoker
	authorize  runtimeorchestrationbinding.AuthorizeOrchestration
	generation runtimecommandidempotency.ProvideExecutionGeneration
}

// New constructs one exact-scope activation composition without wiring it globally.
func New(domain string, target runtimemanagement.Target, versions versionStore, identity identityStore, owner lifecycleOwner, commands *runtimecommandidempotency.Boundary, start managedStartInvoker, authorize runtimeorchestrationbinding.AuthorizeOrchestration, generation runtimecommandidempotency.ProvideExecutionGeneration) (*Orchestrator, error) {
	if domain == "" || target.WorkspaceID() == 0 || target.ConfigurationID() == 0 || target.RuntimeInstanceID() == "" || versions == nil || identity == nil || owner == nil || commands == nil || start == nil || authorize == nil || generation == nil {
		return nil, ErrInvalidRequest
	}
	if observation := owner.Observe(); observation.WorkspaceID() != target.WorkspaceID() || observation.ConfigurationID() != target.ConfigurationID() || observation.RuntimeInstanceID() != target.RuntimeInstanceID() {
		return nil, ErrInvalidRequest
	}
	return &Orchestrator{domain: domain, target: target, versions: versions, identity: identity, owner: owner, commands: commands, start: start, authorize: authorize, generation: generation}, nil
}

// ActivateExact activates one exact Published target or observes satisfaction/progress.
func (o *Orchestrator) ActivateExact(ctx context.Context, request Request) (Result, error) {
	scope, intent, err := o.command(request, runtimecommandidempotency.OperationStart)
	if err != nil || ctx == nil || request.trackedStartKey != "" {
		return Result{}, ErrInvalidRequest
	}
	var selected observed
	admission, disposition, err := o.commands.ExecuteReplayFirstManagedStart(ctx, scope, request.key, intent, o.authorize,
		func(context.Context) (runtimecommandidempotency.AbsentCandidate, error) {
			var decideErr error
			selected, decideErr = o.decide(request)
			if decideErr != nil {
				return runtimecommandidempotency.AbsentCandidate{}, decideErr
			}
			return selected.primitiveCandidate(request.versionID)
		}, o.revalidate(request.versionID), o.generation,
		func(binding runtimeorchestrationbinding.StartExecutionBinding) (runtimecommandidempotency.TerminalOutcome, error) {
			return o.invokeStart(ctx, request.versionID, binding)
		})
	if disposition == runtimecommandidempotency.ReplayFirstNoClaim {
		return Result{category: ResultInProgress}, nil
	}
	return terminalPrimitive(admission, err)
}

// ReplaceExact replaces any different active target with one exact Published target.
func (o *Orchestrator) ReplaceExact(ctx context.Context, request Request) (Result, error) {
	return o.executeParent(ctx, request, runtimecommandidempotency.OperationReplace)
}

// RollbackExact replaces the current execution with one caller-selected exact Published target.
func (o *Orchestrator) RollbackExact(ctx context.Context, request Request) (Result, error) {
	return o.executeParent(ctx, request, runtimecommandidempotency.OperationRollback)
}

type observed struct {
	view      runtimeidentity.RuntimeInstanceView
	attempt   runtimeidentity.LaunchAttemptRecord
	hasActive bool
	starting  bool
	running   bool
	satisfied bool
}

func (o *Orchestrator) command(request Request, operation runtimecommandidempotency.Operation) (runtimecommandidempotency.Scope, runtimecommandidempotency.Intent, error) {
	if o == nil || request.key == "" || request.versionID == 0 {
		return runtimecommandidempotency.Scope{}, runtimecommandidempotency.Intent{}, ErrInvalidRequest
	}
	scope, err := runtimecommandidempotency.NewScope(o.domain, o.target.WorkspaceID(), o.target.ConfigurationID(), o.target.RuntimeInstanceID(), operation)
	if err != nil {
		return scope, runtimecommandidempotency.Intent{}, err
	}
	if operation == runtimecommandidempotency.OperationStart {
		intent, e := runtimecommandidempotency.NewStartIntent(request.versionID)
		return scope, intent, e
	}
	if operation == runtimecommandidempotency.OperationReplace {
		intent, e := runtimecommandidempotency.NewReplaceIntent(request.versionID)
		return scope, intent, e
	}
	intent, err := runtimecommandidempotency.NewRollbackIntent(request.versionID)
	return scope, intent, err
}
func (o *Orchestrator) validateVersion(versionID uint64) error {
	version, err := o.versions.Get(versionID)
	if err != nil {
		return err
	}
	if version.ID != versionID || version.ConfigurationID != o.target.ConfigurationID() || version.State != configurationversion.Published {
		return ErrVersionNotPublished
	}
	return nil
}
func (o *Orchestrator) decide(request Request) (observed, error) {
	if err := o.validateVersion(request.versionID); err != nil {
		return observed{}, err
	}
	facts, err := o.observe()
	if err != nil {
		return observed{}, err
	}
	if facts.running && facts.attempt.ConfigurationVersionID() == request.versionID {
		facts.satisfied = true
	}
	return facts, nil
}
func (o *Orchestrator) observe() (observed, error) {
	id := o.target.RuntimeInstanceID()
	a, err := o.identity.ReadRuntimeInstance(id)
	if err != nil {
		return observed{}, err
	}
	history, err := o.identity.ReadLaunchAttemptHistory(id)
	if err != nil {
		return observed{}, err
	}
	owner := o.owner.Observe()
	b, err := o.identity.ReadRuntimeInstance(id)
	if err != nil || a != b || a.WorkspaceID() != o.target.WorkspaceID() || a.ConfigurationID() != o.target.ConfigurationID() || a.RuntimeInstanceID() != id || owner.WorkspaceID() != a.WorkspaceID() || owner.ConfigurationID() != a.ConfigurationID() || owner.RuntimeInstanceID() != id {
		return observed{}, errIncoherentState
	}
	result := observed{view: a}
	activeID, active := a.ActiveAttempt()
	ownerAttempt, ownerActive := owner.ActiveAttempt()
	if !active {
		if ownerActive || !((a.ActualState() == runtimeidentity.ActualStateStopped && owner.ActualState() == runtimelifecycle.ActualStopped) || (a.ActualState() == runtimeidentity.ActualStateFailed && owner.ActualState() == runtimelifecycle.ActualFailed)) {
			return observed{}, errIncoherentState
		}
		return result, nil
	}
	for _, attempt := range history {
		if attempt.LaunchAttemptID() == activeID {
			result.attempt, result.hasActive = attempt, true
		}
	}
	if !result.hasActive || !ownerActive || ownerAttempt.LaunchAttemptID() != activeID || ownerAttempt.ConfigurationVersionID() != result.attempt.ConfigurationVersionID() {
		return observed{}, errIncoherentState
	}
	result.starting = a.ActualState() == runtimeidentity.ActualStateClaimed && result.attempt.Phase() == runtimeidentity.AttemptPhaseClaimed && owner.ActualState() == runtimelifecycle.ActualStarting
	result.running = a.ActualState() == runtimeidentity.ActualStateRunning && result.attempt.Phase() == runtimeidentity.AttemptPhaseRunning && owner.ActualState() == runtimelifecycle.ActualRunning && ownerAttempt.RunningPublished()
	if !result.starting && !result.running {
		return observed{}, errIncoherentState
	}
	return result, nil
}
func (f observed) primitiveCandidate(versionID uint64) (runtimecommandidempotency.AbsentCandidate, error) {
	revision := runtimeorchestrationbinding.AggregateRevision(f.view.Revision())
	if f.satisfied {
		return runtimecommandidempotency.NewSatisfiedCandidate(revision, f.attempt.LaunchAttemptID(), versionID)
	}
	if f.hasActive {
		return runtimecommandidempotency.NewNoClaimCandidate(), nil
	}
	return runtimecommandidempotency.NewExecutePrimitiveCandidate(revision)
}
func (o *Orchestrator) revalidate(versionID uint64) runtimecommandidempotency.RevalidateCandidate {
	return func(_ context.Context, candidate runtimecommandidempotency.AbsentCandidate) (runtimecommandidempotency.CandidateRevalidation, error) {
		if err := o.validateVersion(versionID); err != nil {
			return runtimecommandidempotency.CandidateUnresolved, err
		}
		facts, err := o.observe()
		if err == nil && facts.running && facts.view.Revision() == runtimeidentity.Revision(candidate.ExpectedAggregateRevision()) && facts.attempt.LaunchAttemptID() == candidate.LaunchAttemptID() && facts.attempt.ConfigurationVersionID() == versionID {
			return runtimecommandidempotency.CandidateRevalidated, nil
		}
		return runtimecommandidempotency.CandidateUnresolved, err
	}
}
func (o *Orchestrator) invokeStart(ctx context.Context, versionID uint64, binding runtimeorchestrationbinding.StartExecutionBinding) (runtimecommandidempotency.TerminalOutcome, error) {
	outcome, err := o.start.InvokeManagedStart(ctx, runtimelifecycle.NewStartRequest(o.target.WorkspaceID(), o.target.ConfigurationID(), versionID), binding)
	if err != nil {
		category := runtimecommandidempotency.OutcomeCategory("")
		if errors.Is(err, context.Canceled) || errors.Is(err, runtimelifecycle.ErrStartConflict) {
			category = runtimecommandidempotency.OutcomeRejected
		} else if errors.Is(err, runtimelifecycle.ErrAttemptIDSourceFailed) {
			category = runtimecommandidempotency.OutcomeFailed
		}
		if category == "" {
			return runtimecommandidempotency.TerminalOutcome{}, err
		}
		terminal, _ := runtimecommandidempotency.NewTerminalOutcome(category, "")
		return terminal, nil
	}
	if err := o.publishStart(outcome, versionID); err != nil {
		return runtimecommandidempotency.TerminalOutcome{}, err
	}
	category := runtimecommandidempotency.OutcomeSucceeded
	if outcome.Kind() == runtimelifecycle.StartPreparationFailed || outcome.Kind() == runtimelifecycle.StartLaunchFailed {
		category = runtimecommandidempotency.OutcomeFailed
	} else if outcome.Kind() == runtimelifecycle.StartStoppedBeforeRunning {
		category = runtimecommandidempotency.OutcomeRejected
	} else if outcome.Kind() != runtimelifecycle.StartRunning {
		return runtimecommandidempotency.TerminalOutcome{}, errIncoherentState
	}
	return runtimecommandidempotency.NewTerminalOutcome(category, outcome.Attempt().LaunchAttemptID())
}
func (o *Orchestrator) publishStart(outcome runtimelifecycle.StartOutcome, versionID uint64) error {
	fact := outcome.Attempt()
	if fact.WorkspaceID() != o.target.WorkspaceID() || fact.ConfigurationID() != o.target.ConfigurationID() || fact.RuntimeInstanceID() != o.target.RuntimeInstanceID() || fact.ConfigurationVersionID() != versionID || fact.LaunchAttemptID() == "" {
		return errIncoherentState
	}
	view, err := o.identity.ReadRuntimeInstance(o.target.RuntimeInstanceID())
	if err != nil {
		return err
	}
	active, ok := view.ActiveAttempt()
	if !ok || active != fact.LaunchAttemptID() {
		if outcome.Kind() == runtimelifecycle.StartStoppedBeforeRunning && view.ActualState() == runtimeidentity.ActualStateStopped {
			history, historyErr := o.identity.ReadLaunchAttemptHistory(o.target.RuntimeInstanceID())
			for _, attempt := range history {
				if historyErr == nil && attempt.LaunchAttemptID() == fact.LaunchAttemptID() && attempt.ConfigurationVersionID() == versionID && attempt.Phase() == runtimeidentity.AttemptPhaseStopped {
					return nil
				}
			}
		}
		return errIncoherentState
	}
	var published runtimeidentity.PublishResult
	switch outcome.Kind() {
	case runtimelifecycle.StartRunning:
		published, err = o.identity.ConditionalPublishRunning(o.target.RuntimeInstanceID(), view.Revision(), active)
	case runtimelifecycle.StartStoppedBeforeRunning:
		published, err = o.identity.ConditionalPublishTerminal(o.target.RuntimeInstanceID(), view.Revision(), active, true)
	case runtimelifecycle.StartPreparationFailed, runtimelifecycle.StartLaunchFailed:
		published, err = o.identity.ConditionalPublishTerminal(o.target.RuntimeInstanceID(), view.Revision(), active, false)
	default:
		return errIncoherentState
	}
	if err != nil || !published.Committed() {
		return errors.Join(errIncoherentState, err)
	}
	return nil
}
func terminalPrimitive(admission runtimecommandidempotency.Admission, callErr error) (Result, error) {
	if admission.Kind() == "" {
		return Result{}, callErr
	}
	if admission.Kind() == runtimecommandidempotency.AdmissionInProgress {
		return Result{category: ResultInProgress}, callErr
	}
	outcome, ok := admission.Record().Outcome()
	if !ok {
		return Result{category: ResultUnresolved}, callErr
	}
	category := map[runtimecommandidempotency.OutcomeCategory]ResultCategory{runtimecommandidempotency.OutcomeSucceeded: ResultSucceeded, runtimecommandidempotency.OutcomeSatisfied: ResultSatisfied, runtimecommandidempotency.OutcomeRejected: ResultRejected, runtimecommandidempotency.OutcomeFailed: ResultFailed}[outcome.Category()]
	return Result{category: category}, callErr
}
func (o *Orchestrator) executeParent(ctx context.Context, request Request, operation runtimecommandidempotency.Operation) (Result, error) {
	scope, intent, err := o.command(request, operation)
	if err != nil || ctx == nil || (request.trackedStartKey == "") != (request.trackedRevision == 0) {
		return Result{}, ErrInvalidRequest
	}
	var selected observed
	admission, disposition, err := o.commands.ExecuteReplayFirstManagedParent(ctx, scope, request.key, intent, o.authorize,
		func(context.Context) (runtimecommandidempotency.AbsentCandidate, error) {
			var decideErr error
			selected, decideErr = o.decide(request)
			if decideErr != nil {
				return runtimecommandidempotency.AbsentCandidate{}, decideErr
			}
			return o.parentCandidate(request, selected)
		}, o.revalidate(request.versionID), o.generation,
		func(execution *runtimecommandidempotency.ReplayFirstManagedParentExecution) error {
			return o.runParent(ctx, request.versionID, selected, execution)
		})
	if disposition == runtimecommandidempotency.ReplayFirstNoClaim {
		return Result{category: ResultInProgress}, nil
	}
	return terminalParent(admission, err)
}
func (o *Orchestrator) parentCandidate(request Request, facts observed) (runtimecommandidempotency.AbsentCandidate, error) {
	revision := runtimeorchestrationbinding.AggregateRevision(facts.view.Revision())
	if facts.satisfied {
		return runtimecommandidempotency.NewSatisfiedCandidate(revision, facts.attempt.LaunchAttemptID(), request.versionID)
	}
	if facts.starting {
		if facts.attempt.ConfigurationVersionID() == request.versionID || request.trackedStartKey == "" {
			return runtimecommandidempotency.NewNoClaimCandidate(), nil
		}
		startScope, _ := runtimecommandidempotency.NewScope(o.domain, o.target.WorkspaceID(), o.target.ConfigurationID(), o.target.RuntimeInstanceID(), runtimecommandidempotency.OperationStart)
		return runtimecommandidempotency.NewExecuteParentFromTrackedStartCandidate(revision+1, startScope, request.trackedStartKey, request.trackedRevision)
	}
	if facts.running {
		revision += 2
	}
	return runtimecommandidempotency.NewExecuteParentCandidate(revision)
}
func (o *Orchestrator) runParent(ctx context.Context, versionID uint64, facts observed, execution *runtimecommandidempotency.ReplayFirstManagedParentExecution) error {
	if facts.hasActive {
		invoke := func() (runtimecommandidempotency.TerminalOutcome, error) { return o.stopOld(ctx, facts) }
		var phase runtimecommandidempotency.PhaseAdmission
		var err error
		if facts.starting {
			phase, err = execution.ExecutePreclaimedStopOld(invoke)
		} else {
			phase, err = execution.InspectOrExecuteStopOld(invoke)
		}
		if err != nil {
			return err
		}
		outcome, ok := phase.Record().Outcome()
		if !ok {
			return runtimecommandidempotency.ErrIndeterminateExecution
		}
		if outcome.Category() != runtimecommandidempotency.OutcomeSucceeded {
			return publishParent(execution, runtimecommandidempotency.ParentOutcomeFailed)
		}
	}
	startCancelled := false
	phase, prevented, err := execution.ContinueOrExecuteManagedStartTarget(ctx,
		func(binding runtimeorchestrationbinding.StartExecutionBinding) (runtimecommandidempotency.TerminalOutcome, error) {
			outcome, invokeErr := o.invokeStart(ctx, versionID, binding)
			startCancelled = invokeErr == nil && ctx.Err() != nil && outcome.Category() == runtimecommandidempotency.OutcomeRejected && outcome.LaunchAttemptID() == ""
			return outcome, invokeErr
		})
	if err != nil {
		if prevented {
			category := runtimecommandidempotency.ParentOutcomeStopped
			if errors.Is(err, context.Canceled) {
				category = runtimecommandidempotency.ParentOutcomeCancelled
			}
			return publishParent(execution, category)
		}
		return err
	}
	if prevented {
		return publishParent(execution, runtimecommandidempotency.ParentOutcomeStopped)
	}
	outcome, ok := phase.Record().Outcome()
	if !ok {
		return runtimecommandidempotency.ErrIndeterminateExecution
	}
	category := runtimecommandidempotency.ParentOutcomeSucceeded
	if startCancelled {
		category = runtimecommandidempotency.ParentOutcomeCancelled
	} else if outcome.Category() == runtimecommandidempotency.OutcomeFailed {
		category = runtimecommandidempotency.ParentOutcomeFailed
	} else if outcome.Category() == runtimecommandidempotency.OutcomeRejected {
		category = runtimecommandidempotency.ParentOutcomeStopped
	}
	return publishParent(execution, category)
}
func (o *Orchestrator) stopOld(ctx context.Context, facts observed) (runtimecommandidempotency.TerminalOutcome, error) {
	revision := facts.view.Revision()
	if facts.running {
		claim, err := o.identity.ConditionalClaimStop(o.target.RuntimeInstanceID(), revision, facts.attempt.LaunchAttemptID())
		if err != nil || !claim.Committed() {
			return runtimecommandidempotency.TerminalOutcome{}, errors.Join(errIncoherentState, err)
		}
		revision = claim.Revision()
	}
	stopped, err := o.owner.StopExpectedAttempt(context.WithoutCancel(ctx), facts.attempt.LaunchAttemptID())
	if err != nil || stopped.Kind() == runtimelifecycle.StopAttemptMismatch {
		return runtimecommandidempotency.TerminalOutcome{}, errors.Join(errIncoherentState, err)
	}
	if attempt, ok := stopped.Attempt(); ok && attempt.LaunchAttemptID() != facts.attempt.LaunchAttemptID() {
		return runtimecommandidempotency.TerminalOutcome{}, errIncoherentState
	}
	if stopped.Kind() == runtimelifecycle.StopFailed && facts.starting {
		return runtimecommandidempotency.TerminalOutcome{}, errIncoherentState
	}
	published, err := o.identity.ConditionalPublishTerminal(o.target.RuntimeInstanceID(), revision, facts.attempt.LaunchAttemptID(), stopped.Kind() == runtimelifecycle.StopStopped)
	if err != nil || !published.Committed() {
		return runtimecommandidempotency.TerminalOutcome{}, errors.Join(errIncoherentState, err)
	}
	category := runtimecommandidempotency.OutcomeSucceeded
	if stopped.Kind() == runtimelifecycle.StopFailed {
		category = runtimecommandidempotency.OutcomeFailed
	}
	return runtimecommandidempotency.NewTerminalOutcome(category, facts.attempt.LaunchAttemptID())
}
func publishParent(execution *runtimecommandidempotency.ReplayFirstManagedParentExecution, category runtimecommandidempotency.ParentOutcomeCategory) error {
	outcome, err := runtimecommandidempotency.NewParentTerminalOutcome(category)
	if err != nil {
		return err
	}
	_, err = execution.PublishTerminal(outcome)
	return err
}
func terminalParent(admission runtimecommandidempotency.ParentAdmission, callErr error) (Result, error) {
	if admission.Kind() == "" {
		return Result{}, callErr
	}
	if admission.Kind() == runtimecommandidempotency.AdmissionInProgress {
		return Result{category: ResultInProgress}, callErr
	}
	outcome, ok := admission.Record().Outcome()
	if !ok {
		return Result{category: ResultUnresolved}, callErr
	}
	categories := map[runtimecommandidempotency.ParentOutcomeCategory]ResultCategory{
		runtimecommandidempotency.ParentOutcomeSucceeded: ResultSucceeded, runtimecommandidempotency.ParentOutcomeSatisfied: ResultSatisfied,
		runtimecommandidempotency.ParentOutcomeStopped: ResultStopped, runtimecommandidempotency.ParentOutcomeCancelled: ResultCancelled,
		runtimecommandidempotency.ParentOutcomeRejected: ResultRejected, runtimecommandidempotency.ParentOutcomeFailed: ResultFailed,
	}
	return Result{category: categories[outcome.Category()]}, callErr
}
