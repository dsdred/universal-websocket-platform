// Package runtimeconfigload defines the neutral immutable handoff contract
// between Runtime Lifecycle Owner, Configuration Loader, and Builder.
package runtimeconfigload

import "github.com/dsdred/universal-websocket-platform/internal/configurationversion"

// RuntimeInstanceID is the opaque operational identity of one Runtime Instance.
type RuntimeInstanceID string

// LaunchAttemptID is the opaque execution identity of one Launch Attempt.
type LaunchAttemptID string

// LoadRequest identifies one exact configuration source and its launch provenance.
// Its fields are immutable after construction.
type LoadRequest struct {
	workspaceID            uint64
	configurationID        uint64
	configurationVersionID uint64
	runtimeInstanceID      RuntimeInstanceID
	launchAttemptID        LaunchAttemptID
}

// NewLoadRequest constructs an immutable request for one exact ConfigurationVersion.
func NewLoadRequest(
	workspaceID uint64,
	configurationID uint64,
	configurationVersionID uint64,
	runtimeInstanceID RuntimeInstanceID,
	launchAttemptID LaunchAttemptID,
) LoadRequest {
	return LoadRequest{
		workspaceID:            workspaceID,
		configurationID:        configurationID,
		configurationVersionID: configurationVersionID,
		runtimeInstanceID:      runtimeInstanceID,
		launchAttemptID:        launchAttemptID,
	}
}

// WorkspaceID returns the requested Workspace identity.
func (r LoadRequest) WorkspaceID() uint64 {
	return r.workspaceID
}

// ConfigurationID returns the requested Configuration identity.
func (r LoadRequest) ConfigurationID() uint64 {
	return r.configurationID
}

// ConfigurationVersionID returns the exact pinned ConfigurationVersion identity.
func (r LoadRequest) ConfigurationVersionID() uint64 {
	return r.configurationVersionID
}

// RuntimeInstanceID returns the Runtime Instance provenance identity.
func (r LoadRequest) RuntimeInstanceID() RuntimeInstanceID {
	return r.runtimeInstanceID
}

// LaunchAttemptID returns the Launch Attempt provenance identity.
func (r LoadRequest) LaunchAttemptID() LaunchAttemptID {
	return r.launchAttemptID
}

// Complete reports whether every mandatory identity is present.
func (r LoadRequest) Complete() bool {
	return r.workspaceID != 0 &&
		r.configurationID != 0 &&
		r.configurationVersionID != 0 &&
		r.runtimeInstanceID != "" &&
		r.launchAttemptID != ""
}

// DetachedLoadResult is immutable source material for one successful load.
// Accessors that return declarative material return detached copies.
type DetachedLoadResult struct {
	workspaceID            uint64
	configurationID        uint64
	configurationVersionID uint64
	configurationNumber    uint32
	published              bool
	schemaIdentity         string
	schemaVersion          uint32
	runtimeInstanceID      RuntimeInstanceID
	launchAttemptID        LaunchAttemptID
	configurationVersion   configurationversion.ConfigurationVersion
}

// NewDetachedLoadResult constructs an immutable result from detached load facts.
func NewDetachedLoadResult(
	request LoadRequest,
	version configurationversion.ConfigurationVersion,
	published bool,
	schemaIdentity string,
	schemaVersion uint32,
) DetachedLoadResult {
	return DetachedLoadResult{
		workspaceID:            request.workspaceID,
		configurationID:        request.configurationID,
		configurationVersionID: request.configurationVersionID,
		configurationNumber:    version.Number,
		published:              published,
		schemaIdentity:         schemaIdentity,
		schemaVersion:          schemaVersion,
		runtimeInstanceID:      request.runtimeInstanceID,
		launchAttemptID:        request.launchAttemptID,
		configurationVersion:   cloneConfigurationVersion(version),
	}
}

// WorkspaceID returns the verified Workspace identity.
func (r DetachedLoadResult) WorkspaceID() uint64 {
	return r.workspaceID
}

// ConfigurationID returns the verified Configuration identity.
func (r DetachedLoadResult) ConfigurationID() uint64 {
	return r.configurationID
}

// ConfigurationVersionID returns the exact loaded ConfigurationVersion identity.
func (r DetachedLoadResult) ConfigurationVersionID() uint64 {
	return r.configurationVersionID
}

// ConfigurationVersionNumber returns the loaded ConfigurationVersion number.
func (r DetachedLoadResult) ConfigurationVersionNumber() uint32 {
	return r.configurationNumber
}

// Published reports the lifecycle fact observed for the loaded version.
func (r DetachedLoadResult) Published() bool {
	return r.published
}

// SchemaIdentity returns the preserved Configuration schema identity.
func (r DetachedLoadResult) SchemaIdentity() string {
	return r.schemaIdentity
}

// SchemaVersion returns the preserved Configuration schema version.
func (r DetachedLoadResult) SchemaVersion() uint32 {
	return r.schemaVersion
}

// RuntimeInstanceID returns the preserved Runtime Instance identity.
func (r DetachedLoadResult) RuntimeInstanceID() RuntimeInstanceID {
	return r.runtimeInstanceID
}

// LaunchAttemptID returns the preserved Launch Attempt identity.
func (r DetachedLoadResult) LaunchAttemptID() LaunchAttemptID {
	return r.launchAttemptID
}

// ConfigurationVersion returns a detached copy of the complete declarative payload.
func (r DetachedLoadResult) ConfigurationVersion() configurationversion.ConfigurationVersion {
	return cloneConfigurationVersion(r.configurationVersion)
}
