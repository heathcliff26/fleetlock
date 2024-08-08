package client

import (
	"bytes"
	"encoding/json"
	"io"
)

// Parse an http request	 body and extract the parameters
func ParseRequest(body io.ReadCloser) (FleetLockRequest, error) {
	var res FleetLockRequest
	err := json.NewDecoder(body).Decode(&res)
	if err != nil {
		return FleetLockRequest{}, err
	}
	return res, nil
}

// Parse an http response body and extract the parameters
func ParseResponse(body io.ReadCloser) (FleetLockResponse, error) {
	var res FleetLockResponse
	err := json.NewDecoder(body).Decode(&res)
	if err != nil {
		return FleetLockResponse{}, err
	}
	return res, nil
}

// Create a new http request pody based on the provided parameters
func PrepareRequest(group, id string) (io.Reader, error) {
	req := FleetLockRequest{
		Client: FleetLockRequestClient{
			ID:    id,
			Group: group,
		},
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(body), nil
}
