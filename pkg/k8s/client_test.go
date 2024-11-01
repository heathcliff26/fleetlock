package k8s

import (
	"context"
	"testing"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/k8s/utils"
	"github.com/stretchr/testify/assert"
	coordv1 "k8s.io/api/coordination/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	testNodeZincatiID = "35ba2101ae3f4d45b96e9c51f461bbff"
	testNodeMachineID = "dfd7882acda64c34aca76193c46f5d4e"
	testNodeName      = "Node1"
	testNamespace     = "fleetlock"
	testPodName       = "Pod1"
)

func initTestCluster(client *fake.Clientset) {
	testNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNodeName,
		},
		Status: v1.NodeStatus{
			NodeInfo: v1.NodeSystemInfo{MachineID: testNodeMachineID},
		},
	}
	_, _ = client.CoreV1().Nodes().Create(context.Background(), testNode, metav1.CreateOptions{})

	testNS := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, _ = client.CoreV1().Namespaces().Create(context.Background(), testNS, metav1.CreateOptions{})

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
	_, _ = client.CoreV1().Pods(testNamespace).Create(context.Background(), testPod, metav1.CreateOptions{})
}

func TestNewClient(t *testing.T) {
	t.Run("NotInCluster", func(t *testing.T) {
		c, err := NewClient("")
		assert.Nil(t, c, "Should not return a client")
		assert.Nil(t, err, "Should not return an error if not in cluster and no kubeconfig provided")
	})
	t.Run("KubeconfigNotFound", func(t *testing.T) {
		c, err := NewClient("not-a-file")
		assert.Nil(t, c, "Should not return a client")
		assert.Error(t, err, "Should return an error if it can't find a kubeconfig")
	})
	t.Run("Kubeconfig", func(t *testing.T) {
		c, err := NewClient("testdata/kubeconfig")
		assert.Nil(t, err, "Should not return an error")
		if !assert.NotNil(t, c, "Should return a client") {
			t.FailNow()
		}
		assert.Equal(t, "fleetlock", c.namespace)
	})
}

func TestDrainNode(t *testing.T) {
	t.Run("NoLease", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(client)

		err := c.DrainNode(testNodeName)

		assert := assert.New(t)

		if !assert.Nilf(err, "Should not throw an error: %v", err) {
			t.FailNow()
		}

		node, _ := client.CoreV1().Nodes().Get(context.Background(), testNodeName, metav1.GetOptions{})
		assert.True(node.Spec.Unschedulable, "Node should be unscheduable")

		lease, _ := client.CoordinationV1().Leases(testNamespace).Get(context.Background(), drainLeaseName(testNodeName), metav1.GetOptions{})
		assert.Equal(utils.Pointer("done"), lease.Spec.HolderIdentity, "Lease should indicate node is drained")
		assert.Equal(utils.Pointer(int32(300)), lease.Spec.LeaseDurationSeconds, "LeaseDurationSeconds should be set")
		assert.Equal(time.Now().Round(time.Second), lease.Spec.AcquireTime.Time.Round(time.Second), "AcquireTime should be now")
	})
	t.Run("LeaseValid", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(client)

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.namespace,
				Name:      drainLeaseName(testNodeName),
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer("draining"),
				LeaseDurationSeconds: utils.Pointer(int32(300)),
				AcquireTime:          &metav1.MicroTime{Time: time.Now()},
			},
		}
		_, _ = client.CoordinationV1().Leases(testNamespace).Create(context.Background(), lease, metav1.CreateOptions{})

		err := c.DrainNode(testNodeName)
		assert.Equal(t, NewErrorDrainIsLocked(), err, "Should return an error signaling that a drain is already in progress")
	})
	t.Run("LeaseExpired", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(client)

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.namespace,
				Name:      drainLeaseName(testNodeName),
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer("draining"),
				LeaseDurationSeconds: utils.Pointer(int32(300)),
				AcquireTime:          &metav1.MicroTime{Time: time.Now().Add(-6 * time.Minute)},
			},
		}
		_, _ = client.CoordinationV1().Leases(testNamespace).Create(context.Background(), lease, metav1.CreateOptions{})

		err := c.DrainNode(testNodeName)

		assert := assert.New(t)

		if !assert.Nilf(err, "Should not throw an error: %v", err) {
			t.FailNow()
		}

		node, _ := client.CoreV1().Nodes().Get(context.Background(), testNodeName, metav1.GetOptions{})
		assert.True(node.Spec.Unschedulable, "Node should be unscheduable")

		lease, _ = client.CoordinationV1().Leases(testNamespace).Get(context.Background(), drainLeaseName(testNodeName), metav1.GetOptions{})
		assert.Equal(utils.Pointer("done"), lease.Spec.HolderIdentity, "Lease should indicate node is drained")
		assert.Equal(time.Now().Round(time.Second), lease.Spec.AcquireTime.Time.Round(time.Second), "AcquireTime should be now")
	})
}

