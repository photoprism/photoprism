package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestDownload_AuthLimit checks the answer to download requests with a header token.
func TestDownload_AuthLimit(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(config.AuthModePublic) })
	GetDownload(router)
	DownloadAlbum(router)
	GetPhotoDownload(router)
	ZipDownload(router)

	origLimit := limiter.Auth
	t.Cleanup(func() { limiter.Auth = origLimit })
	limiter.Auth = limiter.NewLimit(rate.Every(24*time.Hour), 1)

	exhaust := func(clientIp string) {
		for !limiter.Auth.Reject(clientIp) {
			limiter.Auth.Reserve(clientIp)
		}
	}

	request := func(clientIp, path string) *httptest.ResponseRecorder {
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		req.RemoteAddr = clientIp + ":1234"
		header.SetAuthorization(req, rnd.AuthToken())
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		return w
	}

	routes := []string{
		"/api/v1/dl/3cad9168fa6acc5c5c2965ddf6ec465ca42fd818",
		"/api/v1/albums/" + entity.AlbumFixtures.Get("christmas2030").AlbumUID + "/dl",
		"/api/v1/photos/" + entity.PhotoFixtures.Get("Photo01").PhotoUID + "/dl",
		"/api/v1/zip/zzLimitTest.zip",
	}

	t.Run("OverLimit", func(t *testing.T) {
		exhaust("198.51.100.81")
		for _, path := range routes {
			assert.Equal(t, http.StatusTooManyRequests, request("198.51.100.81", path).Code, path)
		}
	})
	t.Run("UnderLimit", func(t *testing.T) {
		for _, path := range routes {
			assert.Equal(t, http.StatusForbidden, request("198.51.100.83", path).Code, path)
		}
	})
	t.Run("OverLimitWithCoarseToken", func(t *testing.T) {
		// A download token still authorizes the request, as it is not checked against the limit.
		exhaust("198.51.100.82")
		prev := tokens.CoarseDownload
		tokens.CoarseDownload = "zzCoarseDownloadTestToken"
		t.Cleanup(func() { tokens.CoarseDownload = prev })
		r := request("198.51.100.82", routes[0]+"?t="+tokens.CoarseDownload)
		assert.NotEqual(t, http.StatusTooManyRequests, r.Code)
		assert.NotEqual(t, http.StatusForbidden, r.Code)
	})
	t.Run("OverLimitWithSignedToken", func(t *testing.T) {
		exhaust("198.51.100.84")
		signed := tokens.SignDownload(entity.SessionFixtures.Pointer("alice").ID)
		require.NotEmpty(t, signed)
		r := request("198.51.100.84", routes[0]+"?t="+signed)
		assert.NotEqual(t, http.StatusTooManyRequests, r.Code)
		assert.NotEqual(t, http.StatusForbidden, r.Code)
	})
}
