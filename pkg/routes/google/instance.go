package google

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
)

func instance(w http.ResponseWriter, r *http.Request) {
	paths := []string{
		"attributes/",
		"hostname",
		"id",
		"service-accounts/",
		"zone",
	}
	render.PlainText(w, r, strings.Join(paths, "\n"))
}

func instanceHostname(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "node0.proj.local")
}

func instanceId(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "1234567890")
}

func instanceZone(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "projects/1234567890/zones/eu-west1-d")
}

func instanceAttributes(w http.ResponseWriter, r *http.Request) {
	paths := []string{
		"cluster-location",
		"cluster-name",
		"cluster-uid",
	}
	render.PlainText(w, r, strings.Join(paths, "\n"))
}

func instanceClusterLocation(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "europe-west1")
}

func instanceClusterName(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "proj")
}

func instanceClusterUid(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "1234567890")
}
