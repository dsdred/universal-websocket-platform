package configurationversion

import (
	"context"
	"errors"
	"strings"
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
	if first.Listener.Host != "127.0.0.1" || first.Listener.Port != 8080 {
		t.Errorf("default Listener = %#v, want 127.0.0.1:8080", first.Listener)
	}
	if first.Listener.TLS != (TLSSettings{Enabled: false, CertificateRef: "", PrivateKeyRef: "", MinVersion: "1.2"}) {
		t.Errorf("default TLS = %#v", first.Listener.TLS)
	}
	if first.Listener.Timeouts != (TimeoutSettings{HandshakeSeconds: 10, ReadSeconds: 0, WriteSeconds: 10, IdleSeconds: 60}) {
		t.Errorf("default Timeouts = %#v", first.Listener.Timeouts)
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

func TestServicePublishLifecycle(t *testing.T) {
	createdAt := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	publishedAt := createdAt.Add(time.Minute)
	secondCreatedAt := publishedAt.Add(time.Minute)
	secondPublishedAt := secondCreatedAt.Add(time.Minute)
	times := []time.Time{createdAt, publishedAt, secondCreatedAt, secondPublishedAt}
	repository := NewMemoryConfigurationVersionRepository()
	service := NewService(repository, configurationCheckerStub{exists: true}, func() time.Time {
		now := times[0]
		times = times[1:]
		return now
	})

	first, err := service.Create(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("Create(first) error = %v", err)
	}
	first, err = service.Publish(context.Background(), 1, 1, first.ID)
	if err != nil {
		t.Fatalf("Publish(first) error = %v", err)
	}
	if first.State != Published || !first.UpdatedAt.Equal(publishedAt) {
		t.Errorf("published first = %#v", first)
	}

	if _, err := service.Publish(context.Background(), 1, 1, first.ID); !errors.Is(err, ErrVersionNotPublishable) {
		t.Errorf("Publish(already Published) error = %v, want ErrVersionNotPublishable", err)
	}

	second, err := service.Create(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("Create(second) error = %v", err)
	}
	second, err = service.Publish(context.Background(), 1, 1, second.ID)
	if err != nil {
		t.Fatalf("Publish(second) error = %v", err)
	}
	if second.State != Published || !second.UpdatedAt.Equal(secondPublishedAt) {
		t.Errorf("published second = %#v", second)
	}

	archivedFirst, err := repository.Get(first.ID)
	if err != nil {
		t.Fatalf("Get(first) error = %v", err)
	}
	if archivedFirst.State != Archived || !archivedFirst.UpdatedAt.Equal(secondPublishedAt) {
		t.Errorf("archived first = %#v", archivedFirst)
	}
	if _, err := service.Publish(context.Background(), 1, 1, archivedFirst.ID); !errors.Is(err, ErrVersionNotPublishable) {
		t.Errorf("Publish(Archived) error = %v, want ErrVersionNotPublishable", err)
	}

	published, err := repository.GetPublished(1)
	if err != nil {
		t.Fatalf("GetPublished() error = %v", err)
	}
	if published.ID != second.ID {
		t.Errorf("GetPublished() ID = %d, want %d", published.ID, second.ID)
	}
	versions, _ := repository.ListByConfiguration(1)
	publishedCount := 0
	for _, version := range versions {
		if version.State == Published {
			publishedCount++
		}
	}
	if publishedCount != 1 {
		t.Errorf("Published count = %d, want 1", publishedCount)
	}
}

func TestServicePublishNotFound(t *testing.T) {
	service := NewService(NewMemoryConfigurationVersionRepository(), configurationCheckerStub{exists: true}, time.Now)

	if _, err := service.Publish(context.Background(), 1, 1, 42); !errors.Is(err, ErrConfigurationVersionNotFound) {
		t.Errorf("Publish() error = %v, want ErrConfigurationVersionNotFound", err)
	}
}

func TestServiceArchiveAllowedStates(t *testing.T) {
	states := []VersionState{Draft, Validated, Published}
	for _, state := range states {
		t.Run(string(state), func(t *testing.T) {
			createdAt := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
			updatedAt := createdAt.Add(time.Hour)
			repository := NewMemoryConfigurationVersionRepository()
			original, err := repository.Create(ConfigurationVersion{
				ConfigurationID: 1,
				Number:          7,
				State:           state,
				Listener: ListenerSettings{
					Host: "ws.internal.local",
					Port: 9000,
					TLS: TLSSettings{
						Enabled:        true,
						CertificateRef: "certificates/archive",
						PrivateKeyRef:  "secrets/archive-key",
						MinVersion:     "1.3",
					},
					Timeouts: TimeoutSettings{HandshakeSeconds: 15, ReadSeconds: 0, WriteSeconds: 20, IdleSeconds: 120},
				},
				CreatedAt: createdAt,
				UpdatedAt: createdAt,
			})
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			service := NewService(repository, configurationCheckerStub{exists: true}, func() time.Time { return updatedAt })

			archived, err := service.Archive(context.Background(), 1, 1, original.ID)
			if err != nil {
				t.Fatalf("Archive() error = %v", err)
			}
			if archived.State != Archived || !archived.UpdatedAt.Equal(updatedAt) || archived.UpdatedAt.Location() != time.UTC {
				t.Errorf("Archive() = %#v", archived)
			}
			if archived.ID != original.ID || archived.ConfigurationID != original.ConfigurationID || archived.Number != original.Number || !archived.CreatedAt.Equal(original.CreatedAt) {
				t.Errorf("Archive() changed immutable fields: %#v", archived)
			}
			if archived.Listener != original.Listener {
				t.Errorf("Archive() changed Listener from %#v to %#v", original.Listener, archived.Listener)
			}

			if _, err := service.Archive(context.Background(), 1, 1, original.ID); !errors.Is(err, ErrVersionNotArchivable) {
				t.Errorf("Archive(Archived) error = %v, want ErrVersionNotArchivable", err)
			}
			if state == Published {
				if _, err := repository.GetPublished(1); !errors.Is(err, ErrConfigurationVersionNotFound) {
					t.Errorf("GetPublished() after Archive error = %v, want ErrConfigurationVersionNotFound", err)
				}
			}
		})
	}
}

func TestServiceArchiveNotFound(t *testing.T) {
	repository := NewMemoryConfigurationVersionRepository()
	version, err := repository.Create(ConfigurationVersion{ConfigurationID: 1, Number: 1, State: Draft})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	service := NewService(repository, configurationCheckerStub{exists: true}, time.Now)

	if _, err := service.Archive(context.Background(), 1, 2, version.ID); !errors.Is(err, ErrConfigurationVersionNotFound) {
		t.Errorf("Archive(other Configuration) error = %v, want ErrConfigurationVersionNotFound", err)
	}
	if _, err := service.Archive(context.Background(), 1, 1, 42); !errors.Is(err, ErrConfigurationVersionNotFound) {
		t.Errorf("Archive(missing Version) error = %v, want ErrConfigurationVersionNotFound", err)
	}

	missingConfigurationService := NewService(repository, configurationCheckerStub{exists: false}, time.Now)
	if _, err := missingConfigurationService.Archive(context.Background(), 1, 1, version.ID); !errors.Is(err, ErrConfigurationNotFound) {
		t.Errorf("Archive(missing Configuration) error = %v, want ErrConfigurationNotFound", err)
	}
}

func TestServiceUpdateListener(t *testing.T) {
	createdAt := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	times := []time.Time{createdAt, updatedAt}
	service := NewService(NewMemoryConfigurationVersionRepository(), configurationCheckerStub{exists: true}, func() time.Time {
		now := times[0]
		times = times[1:]
		return now
	})

	created, err := service.Create(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	updated, err := service.UpdateListener(context.Background(), 1, 1, created.ID, ListenerSettings{Host: "  0.0.0.0  ", Port: 9000})
	if err != nil {
		t.Fatalf("UpdateListener() error = %v", err)
	}
	if updated.Listener.Host != "0.0.0.0" || updated.Listener.Port != 9000 || updated.Listener.TLS.MinVersion != "1.2" {
		t.Errorf("Listener = %#v", updated.Listener)
	}
	if updated.ID != created.ID || updated.ConfigurationID != created.ConfigurationID || updated.Number != created.Number || updated.State != created.State || !updated.CreatedAt.Equal(created.CreatedAt) {
		t.Errorf("UpdateListener() changed immutable fields: %#v", updated)
	}
	if !updated.UpdatedAt.Equal(updatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt) || updated.UpdatedAt.Location() != time.UTC {
		t.Errorf("UpdatedAt = %s, want %s UTC", updated.UpdatedAt, updatedAt)
	}
}

func TestServiceUpdateListenerValidBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		listener ListenerSettings
	}{
		{name: "IPv4 minimum port", listener: ListenerSettings{Host: "127.0.0.1", Port: 1}},
		{name: "IPv4 wildcard", listener: ListenerSettings{Host: "0.0.0.0", Port: 9000}},
		{name: "IPv6 maximum port", listener: ListenerSettings{Host: "::1", Port: 65535}},
		{name: "localhost", listener: ListenerSettings{Host: "localhost", Port: 8080}},
		{name: "hostname", listener: ListenerSettings{Host: "ws.internal.local", Port: 8080}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(NewMemoryConfigurationVersionRepository(), configurationCheckerStub{exists: true}, time.Now)
			created, err := service.Create(context.Background(), 1, 1)
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			if _, err := service.UpdateListener(context.Background(), 1, 1, created.ID, tt.listener); err != nil {
				t.Errorf("UpdateListener() error = %v", err)
			}
		})
	}
}

