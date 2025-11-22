package entity

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dustin/go-humanize/english"
	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/pkg/clean"
)

// Keyword and PhotoKeyword cache expiration times and cleanup interval.
const (
	keywordCacheDefaultExpiration = 15 * time.Minute
	keywordCacheErrorExpiration   = 5 * time.Minute
	keywordCacheCleanupInterval   = 10 * time.Minute
	photoKeywordCacheExpiration   = 24 * time.Hour
)

// Cache Keyword and PhotoKeyword entities for faster indexing.
var (
	keywordCache           = gc.New(keywordCacheDefaultExpiration, keywordCacheCleanupInterval)
	photoKeywordCache      = gc.New(photoKeywordCacheExpiration, keywordCacheCleanupInterval)
	photoKeywordCacheMutex = sync.Mutex{}
)

// photoKeywordCacheKey returns a string key for the photoKeywordCache.
func photoKeywordCacheKey(photoId, keywordId uint) string {
	return fmt.Sprintf("%d-%d", photoId, keywordId)
}

// FlushKeywordCache removes all cached Keyword entities from the cache.
func FlushKeywordCache() {
	keywordCache.Flush()
}

// FlushCachedKeyword deletes a cached Keyword entity from the cache.
func FlushCachedKeyword(m *Keyword) {
	if m == nil {
		return
	} else if m.HasID() {
		keywordCache.Delete(m.Keyword)
	}
}

// FlushPhotoKeywordCache removes all cached PhotoKeyword entities from the cache.
func FlushPhotoKeywordCache() {
	photoKeywordCacheMutex.Lock()
	defer photoKeywordCacheMutex.Unlock()

	start := time.Now()

	photoKeywordCache.Flush()

	log.Debugf("index: flushed photo keywords cache [%s]", time.Since(start))
}

// FlushCachedPhotoKeyword deletes a cached PhotoKeyword entity from the cache.
func FlushCachedPhotoKeyword(m *PhotoKeyword) {
	if m == nil {
		return
	} else if m.HasID() {
		photoKeywordCache.Delete(photoKeywordCacheKey(m.PhotoID, m.KeywordID))
	}
}

// CachePhotoKeywords preloads the photo-keyword cache from the database to speed up lookups.
func CachePhotoKeywords() (err error) {
	photoKeywordCacheMutex.Lock()
	defer photoKeywordCacheMutex.Unlock()

	start := time.Now()

	var photoKeywords []PhotoKeyword

	// Find photo keyword assignments.
	if err = UnscopedDb().
		Raw("SELECT * FROM photos_keywords").
		Scan(&photoKeywords).Error; err != nil {
		return err
	}

	// Cache existing keyword assignments.
	for _, m := range photoKeywords {
		photoKeywordCache.SetDefault(m.CacheKey(), m)
	}

	log.Debugf("index: cached %s [%s]", english.Plural(len(photoKeywords), "photo keyword", "photo keywords"), time.Since(start))

	return nil
}

// FindKeyword resolves a keyword by its normalized name, optionally consulting
// the in-memory cache before hitting the database.
func FindKeyword(keyword string, cached bool) (*Keyword, error) {
	if keyword == "" {
		return &Keyword{}, errors.New("missing keyword name")
	}

	// Return cached keyword, if found.
	if cached {
		if cacheData, ok := keywordCache.Get(keyword); ok {
			log.Tracef("keyword: cache hit for %s", keyword)

			// Get cached data.
			if result := cacheData.(*Keyword); result.HasID() {
				// Return cached entity.
				return result, nil
			} else {
				// Return cached "not found" error.
				return &Keyword{}, fmt.Errorf("keyword not found")
			}
		}
	}

	// Fetch and cache keyword.
	result := &Keyword{}

	if find := Db().First(result, "keyword = ?", keyword); find.RecordNotFound() {
		keywordCache.Set(keyword, result, keywordCacheErrorExpiration)
		return result, fmt.Errorf("keyword not found")
	} else if find.Error != nil {
		keywordCache.Set(keyword, result, keywordCacheErrorExpiration)
		return result, find.Error
	} else {
		keywordCache.SetDefault(result.Keyword, result)
	}

	return result, nil
}

// FindPhotoKeyword loads the photo-keyword join row for the given IDs, using the cache when enabled.
func FindPhotoKeyword(photoId, keywordId uint, cached bool) (*PhotoKeyword, error) {
	if photoId == 0 {
		return &PhotoKeyword{}, errors.New("invalid photo id")
	} else if keywordId == 0 {
		return &PhotoKeyword{}, errors.New("invalid keyword id")
	}

	cacheKey := photoKeywordCacheKey(photoId, keywordId)

	if cacheKey == "" {
		return &PhotoKeyword{}, fmt.Errorf("invalid cache key %s", clean.LogQuote(cacheKey))
	}

	// Return cached keyword, if found.
	if cached {
		if cacheData, ok := photoKeywordCache.Get(cacheKey); ok {
			log.Tracef("photo-keyword: cache hit for %s", cacheKey)

			// Get cached data.
			if result := cacheData.(PhotoKeyword); result.HasID() {
				// Return cached entity.
				return &result, nil
			} else {
				// Return cached "not found" error.
				return &PhotoKeyword{}, fmt.Errorf("photo-keyword not found")
			}
		}
	}

	// Fetch and cache photo-keyword.
	result := &PhotoKeyword{}

	if find := Db().First(result, "photo_id = ? AND keyword_id = ?", photoId, keywordId); find.RecordNotFound() {
		photoKeywordCache.Set(cacheKey, *result, keywordCacheErrorExpiration)
		return result, fmt.Errorf("photo-keyword not found")
	} else if find.Error != nil {
		photoKeywordCache.Set(cacheKey, *result, keywordCacheErrorExpiration)
		return result, find.Error
	} else {
		photoKeywordCache.SetDefault(cacheKey, *result)
	}

	return result, nil
}
