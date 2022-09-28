package entity

import (
	"fmt"
	"time"

	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

var sessionCacheExpiration = 15 * time.Minute
var sessionCache = gc.New(sessionCacheExpiration, 5*time.Minute)

// FlushSessionCache resets the session cache.
func FlushSessionCache() {
	sessionCache.Flush()
}

// DeleteFromSessionCache deletes a cached session.
func DeleteFromSessionCache(id string) {
	if id == "" {
		return
	}

	sessionCache.Delete(id)
}

// FindSession returns an existing session or nil if not found.
func FindSession(id string) (s Session, err error) {
	s = Session{}

	// Valid id?
	if !rnd.IsSessionID(id) {
		return s, fmt.Errorf("id %s is invalid", clean.LogQuote(id))
	}

	// Find cached session.
	if cacheData, ok := sessionCache.Get(id); ok {
		return cacheData.(Session), nil
	}

	// Search database and return session if found.
	if r := Db().First(&s, "id = ?", id); r.RecordNotFound() {
		err = fmt.Errorf("not found")
	} else if r.Error != nil {
		err = r.Error
	} else if !rnd.IsSessionID(s.ID) {
		err = fmt.Errorf("has invalid id %s", clean.LogQuote(s.ID))
	} else {
		sessionCache.SetDefault(s.ID, s)
	}

	return s, err
}
