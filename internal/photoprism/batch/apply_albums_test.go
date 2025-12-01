package batch

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
)

// TestApplyAlbums exercises batch action logic.
func TestApplyAlbums(t *testing.T) {
	t.Run("AddPhotoToExistingAlbumByUID", func(t *testing.T) {
		photo := entity.PhotoFixtures.Get("Photo01")
		albumUID := entity.AlbumFixtures.Get("christmas2030").AlbumUID

		albums := Items{
			Items: []Item{
				{Action: ActionAdd, Value: albumUID},
			},
		}

		if err := ApplyAlbums(&photo, albums); err != nil {
			t.Fatal(err)
		}

		// Verify photo was added to album by checking photos_albums table
		var photoAlbum entity.PhotoAlbum
		if err := entity.Db().Where("album_uid = ? AND photo_uid = ? AND hidden = ?", albumUID, photo.PhotoUID, false).First(&photoAlbum).Error; err != nil {
			t.Fatal(err)
		}
	})
	t.Run("AddPhotoToNewAlbumByTitle", func(t *testing.T) {
		photo := entity.PhotoFixtures.Get("Photo02")
		albumTitle := "Test Album for Actions"

		albums := Items{
			Items: []Item{
				{Action: ActionAdd, Title: albumTitle},
			},
		}

		if err := ApplyAlbums(&photo, albums); err != nil {
			t.Fatal(err)
		}

		// Verify album was created and photo added
		var album entity.Album
		if err := entity.Db().Where("album_title = ?", albumTitle).First(&album).Error; err != nil {
			t.Fatal(err)
		}

		if album.AlbumTitle != albumTitle {
			t.Errorf("expected album title %s, got %s", albumTitle, album.AlbumTitle)
		}

		var photoAlbum entity.PhotoAlbum
		if err := entity.Db().Where("album_uid = ? AND photo_uid = ? AND hidden = ?", album.AlbumUID, photo.PhotoUID, false).First(&photoAlbum).Error; err != nil {
			t.Fatal(err)
		}
	})
	t.Run("RemovePhotoFromAlbum", func(t *testing.T) {
		// First add photo to album
		photo := entity.PhotoFixtures.Get("Photo03")
		album := entity.AlbumFixtures.Get("holiday-2030")

		// Create photo-album relation manually
		var existing entity.PhotoAlbum
		err := entity.Db().Where("album_uid = ? AND photo_uid = ?", album.AlbumUID, photo.PhotoUID).First(&existing).Error
		if err != nil {
			photoAlbumEntry := entity.PhotoAlbum{
				PhotoUID: photo.PhotoUID,
				AlbumUID: album.AlbumUID,
				Hidden:   false,
			}

			if err = entity.Db().Create(&photoAlbumEntry).Error; err != nil {
				t.Fatal(err)
			}
		} else if existing.Hidden {
			existing.Hidden = false
			if err = entity.Db().Save(&existing).Error; err != nil {
				t.Fatal(err)
			}
		}

		// Verify photo is in album
		var checkEntry entity.PhotoAlbum
		if err = entity.Db().Where("album_uid = ? AND photo_uid = ? AND hidden = ?", album.AlbumUID, photo.PhotoUID, false).First(&checkEntry).Error; err != nil {
			t.Fatal(err)
		}

		// Now remove it
		albums := Items{
			Items: []Item{
				{Action: ActionRemove, Value: album.AlbumUID},
			},
		}

		if errs := ApplyAlbums(&photo, albums); errs != nil {
			t.Fatal(errs)
		}

		// Verify photo was removed (should be marked as hidden)
		var removedEntry entity.PhotoAlbum
		if err = entity.Db().Where("album_uid = ? AND photo_uid = ?", album.AlbumUID, photo.PhotoUID).First(&removedEntry).Error; err != nil {
			t.Fatal(err)
		}

		if !removedEntry.Hidden {
			t.Error("expected photo to be marked as hidden in album")
		}
	})
	// Error cases
	t.Run("AddPhotoToNonExistingAlbumByUID", func(t *testing.T) {
		photo := entity.PhotoFixtures.Get("Photo04")
		nonExistingAlbumUID := "at9lxuqxpoaaaaaa" // Invalid/non-existing UID

		albums := Items{
			Items: []Item{
				{Action: ActionAdd, Value: nonExistingAlbumUID},
			},
		}

		err := ApplyAlbums(&photo, albums)
		if err == nil {
			t.Error("expected error when adding photo to non-existing album, but got none")
		}
	})
	t.Run("AddPhotoToAlbumWithInvalidUID", func(t *testing.T) {
		photo := entity.PhotoFixtures.Get("Photo04")
		invalidUID := "invalid-uid-format" // Invalid UID format

		albums := Items{
			Items: []Item{
				{Action: ActionAdd, Value: invalidUID},
			},
		}

		err := ApplyAlbums(&photo, albums)
		if err == nil {
			t.Error("expected error when adding photo to album with invalid UID, but got none")
		}
	})
	t.Run("RemovePhotoFromNonExistingAlbum", func(t *testing.T) {
		photo := entity.PhotoFixtures.Get("Photo05")
		nonExistingAlbumUID := "at9lxuqxpobbbbbb" // Non-existing UID

		albums := Items{
			Items: []Item{
				{Action: ActionRemove, Value: nonExistingAlbumUID},
			},
		}

		err := ApplyAlbums(&photo, albums)
		if err == nil {
			t.Error("expected error when removing photo from non-existing album, but got none")
		}
	})
	t.Run("InvalidActionOnAlbum", func(t *testing.T) {
		photo := entity.PhotoFixtures.Get("Photo06")
		albumUID := entity.AlbumFixtures.Get("christmas2030").AlbumUID

		albums := Items{
			Items: []Item{
				{Action: "invalid-action", Value: albumUID}, // Invalid action
			},
		}

		err := ApplyAlbums(&photo, albums)
		if err == nil {
			t.Error("expected error for invalid action, but got none")
		}
	})
	t.Run("EmptyAlbumItems", func(t *testing.T) {
		photo := entity.PhotoFixtures.Get("Photo07")

		albums := Items{
			Items: []Item{}, // Empty items
		}

		// This should not error, but should be a no-op
		err := ApplyAlbums(&photo, albums)
		if err != nil {
			t.Errorf("expected no error for empty album items, but got: %v", err)
		}
	})
	t.Run("AddPhotoToAlbumWithEmptyValueAndTitle", func(t *testing.T) {
		photo := entity.PhotoFixtures.Get("Photo08")

		albums := Items{
			Items: []Item{
				{Action: ActionAdd, Value: "", Title: ""}, // Both empty
			},
		}

		err := ApplyAlbums(&photo, albums)
		if err == nil {
			t.Error("expected error when both Value and Title are empty, but got none")
		}
	})
	t.Run("InvalidPhotoUID", func(t *testing.T) {
		invalidPhotoUID := "invalid-photo-uid"
		albumUID := entity.AlbumFixtures.Get("christmas2030").AlbumUID

		albums := Items{
			Items: []Item{
				{Action: ActionAdd, Value: albumUID},
			},
		}

		if err := ApplyAlbums(&entity.Photo{PhotoUID: invalidPhotoUID}, albums); err == nil {
			t.Error("expected error for invalid photo UID, but got none")
		}
	})
}
