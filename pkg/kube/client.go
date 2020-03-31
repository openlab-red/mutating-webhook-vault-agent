package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client creates Kubernetes Client with Inner Cluster Config
func Client() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}
