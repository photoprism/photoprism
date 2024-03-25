package photoprism

import (
	"fmt"

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

	// Check mime type, file size, and resolution.
	if f.WrongType() {
		// Skip files that have the wrong mimetype based on their filename extension:
		// https://github.com/photoprism/photoprism/issues/4118
		result.Err = fmt.Errorf("index: skipped %s due to wrong extension %s for mimetype %s", clean.Log(f.RootRelName()), clean.LogQuote(f.Extension()), clean.LogQuote(f.MimeType()))
		result.Status = IndexFailed
		return result
	} else if limitErr, _ := f.ExceedsBytes(o.ByteLimit); limitErr != nil {
		result.Err = fmt.Errorf("index: %s", limitErr)
		result.Status = IndexFailed
		return result
	} else if limitErr, _ = f.ExceedsResolution(o.ResolutionLimit); limitErr != nil {
		result.Err = fmt.Errorf("index: %s", limitErr)
		result.Status = IndexFailed
		return result
	}

	// Create JSON sidecar file, if needed.
	if jsonErr := f.CreateExifToolJson(ind.convert); jsonErr != nil {
		log.Errorf("index: %s", clean.Error(jsonErr))
	}

	// Create JPEG sidecar for media files in other formats so that thumbnails can be created.
	if o.Convert && f.IsMedia() && !f.HasPreviewImage() {
		if jpg, err := ind.convert.ToImage(f, false); err != nil {
			result.Err = fmt.Errorf("index: failed to create a preview for %s (%s)", clean.Log(f.RootRelName()), err.Error())
			result.Status = IndexFailed
			return result
		} else if limitErr, _ := jpg.ExceedsResolution(o.ResolutionLimit); limitErr != nil {
			result.Err = fmt.Errorf("index: %s", limitErr)
			result.Status = IndexFailed
			return result
		} else {
			log.Debugf("index: created %s", clean.Log(jpg.BaseName()))

			if err = jpg.CreateThumbnails(ind.thumbPath(), false); err != nil {
				result.Err = fmt.Errorf("index: failed to create thumbnails for %s (%s)", clean.Log(f.RootRelName()), err.Error())
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
			log.Errorf("index: %s to update main %s file %s", result, f.FileType(), clean.Log(f.RootRelName()))
		} else {
			log.Errorf("index: %s to add main %s file %s", result, f.FileType(), clean.Log(f.RootRelName()))
		}
	} else {
		log.Infof("index: %s main %s file %s", result, f.FileType(), clean.Log(f.RootRelName()))
	}

	return result
}
