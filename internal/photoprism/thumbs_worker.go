package photoprism

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
			log.Errorf("thumbs: %s", err)
		}
	}
}
