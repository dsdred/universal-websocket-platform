package configurationversion

import (
	"encoding/json"
	"io"
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

func TestHandlerPublish(t *testing.T) {
	router := newTestRouter(t, true)
	versionsPath := "/api/v1/workspaces/1/configurations/1/versions"
	first := decodeVersion(t, performRequest(router, http.MethodPost, versionsPath))
	second := decodeVersion(t, performRequest(router, http.MethodPost, versionsPath))

	publishFirst := performRequest(router, http.MethodPost, versionsPath+"/1/publish")
	assertStatus(t, publishFirst, http.StatusOK)
	assertContentType(t, publishFirst)
	if published := decodeVersion(t, publishFirst); published.ID != first.ID || published.State != Published {
		t.Errorf("published first = %#v", published)
	}

	repeat := performRequest(router, http.MethodPost, versionsPath+"/1/publish")
	assertStatus(t, repeat, http.StatusConflict)
	assertErrorCode(t, repeat, "version_not_publishable")

	publishSecond := performRequest(router, http.MethodPost, versionsPath+"/2/publish")
	assertStatus(t, publishSecond, http.StatusOK)
	if published := decodeVersion(t, publishSecond); published.ID != second.ID || published.State != Published {
		t.Errorf("published second = %#v", published)
	}

	list := performRequest(router, http.MethodGet, versionsPath)
	var versions []ConfigurationVersion
	decodeResponse(t, list, &versions)
	if len(versions) != 2 || versions[0].State != Archived || versions[1].State != Published {
		t.Errorf("versions after second publish = %#v, want Archived then Published", versions)
	}

	archived := performRequest(router, http.MethodPost, versionsPath+"/1/publish")
	assertStatus(t, archived, http.StatusConflict)
	assertErrorCode(t, archived, "version_not_publishable")

	notFound := performRequest(router, http.MethodPost, versionsPath+"/999/publish")
	assertStatus(t, notFound, http.StatusNotFound)
	assertErrorCode(t, notFound, "version_not_found")
}

func TestHandlerArchive(t *testing.T) {
	t.Run("Draft", func(t *testing.T) {
		router := newTestRouter(t, true)
		versionsPath := "/api/v1/workspaces/1/configurations/1/versions"
		created := decodeVersion(t, performRequest(router, http.MethodPost, versionsPath))

		archivedResponse := performRequest(router, http.MethodPost, versionsPath+"/1/archive")
		assertStatus(t, archivedResponse, http.StatusOK)
		assertContentType(t, archivedResponse)
		archived := decodeVersion(t, archivedResponse)
		if archived.ID != created.ID || archived.State != Archived {
			t.Errorf("archived Draft = %#v", archived)
		}

		repeat := performRequest(router, http.MethodPost, versionsPath+"/1/archive")
		assertStatus(t, repeat, http.StatusConflict)
		assertErrorCode(t, repeat, "version_not_archivable")
	})

	t.Run("Published", func(t *testing.T) {
		router := newTestRouter(t, true)
		versionsPath := "/api/v1/workspaces/1/configurations/1/versions"
		performRequest(router, http.MethodPost, versionsPath)
		publish := performRequest(router, http.MethodPost, versionsPath+"/1/publish")
		assertStatus(t, publish, http.StatusOK)

		archive := performRequest(router, http.MethodPost, versionsPath+"/1/archive")
		assertStatus(t, archive, http.StatusOK)
		assertContentType(t, archive)
		if archived := decodeVersion(t, archive); archived.State != Archived {
			t.Errorf("archived Published = %#v", archived)
		}
	})
}

func TestHandlerArchiveNotFound(t *testing.T) {
	missingConfiguration := performRequest(
		newTestRouter(t, false),
		http.MethodPost,
		"/api/v1/workspaces/1/configurations/42/versions/1/archive",
	)
	assertStatus(t, missingConfiguration, http.StatusNotFound)
	assertErrorCode(t, missingConfiguration, "configuration_not_found")

	router := newTestRouter(t, true)
	missingVersion := performRequest(
		router,
		http.MethodPost,
		"/api/v1/workspaces/1/configurations/1/versions/42/archive",
	)
	assertStatus(t, missingVersion, http.StatusNotFound)
	assertErrorCode(t, missingVersion, "version_not_found")
}

func TestHandlerUpdateListener(t *testing.T) {
	router := newTestRouter(t, true)
	versionsPath := "/api/v1/workspaces/1/configurations/1/versions"
	created := decodeVersion(t, performRequest(router, http.MethodPost, versionsPath))
	listenerPath := versionsPath + "/1/listener"

	response := performRequestWithBody(router, http.MethodPut, listenerPath, `{"host":"0.0.0.0","port":9000}`)
	assertStatus(t, response, http.StatusOK)
	assertContentType(t, response)
	updated := decodeVersion(t, response)
	if updated.ID != created.ID || updated.State != Draft || updated.Listener.Host != "0.0.0.0" || updated.Listener.Port != 9000 || updated.Listener.TLS.MinVersion != "1.2" {
		t.Errorf("updated Version = %#v", updated)
	}

	publish := performRequest(router, http.MethodPost, versionsPath+"/1/publish")
	assertStatus(t, publish, http.StatusOK)
	published := decodeVersion(t, publish)
	if published.Listener != updated.Listener {
		t.Errorf("Publish changed Listener from %#v to %#v", updated.Listener, published.Listener)
	}

	conflict := performRequestWithBody(router, http.MethodPut, listenerPath, `{"host":"localhost","port":8080}`)
	assertStatus(t, conflict, http.StatusConflict)
	assertErrorCode(t, conflict, "version_not_editable")
}

func TestHandlerUpdateListenerInvalidRequest(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode string
	}{
		{name: "validation failed", body: `{"host":"","port":8080}`, wantCode: "validation_failed"},
		{name: "zero port", body: `{"host":"localhost","port":0}`, wantCode: "validation_failed"},
		{name: "malformed JSON", body: `{"host":`, wantCode: "invalid_request"},
		{name: "unknown field", body: `{"host":"localhost","port":8080,"tls":true}`, wantCode: "invalid_request"},
		{name: "empty body", body: "", wantCode: "invalid_request"},
		{name: "additional JSON", body: `{"host":"localhost","port":8080}{}`, wantCode: "invalid_request"},
		{name: "string port", body: `{"host":"localhost","port":"8080"}`, wantCode: "invalid_request"},
		{name: "fractional port", body: `{"host":"localhost","port":1.5}`, wantCode: "invalid_request"},
		{name: "negative port", body: `{"host":"localhost","port":-1}`, wantCode: "invalid_request"},
		{name: "port above range", body: `{"host":"localhost","port":65536}`, wantCode: "invalid_request"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newTestRouter(t, true)
			performRequest(router, http.MethodPost, "/api/v1/workspaces/1/configurations/1/versions")
			response := performRequestWithBody(
				router,
				http.MethodPut,
				"/api/v1/workspaces/1/configurations/1/versions/1/listener",
				tt.body,
			)
			assertStatus(t, response, http.StatusBadRequest)
			assertContentType(t, response)
			assertErrorCode(t, response, tt.wantCode)
		})
	}
}

