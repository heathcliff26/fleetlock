package api

// Wrapper struct for the request parameters.
// The actual parameters are stored in "client_params".
type FleetLockRequest struct {
	Client FleetLockRequestClient `json:"client_params"`
}

// The parameters for a fleetlock request.
// Needs to contain ID and group.
type FleetLockRequestClient struct {
	// The unique id for the client.
	// Generally this would be the zincati app id,
	// meaning the unique machine id masked by the zincati app id.
	ID string `json:"id"`
	// This would be the group to which the client belongs.
	// Zincati uses "default" here as a default value.
	Group string `json:"group"`
}

// The response sent by the server to the client.
// Please note that success or failure are indicated by the HTTP Status.
// These values here can be any arbitrary value and should mostly be used
// to display the failure reason in the logs.
type FleetLockResponse struct {
	// The kind of response, should be something the client code can identify, e.g. success, error, etc.
	Kind string `json:"kind"`
	// Essentially the reason for the above kind, best to be log/user readable.
	Value string `json:"value"`
}

// Not part of the actual api specification, used for the server to indicate it is up.
type FleetlockHealthResponse struct {
	// The current status of the server. Should be indicated by a 200 response anyway.
	// Should be "ok" if nothing is wrong
	Status string `json:"status"`
	// The specific error if there is a problem.
	// Empty when there is no error.
	Error string `json:"error"`
}
