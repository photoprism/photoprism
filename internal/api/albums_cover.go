package api

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Namespace for caching and logs.
const albumCover = "album-cover"

// coverSize limits a cover image to the largest size that is always pre-rendered.
// Cropped sizes are capped to a cropped size so a square request stays square.
func coverSize(size thumb.Size) thumb.Size {
	if size.Fit {
		return size.Limit(thumb.SizeFit720)
	}

	return size.Limit(thumb.SizeTile500)
}

// AlbumCover returns an album cover image.
//
//	@Summary		returns an album cover image
//	@Id				AlbumCover
//	@Description	Returns a generic placeholder icon when a cover file is assigned to the album; request the picture from /api/v1/t/{hash}/{token}/{size} instead, using its hash. Covers are always served inline; use the download endpoints to obtain a file.
//	@Produce		image/jpeg
//	@Produce		image/svg+xml
//	@Tags			Images, Albums
//	@Failure		403		{file}	image/svg+xml
//	@Failure		200		{file}	image/svg+xml
//	@Success		200		{file}	image/jpg
//	@Param			uid		path	string	true	"Album UID"
//	@Param			token	path	string	true	"user-specific security token provided with session or 'public' when running PhotoPrism in public mode"
//	@Param			size	path	string	true	"cover image size; larger sizes are reduced"	Enums(tile_50, tile_100, left_224, right_224, tile_224, tile_500, fit_720)
//	@Router			/api/v1/albums/{uid}/t/{token}/{size} [get]
func AlbumCover(router *gin.RouterGroup) {
	router.GET("/albums/:uid/t/:token/:size", func(c *gin.Context) {
		if InvalidPreviewToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", albumIconSvg)
			return
		}

		start := time.Now()
		conf := get.Config()
		thumbName := thumb.Name(clean.Token(c.Param("size")))
		uid := clean.UID(c.Param("uid"))

		size, ok := thumb.Sizes[thumbName]

		if !ok {
			log.Errorf("%s: invalid size %s", albumCover, clean.Log(thumbName.String()))
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		// Serve a generic cover if the album has a cover file, as clients
		// must request it from the thumbnail endpoint by its hash.
		if CachedCoverHasThumb(albumCover, uid, query.AlbumHasThumb) {
			log.Debugf("%s: %s has a cover file, use the thumbnail endpoint instead", albumCover, uid)
			AddCoverCacheHeader(c)
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		size = coverSize(size)
		thumbName = size.Name

		cache := get.CoverCache()
		cacheKey := CacheKey(albumCover, uid, string(thumbName))

		if cacheData, ok := cache.Get(cacheKey); ok {
			log.Tracef("api: cache hit for %s [%s]", cacheKey, time.Since(start))

			cached := cacheData.(ThumbCache)

			if !fs.FileExists(cached.FileName) {
				log.Errorf("%s: %s not found", albumCover, uid)
				c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
				return
			}

			AddCoverCacheHeader(c)
			c.File(cached.FileName)

			return
		}

		f, err := query.AlbumCoverByUID(uid, conf.Settings().Features.Private)

		if err != nil {
			log.Debugf("%s: %s contains no pictures, using generic cover", albumCover, uid)
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("%s: found no original for %s", albumCover, clean.Log(f.FileName))
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			log.Warnf("%s: %s is missing", albumCover, clean.Log(f.FileName))
			logErr(albumCover, f.Update("FileMissing", true))
			return
		}

		var thumbnail string

		if conf.ThumbUncached() {
			thumbnail, err = thumb.FromFile(fileName, f.FileHash, conf.ThumbCachePath(), size.Width, size.Height, f.FileOrientation, size.Options...)
		} else {
			thumbnail, err = thumb.FromCache(fileName, f.FileHash, conf.ThumbCachePath(), size.Width, size.Height, size.Options...)
		}

		if err != nil {
			log.Errorf("%s: %s", albumCover, err)
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		} else if thumbnail == "" {
			log.Errorf("%s: %s has empty thumb name - you may have found a bug", albumCover, filepath.Base(fileName))
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		cache.SetDefault(cacheKey, ThumbCache{FileName: thumbnail})
		log.Debugf("cached %s [%s]", cacheKey, time.Since(start))

		AddCoverCacheHeader(c)
		c.File(thumbnail)
	})
}
