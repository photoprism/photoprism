package api

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// TestShareAliases checks the remote destinations assigned below a selected folder.
func TestShareAliases(t *testing.T) {
	newFile := func(hash string) *entity.File {
		return &entity.File{
			FileName:  "alias-helper.jpg",
			FileType:  "jpg",
			FileHash:  hash,
			MediaType: entity.MediaImage,
			Photo:     &entity.Photo{PhotoTitle: "Alias Helper", TakenAtLocal: time.Date(2026, 9, 17, 10, 30, 0, 0, time.UTC)},
		}
	}
	file := newFile("deadbeef00000000000000000000000000000920")
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, "Uploads/Trip/20260917-103000-Alias-Helper.jpg", newShareAliases().Resolve(file, "Uploads/Trip"))
	})
	t.Run("RootFolder", func(t *testing.T) {
		assert.Equal(t, "20260917-103000-Alias-Helper.jpg", newShareAliases().Resolve(file, ""))
	})
	t.Run("Collision", func(t *testing.T) {
		aliases := newShareAliases()
		aliases.Keep("Uploads/Trip/20260917-103000-Alias-Helper.jpg")
		assert.Equal(t, "Uploads/Trip/20260917-103000-Alias-Helper (1).jpg", aliases.Resolve(file, "Uploads/Trip"))
		aliases.Keep("uploads/TRIP/20260917-103000-Alias-Helper (1).jpg")
		assert.Equal(t, "Uploads/Trip/20260917-103000-Alias-Helper (2).jpg", aliases.Resolve(file, "Uploads/Trip"))
	})
	// The sequence number advances past every name already assigned, whoever assigned it.
	t.Run("TakenSequence", func(t *testing.T) {
		aliases := newShareAliases()
		aliases.Keep("Uploads/Trip/20260917-103000-Alias-Helper.jpg")
		for seq := 1; seq < 5; seq++ {
			aliases.Keep(fmt.Sprintf("Uploads/Trip/20260917-103000-Alias-Helper (%d).jpg", seq))
		}
		assert.Equal(t, "Uploads/Trip/20260917-103000-Alias-Helper (5).jpg", aliases.Resolve(file, "Uploads/Trip"))
	})
	// Assigning many equal names keeps them unique without rescanning the numbers already used.
	t.Run("Sequence", func(t *testing.T) {
		aliases := newShareAliases()
		for seq := range 10 {
			alias := aliases.Resolve(newFile(fmt.Sprintf("deadbeef%032d", 920+seq)), "Uploads/Trip")
			assert.False(t, aliases.taken[strings.ToLower(alias)], alias)
			aliases.Keep(alias)
		}
		assert.Equal(t, 10, len(aliases.taken))
		assert.Equal(t, 9, aliases.next["uploads/trip/20260917-103000-alias-helper.jpg"])
	})
	// A missing file hash yields a generated name, which needs no sequence number to stay unique.
	t.Run("GeneratedName", func(t *testing.T) {
		alias := newShareAliases().Resolve(newFile(""), "Uploads/Trip")
		require.Equal(t, "Uploads/Trip", path.Dir(alias))
		assert.Equal(t, ".jpg", path.Ext(alias))
		assert.NotEqual(t, alias, newShareAliases().Resolve(newFile(""), "Uploads/Trip"))
	})
}

