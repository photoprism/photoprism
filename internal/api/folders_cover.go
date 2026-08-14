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

const (
	folderCover = "folder-cover"
)

// FolderCover returns a folder cover image.
//
//	@Summary		returns a folder cover image
//	@Id				FolderCover
//	@Description	Covers are always served inline; use the download endpoints to obtain a file.
//	@Produce		image/jpeg
//	@Produce		image/svg+xml
//	@Tags			Images, Folders
//	@Failure		403		{file}	image/svg+xml
//	@Failure		200		{file}	image/svg+xml
//	@Success		200		{file}	image/jpg
//	@Param			uid		path	string	true	"folder uid"
//	@Param			token	path	string	true	"user-specific security token provided with session or 'public' when running PhotoPrism in public mode"
//	@Param			size	path	string	true	"cover image size; larger sizes are reduced"	Enums(tile_50, tile_100, left_224, right_224, tile_224, tile_500, fit_720)
//	@Router			/api/v1/folders/t/{uid}/{token}/{size} [get]
func FolderCover(router *gin.RouterGroup) {
	router.GET("/folders/t/:uid/:token/:size", func(c *gin.Context) {
		if InvalidPreviewToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", folderIconSvg)
			return
		}

		start := time.Now()
		conf := get.Config()
		uid := clean.UID(c.Param("uid"))
		thumbName := thumb.Name(clean.Token(c.Param("size")))

		size, ok := thumb.Sizes[thumbName]

		if !ok {
			log.Errorf("%s: invalid size %s", folderCover, clean.Log(thumbName.String()))
			c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
			return
		}

		size = coverSize(size)
		thumbName = size.Name

		cache := get.CoverCache()
		cacheKey := CacheKey(folderCover, uid, string(thumbName))

		if cacheData, ok := cache.Get(cacheKey); ok {
			log.Tracef("api: cache hit for %s [%s]", cacheKey, time.Since(start))

			cached := cacheData.(ThumbCache)

			if !fs.FileExists(cached.FileName) {
				log.Errorf("%s: %s not found", folderCover, clean.Log(uid))
				c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
				return
			}

			AddCoverCacheHeader(c)
			c.File(cached.FileName)

			return
		}

		f, err := query.FolderCoverByUID(uid)

		if err != nil {
			log.Debugf("%s: %s contains no pictures, using generic cover", folderCover, clean.Log(uid))
			c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
			return
		}

		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("%s: found no original for %s", folderCover, clean.Log(f.FileName))
			c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			log.Warnf("%s: %s is missing", folderCover, clean.Log(f.FileName))
			logErr(folderCover, f.Update("FileMissing", true))
			return
		}

		var thumbnail string

		if conf.ThumbUncached() {
			thumbnail, err = thumb.FromFile(fileName, f.FileHash, conf.ThumbCachePath(), size.Width, size.Height, f.FileOrientation, size.Options...)
		} else {
			thumbnail, err = thumb.FromCache(fileName, f.FileHash, conf.ThumbCachePath(), size.Width, size.Height, size.Options...)
		}

		if err != nil {
			log.Errorf("%s: %s", folderCover, err)
			c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
			return
		} else if thumbnail == "" {
			log.Errorf("%s: %s has empty thumb name - you may have found a bug", folderCover, filepath.Base(fileName))
			c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
			return
		}

		cache.SetDefault(cacheKey, ThumbCache{FileName: thumbnail})
		log.Debugf("cached %s [%s]", cacheKey, time.Since(start))

		AddCoverCacheHeader(c)
		c.File(thumbnail)
	})
}
