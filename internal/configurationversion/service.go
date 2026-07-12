package configurationversion

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Service applies Configuration Version business rules.
type Service struct {
	repository           ConfigurationVersionRepository
	configurationChecker ConfigurationExistenceChecker
	now                  func() time.Time
	lifecycleMu          sync.Mutex
}

// NewService creates a Configuration Version service.
func NewService(repository ConfigurationVersionRepository, configurationChecker ConfigurationExistenceChecker, now func() time.Time) *Service {
	return &Service{repository: repository, configurationChecker: configurationChecker, now: now}
}

// Create creates the next Draft Version for an existing Configuration.
func (s *Service) Create(ctx context.Context, workspaceID, configurationID uint64) (ConfigurationVersion, error) {
	if err := s.requireConfiguration(ctx, workspaceID, configurationID); err != nil {
		return ConfigurationVersion{}, err
	}

	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	versions, err := s.repository.ListByConfiguration(configurationID)
	if err != nil {
		return ConfigurationVersion{}, err
	}

	var nextNumber uint32 = 1
	for _, version := range versions {
		if version.Number >= nextNumber {
			nextNumber = version.Number + 1
		}
	}

	now := s.now().UTC()
	return s.repository.Create(ConfigurationVersion{
		ConfigurationID: configurationID,
		Number:          nextNumber,
		State:           Draft,
		CreatedAt:       now,
		UpdatedAt:       now,
	})
}

// Publish transitions a Draft Version to Published and archives the previous Published Version.
func (s *Service) Publish(ctx context.Context, workspaceID, configurationID, versionID uint64) (ConfigurationVersion, error) {
	if err := s.requireConfiguration(ctx, workspaceID, configurationID); err != nil {
		return ConfigurationVersion{}, err
	}

	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	target, err := s.repository.Get(versionID)
	if err != nil || target.ConfigurationID != configurationID {
		if err == nil || errors.Is(err, ErrConfigurationVersionNotFound) {
			return ConfigurationVersion{}, ErrConfigurationVersionNotFound
		}
		return ConfigurationVersion{}, err
	}
	if target.State != Draft {
		return ConfigurationVersion{}, ErrVersionNotPublishable
	}

	now := s.now().UTC()
	updates := make([]ConfigurationVersion, 0, 2)
	current, err := s.repository.GetPublished(configurationID)
	switch {
	case err == nil:
		current.State = Archived
		current.UpdatedAt = now
		updates = append(updates, current)
	case !errors.Is(err, ErrConfigurationVersionNotFound):
		return ConfigurationVersion{}, err
	}

	target.State = Published
	target.UpdatedAt = now
	updates = append(updates, target)
	if err := s.repository.UpdateBatch(updates); err != nil {
		return ConfigurationVersion{}, err
	}

	return target, nil
}

// Archive transitions a non-Archived Version to Archived.
func (s *Service) Archive(ctx context.Context, workspaceID, configurationID, versionID uint64) (ConfigurationVersion, error) {
	if err := s.requireConfiguration(ctx, workspaceID, configurationID); err != nil {
		return ConfigurationVersion{}, err
	}

	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	version, err := s.repository.Get(versionID)
	if err != nil || version.ConfigurationID != configurationID {
		if err == nil || errors.Is(err, ErrConfigurationVersionNotFound) {
			return ConfigurationVersion{}, ErrConfigurationVersionNotFound
		}
		return ConfigurationVersion{}, err
	}

	switch version.State {
	case Draft, Validated, Published:
	case Archived:
		return ConfigurationVersion{}, ErrVersionNotArchivable
	default:
		return ConfigurationVersion{}, ErrVersionNotArchivable
	}

	version.State = Archived
	version.UpdatedAt = s.now().UTC()
	return s.repository.Update(version)
}

// List returns all Versions for an existing Configuration.
func (s *Service) List(ctx context.Context, workspaceID, configurationID uint64) ([]ConfigurationVersion, error) {
	if err := s.requireConfiguration(ctx, workspaceID, configurationID); err != nil {
		return nil, err
	}
	return s.repository.ListByConfiguration(configurationID)
}

func (s *Service) requireConfiguration(ctx context.Context, workspaceID, configurationID uint64) error {
	exists, err := s.configurationChecker.Exists(ctx, workspaceID, configurationID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrConfigurationNotFound
	}
	return nil
}
