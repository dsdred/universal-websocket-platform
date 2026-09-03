package runtimecommandidempotency

import (
	"context"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

// AbsentCandidateKind is the closed absent-identity decision set.
type AbsentCandidateKind uint8

const (
	// CandidateNoClaim is a definitive read-only observation that claims nothing.
	CandidateNoClaim AbsentCandidateKind = iota + 1
	// CandidateSatisfied claims before exact aggregate-fact revalidation.
	CandidateSatisfied
	// CandidateExecutePrimitive admits one primitive managed Start.
	CandidateExecutePrimitive
	// CandidateExecuteParent admits one managed Replace or Rollback parent.
	CandidateExecuteParent
)

// ParentCandidateMode distinguishes ordinary parent admission from the
// tracked-Start admission that atomically preclaims StopOld.
type ParentCandidateMode uint8

const (
	// ParentCandidateOrdinary requires an otherwise clear Instance ledger.
	ParentCandidateOrdinary ParentCandidateMode = iota + 1
	// ParentCandidateTrackedStart requires one exact live tracked Start.
	ParentCandidateTrackedStart
)

// AbsentCandidate is an immutable, authority-free absent-identity decision.
// It contains no execution generation, permit, rendezvous, or terminal truth.
type AbsentCandidate struct {
	kind                      AbsentCandidateKind
	expectedAggregateRevision runtimeorchestrationbinding.AggregateRevision
	launchAttemptID           runtimeconfigload.LaunchAttemptID
	configurationVersionID    uint64
	parentMode                ParentCandidateMode
	trackedStartScope         Scope
	trackedStartKey           CommandKey
	trackedStartRevision      Revision
}

// NewNoClaimCandidate constructs a definitive no-claim decision.
func NewNoClaimCandidate() AbsentCandidate {
	return AbsentCandidate{kind: CandidateNoClaim}
}

// NewSatisfiedCandidate constructs exact facts that must be revalidated only
// after the durable command or parent claim is committed.
func NewSatisfiedCandidate(
	expectedAggregateRevision runtimeorchestrationbinding.AggregateRevision,
	launchAttemptID runtimeconfigload.LaunchAttemptID,
	configurationVersionID uint64,
) (AbsentCandidate, error) {
	candidate := AbsentCandidate{
		kind: CandidateSatisfied, expectedAggregateRevision: expectedAggregateRevision,
		launchAttemptID: launchAttemptID, configurationVersionID: configurationVersionID,
	}
	if expectedAggregateRevision == 0 || launchAttemptID == "" || configurationVersionID == 0 {
		return AbsentCandidate{}, ErrInvalidSubmission
	}
	return candidate, nil
}

// NewExecutePrimitiveCandidate constructs a primitive execution decision.
func NewExecutePrimitiveCandidate(
	expectedAggregateRevision runtimeorchestrationbinding.AggregateRevision,
) (AbsentCandidate, error) {
	if expectedAggregateRevision == 0 {
		return AbsentCandidate{}, ErrInvalidSubmission
	}
	return AbsentCandidate{
		kind: CandidateExecutePrimitive, expectedAggregateRevision: expectedAggregateRevision,
	}, nil
}

// NewExecuteParentCandidate constructs an ordinary managed-parent decision.
func NewExecuteParentCandidate(
	expectedAggregateRevision runtimeorchestrationbinding.AggregateRevision,
) (AbsentCandidate, error) {
	if expectedAggregateRevision == 0 {
		return AbsentCandidate{}, ErrInvalidSubmission
	}
	return AbsentCandidate{
		kind: CandidateExecuteParent, expectedAggregateRevision: expectedAggregateRevision,
		parentMode: ParentCandidateOrdinary,
	}, nil
}

// NewExecuteParentFromTrackedStartCandidate constructs an exact tracked-Start
// parent decision. The Start identity and revision are rechecked atomically
// with the parent plus preclaimed StopOld transition.
func NewExecuteParentFromTrackedStartCandidate(
	expectedAggregateRevision runtimeorchestrationbinding.AggregateRevision,
	trackedStartScope Scope,
	trackedStartKey CommandKey,
	trackedStartRevision Revision,
) (AbsentCandidate, error) {
	if expectedAggregateRevision == 0 || !trackedStartScope.validPrimitive() ||
		trackedStartScope.operation != OperationStart || trackedStartKey == "" || trackedStartRevision == 0 {
		return AbsentCandidate{}, ErrInvalidSubmission
	}
	return AbsentCandidate{
		kind: CandidateExecuteParent, expectedAggregateRevision: expectedAggregateRevision,
		parentMode: ParentCandidateTrackedStart, trackedStartScope: trackedStartScope,
		trackedStartKey: trackedStartKey, trackedStartRevision: trackedStartRevision,
	}, nil
}

// Kind returns the closed candidate discriminator.
func (c AbsentCandidate) Kind() AbsentCandidateKind { return c.kind }

// ExpectedAggregateRevision returns the exact read-only aggregate fact.
func (c AbsentCandidate) ExpectedAggregateRevision() runtimeorchestrationbinding.AggregateRevision {
	return c.expectedAggregateRevision
}

// LaunchAttemptID returns the exact read-only satisfied-attempt fact.
func (c AbsentCandidate) LaunchAttemptID() runtimeconfigload.LaunchAttemptID {
	return c.launchAttemptID
}

// ConfigurationVersionID returns the exact read-only satisfied-version fact.
func (c AbsentCandidate) ConfigurationVersionID() uint64 { return c.configurationVersionID }

func (c AbsentCandidate) validFor(scope Scope, intent Intent) bool {
	switch c.kind {
	case CandidateNoClaim:
		return c == (AbsentCandidate{}) || c == NewNoClaimCandidate()
	case CandidateSatisfied:
		return c.expectedAggregateRevision != 0 && c.launchAttemptID != "" &&
			c.configurationVersionID == intent.configurationVersionID
	case CandidateExecutePrimitive:
		return scope.operation == OperationStart && c.expectedAggregateRevision != 0 &&
			c.launchAttemptID == "" && c.configurationVersionID == 0 && c.parentMode == 0
	case CandidateExecuteParent:
		if !scope.validParent() || c.expectedAggregateRevision == 0 {
			return false
		}
		if c.parentMode == ParentCandidateOrdinary {
			return c.trackedStartScope == (Scope{}) && c.trackedStartKey == "" && c.trackedStartRevision == 0
		}
		return c.parentMode == ParentCandidateTrackedStart &&
			c.trackedStartScope.operation == OperationStart &&
			c.trackedStartScope.instanceScope() == scope.instanceScope() &&
			c.trackedStartKey != "" && c.trackedStartRevision != 0
	default:
		return false
	}
}

// DecideAbsent is a synchronous read-only orchestration decision callback.
type DecideAbsent func(context.Context) (AbsentCandidate, error)

// CandidateRevalidation is the closed satisfied-fact revalidation result.
type CandidateRevalidation uint8

const (
	// CandidateRevalidated confirms every exact satisfied fact.
	CandidateRevalidated CandidateRevalidation = iota + 1
	// CandidateUnresolved represents stale, unavailable, or ambiguous facts.
	CandidateUnresolved
)

// RevalidateCandidate synchronously revalidates a claimed satisfied candidate.
type RevalidateCandidate func(context.Context, AbsentCandidate) (CandidateRevalidation, error)

// ProvideExecutionGeneration supplies one non-empty composition-owned
// generation after the execution claim wins.
type ProvideExecutionGeneration func(context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error)

// ReplayFirstDisposition distinguishes a durable admission from definitive no-claim.
type ReplayFirstDisposition uint8

const (
	// ReplayFirstNoClaim means no durable record was created.
	ReplayFirstNoClaim ReplayFirstDisposition = iota + 1
	// ReplayFirstAdmitted returns an existing or newly claimed record.
	ReplayFirstAdmitted
)

// ExecuteReplayFirstManagedStart performs authorized replay-first primitive
// admission and allocates generation only for the atomic absent winner.
func (b *Boundary) ExecuteReplayFirstManagedStart(
	ctx context.Context,
	scope Scope,
	key CommandKey,
	intent Intent,
	authorize AuthorizeOrchestration,
	decideAbsent DecideAbsent,
	revalidate RevalidateCandidate,
	provideGeneration ProvideExecutionGeneration,
	invoke func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error),
) (Admission, ReplayFirstDisposition, error) {
	if b == nil || b.storage == nil || ctx == nil || scope.operation != OperationStart ||
		!scope.validPrimitive() || key == "" || !intent.validFor(scope) || authorize == nil ||
		decideAbsent == nil || revalidate == nil || provideGeneration == nil || invoke == nil {
		return Admission{}, 0, ErrInvalidSubmission
	}
	if err := authorizeOrchestration(authorize, ctx, scope, intent); err != nil {
		return Admission{}, 0, err
	}
	if err := ctx.Err(); err != nil {
		return Admission{}, 0, err
	}
	if admission, found, err := b.inspectPrimitive(scope, key, intent); found || err != nil {
		return admission, ReplayFirstAdmitted, err
	}
	candidate, err := decideAbsentSafely(decideAbsent, ctx)
	if err != nil || !candidate.validFor(scope, intent) {
		if err == nil {
			err = ErrInvalidSubmission
		}
		return Admission{}, 0, err
	}
	switch candidate.kind {
	case CandidateNoClaim:
		return Admission{}, ReplayFirstNoClaim, nil
	case CandidateSatisfied:
		admission, claimErr := b.Execute(ctx, scope, key, intent, allowPrimitive,
			func() (TerminalOutcome, error) {
				if !revalidateSatisfiedSafely(revalidate, ctx, candidate) {
					return TerminalOutcome{}, ErrIndeterminateExecution
				}
				return NewTerminalOutcome(OutcomeSucceeded, candidate.launchAttemptID)
			})
		return admission, ReplayFirstAdmitted, claimErr
	case CandidateExecutePrimitive:
		admission, claimErr := b.executeLateManagedStart(
			ctx, scope, key, intent, candidate.expectedAggregateRevision,
			allowOrchestrationInternal, provideGeneration, invoke,
		)
		return admission, ReplayFirstAdmitted, claimErr
	default:
		return Admission{}, 0, ErrInvalidSubmission
	}
}

