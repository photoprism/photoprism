package photoprism

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
)

type ConvertJob struct {
	force   bool
	file    *MediaFile
	convert *Convert
}

func ConvertWorker(jobs <-chan ConvertJob) {
	logError := func(err error, job ConvertJob) {
		fileName := job.file.RelName(job.convert.conf.OriginalsPath())
		log.Errorf("convert: %s for %s", strings.TrimSpace(err.Error()), clean.Log(fileName))
	}

	for job := range jobs {
		switch {
		case job.file == nil:
			continue
		case job.convert == nil:
			continue
		case job.file.IsAnimated():
			_, _ = job.convert.ToJson(job.file, false)

			// Create JPEG preview and AVC encoded version for videos.
			if _, err := job.convert.ToImage(job.file, job.force); err != nil {
				logError(err, job)
			} else if metaData := job.file.MetaData(); metaData.CodecAvc() {
				continue
			} else if _, err := job.convert.ToAvc(job.file, job.convert.conf.FFmpegEncoder(), false, false); err != nil {
				logError(err, job)
			}
		default:
			if _, err := job.convert.ToImage(job.file, job.force); err != nil {
				logError(err, job)
			}
		}
	}
}
