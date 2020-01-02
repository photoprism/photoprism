package photoprism

type IndexJob struct {
	related RelatedFiles
	opt     IndexOptions
	ind     *Index
}

func indexWorker(jobs <-chan IndexJob) {
	for job := range jobs {
		done := make(map[string]bool)
		related := job.related
		opt := job.opt
		ind := job.ind

		res := ind.MediaFile(related.main, opt)
		done[related.main.Filename()] = true

		log.Infof("index: %s main %s file \"%s\"", res, related.main.Type(), related.main.RelativeFilename(ind.originalsPath()))

		for _, f := range related.files {
			if done[f.Filename()] {
				continue
			}

			res := ind.MediaFile(f, opt)
			done[f.Filename()] = true

			log.Infof("index: %s related %s file \"%s\"", res, f.Type(), f.RelativeFilename(ind.originalsPath()))
		}
	}

}
