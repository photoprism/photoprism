package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPhotoAlbum(t *testing.T) {
	t.Run("new album", func(t *testing.T) {
		m := NewPhotoAlbum("ABC", "EFG")
		assert.Equal(t, "ABC", m.PhotoUUID)
		assert.Equal(t, "EFG", m.AlbumUUID)
	})
}
