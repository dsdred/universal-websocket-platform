package configuration

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestMemoryConfigurationRepositoryCRUD(t *testing.T) {
	repository := NewMemoryConfigurationRepository()
	createdAt := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)

	first, err := repository.Create(Configuration{WorkspaceID: 1, Name: "First", CreatedAt: createdAt, UpdatedAt: createdAt})
	if err != nil {
		t.Fatalf("Create(first) error = %v", err)
	}
	other, err := repository.Create(Configuration{WorkspaceID: 2, Name: "Other", CreatedAt: createdAt, UpdatedAt: createdAt})
	if err != nil {
		t.Fatalf("Create(other) error = %v", err)
	}
	second, err := repository.Create(Configuration{WorkspaceID: 1, Name: "Second", CreatedAt: createdAt, UpdatedAt: createdAt})
	if err != nil {
		t.Fatalf("Create(second) error = %v", err)
	}
	if first.ID != 1 || other.ID != 2 || second.ID != 3 {
		t.Errorf("assigned IDs = [%d %d %d], want [1 2 3]", first.ID, other.ID, second.ID)
	}

	got, err := repository.Get(first.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != first || !got.CreatedAt.Equal(createdAt) {
		t.Errorf("Get() = %#v, want %#v", got, first)
	}

	listed, err := repository.ListByWorkspace(1)
	if err != nil {
		t.Fatalf("ListByWorkspace() error = %v", err)
	}
	if len(listed) != 2 || listed[0].ID != first.ID || listed[1].ID != second.ID {
		t.Errorf("ListByWorkspace() IDs = %v, want [1 3]", configurationIDs(listed))
	}
	for _, configuration := range listed {
		if configuration.WorkspaceID != 1 {
			t.Errorf("ListByWorkspace() included WorkspaceID %d", configuration.WorkspaceID)
		}
	}

	first.Name = "Updated"
	updated, err := repository.Update(first)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name != "Updated" || updated.ID != first.ID || !updated.CreatedAt.Equal(createdAt) {
		t.Errorf("Update() = %#v", updated)
	}

	exists, err := repository.ExistsByWorkspace(context.Background(), 1)
	if err != nil || !exists {
		t.Errorf("ExistsByWorkspace(1) = %t, %v; want true, nil", exists, err)
	}
	exists, err = repository.ExistsByWorkspace(context.Background(), 99)
	if err != nil || exists {
		t.Errorf("ExistsByWorkspace(99) = %t, %v; want false, nil", exists, err)
	}

	if err := repository.Delete(first.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repository.Get(first.ID); !errors.Is(err, ErrConfigurationNotFound) {
		t.Errorf("Get(deleted) error = %v, want ErrConfigurationNotFound", err)
	}
}

func TestMemoryConfigurationRepositoryNotFound(t *testing.T) {
	repository := NewMemoryConfigurationRepository()

	if _, err := repository.Get(42); !errors.Is(err, ErrConfigurationNotFound) {
		t.Errorf("Get() error = %v, want ErrConfigurationNotFound", err)
	}
	if _, err := repository.Update(Configuration{ID: 42}); !errors.Is(err, ErrConfigurationNotFound) {
		t.Errorf("Update() error = %v, want ErrConfigurationNotFound", err)
	}
	if err := repository.Delete(42); !errors.Is(err, ErrConfigurationNotFound) {
		t.Errorf("Delete() error = %v, want ErrConfigurationNotFound", err)
	}
}

func TestMemoryConfigurationRepositoryConcurrentAccess(t *testing.T) {
	repository := NewMemoryConfigurationRepository()
	const count = 100

	var waitGroup sync.WaitGroup
	waitGroup.Add(count)
	for range count {
		go func() {
			defer waitGroup.Done()
			created, err := repository.Create(Configuration{WorkspaceID: 1, Name: "Concurrent"})
			if err != nil {
				t.Errorf("Create() error = %v", err)
				return
			}
			if _, err := repository.Get(created.ID); err != nil {
				t.Errorf("Get(%d) error = %v", created.ID, err)
			}
			if _, err := repository.ExistsByWorkspace(context.Background(), 1); err != nil {
				t.Errorf("ExistsByWorkspace() error = %v", err)
			}
		}()
	}
	waitGroup.Wait()

	configurations, err := repository.ListByWorkspace(1)
	if err != nil {
		t.Fatalf("ListByWorkspace() error = %v", err)
	}
	if len(configurations) != count {
		t.Fatalf("ListByWorkspace() length = %d, want %d", len(configurations), count)
	}
	for index, configuration := range configurations {
		if wantID := uint64(index + 1); configuration.ID != wantID {
			t.Errorf("configurations[%d].ID = %d, want %d", index, configuration.ID, wantID)
		}
	}
}

func configurationIDs(configurations []Configuration) []uint64 {
	ids := make([]uint64, len(configurations))
	for index, configuration := range configurations {
		ids[index] = configuration.ID
	}
	return ids
}
