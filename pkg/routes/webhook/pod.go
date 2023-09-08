package webhook

import (
	"fmt"

	"github.com/distribution/reference"
	"github.com/magnm/lcm/config"
	"github.com/magnm/lcm/pkg/kubernetes"
	kubegoogle "github.com/magnm/lcm/pkg/kubernetes/google"
	"github.com/samber/lo"
	"golang.org/x/exp/slog"
	corev1 "k8s.io/api/core/v1"
)

func patchesForPod(pod *corev1.Pod, dryRun bool) ([]kubernetes.PatchOperation, error) {
	var patches []kubernetes.PatchOperation

	// Check if we should add imagePullSecret
	for i, container := range pod.Spec.Containers {
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
						Op:   "add",
						Path: "/spec/imagePullSecrets",
						Value: []corev1.LocalObjectReference{
							{Name: secretReference.Name},
						},
					})
				} else {
					if !lo.ContainsBy(pod.Spec.ImagePullSecrets, func(secret corev1.LocalObjectReference) bool {
						return secret.Name == secretReference.Name
					}) {
						patches = append(patches, kubernetes.PatchOperation{
							Op:   "add",
							Path: "/spec/imagePullSecrets/-",
							Value: []corev1.LocalObjectReference{
								{Name: secretReference.Name},
							},
						})
					}
				}
			}

			// Add GCE_METADATA_IP/HOST env var to the pod, which libraries
			// will use to detect that they are running on GCP
			if len(pod.Spec.Containers[i].Env) == 0 {
				patches = append(patches, kubernetes.PatchOperation{
					Op:   "add",
					Path: fmt.Sprintf("/spec/containers/%d/env", i),
					Value: []corev1.EnvVar{
						{
							Name:  "GCE_METADATA_IP",
							Value: kubernetes.GetOurServiceIp(),
						},
						{
							Name:  "GCE_METADATA_HOST",
							Value: "metadata.google.internal",
						},
					},
				})
			} else {
				if !lo.ContainsBy(pod.Spec.Containers[i].Env, func(env corev1.EnvVar) bool {
					return env.Name == "GCE_METADATA_IP"
				}) {
					patches = append(patches, kubernetes.PatchOperation{
						Op:   "add",
						Path: fmt.Sprintf("/spec/containers/%d/env/-", i),
						Value: corev1.EnvVar{
							Name:  "GCE_METADATA_IP",
							Value: kubernetes.GetOurServiceIp(),
						},
					})
					patches = append(patches, kubernetes.PatchOperation{
						Op:   "add",
						Path: fmt.Sprintf("/spec/containers/%d/env/-", i),
						Value: corev1.EnvVar{
							Name:  "GCE_METADATA_HOST",
							Value: "metadata.google.internal",
						},
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
			if !lo.ContainsBy(pod.Spec.HostAliases, func(alias corev1.HostAlias) bool {
				return alias.IP == kubernetes.GetOurServiceIp()
			}) {
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
	}

	return patches, nil
}
