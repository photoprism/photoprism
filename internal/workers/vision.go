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
)

// Vision represents a computer vision worker.
type Vision struct {
	conf *config.Config
}

// NewVision returns a new Vision worker.
func NewVision(conf *config.Config) *Vision {
	return &Vision{conf: conf}
}

// originalsPath returns the original media files path as string.
func (w *Vision) originalsPath() string {
	return w.conf.OriginalsPath()
}

// Start runs the specified model types for the photos that match the search query.
func (w *Vision) Start(q string, models []string, force bool) (err error) {
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

	updateLabels := slices.Contains(models, vision.ModelTypeLabels)
	updateNsfw := slices.Contains(models, vision.ModelTypeNsfw)
	updateCaptions := slices.Contains(models, vision.ModelTypeCaption)

	// Refresh index metadata.
	if n := len(models); n == 0 {
		log.Warnf("vision: no models were specified")
		return nil
	} else if n == 1 {
		log.Infof("vision: running %s model", models[0])
	} else {
		log.Infof("vision: running %s models", strings.Join(models, " and "))
	}

	// Check time when worker was last executed.
	updateIndex := false

	start := time.Now()
	done := make(map[string]bool)
	limit := 1000
	offset := 0
	updated := 0

	ind := get.Index()

	for {
		frm := form.SearchPhotos{
			Query:   strings.TrimSpace(q),
			Primary: true,
			Merged:  false,
			Count:   limit,
			Offset:  offset,
			Order:   sortby.Oldest,
		}

		photos, _, queryErr := search.Photos(frm)

		if queryErr != nil {
			return queryErr
		}

		if len(photos) == 0 {
			break
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
			fileName := photoprism.FileName(photo.FileRoot, photo.FileName)
			file, fileErr := photoprism.NewMediaFile(fileName)

			if fileErr != nil {
				log.Errorf("vision: failed to open %s (%s)", photoName, fileErr)
				continue
			}

			m, loadErr := query.PhotoByUID(photo.PhotoUID)

			if loadErr != nil {
				log.Errorf("vision: failed to load %s (%s)", photoName, loadErr)
				continue
			}

			changed := false

			if updateLabels && (len(m.Labels) == 0 || force) {
				if labels := ind.Labels(file); len(labels) > 0 {
					m.AddLabels(labels)
					changed = true
				}
			}

			if updateNsfw && (!photo.PhotoPrivate || force) {
				if isNsfw := ind.IsNsfw(file); photo.PhotoPrivate != isNsfw {
					photo.PhotoPrivate = isNsfw
					changed = true
					log.Infof("vision: changed private flag of %s to %t", photoName, photo.PhotoPrivate)
				}
			}

			if updateCaptions && (m.PhotoCaption == "" || force) {
				if caption, captionErr := ind.Caption(file); captionErr != nil {
					log.Warnf("vision: %s in %s (generate caption)", clean.Error(captionErr), photoName)
				} else if caption.Text != "" {
					if caption.Source == "" {
						caption.Source = entity.SrcImage
					}

					if (entity.SrcPriority[caption.Source] > entity.SrcPriority[m.CaptionSrc]) || !m.HasCaption() {
						m.SetCaption(caption.Text, caption.Source)
						changed = true
						log.Infof("vision: changed caption of %s to %s", photoName, clean.Log(m.PhotoCaption))
					}
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
		}

		if mutex.VisionWorker.Canceled() {
			return errors.New("vision: worker canceled")
		}

		offset += limit
	}

	if updated > 0 {
		log.Infof("vision: updated %s [%s]", english.Plural(updated, "photo", "photos"), time.Since(start))
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
