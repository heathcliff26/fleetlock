package k8s

import (
	"context"
	"testing"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/k8s/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	coordv1 "k8s.io/api/coordination/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

const (
	testNodeZincatiID = "35ba2101ae3f4d45b96e9c51f461bbff"
	testNodeMachineID = "dfd7882acda64c34aca76193c46f5d4e"
	testNodeName      = "Node1"
	testNamespace     = "fleetlock"
	testPodName       = "Pod1"
)

func initTestCluster(t *testing.T, client *fake.Clientset) {
	ctx := t.Context()

	testNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNodeName,
		},
		Status: v1.NodeStatus{
			NodeInfo: v1.NodeSystemInfo{MachineID: testNodeMachineID},
		},
	}
	_, _ = client.CoreV1().Nodes().Create(ctx, testNode, metav1.CreateOptions{})

	testNS := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, _ = client.CoreV1().Namespaces().Create(ctx, testNS, metav1.CreateOptions{})

	testPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testPodName,
			Namespace: testNamespace,
		},
		Spec: v1.PodSpec{
			NodeName:                      testNodeName,
			TerminationGracePeriodSeconds: utils.Pointer(int64(1)),
		},
	}
	_, _ = client.CoreV1().Pods(testNamespace).Create(ctx, testPod, metav1.CreateOptions{})
}

func TestNewClient(t *testing.T) {
	t.Run("NotInCluster", func(t *testing.T) {
		c, err := NewClient(NewDefaultConfig())
		assert.Nil(t, c, "Should not return a client")
		assert.Nil(t, err, "Should not return an error if not in cluster and no kubeconfig provided")
	})
	t.Run("KubeconfigNotFound", func(t *testing.T) {
		cfg := NewDefaultConfig()
		cfg.Kubeconfig = "not-a-file"

		c, err := NewClient(cfg)
		assert.Nil(t, c, "Should not return a client")
		assert.Error(t, err, "Should return an error if it can't find a kubeconfig")
	})
	t.Run("InvalidDrainTimeout", func(t *testing.T) {
		cfg := NewDefaultConfig()
		cfg.Kubeconfig = "testdata/kubeconfig"
		cfg.DrainTimeoutSeconds = 0

		c, err := NewClient(cfg)
		assert.Nil(t, c, "Should not return a client")
		assert.Error(t, err, "Should return an error")
	})
	t.Run("Success", func(t *testing.T) {
		cfg := NewDefaultConfig()
		cfg.Kubeconfig = "testdata/kubeconfig"
		cfg.DrainTimeoutSeconds = 5
		cfg.DrainRetries = 3

		c, err := NewClient(cfg)

		assert := assert.New(t)
		require := require.New(t)

		require.NoError(err, "Should not return an error")
		require.NotNil(c, "Should return a client")
		assert.Equal("fleetlock", c.namespace, "Namespace should be set")
		assert.Equal(cfg.DrainTimeoutSeconds, c.drainTimeoutSeconds, "Drain timeout should be set")
		assert.Equal(3, c.drainRetries, "Drain retries should be set")

	})
}

