package photoprism

type IndexJob struct {
	FileName string
	Related  RelatedFiles
	IndexOpt IndexOptions
	Ind      *Index
}

func IndexWorker(jobs <-chan IndexJob) {
	for job := range jobs {
		if result := IndexRelated(job.Related, job.Ind, job.IndexOpt); result.Failed() {
			log.Debugf("index: failed to index %s with error: %s", job.FileName, result.Err)
		}
	}
}
