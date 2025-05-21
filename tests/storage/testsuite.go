package storage

import (
	"strconv"
	"sync"
	"testing"

	lockmanager "github.com/heathcliff26/fleetlock/pkg/lock-manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GetGroups() lockmanager.Groups {
	testGroups := make(lockmanager.Groups, 7)
	testGroups["basic"] = lockmanager.GroupConfig{
		Slots: 1,
	}
	testGroups["NoDuplicates"] = lockmanager.GroupConfig{
		Slots: 3,
	}
	testGroups["GetLocks"] = lockmanager.GroupConfig{
		Slots: 10,
	}
	testGroups["ConcurrentReserve"] = lockmanager.GroupConfig{
		Slots: 10,
	}
	testGroups["ConcurrentRelease"] = lockmanager.GroupConfig{
		Slots: 10,
	}
	testGroups["ReserveRace"] = lockmanager.GroupConfig{
		Slots: 5,
	}
	testGroups["ReserveReturnTrueIfAlreadyExists"] = lockmanager.GroupConfig{
		Slots: 1,
	}
	return testGroups
}

func RunLockManagerTestsuiteWithStorage(t *testing.T, storage lockmanager.StorageBackend) {
	lm := lockmanager.NewManagerWithStorage(GetGroups(), storage)
	t.Cleanup(func() {
		err := lm.Close()
		if err != nil {
			t.Logf("Failed to close manager: %v", err)
		}
	})

	t.Run("Basic", func(t *testing.T) {
		t.Run("Reserve", func(t *testing.T) {
			assert := assert.New(t)

			ok, err := lm.Reserve("basic", "User1")
			assert.True(ok)
			assert.Nil(err)

			ok, err = lm.Reserve("basic", "User2")
			assert.False(ok)
			assert.Nil(err)
		})
		t.Run("Release", func(t *testing.T) {
			assert := assert.New(t)

			err := lm.Release("basic", "User1")
			assert.Nil(err)

			err = lm.Release("basic", "User1")
			assert.Nil(err)

			err = lm.Release("basic", "UnkownUser")
			assert.Nil(err)
		})
	})
	t.Run("NoDuplicates", func(t *testing.T) {
		assert := assert.New(t)
		for range 10 {
			ok, err := lm.Reserve("NoDuplicates", "same-id")
			assert.True(ok)
			assert.Nil(err)

			count, err := storage.GetLocks("NoDuplicates")
			assert.Equal(1, count)
			assert.Nil(err)
		}
	})
	t.Run("GetLocks", func(t *testing.T) {
		assert := assert.New(t)

		res, err := storage.GetLocks("GetLocks")
		assert.Equal(0, res)
		assert.Nil(err)

		for i := range 10 {
			ok, err := lm.Reserve("GetLocks", "User"+strconv.Itoa(i))
			assert.True(ok)
			assert.Nil(err)

			res, err := storage.GetLocks("GetLocks")
			assert.Equal(i+1, res)
			assert.Nil(err)
		}
		for i := range 10 {
			err := lm.Release("GetLocks", "User"+strconv.Itoa(i))
			assert.Nil(err)

			res, err := storage.GetLocks("GetLocks")
			assert.Equal(9-i, res)
			assert.Nil(err)
		}
	})
	t.Run("ConcurrentReserve", func(t *testing.T) {
		assert := assert.New(t)

		var wg sync.WaitGroup
		wg.Add(10)
		for i := range 10 {
			go func() {
				defer wg.Done()
				ok, err := lm.Reserve("ConcurrentReserve", "User"+strconv.Itoa(i))
				assert.True(ok)
				assert.Nil(err)
			}()
		}
		wg.Wait()

		res, err := storage.GetLocks("ConcurrentReserve")
		assert.Equal(10, res)
		assert.Nil(err)
	})
	t.Run("ConcurrentRelease", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		for i := range 10 {
			ok, err := lm.Reserve("ConcurrentRelease", "User"+strconv.Itoa(i))
			require.True(ok)
			require.NoError(err)
		}

		var wg sync.WaitGroup
		wg.Add(10)
		for i := range 10 {
			go func() {
				defer wg.Done()
				err := lm.Release("ConcurrentRelease", "User"+strconv.Itoa(i))
				assert.Nil(err)
			}()
		}
		wg.Wait()

		res, err := storage.GetLocks("ConcurrentRelease")
		assert.Equal(0, res)
		assert.Nil(err)
	})
	t.Run("ReserveRace", func(t *testing.T) {
		assert := assert.New(t)

		result := make(chan bool, 10)
		for i := range 10 {
			go func() {
				ok, err := lm.Reserve("ReserveRace", "User"+strconv.Itoa(i))
				assert.Nil(err)
				result <- ok
			}()
		}
		count := 0
		for range 10 {
			if <-result {
				count++
			}
		}
		assert.Equal(5, count)

		count, err := storage.GetLocks("ReserveRace")
		assert.Equal(5, count)
		assert.Nil(err)
	})
	t.Run("ReserveReturnTrueIfAlreadyExists", func(t *testing.T) {
		assert := assert.New(t)

		ok, err := lm.Reserve("ReserveReturnTrueIfAlreadyExists", "User1")
		assert.True(ok)
		assert.Nil(err)

		ok, err = lm.Reserve("ReserveReturnTrueIfAlreadyExists", "User1")
		assert.True(ok)
		assert.Nil(err)
	})
}
