package api

import (
	"net/http"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/thumb"

	"github.com/stretchr/testify/assert"
)

func TestLabelCover(t *testing.T) {
	t.Run("InvalidType", func(t *testing.T) {
		app, router, conf := NewApiTest()
		LabelCover(router)
		r := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/t/"+conf.PreviewToken()+"/xxx")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidLabel", func(t *testing.T) {
		app, router, conf := NewApiTest()
		LabelCover(router)
		r := PerformRequest(app, "GET", "/api/v1/labels/xxx/t/"+conf.PreviewToken()+"/tile_500")

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("CouldNotFindOriginal", func(t *testing.T) {
		app, router, conf := NewApiTest()
		// Cleared so that the request reaches the cover query, since a cover file assigned by
		// another test would answer with the generic icon before it gets there.
		SetTestCoverFile(t, entity.Label{}, "label_uid = ?", "ls6sg6b1wowuy3c2", "")
		LabelCover(router)
		r := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/t/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "image/svg+xml", r.Header().Get("Content-Type"))
		// An icon returned because something was missing must not be cached, unlike the one an
		// assigned cover file returns.
		assert.Empty(t, r.Header().Get("Cache-Control"))
	})
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		LabelCover(router)
		r := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c3/t/xxx/tile_500")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("SizeExceedsLimit", func(t *testing.T) {
		app, router, conf := NewApiTest()
		SetTestThumbUncached(t, true)
		SetTestCoverFile(t, entity.Label{}, "label_uid = ?", "ls6sg6b1wowuy3c2", "")
		original := CreateTestLabelCover(t, "ls6sg6b1wowuy3c2", "2007/12/PhotoWithEditedAt.jpg")
		LabelCover(router)
		r := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/t/"+conf.PreviewToken()+"/fit_15360")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "image/jpeg", r.Header().Get("Content-Type"))
		assert.NotEqual(t, original, r.Body.Bytes())
	})
	t.Run("LimitsSize", func(t *testing.T) {
		app, router, conf := NewApiTest()
		SetTestThumbUncached(t, true)
		SetTestCoverFile(t, entity.Label{}, "label_uid = ?", "ls6sg6b1wowuy3c2", "")
		CreateTestLabelCover(t, "ls6sg6b1wowuy3c2", "2007/12/PhotoWithEditedAt.jpg")
		LabelCover(router)
		small := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/t/"+conf.PreviewToken()+"/fit_720")
		large := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/t/"+conf.PreviewToken()+"/fit_1920")
		assert.Equal(t, http.StatusOK, large.Code)
		assert.Equal(t, "image/jpeg", large.Header().Get("Content-Type"))
		assert.Equal(t, small.Body.Bytes(), large.Body.Bytes())
	})
	t.Run("ServedInline", func(t *testing.T) {
		// Cover responses are always inline.
		app, router, conf := NewApiTest()
		SetTestThumbUncached(t, true)
		SetTestCoverFile(t, entity.Label{}, "label_uid = ?", "ls6sg6b1wowuy3c2", "")
		CreateTestLabelCover(t, "ls6sg6b1wowuy3c2", "2007/12/PhotoWithEditedAt.jpg")
		LabelCover(router)
		// Repeated so that the cache hit is covered as well as the render path.
		for i := range 2 {
			r := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/t/"+conf.PreviewToken()+"/tile_500?download=1&name=original")
			assert.Equal(t, http.StatusOK, r.Code)
			assert.Equal(t, "image/jpeg", r.Header().Get("Content-Type"))
			assert.Empty(t, r.Header().Get("Content-Disposition"))

			// Without this the second pass would render again and prove nothing about the hit.
			_, hit := get.CoverCache().Get(CacheKey(labelCover, "ls6sg6b1wowuy3c2", thumb.Tile500.String()))
			assert.True(t, hit, "pass %d must leave the rendition cached", i)
		}
	})
	t.Run("HasCoverFile", func(t *testing.T) {
		app, router, conf := NewApiTest()
		SetTestThumbUncached(t, true)
		original := CreateTestLabelCover(t, "ls6sg6b1wowuy3c2", "2007/12/PhotoWithEditedAt.jpg")
		// The flag requires the cover file to still resolve, and another test may have flagged
		// that fixture as missing on its way out.
		CreateTestFileOriginal(t, "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		SetTestCoverFile(t, entity.Label{}, "label_uid = ?", "ls6sg6b1wowuy3c2", "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		LabelCover(router)
		r := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/t/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "image/svg+xml", r.Header().Get("Content-Type"))
		assert.Equal(t, "private, max-age=3600", r.Header().Get("Cache-Control"))
		assert.NotEqual(t, original, r.Body.Bytes())
	})
}
