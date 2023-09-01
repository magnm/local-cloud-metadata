package google

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/magnm/lcm/config"
)

func project(w http.ResponseWriter, r *http.Request) {
	paths := []string{
		"project-id",
		"numeric-project-id",
	}
	render.PlainText(w, r, strings.Join(paths, "\n"))
}

func projectId(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, config.Current.ProjectId)
}

func projectNumericId(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "1234567890")
}
