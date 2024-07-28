package kubernetes

import (
	"context"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConflictingGroupNames(t *testing.T) {
	nsName := "fleetlock"
	storage, client := NewKubernetesBackendWithFakeClient(nsName)
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName,
		},
	}
	_, _ = client.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})

	assert := assert.New(t)

	group := "default"
	id := "user"
	err := storage.Reserve(group, id)
	assert.Nil(err, "Should reserve slot")

	for i := 1; i < 10; i++ {
		err := storage.Reserve(group+"-"+strconv.Itoa(i), id+strconv.Itoa(i))
		assert.Nil(err, "Should reserve slot")
	}

	count, err := storage.GetLocks(group)
	assert.Nil(err, "Should not fail to obtain the locks")
	assert.Equal(1, count, "Should only count one lock")
}

func TestCompliantLeaseNames(t *testing.T) {
	nsName := "fleetlock"
	storage, client := NewKubernetesBackendWithFakeClient(nsName)
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName,
		},
	}
	_, _ = client.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})

	assert := assert.New(t)

	err := storage.Reserve("default", "User")
	assert.Nil(err, "Should reserve slot")

	leases, _ := client.CoordinationV1().Leases(nsName).List(context.TODO(), metav1.ListOptions{})

	validationRegex := regexp.MustCompile("^[a-z0-9.-]+$")

	for _, lease := range leases.Items {
		assert.True(validationRegex.MatchString(lease.GetName()), "Name should be compliant with k8s")
	}
}
