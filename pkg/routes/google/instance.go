package google

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/magnm/lcm/config"
)

func instance(w http.ResponseWriter, r *http.Request) {
	paths := []string{
		"attributes/",
		"hostname",
		"id",
		"service-accounts/",
		"zone",
	}
	writeText(w, r, strings.Join(paths, "\n"))
}

func instanceHostname(w http.ResponseWriter, r *http.Request) {
	writeText(w, r, fmt.Sprintf("node0.c.%s.internal", config.Current.ProjectId))
}

func instanceId(w http.ResponseWriter, r *http.Request) {
	writeText(w, r, "1234567890")
}

func instanceZone(w http.ResponseWriter, r *http.Request) {
	writeText(w, r, "projects/1234567890/zones/eu-west1-d")
}

func instanceAttributes(w http.ResponseWriter, r *http.Request) {
	paths := []string{
		"cluster-location",
		"cluster-name",
		"cluster-uid",
	}
	writeText(w, r, strings.Join(paths, "\n"))
}

func instanceClusterLocation(w http.ResponseWriter, r *http.Request) {
	writeText(w, r, "europe-west1")
}

func instanceClusterName(w http.ResponseWriter, r *http.Request) {
	writeText(w, r, "dev-cluster")
}

func instanceClusterUid(w http.ResponseWriter, r *http.Request) {
	writeText(w, r, "1234567890")
}
