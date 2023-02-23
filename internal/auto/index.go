package auto

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
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

	return !autoIndex.IsZero() && autoIndex.Sub(time.Now()) < -1*delay && !mutex.MainWorker.Running()
}

// Index starts indexing originals e.g. after WebDAV uploads.
func Index() error {
	if mutex.MainWorker.Running() {
		return nil
	}

	conf := get.Config()
	settings := conf.Settings()

	start := time.Now()

	path := conf.OriginalsPath()

	ind := get.Index()

	convert := settings.Index.Convert && conf.SidecarWritable()
	indOpt := photoprism.NewIndexOptions(entity.RootPath, false, convert, true, false, true)

	lastRun := ind.LastRun()
	found, updated := ind.Start(indOpt)

	if updated == 0 && lastRun.IsZero() {
		return nil
	}

	api.RemoveFromFolderCache(entity.RootOriginals)

	prg := get.Purge()

	prgOpt := photoprism.PurgeOptions{
		Path:   filepath.Clean(entity.RootPath),
		Ignore: found,
		Force:  true,
	}

	if files, photos, err := prg.Start(prgOpt); err != nil {
		return err
	} else if len(files) > 0 || len(photos) > 0 {
		event.InfoMsg(i18n.MsgRemovedFilesAndPhotos, len(files), len(photos))
	}

	event.Publish("index.updating", event.Data{
		"step": "moments",
	})

	moments := get.Moments()

	if err := moments.Start(); err != nil {
		log.Warnf("moments: %s", err)
	}

	elapsed := int(time.Since(start).Seconds())

	msg := i18n.Msg(i18n.MsgIndexingCompletedIn, elapsed)

	event.Success(msg)
	event.Publish("index.completed", event.Data{"path": path, "seconds": elapsed})

	api.UpdateClientConfig()

	return nil
}