func TestServiceUpdateListenerValidation(t *testing.T) {
	tests := []struct {
		name     string
		listener ListenerSettings
		field    string
	}{
		{name: "empty host", listener: ListenerSettings{Host: "", Port: 8080}, field: "host"},
		{name: "whitespace host", listener: ListenerSettings{Host: "  ", Port: 8080}, field: "host"},
		{name: "URL scheme", listener: ListenerSettings{Host: "http://localhost", Port: 8080}, field: "host"},
		{name: "host with port", listener: ListenerSettings{Host: "localhost:8080", Port: 8080}, field: "host"},
		{name: "WebSocket URL", listener: ListenerSettings{Host: "ws://127.0.0.1", Port: 8080}, field: "host"},
		{name: "spaces", listener: ListenerSettings{Host: "name with spaces", Port: 8080}, field: "host"},
		{name: "long host", listener: ListenerSettings{Host: strings.Repeat("a", 256), Port: 8080}, field: "host"},
		{name: "zero port", listener: ListenerSettings{Host: "localhost", Port: 0}, field: "port"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(NewMemoryConfigurationVersionRepository(), configurationCheckerStub{exists: true}, time.Now)
			created, err := service.Create(context.Background(), 1, 1)
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			_, err = service.UpdateListener(context.Background(), 1, 1, created.ID, tt.listener)
			var validationError *ValidationError
			if !errors.As(err, &validationError) || validationError.Field != tt.field {
				t.Errorf("UpdateListener() error = %v, want ValidationError for %s", err, tt.field)
			}
		})
	}
}

