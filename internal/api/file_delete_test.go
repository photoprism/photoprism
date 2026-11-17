package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestDeleteFile(t *testing.T) {
	t.Run("DeleteNotExistingFile", func(t *testing.T) {
		app, router, _ := NewApiTest()

		DeleteFile(router)

		r := PerformRequest(app, "DELETE", "/api/v1/photos/5678/files/23456hbg")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("DeletePrimaryFile", func(t *testing.T) {
		app, router, _ := NewApiTest()

		DeleteFile(router)

		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh7/files/fs6sg6bw45bnlqdw")
		assert.Equal(t, http.StatusInternalServerError, r.Code)
	})
	t.Run("TryToDeleteFile", func(t *testing.T) {
		app, router, _ := NewApiTest()

		DeleteFile(router)

		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh8/files/fs6sg6bw45bn0001")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("PhotoMismatch", func(t *testing.T) {
		app, router, conf := NewApiTest()

		DeleteFile(router)

		options := *conf.Options()
		t.Cleanup(func() {
			*conf.Options() = options
			conf.Propagate()
		})
		conf.Options().OriginalsPath = t.TempDir()
		conf.Propagate()

		photo := entity.PhotoFixtures.Pointer("Photo01")
		other := entity.PhotoFixtures.Pointer("19800101_000002_D640C559")
		require.NotEqual(t, photo.PhotoUID, other.PhotoUID)

		name := "file-delete/" + rnd.Base36(8) + ".jpg"
		fileName := filepath.Join(conf.OriginalsPath(), name)
		require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))
		require.NoError(t, os.WriteFile(fileName, []byte("not empty"), fs.ModeFile)) //nolint:gosec // Test writes to its own temporary folder.

		f := &entity.File{
			PhotoID:     photo.ID,
			PhotoUID:    photo.PhotoUID,
			FileUID:     rnd.GenerateUID(entity.FileUID),
			FileRoot:    entity.RootOriginals,
			FileName:    name,
			FileHash:    rnd.Base36(40),
			FileType:    fs.ImageJpeg.String(),
			FilePrimary: false,
		}
		require.NoError(t, f.Create())
		t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "file_uid = ?", f.FileUID) })

		// A file named under another photo is not found and stays in place.
		r := PerformRequest(app, http.MethodDelete, "/api/v1/photos/"+other.PhotoUID+"/files/"+f.FileUID)
		assert.Equal(t, http.StatusNotFound, r.Code)
		assert.FileExists(t, fileName)
		var count int64
		require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("file_uid = ?", f.FileUID).Count(&count).Error)
		assert.Equal(t, int64(1), count)

		// The file's own photo deletes it and returns that photo.
		r = PerformRequest(app, http.MethodDelete, "/api/v1/photos/"+photo.PhotoUID+"/files/"+f.FileUID)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		var result entity.Photo
		require.NoError(t, json.Unmarshal(r.Body.Bytes(), &result))
		assert.Equal(t, photo.PhotoUID, result.PhotoUID)
		assert.NoFileExists(t, fileName)
		require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("file_uid = ?", f.FileUID).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})
}
