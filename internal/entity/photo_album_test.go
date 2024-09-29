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
	t.Run("existing photo_album", func(t *testing.T) {
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

	t.Run("not yet existing photo_album", func(t *testing.T) {
		newPhoto := NewPhoto(false)
		newPhoto.ID = 56789 // Can't add details if there isn't a photo in the database.
		Db().Create(&newPhoto)
		newAlbum := &Album{ID: 56789} // Can't add details if there isn't a photo in the database.
		Db().Create(newAlbum)

		model := &PhotoAlbum{PhotoUID: newPhoto.PhotoUID, AlbumUID: newAlbum.AlbumUID}
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
		UnscopedDb().Delete(model)
		UnscopedDb().Delete(newAlbum)
		UnscopedDb().Delete(&newPhoto)
	})
}

func TestPhotoAlbum_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		newPhoto := NewPhoto(false)
		newPhoto.ID = 56786 // Can't add details if there isn't a photo in the database.
		Db().Create(&newPhoto)
		newAlbum := &Album{ID: 56783}
		Db().Create(newAlbum)

		p := PhotoAlbum{PhotoUID: newPhoto.PhotoUID, AlbumUID: newAlbum.AlbumUID} // Prevent Unique Constraint violation.

		err := p.Create()

		if err != nil {
			t.Fatal(err)
		}
		// Cleanup
		result := UnscopedDb().Model(PhotoAlbum{}).Delete(p)
		assert.Equal(t, int64(1), result.RowsAffected)
		UnscopedDb().Delete(newAlbum)
		UnscopedDb().Delete(&newPhoto)
	})
}
