// Package runtimemanagement routes authorized lifecycle commands to one exact
// process-local Runtime Instance scope.
package runtimemanagement

import (
	"context"
	"errors"

	"github.com/dsdred/universal-websocket-platform/internal/configurationloader"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelaunchflow"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle"
)

var (
	// ErrInvalidBinding indicates an incomplete or identity-mismatched binding.
	ErrInvalidBinding = errors.New("invalid Runtime management binding")
	// ErrInvalidDirectory indicates an invalid directory or composition.
	ErrInvalidDirectory = errors.New("invalid Runtime management directory")
	// ErrInvalidContext indicates a nil command context.
	ErrInvalidContext = errors.New("invalid Runtime management context")
	// ErrInvalidTarget indicates an incomplete target or Start version identity.
	ErrInvalidTarget = errors.New("invalid Runtime management target")
	// ErrRuntimeInstanceNotFound normalizes absent and mismatched target scopes.
	ErrRuntimeInstanceNotFound = errors.New("Runtime Instance not found")
)

// Action identifies one management operation for authorization.
type Action string

const (
	// ActionStart identifies exact-version Runtime start authorization.
	ActionStart Action = "start"
	// ActionStop identifies Runtime stop authorization.
	ActionStop Action = "stop"
	// ActionObserve identifies Runtime observation authorization.
	ActionObserve Action = "observe"
)

// Target is one immutable Runtime Instance ownership scope.
type Target struct {
	workspaceID       uint64
	configurationID   uint64
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID
}

// NewTarget constructs one exact management target.
func NewTarget(
	workspaceID uint64,
	configurationID uint64,
	runtimeInstanceID runtimeconfigload.RuntimeInstanceID,
) (Target, error) {
	target := Target{
		workspaceID:       workspaceID,
		configurationID:   configurationID,
		runtimeInstanceID: runtimeInstanceID,
	}
	if !target.valid() {
		return Target{}, ErrInvalidTarget
	}
	return target, nil
}

// WorkspaceID returns the exact Workspace identity.
func (t Target) WorkspaceID() uint64 { return t.workspaceID }

// ConfigurationID returns the exact Configuration identity.
func (t Target) ConfigurationID() uint64 { return t.configurationID }

// RuntimeInstanceID returns the exact Runtime Instance identity.
func (t Target) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID {
	return t.runtimeInstanceID
}

func (t Target) valid() bool {
	return t.workspaceID != 0 && t.configurationID != 0 && t.runtimeInstanceID != ""
}

// Authorize admits one exact action and target. The version is non-zero only
// for ActionStart.
type Authorize func(context.Context, Action, Target, uint64) error

// Binding declares one immutable Target, Owner, and Loader composition input.
type Binding struct {
	target Target
	owner  *runtimelifecycle.Owner
	loader *configurationloader.Loader
}

// NewBinding validates and constructs one management scope declaration.
func NewBinding(
	target Target,
	owner *runtimelifecycle.Owner,
	loader *configurationloader.Loader,
) (Binding, error) {
	if !target.valid() || owner == nil || loader == nil {
		return Binding{}, ErrInvalidBinding
	}
	observation := owner.Observe()
	if observation.WorkspaceID() != target.WorkspaceID() ||
		observation.ConfigurationID() != target.ConfigurationID() ||
		observation.RuntimeInstanceID() != target.RuntimeInstanceID() {
		return Binding{}, ErrInvalidBinding
	}
	return Binding{target: target, owner: owner, loader: loader}, nil
}

func (b Binding) valid() bool {
	if !b.target.valid() || b.owner == nil || b.loader == nil {
		return false
	}
	observation := b.owner.Observe()
	return observation.WorkspaceID() == b.target.WorkspaceID() &&
		observation.ConfigurationID() == b.target.ConfigurationID() &&
		observation.RuntimeInstanceID() == b.target.RuntimeInstanceID()
}

type scope struct {
	target Target
	owner  *runtimelifecycle.Owner
	flow   *runtimelaunchflow.Flow
}

// Directory is one immutable process-local Runtime management boundary.
type Directory struct {
	authorize Authorize
	scopes    map[runtimeconfigload.RuntimeInstanceID]scope
}

