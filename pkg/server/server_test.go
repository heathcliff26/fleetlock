package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/api"
	"github.com/heathcliff26/fleetlock/pkg/k8s"
	lockmanager "github.com/heathcliff26/fleetlock/pkg/lock-manager"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/memory"
	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	testNodeZincatiID = "35ba2101ae3f4d45b96e9c51f461bbff"
	testNodeMachineID = "dfd7882acda64c34aca76193c46f5d4e"
	testNodeName      = "Node1"
	testNamespace     = "fleetlock"
)

func TestNewServer(t *testing.T) {
	serverCfg := NewDefaultServerConfig()
	serverCfg.Defaults()
	groups := lockmanager.NewDefaultGroups()
	storageCfg := lockmanager.NewDefaultStorageConfig()
	k8s, _ := k8s.NewFakeClient()
	s, err := NewServer(serverCfg, groups, storageCfg, k8s)

	assert := assert.New(t)

	assert.NoError(err)
	assert.Equal(serverCfg, s.cfg)
	assert.NotNil(s.lm)
	assert.Equal(k8s, s.k8s)

	storageCfg.Type = "Unknown"

	s, err = NewServer(serverCfg, groups, storageCfg, nil)

	assert.Nil(s)
	assert.Equal("*errors.ErrorUnkownStorageType", reflect.TypeOf(err).String())
}

func TestRequestHandler(t *testing.T) {
	storage := memory.NewMemoryBackend([]string{"default"})
	lm := lockmanager.NewManagerWithStorage(lockmanager.NewDefaultGroups(), storage)
	s := &Server{
		cfg: &ServerConfig{},
		lm:  lm,
	}

	s.createHTTPServer()

	t.Run("Reserve", func(t *testing.T) {
		req := createRequest("/v1/pre-reboot", "default", "testUser")
		rr := httptest.NewRecorder()
		s.httpServer.Handler.ServeHTTP(rr, req)
		res, response, err := parseResponse(rr)

		assert := assert.New(t)

		assert.NoError(err)
		assert.Equal(http.StatusOK, res.StatusCode)
		assert.Equal(msgSuccess, response)

		ok, _ := storage.HasLock("default", "testUser")
		assert.True(ok)
	})
	t.Run("Release", func(t *testing.T) {
		req := createRequest("/v1/steady-state", "default", "testUser")
		rr := httptest.NewRecorder()
		s.httpServer.Handler.ServeHTTP(rr, req)
		res, response, err := parseResponse(rr)

		assert := assert.New(t)

		assert.NoError(err)
		assert.Equal(http.StatusOK, res.StatusCode)
		assert.Equal(msgSuccess, response)

		ok, _ := storage.HasLock("default", "t1")
		assert.False(ok)
	})
	t.Run("NotFound", func(t *testing.T) {
		req := createRequest("/foo/bar/nothing", "default", "testUser")
		rr := httptest.NewRecorder()
		s.httpServer.Handler.ServeHTTP(rr, req)

		assert := assert.New(t)

		assert.Equal(http.StatusNotFound, rr.Result().StatusCode)
	})
	t.Run("WrongMethod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/steady-state", createFleetLockRequest("default", "testUser"))
		rr := httptest.NewRecorder()
		s.httpServer.Handler.ServeHTTP(rr, req)

		assert := assert.New(t)

		assert.Equal(http.StatusMethodNotAllowed, rr.Result().StatusCode)
	})
	t.Run("MissingHeader", func(t *testing.T) {
		req := createRequest("/v1/steady-state", "default", "testUser")
		rr := httptest.NewRecorder()

		req.Header.Del("fleet-lock-protocol")

		s.httpServer.Handler.ServeHTTP(rr, req)
		res, response, err := parseResponse(rr)

		assert := assert.New(t)

		assert.NoError(err)
		assert.Equal(http.StatusBadRequest, res.StatusCode)
		assert.Equal(msgMissingFleetLockHeader, response)
	})
	t.Run("ParseError", func(t *testing.T) {
		req := createRequest("/v1/steady-state", "default", "testUser")
		rr := httptest.NewRecorder()

		req.Body = io.NopCloser(strings.NewReader("This is not json"))

		s.httpServer.Handler.ServeHTTP(rr, req)
		res, response, err := parseResponse(rr)

		assert := assert.New(t)

		assert.NoError(err)
		assert.Equal(http.StatusBadRequest, res.StatusCode)
		assert.Equal(msgRequestParseFailed, response)
	})
	t.Run("MissingGroup", func(t *testing.T) {
		req := createRequest("/v1/steady-state", "", "testUser")
		rr := httptest.NewRecorder()
		s.httpServer.Handler.ServeHTTP(rr, req)
		res, response, err := parseResponse(rr)

		assert := assert.New(t)

		assert.NoError(err)
		assert.Equal(http.StatusBadRequest, res.StatusCode)
		assert.Equal(msgInvalidGroupValue, response)
	})
	t.Run("MissingID", func(t *testing.T) {
		req := createRequest("/v1/steady-state", "default", "")
		rr := httptest.NewRecorder()
		s.httpServer.Handler.ServeHTTP(rr, req)
		res, response, err := parseResponse(rr)

		assert := assert.New(t)

		assert.NoError(err)
		assert.Equal(http.StatusBadRequest, res.StatusCode)
		assert.Equal(msgEmptyID, response)
	})
}

