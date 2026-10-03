package workers

import (
	"fmt"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// Index represents a background indexing worker.
type Index struct {
	conf *config.Config
}

// NewIndex returns a new Index worker.
func NewIndex(conf *config.Config) *Index {
	return &Index{conf: conf}
}

// StartScheduled starts a scheduled run of the indexing worker based on the current configuration.
func (w *Index) StartScheduled() {
	if err := w.Start(); err != nil {
		log.Errorf("scheduler: %s (index)", err)
	}
}

// Start runs the indexing worker once.
func (w *Index) Start() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("index: %s (worker panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if mutex.IndexWorker.Running() || mutex.BackupWorker.Running() {
		return nil
	}

	conf := w.conf
	settings := conf.Settings()

	start := time.Now()

	ind := get.Index()

	convert := settings.Index.Convert && conf.SidecarWritable()
	indOpt := photoprism.NewIndexOptions(entity.RootPath, false, convert, true, false, true, conf)
	indOpt.Action = photoprism.ActionAutoIndex

	lastRun, lastFound := ind.LastRun()
	found, indexed := ind.Start(indOpt)

	if !lastRun.IsZero() && indexed == 0 && len(found) == lastFound {
		return nil
	}

	prg := get.Purge()

	prgOpt := photoprism.PurgeOptions{
		Path:   filepath.Clean(entity.RootPath),
		Ignore: found,
		Force:  true,
	}

	if files, photos, updated, err := prg.Start(prgOpt); err != nil {
		return err
	} else if updated > 0 {
		event.InfoMsg(i18n.MsgRemovedFilesAndPhotos, len(files), len(photos))
	}

	event.Publish("index.updating", event.Data{
		"uid":    indOpt.UID,
		"action": indOpt.Action,
		"step":   "moments",
	})

	moments := get.Moments()

	if err := moments.Start(); err != nil {
		log.Warnf("moments: %s", err)
	}

	elapsed := time.Since(start)
	seconds := int(elapsed.Seconds())

	log.Infof("library: indexed %s in %s", english.Plural(len(found), "file", "files"), elapsed)

	event.PublishSuccessMsg(i18n.MsgIndexingCompletedIn, seconds)

	event.PublishCompleted([]string{"index.completed"}, indOpt.UID, indOpt.Action, seconds)

	return nil
}
