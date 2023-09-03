package google

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/distribution/reference"
	"github.com/magnm/lcm/config"
	googleclient "github.com/magnm/lcm/pkg/cloud/client/google"
	"github.com/magnm/lcm/pkg/kubernetes"
	"golang.org/x/exp/slog"
	corev1 "k8s.io/api/core/v1"
)

var GCPServiceAccountAnnotation = "iam.gke.io/gcp-service-account"
var MetadataServerDomain = "metadata.google.internal"

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

func ShouldAddImagePullSecret(image reference.Named) bool {
	return strings.Contains(image.Name(), "gcr.io/") ||
		strings.Contains(image.Name(), "docker.pkg.dev/")
}

func PullSecretForImage(image reference.Named, namespace string, dryRun bool) (*corev1.LocalObjectReference, error) {
	domain := reference.Domain(image)
	secretName := fmt.Sprintf("registry-%s", strings.ReplaceAll(domain, ".", "-"))

	registryAuth := constructRegistryAuth(domain)
	if !dryRun {
		err := kubernetes.CreateImagePullSecret(secretName, namespace, registryAuth)
		if err != nil {
			return nil, err
		}
	}

	return &corev1.LocalObjectReference{
		Name: secretName,
	}, nil
}

func constructRegistryAuth(domain string) kubernetes.RegistryAuth {
	username := "oauth2accesstoken"
	token := googleclient.GetMainAccountAccessToken()
	encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, token)))

	auth := kubernetes.RegistryAuth{
		Auths: map[string]kubernetes.RegistryAuthEntry{
			domain: {
				Username: username,
				Password: token,
				Email:    "",
				Auth:     encoded,
			},
		},
	}
	return auth
}