func TestHandleReserve(t *testing.T) {
	lm := lockmanager.NewManagerWithStorage(lockmanager.NewDefaultGroups(), memory.NewMemoryBackend([]string{"default"}))
	s := &Server{lm: lm}

	rr := httptest.NewRecorder()
	params := newFleetlockRequest("default", "testUser-1")
	s.handleReserve(rr, params)
	res, response, err := parseResponse(rr)

	assert := assert.New(t)

	assert.NoError(err)
	assert.Equal(http.StatusOK, res.StatusCode)
	assert.Equal(msgSuccess, response)

	rr = httptest.NewRecorder()
	params.Client.ID = "testUser-2"
	s.handleReserve(rr, params)
	res, response, err = parseResponse(rr)

	assert.NoError(err)
	assert.Equal(http.StatusLocked, res.StatusCode)
	assert.Equal(msgSlotsFull, response)

	rr = httptest.NewRecorder()
	params = newFleetlockRequest("", "testUser-3")
	s.handleReserve(rr, params)
	res, response, err = parseResponse(rr)

	assert.NoError(err)
	assert.Equal(http.StatusInternalServerError, res.StatusCode)
	assert.Equal(msgUnexpectedError, response)
}

func TestHandleRelease(t *testing.T) {
	lm := lockmanager.NewManagerWithStorage(lockmanager.NewDefaultGroups(), memory.NewMemoryBackend([]string{"default"}))
	s := &Server{lm: lm}

	rr := httptest.NewRecorder()
	params := newFleetlockRequest("", "testUser")
	s.handleRelease(rr, params)
	res, response, err := parseResponse(rr)

	assert := assert.New(t)

	assert.NoError(err)
	assert.Equal(http.StatusInternalServerError, res.StatusCode)
	assert.Equal(msgUnexpectedError, response)
}

func TestDrainNode(t *testing.T) {
	groups := lockmanager.NewDefaultGroups()
	groups["default"] = lockmanager.GroupConfig{
		Slots: 2,
	}
	lm := lockmanager.NewManagerWithStorage(groups, memory.NewMemoryBackend([]string{"default"}))
	k8s, fakeclient := k8s.NewFakeClient()
	s := &Server{
		lm:  lm,
		k8s: k8s,
	}
	initTestCluster(t, fakeclient)

	assert := assert.New(t)

	rr := httptest.NewRecorder()
	params := newFleetlockRequest("default", "abcdef123456789")
	s.handleReserve(rr, params)
	res, response, err := parseResponse(rr)

	assert.NoError(err, "Requests should be handled without error")
	assert.Equal(http.StatusOK, res.StatusCode, "Should return 200 OK when k8s node not found")
	assert.Equal(msgSuccess, response, "Should return success message when k8s node not found")

	rr = httptest.NewRecorder()
	params.Client.ID = testNodeZincatiID
	s.handleReserve(rr, params)
	res, response, err = parseResponse(rr)

	assert.NoError(err, "Requests should be handled without error")
	assert.Equal(http.StatusOK, res.StatusCode, "Should return 200 OK when node is still being drained")
	assert.Equal(msgWaitingForNodeDrain, response, "Should return message for node drain in progress")

	time.Sleep(10 * time.Millisecond)

	rr = httptest.NewRecorder()
	s.handleReserve(rr, params)
	res, response, err = parseResponse(rr)

	assert.NoError(err, "Requests should be handled without error")
	assert.Equal(http.StatusOK, res.StatusCode, "Should return 200 OK when node has been drained")
	assert.Equal(msgSuccess, response, "Should return success message when node has been drained")
}

