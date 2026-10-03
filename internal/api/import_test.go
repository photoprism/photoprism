package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestCancelImport(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CancelImport(router)
		r := PerformRequest(app, "DELETE", "/api/v1/import")

		var resp i18n.Response

		if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}

		assert.True(t, resp.Success())
		assert.Equal(t, i18n.Msg(i18n.MsgImportCanceled), resp.Message)
		assert.Equal(t, i18n.Msg(i18n.MsgImportCanceled), resp.String())
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, http.StatusOK, resp.Code)
	})
}

func TestStartImport(t *testing.T) {
	t.Run("ReadOnlyMode", func(t *testing.T) {
		app, router, config := NewApiTest()
		config.Options().ReadOnly = true
		StartImport(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/import/test", "{foo:123}")

		assert.Equal(t, http.StatusForbidden, r.Code)
		config.Options().ReadOnly = false
	})
	t.Run("QuotaExceeded", func(t *testing.T) {
		app, router, config := NewApiTest()
		config.Options().FilesQuota = 1
		StartImport(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/import/test", "{foo:123}")

		assert.Equal(t, http.StatusInsufficientStorage, r.Code)
		config.Options().FilesQuota = 0
	})
}

func TestStartImportAlbums(t *testing.T) {
	app, router, conf := NewApiTest()
	StartImport(router)
	options := *conf.Options()
	mode := conf.AuthMode()
	t.Cleanup(func() { *conf.Options() = options; conf.SetAuthMode(mode) })
	conf.SetAuthMode(config.AuthModePasswd)
	conf.Options().StoragePath = t.TempDir()
	conf.Options().OriginalsPath = t.TempDir()
	conf.Options().SidecarPath = t.TempDir()
	conf.Options().ImportPath = t.TempDir()
	conf.Options().ImportAllow = ""
	user := entity.UserFixtures.Pointer("alice")
	sess := clientCredentialSession(t, conf, "client", "*", user)

	// Import a folder of the import path; only the album that exists receives the imported picture.
	album := entity.NewUserAlbum("Import Target "+rnd.Base36(6), entity.AlbumManual, "", user.UserUID)
	require.NoError(t, album.Create())
	t.Cleanup(func() { _ = entity.UnscopedDb().Unscoped().Delete(album).Error })
	missing := rnd.GenerateUID(entity.AlbumUID)

	folder := "zz-import-" + rnd.Base36(8)
	dir := filepath.Join(conf.ImportPath(), folder)
	require.NoError(t, fs.MkdirAll(dir))
	filename := filepath.Join(dir, "import.jpg")
	require.NoError(t, os.WriteFile(filename, NewTestJpeg(t, 155, 105), fs.ModeFile))
	hash := fs.Hash(filename)
	t.Cleanup(func() {
		file, err := entity.FirstFileByHash(hash)
		if err != nil {
			return
		}
		entity.UnscopedDb().Unscoped().Delete(&entity.PhotoAlbum{}, "photo_uid = ?", file.PhotoUID)
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", file.PhotoID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", file.PhotoID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Photo{}, "id = ?", file.PhotoID)
	})

	body := fmt.Sprintf(`{"albums":[%q, %q]}`, missing, album.AlbumUID)
	result := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/import/"+folder, body, sess.AuthToken())
	require.Equal(t, http.StatusOK, result.Code, result.Body.String())

	file, err := entity.FirstFileByHash(hash)
	require.NoError(t, err)

	var count int
	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoAlbum{}).Where("photo_uid = ? AND album_uid = ?", file.PhotoUID, album.AlbumUID).Count(&count).Error)
	assert.Equal(t, 1, count)
	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoAlbum{}).Where("album_uid = ?", missing).Count(&count).Error)
	assert.Equal(t, 0, count)
}

func TestStartImportUploadFolder(t *testing.T) {
	app, router, conf := NewApiTest()
	StartImport(router)
	options := *conf.Options()
	mode := conf.AuthMode()
	t.Cleanup(func() { *conf.Options() = options; conf.SetAuthMode(mode) })
	conf.SetAuthMode(config.AuthModePasswd)
	conf.Options().StoragePath = t.TempDir()
	conf.Options().OriginalsPath = t.TempDir()
	conf.Options().SidecarPath = t.TempDir()
	conf.Options().ImportPath = t.TempDir()
	conf.Options().ImportAllow = ""

	// stage returns the hash of a picture written to the given folder of the import path.
	stage := func(t *testing.T, folder string, w, h int) string {
		dir := filepath.Join(conf.ImportPath(), folder)
		require.NoError(t, fs.MkdirAll(dir))
		filename := filepath.Join(dir, "import.jpg")
		require.NoError(t, os.WriteFile(filename, NewTestJpeg(t, w, h), fs.ModeFile))
		hash := fs.Hash(filename)
		t.Cleanup(func() {
			file, err := entity.FirstFileByHash(hash)
			if err != nil {
				return
			}
			entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", file.PhotoID)
			entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", file.PhotoID)
			entity.UnscopedDb().Unscoped().Delete(&entity.Photo{}, "id = ?", file.PhotoID)
		})
		return hash
	}

	t.Run("Admin", func(t *testing.T) {
		// A folder named upload is imported like any other folder of the import path.
		sess := clientCredentialSession(t, conf, "client", "*", entity.UserFixtures.Pointer("alice"))
		token := rnd.Base36(10)
		hash := stage(t, "upload/"+token, 163, 113)
		result := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/import/upload/"+token, `{}`, sess.AuthToken())
		require.Equal(t, http.StatusOK, result.Code, result.Body.String())
		_, err := entity.FirstFileByHash(hash)
		assert.NoError(t, err)
	})
	t.Run("Contributor", func(t *testing.T) {
		role := acl.RoleContributor
		if _, ok := acl.UserRoles[role.String()]; !ok {
			acl.UserRoles[role.String()] = role
			t.Cleanup(func() { delete(acl.UserRoles, role.String()) })
		}
		if _, ok := acl.Rules[acl.ResourceFiles][role]; !ok {
			acl.Rules[acl.ResourceFiles][role] = acl.GrantUploadAccess
			t.Cleanup(func() { delete(acl.Rules[acl.ResourceFiles], role) })
		}

		user := &entity.User{UserName: "contrib" + rnd.Base36(6), UserRole: role.String(), UploadPath: "inbox", CanLogin: true}
		require.NoError(t, user.Create())
		t.Cleanup(func() {
			_ = entity.UnscopedDb().Unscoped().Delete(&entity.UserDetails{}, "user_uid = ?", user.UserUID).Error
			_ = entity.UnscopedDb().Unscoped().Delete(&entity.UserSettings{}, "user_uid = ?", user.UserUID).Error
			_ = entity.UnscopedDb().Unscoped().Delete(user).Error
		})
		sess := clientCredentialSession(t, conf, "client", "*", user)

		// Importing needs permission to manage files, also for uploads staged in the import path.
		token := rnd.Base36(10)
		hash := stage(t, "upload/"+sess.RefID+token, 165, 115)
		result := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/import/upload/"+token, `{}`, sess.AuthToken())
		assert.Equal(t, http.StatusForbidden, result.Code)
		_, err := entity.FirstFileByHash(hash)
		assert.Error(t, err)
	})
}
