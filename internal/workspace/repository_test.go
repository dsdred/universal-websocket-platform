package workspace

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestMemoryWorkspaceRepositoryCRUD(t *testing.T) {
	repository := NewMemoryWorkspaceRepository()
	createdAt := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	workspace := Workspace{Name: "Main", Description: "Primary", CreatedAt: createdAt, UpdatedAt: createdAt}

	created, err := repository.Create(workspace)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID != 1 {
		t.Errorf("Create() ID = %d, want 1", created.ID)
	}
	if !created.CreatedAt.Equal(createdAt) {
		t.Errorf("Create() CreatedAt = %s, want %s", created.CreatedAt, createdAt)
	}

	got, err := repository.Get(created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != created {
		t.Errorf("Get() = %#v, want %#v", got, created)
	}

	second, err := repository.Create(Workspace{Name: "Second", CreatedAt: createdAt, UpdatedAt: createdAt})
	if err != nil {
		t.Fatalf("Create(second) error = %v", err)
	}
	listed, err := repository.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(listed) != 2 || listed[0].ID != created.ID || listed[1].ID != second.ID {
		t.Errorf("List() IDs = %v, want [1 2]", workspaceIDs(listed))
	}

	updatedAt := createdAt.Add(time.Hour)
	created.Name = "Updated"
	created.UpdatedAt = updatedAt
	updated, err := repository.Update(created)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name != "Updated" || !updated.CreatedAt.Equal(createdAt) || !updated.UpdatedAt.Equal(updatedAt) {
		t.Errorf("Update() = %#v", updated)
	}

	if err := repository.Delete(created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repository.Get(created.ID); !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("Get(deleted) error = %v, want ErrWorkspaceNotFound", err)
	}
}

func TestMemoryWorkspaceRepositoryNotFound(t *testing.T) {
	repository := NewMemoryWorkspaceRepository()

	if _, err := repository.Get(42); !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("Get() error = %v, want ErrWorkspaceNotFound", err)
	}
	if _, err := repository.Update(Workspace{ID: 42}); !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("Update() error = %v, want ErrWorkspaceNotFound", err)
	}
	if err := repository.Delete(42); !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("Delete() error = %v, want ErrWorkspaceNotFound", err)
	}
}

func TestMemoryWorkspaceRepositoryConcurrentAccess(t *testing.T) {
	repository := NewMemoryWorkspaceRepository()
	const count = 100

	var waitGroup sync.WaitGroup
	waitGroup.Add(count)
	for range count {
		go func() {
			defer waitGroup.Done()
			created, err := repository.Create(Workspace{Name: "Concurrent"})
			if err != nil {
				t.Errorf("Create() error = %v", err)
				return
			}
			if _, err := repository.Get(created.ID); err != nil {
				t.Errorf("Get(%d) error = %v", created.ID, err)
			}
		}()
	}
	waitGroup.Wait()

	workspaces, err := repository.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(workspaces) != count {
		t.Fatalf("List() length = %d, want %d", len(workspaces), count)
	}
	for index, workspace := range workspaces {
		wantID := uint64(index + 1)
		if workspace.ID != wantID {
			t.Errorf("List()[%d].ID = %d, want %d", index, workspace.ID, wantID)
		}
	}
}

func workspaceIDs(workspaces []Workspace) []uint64 {
	ids := make([]uint64, len(workspaces))
	for index, workspace := range workspaces {
		ids[index] = workspace.ID
	}
	return ids
}
