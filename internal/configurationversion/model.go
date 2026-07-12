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

// ListenerSettings describes where a future WebSocket Listener will accept connections.
type ListenerSettings struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

// ConfigurationVersion contains metadata for a Configuration Version.
type ConfigurationVersion struct {
	ID              uint64           `json:"id"`
	ConfigurationID uint64           `json:"configurationId"`
	Number          uint32           `json:"number"`
	State           VersionState     `json:"state"`
	Listener        ListenerSettings `json:"listener"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
}
