package k8s

import (
	"context"
	"strconv"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/k8s/utils"
	coordv1 "k8s.io/api/coordination/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes/typed/coordination/v1"
)

const (
	leaseStateDone     = "done"
	leaseStateDraining = "draining"
	leaseStateError    = "error"
)

const leaseFailCounterName = "fleetlock.heathcliff.eu/DrainFailCount"

type lease struct {
	name   string
	lease  *coordv1.Lease
	client client.LeaseInterface
}

// Create a new lease instance. Does not create a lease on the kubernetes side.
func NewLease(name string, client client.LeaseInterface) *lease {
	return &lease{
		name:   name,
		client: client,
	}
}

// Fetch the lease from the kubernetes server and store it in the local object.
// Does not need to be called manually.
func (l *lease) get(ctx context.Context) error {
	lease, err := l.client.Get(ctx, l.name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	l.lease = lease
	return nil
}

// Create the lease in the kubernetes server.
func (l *lease) create(ctx context.Context, duration int32) error {
	lease := &coordv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name: l.name,
		},
		Spec: coordv1.LeaseSpec{
			HolderIdentity:       utils.Pointer(leaseStateDraining),
			LeaseDurationSeconds: utils.Pointer(duration),
			AcquireTime:          &metav1.MicroTime{Time: time.Now()},
		},
	}

	lease, err := l.client.Create(ctx, lease, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	l.lease = lease
	return nil
}

// Update the lease on the kubernetes server
func (l *lease) update(ctx context.Context) error {
	lease, err := l.client.Update(ctx, l.lease, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	l.lease = lease
	return nil
}

func (l *lease) getFailCounter(ctx context.Context) (int, error) {
	if l.lease == nil {
		err := l.get(ctx)
		if err != nil {
			return 0, err
		}
	}

	failCounterStr, ok := l.lease.GetAnnotations()[leaseFailCounterName]
	if !ok {
		return 0, nil
	}
	return strconv.Atoi(failCounterStr)
}

// Return the count of failed attempts, defaults to 0 if no lease is found
func (l *lease) GetFailCounter(ctx context.Context) (int, error) {
	count, err := l.getFailCounter(ctx)
	if errors.IsNotFound(err) {
		return 0, nil
	}
	return count, err
}

// Increase the fail counter by 1.
// Does not call update!!!
func (l *lease) increaseFailCounter(ctx context.Context) error {
	failCount, err := l.getFailCounter(ctx)
	if err != nil {
		return err
	}

	if l.lease.Annotations == nil {
		l.lease.Annotations = make(map[string]string)
	}

	failCount++
	l.lease.Annotations[leaseFailCounterName] = strconv.Itoa(failCount)

	return nil
}

// Aquire the lease for draining
func (l *lease) Lock(ctx context.Context, duration int32) error {
	err := l.get(ctx)
	if errors.IsNotFound(err) {
		return l.create(ctx, duration)
	} else if err != nil {
		return err
	}

	if l.lease.Spec.AcquireTime == nil || l.lease.Spec.LeaseDurationSeconds == nil || l.lease.Spec.HolderIdentity == nil {
		return NewErrorInvalidLease()
	}

	validUntil := l.lease.Spec.AcquireTime.Add(time.Duration(*l.lease.Spec.LeaseDurationSeconds) * time.Second)

	if time.Now().After(validUntil) {
		if *l.lease.Spec.HolderIdentity == leaseStateDraining {
			err = l.increaseFailCounter(ctx)
			if err != nil {
				return err
			}
		}

		*l.lease.Spec.HolderIdentity = leaseStateDraining
		l.lease.Spec.AcquireTime = &metav1.MicroTime{Time: time.Now()}

		err = l.update(ctx)
		if err != nil {
			return err
		}
	} else {
		return NewErrorDrainIsLocked()
	}

	return nil
}

// Set the lease to done
func (l *lease) Done(ctx context.Context) error {
	if l.lease == nil {
		err := l.get(ctx)
		if err != nil {
			return err
		}
	}

	*l.lease.Spec.HolderIdentity = leaseStateDone
	err := l.update(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Delete the lease
func (l *lease) Delete(ctx context.Context) error {
	err := l.client.Delete(ctx, l.name, metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		return nil
	} else {
		return err
	}
}

// Return true if the lease is done.
// Does not return an error if the lease does not exist.
func (l *lease) IsDone(ctx context.Context) (bool, error) {
	if l.lease == nil {
		err := l.get(ctx)
		if errors.IsNotFound(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}
	}

	return *l.lease.Spec.HolderIdentity == leaseStateDone, nil
}

// Set the lease to an error state and increase the fail counter by one
func (l *lease) Error(ctx context.Context) error {
	err := l.increaseFailCounter(ctx)
	if err != nil {
		return err
	}

	*l.lease.Spec.HolderIdentity = leaseStateError

	err = l.update(ctx)
	if err != nil {
		return err
	}
	return nil
}
