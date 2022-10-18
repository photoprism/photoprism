package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileSyncMap_Get(t *testing.T) {
	t.Run("get existing filesync", func(t *testing.T) {
		r := FileSyncFixtures.Get("FileSync1", 0, "")
		assert.Equal(t, uint(1000000), r.ServiceID)
		assert.Equal(t, "/20200706-092527-Landscape-München-2020.jpg", r.RemoteName)
		assert.IsType(t, FileSync{}, r)
	})
	t.Run("get not existing filesync", func(t *testing.T) {
		r := FileSyncFixtures.Get("FileSyncXXX", 123, "new remote name for sync")
		assert.Equal(t, uint(123), r.ServiceID)
		assert.Equal(t, "new remote name for sync", r.RemoteName)
		assert.IsType(t, FileSync{}, r)
	})
}

func TestFileSyncMap_Pointer(t *testing.T) {
	t.Run("get existing filesync pointer", func(t *testing.T) {
		r := FileSyncFixtures.Pointer("FileSync1", 0, "")
		assert.Equal(t, uint(1000000), r.ServiceID)
		assert.Equal(t, "/20200706-092527-Landscape-München-2020.jpg", r.RemoteName)
		assert.IsType(t, &FileSync{}, r)
	})
	t.Run("get not existing filesync pointer", func(t *testing.T) {
		r := FileSyncFixtures.Pointer("FileSyncYYY", 456, "new remote name for sync pointer")
		assert.Equal(t, uint(456), r.ServiceID)
		assert.Equal(t, "new remote name for sync pointer", r.RemoteName)
		assert.IsType(t, &FileSync{}, r)
	})
}
