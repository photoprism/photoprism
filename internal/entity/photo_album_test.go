package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPhotoAlbum(t *testing.T) {
	t.Run("NewAlbum", func(t *testing.T) {
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
	t.Run("ExistingAlbum", func(t *testing.T) {
		model := PhotoAlbumFixtures.Get("1", "ps6sg6be2lvl0yh7", "as6sg6bxpogaaba8")
		result := FirstOrCreatePhotoAlbum(&model)

		if result == nil {
			t.Fatal("result must not be nil")
		}

		if result.AlbumUID != model.AlbumUID {
			t.Errorf("AlbumUID should be the same: %s %s", result.AlbumUID, model.AlbumUID)
		}

		if result.PhotoUID != model.PhotoUID {
			t.Errorf("PhotoUID should be the same: %s %s", result.PhotoUID, model.PhotoUID)
		}
	})
	t.Run("NotYetExistingAlbum", func(t *testing.T) {
		model := &PhotoAlbum{PhotoUID: "ps6sg6be2lvl0y14", AlbumUID: "as6sg6bipotaab29"}
		result := FirstOrCreatePhotoAlbum(model)

		if result == nil {
			t.Fatal("result must not be nil")
		}
		t.Cleanup(func() {
			require.NoError(t, UnscopedDb().Delete(&result).Error)
		})

		if result.AlbumUID != model.AlbumUID {
			t.Errorf("AlbumUID should be the same: %s %s", result.AlbumUID, model.AlbumUID)
		}

		if result.PhotoUID != model.PhotoUID {
			t.Errorf("PhotoUID should be the same: %s %s", result.PhotoUID, model.PhotoUID)
		}
	})
}

func TestPhotoAlbum_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		p := PhotoAlbum{PhotoUID: "ps6sg6be2lvl0y14", AlbumUID: "as6sg6bipogaab11"}

		err := p.Create()

		if err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			require.NoError(t, UnscopedDb().Delete(&p).Error)
		})
	})
}
