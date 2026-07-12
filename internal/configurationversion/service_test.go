package configurationversion

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type configurationCheckerStub struct {
	exists bool
	err    error
}

func (s configurationCheckerStub) Exists(context.Context, uint64, uint64) (bool, error) {
	return s.exists, s.err
}

func TestServiceCreatesDraftVersionsWithAutomaticNumbers(t *testing.T) {
	firstTime := time.Date(2026, 7, 12, 17, 0, 0, 0, time.FixedZone("test", 5*60*60))
	nextTime := firstTime
	service := NewService(
		NewMemoryConfigurationVersionRepository(),
		configurationCheckerStub{exists: true},
		func() time.Time {
			current := nextTime
			nextTime = nextTime.Add(time.Minute)
			return current
		},
	)

	first, err := service.Create(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("Create(first) error = %v", err)
	}
	second, err := service.Create(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("Create(second) error = %v", err)
	}
	if first.Number != 1 || second.Number != 2 {
		t.Errorf("Numbers = [%d %d], want [1 2]", first.Number, second.Number)
	}
	if first.State != Draft || second.State != Draft {
		t.Errorf("States = [%q %q], want Draft", first.State, second.State)
	}
	if first.ConfigurationID != 1 || first.CreatedAt.Location() != time.UTC || !first.CreatedAt.Equal(firstTime) {
		t.Errorf("first Version = %#v", first)
	}
	if !first.UpdatedAt.Equal(first.CreatedAt) {
		t.Errorf("UpdatedAt = %s, want %s", first.UpdatedAt, first.CreatedAt)
	}

	versions, err := service.List(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(versions) != 2 || versions[0].Number != 1 || versions[1].Number != 2 {
		t.Errorf("List() numbers = %v, want [1 2]", versionNumbers(versions))
	}
}

func TestServiceNumbersVersionsPerConfiguration(t *testing.T) {
	service := NewService(NewMemoryConfigurationVersionRepository(), configurationCheckerStub{exists: true}, time.Now)

	first, err := service.Create(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("Create(configuration 1) error = %v", err)
	}
	other, err := service.Create(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("Create(configuration 2) error = %v", err)
	}
	if first.Number != 1 || other.Number != 1 {
		t.Errorf("Numbers = [%d %d], want [1 1]", first.Number, other.Number)
	}
}

func TestServiceConfigurationNotFound(t *testing.T) {
	service := NewService(NewMemoryConfigurationVersionRepository(), configurationCheckerStub{exists: false}, time.Now)

	if _, err := service.Create(context.Background(), 1, 42); !errors.Is(err, ErrConfigurationNotFound) {
		t.Errorf("Create() error = %v, want ErrConfigurationNotFound", err)
	}
	if _, err := service.List(context.Background(), 1, 42); !errors.Is(err, ErrConfigurationNotFound) {
		t.Errorf("List() error = %v, want ErrConfigurationNotFound", err)
	}
}

func TestServiceAssignsUniqueNumbersConcurrently(t *testing.T) {
	service := NewService(NewMemoryConfigurationVersionRepository(), configurationCheckerStub{exists: true}, time.Now)
	const count = 100

	var waitGroup sync.WaitGroup
	waitGroup.Add(count)
	for range count {
		go func() {
			defer waitGroup.Done()
			if _, err := service.Create(context.Background(), 1, 1); err != nil {
				t.Errorf("Create() error = %v", err)
			}
		}()
	}
	waitGroup.Wait()

	versions, err := service.List(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(versions) != count {
		t.Fatalf("List() length = %d, want %d", len(versions), count)
	}
	for index, version := range versions {
		if want := uint32(index + 1); version.Number != want {
			t.Errorf("versions[%d].Number = %d, want %d", index, version.Number, want)
		}
	}
}
