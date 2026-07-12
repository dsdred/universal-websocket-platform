package configurationversion

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/dsdred/universal-websocket-platform/internal/configuration"
	httpapi "github.com/dsdred/universal-websocket-platform/internal/http"
)

func TestHandlerCreateAndList(t *testing.T) {
	router := newTestRouter(t, true)
	path := "/api/v1/workspaces/1/configurations/1/versions"

	empty := performRequest(router, http.MethodGet, path)
	assertStatus(t, empty, http.StatusOK)
	assertContentType(t, empty)
	if body := strings.TrimSpace(empty.Body.String()); body != "[]" {
		t.Errorf("empty list = %q, want []", body)
	}

	first := performRequest(router, http.MethodPost, path)
	assertStatus(t, first, http.StatusCreated)
	assertContentType(t, first)
	firstVersion := decodeVersion(t, first)
	if firstVersion.Number != 1 || firstVersion.State != Draft || firstVersion.ConfigurationID != 1 {
		t.Errorf("first Version = %#v", firstVersion)
	}

	second := performRequest(router, http.MethodPost, path)
	assertStatus(t, second, http.StatusCreated)
	secondVersion := decodeVersion(t, second)
	if secondVersion.Number != 2 || secondVersion.State != Draft {
		t.Errorf("second Version = %#v", secondVersion)
	}

	list := performRequest(router, http.MethodGet, path)
	assertStatus(t, list, http.StatusOK)
	assertContentType(t, list)
	var versions []ConfigurationVersion
	decodeResponse(t, list, &versions)
	if len(versions) != 2 || versions[0].Number != 1 || versions[1].Number != 2 {
		t.Errorf("list numbers = %v, want [1 2]", versionNumbers(versions))
	}
}

func TestHandlerConfigurationNotFound(t *testing.T) {
	router := newTestRouter(t, false)
	path := "/api/v1/workspaces/1/configurations/42/versions"

	for _, method := range []string{http.MethodPost, http.MethodGet} {
		response := performRequest(router, method, path)
		assertStatus(t, response, http.StatusNotFound)
		assertContentType(t, response)
		var body httpapi.ErrorResponse
		decodeResponse(t, response, &body)
		if body.Error.Code != "configuration_not_found" {
			t.Errorf("%s error code = %q, want configuration_not_found", method, body.Error.Code)
		}
	}
}

func newTestRouter(t *testing.T, configurationExists bool) http.Handler {
	t.Helper()
	configurationRepository := configuration.NewMemoryConfigurationRepository()
	if configurationExists {
		created, err := configurationRepository.Create(configuration.Configuration{WorkspaceID: 1, Name: "Existing"})
		if err != nil {
			t.Fatalf("create Configuration: %v", err)
		}
		if created.ID != 1 {
			t.Fatalf("Configuration ID = %d, want 1", created.ID)
		}
	}
	service := NewService(NewMemoryConfigurationVersionRepository(), configurationRepository, time.Now)
	handler := NewHandler(service)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)
	return router
}

func performRequest(handler http.Handler, method, path string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func decodeVersion(t *testing.T, response *httptest.ResponseRecorder) ConfigurationVersion {
	t.Helper()
	var version ConfigurationVersion
	decodeResponse(t, response, &version)
	return version
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, destination any) {
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

func assertContentType(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if contentType := response.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}
}
