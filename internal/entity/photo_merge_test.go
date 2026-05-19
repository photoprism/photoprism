package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestPhoto_Stackable(t *testing.T) {
	t.Run("IsStackable", func(t *testing.T) {
		m := Photo{ID: 1, PhotoUID: "pr32t8j3feogit2t", PhotoName: "foo", PhotoStack: IsStackable, TakenAt: Now(), TakenAtLocal: time.Time{}, TakenSrc: SrcMeta, TimeZone: "Europe/Berlin"}
		assert.True(t, m.Stackable())
	})
	t.Run("IsStacked", func(t *testing.T) {
		m := Photo{ID: 1, PhotoUID: "pr32t8j3feogit2t", PhotoName: "foo", PhotoStack: IsStacked, TakenAt: Now(), TakenAtLocal: time.Time{}, TakenSrc: SrcMeta, TimeZone: "Europe/Berlin"}
		assert.True(t, m.Stackable())
	})
	t.Run("NoName", func(t *testing.T) {
		m := Photo{ID: 1, PhotoUID: "pr32t8j3feogit2t", PhotoName: "", TakenAt: time.Time{}, TakenAtLocal: Now(), TakenSrc: SrcMeta, TimeZone: "Europe/Berlin"}
		assert.False(t, m.Stackable())
	})
	t.Run("IsUnstacked", func(t *testing.T) {
		m := Photo{ID: 1, PhotoUID: "pr32t8j3feogit2t", PhotoName: "foo", PhotoStack: IsUnstacked, TakenAt: Now(), TakenAtLocal: time.Time{}, TakenSrc: SrcMeta, TimeZone: "Europe/Berlin"}
		assert.False(t, m.Stackable())
	})
	t.Run("NoID", func(t *testing.T) {
		m := Photo{ID: 0, PhotoUID: "pr32t8j3feogit2t", PhotoName: "foo", PhotoStack: IsStacked, TakenAt: Now(), TakenAtLocal: time.Time{}, TakenSrc: SrcMeta, TimeZone: "Europe/Berlin"}
		assert.False(t, m.Stackable())
	})
	t.Run("NoPhotoUID", func(t *testing.T) {
		m := Photo{ID: 1, PhotoUID: "", PhotoName: "foo", PhotoStack: IsStacked, TakenAt: Now(), TakenAtLocal: time.Time{}, TakenSrc: SrcMeta, TimeZone: "Europe/Berlin"}
		assert.False(t, m.Stackable())
	})
}

func TestPhoto_IdenticalIdentical(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo19")

		result, err := photo.Identical(true, true)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("result: %#v", result)
		assert.Equal(t, 1, len(result))
	})
	t.Run("UnstackedPhoto", func(t *testing.T) {
		photo := &Photo{PhotoStack: IsUnstacked, PhotoName: "testName"}

		result, err := photo.Identical(true, true)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(result))
	})
	t.Run("Success", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo23")

		result, err := photo.Identical(true, true)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("result: %#v", result)
		assert.Equal(t, 2, len(result))
	})
	t.Run("Success", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo23")
		result, err := photo.Identical(true, false)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("result: %#v", result)
		assert.Equal(t, 2, len(result))
	})
}

func TestPhoto_Merge(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo23")
		original, merged, err := photo.Merge(true, false)

		if err != nil {
			t.Fatal(err)
		}
		assert.EqualValues(t, 1000023, original.ID)
		assert.EqualValues(t, 1000024, merged[0].ID)
	})
	t.Run("SyncVideoTypeFromMergedFiles", func(t *testing.T) {
		takenAt := time.Date(2026, 1, 21, 12, 34, 56, 0, time.UTC)

		imagePhoto := NewPhoto(true)
		imagePhoto.PhotoUID = rnd.GenerateUID(PhotoUID)
		imagePhoto.PhotoPath = "merge"
		imagePhoto.PhotoName = "PhotoMerge"
		imagePhoto.PhotoType = MediaImage
		imagePhoto.TypeSrc = SrcAuto
		imagePhoto.PhotoQuality = 5
		imagePhoto.TakenAt = takenAt
		imagePhoto.TakenAtLocal = takenAt
		imagePhoto.TakenSrc = SrcMeta
		imagePhoto.PlaceSrc = SrcMeta
		imagePhoto.CellID = UnknownLocation.ID
		imagePhoto.CameraID = UnknownCamera.ID
		imagePhoto.CameraSerial = "merge-camera"

		if err := imagePhoto.Create(); err != nil {
			t.Fatal(err)
		}

		videoPhoto := NewPhoto(true)
		videoPhoto.PhotoUID = rnd.GenerateUID(PhotoUID)
		videoPhoto.PhotoPath = "merge"
		videoPhoto.PhotoName = "PhotoMerge"
		videoPhoto.PhotoType = MediaVideo
		videoPhoto.TypeSrc = SrcAuto
		videoPhoto.PhotoQuality = 4
		videoPhoto.TakenAt = takenAt
		videoPhoto.TakenAtLocal = takenAt
		videoPhoto.TakenSrc = SrcMeta
		videoPhoto.PlaceSrc = SrcMeta
		videoPhoto.CellID = UnknownLocation.ID
		videoPhoto.CameraID = UnknownCamera.ID
		videoPhoto.CameraSerial = "merge-camera"

		if err := videoPhoto.Create(); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			photoIDs := []uint{imagePhoto.ID, videoPhoto.ID}

			_ = UnscopedDb().Where("photo_id IN (?)", photoIDs).Delete(File{}).Error
			_ = UnscopedDb().Where("photo_id IN (?)", photoIDs).Delete(Details{}).Error
			_ = UnscopedDb().Where("id IN (?)", photoIDs).Delete(Photo{}).Error
		})

		imageFile := File{
			PhotoID:      imagePhoto.ID,
			PhotoUID:     imagePhoto.PhotoUID,
			FileUID:      rnd.GenerateUID(FileUID),
			FileName:     "merge/" + imagePhoto.PhotoUID + ".jpg",
			FileRoot:     RootOriginals,
			FileHash:     "merge-image-" + imagePhoto.PhotoUID,
			FileType:     "jpg",
			MediaType:    media.Image.String(),
			FilePrimary:  true,
			FileMissing:  false,
			FileSidecar:  false,
			FileDuration: 0,
		}

		if err := imageFile.Create(); err != nil {
			t.Fatal(err)
		}

		videoFile := File{
			PhotoID:      videoPhoto.ID,
			PhotoUID:     videoPhoto.PhotoUID,
			FileUID:      rnd.GenerateUID(FileUID),
			FileName:     "merge/" + videoPhoto.PhotoUID + ".mp4",
			FileRoot:     RootOriginals,
			FileHash:     "merge-video-" + videoPhoto.PhotoUID,
			FileType:     "mp4",
			MediaType:    media.Video.String(),
			FilePrimary:  true,
			FileVideo:    true,
			FileMissing:  false,
			FileSidecar:  false,
			FileDuration: time.Second * 8,
		}

		if err := videoFile.Create(); err != nil {
			t.Fatal(err)
		}

		original, merged, err := videoPhoto.Merge(true, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, imagePhoto.ID, original.ID)
		assert.Equal(t, 1, len(merged))

		var refreshed Photo

		if err := Db().First(&refreshed, "id = ?", original.ID).Error; err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, MediaVideo, refreshed.PhotoType)
		assert.Equal(t, SrcFile, refreshed.TypeSrc)
	})
}

