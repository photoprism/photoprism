package entity

import (
	"errors"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// SessionActive reports whether the session with the specified ID exists and has not expired. It reads a
// cached copy if there is one, but neither records activity nor adds the session to the cache. It returns
// ErrSessionIdInvalid for a malformed ID and any other error only if the session could not be read.
func SessionActive(id string) (bool, error) {
	if !rnd.IsSessionID(id) {
		return false, ErrSessionIdInvalid
	}

	if cacheData, ok := sessionCache.Get(id); ok && cacheData != nil {
		return !cacheData.(*Session).Expired(), nil
	}

	found := &Session{}

	// Read only the columns needed to check the expiry.
	if err := Db().Select("id", "sess_expires", "sess_timeout", "last_active").First(found, "id = ?", id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return !found.Expired(), nil
}
