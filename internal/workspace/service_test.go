package workspace

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type configurationCheckerStub struct {
	exists bool
	err    error
}

func (s configurationCheckerStub) ExistsByWorkspace(context.Context, uint64) (bool, error) {
	return s.exists, s.err
}

func TestWorkspaceServiceCreate(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	service := NewWorkspaceService(NewMemoryWorkspaceRepository(), nil, func() time.Time { return now })

	workspace, err := service.Create(CreateWorkspace{Name: "  Main Workspace  ", Description: "Primary"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if workspace.ID != 1 || workspace.Name != "Main Workspace" || workspace.Description != "Primary" {
		t.Errorf("Create() = %#v", workspace)
	}
	if !workspace.CreatedAt.Equal(now) || !workspace.UpdatedAt.Equal(now) {
		t.Errorf("Create() timestamps = %s, %s; want %s", workspace.CreatedAt, workspace.UpdatedAt, now)
	}
}

func TestWorkspaceServiceValidation(t *testing.T) {
	service := NewWorkspaceService(NewMemoryWorkspaceRepository(), nil, time.Now)
	tests := []struct {
		name        string
		input       CreateWorkspace
		wantField   string
		wantMessage string
	}{
		{name: "empty name", input: CreateWorkspace{Name: ""}, wantField: "name", wantMessage: "must not be empty"},
		{name: "whitespace name", input: CreateWorkspace{Name: " \t\n "}, wantField: "name", wantMessage: "must not be empty"},
		{name: "long unicode name", input: CreateWorkspace{Name: strings.Repeat("я", 101)}, wantField: "name", wantMessage: "100"},
		{name: "long unicode description", input: CreateWorkspace{Name: "Valid", Description: strings.Repeat("я", 1001)}, wantField: "description", wantMessage: "1000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Create(tt.input)
			var validationError *ValidationError
			if !errors.As(err, &validationError) {
				t.Fatalf("Create() error = %v, want ValidationError", err)
			}
			if validationError.Field != tt.wantField || !strings.Contains(validationError.Message, tt.wantMessage) {
				t.Errorf("ValidationError = %#v, want field %q and message containing %q", validationError, tt.wantField, tt.wantMessage)
			}
		})
	}
}

func TestWorkspaceServiceUpdatePreservesImmutableFields(t *testing.T) {
	createdAt := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	times := []time.Time{createdAt, updatedAt}
	service := NewWorkspaceService(NewMemoryWorkspaceRepository(), nil, func() time.Time {
		now := times[0]
		times = times[1:]
		return now
	})

	created, err := service.Create(CreateWorkspace{Name: "Main"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	updated, err := service.Update(created.ID, UpdateWorkspace{Name: "Updated", Description: "Changed"})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updated.ID != created.ID {
		t.Errorf("Update() ID = %d, want %d", updated.ID, created.ID)
	}
	if !updated.CreatedAt.Equal(created.CreatedAt) {
		t.Errorf("Update() CreatedAt = %s, want %s", updated.CreatedAt, created.CreatedAt)
	}
	if !updated.UpdatedAt.Equal(updatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt) {
		t.Errorf("Update() UpdatedAt = %s, want changed to %s", updated.UpdatedAt, updatedAt)
	}
}

func TestWorkspaceServiceNotFound(t *testing.T) {
	service := NewWorkspaceService(NewMemoryWorkspaceRepository(), nil, time.Now)

	if _, err := service.Get(42); !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("Get() error = %v, want ErrWorkspaceNotFound", err)
	}
	if _, err := service.Update(42, UpdateWorkspace{Name: "Valid"}); !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("Update() error = %v, want ErrWorkspaceNotFound", err)
	}
	if err := service.Delete(context.Background(), 42); !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("Delete() error = %v, want ErrWorkspaceNotFound", err)
	}
}

func TestWorkspaceServiceDeleteNonEmpty(t *testing.T) {
	repository := NewMemoryWorkspaceRepository()
	service := NewWorkspaceService(repository, configurationCheckerStub{exists: true}, time.Now)
	created, err := service.Create(CreateWorkspace{Name: "Non-empty"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := service.Delete(context.Background(), created.ID); !errors.Is(err, ErrWorkspaceNotEmpty) {
		t.Errorf("Delete() error = %v, want ErrWorkspaceNotEmpty", err)
	}
	if _, err := service.Get(created.ID); err != nil {
		t.Errorf("Get() after blocked Delete error = %v", err)
	}
}
