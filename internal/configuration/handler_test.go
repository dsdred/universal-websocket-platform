package configuration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	httpapi "github.com/dsdred/universal-websocket-platform/internal/http"
	"github.com/dsdred/universal-websocket-platform/internal/workspace"
)

func TestHandlerCRUDAndWorkspaceDeleteProtection(t *testing.T) {
	router := newAPITestRouter()
	workspaceOne := createWorkspace(t, router, "Workspace One")
	workspaceTwo := createWorkspace(t, router, "Workspace Two")

	emptyList := request(t, router, http.MethodGet, workspaceConfigurationsPath(workspaceOne.ID), "")
	assertResponseStatus(t, emptyList, http.StatusOK)
	assertResponseContentType(t, emptyList)
	if body := strings.TrimSpace(emptyList.Body.String()); body != "[]" {
		t.Errorf("empty list body = %q, want []", body)
	}

	created := request(t, router, http.MethodPost, workspaceConfigurationsPath(workspaceOne.ID),
		`{"name":"Notification Server","description":"Configuration for realtime notifications"}`)
	assertResponseStatus(t, created, http.StatusCreated)
	assertResponseContentType(t, created)
	if location := created.Header().Get("Location"); location != "/api/v1/workspaces/1/configurations/1" {
		t.Errorf("Location = %q, want %q", location, "/api/v1/workspaces/1/configurations/1")
	}
	createdConfiguration := decodeConfiguration(t, created)
	if createdConfiguration.ID != 1 || createdConfiguration.WorkspaceID != workspaceOne.ID {
		t.Errorf("created Configuration = %#v", createdConfiguration)
	}

	list := request(t, router, http.MethodGet, workspaceConfigurationsPath(workspaceOne.ID), "")
	assertResponseStatus(t, list, http.StatusOK)
	var configurations []Configuration
	decodeResponse(t, list, &configurations)
	if len(configurations) != 1 || configurations[0] != createdConfiguration {
		t.Errorf("list = %#v, want created Configuration", configurations)
	}

	get := request(t, router, http.MethodGet, configurationPath(workspaceOne.ID, createdConfiguration.ID), "")
	assertResponseStatus(t, get, http.StatusOK)
	assertResponseContentType(t, get)
	if got := decodeConfiguration(t, get); got != createdConfiguration {
		t.Errorf("GET = %#v, want %#v", got, createdConfiguration)
	}

	otherWorkspace := request(t, router, http.MethodGet, configurationPath(workspaceTwo.ID, createdConfiguration.ID), "")
	assertResponseStatus(t, otherWorkspace, http.StatusNotFound)
	assertErrorCode(t, otherWorkspace, "configuration_not_found")

	updated := request(t, router, http.MethodPut, configurationPath(workspaceOne.ID, createdConfiguration.ID),
		`{"name":"Updated Notification Server","description":"Updated description"}`)
	assertResponseStatus(t, updated, http.StatusOK)
	assertResponseContentType(t, updated)
	updatedConfiguration := decodeConfiguration(t, updated)
	if updatedConfiguration.WorkspaceID != workspaceOne.ID || updatedConfiguration.Name != "Updated Notification Server" {
		t.Errorf("updated Configuration = %#v", updatedConfiguration)
	}

	blockedDelete := request(t, router, http.MethodDelete, "/api/v1/workspaces/1", "")
	assertResponseStatus(t, blockedDelete, http.StatusConflict)
	assertResponseContentType(t, blockedDelete)
	assertErrorCode(t, blockedDelete, "workspace_not_empty")

	deletedConfiguration := request(t, router, http.MethodDelete, configurationPath(workspaceOne.ID, createdConfiguration.ID), "")
	assertResponseStatus(t, deletedConfiguration, http.StatusNoContent)
	if deletedConfiguration.Body.Len() != 0 || deletedConfiguration.Header().Get("Content-Type") != "" {
		t.Errorf("Configuration DELETE body/content-type = %q/%q, want empty", deletedConfiguration.Body.String(), deletedConfiguration.Header().Get("Content-Type"))
	}

	deletedWorkspace := request(t, router, http.MethodDelete, "/api/v1/workspaces/1", "")
	assertResponseStatus(t, deletedWorkspace, http.StatusNoContent)
}

