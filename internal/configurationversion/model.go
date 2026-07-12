package configurationversion

import "time"

// VersionState represents the lifecycle state of a Configuration Version.
type VersionState string

const (
	Draft     VersionState = "Draft"
	Validated VersionState = "Validated"
	Published VersionState = "Published"
	Archived  VersionState = "Archived"
)

// ConfigurationVersion contains metadata for a Configuration Version.
type ConfigurationVersion struct {
	ID              uint64       `json:"id"`
	ConfigurationID uint64       `json:"configurationId"`
	Number          uint32       `json:"number"`
	State           VersionState `json:"state"`
	CreatedAt       time.Time    `json:"createdAt"`
	UpdatedAt       time.Time    `json:"updatedAt"`
}
