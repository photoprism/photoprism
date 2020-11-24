package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileShareMap_Get(t *testing.T) {
	t.Run("get existing fileshare", func(t *testing.T) {
		r := FileShareFixtures.Get("FileShare1", 0, 0, "")
		assert.Equal(t, uint(1000000), r.AccountID)
		assert.Equal(t, "name for remote", r.RemoteName)
		assert.IsType(t, FileShare{}, r)
	})
	t.Run("get not existing fileshare", func(t *testing.T) {
		r := FileShareFixtures.Get("FileShareXXX", 123, 888, "new remote name")
		assert.Equal(t, uint(888), r.AccountID)
		assert.Equal(t, "new remote name", r.RemoteName)
		assert.IsType(t, FileShare{}, r)
	})
}

func TestFileShareMap_Pointer(t *testing.T) {
	t.Run("get existing fileshare pointer", func(t *testing.T) {
		r := FileShareFixtures.Pointer("FileShare1", 0, 0, "")
		assert.Equal(t, uint(1000000), r.AccountID)
		assert.Equal(t, "name for remote", r.RemoteName)
		assert.IsType(t, &FileShare{}, r)
	})
	t.Run("get not existing fileshare pointer", func(t *testing.T) {
		r := FileShareFixtures.Pointer("FileShareYYY", 456, 889, "new remote name for pointer")
		assert.Equal(t, uint(889), r.AccountID)
		assert.Equal(t, "new remote name for pointer", r.RemoteName)
		assert.IsType(t, &FileShare{}, r)
	})
}
