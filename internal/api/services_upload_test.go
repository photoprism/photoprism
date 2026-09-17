package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestUploadToService(t *testing.T) {
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UploadToService(router)
		r := PerformRequest(app, "POST", "/api/v1/services/1000000/upload")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrBadRequest), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("AccountNotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UploadToService(router)
		r := PerformRequest(app, "POST", "/api/v1/services/999000/upload")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrAccountNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	// Covers the file selection, which the cases above never reach: they abort in AccountByID or
	// BindJSON, both of which run before it.
	t.Run("EmptySelection", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UploadToService(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/services/1000000/upload", `{"folder": "/Photos"}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrEntityNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

// TestUploadToServiceReservedPaths checks original paths before remote aliases are queued.
func TestUploadToServiceReservedPaths(t *testing.T) {
	app, router, conf := NewApiTest()
	previous := conf.Options().OriginalsPath
	conf.Options().OriginalsPath = filepath.Join(t.TempDir(), ".config", "originals")
	t.Cleanup(func() { conf.Options().OriginalsPath = previous })
	require.NoError(t, fs.MkdirAll(filepath.Join(conf.OriginalsPath(), ".config")))
	require.NoError(t, os.WriteFile(filepath.Join(conf.OriginalsPath(), ".config/private.jpg"), []byte("control"), fs.ModeFile))
	require.NoError(t, os.Symlink(filepath.Join(conf.OriginalsPath(), ".config/private.jpg"), filepath.Join(conf.OriginalsPath(), "alias.jpg")))
	UploadToService(router)
	require.NoError(t, mutex.ShareWorker.Start())
	t.Cleanup(mutex.ShareWorker.Stop)
	account := entity.Service{AccName: "Reserved Path Control", AccURL: "http://127.0.0.1/", AccType: "webdav", AccShare: true, ShareSize: "original"}
	require.NoError(t, entity.Db().Create(&account).Error)
	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.FileShare{}, "service_id = ?", account.ID)
		entity.UnscopedDb().Unscoped().Delete(&account)
	})
	photo := entity.NewPhoto(false)
	photo.PhotoQuality = 3
	require.NoError(t, photo.Save())
	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(photo)
	})
	var files []*entity.File
	for i, name := range []string{".config/private.jpg", "alias.jpg", "ordinary.jpg"} {
		file := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: name, FileType: "jpg", MediaType: entity.MediaImage, FileHash: fmt.Sprintf("%040d", 900+i), FilePrimary: true}
		require.NoError(t, file.Create())
		files = append(files, file)
	}
	uri := fmt.Sprintf("/api/v1/services/%d/upload", account.ID)
	result := PerformRequestWithBody(app, http.MethodPost, uri, `{"selection":{"photos":["`+photo.PhotoUID+`"]},"folder":"ordinary"}`)
	require.Equal(t, http.StatusOK, result.Code, result.Body.String())
	assert.Equal(t, int64(2), gjson.GetBytes(result.Body.Bytes(), "#").Int())
	var shares []entity.FileShare
	require.NoError(t, entity.Db().Where("service_id = ?", account.ID).Find(&shares).Error)
	require.Len(t, shares, 2)
	assert.ElementsMatch(t, []uint{files[1].ID, files[2].ID}, []uint{shares[0].FileID, shares[1].FileID})
	result = PerformRequestWithBody(app, http.MethodPost, uri, `{"selection":{"photos":["`+photo.PhotoUID+`"]},"folder":".ssh"}`)
	require.Equal(t, http.StatusBadRequest, result.Code)
	assert.Equal(t, int64(0), gjson.GetBytes(result.Body.Bytes(), "#").Int())
	var count int
	require.NoError(t, entity.Db().Model(&entity.FileShare{}).Where("service_id = ?", account.ID).Count(&count).Error)
	assert.Equal(t, 2, count)
}
