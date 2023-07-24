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
		// File and convert service must not be nil.
		if job.file == nil || job.convert == nil {
			continue
		}

		// f is the media file to be converted.
		f := job.file

		switch {
		case f.IsAnimated():
			// Extract metadata.
			_, _ = job.convert.ToJson(f, false)

			// Create cover image.
			if _, err := job.convert.ToImage(f, job.force); err != nil {
				logError(err, job)
			}

			// Check if the file has a playable format or has already been transcoded.
			if f.SkipTranscoding() {
				log.Debugf("convert: %s does not require transcoding", clean.Log(f.RelName(job.convert.conf.OriginalsPath())))
				continue
			}

			// Transcode to MP4 AVC.
			if _, err := job.convert.ToAvc(f, job.convert.conf.FFmpegEncoder(), false, false); err != nil {
				logError(err, job)
			}
		default:
			// Create preview image.
			if _, err := job.convert.ToImage(f, job.force); err != nil {
				logError(err, job)
			}
		}
	}
}
