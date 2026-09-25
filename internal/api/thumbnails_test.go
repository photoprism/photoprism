package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/thumb"
)

func TestGetThumb(t *testing.T) {
	t.Run("InvalidType", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/1/"+conf.PreviewToken()+"/xxx")

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("WrongHash", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/1/"+conf.PreviewToken()+"/tile_500")

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("WrongFile", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818/"+conf.PreviewToken()+"/fit_7680")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("WrongToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818/xxx/tile_500")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("NoJPEG", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/pcad9168fa6acc5c5ba965adf6ec465ca42fd819/"+conf.PreviewToken()+"/fit_7680")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("FileError", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/acad9168fa6acc5c5c2965ddf6ec465ca42fd832/"+conf.PreviewToken()+"/fit_7680")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/46f5b5c0c027f0c1b15136644f404c57210bf20c-016014058037/"+conf.PreviewToken()+"/tile_160")

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("SizeExceedsLimit", func(t *testing.T) {
		app, router, conf := NewApiTest()
		SetTestThumbUncached(t, true)
		original := CreateTestFileOriginal(t, "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818/"+conf.PreviewToken()+"/fit_15360")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "image/jpeg", r.Header().Get("Content-Type"))
		assert.NotEqual(t, original, r.Body.Bytes())
	})
	t.Run("SizeExceedsLimitAtDefaults", func(t *testing.T) {
		// Matches a stock install, where on-demand rendering is disabled and an oversized request
		// is answered from the largest size indexing pre-generates. Asserting on the content of
		// that size is what separates it from the on-demand limit, which is higher.
		app, router, conf := NewApiTest()
		SetTestThumbSizes(t, thumb.SizeFit1920.Width, thumb.SizeFit7680.Width)
		fileHash := "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818"
		original := CreateTestFileOriginal(t, fileHash)
		_, precached := thumb.Find(conf.ThumbSizePrecached())
		require.Equal(t, thumb.Fit1920, precached.Name)
		cached := WriteTestThumb(t, fileHash, precached)

		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/"+fileHash+"/"+conf.PreviewToken()+"/fit_15360")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "image/jpeg", r.Header().Get("Content-Type"))
		assert.Equal(t, cached, r.Body.Bytes(), "the pre-generated size must be served unchanged")
		assert.NotEqual(t, original, r.Body.Bytes())
	})
	t.Run("SizeExceedsLimitDownload", func(t *testing.T) {
		// Unlike the cover endpoints, this one honors ?download=, and the attachment must be the
		// reduced size rather than the original. The attachment path skips the shortcut that
		// serves an existing rendition, so it reaches the cache lookup instead.
		app, router, conf := NewApiTest()
		SetTestThumbSizes(t, thumb.SizeFit1920.Width, thumb.SizeFit7680.Width)
		fileHash := "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818"
		original := CreateTestFileOriginal(t, fileHash)
		_, precached := thumb.Find(conf.ThumbSizePrecached())
		require.Equal(t, thumb.Fit1920, precached.Name)
		cached := WriteTestThumb(t, fileHash, precached)

		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/"+fileHash+"/"+conf.PreviewToken()+"/fit_15360?download=1")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "image/jpeg", r.Header().Get("Content-Type"))
		assert.Contains(t, r.Header().Get("Content-Disposition"), "attachment")
		assert.Equal(t, cached, r.Body.Bytes(), "the pre-generated size must be served unchanged")
		assert.NotEqual(t, original, r.Body.Bytes())
	})
	t.Run("SmallerOriginalFitsLargerSize", func(t *testing.T) {
		// FitSizes is neither square nor a superset of the fit sizes, so the smallest
		// fitting size can resolve above the limit.
		app, router, conf := NewApiTest()
		SetTestThumbUncached(t, true)
		SetTestThumbSizes(t, thumb.SizeFit2048.Width, thumb.SizeFit2048.Width)

		fileHash := "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818"
		original := CreateTestFileOriginal(t, fileHash)
		SetTestFileBounds(t, fileHash, 2000, 2000)

		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/"+fileHash+"/"+conf.PreviewToken()+"/fit_2048")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "image/jpeg", r.Header().Get("Content-Type"))
		assert.NotEqual(t, original, r.Body.Bytes())
	})
}