func TestServiceUpdateListenerStateAndScope(t *testing.T) {
	for _, state := range []VersionState{Published, Archived, Validated} {
		t.Run(string(state), func(t *testing.T) {
			repository := NewMemoryConfigurationVersionRepository()
			version, err := repository.Create(ConfigurationVersion{ConfigurationID: 1, Number: 1, State: state})
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			service := NewService(repository, configurationCheckerStub{exists: true}, time.Now)
			_, err = service.UpdateListener(context.Background(), 1, 1, version.ID, ListenerSettings{Host: "localhost", Port: 8080})
			if !errors.Is(err, ErrVersionNotEditable) {
				t.Errorf("UpdateListener(%s) error = %v, want ErrVersionNotEditable", state, err)
			}
		})
	}

	repository := NewMemoryConfigurationVersionRepository()
	version, _ := repository.Create(ConfigurationVersion{ConfigurationID: 1, Number: 1, State: Draft})
	service := NewService(repository, configurationCheckerStub{exists: true}, time.Now)
	if _, err := service.UpdateListener(context.Background(), 1, 2, version.ID, ListenerSettings{Host: "localhost", Port: 8080}); !errors.Is(err, ErrConfigurationVersionNotFound) {
		t.Errorf("UpdateListener(other Configuration) error = %v", err)
	}
}

