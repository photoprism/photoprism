package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestUploadUserFiles(t *testing.T) {
	t.Run("BadRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s/upload/abc123456789", adminUid)
		// t.Logf("Request URL: %s", reqUrl)
		UploadUserFiles(router)
		r := PerformRequestWithBody(app, "POST", reqUrl, "{foo:123}")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("ReadOnlyMode", func(t *testing.T) {
		app, router, config := NewApiTest()
		config.Options().ReadOnly = true
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s/upload/abc123456789", adminUid)
		// t.Logf("Request URL: %s", reqUrl)
		UploadUserFiles(router)
		r := PerformRequestWithBody(app, "POST", reqUrl, "{foo:123}")
		assert.Equal(t, http.StatusForbidden, r.Code)
		config.Options().ReadOnly = false
	})
	t.Run("QuotaExceeded", func(t *testing.T) {
		app, router, config := NewApiTest()
		config.Options().FilesQuota = 1
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s/upload/abc123456789", adminUid)
		// t.Logf("Request URL: %s", reqUrl)
		UploadUserFiles(router)
		r := PerformRequestWithBody(app, "POST", reqUrl, "{foo:123}")
		assert.Equal(t, http.StatusInsufficientStorage, r.Code)
		config.Options().FilesQuota = 0
	})
	t.Run("ProcessRequestTooLarge", func(t *testing.T) {
		app, router, _ := NewApiTest()
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s/upload/abc123456789", adminUid)
		ProcessUserUpload(router)
		token := AuthenticateAdmin(app, router)

		body := `{"albums":["` + strings.Repeat("a", 300*1024) + `"]}`
		req := httptest.NewRequest(http.MethodPut, reqUrl, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		header.SetAuthorization(req, token)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	})
}

func TestUploadCheckFile_AcceptsAndReducesLimit(t *testing.T) {
	dir := t.TempDir()
	// Copy a small known-good JPEG test file from pkg/fs/testdata
	src := filepath.Clean("../../pkg/fs/testdata/directory/example.jpg")
	dst := filepath.Join(dir, "example.jpg")
	b, err := os.ReadFile(src)
	if err != nil {
		t.Skipf("skip if test asset not present: %v", err)
	}
	if err := os.WriteFile(dst, b, 0o600); err != nil { //nolint:gosec // test writes to a temp path under the test's control
		t.Fatal(err)
	}

	orig := int64(len(b))
	rem, err := UploadCheckFile(dst, false, orig+100)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), rem)
	// file remains
	assert.FileExists(t, dst)
}

func TestUploadCheckFile_TotalLimitReachedDeletes(t *testing.T) {
	dir := t.TempDir()
	// Make a tiny file
	dst := filepath.Join(dir, "tiny.txt")
	assert.NoError(t, os.WriteFile(dst, []byte("hello"), 0o600))
	// Very small total limit (0) → should remove file and error
	_, err := UploadCheckFile(dst, false, 0)
	assert.Error(t, err)
	_, statErr := os.Stat(dst)
	assert.True(t, os.IsNotExist(statErr), "file should be removed when limit reached")
}

func TestUploadCheckFile_UnsupportedTypeDeletes(t *testing.T) {
	dir := t.TempDir()
	// Create a file with an unknown extension; should be rejected
	dst := filepath.Join(dir, "unknown.xyz")
	assert.NoError(t, os.WriteFile(dst, []byte("not-an-image"), 0o600))
	_, err := UploadCheckFile(dst, false, 1<<20)
	assert.Error(t, err)
	// The message names the rejected file and reports the cause it was given.
	assert.Contains(t, err.Error(), "rejected")
	assert.Contains(t, err.Error(), "unknown.xyz")
	assert.NotContains(t, err.Error(), "no error")
	assert.NotNil(t, errors.Unwrap(err), "the cause stays reachable for a renderer")
	// The path the cause carries is removed when the error is rendered for the log.
	assert.NotContains(t, clean.Error(err), dir)
	_, statErr := os.Stat(dst)
	assert.True(t, os.IsNotExist(statErr), "unsupported file should be removed")
}

func TestUploadCheckFile_SizeAccounting(t *testing.T) {
	dir := t.TempDir()
	// Use known-good JPEG
	src := filepath.Clean("../../pkg/fs/testdata/directory/example.jpg")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Skip("asset missing; skip")
	}
	f := filepath.Join(dir, "a.jpg")
	assert.NoError(t, os.WriteFile(f, data, 0o600)) //nolint:gosec // test writes to a temp path under the test's control
	size := int64(len(data))
	// Set remaining limit to size+1 so it does not hit the removal branch (which triggers on <=0)
	rem, err := UploadCheckFile(f, false, size+1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), rem)
}