// ReplayFirstManagedParentExecution is callback-scoped managed-parent
// authority. Its generation provider remains private and can be consumed only
// by a newly claimed StartTarget phase.
type ReplayFirstManagedParentExecution struct {
	managed           *ManagedParentExecution
	tracked           *TrackedStartManagedParentExecution
	expectedRevision  runtimeorchestrationbinding.AggregateRevision
	provideGeneration ProvideExecutionGeneration
}

// InspectOrExecuteStopOld preserves the ordinary optional StopOld path.
func (p *ReplayFirstManagedParentExecution) InspectOrExecuteStopOld(
	invoke func() (TerminalOutcome, error),
) (PhaseAdmission, error) {
	if p == nil || p.managed == nil || p.tracked != nil {
		return PhaseAdmission{}, ErrInvalidSubmission
	}
	return p.managed.InspectOrExecuteStopOld(invoke)
}

// ExecutePreclaimedStopOld consumes the tracked-Start candidate's preclaim.
func (p *ReplayFirstManagedParentExecution) ExecutePreclaimedStopOld(
	invoke func() (TerminalOutcome, error),
) (PhaseAdmission, error) {
	if p == nil || p.tracked == nil {
		return PhaseAdmission{}, ErrInvalidSubmission
	}
	return p.tracked.ExecutePreclaimedStopOld(invoke)
}

