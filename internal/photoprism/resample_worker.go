package photoprism

type ResampleJob struct {
	mediaFile *MediaFile
	path      string
	force     bool
}

func resampleWorker(jobs <-chan ResampleJob) {
	for job := range jobs {
		mf := job.mediaFile

		if mf == nil {
			log.Error("resample: media file is nil - might be a bug")
			continue
		}

		if err := mf.ResampleDefault(job.path, job.force); err != nil {
			log.Errorf("resample: %s", err)
		}
	}
}
