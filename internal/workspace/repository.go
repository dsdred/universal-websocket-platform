package workspace

import (
	"context"
	"errors"
)

// ErrWorkspaceNotFound indicates that a Workspace does not exist.
var ErrWorkspaceNotFound = errors.New("workspace not found")

// ErrWorkspaceNotEmpty indicates that a Workspace contains Configurations.
var ErrWorkspaceNotEmpty = errors.New("workspace contains configurations")

// ConfigurationExistenceChecker checks whether a Workspace contains Configurations.
type ConfigurationExistenceChecker interface {
	ExistsByWorkspace(context.Context, uint64) (bool, error)
}

// WorkspaceRepository stores Workspace entities.
type WorkspaceRepository interface {
	Create(Workspace) (Workspace, error)
	Get(uint64) (Workspace, error)
	List() ([]Workspace, error)
	Update(Workspace) (Workspace, error)
	Delete(uint64) error
}