// ContinueOrExecuteManagedStartTarget requests generation only after the
// StartTarget phase claim wins, then installs its binding before invocation.
func (p *ReplayFirstManagedParentExecution) ContinueOrExecuteManagedStartTarget(
	ctx context.Context,
	invoke func(runtimeorchestrationbinding.StartExecutionBinding) (TerminalOutcome, error),
) (PhaseAdmission, bool, error) {
	if p == nil || p.managed == nil || p.expectedRevision == 0 || p.provideGeneration == nil {
		return PhaseAdmission{}, false, ErrInvalidSubmission
	}
	return p.managed.continueOrExecuteLateManagedStartTarget(
		ctx, p.expectedRevision, p.provideGeneration, invoke,
	)
}

// PublishTerminal preserves managed parent terminal publication.
func (p *ReplayFirstManagedParentExecution) PublishTerminal(
	outcome ParentTerminalOutcome,
) (ParentRecordView, error) {
	if p == nil || p.managed == nil {
		return ParentRecordView{}, ErrInvalidSubmission
	}
	return p.managed.PublishTerminal(outcome)
}

// ExecuteReplayFirstManagedParent performs authorized replay-first parent
// admission. StartTarget generation stays late and callback-scoped.
func (b *Boundary) ExecuteReplayFirstManagedParent(
	ctx context.Context,
	scope Scope,
	key CommandKey,
	intent Intent,
	authorize AuthorizeOrchestration,
	decideAbsent DecideAbsent,
	revalidate RevalidateCandidate,
	provideGeneration ProvideExecutionGeneration,
	invoke func(*ReplayFirstManagedParentExecution) error,
) (ParentAdmission, ReplayFirstDisposition, error) {
	if b == nil || b.storage == nil || ctx == nil || !scope.validParent() || key == "" ||
		!intent.validParentFor(scope) || authorize == nil || decideAbsent == nil ||
		revalidate == nil || provideGeneration == nil || invoke == nil {
		return ParentAdmission{}, 0, ErrInvalidSubmission
	}
	if err := authorizeOrchestration(authorize, ctx, scope, intent); err != nil {
		return ParentAdmission{}, 0, err
	}
	if err := ctx.Err(); err != nil {
		return ParentAdmission{}, 0, err
	}
	if admission, found, err := b.inspectParent(scope, key, intent); found || err != nil {
		return admission, ReplayFirstAdmitted, err
	}
	candidate, err := decideAbsentSafely(decideAbsent, ctx)
	if err != nil || !candidate.validFor(scope, intent) {
		if err == nil {
			err = ErrInvalidSubmission
		}
		return ParentAdmission{}, 0, err
	}
	if candidate.kind == CandidateNoClaim {
		return ParentAdmission{}, ReplayFirstNoClaim, nil
	}
	if candidate.kind == CandidateSatisfied {
		admission, claimErr := b.ExecuteParent(ctx, scope, key, intent, allowPrimitive,
			func(execution *ParentExecution) error {
				if !revalidateSatisfiedSafely(revalidate, ctx, candidate) {
					return ErrIndeterminateExecution
				}
				_, publishErr := execution.PublishTerminal(mustParentSatisfied())
				return publishErr
			})
		return admission, ReplayFirstAdmitted, claimErr
	}
	if candidate.kind != CandidateExecuteParent {
		return ParentAdmission{}, 0, ErrInvalidSubmission
	}
	wrap := func(managed *ManagedParentExecution, tracked *TrackedStartManagedParentExecution) error {
		return invoke(&ReplayFirstManagedParentExecution{
			managed: managed, tracked: tracked, expectedRevision: candidate.expectedAggregateRevision,
			provideGeneration: provideGeneration,
		})
	}
	var admission ParentAdmission
	if candidate.parentMode == ParentCandidateTrackedStart {
		admission, err = b.executeManagedParentFromTrackedStartCandidate(
			ctx, scope, key, intent, allowOrchestrationInternal, candidate,
			func(execution *TrackedStartManagedParentExecution) error {
				return wrap(execution.managed, execution)
			},
		)
	} else {
		admission, err = b.ExecuteManagedParent(
			ctx, scope, key, intent, allowOrchestrationInternal,
			func(execution *ManagedParentExecution) error { return wrap(execution, nil) },
		)
	}
	return admission, ReplayFirstAdmitted, err
}

