package configurationversion

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	httpapi "github.com/dsdred/universal-websocket-platform/internal/http"
)

// Handler exposes Configuration Version operations over HTTP.
type Handler struct {
	service *Service
}

type timeoutSettingsRequest struct {
	HandshakeSeconds *uint32 `json:"handshakeSeconds"`
	ReadSeconds      *uint32 `json:"readSeconds"`
	WriteSeconds     *uint32 `json:"writeSeconds"`
	IdleSeconds      *uint32 `json:"idleSeconds"`
}

const (
	defaultAPIKeyHeader        = "X-API-Key"
	defaultJWTClockSkewSeconds = 60
)

type authenticationSettingsRequest struct {
	Enabled   bool                            `json:"enabled"`
	Providers []authenticationProviderRequest `json:"providers"`
}

type authenticationProviderRequest struct {
	Name     string                     `json:"name"`
	Type     AuthenticationProviderType `json:"type"`
	Enabled  bool                       `json:"enabled"`
	Priority uint32                     `json:"priority"`
	APIKey   *apiKeySettingsRequest     `json:"apiKey,omitempty"`
	JWT      *jwtSettingsRequest        `json:"jwt,omitempty"`
}

type apiKeySettingsRequest struct {
	Header    *string `json:"header"`
	SecretRef string  `json:"secretRef"`
}

type jwtSettingsRequest struct {
	SigningKeys       []JWTSigningKey    `json:"signingKeys"`
	AllowedAlgorithms []JWTAlgorithm     `json:"allowedAlgorithms"`
	AllowedIssuers    []string           `json:"allowedIssuers"`
	AllowedAudiences  []string           `json:"allowedAudiences"`
	RequiredClaims    []JWTRequiredClaim `json:"requiredClaims"`
	ClockSkewSeconds  *uint32            `json:"clockSkewSeconds"`
}

// NewHandler creates a Configuration Version HTTP handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers nested Configuration Version API routes.
func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Post("/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions", h.create)
	router.Get("/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions", h.list)
	router.Post("/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions/{versionID}/publish", h.publish)
	router.Post("/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions/{versionID}/archive", h.archive)
	router.Put("/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions/{versionID}/listener", h.updateListener)
	router.Put("/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions/{versionID}/listener/tls", h.updateTLS)
	router.Put("/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions/{versionID}/listener/timeouts", h.updateTimeouts)
	router.Put("/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions/{versionID}/authentication", h.updateAuthentication)
}

func (h *Handler) updateAuthentication(w http.ResponseWriter, r *http.Request) {
	workspaceID, configurationID, ok := requestIDs(w, r)
	if !ok {
		return
	}
	versionID, ok := pathID(w, r, "versionID", "Invalid version ID")
	if !ok {
		return
	}

	var request authenticationSettingsRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	version, err := h.service.UpdateAuthentication(r.Context(), workspaceID, configurationID, versionID, request.settings())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, version)
}

func (r authenticationSettingsRequest) settings() AuthenticationSettings {
	providers := make([]AuthenticationProvider, len(r.Providers))
	for index, requestProvider := range r.Providers {
		provider := AuthenticationProvider{
			Name:     requestProvider.Name,
			Type:     requestProvider.Type,
			Enabled:  requestProvider.Enabled,
			Priority: requestProvider.Priority,
		}
		if requestProvider.APIKey != nil {
			header := defaultAPIKeyHeader
			if requestProvider.APIKey.Header != nil {
				header = *requestProvider.APIKey.Header
			}
			provider.APIKey = &APIKeySettings{Header: header, SecretRef: requestProvider.APIKey.SecretRef}
		}
		if requestProvider.JWT != nil {
			clockSkewSeconds := uint32(defaultJWTClockSkewSeconds)
			if requestProvider.JWT.ClockSkewSeconds != nil {
				clockSkewSeconds = *requestProvider.JWT.ClockSkewSeconds
			}
			provider.JWT = &JWTSettings{
				SigningKeys:       append(make([]JWTSigningKey, 0, len(requestProvider.JWT.SigningKeys)), requestProvider.JWT.SigningKeys...),
				AllowedAlgorithms: append(make([]JWTAlgorithm, 0, len(requestProvider.JWT.AllowedAlgorithms)), requestProvider.JWT.AllowedAlgorithms...),
				AllowedIssuers:    append(make([]string, 0, len(requestProvider.JWT.AllowedIssuers)), requestProvider.JWT.AllowedIssuers...),
				AllowedAudiences:  append(make([]string, 0, len(requestProvider.JWT.AllowedAudiences)), requestProvider.JWT.AllowedAudiences...),
				RequiredClaims:    append(make([]JWTRequiredClaim, 0, len(requestProvider.JWT.RequiredClaims)), requestProvider.JWT.RequiredClaims...),
				ClockSkewSeconds:  clockSkewSeconds,
			}
		}
		providers[index] = provider
	}
	return AuthenticationSettings{Enabled: r.Enabled, Providers: providers}
}

