package google

import (
	"net/http"
	"strings"

	"github.com/magnm/lcm/config"
)

func project(w http.ResponseWriter, r *http.Request) {
	paths := []string{
		"project-id",
		"numeric-project-id",
	}
	writeText(w, r, strings.Join(paths, "\n"))
}

func projectId(w http.ResponseWriter, r *http.Request) {
	writeText(w, r, config.Current.ProjectId)
}

func projectNumericId(w http.ResponseWriter, r *http.Request) {
	writeText(w, r, "1234567890")
}
