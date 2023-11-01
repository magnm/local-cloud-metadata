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
	var (
		patches []kubernetes.PatchOperation
		err     error
	)

	var dnsEntries []corev1.HostAlias
	var envVars []corev1.EnvVar

	switch config.Current.Type {
	case config.GoogleMetadata:
		dnsEntries = []corev1.HostAlias{
			{IP: kubernetes.GetOurServiceIp(), Hostnames: []string{kubegoogle.MetadataServerDomain}},
		}

		envVars = []corev1.EnvVar{
			{Name: "GCE_METADATA_IP", Value: kubernetes.GetOurServiceIp()},
			{Name: "GCE_METADATA_HOST", Value: "metadata.google.internal"},
		}
	}

	// Check if we should add imagePullSecret or envVars
	for i, container := range pod.Spec.Containers {
		patches, err = patchesForContainer(patches, envVars, "containers", pod, container, i, dryRun)
		if err != nil {
			return nil, err
		}
	}
	for i, initContainer := range pod.Spec.InitContainers {
		patches, err = patchesForContainer(patches, envVars, "initContainers", pod, initContainer, i, dryRun)
		if err != nil {
			return nil, err
		}
	}

	if len(dnsEntries) > 0 {
		if len(pod.Spec.HostAliases) == 0 {
			patches = append(patches, kubernetes.PatchOperation{
				Op:    "add",
				Path:  "/spec/hostAliases",
				Value: dnsEntries,
			})
		} else {
			for _, dnsEntry := range dnsEntries {
				if !lo.ContainsBy(pod.Spec.HostAliases, func(alias corev1.HostAlias) bool {
					return alias.IP == dnsEntry.IP
				}) {
					patches = append(patches, kubernetes.PatchOperation{
						Op:    "add",
						Path:  "/spec/hostAliases/-",
						Value: dnsEntry,
					})
				}
			}
		}
	}

	return patches, nil
}

func patchesForContainer(
	patches []kubernetes.PatchOperation,
	envVars []corev1.EnvVar,
	containerTypeJsonPath string,
	pod *corev1.Pod,
	container corev1.Container,
	index int,
	dryRun bool,
) ([]kubernetes.PatchOperation, error) {
	if container.ImagePullPolicy == corev1.PullNever {
		return patches, nil
	}

	image, err := reference.ParseNormalizedNamed(container.Image)
	if err != nil {
		slog.Error("failed to parse image tag", "tag", container.Image, "err", err)
		return nil, err
	}

	var pullSecretRef *corev1.LocalObjectReference

	switch config.Current.Type {
	case config.GoogleMetadata:
		if kubegoogle.ShouldAddImagePullSecret(image) {
			pullSecretRef, err = kubegoogle.PullSecretForImage(image, pod.Namespace, dryRun)
			if err != nil {
				slog.Error("failed to create image pull secret", "err", err)
				return nil, err
			}
		}
	}

	if pullSecretRef != nil {
		if len(pod.Spec.ImagePullSecrets) == 0 {
			patches = append(patches, kubernetes.PatchOperation{
				Op:   "add",
				Path: "/spec/imagePullSecrets",
				Value: []corev1.LocalObjectReference{
					{Name: pullSecretRef.Name},
				},
			})
		} else {
			if !lo.ContainsBy(pod.Spec.ImagePullSecrets, func(secret corev1.LocalObjectReference) bool {
				return secret.Name == pullSecretRef.Name
			}) {
				patches = append(patches, kubernetes.PatchOperation{
					Op:   "add",
					Path: "/spec/imagePullSecrets/-",
					Value: []corev1.LocalObjectReference{
						{Name: pullSecretRef.Name},
					},
				})
			}
		}
	}

	if len(envVars) > 0 {
		if len(pod.Spec.Containers[index].Env) == 0 {
			patches = append(patches, kubernetes.PatchOperation{
				Op:    "add",
				Path:  fmt.Sprintf("/spec/%s/%d/env", containerTypeJsonPath, index),
				Value: envVars,
			})
		} else {
			for _, env := range envVars {
				if !lo.ContainsBy(pod.Spec.Containers[index].Env, func(e corev1.EnvVar) bool {
					return e.Name == env.Name
				}) {
					patches = append(patches, kubernetes.PatchOperation{
						Op:    "add",
						Path:  fmt.Sprintf("/spec/%s/%d/env/-", containerTypeJsonPath, index),
						Value: env,
					})
				}
			}
		}
	}

	return patches, nil
}
