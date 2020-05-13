package photoprism

import (
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

		if related.Main == nil {
			log.Warnf("index: no media file found for %s", txt.Quote(fs.RelativeName(job.FileName, ind.originalsPath())))
			continue
		}

		f := related.Main

		if opt.Convert && !f.HasJpeg() {
			if converted, err := ind.convert.ToJpeg(f); err != nil {
				log.Errorf("index: creating jpeg failed (%s)", err.Error())
			} else {
				if err := converted.ResampleDefault(ind.thumbPath(), false); err != nil {
					log.Errorf("index: could not create default thumbnails (%s)", err.Error())
				}

				related.Files = append(related.Files, converted)
			}
		}

		if ind.conf.SidecarJson() && !f.HasJson() {
			if jsonFile, err := ind.convert.ToJson(f); err != nil {
				log.Errorf("index: creating json sidecar file failed (%s)", err.Error())
			} else {
				related.Files = append(related.Files, jsonFile)
			}
		}

		res := ind.MediaFile(f, opt, "")
		done[f.FileName()] = true

		if (res.Status == IndexAdded || res.Status == IndexUpdated) && f.IsJpeg() {
			if err := f.ResampleDefault(ind.thumbPath(), false); err != nil {
				log.Errorf("index: could not create default thumbnails (%s)", err.Error())
			}
		}

		log.Infof("index: %s main %s file %s", res, f.FileType(), txt.Quote(f.RelativeName(ind.originalsPath())))

		for _, f := range related.Files {
			if done[f.FileName()] {
				continue
			}

			res := ind.MediaFile(f, opt, "")
			done[f.FileName()] = true

			if (res.Status == IndexAdded || res.Status == IndexUpdated) && f.IsJpeg() {
				if err := f.ResampleDefault(ind.thumbPath(), false); err != nil {
					log.Errorf("index: could not create default thumbnails (%s)", err.Error())
				}
			}

			log.Infof("index: %s related %s file %s", res, f.FileType(), txt.Quote(f.RelativeName(ind.originalsPath())))
		}
	}
}
