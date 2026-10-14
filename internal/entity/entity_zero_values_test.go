package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestZeroValues_DefaultTagged pins that each model saves the Go zero value of a field whose column
// has a non-zero default, through the write path production code uses for it: its Save method, an
// update with a values map, or both.
func TestZeroValues_DefaultTagged(t *testing.T) {
	t.Run("Album", func(t *testing.T) {
		m := NewAlbum("Zero Values Album", AlbumManual)
		m.AlbumCountry = "de"
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

		m.AlbumCountry = ""
		require.NoError(t, m.Save())

		found := Album{}
		require.NoError(t, UnscopedDb().Where("album_uid = ?", m.AlbumUID).First(&found).Error)
		assert.Equal(t, "", found.AlbumCountry)

		require.NoError(t, UnscopedDb().Model(m).Updates(Values{"album_type": ""}).Error)
		require.NoError(t, UnscopedDb().Where("album_uid = ?", m.AlbumUID).First(&found).Error)
		assert.Equal(t, "", found.AlbumType)
	})
	t.Run("Photo", func(t *testing.T) {
		m := &Photo{PhotoUID: rnd.GenerateUID(PhotoUID), TimeZone: "Europe/Berlin", PhotoCountry: "de"}
		createShapePhoto(t, m)

		m.TimeZone = ""
		m.PhotoCountry = ""
		m.CameraID = 0
		m.LensID = 0
		require.NoError(t, m.Save())

		found, err := findShapePhoto(m.PhotoUID)
		require.NoError(t, err)
		assert.Equal(t, "", found.TimeZone)
		assert.Equal(t, "", found.PhotoCountry)
		assert.Equal(t, uint(0), found.CameraID)
		assert.Equal(t, uint(0), found.LensID)

		require.NoError(t, m.Updates(Values{"photo_type": "", "place_id": "", "cell_id": ""}))
		found, err = findShapePhoto(m.PhotoUID)
		require.NoError(t, err)
		assert.Equal(t, "", found.PhotoType)
		assert.Equal(t, "", found.PlaceID)
		assert.Equal(t, "", found.CellID)
	})
	t.Run("File", func(t *testing.T) {
		photo := &Photo{PhotoUID: rnd.GenerateUID(PhotoUID)}
		createShapePhoto(t, photo)

		m := &File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileUID: rnd.GenerateUID(FileUID), FileName: "zero/root.jpg", FileRoot: RootSidecar, FileHash: rnd.GenerateUID('h')}
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Unscoped().Delete(m).Error })

		m.FileRoot = ""
		require.NoError(t, m.Save())

		found := File{}
		require.NoError(t, UnscopedDb().Where("file_uid = ?", m.FileUID).First(&found).Error)
		assert.Equal(t, "", found.FileRoot)
	})
	t.Run("Label", func(t *testing.T) {
		m := NewLabel("Zero Values Label "+rnd.Base36(6), 0)
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Unscoped().Delete(m).Error })

		m.PhotoCount = 0
		require.NoError(t, m.Save())

		found := Label{}
		require.NoError(t, UnscopedDb().Where("id = ?", m.ID).First(&found).Error)
		assert.Equal(t, 0, found.PhotoCount)
	})
	t.Run("Place", func(t *testing.T) {
		m := &Place{ID: "zero-" + rnd.Base36(8), PlaceLabel: "Zero Values", PhotoCount: 5}
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

		m.PhotoCount = 0
		require.NoError(t, m.Save())

		found := Place{}
		require.NoError(t, UnscopedDb().Where("id = ?", m.ID).First(&found).Error)
		assert.Equal(t, 0, found.PhotoCount)
	})
	t.Run("Cell", func(t *testing.T) {
		m := &Cell{ID: "zero-" + rnd.Base36(8), PlaceID: UnknownPlace.ID}
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

		// Save leaves a blank foreign key with a default tag out of the update, so the stored place stays.
		m.PlaceID = ""
		require.NoError(t, m.Save())

		found := Cell{}
		require.NoError(t, UnscopedDb().Where("id = ?", m.ID).First(&found).Error)
		assert.Equal(t, UnknownPlace.ID, found.PlaceID)

		// A values map writes it.
		require.NoError(t, UnscopedDb().Model(m).Updates(Values{"place_id": ""}).Error)
		require.NoError(t, UnscopedDb().Where("id = ?", m.ID).First(&found).Error)
		assert.Equal(t, "", found.PlaceID)
	})
	t.Run("UserDetails", func(t *testing.T) {
		m := NewUserDetails(rnd.GenerateUID(UserUID))
		m.BirthYear, m.BirthMonth, m.BirthDay = 1990, 5, 17
		m.UserCountry = "de"
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

		m.BirthYear, m.BirthMonth, m.BirthDay = 0, 0, 0
		m.UserCountry = ""
		m.PlaceID, m.CellID = "", ""
		require.NoError(t, m.Save())

		found := UserDetails{}
		require.NoError(t, UnscopedDb().Where("user_uid = ?", m.UserUID).First(&found).Error)
		assert.Equal(t, 0, found.BirthYear)
		assert.Equal(t, 0, found.BirthMonth)
		assert.Equal(t, 0, found.BirthDay)
		assert.Equal(t, "", found.UserCountry)
		assert.Equal(t, "", found.PlaceID)
		assert.Equal(t, "", found.CellID)

		m.PlaceID, m.CellID = "de", "de"
		require.NoError(t, m.Save())
		require.NoError(t, m.Updates(Values{"place_id": "", "cell_id": ""}))
		require.NoError(t, UnscopedDb().Where("user_uid = ?", m.UserUID).First(&found).Error)
		assert.Equal(t, "", found.PlaceID)
		assert.Equal(t, "", found.CellID)
	})
	t.Run("UserSettings", func(t *testing.T) {
		m := NewUserSettings(rnd.GenerateUID(UserUID))
		m.UIStartPage = "browse"
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

		m.UIStartPage = ""
		require.NoError(t, m.Save())

		found := UserSettings{}
		require.NoError(t, UnscopedDb().Where("user_uid = ?", m.UserUID).First(&found).Error)
		assert.Equal(t, "", found.UIStartPage)
	})
	t.Run("Marker", func(t *testing.T) {
		m := MarkerFixtures.Get("1000003-4")
		require.NotEmpty(t, m.MarkerUID)

		stored := Marker{}
		require.NoError(t, UnscopedDb().Where("marker_uid = ?", m.MarkerUID).First(&stored).Error)
		t.Cleanup(func() {
			_ = UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", m.MarkerUID).
				UpdateColumns(Values{"face_dist": stored.FaceDist, "size": stored.Size, "updated_at": stored.UpdatedAt}).Error
		})

		require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", m.MarkerUID).
			Updates(Values{"face_dist": 0, "size": 0}).Error)

		found := Marker{}
		require.NoError(t, UnscopedDb().Where("marker_uid = ?", m.MarkerUID).First(&found).Error)
		assert.Equal(t, float64(0), found.FaceDist)
		assert.Equal(t, 0, found.Size)

		// The struct save production code uses for markers writes them as well.
		found.FaceDist, found.Size = 0.5, 100
		require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", m.MarkerUID).
			Updates(Values{"face_dist": found.FaceDist, "size": found.Size}).Error)
		found.FaceDist, found.Size = 0, 0
		require.NoError(t, found.Save())

		require.NoError(t, UnscopedDb().Where("marker_uid = ?", m.MarkerUID).First(&found).Error)
		assert.Equal(t, float64(0), found.FaceDist)
		assert.Equal(t, 0, found.Size)
	})
	t.Run("Duplicate", func(t *testing.T) {
		// GORM v1 ignores the default of a primary key column such as file_root; v2 does not.
		m := &Duplicate{FileName: "zero/duplicate-" + rnd.Base36(8) + ".jpg", FileRoot: RootSidecar, FileHash: rnd.GenerateUID('h'), FileSize: 1, ModTime: time.Now().Unix()}
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Where("file_name = ?", m.FileName).Delete(&Duplicate{}).Error })

		require.NoError(t, UnscopedDb().Model(&Duplicate{}).Where("file_name = ?", m.FileName).Updates(Values{"file_root": ""}).Error)

		found := Duplicate{}
		require.NoError(t, UnscopedDb().Where("file_name = ?", m.FileName).First(&found).Error)
		assert.Equal(t, "", found.FileRoot)
	})
}
