package photoprism

import (
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/query"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
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

		o := job.IndexOpt
		imp := job.Imp
		impOpt := job.ImportOpt
		impPath := job.ImportOpt.Path
		related := job.Related

		if related.Main == nil {
			log.Warnf("import: %s belongs to no supported media file", clean.Log(fs.RelName(job.FileName, impPath)))
			continue
		}

		// Extract metadata to a JSON file with Exiftool.
		if related.Main.NeedsExifToolJson() {
			if jsonName, err := imp.convert.ToJson(related.Main); err != nil {
				log.Debugf("import: %s in %s (extract metadata)", clean.Log(err.Error()), clean.Log(related.Main.BaseName()))
			} else if err := related.Main.ReadExifToolJson(); err != nil {
				log.Errorf("import: %s in %s (read metadata)", clean.Log(err.Error()), clean.Log(related.Main.BaseName()))
			} else {
				log.Debugf("import: created %s", filepath.Base(jsonName))
			}
		}

		originalName := related.Main.RelName(impPath)

		event.Publish("import.file", event.Data{
			"fileName": originalName,
			"baseName": filepath.Base(related.Main.FileName()),
		})

		for _, f := range related.Files {
			relFileName := f.RelName(impPath)

			if destFileName, err := imp.DestinationFilename(related.Main, f); err == nil {
				destDir := filepath.Dir(destFileName)

				if fs.PathExists(destDir) {
					// Do nothing.
				} else if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
					log.Errorf("import: failed creating folder for %s (%s)", clean.Log(f.BaseName()), err.Error())
				} else {
					destDirRel := fs.RelName(destDir, imp.originalsPath())

					folder := entity.NewFolder(entity.RootOriginals, destDirRel, fs.BirthTime(destDir))

					if err := folder.Create(); err == nil {
						log.Infof("import: created folder /%s", folder.Path)
					}
				}

				if related.Main.HasSameName(f) {
					destMainFileName = destFileName
					log.Infof("import: moving main %s file %s to %s", f.FileType(), clean.Log(relFileName), clean.Log(fs.RelName(destFileName, imp.originalsPath())))
				} else {
					log.Infof("import: moving related %s file %s to %s", f.FileType(), clean.Log(relFileName), clean.Log(fs.RelName(destFileName, imp.originalsPath())))
				}

				if impOpt.Move {
					if err := f.Move(destFileName); err != nil {
						logRelName := clean.Log(fs.RelName(destMainFileName, imp.originalsPath()))
						log.Debugf("import: %s", err.Error())
						log.Warnf("import: failed moving file to %s, is another import running at the same time?", logRelName)
					}
				} else {
					if err := f.Copy(destFileName); err != nil {
						logRelName := clean.Log(fs.RelName(destMainFileName, imp.originalsPath()))
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
				} else if err := entity.AddPhotoToAlbums(file.PhotoUID, impOpt.Albums); err != nil {
					log.Warn(err)
				}

				// Remove duplicates to save storage.
				if impOpt.RemoveExistingFiles {
					if err := f.Remove(); err != nil {
						log.Errorf("import: failed deleting %s (%s)", clean.Log(f.BaseName()), err.Error())
					} else {
						log.Infof("import: deleted %s (already exists)", clean.Log(relFileName))
					}
				}
			}
		}

		if destMainFileName != "" {
			f, err := NewMediaFile(destMainFileName)

			if err != nil {
				log.Errorf("import: %s in %s", err.Error(), clean.Log(fs.RelName(destMainFileName, imp.originalsPath())))
				continue
			}

			// Extract metadata to a JSON file with Exiftool.
			if f.NeedsExifToolJson() {
				if jsonName, err := imp.convert.ToJson(f); err != nil {
					log.Debugf("import: %s in %s (extract metadata)", clean.Log(err.Error()), clean.Log(f.RootRelName()))
				} else {
					log.Debugf("import: created %s", filepath.Base(jsonName))
				}
			}

			// Create JPEG sidecar for media files in other formats so that thumbnails can be created.
			if o.Convert && f.IsMedia() && !f.HasJpeg() {
				if jpegFile, err := imp.convert.ToJpeg(f, false); err != nil {
					log.Errorf("import: %s in %s (convert to jpeg)", err.Error(), clean.Log(f.RootRelName()))
					continue
				} else {
					log.Debugf("import: created %s", clean.Log(jpegFile.BaseName()))
				}
			}

			// Ensure that a JPEG and the configured default thumbnail sizes exist.
			if jpg, err := f.Jpeg(); err != nil {
				log.Error(err)
			} else if exceeds, actual := jpg.ExceedsResolution(o.ResolutionLimit); exceeds {
				log.Errorf("index: %s exceeds resolution limit (%d / %d MP)", clean.Log(f.RootRelName()), actual, o.ResolutionLimit)
				continue
			} else if err := jpg.CreateThumbnails(imp.thumbPath(), false); err != nil {
				log.Errorf("import: failed creating thumbnails for %s (%s)", clean.Log(f.RootRelName()), err.Error())
				continue
			}

			// Find related files.
			related, err := f.RelatedFiles(imp.conf.Settings().StackSequences())

			// Skip import if the finding related files results in an error.
			if err != nil {
				log.Errorf("import: %s in %s (find related files)", err.Error(), clean.Log(fs.RelName(destMainFileName, imp.originalsPath())))
				continue
			}

			done := make(map[string]bool)
			ind := imp.index
			photoUID := ""

			if related.Main != nil {
				f := related.Main

				// Enforce file size and resolution limits.
				if exceeds, actual := f.ExceedsFileSize(o.OriginalsLimit); exceeds {
					log.Warnf("import: %s exceeds file size limit (%d / %d MB)", clean.Log(f.RootRelName()), actual, o.OriginalsLimit)
					continue
				} else if exceeds, actual = f.ExceedsResolution(o.ResolutionLimit); exceeds {
					log.Warnf("import: %s exceeds resolution limit (%d / %d MP)", clean.Log(f.RootRelName()), actual, o.ResolutionLimit)
					continue
				}

				// Index main MediaFile.
				res := ind.MediaFile(f, o, originalName, "")

				// Log result.
				log.Infof("import: %s main %s file %s", res, f.FileType(), clean.Log(f.RootRelName()))
				done[f.FileName()] = true

				if !res.Success() {
					// Skip importing related files if the main file was not indexed successfully.
					continue
				} else if res.PhotoUID != "" {
					photoUID = res.PhotoUID

					// Add photo to album if a list of albums was provided when importing.
					if err := entity.AddPhotoToAlbums(photoUID, impOpt.Albums); err != nil {
						log.Warn(err)
					}
				}
			} else {
				log.Warnf("import: found no main file for %s, conversion to jpeg may have failed", clean.Log(f.RootRelName()))
			}

			for _, f := range related.Files {
				if f == nil {
					continue
				}

				if done[f.FileName()] {
					continue
				}

				done[f.FileName()] = true

				// Show warning if sidecar file exceeds size or resolution limit.
				if exceeds, actual := f.ExceedsFileSize(o.OriginalsLimit); exceeds {
					log.Warnf("import: sidecar file %s exceeds size limit (%d / %d MB)", clean.Log(f.RootRelName()), actual, o.OriginalsLimit)
				} else if exceeds, actual = f.ExceedsResolution(o.ResolutionLimit); exceeds {
					log.Warnf("import: sidecar file %s exceeds resolution limit (%d / %d MP)", clean.Log(f.RootRelName()), actual, o.ResolutionLimit)
				}

				// Extract metadata to a JSON file with Exiftool.
				if f.NeedsExifToolJson() {
					if jsonName, err := imp.convert.ToJson(f); err != nil {
						log.Debugf("import: %s in %s (extract metadata)", clean.Log(err.Error()), clean.Log(f.RootRelName()))
					} else {
						log.Debugf("import: created %s", filepath.Base(jsonName))
					}
				}

				// Index related MediaFile.
				res := ind.MediaFile(f, o, "", photoUID)

				// Save file error.
				if fileUid, err := res.FileError(); err != nil {
					query.SetFileError(fileUid, err.Error())
				}

				// Log result.
				log.Infof("import: %s related %s file %s", res, f.FileType(), clean.Log(f.RootRelName()))
			}

		}
	}
}
