package runtimecontainmentcomposition

import (
	"context"

	"github.com/dsdred/universal-websocket-platform/internal/runtimecommandidempotency"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecontainment"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecoordination"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeidentity"
	"github.com/dsdred/universal-websocket-platform/internal/runtimemanagement"
)

// ClosedAssembly owns one complete source-bound pair of permanently closed
// coordinated owners. It exposes only detached observations, no adoption,
// activation, participant, root, owner, boundary, or recovery permit.
type ClosedAssembly struct{ data *closedOwners }
type closedOwners struct {
	scope      runtimecoordination.Scope
	identities *runtimeidentity.Store
	commands   *runtimecommandidempotency.MemoryStorage
	boundary   *runtimecommandidempotency.Boundary
}

// ClosedObservation contains complete detached observations from both owners.
// It grants no mutation authority or assertion of live recovery readiness.
type ClosedObservation struct {
	identities runtimeidentity.CoordinationSnapshot
	commands   runtimecommandidempotency.CoordinationSnapshot
}

// Identities returns the detached DP-014 observation.
func (o ClosedObservation) Identities() runtimeidentity.CoordinationSnapshot { return o.identities }

// Commands returns the detached DP-015 observation.
func (o ClosedObservation) Commands() runtimecommandidempotency.CoordinationSnapshot {
	return o.commands
}

// NewClosedAssembly supplies this exact keeper's root and one source-registered
// participant to both real owners before any absence can escape. Failed partial
// construction returns no assembly; its source registrations remain retained.
func NewClosedAssembly(operationalDomain string, domain runtimecontainment.Domain, containment runtimecontainment.Result, target runtimemanagement.Target) (*ClosedAssembly, error) {
	parsed, err := runtimecontainment.ParseDomain(operationalDomain)
	if err != nil || parsed != domain {
		return nil, errContainmentUnavailable
	}
	scope, err := runtimecoordination.NewScope(operationalDomain, target.WorkspaceID(), target.ConfigurationID(), string(target.RuntimeInstanceID()))
	if err != nil {
		return nil, err
	}
	authority, ok := containment.Authority()
	if !ok || !authorityCurrent(context.Background(), domain, authority) {
		return nil, errContainmentUnavailable
	}
	root, err := authority.CoordinationRoot(domain)
	if err != nil {
		return nil, err
	}
	participant, err := runtimecoordination.NewSourceParticipant(root, scope)
	if err != nil {
		return nil, err
	}
	identities, err := runtimeidentity.NewCoordinatedStore(scope, participant)
	if err != nil {
		return nil, err
	}
	commands, err := runtimecommandidempotency.NewCoordinatedMemoryStorage(root)
	if err != nil {
		return nil, err
	}
	boundary, err := runtimecommandidempotency.NewBoundary(commands)
	if err != nil {
		return nil, err
	}
	if err = commands.AttachInstance(scope, participant); err != nil {
		return nil, err
	}
	assembly := &ClosedAssembly{data: &closedOwners{scope, identities, commands, boundary}}
	if _, err = assembly.Capture(scope); err != nil {
		return nil, err
	}
	return assembly, nil
}

// Capture returns both detached owner observations only for the exact scope.
// No incomplete observation escapes when either owner rejects the read.
func (a *ClosedAssembly) Capture(scope runtimecoordination.Scope) (ClosedObservation, error) {
	if a == nil || a.data == nil || !scope.Valid() {
		return ClosedObservation{}, runtimecoordination.ErrInvalidScope
	}
	if scope != a.data.scope {
		return ClosedObservation{}, runtimecoordination.ErrForeignScope
	}
	i, err := a.data.identities.CaptureCoordinationSnapshot(scope)
	if err != nil {
		return ClosedObservation{}, err
	}
	c, err := a.data.commands.CaptureCoordinationSnapshot(scope)
	if err != nil {
		return ClosedObservation{}, err
	}
	return ClosedObservation{i, c}, nil
}
