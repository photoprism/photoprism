package api

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/thumb"
)

// CoverCacheTTL specifies the number of seconds to cache album covers.
var CoverCacheTTL thumb.MaxAge = 3600 // 1 hour

type ThumbCache struct {
	FileName  string
	ShareName string
}

type ByteCache struct {
	Data []byte
}

// CacheKey returns a cache key string based on namespace, uid and name.
func CacheKey(ns, uid, name string) string {
	return fmt.Sprintf("%s:%s:%s", ns, uid, name)
}

// RemoveFromFolderCache removes an item from the folder cache e.g. after indexing.
func RemoveFromFolderCache(rootName string) {
	cache := get.FolderCache()

	cacheKey := fmt.Sprintf("folder:%s:%t:%t", rootName, true, false)

	cache.Delete(cacheKey)

	if err := query.UpdateAlbumFolderCovers(); err != nil {
		log.Error(err)
	}

	log.Debugf("removed %s from cache", cacheKey)
}

// RemoveFromAlbumCoverCache removes covers by album UID e.g. after adding or removing photos.
func RemoveFromAlbumCoverCache(uid string) {
	cache := get.CoverCache()

	for thumbName := range thumb.Sizes {
		cacheKey := CacheKey(albumCover, uid, string(thumbName))

		cache.Delete(cacheKey)

		log.Debugf("removed %s from cache", cacheKey)
	}

	if err := query.UpdateAlbumCovers(); err != nil {
		log.Error(err)
	}
}

// FlushCoverCache clears the complete cover cache.
func FlushCoverCache() {
	get.CoverCache().Flush()

	if err := query.UpdateCovers(); err != nil {
		log.Error(err)
	}

	log.Debugf("albums: flushed cover cache")
}

// AddCacheHeader adds a cache control header to the response.
func AddCacheHeader(c *gin.Context, maxAge thumb.MaxAge, public bool) {
	if public {
		c.Header("Cache-Control", fmt.Sprintf("public, max-age=%s, no-transform", maxAge.String()))
	} else {
		c.Header("Cache-Control", fmt.Sprintf("private, max-age=%s, no-transform", maxAge.String()))
	}
}

// AddCoverCacheHeader adds cover image cache control headers to the response.
func AddCoverCacheHeader(c *gin.Context) {
	AddCacheHeader(c, CoverCacheTTL, thumb.CachePublic)
}

// AddThumbCacheHeader adds thumbnail cache control headers to the response.
func AddThumbCacheHeader(c *gin.Context) {
	if thumb.CachePublic {
		c.Header("Cache-Control", fmt.Sprintf("public, max-age=%s, no-transform, immutable", thumb.CacheTTL.String()))
	} else {
		c.Header("Cache-Control", fmt.Sprintf("private, max-age=%s, no-transform, immutable", thumb.CacheTTL.String()))
	}
}
