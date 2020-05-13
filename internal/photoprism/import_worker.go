package photoprism

import (
	"os"
	"path"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
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
			log.Warnf("import: no media file found for %s", txt.Quote(fs.RelativeName(job.FileName, importPath)))
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
					log.Errorf("import: could not create folders (%s)", err.Error())
				}

				if related.Main.HasSameName(f) {
					destinationMainFilename = destinationFilename
					log.Infof("import: moving main %s file %s to %s", f.FileType(), txt.Quote(relativeFilename), txt.Quote(fs.RelativeName(destinationFilename, imp.originalsPath())))
				} else {
					log.Infof("import: moving related %s file %s to %s", f.FileType(), txt.Quote(relativeFilename), txt.Quote(fs.RelativeName(destinationFilename, imp.originalsPath())))
				}

				if opt.Move {
					if err := f.Move(destinationFilename); err != nil {
						log.Errorf("import: could not move file to %s (%s)", txt.Quote(fs.RelativeName(destinationMainFilename, imp.originalsPath())), err.Error())
					}
				} else {
					if err := f.Copy(destinationFilename); err != nil {
						log.Errorf("import: could not copy file to %s (%s)", txt.Quote(fs.RelativeName(destinationMainFilename, imp.originalsPath())), err.Error())
					}
				}
			} else {
				log.Warnf("import: %s", err)

				if opt.RemoveExistingFiles {
					if err := f.Remove(); err != nil {
						log.Errorf("import: could not delete %s (%s)", txt.Quote(fs.RelativeName(f.FileName(), importPath)), err.Error())
					} else {
						log.Infof("import: deleted %s (already exists)", txt.Quote(relativeFilename))
					}
				}
			}
		}

		if destinationMainFilename != "" {
			f, err := NewMediaFile(destinationMainFilename)

			if err != nil {
				log.Errorf("import: could not index %s (%s)", txt.Quote(fs.RelativeName(destinationMainFilename, imp.originalsPath())), err.Error())

				continue
			}

			if !f.HasJpeg() {
				if _, err := imp.convert.ToJpeg(f); err != nil {
					log.Errorf("import: creating jpeg failed (%s)", err.Error())
				}
			}

			if jpg, err := f.Jpeg(); err != nil {
				log.Error(err)
			} else {
				if err := jpg.ResampleDefault(imp.conf.ThumbPath(), false); err != nil {
					log.Errorf("import: could not create default thumbnails (%s)", err.Error())
				}
			}

			if imp.conf.SidecarJson() && !f.HasJson() {
				if jsonFile, err := imp.convert.ToJson(f); err != nil {
					log.Errorf("import: creating json sidecar file failed (%s)", err.Error())
				} else {
					related.Files = append(related.Files, jsonFile)
				}
			}

			related, err := f.RelatedFiles(imp.conf.Settings().Index.Group)

			if err != nil {
				log.Errorf("import: could not index %s (%s)", txt.Quote(fs.RelativeName(destinationMainFilename, imp.originalsPath())), err.Error())

				continue
			}

			done := make(map[string]bool)
			ind := imp.index

			if related.Main != nil {
				res := ind.MediaFile(related.Main, indexOpt, originalName)
				log.Infof("import: %s main %s file %s", res, related.Main.FileType(), txt.Quote(related.Main.RelativeName(ind.originalsPath())))
				done[related.Main.FileName()] = true
			} else {
				log.Warnf("import: no main file for %s (conversion to jpeg failed?)", fs.RelativeName(destinationMainFilename, imp.originalsPath()))
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

				log.Infof("import: %s related %s file %s", res, f.FileType(), txt.Quote(f.RelativeName(ind.originalsPath())))
			}
		}
	}
}
