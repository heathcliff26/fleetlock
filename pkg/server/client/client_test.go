package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	tMatrix := []struct {
		Name       string
		Url, Group string
		Success    bool
	}{
		{
			Name:  "MissingUrl",
			Group: "default",
		},
		{
			Name: "MissingGroup",
			Url:  "https://fleetlock.example.com",
		},
		{
			Name:    "Success",
			Group:   "default",
			Url:     "https://fleetlock.example.com",
			Success: true,
		},
	}
	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			res, err := NewClient(tCase.Url, tCase.Group)

			if tCase.Success {
				if !assert.NoError(err, "Should succeed") || !assert.NotNil(res, "Should return a client") {
					t.FailNow()
				}
				assert.Equal(tCase.Url, res.url, "Client should have the url set")
				assert.Equal(tCase.Group, res.group, "Client should have the group set")
				assert.NotEmpty(res.appID, "Client should have the appID set")
			} else {
				assert.Error(err, "Client creation should not succeed")
				assert.Nil(res, "Should not return a client")
			}
		})
	}
}

func TestNewEmptyClient(t *testing.T) {
	assert := assert.New(t)

	res, err := NewEmptyClient()

	if !assert.NoError(err, "Should succeed") || !assert.NotNil(res, "Should return a client") {
		t.FailNow()
	}

	assert.Empty(res.url, "Client should not have the url set")
	assert.Empty(res.group, "Client should not have the group set")
	assert.NotEmpty(res.appID, "Client should have the appID set")
}

func TestLock(t *testing.T) {
	assert := assert.New(t)

	c, srv := NewFakeServer(t, http.StatusOK, "/v1/pre-reboot")
	defer srv.Close()

	err := c.Lock()
	assert.NoError(err, "Should succeed")

	c2, srv2 := NewFakeServer(t, http.StatusLocked, "/v1/pre-reboot")
	defer srv2.Close()

	err = c2.Lock()
	assert.Error(err, "Should not succeed")
}

func TestRelease(t *testing.T) {
	assert := assert.New(t)

	c, srv := NewFakeServer(t, http.StatusOK, "/v1/steady-state")
	defer srv.Close()

	err := c.Release()
	assert.NoError(err, "Should succeed")

	c2, srv2 := NewFakeServer(t, http.StatusLocked, "/v1/steady-state")
	defer srv2.Close()

	err = c2.Release()
	assert.Error(err, "Should not succeed")
}

func TestGetAndSet(t *testing.T) {
	t.Run("URL", func(t *testing.T) {
		assert := assert.New(t)

		var c *FleetlockClient
		assert.Empty(c.GetURL(), "Should not panic when reading URL from nil pointer")

		c = &FleetlockClient{}

		assert.NoError(c.SetURL("https://fleetlock.example.com"), "Should set URL without error")
		assert.Equal("https://fleetlock.example.com", c.GetURL(), "URL should match")

		assert.NoError(c.SetURL("https://fleetlock.example.com/"), "Should set URL without trailing slash")
		assert.Equal("https://fleetlock.example.com", c.GetURL(), "URL should not have trailing /")

		assert.Error(c.SetURL(""), "Should not accept empty URL")
	})
	t.Run("Group", func(t *testing.T) {
		assert := assert.New(t)

		var c *FleetlockClient
		assert.Empty(c.GetGroup(), "Should not panic when reading group from nil pointer")

		c = &FleetlockClient{}

		assert.NoError(c.SetGroup("default"), "Should set group without error")
		assert.Equal("default", c.GetGroup(), "group should match")

		assert.Error(c.SetGroup(""), "Should not accept empty group")
	})
}

func NewFakeServer(t *testing.T, statusCode int, path string) (*FleetlockClient, *httptest.Server) {
	assert := assert.New(t)

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(path, req.URL.String(), "Request use the correct request URL")
		assert.Equal(http.MethodPost, req.Method, "Should be POST request")
		assert.Equal("true", strings.ToLower(req.Header.Get("fleet-lock-protocol")), "fleet-lock-protocol header should be set")

		params, err := ParseRequest(req.Body)
		assert.NoError(err, "Request should have the correct format")
		assert.Equal("testGroup", params.Client.Group, "Should have Group set")
		assert.Equal("testID", params.Client.ID, "Should have ID set")

		rw.WriteHeader(statusCode)
		b, err := json.MarshalIndent(FleetLockResponse{
			Kind:  "ok",
			Value: "Success",
		}, "", "  ")
		if !assert.NoError(err, "Error in fake server: failed to prepare response") {
			return
		}

		_, err = rw.Write(b)
		assert.NoError(err, "Error in fake server: failed to send response")
	}))
	c := &FleetlockClient{
		url:   srv.URL,
		group: "testGroup",
		appID: "testID",
	}
	return c, srv
}
