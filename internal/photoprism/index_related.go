package photoprism

import (
	"errors"
	"fmt"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// IndexRelated indexes a group of related files and returns the result.
func IndexRelated(related RelatedFiles, ind *Index, o IndexOptions) (result IndexResult) {
	// Skip if main file is nil.
	if related.Main == nil {
		result.Err = fmt.Errorf("index: no main media file for %s", clean.Log(related.String()))
		result.Status = IndexFailed
		return result
	}

	done := make(map[string]bool)
	result = IndexMain(&related, ind, o)

	switch {
	case result.Failed():
		return result
	case !result.Success():
		// Skip related files if indexing was not completely successful.
		return result
	case result.Stacked() && related.Len() > 1:
		// Show info if main file was stacked and has additional related files.
		log.Infof("index: %s has %s", related.MainLogName(), english.Plural(related.Count(), "related file", "related files"))
	}

	done[related.Main.FileName()] = true
	photoUID := result.PhotoUID

	i := 0

	for i < len(related.Files) {
		f := related.Files[i]
		i++

		if f == nil {
			continue
		}

		if done[f.FileName()] {
			continue
		}

		done[f.FileName()] = true

		// Skip files if the filename extension does not match their mime type,
		// see https://github.com/photoprism/photoprism/issues/3518 for details.
		if typeErr := f.CheckType(); typeErr != nil {
			result.Err = fmt.Errorf("index: skipped %s because it %w", clean.Log(f.RootRelName()), typeErr)
			result.Status = IndexFailed
			continue
		}

		// Show warning if sidecar file exceeds size or resolution limit.
		if _, limitErr := f.ExceedsBytes(o.ByteLimit); limitErr != nil {
			log.Warnf("index: %s", limitErr)
		} else if _, limitErr = f.ExceedsResolution(o.ResolutionLimit); limitErr != nil {
			log.Warnf("index: %s", limitErr)
		}

		// Create JSON sidecar file, if needed.
		if jsonErr := f.CreateExifToolJson(ind.convert); jsonErr != nil {
			log.Warnf("index: %s", clean.Error(jsonErr))
		}

		// Create JPEG sidecar for media files in other formats so that thumbnails can be created.
		if o.Convert && f.IsMedia() && !f.HasPreviewImage() {
			// Try to create a preview image; if this fails, log and continue without failing the whole group.
			if img, imgErr := ind.convert.ToImage(f, false); imgErr != nil {
				// Stop the run instead of masking a full disk as a generic preview error.
				if errors.Is(imgErr, status.ErrInsufficientStorage) {
					ind.abortInsufficientStorage()
					result.Err = imgErr
					result.Status = IndexFailed
					return result
				}

				log.Warnf("index: could not create preview image for %s (%s)", clean.Log(f.RootRelName()), imgErr)
				// Continue indexing other related files without changing the overall success status.
				continue
			} else if img == nil {
				log.Debugf("index: skipped creating preview image for %s", clean.Log(f.RootRelName()))
			} else {
				log.Debugf("index: created %s", clean.Log(img.BaseName()))

				// Skip with warning if thumbs could not be created.
				if thumbsErr := img.GenerateThumbnails(ind.thumbPath(), false); thumbsErr != nil {
					// Stop the run instead of masking a full disk as a generic thumbnail error.
					if errors.Is(thumbsErr, status.ErrInsufficientStorage) {
						ind.abortInsufficientStorage()
						result.Err = thumbsErr
						result.Status = IndexFailed
						return result
					}

					log.Warnf("index: failed to generate thumbnails for %s (%s)", clean.Log(f.RootRelName()), thumbsErr.Error())
					// Continue indexing; preview image exists and other related files may still succeed.
					continue
				}

				// Add preview image to list of files.
				img.SetRelatedMain(related.Main)
				related.Files = append(related.Files, img)
			}
		}

		// Index related MediaFile.
		f.SetRelatedMain(related.Main)
		res := ind.MediaFile(f, o, "", result.PhotoUID)
		if photoUID == "" && res.Success() {
			photoUID = res.PhotoUID
		}

		// Save file error.
		if fileUid, err := res.FileError(); err != nil {
			query.SetFileError(fileUid, err.Error())
		}

		// Log index result.
		log.Infof("index: %s related %s file %s", res, f.FileType(), clean.Log(f.RootRelName()))
	}

	// Reconcile the logical source once the group is done. UserMediaFile already
	// covered the primary preview and a distinct RAW/HEIC source's sidecars while
	// indexing it, so this pass only serves incremental sidecar-only updates
	// where the unchanged preview was filtered out of the list.
	if o.ImportFaceTags && photoUID != "" && !related.ContainsPreview() && isXmpFaceSource(related.Main) {
		// Collect first so a source that declares no region container skips both
		// the primary-file and marker queries. A declared but empty set still
		// proceeds: that is what removes markers whose regions the user deleted.
		if regions, collectErr := collectXmpFaces(related.Main); collectErr != nil {
			log.Warnf("index: %s while importing xmp face regions for %s", clean.Error(collectErr), related.MainLogName())
		} else if !regions.Declared {
			log.Tracef("index: no xmp face regions declared for %s", related.MainLogName())
		} else if primary, primaryErr := entity.PrimaryFile(photoUID); primaryErr != nil {
			log.Debugf("index: could not find primary file for xmp faces in %s (%s)", related.MainLogName(), clean.Error(primaryErr))
		} else if saved, _, applyErr := applyXmpFaceRegions(regions, primary); applyErr != nil {
			log.Warnf("index: %s while importing xmp face regions for %s", clean.Error(applyErr), related.MainLogName())
		} else if saved {
			log.Debugf("index: imported xmp face regions for %s", related.MainLogName())
		}
	}

	return result
}
