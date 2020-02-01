package photoprism

import "strings"

type ConvertJob struct {
	image   *MediaFile
	convert *Convert
}

func convertWorker(jobs <-chan ConvertJob) {
	for job := range jobs {
		if _, err := job.convert.ToJpeg(job.image); err != nil {
			fileName := job.image.RelativeName(job.convert.conf.OriginalsPath())
			log.Errorf("convert: could not create jpeg for %s (%s)", fileName, strings.TrimSpace(err.Error()))
		}
	}
}
