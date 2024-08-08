package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/heathcliff26/fleetlock/pkg/server/client"
)

func ReadUserIP(req *http.Request) string {
	IPAddress := req.Header.Get("x-real-ip")
	if IPAddress == "" {
		IPAddress = req.Header.Get("x-forwarded-for")
	}
	if IPAddress == "" {
		IPAddress = req.RemoteAddr
	}
	return IPAddress
}

// Send a response to the writer and handle impossible parse errors
func sendResponse(rw http.ResponseWriter, res client.FleetLockResponse) {
	b, err := json.Marshal(res)
	if err != nil {
		slog.Error("Failed to create Response", "err", err)
		return
	}

	_, err = rw.Write(b)
	if err != nil {
		slog.Error("Failed to send response to client", "err", err)
	}
}