// NewDirectory validates all bindings before constructing their bound launch
// flows and an immutable routing map.
func NewDirectory(authorize Authorize, bindings ...Binding) (*Directory, error) {
	if authorize == nil || len(bindings) == 0 {
		return nil, ErrInvalidDirectory
	}

	instanceIDs := make(map[runtimeconfigload.RuntimeInstanceID]struct{}, len(bindings))
	owners := make(map[*runtimelifecycle.Owner]struct{}, len(bindings))
	for _, binding := range bindings {
		if !binding.valid() {
			return nil, ErrInvalidDirectory
		}
		if _, exists := instanceIDs[binding.target.RuntimeInstanceID()]; exists {
			return nil, ErrInvalidDirectory
		}
		if _, exists := owners[binding.owner]; exists {
			return nil, ErrInvalidDirectory
		}
		instanceIDs[binding.target.RuntimeInstanceID()] = struct{}{}
		owners[binding.owner] = struct{}{}
	}

	scopes := make(map[runtimeconfigload.RuntimeInstanceID]scope, len(bindings))
	for _, binding := range bindings {
		flow, err := runtimelaunchflow.New(binding.owner, binding.loader)
		if err != nil {
			return nil, ErrInvalidDirectory
		}
		scopes[binding.target.RuntimeInstanceID()] = scope{
			target: binding.target,
			owner:  binding.owner,
			flow:   flow,
		}
	}

	return &Directory{authorize: authorize, scopes: scopes}, nil
}

// Start authorizes and delegates one exact version launch to the target's Flow.
func (d *Directory) Start(
	ctx context.Context,
	target Target,
	configurationVersionID uint64,
) (runtimelifecycle.StartOutcome, error) {
	scope, err := d.admit(ctx, ActionStart, target, configurationVersionID)
	if err != nil {
		return runtimelifecycle.StartOutcome{}, err
	}
	return scope.flow.Start(ctx, runtimelifecycle.NewStartRequest(
		scope.target.WorkspaceID(),
		scope.target.ConfigurationID(),
		configurationVersionID,
	))
}

// Stop authorizes and delegates shutdown to the target's Owner.
func (d *Directory) Stop(
	ctx context.Context,
	target Target,
) (runtimelifecycle.StopOutcome, error) {
	scope, err := d.admit(ctx, ActionStop, target, 0)
	if err != nil {
		return runtimelifecycle.StopOutcome{}, err
	}
	return scope.owner.Stop(ctx)
}

// Observe authorizes and returns one coherent observation from the target's Owner.
func (d *Directory) Observe(
	ctx context.Context,
	target Target,
) (runtimelifecycle.Observation, error) {
	scope, err := d.admit(ctx, ActionObserve, target, 0)
	if err != nil {
		return runtimelifecycle.Observation{}, err
	}
	if err := ctx.Err(); err != nil {
		return runtimelifecycle.Observation{}, err
	}
	return scope.owner.Observe(), nil
}

func (d *Directory) admit(
	ctx context.Context,
	action Action,
	target Target,
	configurationVersionID uint64,
) (scope, error) {
	if !d.valid() {
		return scope{}, ErrInvalidDirectory
	}
	if ctx == nil {
		return scope{}, ErrInvalidContext
	}
	if !target.valid() || action == ActionStart && configurationVersionID == 0 {
		return scope{}, ErrInvalidTarget
	}
	if err := ctx.Err(); err != nil {
		return scope{}, err
	}

	selected, exists := d.scopes[target.RuntimeInstanceID()]
	if !exists || selected.target != target {
		return scope{}, ErrRuntimeInstanceNotFound
	}
	if err := d.authorize(ctx, action, target, configurationVersionID); err != nil {
		return scope{}, err
	}
	return selected, nil
}

func (d *Directory) valid() bool {
	if d == nil || d.authorize == nil || len(d.scopes) == 0 {
		return false
	}
	for instanceID, scope := range d.scopes {
		if instanceID == "" || scope.target.RuntimeInstanceID() != instanceID ||
			!scope.target.valid() || scope.owner == nil || scope.flow == nil {
			return false
		}
	}
	return true
}