func TestDrainNode(t *testing.T) {
	t.Run("NoLease", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(t, client)

		ctx := t.Context()

		err := c.DrainNode(testNodeName)

		assert := assert.New(t)

		require.NoError(t, err, "Should not throw an error")

		node, _ := client.CoreV1().Nodes().Get(ctx, testNodeName, metav1.GetOptions{})
		assert.True(node.Spec.Unschedulable, "Node should be unscheduable")

		lease, _ := client.CoordinationV1().Leases(testNamespace).Get(ctx, drainLeaseName(testNodeName), metav1.GetOptions{})
		assert.Equal(utils.Pointer(leaseStateDone), lease.Spec.HolderIdentity, "Lease should indicate node is drained")
		assert.Equal(utils.Pointer(int32(300)), lease.Spec.LeaseDurationSeconds, "LeaseDurationSeconds should be set")
		assert.Equal(time.Now().Round(time.Second), lease.Spec.AcquireTime.Round(time.Second), "AcquireTime should be now")
	})
	t.Run("LeaseValid", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(t, client)

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.namespace,
				Name:      drainLeaseName(testNodeName),
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer(leaseStateDraining),
				LeaseDurationSeconds: utils.Pointer(int32(300)),
				AcquireTime:          &metav1.MicroTime{Time: time.Now()},
			},
		}
		_, _ = client.CoordinationV1().Leases(testNamespace).Create(t.Context(), lease, metav1.CreateOptions{})

		err := c.DrainNode(testNodeName)
		assert.Equal(t, NewErrorDrainIsLocked(), err, "Should return an error signaling that a drain is already in progress")
	})
	t.Run("LeaseInvalid", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(t, client)

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.namespace,
				Name:      drainLeaseName(testNodeName),
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity: utils.Pointer(leaseStateDraining),
			},
		}
		_, _ = client.CoordinationV1().Leases(testNamespace).Create(t.Context(), lease, metav1.CreateOptions{})

		err := c.DrainNode(testNodeName)
		assert.Equal(t, NewErrorInvalidLease(), err, "Should return an error signaling that the lease is invalid")
	})
	t.Run("LeaseExpired", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(t, client)

		ctx := t.Context()

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.namespace,
				Name:      drainLeaseName(testNodeName),
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer(leaseStateDraining),
				LeaseDurationSeconds: utils.Pointer(int32(300)),
				AcquireTime:          &metav1.MicroTime{Time: time.Now().Add(-6 * time.Minute)},
			},
		}
		_, _ = client.CoordinationV1().Leases(testNamespace).Create(ctx, lease, metav1.CreateOptions{})

		err := c.DrainNode(testNodeName)

		assert := assert.New(t)

		require.NoError(t, err, "Should not throw an error")

		node, _ := client.CoreV1().Nodes().Get(ctx, testNodeName, metav1.GetOptions{})
		assert.True(node.Spec.Unschedulable, "Node should be unscheduable")

		lease, _ = client.CoordinationV1().Leases(testNamespace).Get(ctx, drainLeaseName(testNodeName), metav1.GetOptions{})
		assert.Equal(utils.Pointer(leaseStateDone), lease.Spec.HolderIdentity, "Lease should indicate node is drained")
		assert.Equal(time.Now().Round(time.Second), lease.Spec.AcquireTime.Round(time.Second), "AcquireTime should be now")
	})
	t.Run("DrainTimeout", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(t, client)

		client.PrependReactor("create", "pods", func(action clienttesting.Action) (bool, runtime.Object, error) {
			time.Sleep(5 * time.Second)
			return false, nil, nil
		})

		c.drainTimeoutSeconds = 1

		err := c.DrainNode(testNodeName)

		assert := assert.New(t)

		assert.Equal(context.DeadlineExceeded, err, "Should exceed deadline")
		lease, _ := client.CoordinationV1().Leases(testNamespace).Get(t.Context(), drainLeaseName(testNodeName), metav1.GetOptions{})
		assert.Equal(leaseStateError, *lease.Spec.HolderIdentity, "Lease state should be error")
		assert.Equal("1", lease.GetAnnotations()[leaseFailCounterName], "Lease fail counter should be 1")
	})
}

func TestFindNodeByZincatiID(t *testing.T) {
	c, client := NewFakeClient()
	initTestCluster(t, client)

	node, err := c.FindNodeByZincatiID(testNodeZincatiID)

	assert := assert.New(t)

	assert.Equal(testNodeName, node, "Should have found correct node")
	assert.Nil(err, "Should have found correct node")

	node, err = c.FindNodeByZincatiID("abcdef123456789")
	assert.Equal("", node, "Should return an empty string if no node has been found")
	assert.Nil(err, "Should not return an error when no node has been found")
}

