package client

import (
	"net/http"
	"testing"

	"github.com/heathcliff26/fleetlock/pkg/fake"
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
	t.Run("ID", func(t *testing.T) {
		assert := assert.New(t)

		var c *FleetlockClient
		assert.Empty(c.GetID(), "Should not panic when reading id from nil pointer")

		c = &FleetlockClient{}

		assert.NoError(c.SetID("testid"), "Should set id without error")
		assert.Equal("testid", c.GetID(), "id should match")

		assert.Error(c.SetID(""), "Should not accept empty id")
	})
}

func NewFakeServer(t *testing.T, statusCode int, path string) (*FleetlockClient, *fake.FakeServer) {
	testGroup, testID := "testGroup", "testID"

	srv := fake.NewFakeServer(t, statusCode, path)
	srv.Group = testGroup
	srv.ID = testID

	c := &FleetlockClient{
		url:   srv.URL(),
		group: testGroup,
		appID: testID,
	}
	return c, srv
}
