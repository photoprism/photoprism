package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestSelectedPhotoUIDsForSession(t *testing.T) {
	const (
		normalUID  = "ps6sg6be2lvl0yh7" // not private, not archived, not shared with guests
		privateUID = "ps6sg6be2lvl0y13" // "Photo06", private
	)
	uids := []string{normalUID, privateUID}

	t.Run("AdminSeesAll", func(t *testing.T) {
		scoped, err := SelectedPhotoUIDsForSession(uids, aclSession("alice"))
		assert.NoError(t, err)
		assert.ElementsMatch(t, uids, scoped)
	})
	t.Run("GuestExcludesPrivateAndUnshared", func(t *testing.T) {
		scoped, err := SelectedPhotoUIDsForSession(uids, aclSession("guest"))
		assert.NoError(t, err)
		assert.NotContains(t, scoped, privateUID)
		assert.NotContains(t, scoped, normalUID)
	})
	t.Run("NilSessionUnchanged", func(t *testing.T) {
		scoped, err := SelectedPhotoUIDsForSession(uids, nil)
		assert.NoError(t, err)
		assert.ElementsMatch(t, uids, scoped)
	})
	t.Run("AdminShortCircuitSkipsQuery", func(t *testing.T) {
		// Full-access sessions return the input verbatim without a scope query, so even an unknown
		// UID passes through (existence is checked by the caller's own lookup).
		in := []string{normalUID, "ps000000000unknown"}
		scoped, err := SelectedPhotoUIDsForSession(in, aclSession("alice"))
		assert.NoError(t, err)
		assert.Equal(t, in, scoped)
	})
	t.Run("EmptyInput", func(t *testing.T) {
		scoped, err := SelectedPhotoUIDsForSession(nil, aclSession("guest"))
		assert.NoError(t, err)
		assert.Empty(t, scoped)
	})
}

func TestPhotoSelection(t *testing.T) {
	albums := form.Selection{Albums: []string{"as6sg6bxpogaaba9", "as6sg6bitoga0004", "as6sg6bxpogaaba8", "as6sg6bxpogaaba7"}}

	months := form.Selection{Albums: []string{"as6sg6bipogaabj9"}}

	folders := form.Selection{Albums: []string{"as6sg6bipogaaba1", "as6sg6bipogaabj8"}}

	states := form.Selection{Albums: []string{"as6sg6bipogaab11", "as6sg6bipotaab12", "asjv2cw2eikl3cb3"}}

	t.Run("NoItemsSelected", func(t *testing.T) {
		f := form.Selection{
			Photos: []string{},
		}

		r, err := SelectedPhotos(f)

		assert.Equal(t, "no items selected", err.Error())
		assert.Empty(t, r)
	})
	t.Run("PhotosSelected", func(t *testing.T) {
		f := form.Selection{
			Photos: []string{"ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8"},
		}

		r, err := SelectedPhotos(f)

		if assert.Nil(t, err) {
			assert.Equal(t, 2, len(r))
			assert.IsType(t, entity.Photos{}, r)
		}
	})
	t.Run("FindAlbums", func(t *testing.T) {
		r, err := SelectedPhotos(albums)

		if assert.Nil(t, err) {
			assert.Equal(t, 9, len(r))
			assert.IsType(t, entity.Photos{}, r)
		}
	})
	t.Run("FindMonths", func(t *testing.T) {
		r, err := SelectedPhotos(months)

		if assert.Nil(t, err) {
			assert.Equal(t, 0, len(r))
			assert.IsType(t, entity.Photos{}, r)
		}
	})
	t.Run("FindFolders", func(t *testing.T) {
		r, err := SelectedPhotos(folders)

		if assert.Nil(t, err) {
			assert.Equal(t, 2, len(r))
			assert.IsType(t, entity.Photos{}, r)
		}
	})
	t.Run("FindStates", func(t *testing.T) {
		r, err := SelectedPhotos(states)

		if assert.Nil(t, err) {
			assert.Equal(t, 4, len(r))
			assert.IsType(t, entity.Photos{}, r)
		}
	})
	t.Run("NotNil", func(t *testing.T) {
		f := form.Selection{
			Photos: []string{"pszzzzzzzzzzzzzz"},
		}

		r, err := SelectedPhotos(f)

		if assert.Nil(t, err) {
			assert.NotNil(t, r)
			assert.Len(t, r, 0)
		}
	})

}

