package runtime

import (
	"context"
	"reflect"

	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

// BootstrapFailureStage identifies the step that rejected a Bootstrap request.
type BootstrapFailureStage string

const (
	BootstrapStageInputValidation   BootstrapFailureStage = "input-validation"
	BootstrapStageDependencyBinding BootstrapFailureStage = "dependency-binding"
	BootstrapStageHostConstruction  BootstrapFailureStage = "host-construction"
	BootstrapStageHostPreparation   BootstrapFailureStage = "host-preparation"
)

// BootstrapFailureCode identifies one stable Bootstrap failure.
type BootstrapFailureCode string

const (
	BootstrapCodeInvalidStartupContext  BootstrapFailureCode = "invalid-startup-context"
	BootstrapCodeInvalidSnapshot        BootstrapFailureCode = "invalid-snapshot"
	BootstrapCodeMissingSecretResolver  BootstrapFailureCode = "missing-secret-resolver"
	BootstrapCodeHostConstructionFailed BootstrapFailureCode = "host-construction-failed"
	BootstrapCodeHostBuildFailed        BootstrapFailureCode = "host-build-failed"
)

const (
	invalidStartupContextDescription  = "Bootstrap startup context is missing"
	invalidSnapshotDescription        = "Bootstrap Snapshot is invalid"
	missingSecretResolverDescription  = "Bootstrap Secret Resolver binding is missing"
	hostConstructionFailedDescription = "Runtime Host construction failed"
	hostBuildFailedDescription        = "Runtime Host build failed"
	hostStartupFailedDescription      = "Runtime Host startup failed"
)

// DependencyBindings contains the fixed capabilities borrowed by Runtime Host composition.
type DependencyBindings struct {
	SecretResolver        secretresolver.Resolver
	LegacyMessageHandler  message.Handler
	TerminalErrorReporter func(error)
}

// BootstrapRequest is the complete static input for one Runtime Host launch.
type BootstrapRequest struct {
	Snapshot       runtimeconfig.Snapshot
	StartupContext context.Context
	Dependencies   *DependencyBindings
}

type bootstrapOutcomeKind uint8

const (
	bootstrapOutcomeSuccess bootstrapOutcomeKind = iota + 1
	bootstrapOutcomeBootstrapFailure
	bootstrapOutcomeStartupFailure
)

// BootstrapOutcome is a closed, mutually exclusive Bootstrap result.
type BootstrapOutcome struct {
	kind             bootstrapOutcomeKind
	host             Host
	bootstrapFailure *BootstrapFailure
	startupFailure   *StartupFailure
}

// Success returns the active Host only for a successful Bootstrap outcome.
func (outcome BootstrapOutcome) Success() (Host, bool) {
	return outcome.host, outcome.kind == bootstrapOutcomeSuccess
}

// BootstrapFailure returns the failure only when Host.Start was not invoked.
func (outcome BootstrapOutcome) BootstrapFailure() (*BootstrapFailure, bool) {
	return outcome.bootstrapFailure, outcome.kind == bootstrapOutcomeBootstrapFailure
}

// StartupFailure returns the failure only when Host.Start was invoked and failed.
func (outcome BootstrapOutcome) StartupFailure() (*StartupFailure, bool) {
	return outcome.startupFailure, outcome.kind == bootstrapOutcomeStartupFailure
}

// BootstrapFailure describes a failure before Host.Start begins.
type BootstrapFailure struct {
	stage BootstrapFailureStage
	code  BootstrapFailureCode
	cause error
}

// Stage returns the stable failure stage.
func (failure *BootstrapFailure) Stage() BootstrapFailureStage {
	if failure == nil {
		return ""
	}
	return failure.stage
}

// Code returns the stable failure code.
func (failure *BootstrapFailure) Code() BootstrapFailureCode {
	if failure == nil {
		return ""
	}
	return failure.code
}

// Error returns the fixed description for this failure identity.
func (failure *BootstrapFailure) Error() string {
	if failure == nil {
		return ""
	}
	switch failure.code {
	case BootstrapCodeInvalidStartupContext:
		return invalidStartupContextDescription
	case BootstrapCodeInvalidSnapshot:
		return invalidSnapshotDescription
	case BootstrapCodeMissingSecretResolver:
		return missingSecretResolverDescription
	case BootstrapCodeHostConstructionFailed:
		return hostConstructionFailedDescription
	case BootstrapCodeHostBuildFailed:
		return hostBuildFailedDescription
	default:
		return ""
	}
}

// Unwrap returns the original construction or preparation cause, when present.
func (failure *BootstrapFailure) Unwrap() error {
	if failure == nil {
		return nil
	}
	return failure.cause
}

// StartupFailure contains the unmodified error returned by Host.Start.
type StartupFailure struct {
	cause error
}

// Error is constant and intentionally excludes the startup cause text.
func (*StartupFailure) Error() string {
	return hostStartupFailedDescription
}

// Unwrap returns the unmodified error returned by Host.Start.
func (failure *StartupFailure) Unwrap() error {
	if failure == nil {
		return nil
	}
	return failure.cause
}

type bootstrapHostFactory func(
	runtimeconfig.Snapshot,
	secretresolver.Resolver,
	message.Handler,
	func(error),
) (Host, error)

// Bootstrap validates one request, constructs one Host, and synchronously starts it.
func Bootstrap(request *BootstrapRequest) BootstrapOutcome {
	return bootstrap(request, newProductionBootstrapHost)
}

func newProductionBootstrapHost(
	snapshot runtimeconfig.Snapshot,
	resolver secretresolver.Resolver,
	handler message.Handler,
	reportError func(error),
) (Host, error) {
	return newHostWithTerminalErrorReporter(snapshot, resolver, handler, nil, reportError)
}

func bootstrap(request *BootstrapRequest, factory bootstrapHostFactory) BootstrapOutcome {
	if request == nil || isNilBootstrapCapability(request.StartupContext) {
		return newBootstrapFailure(
			BootstrapStageInputValidation,
			BootstrapCodeInvalidStartupContext,
			nil,
		)
	}
	if !completeBootstrapSnapshot(request.Snapshot) {
		return newBootstrapFailure(
			BootstrapStageInputValidation,
			BootstrapCodeInvalidSnapshot,
			nil,
		)
	}
	if request.Dependencies == nil ||
		isNilBootstrapCapability(request.Dependencies.SecretResolver) {
		return newBootstrapFailure(
			BootstrapStageDependencyBinding,
			BootstrapCodeMissingSecretResolver,
			nil,
		)
	}

	handler := request.Dependencies.LegacyMessageHandler
	if isNilBootstrapCapability(handler) {
		handler = nil
	}
	reporter := request.Dependencies.TerminalErrorReporter

	host, err := factory(
		request.Snapshot,
		request.Dependencies.SecretResolver,
		handler,
		reporter,
	)
	if err != nil {
		return newBootstrapFailure(
			BootstrapStageHostConstruction,
			BootstrapCodeHostConstructionFailed,
			err,
		)
	}
	if err := host.Build(); err != nil {
		return newBootstrapFailure(
			BootstrapStageHostPreparation,
			BootstrapCodeHostBuildFailed,
			err,
		)
	}
	if err := host.Start(request.StartupContext); err != nil {
		return BootstrapOutcome{
			kind:           bootstrapOutcomeStartupFailure,
			startupFailure: &StartupFailure{cause: err},
		}
	}
	return BootstrapOutcome{kind: bootstrapOutcomeSuccess, host: host}
}

func newBootstrapFailure(
	stage BootstrapFailureStage,
	code BootstrapFailureCode,
	cause error,
) BootstrapOutcome {
	return BootstrapOutcome{
		kind: bootstrapOutcomeBootstrapFailure,
		bootstrapFailure: &BootstrapFailure{
			stage: stage,
			code:  code,
			cause: cause,
		},
	}
}

func completeBootstrapSnapshot(snapshot runtimeconfig.Snapshot) bool {
	return completeBootstrapProvenance(snapshot.Provenance())
}

func completeBootstrapProvenance(provenance runtimeconfig.Provenance) bool {
	return provenance.WorkspaceID != 0 &&
		provenance.ConfigurationID != 0 &&
		provenance.ConfigurationVersionID != 0 &&
		provenance.ConfigurationVersionNumber != 0 &&
		provenance.SchemaIdentity != "" &&
		provenance.SchemaVersion != 0 &&
		provenance.RuntimeInstanceID != "" &&
		provenance.LaunchAttemptID != ""
}

func isNilBootstrapCapability(capability any) bool {
	if capability == nil {
		return true
	}
	value := reflect.ValueOf(capability)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
