package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	lockmanager "github.com/heathcliff26/fleetlock/pkg/lock-manager"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/memory"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	serverCfg := NewDefaultServerConfig()
	serverCfg.Defaults()
	groups := lockmanager.NewDefaultGroups()
	storageCfg := lockmanager.NewDefaultStorageConfig()
	s, err := NewServer(serverCfg, groups, storageCfg)

	assert := assert.New(t)

	assert.Nil(err)
	assert.Equal(serverCfg, s.cfg)
	assert.NotNil(s.lm)

	storageCfg.Type = "Unknown"

	s, err = NewServer(serverCfg, groups, storageCfg)

	assert.Nil(s)
	assert.Equal("*errors.ErrorUnkownStorageType", reflect.TypeOf(err).String())
}

func TestRequestHandler(t *testing.T) {
	storage := memory.NewMemoryBackend([]string{"default"})
	lm := lockmanager.NewManagerWithStorage(lockmanager.NewDefaultGroups(), storage)
	s := &Server{lm: lm}

	t.Run("Reserve", func(t *testing.T) {
		req := createRequest("/v1/pre-reboot", "default", "testUser")
		rr := httptest.NewRecorder()
		s.requestHandler(rr, req)
		res, response, err := parseResponse(rr)

		assert := assert.New(t)

		assert.Nil(err)
		assert.Equal(http.StatusOK, res.StatusCode)
		assert.Equal(msgSuccess, response)

		ok, _ := storage.HasLock("default", "testUser")
		assert.True(ok)
	})
	t.Run("Release", func(t *testing.T) {
		req := createRequest("/v1/steady-state", "default", "testUser")
		rr := httptest.NewRecorder()
		s.requestHandler(rr, req)
		res, response, err := parseResponse(rr)

		assert := assert.New(t)

		assert.Nil(err)
		assert.Equal(http.StatusOK, res.StatusCode)
		assert.Equal(msgSuccess, response)

		ok, _ := storage.HasLock("default", "t1")
		assert.False(ok)
	})
	t.Run("NotFound", func(t *testing.T) {
		req := createRequest("/foo/bar/nothing", "default", "testUser")
		rr := httptest.NewRecorder()
		s.requestHandler(rr, req)
		res, response, err := parseResponse(rr)

		assert := assert.New(t)

		assert.Nil(err)
		assert.Equal(http.StatusNotFound, res.StatusCode)
		assert.Equal(msgNotFound, response)
	})
	t.Run("WrongMethod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/steady-state", createFleetLockRequest("default", "testUser"))
		rr := httptest.NewRecorder()
		s.requestHandler(rr, req)
		res, response, err := parseResponse(rr)

		assert := assert.New(t)

		assert.Nil(err)
		assert.Equal(http.StatusMethodNotAllowed, res.StatusCode)
		assert.Equal(msgWrongMethod, response)
	})
	t.Run("MissingHeader", func(t *testing.T) {
		req := createRequest("/v1/steady-state", "default", "testUser")
		rr := httptest.NewRecorder()

		req.Header.Del("fleet-lock-protocol")

		s.requestHandler(rr, req)
		res, response, err := parseResponse(rr)

		assert := assert.New(t)

		assert.Nil(err)
		assert.Equal(http.StatusBadRequest, res.StatusCode)
		assert.Equal(msgMissingFleetLockHeader, response)
	})
	t.Run("ParseError", func(t *testing.T) {
		req := createRequest("/v1/steady-state", "default", "testUser")
		rr := httptest.NewRecorder()

		req.Body = io.NopCloser(strings.NewReader("This is not json"))

		s.requestHandler(rr, req)
		res, response, err := parseResponse(rr)

		assert := assert.New(t)

		assert.Nil(err)
		assert.Equal(http.StatusBadRequest, res.StatusCode)
		assert.Equal(msgRequestParseFailed, response)
	})
	t.Run("MissingGroup", func(t *testing.T) {
		req := createRequest("/v1/steady-state", "", "testUser")
		rr := httptest.NewRecorder()
		s.requestHandler(rr, req)
		res, response, err := parseResponse(rr)

		assert := assert.New(t)

		assert.Nil(err)
		assert.Equal(http.StatusBadRequest, res.StatusCode)
		assert.Equal(msgInvalidGroupValue, response)
	})
	t.Run("MissingID", func(t *testing.T) {
		req := createRequest("/v1/steady-state", "default", "")
		rr := httptest.NewRecorder()
		s.requestHandler(rr, req)
		res, response, err := parseResponse(rr)

		assert := assert.New(t)

		assert.Nil(err)
		assert.Equal(http.StatusBadRequest, res.StatusCode)
		assert.Equal(msgEmptyID, response)
	})
}

func TestHandleReserve(t *testing.T) {
	lm := lockmanager.NewManagerWithStorage(lockmanager.NewDefaultGroups(), memory.NewMemoryBackend([]string{"default"}))
	s := &Server{lm: lm}

	rr := httptest.NewRecorder()
	params := FleetLockRequest{
		Client: struct {
			ID    string "json:\"id\""
			Group string "json:\"group\""
		}{
			ID:    "testUser-1",
			Group: "default",
		},
	}
	s.handleReserve(rr, params)
	res, response, err := parseResponse(rr)

	assert := assert.New(t)

	assert.Nil(err)
	assert.Equal(http.StatusOK, res.StatusCode)
	assert.Equal(msgSuccess, response)

	rr = httptest.NewRecorder()
	params.Client.ID = "testUser-2"
	s.handleReserve(rr, params)
	res, response, err = parseResponse(rr)

	assert.Nil(err)
	assert.Equal(http.StatusLocked, res.StatusCode)
	assert.Equal(msgSlotsFull, response)

	rr = httptest.NewRecorder()
	params.Client.ID = "testUser-3"
	params.Client.Group = ""
	s.handleReserve(rr, params)
	res, response, err = parseResponse(rr)

	assert.Nil(err)
	assert.Equal(http.StatusInternalServerError, res.StatusCode)
	assert.Equal(msgUnexpectedError, response)
}

func TestHandleRelease(t *testing.T) {
	lm := lockmanager.NewManagerWithStorage(lockmanager.NewDefaultGroups(), memory.NewMemoryBackend([]string{"default"}))
	s := &Server{lm: lm}

	rr := httptest.NewRecorder()
	params := FleetLockRequest{
		Client: struct {
			ID    string "json:\"id\""
			Group string "json:\"group\""
		}{
			ID:    "testUser",
			Group: "",
		},
	}
	s.handleRelease(rr, params)
	res, response, err := parseResponse(rr)

	assert := assert.New(t)

	assert.Nil(err)
	assert.Equal(http.StatusInternalServerError, res.StatusCode)
	assert.Equal(msgUnexpectedError, response)
}

func createFleetLockRequest(group, id string) io.Reader {
	msg := FleetLockRequest{
		Client: struct {
			ID    string "json:\"id\""
			Group string "json:\"group\""
		}{
			ID:    id,
			Group: group,
		},
	}
	body, _ := json.Marshal(msg)
	return bytes.NewReader(body)
}

func createRequest(target, group, id string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, target, createFleetLockRequest(group, id))

	req.Header.Set("fleet-lock-protocol", "true")

	return req
}

func parseResponse(rr *httptest.ResponseRecorder) (*http.Response, FleetLockResponse, error) {
	res := rr.Result()

	var response FleetLockResponse
	err := json.NewDecoder(res.Body).Decode(&response)

	return res, response, err
}
