package photoprism

type ConvertJob struct {
	image   *MediaFile
	convert *Convert
}

func convertWorker(jobs <-chan ConvertJob) {
	for job := range jobs {
		if _, err := job.convert.ToJpeg(job.image); err != nil {
			log.Warnf("convert: %s (%s)", err.Error(), job.image.FileName())
		}
	}
}