func TestPhoto_SyncMediaTypeFromFiles(t *testing.T) {
	t.Run("NoMainFiles", func(t *testing.T) {
		photo := NewPhoto(true)
		photo.PhotoUID = rnd.GenerateUID(PhotoUID)
		photo.PhotoPath = "merge"
		photo.PhotoName = "NoMainFiles"
		photo.PhotoType = MediaVideo
		photo.TypeSrc = SrcAuto

		if err := photo.Create(); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			_ = UnscopedDb().Where("photo_id = ?", photo.ID).Delete(File{}).Error
			_ = UnscopedDb().Where("photo_id = ?", photo.ID).Delete(Details{}).Error
			_ = UnscopedDb().Where("id = ?", photo.ID).Delete(Photo{}).Error
		})

		sidecar := File{
			PhotoID:     photo.ID,
			PhotoUID:    photo.PhotoUID,
			FileUID:     rnd.GenerateUID(FileUID),
			FileName:    "merge/" + photo.PhotoUID + ".xmp",
			FileRoot:    RootOriginals,
			FileHash:    "merge-sidecar-" + photo.PhotoUID,
			FileType:    "xmp",
			MediaType:   media.Sidecar.String(),
			FilePrimary: false,
			FileMissing: false,
			FileSidecar: true,
		}

		if err := sidecar.Create(); err != nil {
			t.Fatal(err)
		}

		if err := photo.SyncMediaTypeFromFiles(SrcFile); err != nil {
			t.Fatal(err)
		}

		var refreshed Photo

		if err := Db().First(&refreshed, "id = ?", photo.ID).Error; err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, MediaVideo, refreshed.PhotoType)
		assert.Equal(t, SrcAuto, refreshed.TypeSrc)
	})
	t.Run("PreservesManualOverride", func(t *testing.T) {
		photo := NewPhoto(true)
		photo.PhotoUID = rnd.GenerateUID(PhotoUID)
		photo.PhotoPath = "merge"
		photo.PhotoName = "ManualOverride"
		photo.PhotoType = MediaImage
		photo.TypeSrc = SrcManual

		if err := photo.Create(); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			_ = UnscopedDb().Where("photo_id = ?", photo.ID).Delete(File{}).Error
			_ = UnscopedDb().Where("photo_id = ?", photo.ID).Delete(Details{}).Error
			_ = UnscopedDb().Where("id = ?", photo.ID).Delete(Photo{}).Error
		})

		video := File{
			PhotoID:      photo.ID,
			PhotoUID:     photo.PhotoUID,
			FileUID:      rnd.GenerateUID(FileUID),
			FileName:     "merge/" + photo.PhotoUID + ".mp4",
			FileRoot:     RootOriginals,
			FileHash:     "merge-manual-video-" + photo.PhotoUID,
			FileType:     "mp4",
			MediaType:    media.Video.String(),
			FilePrimary:  true,
			FileMissing:  false,
			FileSidecar:  false,
			FileVideo:    true,
			FileDuration: time.Second * 3,
		}

		if err := video.Create(); err != nil {
			t.Fatal(err)
		}

		if err := photo.SyncMediaTypeFromFiles(SrcFile); err != nil {
			t.Fatal(err)
		}

		var refreshed Photo

		if err := Db().First(&refreshed, "id = ?", photo.ID).Error; err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, MediaImage, refreshed.PhotoType)
		assert.Equal(t, SrcManual, refreshed.TypeSrc)
	})
}
