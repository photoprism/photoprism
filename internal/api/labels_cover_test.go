package api

import (
	"net/http"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"

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
		LabelCover(router)
		// r := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c3/t/"+conf.PreviewToken()+"/tile_500")
		// ls6sg6b1wowuy3c2
		r := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/t/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
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
		conf.Options().ThumbUncached = true
		defer func() { conf.Options().ThumbUncached = false }()
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
		conf.Options().ThumbUncached = true
		defer func() { conf.Options().ThumbUncached = false }()
		SetTestCoverFile(t, entity.Label{}, "label_uid = ?", "ls6sg6b1wowuy3c2", "")
		CreateTestLabelCover(t, "ls6sg6b1wowuy3c2", "2007/12/PhotoWithEditedAt.jpg")
		LabelCover(router)
		small := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/t/"+conf.PreviewToken()+"/fit_720")
		large := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/t/"+conf.PreviewToken()+"/fit_1920")
		assert.Equal(t, http.StatusOK, large.Code)
		assert.Equal(t, "image/jpeg", large.Header().Get("Content-Type"))
		assert.Equal(t, small.Body.Bytes(), large.Body.Bytes())
	})
	t.Run("HasCoverFile", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().ThumbUncached = true
		defer func() { conf.Options().ThumbUncached = false }()
		original := CreateTestLabelCover(t, "ls6sg6b1wowuy3c2", "2007/12/PhotoWithEditedAt.jpg")
		SetTestCoverFile(t, entity.Label{}, "label_uid = ?", "ls6sg6b1wowuy3c2", "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		LabelCover(router)
		r := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/t/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "image/svg+xml", r.Header().Get("Content-Type"))
		assert.NotEqual(t, original, r.Body.Bytes())
	})
}
