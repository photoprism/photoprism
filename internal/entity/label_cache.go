package entity

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dustin/go-humanize/english"
	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Label and PhotoLabel cache expiration times and cleanup interval.
const (
	labelCacheDefaultExpiration = 15 * time.Minute
	labelCacheErrorExpiration   = 5 * time.Minute
	labelCacheCleanupInterval   = 10 * time.Minute
	photoLabelCacheExpiration   = 24 * time.Hour
)

// Cache Label and PhotoLabel entities for faster indexing.
var (
	UsePhotoLabelsCache  = true
	labelCache           = gc.New(labelCacheDefaultExpiration, labelCacheCleanupInterval)
	photoLabelCache      = gc.New(photoLabelCacheExpiration, labelCacheCleanupInterval)
	photoLabelCacheMutex = sync.Mutex{}
)

// photoLabelCacheKey returns a string key for the photoLabelCache.
func photoLabelCacheKey(photoId, labelId uint) string {
	return fmt.Sprintf("%d-%d", photoId, labelId)
}

// FlushLabelCache removes all cached Label entities from the cache.
func FlushLabelCache() {
	labelCache.Flush()
}

// FlushPhotoLabelCache removes all cached PhotoLabel entities from the cache.
func FlushPhotoLabelCache() {
	if !UsePhotoLabelsCache {
		return
	}

	photoLabelCacheMutex.Lock()
	defer photoLabelCacheMutex.Unlock()

	start := time.Now()

	photoLabelCache.Flush()

	log.Debugf("index: flushed photo labels cache [%s]", time.Since(start))
}

// FlushCachedPhotoLabel deletes a cached PhotoLabel entity from the cache.
func FlushCachedPhotoLabel(m *PhotoLabel) {
	if m == nil || !UsePhotoLabelsCache {
		return
	} else if m.HasID() {
		photoLabelCache.Delete(photoLabelCacheKey(m.PhotoID, m.LabelID))
	}
}

// CachePhotoLabels preloads the photo-label cache from the database to speed up lookups.
func CachePhotoLabels() (err error) {
	if !UsePhotoLabelsCache {
		return nil
	}

	photoLabelCacheMutex.Lock()
	defer photoLabelCacheMutex.Unlock()

	start := time.Now()

	var photoLabels []PhotoLabel

	// Find photo label assignments.
	if err = UnscopedDb().
		Raw("SELECT * FROM photos_labels").
		Scan(&photoLabels).Error; err != nil {
		return err
	}

	// Cache existing label assignments.
	for _, m := range photoLabels {
		photoLabelCache.SetDefault(m.CacheKey(), m)
	}

	log.Debugf("index: cached %s [%s]", english.Plural(len(photoLabels), "photo label", "photo labels"), time.Since(start))

	return nil
}

// FindLabel resolves a label by name/slug, optionally consulting the in-memory cache before querying the database.
func FindLabel(name string, cached bool) (*Label, error) {
	if name == "" {
		return &Label{}, errors.New("missing label name")
	}

	// Use the label slug as natural key cache.
	cacheKey := txt.Slug(name)

	if cacheKey == "" {
		return &Label{}, fmt.Errorf("invalid label slug %s", clean.LogQuote(cacheKey))
	}

	// Return cached label, if found.
	if cached {
		if cacheData, ok := labelCache.Get(cacheKey); ok {
			log.Tracef("label: cache hit for %s", cacheKey)

			// Get cached data.
			if result := cacheData.(*Label); result.HasID() {
				// Return cached entity.
				return result, nil
			} else {
				// Return cached "not found" error.
				return &Label{}, fmt.Errorf("label not found")
			}
		}
	}

	// Fetch and cache label.
	result := &Label{}

	if find := Db().First(result, "(label_slug <> '' AND label_slug = ? OR custom_slug <> '' AND custom_slug = ?)", cacheKey, cacheKey); find.RecordNotFound() {
		labelCache.Set(cacheKey, result, labelCacheErrorExpiration)
		return result, fmt.Errorf("label not found")
	} else if find.Error != nil {
		labelCache.Set(cacheKey, result, labelCacheErrorExpiration)
		return result, find.Error
	} else {
		labelCache.SetDefault(result.LabelSlug, result)
	}

	return result, nil
}

// FindPhotoLabel loads the photo-label join row for the given IDs, using the cache when enabled.
func FindPhotoLabel(photoId, labelId uint, cached bool) (*PhotoLabel, error) {
	if photoId == 0 {
		return &PhotoLabel{}, errors.New("invalid photo id")
	} else if labelId == 0 {
		return &PhotoLabel{}, errors.New("invalid label id")
	}

	cacheKey := photoLabelCacheKey(photoId, labelId)

	if cacheKey == "" {
		return &PhotoLabel{}, fmt.Errorf("invalid cache key %s", clean.LogQuote(cacheKey))
	}

	// Return cached label, if found.
	if cached && UsePhotoLabelsCache {
		if cacheData, ok := photoLabelCache.Get(cacheKey); ok {
			log.Tracef("photo-label: cache hit for %s", cacheKey)

			// Get cached data.
			if result := cacheData.(PhotoLabel); result.HasID() {
				// Return cached entity.
				return &result, nil
			} else {
				// Return cached "not found" error.
				return &PhotoLabel{}, fmt.Errorf("photo-label not found")
			}
		}
	}

	// Fetch and cache photo-label.
	result := &PhotoLabel{}

	if find := Db().First(result, "photo_id = ? AND label_id = ?", photoId, labelId); find.RecordNotFound() {
		if cached && UsePhotoLabelsCache {
			photoLabelCache.Set(cacheKey, *result, labelCacheErrorExpiration)
		}
		return result, fmt.Errorf("photo-label not found")
	} else if find.Error != nil {
		if cached && UsePhotoLabelsCache {
			photoLabelCache.Set(cacheKey, *result, labelCacheErrorExpiration)
		}
		return result, find.Error
	} else if cached && UsePhotoLabelsCache {
		photoLabelCache.SetDefault(cacheKey, *result)
	}

	return result, nil
}
