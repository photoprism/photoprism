package photoprism

type IndexJob struct {
	filename string
	related  RelatedFiles
	opt      IndexOptions
	ind      *Index
}

func indexWorker(jobs <-chan IndexJob) {
	for job := range jobs {
		done := make(map[string]bool)
		related := job.related
		opt := job.opt
		ind := job.ind

		if related.main != nil {
			res := ind.MediaFile(related.main, opt)
			done[related.main.Filename()] = true

			log.Infof("index: %s main %s file \"%s\"", res, related.main.Type(), related.main.RelativeFilename(ind.originalsPath()))
		} else {
			log.Warnf("index: no main file for %s (conversion to jpeg failed?)", job.filename)
		}

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
