package webhook

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/magnm/lcm/pkg/kubernetes"
	"golang.org/x/exp/slog"
)

func Routes() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/mutate", handleRequest)
	return r
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	slog.Debug("admission request", "length", r.ContentLength)
	review, pod, err := kubernetes.DecodePodMutationRequest(r)
	if err != nil {
		slog.Error("failed to decode pod mutation request", "err", err)
		http.Error(w, "failed to decode pod mutation request", http.StatusBadRequest)
		return
	}
	slog.Debug("admission review", "version", review.APIVersion)

	patches, err := patchesForPod(pod, *review.Request.DryRun)
	if err != nil {
		slog.Error("failed to generate patches for pod", "err", err)
		http.Error(w, "failed to generate patches for pod", http.StatusInternalServerError)
		return
	}

	response, err := kubernetes.EncodeMutationPatches(review, patches)
	if err != nil {
		slog.Error("failed to encode mutation patches", "err", err)
		http.Error(w, "failed to encode mutation patches", http.StatusInternalServerError)
		return
	}
	slog.Debug("admission response", "patches", patches)

	render.JSON(w, r, response)
}
