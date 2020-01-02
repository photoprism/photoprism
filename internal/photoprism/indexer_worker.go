package photoprism

type IndexJob struct {
	r RelatedFiles
	o IndexerOptions
	i *Indexer
}

func indexerWorker(jobs <-chan IndexJob) {
	for job := range jobs {
		indexed := make(map[string]bool)
		r := job.r
		o := job.o
		i := job.i

		mainIndexResult := i.indexMediaFile(r.main, o)
		indexed[r.main.Filename()] = true

		log.Infof("index: %s main %s file \"%s\"", mainIndexResult, r.main.Type(), r.main.RelativeFilename(i.originalsPath()))

		for _, relatedMediaFile := range r.files {
			if indexed[relatedMediaFile.Filename()] {
				continue
			}

			indexResult := i.indexMediaFile(relatedMediaFile, o)
			indexed[relatedMediaFile.Filename()] = true

			log.Infof("index: %s related %s file \"%s\"", indexResult, relatedMediaFile.Type(), relatedMediaFile.RelativeFilename(i.originalsPath()))
		}
	}

}
