package photoprism

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

type ConvertJob struct {
	file    *MediaFile
	convert *Convert
}

func ConvertWorker(jobs <-chan ConvertJob) {
	logError := func(err error, job ConvertJob) {
		fileName := job.file.RelName(job.convert.conf.OriginalsPath())
		log.Errorf("%s in %s", strings.TrimSpace(err.Error()), txt.Quote(fileName))
	}

	for job := range jobs {
		switch {
		case job.file == nil:
			continue
		case job.convert == nil:
			continue
		case job.file.IsVideo():
			if _, err := job.convert.ToJson(job.file); err != nil {
				logError(err, job)
			} else if _, err := job.convert.ToJpeg(job.file); err != nil {
				logError(err, job)
			} else if metaData := job.file.MetaData(); metaData.CodecAvc() {
				continue
			} else if _, err := job.convert.ToAvc(job.file, job.convert.conf.FFmpegCodec()); err != nil {
				logError(err, job)
			}
		default:
			if _, err := job.convert.ToJpeg(job.file); err != nil {
				logError(err, job)
			}
		}
	}
}
