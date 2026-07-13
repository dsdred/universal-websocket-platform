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

	"golang.org/x/net/http/httpguts"
)

const (
	defaultListenerHost                 = "127.0.0.1"
	defaultListenerPort                 = 8080
	defaultTLSMinVersion                = "1.2"
	defaultHandshakeSeconds             = 10
	defaultReadSeconds                  = 0
	defaultWriteSeconds                 = 10
	defaultIdleSeconds                  = 60
	maxListenerHostLength               = 255
	maxTLSReferenceLength               = 255
	maxAuthenticationProviderNameLength = 255
	maxAPIKeyHeaderLength               = 255
	maxSecretReferenceLength            = 255
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
			TLS: TLSSettings{
				Enabled:    false,
				MinVersion: defaultTLSMinVersion,
			},
			Timeouts: TimeoutSettings{
				HandshakeSeconds: defaultHandshakeSeconds,
				ReadSeconds:      defaultReadSeconds,
				WriteSeconds:     defaultWriteSeconds,
				IdleSeconds:      defaultIdleSeconds,
			},
		},
		Authentication: AuthenticationSettings{
			Enabled:   false,
			Providers: make([]AuthenticationProvider, 0),
		},
		CreatedAt: now,
		UpdatedAt: now,
	})
}

// UpdateAuthentication validates and replaces Authentication settings for a Draft Version.
func (s *Service) UpdateAuthentication(
	ctx context.Context,
	workspaceID, configurationID, versionID uint64,
	authentication AuthenticationSettings,
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

	normalized, err := validateAuthentication(authentication)
	if err != nil {
		return ConfigurationVersion{}, err
	}

	version.Authentication = normalized
	version.UpdatedAt = s.now().UTC()
	return s.repository.Update(version)
}

// UpdateTimeouts validates and updates timeout settings for a Draft Version.
func (s *Service) UpdateTimeouts(
	ctx context.Context,
	workspaceID, configurationID, versionID uint64,
	timeouts TimeoutSettings,
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
	if err := validateTimeouts(timeouts); err != nil {
		return ConfigurationVersion{}, err
	}

	version.Listener.Timeouts = timeouts
	version.UpdatedAt = s.now().UTC()
	return s.repository.Update(version)
}

