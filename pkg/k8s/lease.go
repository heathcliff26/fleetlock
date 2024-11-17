package k8s

import (
	"context"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/k8s/utils"
	coordv1 "k8s.io/api/coordination/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes/typed/coordination/v1"
)

type lease struct {
	name   string
	lease  *coordv1.Lease
	client client.LeaseInterface
}

func NewLease(name string, client client.LeaseInterface) *lease {
	return &lease{
		name:   name,
		client: client,
	}
}

func (l *lease) Lock(ctx context.Context, duration int32) error {
	lease, err := l.client.Get(ctx, l.name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return l.create(ctx, duration)
	} else if err != nil {
		return err
	}

	if lease.Spec.AcquireTime == nil || lease.Spec.LeaseDurationSeconds == nil {
		return NewErrorInvalidLease()
	}

	validUntil := lease.Spec.AcquireTime.Time.Add(time.Duration(*lease.Spec.LeaseDurationSeconds) * time.Second)

	if time.Now().After(validUntil) {
		lease.Spec.AcquireTime = &metav1.MicroTime{Time: time.Now()}
		lease, err = l.client.Update(ctx, lease, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	} else {
		return NewErrorDrainIsLocked()
	}

	l.lease = lease
	return nil
}

func (l *lease) create(ctx context.Context, duration int32) error {
	lease := &coordv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name: l.name,
		},
		Spec: coordv1.LeaseSpec{
			HolderIdentity:       utils.Pointer("draining"),
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

func (l *lease) Done(ctx context.Context) error {
	if l.lease == nil {
		return NewErrorLeaseNil()
	}

	*l.lease.Spec.HolderIdentity = "done"
	lease, err := l.client.Update(ctx, l.lease, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	l.lease = lease
	return nil
}

func (l *lease) Delete(ctx context.Context) error {
	err := l.client.Delete(ctx, l.name, metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		return nil
	} else {
		return err
	}
}

func (l *lease) IsDone(ctx context.Context) (bool, error) {
	lease, err := l.client.Get(context.Background(), l.name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return *lease.Spec.HolderIdentity == "done", nil
}