func TestUncordonNode(t *testing.T) {
	lm := lockmanager.NewManagerWithStorage(lockmanager.NewDefaultGroups(), memory.NewMemoryBackend([]string{"default"}))
	k8s, fakeclient := k8s.NewFakeClient()
	s := &Server{
		lm:  lm,
		k8s: k8s,
	}
	initTestCluster(t, fakeclient)

	assert := assert.New(t)

	rr := httptest.NewRecorder()
	params := newFleetlockRequest("default", "abcdef123456789")
	s.handleRelease(rr, params)
	res, response, err := parseResponse(rr)

	assert.NoError(err)
	assert.Equal(http.StatusOK, res.StatusCode)
	assert.Equal(msgSuccess, response)

	rr = httptest.NewRecorder()
	params.Client.ID = testNodeZincatiID
	assert.True(s.uncordonNode(rr, params))
}

func TestHealthCheck(t *testing.T) {
	s := &Server{}
	rr := httptest.NewRecorder()

	assert := assert.New(t)

	s.handleHealthCheck(rr, nil)

	assert.Equal(http.StatusOK, rr.Result().StatusCode, "Health Check should return with 200")
	assert.Equal("application/json", rr.Header().Get("Content-Type"), "Content type should be json")

	var res api.FleetlockHealthResponse
	err := json.NewDecoder(rr.Result().Body).Decode(&res)

	assert.NoError(err, "Response should be parsable")
	expectedRes := api.FleetlockHealthResponse{
		Status: "ok",
	}
	assert.Equal(expectedRes, res, "Response should match")
}

func TestServerShutdown(t *testing.T) {
	s := &Server{}

	assert := assert.New(t)

	assert.NotPanics(func() {
		assert.NoError(s.Shutdown(), "Should shutdown just return nil when no server is running")
	}, "Should not panic when no httpServer is nil")

	s.httpServer = &http.Server{}

	assert.NoError(s.Shutdown(), "Should succeed in shutting down the server")
}

func newFleetlockRequest(group, id string) api.FleetLockRequest {
	return api.FleetLockRequest{
		Client: api.FleetLockRequestClient{
			ID:    id,
			Group: group,
		},
	}
}

func createFleetLockRequest(group, id string) io.Reader {
	body, _ := api.PrepareRequest(group, id)
	return body
}

func createRequest(target, group, id string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, target, createFleetLockRequest(group, id))

	req.Header.Set("fleet-lock-protocol", "true")

	return req
}

func parseResponse(rr *httptest.ResponseRecorder) (*http.Response, api.FleetLockResponse, error) {
	res := rr.Result()

	response, err := api.ParseResponse(res.Body)

	return res, response, err
}

func initTestCluster(t *testing.T, client *fake.Clientset) {
	testNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNodeName,
		},
		Status: v1.NodeStatus{
			NodeInfo: v1.NodeSystemInfo{MachineID: testNodeMachineID},
		},
	}
	_, _ = client.CoreV1().Nodes().Create(t.Context(), testNode, metav1.CreateOptions{})

	testNS := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, _ = client.CoreV1().Namespaces().Create(t.Context(), testNS, metav1.CreateOptions{})
}
