package api

import (
	"fmt"
	"os"
	"path"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config/ttl"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

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
	if !rnd.IsAlnum(uid) {
		return
	}

	cache := get.CoverCache()

	// Flush album cover cache.
	for thumbName := range thumb.Sizes {
		cacheKey := CacheKey(albumCover, uid, string(thumbName))

		cache.Delete(cacheKey)

		log.Debugf("removed %s from cache", cacheKey)
	}

	// Delete share preview, if exists.
	if sharePreview := path.Join(get.Config().ThumbCachePath(), "share", uid+fs.ExtJPEG); fs.FileExists(sharePreview) {
		_ = os.Remove(sharePreview)
	}

	// Update album cover images.
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
func AddCacheHeader(c *gin.Context, duration ttl.Duration, public bool) {
	header.SetCacheControl(c, duration.Int(), public)
}

// AddCoverCacheHeader adds cover image cache control headers to the response.
func AddCoverCacheHeader(c *gin.Context) {
	AddCacheHeader(c, ttl.CacheCover, thumb.CachePublic)
}

// AddImmutableCacheHeader adds cache control headers to the response for immutable content like thumbnails.
func AddImmutableCacheHeader(c *gin.Context) {
	header.SetCacheControlImmutable(c, ttl.CacheDefault.Int(), thumb.CachePublic)
}

// AddVideoCacheHeader adds video cache control headers to the response.
func AddVideoCacheHeader(c *gin.Context, cdn bool) {
	header.SetCacheControlImmutable(c, ttl.CacheVideo.Int(), cdn || thumb.CachePublic)
}