func (h *Handler) updateTimeouts(w http.ResponseWriter, r *http.Request) {
	workspaceID, configurationID, ok := requestIDs(w, r)
	if !ok {
		return
	}
	versionID, ok := pathID(w, r, "versionID", "Invalid version ID")
	if !ok {
		return
	}

	var request timeoutSettingsRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil || request.missingValue() {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	version, err := h.service.UpdateTimeouts(r.Context(), workspaceID, configurationID, versionID, TimeoutSettings{
		HandshakeSeconds: *request.HandshakeSeconds,
		ReadSeconds:      *request.ReadSeconds,
		WriteSeconds:     *request.WriteSeconds,
		IdleSeconds:      *request.IdleSeconds,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, version)
}

func (r timeoutSettingsRequest) missingValue() bool {
	return r.HandshakeSeconds == nil || r.ReadSeconds == nil || r.WriteSeconds == nil || r.IdleSeconds == nil
}

func (h *Handler) updateTLS(w http.ResponseWriter, r *http.Request) {
	workspaceID, configurationID, ok := requestIDs(w, r)
	if !ok {
		return
	}
	versionID, ok := pathID(w, r, "versionID", "Invalid version ID")
	if !ok {
		return
	}

	var tls TLSSettings
	if err := httpapi.DecodeJSON(r, &tls); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	version, err := h.service.UpdateTLS(r.Context(), workspaceID, configurationID, versionID, tls)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, version)
}

func (h *Handler) updateListener(w http.ResponseWriter, r *http.Request) {
	workspaceID, configurationID, ok := requestIDs(w, r)
	if !ok {
		return
	}
	versionID, ok := pathID(w, r, "versionID", "Invalid version ID")
	if !ok {
		return
	}

	var listener ListenerSettings
	if err := httpapi.DecodeJSON(r, &listener); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	version, err := h.service.UpdateListener(r.Context(), workspaceID, configurationID, versionID, listener)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, version)
}

func (h *Handler) archive(w http.ResponseWriter, r *http.Request) {
	workspaceID, configurationID, ok := requestIDs(w, r)
	if !ok {
		return
	}
	versionID, ok := pathID(w, r, "versionID", "Invalid version ID")
	if !ok {
		return
	}

	version, err := h.service.Archive(r.Context(), workspaceID, configurationID, versionID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, version)
}

func (h *Handler) publish(w http.ResponseWriter, r *http.Request) {
	workspaceID, configurationID, ok := requestIDs(w, r)
	if !ok {
		return
	}
	versionID, ok := pathID(w, r, "versionID", "Invalid version ID")
	if !ok {
		return
	}

	version, err := h.service.Publish(r.Context(), workspaceID, configurationID, versionID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, version)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	workspaceID, configurationID, ok := requestIDs(w, r)
	if !ok {
		return
	}

	version, err := h.service.Create(r.Context(), workspaceID, configurationID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpapi.WriteJSON(w, http.StatusCreated, version)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	workspaceID, configurationID, ok := requestIDs(w, r)
	if !ok {
		return
	}

	versions, err := h.service.List(r.Context(), workspaceID, configurationID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, versions)
}

func requestIDs(w http.ResponseWriter, r *http.Request) (uint64, uint64, bool) {
	workspaceID, ok := pathID(w, r, "workspaceID", "Invalid workspace ID")
	if !ok {
		return 0, 0, false
	}
	configurationID, ok := pathID(w, r, "configurationID", "Invalid configuration ID")
	if !ok {
		return 0, 0, false
	}
	return workspaceID, configurationID, true
}

func pathID(w http.ResponseWriter, r *http.Request, parameter, message string) (uint64, bool) {
	id, err := strconv.ParseUint(chi.URLParam(r, parameter), 10, 64)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", message)
		return 0, false
	}
	return id, true
}

func writeServiceError(w http.ResponseWriter, err error) {
	var validationError *ValidationError

	switch {
	case errors.As(err, &validationError):
		httpapi.WriteError(w, http.StatusBadRequest, "validation_failed", validationError.Error())
	case errors.Is(err, ErrConfigurationNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "configuration_not_found", "Configuration not found")
	case errors.Is(err, ErrConfigurationVersionNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "version_not_found", "Configuration version not found")
	case errors.Is(err, ErrVersionNotPublishable):
		httpapi.WriteError(w, http.StatusConflict, "version_not_publishable", "Configuration version cannot be published")
	case errors.Is(err, ErrVersionNotArchivable):
		httpapi.WriteError(w, http.StatusConflict, "version_not_archivable", "Configuration version cannot be archived")
	case errors.Is(err, ErrVersionNotEditable):
		httpapi.WriteError(w, http.StatusConflict, "version_not_editable", "Configuration version cannot be edited")
	default:
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
	}
}
