package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPhotoAlbum(t *testing.T) {
	t.Run("new album", func(t *testing.T) {
		m := NewPhotoAlbum("ABC", "EFG")
		assert.Equal(t, "ABC", m.PhotoUUID)
		assert.Equal(t, "EFG", m.AlbumUUID)
	})
}

func TestPhotoAlbum_TableName(t *testing.T) {
	photoAlbum := &PhotoAlbum{}
	tableName := photoAlbum.TableName()

	assert.Equal(t, "photos_albums", tableName)
}

func TestPhotoAlbum_FirstOrCreate(t *testing.T) {
	m := PhotoAlbumFixtures.Get("1", "pt9jtdre2lvl0yh7", "at9lxuqxpogaaba8")
	r := m.FirstOrCreate()
	assert.Equal(t, "at9lxuqxpogaaba8", r.AlbumUUID)
}
