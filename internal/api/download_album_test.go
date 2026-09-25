package api

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/entity"
)

func TestAlbumDownloadName(t *testing.T) {
	t.Run("File", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/albums?name=file", nil)
		assert.NoError(t, err)

		c := &gin.Context{
			Request: req,
		}

		assert.Equal(t, customize.DownloadNameFile, AlbumDownloadName(c))
	})
	t.Run("Share", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/albums?name=share", nil)
		assert.NoError(t, err)

		c := &gin.Context{
			Request: req,
		}

		assert.Equal(t, customize.DownloadNameShare, AlbumDownloadName(c))
	})
	t.Run("Original", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/albums?name=original", nil)
		assert.NoError(t, err)

		c := &gin.Context{
			Request: req,
		}

		assert.Equal(t, customize.DownloadNameOriginal, AlbumDownloadName(c))
	})
}

// TestDownloadAlbum covers album download admission and archive contents.
func TestDownloadAlbum(t *testing.T) {
	t.Run("DownloadNotExistingAlbum", func(t *testing.T) {
		app, router, conf := NewApiTest()

		DownloadAlbum(router)

		r := PerformRequest(app, "GET", "/api/v1/albums/5678/dl?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("DownloadExistingAlbum", func(t *testing.T) {
		app, router, conf := NewApiTest()

		DownloadAlbum(router)

		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/dl?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("DownloadDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Settings().Features.Download = false

		DownloadAlbum(router)

		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/dl?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusForbidden, r.Code)

		conf.Settings().Features.Download = true
	})
	t.Run("SignedAdminToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DownloadAlbum(router)

		q := tokens.SignDownload(entity.SessionFixtures.Get("alice").ID)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/dl?t="+q)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("SignedVisitorTokenSharedAlbum", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DownloadAlbum(router)

		// The visitor fixture has redeemed a share for as6sg6bxpogaaba8.
		q := tokens.SignDownload(entity.SessionFixtures.Get("visitor").ID)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/dl?t="+q)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("SignedVisitorTokenUnsharedAlbum", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DownloadAlbum(router)

		// The visitor has no share for as6sg6bxpogaaba9, so the album is reported as not found.
		q := tokens.SignDownload(entity.SessionFixtures.Get("visitor").ID)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba9/dl?t="+q)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	// A credential is admitted on the album row by its client role, not by holding a share, while
	// the token still answers for pictures - the resource this transport is authorized on.
	t.Run("SignedClientToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		options := *conf.Options()
		t.Cleanup(func() {
			*conf.Options() = options
			conf.Propagate()
		})
		conf.SetAuthMode(config.AuthModePasswd)
		conf.Options().OriginalsPath = t.TempDir()
		conf.Propagate()

		DownloadAlbum(router)

		photo := entity.NewPhoto(false)
		require.NoError(t, photo.Save())
		t.Cleanup(func() {
			entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
			entity.UnscopedDb().Unscoped().Delete(&photo)
		})

		file := &entity.File{
			PhotoID: photo.ID, PhotoUID: photo.PhotoUID,
			FileRoot: entity.RootOriginals, FileName: "client-download.jpg",
			FileHash: "aed71fbf713047751c3166ac8bd718c2fdbb909f",
			FileType: "jpg", MediaType: entity.MediaImage, FilePrimary: true,
		}
		require.NoError(t, file.Create())
		t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(file) })
		original := CreateTestOriginal(t, file)

		album := entity.NewAlbum("Signed Client Download", entity.AlbumManual)
		require.NoError(t, album.Create())
		t.Cleanup(func() {
			entity.UnscopedDb().Unscoped().Delete(&entity.PhotoAlbum{}, "album_uid = ?", album.AlbumUID)
			entity.UnscopedDb().Unscoped().Delete(album)
			entity.FlushAlbumCache()
		})
		require.NoError(t, entity.NewPhotoAlbum(file.PhotoUID, album.AlbumUID).Create())

		sess := clientCredentialSession(t, conf, acl.RoleClient.String(), "photos", nil)
		r := PerformRequest(app, http.MethodGet,
			"/api/v1/albums/"+album.AlbumUID+"/dl?name=file&t="+tokens.SignDownload(sess.ID))
		require.Equal(t, http.StatusOK, r.Code)

		archive, err := zip.NewReader(bytes.NewReader(r.Body.Bytes()), int64(r.Body.Len()))
		require.NoError(t, err)
		require.Len(t, archive.File, 1)
		assert.Equal(t, filepath.Base(file.FileName), archive.File[0].Name)

		entry, err := archive.File[0].Open()
		require.NoError(t, err)
		contents, err := io.ReadAll(entry)
		require.NoError(t, entry.Close())
		require.NoError(t, err)
		assert.Equal(t, original, contents)
	})
	t.Run("SignedClientTokenOutOfScope", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DownloadAlbum(router)

		sess := clientCredentialSession(t, conf, acl.RoleClient.String(), "metrics", nil)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/dl?t="+tokens.SignDownload(sess.ID))
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("SignedClientTokenForARestrictedAccount", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DownloadAlbum(router)

		// The attached account holds no share for this album, and a client role does not lift it.
		sess := clientCredentialSession(t, conf, acl.RoleClient.String(), "photos",
			entity.UserFixtures.Pointer("guest"))
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/dl?t="+tokens.SignDownload(sess.ID))
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("NonJwtHeaderCannotDownloadWithoutToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DownloadAlbum(router)

		// Header auth on downloads is restricted to cluster JWTs; a regular session bearer cannot use it
		// and must present a "?t=" token (which it can, being a persisted session). With no "?t=" it is
		// denied.
		token := AuthenticateAdmin(app, router)
		r := AuthenticatedRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/dl", token)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("CoarseDownloadTokenWorks", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DownloadAlbum(router)

		// A configured static token is not session-bound, so it downloads the public, non-private
		// subset of the album. It only exists when an operator sets one; nothing is auto-generated.
		const coarse = "static-download-token"
		orig := tokens.CoarseDownload
		tokens.CoarseDownload = coarse
		defer func() { tokens.CoarseDownload = orig }()
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/dl?t="+coarse)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
