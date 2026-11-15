// Package configurationstudio serves the built-in local Configuration Studio.
package configurationstudio

import (
	_ "embed"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const contentSecurityPolicy = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'"

var (
	//go:embed assets/index.html
	indexHTML []byte

	//go:embed assets/studio.css
	studioCSS []byte

	//go:embed assets/studio.js
	studioJS []byte

	//go:embed assets/dashboard.html
	dashboardHTML []byte

	//go:embed assets/control.css
	controlCSS []byte

	//go:embed assets/dashboard.js
	dashboardJS []byte

	//go:embed assets/branding/viking-128.png
	vikingPNG []byte

	//go:embed assets/branding/favicon.ico
	faviconICO []byte

	//go:embed assets/branding/uwp-viking.ico
	windowsICO []byte

	//go:embed assets/branding/uwp-logo.png
	logoPNG []byte
)

// RegisterRoutes registers the exact routes for the embedded local/dev UI.
func RegisterRoutes(router chi.Router) {
	router.Get("/", serveAsset("text/html; charset=utf-8", indexHTML))
	router.Get("/dashboard", serveAsset("text/html; charset=utf-8", dashboardHTML))
	router.Get("/configuration-studio/control.css", serveAsset("text/css; charset=utf-8", controlCSS))
	router.Get("/configuration-studio/dashboard.js", serveAsset("text/javascript; charset=utf-8", dashboardJS))
	router.Get("/configuration-studio/studio.css", serveAsset("text/css; charset=utf-8", studioCSS))
	router.Get("/configuration-studio/studio.js", serveAsset("text/javascript; charset=utf-8", studioJS))
	router.Get("/favicon.ico", serveBrandAsset("image/vnd.microsoft.icon", faviconICO))
	router.Get("/configuration-studio/branding/viking-128.png", serveBrandAsset("image/png", vikingPNG))
	router.Get("/configuration-studio/branding/uwp-viking.ico", serveBrandAsset("image/vnd.microsoft.icon", windowsICO))
	router.Get("/configuration-studio/branding/uwp-logo.png", serveBrandAsset("image/png", logoPNG))
}

func serveBrandAsset(contentType string, content []byte) http.HandlerFunc {
	asset := serveAsset(contentType, content)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		asset(w, r)
	}
}

func serveAsset(contentType string, content []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Security-Policy", contentSecurityPolicy)
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}
}
