package lockmanager

import (
	"reflect"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/etcd"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/kubernetes"

	//nolint:staticcheck
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/redis"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/sql"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/valkey"
	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	mr := miniredis.RunT(t)
	type result struct {
		groups  map[string]*lockGroup
		storage string
	}

	tMatrix := []struct {
		Name    string
		Storage StorageConfig
		Result  result
		Error   string
	}{
		{
			Name: "MemoryBackend",
			Storage: StorageConfig{
				Type: "memory",
			},
			Result: result{
				groups:  initGroups(NewDefaultGroups()),
				storage: "*memory.MemoryBackend",
			},
			Error: "",
		},
		{
			Name: "SQLiteBackend",
			Storage: StorageConfig{
				Type: "sqlite",
				SQLite: sql.SQLiteConfig{
					File: "file:test.db?mode=memory",
				},
			},
			Result: result{
				groups:  initGroups(NewDefaultGroups()),
				storage: "*sql.SQLBackend",
			},
			Error: "",
		},
		{
			Name: "ErrorNewSQLiteBackend",
			Storage: StorageConfig{
				Type: "sqlite",
				SQLite: sql.SQLiteConfig{
					File: "/not/a/valid/path/to/database",
				},
			},
			Error: "failed to \"ping\" sqlite database",
		},
		{
			Name: "ErrorNewPostgresBackend",
			Storage: StorageConfig{
				Type: "postgres",
				Postgres: sql.PostgresConfig{
					Address:  "localhost:1234",
					Database: "nothing",
				},
			},
			Error: "failed to ping postgres database",
		},
		{
			Name: "ErrorNewMySQLBackend",
			Storage: StorageConfig{
				Type: "mysql",
				MySQL: sql.MySQLConfig{
					Address:  "localhost:1234",
					Database: "nothing",
				},
			},
			Error: "failed to open mysql database",
		},
		{
			Name: "RedisBackend",
			Storage: StorageConfig{
				Type: "redis",
				Redis: redis.RedisConfig{
					Addr: mr.Addr(),
				},
			},
			Result: result{
				groups:  initGroups(NewDefaultGroups()),
				storage: "*redis.RedisBackend",
			},
			Error: "",
		},
		{
			Name: "ErrorNewRedisBackend",
			Storage: StorageConfig{
				Type: "redis",
				Redis: redis.RedisConfig{
					Addr: "",
				},
			},
			Error: "no alive address in InitAddress",
		},
		{
			Name: "ValkeyBackend",
			Storage: StorageConfig{
				Type: "valkey",
				Valkey: valkey.ValkeyConfig{
					Addrs: []string{mr.Addr()},
				},
			},
			Result: result{
				groups:  initGroups(NewDefaultGroups()),
				storage: "*valkey.ValkeyBackend",
			},
			Error: "",
		},
		{
			Name: "ErrorNewValkeyBackend",
			Storage: StorageConfig{
				Type: "valkey",
				Valkey: valkey.ValkeyConfig{
					Addrs: []string{},
				},
			},
			Error: "no alive address in InitAddress",
		},
		{
			Name: "ErrorNewEtcdBackend",
			Storage: StorageConfig{
				Type: "etcd",
				Etcd: etcd.EtcdConfig{
					Endpoints: []string{},
				},
			},
			Error: "failed to create etcd client",
		},
		{
			Name: "ErrorNewKubernetesBackend",
			Storage: StorageConfig{
				Type:       "kubernetes",
				Kubernetes: kubernetes.KubernetesConfig{},
			},
			Error: "unable to load in-cluster configuration",
		},
		{
			Name: "UnknownStorageType",
			Storage: StorageConfig{
				Type: "not-a-valid-type",
			},
			Error: "Unsupported storage type",
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			lm, err := NewManager(NewDefaultGroups(), tCase.Storage)

			if tCase.Error == "" {
				assert.Nilf(err, "Should not return an error but returned: %v", err)
				if !assert.NotNil(lm) {
					t.FailNow()
				}
				assert.Equal(tCase.Result.groups, lm.groups)
				assert.Equal(tCase.Result.storage, reflect.TypeOf(lm.storage).String())
			} else {
				assert.Nil(lm)
				assert.ErrorContains(err, tCase.Error)
			}
		})
	}
}

func TestReserve(t *testing.T) {
	storageCfg := StorageConfig{
		Type: "memory",
	}
	lm, err := NewManager(NewDefaultGroups(), storageCfg)

	assert := assert.New(t)

	if !assert.Nil(err) {
		t.FailNow()
	}

	ok, err := lm.Reserve("unknown", "somebody")
	assert.False(ok)
	assert.Equal("*errors.ErrorUnknownGroup", reflect.TypeOf(err).String())

	ok, err = lm.Reserve("default", "")
	assert.False(ok)
	assert.Equal("errors.ErrorEmptyID", reflect.TypeOf(err).String())
}

func TestRelease(t *testing.T) {
	storageCfg := StorageConfig{
		Type: "memory",
	}
	lm, err := NewManager(NewDefaultGroups(), storageCfg)

	assert := assert.New(t)

	if !assert.Nil(err) {
		t.FailNow()
	}

	err = lm.Release("unknown", "somebody")
	assert.Equal("*errors.ErrorUnknownGroup", reflect.TypeOf(err).String())

	err = lm.Release("default", "")
	assert.Equal("errors.ErrorEmptyID", reflect.TypeOf(err).String())
}
