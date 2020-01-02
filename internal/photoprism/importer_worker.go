package photoprism

import (
	"os"
	"path"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/event"
)

type ImportJob struct {
	related    RelatedFiles
	options    IndexerOptions
	importPath string
	imp        *Importer
}

func importerWorker(jobs <-chan ImportJob) {
	for job := range jobs {
		var destinationMainFilename string
		related := job.related
		imp := job.imp
		options := job.options
		importPath := job.importPath

		event.Publish("import.file", event.Data{
			"fileName": related.main.Filename(),
			"baseName": filepath.Base(related.main.Filename()),
		})

		for _, relatedMediaFile := range related.files {
			relativeFilename := relatedMediaFile.RelativeFilename(importPath)

			if destinationFilename, err := imp.DestinationFilename(related.main, relatedMediaFile); err == nil {
				if err := os.MkdirAll(path.Dir(destinationFilename), os.ModePerm); err != nil {
					log.Errorf("import: could not create directories (%s)", err.Error())
				}

				if related.main.HasSameFilename(relatedMediaFile) {
					destinationMainFilename = destinationFilename
					log.Infof("import: moving main %s file \"%s\" to \"%s\"", relatedMediaFile.Type(), relativeFilename, destinationFilename)
				} else {
					log.Infof("import: moving related %s file \"%s\" to \"%s\"", relatedMediaFile.Type(), relativeFilename, destinationFilename)
				}

				if err := relatedMediaFile.Move(destinationFilename); err != nil {
					log.Errorf("import: could not move file to \"%s\" (%s)", destinationMainFilename, err.Error())
				}
			} else if imp.removeExistingFiles {
				if err := relatedMediaFile.Remove(); err != nil {
					log.Errorf("import: could not delete file \"%s\" (%s)", relatedMediaFile.Filename(), err.Error())
				} else {
					log.Infof("import: deleted \"%s\" (already exists)", relativeFilename)
				}
			}
		}

		if destinationMainFilename != "" {
			importedMainFile, err := NewMediaFile(destinationMainFilename)

			if err != nil {
				log.Errorf("import: could not index \"%s\" (%s)", destinationMainFilename, err.Error())

				continue
			}

			if importedMainFile.IsRaw() {
				if _, err := imp.converter.ConvertToJpeg(importedMainFile); err != nil {
					log.Errorf("import: could not create jpeg from raw (%s)", err)
				}
			}
			if importedMainFile.IsHEIF() {
				if _, err := imp.converter.ConvertToJpeg(importedMainFile); err != nil {
					log.Errorf("import: could not create jpeg from heif (%s)", err)
				}
			}

			if jpg, err := importedMainFile.Jpeg(); err != nil {
				log.Error(err)
			} else {
				if err := jpg.CreateDefaultThumbnails(imp.conf.ThumbnailsPath(), false); err != nil {
					log.Errorf("import: could not create default thumbnails (%s)", err)
				}
			}

			indexed := make(map[string]bool)
			ind := imp.indexer
			mainIndexResult := ind.indexMediaFile(related.main, options)
			indexed[related.main.Filename()] = true

			log.Infof("import: indexed %s main %s file \"%s\"", mainIndexResult, related.main.Type(), related.main.RelativeFilename(ind.originalsPath()))

			for _, relatedMediaFile := range related.files {
				if indexed[relatedMediaFile.Filename()] {
					continue
				}

				indexResult := ind.indexMediaFile(relatedMediaFile, options)
				indexed[relatedMediaFile.Filename()] = true

				log.Infof("import: indexed %s related %s file \"%s\"", indexResult, relatedMediaFile.Type(), relatedMediaFile.RelativeFilename(ind.originalsPath()))
			}
		}
	}
}
