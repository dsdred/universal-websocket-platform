package configurationstudio

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"regexp"
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

func TestBrandingRoutesDeliverExactAssetsAndSecurityHeaders(t *testing.T) {
	const expectedCSP = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'"
	for _, asset := range []struct {
		path string
		mime string
		body []byte
	}{
		{"/favicon.ico?v=128", "image/vnd.microsoft.icon", faviconICO},
		{"/configuration-studio/branding/viking-128.png", "image/png", vikingPNG},
		{"/configuration-studio/branding/uwp-viking.ico", "image/vnd.microsoft.icon", windowsICO},
		{"/configuration-studio/branding/uwp-logo.png", "image/png", logoPNG},
	} {
		t.Run(asset.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			newRouter().ServeHTTP(response, httptest.NewRequest(http.MethodGet, asset.path, nil))
			if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), asset.body) {
				t.Fatal("branding route did not deliver exact embedded bytes with status 200")
			}
			for key, want := range map[string]string{
				"Content-Type": asset.mime, "Cache-Control": "no-cache",
				"X-Content-Type-Options": "nosniff", "Content-Security-Policy": expectedCSP,
			} {
				if got := response.Header().Get(key); got != want {
					t.Errorf("%s = %q, want %q", key, got, want)
				}
			}
			for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
				response = httptest.NewRecorder()
				newRouter().ServeHTTP(response, httptest.NewRequest(method, asset.path, nil))
				if response.Code != http.StatusMethodNotAllowed {
					t.Errorf("%s status = %d, want 405", method, response.Code)
				}
			}
		})
	}
	for _, path := range []string{"/favicon.ico/", "/configuration-studio/branding/", "/configuration-studio/branding/missing.png", "/configuration-studio/branding/README.md", "/configuration-studio/branding/favicon.ico"} {
		response := httptest.NewRecorder()
		newRouter().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNotFound {
			t.Errorf("GET %s status = %d, want 404", path, response.Code)
		}
	}
}

func TestBothPagesReferenceVersionedFaviconAndDecorativePortrait(t *testing.T) {
	for _, page := range []struct {
		name string
		body []byte
	}{{"Studio", indexHTML}, {"Dashboard", dashboardHTML}} {
		t.Run(page.name, func(t *testing.T) {
			html := string(page.body)
			icons := regexp.MustCompile(`<link\b[^>]*\brel="icon"[^>]*>`).FindAllString(html, -1)
			if len(icons) != 1 {
				t.Fatalf("favicon links = %d, want one", len(icons))
			}
			for _, required := range []string{`href="/favicon.ico?v=128"`, `type="image/vnd.microsoft.icon"`, `sizes="16x16 32x32 48x48"`} {
				if !strings.Contains(icons[0], required) {
					t.Errorf("favicon link missing %s", required)
				}
			}
			portraits := regexp.MustCompile(`<img\b[^>]*\bsrc="/configuration-studio/branding/viking-128.png"[^>]*>`).FindAllString(html, -1)
			if len(portraits) != 1 {
				t.Fatalf("portrait images = %d, want one", len(portraits))
			}
			for _, required := range []string{`alt=""`, `width="40"`, `height="40"`} {
				if !strings.Contains(portraits[0], required) {
					t.Errorf("decorative portrait missing %s", required)
				}
			}
			if strings.Contains(html, `class="brand-mark">U</span>`) {
				t.Error("temporary U brand mark remains")
			}
		})
	}
}

func TestBrandingPNGDimensionsAndPortraitTransparency(t *testing.T) {
	portrait := decodeBrandPNG(t, vikingPNG, 128)
	assertTransparentInset(t, portrait, 4)
	// Pin the independently admitted face asset, preventing accidental helmet substitution.
	if got := fmt.Sprintf("%x", sha256.Sum256(vikingPNG)); got != "61a14ed986a4260ec214d5912a93d1a5deadcc4e9b47d0ab6d413bd43d82799c" {
		t.Errorf("sidebar asset differs from approved standalone portrait: %s", got)
	}
	logo := decodeBrandPNG(t, logoPNG, 1254)
	if got := fmt.Sprintf("%x", sha256.Sum256(logoPNG)); got != "883163ba0332fb3f992450ce17fe78399c2dff6904738254306535cdf8128743" {
		t.Errorf("full branding is not byte-exact approved archive logo: %s", got)
	}
	if _, _, _, alpha := logo.At(0, 0).RGBA(); alpha != 65535 {
		t.Error("archive full logo should preserve its documented opaque background")
	}
}

