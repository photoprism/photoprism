package api

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

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
