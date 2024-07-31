package k8s

import (
	"fmt"
)

func drainLeaseName(id string) string {
	return fmt.Sprintf("fleetlock-drain-%s", id)
}

func nodeUnschedulablePatch(desired bool) []byte {
	return []byte(fmt.Sprintf("{\"spec\":{\"unschedulable\":%t}}", desired))
}
