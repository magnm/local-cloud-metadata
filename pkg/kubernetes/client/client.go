package client

import (
	"os"
	"path/filepath"

	"golang.org/x/exp/slog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var clientset *kubernetes.Clientset

func GetKubernetesClient() (*kubernetes.Clientset, error) {
	if clientset != nil {
		return clientset, nil
	}

	var config *rest.Config
	var err error
	// Check if we are running in-cluster
	if os.Getenv("KUBERNETES_SERVICE_HOST") == "" {
		home := homedir.HomeDir()
		kubeconfig := filepath.Join(home, ".kube", "config")

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			slog.Error("couldn't find kubeconfig", "err", err)
			return nil, err
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	clientset, err = kubernetes.NewForConfig(config)
	return clientset, err
}
