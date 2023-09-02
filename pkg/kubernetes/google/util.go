package google

import (
	"fmt"

	"github.com/magnm/lcm/config"
	"github.com/magnm/lcm/pkg/kubernetes"
	"golang.org/x/exp/slog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IAMServiceAccount struct {
	Metadata metav1.ObjectMeta     `json:"metadata"`
	Spec     IAMServiceAccountSpec `json:"spec"`
}

type IAMServiceAccountSpec struct {
	ResourceID string `json:"resourceID"`
}

type IAMPartialPolicy struct {
	Spec IAMPartialPolicySpec `json:"spec"`
}

type IAMPartialPolicySpec struct {
	Bindings    []IAMPartialPolicyBinding `json:"bindings"`
	ResourceRef corev1.ObjectReference    `json:"resourceRef"`
}

type IAMPartialPolicyBinding struct {
	Role    string                   `json:"role"`
	Members []IAMPartialPolicyMember `json:"members"`
}

type IAMPartialPolicyMember struct {
	Member string `json:"member"`
}

func GetGsaForKsa(ksa string, namespace string) string {
	crs, err := findIAMPartialPolicies(namespace)
	if err != nil {
		slog.Error("failed to find IAMPartialPolicies", "err", err)
		return ""
	}

	// Expected member binding is "serviceAccount:project-id.svc.id.goog[namespace/ksa-name]"
	expected := fmt.Sprintf(
		"serviceAccount:%s.svc.id.goog[%s/%s]",
		config.Current.ProjectId,
		namespace,
		ksa,
	)

	for _, policy := range crs {
		for _, binding := range policy.Spec.Bindings {
			for _, member := range binding.Members {
				if member.Member == expected && binding.Role == "roles/iam.workloadIdentityUser" {
					sa, err := findIAMServiceAccount(namespace, policy.Spec.ResourceRef.Name)
					if err != nil || sa == "" {
						slog.Error("failed to find IAMServiceAccount", "err", err, "resourceRef", policy.Spec.ResourceRef.Name)
						return ""
					}
					return sa
				}
			}
		}
	}

	// If we didn't find anything in the given namespace, try again with no namespace (i.e. cluster-wide)

	crs, err = findIAMPartialPolicies("")
	if err != nil {
		slog.Error("failed to find IAMPartialPolicies", "err", err)
		return ""
	}

	for _, policy := range crs {
		for _, binding := range policy.Spec.Bindings {
			for _, member := range binding.Members {
				if member.Member == expected && binding.Role == "roles/iam.workloadIdentityUser" {
					sa, err := findIAMServiceAccount("", policy.Spec.ResourceRef.Name)
					if err != nil || sa == "" {
						slog.Error("failed to find IAMServiceAccount", "err", err, "resourceRef", policy.Spec.ResourceRef.Name)
						return ""
					}
					return sa
				}
			}
		}
	}

	return ""
}

func findIAMPartialPolicies(namespace string) ([]IAMPartialPolicy, error) {
	return kubernetes.FindCustomResource[IAMPartialPolicy](
		"iam.cnrm.cloud.google.com",
		"v1beta1",
		"iampartialpolicy",
		namespace,
	)
}

func findIAMServiceAccount(namespace string, resourceName string) (string, error) {
	cr, err := kubernetes.FindCustomResource[IAMServiceAccount](
		"iam.cnrm.cloud.google.com",
		"v1beta1",
		"iamserviceaccount",
		namespace,
	)
	if err != nil {
		return "", err
	}

	var localName string

	for _, serviceAccount := range cr {
		if serviceAccount.Metadata.Name == resourceName {
			if serviceAccount.Spec.ResourceID != "" {
				localName = serviceAccount.Spec.ResourceID
			} else {
				localName = serviceAccount.Metadata.Name
			}
		}
	}

	if localName != "" {
		return localName + "@" + config.Current.ProjectId + ".iam.gserviceaccount.com", nil
	} else {
		return "", nil
	}
}
