package server

var (
	msgNotFound = FleetLockResponse{
		Kind:  "not_found",
		Value: "The requested url is not found on this server",
	}
	msgWrongMethod = FleetLockResponse{
		Kind:  "bad_request",
		Value: "Only accepts POST request",
	}
	msgMissingFleetLockHeader = FleetLockResponse{
		Kind:  "missing_fleetlock_header",
		Value: "The header fleet-lock-protocol must be set to true",
	}
	msgRequestParseFailed = FleetLockResponse{
		Kind:  "bad_request",
		Value: "The request json could not be parsed",
	}
	msgInvalidGroupValue = FleetLockResponse{
		Kind:  "bad_request",
		Value: "The value of group is invalid or empty. It must conform to \"" + groupValidationPattern + "\"",
	}
	msgEmptyID = FleetLockResponse{
		Kind:  "bad_request",
		Value: "The value of id is empty",
	}
	msgUnexpectedError = FleetLockResponse{
		Kind:  "error",
		Value: "An unexpected error occured",
	}
	msgSuccess = FleetLockResponse{
		Kind:  "success",
		Value: "The operation was succesfull",
	}
	msgSlotsFull = FleetLockResponse{
		Kind:  "all_slots_full",
		Value: "Could not reserve a slot as all slots in the group are currently locked already",
	}
	msgWaitingForNodeDrain = FleetLockResponse{
		Kind:  "waiting_for_node_drain",
		Value: "The Slot has been reserved, but the node is not yet drained",
	}
)
