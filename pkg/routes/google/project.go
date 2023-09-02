package google

import (
	"net/http"
	"strings"

	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"github.com/magnm/lcm/config"
	googleclient "github.com/magnm/lcm/pkg/cloud/client/google"
)

var cachedProject *resourcemanagerpb.Project

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
	if cachedProject == nil {
		cachedProject = googleclient.GetProject(config.Current.ProjectId)
	}
	if cachedProject == nil {
		http.Error(w, "failed to get project", http.StatusInternalServerError)
		return
	}

	numericId := strings.TrimPrefix(cachedProject.Name, "projects/")

	writeText(w, r, numericId)
}
