package k8s

type ErrorFailedToEvictAllPods struct{}

func NewErrorFailedToEvictAllPods() error {
	return ErrorFailedToEvictAllPods{}
}

func (e ErrorFailedToEvictAllPods) Error() string {
	return "Failed to evict all pods from node"
}

type ErrorDrainIsLocked struct{}

func NewErrorDrainIsLocked() error {
	return ErrorDrainIsLocked{}
}

func (e ErrorDrainIsLocked) Error() string {
	return "Can't drain node, as another drain is already in progress"
}

type ErrorInvalidLease struct{}

func NewErrorInvalidLease() error {
	return ErrorInvalidLease{}
}

func (e ErrorInvalidLease) Error() string {
	return "Invalid lease, either AcquireTime or LeaseDurationSeconds are nil"
}

type ErrorLeaseNil struct{}

func NewErrorLeaseNil() error {
	return ErrorLeaseNil{}
}

func (e ErrorLeaseNil) Error() string {
	return "Tried changing lease, but lease status on cluster is unknown"
}
