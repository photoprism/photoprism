package photoprism

import (
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

type IndexJob struct {
	FileName string
	Related  RelatedFiles
	IndexOpt IndexOptions
	Ind      *Index
}

func IndexWorker(jobs <-chan IndexJob) {
	for job := range jobs {
		done := make(map[string]bool)
		related := job.Related
		opt := job.IndexOpt
		ind := job.Ind

		if related.Main != nil {
			res := ind.MediaFile(related.Main, opt, "")
			done[related.Main.FileName()] = true

			log.Infof("index: %s main %s file %s", res, related.Main.FileType(), txt.Quote(related.Main.RelativeName(ind.originalsPath())))
		} else {
			log.Warnf("index: no main file for %s (conversion failed?)", txt.Quote(fs.RelativeName(job.FileName, ind.originalsPath())))
		}

		for _, f := range related.Files {
			if done[f.FileName()] {
				continue
			}

			res := ind.MediaFile(f, opt, "")
			done[f.FileName()] = true

			log.Infof("index: %s related %s file %s", res, f.FileType(), txt.Quote(f.RelativeName(ind.originalsPath())))
		}
	}
}