// UpdateTLS validates and updates TLS settings for a Draft Version.
func (s *Service) UpdateTLS(
	ctx context.Context,
	workspaceID, configurationID, versionID uint64,
	tls TLSSettings,
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

	normalized, err := validateTLS(tls)
	if err != nil {
		return ConfigurationVersion{}, err
	}

	version.Listener.TLS = normalized
	version.UpdatedAt = s.now().UTC()
	return s.repository.Update(version)
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

	version.Listener.Host = normalized.Host
	version.Listener.Port = normalized.Port
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

func validateTLS(tls TLSSettings) (TLSSettings, error) {
	tls.CertificateRef = strings.TrimSpace(tls.CertificateRef)
	tls.PrivateKeyRef = strings.TrimSpace(tls.PrivateKeyRef)
	tls.MinVersion = strings.TrimSpace(tls.MinVersion)

	if tls.MinVersion != "1.2" && tls.MinVersion != "1.3" {
		return TLSSettings{}, &ValidationError{Field: "minVersion", Message: "must be one of 1.2 or 1.3"}
	}
	if tls.Enabled && tls.CertificateRef == "" {
		return TLSSettings{}, &ValidationError{Field: "certificateRef", Message: "must not be empty when TLS is enabled"}
	}
	if tls.Enabled && tls.PrivateKeyRef == "" {
		return TLSSettings{}, &ValidationError{Field: "privateKeyRef", Message: "must not be empty when TLS is enabled"}
	}
	if tls.CertificateRef != "" && !validTLSReference(tls.CertificateRef) {
		return TLSSettings{}, &ValidationError{Field: "certificateRef", Message: "must be a valid certificate reference"}
	}
	if tls.PrivateKeyRef != "" && !validTLSReference(tls.PrivateKeyRef) {
		return TLSSettings{}, &ValidationError{Field: "privateKeyRef", Message: "must be a valid private key reference"}
	}

	return tls, nil
}

func validTLSReference(reference string) bool {
	if utf8.RuneCountInString(reference) > maxTLSReferenceLength ||
		strings.HasPrefix(reference, "/") || strings.HasSuffix(reference, "/") ||
		strings.Contains(reference, "//") || strings.Contains(reference, "://") ||
		strings.Contains(reference, "-----BEGIN") {
		return false
	}

	for _, character := range reference {
		if !((character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			character == '/' || character == '-' || character == '_' || character == '.') {
			return false
		}
	}
	return true
}

func validateTimeouts(timeouts TimeoutSettings) error {
	if timeouts.HandshakeSeconds < 1 || timeouts.HandshakeSeconds > 300 {
		return &ValidationError{Field: "handshakeSeconds", Message: "must be between 1 and 300"}
	}
	if timeouts.ReadSeconds > 86400 {
		return &ValidationError{Field: "readSeconds", Message: "must be between 0 and 86400"}
	}
	if timeouts.WriteSeconds < 1 || timeouts.WriteSeconds > 300 {
		return &ValidationError{Field: "writeSeconds", Message: "must be between 1 and 300"}
	}
	if timeouts.IdleSeconds > 86400 {
		return &ValidationError{Field: "idleSeconds", Message: "must be between 0 and 86400"}
	}
	return nil
}

func validateAuthentication(authentication AuthenticationSettings) (AuthenticationSettings, error) {
	if authentication.Enabled && len(authentication.Providers) == 0 {
		return AuthenticationSettings{}, &ValidationError{Field: "providers", Message: "must contain at least one provider when Authentication is enabled"}
	}

	providers := make([]AuthenticationProvider, len(authentication.Providers))
	names := make(map[string]struct{}, len(authentication.Providers))
	priorities := make(map[uint32]struct{}, len(authentication.Providers))
	for index, provider := range authentication.Providers {
		provider.Name = strings.TrimSpace(provider.Name)
		if provider.Name == "" {
			return AuthenticationSettings{}, &ValidationError{Field: "providers.name", Message: "must not be empty"}
		}
		if utf8.RuneCountInString(provider.Name) > maxAuthenticationProviderNameLength {
			return AuthenticationSettings{}, &ValidationError{Field: "providers.name", Message: "must not exceed 255 characters"}
		}
		if _, exists := names[provider.Name]; exists {
			return AuthenticationSettings{}, &ValidationError{Field: "providers.name", Message: "must be unique"}
		}
		names[provider.Name] = struct{}{}

		switch provider.Type {
		case AuthenticationProviderAPIKey:
			if provider.JWT != nil {
				return AuthenticationSettings{}, &ValidationError{Field: "providers.jwt", Message: "must be omitted for non-jwt Provider"}
			}
			if provider.APIKey == nil {
				return AuthenticationSettings{}, &ValidationError{Field: "providers.apiKey", Message: "must be provided for api-key Provider"}
			}
			normalized, err := validateAPIKey(*provider.APIKey)
			if err != nil {
				return AuthenticationSettings{}, err
			}
			provider.APIKey = &normalized
		case AuthenticationProviderJWT:
			if provider.APIKey != nil {
				return AuthenticationSettings{}, &ValidationError{Field: "providers.apiKey", Message: "must be omitted for non-api-key Provider"}
			}
			if provider.JWT == nil {
				return AuthenticationSettings{}, &ValidationError{Field: "providers.jwt", Message: "must be provided for jwt Provider"}
			}
			normalized, err := validateJWT(*provider.JWT)
			if err != nil {
				return AuthenticationSettings{}, err
			}
			provider.JWT = &normalized
		case AuthenticationProviderBasic:
			if provider.APIKey != nil {
				return AuthenticationSettings{}, &ValidationError{Field: "providers.apiKey", Message: "must be omitted for non-api-key Provider"}
			}
			if provider.JWT != nil {
				return AuthenticationSettings{}, &ValidationError{Field: "providers.jwt", Message: "must be omitted for non-jwt Provider"}
			}
		default:
			return AuthenticationSettings{}, &ValidationError{Field: "providers.type", Message: "must be one of jwt, api-key, or basic"}
		}
		if _, exists := priorities[provider.Priority]; exists {
			return AuthenticationSettings{}, &ValidationError{Field: "providers.priority", Message: "must be unique"}
		}
		priorities[provider.Priority] = struct{}{}
		providers[index] = provider
	}

	return AuthenticationSettings{Enabled: authentication.Enabled, Providers: providers}, nil
}

func validateAPIKey(apiKey APIKeySettings) (APIKeySettings, error) {
	apiKey.Header = strings.TrimSpace(apiKey.Header)
	if apiKey.Header == "" {
		return APIKeySettings{}, &ValidationError{Field: "providers.apiKey.header", Message: "must not be empty"}
	}
	if utf8.RuneCountInString(apiKey.Header) > maxAPIKeyHeaderLength {
		return APIKeySettings{}, &ValidationError{Field: "providers.apiKey.header", Message: "must not exceed 255 characters"}
	}
	if !httpguts.ValidHeaderFieldName(apiKey.Header) {
		return APIKeySettings{}, &ValidationError{Field: "providers.apiKey.header", Message: "must be a valid HTTP header field name"}
	}

	apiKey.SecretRef = strings.TrimSpace(apiKey.SecretRef)
	if apiKey.SecretRef == "" {
		return APIKeySettings{}, &ValidationError{Field: "providers.apiKey.secretRef", Message: "must not be empty"}
	}
	if utf8.RuneCountInString(apiKey.SecretRef) > maxSecretReferenceLength || !validTLSReference(apiKey.SecretRef) {
		return APIKeySettings{}, &ValidationError{Field: "providers.apiKey.secretRef", Message: "must be a valid Secret Reference"}
	}

	return apiKey, nil
}

func validateJWT(jwt JWTSettings) (JWTSettings, error) {
	if len(jwt.SigningKeys) == 0 {
		return JWTSettings{}, &ValidationError{Field: "providers.jwt.signingKeys", Message: "must contain at least one signing key"}
	}
	signingKeys := make([]JWTSigningKey, len(jwt.SigningKeys))
	keyNames := make(map[string]struct{}, len(jwt.SigningKeys))
	for index, key := range jwt.SigningKeys {
		key.Name = strings.TrimSpace(key.Name)
		if key.Name == "" {
			return JWTSettings{}, &ValidationError{Field: "providers.jwt.signingKeys.name", Message: "must not be empty"}
		}
		if _, exists := keyNames[key.Name]; exists {
			return JWTSettings{}, &ValidationError{Field: "providers.jwt.signingKeys.name", Message: "must be unique"}
		}
		keyNames[key.Name] = struct{}{}
		key.SecretRef = strings.TrimSpace(key.SecretRef)
		if key.SecretRef == "" {
			return JWTSettings{}, &ValidationError{Field: "providers.jwt.signingKeys.secretRef", Message: "must not be empty"}
		}
		if utf8.RuneCountInString(key.SecretRef) > maxSecretReferenceLength || !validTLSReference(key.SecretRef) {
			return JWTSettings{}, &ValidationError{Field: "providers.jwt.signingKeys.secretRef", Message: "must be a valid Secret Reference"}
		}
		signingKeys[index] = key
	}

	if len(jwt.AllowedAlgorithms) == 0 {
		return JWTSettings{}, &ValidationError{Field: "providers.jwt.allowedAlgorithms", Message: "must contain at least one algorithm"}
	}
	algorithms := make([]JWTAlgorithm, len(jwt.AllowedAlgorithms))
	seenAlgorithms := make(map[JWTAlgorithm]struct{}, len(jwt.AllowedAlgorithms))
	for index, algorithm := range jwt.AllowedAlgorithms {
		if !validJWTAlgorithm(algorithm) {
			return JWTSettings{}, &ValidationError{Field: "providers.jwt.allowedAlgorithms", Message: "contains an unsupported algorithm"}
		}
		if _, exists := seenAlgorithms[algorithm]; exists {
			return JWTSettings{}, &ValidationError{Field: "providers.jwt.allowedAlgorithms", Message: "must not contain duplicates"}
		}
		seenAlgorithms[algorithm] = struct{}{}
		algorithms[index] = algorithm
	}

	issuers, err := normalizeUniqueStrings(jwt.AllowedIssuers, "providers.jwt.allowedIssuers")
	if err != nil {
		return JWTSettings{}, err
	}
	audiences, err := normalizeUniqueStrings(jwt.AllowedAudiences, "providers.jwt.allowedAudiences")
	if err != nil {
		return JWTSettings{}, err
	}

	requiredClaims := make([]JWTRequiredClaim, len(jwt.RequiredClaims))
	claimNames := make(map[string]struct{}, len(jwt.RequiredClaims))
	for index, claim := range jwt.RequiredClaims {
		claim.Name = strings.TrimSpace(claim.Name)
		claim.Value = strings.TrimSpace(claim.Value)
		if claim.Name == "" {
			return JWTSettings{}, &ValidationError{Field: "providers.jwt.requiredClaims.name", Message: "must not be empty"}
		}
		if claim.Value == "" {
			return JWTSettings{}, &ValidationError{Field: "providers.jwt.requiredClaims.value", Message: "must not be empty"}
		}
		if _, exists := claimNames[claim.Name]; exists {
			return JWTSettings{}, &ValidationError{Field: "providers.jwt.requiredClaims.name", Message: "must be unique"}
		}
		claimNames[claim.Name] = struct{}{}
		requiredClaims[index] = claim
	}

	if jwt.ClockSkewSeconds > 300 {
		return JWTSettings{}, &ValidationError{Field: "providers.jwt.clockSkewSeconds", Message: "must be between 0 and 300"}
	}

	return JWTSettings{
		SigningKeys:       signingKeys,
		AllowedAlgorithms: algorithms,
		AllowedIssuers:    issuers,
		AllowedAudiences:  audiences,
		RequiredClaims:    requiredClaims,
		ClockSkewSeconds:  jwt.ClockSkewSeconds,
	}, nil
}

func validJWTAlgorithm(algorithm JWTAlgorithm) bool {
	switch algorithm {
	case HS256, HS384, HS512, RS256, RS384, RS512, ES256, ES384, ES512, PS256, PS384, PS512:
		return true
	default:
		return false
	}
}

func normalizeUniqueStrings(values []string, field string) ([]string, error) {
	normalized := make([]string, len(values))
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, &ValidationError{Field: field, Message: "must not contain empty values"}
		}
		if _, exists := seen[value]; exists {
			return nil, &ValidationError{Field: field, Message: "must not contain duplicates"}
		}
		seen[value] = struct{}{}
		normalized[index] = value
	}
	return normalized, nil
}