func TestICOFramesSeparateHelmetFaviconsFromWindowsFace(t *testing.T) {
	faviconFrames := decodeBrandICO(t, faviconICO, []int{16, 32, 48})
	approvedHelmetHashes := []string{
		"eb74111fe6a752e55241fac6b117c8f397ed2faaccbda1a5e7f81edaeab9b064",
		"002b1973a4c5118c7ae06f490dd58857a86c0db1c0a990929394844c82609b34",
		"777d87a28986e9ae44adc4b28abbdaeb7ce5d0c3fc98099a9f7a1ff4b907425e",
	}
	for i, frame := range faviconFrames {
		if got := fmt.Sprintf("%x", sha256.Sum256(frame)); got != approvedHelmetHashes[i] {
			t.Errorf("favicon frame %d differs from actual approved archive helmet PNG: %s", i, got)
		}
	}
	face := decodeBrandICO(t, windowsICO, []int{256})[0]
	assertTransparentInset(t, decodeBrandPNG(t, face, 256), 8)
	if got := fmt.Sprintf("%x", sha256.Sum256(face)); got != "1ce27684c53a1356bce42bfdd956e79eedb36fe0b677748c0319d3d114004b0c" {
		t.Errorf("Windows frame is not the admitted standalone face rendering: %s", got)
	}
}

func decodeBrandPNG(t *testing.T, data []byte, size int) image.Image {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("invalid PNG: %v", err)
	}
	if img.Bounds() != image.Rect(0, 0, size, size) {
		t.Fatalf("PNG bounds = %v, want %dx%d", img.Bounds(), size, size)
	}
	return img
}

func assertTransparentInset(t *testing.T, img image.Image, inset int) {
	t.Helper()
	bounds := img.Bounds()
	var transparent, visible int
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			_, _, _, alpha := img.At(x, y).RGBA()
			if alpha == 0 {
				transparent++
			} else {
				visible++
			}
			if (x < inset || y < inset || x >= bounds.Dx()-inset || y >= bounds.Dy()-inset) && alpha != 0 {
				t.Fatalf("portrait inset is not transparent at (%d,%d)", x, y)
			}
		}
	}
	if transparent == 0 || visible == 0 {
		t.Fatal("portrait must contain both actual visible artwork and transparent background")
	}
}

func decodeBrandICO(t *testing.T, data []byte, sizes []int) [][]byte {
	t.Helper()
	if len(data) < 6 || binary.LittleEndian.Uint16(data[0:2]) != 0 || binary.LittleEndian.Uint16(data[2:4]) != 1 || int(binary.LittleEndian.Uint16(data[4:6])) != len(sizes) {
		t.Fatal("invalid ICO header or unexpected frame count")
	}
	directoryEnd := 6 + 16*len(sizes)
	if len(data) < directoryEnd {
		t.Fatal("ICO directory is truncated")
	}
	previousEnd := uint64(directoryEnd)
	frames := make([][]byte, len(sizes))
	for i, size := range sizes {
		entry := data[6+16*i : 6+16*(i+1)]
		width, height := int(entry[0]), int(entry[1])
		if width == 0 {
			width = 256
		}
		if height == 0 {
			height = 256
		}
		if width != size || height != size {
			t.Fatalf("frame %d dimensions = %dx%d, want %dx%d", i, width, height, size, size)
		}
		length := uint64(binary.LittleEndian.Uint32(entry[8:12]))
		offset := uint64(binary.LittleEndian.Uint32(entry[12:16]))
		end := offset + length
		if length == 0 || offset != previousEnd || end > uint64(len(data)) {
			t.Fatalf("frame %d has invalid/overlapping/out-of-bounds payload", i)
		}
		frames[i] = data[offset:end]
		decodeBrandPNG(t, frames[i], size)
		previousEnd = end
	}
	if previousEnd != uint64(len(data)) {
		t.Fatal("unexpected data outside ICO payloads")
	}
	return frames
}

func newRouter() http.Handler {
	router := chi.NewRouter()
	RegisterRoutes(router)
	return router
}