func TestHandlerNotFound(t *testing.T) {
	router := newAPITestRouter()
	workspace := createWorkspace(t, router, "Existing")

	missingWorkspace := request(t, router, http.MethodGet, workspaceConfigurationsPath(42), "")
	assertResponseStatus(t, missingWorkspace, http.StatusNotFound)
	assertErrorCode(t, missingWorkspace, "workspace_not_found")

	missingConfiguration := request(t, router, http.MethodGet, configurationPath(workspace.ID, 42), "")
	assertResponseStatus(t, missingConfiguration, http.StatusNotFound)
	assertErrorCode(t, missingConfiguration, "configuration_not_found")
}

func TestHandlerValidationError(t *testing.T) {
	router := newAPITestRouter()
	workspace := createWorkspace(t, router, "Existing")
	response := request(t, router, http.MethodPost, workspaceConfigurationsPath(workspace.ID), `{"name":"   ","description":""}`)

	assertResponseStatus(t, response, http.StatusBadRequest)
	assertResponseContentType(t, response)
	assertErrorCode(t, response, "validation_failed")
}

func TestHandlerInvalidJSON(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed", body: `{"name":`},
		{name: "unknown field", body: `{"name":"Main","workspaceId":1}`},
		{name: "empty", body: ""},
		{name: "additional JSON", body: `{"name":"First"}{"name":"Second"}`},
		{name: "additional content", body: `{"name":"First"} trailing`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newAPITestRouter()
			workspace := createWorkspace(t, router, "Existing")
			response := request(t, router, http.MethodPost, workspaceConfigurationsPath(workspace.ID), tt.body)
			assertResponseStatus(t, response, http.StatusBadRequest)
			assertResponseContentType(t, response)
			assertErrorCode(t, response, "invalid_request")
		})
	}
}

func newAPITestRouter() http.Handler {
	workspaceRepository := workspace.NewMemoryWorkspaceRepository()
	configurationRepository := NewMemoryConfigurationRepository()
	nextTime := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	now := func() time.Time {
		current := nextTime
		nextTime = nextTime.Add(time.Minute)
		return current
	}

	workspaceService := workspace.NewWorkspaceService(workspaceRepository, configurationRepository, now)
	configurationService := NewService(configurationRepository, workspaceRepository, now)
	workspaceHandler := workspace.NewHandler(workspaceService)
	configurationHandler := NewHandler(configurationService)
	router := chi.NewRouter()
	workspaceHandler.RegisterRoutes(router)
	configurationHandler.RegisterRoutes(router)
	return router
}

func createWorkspace(t *testing.T, router http.Handler, name string) workspace.Workspace {
	t.Helper()
	response := request(t, router, http.MethodPost, "/api/v1/workspaces", `{"name":"`+name+`","description":""}`)
	assertResponseStatus(t, response, http.StatusCreated)
	var created workspace.Workspace
	decodeResponse(t, response, &created)
	return created
}

func workspaceConfigurationsPath(workspaceID uint64) string {
	return "/api/v1/workspaces/" + uintString(workspaceID) + "/configurations"
}

func configurationPath(workspaceID, configurationID uint64) string {
	return workspaceConfigurationsPath(workspaceID) + "/" + uintString(configurationID)
}

func uintString(value uint64) string {
	return strconv.FormatUint(value, 10)
}

func request(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}

func decodeConfiguration(t *testing.T, response *httptest.ResponseRecorder) Configuration {
	t.Helper()
	var configuration Configuration
	decodeResponse(t, response, &configuration)
	return configuration
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, destination any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func assertResponseStatus(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if response.Code != want {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, want, response.Body.String())
	}
}

func assertResponseContentType(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if contentType := response.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}
}

func assertErrorCode(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	var body httpapi.ErrorResponse
	decodeResponse(t, response, &body)
	if body.Error.Code != want {
		t.Errorf("error code = %q, want %q", body.Error.Code, want)
	}
}
