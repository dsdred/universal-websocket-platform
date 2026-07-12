package configurationversion

import (
	"context"
	"errors"
)

// ErrConfigurationVersionNotFound indicates that a Configuration Version does not exist.
var ErrConfigurationVersionNotFound = errors.New("configuration version not found")

// ErrConfigurationNotFound indicates that the parent Configuration does not exist.
var ErrConfigurationNotFound = errors.New("configuration not found")

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
	Delete(uint64) error
	GetPublished(uint64) (ConfigurationVersion, error)
}
