package lockmanager

import (
	"os"
	"reflect"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/redis"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/sqlite"
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
		Storage *StorageConfig
		Result  result
		Error   string
	}{
		{
			Name: "MemoryBackend",
			Storage: &StorageConfig{
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
			Storage: &StorageConfig{
				Type: "sqlite",
				SQLite: &sqlite.SQLiteConfig{
					File: "test.db",
				},
			},
			Result: result{
				groups:  initGroups(NewDefaultGroups()),
				storage: "*sqlite.SQLBackend",
			},
			Error: "",
		},
		{
			Name: "ErrorNewSQLiteBackend",
			Storage: &StorageConfig{
				Type: "sqlite",
				SQLite: &sqlite.SQLiteConfig{
					File: "/not/a/valid/path/to/database",
				},
			},
			Error: "sqlite3.Error",
		},
		{
			Name: "RedisBackend",
			Storage: &StorageConfig{
				Type: "redis",
				Redis: &redis.RedisConfig{
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
			Storage: &StorageConfig{
				Type: "redis",
				Redis: &redis.RedisConfig{
					Addr: "",
				},
			},
			Error: "*net.OpError",
		},
		{
			Name: "UnknownStorageType",
			Storage: &StorageConfig{
				Type: "not-a-valid-type",
			},
			Error: "*errors.ErrorUnkownStorageType",
		},
	}
	t.Cleanup(func() {
		err := os.Remove("test.db")
		if err != nil {
			t.Logf("Failed to cleanup sqlite database file: %v", err)
		}
	})

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			lm, err := NewManager(NewDefaultGroups(), tCase.Storage)

			if tCase.Error == "" {
				assert.Nil(err)
				if !assert.NotNil(lm) {
					t.FailNow()
				}
				assert.Equal(tCase.Result.groups, lm.groups)
				assert.Equal(tCase.Result.storage, reflect.TypeOf(lm.storage).String())
			} else {
				assert.Nil(lm)
				assert.Equal(tCase.Error, reflect.TypeOf(err).String())
			}
		})
	}
}

func TestReserve(t *testing.T) {
	storageCfg := StorageConfig{
		Type: "memory",
	}
	lm, err := NewManager(NewDefaultGroups(), &storageCfg)

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
	lm, err := NewManager(NewDefaultGroups(), &storageCfg)

	assert := assert.New(t)

	if !assert.Nil(err) {
		t.FailNow()
	}

	err = lm.Release("unknown", "somebody")
	assert.Equal("*errors.ErrorUnknownGroup", reflect.TypeOf(err).String())

	err = lm.Release("default", "")
	assert.Equal("errors.ErrorEmptyID", reflect.TypeOf(err).String())
}
