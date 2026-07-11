package configuration

import "time"

// Configuration describes metadata for a future server configuration.
type Configuration struct {
	ID          uint64    `json:"id"`
	WorkspaceID uint64    `json:"workspaceId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateConfiguration contains values accepted when creating a Configuration.
type CreateConfiguration struct {
	Name        string
	Description string
}

// UpdateConfiguration contains values accepted when updating a Configuration.
type UpdateConfiguration struct {
	Name        string
	Description string
}
