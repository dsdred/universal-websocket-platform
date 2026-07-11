package configuration

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	httpapi "github.com/dsdred/universal-websocket-platform/internal/http"
)

// Handler exposes Configuration operations over HTTP.
type Handler struct {
	service *Service
}

type configurationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NewHandler creates a Configuration HTTP handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers nested Configuration API routes.
func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Post("/api/v1/workspaces/{workspaceID}/configurations", h.create)
	router.Get("/api/v1/workspaces/{workspaceID}/configurations", h.list)
	router.Get("/api/v1/workspaces/{workspaceID}/configurations/{configurationID}", h.get)
	router.Put("/api/v1/workspaces/{workspaceID}/configurations/{configurationID}", h.update)
	router.Delete("/api/v1/workspaces/{workspaceID}/configurations/{configurationID}", h.delete)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := pathID(w, r, "workspaceID", "Invalid workspace ID")
	if !ok {
		return
	}

	request, ok := decodeConfigurationRequest(w, r)
	if !ok {
		return
	}

	configuration, err := h.service.Create(r.Context(), workspaceID, CreateConfiguration{
		Name:        request.Name,
		Description: request.Description,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	location := "/api/v1/workspaces/" + strconv.FormatUint(workspaceID, 10) +
		"/configurations/" + strconv.FormatUint(configuration.ID, 10)
	w.Header().Set("Location", location)
	httpapi.WriteJSON(w, http.StatusCreated, configuration)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := pathID(w, r, "workspaceID", "Invalid workspace ID")
	if !ok {
		return
	}

	configurations, err := h.service.List(r.Context(), workspaceID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, configurations)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	workspaceID, configurationID, ok := requestIDs(w, r)
	if !ok {
		return
	}

	configuration, err := h.service.Get(r.Context(), workspaceID, configurationID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, configuration)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	workspaceID, configurationID, ok := requestIDs(w, r)
	if !ok {
		return
	}

	request, ok := decodeConfigurationRequest(w, r)
	if !ok {
		return
	}

	configuration, err := h.service.Update(r.Context(), workspaceID, configurationID, UpdateConfiguration{
		Name:        request.Name,
		Description: request.Description,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, configuration)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	workspaceID, configurationID, ok := requestIDs(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), workspaceID, configurationID); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

func decodeConfigurationRequest(w http.ResponseWriter, r *http.Request) (configurationRequest, bool) {
	var request configurationRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return configurationRequest{}, false
	}
	return request, true
}

func writeServiceError(w http.ResponseWriter, err error) {
	var validationError *ValidationError

	switch {
	case errors.As(err, &validationError):
		httpapi.WriteError(w, http.StatusBadRequest, "validation_failed", validationError.Error())
	case errors.Is(err, ErrWorkspaceNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "workspace_not_found", "Workspace not found")
	case errors.Is(err, ErrConfigurationNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "configuration_not_found", "Configuration not found")
	default:
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
	}
}
