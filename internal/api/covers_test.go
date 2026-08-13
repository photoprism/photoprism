package api

import (
	"net/http"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/thumb"

	"github.com/stretchr/testify/assert"
)

func TestCoverSize(t *testing.T) {
	t.Run("Fit", func(t *testing.T) {
		assert.Equal(t, thumb.Fit720, coverSize(thumb.Sizes[thumb.Fit15360]).Name)
		assert.Equal(t, thumb.Fit720, coverSize(thumb.Sizes[thumb.Fit1920]).Name)
		assert.Equal(t, thumb.Fit720, coverSize(thumb.Sizes[thumb.Fit720]).Name)
	})
	t.Run("Tile", func(t *testing.T) {
		// A square request must stay square rather than turn into a fitted image.
		assert.Equal(t, thumb.Tile500, coverSize(thumb.Sizes[thumb.Tile1080]).Name)
		assert.Equal(t, thumb.Tile500, coverSize(thumb.Sizes[thumb.Tile500]).Name)
		assert.Equal(t, thumb.Tile224, coverSize(thumb.Sizes[thumb.Tile224]).Name)
	})
	t.Run("NeverUncached", func(t *testing.T) {
		for name, size := range thumb.Sizes {
			assert.False(t, coverSize(size).Uncached(), "%s", name)
			assert.False(t, coverSize(size).ExceedsLimit(), "%s", name)
		}
	})
}

func TestAlbumCover(t *testing.T) {
	t.Run("InvalidType", func(t *testing.T) {
		app, router, conf := NewApiTest()
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba7/t/"+conf.PreviewToken()+"/xxx")

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("AlbumContainsNoPhotosBecauseIsNotExisting", func(t *testing.T) {
		app, router, conf := NewApiTest()
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/987-986435/t/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("AlbumCouldNotFindOriginal", func(t *testing.T) {
		app, router, conf := NewApiTest()
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba9/t/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/t/xxx/tile_500")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("SizeExceedsLimit", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().ThumbUncached = true
		defer func() { conf.Options().ThumbUncached = false }()
		SetTestCoverFile(t, entity.Album{}, "album_uid = ?", "as6sg6bxpogaaba8", "")
		original := CreateTestAlbumCover(t, "as6sg6bxpogaaba8", "2023/11/IMG_57.jpg")
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/t/"+conf.PreviewToken()+"/fit_15360")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "image/jpeg", r.Header().Get("Content-Type"))
		assert.NotEqual(t, original, r.Body.Bytes())
	})
	t.Run("SizeExceedsLimitAtDefaults", func(t *testing.T) {
		// Matches a stock install, where on-demand rendering is disabled and the cover is
		// therefore served from the thumbnail that indexing pre-generated.
		app, router, conf := NewApiTest()
		SetTestCoverFile(t, entity.Album{}, "album_uid = ?", "as6sg6bxpogaaba8", "")
		original := CreateTestAlbumCover(t, "as6sg6bxpogaaba8", "2023/11/IMG_57.jpg")
		f, err := query.AlbumCoverByUID("as6sg6bxpogaaba8", conf.Settings().Features.Private)
		if err != nil {
			t.Fatal(err)
		}
		CreateTestThumb(t, photoprism.FileName(f.FileRoot, f.FileName), f.FileHash, thumb.SizeFit720)
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/t/"+conf.PreviewToken()+"/fit_15360")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "image/jpeg", r.Header().Get("Content-Type"))
		assert.NotEqual(t, original, r.Body.Bytes())
	})
	t.Run("LimitsSize", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().ThumbUncached = true
		defer func() { conf.Options().ThumbUncached = false }()
		SetTestCoverFile(t, entity.Album{}, "album_uid = ?", "as6sg6bxpogaaba8", "")
		CreateTestAlbumCover(t, "as6sg6bxpogaaba8", "2023/11/IMG_57.jpg")
		AlbumCover(router)
		small := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/t/"+conf.PreviewToken()+"/fit_720")
		large := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/t/"+conf.PreviewToken()+"/fit_1920")
		assert.Equal(t, http.StatusOK, large.Code)
		assert.Equal(t, "image/jpeg", large.Header().Get("Content-Type"))
		assert.Equal(t, small.Body.Bytes(), large.Body.Bytes())
	})
	t.Run("HasCoverFile", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().ThumbUncached = true
		defer func() { conf.Options().ThumbUncached = false }()
		original := CreateTestAlbumCover(t, "as6sg6bxpogaaba8", "2023/11/IMG_57.jpg")
		SetTestCoverFile(t, entity.Album{}, "album_uid = ?", "as6sg6bxpogaaba8", "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/t/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "image/svg+xml", r.Header().Get("Content-Type"))
		assert.NotEqual(t, original, r.Body.Bytes())
	})
}

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
