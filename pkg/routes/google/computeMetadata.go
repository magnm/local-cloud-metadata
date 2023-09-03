package google

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/magnm/lcm/pkg/routes/util"
)

func computeMetadataRoutes(r chi.Router) {
	r.Get("/", metadataIndex)
	r.Get("/project", util.RedirectTo("/project/"))
	r.Get("/project/", project)
	r.Get("/project/project-id", projectId)
	r.Get("/project/numeric-project-id", projectNumericId)

	r.Get("/instance", util.RedirectTo("/instance/"))
	r.Get("/instance/", instance)
	r.Get("/instance/hostname", instanceHostname)
	r.Get("/instance/id", instanceId)
	r.Get("/instance/zone", instanceZone)
	r.Get("/instance/attributes", util.RedirectTo("/instance/attributes/"))
	r.Get("/instance/attributes/", instanceAttributes)
	r.Get("/instance/attributes/cluster-location", instanceClusterLocation)
	r.Get("/instance/attributes/cluster-name", instanceClusterName)
	r.Get("/instance/attributes/cluster-uid", instanceClusterUid)
	r.Get("/instance/service-accounts", util.RedirectTo("/instance/service-accounts/"))
	r.Get("/instance/service-accounts/", serviceAccounts)
	r.Get("/instance/service-accounts/{acc}", util.RedirectToKey("/instance/service-accounts/{acc}/", []string{"acc"}))
	r.Get("/instance/service-accounts/{acc}/", serviceAccount)
	r.Get("/instance/service-accounts/{acc}/{key}", serviceAccountAttr)
}

func metadataIndex(w http.ResponseWriter, r *http.Request) {
	paths := []string{
		"project/",
		"instance/",
	}
	writeText(w, r, strings.Join(paths, "\n"))
}
