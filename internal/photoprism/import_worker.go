package photoprism

import (
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/query"

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
		var destMainFileName string
		related := job.Related
		imp := job.Imp
		opt := job.ImportOpt
		indexOpt := job.IndexOpt
		importPath := job.ImportOpt.Path

		if related.Main == nil {
			log.Warnf("import: %s belongs to no supported media file", txt.Quote(fs.RelName(job.FileName, importPath)))
			continue
		}

		if related.Main.NeedsExifToolJson() {
			if jsonName, err := imp.convert.ToJson(related.Main); err != nil {
				log.Debugf("import: %s in %s (extract metadata)", txt.Quote(err.Error()), txt.Quote(related.Main.BaseName()))
			} else if err := related.Main.ReadExifToolJson(); err != nil {
				log.Errorf("import: %s in %s (read metadata)", txt.Quote(err.Error()), txt.Quote(related.Main.BaseName()))
			} else {
				log.Debugf("import: created %s", filepath.Base(jsonName))
			}
		}

		originalName := related.Main.RelName(importPath)

		event.Publish("import.file", event.Data{
			"fileName": originalName,
			"baseName": filepath.Base(related.Main.FileName()),
		})

		for _, f := range related.Files {
			relFileName := f.RelName(importPath)

			if destFileName, err := imp.DestinationFilename(related.Main, f); err == nil {
				destDir := filepath.Dir(destFileName)

				if fs.PathExists(destDir) {
					// Do nothing.
				} else if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
					log.Errorf("import: failed creating folder for %s (%s)", txt.Quote(f.BaseName()), err.Error())
				} else {
					destDirRel := fs.RelName(destDir, imp.originalsPath())

					folder := entity.NewFolder(entity.RootOriginals, destDirRel, fs.BirthTime(destDir))

					if err := folder.Create(); err == nil {
						log.Infof("import: created folder /%s", folder.Path)
					}
				}

				if related.Main.HasSameName(f) {
					destMainFileName = destFileName
					log.Infof("import: moving main %s file %s to %s", f.FileType(), txt.Quote(relFileName), txt.Quote(fs.RelName(destFileName, imp.originalsPath())))
				} else {
					log.Infof("import: moving related %s file %s to %s", f.FileType(), txt.Quote(relFileName), txt.Quote(fs.RelName(destFileName, imp.originalsPath())))
				}

				if opt.Move {
					if err := f.Move(destFileName); err != nil {
						logRelName := txt.Quote(fs.RelName(destMainFileName, imp.originalsPath()))
						log.Debugf("import: %s", err.Error())
						log.Warnf("import: failed moving file to %s, is another import running at the same time?", logRelName)
					}
				} else {
					if err := f.Copy(destFileName); err != nil {
						logRelName := txt.Quote(fs.RelName(destMainFileName, imp.originalsPath()))
						log.Debugf("import: %s", err.Error())
						log.Warnf("import: failed copying file to %s, is another import running at the same time?", logRelName)
					}
				}
			} else {
				log.Infof("import: %s", err)

				// Try to add duplicates to selected album(s) as well, see #991.
				if fileHash := f.Hash(); fileHash == "" {
					// Do nothing.
				} else if file, err := entity.FirstFileByHash(fileHash); err != nil {
					// Do nothing.
				} else if err := entity.AddPhotoToAlbums(file.PhotoUID, opt.Albums); err != nil {
					log.Warn(err)
				}

				// Remove duplicates to save storage.
				if opt.RemoveExistingFiles {
					if err := f.Remove(); err != nil {
						log.Errorf("import: failed deleting %s (%s)", txt.Quote(f.BaseName()), err.Error())
					} else {
						log.Infof("import: deleted %s (already exists)", txt.Quote(relFileName))
					}
				}
			}
		}

		if destMainFileName != "" {
			f, err := NewMediaFile(destMainFileName)

			if err != nil {
				log.Errorf("import: %s in %s", err.Error(), txt.Quote(fs.RelName(destMainFileName, imp.originalsPath())))
				continue
			}

			if f.NeedsExifToolJson() {
				if jsonName, err := imp.convert.ToJson(f); err != nil {
					log.Debugf("import: %s in %s (extract metadata)", txt.Quote(err.Error()), txt.Quote(f.BaseName()))
				} else {
					log.Debugf("import: created %s", filepath.Base(jsonName))
				}
			}

			if indexOpt.Convert && f.IsMedia() && !f.HasJpeg() {
				if jpegFile, err := imp.convert.ToJpeg(f); err != nil {
					log.Errorf("import: %s in %s (convert to jpeg)", err.Error(), txt.Quote(fs.RelName(destMainFileName, imp.originalsPath())))
					continue
				} else {
					log.Debugf("import: created %s", txt.Quote(jpegFile.BaseName()))
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

			related, err := f.RelatedFiles(imp.conf.Settings().StackSequences())

			if err != nil {
				log.Errorf("import: %s in %s (find related files)", err.Error(), txt.Quote(fs.RelName(destMainFileName, imp.originalsPath())))

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
				log.Warnf("import: found no main file for %s, conversion to jpeg may have failed", fs.RelName(destMainFileName, imp.originalsPath()))
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

				if f.NeedsExifToolJson() {
					if jsonName, err := imp.convert.ToJson(f); err != nil {
						log.Debugf("import: %s in %s (extract metadata)", txt.Quote(err.Error()), txt.Quote(f.BaseName()))
					} else {
						log.Debugf("import: created %s", filepath.Base(jsonName))
					}
				}

				res := ind.MediaFile(f, indexOpt, "")

				if res.Indexed() && f.IsJpeg() {
					if err := f.ResampleDefault(ind.thumbPath(), false); err != nil {
						log.Errorf("import: failed creating thumbnails for %s (%s)", txt.Quote(f.BaseName()), err.Error())
						query.SetFileError(res.FileUID, err.Error())
					}
				}

				log.Infof("import: %s related %s file %s", res, f.FileType(), txt.Quote(f.RelName(ind.originalsPath())))
			}

		}
	}
}