// likeTestPhoto creates a photo with one original in dir and removes both when the test ends.
func likeTestPhoto(t *testing.T, dir, name string) *entity.Photo {
	t.Helper()

	photo := &entity.Photo{PhotoPath: dir, PhotoName: name, PhotoType: entity.MediaImage, PhotoQuality: 3}
	require.NoError(t, photo.Create())

	file := &entity.File{
		PhotoID:     photo.ID,
		PhotoUID:    photo.PhotoUID,
		FileName:    dir + "/" + name + ".jpg",
		FileRoot:    entity.RootOriginals,
		FileHash:    rnd.GenerateUID(entity.FileUID),
		FileType:    "jpg",
		FilePrimary: true,
	}
	require.NoError(t, file.Create())

	t.Cleanup(func() {
		_ = entity.UnscopedDb().Where("photo_id = ?", photo.ID).Delete(&entity.Details{}).Error
		_ = entity.UnscopedDb().Where("photo_id = ?", photo.ID).Delete(&entity.File{}).Error
		_ = entity.UnscopedDb().Where("id = ?", photo.ID).Delete(&entity.Photo{}).Error
	})

	return photo
}

// likeTestFolder creates an originals folder row for dir and removes it and its album when the test ends.
func likeTestFolder(t *testing.T, dir string) entity.Folder {
	t.Helper()

	folder := entity.NewFolder(entity.RootOriginals, dir, time.Now())
	require.NoError(t, folder.Create())

	t.Cleanup(func() {
		_ = entity.UnscopedDb().Where("folder_uid = ?", folder.FolderUID).Delete(&entity.Folder{}).Error
		_ = entity.UnscopedDb().Where("album_type = ? AND album_path = ?", entity.AlbumFolder, dir).Delete(&entity.Album{}).Error
	})

	return folder
}

func TestSubfolderCond(t *testing.T) {
	t.Run("MySQL", func(t *testing.T) {
		cond, err := subfolderCond(dsn.DriverMySQL)
		require.NoError(t, err)
		assert.Equal(t, "b.path LIKE CONCAT(REPLACE(REPLACE(REPLACE(a.path, '!', '!!'), '%', '!%'), '_', '!_'), '/%') ESCAPE '!'"+
			" AND SUBSTR(b.path, 1, LENGTH(a.path) + 1) = CONCAT(a.path, '/')", cond)
	})
	t.Run("SQLite", func(t *testing.T) {
		cond, err := subfolderCond(dsn.DriverSQLite3)
		require.NoError(t, err)
		assert.Equal(t, "b.path LIKE REPLACE(REPLACE(REPLACE(a.path, '!', '!!'), '%', '!%'), '_', '!_') || '/%' ESCAPE '!'"+
			" AND SUBSTR(b.path, 1, LENGTH(a.path) + 1) = a.path || '/'", cond)
	})
	t.Run("UnknownDialect", func(t *testing.T) {
		cond, err := subfolderCond("postgres")
		assert.Error(t, err)
		assert.Empty(t, cond)
	})
}

func TestSelectedPhotos_SubfolderContainment(t *testing.T) {
	base := "zz-like-" + rnd.Base36(6)
	folder := likeTestFolder(t, base+"_a!b!%")
	likeTestFolder(t, base+"_a!b!%/sub")
	likeTestFolder(t, base+"Xa!b!Y/sub")
	likeTestFolder(t, base+"_a!b!Z/sub")
	likeTestFolder(t, base+"_A!b!%/sub")
	inFolder := likeTestPhoto(t, base+"_a!b!%", "in-folder")
	inSubfolder := likeTestPhoto(t, base+"_a!b!%/sub", "in-subfolder")
	sibling := likeTestPhoto(t, base+"Xa!b!Y/sub", "sibling")
	likeTestPhoto(t, base+"_a!b!Z/sub", "sibling-z")
	likeTestPhoto(t, base+"_A!b!%/sub", "sibling-case")

	results, err := SelectedPhotos(form.Selection{Files: []string{folder.FolderUID}})
	require.NoError(t, err)

	uids := results.UIDs()
	assert.ElementsMatch(t, []string{inFolder.PhotoUID, inSubfolder.PhotoUID}, uids)
	assert.NotContains(t, uids, sibling.PhotoUID)
}
