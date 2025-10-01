package workers

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
)

// Meta represents a background index and metadata optimization worker.
type Meta struct {
	conf *config.Config
}

// NewMeta returns a new Meta worker.
func NewMeta(conf *config.Config) *Meta {
	return &Meta{conf: conf}
}

// originalsPath returns the original media files path as string.
func (w *Meta) originalsPath() string {
	return w.conf.OriginalsPath()
}

// Start metadata optimization routine.
func (w *Meta) Start(delay, interval time.Duration, force bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("index: %s (worker panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if err = mutex.MetaWorker.Start(); err != nil {
		return err
	}

	defer mutex.MetaWorker.Stop()

	// Check time when worker was last executed.
	updateIndex := force || mutex.MetaWorker.LastRun().Before(time.Now().Add(-1*entity.IndexUpdateInterval))

	// Run faces worker if needed.
	if updateIndex || entity.UpdateFaces.Load() {
		log.Debugf("index: running face recognition")
		if faces := photoprism.NewFaces(w.conf); faces.Disabled() {
			log.Debugf("index: skipping face recognition")
		} else if facesErr := faces.Start(photoprism.FacesOptions{}); facesErr != nil {
			log.Warn(facesErr)
		}
	}

	// Refresh index metadata.
	log.Debugf("index: updating metadata")

	start := time.Now()
	settings := w.conf.Settings()
	done := make(map[string]bool)
	limit := 1000
	offset := 0
	optimized := 0

	ind := get.Index()

	labelsModelShouldRun := w.conf.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunNewlyIndexed)
	captionModelShouldRun := w.conf.VisionModelShouldRun(vision.ModelTypeCaption, vision.RunNewlyIndexed)

	for {
		photos, queryErr := query.PhotosMetadataUpdate(limit, offset, delay, interval)

		if queryErr != nil {
			return queryErr
		}

		if len(photos) == 0 {
			break
		}

		for _, photo := range photos {
			if mutex.MetaWorker.Canceled() {
				return errors.New("index: metadata worker canceled")
			}

			if done[photo.PhotoUID] {
				continue
			}

			done[photo.PhotoUID] = true

			generateLabels := labelsModelShouldRun && photo.ShouldGenerateLabels(false)
			generateCaption := captionModelShouldRun && photo.ShouldGenerateCaption(entity.SrcAuto, false)

			// If configured, generate metadata for newly indexed photos using external vision services.
			if photo.IsNewlyIndexed() && (generateLabels || generateCaption) {
				primaryFile, fileErr := photo.PrimaryFile()

				if fileErr != nil {
					log.Debugf("index: photo %s has invalid primary file (%s)", photo.PhotoUID, clean.Error(fileErr))
				} else {
					fileName := photoprism.FileName(primaryFile.FileRoot, primaryFile.FileName)
					mediaFile, mediaErr := photoprism.NewMediaFile(fileName)

					if mediaErr != nil || mediaFile == nil || !mediaFile.Ok() {
						if mediaErr != nil {
							log.Debugf("index: could not open primary file %s (generate metadata)", clean.Error(mediaErr))
						}
					} else {
						if generateLabels {
							if labels := ind.Labels(mediaFile, entity.SrcAuto); len(labels) > 0 {
								photo.AddLabels(labels)
							}
						}

						if generateCaption {
							if caption, captionErr := ind.Caption(mediaFile, entity.SrcAuto); captionErr != nil {
								log.Debugf("index: %s (generate caption for %s)", clean.Error(captionErr), photo.PhotoUID)
							} else if text := strings.TrimSpace(caption.Text); text != "" {
								photo.SetCaption(text, caption.Source)
								if updateErr := photo.UpdateCaptionLabels(); updateErr != nil {
									log.Warnf("index: %s (update caption labels for %s)", clean.Error(updateErr), photo.PhotoUID)
								}
							}
						}
					}
				}
			}

			updated, merged, optimizeErr := photo.Optimize(settings.StackMeta(), settings.StackUUID(), settings.Features.Estimates, force)

			if optimizeErr != nil {
				log.Errorf("index: %s in optimization worker", optimizeErr)
			} else if updated {
				optimized++
				log.Debugf("index: updated photo %s", photo.String())
			}

			for _, p := range merged {
				log.Infof("index: merged %s", p.PhotoUID)
				done[p.PhotoUID] = true
			}
		}

		if mutex.MetaWorker.Canceled() {
			return errors.New("index: metadata worker canceled")
		}

		offset += limit
	}

	if optimized > 0 {
		log.Infof("index: updated %s [%s]", english.Plural(optimized, "photo", "photos"), time.Since(start))
		updateIndex = true
	}

	// Only update index if necessary.
	if updateIndex {
		// Set photo quality scores to -1 if files are missing.
		if err = query.FlagHiddenPhotos(); err != nil {
			log.Warnf("index: %s in optimization worker", err)
		}

		// Run moments worker.
		if moments := photoprism.NewMoments(w.conf); moments == nil {
			log.Errorf("index: failed to update moments")
		} else if err = moments.Start(); err != nil {
			log.Warnf("moments: %s in optimization worker", err)
		}

		// Update precalculated photo and file counts.
		if err = entity.UpdateCounts(); err != nil {
			log.Warnf("index: %s in optimization worker", err)
		}

		// Update album, subject, and label cover thumbs.
		if err = query.UpdateCovers(); err != nil {
			log.Warnf("index: %s in optimization worker", err)
		}
	}

	return nil
}
