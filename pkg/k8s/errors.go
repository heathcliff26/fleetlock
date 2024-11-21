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
	return "Invalid lease, either AcquireTime, LeaseDurationSeconds or HolderIdentity are nil"
}

type ErrorDrainTimeoutSecondsInvalid struct{}

func NewErrorDrainTimeoutSecondsInvalid() error {
	return ErrorDrainTimeoutSecondsInvalid{}
}

func (e ErrorDrainTimeoutSecondsInvalid) Error() string {
	return "drainTimeoutSeconds value needs to be greater than 0"
}
