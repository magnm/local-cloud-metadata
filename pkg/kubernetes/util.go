package kubernetes

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	kubeclient "github.com/magnm/lcm/pkg/kubernetes/client"
	"github.com/magnm/lcm/pkg/util"
	"golang.org/x/exp/slog"
	corev1 "k8s.io/api/core/v1"
	errorv1 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type RegistryAuth struct {
	Auths map[string]RegistryAuthEntry `json:"auths"`
}

type RegistryAuthEntry struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Auth     string `json:"auth"`
}

var podCache = map[string]*corev1.Pod{}

func CallingPod(r *http.Request) (*corev1.Pod, error) {
	ip := util.RequestIp(r)

	// Check cache first before looking up in kube api
	if pod, ok := podCache[ip]; ok {
		return pod, nil
	}

	client, err := kubeclient.GetKubernetesClient()
	if err != nil {
		return nil, err
	}

	podList, err := client.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("status.podIP=%s,status.phase!=Failed,status.phase!=Succeeded", ip),
	})
	if err != nil {
		return nil, err
	}
	if len(podList.Items) == 0 {
		slog.Error("no pod found for ip", "ip", ip)
		return nil, errorv1.NewNotFound(corev1.Resource("Pod"), ip)
	}
	pod := podList.Items[0]

	podCache[ip] = &pod

	return &pod, nil
}

func ServiceAccountForPod(pod *corev1.Pod) (*corev1.ServiceAccount, error) {
	name := pod.Spec.ServiceAccountName
	if name == "" {
		name = "default"
	}

	client, err := kubeclient.GetKubernetesClient()
	if err != nil {
		return nil, err
	}

	return client.CoreV1().ServiceAccounts(pod.Namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func FindCustomResource[T any](group string, version string, resource string, namespace string) ([]T, error) {
	client, err := kubeclient.GetKubernetesDynamicClient()
	if err != nil {
		return nil, err
	}

	resourceType := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	list, err := client.Resource(resourceType).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	converted := []T{}
	for _, item := range list.Items {
		var binding T
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &binding)
		if err != nil {
			return nil, err
		}
		converted = append(converted, binding)
	}
	return converted, nil
}

func CreateSecret(secret *corev1.Secret) error {
	client, err := kubeclient.GetKubernetesClient()
	if err != nil {
		return err
	}

	_, err = client.CoreV1().Secrets(secret.Namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	return err
}

func CreateImagePullSecret(name string, namespace string, registryAuth RegistryAuth) error {
	jsonEncoded, err := json.Marshal(registryAuth)
	if err != nil {
		return err
	}
	base64Encoded := base64.StdEncoding.EncodeToString(jsonEncoded)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			".dockerconfigjson": []byte(base64Encoded),
		},
	}

	return CreateSecret(secret)
}
