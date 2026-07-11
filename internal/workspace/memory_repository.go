package workspace

import (
	"context"
	"sort"
	"sync"
)

// MemoryWorkspaceRepository stores Workspaces in process memory.
type MemoryWorkspaceRepository struct {
	mu         sync.RWMutex
	workspaces map[uint64]Workspace
	nextID     uint64
}

// Exists reports whether a Workspace exists.
func (r *MemoryWorkspaceRepository) Exists(_ context.Context, id uint64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.workspaces[id]
	return exists, nil
}

// NewMemoryWorkspaceRepository creates an empty in-memory repository.
func NewMemoryWorkspaceRepository() *MemoryWorkspaceRepository {
	return &MemoryWorkspaceRepository{
		workspaces: make(map[uint64]Workspace),
		nextID:     1,
	}
}

// Create stores a new Workspace and assigns its ID.
func (r *MemoryWorkspaceRepository) Create(workspace Workspace) (Workspace, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	workspace.ID = r.nextID
	r.nextID++
	r.workspaces[workspace.ID] = workspace

	return workspace, nil
}

// Get returns a Workspace by ID.
func (r *MemoryWorkspaceRepository) Get(id uint64) (Workspace, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	workspace, exists := r.workspaces[id]
	if !exists {
		return Workspace{}, ErrWorkspaceNotFound
	}

	return workspace, nil
}

// List returns all Workspaces ordered by ID.
func (r *MemoryWorkspaceRepository) List() ([]Workspace, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	workspaces := make([]Workspace, 0, len(r.workspaces))
	for _, workspace := range r.workspaces {
		workspaces = append(workspaces, workspace)
	}

	sort.Slice(workspaces, func(i, j int) bool {
		return workspaces[i].ID < workspaces[j].ID
	})

	return workspaces, nil
}

// Update replaces an existing Workspace.
func (r *MemoryWorkspaceRepository) Update(workspace Workspace) (Workspace, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.workspaces[workspace.ID]; !exists {
		return Workspace{}, ErrWorkspaceNotFound
	}

	r.workspaces[workspace.ID] = workspace
	return workspace, nil
}

// Delete removes a Workspace by ID.
func (r *MemoryWorkspaceRepository) Delete(id uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.workspaces[id]; !exists {
		return ErrWorkspaceNotFound
	}

	delete(r.workspaces, id)
	return nil
}
