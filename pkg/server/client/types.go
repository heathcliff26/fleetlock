package client

type FleetLockRequest struct {
	Client FleetLockRequestClient `json:"client_params"`
}

type FleetLockRequestClient struct {
	ID    string `json:"id"`
	Group string `json:"group"`
}

type FleetLockResponse struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type FleetlockHealthResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
