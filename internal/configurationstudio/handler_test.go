package configurationstudio

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestEmbeddedAssetRoutes(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		contentType string
		body        []byte
	}{
		{name: "index", path: "/", contentType: "text/html; charset=utf-8", body: indexHTML},
		{name: "stylesheet", path: "/configuration-studio/studio.css", contentType: "text/css; charset=utf-8", body: studioCSS},
		{name: "script", path: "/configuration-studio/studio.js", contentType: "text/javascript; charset=utf-8", body: studioJS},
		{name: "dashboard", path: "/dashboard", contentType: "text/html; charset=utf-8", body: dashboardHTML},
		{name: "control theme", path: "/configuration-studio/control.css", contentType: "text/css; charset=utf-8", body: controlCSS},
		{name: "dashboard script", path: "/configuration-studio/dashboard.js", contentType: "text/javascript; charset=utf-8", body: dashboardJS},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			newRouter().ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
			if got := response.Header().Get("Content-Type"); got != test.contentType {
				t.Errorf("Content-Type = %q, want %q", got, test.contentType)
			}
			if got := response.Header().Get("X-Content-Type-Options"); got != "nosniff" {
				t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
			}
			if got := response.Header().Get("Content-Security-Policy"); got != contentSecurityPolicy {
				t.Errorf("Content-Security-Policy = %q, want %q", got, contentSecurityPolicy)
			}
			if got := response.Body.Bytes(); string(got) != string(test.body) {
				t.Error("response bytes differ from embedded asset")
			}
		})
	}
}

func TestRoutesAreExact(t *testing.T) {
	tests := []struct {
		method string
		path   string
		status int
	}{
		{method: http.MethodGet, path: "/configuration-studio", status: http.StatusNotFound},
		{method: http.MethodGet, path: "/configuration-studio/", status: http.StatusNotFound},
		{method: http.MethodGet, path: "/configuration-studio/missing.js", status: http.StatusNotFound},
		{method: http.MethodPost, path: "/", status: http.StatusMethodNotAllowed},
		{method: http.MethodGet, path: "/dashboard/", status: http.StatusNotFound},
		{method: http.MethodGet, path: "/dashboard/missing", status: http.StatusNotFound},
		{method: http.MethodPost, path: "/dashboard", status: http.StatusMethodNotAllowed},
		{method: http.MethodPut, path: "/dashboard", status: http.StatusMethodNotAllowed},
		{method: http.MethodDelete, path: "/dashboard", status: http.StatusMethodNotAllowed},
		{method: http.MethodPost, path: "/configuration-studio/dashboard.js", status: http.StatusMethodNotAllowed},
	}

	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.path, nil)
		response := httptest.NewRecorder()
		newRouter().ServeHTTP(response, request)
		if response.Code != test.status {
			t.Errorf("%s %s status = %d, want %d", test.method, test.path, response.Code, test.status)
		}
	}
}

func TestAssetsAreSelfContainedAndShowRequiredNotices(t *testing.T) {
	html := string(indexHTML)
	for _, required := range []string{
		`href="/configuration-studio/studio.css"`,
		`src="/configuration-studio/studio.js"`,
		"trusted local development environment",
		"Published does not mean Running.",
		"Control Service data is stored only in memory",
		"is lost after the Control Service restarts.",
		"Any WebSocket URL shown here is preliminary.",
	} {
		if !strings.Contains(html, required) {
			t.Errorf("index asset is missing %q", required)
		}
	}

	allAssets := string(indexHTML) + string(studioCSS) + string(studioJS) + string(dashboardHTML) + string(controlCSS) + string(dashboardJS)
	for _, forbidden := range []string{"https://", "http://", "//cdn", "@import", "innerHTML", "new WebSocket", "window.open"} {
		if strings.Contains(allAssets, forbidden) {
			t.Errorf("embedded assets contain forbidden dependency or behavior %q", forbidden)
		}
	}
}

func TestDashboardReadOnlyBoundaries(t *testing.T) {
	html := string(dashboardHTML)
	for _, required := range []string{
		`href="/"`, `href="/dashboard#workspaces"`, `href="/dashboard#configurations"`,
		`href="/configuration-studio/control.css"`, `src="/configuration-studio/dashboard.js"`,
		"Published does not mean Running.", "No HTTP GET API", "Unavailable", "Refresh",
		"Unauthenticated Control Service", `aria-live="polite"`,
	} {
		if !strings.Contains(html, required) {
			t.Errorf("Dashboard is missing required boundary/navigation %q", required)
		}
	}
	for _, forbidden := range []string{`"POST"`, `"PUT"`, `"DELETE"`, `"PATCH"`, "setInterval", "WebSocket", "/runtime", "/diagnostics", "secretRef", "certificateRef", "privateKeyRef", "innerHTML"} {
		if strings.Contains(string(dashboardJS), forbidden) {
			t.Errorf("Dashboard contains forbidden behavior %q", forbidden)
		}
	}
	if strings.Contains(html, `src="/configuration-studio/studio.js"`) {
		t.Error("Dashboard must not load the Studio mutation script")
	}
}

func TestScriptStaysWithinApprovedAPIScope(t *testing.T) {
	script := string(studioJS)
	for _, required := range []string{
		`"/api/v1/workspaces"`,
		`/configurations`,
		`/versions`,
		`/listener`,
		`/publish`,
		`Number.isSafeInteger`,
		`textContent`,
		`ws://${renderedHost}:${port}/ws`,
	} {
		if !strings.Contains(script, required) {
			t.Errorf("script is missing required behavior %q", required)
		}
	}

	for _, forbidden := range []string{
		`"DELETE"`,
		`/archive`,
		`/runtime`,
		`/recovery`,
		`/authentication`,
		`/routing`,
		`/timeouts`,
		`listener/tls`,
		`setInterval`,
		`setTimeout`,
	} {
		if strings.Contains(script, forbidden) {
			t.Errorf("script contains out-of-scope behavior %q", forbidden)
		}
	}
}

func TestScriptUsesExactMutationSequence(t *testing.T) {
	normalized := strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "").Replace(string(studioJS))
	steps := []string{
		`requestJSON("/api/v1/workspaces","POST",201,`,
		"requestJSON(`/api/v1/workspaces/${workspaceID}/configurations`,\"POST\",201,",
		"requestJSON(`/api/v1/workspaces/${workspaceID}/configurations/${configurationID}/versions`,\"POST\",201,",
		"requestJSON(`/api/v1/workspaces/${workspaceID}/configurations/${configurationID}/versions/${versionID}/listener`,\"PUT\",200,",
		"requestJSON(`/api/v1/workspaces/${workspaceID}/configurations/${configurationID}/versions/${versionID}/publish`,\"POST\",200,",
	}

	previous := -1
	for _, step := range steps {
		index := strings.Index(normalized, step)
		if index < 0 {
			t.Fatalf("script is missing exact mutation step %q", step)
		}
		if index <= previous {
			t.Fatalf("mutation step %q is out of order", step)
		}
		previous = index
	}
}

func newRouter() http.Handler {
	router := chi.NewRouter()
	RegisterRoutes(router)
	return router
}