func TestListenerLifecycleRegression(t *testing.T) {
	repository := NewMemoryConfigurationVersionRepository()
	service := NewService(repository, configurationCheckerStub{exists: true}, time.Now)

	first, _ := service.Create(context.Background(), 1, 1)
	first, _ = service.UpdateListener(context.Background(), 1, 1, first.ID, ListenerSettings{Host: "0.0.0.0", Port: 9000})
	first, _ = service.Publish(context.Background(), 1, 1, first.ID)
	if first.Listener.Host != "0.0.0.0" || first.Listener.Port != 9000 {
		t.Errorf("Publish changed first Listener: %#v", first.Listener)
	}

	second, _ := service.Create(context.Background(), 1, 1)
	second, _ = service.UpdateListener(context.Background(), 1, 1, second.ID, ListenerSettings{Host: "localhost", Port: 9001})
	second, _ = service.Publish(context.Background(), 1, 1, second.ID)
	archivedFirst, _ := repository.Get(first.ID)
	if archivedFirst.State != Archived || archivedFirst.Listener.Host != "0.0.0.0" || archivedFirst.Listener.Port != 9000 {
		t.Errorf("auto-archive changed first Version: %#v", archivedFirst)
	}
	if second.Listener.Host != "localhost" || second.Listener.Port != 9001 {
		t.Errorf("Publish changed second Listener: %#v", second.Listener)
	}
}

func TestServiceUpdateTLS(t *testing.T) {
	createdAt := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	repository := NewMemoryConfigurationVersionRepository()
	original, err := repository.Create(ConfigurationVersion{
		ConfigurationID: 1,
		Number:          7,
		State:           Draft,
		Listener: ListenerSettings{
			Host: "0.0.0.0",
			Port: 9443,
			TLS:  TLSSettings{MinVersion: "1.2"},
		},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	service := NewService(repository, configurationCheckerStub{exists: true}, func() time.Time { return updatedAt })

	updated, err := service.UpdateTLS(context.Background(), 1, 1, original.ID, TLSSettings{
		Enabled:        true,
		CertificateRef: "  certificates/main  ",
		PrivateKeyRef:  "  secrets/tls-key  ",
		MinVersion:     " 1.3 ",
	})
	if err != nil {
		t.Fatalf("UpdateTLS() error = %v", err)
	}
	wantTLS := TLSSettings{Enabled: true, CertificateRef: "certificates/main", PrivateKeyRef: "secrets/tls-key", MinVersion: "1.3"}
	if updated.Listener.TLS != wantTLS {
		t.Errorf("TLS = %#v, want %#v", updated.Listener.TLS, wantTLS)
	}
	if updated.Listener.Host != original.Listener.Host || updated.Listener.Port != original.Listener.Port {
		t.Errorf("UpdateTLS() changed Host/Port: %#v", updated.Listener)
	}
	if updated.ID != original.ID || updated.ConfigurationID != original.ConfigurationID || updated.Number != original.Number || updated.State != original.State || !updated.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("UpdateTLS() changed immutable fields: %#v", updated)
	}
	if !updated.UpdatedAt.Equal(updatedAt) || updated.UpdatedAt.Equal(original.UpdatedAt) || updated.UpdatedAt.Location() != time.UTC {
		t.Errorf("UpdatedAt = %s, want %s UTC", updated.UpdatedAt, updatedAt)
	}

	disabled, err := service.UpdateTLS(context.Background(), 1, 1, original.ID, TLSSettings{Enabled: false, MinVersion: "1.2"})
	if err != nil {
		t.Fatalf("UpdateTLS(disable) error = %v", err)
	}
	if disabled.Listener.TLS != (TLSSettings{Enabled: false, MinVersion: "1.2"}) {
		t.Errorf("disabled TLS = %#v", disabled.Listener.TLS)
	}
}

func TestServiceUpdateTLSValidation(t *testing.T) {
	validCertificate := "certificates/main"
	validKey := "secrets/tls-key"
	tests := []struct {
		name  string
		tls   TLSSettings
		field string
	}{
		{name: "missing certificate", tls: TLSSettings{Enabled: true, PrivateKeyRef: validKey, MinVersion: "1.2"}, field: "certificateRef"},
		{name: "missing private key", tls: TLSSettings{Enabled: true, CertificateRef: validCertificate, MinVersion: "1.2"}, field: "privateKeyRef"},
		{name: "unknown minimum version", tls: TLSSettings{MinVersion: "1.1"}, field: "minVersion"},
		{name: "empty minimum version", tls: TLSSettings{}, field: "minVersion"},
		{name: "PEM certificate", tls: TLSSettings{Enabled: true, CertificateRef: "-----BEGIN CERTIFICATE-----", PrivateKeyRef: validKey, MinVersion: "1.2"}, field: "certificateRef"},
		{name: "Windows path", tls: TLSSettings{Enabled: true, CertificateRef: `C:\certs\server.crt`, PrivateKeyRef: validKey, MinVersion: "1.2"}, field: "certificateRef"},
		{name: "Unix path", tls: TLSSettings{Enabled: true, CertificateRef: "/tmp/server.crt", PrivateKeyRef: validKey, MinVersion: "1.2"}, field: "certificateRef"},
		{name: "URL", tls: TLSSettings{Enabled: true, CertificateRef: "https://example.com/cert.pem", PrivateKeyRef: validKey, MinVersion: "1.2"}, field: "certificateRef"},
		{name: "double slash", tls: TLSSettings{Enabled: true, CertificateRef: "certificates//main", PrivateKeyRef: validKey, MinVersion: "1.2"}, field: "certificateRef"},
		{name: "spaces", tls: TLSSettings{Enabled: true, CertificateRef: "certificates/main cert", PrivateKeyRef: validKey, MinVersion: "1.2"}, field: "certificateRef"},
		{name: "long reference", tls: TLSSettings{Enabled: true, CertificateRef: strings.Repeat("a", 256), PrivateKeyRef: validKey, MinVersion: "1.2"}, field: "certificateRef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(NewMemoryConfigurationVersionRepository(), configurationCheckerStub{exists: true}, time.Now)
			created, err := service.Create(context.Background(), 1, 1)
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			_, err = service.UpdateTLS(context.Background(), 1, 1, created.ID, tt.tls)
			var validationError *ValidationError
			if !errors.As(err, &validationError) || validationError.Field != tt.field {
				t.Errorf("UpdateTLS() error = %v, want ValidationError for %s", err, tt.field)
			}
		})
	}
}