func TestUploadUserFilesStorageFolderError(t *testing.T) {
	app, router, _ := NewApiTest()
	UploadUserFiles(router)
	ProcessUserUpload(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID

	// A token longer than a path component makes the upload folder unusable, so both
	// handlers take the branch that reports the failure.
	longToken := strings.Repeat("t", 300)

	// storageFolderLogLine returns the line the handler logged for a failed upload folder.
	storageFolderLogLine := func(t *testing.T, method string, body *bytes.Buffer, contentType string) string {
		t.Helper()

		hook := captureLog(t)

		req := httptest.NewRequest(method, "/api/v1/users/"+adminUid+"/upload/"+longToken, body)
		req.Header.Set("Content-Type", contentType)
		header.SetAuthorization(req, token)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		for _, entry := range hook.AllEntries() {
			if strings.Contains(entry.Message, "storage folder") {
				return entry.Message
			}
		}

		t.Fatalf("expected a log entry for the storage folder, got %d entries (status %d)", len(hook.AllEntries()), w.Code)

		return ""
	}

	t.Run("Upload", func(t *testing.T) {
		body, ctype, err := buildMultipart(map[string][]byte{"a.jpg": []byte("x")})

		if err != nil {
			t.Fatal(err)
		}

		line := storageFolderLogLine(t, http.MethodPost, body, ctype)

		assert.Contains(t, line, "***")
		assert.NotContains(t, line, adminUid)
		assert.NotContains(t, line, longToken)
	})
	t.Run("Process", func(t *testing.T) {
		line := storageFolderLogLine(t, http.MethodPut, bytes.NewBufferString(`{"albums":[]}`), "application/json")

		assert.Contains(t, line, "***")
		assert.NotContains(t, line, adminUid)
		assert.NotContains(t, line, longToken)
	})
}

func TestUploadPathDenied(t *testing.T) {
	key := acl.RoleContributor.String()

	if _, had := acl.UserRoles[key]; !had {
		acl.UserRoles[key] = acl.RoleContributor
		t.Cleanup(func() { delete(acl.UserRoles, key) })
	}

	t.Run("ContributorWithoutPaths", func(t *testing.T) {
		assert.True(t, uploadPathDenied(&entity.User{UserName: ".", UserRole: key}))
	})
	t.Run("ContributorWithBasePath", func(t *testing.T) {
		assert.False(t, uploadPathDenied(&entity.User{UserName: "jane", UserRole: key}))
	})
	t.Run("ContributorWithUploadPath", func(t *testing.T) {
		assert.False(t, uploadPathDenied(&entity.User{UserName: ".", UserRole: key, UploadPath: "inbox"}))
	})
	t.Run("Admin", func(t *testing.T) {
		assert.False(t, uploadPathDenied(&entity.User{UserName: "admin", UserRole: acl.RoleAdmin.String()}))
	})
}

func TestUploadAlbumsAllowed(t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		s := &entity.Session{}
		s.SetUser(entity.UserFixtures.Pointer("alice"))
		assert.True(t, uploadAlbumsAllowed(s))
	})
	t.Run("AlbumsScope", func(t *testing.T) {
		s := &entity.Session{AuthScope: "files albums"}
		s.SetUser(entity.UserFixtures.Pointer("alice"))
		assert.True(t, uploadAlbumsAllowed(s))
	})
	t.Run("FilesScope", func(t *testing.T) {
		s := &entity.Session{AuthScope: "files"}
		s.SetUser(entity.UserFixtures.Pointer("alice"))
		assert.False(t, uploadAlbumsAllowed(s))
	})
	t.Run("RestrictedClientForAdmin", func(t *testing.T) {
		assert.False(t, uploadAlbumsAllowed(mixedPrincipalSession()))
	})
	t.Run("ClientWithoutUser", func(t *testing.T) {
		s := &entity.Session{}
		s.SetClient(&entity.Client{ClientRole: acl.RoleClient.String(), AuthProvider: authn.ProviderClient.String()})
		require.True(t, s.GrantsAny(acl.ResourceAlbums, acl.Permissions{acl.ActionCreate, acl.ActionUpload}))
		assert.False(t, uploadAlbumsAllowed(s))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, uploadAlbumsAllowed(nil))
	})
}

