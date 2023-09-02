package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/magnm/lcm/pkg/kubernetes"
	"golang.org/x/exp/slog"
)

func Routes() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/mutate", handleRequest)
	return r
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	podRequest, dryRun, err := kubernetes.DecodePodMutationRequest(r)
	if err != nil {
		slog.Error("failed to decode pod mutation request", "err", err)
		http.Error(w, "failed to decode pod mutation request", http.StatusBadRequest)
		return
	}

	patches, err := patchesForPod(podRequest, dryRun)
	if err != nil {
		slog.Error("failed to generate patches for pod", "err", err)
		http.Error(w, "failed to generate patches for pod", http.StatusInternalServerError)
		return
	}

	response, err := kubernetes.EncodeMutationPatches(patches)
	if err != nil {
		slog.Error("failed to encode mutation patches", "err", err)
		http.Error(w, "failed to encode mutation patches", http.StatusInternalServerError)
		return
	}
	slog.Debug("admission response", "response", response)

	bytes, err := json.Marshal(response)
	if err != nil {
		slog.Error("failed to marshal admission response", "err", err)
		http.Error(w, "failed to marshal admission response", http.StatusInternalServerError)
		return
	}

	render.Data(w, r, bytes)
}
