package workspace

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Handler exposes Workspace operations over HTTP.
type Handler struct {
	service *WorkspaceService
}

type workspaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewHandler creates a Workspace HTTP handler.
func NewHandler(service *WorkspaceService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers Workspace API routes.
func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Post("/api/v1/workspaces", h.create)
	router.Get("/api/v1/workspaces", h.list)
	router.Get("/api/v1/workspaces/{id}", h.get)
	router.Put("/api/v1/workspaces/{id}", h.update)
	router.Delete("/api/v1/workspaces/{id}", h.delete)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	request, ok := decodeWorkspaceRequest(w, r)
	if !ok {
		return
	}

	workspace, err := h.service.Create(CreateWorkspace{
		Name:        request.Name,
		Description: request.Description,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Location", "/api/v1/workspaces/"+strconv.FormatUint(workspace.ID, 10))
	writeJSON(w, http.StatusCreated, workspace)
}

func (h *Handler) list(w http.ResponseWriter, _ *http.Request) {
	workspaces, err := h.service.List()
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, workspaces)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := workspaceID(w, r)
	if !ok {
		return
	}

	workspace, err := h.service.Get(id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, workspace)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := workspaceID(w, r)
	if !ok {
		return
	}

	request, ok := decodeWorkspaceRequest(w, r)
	if !ok {
		return
	}

	workspace, err := h.service.Update(id, UpdateWorkspace{
		Name:        request.Name,
		Description: request.Description,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, workspace)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := workspaceID(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(id); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func workspaceID(w http.ResponseWriter, r *http.Request) (uint64, bool) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid workspace ID")
		return 0, false
	}

	return id, true
}

func decodeWorkspaceRequest(w http.ResponseWriter, r *http.Request) (workspaceRequest, bool) {
	var request workspaceRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return workspaceRequest{}, false
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return workspaceRequest{}, false
	}

	return request, true
}

func writeServiceError(w http.ResponseWriter, err error) {
	var validationError *ValidationError

	switch {
	case errors.As(err, &validationError):
		writeAPIError(w, http.StatusBadRequest, "validation_failed", validationError.Error())
	case errors.Is(err, ErrWorkspaceNotFound):
		writeAPIError(w, http.StatusNotFound, "workspace_not_found", "Workspace not found")
	default:
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
	}
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Error: errorBody{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
