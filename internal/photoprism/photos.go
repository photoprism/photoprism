package photoprism

import (
	"fmt"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/photoprism/photoprism/internal/entity/query"
)

// Photos represents photo id lookup table, sorted by date and S2 cell id.
type Photos struct {
	count  int
	photos query.PhotoMap
	mutex  sync.RWMutex
}

// NewPhotos returns a new Photos instance.
func NewPhotos() *Photos {
	m := &Photos{
		photos: make(query.PhotoMap),
	}

	return m
}

// Init fetches the list from the database once.
func (m *Photos) Init() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.photos) > 0 {
		m.count = len(m.photos)
		return nil
	}

	photos, err := query.IndexedPhotos()

	if err != nil {
		return fmt.Errorf("%s (find indexed photos)", err.Error())
	} else {
		m.photos = photos
		m.count = len(photos)
		return nil
	}
}

// Remove a photo from the lookup table.
func (m *Photos) Remove(takenAt time.Time, cellId string) {
	key := entity.MapKey(takenAt, cellId)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.photos, key)
}

// Find returns the photo ID for a time and cell id.
func (m *Photos) Find(takenAt time.Time, cellId string) uint {
	key := entity.MapKey(takenAt, cellId)

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.photos[key]
}
