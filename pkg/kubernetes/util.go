package kubernetes

import (
	"context"
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

func ServiceAccountForPod(pod *corev1.Pod) string {
	return pod.Spec.ServiceAccountName
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