func TestHandlerUpdateListenerNotFound(t *testing.T) {
	missingConfiguration := performRequestWithBody(
		newTestRouter(t, false),
		http.MethodPut,
		"/api/v1/workspaces/1/configurations/42/versions/1/listener",
		`{"host":"localhost","port":8080}`,
	)
	assertStatus(t, missingConfiguration, http.StatusNotFound)
	assertErrorCode(t, missingConfiguration, "configuration_not_found")

	router := newTestRouter(t, true)
	missingVersion := performRequestWithBody(
		router,
		http.MethodPut,
		"/api/v1/workspaces/1/configurations/1/versions/42/listener",
		`{"host":"localhost","port":8080}`,
	)
	assertStatus(t, missingVersion, http.StatusNotFound)
	assertErrorCode(t, missingVersion, "version_not_found")
}

func TestHandlerUpdateTLS(t *testing.T) {
	router := newTestRouter(t, true)
	versionsPath := "/api/v1/workspaces/1/configurations/1/versions"
	created := decodeVersion(t, performRequest(router, http.MethodPost, versionsPath))
	tlsPath := versionsPath + "/1/listener/tls"
	body := `{"enabled":true,"certificateRef":"certificates/main","privateKeyRef":"secrets/tls-key","minVersion":"1.3"}`

	response := performRequestWithBody(router, http.MethodPut, tlsPath, body)
	assertStatus(t, response, http.StatusOK)
	assertContentType(t, response)
	updated := decodeVersion(t, response)
	wantTLS := TLSSettings{Enabled: true, CertificateRef: "certificates/main", PrivateKeyRef: "secrets/tls-key", MinVersion: "1.3"}
	if updated.ID != created.ID || updated.State != Draft || updated.Listener.TLS != wantTLS {
		t.Errorf("updated Version = %#v", updated)
	}
	if updated.Listener.Host != "127.0.0.1" || updated.Listener.Port != 8080 {
		t.Errorf("UpdateTLS changed Host/Port: %#v", updated.Listener)
	}
}

