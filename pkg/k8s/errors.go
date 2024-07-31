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
