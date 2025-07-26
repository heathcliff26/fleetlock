package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryBackendClose(t *testing.T) {
	// Test that Close() method exists and works correctly
	backend := NewMemoryBackend([]string{"default"})
	
	err := backend.Close()
	assert.NoError(t, err, "Close should not return an error")
}

func TestMemoryBackendHasLockInternal(t *testing.T) {
	// Test the internal hasLock function through HasLock
	backend := NewMemoryBackend([]string{"default"})
	
	// Test when no locks exist
	hasLock, err := backend.HasLock("default", "test-id")
	assert.NoError(t, err)
	assert.False(t, hasLock, "Should return false when no locks exist")
	
	// Reserve a lock
	err = backend.Reserve("default", "test-id")
	assert.NoError(t, err)
	
	// Test when lock exists
	hasLock, err = backend.HasLock("default", "test-id")
	assert.NoError(t, err)
	assert.True(t, hasLock, "Should return true when lock exists")
	
	// Test with different ID
	hasLock, err = backend.HasLock("default", "different-id")
	assert.NoError(t, err)
	assert.False(t, hasLock, "Should return false for different ID")
}

func TestMemoryBackendUnknownGroup(t *testing.T) {
	// Test behavior with unknown groups
	backend := NewMemoryBackend([]string{"default"})
	
	hasLock, err := backend.HasLock("unknown-group", "test-id")
	assert.Error(t, err, "Should return error for unknown group")
	assert.False(t, hasLock, "Should return false for unknown group")
}