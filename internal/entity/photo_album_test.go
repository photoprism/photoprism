package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPhotoAlbum(t *testing.T) {
	t.Run("new album", func(t *testing.T) {
		m := NewPhotoAlbum("ABC", "EFG")
		assert.Equal(t, "ABC", m.PhotoUID)
		assert.Equal(t, "EFG", m.AlbumUID)
	})
}

func TestPhotoAlbum_TableName(t *testing.T) {
	photoAlbum := &PhotoAlbum{}
	tableName := photoAlbum.TableName()

	assert.Equal(t, "photos_albums", tableName)
}

func TestFirstOrCreatePhotoAlbum(t *testing.T) {
	t.Run("existing album", func(t *testing.T) {
		model := PhotoAlbumFixtures.Get("1", "ps6sg6be2lvl0yh7", "as6sg6bxpogaaba8")
		result := FirstOrCreatePhotoAlbum(&model)

		if result == nil {
			t.Fatal("result should not be nil")
		}

		if result.AlbumUID != model.AlbumUID {
			t.Errorf("AlbumUID should be the same: %s %s", result.AlbumUID, model.AlbumUID)
		}

		if result.PhotoUID != model.PhotoUID {
			t.Errorf("PhotoUID should be the same: %s %s", result.PhotoUID, model.PhotoUID)
		}
	})
	//TODO fails on mariadb
	t.Run("not yet existing album", func(t *testing.T) {
		model := &PhotoAlbum{}
		result := FirstOrCreatePhotoAlbum(model)

		if result == nil {
			t.Fatal("result should not be nil")
		}

		if result.AlbumUID != model.AlbumUID {
			t.Errorf("AlbumUID should be the same: %s %s", result.AlbumUID, model.AlbumUID)
		}

		if result.PhotoUID != model.PhotoUID {
			t.Errorf("PhotoUID should be the same: %s %s", result.PhotoUID, model.PhotoUID)
		}
	})
}

// TODO fails on mariadb
func TestPhotoAlbum_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		p := PhotoAlbum{}

		err := p.Create()

		if err != nil {
			t.Fatal(err)
		}
	})
}
