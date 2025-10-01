package workers

import (
	"errors"
	"fmt"
	"path"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Vision orchestrates background computer-vision tasks (labels, captions,
// NSFW detection). It wraps configuration lookups and scheduling helpers.
type Vision struct {
	conf *config.Config
}

// NewVision constructs a Vision worker bound to the provided configuration.
func NewVision(conf *config.Config) *Vision {
	return &Vision{conf: conf}
}

// StartScheduled executes the worker in scheduled mode, selecting models that
// are allowed to run in the RunOnSchedule context.
func (w *Vision) StartScheduled() {
	models := w.scheduledModels()

	if len(models) == 0 {
		return
	}

	if err := w.Start(
		w.conf.VisionFilter(),
		0,
		models,
		entity.SrcAuto,
		false,
		vision.RunOnSchedule,
	); err != nil {
		log.Errorf("scheduler: %s (vision)", err)
	}
}

// scheduledModels returns the model types that should run for scheduled jobs.
func (w *Vision) scheduledModels() []string {
	models := make([]string, 0, 3)

	if w.conf.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunOnSchedule) {
		models = append(models, vision.ModelTypeLabels)
	}

	if w.conf.VisionModelShouldRun(vision.ModelTypeNsfw, vision.RunOnSchedule) {
		models = append(models, vision.ModelTypeNsfw)
	}

	if w.conf.VisionModelShouldRun(vision.ModelTypeCaption, vision.RunOnSchedule) {
		models = append(models, vision.ModelTypeCaption)
	}

	return models
}

// originalsPath returns the path that holds original media files.
func (w *Vision) originalsPath() string {
	return w.conf.OriginalsPath()
}

// Start runs the requested vision models against photos matching the search
// filter. `customSrc` allows the caller to override the metadata source string,
// `force` regenerates metadata regardless of existing values, and `runType`
// describes the scheduling context (manual, scheduled, etc.). A global worker
// mutex prevents multiple vision jobs from running concurrently.
func (w *Vision) Start(filter string, count int, models []string, customSrc string, force bool, runType vision.RunType) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("vision: %s (worker panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if err = mutex.VisionWorker.Start(); err != nil {
		return err
	}

	defer mutex.VisionWorker.Stop()

	models = vision.FilterModels(models, runType, func(mt vision.ModelType, when vision.RunType) bool {
		return w.conf.VisionModelShouldRun(mt, when)
	})

	updateLabels := slices.Contains(models, vision.ModelTypeLabels)
	updateNsfw := slices.Contains(models, vision.ModelTypeNsfw)
	updateCaptions := slices.Contains(models, vision.ModelTypeCaption)

	// Refresh index metadata.
	if n := len(models); n == 0 {
		log.Warnf("vision: no models were specified")
		return nil
	} else {
		log.Infof("vision: running %s models", txt.JoinAnd(models))
	}

	customSrc = clean.ShortTypeLower(customSrc)

	// Check time when worker was last executed.
	updateIndex := false

	start := time.Now()
	done := make(map[string]bool)
	offset := 0
	updated := 0

	// Make sure count is within
	if count < 1 || count > search.MaxResults {
		count = search.MaxResults
	}

	ind := get.Index()

	frm := form.SearchPhotos{
		Query:   filter,
		Primary: true,
		Merged:  false,
		Count:   count,
		Offset:  offset,
		Order:   sortby.Added,
	}

	// Find photos without captions when only
	// captions are updated without force flag.
	if !updateLabels && !updateNsfw && !force {
		frm.Caption = txt.False
	}

	photos, _, queryErr := search.Photos(frm)

	if queryErr != nil {
		return queryErr
	}

	if n := len(photos); n == 0 {
		log.Info("vision: no pictures to process")
		return nil
	} else {
		log.Infof("vision: processing %s", english.Plural(n, "picture", "pictures"))
	}

	for _, photo := range photos {
		if mutex.VisionWorker.Canceled() {
			return errors.New("vision: worker canceled")
		}

		if done[photo.PhotoUID] {
			continue
		}

		done[photo.PhotoUID] = true

		photoName := path.Join(photo.PhotoPath, photo.PhotoName)

		m, loadErr := query.PhotoByUID(photo.PhotoUID)

		if loadErr != nil {
			log.Errorf("vision: failed to load %s (%s)", photoName, loadErr)
			continue
		}

		generateLabels := updateLabels && m.ShouldGenerateLabels(force)
		generateCaptions := updateCaptions && m.ShouldGenerateCaption(customSrc, force)
		generateNsfw := updateNsfw && (!photo.PhotoPrivate || force)

		if !(generateLabels || generateCaptions || generateNsfw) {
			continue
		}

		fileName := photoprism.FileName(photo.FileRoot, photo.FileName)
		file, fileErr := photoprism.NewMediaFile(fileName)

		if fileErr != nil {
			log.Errorf("vision: failed to open %s (%s)", photoName, fileErr)
			continue
		}

		changed := false

		// Generate labels.
		if generateLabels {
			if labels := ind.Labels(file, customSrc); len(labels) > 0 {
				m.AddLabels(labels)
				changed = true
			}
		}

		// Detect NSFW content.
		if generateNsfw {
			if isNsfw := ind.IsNsfw(file); m.PhotoPrivate != isNsfw {
				m.PhotoPrivate = isNsfw
				changed = true
				log.Infof("vision: changed private flag of %s to %t", photoName, m.PhotoPrivate)
			}
		}

		// Generate a caption if none exists or the force flag is used,
		// and only if no caption was set or removed by a higher-priority source.
		if generateCaptions {
			if caption, captionErr := ind.Caption(file, customSrc); captionErr != nil {
				log.Warnf("vision: %s in %s (generate caption)", clean.Error(captionErr), photoName)
			} else if text := strings.TrimSpace(caption.Text); text != "" {
				m.SetCaption(text, caption.Source)
				if updateErr := m.UpdateCaptionLabels(); updateErr != nil {
					log.Warnf("vision: %s in %s (update caption labels)", clean.Error(updateErr), photoName)
				}
				changed = true
				log.Infof("vision: changed caption of %s to %s", photoName, clean.Log(m.PhotoCaption))
			}
		}

		if changed {
			if saveErr := m.GenerateAndSaveTitle(); saveErr != nil {
				log.Infof("vision: failed to updated %s (%s)", photoName, clean.Error(saveErr))
			} else {
				updated++
				log.Debugf("vision: updated %s", photoName)
			}
		}

		if mutex.VisionWorker.Canceled() {
			return errors.New("vision: worker canceled")
		}
	}

	log.Infof("vision: updated %s [%s]", english.Plural(updated, "picture", "pictures"), time.Since(start))

	if updated > 0 {
		updateIndex = true
	}

	// Only update index if photo metadata has changed or the force flag was used.
	if updateIndex {
		// Run moments worker.
		if moments := photoprism.NewMoments(w.conf); moments == nil {
			log.Errorf("vision: failed to update moments")
		} else if err = moments.Start(); err != nil {
			log.Warnf("moments: %s in optimization worker", err)
		}

		// Update precalculated photo and file counts.
		if err = entity.UpdateCounts(); err != nil {
			log.Warnf("vision: %s in optimization worker", err)
		}

		// Update album, subject, and label cover thumbs.
		if err = query.UpdateCovers(); err != nil {
			log.Warnf("vision: %s in optimization worker", err)
		}
	}

	return nil
}
