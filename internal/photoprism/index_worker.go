package photoprism

import (
	"path/filepath"

	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

type IndexJob struct {
	FileName string
	Related  RelatedFiles
	IndexOpt IndexOptions
	Ind      *Index
}

func IndexWorker(jobs <-chan IndexJob) {
	for job := range jobs {
		done := make(map[string]bool)
		related := job.Related
		opt := job.IndexOpt
		ind := job.Ind

		// Skip sidecar files without related media file.
		if related.Main == nil {
			log.Warnf("index: no media file found for %s", txt.Quote(fs.RelativeName(job.FileName, ind.originalsPath())))
			continue
		}

		// Enforce file size limit for originals.
		if ind.conf.OriginalsLimit() > 0 && related.Main.FileSize() > ind.conf.OriginalsLimit() {
			log.Warnf("index: %s exceeds file size limit for originals [%d / %d MB]", filepath.Base(related.Main.FileName()), related.Main.FileSize()/(1024*1024), ind.conf.OriginalsLimit()/(1024*1024))
			continue
		}

		f := related.Main

		if opt.Convert && !f.HasJpeg() {
			if jpegFile, err := ind.convert.ToJpeg(f, ind.conf.JpegHidden()); err != nil {
				log.Errorf("index: creating jpeg failed (%s)", err.Error())
				continue
			} else {
				log.Infof("index: %s created", fs.RelativeName(jpegFile.FileName(), ind.originalsPath()))

				if err := jpegFile.ResampleDefault(ind.thumbPath(), false); err != nil {
					log.Errorf("index: could not create default thumbnails (%s)", err.Error())
					continue
				}

				related.Files = append(related.Files, jpegFile)
			}
		}

		if ind.conf.SidecarJson() && !f.HasJson() {
			if jsonFile, err := ind.convert.ToJson(f, ind.conf.SidecarHidden()); err != nil {
				log.Errorf("index: creating json sidecar file failed (%s)", err.Error())
			} else {
				log.Infof("index: %s created", fs.RelativeName(jsonFile.FileName(), ind.originalsPath()))
			}
		}

		res := ind.MediaFile(f, opt, "")
		done[f.FileName()] = true

		if (res.Status == IndexAdded || res.Status == IndexUpdated) && f.IsJpeg() {
			if err := f.ResampleDefault(ind.thumbPath(), false); err != nil {
				log.Errorf("index: could not create default thumbnails (%s)", err.Error())
				query.SetFileError(res.FileUID, err.Error())
			}
		}

		log.Infof("index: %s main %s file %s", res, f.FileType(), txt.Quote(f.RelativeName(ind.originalsPath())))

		if !res.Success() {
			continue
		}

		for _, f := range related.Files {
			if done[f.FileName()] {
				continue
			}

			res := ind.MediaFile(f, opt, "")
			done[f.FileName()] = true

			if (res.Status == IndexAdded || res.Status == IndexUpdated) && f.IsJpeg() {
				if err := f.ResampleDefault(ind.thumbPath(), false); err != nil {
					log.Errorf("index: could not create default thumbnails (%s)", err.Error())
					query.SetFileError(res.FileUID, err.Error())
				}
			}

			log.Infof("index: %s related %s file %s", res, f.FileType(), txt.Quote(f.RelativeName(ind.originalsPath())))
		}
	}
}
