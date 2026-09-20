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
		// The cap relies on the pre-cached size never being configurable below fit_720.
		cached, onDemand := thumb.SizeCached, thumb.SizeOnDemand
		thumb.SizeCached, thumb.SizeOnDemand = thumb.SizeFit720.Width, thumb.SizeFit720.Width
		defer func() { thumb.SizeCached, thumb.SizeOnDemand = cached, onDemand }()

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
		// Cleared so that the request reaches the cover query, since a cover file assigned by
		// another test would answer with the generic icon before it gets there.
		SetTestCoverFile(t, entity.Album{}, "album_uid = ?", "as6sg6bxpogaaba9", "")
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba9/t/"+conf.PreviewToken()+"/tile_500")
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
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/t/xxx/tile_500")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("SizeExceedsLimit", func(t *testing.T) {
		app, router, conf := NewApiTest()
		SetTestThumbUncached(t, true)
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
		SetTestThumbUncached(t, true)
		SetTestCoverFile(t, entity.Album{}, "album_uid = ?", "as6sg6bxpogaaba8", "")
		CreateTestAlbumCover(t, "as6sg6bxpogaaba8", "2023/11/IMG_57.jpg")
		AlbumCover(router)
		small := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/t/"+conf.PreviewToken()+"/fit_720")
		large := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/t/"+conf.PreviewToken()+"/fit_1920")
		assert.Equal(t, http.StatusOK, large.Code)
		assert.Equal(t, "image/jpeg", large.Header().Get("Content-Type"))
		assert.Equal(t, small.Body.Bytes(), large.Body.Bytes())
	})
	t.Run("ServedInline", func(t *testing.T) {
		// Cover responses are always inline.
		app, router, conf := NewApiTest()
		SetTestThumbUncached(t, true)
		SetTestCoverFile(t, entity.Album{}, "album_uid = ?", "as6sg6bxpogaaba8", "")
		CreateTestAlbumCover(t, "as6sg6bxpogaaba8", "2023/11/IMG_57.jpg")
		AlbumCover(router)
		// Repeated so that the cache hit is covered as well as the render path.
		for range 2 {
			r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/t/"+conf.PreviewToken()+"/tile_500?download=1&name=original")
			assert.Equal(t, http.StatusOK, r.Code)
			assert.Equal(t, "image/jpeg", r.Header().Get("Content-Type"))
			assert.Empty(t, r.Header().Get("Content-Disposition"))
		}
	})
	t.Run("HasCoverFile", func(t *testing.T) {
		app, router, conf := NewApiTest()
		SetTestThumbUncached(t, true)
		original := CreateTestAlbumCover(t, "as6sg6bxpogaaba8", "2023/11/IMG_57.jpg")
		// The flag requires the cover file to still resolve, and another test may have flagged
		// that fixture as missing on its way out.
		CreateTestFileOriginal(t, "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		SetTestCoverFile(t, entity.Album{}, "album_uid = ?", "as6sg6bxpogaaba8", "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/t/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "image/svg+xml", r.Header().Get("Content-Type"))
		// The one icon response clients may cache, since it is chosen from the index rather than
		// returned because something was missing.
		assert.Equal(t, "private, max-age=3600", r.Header().Get("Cache-Control"))
		assert.NotEqual(t, original, r.Body.Bytes())
	})
}
