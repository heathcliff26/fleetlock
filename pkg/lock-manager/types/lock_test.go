package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLock(t *testing.T) {
	// Test basic Lock struct creation and field access
	now := time.Now()
	lock := Lock{
		Group:   "test-group",
		ID:      "test-id",
		Created: now,
	}

	assert.Equal(t, "test-group", lock.Group)
	assert.Equal(t, "test-id", lock.ID)
	assert.Equal(t, now, lock.Created)
}

func TestLockZeroValue(t *testing.T) {
	// Test zero value behavior
	var lock Lock
	
	assert.Equal(t, "", lock.Group)
	assert.Equal(t, "", lock.ID)
	assert.True(t, lock.Created.IsZero())
}

func TestLockComparison(t *testing.T) {
	// Test Lock struct comparison
	now := time.Now()
	
	lock1 := Lock{
		Group:   "group",
		ID:      "id1",
		Created: now,
	}
	
	lock2 := Lock{
		Group:   "group",
		ID:      "id1",
		Created: now,
	}
	
	lock3 := Lock{
		Group:   "group",
		ID:      "id2",
		Created: now,
	}

	assert.Equal(t, lock1, lock2, "Locks with same values should be equal")
	assert.NotEqual(t, lock1, lock3, "Locks with different IDs should not be equal")
}