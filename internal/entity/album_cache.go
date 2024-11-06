package entity

import (
	"fmt"
	"time"

	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

var albumCache = gc.New(15*time.Minute, 15*time.Minute)

// FlushAlbumCache resets the album cache.
func FlushAlbumCache() {
	albumCache.Flush()
}

// CachedAlbumByUID returns an existing album or an error if not found.
func CachedAlbumByUID(uid string) (m Album, err error) {
	// Valid album UID?
	if uid == "" || rnd.InvalidUID(uid, AlbumUID) {
		return m, fmt.Errorf("invalid album uid %s", clean.LogQuote(uid))
	}

	// Cached?
	if cacheData, ok := albumCache.Get(uid); ok {
		log.Tracef("album: cache hit for %s", uid)
		return cacheData.(Album), nil
	}

	// Find in database.
	m = Album{}

	if r := Db().First(&m, "album_uid = ?", uid); r.RecordNotFound() {
		return m, fmt.Errorf("album not found")
	} else if r.Error != nil {
		return m, r.Error
	} else {
		albumCache.SetDefault(m.AlbumUID, m)
		return m, nil
	}
}
