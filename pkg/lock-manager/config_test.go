package lockmanager

import (
	"testing"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultStorageConfig(t *testing.T) {
	cfg := NewDefaultStorageConfig()

	assert := assert.New(t)

	assert.Equal("memory", cfg.Type)
	assert.Nil(cfg.SQLite)
}

func TestNewDefaultGroups(t *testing.T) {
	groups := NewDefaultGroups()

	assert := assert.New(t)

	assert.Equal(1, len(groups))

	defaultGroup := GroupConfig{
		Slots: 1,
	}

	assert.Equal(defaultGroup, groups["default"])
}

func TestGroupsValidate(t *testing.T) {
	tMatrix := []struct {
		Name   string
		Groups Groups
		Result error
	}{
		{
			Name: "Valid",
			Groups: Groups{
				"default": GroupConfig{
					Slots: 1,
				},
				"foo": GroupConfig{
					Slots: 10,
				},
				"bar": GroupConfig{
					Slots: 10000,
				},
			},
			Result: nil,
		},
		{
			Name: "Invalid",
			Groups: Groups{
				"default": GroupConfig{
					Slots: 0,
				},
			},
			Result: errors.NewErrorGroupSlotsOutOfRange(),
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			err := tCase.Groups.Validate()

			assert.Equal(t, tCase.Result, err)
		})
	}
}
