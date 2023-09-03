package google

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/magnm/lcm/pkg/routes/util"
	"golang.org/x/exp/slog"
)

func Routes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(verifyRequestHeaders)
	r.Get("/", index)
	r.Get("/computeMetadata", util.RedirectTo("/computeMetadata/v1/"))
	r.Get("/computeMetadata/v1", util.RedirectTo("/computeMetadata/v1/"))
	r.Route("/computeMetadata/v1/", computeMetadataRoutes)
	return r
}

func index(w http.ResponseWriter, r *http.Request) {
	writeText(w, r, "computeMetadata/")
}

func verifyRequestHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("google metadata request", "path", r.URL.Path)
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
