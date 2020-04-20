/*
This package encapsulates session storage.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki
*/
package session

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"time"

	gc "github.com/patrickmn/go-cache"
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

// Session represents a session store.
type Session struct {
	cacheFile string
	cache     *gc.Cache
}

// New returns a new session store with an optional cachePath.
func New(expiration time.Duration, cachePath string) *Session {
	s := &Session{}

	cleanupInterval := 15 * time.Minute

	if cachePath != "" {
		var items map[string]gc.Item

		s.cacheFile = path.Join(cachePath, "sessions.json")

		if cached, err := ioutil.ReadFile(s.cacheFile); err != nil {
			log.Infof("session: %s", err)
		} else if err := json.Unmarshal(cached, &items); err != nil {
			log.Errorf("session: %s", err)
		} else {
			s.cache = gc.NewFrom(expiration, cleanupInterval, items)
		}
	}

	if s.cache == nil {
		s.cache = gc.New(expiration, cleanupInterval)
	}

	return s
}
