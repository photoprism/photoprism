package api

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestPhotoEditsDuringFaceMigration pins that the edits which index a file again answer 409 while a
// face migration holds the lock, after authorization and before anything is changed.
func TestPhotoEditsDuringFaceMigration(t *testing.T) {
	app, router, conf := NewApiTest()
	ChangeFileOrientation(router)
	PhotoUnstack(router)

	folder := "migrationedit" + rnd.Base36(10)
	dir := filepath.Join(conf.OriginalsPath(), folder)
	require.NoError(t, fs.MkdirAll(filepath.Join(dir, "London")))
	require.NoError(t, fs.Copy("./testdata/face_160x160.jpg", filepath.Join(dir, "bridge2.jpg"), true))
	require.NoError(t, fs.Copy("./testdata/london_160x160.jpg", filepath.Join(dir, "London", "bridge3.jpg"), true))

	photo := entity.NewUserPhoto(true, "")
	photo.PhotoPath = folder
	photo.PhotoName = "bridge2"
	require.NoError(t, photo.Create())

	primary := &entity.File{
		FileUID: rnd.GenerateUID(entity.FileUID), PhotoID: photo.ID, PhotoUID: photo.PhotoUID,
		FileName: folder + "/bridge2.jpg", FileRoot: entity.RootOriginals, FilePrimary: true,
		FileHash: rnd.GenerateUID(entity.FileUID), FileOrientation: 1,
	}
	require.NoError(t, primary.Create())

	secondary := &entity.File{
		FileUID: rnd.GenerateUID(entity.FileUID), PhotoID: photo.ID, PhotoUID: photo.PhotoUID,
		FileName: folder + "/London/bridge3.jpg", FileRoot: entity.RootOriginals,
		FileHash: rnd.GenerateUID(entity.FileUID),
	}
	require.NoError(t, secondary.Create())

	t.Cleanup(func() {
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "file_name LIKE ?", folder+"/%").Error)
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.Photo{}, "photo_path LIKE ?", folder+"%").Error)
		assert.NoError(t, os.RemoveAll(dir))
	})

	lock, err := mutex.AcquireFileLock(conf.FacesLockFile(), "faces migration")
	require.NoError(t, err)
	t.Cleanup(lock.Release)

	orientationUri := "/api/v1/photos/" + photo.PhotoUID + "/files/" + primary.FileUID + "/orientation"
	unstackUri := "/api/v1/photos/" + photo.PhotoUID + "/files/" + secondary.FileUID + "/unstack"

	t.Run("Orientation", func(t *testing.T) {
		before, err := os.ReadFile(filepath.Join(dir, "bridge2.jpg")) //nolint:gosec // test-owned path
		require.NoError(t, err)

		r := PerformRequestWithBody(app, http.MethodPut, orientationUri, `{"Orientation":6}`)
		assert.Equal(t, http.StatusConflict, r.Code, r.Body.String())

		var file entity.File
		require.NoError(t, entity.UnscopedDb().First(&file, "file_uid = ?", primary.FileUID).Error)
		assert.Equal(t, 1, file.FileOrientation, "the orientation is not changed")

		after, err := os.ReadFile(filepath.Join(dir, "bridge2.jpg")) //nolint:gosec // test-owned path
		require.NoError(t, err)
		assert.Equal(t, before, after, "the file header is not changed")
	})
	t.Run("SameOrientation", func(t *testing.T) {
		// A request that changes nothing indexes nothing, so it succeeds as it does outside a migration.
		r := PerformRequestWithBody(app, http.MethodPut, orientationUri, `{"Orientation":1}`)
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
	})
	t.Run("Unstack", func(t *testing.T) {
		r := PerformRequest(app, http.MethodPost, unstackUri)
		assert.Equal(t, http.StatusConflict, r.Code, r.Body.String())

		var file entity.File
		require.NoError(t, entity.UnscopedDb().First(&file, "file_uid = ?", secondary.FileUID).Error)
		assert.Equal(t, photo.ID, file.PhotoID, "the file stays in its stack")

		stack := entity.FindPhoto(photo)
		require.NotNil(t, stack)
		assert.Equal(t, photo.PhotoStack, stack.PhotoStack, "the stack flag is not changed")
	})
	t.Run("AfterAuthorization", func(t *testing.T) {
		prevAuthMode := conf.AuthMode()
		conf.SetAuthMode(config.AuthModePasswd)
		t.Cleanup(func() { conf.SetAuthMode(prevAuthMode) })

		r := PerformRequestWithBody(app, http.MethodPut, orientationUri, `{"Orientation":6}`)
		assert.Equal(t, http.StatusUnauthorized, r.Code, r.Body.String())

		r = PerformRequest(app, http.MethodPost, unstackUri)
		assert.Equal(t, http.StatusUnauthorized, r.Code, r.Body.String())
	})
}
