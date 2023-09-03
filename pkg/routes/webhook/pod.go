package webhook

import (
	"github.com/distribution/reference"
	"github.com/magnm/lcm/config"
	"github.com/magnm/lcm/pkg/kubernetes"
	kubegoogle "github.com/magnm/lcm/pkg/kubernetes/google"
	"golang.org/x/exp/slog"
	corev1 "k8s.io/api/core/v1"
)

func patchesForPod(pod *corev1.Pod, dryRun bool) ([]kubernetes.PatchOperation, error) {
	var patches []kubernetes.PatchOperation

	// Check if we should add imagePullSecret
	for _, container := range pod.Spec.Containers {
		if container.ImagePullPolicy == corev1.PullNever {
			continue
		}

		image, err := reference.ParseNormalizedNamed(container.Image)
		if err != nil {
			slog.Error("failed to parse image tag", "tag", container.Image, "err", err)
			return nil, err
		}

		switch config.Current.Type {
		case config.GoogleMetadata:
			if kubegoogle.ShouldAddImagePullSecret(image) {
				secretReference, err := kubegoogle.PullSecretForImage(image, pod.Namespace, dryRun)
				if err != nil {
					slog.Error("failed to create image pull secret", "err", err)
					return nil, err
				}

				if len(pod.Spec.ImagePullSecrets) == 0 {
					patches = append(patches, kubernetes.PatchOperation{
						Op:    "add",
						Path:  "/spec/imagePullSecrets",
						Value: []string{secretReference.Name},
					})
				} else {
					patches = append(patches, kubernetes.PatchOperation{
						Op:    "add",
						Path:  "/spec/imagePullSecrets/-",
						Value: secretReference.Name,
					})
				}
			}

		}
	}

	// Add a DNS entry for the metadata server to the pod
	if config.Current.Type == config.GoogleMetadata {
		if len(pod.Spec.HostAliases) == 0 {
			patches = append(patches, kubernetes.PatchOperation{
				Op:   "add",
				Path: "/spec/hostAliases",
				Value: []corev1.HostAlias{
					{
						IP:        kubernetes.GetOurServiceIp(),
						Hostnames: []string{kubegoogle.MetadataServerDomain},
					},
				},
			})
		} else {
			patches = append(patches, kubernetes.PatchOperation{
				Op:   "add",
				Path: "/spec/hostAliases/-",
				Value: corev1.HostAlias{
					IP:        kubernetes.GetOurServiceIp(),
					Hostnames: []string{kubegoogle.MetadataServerDomain},
				},
			})
		}
	}

	return patches, nil
}
