package configurationversion

import (
	"context"
	"sync"
	"time"
)

// Service applies Configuration Version business rules.
type Service struct {
	repository           ConfigurationVersionRepository
	configurationChecker ConfigurationExistenceChecker
	now                  func() time.Time
	numberMu             sync.Mutex
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

	s.numberMu.Lock()
	defer s.numberMu.Unlock()

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
