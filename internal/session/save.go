package session

import (
	"encoding/json"
	"os"
	"path"
	"sync"
	"time"

	gc "github.com/patrickmn/go-cache"
	"github.com/photoprism/photoprism/internal/entity"
)

const cacheFileName = "sessions.json"

var fileMutex sync.RWMutex

// New returns a new session store with an optional cachePath.
func New(expiration time.Duration, cachePath string) *Session {
	s := &Session{}

	cleanupInterval := 15 * time.Minute

	if cachePath != "" {
		fileMutex.RLock()
		defer fileMutex.RUnlock()

		var savedItems map[string]Saved

		items := make(map[string]gc.Item)
		s.cacheFile = path.Join(cachePath, cacheFileName)

		if cached, err := os.ReadFile(s.cacheFile); err != nil {
			log.Debugf("session: %s", err)
		} else if err := json.Unmarshal(cached, &savedItems); err != nil {
			log.Errorf("session: %s", err)
		} else {
			for key, saved := range savedItems {
				user := entity.FindUserByUID(saved.User)

				if user == nil {
					continue
				}

				var tokens []string
				var shared []string

				for _, token := range saved.Tokens {
					links := entity.FindValidLinks(token, "")

					if len(links) > 0 {
						for _, link := range links {
							shared = append(shared, link.LinkUID)
						}

						tokens = append(tokens, token)
					}
				}

				data := Data{User: *user, Tokens: tokens, Shares: shared}
				items[key] = gc.Item{Expiration: saved.Expiration, Object: data}
			}

			s.cache = gc.NewFrom(expiration, cleanupInterval, items)
		}
	}

	if s.cache == nil {
		s.cache = gc.New(expiration, cleanupInterval)
	}

	return s
}

// Save stores all sessions in a JSON file.
func (s *Session) Save() error {
	if s.cacheFile == "" {
		return nil
	}

	fileMutex.Lock()
	defer fileMutex.Unlock()

	items := s.cache.Items()
	savedItems := make(map[string]Saved, len(items))

	for key, item := range items {
		saved := item.Object.(Data).Saved()
		saved.Expiration = item.Expiration
		savedItems[key] = saved
	}

	if serialized, err := json.MarshalIndent(savedItems, "", " "); err != nil {
		return err
	} else if err = os.WriteFile(s.cacheFile, serialized, 0600); err != nil {
		return err
	}

	return nil
}
