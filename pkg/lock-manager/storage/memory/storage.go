package memory

import (
	"time"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/errors"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/types"
)

type MemoryBackend struct {
	// This map will be assumed to be read-only after creation
	groups map[string]*group
}

type group struct {
	slots []lock
}

type lock struct {
	id      string
	created time.Time
}

const initialArraySize = 10

func NewMemoryBackend(groups []string) *MemoryBackend {
	g := make(map[string]*group)

	for _, groupName := range groups {
		// Pre-reserve some slots
		g[groupName] = &group{
			slots: make([]lock, 0, initialArraySize),
		}
	}

	return &MemoryBackend{
		groups: g,
	}
}

// Reserve a lock for the given group.
// Returns true if the lock is successfully reserved, even if the lock is already held by the specific id
func (m *MemoryBackend) Reserve(group string, id string) error {
	g := m.groups[group]
	if g == nil {
		// All groups should be initialized at the beginning
		return errors.NewErrorUnknownGroup(group)
	}

	if g.hasLock(id) {
		return nil
	}

	lock := lock{
		id:      id,
		created: time.Now(),
	}
	g.slots = append(g.slots, lock)

	return nil
}

// Returns the current number of locks for the given group
func (m *MemoryBackend) GetLocks(group string) (int, error) {
	g := m.groups[group]
	if g == nil {
		return 0, nil
	}

	return len(g.slots), nil
}

// Release the lock currently held by the id.
// Does not fail when no lock is held.
func (m *MemoryBackend) Release(group string, id string) error {
	g := m.groups[group]
	if g == nil {
		return nil
	}

	var i int
	var l lock
	for i, l = range g.slots {
		if l.id == id {
			break
		}
	}
	if len(g.slots) > 1 {
		g.slots[i] = g.slots[len(g.slots)-1]
		g.slots = g.slots[:len(g.slots)-1]
	} else {
		g.slots = g.slots[:0]
	}

	return nil
}

// Return all locks older than x
func (m *MemoryBackend) GetStaleLocks(ts time.Duration) ([]types.Lock, error) {
	panic("TODO")
}

// Check if a given id already has a lock for this group
func (m *MemoryBackend) HasLock(group, id string) (bool, error) {
	g := m.groups[group]
	if g == nil {
		// All groups should be initialized at the beginning
		return false, errors.NewErrorUnknownGroup(group)
	}
	return g.hasLock(id), nil
}

// Calls all necessary finalization if necessary
func (m *MemoryBackend) Close() error {
	return nil
}

func (g *group) hasLock(id string) bool {
	for _, lock := range g.slots {
		if id == lock.id {
			return true
		}
	}
	return false
}
