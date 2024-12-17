package fake

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/heathcliff26/fleetlock/pkg/api"
	"github.com/stretchr/testify/assert"
)

type FakeServer struct {
	server *httptest.Server
	assert *assert.Assertions

	Path       string
	StatusCode int
	Group      string
	ID         string
}

func NewFakeServer(t *testing.T, statusCode int, path string) *FakeServer {
	s := &FakeServer{
		assert:     assert.New(t),
		Path:       path,
		StatusCode: statusCode,
	}

	s.server = httptest.NewServer(http.HandlerFunc(s.handleRequest))

	return s
}

func (s *FakeServer) handleRequest(rw http.ResponseWriter, req *http.Request) {
	s.assert.Contains([]string{"/v1/pre-reboot", "/v1/steady-state"}, req.URL.String(), "Should request a valid url")
	if s.Path != "" {
		s.assert.Equal(s.Path, req.URL.String(), "Should use the specified URL")
	}

	s.assert.Equal(http.MethodPost, req.Method, "Should be POST request")
	s.assert.Equal("true", strings.ToLower(req.Header.Get("fleet-lock-protocol")), "fleet-lock-protocol header should be set")

	params, err := api.ParseRequest(req.Body)
	s.assert.NoError(err, "Request should have the correct format")

	if s.Group != "" {
		s.assert.Equal(s.Group, params.Client.Group, "Should have expected group")
	} else {
		s.assert.NotEmpty(params.Client.Group, "Should have group set")
	}
	if s.ID != "" {
		s.assert.Equal(s.ID, params.Client.ID, "Should have expected id")
	} else {
		s.assert.NotEmpty(params.Client.ID, "Should have id set")
	}

	rw.WriteHeader(s.StatusCode)
	b, err := json.MarshalIndent(api.FleetLockResponse{
		Kind:  "ok",
		Value: "Success",
	}, "", "  ")
	if !s.assert.NoError(err, "Error in fake server: failed to prepare response") {
		return
	}

	_, err = rw.Write(b)
	s.assert.NoError(err, "Error in fake server: failed to send response")
}

func (s *FakeServer) URL() string {
	return s.server.URL
}

func (s *FakeServer) Close() {
	if s != nil && s.server != nil {
		s.server.Close()
	}
}