func TestUploadAlbums(t *testing.T) {
	// An album alice owns, one she holds a share for, and one she neither owns nor holds a share for.
	owned := entity.NewUserAlbum("Upload Albums "+rnd.Base36(6), entity.AlbumManual, "", entity.UserFixtures.Pointer("alice").UserUID)
	require.NoError(t, owned.Create())
	t.Cleanup(func() { _ = entity.UnscopedDb().Delete(owned).Error })
	other := entity.AlbumFixtures.Get("holiday-2030").AlbumUID
	missing := rnd.GenerateUID(entity.AlbumUID)
	deleted := entity.NewUserAlbum("Upload Albums "+rnd.Base36(6), entity.AlbumManual, "", entity.UserFixtures.Pointer("alice").UserUID)
	require.NoError(t, deleted.Create())
	t.Cleanup(func() { _ = entity.UnscopedDb().Unscoped().Delete(deleted).Error })
	require.NoError(t, deleted.Delete())

	// Look up the deleted album, which caches it, so the check must not depend on the cache.
	require.NotNil(t, entity.FindAlbum(entity.Album{AlbumUID: deleted.AlbumUID}))

	albums := []string{owned.AlbumUID, sharedAlbumUID, other, missing, deleted.AlbumUID, "New Album", owned.AlbumUID, "New Album"}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users/uqxetse3cy5eo9z2/upload/abc", nil)

	t.Run("FullAccess", func(t *testing.T) {
		s := &entity.Session{}
		s.SetUser(entity.UserFixtures.Pointer("alice"))
		assert.Equal(t, []string{owned.AlbumUID, sharedAlbumUID, other, "New Album"}, uploadAlbums(c, s, albums))
	})
	t.Run("SharedAccessOnly", func(t *testing.T) {
		s := mixedPrincipalSession()
		require.True(t, s.HasSharedAccessOnly(acl.ResourceAlbums))
		assert.Equal(t, []string{owned.AlbumUID, sharedAlbumUID, "New Album"}, uploadAlbums(c, s, albums))
	})
	t.Run("None", func(t *testing.T) {
		s := mixedPrincipalSession()
		assert.Empty(t, uploadAlbums(c, s, nil))
	})
	t.Run("Limit", func(t *testing.T) {
		s := &entity.Session{}
		s.SetUser(entity.UserFixtures.Pointer("alice"))
		titles := make([]string, 0, MaxUploadAlbums+4)
		titles = append(titles, other, missing, "", "")
		for i := 0; i < MaxUploadAlbums+1; i++ {
			titles = append(titles, fmt.Sprintf("Limit %d", i))
		}
		titles = append(titles, "Limit 0")

		// Refused, empty and repeated entries take no place; the album after the limit is skipped.
		result := uploadAlbums(c, s, titles)
		require.Len(t, result, MaxUploadAlbums)
		assert.Equal(t, other, result[0])
		assert.Equal(t, fmt.Sprintf("Limit %d", MaxUploadAlbums-2), result[MaxUploadAlbums-1])
	})
}

func TestProcessUserUploadAlbums(t *testing.T) {
	app, router, conf := NewApiTest()
	ProcessUserUpload(router)
	options := *conf.Options()
	mode := conf.AuthMode()
	t.Cleanup(func() { *conf.Options() = options; conf.SetAuthMode(mode) })
	conf.SetAuthMode(config.AuthModePasswd)
	conf.Options().StoragePath = t.TempDir()
	conf.Options().OriginalsPath = t.TempDir()
	conf.Options().SidecarPath = t.TempDir()
	conf.Options().ImportAllow = ""
	conf.Options().BackupAlbums = true
	user := entity.UserFixtures.Pointer("alice")
	sess := clientCredentialSession(t, conf, "client", "*", user)

	// Only the album that exists receives the uploaded picture.
	album := entity.NewUserAlbum("Upload Target "+rnd.Base36(6), entity.AlbumManual, "", user.UserUID)
	require.NoError(t, album.Create())
	t.Cleanup(func() { _ = entity.UnscopedDb().Unscoped().Delete(album).Error })
	missing := rnd.GenerateUID(entity.AlbumUID)

	// A title resolves among the user's own albums, so another user's album with the same title is not used.
	title := "Upload Title " + rnd.Base36(6)
	foreign := entity.NewUserAlbum(title, entity.AlbumManual, "", rnd.GenerateUID(entity.UserUID))
	require.NoError(t, foreign.Create())
	t.Cleanup(func() {
		_ = entity.UnscopedDb().Unscoped().Delete(&entity.PhotoAlbum{}, "album_uid IN (SELECT album_uid FROM albums WHERE album_title = ?)", title).Error
		_ = entity.UnscopedDb().Unscoped().Delete(&entity.Album{}, "album_title = ?", title).Error
	})

	token := rnd.Base36(10)
	dir, err := conf.UserUploadPath(user.UserUID, sess.RefID+token)
	require.NoError(t, err)
	filename := filepath.Join(dir, "upload.jpg")
	require.NoError(t, os.WriteFile(filename, NewTestJpeg(t, 153, 103), fs.ModeFile))
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

	body := fmt.Sprintf(`{"albums":[%q, %q, %q]}`, missing, album.AlbumUID, title)
	result := AuthenticatedRequestWithBody(app, http.MethodPut, "/api/v1/users/"+user.UserUID+"/upload/"+token, body, sess.AuthToken())
	require.Equal(t, http.StatusOK, result.Code, result.Body.String())

	file, err := entity.FirstFileByHash(hash)
	require.NoError(t, err)

	var count int
	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoAlbum{}).Where("photo_uid = ? AND album_uid = ?", file.PhotoUID, album.AlbumUID).Count(&count).Error)
	assert.Equal(t, 1, count)
	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoAlbum{}).Where("album_uid = ?", missing).Count(&count).Error)
	assert.Equal(t, 0, count)

	// The picture is added to a new album of the user, and only that album's backup file is written.
	var created entity.Album
	require.NoError(t, entity.UnscopedDb().Where("album_title = ? AND created_by = ?", title, user.UserUID).First(&created).Error)
	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoAlbum{}).Where("photo_uid = ? AND album_uid = ?", file.PhotoUID, created.AlbumUID).Count(&count).Error)
	assert.Equal(t, 1, count)
	createdYaml, _, err := created.YamlFileName(conf.BackupAlbumsPath())
	require.NoError(t, err)
	assert.FileExists(t, createdYaml)
	foreignYaml, _, err := foreign.YamlFileName(conf.BackupAlbumsPath())
	require.NoError(t, err)
	assert.NoFileExists(t, foreignYaml)
}

