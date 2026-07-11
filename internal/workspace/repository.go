package workspace

import "errors"

// ErrWorkspaceNotFound indicates that a Workspace does not exist.
var ErrWorkspaceNotFound = errors.New("workspace not found")

// WorkspaceRepository stores Workspace entities.
type WorkspaceRepository interface {
	Create(Workspace) (Workspace, error)
	Get(uint64) (Workspace, error)
	List() ([]Workspace, error)
	Update(Workspace) (Workspace, error)
	Delete(uint64) error
}
