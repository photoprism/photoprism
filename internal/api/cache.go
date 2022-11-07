package api

import (
	"fmt"
	"strconv"

	"github.com/photoprism/photoprism/internal/query"

	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/thumb"
)

// MaxAge represents a cache TTL in seconds.
type MaxAge int

// String returns the cache TTL in seconds as string.
func (a MaxAge) String() string {
	return strconv.Itoa(int(a))
}

// Default cache TTL times in seconds.
var (
	CoverCacheTTL MaxAge = 3600           // 1 hour
	ThumbCacheTTL MaxAge = 3600 * 24 * 90 // ~ 3 months
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
