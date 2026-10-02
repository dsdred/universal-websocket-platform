// Package runtimecontainmentcomposition assembles the repository-private
// containment-gated activation and execution-evidence boundary.
//
// It owns no containment, command, lifecycle, identity, or recovery truth.
package runtimecontainmentcomposition

import (
	"context"
	"errors"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeactivation"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecommandidempotency"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecontainment"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeexecutionevidence"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeidentity"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle"
	"github.com/dsdred/universal-websocket-platform/internal/runtimemanagement"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

var errContainmentUnavailable = errors.New("Runtime containment authority unavailable")

type versionStore interface {
	Get(uint64) (configurationversion.ConfigurationVersion, error)
}

type identityStore interface {
	ReadRuntimeInstance(runtimeconfigload.RuntimeInstanceID) (runtimeidentity.RuntimeInstanceView, error)
	ReadLaunchAttemptHistory(runtimeconfigload.RuntimeInstanceID) ([]runtimeidentity.LaunchAttemptRecord, error)
	ConditionalClaimStop(runtimeconfigload.RuntimeInstanceID, runtimeidentity.Revision, runtimeconfigload.LaunchAttemptID) (runtimeidentity.PublishResult, error)
	ConditionalPublishRunning(runtimeconfigload.RuntimeInstanceID, runtimeidentity.Revision, runtimeconfigload.LaunchAttemptID) (runtimeidentity.PublishResult, error)
	OwnerTerminalPublisher() runtimeidentity.OwnerTerminalPublication
}

type lifecycleOwner interface {
	Observe() runtimelifecycle.Observation
	StopExpectedAttempt(context.Context, runtimeconfigload.LaunchAttemptID) (runtimelifecycle.StopOutcome, error)
}

type managedStartInvoker interface {
	InvokeManagedStart(context.Context, runtimelifecycle.StartRequest, runtimeorchestrationbinding.StartExecutionBinding) (runtimelifecycle.StartOutcome, error)
}

// Composition is one immutable, repository-private assembly of an exact
// containment authority, activation orchestrator, and evidence composer. It
// exposes no underlying dependency or generation accessor.
type Composition struct {
	domain       runtimecontainment.Domain
	authority    *runtimecontainment.ActiveAuthority
	orchestrator *runtimeactivation.Orchestrator
	evidence     *runtimeexecutionevidence.Composer
}

// New constructs one containment-gated composition from an already acquired
// authority. The operational domain must be the canonical representation that
// parses to the supplied containment domain.
func New(
	operationalDomain string,
	domain runtimecontainment.Domain,
	containment runtimecontainment.Result,
	target runtimemanagement.Target,
	versions versionStore,
	identities identityStore,
	owner lifecycleOwner,
	commands *runtimecommandidempotency.Boundary,
	start managedStartInvoker,
	authorize runtimeorchestrationbinding.AuthorizeOrchestration,
) (*Composition, error) {
	parsed, err := runtimecontainment.ParseDomain(operationalDomain)
	if err != nil || parsed != domain || identities == nil || authorize == nil {
		return nil, errContainmentUnavailable
	}
	authority, ok := containment.Authority()
	if !ok || !authorityCurrent(context.Background(), domain, authority) {
		return nil, errContainmentUnavailable
	}

	composition := &Composition{domain: domain, authority: authority}
	gatedAuthorize := func(ctx context.Context, request runtimeorchestrationbinding.OrchestrationAuthorizationRequest) error {
		if err := composition.requireCurrent(ctx); err != nil {
			return err
		}
		policyErr := authorize(ctx, request)
		if err := composition.requireCurrent(ctx); err != nil {
			return err
		}
		return policyErr
	}
	provideGeneration := func(ctx context.Context) (runtimeorchestrationbinding.ExecutionGeneration, error) {
		if err := composition.requireCurrent(ctx); err != nil {
			return "", err
		}
		return runtimeorchestrationbinding.ExecutionGeneration(authority.CurrentGeneration()), nil
	}

	orchestrator, err := runtimeactivation.New(
		operationalDomain, target, versions, identities, owner, commands, start,
		gatedAuthorize, provideGeneration,
	)
	if err != nil {
		return nil, err
	}
	composition.orchestrator = orchestrator
	composition.evidence = runtimeexecutionevidence.NewComposer(domain, authority, identities)
	return composition, nil
}

// ActivateExact submits one exact activation through the live containment
// gate and the existing activation orchestrator.
func (c *Composition) ActivateExact(ctx context.Context, request runtimeactivation.Request) (runtimeactivation.Result, error) {
	if err := c.requireCurrent(ctx); err != nil {
		return runtimeactivation.Result{}, err
	}
	return c.orchestrator.ActivateExact(ctx, request)
}

// ReplaceExact submits one exact replacement through the live containment
// gate and the existing activation orchestrator.
func (c *Composition) ReplaceExact(ctx context.Context, request runtimeactivation.Request) (runtimeactivation.Result, error) {
	if err := c.requireCurrent(ctx); err != nil {
		return runtimeactivation.Result{}, err
	}
	return c.orchestrator.ReplaceExact(ctx, request)
}

// RollbackExact submits one exact rollback through the live containment gate
// and the existing activation orchestrator.
func (c *Composition) RollbackExact(ctx context.Context, request runtimeactivation.Request) (runtimeactivation.Result, error) {
	if err := c.requireCurrent(ctx); err != nil {
		return runtimeactivation.Result{}, err
	}
	return c.orchestrator.RollbackExact(ctx, request)
}

// QueryEvidence delegates one exact-tuple question to the evidence composer
// bound to this composition. Foreign domains remain fail-closed there.
func (c *Composition) QueryEvidence(
	ctx context.Context,
	domain runtimecontainment.Domain,
	instanceID runtimeconfigload.RuntimeInstanceID,
	attemptID runtimeconfigload.LaunchAttemptID,
	generation runtimeidentity.ExecutionGeneration,
	question runtimeexecutionevidence.Question,
) runtimeexecutionevidence.EvidenceHandle {
	if c == nil {
		var evidence *runtimeexecutionevidence.Composer
		return evidence.Query(ctx, domain, instanceID, attemptID, generation, question)
	}
	return c.evidence.Query(ctx, domain, instanceID, attemptID, generation, question)
}

func (c *Composition) requireCurrent(ctx context.Context) error {
	if c == nil || c.orchestrator == nil || c.evidence == nil ||
		!authorityCurrent(ctx, c.domain, c.authority) {
		return errContainmentUnavailable
	}
	return nil
}

func authorityCurrent(
	ctx context.Context,
	domain runtimecontainment.Domain,
	authority *runtimecontainment.ActiveAuthority,
) bool {
	if authority == nil || ctx == nil {
		return false
	}
	generation := authority.CurrentGeneration()
	if generation == "" {
		return false
	}
	evidence := authority.ReadGeneration(ctx, domain, generation)
	return evidence.Kind() == runtimecontainment.EvidenceGenerationLive &&
		authority.CurrentGeneration() == generation && authority.IsAuthoritative()
}
