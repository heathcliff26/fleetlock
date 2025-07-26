package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorUnknownGroup(t *testing.T) {
	tests := []struct {
		name        string
		group       string
		expectedMsg string
	}{
		{
			name:        "ValidGroup",
			group:       "test-group",
			expectedMsg: "Unknown group: test-group",
		},
		{
			name:        "EmptyGroup",
			group:       "",
			expectedMsg: "Unknown group: ",
		},
		{
			name:        "SpecialCharacters",
			group:       "group@#$%",
			expectedMsg: "Unknown group: group@#$%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewErrorUnknownGroup(tt.group)
			assert.Error(t, err)
			assert.Equal(t, tt.expectedMsg, err.Error())
			
			// Test type assertion
			var unknownGroupErr *ErrorUnknownGroup
			assert.True(t, errors.As(err, &unknownGroupErr))
			assert.Equal(t, tt.group, unknownGroupErr.group)
		})
	}
}

func TestErrorEmptyID(t *testing.T) {
	err := ErrorEmptyID{}
	expectedMsg := "Received empty id, can't reserve a slot without an id"
	
	assert.Equal(t, expectedMsg, err.Error())
	
	// Test type assertion
	var emptyIDErr ErrorEmptyID
	assert.True(t, errors.As(err, &emptyIDErr))
}

func TestErrorUnkownStorageType(t *testing.T) {
	tests := []struct {
		name        string
		storageType string
		expectedMsg string
	}{
		{
			name:        "InvalidType",
			storageType: "invalid-storage",
			expectedMsg: "Unsupported storage type \"invalid-storage\" selected",
		},
		{
			name:        "EmptyType",
			storageType: "",
			expectedMsg: "Unsupported storage type \"\" selected",
		},
		{
			name:        "NumericType",
			storageType: "123",
			expectedMsg: "Unsupported storage type \"123\" selected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewErrorUnkownStorageType(tt.storageType)
			assert.Error(t, err)
			assert.Equal(t, tt.expectedMsg, err.Error())
			
			// Test type assertion
			var unknownStorageErr *ErrorUnkownStorageType
			assert.True(t, errors.As(err, &unknownStorageErr))
			assert.Equal(t, tt.storageType, unknownStorageErr.Type)
		})
	}
}

func TestErrorGroupSlotsOutOfRange(t *testing.T) {
	err := NewErrorGroupSlotsOutOfRange()
	expectedMsg := "At least one group has not enough slots, need at least 1"
	
	assert.Error(t, err)
	assert.Equal(t, expectedMsg, err.Error())
	
	// Test type assertion
	var slotsErr ErrorGroupSlotsOutOfRange
	assert.True(t, errors.As(err, &slotsErr))
}