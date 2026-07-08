package photoprism

import (
	"errors"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// ThumbsJob encapsulates thumbnail generation parameters for a media file.
type ThumbsJob struct {
	mediaFile *MediaFile
	path      string
	force     bool
}

// ThumbsWorker consumes thumbnail jobs and generates the requested previews.
func ThumbsWorker(jobs <-chan ThumbsJob) {
	for job := range jobs {
		mf := job.mediaFile

		if mf == nil {
			log.Error("thumbs: media file is nil - might be a bug")
			continue
		}

		if err := mf.GenerateThumbnails(job.path, job.force); err != nil {
			// Stop the run if the disk filled mid-scan instead of failing every remaining file.
			if errors.Is(err, status.ErrInsufficientStorage) {
				cancelThumbsInsufficientStorage()
				continue
			}

			log.Errorf("thumbs: %s", err)
		}
	}
}

// cancelThumbsInsufficientStorage logs the insufficient-storage cause once and cancels the
// index worker so the directory walk stops, instead of failing every remaining thumbnail job.
func cancelThumbsInsufficientStorage() {
	if mutex.IndexWorker.Canceled() {
		return
	}

	log.Errorf("thumbs: aborting due to insufficient storage")
	event.ErrorMsg(i18n.ErrInsufficientStorage)
	mutex.IndexWorker.Cancel()
}
