package runtimelifecycle

import (
	"context"
	"fmt"
	"sync"

	"github.com/dsdred/universal-websocket-platform/internal/runtime"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

type launchResult struct {
	outcome runtime.BootstrapOutcome
	host    runtime.Host
	success bool
}

type launchFunction func(*runtime.BootstrapRequest) launchResult

type attemptState struct {
	fact AttemptFact

	context context.Context
	cancel  context.CancelFunc

	accepted           bool
	preparationFailure error
	launchOutcome      runtime.BootstrapOutcome
	host               runtime.Host

	startDone    chan struct{}
	startOutcome StartOutcome
	startReady   bool

	stopDone    chan struct{}
	stopOutcome StopOutcome
	stopReady   bool
	stopStarted bool
}

type Owner struct {
	mu sync.Mutex

	workspaceID     uint64
	configurationID uint64
	instanceID      runtimeconfigload.RuntimeInstanceID
	nextAttemptID   LaunchAttemptIDSource
	dependencies    *runtime.DependencyBindings
	launch          launchFunction

	desired DesiredState
	actual  ActualState

	usedAttempts map[runtimeconfigload.LaunchAttemptID]struct{}
	attempts     map[*attemptState]struct{}
	active       *attemptState
	last         *attemptState
}

func NewOwner(
	workspaceID uint64,
	configurationID uint64,
	instanceID runtimeconfigload.RuntimeInstanceID,
	nextAttemptID LaunchAttemptIDSource,
	dependencies *runtime.DependencyBindings,
) (*Owner, error) {
	return newOwner(
		workspaceID,
		configurationID,
		instanceID,
		nextAttemptID,
		dependencies,
		launchRuntime,
	)
}

func newOwner(
	workspaceID uint64,
	configurationID uint64,
	instanceID runtimeconfigload.RuntimeInstanceID,
	nextAttemptID LaunchAttemptIDSource,
	dependencies *runtime.DependencyBindings,
	launch launchFunction,
) (*Owner, error) {
	if workspaceID == 0 ||
		configurationID == 0 ||
		instanceID == "" ||
		nextAttemptID == nil ||
		dependencies == nil ||
		launch == nil {
		return nil, ErrInvalidOwner
	}

	return &Owner{
		workspaceID:     workspaceID,
		configurationID: configurationID,
		instanceID:      instanceID,
		nextAttemptID:   nextAttemptID,
		dependencies:    dependencies,
		launch:          launch,
		desired:         DesiredStopped,
		actual:          ActualStopped,
		usedAttempts:    make(map[runtimeconfigload.LaunchAttemptID]struct{}),
		attempts:        make(map[*attemptState]struct{}),
	}, nil
}

func launchRuntime(request *runtime.BootstrapRequest) launchResult {
	outcome := runtime.Launch(request)
	host, success := outcome.Success()
	return launchResult{
		outcome: outcome,
		host:    host,
		success: success,
	}
}

func (o *Owner) PrepareStart(request StartRequest) (LaunchPreparation, error) {
	if o == nil ||
		request.workspaceID == 0 ||
		request.configurationID == 0 ||
		request.configurationVersionID == 0 ||
		request.workspaceID != o.workspaceID ||
		request.configurationID != o.configurationID {
		return LaunchPreparation{}, ErrInvalidStartRequest
	}

	candidate, err := o.nextAttemptID()
	if err != nil {
		return LaunchPreparation{}, fmt.Errorf("%w: %w", ErrAttemptIDSourceFailed, err)
	}
	if candidate == "" {
		return LaunchPreparation{}, ErrAttemptIDSourceFailed
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	if _, reused := o.usedAttempts[candidate]; reused {
		return LaunchPreparation{}, ErrAttemptIDReused
	}
	if !o.startClaimAllowedLocked() {
		return LaunchPreparation{}, ErrStartConflict
	}

	preparationContext, cancel := context.WithCancel(context.Background())
	attempt := &attemptState{
		fact: AttemptFact{
			workspaceID:            o.workspaceID,
			configurationID:        o.configurationID,
			configurationVersionID: request.configurationVersionID,
			runtimeInstanceID:      o.instanceID,
			launchAttemptID:        candidate,
			phase:                  AttemptPreparing,
		},
		context:   preparationContext,
		cancel:    cancel,
		startDone: make(chan struct{}),
	}

	loadRequest := runtimeconfigload.NewLoadRequest(
		o.workspaceID,
		o.configurationID,
		request.configurationVersionID,
		o.instanceID,
		candidate,
	)
	preparation := LaunchPreparation{
		owner:       o,
		attempt:     attempt,
		loadRequest: loadRequest,
		context:     preparationContext,
	}

	o.usedAttempts[candidate] = struct{}{}
	o.attempts[attempt] = struct{}{}
	o.active = attempt
	o.desired = DesiredRunning
	o.actual = ActualStarting

	return preparation, nil
}

func (o *Owner) startClaimAllowedLocked() bool {
	if o.active != nil {
		return false
	}
	switch o.actual {
	case ActualStopped:
		return true
	case ActualFailed:
		return o.active == nil
	default:
		return false
	}
}

func (o *Owner) Start(
	ctx context.Context,
	preparation LaunchPreparation,
	result PreparationResult,
) (StartOutcome, error) {
	if o == nil {
		return StartOutcome{}, ErrPreparationNotOwned
	}

	o.mu.Lock()
	attempt, owned := o.authenticAttemptLocked(preparation)
	if !owned {
		o.mu.Unlock()
		return StartOutcome{}, ErrPreparationNotOwned
	}
	if attempt.startReady || attempt.accepted {
		if err := ctx.Err(); err != nil {
			o.mu.Unlock()
			return StartOutcome{}, err
		}
		if attempt.startReady {
			outcome := attempt.startOutcome
			o.mu.Unlock()
			return outcome, nil
		}
		done := attempt.startDone
		o.mu.Unlock()
		return o.waitStart(ctx, attempt, done)
	}
	o.mu.Unlock()

	snapshot, preparationFailure, validationErr := validatePreparationResult(
		preparation.loadRequest,
		result,
	)
	if validationErr != nil {
		return StartOutcome{}, validationErr
	}

	o.mu.Lock()
	attempt, owned = o.authenticAttemptLocked(preparation)
	if !owned {
		o.mu.Unlock()
		return StartOutcome{}, ErrPreparationNotOwned
	}
	if attempt.startReady || attempt.accepted {
		if err := ctx.Err(); err != nil {
			o.mu.Unlock()
			return StartOutcome{}, err
		}
		if attempt.startReady {
			outcome := attempt.startOutcome
			o.mu.Unlock()
			return outcome, nil
		}
		done := attempt.startDone
		o.mu.Unlock()
		return o.waitStart(ctx, attempt, done)
	}
	if o.active != attempt || attempt.fact.phase != AttemptPreparing {
		o.mu.Unlock()
		return StartOutcome{}, ErrPreparationNotOwned
	}
	if err := ctx.Err(); err != nil {
		o.mu.Unlock()
		return StartOutcome{}, err
	}

	attempt.accepted = true
	if preparationFailure != nil {
		attempt.preparationFailure = preparationFailure
		attempt.fact.phase = AttemptHistorical
		attempt.fact.terminalKind = AttemptPreparationFailed
		o.active = nil
		o.last = attempt
		o.actual = ActualFailed
		outcome := StartOutcome{
			kind:               StartPreparationFailed,
			attempt:            attempt.fact,
			preparationFailure: preparationFailure,
		}
		o.publishStartLocked(attempt, outcome)
		cancel := attempt.cancel
		o.mu.Unlock()
		cancel()
		return outcome, nil
	}

	attempt.fact.phase = AttemptLaunching
	done := attempt.startDone
	o.mu.Unlock()

	go o.runLaunch(attempt, snapshot)
	return o.waitStart(ctx, attempt, done)
}

func validatePreparationResult(
	loadRequest runtimeconfigload.LoadRequest,
	result PreparationResult,
) (runtimeconfig.Snapshot, error, error) {
	switch result.kind {
	case PreparationSnapshot:
		snapshot, ok := result.Snapshot()
		if !ok || !snapshotMatchesLoadRequest(snapshot, loadRequest) {
			return runtimeconfig.Snapshot{}, nil, ErrInvalidPreparationResult
		}
		return snapshot, nil, nil
	case PreparationFailure:
		cause, ok := result.Failure()
		if !ok {
			return runtimeconfig.Snapshot{}, nil, ErrInvalidPreparationResult
		}
		return runtimeconfig.Snapshot{}, cause, nil
	default:
		return runtimeconfig.Snapshot{}, nil, ErrInvalidPreparationResult
	}
}

func snapshotMatchesLoadRequest(
	snapshot runtimeconfig.Snapshot,
	request runtimeconfigload.LoadRequest,
) bool {
	provenance := snapshot.Provenance()
	return provenance.WorkspaceID == request.WorkspaceID() &&
		provenance.ConfigurationID == request.ConfigurationID() &&
		provenance.ConfigurationVersionID == request.ConfigurationVersionID() &&
		provenance.RuntimeInstanceID == request.RuntimeInstanceID() &&
		provenance.LaunchAttemptID == request.LaunchAttemptID()
}

func (o *Owner) Stop(ctx context.Context) (StopOutcome, error) {
	if o == nil {
		return StopOutcome{}, ErrInvalidOwner
	}

	o.mu.Lock()
	if err := ctx.Err(); err != nil {
		o.mu.Unlock()
		return StopOutcome{}, err
	}
	return o.stopLocked(ctx, o.active, false)
}

// StopExpectedAttempt stops only the exact relevant Owner-issued Launch Attempt.
func (o *Owner) StopExpectedAttempt(
	ctx context.Context,
	expectedAttemptID runtimeconfigload.LaunchAttemptID,
) (StopOutcome, error) {
	if o == nil {
		return StopOutcome{}, ErrInvalidOwner
	}
	if expectedAttemptID == "" {
		return StopOutcome{}, ErrInvalidExpectedAttempt
	}

	o.mu.Lock()
	if err := ctx.Err(); err != nil {
		o.mu.Unlock()
		return StopOutcome{}, err
	}

	attempt := o.active
	if attempt == nil {
		attempt = o.last
	}
	if attempt == nil || attempt.fact.launchAttemptID != expectedAttemptID {
		outcome := StopOutcome{kind: StopAttemptMismatch}
		if attempt != nil {
			outcome.attempt = attempt.fact
			outcome.hasAttempt = true
		}
		o.mu.Unlock()
		return outcome, nil
	}

	return o.stopLocked(ctx, attempt, attempt == o.last && o.active == nil)
}

func (o *Owner) stopLocked(
	ctx context.Context,
	attempt *attemptState,
	retainedMatch bool,
) (StopOutcome, error) {
	if attempt == nil {
		switch o.actual {
		case ActualStopped:
			o.mu.Unlock()
			return StopOutcome{kind: StopStopped}, nil
		case ActualFailed:
			o.desired = DesiredStopped
			o.actual = ActualStopped
			outcome := StopOutcome{kind: StopStopped}
			if o.last != nil {
				outcome.attempt = o.last.fact
				outcome.hasAttempt = true
			}
			o.mu.Unlock()
			return outcome, nil
		default:
			o.mu.Unlock()
			return StopOutcome{}, ErrStartConflict
		}
	}
	if retainedMatch {
		switch attempt.fact.terminalKind {
		case AttemptStopped, AttemptStoppedBeforeRunning:
			outcome := StopOutcome{
				kind:       StopStopped,
				attempt:    attempt.fact,
				hasAttempt: true,
			}
			o.mu.Unlock()
			return outcome, nil
		case AttemptPreparationFailed, AttemptLaunchFailed:
			o.desired = DesiredStopped
			o.actual = ActualStopped
			outcome := StopOutcome{
				kind:       StopStopped,
				attempt:    attempt.fact,
				hasAttempt: true,
			}
			o.mu.Unlock()
			return outcome, nil
		default:
			o.mu.Unlock()
			return StopOutcome{}, ErrStartConflict
		}
	}

	if attempt.stopReady {
		outcome := attempt.stopOutcome
		o.mu.Unlock()
		return outcome, nil
	}

	switch attempt.fact.phase {
	case AttemptPreparing:
		attempt.fact.stopOrigin = StopBeforeRunning
		attempt.fact.phase = AttemptHistorical
		attempt.fact.terminalKind = AttemptStoppedBeforeRunning
		o.desired = DesiredStopped
		o.actual = ActualStopped
		o.active = nil
		o.last = attempt
		startOutcome := StartOutcome{
			kind:    StartStoppedBeforeRunning,
			attempt: attempt.fact,
		}
		stopOutcome := StopOutcome{
			kind:       StopStopped,
			attempt:    attempt.fact,
			hasAttempt: true,
		}
		o.publishStartLocked(attempt, startOutcome)
		o.publishStopLocked(attempt, stopOutcome)
		cancel := attempt.cancel
		o.mu.Unlock()
		cancel()
		return stopOutcome, nil

	case AttemptLaunching:
		attempt.fact.stopOrigin = StopBeforeRunning
		attempt.fact.phase = AttemptStopping
		o.desired = DesiredStopped
		o.actual = ActualStopping
		attempt.stopDone = make(chan struct{})
		done := attempt.stopDone
		cancel := attempt.cancel
		o.mu.Unlock()
		cancel()
		return o.waitStop(ctx, attempt, done)

	case AttemptRunning:
		attempt.fact.stopOrigin = StopAfterRunning
		attempt.fact.phase = AttemptStopping
		o.desired = DesiredStopped
		o.actual = ActualStopping
		attempt.stopDone = make(chan struct{})
		attempt.stopStarted = true
		done := attempt.stopDone
		host := attempt.host
		o.mu.Unlock()
		go o.runHostStop(attempt, host)
		return o.waitStop(ctx, attempt, done)

	case AttemptStopping:
		done := attempt.stopDone
		o.mu.Unlock()
		return o.waitStop(ctx, attempt, done)

	default:
		o.mu.Unlock()
		return StopOutcome{}, ErrStartConflict
	}
}

func (o *Owner) runLaunch(attempt *attemptState, snapshot runtimeconfig.Snapshot) {
	result := o.launch(&runtime.BootstrapRequest{
		Snapshot:       snapshot,
		StartupContext: attempt.context,
		Dependencies:   o.dependencies,
	})

	var stopHost runtime.Host
	var runStop bool

	o.mu.Lock()
	if !result.success {
		attempt.launchOutcome = result.outcome
	}
	if result.success {
		attempt.host = result.host
		if attempt.fact.stopOrigin == StopNotClaimed &&
			o.active == attempt &&
			o.desired == DesiredRunning {
			attempt.fact.phase = AttemptRunning
			attempt.fact.runningPublished = true
			o.actual = ActualRunning
			o.publishStartLocked(attempt, StartOutcome{
				kind:    StartRunning,
				attempt: attempt.fact,
			})
		} else {
			if !attempt.stopStarted {
				attempt.stopStarted = true
				stopHost = result.host
				runStop = true
			}
		}
	} else if attempt.fact.stopOrigin == StopBeforeRunning {
		attempt.fact.phase = AttemptHistorical
		attempt.fact.terminalKind = AttemptStoppedBeforeRunning
		o.active = nil
		o.last = attempt
		o.actual = ActualStopped
		o.publishStartLocked(attempt, StartOutcome{
			kind:    StartStoppedBeforeRunning,
			attempt: attempt.fact,
		})
		o.publishStopLocked(attempt, StopOutcome{
			kind:       StopStopped,
			attempt:    attempt.fact,
			hasAttempt: true,
		})
	} else {
		attempt.fact.phase = AttemptHistorical
		attempt.fact.terminalKind = AttemptLaunchFailed
		o.active = nil
		o.last = attempt
		o.actual = ActualFailed
		o.publishStartLocked(attempt, StartOutcome{
			kind:          StartLaunchFailed,
			attempt:       attempt.fact,
			launchOutcome: result.outcome,
		})
	}
	cancel := attempt.cancel
	o.mu.Unlock()

	cancel()
	if runStop {
		go o.runHostStop(attempt, stopHost)
	}
}

func (o *Owner) runHostStop(attempt *attemptState, host runtime.Host) {
	err := host.Stop(context.Background())

	o.mu.Lock()
	if err == nil {
		attempt.host = nil
		attempt.fact.phase = AttemptHistorical
		attempt.fact.terminalKind = AttemptStopped
		o.active = nil
		o.last = attempt
		o.actual = ActualStopped
		if attempt.fact.stopOrigin == StopBeforeRunning {
			o.publishStartLocked(attempt, StartOutcome{
				kind:    StartStoppedBeforeRunning,
				attempt: attempt.fact,
			})
		}
		o.publishStopLocked(attempt, StopOutcome{
			kind:       StopStopped,
			attempt:    attempt.fact,
			hasAttempt: true,
		})
	} else {
		attempt.fact.phase = AttemptStopping
		attempt.fact.terminalKind = AttemptStopFailed
		o.actual = ActualFailed
		if attempt.fact.stopOrigin == StopBeforeRunning {
			o.publishStartLocked(attempt, StartOutcome{
				kind:    StartStoppedBeforeRunning,
				attempt: attempt.fact,
			})
		}
		o.publishStopLocked(attempt, StopOutcome{
			kind:       StopFailed,
			attempt:    attempt.fact,
			hasAttempt: true,
			failure:    err,
		})
	}
	o.mu.Unlock()
}

func (o *Owner) Observe() Observation {
	if o == nil {
		return Observation{}
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	observation := Observation{
		runtimeInstanceID: o.instanceID,
		workspaceID:       o.workspaceID,
		configurationID:   o.configurationID,
		desired:           o.desired,
		actual:            o.actual,
	}
	if o.active != nil {
		observation.active = o.active.fact
		observation.hasActive = true
	}
	if o.last != nil {
		observation.last = o.last.fact
		observation.hasLast = true
	}
	return observation
}

func (o *Owner) authenticAttemptLocked(preparation LaunchPreparation) (*attemptState, bool) {
	if preparation.owner != o || preparation.attempt == nil {
		return nil, false
	}
	if _, exists := o.attempts[preparation.attempt]; !exists {
		return nil, false
	}
	return preparation.attempt, true
}

func (o *Owner) waitStart(
	ctx context.Context,
	attempt *attemptState,
	done <-chan struct{},
) (StartOutcome, error) {
	select {
	case <-done:
		o.mu.Lock()
		outcome := attempt.startOutcome
		o.mu.Unlock()
		return outcome, nil
	case <-ctx.Done():
		return StartOutcome{}, ctx.Err()
	}
}

func (o *Owner) waitStop(
	ctx context.Context,
	attempt *attemptState,
	done <-chan struct{},
) (StopOutcome, error) {
	select {
	case <-done:
		o.mu.Lock()
		outcome := attempt.stopOutcome
		o.mu.Unlock()
		return outcome, nil
	case <-ctx.Done():
		return StopOutcome{}, ctx.Err()
	}
}

func (o *Owner) publishStartLocked(attempt *attemptState, outcome StartOutcome) {
	if attempt.startReady {
		return
	}
	attempt.startOutcome = outcome
	attempt.startReady = true
	close(attempt.startDone)
}

func (o *Owner) publishStopLocked(attempt *attemptState, outcome StopOutcome) {
	if attempt.stopReady {
		return
	}
	attempt.stopOutcome = outcome
	attempt.stopReady = true
	if attempt.stopDone != nil {
		close(attempt.stopDone)
	}
}