func TestProcessUserUploadAlbumLimit(t *testing.T) {
	app, router, conf := NewApiTest()
	ProcessUserUpload(router)
	options := *conf.Options()
	mode := conf.AuthMode()
	t.Cleanup(func() { *conf.Options() = options; conf.SetAuthMode(mode) })
	conf.SetAuthMode(config.AuthModePasswd)
	conf.Options().StoragePath = t.TempDir()
	conf.Options().OriginalsPath = t.TempDir()
	conf.Options().SidecarPath = t.TempDir()
	conf.Options().ImportAllow = ""
	user := entity.UserFixtures.Pointer("alice")
	sess := clientCredentialSession(t, conf, "client", "*", user)
	prefix := "Too Many " + rnd.Base36(6)
	titles := make([]string, MaxUploadAlbums+1)

	for i := range titles {
		titles[i] = fmt.Sprintf("%s %d", prefix, i)
	}

	// The picture is imported and added to the first MaxUploadAlbums albums only.
	token := rnd.Base36(10)
	dir, err := conf.UserUploadPath(user.UserUID, sess.RefID+token)
	require.NoError(t, err)
	filename := filepath.Join(dir, "upload.jpg")
	require.NoError(t, os.WriteFile(filename, NewTestJpeg(t, 157, 107), fs.ModeFile))
	hash := fs.Hash(filename)
	t.Cleanup(func() {
		_ = entity.UnscopedDb().Unscoped().Delete(&entity.Album{}, "album_title LIKE ?", prefix+"%").Error
		file, err := entity.FirstFileByHash(hash)
		if err != nil {
			return
		}
		entity.UnscopedDb().Unscoped().Delete(&entity.PhotoAlbum{}, "photo_uid = ?", file.PhotoUID)
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", file.PhotoID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", file.PhotoID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Photo{}, "id = ?", file.PhotoID)
	})

	body, err := json.Marshal(form.UploadOptions{Albums: titles})
	require.NoError(t, err)

	result := AuthenticatedRequestWithBody(app, http.MethodPut, "/api/v1/users/"+user.UserUID+"/upload/"+token, string(body), sess.AuthToken())
	require.Equal(t, http.StatusOK, result.Code, result.Body.String())

	file, err := entity.FirstFileByHash(hash)
	require.NoError(t, err)

	var count int
	require.NoError(t, entity.UnscopedDb().Model(&entity.Album{}).Where("album_title LIKE ?", prefix+"%").Count(&count).Error)
	assert.Equal(t, MaxUploadAlbums, count)
	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoAlbum{}).Where("photo_uid = ?", file.PhotoUID).Count(&count).Error)
	assert.Equal(t, MaxUploadAlbums, count)
	require.NoError(t, entity.UnscopedDb().Model(&entity.Album{}).Where("album_title = ?", titles[MaxUploadAlbums]).Count(&count).Error)
	assert.Equal(t, 0, count)
}
