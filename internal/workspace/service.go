package workspace

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxNameLength        = 100
	maxDescriptionLength = 1000
)

// ValidationError describes invalid Workspace data.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// WorkspaceService applies Workspace business rules.
type WorkspaceService struct {
	repository WorkspaceRepository
	now        func() time.Time
}

// NewWorkspaceService creates a Workspace service.
func NewWorkspaceService(repository WorkspaceRepository, now func() time.Time) *WorkspaceService {
	return &WorkspaceService{repository: repository, now: now}
}

// Create validates and creates a Workspace.
func (s *WorkspaceService) Create(input CreateWorkspace) (Workspace, error) {
	name, err := validateWorkspace(input.Name, input.Description)
	if err != nil {
		return Workspace{}, err
	}

	now := s.now()
	return s.repository.Create(Workspace{
		Name:        name,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

// Get returns a Workspace by ID.
func (s *WorkspaceService) Get(id uint64) (Workspace, error) {
	return s.repository.Get(id)
}

// List returns all Workspaces ordered by ID.
func (s *WorkspaceService) List() ([]Workspace, error) {
	return s.repository.List()
}

// Update validates and updates a Workspace while preserving immutable fields.
func (s *WorkspaceService) Update(id uint64, input UpdateWorkspace) (Workspace, error) {
	name, err := validateWorkspace(input.Name, input.Description)
	if err != nil {
		return Workspace{}, err
	}

	existing, err := s.repository.Get(id)
	if err != nil {
		return Workspace{}, err
	}

	existing.Name = name
	existing.Description = input.Description
	existing.UpdatedAt = s.now()

	return s.repository.Update(existing)
}

// Delete removes a Workspace by ID.
func (s *WorkspaceService) Delete(id uint64) error {
	return s.repository.Delete(id)
}

func validateWorkspace(name, description string) (string, error) {
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
