// Package runtimelaunchflow connects Runtime launch preparation to the
// Configuration Loader, Snapshot Builder, and Runtime Lifecycle Owner.
package runtimelaunchflow

import (
	"context"
	"errors"

	"github.com/dsdred/universal-websocket-platform/internal/configurationloader"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle"
)

const buildFailureDescription = "Runtime Snapshot build failed"

var (
	ErrInvalidFlow         = errors.New("invalid Runtime launch Flow")
	ErrInvalidStartContext = errors.New("invalid Runtime launch start context")
)

type buildFunction func(
	runtimeconfigload.DetachedLoadResult,
) (runtimeconfig.Snapshot, []runtimeconfig.Diagnostic)

// Flow synchronously prepares one Runtime launch through its bound Owner and
// Configuration Loader.
type Flow struct {
	owner  *runtimelifecycle.Owner
	loader *configurationloader.Loader
	build  buildFunction
}

// New constructs an immutable Runtime launch Flow.
func New(
	owner *runtimelifecycle.Owner,
	loader *configurationloader.Loader,
) (*Flow, error) {
	builder := runtimeconfig.NewBuilder()
	return newFlow(owner, loader, builder.Build)
}

func newFlow(
	owner *runtimelifecycle.Owner,
	loader *configurationloader.Loader,
	build buildFunction,
) (*Flow, error) {
	if owner == nil || loader == nil || build == nil {
		return nil, ErrInvalidFlow
	}
	return &Flow{
		owner:  owner,
		loader: loader,
		build:  build,
	}, nil
}

// Start performs one synchronous PrepareStart, Load, Build, and Owner.Start
// operation. Caller cancellation is observed exactly once before PrepareStart.
func (f *Flow) Start(
	ctx context.Context,
	request runtimelifecycle.StartRequest,
) (runtimelifecycle.StartOutcome, error) {
	if f == nil || f.owner == nil || f.loader == nil || f.build == nil {
		return runtimelifecycle.StartOutcome{}, ErrInvalidFlow
	}
	if ctx == nil {
		return runtimelifecycle.StartOutcome{}, ErrInvalidStartContext
	}
	if err := ctx.Err(); err != nil {
		return runtimelifecycle.StartOutcome{}, err
	}

	preparation, err := f.owner.PrepareStart(request)
	if err != nil {
		return runtimelifecycle.StartOutcome{}, err
	}
	if preparation.Context().Err() != nil {
		return convergeStoppedPreparation(f.owner, preparation)
	}
	return f.startPrepared(preparation)
}

func (f *Flow) startPrepared(
	preparation runtimelifecycle.LaunchPreparation,
) (runtimelifecycle.StartOutcome, error) {
	loadResult, err := f.loader.Load(preparation.LoadRequest())
	if preparation.Context().Err() != nil {
		return convergeStoppedPreparation(f.owner, preparation)
	}
	if err != nil {
		return f.owner.Start(
			context.Background(),
			preparation,
			runtimelifecycle.FailedPreparation(err),
		)
	}

	snapshot, diagnostics := f.build(loadResult)
	if preparation.Context().Err() != nil {
		return convergeStoppedPreparation(f.owner, preparation)
	}
	if len(diagnostics) != 0 {
		return f.owner.Start(
			context.Background(),
			preparation,
			runtimelifecycle.FailedPreparation(newBuildFailure(diagnostics)),
		)
	}

	return f.owner.Start(
		context.Background(),
		preparation,
		runtimelifecycle.PreparedSnapshot(snapshot),
	)
}

func convergeStoppedPreparation(
	owner *runtimelifecycle.Owner,
	preparation runtimelifecycle.LaunchPreparation,
) (runtimelifecycle.StartOutcome, error) {
	return owner.Start(
		context.Background(),
		preparation,
		runtimelifecycle.PreparationResult{},
	)
}

// BuildFailure is one immutable complete set of blocking Builder Diagnostics.
type BuildFailure struct {
	diagnostics []runtimeconfig.Diagnostic
}

func newBuildFailure(diagnostics []runtimeconfig.Diagnostic) *BuildFailure {
	return &BuildFailure{
		diagnostics: append([]runtimeconfig.Diagnostic(nil), diagnostics...),
	}
}

// Error returns a constant description without exposing configuration data.
func (*BuildFailure) Error() string {
	return buildFailureDescription
}

// Diagnostics returns an independent copy of the complete blocking set.
func (f *BuildFailure) Diagnostics() []runtimeconfig.Diagnostic {
	if f == nil {
		return nil
	}
	return append([]runtimeconfig.Diagnostic(nil), f.diagnostics...)
}
