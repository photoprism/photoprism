package api

import (
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/thumb"
)

// TestDownloadFileName checks sanitization, Unicode names, and output extension enforcement.
func TestDownloadFileName(t *testing.T) {
	cases := []struct {
		name, requested, extension, expected string
	}{
		{"Absent", "", ".jpg", "fallback.jpg"},
		{"Unicode", "旅行_preview.png", ".jpg", "旅行_preview.jpg"},
		{"Trim", "  photo  .png ", ".jpg", "photo.jpg"},
		{"Controls", "photo\r\n\x00.png", ".jpg", "photo.jpg"},
		{"InvalidCharacters", "ph:o?to*.png", ".jpg", "photo.jpg"},
		{"Traversal", "../photo.jpg", ".jpg", "fallback.jpg"},
		{"Slash", "dir/photo.jpg", ".jpg", "fallback.jpg"},
		{"EmptyAfterCleaning", "<>:*?", ".jpg", "fallback.jpg"},
		{"EmptyStem", ".png", ".jpg", "fallback.jpg"},
		{"NoExtension", "photo", ".heic", "photo.heic"},
		{"RawExtension", "photo.jpg", ".CR2", "photo.CR2"},
		{"FinalExtension", "photo.preview.exe", ".png", "photo.preview.png"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/?filename="+url.QueryEscape(tc.requested), nil)
			require.NoError(t, err)
			c := &gin.Context{Request: req}
			assert.Equal(t, tc.expected, DownloadFileName(c, "fallback.jpg", tc.extension))
		})
	}
}

// TestGetDownloadFilename checks custom hash-download names while preserving content and legacy naming.
func TestGetDownloadFilename(t *testing.T) {
	app, router, conf := NewApiTest()
	const hash = "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818"
	original := CreateTestFileOriginal(t, hash)
	f, err := query.FileByHash(hash)
	require.NoError(t, err)
	GetDownload(router)

	t.Run("Custom", func(t *testing.T) {
		r := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+hash+"?t="+conf.DownloadToken()+"&name=share&filename="+url.QueryEscape("旅行.png"))
		require.Equal(t, http.StatusOK, r.Code)
		_, params, err := mime.ParseMediaType(r.Header().Get("Content-Disposition"))
		require.NoError(t, err)
		assert.Equal(t, "旅行"+filepath.Ext(f.FileName), params["filename"])
		assert.Equal(t, original, r.Body.Bytes())
	})
	t.Run("Legacy", func(t *testing.T) {
		for _, name := range []customize.DownloadName{customize.DownloadNameFile, customize.DownloadNameOriginal, customize.DownloadNameShare} {
			r := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+hash+"?t="+conf.DownloadToken()+"&name="+string(name))
			require.Equal(t, http.StatusOK, r.Code)
			_, params, err := mime.ParseMediaType(r.Header().Get("Content-Disposition"))
			require.NoError(t, err)
			assert.Equal(t, f.DownloadName(name, 0), params["filename"])
		}
	})
	t.Run("InvalidFallsBack", func(t *testing.T) {
		r := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+hash+"?t="+conf.DownloadToken()+"&name=original&filename=..%2Fphoto.png")
		require.Equal(t, http.StatusOK, r.Code)
		_, params, err := mime.ParseMediaType(r.Header().Get("Content-Disposition"))
		require.NoError(t, err)
		assert.Equal(t, f.DownloadName(customize.DownloadNameOriginal, 0), params["filename"])
	})
	t.Run("Unauthorized", func(t *testing.T) {
		conf.SetAuthMode(config.AuthModePasswd)
		t.Cleanup(func() { conf.SetAuthMode(config.AuthModePublic) })
		r := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+hash+"?t=invalid&filename=photo.jpg")
		assert.Equal(t, http.StatusForbidden, r.Code)
		assert.Empty(t, r.Header().Get("Content-Disposition"))
	})
}

// TestGetThumbFilename checks request-specific names without changing cached thumbnail identity.
func TestGetThumbFilename(t *testing.T) {
	app, router, conf := NewApiTest()
	const hash = "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818"
	CreateTestFileOriginal(t, hash)
	SetTestThumbUncached(t, true)
	SetTestThumbSizes(t, thumb.SizeFit2048.Width, thumb.SizeFit2048.Width)
	cache := get.ThumbCache()
	key := CacheKey("thumbs", hash, string(thumb.Fit720))
	cache.Delete(key)
	t.Cleanup(func() { cache.Delete(key) })
	GetThumb(router)

	base := "/api/v1/t/" + hash + "/" + conf.PreviewToken() + "/fit_720?download=1"
	miss := PerformRequest(app, http.MethodGet, base+"&filename="+url.QueryEscape("旅行_fit_720.png"))
	require.Equal(t, http.StatusOK, miss.Code)
	require.Equal(t, "image/jpeg", miss.Header().Get("Content-Type"))
	_, params, err := mime.ParseMediaType(miss.Header().Get("Content-Disposition"))
	require.NoError(t, err)
	assert.Equal(t, "旅行_fit_720.jpg", params["filename"])
	cached, found := cache.Get(key)
	require.True(t, found)
	assert.NotContains(t, cached.(ThumbCache).ShareName, "旅行")

	hit := PerformRequest(app, http.MethodGet, base+"&filename="+url.QueryEscape("旅行_fit_720.png"))
	assert.Equal(t, miss.Header().Get("Content-Disposition"), hit.Header().Get("Content-Disposition"))
	assert.Equal(t, miss.Body.Bytes(), hit.Body.Bytes())

	other := PerformRequest(app, http.MethodGet, base+"&filename=another.heic")
	_, params, err = mime.ParseMediaType(other.Header().Get("Content-Disposition"))
	require.NoError(t, err)
	assert.Equal(t, "another.jpg", params["filename"])
	assert.Equal(t, miss.Body.Bytes(), other.Body.Bytes())
	after, found := cache.Get(key)
	require.True(t, found)
	assert.Equal(t, cached, after)

	legacy := PerformRequest(app, http.MethodGet, base)
	_, params, err = mime.ParseMediaType(legacy.Header().Get("Content-Disposition"))
	require.NoError(t, err)
	assert.Equal(t, cached.(ThumbCache).ShareName, params["filename"])

	inline := PerformRequest(app, http.MethodGet, "/api/v1/t/"+hash+"/"+conf.PreviewToken()+"/fit_720?filename=another.png")
	assert.Empty(t, inline.Header().Get("Content-Disposition"))

	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(config.AuthModePublic) })
	denied := PerformRequest(app, http.MethodGet, "/api/v1/t/"+hash+"/invalid/fit_720?download=1&filename=photo.jpg")
	assert.Equal(t, http.StatusForbidden, denied.Code)
	assert.Empty(t, denied.Header().Get("Content-Disposition"))
}