func TestUncordonNode(t *testing.T) {
	c, client := NewFakeClient()
	initTestCluster(t, client)

	ctx := t.Context()

	node, _ := client.CoreV1().Nodes().Get(ctx, testNodeName, metav1.GetOptions{})
	node.Spec.Unschedulable = true
	_, _ = client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	lease := &coordv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.namespace,
			Name:      drainLeaseName(testNodeName),
		},
		Spec: coordv1.LeaseSpec{
			HolderIdentity:       utils.Pointer(leaseStateDone),
			LeaseDurationSeconds: utils.Pointer(int32(300)),
			AcquireTime:          &metav1.MicroTime{Time: time.Now()},
		},
	}
	_, _ = client.CoordinationV1().Leases(testNamespace).Create(ctx, lease, metav1.CreateOptions{})

	err := c.UncordonNode(testNodeName)

	assert := assert.New(t)

	assert.Nil(err, "Should not return an error")

	_, err = client.CoordinationV1().Leases(testNamespace).Get(ctx, drainLeaseName(testNodeName), metav1.GetOptions{})
	assert.True(errors.IsNotFound(err), "Lease should be deleted")

	node, _ = client.CoreV1().Nodes().Get(ctx, testNodeName, metav1.GetOptions{})
	assert.False(node.Spec.Unschedulable, "Node should be schedulable")

	err = c.UncordonNode(testNodeName)
	assert.Nil(err, "Should not return an error")
}

func TestIsDrained(t *testing.T) {
	t.Run("NoLease", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(t, client)

		res, err := c.IsDrained(testNodeName)

		assert := assert.New(t)

		assert.Nil(err, "Should not return an error")
		assert.False(res, "Should return false")
	})
	t.Run("LeaseDone", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(t, client)

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.namespace,
				Name:      drainLeaseName(testNodeName),
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer(leaseStateDone),
				LeaseDurationSeconds: utils.Pointer(int32(300)),
				AcquireTime:          &metav1.MicroTime{Time: time.Now()},
			},
		}
		_, _ = client.CoordinationV1().Leases(testNamespace).Create(t.Context(), lease, metav1.CreateOptions{})

		res, err := c.IsDrained(testNodeName)

		assert := assert.New(t)

		assert.Nil(err, "Should not return an error")
		assert.True(res, "Should return true")
	})
	t.Run("LeaseDraining", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(t, client)

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.namespace,
				Name:      drainLeaseName(testNodeName),
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer(leaseStateDraining),
				LeaseDurationSeconds: utils.Pointer(int32(300)),
				AcquireTime:          &metav1.MicroTime{Time: time.Now()},
			},
		}
		_, _ = client.CoordinationV1().Leases(testNamespace).Create(t.Context(), lease, metav1.CreateOptions{})

		res, err := c.IsDrained(testNodeName)

		assert := assert.New(t)

		assert.Nil(err, "Should not return an error")
		assert.False(res, "Should return false")
	})
	t.Run("LeaseError", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(t, client)

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.namespace,
				Name:      drainLeaseName(testNodeName),
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer(leaseStateError),
				LeaseDurationSeconds: utils.Pointer(int32(300)),
				AcquireTime:          &metav1.MicroTime{Time: time.Now()},
			},
		}
		_, _ = client.CoordinationV1().Leases(testNamespace).Create(t.Context(), lease, metav1.CreateOptions{})

		res, err := c.IsDrained(testNodeName)

		assert := assert.New(t)

		assert.Nil(err, "Should not return an error")
		assert.False(res, "Should return false")
	})
	t.Run("LeaseRetries", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(t, client)
		c.drainRetries = 3

		ctx := t.Context()

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.namespace,
				Name:      drainLeaseName(testNodeName),
				Annotations: map[string]string{
					leaseFailCounterName: "2",
				},
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer(leaseStateError),
				LeaseDurationSeconds: utils.Pointer(int32(300)),
				AcquireTime:          &metav1.MicroTime{Time: time.Now()},
			},
		}
		lease, _ = client.CoordinationV1().Leases(testNamespace).Create(ctx, lease, metav1.CreateOptions{})

		res, err := c.IsDrained(testNodeName)

		assert := assert.New(t)

		assert.Nil(err, "Should not return an error")
		assert.False(res, "Should return false")

		lease.Annotations[leaseFailCounterName] = "3"
		_, _ = client.CoordinationV1().Leases(testNamespace).Update(ctx, lease, metav1.UpdateOptions{})

		res, err = c.IsDrained(testNodeName)

		assert.Nil(err, "Should not return an error")
		assert.True(res, "Should return true")
	})
}
