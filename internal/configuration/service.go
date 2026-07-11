package configuration

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxNameLength        = 100
	maxDescriptionLength = 1000
)

// ValidationError describes invalid Configuration data.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Service applies Configuration business rules.
type Service struct {
	repository       ConfigurationRepository
	workspaceChecker WorkspaceExistenceChecker
	now              func() time.Time
}

// NewService creates a Configuration service.
func NewService(repository ConfigurationRepository, workspaceChecker WorkspaceExistenceChecker, now func() time.Time) *Service {
	return &Service{repository: repository, workspaceChecker: workspaceChecker, now: now}
}

// Create validates and creates a Configuration in an existing Workspace.
func (s *Service) Create(ctx context.Context, workspaceID uint64, input CreateConfiguration) (Configuration, error) {
	if err := s.requireWorkspace(ctx, workspaceID); err != nil {
		return Configuration{}, err
	}

	name, err := validateConfiguration(input.Name, input.Description)
	if err != nil {
		return Configuration{}, err
	}

	now := s.now().UTC()
	return s.repository.Create(Configuration{
		WorkspaceID: workspaceID,
		Name:        name,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

// Get returns a Configuration scoped to a Workspace.
func (s *Service) Get(ctx context.Context, workspaceID, configurationID uint64) (Configuration, error) {
	if err := s.requireWorkspace(ctx, workspaceID); err != nil {
		return Configuration{}, err
	}

	configuration, err := s.repository.Get(configurationID)
	if err != nil {
		return Configuration{}, err
	}
	if configuration.WorkspaceID != workspaceID {
		return Configuration{}, ErrConfigurationNotFound
	}

	return configuration, nil
}

// List returns all Configurations in an existing Workspace.
func (s *Service) List(ctx context.Context, workspaceID uint64) ([]Configuration, error) {
	if err := s.requireWorkspace(ctx, workspaceID); err != nil {
		return nil, err
	}

	return s.repository.ListByWorkspace(workspaceID)
}

// Update validates and updates a Configuration without changing immutable fields.
func (s *Service) Update(ctx context.Context, workspaceID, configurationID uint64, input UpdateConfiguration) (Configuration, error) {
	if err := s.requireWorkspace(ctx, workspaceID); err != nil {
		return Configuration{}, err
	}

	name, err := validateConfiguration(input.Name, input.Description)
	if err != nil {
		return Configuration{}, err
	}

	existing, err := s.repository.Get(configurationID)
	if err != nil {
		return Configuration{}, err
	}
	if existing.WorkspaceID != workspaceID {
		return Configuration{}, ErrConfigurationNotFound
	}

	existing.Name = name
	existing.Description = input.Description
	existing.UpdatedAt = s.now().UTC()
	return s.repository.Update(existing)
}

// Delete removes a Configuration scoped to a Workspace.
func (s *Service) Delete(ctx context.Context, workspaceID, configurationID uint64) error {
	if err := s.requireWorkspace(ctx, workspaceID); err != nil {
		return err
	}

	existing, err := s.repository.Get(configurationID)
	if err != nil {
		return err
	}
	if existing.WorkspaceID != workspaceID {
		return ErrConfigurationNotFound
	}

	return s.repository.Delete(configurationID)
}

func (s *Service) requireWorkspace(ctx context.Context, workspaceID uint64) error {
	exists, err := s.workspaceChecker.Exists(ctx, workspaceID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrWorkspaceNotFound
	}
	return nil
}

func validateConfiguration(name, description string) (string, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return "", &ValidationError{Field: "name", Message: "must not be empty"}
	}
	if utf8.RuneCountInString(trimmedName) > maxNameLength {
		return "", &ValidationError{Field: "name", Message: "must not exceed 100 characters"}
	}
	if utf8.RuneCountInString(description) > maxDescriptionLength {
		return "", &ValidationError{Field: "description", Message: "must not exceed 1000 characters"}
	}

	return trimmedName, nil
}
