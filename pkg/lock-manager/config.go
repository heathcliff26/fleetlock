package lockmanager

import (
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/errors"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/etcd"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/kubernetes"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/sql"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/valkey"
)

type StorageConfig struct {
	Type       string                      `json:"type"`
	SQLite     sql.SQLiteConfig            `json:"sqlite,omitempty"`
	Postgres   sql.PostgresConfig          `json:"postgres,omitempty"`
	MySQL      sql.MySQLConfig             `json:"mysql,omitempty"`
	Valkey     valkey.ValkeyConfig         `json:"valkey,omitempty"`
	Etcd       etcd.EtcdConfig             `json:"etcd,omitempty"`
	Kubernetes kubernetes.KubernetesConfig `json:"kubernetes,omitempty"`
}

type Groups map[string]GroupConfig

type GroupConfig struct {
	Slots int `json:"slots"`
}

// Create a new storage config with default values
func NewDefaultStorageConfig() StorageConfig {
	return StorageConfig{
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
