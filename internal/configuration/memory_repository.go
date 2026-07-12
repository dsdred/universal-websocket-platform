package configuration

import (
	"context"
	"sort"
	"sync"
)

// MemoryConfigurationRepository stores Configurations in process memory.
type MemoryConfigurationRepository struct {
	mu             sync.RWMutex
	configurations map[uint64]Configuration
	nextID         uint64
}

// NewMemoryConfigurationRepository creates an empty in-memory repository.
func NewMemoryConfigurationRepository() *MemoryConfigurationRepository {
	return &MemoryConfigurationRepository{
		configurations: make(map[uint64]Configuration),
		nextID:         1,
	}
}

// Create stores a new Configuration and assigns its ID.
func (r *MemoryConfigurationRepository) Create(configuration Configuration) (Configuration, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	configuration.ID = r.nextID
	r.nextID++
	r.configurations[configuration.ID] = configuration
	return configuration, nil
}

// Get returns a Configuration by ID.
func (r *MemoryConfigurationRepository) Get(id uint64) (Configuration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	configuration, exists := r.configurations[id]
	if !exists {
		return Configuration{}, ErrConfigurationNotFound
	}

	return configuration, nil
}

// ListByWorkspace returns Configurations for a Workspace ordered by ID.
func (r *MemoryConfigurationRepository) ListByWorkspace(workspaceID uint64) ([]Configuration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	configurations := make([]Configuration, 0)
	for _, configuration := range r.configurations {
		if configuration.WorkspaceID == workspaceID {
			configurations = append(configurations, configuration)
		}
	}

	sort.Slice(configurations, func(i, j int) bool {
		return configurations[i].ID < configurations[j].ID
	})

	return configurations, nil
}

// Update replaces an existing Configuration.
func (r *MemoryConfigurationRepository) Update(configuration Configuration) (Configuration, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configurations[configuration.ID]; !exists {
		return Configuration{}, ErrConfigurationNotFound
	}

	r.configurations[configuration.ID] = configuration
	return configuration, nil
}

// Delete removes a Configuration by ID.
func (r *MemoryConfigurationRepository) Delete(id uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configurations[id]; !exists {
		return ErrConfigurationNotFound
	}

	delete(r.configurations, id)
	return nil
}

// ExistsByWorkspace reports whether a Workspace contains any Configuration.
func (r *MemoryConfigurationRepository) ExistsByWorkspace(_ context.Context, workspaceID uint64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, configuration := range r.configurations {
		if configuration.WorkspaceID == workspaceID {
			return true, nil
		}
	}

	return false, nil
}

// Exists reports whether a Configuration exists in a Workspace.
func (r *MemoryConfigurationRepository) Exists(_ context.Context, workspaceID, configurationID uint64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	configuration, exists := r.configurations[configurationID]
	return exists && configuration.WorkspaceID == workspaceID, nil
}
