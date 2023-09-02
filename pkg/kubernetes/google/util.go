package google

import (
	"fmt"

	"github.com/magnm/lcm/config"
	googleclient "github.com/magnm/lcm/pkg/cloud/client/google"
	"golang.org/x/exp/slog"
	corev1 "k8s.io/api/core/v1"
)

var GCPServiceAccountAnnotation = "iam.gke.io/gcp-service-account"

func GetGsaForKsa(ksa *corev1.ServiceAccount) string {
	gcpServiceAccount, ok := ksa.GetAnnotations()[GCPServiceAccountAnnotation]

	if !ok {
		slog.Error("no gsa annotation found on ksa", "ksa", ksa.Name)
		return ""
	}

	// Validate that the ksa is bound to the gsa
	// Expected member binding is "serviceAccount:project-id.svc.id.goog[ksa-namespace/ksa-name]"
	ksaBinding := fmt.Sprintf(
		"serviceAccount:%s.svc.id.goog[%s/%s]",
		config.Current.ProjectId,
		ksa.Namespace,
		ksa.Name,
	)
	if !googleclient.ValidateKsaGsaBinding(ksaBinding, gcpServiceAccount) {
		slog.Error("ksa is not bound to the gsa", "ksa", ksa.Name, "gsa", gcpServiceAccount)
		return ""
	}

	return gcpServiceAccount
}
