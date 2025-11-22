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

	// Refresh index metadata.
	log.Debugf("index: updating metadata")

	start := time.Now()
	settings := w.conf.Settings()
	done := make(map[string]bool)
	limit := 1000
	offset := 0
	optimized := 0
	updateFaces := updateIndex || entity.UpdateFaces.Load()

	labelsModelShouldRun := w.conf.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunNewlyIndexed)
	captionModelShouldRun := w.conf.VisionModelShouldRun(vision.ModelTypeCaption, vision.RunNewlyIndexed)
	nsfwModelShouldRun := w.conf.VisionModelShouldRun(vision.ModelTypeNsfw, vision.RunNewlyIndexed)
	detectFaces := w.conf.VisionModelShouldRun(vision.ModelTypeFace, vision.RunNewlyIndexed)

	if nsfwModelShouldRun {
		log.Debugf("index: cannot run %s model on %s", vision.ModelTypeNsfw, vision.RunNewlyIndexed)
	}

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

			logName := photo.String()

			// Track whether any persistence happened for the current photo.
			updated := false

			generateLabels := labelsModelShouldRun && photo.ShouldGenerateLabels(false)
			generateCaption := captionModelShouldRun && photo.ShouldGenerateCaption(entity.SrcAuto, false)

			// If configured, generate metadata for newly indexed photos using external vision services.
			if photo.IsNewlyIndexed() && (detectFaces || generateLabels || generateCaption) {
				primaryFile, fileErr := photo.PrimaryFile()

				if fileErr != nil {
					log.Debugf("index: photo %s has invalid primary file (%s)", logName, clean.Error(fileErr))
				} else {
					fileName := photoprism.FileName(primaryFile.FileRoot, primaryFile.FileName)

					// Load original media file.
					mediaFile, mediaErr := photoprism.NewMediaFile(fileName)

					if mediaErr != nil || mediaFile == nil || !mediaFile.Ok() {
						if mediaErr != nil {
							log.Debugf("index: could not open primary file %s (generate metadata)", clean.Error(mediaErr))
						}
					} else {
						// Record whether any in-memory metadata changed before saving.
						changed := false

						// Detect faces.
						if detectFaces {
							if markers := primaryFile.Markers(); markers == nil {
								log.Errorf("index: failed loading markers for %s", logName)
							} else {
								expected := markers.DetectedFaceCount()
								faces, detectErr := photoprism.DetectFaces(mediaFile, expected)

								if detectErr != nil {
									log.Debugf("vision: %s in %s (detect faces)", detectErr, clean.Log(mediaFile.BaseName()))
								} else if saved, count, applyErr := photoprism.ApplyDetectedFaces(primaryFile, faces); applyErr != nil {
									log.Warnf("index: %s in %s (save faces)", clean.Error(applyErr), logName)
								} else if saved {
									photo.PhotoFaces = count
									updateFaces = true
									changed = true
								}
							}
						}

						// Generate photo labels if needed.
						if generateLabels {
							if labels := mediaFile.GenerateLabels(entity.SrcAuto); len(labels) > 0 {
								if w.conf.DetectNSFW() && !photo.PhotoPrivate {
									if labels.IsNSFW(vision.Config.Thresholds.GetNSFW()) {
										photo.PhotoPrivate = true
										log.Infof("vision: changed private flag of %s to %t (labels)", logName, photo.PhotoPrivate)
									}
								}
								photo.AddLabels(labels)
								changed = true
							}
						}

						// Generate photo caption if needed.
						if generateCaption {
							if caption, captionErr := mediaFile.GenerateCaption(entity.SrcAuto); captionErr != nil {
								log.Debugf("index: failed to generate caption for %s (%s)", logName, clean.Error(captionErr))
							} else if text := strings.TrimSpace(caption.Text); text != "" {
								photo.SetCaption(text, caption.Source)
								if updateErr := photo.UpdateCaptionLabels(); updateErr != nil {
									log.Warnf("index: failed to update caption labels for %s (%s)", logName, clean.Error(updateErr))
								}
								changed = true
							}
						}

						// Persist the derived metadata updates (title, label counts) once per photo.
						if changed {
							if saveErr := photo.SaveVision(); saveErr == nil {
								updated = true
							}
						}
					}
				}

			}

			saved, merged, optimizeErr := photo.Optimize(settings.StackMeta(), settings.StackUUID(), settings.Features.Estimates, force)

			if optimizeErr != nil {
				log.Errorf("index: %s in optimization worker", optimizeErr)
			} else if saved {
				updated = true
				log.Debugf("index: updated photo %s", logName)
			}

			if updated {
				optimized++
			}

			for _, p := range merged {
				if p != nil {
					log.Infof("index: merged %s", p.String())
					done[p.PhotoUID] = true
				}
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

	// Perform face recognition.
	if updateFaces {
		log.Debugf("index: running face recognition")
		if faces := photoprism.NewFaces(w.conf); faces.Disabled() {
			log.Debugf("index: skipping face recognition")
		} else if facesErr := faces.Start(photoprism.FacesOptions{}); facesErr != nil {
			log.Warn(facesErr)
		}
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
