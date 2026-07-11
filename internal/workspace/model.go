package workspace

import "time"

// Workspace logically groups platform resources.
type Workspace struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateWorkspace contains values accepted when creating a Workspace.
type CreateWorkspace struct {
	Name        string
	Description string
}

// UpdateWorkspace contains values accepted when updating a Workspace.
type UpdateWorkspace struct {
	Name        string
	Description string
}
