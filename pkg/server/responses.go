package server

import "github.com/heathcliff26/fleetlock/pkg/api"

var (
	msgMissingFleetLockHeader = api.FleetLockResponse{
		Kind:  "missing_fleetlock_header",
		Value: "The header fleet-lock-protocol must be set to true",
	}
	msgRequestParseFailed = api.FleetLockResponse{
		Kind:  "bad_request",
		Value: "The request json could not be parsed",
	}
	msgInvalidGroupValue = api.FleetLockResponse{
		Kind:  "bad_request",
		Value: "The value of group is invalid or empty. It must conform to \"" + groupValidationPattern + "\"",
	}
	msgEmptyID = api.FleetLockResponse{
		Kind:  "bad_request",
		Value: "The value of id is empty",
	}
	msgUnexpectedError = api.FleetLockResponse{
		Kind:  "error",
		Value: "An unexpected error occured",
	}
	msgSuccess = api.FleetLockResponse{
		Kind:  "success",
		Value: "The operation was succesfull",
	}
	msgSlotsFull = api.FleetLockResponse{
		Kind:  "all_slots_full",
		Value: "Could not reserve a slot as all slots in the group are currently locked already",
	}
	msgWaitingForNodeDrain = api.FleetLockResponse{
		Kind:  "waiting_for_node_drain",
		Value: "The Slot has been reserved, but the node is not yet drained",
	}
)
