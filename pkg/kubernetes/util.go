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
)

func CallingPod(r *http.Request) (*corev1.Pod, error) {
	ip := util.RequestIp(r)
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
	return &pod, nil
}

func ServiceAccountForPod(pod *corev1.Pod) string {
	return pod.Spec.ServiceAccountName
}
