package client

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/heathcliff26/fleetlock/pkg/api"
)

type FleetlockClient struct {
	url   string
	group string
	appID string

	mutex sync.RWMutex
}

// Create a new client for fleetlock
func NewClient(url, group string) (*FleetlockClient, error) {
	c, err := NewEmptyClient()
	if err != nil {
		return nil, err
	}

	err = c.SetURL(url)
	if err != nil {
		return nil, err
	}

	err = c.SetGroup(group)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Create a new fleetlock client without url or group set
func NewEmptyClient() (*FleetlockClient, error) {
	appID, err := GetZincateAppID()
	if err != nil {
		return nil, fmt.Errorf("failed to create zincati app id: %v", err)
	}

	return &FleetlockClient{
		appID: appID,
	}, nil
}

// Aquire a lock for this machine
func (c *FleetlockClient) Lock() error {
	ok, res, err := c.doRequest("/v1/pre-reboot")
	if err != nil {
		return err
	} else if ok {
		return nil
	}
	return fmt.Errorf("failed to aquire lock kind=\"%s\" reason=\"%s\"", res.Kind, res.Value)
}

// Release the hold lock
func (c *FleetlockClient) Release() error {
	ok, res, err := c.doRequest("/v1/steady-state")
	if err != nil {
		return err
	} else if ok {
		return nil
	}
	return fmt.Errorf("failed to release lock kind=\"%s\" reason=\"%s\"", res.Kind, res.Value)
}

func (c *FleetlockClient) doRequest(path string) (bool, api.FleetLockResponse, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	body, err := api.PrepareRequest(c.group, c.appID)
	if err != nil {
		return false, api.FleetLockResponse{}, fmt.Errorf("failed to prepare request body: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, c.url+path, body)
	if err != nil {
		return false, api.FleetLockResponse{}, fmt.Errorf("failed to create http post request: %v", err)
	}
	req.Header.Set("fleet-lock-protocol", "true")
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, api.FleetLockResponse{}, fmt.Errorf("failed to send request to server: %v", err)
	}

	resBody, err := api.ParseResponse(res.Body)
	if err != nil {
		return false, api.FleetLockResponse{}, fmt.Errorf("failed to prepare response body: %v", err)
	}

	return res.StatusCode == http.StatusOK, resBody, nil
}

// Get the fleetlock server url
func (c *FleetlockClient) GetURL() string {
	if c == nil {
		return ""
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.url
}

// Change the fleetlock server url
func (c *FleetlockClient) SetURL(url string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if url == "" {
		return fmt.Errorf("the fleetlock server url can't be empty")
	}
	c.url = TrimTrailingSlash(url)
	return nil
}

// Get the fleetlock group
func (c *FleetlockClient) GetGroup() string {
	if c == nil {
		return ""
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.group
}

// Change the fleetlock group
func (c *FleetlockClient) SetGroup(group string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if group == "" {
		return fmt.Errorf("the fleetlock group can't be empty")
	}
	c.group = group
	return nil
}
