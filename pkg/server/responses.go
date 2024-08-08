package server

import "github.com/heathcliff26/fleetlock/pkg/server/client"

var (
	msgNotFound = client.FleetLockResponse{
		Kind:  "not_found",
		Value: "The requested url is not found on this server",
	}
	msgWrongMethod = client.FleetLockResponse{
		Kind:  "bad_request",
		Value: "Only accepts POST request",
	}
	msgMissingFleetLockHeader = client.FleetLockResponse{
		Kind:  "missing_fleetlock_header",
		Value: "The header fleet-lock-protocol must be set to true",
	}
	msgRequestParseFailed = client.FleetLockResponse{
		Kind:  "bad_request",
		Value: "The request json could not be parsed",
	}
	msgInvalidGroupValue = client.FleetLockResponse{
		Kind:  "bad_request",
		Value: "The value of group is invalid or empty. It must conform to \"" + groupValidationPattern + "\"",
	}
	msgEmptyID = client.FleetLockResponse{
		Kind:  "bad_request",
		Value: "The value of id is empty",
	}
	msgUnexpectedError = client.FleetLockResponse{
		Kind:  "error",
		Value: "An unexpected error occured",
	}
	msgSuccess = client.FleetLockResponse{
		Kind:  "success",
		Value: "The operation was succesfull",
	}
	msgSlotsFull = client.FleetLockResponse{
		Kind:  "all_slots_full",
		Value: "Could not reserve a slot as all slots in the group are currently locked already",
	}
	msgWaitingForNodeDrain = client.FleetLockResponse{
		Kind:  "waiting_for_node_drain",
		Value: "The Slot has been reserved, but the node is not yet drained",
	}
)
