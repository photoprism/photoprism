package photoprism

import (
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
		opt := job.ImportOpt
		src := job.ImportOpt.Path

		related := job.Related

		// relatedOriginalNames contains the original filenames of related files.
		relatedOriginalNames := make(map[string]string, len(related.Files))

		if related.Main == nil {
			log.Errorf("import: %s does not belong to a supported media file", clean.Log(fs.RelName(job.FileName, src)))
			continue
		}

		// Create JSON sidecar file, if needed.
		if jsonErr := related.Main.CreateExifToolJson(imp.convert); jsonErr != nil {
			log.Warnf("import: %s", clean.Error(jsonErr))
		}

		originalName := related.Main.RelName(src)

		event.Publish("import.file", event.Data{
			"fileName":  originalName,
			"baseName":  filepath.Base(related.Main.FileName()),
			"subFolder": opt.DestFolder,
		})

		for _, f := range related.Files {
			relFileName := f.RelName(src)

			if destFileName, err := imp.DestinationFilename(related.Main, f, opt.DestFolder); err == nil {
				destDir := filepath.Dir(destFileName)

				// Remember the original filenames of related files, so they can later be indexed and searched.
				relatedOriginalNames[destFileName] = relFileName

				if fs.PathExists(destDir) {
					// Do nothing.
				} else if mkdirErr := fs.MkdirAll(destDir); mkdirErr != nil {
					log.Errorf("import: failed to create folder for %s (%s)", clean.Log(f.BaseName()), mkdirErr.Error())
				} else {
					destDirRel := fs.RelName(destDir, imp.originalsPath())

					folder := entity.NewFolder(entity.RootOriginals, destDirRel, fs.BirthTime(destDir))

					if createErr := folder.Create(); createErr == nil {
						log.Infof("import: created folder /%s", folder.Path)
					}
				}

				if related.Main.HasSameName(f) {
					destMainFileName = destFileName
					log.Infof("import: moving main %s file %s to %s", f.FileType(), clean.Log(relFileName), clean.Log(fs.RelName(destFileName, imp.originalsPath())))
				} else {
					log.Infof("import: moving related %s file %s to %s", f.FileType(), clean.Log(relFileName), clean.Log(fs.RelName(destFileName, imp.originalsPath())))
				}

				if opt.Move {
					if moveErr := f.Move(destFileName); moveErr != nil {
						logRelName := clean.Log(fs.RelName(destMainFileName, imp.originalsPath()))
						log.Debugf("import: %s", clean.Error(moveErr))
						log.Warnf("import: failed moving file to %s, is another import running at the same time?", logRelName)
					}
				} else {
					if copyErr := f.Copy(destFileName); copyErr != nil {
						logRelName := clean.Log(fs.RelName(destMainFileName, imp.originalsPath()))
						log.Debugf("import: %s", clean.Error(copyErr))
						log.Warnf("import: failed copying file to %s, is another import running at the same time?", logRelName)
					}
				}
			} else {
				log.Infof("import: %s", err)

				// Try to add duplicates to selected album(s) as well, see #991.
				if fileHash := f.Hash(); fileHash == "" {
					// Do nothing.
				} else if file, fileErr := entity.FirstFileByHash(fileHash); fileErr != nil {
					// Do nothing.
				} else if albumErr := entity.AddPhotoToUserAlbums(file.PhotoUID, opt.Albums, opt.UID); albumErr != nil {
					log.Warn(albumErr)
				}

				// Remove duplicates to save storage.
				if opt.RemoveExistingFiles {
					if removeErr := f.Remove(); removeErr != nil {
						log.Errorf("import: failed to delete %s (%s)", clean.Log(f.BaseName()), removeErr.Error())
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

			// Create JSON sidecar file, if needed.
			if jsonErr := f.CreateExifToolJson(imp.convert); jsonErr != nil {
				log.Warnf("import: %s", clean.Error(jsonErr))
			}

			// Create JPEG sidecar for media files in other formats so that thumbnails can be created.
			if o.Convert && f.IsMedia() && !f.HasPreviewImage() {
				if jpegFile, err := imp.convert.ToImage(f, false); err != nil {
					log.Errorf("import: %s in %s (convert to jpeg)", clean.Error(err), clean.Log(f.RootRelName()))
					continue
				} else {
					log.Debugf("import: created %s", clean.Log(jpegFile.BaseName()))
				}
			}

			// Ensure that a JPEG and the configured default thumbnail sizes exist.
			if jpg, convertErr := f.PreviewImage(); convertErr != nil {
				log.Error(convertErr)
			} else if limitErr, _ := jpg.ExceedsResolution(o.ResolutionLimit); limitErr != nil {
				log.Errorf("index: %s", limitErr)
				continue
			} else if thumbsErr := jpg.CreateThumbnails(imp.thumbPath(), false); thumbsErr != nil {
				log.Errorf("import: failed to create thumbnails for %s (%s)", clean.Log(f.RootRelName()), clean.Error(thumbsErr))
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
				if limitErr, _ := f.ExceedsBytes(o.ByteLimit); limitErr != nil {
					log.Warnf("import: %s", limitErr)
					continue
				} else if limitErr, _ = f.ExceedsResolution(o.ResolutionLimit); limitErr != nil {
					log.Warnf("import: %s", limitErr)
					continue
				}

				// Index main MediaFile.
				res := ind.UserMediaFile(f, o, originalName, "", opt.UID)

				// Log result.
				log.Infof("import: %s main %s file %s", res, f.FileType(), clean.Log(f.RootRelName()))
				done[f.FileName()] = true

				if !res.Success() {
					// Skip importing related files if the main file was not indexed successfully.
					continue
				} else if res.PhotoUID != "" {
					photoUID = res.PhotoUID

					// Add photo to album if a list of albums was provided when importing.
					if albumErr := entity.AddPhotoToUserAlbums(photoUID, opt.Albums, opt.UID); albumErr != nil {
						log.Warn(albumErr)
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
				if limitErr, _ := f.ExceedsBytes(o.ByteLimit); limitErr != nil {
					log.Warnf("import: %s", limitErr)
				} else if limitErr, _ = f.ExceedsResolution(o.ResolutionLimit); limitErr != nil {
					log.Warnf("import: %s", limitErr)
				}

				// Extract metadata to a JSON file with Exiftool.
				if f.NeedsExifToolJson() {
					if jsonName, err := imp.convert.ToJson(f, false); err != nil {
						log.Tracef("exiftool: %s", clean.Error(err))
						log.Debugf("exiftool: failed parsing %s", clean.Log(f.RootRelName()))
					} else {
						log.Debugf("import: created %s", filepath.Base(jsonName))
					}
				}

				// Index related media file including its original filename.
				res := ind.UserMediaFile(f, o, relatedOriginalNames[f.FileName()], photoUID, opt.UID)

				// Save file error.
				if fileUid, fileErr := res.FileError(); fileErr != nil {
					query.SetFileError(fileUid, clean.Error(fileErr))
				}

				// Log result.
				log.Infof("import: %s related %s file %s", res, f.FileType(), clean.Log(f.RootRelName()))
			}
		}
	}
}
