package workers

import (
	"fmt"
	"path/filepath"
	"runtime/debug"
	"time"

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

	path := conf.OriginalsPath()

	ind := get.Index()

	convert := settings.Index.Convert && conf.SidecarWritable()
	indOpt := photoprism.NewIndexOptions(entity.RootPath, false, convert, true, false, true)
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

	elapsed := int(time.Since(start).Seconds())

	msg := i18n.Msg(i18n.MsgIndexingCompletedIn, elapsed)

	event.Success(msg)

	eventData := event.Data{
		"uid":     indOpt.UID,
		"action":  indOpt.Action,
		"path":    path,
		"seconds": elapsed,
	}

	event.Publish("index.completed", eventData)

	return nil
}
