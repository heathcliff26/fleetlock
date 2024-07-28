package lockmanager

import (
	"sync"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/errors"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/etcd"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/kubernetes"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/memory"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/redis"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/sql"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/types"
)

type LockManager struct {
	groups  map[string]*lockGroup
	storage StorageBackend
}

type lockGroup struct {
	Config GroupConfig
	RWLock sync.RWMutex
}

// It is assumed that each group itself is multi-read, single-write.
// There can be multiple writes to different groups happening in parallel though.
type StorageBackend interface {
	// Reserve a lock for the given group.
	// Returns true if the lock is successfully reserved, even if the lock is already held by the specific id
	Reserve(group, id string) error
	// Returns the current number of locks for the given group
	GetLocks(group string) (int, error)
	// Release the lock currently held by the id.
	// Does not fail when no lock is held.
	Release(group, id string) error
	// Return all locks older than x
	GetStaleLocks(ts time.Duration) ([]types.Lock, error)
	// Check if a given id already has a lock for this group
	HasLock(group, id string) (bool, error)
	// Calls all necessary finalization if necessary
	Close() error
}

// Create a new LockManager from the given configuration
func NewManager(groups Groups, storageCfg *StorageConfig) (*LockManager, error) {
	var storage StorageBackend
	var err error
	switch storageCfg.Type {
	case "memory":
		i := 0
		groupNames := make([]string, len(groups))
		for k := range groups {
			groupNames[i] = k
			i++
		}
		storage = memory.NewMemoryBackend(groupNames)
	case "sqlite":
		storage, err = sql.NewSQLiteBackend(storageCfg.SQLite)
	case "postgres":
		storage, err = sql.NewPostgresBackend(storageCfg.Postgres)
	case "mysql":
		storage, err = sql.NewMySQLBackend(storageCfg.MySQL)
	case "redis":
		storage, err = redis.NewRedisBackend(storageCfg.Redis)
	case "etcd":
		storage, err = etcd.NewEtcdBackend(storageCfg.Etcd)
	case "kubernetes":
		storage, err = kubernetes.NewKubernetesBackend(storageCfg.Kubernetes)
	default:
		err = errors.NewErrorUnkownStorageType(storageCfg.Type)
	}
	if err != nil {
		return nil, err
	}

	return &LockManager{
		groups:  initGroups(groups),
		storage: storage,
	}, nil
}

// Create a new LockManager with custom StorageBackend
func NewManagerWithStorage(groups Groups, storage StorageBackend) *LockManager {
	return &LockManager{
		groups:  initGroups(groups),
		storage: storage,
	}
}

func initGroups(groups Groups) map[string]*lockGroup {
	g := make(map[string]*lockGroup, len(groups))
	for name, cfg := range groups {
		g[name] = &lockGroup{
			Config: cfg,
		}
	}
	return g
}

func (lm *LockManager) Reserve(group, id string) (bool, error) {
	lGroup := lm.groups[group]
	if lGroup == nil {
		return false, errors.NewErrorUnknownGroup(group)
	}
	if id == "" {
		return false, errors.ErrorEmptyID{}
	}

	checkHasLock := func() (bool, error) {
		// Lock group for reading to ensure that no writing is happening during it and result is accurate
		lGroup.RWLock.RLock()
		defer lGroup.RWLock.RUnlock()

		return lm.storage.HasLock(group, id)
	}
	ok, err := checkHasLock()
	if ok && err == nil {
		return true, nil
	} else if err != nil {
		return false, err
	}

	// Use own function to ensure lock is released
	checkAvailableSlots := func() (bool, error) {
		// Lock group for reading to ensure that no writing is happening during it and result is accurate
		lGroup.RWLock.RLock()
		defer lGroup.RWLock.RUnlock()

		return lm.checkSlots(group, lGroup.Config)
	}
	ok, err = checkAvailableSlots()
	if err != nil || !ok {
		return false, err
	}

	// Get Write Lock
	lGroup.RWLock.Lock()
	defer lGroup.RWLock.Unlock()

	// Re-check, since another write could have happened between checking the first time and now
	ok, err = lm.checkSlots(group, lGroup.Config)
	if err != nil || !ok {
		return false, err
	}

	err = lm.storage.Reserve(group, id)
	return err == nil, err
}

func (lm *LockManager) checkSlots(group string, cfg GroupConfig) (bool, error) {
	usedSlots, err := lm.storage.GetLocks(group)
	if err != nil {
		return false, err
	}

	return usedSlots < cfg.Slots, nil
}

func (lm *LockManager) Release(group, id string) error {
	lGroup := lm.groups[group]
	if lGroup == nil {
		return errors.NewErrorUnknownGroup(group)
	}
	if id == "" {
		return errors.ErrorEmptyID{}
	}

	lGroup.RWLock.Lock()
	defer lGroup.RWLock.Unlock()

	return lm.storage.Release(group, id)
}

func (lm *LockManager) Close() error {
	return lm.storage.Close()
}
