package photoprism

import (
	"fmt"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
)

// IndexMain indexes the main file from a group of related files and returns the result.
func IndexMain(related *RelatedFiles, ind *Index, o IndexOptions) (result IndexResult) {
	// Skip if main file is nil.
	if related.Main == nil {
		result.Err = fmt.Errorf("index: no main file for %s", clean.Log(related.String()))
		result.Status = IndexFailed
		return result
	}

	f := related.Main

	// Enforce file size and resolution limits.
	if limitErr, _ := f.ExceedsBytes(o.ByteLimit); limitErr != nil {
		result.Err = fmt.Errorf("index: %s", limitErr)
		result.Status = IndexFailed
		return result
	} else if limitErr, _ = f.ExceedsResolution(o.ResolutionLimit); limitErr != nil {
		result.Err = fmt.Errorf("index: %s", limitErr)
		result.Status = IndexFailed
		return result
	}

	// Extract metadata to a JSON file with Exiftool.
	if f.NeedsExifToolJson() {
		if jsonName, err := ind.convert.ToJson(f, false); err != nil {
			log.Tracef("exiftool: %s", clean.Log(err.Error()))
			log.Debugf("exiftool: failed parsing %s", clean.Log(f.RootRelName()))
		} else {
			log.Debugf("index: created %s", filepath.Base(jsonName))
		}
	}

	// Create JPEG sidecar for media files in other formats so that thumbnails can be created.
	if o.Convert && f.IsMedia() && !f.HasPreviewImage() {
		if jpg, err := ind.convert.ToImage(f, false); err != nil {
			result.Err = fmt.Errorf("index: failed creating preview for %s (%s)", clean.Log(f.RootRelName()), err.Error())
			result.Status = IndexFailed
			return result
		} else if limitErr, _ := jpg.ExceedsResolution(o.ResolutionLimit); limitErr != nil {
			result.Err = fmt.Errorf("index: %s", limitErr)
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

	// Index main MediaFile.
	exists := ind.files.Exists(f.RootRelName(), f.Root())
	result = ind.MediaFile(f, o, "", "")

	// Save file error.
	if fileUid, err := result.FileError(); err != nil {
		query.SetFileError(fileUid, err.Error())
	}

	// Log index result.
	if result.Failed() {
		log.Error(result.Err)

		if exists {
			log.Errorf("index: %s updating main %s file %s", result, f.FileType(), clean.Log(f.RootRelName()))
		} else {
			log.Errorf("index: %s adding main %s file %s", result, f.FileType(), clean.Log(f.RootRelName()))
		}
	} else {
		log.Infof("index: %s main %s file %s", result, f.FileType(), clean.Log(f.RootRelName()))
	}

	return result
}
