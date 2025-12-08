package photoprism

import (
	"path/filepath"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// ImportJob describes a media import task pulled from the worker queue.
type ImportJob struct {
	FileName  string
	Related   RelatedFiles
	IndexOpt  IndexOptions
	ImportOpt ImportOptions
	Imp       *Import
}

// ImportWorker consumes ImportJob messages and performs the on-disk moves/copies plus indexing.
func ImportWorker(jobs <-chan ImportJob) {
	for job := range jobs {
		var destMainFileName string

		o := job.IndexOpt

		imp := job.Imp
		opt := job.ImportOpt
		src := job.ImportOpt.Path
		allowExt := job.Imp.AllowExt

		related := job.Related

		// relatedOriginalNames contains the original filenames of related files.
		relatedOriginalNames := make(map[string]string, len(related.Files))

		if related.Main == nil {
			log.Errorf("import: %s does not belong to a supported media file", clean.Log(fs.RelName(job.FileName, src)))
			continue
		}

		originalName := related.Main.RelName(src)
		mainFileType := related.Main.FileType()

		if mainFileType == fs.TypeUnknown {
			log.Warnf("import: skipped %s due to unknown or unsupported extension", clean.Log(originalName))
			continue
		} else if allowExt.Excludes(mainFileType.DefaultExt()) {
			log.Warnf("import: skipped %s because the file type is not allowed", clean.Log(originalName))
			continue
		}

		event.Publish("import.file", event.Data{
			"fileName":  originalName,
			"baseName":  filepath.Base(related.Main.FileName()),
			"subFolder": opt.DestFolder,
		})

		// Create JSON sidecar file, if needed.
		if jsonErr := related.Main.CreateExifToolJson(imp.convert); jsonErr != nil {
			log.Warnf("import: %s", clean.Error(jsonErr))
		}

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

					folder := entity.NewFolder(entity.RootOriginals, destDirRel, fs.ModTime(destDir))

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
					if moveErr := f.Move(destFileName, false); moveErr != nil {
						logRelName := clean.Log(fs.RelName(destMainFileName, imp.originalsPath()))
						log.Error(moveErr)
						log.Warnf("import: could not move file to %s, is another import running?", logRelName)
					}
				} else {
					if copyErr := f.Copy(destFileName, false); copyErr != nil {
						logRelName := clean.Log(fs.RelName(destMainFileName, imp.originalsPath()))
						log.Error(copyErr)
						log.Warnf("import: could not copy file to %s, is another import running?", logRelName)
					}
				}
			} else {
				log.Infof("import: %s", err)

				// Try to add duplicates to selected album(s) as well, see #991.
				if fileHash := f.Hash(); fileHash == "" {
					// Do nothing.
				} else if file, fileErr := entity.FirstFileByHash(fileHash); fileErr != nil {
					// Do nothing.
				} else if albumErr := entity.AddPhotoToUserAlbums(file.PhotoUID, opt.Albums, imp.conf.Settings().Albums.Order.Album, opt.UID); albumErr != nil {
					log.Warn(albumErr)
				}

				// Remember the original filename for duplicates so that indexing can still persist
				// OriginalName even when the file was not copied due to an existing identical file.
				if fileHash := f.Hash(); fileHash != "" {
					if existing, findErr := entity.FirstFileByHash(fileHash); findErr == nil {
						existingPath := FileName(existing.FileRoot, existing.FileName)
						if existingPath != "" {
							relatedOriginalNames[existingPath] = relFileName
						}
					}
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
				if img, imgErr := imp.convert.ToImage(f, false); imgErr != nil {
					log.Errorf("import: failed to create preview image for %s (%s)", clean.Log(f.RootRelName()), clean.Error(imgErr))
					continue
				} else if img == nil {
					log.Debugf("import: skipped creating preview image for %s", clean.Log(f.RootRelName()))
				} else {
					log.Debugf("import: created %s", clean.Log(img.BaseName()))
				}
			}

			// Ensure that a JPEG and the configured default thumbnail sizes exist.
			if img, imgErr := f.PreviewImage(); imgErr != nil {
				log.Error(imgErr)
			} else if _, limitErr := img.ExceedsResolution(o.ResolutionLimit); limitErr != nil {
				log.Errorf("import: %s", limitErr)
				continue
			} else if thumbsErr := img.GenerateThumbnails(imp.thumbPath(), false); thumbsErr != nil {
				log.Errorf("import: failed to generate thumbnails for %s (%s)", clean.Log(f.RootRelName()), clean.Error(thumbsErr))
				continue
			}

			// Find and index related originals.
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
				main := related.Main

				// Enforce file size and resolution limits.
				if _, limitErr := main.ExceedsBytes(o.ByteLimit); limitErr != nil {
					log.Warnf("import: %s", limitErr)
					continue
				} else if _, limitErr = main.ExceedsResolution(o.ResolutionLimit); limitErr != nil {
					log.Warnf("import: %s", limitErr)
					continue
				}

				// Index main MediaFile.
				res := ind.UserMediaFile(main, o, originalName, "", opt.UID)

				// Log result.
				log.Infof("import: %s main %s file %s", res, main.FileType(), clean.Log(main.RootRelName()))
				done[main.FileName()] = true

				if !res.Success() {
					// Skip importing related files if the main file was not indexed successfully.
					continue
				} else if res.PhotoUID != "" {
					photoUID = res.PhotoUID

					// Add photo to album if a list of albums was provided when importing.
					if albumErr := entity.AddPhotoToUserAlbums(photoUID, opt.Albums, imp.conf.Settings().Albums.Order.Album, opt.UID); albumErr != nil {
						log.Warn(albumErr)
					}
				}
			} else {
				log.Warnf("import: no main media file found for %s, creation of a preview image may have failed", clean.Log(f.RootRelName()))
			}

			for _, file := range related.Files {
				if file == nil {
					continue
				}

				if done[file.FileName()] {
					continue
				}

				done[file.FileName()] = true

				// Show warning if sidecar file exceeds size or resolution limit.
				if _, limitErr := file.ExceedsBytes(o.ByteLimit); limitErr != nil {
					log.Warnf("import: %s", limitErr)
				} else if _, limitErr = file.ExceedsResolution(o.ResolutionLimit); limitErr != nil {
					log.Warnf("import: %s", limitErr)
				}

				// Extract metadata to a JSON file with Exiftool.
				if file.NeedsExifToolJson() {
					if jsonName, err := imp.convert.ToJson(file, false); err != nil {
						log.Tracef("exiftool: %s", clean.Error(err))
						log.Debugf("exiftool: failed parsing %s", clean.Log(file.RootRelName()))
					} else {
						log.Debugf("import: created %s", filepath.Base(jsonName))
					}
				}

				// Index related media file including its original filename.
				res := ind.UserMediaFile(file, o, relatedOriginalNames[file.FileName()], photoUID, opt.UID)

				// Save file error.
				if fileUid, fileErr := res.FileError(); fileErr != nil {
					query.SetFileError(fileUid, clean.Error(fileErr))
				}

				// Log result.
				log.Infof("import: %s related %s file %s", res, file.FileType(), clean.Log(file.RootRelName()))
			}
		}
	}
}
