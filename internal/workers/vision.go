package workers

import (
	"errors"
	"fmt"
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
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/enum"
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
	if w.conf == nil {
		return nil
	}

	models := make([]string, 0, 4)

	if w.conf.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunOnSchedule) {
		models = append(models, vision.ModelTypeLabels)
	}

	if w.conf.VisionModelShouldRun(vision.ModelTypeNsfw, vision.RunOnSchedule) {
		models = append(models, vision.ModelTypeNsfw)
	}

	if w.conf.VisionModelShouldRun(vision.ModelTypeCaption, vision.RunOnSchedule) {
		models = append(models, vision.ModelTypeCaption)
	}

	if w.conf.VisionModelShouldRun(vision.ModelTypeFace, vision.RunOnSchedule) {
		models = append(models, vision.ModelTypeFace)
	}

	return models
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
	detectFaces := slices.Contains(models, vision.ModelTypeFace)

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
	// Remember if we saved new face markers so recognition can run after the loop.
	updateFaces := false

	start := time.Now()
	done := make(map[string]bool)
	offset := 0
	updated := 0
	processed := 0

	// Make sure count is within
	if count < 1 || count > search.MaxResults {
		count = search.MaxResults
	}

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
		frm.Caption = enum.False
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

		logName := photo.String()

		m, loadErr := query.PhotoByUID(photo.PhotoUID)

		if loadErr != nil {
			log.Errorf("vision: failed to load %s (%s)", logName, loadErr)
			continue
		}

		generateLabels := updateLabels && m.ShouldGenerateLabels(force)
		generateCaptions := updateCaptions && m.ShouldGenerateCaption(customSrc, force)
		detectNsfw := updateNsfw && (!photo.PhotoPrivate || force)

		if !generateLabels && !generateCaptions && !detectNsfw && !detectFaces {
			continue
		}

		processed++

		fileName := photoprism.FileName(photo.FileRoot, photo.FileName)
		file, fileErr := photoprism.NewMediaFile(fileName)

		if fileErr != nil {
			log.Errorf("vision: failed to open %s (%s)", logName, fileErr)
			continue
		}

		// Track whether this iteration produced metadata that needs persisting.
		changed := false

		if detectFaces {
			if primaryFile, err := m.PrimaryFile(); err != nil {
				log.Debugf("vision: photo %s has invalid primary file (%s)", logName, clean.Error(err))
			} else if primaryFile == nil {
				log.Debugf("vision: missing primary file for %s", logName)
			} else if markers := primaryFile.Markers(); markers == nil {
				log.Errorf("vision: failed loading markers for %s", logName)
			} else {
				expected := markers.DetectedFaceCount()
				faces, detectErr := photoprism.DetectFaces(file, expected)
				if detectErr != nil {
					log.Debugf("vision: %s in %s (detect faces)", detectErr, clean.Log(file.BaseName()))
				} else if saved, faceCount, applyErr := photoprism.ApplyDetectedFaces(primaryFile, faces); applyErr != nil {
					log.Warnf("vision: %s in %s (save faces)", clean.Error(applyErr), logName)
				} else if saved {
					m.PhotoFaces = faceCount
					updateFaces = true
					changed = true
				}
			}
		}

		// Generate labels.
		if generateLabels {
			if labels := file.GenerateLabels(customSrc); len(labels) > 0 {
				if w.conf.DetectNSFW() && !m.PhotoPrivate {
					if labels.IsNSFW(vision.Config.Thresholds.GetNSFW()) {
						m.PhotoPrivate = true
						log.Infof("vision: changed private flag of %s to %t (labels)", logName, m.PhotoPrivate)
					}
				}
				m.AddLabels(labels)
				changed = true
			}
		}

		// Detect NSFW content.
		if detectNsfw {
			if isNsfw := file.DetectNSFW(); m.PhotoPrivate != isNsfw {
				m.PhotoPrivate = isNsfw
				changed = true
				log.Infof("vision: changed private flag of %s to %t", logName, m.PhotoPrivate)
			}
		}

		// Generate a caption if none exists or the force flag is used,
		// and only if no caption was set or removed by a higher-priority source.
		if generateCaptions {
			if caption, captionErr := file.GenerateCaption(customSrc); captionErr != nil {
				log.Warnf("vision: %s in %s (generate caption)", clean.Error(captionErr), logName)
			} else if text := strings.TrimSpace(caption.Text); text != "" {
				m.SetCaption(text, caption.Source)
				if updateErr := m.UpdateCaptionLabels(); updateErr != nil {
					log.Warnf("vision: %s in %s (update caption labels)", clean.Error(updateErr), logName)
				}
				changed = true
				log.Infof("vision: changed caption of %s to %s", logName, clean.Log(m.PhotoCaption))
			}
		}

		if changed {
			if saveErr := m.SaveVision(); saveErr == nil {
				updated++
			}
		}

		if mutex.VisionWorker.Canceled() {
			return errors.New("vision: worker canceled")
		}
	}

	elapsed := time.Since(start)

	switch {
	case processed == 0:
		log.Infof("vision: no pictures required processing [%s]", elapsed)
	case updated == processed:
		log.Infof("vision: updated %s [%s]", english.Plural(updated, "picture", "pictures"), elapsed)
	case updated == 0:
		log.Infof("vision: processed %s (no metadata changes detected) [%s]", english.Plural(processed, "picture", "pictures"), elapsed)
	default:
		log.Infof("vision: updated %s out of %s [%s]", english.Plural(updated, "picture", "pictures"), english.Plural(processed, "picture", "pictures"), elapsed)
	}

	if updated > 0 {
		updateIndex = true
	}

	if updateFaces {
		// Perform face recognition after saving new face markers.
		log.Debugf("vision: running face recognition")
		if faces := photoprism.NewFaces(w.conf); faces.Disabled() {
			log.Debugf("vision: skipping face recognition")
		} else if facesErr := faces.Start(photoprism.FacesOptions{}); facesErr != nil {
			log.Warn(facesErr)
		}
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