func TestServiceUpdateTLSVersions(t *testing.T) {
	for _, minVersion := range []string{"1.2", "1.3"} {
		t.Run(minVersion, func(t *testing.T) {
			service := NewService(NewMemoryConfigurationVersionRepository(), configurationCheckerStub{exists: true}, time.Now)
			created, _ := service.Create(context.Background(), 1, 1)
			if _, err := service.UpdateTLS(context.Background(), 1, 1, created.ID, TLSSettings{MinVersion: minVersion}); err != nil {
				t.Errorf("UpdateTLS(%s) error = %v", minVersion, err)
			}
		})
	}
}

func TestServiceUpdateTLSStateAndScope(t *testing.T) {
	for _, state := range []VersionState{Published, Archived, Validated} {
		t.Run(string(state), func(t *testing.T) {
			repository := NewMemoryConfigurationVersionRepository()
			version, _ := repository.Create(ConfigurationVersion{ConfigurationID: 1, Number: 1, State: state})
			service := NewService(repository, configurationCheckerStub{exists: true}, time.Now)
			_, err := service.UpdateTLS(context.Background(), 1, 1, version.ID, TLSSettings{MinVersion: "1.2"})
			if !errors.Is(err, ErrVersionNotEditable) {
				t.Errorf("UpdateTLS(%s) error = %v, want ErrVersionNotEditable", state, err)
			}
		})
	}

	repository := NewMemoryConfigurationVersionRepository()
	version, _ := repository.Create(ConfigurationVersion{ConfigurationID: 1, Number: 1, State: Draft})
	service := NewService(repository, configurationCheckerStub{exists: true}, time.Now)
	if _, err := service.UpdateTLS(context.Background(), 1, 2, version.ID, TLSSettings{MinVersion: "1.2"}); !errors.Is(err, ErrConfigurationVersionNotFound) {
		t.Errorf("UpdateTLS(other Configuration) error = %v", err)
	}
}

