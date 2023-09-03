package google

import (
	"net/http"
	"strings"

	"github.com/magnm/lcm/config"
	googleclient "github.com/magnm/lcm/pkg/cloud/client/google"
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
	project := googleclient.GetProject(config.Current.ProjectId)
	if project == nil {
		http.Error(w, "failed to get project", http.StatusInternalServerError)
		return
	}

	numericId := strings.TrimPrefix(project.Name, "projects/")

	writeText(w, r, numericId)
}
