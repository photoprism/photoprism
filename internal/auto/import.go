package auto

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
)

var autoImport = time.Time{}
var importMutex = sync.Mutex{}

// ResetImport resets the auto import trigger time.
func ResetImport() {
	importMutex.Lock()
	defer importMutex.Unlock()

	autoImport = time.Time{}
}

// ShouldImport sets the auto import trigger to the current time.
func ShouldImport() {
	importMutex.Lock()
	defer importMutex.Unlock()

	autoImport = time.Now()
}

// mustImport tests if auto import must be started.
func mustImport(delay time.Duration) bool {
	if delay.Seconds() <= 0 {
		return false
	}

	importMutex.Lock()
	defer importMutex.Unlock()

	return !autoImport.IsZero() && autoImport.Sub(time.Now()) < -1*delay && !mutex.MainWorker.Running()
}

// Import starts importing originals e.g. after WebDAV uploads.
func Import() error {
	if mutex.MainWorker.Running() {
		return nil
	}

	conf := service.Config()

	if conf.ReadOnly() || !conf.Settings().Features.Import {
		return nil
	}

	start := time.Now()

	path := filepath.Clean(conf.ImportPath())

	imp := service.Import()

	api.RemoveFromFolderCache(entity.RootImport)

	event.InfoMsg(i18n.MsgCopyingFilesFrom, clean.Log(filepath.Base(path)))

	var opt photoprism.ImportOptions

	if conf.Settings().Import.Move {
		opt = photoprism.ImportOptionsMove(path, conf.ImportDest())
	} else {
		opt = photoprism.ImportOptionsCopy(path, conf.ImportDest())
	}

	imported := imp.Start(opt)

	if len(imported) == 0 {
		return nil
	}

	moments := service.Moments()

	if err := moments.Start(); err != nil {
		log.Warnf("moments: %s", err)
	}

	elapsed := int(time.Since(start).Seconds())

	msg := i18n.Msg(i18n.MsgImportCompletedIn, elapsed)

	event.Success(msg)
	event.Publish("import.completed", event.Data{"path": path, "seconds": elapsed})
	event.Publish("index.completed", event.Data{"path": path, "seconds": elapsed})

	api.UpdateClientConfig()

	return nil
}
