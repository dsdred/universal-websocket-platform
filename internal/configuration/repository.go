package configuration

import (
	"context"
	"errors"
)

// ErrConfigurationNotFound indicates that a Configuration does not exist in the requested Workspace.
var ErrConfigurationNotFound = errors.New("configuration not found")

// ErrWorkspaceNotFound indicates that the parent Workspace does not exist.
var ErrWorkspaceNotFound = errors.New("workspace not found")

// WorkspaceExistenceChecker checks whether a Workspace exists.
type WorkspaceExistenceChecker interface {
	Exists(context.Context, uint64) (bool, error)
}

// ConfigurationRepository stores Configuration entities.
type ConfigurationRepository interface {
	Create(Configuration) (Configuration, error)
	Get(uint64) (Configuration, error)
	ListByWorkspace(uint64) ([]Configuration, error)
	Update(Configuration) (Configuration, error)
	Delete(uint64) error
	ExistsByWorkspace(context.Context, uint64) (bool, error)
}
