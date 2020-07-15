package photoprism

import (
	"os"
	"path"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/entity"
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
			log.Warnf("import: %s belongs to no supported media file", txt.Quote(fs.RelName(job.FileName, importPath)))
			continue
		}

		originalName := related.Main.RelName(importPath)

		event.Publish("import.file", event.Data{
			"fileName": originalName,
			"baseName": filepath.Base(related.Main.FileName()),
		})

		for _, f := range related.Files {
			relativeFilename := f.RelName(importPath)

			if destinationFilename, err := imp.DestinationFilename(related.Main, f); err == nil {
				if err := os.MkdirAll(path.Dir(destinationFilename), os.ModePerm); err != nil {
					log.Errorf("import: failed creating folders for %s (%s)", txt.Quote(f.BaseName()), err.Error())
				}

				if related.Main.HasSameName(f) {
					destinationMainFilename = destinationFilename
					log.Infof("import: moving main %s file %s to %s", f.FileType(), txt.Quote(relativeFilename), txt.Quote(fs.RelName(destinationFilename, imp.originalsPath())))
				} else {
					log.Infof("import: moving related %s file %s to %s", f.FileType(), txt.Quote(relativeFilename), txt.Quote(fs.RelName(destinationFilename, imp.originalsPath())))
				}

				if opt.Move {
					if err := f.Move(destinationFilename); err != nil {
						log.Errorf("import: failed moving file to %s (%s)", txt.Quote(fs.RelName(destinationMainFilename, imp.originalsPath())), err.Error())
					}
				} else {
					if err := f.Copy(destinationFilename); err != nil {
						log.Errorf("import: failed copying file to %s (%s)", txt.Quote(fs.RelName(destinationMainFilename, imp.originalsPath())), err.Error())
					}
				}
			} else {
				log.Warnf("import: %s", err)

				if opt.RemoveExistingFiles {
					if err := f.Remove(); err != nil {
						log.Errorf("import: failed deleting %s (%s)", txt.Quote(f.BaseName()), err.Error())
					} else {
						log.Infof("import: deleted %s (already exists)", txt.Quote(relativeFilename))
					}
				}
			}
		}

		if destinationMainFilename != "" {
			f, err := NewMediaFile(destinationMainFilename)

			if err != nil {
				log.Errorf("import: %s in %s", err.Error(), txt.Quote(fs.RelName(destinationMainFilename, imp.originalsPath())))
				continue
			}

			if !f.HasJpeg() {
				if jpegFile, err := imp.convert.ToJpeg(f); err != nil {
					log.Errorf("import: %s in %s (convert to jpeg)", err.Error(), txt.Quote(fs.RelName(destinationMainFilename, imp.originalsPath())))
					continue
				} else {
					log.Infof("import: %s created", fs.RelName(jpegFile.FileName(), imp.originalsPath()))
				}
			}

			if jpg, err := f.Jpeg(); err != nil {
				log.Error(err)
			} else {
				if err := jpg.ResampleDefault(imp.thumbPath(), false); err != nil {
					log.Errorf("import: %s in %s (resample)", err.Error(), txt.Quote(jpg.BaseName()))
					continue
				}
			}

			if imp.conf.SidecarJson() && !f.HasJson() {
				if jsonFile, err := imp.convert.ToJson(f); err != nil {
					log.Errorf("import: %s in %s (create json sidecar)", err.Error(), txt.Quote(f.BaseName()))
				} else {
					log.Infof("import: %s created", fs.RelName(jsonFile.FileName(), imp.originalsPath()))
				}
			}

			related, err := f.RelatedFiles(imp.conf.Settings().Index.Sequences)

			if err != nil {
				log.Errorf("import: %s in %s (find related files)", err.Error(), txt.Quote(fs.RelName(destinationMainFilename, imp.originalsPath())))

				continue
			}

			done := make(map[string]bool)
			ind := imp.index
			sizeLimit := ind.conf.OriginalsLimit()

			if related.Main != nil {
				f := related.Main

				// Enforce file size limit for originals.
				if sizeLimit > 0 && f.FileSize() > sizeLimit {
					log.Warnf("import: %s exceeds file size limit (%d / %d MB)", txt.Quote(f.BaseName()), f.FileSize()/(1024*1024), sizeLimit/(1024*1024))
					continue
				}

				res := ind.MediaFile(f, indexOpt, originalName)

				log.Infof("import: %s main %s file %s", res, f.FileType(), txt.Quote(f.RelName(ind.originalsPath())))
				done[f.FileName()] = true

				if res.Success() {
					if err := entity.AddPhotoToAlbums(res.PhotoUID, opt.Albums); err != nil {
						log.Warn(err)
					}
				} else {
					continue
				}
			} else {
				log.Warnf("import: no main file for %s, conversion to jpeg failed?", fs.RelName(destinationMainFilename, imp.originalsPath()))
			}

			for _, f := range related.Files {
				if f == nil {
					continue
				}

				if done[f.FileName()] {
					continue
				}

				done[f.FileName()] = true

				// Enforce file size limit for originals.
				if sizeLimit > 0 && f.FileSize() > sizeLimit {
					log.Warnf("import: %s exceeds file size limit (%d / %d MB)", txt.Quote(f.BaseName()), f.FileSize()/(1024*1024), sizeLimit/(1024*1024))
					continue
				}

				res := ind.MediaFile(f, indexOpt, "")

				log.Infof("import: %s related %s file %s", res, f.FileType(), txt.Quote(f.RelName(ind.originalsPath())))
			}

		}
	}
}