func TestFindNodeByZincatiID(t *testing.T) {
	c, client := NewFakeClient()
	initTestCluster(client)

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
	initTestCluster(client)

	node, _ := client.CoreV1().Nodes().Get(context.Background(), testNodeName, metav1.GetOptions{})
	node.Spec.Unschedulable = true
	_, _ = client.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
	lease := &coordv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.namespace,
			Name:      drainLeaseName(testNodeName),
		},
		Spec: coordv1.LeaseSpec{
			HolderIdentity:       utils.Pointer("done"),
			LeaseDurationSeconds: utils.Pointer(int32(300)),
			AcquireTime:          &metav1.MicroTime{Time: time.Now()},
		},
	}
	_, _ = client.CoordinationV1().Leases(testNamespace).Create(context.Background(), lease, metav1.CreateOptions{})

	err := c.UncordonNode(testNodeName)

	assert := assert.New(t)

	assert.Nil(err, "Should not return an error")

	_, err = client.CoordinationV1().Leases(testNamespace).Get(context.Background(), drainLeaseName(testNodeName), metav1.GetOptions{})
	assert.True(errors.IsNotFound(err), "Lease should be deleted")

	node, _ = client.CoreV1().Nodes().Get(context.Background(), testNodeName, metav1.GetOptions{})
	assert.False(node.Spec.Unschedulable, "Node should be schedulable")

	err = c.UncordonNode(testNodeName)
	assert.Nil(err, "Should not return an error")
}

func TestIsDrained(t *testing.T) {
	t.Run("NoLease", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(client)

		res, err := c.IsDrained(testNodeName)

		assert := assert.New(t)

		assert.Nil(err, "Should not return an error")
		assert.False(res, "Should return false")
	})
	t.Run("LeaseDone", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(client)

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.namespace,
				Name:      drainLeaseName(testNodeName),
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer("done"),
				LeaseDurationSeconds: utils.Pointer(int32(300)),
				AcquireTime:          &metav1.MicroTime{Time: time.Now()},
			},
		}
		_, _ = client.CoordinationV1().Leases(testNamespace).Create(context.Background(), lease, metav1.CreateOptions{})

		res, err := c.IsDrained(testNodeName)

		assert := assert.New(t)

		assert.Nil(err, "Should not return an error")
		assert.True(res, "Should return true")
	})
	t.Run("LeaseDraining", func(t *testing.T) {
		c, client := NewFakeClient()
		initTestCluster(client)

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.namespace,
				Name:      drainLeaseName(testNodeName),
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer("draining"),
				LeaseDurationSeconds: utils.Pointer(int32(300)),
				AcquireTime:          &metav1.MicroTime{Time: time.Now()},
			},
		}
		_, _ = client.CoordinationV1().Leases(testNamespace).Create(context.Background(), lease, metav1.CreateOptions{})

		res, err := c.IsDrained(testNodeName)

		assert := assert.New(t)

		assert.Nil(err, "Should not return an error")
		assert.False(res, "Should return false")
	})

}
