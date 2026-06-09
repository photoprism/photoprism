package photoprism

import (
	"errors"
	"strings"

	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// ConvertJob represents a single media conversion task.
type ConvertJob struct {
	force   bool
	file    *MediaFile
	convert *Convert
}

// ConvertWorker processes ConvertJob messages serially on a worker goroutine.
func ConvertWorker(jobs <-chan ConvertJob) {
	// handleErr logs a conversion failure, or stops the run if the disk filled mid-convert.
	// It returns true when the caller should skip the remaining steps for the current file.
	handleErr := func(err error, job ConvertJob) bool {
		if errors.Is(err, status.ErrInsufficientStorage) {
			job.convert.cancelInsufficientStorage()
			return true
		}

		fileName := job.file.RelName(job.convert.conf.OriginalsPath())
		log.Errorf("convert: %s for %s", strings.TrimSpace(err.Error()), clean.Log(fileName))
		return false
	}

	for job := range jobs {
		// File and convert service must not be nil.
		if job.file == nil || job.convert == nil {
			continue
		}

		// Drain remaining queued jobs without processing once the run was canceled.
		if mutex.IndexWorker.Canceled() {
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
				if handleErr(err, job) {
					continue
				}
			}

			// Check if the file has a playable format or has already been transcoded.
			if f.SkipTranscoding() {
				log.Debugf("convert: %s does not require transcoding", clean.Log(f.RelName(job.convert.conf.OriginalsPath())))
				continue
			}

			// Transcode to MP4 AVC.
			if _, err := job.convert.ToAvc(f, job.convert.conf.FFmpegEncoder(), false, false); err != nil {
				handleErr(err, job)
			}
		default:
			// Create preview image.
			if _, err := job.convert.ToImage(f, job.force); err != nil {
				handleErr(err, job)
			}
		}
	}
}
