package photoprism

import (
	"fmt"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
)

// IndexRelated indexes a group of related files and returns the result.
func IndexRelated(related RelatedFiles, ind *Index, o IndexOptions) (result IndexResult) {
	// Skip if main file is nil.
	if related.Main == nil {
		result.Err = fmt.Errorf("index: no main file for %s", clean.Log(related.String()))
		result.Status = IndexFailed
		return result
	}

	done := make(map[string]bool)
	result = IndexMain(&related, ind, o)

	if result.Failed() {
		return result
	} else if !result.Success() {
		// Skip related files if indexing was not completely successful.
		return result
	} else if result.Stacked() && related.Len() > 1 {
		// Show info if main file was stacked and has additional related files.
		log.Infof("index: %s has %s", related.MainLogName(), english.Plural(related.Count(), "related file", "related files"))
	}

	done[related.Main.FileName()] = true

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

		// Show warning if sidecar file exceeds size or resolution limit.
		if limitErr, _ := f.ExceedsBytes(o.ByteLimit); limitErr != nil {
			log.Warnf("index: %s", limitErr)
		} else if limitErr, _ = f.ExceedsResolution(o.ResolutionLimit); limitErr != nil {
			log.Warnf("index: %s", limitErr)
		}

		// Create JSON sidecar file, if needed.
		if jsonErr := f.CreateExifToolJson(ind.convert); jsonErr != nil {
			log.Errorf("index: %s", clean.Log(jsonErr.Error()))
		}

		// Create JPEG sidecar for media files in other formats so that thumbnails can be created.
		if o.Convert && f.IsMedia() && !f.HasPreviewImage() {
			if jpg, err := ind.convert.ToImage(f, false); err != nil {
				result.Err = fmt.Errorf("index: failed creating preview for %s (%s)", clean.Log(f.RootRelName()), err.Error())
				result.Status = IndexFailed
				return result
			} else {
				log.Debugf("index: created %s", clean.Log(jpg.BaseName()))

				if err := jpg.CreateThumbnails(ind.thumbPath(), false); err != nil {
					result.Err = fmt.Errorf("index: failed creating thumbnails for %s (%s)", clean.Log(f.RootRelName()), err.Error())
					result.Status = IndexFailed
					return result
				}

				related.Files = append(related.Files, jpg)
			}
		}

		// Index related MediaFile.
		res := ind.MediaFile(f, o, "", result.PhotoUID)

		// Save file error.
		if fileUid, err := res.FileError(); err != nil {
			query.SetFileError(fileUid, err.Error())
		}

		// Log index result.
		log.Infof("index: %s related %s file %s", res, f.FileType(), clean.Log(f.RootRelName()))
	}

	return result
}
