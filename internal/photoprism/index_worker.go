package photoprism

type IndexJob struct {
	FileName string
	Related  RelatedFiles
	IndexOpt IndexOptions
	Ind      *Index
}

func IndexWorker(jobs <-chan IndexJob) {
	for job := range jobs {
		IndexRelated(job.Related, job.Ind, job.IndexOpt)
	}
}
