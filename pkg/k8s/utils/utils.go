package utils

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var serviceAccountNamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

const namespaceFleetlock = "fleetlock"

// Read the namespace from the inserted serviceaccount file. Fallback to default if the file does not exist.
func GetNamespace() (string, error) {
	data, err := os.ReadFile(serviceAccountNamespaceFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", NewErrorGetNamespace(serviceAccountNamespaceFile, err)
		} else {
			return namespaceFleetlock, nil
		}
	}

	ns := strings.TrimSpace(string(data))
	if len(ns) == 0 {
		return "", NewErrorGetNamespace(serviceAccountNamespaceFile, fmt.Errorf("file was empty"))
	}
	return ns, nil
}

// Create a new kubernetes clientset from the provided kubeconfig. Default to in-cluster if none is provided.
func CreateNewClientset(kubeconfig string) (kubernetes.Interface, error) {
	var config *rest.Config
	var err error
	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

// Return a pointer to the variable value
func Pointer[T any](v T) *T {
	return &v
}
