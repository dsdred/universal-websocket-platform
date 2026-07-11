package configuration

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type workspaceCheckerStub struct {
	existing map[uint64]bool
	err      error
}

func (s workspaceCheckerStub) Exists(_ context.Context, id uint64) (bool, error) {
	return s.existing[id], s.err
}

func TestServiceCreate(t *testing.T) {
	now := time.Date(2026, 7, 12, 17, 0, 0, 0, time.FixedZone("test", 5*60*60))
	service := NewService(
		NewMemoryConfigurationRepository(),
		workspaceCheckerStub{existing: map[uint64]bool{1: true}},
		func() time.Time { return now },
	)

	configuration, err := service.Create(context.Background(), 1, CreateConfiguration{
		Name:        "  Notification Server  ",
		Description: "Realtime notifications",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if configuration.ID != 1 || configuration.WorkspaceID != 1 || configuration.Name != "Notification Server" {
		t.Errorf("Create() = %#v", configuration)
	}
	if configuration.CreatedAt.Location() != time.UTC || !configuration.CreatedAt.Equal(now) {
		t.Errorf("Create() CreatedAt = %s, want UTC equivalent of %s", configuration.CreatedAt, now)
	}
	if !configuration.UpdatedAt.Equal(configuration.CreatedAt) {
		t.Errorf("Create() UpdatedAt = %s, want %s", configuration.UpdatedAt, configuration.CreatedAt)
	}
}

func TestServiceWorkspaceNotFound(t *testing.T) {
	service := NewService(NewMemoryConfigurationRepository(), workspaceCheckerStub{existing: map[uint64]bool{}}, time.Now)

	if _, err := service.Create(context.Background(), 42, CreateConfiguration{Name: "Valid"}); !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("Create() error = %v, want ErrWorkspaceNotFound", err)
	}
	if _, err := service.List(context.Background(), 42); !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("List() error = %v, want ErrWorkspaceNotFound", err)
	}
}

func TestServiceValidation(t *testing.T) {
	service := NewService(
		NewMemoryConfigurationRepository(),
		workspaceCheckerStub{existing: map[uint64]bool{1: true}},
		time.Now,
	)
	tests := []struct {
		name        string
		input       CreateConfiguration
		wantField   string
		wantMessage string
	}{
		{name: "empty name", input: CreateConfiguration{}, wantField: "name", wantMessage: "must not be empty"},
		{name: "whitespace name", input: CreateConfiguration{Name: " \t\n "}, wantField: "name", wantMessage: "must not be empty"},
		{name: "long unicode name", input: CreateConfiguration{Name: strings.Repeat("я", 101)}, wantField: "name", wantMessage: "100"},
		{name: "long unicode description", input: CreateConfiguration{Name: "Valid", Description: strings.Repeat("я", 1001)}, wantField: "description", wantMessage: "1000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Create(context.Background(), 1, tt.input)
			var validationError *ValidationError
			if !errors.As(err, &validationError) {
				t.Fatalf("Create() error = %v, want ValidationError", err)
			}
			if validationError.Field != tt.wantField || !strings.Contains(validationError.Message, tt.wantMessage) {
				t.Errorf("ValidationError = %#v", validationError)
			}
		})
	}
}

func TestServiceUpdatePreservesImmutableFields(t *testing.T) {
	createdAt := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	times := []time.Time{createdAt, updatedAt}
	service := NewService(
		NewMemoryConfigurationRepository(),
		workspaceCheckerStub{existing: map[uint64]bool{1: true, 2: true}},
		func() time.Time {
			now := times[0]
			times = times[1:]
			return now
		},
	)

	created, err := service.Create(context.Background(), 1, CreateConfiguration{Name: "Original"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	updated, err := service.Update(context.Background(), 1, created.ID, UpdateConfiguration{Name: "Updated"})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.ID != created.ID || updated.WorkspaceID != created.WorkspaceID || !updated.CreatedAt.Equal(created.CreatedAt) {
		t.Errorf("Update() changed immutable fields: %#v", updated)
	}
	if !updated.UpdatedAt.Equal(updatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt) {
		t.Errorf("Update() UpdatedAt = %s, want %s", updated.UpdatedAt, updatedAt)
	}

	if _, err := service.Get(context.Background(), 2, created.ID); !errors.Is(err, ErrConfigurationNotFound) {
		t.Errorf("Get(other Workspace) error = %v, want ErrConfigurationNotFound", err)
	}
}

func TestServiceDeleteAndNotFound(t *testing.T) {
	service := NewService(
		NewMemoryConfigurationRepository(),
		workspaceCheckerStub{existing: map[uint64]bool{1: true}},
		time.Now,
	)
	created, err := service.Create(context.Background(), 1, CreateConfiguration{Name: "Delete me"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := service.Delete(context.Background(), 1, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := service.Get(context.Background(), 1, created.ID); !errors.Is(err, ErrConfigurationNotFound) {
		t.Errorf("Get(deleted) error = %v, want ErrConfigurationNotFound", err)
	}
	if err := service.Delete(context.Background(), 1, 42); !errors.Is(err, ErrConfigurationNotFound) {
		t.Errorf("Delete(missing) error = %v, want ErrConfigurationNotFound", err)
	}
}
