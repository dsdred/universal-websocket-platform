package configurationversion

import (
	"context"
	"errors"
)

// ErrConfigurationVersionNotFound indicates that a Configuration Version does not exist.
var ErrConfigurationVersionNotFound = errors.New("configuration version not found")

// ErrConfigurationNotFound indicates that the parent Configuration does not exist.
var ErrConfigurationNotFound = errors.New("configuration not found")

// ErrVersionNotPublishable indicates that a Version cannot transition to Published.
var ErrVersionNotPublishable = errors.New("configuration version is not publishable")

// ErrVersionNotArchivable indicates that a Version cannot transition to Archived.
var ErrVersionNotArchivable = errors.New("configuration version is not archivable")

// ConfigurationExistenceChecker checks whether a Configuration exists in a Workspace.
type ConfigurationExistenceChecker interface {
	Exists(context.Context, uint64, uint64) (bool, error)
}

// ConfigurationVersionRepository stores Configuration Version entities.
type ConfigurationVersionRepository interface {
	Create(ConfigurationVersion) (ConfigurationVersion, error)
	Get(uint64) (ConfigurationVersion, error)
	ListByConfiguration(uint64) ([]ConfigurationVersion, error)
	Update(ConfigurationVersion) (ConfigurationVersion, error)
	UpdateBatch([]ConfigurationVersion) error
	Delete(uint64) error
	GetPublished(uint64) (ConfigurationVersion, error)
}
