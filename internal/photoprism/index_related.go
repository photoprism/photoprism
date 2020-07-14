package photoprism

import (
	"fmt"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// IndexMain indexes the main file from a group of related files and returns the result.
func IndexMain(related *RelatedFiles, ind *Index, opt IndexOptions) (result IndexResult) {
	// Skip sidecar files without related media file.
	if related.Main == nil {
		result.Err = fmt.Errorf("index: no main file found for %s", txt.Quote(related.String()))
		result.Status = IndexFailed
		return result
	}

	// Enforce file size limit for originals.
	if ind.conf.OriginalsLimit() > 0 && related.Main.FileSize() > ind.conf.OriginalsLimit() {
		result.Err = fmt.Errorf("index: %s exceeds file size limit for originals [%d / %d MB]", filepath.Base(related.Main.FileName()), related.Main.FileSize()/(1024*1024), ind.conf.OriginalsLimit()/(1024*1024))
		result.Status = IndexFailed
		return result
	}

	f := related.Main

	if opt.Convert && !f.HasJpeg() {
		if jpegFile, err := ind.convert.ToJpeg(f); err != nil {
			result.Err = fmt.Errorf("index: creating jpeg failed (%s)", err.Error())
			result.Status = IndexFailed

			return result
		} else {
			log.Infof("index: %s created", fs.RelName(jpegFile.FileName(), ind.originalsPath()))

			if err := jpegFile.ResampleDefault(ind.thumbPath(), false); err != nil {
				result.Err = fmt.Errorf("index: could not create default thumbnails (%s)", err.Error())
				result.Status = IndexFailed

				return result
			}

			related.Files = append(related.Files, jpegFile)
		}
	}

	if ind.conf.SidecarJson() && !f.HasJson() {
		if jsonFile, err := ind.convert.ToJson(f); err != nil {
			log.Errorf("index: creating json sidecar file failed (%s)", err.Error())
		} else {
			log.Infof("index: %s created", fs.RelName(jsonFile.FileName(), ind.originalsPath()))
		}
	}

	result = ind.MediaFile(f, opt, "")

	if result.Indexed() && f.IsJpeg() {
		if err := f.ResampleDefault(ind.thumbPath(), false); err != nil {
			log.Errorf("index: could not create default thumbnails (%s)", err.Error())
			query.SetFileError(result.FileUID, err.Error())
		}
	}

	log.Infof("index: %s main %s file %s", result, f.FileType(), txt.Quote(f.RelName(ind.originalsPath())))

	return result
}

// IndexMain indexes a group of related files and returns the result.
func IndexRelated(related RelatedFiles, ind *Index, opt IndexOptions) (result IndexResult) {
	done := make(map[string]bool)

	result = IndexMain(&related, ind, opt)

	if result.Failed() {
		log.Error(result.Err)
		return result
	} else if !result.Success() || result.Stacked() {
		// Skip related files if main file was stacked or indexing was not completely successful.
		return result
	}

	done[related.Main.FileName()] = true

	for _, f := range related.Files {
		if done[f.FileName()] {
			continue
		}

		res := ind.MediaFile(f, opt, "")
		done[f.FileName()] = true

		if res.Indexed() && f.IsJpeg() {
			if err := f.ResampleDefault(ind.thumbPath(), false); err != nil {
				log.Errorf("index: could not create default thumbnails (%s)", err.Error())
				query.SetFileError(res.FileUID, err.Error())
			}
		}

		log.Infof("index: %s related %s file %s", res, f.FileType(), txt.Quote(f.RelName(ind.originalsPath())))
	}

	return result
}
