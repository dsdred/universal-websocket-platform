package configurationversion

import (
	"sort"
	"sync"
)

// MemoryConfigurationVersionRepository stores Configuration Versions in process memory.
type MemoryConfigurationVersionRepository struct {
	mu       sync.RWMutex
	versions map[uint64]ConfigurationVersion
	nextID   uint64
}

// NewMemoryConfigurationVersionRepository creates an empty in-memory repository.
func NewMemoryConfigurationVersionRepository() *MemoryConfigurationVersionRepository {
	return &MemoryConfigurationVersionRepository{
		versions: make(map[uint64]ConfigurationVersion),
		nextID:   1,
	}
}

// Create stores a Configuration Version and assigns its ID.
func (r *MemoryConfigurationVersionRepository) Create(version ConfigurationVersion) (ConfigurationVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	version.ID = r.nextID
	r.nextID++
	r.versions[version.ID] = version
	return version, nil
}

// Get returns a Configuration Version by ID.
func (r *MemoryConfigurationVersionRepository) Get(id uint64) (ConfigurationVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	version, exists := r.versions[id]
	if !exists {
		return ConfigurationVersion{}, ErrConfigurationVersionNotFound
	}
	return version, nil
}

// ListByConfiguration returns Configuration Versions ordered by Number.
func (r *MemoryConfigurationVersionRepository) ListByConfiguration(configurationID uint64) ([]ConfigurationVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions := make([]ConfigurationVersion, 0)
	for _, version := range r.versions {
		if version.ConfigurationID == configurationID {
			versions = append(versions, version)
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Number < versions[j].Number
	})
	return versions, nil
}

// Update replaces an existing Configuration Version.
func (r *MemoryConfigurationVersionRepository) Update(version ConfigurationVersion) (ConfigurationVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.versions[version.ID]; !exists {
		return ConfigurationVersion{}, ErrConfigurationVersionNotFound
	}
	r.versions[version.ID] = version
	return version, nil
}

// UpdateBatch atomically replaces existing Configuration Versions.
func (r *MemoryConfigurationVersionRepository) UpdateBatch(versions []ConfigurationVersion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, version := range versions {
		if _, exists := r.versions[version.ID]; !exists {
			return ErrConfigurationVersionNotFound
		}
	}
	for _, version := range versions {
		r.versions[version.ID] = version
	}
	return nil
}

// Delete removes a Configuration Version by ID.
func (r *MemoryConfigurationVersionRepository) Delete(id uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.versions[id]; !exists {
		return ErrConfigurationVersionNotFound
	}
	delete(r.versions, id)
	return nil
}

// GetPublished returns the Published Version of a Configuration.
func (r *MemoryConfigurationVersionRepository) GetPublished(configurationID uint64) (ConfigurationVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, version := range r.versions {
		if version.ConfigurationID == configurationID && version.State == Published {
			return version, nil
		}
	}
	return ConfigurationVersion{}, ErrConfigurationVersionNotFound
}
