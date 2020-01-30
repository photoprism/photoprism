package photoprism

import (
	"os"
	"path"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/event"
)

type ImportJob struct {
	filename string
	related  RelatedFiles
	opt      IndexOptions
	path     string
	imp      *Import
}

func importWorker(jobs <-chan ImportJob) {
	for job := range jobs {
		var destinationMainFilename string
		related := job.related
		imp := job.imp
		opt := job.opt
		importPath := job.path

		event.Publish("import.file", event.Data{
			"fileName": related.main.Filename(),
			"baseName": filepath.Base(related.main.Filename()),
		})

		for _, f := range related.files {
			relativeFilename := f.RelativeFilename(importPath)

			if destinationFilename, err := imp.DestinationFilename(related.main, f); err == nil {
				if err := os.MkdirAll(path.Dir(destinationFilename), os.ModePerm); err != nil {
					log.Errorf("import: could not create directories (%s)", err.Error())
				}

				if related.main.HasSameFilename(f) {
					destinationMainFilename = destinationFilename
					log.Infof("import: moving main %s file \"%s\" to \"%s\"", f.Type(), relativeFilename, destinationFilename)
				} else {
					log.Infof("import: moving related %s file \"%s\" to \"%s\"", f.Type(), relativeFilename, destinationFilename)
				}

				if err := f.Move(destinationFilename); err != nil {
					log.Errorf("import: could not move file to \"%s\" (%s)", destinationMainFilename, err.Error())
				}
			} else if imp.removeExistingFiles {
				if err := f.Remove(); err != nil {
					log.Errorf("import: could not delete file \"%s\" (%s)", f.Filename(), err.Error())
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

			if importedMainFile.IsRaw() || importedMainFile.IsHEIF() || importedMainFile.IsImageOther() {
				if _, err := imp.convert.ToJpeg(importedMainFile); err != nil {
					log.Errorf("import: creating jpeg failed (%s)", err.Error())
				}
			}

			if jpg, err := importedMainFile.Jpeg(); err != nil {
				log.Error(err)
			} else {
				if err := jpg.RenderDefaultThumbnails(imp.conf.ThumbnailsPath(), false); err != nil {
					log.Errorf("import: could not create default thumbnails (%s)", err.Error())
				}
			}

			related, err := importedMainFile.RelatedFiles()

			if err != nil {
				log.Errorf("import: could not index \"%s\" (%s)", destinationMainFilename, err.Error())

				continue
			}

			done := make(map[string]bool)
			ind := imp.index

			if related.main != nil {
				res := ind.MediaFile(related.main, opt)
				log.Infof("import: %s main %s file \"%s\"", res, related.main.Type(), related.main.RelativeFilename(ind.originalsPath()))
				done[related.main.Filename()] = true
			} else {
				log.Warnf("import: no main file for %s (conversion to jpeg failed?)", destinationMainFilename)
			}

			for _, f := range related.files {
				if f == nil {
					continue
				}

				if done[f.Filename()] {
					continue
				}

				res := ind.MediaFile(f, opt)
				done[f.Filename()] = true

				log.Infof("import: %s related %s file \"%s\"", res, f.Type(), f.RelativeFilename(ind.originalsPath()))
			}
		}
	}
}