func TestTLSLifecycleRegression(t *testing.T) {
	repository := NewMemoryConfigurationVersionRepository()
	service := NewService(repository, configurationCheckerStub{exists: true}, time.Now)
	firstTLS := TLSSettings{Enabled: true, CertificateRef: "certificates/main", PrivateKeyRef: "secrets/tls-key", MinVersion: "1.3"}
	secondTLS := TLSSettings{Enabled: true, CertificateRef: "certificates/next", PrivateKeyRef: "secrets/next-key", MinVersion: "1.2"}
	firstTimeouts := TimeoutSettings{HandshakeSeconds: 15, ReadSeconds: 0, WriteSeconds: 20, IdleSeconds: 120}
	secondTimeouts := TimeoutSettings{HandshakeSeconds: 20, ReadSeconds: 30, WriteSeconds: 25, IdleSeconds: 180}

	first, _ := service.Create(context.Background(), 1, 1)
	first, _ = service.UpdateListener(context.Background(), 1, 1, first.ID, ListenerSettings{Host: "0.0.0.0", Port: 9443})
	first, _ = service.UpdateTLS(context.Background(), 1, 1, first.ID, firstTLS)
	first, _ = service.UpdateTimeouts(context.Background(), 1, 1, first.ID, firstTimeouts)
	first, _ = service.Publish(context.Background(), 1, 1, first.ID)
	if first.Listener.Host != "0.0.0.0" || first.Listener.Port != 9443 || first.Listener.TLS != firstTLS || first.Listener.Timeouts != firstTimeouts {
		t.Errorf("Publish changed first Listener/TLS: %#v", first.Listener)
	}

	second, _ := service.Create(context.Background(), 1, 1)
	second, _ = service.UpdateTLS(context.Background(), 1, 1, second.ID, secondTLS)
	second, _ = service.UpdateTimeouts(context.Background(), 1, 1, second.ID, secondTimeouts)
	second, _ = service.Publish(context.Background(), 1, 1, second.ID)
	archivedFirst, _ := repository.Get(first.ID)
	if archivedFirst.State != Archived || archivedFirst.Listener.TLS != firstTLS || archivedFirst.Listener.Timeouts != firstTimeouts {
		t.Errorf("auto-archive changed first TLS: %#v", archivedFirst)
	}
	if second.Listener.TLS != secondTLS || second.Listener.Timeouts != secondTimeouts {
		t.Errorf("Publish changed second TLS: %#v", second.Listener.TLS)
	}
}

