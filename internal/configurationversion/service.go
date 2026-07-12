package configurationversion

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	defaultListenerHost   = "127.0.0.1"
	defaultListenerPort   = 8080
	maxListenerHostLength = 255
)

// ValidationError describes invalid Configuration Version settings.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Service applies Configuration Version business rules.
type Service struct {
	repository           ConfigurationVersionRepository
	configurationChecker ConfigurationExistenceChecker
	now                  func() time.Time
	lifecycleMu          sync.Mutex
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

	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

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
		Listener: ListenerSettings{
			Host: defaultListenerHost,
			Port: defaultListenerPort,
		},
		CreatedAt: now,
		UpdatedAt: now,
	})
}

// UpdateListener validates and updates Listener settings for a Draft Version.
func (s *Service) UpdateListener(
	ctx context.Context,
	workspaceID, configurationID, versionID uint64,
	listener ListenerSettings,
) (ConfigurationVersion, error) {
	if err := s.requireConfiguration(ctx, workspaceID, configurationID); err != nil {
		return ConfigurationVersion{}, err
	}

	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	version, err := s.repository.Get(versionID)
	if err != nil || version.ConfigurationID != configurationID {
		if err == nil || errors.Is(err, ErrConfigurationVersionNotFound) {
			return ConfigurationVersion{}, ErrConfigurationVersionNotFound
		}
		return ConfigurationVersion{}, err
	}
	if version.State != Draft {
		return ConfigurationVersion{}, ErrVersionNotEditable
	}

	normalized, err := validateListener(listener)
	if err != nil {
		return ConfigurationVersion{}, err
	}

	version.Listener = normalized
	version.UpdatedAt = s.now().UTC()
	return s.repository.Update(version)
}

// Publish transitions a Draft Version to Published and archives the previous Published Version.
func (s *Service) Publish(ctx context.Context, workspaceID, configurationID, versionID uint64) (ConfigurationVersion, error) {
	if err := s.requireConfiguration(ctx, workspaceID, configurationID); err != nil {
		return ConfigurationVersion{}, err
	}

	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	target, err := s.repository.Get(versionID)
	if err != nil || target.ConfigurationID != configurationID {
		if err == nil || errors.Is(err, ErrConfigurationVersionNotFound) {
			return ConfigurationVersion{}, ErrConfigurationVersionNotFound
		}
		return ConfigurationVersion{}, err
	}
	if target.State != Draft {
		return ConfigurationVersion{}, ErrVersionNotPublishable
	}

	now := s.now().UTC()
	updates := make([]ConfigurationVersion, 0, 2)
	current, err := s.repository.GetPublished(configurationID)
	switch {
	case err == nil:
		current.State = Archived
		current.UpdatedAt = now
		updates = append(updates, current)
	case !errors.Is(err, ErrConfigurationVersionNotFound):
		return ConfigurationVersion{}, err
	}

	target.State = Published
	target.UpdatedAt = now
	updates = append(updates, target)
	if err := s.repository.UpdateBatch(updates); err != nil {
		return ConfigurationVersion{}, err
	}

	return target, nil
}

// Archive transitions a non-Archived Version to Archived.
func (s *Service) Archive(ctx context.Context, workspaceID, configurationID, versionID uint64) (ConfigurationVersion, error) {
	if err := s.requireConfiguration(ctx, workspaceID, configurationID); err != nil {
		return ConfigurationVersion{}, err
	}

	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	version, err := s.repository.Get(versionID)
	if err != nil || version.ConfigurationID != configurationID {
		if err == nil || errors.Is(err, ErrConfigurationVersionNotFound) {
			return ConfigurationVersion{}, ErrConfigurationVersionNotFound
		}
		return ConfigurationVersion{}, err
	}

	switch version.State {
	case Draft, Validated, Published:
	case Archived:
		return ConfigurationVersion{}, ErrVersionNotArchivable
	default:
		return ConfigurationVersion{}, ErrVersionNotArchivable
	}

	version.State = Archived
	version.UpdatedAt = s.now().UTC()
	return s.repository.Update(version)
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

func validateListener(listener ListenerSettings) (ListenerSettings, error) {
	host := strings.TrimSpace(listener.Host)
	if host == "" {
		return ListenerSettings{}, &ValidationError{Field: "host", Message: "must not be empty"}
	}
	if utf8.RuneCountInString(host) > maxListenerHostLength {
		return ListenerSettings{}, &ValidationError{Field: "host", Message: "must not exceed 255 characters"}
	}
	if net.ParseIP(host) == nil && !validHostname(host) {
		return ListenerSettings{}, &ValidationError{Field: "host", Message: "must be a valid IP address or hostname"}
	}
	if listener.Port == 0 {
		return ListenerSettings{}, &ValidationError{Field: "port", Message: "must be between 1 and 65535"}
	}

	return ListenerSettings{Host: host, Port: listener.Port}, nil
}

func validHostname(host string) bool {
	if len(host) > maxListenerHostLength {
		return false
	}
	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return false
	}

	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, character := range label {
			if !((character >= 'a' && character <= 'z') ||
				(character >= 'A' && character <= 'Z') ||
				(character >= '0' && character <= '9') || character == '-') {
				return false
			}
		}
	}
	return true
}
