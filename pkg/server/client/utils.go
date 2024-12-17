package client

import (
	"log/slog"
	"os"
	"strings"

	systemdutils "github.com/heathcliff26/fleetlock/pkg/systemd-utils"
)

// Read the machine-id from /etc/machine-id
func GetMachineID() (string, error) {
	b, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		return "", err
	}
	machineID := strings.TrimRight(string(b), "\r\n")
	return machineID, nil
}

// Find the machine-id of the current node and generate a zincati appID from it.
func GetZincateAppID() (string, error) {
	machineID, err := GetMachineID()
	if err != nil {
		return "", err
	}

	appID, err := systemdutils.ZincatiMachineID(machineID)
	if err != nil {
		return "", err
	}
	return appID, nil
}

// When having // in a URL, it somehow converts the request from POST to GET.
// See: https://github.com/golang/go/issues/69063
// In general it could lead to unintended behaviour.
func TrimTrailingSlash(url string) string {
	res, found := strings.CutSuffix(url, "/")
	if found {
		slog.Info("Removed trailing slash in URL, as this could lead to undefined behaviour")
	}
	return res
}
