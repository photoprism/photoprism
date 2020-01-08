package photoprism

type ThumbnailsJob struct {
	mediaFile *MediaFile
	path      string
	force     bool
}

func thumbnailsWorker(jobs <-chan ThumbnailsJob) {
	for job := range jobs {
		if err := job.mediaFile.CreateDefaultThumbnails(job.path, job.force); err != nil {
			log.Errorf("thumbs: %s", err)
		}
	}
}
