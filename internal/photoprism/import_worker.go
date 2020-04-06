package photoprism

import (
	"os"
	"path"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/event"
)

type ImportJob struct {
	FileName  string
	Related   RelatedFiles
	IndexOpt  IndexOptions
	ImportOpt ImportOptions
	Imp       *Import
}

func ImportWorker(jobs <-chan ImportJob) {
	for job := range jobs {
		var destinationMainFilename string
		related := job.Related
		imp := job.Imp
		opt := job.ImportOpt
		indexOpt := job.IndexOpt
		importPath := job.ImportOpt.Path

		if related.Main == nil {
			log.Warnf("import: no main file found for %s", job.FileName)
			continue
		}

		originalName := related.Main.RelativeName(importPath)

		event.Publish("import.file", event.Data{
			"fileName": originalName,
			"baseName": filepath.Base(related.Main.FileName()),
		})

		for _, f := range related.Files {
			relativeFilename := f.RelativeName(importPath)

			if destinationFilename, err := imp.DestinationFilename(related.Main, f); err == nil {
				if err := os.MkdirAll(path.Dir(destinationFilename), os.ModePerm); err != nil {
					log.Errorf("import: could not create directories (%s)", err.Error())
				}

				if related.Main.HasSameName(f) {
					destinationMainFilename = destinationFilename
					log.Infof("import: moving main %s file \"%s\" to \"%s\"", f.Type(), relativeFilename, destinationFilename)
				} else {
					log.Infof("import: moving related %s file \"%s\" to \"%s\"", f.Type(), relativeFilename, destinationFilename)
				}

				if opt.Move {
					if err := f.Move(destinationFilename); err != nil {
						log.Errorf("import: could not move file to %s (%s)", destinationMainFilename, err.Error())
					}
				} else {
					if err := f.Copy(destinationFilename); err != nil {
						log.Errorf("import: could not copy file to %s (%s)", destinationMainFilename, err.Error())
					}
				}
			} else if opt.RemoveExistingFiles {
				if err := f.Remove(); err != nil {
					log.Errorf("import: could not delete %s (%s)", f.FileName(), err.Error())
				} else {
					log.Infof("import: deleted %s (already exists)", relativeFilename)
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
				if err := jpg.ResampleDefault(imp.conf.ThumbnailsPath(), false); err != nil {
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

			if related.Main != nil {
				res := ind.MediaFile(related.Main, indexOpt, originalName)
				log.Infof("import: %s main %s file \"%s\"", res, related.Main.Type(), related.Main.RelativeName(ind.originalsPath()))
				done[related.Main.FileName()] = true
			} else {
				log.Warnf("import: no main file for %s (conversion to jpeg failed?)", destinationMainFilename)
			}

			for _, f := range related.Files {
				if f == nil {
					continue
				}

				if done[f.FileName()] {
					continue
				}

				res := ind.MediaFile(f, indexOpt, "")
				done[f.FileName()] = true

				log.Infof("import: %s related %s file \"%s\"", res, f.Type(), f.RelativeName(ind.originalsPath()))
			}
		}
	}
}
