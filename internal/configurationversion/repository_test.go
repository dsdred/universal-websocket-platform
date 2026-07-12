package configurationversion

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestMemoryRepositoryCRUD(t *testing.T) {
	repository := NewMemoryConfigurationVersionRepository()
	createdAt := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)

	second, err := repository.Create(ConfigurationVersion{ConfigurationID: 1, Number: 2, State: Draft, CreatedAt: createdAt, UpdatedAt: createdAt})
	if err != nil {
		t.Fatalf("Create(second) error = %v", err)
	}
	first, err := repository.Create(ConfigurationVersion{ConfigurationID: 1, Number: 1, State: Draft, CreatedAt: createdAt, UpdatedAt: createdAt})
	if err != nil {
		t.Fatalf("Create(first) error = %v", err)
	}
	other, err := repository.Create(ConfigurationVersion{ConfigurationID: 2, Number: 1, State: Draft})
	if err != nil {
		t.Fatalf("Create(other) error = %v", err)
	}
	if second.ID != 1 || first.ID != 2 || other.ID != 3 {
		t.Errorf("IDs = [%d %d %d], want [1 2 3]", second.ID, first.ID, other.ID)
	}

	got, err := repository.Get(first.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !reflect.DeepEqual(got, first) {
		t.Errorf("Get() = %#v, want %#v", got, first)
	}

	listed, err := repository.ListByConfiguration(1)
	if err != nil {
		t.Fatalf("ListByConfiguration() error = %v", err)
	}
	if len(listed) != 2 || listed[0].Number != 1 || listed[1].Number != 2 {
		t.Errorf("ListByConfiguration() numbers = %v, want [1 2]", versionNumbers(listed))
	}
	for _, version := range listed {
		if version.ConfigurationID != 1 {
			t.Errorf("ListByConfiguration() included ConfigurationID %d", version.ConfigurationID)
		}
	}

	first.State = Published
	updated, err := repository.Update(first)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.State != Published || updated.ID != first.ID {
		t.Errorf("Update() = %#v", updated)
	}
	published, err := repository.GetPublished(1)
	if err != nil {
		t.Fatalf("GetPublished() error = %v", err)
	}
	if published.ID != first.ID {
		t.Errorf("GetPublished() ID = %d, want %d", published.ID, first.ID)
	}

	if err := repository.Delete(first.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repository.Get(first.ID); !errors.Is(err, ErrConfigurationVersionNotFound) {
		t.Errorf("Get(deleted) error = %v, want ErrConfigurationVersionNotFound", err)
	}
}

func TestMemoryRepositoryNotFound(t *testing.T) {
	repository := NewMemoryConfigurationVersionRepository()

	if _, err := repository.Get(42); !errors.Is(err, ErrConfigurationVersionNotFound) {
		t.Errorf("Get() error = %v", err)
	}
	if _, err := repository.Update(ConfigurationVersion{ID: 42}); !errors.Is(err, ErrConfigurationVersionNotFound) {
		t.Errorf("Update() error = %v", err)
	}
	if err := repository.Delete(42); !errors.Is(err, ErrConfigurationVersionNotFound) {
		t.Errorf("Delete() error = %v", err)
	}
	if _, err := repository.GetPublished(42); !errors.Is(err, ErrConfigurationVersionNotFound) {
		t.Errorf("GetPublished() error = %v", err)
	}
}

func TestMemoryRepositoryUpdateBatchIsAtomic(t *testing.T) {
	repository := NewMemoryConfigurationVersionRepository()
	first, err := repository.Create(ConfigurationVersion{ConfigurationID: 1, Number: 1, State: Draft})
	if err != nil {
		t.Fatalf("Create(first) error = %v", err)
	}
	second, err := repository.Create(ConfigurationVersion{ConfigurationID: 1, Number: 2, State: Draft})
	if err != nil {
		t.Fatalf("Create(second) error = %v", err)
	}

	first.State = Archived
	second.State = Published
	if err := repository.UpdateBatch([]ConfigurationVersion{first, second}); err != nil {
		t.Fatalf("UpdateBatch() error = %v", err)
	}
	updatedFirst, _ := repository.Get(first.ID)
	updatedSecond, _ := repository.Get(second.ID)
	if updatedFirst.State != Archived || updatedSecond.State != Published {
		t.Errorf("states after UpdateBatch = [%s %s]", updatedFirst.State, updatedSecond.State)
	}

	updatedFirst.State = Published
	missing := ConfigurationVersion{ID: 999, State: Archived}
	if err := repository.UpdateBatch([]ConfigurationVersion{updatedFirst, missing}); !errors.Is(err, ErrConfigurationVersionNotFound) {
		t.Fatalf("UpdateBatch(with missing) error = %v, want ErrConfigurationVersionNotFound", err)
	}
	unchangedFirst, _ := repository.Get(first.ID)
	if unchangedFirst.State != Archived {
		t.Errorf("failed UpdateBatch changed first State to %s, want Archived", unchangedFirst.State)
	}
}

func TestMemoryRepositoryConcurrentAccess(t *testing.T) {
	repository := NewMemoryConfigurationVersionRepository()
	const count = 100

	var waitGroup sync.WaitGroup
	waitGroup.Add(count)
	for index := range count {
		go func(number int) {
			defer waitGroup.Done()
			created, err := repository.Create(ConfigurationVersion{ConfigurationID: 1, Number: uint32(number + 1)})
			if err != nil {
				t.Errorf("Create() error = %v", err)
				return
			}
			if _, err := repository.Get(created.ID); err != nil {
				t.Errorf("Get(%d) error = %v", created.ID, err)
			}
		}(index)
	}
	waitGroup.Wait()

	versions, err := repository.ListByConfiguration(1)
	if err != nil {
		t.Fatalf("ListByConfiguration() error = %v", err)
	}
	if len(versions) != count {
		t.Fatalf("ListByConfiguration() length = %d, want %d", len(versions), count)
	}
}

func versionNumbers(versions []ConfigurationVersion) []uint32 {
	numbers := make([]uint32, len(versions))
	for index, version := range versions {
		numbers[index] = version.Number
	}
	return numbers
}