func TestServiceUpdateTimeouts(t *testing.T) {
	createdAt := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	originalTLS := TLSSettings{Enabled: true, CertificateRef: "certificates/main", PrivateKeyRef: "secrets/key", MinVersion: "1.3"}
	repository := NewMemoryConfigurationVersionRepository()
	original, err := repository.Create(ConfigurationVersion{
		ConfigurationID: 1,
		Number:          7,
		State:           Draft,
		Listener: ListenerSettings{
			Host:     "0.0.0.0",
			Port:     9443,
			TLS:      originalTLS,
			Timeouts: TimeoutSettings{HandshakeSeconds: 10, ReadSeconds: 0, WriteSeconds: 10, IdleSeconds: 60},
		},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	service := NewService(repository, configurationCheckerStub{exists: true}, func() time.Time { return updatedAt })
	want := TimeoutSettings{HandshakeSeconds: 15, ReadSeconds: 0, WriteSeconds: 20, IdleSeconds: 120}

	updated, err := service.UpdateTimeouts(context.Background(), 1, 1, original.ID, want)
	if err != nil {
		t.Fatalf("UpdateTimeouts() error = %v", err)
	}
	if updated.Listener.Timeouts != want {
		t.Errorf("Timeouts = %#v, want %#v", updated.Listener.Timeouts, want)
	}
	if updated.Listener.Host != original.Listener.Host || updated.Listener.Port != original.Listener.Port || updated.Listener.TLS != originalTLS {
		t.Errorf("UpdateTimeouts() changed Listener/TLS: %#v", updated.Listener)
	}
	if updated.ID != original.ID || updated.ConfigurationID != original.ConfigurationID || updated.Number != original.Number || updated.State != original.State || !updated.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("UpdateTimeouts() changed immutable fields: %#v", updated)
	}
	if !updated.UpdatedAt.Equal(updatedAt) || updated.UpdatedAt.Equal(original.UpdatedAt) || updated.UpdatedAt.Location() != time.UTC {
		t.Errorf("UpdatedAt = %s, want %s UTC", updated.UpdatedAt, updatedAt)
	}
}

func TestServiceUpdateTimeoutsBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		timeouts TimeoutSettings
		valid    bool
		field    string
	}{
		{name: "handshake zero", timeouts: TimeoutSettings{HandshakeSeconds: 0, WriteSeconds: 10}, field: "handshakeSeconds"},
		{name: "handshake one", timeouts: TimeoutSettings{HandshakeSeconds: 1, WriteSeconds: 10}, valid: true},
		{name: "handshake 300", timeouts: TimeoutSettings{HandshakeSeconds: 300, WriteSeconds: 10}, valid: true},
		{name: "handshake above", timeouts: TimeoutSettings{HandshakeSeconds: 301, WriteSeconds: 10}, field: "handshakeSeconds"},
		{name: "read zero", timeouts: TimeoutSettings{HandshakeSeconds: 10, ReadSeconds: 0, WriteSeconds: 10}, valid: true},
		{name: "read 86400", timeouts: TimeoutSettings{HandshakeSeconds: 10, ReadSeconds: 86400, WriteSeconds: 10}, valid: true},
		{name: "read above", timeouts: TimeoutSettings{HandshakeSeconds: 10, ReadSeconds: 86401, WriteSeconds: 10}, field: "readSeconds"},
		{name: "write zero", timeouts: TimeoutSettings{HandshakeSeconds: 10, WriteSeconds: 0}, field: "writeSeconds"},
		{name: "write one", timeouts: TimeoutSettings{HandshakeSeconds: 10, WriteSeconds: 1}, valid: true},
		{name: "write 300", timeouts: TimeoutSettings{HandshakeSeconds: 10, WriteSeconds: 300}, valid: true},
		{name: "write above", timeouts: TimeoutSettings{HandshakeSeconds: 10, WriteSeconds: 301}, field: "writeSeconds"},
		{name: "idle zero", timeouts: TimeoutSettings{HandshakeSeconds: 10, WriteSeconds: 10, IdleSeconds: 0}, valid: true},
		{name: "idle 86400", timeouts: TimeoutSettings{HandshakeSeconds: 10, WriteSeconds: 10, IdleSeconds: 86400}, valid: true},
		{name: "idle above", timeouts: TimeoutSettings{HandshakeSeconds: 10, WriteSeconds: 10, IdleSeconds: 86401}, field: "idleSeconds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(NewMemoryConfigurationVersionRepository(), configurationCheckerStub{exists: true}, time.Now)
			created, err := service.Create(context.Background(), 1, 1)
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			_, err = service.UpdateTimeouts(context.Background(), 1, 1, created.ID, tt.timeouts)
			if tt.valid {
				if err != nil {
					t.Errorf("UpdateTimeouts() error = %v", err)
				}
				return
			}
			var validationError *ValidationError
			if !errors.As(err, &validationError) || validationError.Field != tt.field {
				t.Errorf("UpdateTimeouts() error = %v, want ValidationError for %s", err, tt.field)
			}
		})
	}
}

func TestServiceUpdateTimeoutsStateAndScope(t *testing.T) {
	valid := TimeoutSettings{HandshakeSeconds: 10, WriteSeconds: 10, IdleSeconds: 60}
	for _, state := range []VersionState{Published, Archived, Validated} {
		t.Run(string(state), func(t *testing.T) {
			repository := NewMemoryConfigurationVersionRepository()
			version, _ := repository.Create(ConfigurationVersion{ConfigurationID: 1, Number: 1, State: state})
			service := NewService(repository, configurationCheckerStub{exists: true}, time.Now)
			if _, err := service.UpdateTimeouts(context.Background(), 1, 1, version.ID, valid); !errors.Is(err, ErrVersionNotEditable) {
				t.Errorf("UpdateTimeouts(%s) error = %v", state, err)
			}
		})
	}

	repository := NewMemoryConfigurationVersionRepository()
	version, _ := repository.Create(ConfigurationVersion{ConfigurationID: 1, Number: 1, State: Draft})
	service := NewService(repository, configurationCheckerStub{exists: true}, time.Now)
	if _, err := service.UpdateTimeouts(context.Background(), 1, 2, version.ID, valid); !errors.Is(err, ErrConfigurationVersionNotFound) {
		t.Errorf("UpdateTimeouts(other Configuration) error = %v", err)
	}
}
