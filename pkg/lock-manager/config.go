package lockmanager

import (
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/errors"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/redis"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/sql"
)

type StorageConfig struct {
	Type     string              `yaml:"type"`
	SQLite   *sql.SQLiteConfig   `yaml:"sqlite,omitempty"`
	Postgres *sql.PostgresConfig `yaml:"postgres,omitempty"`
	Redis    *redis.RedisConfig  `yaml:"redis,omitempty"`
}

type Groups map[string]GroupConfig

type GroupConfig struct {
	Slots int `yaml:"slots"`
}

// Create a new storage config with default values
func NewDefaultStorageConfig() *StorageConfig {
	return &StorageConfig{
		Type: "memory",
	}
}

func NewDefaultGroups() Groups {
	groups := make(Groups, 1)
	groups["default"] = GroupConfig{
		Slots: 1,
	}
	return groups
}

func (g Groups) Validate() error {
	for _, v := range g {
		if v.Slots < 1 {
			return errors.NewErrorGroupSlotsOutOfRange()
		}
	}
	return nil
}
