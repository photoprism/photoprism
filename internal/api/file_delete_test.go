package api

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism"
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
	t.Run("DeleteDuplicateMissingName", func(t *testing.T) {
		app, router, _ := NewApiTest()

		DeleteFile(router)

		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh7/duplicates")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("DeleteDuplicateNotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()

		DeleteFile(router)

		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh7/duplicates?name=missing.jpg&root=/")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("DeleteDuplicateSuccess", func(t *testing.T) {
		app, router, conf := NewApiTest()

		photo, err := query.PhotoPreloadByUID("ps6sg6be2lvl0yh7")
		require.NoError(t, err)
		require.NotEmpty(t, photo.Files)

		duplicateName := "api-test/delete-duplicate-success.jpg"
		duplicatePath := photoprism.FileName(entity.RootOriginals, duplicateName)

		require.NoError(t, os.MkdirAll(filepath.Dir(duplicatePath), 0o755))
		require.NoError(t, os.WriteFile(duplicatePath, []byte("duplicate"), 0o644))
		require.NoError(t, entity.AddDuplicate(duplicateName, entity.RootOriginals, photo.Files[0].FileHash, int64(len("duplicate")), time.Now().Unix()))

		defer func() {
			_ = os.Remove(duplicatePath)
			_ = entity.PurgeDuplicate(duplicateName, entity.RootOriginals)
		}()

		DeleteFile(router)

		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh7/duplicates?name=api-test/delete-duplicate-success.jpg&root=/")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, conf.OriginalsPath(), filepath.Dir(filepath.Dir(duplicatePath)))
		assert.Empty(t, gjson.Get(r.Body.String(), "Duplicates").Array())

		duplicate := entity.Duplicate{FileName: duplicateName, FileRoot: entity.RootOriginals}
		assert.Error(t, duplicate.Find())
		assert.NoFileExists(t, duplicatePath)
	})
}
