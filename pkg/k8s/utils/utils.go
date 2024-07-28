package utils

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var serviceAccountNamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

const namespaceFleetlock = "fleetlock"

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
