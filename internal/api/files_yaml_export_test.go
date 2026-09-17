package api

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestStoredYamlExports checks indexed metadata eligibility across direct and archive downloads.
func TestStoredYamlExports(t *testing.T) {
	app, router, conf := NewApiTest()
	GetDownload(router)
	GetPhotoDownload(router)
	DownloadAlbum(router)
	ZipCreate(router)
	options := *conf.Options()
	settings := *conf.Settings()
	mode := conf.AuthMode()
	t.Cleanup(func() {
		*conf.Options() = options
		*conf.Settings() = settings
		conf.SetAuthMode(mode)
		conf.Propagate()
	})
	conf.SetAuthMode(config.AuthModePasswd)
	conf.Options().OriginalsPath = t.TempDir()
	conf.Options().SidecarPath = t.TempDir()
	conf.Options().TempPath = t.TempDir()
	conf.Propagate()
	conf.Settings().Download.MediaSidecar = true
	conf.Settings().Albums.Download.MediaSidecar = true
	const note = "stored-metadata-note-control"
	original := NewTestJpeg(t, 73, 59)
	assert.NotContains(t, string(original), note)
	name := filepath.Join(conf.OriginalsPath(), "export-control.jpg")
	require.NoError(t, os.WriteFile(name, original, fs.ModeFile))
	seed := entity.NewPhoto(false)
	seed.PhotoUID = rnd.GenerateUID(entity.PhotoUID)
	seed.PhotoTitle = "Stored Metadata Export Control"
	seed.GetDetails().Notes = note
	require.NoError(t, seed.SaveAsYaml(filepath.Join(conf.OriginalsPath(), "export-control.yml")))
	media, err := photoprism.NewMediaFile(name)
	require.NoError(t, err)
	related, err := media.RelatedFiles(false)
	require.NoError(t, err)
	ind := photoprism.NewIndex(conf, photoprism.NewConvert(conf), photoprism.NewFiles(), photoprism.NewPhotos())
	result := photoprism.IndexRelated(related, ind, photoprism.IndexOptionsNone(conf))
	require.True(t, result.Success(), "%+v", result)
	photo, err := query.PhotoByUID(result.PhotoUID)
	require.NoError(t, err)
	require.NoError(t, photo.Update("PhotoQuality", 3))
	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.PhotoAlbum{}, "photo_uid = ?", photo.PhotoUID)
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&photo)
	})
	var files entity.Files
	require.NoError(t, entity.Db().Where("photo_id = ?", photo.ID).Find(&files).Error)
	var yamlFile, imageFile *entity.File
	for i := range files {
		if files[i].FileType == "yml" {
			yamlFile = &files[i]
		}
		if files[i].FileType == "jpg" {
			imageFile = &files[i]
		}
	}
	require.NotNil(t, yamlFile, "normal indexing must register the YAML")
	require.NotNil(t, imageFile)
	yamlData, err := os.ReadFile(filepath.Join(conf.OriginalsPath(), yamlFile.FileName)) //nolint:gosec // Test reads a controlled fixture or generated output.
	require.NoError(t, err)
	require.Contains(t, string(yamlData), note)
	const albumUID = "as6sg6bxpogaaba8"
	require.NoError(t, entity.NewPhotoAlbum(photo.PhotoUID, albumUID).Create())
	visitor := entity.SessionFixtures.Get("visitor")
	visitorToken := tokens.DownloadToken(visitor.ID)
	guestUser := entity.NewUser()
	guestUser.UserName = "export-guest-" + rnd.Base36(6)
	guestUser.UserRole = "guest"
	guestUser.CanLogin = true
	require.NoError(t, guestUser.Create())
	share := entity.NewUserShare(guestUser.UserUID, albumUID, entity.PermView, nil)
	require.NoError(t, share.Save())
	guestUser.RefreshShares()
	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(share)
		entity.UnscopedDb().Unscoped().Delete(&entity.UserDetails{}, "user_uid = ?", guestUser.UserUID)
		entity.UnscopedDb().Unscoped().Delete(&entity.UserSettings{}, "user_uid = ?", guestUser.UserUID)
		entity.UnscopedDb().Unscoped().Delete(guestUser)
	})
	guest := entity.NewSession(3600, 0).SetUser(guestUser)
	require.NoError(t, guest.Create())
	t.Cleanup(func() { require.NoError(t, guest.Delete()) })
	guestToken := tokens.DownloadToken(guest.ID)
	reader := clientCredentialSession(t, conf, "client", "photos albums files", nil)
	readerToken := tokens.DownloadToken(reader.ID)
	// archiveContents reads every populated ZIP entry for metadata and original controls.
	archiveContents := func(data []byte) string {
		z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		require.NoError(t, err)
		var out bytes.Buffer
		for _, f := range z.File {
			r, err := f.Open()
			require.NoError(t, err)
			_, err = io.Copy(&out, io.LimitReader(r, 8*1024*1024))
			require.NoError(t, err)
			require.NoError(t, r.Close())
		}
		return out.String()
	}
	t.Run("CoarseToken", func(t *testing.T) {
		conf.Options().DownloadToken = "coarse-export-test-token"
		conf.Propagate()
		defer func() { conf.Options().DownloadToken = options.DownloadToken; conf.Propagate() }()
		denied := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+yamlFile.FileHash+"?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusNotFound, denied.Code)
		assert.NotContains(t, denied.Body.String(), note)
		allowed := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+imageFile.FileHash+"?t="+conf.DownloadToken())
		require.Equal(t, http.StatusOK, allowed.Code)
		assert.Equal(t, original, allowed.Body.Bytes())
	})
	t.Run("Direct", func(t *testing.T) {
		denied := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+yamlFile.FileHash+"?t="+visitorToken)
		assert.Equal(t, http.StatusNotFound, denied.Code)
		assert.NotContains(t, denied.Body.String(), note)
		allowed := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+yamlFile.FileHash+"?t="+readerToken)
		require.Equal(t, http.StatusOK, allowed.Code)
		assert.Contains(t, allowed.Body.String(), note)
		img := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+imageFile.FileHash+"?t="+visitorToken)
		require.Equal(t, http.StatusOK, img.Code)
		assert.Equal(t, original, img.Body.Bytes())
	})
	t.Run("RegisteredGuest", func(t *testing.T) {
		r := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+yamlFile.FileHash+"?t="+guestToken)
		require.Equal(t, http.StatusOK, r.Code)
		assert.Contains(t, r.Body.String(), note)
	})
	t.Run("FilesOnlyReader", func(t *testing.T) {
		sess := clientCredentialSession(t, conf, "client", "read files", nil)
		r := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+yamlFile.FileHash+"?t="+tokens.DownloadToken(sess.ID))
		require.Equal(t, http.StatusOK, r.Code)
		assert.Contains(t, r.Body.String(), note)
	})
	t.Run("Primary", func(t *testing.T) {
		require.NoError(t, imageFile.Update("FilePrimary", false))
		require.NoError(t, yamlFile.Update("FilePrimary", true))
		defer func() {
			require.NoError(t, yamlFile.Update("FilePrimary", false))
			require.NoError(t, imageFile.Update("FilePrimary", true))
		}()
		denied := PerformRequest(app, http.MethodGet, "/api/v1/photos/"+photo.PhotoUID+"/dl?t="+visitorToken)
		assert.Equal(t, http.StatusNotFound, denied.Code)
		assert.NotContains(t, denied.Body.String(), note)
		allowed := PerformRequest(app, http.MethodGet, "/api/v1/photos/"+photo.PhotoUID+"/dl?t="+readerToken)
		require.Equal(t, http.StatusOK, allowed.Code)
		assert.Contains(t, allowed.Body.String(), note)
	})
	for _, sidecars := range []bool{false, true} {
		conf.Settings().Download.MediaSidecar = sidecars
		conf.Settings().Albums.Download.MediaSidecar = sidecars
		for _, tc := range []struct {
			name, auth, download string
			permitted            bool
		}{
			{"Visitor", visitor.AuthToken(), visitorToken, false},
			{"RegisteredGuest", guest.AuthToken(), guestToken, true},
			{"Reader", reader.AuthToken(), readerToken, true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				album := PerformRequest(app, http.MethodGet, "/api/v1/albums/"+albumUID+"/dl?t="+tc.download)
				require.Equal(t, http.StatusOK, album.Code)
				contents := archiveContents(album.Body.Bytes())
				assert.Contains(t, contents, string(original))
				if tc.permitted && sidecars {
					assert.Contains(t, contents, note)
				} else {
					assert.NotContains(t, contents, note)
				}
				created := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/zip", `{"photos":["`+photo.PhotoUID+`"]}`, tc.auth)
				require.Equal(t, http.StatusOK, created.Code, created.Body.String())
				zipName := gjson.GetBytes(created.Body.Bytes(), "filename").String()
				require.NotEmpty(t, zipName)
				data, err := os.ReadFile(filepath.Join(conf.TempPath(), fs.ZipDir, zipName)) //nolint:gosec // Test reads its generated archive.
				require.NoError(t, err)
				contents = archiveContents(data)
				assert.Contains(t, contents, string(original))
				if tc.permitted && sidecars {
					assert.Contains(t, contents, note)
				} else {
					assert.NotContains(t, contents, note)
				}
			})
		}
	}
	require.NoError(t, entity.Db().Where("photo_uid = ? AND album_uid = ?", photo.PhotoUID, albumUID).Delete(&entity.PhotoAlbum{}).Error)
	denied := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+imageFile.FileHash+"?t="+visitorToken)
	assert.Equal(t, http.StatusNotFound, denied.Code)
}
