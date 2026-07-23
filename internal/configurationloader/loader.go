// Package configurationloader loads one exact Published ConfigurationVersion
// into detached source material for Runtime Snapshot construction.
package configurationloader

import (
	"errors"
	"strings"

	"github.com/dsdred/universal-websocket-platform/internal/configuration"
	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

var (
	// ErrInvalidLoadRequest indicates that a mandatory load identity is absent.
	ErrInvalidLoadRequest = errors.New("invalid configuration load request")
	// ErrSourceUnavailable indicates that the configured source could not complete the read.
	ErrSourceUnavailable = errors.New("configuration source unavailable")
	// ErrSourceNotFound indicates that a requested source entity does not exist.
	ErrSourceNotFound = errors.New("configuration source not found")
	// ErrIdentityMismatch indicates that loaded identities do not form the requested ownership chain.
	ErrIdentityMismatch = errors.New("configuration identity mismatch")
	// ErrVersionNotPublished indicates that the exact pinned version is not Published.
	ErrVersionNotPublished = errors.New("configuration version not published")
	// ErrInconsistentSourceObservation indicates that one logical source state could not be established.
	ErrInconsistentSourceObservation = errors.New("inconsistent configuration source observation")
	// ErrSourceIntegrity indicates that source material is malformed or representation-incomplete.
	ErrSourceIntegrity = errors.New("configuration source integrity failure")
)

// RuntimeInstanceID is the neutral contract identity of one Runtime Instance.
type RuntimeInstanceID = runtimeconfigload.RuntimeInstanceID

// LaunchAttemptID is the neutral contract identity of one Launch Attempt.
type LaunchAttemptID = runtimeconfigload.LaunchAttemptID

// LoadRequest is the neutral immutable Loader handoff request.
type LoadRequest = runtimeconfigload.LoadRequest

// DetachedLoadResult is the neutral immutable Loader-to-Builder handoff result.
type DetachedLoadResult = runtimeconfigload.DetachedLoadResult

// NewLoadRequest constructs an immutable request for one exact ConfigurationVersion.
func NewLoadRequest(
	workspaceID uint64,
	configurationID uint64,
	configurationVersionID uint64,
	runtimeInstanceID RuntimeInstanceID,
	launchAttemptID LaunchAttemptID,
) (LoadRequest, error) {
	request := runtimeconfigload.NewLoadRequest(
		workspaceID,
		configurationID,
		configurationVersionID,
		runtimeInstanceID,
		launchAttemptID,
	)
	if !request.Complete() {
		return LoadRequest{}, ErrInvalidLoadRequest
	}
	return request, nil
}

// SourceObservation is one detached, consistent observation supplied by a
// configured source boundary. Complete asserts representation completeness;
// Loader still validates identity, lifecycle, and mandatory schema facts.
type SourceObservation struct {
	WorkspaceID            uint64
	Configuration          configuration.Configuration
	ConfigurationVersion   configurationversion.ConfigurationVersion
	SchemaIdentity         string
	SchemaVersion          uint32
	RepresentationComplete bool
}

// Source loads one exact ConfigurationVersion together with its ownership and
// schema facts. It must not select a latest or replacement version.
type Source interface {
	LoadExact(workspaceID, configurationID, configurationVersionID uint64) (SourceObservation, error)
}

// Loader performs the single Load Exact Published Configuration operation.
type Loader struct {
	source Source
}

// New constructs a Loader over one configured source boundary.
func New(source Source) *Loader {
	return &Loader{source: source}
}

// Load loads and validates the exact pinned Published ConfigurationVersion.
func (l *Loader) Load(request LoadRequest) (DetachedLoadResult, error) {
	if !request.Complete() {
		return DetachedLoadResult{}, ErrInvalidLoadRequest
	}
	if l == nil || l.source == nil {
		return DetachedLoadResult{}, ErrSourceUnavailable
	}

	observation, err := l.source.LoadExact(
		request.WorkspaceID(),
		request.ConfigurationID(),
		request.ConfigurationVersionID(),
	)
	if err != nil {
		return DetachedLoadResult{}, normalizeSourceError(err)
	}
	if !observation.RepresentationComplete ||
		strings.TrimSpace(observation.SchemaIdentity) == "" ||
		observation.SchemaVersion == 0 ||
		observation.ConfigurationVersion.Number == 0 ||
		!validVersionState(observation.ConfigurationVersion.State) {
		return DetachedLoadResult{}, ErrSourceIntegrity
	}
	if observation.WorkspaceID != request.WorkspaceID() ||
		observation.Configuration.ID != request.ConfigurationID() ||
		observation.Configuration.WorkspaceID != request.WorkspaceID() ||
		observation.ConfigurationVersion.ID != request.ConfigurationVersionID() ||
		observation.ConfigurationVersion.ConfigurationID != request.ConfigurationID() {
		return DetachedLoadResult{}, ErrIdentityMismatch
	}
	if observation.ConfigurationVersion.State != configurationversion.Published {
		return DetachedLoadResult{}, ErrVersionNotPublished
	}

	return runtimeconfigload.NewDetachedLoadResult(
		request,
		observation.ConfigurationVersion,
		true,
		observation.SchemaIdentity,
		observation.SchemaVersion,
	), nil
}

func validVersionState(state configurationversion.VersionState) bool {
	switch state {
	case configurationversion.Draft,
		configurationversion.Validated,
		configurationversion.Published,
		configurationversion.Archived:
		return true
	default:
		return false
	}
}

func normalizeSourceError(err error) error {
	switch {
	case errors.Is(err, ErrSourceNotFound):
		return ErrSourceNotFound
	case errors.Is(err, ErrIdentityMismatch):
		return ErrIdentityMismatch
	case errors.Is(err, ErrVersionNotPublished):
		return ErrVersionNotPublished
	case errors.Is(err, ErrInconsistentSourceObservation):
		return ErrInconsistentSourceObservation
	case errors.Is(err, ErrSourceIntegrity):
		return ErrSourceIntegrity
	case errors.Is(err, ErrSourceUnavailable):
		return ErrSourceUnavailable
	default:
		return ErrSourceUnavailable
	}
}
