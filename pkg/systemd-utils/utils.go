package systemdutils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// https://github.com/coreos/zincati/blob/main/src/identity/mod.rs
const ZincatiAppID = "de35106b6ec24688b63afddaa156679b"

// Derive the local app id for zincati from the machine id
func ZincatiMachineID(machineID string) (string, error) {
	return AppSpecificID(machineID, ZincatiAppID)
}

// Derive the machine specific app id from a given apps id and the local machine id
func AppSpecificID(machineID, appID string) (string, error) {
	// Remove any UUID dash formatting
	machineID = strings.ReplaceAll(machineID, "-", "")

	machineBytes, err := hex.DecodeString(machineID)
	if err != nil {
		return "", err
	}
	appBytes, err := hex.DecodeString(appID)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, machineBytes)
	mac.Write(appBytes)
	sum := mac.Sum(nil)

	// UUID v4 settings
	// https://docs.rs/libsystemd/0.3.1/src/libsystemd/id128.rs.html#52-54
	// https://github.com/systemd/systemd/blob/5a7eb46c0206411d380543021291b4bca0b6f59f/src/libsystemd/sd-id128/id128-util.c#L199
	sum[6] = (sum[6] & 0x0F) | 0x40
	sum[8] = (sum[8] & 0x3F) | 0x80

	id := string(sum)[:16]
	return fmt.Sprintf("%x", id), nil
}
