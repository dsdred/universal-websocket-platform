package workspace

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestHandlerCRUD(t *testing.T) {
	router := newTestRouter()

	created := performRequest(t, router, http.MethodPost, "/api/v1/workspaces", `{"name":"Main Workspace","description":"Primary workspace"}`)
	assertStatus(t, created, http.StatusCreated)
	assertJSONContentType(t, created)
	if location := created.Header().Get("Location"); location != "/api/v1/workspaces/1" {
		t.Errorf("Location = %q, want %q", location, "/api/v1/workspaces/1")
	}
	createdWorkspace := decodeWorkspace(t, created)
	if createdWorkspace.ID != 1 || createdWorkspace.Name != "Main Workspace" {
		t.Errorf("created Workspace = %#v", createdWorkspace)
	}

	second := performRequest(t, router, http.MethodPost, "/api/v1/workspaces", `{"name":"Second","description":""}`)
	assertStatus(t, second, http.StatusCreated)

	list := performRequest(t, router, http.MethodGet, "/api/v1/workspaces", "")
	assertStatus(t, list, http.StatusOK)
	assertJSONContentType(t, list)
	var workspaces []Workspace
	decodeJSON(t, list, &workspaces)
	if len(workspaces) != 2 || workspaces[0].ID != 1 || workspaces[1].ID != 2 {
		t.Errorf("list IDs = %v, want [1 2]", workspaceIDs(workspaces))
	}

	get := performRequest(t, router, http.MethodGet, "/api/v1/workspaces/1", "")
	assertStatus(t, get, http.StatusOK)
	assertJSONContentType(t, get)
	gotWorkspace := decodeWorkspace(t, get)
	if gotWorkspace != createdWorkspace {
		t.Errorf("GET Workspace = %#v, want %#v", gotWorkspace, createdWorkspace)
	}

	update := performRequest(t, router, http.MethodPut, "/api/v1/workspaces/1", `{"name":"Updated Workspace","description":"Updated description"}`)
	assertStatus(t, update, http.StatusOK)
	assertJSONContentType(t, update)
	updatedWorkspace := decodeWorkspace(t, update)
	if updatedWorkspace.ID != createdWorkspace.ID || !updatedWorkspace.CreatedAt.Equal(createdWorkspace.CreatedAt) {
		t.Errorf("PUT changed immutable fields: %#v", updatedWorkspace)
	}
	if updatedWorkspace.Name != "Updated Workspace" || updatedWorkspace.Description != "Updated description" {
		t.Errorf("PUT Workspace = %#v", updatedWorkspace)
	}

	deleted := performRequest(t, router, http.MethodDelete, "/api/v1/workspaces/1", "")
	assertStatus(t, deleted, http.StatusNoContent)
	if deleted.Body.Len() != 0 {
		t.Errorf("DELETE body = %q, want empty", deleted.Body.String())
	}
	if contentType := deleted.Header().Get("Content-Type"); contentType != "" {
		t.Errorf("DELETE Content-Type = %q, want empty", contentType)
	}

	notFound := performRequest(t, router, http.MethodGet, "/api/v1/workspaces/1", "")
	assertStatus(t, notFound, http.StatusNotFound)
	assertAPIError(t, notFound, "workspace_not_found", "Workspace not found")
}

func TestHandlerValidationError(t *testing.T) {
	router := newTestRouter()
	response := performRequest(t, router, http.MethodPost, "/api/v1/workspaces", `{"name":"   ","description":""}`)

	assertStatus(t, response, http.StatusBadRequest)
	assertJSONContentType(t, response)
	assertAPIErrorCode(t, response, "validation_failed")
}

func TestHandlerInvalidRequestBody(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed JSON", body: `{"name":`},
		{name: "unknown field", body: `{"name":"Main","unknown":true}`},
		{name: "empty body", body: ""},
		{name: "additional JSON", body: `{"name":"Main"}{"name":"Second"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := performRequest(t, newTestRouter(), http.MethodPost, "/api/v1/workspaces", tt.body)
			assertStatus(t, response, http.StatusBadRequest)
			assertJSONContentType(t, response)
			assertAPIErrorCode(t, response, "invalid_request")
		})
	}
}

func TestHandlerInvalidWorkspaceID(t *testing.T) {
	response := performRequest(t, newTestRouter(), http.MethodGet, "/api/v1/workspaces/not-a-number", "")

	assertStatus(t, response, http.StatusBadRequest)
	assertAPIErrorCode(t, response, "invalid_request")
}

func newTestRouter() http.Handler {
	repository := NewMemoryWorkspaceRepository()
	nextTime := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	service := NewWorkspaceService(repository, func() time.Time {
		current := nextTime
		nextTime = nextTime.Add(time.Minute)
		return current
	})
	handler := NewHandler(service)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)
	return router
}

func performRequest(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func decodeWorkspace(t *testing.T, response *httptest.ResponseRecorder) Workspace {
	t.Helper()
	var workspace Workspace
	decodeJSON(t, response, &workspace)
	return workspace
}

func decodeJSON(t *testing.T, response *httptest.ResponseRecorder, destination any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func assertStatus(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if response.Code != want {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, want, response.Body.String())
	}
}

func assertJSONContentType(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if contentType := response.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}
}

func assertAPIErrorCode(t *testing.T, response *httptest.ResponseRecorder, wantCode string) {
	t.Helper()
	var body errorResponse
	decodeJSON(t, response, &body)
	if body.Error.Code != wantCode {
		t.Errorf("error code = %q, want %q", body.Error.Code, wantCode)
	}
}

func assertAPIError(t *testing.T, response *httptest.ResponseRecorder, wantCode, wantMessage string) {
	t.Helper()
	var body errorResponse
	decodeJSON(t, response, &body)
	if body.Error.Code != wantCode || body.Error.Message != wantMessage {
		t.Errorf("error = %#v, want code %q and message %q", body.Error, wantCode, wantMessage)
	}
}
