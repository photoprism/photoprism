package auto

import (
	"errors"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/workers"
)

var autoIndex = time.Time{}
var indexMutex = sync.Mutex{}

// ResetIndex resets the auto index trigger time.
func ResetIndex() {
	indexMutex.Lock()
	defer indexMutex.Unlock()

	autoIndex = time.Time{}
}

// ShouldIndex sets the auto index trigger to the current time.
func ShouldIndex() {
	indexMutex.Lock()
	defer indexMutex.Unlock()

	autoIndex = time.Now()
}

// mustIndex tests if the index must be updated.
func mustIndex(delay time.Duration) bool {
	if delay.Seconds() <= 0 {
		return false
	}

	indexMutex.Lock()
	defer indexMutex.Unlock()

	return !autoIndex.IsZero() && time.Until(autoIndex) < -1*delay && !mutex.IndexWorker.Running()
}

// Index starts indexing originals e.g. after WebDAV uploads. A run that waits for a face migration
// sets the trigger again, so the uploads are indexed once the migration has completed.
func Index() (err error) {
	if mutex.IndexWorker.Running() {
		return nil
	}

	api.RemoveFromFolderCache(entity.RootOriginals)

	if err = workers.NewIndex(get.Config()).Start(); errors.Is(err, photoprism.ErrFaceMigration) {
		log.Debugf("auto-index: %s, trying again later", err)
		ShouldIndex()
		return nil
	}

	api.UpdateClientConfig()

	return err
}
