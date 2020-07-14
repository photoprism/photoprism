package photoprism

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

type ConvertJob struct {
	image   *MediaFile
	convert *Convert
}

func ConvertWorker(jobs <-chan ConvertJob) {
	for job := range jobs {
		if _, err := job.convert.ToJpeg(job.image); err != nil {
			fileName := job.image.RelName(job.convert.conf.OriginalsPath())
			log.Errorf("convert: %s in %s (jpeg)", strings.TrimSpace(err.Error()), txt.Quote(fileName))
		}
	}
}