func TestHandlerUpdateTLSInvalidRequest(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode string
	}{
		{name: "validation failed", body: `{"enabled":true,"certificateRef":"","privateKeyRef":"secrets/key","minVersion":"1.2"}`, wantCode: "validation_failed"},
		{name: "null certificate", body: `{"enabled":true,"certificateRef":null,"privateKeyRef":"secrets/key","minVersion":"1.2"}`, wantCode: "validation_failed"},
		{name: "null minimum version", body: `{"enabled":false,"certificateRef":"","privateKeyRef":"","minVersion":null}`, wantCode: "validation_failed"},
		{name: "long reference", body: `{"enabled":true,"certificateRef":"` + strings.Repeat("a", 256) + `","privateKeyRef":"secrets/key","minVersion":"1.2"}`, wantCode: "validation_failed"},
		{name: "malformed JSON", body: `{"enabled":`, wantCode: "invalid_request"},
		{name: "unknown field", body: `{"enabled":false,"certificateRef":"","privateKeyRef":"","minVersion":"1.2","acme":true}`, wantCode: "invalid_request"},
		{name: "empty body", body: "", wantCode: "invalid_request"},
		{name: "additional JSON", body: `{"enabled":false,"certificateRef":"","privateKeyRef":"","minVersion":"1.2"}{}`, wantCode: "invalid_request"},
		{name: "string enabled", body: `{"enabled":"true","certificateRef":"certificates/main","privateKeyRef":"secrets/key","minVersion":"1.2"}`, wantCode: "invalid_request"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newTestRouter(t, true)
			performRequest(router, http.MethodPost, "/api/v1/workspaces/1/configurations/1/versions")
			response := performRequestWithBody(
				router,
				http.MethodPut,
				"/api/v1/workspaces/1/configurations/1/versions/1/listener/tls",
				tt.body,
			)
			assertStatus(t, response, http.StatusBadRequest)
			assertContentType(t, response)
			assertErrorCode(t, response, tt.wantCode)
		})
	}
}

func TestHandlerUpdateTLSNotFoundAndState(t *testing.T) {
	validBody := `{"enabled":false,"certificateRef":"","privateKeyRef":"","minVersion":"1.2"}`
	missingConfiguration := performRequestWithBody(
		newTestRouter(t, false),
		http.MethodPut,
		"/api/v1/workspaces/1/configurations/42/versions/1/listener/tls",
		validBody,
	)
	assertStatus(t, missingConfiguration, http.StatusNotFound)
	assertErrorCode(t, missingConfiguration, "configuration_not_found")

	router := newTestRouter(t, true)
	missingVersion := performRequestWithBody(
		router,
		http.MethodPut,
		"/api/v1/workspaces/1/configurations/1/versions/42/listener/tls",
		validBody,
	)
	assertStatus(t, missingVersion, http.StatusNotFound)
	assertErrorCode(t, missingVersion, "version_not_found")

	versionsPath := "/api/v1/workspaces/1/configurations/1/versions"
	performRequest(router, http.MethodPost, versionsPath)
	performRequest(router, http.MethodPost, versionsPath+"/1/publish")
	conflict := performRequestWithBody(router, http.MethodPut, versionsPath+"/1/listener/tls", validBody)
	assertStatus(t, conflict, http.StatusConflict)
	assertErrorCode(t, conflict, "version_not_editable")
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
	return performRequestWithBody(handler, method, path, "")
}

func performRequestWithBody(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
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

func assertErrorCode(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	assertContentType(t, response)
	var body httpapi.ErrorResponse
	decodeResponse(t, response, &body)
	if body.Error.Code != want {
		t.Errorf("error code = %q, want %q", body.Error.Code, want)
	}
}
