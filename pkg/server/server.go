package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	lockmanager "github.com/heathcliff26/fleetlock/pkg/lock-manager"
)

const groupValidationPattern = "^[a-zA-Z0-9.-]+$"

var groupValidationRegex = regexp.MustCompile(groupValidationPattern)

type Server struct {
	cfg *ServerConfig
	lm  *lockmanager.LockManager
}

type FleetLockRequest struct {
	Client struct {
		ID    string `json:"id"`
		Group string `json:"group"`
	} `json:"client_params"`
}

type FleetLockResponse struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// Create a new Server
func NewServer(cfg *ServerConfig, groups lockmanager.Groups, storageCfg *lockmanager.StorageConfig) (*Server, error) {
	lm, err := lockmanager.NewManager(groups, storageCfg)
	if err != nil {
		return nil, err
	}

	return &Server{
		cfg: cfg,
		lm:  lm,
	}, nil
}

// Main entrypoint for new requests
func (s *Server) requestHandler(rw http.ResponseWriter, req *http.Request) {
	slog.Debug("Received request", slog.String("method", req.Method), slog.String("uri", req.RequestURI), slog.String("remote", ReadUserIP(req)))

	var handleFunc func(http.ResponseWriter, FleetLockRequest)
	switch req.URL.String() {
	case "/v1/pre-reboot":
		handleFunc = s.handleReserve
	case "/v1/steady-state":
		handleFunc = s.handleRelease
	default:
		slog.Debug("Unknown URL", slog.String("url", req.URL.String()), slog.String("remote", ReadUserIP(req)))
		rw.WriteHeader(http.StatusNotFound)
		sendResponse(rw, msgNotFound)
		return
	}

	// Verify right method
	if req.Method != http.MethodPost {
		slog.Debug("Received request with wrong method", slog.String("method", req.Method), slog.String("remote", ReadUserIP(req)))
		rw.WriteHeader(http.StatusMethodNotAllowed)
		sendResponse(rw, msgWrongMethod)
		return
	}

	// Verify FleetLock header is set
	if strings.ToLower(req.Header.Get("fleet-lock-protocol")) != "true" {
		slog.Debug("Received request with missing or wrong fleet-lock-protocol header", slog.String("remote", ReadUserIP(req)))
		rw.WriteHeader(http.StatusBadRequest)
		sendResponse(rw, msgMissingFleetLockHeader)
		return
	}

	var params FleetLockRequest
	err := json.NewDecoder(req.Body).Decode(&params)
	if err != nil {
		slog.Debug("Failed to parse request", "error", err, slog.String("remote", ReadUserIP(req)))
		rw.WriteHeader(http.StatusBadRequest)
		sendResponse(rw, msgRequestParseFailed)
		return
	}

	if strings.Contains(params.Client.Group, "\n") || !groupValidationRegex.MatchString(params.Client.Group) {
		slog.Debug("Request contained invalid characters for group", slog.String("group", params.Client.Group), slog.String("remote", ReadUserIP(req)))
		rw.WriteHeader(http.StatusBadRequest)
		sendResponse(rw, msgInvalidGroupValue)
		return
	}

	if params.Client.ID == "" {
		slog.Debug("Request did not contain an id", slog.String("remote", ReadUserIP(req)))
		rw.WriteHeader(http.StatusBadRequest)
		sendResponse(rw, msgEmptyID)
		return
	}

	handleFunc(rw, params)
}

// Handle requests to reserve a slot
//
//	URL: /v1/pre-reboot
func (s *Server) handleReserve(rw http.ResponseWriter, params FleetLockRequest) {
	ok, err := s.lm.Reserve(params.Client.Group, params.Client.ID)
	if err != nil {
		slog.Error("Failed to reserve slot", "error", err, slog.String("group", params.Client.Group), slog.String("group", params.Client.ID))
		rw.WriteHeader(http.StatusInternalServerError)
		sendResponse(rw, msgUnexpectedError)
		return
	}
	if ok {
		slog.Info("Reserved slot", slog.String("group", params.Client.Group), slog.String("group", params.Client.ID))
		sendResponse(rw, msgSuccess)
	} else {
		slog.Debug("Could not reserve slot, all slots where filled", slog.String("group", params.Client.Group), slog.String("group", params.Client.ID))
		rw.WriteHeader(http.StatusLocked)
		sendResponse(rw, msgSlotsFull)
	}
}

// Handle requests to release a slot
//
//	URL: /v1/steady-state
func (s *Server) handleRelease(rw http.ResponseWriter, params FleetLockRequest) {
	err := s.lm.Release(params.Client.Group, params.Client.ID)
	if err != nil {
		slog.Error("Failed to release slot", "error", err, slog.String("group", params.Client.Group), slog.String("group", params.Client.ID))
		rw.WriteHeader(http.StatusInternalServerError)
		sendResponse(rw, msgUnexpectedError)
		return
	}
	slog.Info("Released slot", slog.String("group", params.Client.Group), slog.String("group", params.Client.ID))
	sendResponse(rw, msgSuccess)
}

// Starts the server and exits with error if that fails
func (s *Server) Run() error {
	http.HandleFunc("/", s.requestHandler)

	slog.Info("Starting server", slog.String("listen", s.cfg.Listen), slog.Bool("ssl", s.cfg.SSL.Enabled))

	var err error
	if s.cfg.SSL.Enabled {
		err = http.ListenAndServeTLS(s.cfg.Listen, s.cfg.SSL.Cert, s.cfg.SSL.Key, nil)
	} else {
		err = http.ListenAndServe(s.cfg.Listen, nil)
	}
	// This just means the server was closed after running
	if errors.Is(err, http.ErrServerClosed) {
		slog.Info("Server closed, exiting")
		return nil
	}
	return err
}