func (b *Boundary) inspectPrimitive(scope Scope, key CommandKey, intent Intent) (Admission, bool, error) {
	b.storage.clientMu.RLock()
	defer b.storage.clientMu.RUnlock()
	if b.storage.generation != b.generation {
		return Admission{}, false, ErrBoundaryExpired
	}
	ledger := b.storage.existingLedger(scope.instanceScope())
	if ledger == nil {
		return Admission{}, false, nil
	}
	ledger.mu.Lock()
	defer ledger.mu.Unlock()
	record := ledger.records[commandIdentity{scope: scope, key: key}]
	if record == nil {
		return Admission{}, false, nil
	}
	if record.intent != intent {
		return Admission{}, true, ErrCommandKeyConflict
	}
	kind := AdmissionInProgress
	if record.state == CommandStateTerminal {
		kind = AdmissionReplay
	}
	return Admission{kind: kind, record: record.view()}, true, nil
}

func (b *Boundary) inspectParent(scope Scope, key CommandKey, intent Intent) (ParentAdmission, bool, error) {
	b.storage.clientMu.RLock()
	defer b.storage.clientMu.RUnlock()
	if b.storage.generation != b.generation {
		return ParentAdmission{}, false, ErrBoundaryExpired
	}
	ledger := b.storage.existingLedger(scope.instanceScope())
	if ledger == nil {
		return ParentAdmission{}, false, nil
	}
	ledger.mu.Lock()
	defer ledger.mu.Unlock()
	record := ledger.parents[commandIdentity{scope: scope, key: key}]
	if record == nil {
		return ParentAdmission{}, false, nil
	}
	if record.intent != intent {
		return ParentAdmission{}, true, ErrCommandKeyConflict
	}
	kind := AdmissionInProgress
	if record.state == CommandStateTerminal {
		kind = AdmissionReplay
	}
	return ParentAdmission{kind: kind, record: record.view()}, true, nil
}

func decideAbsentSafely(decide DecideAbsent, ctx context.Context) (candidate AbsentCandidate, err error) {
	defer func() {
		if recover() != nil {
			candidate = AbsentCandidate{}
			err = ErrInvalidSubmission
		}
	}()
	return decide(ctx)
}

func revalidateSatisfiedSafely(revalidate RevalidateCandidate, ctx context.Context, candidate AbsentCandidate) (confirmed bool) {
	defer func() {
		if recover() != nil {
			confirmed = false
		}
	}()
	result, err := revalidate(ctx, candidate)
	return err == nil && result == CandidateRevalidated
}

func allowPrimitive(context.Context, Scope, Intent) error { return nil }

func allowOrchestrationInternal(context.Context, OrchestrationAuthorizationRequest) error { return nil }

func mustParentSatisfied() ParentTerminalOutcome {
	outcome, _ := NewParentTerminalOutcome(ParentOutcomeSatisfied)
	return outcome
}
