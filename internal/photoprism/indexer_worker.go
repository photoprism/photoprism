package photoprism

type IndexJob struct {
	related RelatedFiles
	options IndexerOptions
	ind     *Indexer
}

func indexerWorker(jobs <-chan IndexJob) {
	for job := range jobs {
		indexed := make(map[string]bool)
		related := job.related
		options := job.options
		ind := job.ind

		mainIndexResult := ind.indexMediaFile(related.main, options)
		indexed[related.main.Filename()] = true

		log.Infof("index: %s main %s file \"%s\"", mainIndexResult, related.main.Type(), related.main.RelativeFilename(ind.originalsPath()))

		for _, relatedMediaFile := range related.files {
			if indexed[relatedMediaFile.Filename()] {
				continue
			}

			indexResult := ind.indexMediaFile(relatedMediaFile, options)
			indexed[relatedMediaFile.Filename()] = true

			log.Infof("index: %s related %s file \"%s\"", indexResult, relatedMediaFile.Type(), relatedMediaFile.RelativeFilename(ind.originalsPath()))
		}
	}

}
