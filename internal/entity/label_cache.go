package entity

import (
	"fmt"
	"time"

	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// labelCache expiration times and cleanup interval.
const (
	labelDefaultExpiration = 15 * time.Minute
	labelErrorExpiration   = 5 * time.Minute
	labelCleanupInterval   = 5 * time.Minute
)

// labelCache stores Label entities for faster indexing.
var labelCache = gc.New(labelDefaultExpiration, labelCleanupInterval)

// FindLabel find the matching label based on the name provided or an error if not found.
func FindLabel(name string, cached bool) (*Label, error) {
	labelSlug := txt.Slug(name)

	if labelSlug == "" {
		return &Label{}, fmt.Errorf("invalid label slug %s", clean.LogQuote(labelSlug))
	}

	// Return cached label, if found.
	if cached {
		if cacheData, ok := labelCache.Get(labelSlug); ok {
			log.Tracef("label: cache hit for %s", labelSlug)

			if result := cacheData.(*Label); !result.HasID() {
				return &Label{}, fmt.Errorf("label not found")
			} else {
				return result, nil
			}
		}
	}

	// Fetch and cache label from database.
	result := &Label{}

	if find := Db().First(result, "(label_slug <> '' AND label_slug = ? OR custom_slug <> '' AND custom_slug = ?)", labelSlug, labelSlug); find.RecordNotFound() {
		labelCache.Set(labelSlug, result, labelErrorExpiration)
		return result, fmt.Errorf("label not found")
	} else if find.Error != nil {
		labelCache.Set(labelSlug, result, labelErrorExpiration)
		return result, find.Error
	} else {
		labelCache.SetDefault(result.LabelSlug, result)
	}

	return result, nil
}

// FlushLabelCache removes all cached Label entities from the cache.
func FlushLabelCache() {
	labelCache.Flush()
}
