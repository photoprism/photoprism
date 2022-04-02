package photoprism

import (
	"fmt"
	"path/filepath"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/query"

	"github.com/photoprism/photoprism/pkg/sanitize"
)

// IndexRelated indexes a group of related files and returns the result.
func IndexRelated(related RelatedFiles, ind *Index, o IndexOptions) (result IndexResult) {
	// Skip if main file is nil.
	if related.Main == nil {
		result.Err = fmt.Errorf("index: no main file for %s", sanitize.Log(related.String()))
		result.Status = IndexFailed
		return result
	}

	done := make(map[string]bool)
	result = IndexMain(&related, ind, o)

	if result.Failed() {
		log.Error(result.Err)
		return result
	} else if !result.Success() {
		// Skip related files if indexing was not completely successful.
		return result
	} else if !result.Indexed() {
		// Skip related files if main file was not indexed but for example skipped.
		if related.Len() > 1 {
			log.Warnf("index: %s has %s", related.MainLogName(), english.Plural(related.Count(), "related file", "related files"))
		}
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
		if exceeds, actual := f.ExceedsFileSize(o.OriginalsLimit); exceeds {
			log.Warnf("index: sidecar file %s exceeds size limit (%d / %d MB)", sanitize.Log(f.RootRelName()), actual, o.OriginalsLimit)
		} else if exceeds, actual = f.ExceedsResolution(o.ResolutionLimit); exceeds {
			log.Warnf("index: sidecar file %s exceeds resolution limit (%d / %d MP)", sanitize.Log(f.RootRelName()), actual, o.ResolutionLimit)
		}

		// Extract metadata to a JSON file with Exiftool.
		if f.NeedsExifToolJson() {
			if jsonName, err := ind.convert.ToJson(f); err != nil {
				log.Debugf("index: %s in %s (extract metadata)", sanitize.Log(err.Error()), sanitize.Log(f.RootRelName()))
			} else {
				log.Debugf("index: created %s", filepath.Base(jsonName))
			}
		}

		// Create JPEG sidecar for media files in other formats so that thumbnails can be created.
		if o.Convert && f.IsMedia() && !f.HasJpeg() {
			if jpg, err := ind.convert.ToJpeg(f); err != nil {
				result.Err = fmt.Errorf("index: failed converting %s to jpeg (%s)", sanitize.Log(f.RootRelName()), err.Error())
				result.Status = IndexFailed
				return result
			} else {
				log.Debugf("index: created %s", sanitize.Log(jpg.BaseName()))

				if err := jpg.CreateThumbnails(ind.thumbPath(), false); err != nil {
					result.Err = fmt.Errorf("index: failed creating thumbnails for %s (%s)", sanitize.Log(f.RootRelName()), err.Error())
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
		log.Infof("index: %s related %s file %s", res, f.FileType(), sanitize.Log(f.RootRelName()))
	}

	return result
}
