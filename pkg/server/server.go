package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/api"
	"github.com/heathcliff26/fleetlock/pkg/k8s"
	lockmanager "github.com/heathcliff26/fleetlock/pkg/lock-manager"
	"github.com/heathcliff26/simple-fileserver/pkg/middleware"
)

const groupValidationPattern = "^[a-zA-Z0-9.-]+$"

var groupValidationRegex = regexp.MustCompile(groupValidationPattern)

type Server struct {
	cfg *ServerConfig
	lm  *lockmanager.LockManager
	k8s *k8s.Client

	httpServer *http.Server
}

// Create a new Server
func NewServer(cfg *ServerConfig, groups lockmanager.Groups, storageCfg lockmanager.StorageConfig, k8s *k8s.Client) (*Server, error) {
	lm, err := lockmanager.NewManager(groups, storageCfg)
	if err != nil {
		return nil, err
	}

	if k8s == nil {
		slog.Info("No kubernetes client available, will not drain nodes")
	}

	return &Server{
		cfg: cfg,
		lm:  lm,
		k8s: k8s,
	}, nil
}

// Main entrypoint for new requests
func (s *Server) requestHandler(rw http.ResponseWriter, req *http.Request) {
	var handleFunc func(http.ResponseWriter, api.FleetLockRequest)
	switch req.URL.String() {
	case "/v1/pre-reboot":
		handleFunc = s.handleReserve
	case "/v1/steady-state":
		handleFunc = s.handleRelease
	}

	// Verify FleetLock header is set
	if strings.ToLower(req.Header.Get("fleet-lock-protocol")) != "true" {
		slog.Debug("Received request with missing or wrong fleet-lock-protocol header", slog.String("remote", ReadUserIP(req)))
		rw.WriteHeader(http.StatusBadRequest)
		sendResponse(rw, msgMissingFleetLockHeader)
		return
	}

	params, err := api.ParseRequest(req.Body)
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
func (s *Server) handleReserve(rw http.ResponseWriter, params api.FleetLockRequest) {
	ok, err := s.lm.Reserve(params.Client.Group, params.Client.ID)
	if err != nil {
		slog.Error("Failed to reserve slot", "error", err, slog.String("group", params.Client.Group), slog.String("id", params.Client.ID))
		rw.WriteHeader(http.StatusInternalServerError)
		sendResponse(rw, msgUnexpectedError)
		return
	}

	if ok {
		slog.Info("Reserved slot", slog.String("group", params.Client.Group), slog.String("id", params.Client.ID))
		if s.k8s != nil && !s.drainNode(rw, params) {
			return
		}
		sendResponse(rw, msgSuccess)
	} else {
		slog.Debug("Could not reserve slot, all slots where filled", slog.String("group", params.Client.Group), slog.String("id", params.Client.ID))
		rw.WriteHeader(http.StatusLocked)
		sendResponse(rw, msgSlotsFull)
	}
}

// Handle requests to release a slot
//
//	URL: /v1/steady-state
func (s *Server) handleRelease(rw http.ResponseWriter, params api.FleetLockRequest) {
	if s.k8s != nil && !s.uncordonNode(rw, params) {
		return
	}

	err := s.lm.Release(params.Client.Group, params.Client.ID)
	if err != nil {
		slog.Error("Failed to release slot", "error", err, slog.String("group", params.Client.Group), slog.String("id", params.Client.ID))
		rw.WriteHeader(http.StatusInternalServerError)
		sendResponse(rw, msgUnexpectedError)
		return
	}
	slog.Info("Released slot", slog.String("group", params.Client.Group), slog.String("id", params.Client.ID))
	sendResponse(rw, msgSuccess)
}

// Drain the node after reservation and before sending success to api.
// Requires k8s client to be non-nil.
func (s *Server) drainNode(rw http.ResponseWriter, params api.FleetLockRequest) bool {
	node, ok := s.matchNodeToId(rw, params)
	if node == "" {
		return ok
	}

	drained, err := s.k8s.IsDrained(node)
	if err != nil {
		slog.Error("Could not check if node has been drained", "error", err, slog.String("group", params.Client.Group), slog.String("id", params.Client.ID), slog.String("node", node))
		rw.WriteHeader(http.StatusInternalServerError)
		sendResponse(rw, msgUnexpectedError)
		return false
	}
	if drained {
		slog.Info("Node is drained, client can continue", slog.String("group", params.Client.Group), slog.String("id", params.Client.ID), slog.String("node", node))
		return true
	}

	go func() {
		err := s.k8s.DrainNode(node)
		if err != nil {
			slog.Error("Failed to drain node", "error", err, slog.String("group", params.Client.Group), slog.String("id", params.Client.ID), slog.String("node", node))
		} else {
			slog.Info("Node finished draining, waiting for client to call again", slog.String("group", params.Client.Group), slog.String("id", params.Client.ID), slog.String("node", node))
		}
	}()

	// Return non-200 status to indicate the request is successful but the client needs to wait as it is still being processed.
	rw.WriteHeader(http.StatusAccepted)
	sendResponse(rw, msgWaitingForNodeDrain)
	return false
}

// Uncordon the node before release.
// Requires k8s client to be non-nil.
func (s *Server) uncordonNode(rw http.ResponseWriter, params api.FleetLockRequest) bool {
	node, ok := s.matchNodeToId(rw, params)
	if node == "" {
		return ok
	}

	err := s.k8s.UncordonNode(node)
	if err != nil {
		slog.Error("Failed to uncordon node", "error", err, slog.String("group", params.Client.Group), slog.String("id", params.Client.ID), slog.String("node", node))
		rw.WriteHeader(http.StatusInternalServerError)
		sendResponse(rw, msgUnexpectedError)
		return false
	}
	slog.Info("Uncordoned node", slog.String("group", params.Client.Group), slog.String("id", params.Client.ID), slog.String("node", node))
	return true
}

func (s *Server) matchNodeToId(rw http.ResponseWriter, params api.FleetLockRequest) (string, bool) {
	node, err := s.k8s.FindNodeByZincatiID(params.Client.ID)
	if err != nil {
		slog.Error("An error occured when matching client id to node", "error", err, slog.String("group", params.Client.Group), slog.String("id", params.Client.ID))
		rw.WriteHeader(http.StatusInternalServerError)
		sendResponse(rw, msgUnexpectedError)
		return "", false
	}

	if node == "" {
		slog.Info("Did not find a matching node for id", slog.String("group", params.Client.Group), slog.String("id", params.Client.ID))
	}

	return node, true
}

// Return a health status of the server
// URL: /healthz
func (s *Server) handleHealthCheck(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	status := api.FleetlockHealthResponse{
		Status: "ok",
	}
	sendResponse(rw, status)
}

// Prepare the http server for usage.
// This is in a separate function to allow testing the handler without running the server.
func (s *Server) createHTTPServer() {
	router := http.NewServeMux()
	router.HandleFunc("POST /v1/pre-reboot", s.requestHandler)
	router.HandleFunc("POST /v1/steady-state", s.requestHandler)
	router.HandleFunc("GET /healthz", s.handleHealthCheck)

	s.httpServer = &http.Server{
		Addr:         s.cfg.Listen,
		Handler:      middleware.Logging(router),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

// Starts the server and exits with error if that fails
func (s *Server) Run() error {
	if s.httpServer != nil {
		return fmt.Errorf("server already started")
	}

	s.createHTTPServer()
	defer func() {
		s.httpServer = nil
	}()

	var err error
	if s.cfg.SSL.Enabled {
		slog.Info("Starting server with SSL", slog.String("address", s.cfg.Listen))
		err = s.httpServer.ListenAndServeTLS(s.cfg.SSL.Cert, s.cfg.SSL.Key)
	} else {
		slog.Info("Starting server", slog.String("address", s.cfg.Listen))
		err = s.httpServer.ListenAndServe()
	}
	// This just means the server was closed after running
	if errors.Is(err, http.ErrServerClosed) {
		slog.Info("Server closed, exiting")
		return nil
	}
	return fmt.Errorf("failed to start server: %w", err)
}

func (s *Server) Shutdown() error {
	if s.httpServer == nil {
		return nil
	}

	slog.Info("Shutting down server")
	err := s.httpServer.Shutdown(context.Background())
	if err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}
	slog.Info("Server shutdown complete")
	return nil
}
