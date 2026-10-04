// Package configurationstudio serves the built-in local Configuration Studio.
package configurationstudio

import (
	_ "embed"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const contentSecurityPolicy = "default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'"

var (
	//go:embed assets/index.html
	indexHTML []byte

	//go:embed assets/studio.css
	studioCSS []byte

	//go:embed assets/studio.js
	studioJS []byte
)

// RegisterRoutes registers the exact routes for the embedded local/dev UI.
func RegisterRoutes(router chi.Router) {
	router.Get("/", serveAsset("text/html; charset=utf-8", indexHTML))
	router.Get("/configuration-studio/studio.css", serveAsset("text/css; charset=utf-8", studioCSS))
	router.Get("/configuration-studio/studio.js", serveAsset("text/javascript; charset=utf-8", studioJS))
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
