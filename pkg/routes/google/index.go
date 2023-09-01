package google

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/samber/lo"
)

func Routes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(verifyRequestHeaders)
	r.Get("/", redirectTo("/computeMetadata/v1/"))
	r.Get("/computeMetadata", redirectTo("/computeMetadata/v1/"))
	r.Get("/computeMetadata/v1", redirectTo("/computeMetadata/v1/"))
	r.Route("/computeMetadata/v1/", computeMetadataRoutes)
	return r
}

func redirectTo(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		patterns := lo.Map(chi.RouteContext(r.Context()).RoutePatterns, func(pattern string, i int) string {
			return strings.ReplaceAll(pattern, "/*", "")
		})
		initialPath := strings.Join(lo.Slice(patterns, 0, len(patterns)-1), "/")
		http.Redirect(w, r, initialPath+path, http.StatusPermanentRedirect)
	}
}

func redirectToKey(path string, keys []string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		resolvedPath := path
		for _, key := range keys {
			value := chi.URLParam(r, key)
			resolvedPath = strings.ReplaceAll(resolvedPath, fmt.Sprintf("{%s}", key), value)
		}
		patterns := lo.Map(chi.RouteContext(r.Context()).RoutePatterns, func(pattern string, i int) string {
			return strings.ReplaceAll(pattern, "/*", "")
		})
		initialPath := strings.Join(lo.Slice(patterns, 0, len(patterns)-1), "/")
		http.Redirect(w, r, initialPath+resolvedPath, http.StatusPermanentRedirect)
	}
}

func verifyRequestHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Metadata-Flavor
		header := r.Header.Get("Metadata-Flavor")
		if header != "Google" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		// Check no X-Forwarded-For
		header = r.Header.Get("X-Forwarded-For")
		if header != "" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		w.Header().Set("Metadata-Flavor", "Google")
		w.Header().Set("Server", "GKE Metadata Server")

		next.ServeHTTP(w, r)
	})
}

func writeText(w http.ResponseWriter, r *http.Request, text string) {
	w.Header().Set("Content-Type", "application/text")
	if status, ok := r.Context().Value(render.StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}
	w.Write([]byte(text)) //nolint:errcheck
}