// TestUploadToServiceAliases checks the remote names queued for selected files.
func TestUploadToServiceAliases(t *testing.T) {
	app, router, _ := NewApiTest()
	UploadToService(router)
	require.NoError(t, mutex.ShareWorker.Start())
	t.Cleanup(mutex.ShareWorker.Stop)
	photo := entity.NewPhoto(false)
	photo.PhotoTitle = "Alias Control"
	photo.TakenAt = time.Date(2026, 9, 17, 10, 30, 0, 0, time.UTC)
	photo.TakenAtLocal = photo.TakenAt
	photo.PhotoQuality = 3
	require.NoError(t, photo.Save())
	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(photo)
	})
	// The first three files share a preferred alias, while the fourth has a name of its own.
	var files []*entity.File
	for i, fileType := range []string{"jpg", "jpg", "jpg", "png"} {
		file := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: fmt.Sprintf("alias-%d.%s", i, fileType), FileType: fileType, MediaType: entity.MediaImage, FileHash: fmt.Sprintf("deadbeef%032d", 910+i), FilePrimary: i == 0}
		require.NoError(t, file.Create())
		files = append(files, file)
	}
	// uploadAliases queues the selection against a new service and returns the remote names.
	uploadAliases := func(t *testing.T, name, folder string) []string {
		account := entity.Service{AccName: name, AccURL: "http://127.0.0.1/", AccType: "webdav", AccShare: true}
		require.NoError(t, entity.Db().Create(&account).Error)
		t.Cleanup(func() {
			entity.UnscopedDb().Unscoped().Delete(&entity.FileShare{}, "service_id = ?", account.ID)
			entity.UnscopedDb().Unscoped().Delete(&account)
		})
		uri := fmt.Sprintf("/api/v1/services/%d/upload", account.ID)
		body := fmt.Sprintf(`{"selection":{"photos":[%q]},"folder":%q}`, photo.PhotoUID, folder)
		result := PerformRequestWithBody(app, http.MethodPost, uri, body)
		require.Equal(t, http.StatusOK, result.Code, result.Body.String())
		var shares []entity.FileShare
		require.NoError(t, entity.Db().Where("service_id = ?", account.ID).Find(&shares).Error)
		names := make([]string, 0, len(shares))
		for _, share := range shares {
			names = append(names, share.RemoteName)
		}
		return names
	}
	t.Run("NestedFolder", func(t *testing.T) {
		names := uploadAliases(t, "Alias Control Nested", "Uploads/Trip")
		require.Len(t, names, 4)
		assert.ElementsMatch(t, []string{
			path.Join("Uploads/Trip", files[0].ShareBase(0)),
			path.Join("Uploads/Trip", files[0].ShareBase(1)),
			path.Join("Uploads/Trip", files[0].ShareBase(2)),
			path.Join("Uploads/Trip", files[3].ShareBase(0)),
		}, names)
		for _, name := range names {
			assert.Equal(t, "Uploads/Trip", path.Dir(name))
		}
	})
	// A destination with a parent-directory segment is rejected before any file is queued.
	t.Run("UnsafeFolder", func(t *testing.T) {
		account := entity.Service{AccName: "Alias Control Unsafe", AccURL: "http://127.0.0.1/", AccType: "webdav", AccShare: true}
		require.NoError(t, entity.Db().Create(&account).Error)
		t.Cleanup(func() {
			entity.UnscopedDb().Unscoped().Delete(&entity.FileShare{}, "service_id = ?", account.ID)
			entity.UnscopedDb().Unscoped().Delete(&account)
		})
		uri := fmt.Sprintf("/api/v1/services/%d/upload", account.ID)
		for _, folder := range []string{"../outside", "Uploads/../../outside", `..\outside`} {
			body := fmt.Sprintf(`{"selection":{"photos":[%q]},"folder":%q}`, photo.PhotoUID, folder)
			result := PerformRequestWithBody(app, http.MethodPost, uri, body)
			require.Equal(t, http.StatusBadRequest, result.Code, folder)
		}
		var count int
		require.NoError(t, entity.Db().Model(&entity.FileShare{}).Where("service_id = ?", account.ID).Count(&count).Error)
		assert.Equal(t, 0, count)
	})
	t.Run("RootFolder", func(t *testing.T) {
		names := uploadAliases(t, "Alias Control Root", "")
		require.Len(t, names, 4)
		assert.ElementsMatch(t, []string{
			files[0].ShareBase(0),
			files[0].ShareBase(1),
			files[0].ShareBase(2),
			files[3].ShareBase(0),
		}, names)
		for _, name := range names {
			assert.Equal(t, ".", path.Dir(name))
		}
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

func TestUploadToServiceYaml(t *testing.T) {
	app, router, conf := NewApiTest()
	previous := conf.Options().OriginalsPath
	conf.Options().OriginalsPath = t.TempDir()
	t.Cleanup(func() { conf.Options().OriginalsPath = previous })
	UploadToService(router)
	require.NoError(t, mutex.ShareWorker.Start())
	t.Cleanup(mutex.ShareWorker.Stop)
	account := entity.Service{AccName: "YAML Control", AccURL: "http://127.0.0.1/", AccType: "webdav", AccShare: true}
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
	jpeg := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: "yaml-control.jpg", FileType: "jpg", MediaType: entity.MediaImage, FileHash: fmt.Sprintf("%040d", 960), FilePrimary: true}
	require.NoError(t, jpeg.Create())
	yaml := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: "yaml-control.yml", FileType: "yml", MediaType: "sidecar", FileSidecar: true, FileHash: fmt.Sprintf("%040d", 961)}
	require.NoError(t, yaml.Create())
	uri := fmt.Sprintf("/api/v1/services/%d/upload", account.ID)
	body := `{"selection":{"photos":["` + photo.PhotoUID + `"]},"folder":"/"}`
	shared := func(t *testing.T) (result []uint) {
		var shares []entity.FileShare
		require.NoError(t, entity.Db().Where("service_id = ?", account.ID).Find(&shares).Error)
		seen := make(map[uint]bool)
		for _, s := range shares {
			if !seen[s.FileID] {
				seen[s.FileID] = true
				result = append(result, s.FileID)
			}
		}
		return result
	}
	t.Run("Disabled", func(t *testing.T) {
		require.NoError(t, account.Update("SyncYaml", -1))
		result := PerformRequestWithBody(app, http.MethodPost, uri, body)
		require.Equal(t, http.StatusOK, result.Code, result.Body.String())
		assert.Equal(t, int64(1), gjson.GetBytes(result.Body.Bytes(), "#").Int())
		assert.ElementsMatch(t, []uint{jpeg.ID}, shared(t))
	})
	t.Run("Default", func(t *testing.T) {
		require.NoError(t, account.Update("SyncYaml", 0))
		result := PerformRequestWithBody(app, http.MethodPost, uri, body)
		require.Equal(t, http.StatusOK, result.Code, result.Body.String())
		assert.Equal(t, int64(2), gjson.GetBytes(result.Body.Bytes(), "#").Int())
		assert.ElementsMatch(t, []uint{jpeg.ID, yaml.ID}, shared(t))
	})
	t.Run("Enabled", func(t *testing.T) {
		require.NoError(t, account.Update("SyncYaml", 1))
		result := PerformRequestWithBody(app, http.MethodPost, uri, body)
		require.Equal(t, http.StatusOK, result.Code, result.Body.String())
		assert.Equal(t, int64(2), gjson.GetBytes(result.Body.Bytes(), "#").Int())
		assert.ElementsMatch(t, []uint{jpeg.ID, yaml.ID}, shared(t))
	})
}
