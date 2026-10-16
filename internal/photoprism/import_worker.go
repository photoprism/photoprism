package photoprism

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sync/atomic"
	"syscall"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// ImportJob describes a media import task pulled from the worker queue.
type ImportJob struct {
	FileName  string
	Related   RelatedFiles
	IndexOpt  IndexOptions
	ImportOpt ImportOptions
	Imp       *Import
	Failures  *ImportFailures // Counts files whose content did not reach the originals, if set.
}

// ImportFailures counts the files of an import run whose content could not be moved or copied to
// the originals folder, and whether the storage was full.
type ImportFailures struct {
	files   atomic.Int64
	noSpace atomic.Bool
}

// add counts a file that could not be imported because of err.
func (f *ImportFailures) add(err error) {
	if f == nil {
		return
	}
	f.files.Add(1)
	if errors.Is(err, syscall.ENOSPC) || errors.Is(err, syscall.EDQUOT) {
		f.noSpace.Store(true)
	}
}

// Err returns status.ErrInsufficientStorage if the storage was full, ErrImportIncomplete if other
// files could not be imported, or nil.
func (f *ImportFailures) Err() error {
	if f == nil {
		return nil
	} else if n := f.files.Load(); n == 0 {
		return nil
	} else if f.noSpace.Load() {
		return fmt.Errorf("%w (%d)", status.ErrInsufficientStorage, n)
	} else {
		return fmt.Errorf("%w (%d)", ErrImportIncomplete, n)
	}
}

// importedContent reports whether the file at dest is a regular file with the given hash, so a failed
// move or copy, e.g. of a source that could not be removed afterwards, still placed the content.
func importedContent(dest, hash string) bool {
	return hash != "" && !fs.IsSymlink(dest) && fs.FileExists(dest) && fs.Hash(dest) == hash
}

// storedFilesOf returns how the files of a group relate to the indexed files with the same content, and whether
// all files that are not sidecars, including the main file, are or may be those files themselves. Files of unknown
// types count as sidecars.
func storedFilesOf(related RelatedFiles) (results map[string]storedResult, originals bool) {
	results = make(map[string]storedResult, len(related.Files)+1)
	originals = related.Main != nil

	for _, f := range append(MediaFiles{related.Main}, related.Files...) {
		if f == nil {
			continue
		} else if _, done := results[f.FileName()]; done {
			continue
		}

		result := storedFileOf(f)
		results[f.FileName()] = result

		if !f.IsSidecar() && result.match != storedSame && result.match != storedMaybeSame {
			originals = false
		}
	}

	return results, originals
}

// mainTypeRefused returns why the type of a file keeps it from being imported as a main file, or an empty string.
func mainTypeRefused(f *MediaFile, allowExt fs.ExtList) string {
	switch fileType := f.FileType(); {
	case fileType == fs.TypeUnknown:
		return "due to unknown or unsupported extension"
	case allowExt.Excludes(fileType.DefaultExt()):
		return "because the file type is not allowed"
	default:
		return ""
	}
}

// storedPhotoUIDs returns the sorted, unique picture UIDs of the stored files that files in results are or may be.
func storedPhotoUIDs(results map[string]storedResult) []string {
	uids := make([]string, 0, len(results))

	for _, r := range results {
		if r.file == nil || r.file.PhotoUID == "" || slices.Contains(uids, r.file.PhotoUID) {
			continue
		} else if r.match == storedSame || r.match == storedMaybeSame {
			uids = append(uids, r.file.PhotoUID)
		}
	}

	slices.Sort(uids)

	return uids
}

// ImportWorker consumes ImportJob messages and performs the on-disk moves/copies plus indexing.
func ImportWorker(jobs <-chan ImportJob) {
	for job := range jobs {
		var destMainFileName, destOtherFileName string

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

		if reason := mainTypeRefused(related.Main, allowExt); reason != "" {
			log.Warnf("import: skipped %s %s", clean.Log(originalName), reason)
			continue
		}

		// A group whose files are all originals themselves, e.g. in a folder linked into originals, is not imported,
		// and neither are its sidecars, but its pictures are added to the selected albums.
		results, originals := storedFilesOf(related)

		if originals {
			if results[related.Main.FileName()].match == storedSame {
				log.Infof("import: skipped %s and its related files, as it is the same file as the original", clean.Log(originalName))
			} else {
				log.Infof("import: skipped %s and its related files, as it may be the same file as the original", clean.Log(originalName))
			}

			for _, photoUID := range storedPhotoUIDs(results) {
				if albumErr := entity.AddPhotoToUserAlbums(photoUID, opt.Albums, imp.conf.Settings().Albums.Order.Album, opt.UID); albumErr != nil {
					log.Warnf("import: %s (add picture to albums)", clean.Error(albumErr))
				}
			}

			continue
		}

		event.Publish("import.file", event.Data{
			"fileName":  originalName,
			"baseName":  filepath.Base(related.Main.FileName()),
			"subFolder": opt.DestFolder,
		})

		// Files whose content does not match their extension are moved or copied, so that no file of
		// a stack is left behind, but no external tool reads them.
		mainTypeErr := related.Main.CheckType()

		// Create JSON sidecar file, if needed.
		if mainTypeErr == nil {
			if jsonErr := related.Main.CreateExifToolJson(imp.convert); jsonErr != nil {
				log.Warnf("import: %s", clean.Error(jsonErr))
			}
		}

		// refusedNew reports whether f is a new file other than the main file and its sidecars whose type the main
		// file could not have.
		refusedNew := func(f *MediaFile) bool {
			return f != nil && !f.IsSidecar() && !related.Main.HasSameName(f) &&
				results[f.FileName()].match == storedNone && mainTypeRefused(f, allowExt) != ""
		}

		// A new file is indexed in place of a main file that is already indexed, so new files of a type the main file
		// could not have are not placed then. Whether the main file is held is checked first, if there are any.
		mainChecked := results[related.Main.FileName()].match
		mainHeld := mainChecked == storedSame || mainChecked == storedMaybeSame

		if !mainHeld && slices.ContainsFunc(related.Files, refusedNew) {
			_, mainErr := imp.DestinationFilename(related.Main, related.Main, opt.DestFolder)
			mainHeld = mainErr != nil
		}

		for _, f := range insta360ImportOrder(related) {
			relFileName := f.RelName(src)
			checked := results[f.FileName()].match
			destFileName, err := "", error(nil)

			if mainHeld && refusedNew(f) {
				log.Warnf("import: skipped %s %s", clean.Log(relFileName), mainTypeRefused(f, allowExt))
				continue
			}

			// A file that is or may be an original stays where it is, even if the name of another row is free.
			if checked != storedSame && checked != storedMaybeSame {
				destFileName, err = imp.DestinationFilename(related.Main, f, opt.DestFolder)
			}

			if destFileName != "" && err == nil {
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

					if _, created, createErr := entity.FirstOrCreateFolder(&folder); createErr == nil && created {
						log.Infof("import: created folder /%s", clean.Log(folder.Path))
					}
				}

				if related.Main.HasSameName(f) {
					destMainFileName = destFileName
					log.Infof("import: moving main %s file %s to %s", f.FileType(), clean.Log(relFileName), clean.Log(fs.RelName(destFileName, imp.originalsPath())))
				} else {
					log.Infof("import: moving related %s file %s to %s", f.FileType(), clean.Log(relFileName), clean.Log(fs.RelName(destFileName, imp.originalsPath())))
				}

				logRelName := clean.Log(fs.RelName(destFileName, imp.originalsPath()))
				srcHash := f.Hash()

				var importErr error

				if opt.Move {
					if importErr = f.Move(destFileName, false); importErr != nil {
						log.Error(clean.Error(importErr))
						log.Warnf("import: could not move file to %s", logRelName)
					}
				} else if importErr = f.Copy(destFileName, false); importErr != nil {
					log.Error(clean.Error(importErr))
					log.Warnf("import: could not copy file to %s", logRelName)
				}

				// A file whose content did not reach the originals is counted and, if it is the main
				// file, not indexed, since whatever the destination holds is not this file.
				switch {
				case importErr == nil:
				case importedContent(destFileName, srcHash):
					log.Warnf("import: %s already holds the content, the staged file remains", logRelName)
				default:
					job.Failures.add(importErr)

					if destMainFileName == destFileName {
						destMainFileName = ""
					}

					continue
				}

				if destOtherFileName == "" && !f.IsSidecar() && !related.Main.HasSameName(f) && mainTypeRefused(f, allowExt) == "" {
					destOtherFileName = destFileName
				}
			} else {
				if err != nil {
					log.Infof("import: %s", err)
				}

				if related.Main.HasSameName(f) {
					mainHeld = true
				}

				// What follows is for a file the library already holds. The destination name is not
				// evidence of that, so it is established once here and every step below reads it.
				result, found := results[f.FileName()]

				if !found {
					result = storedFileOf(f)
				}

				stored, match := result.file, result.match

				// Add the picture of a copy or of a kept original to the selected albums.
				if stored != nil {
					if albumErr := entity.AddPhotoToUserAlbums(stored.PhotoUID, opt.Albums, imp.conf.Settings().Albums.Order.Album, opt.UID); albumErr != nil {
						log.Warnf("import: %s (add duplicate to albums)", clean.Error(albumErr))
					}
				}

				// Remember the original filename for duplicates so that indexing can still persist
				// OriginalName even when the file was not copied due to an existing identical file.
				if match == storedCopy {
					if storedName := FileName(stored.FileRoot, stored.FileName); storedName != "" {
						relatedOriginalNames[storedName] = relFileName
					}
				}

				// Remove duplicates to save storage, and only where the library demonstrably holds
				// the same content, also after the stored file was compared.
				switch {
				case !opt.RemoveExistingFiles:
				case match == storedCopy && !result.unchanged():
					log.Infof("import: kept %s, as the stored file has changed", clean.Log(relFileName))
				case match == storedCopy:
					if removeErr := f.Remove(); removeErr != nil {
						log.Errorf("import: failed to delete %s (%s)", clean.Log(f.BaseName()), removeErr.Error())
					} else {
						log.Infof("import: deleted %s (already exists)", clean.Log(relFileName))
					}
				case match == storedSame:
					log.Infof("import: kept %s, as it is the same file as the original", clean.Log(relFileName))
				case match == storedMaybeSame:
					log.Infof("import: kept %s, as it may be the same file as the original", clean.Log(relFileName))
				default:
					// Saying nothing here reads as a failure to clean up, since the line above
					// reports a match. The source is kept because the library does not demonstrably
					// hold the same content, which a stale index is enough to cause.
					log.Infof("import: kept %s, as no stored file holds its contents", clean.Log(relFileName))
				}
			}
		}

		// The first other file placed in originals is indexed with its related files if the main file was not
		// placed because it is already indexed, e.g. a new RAW file next to a hard link to an original.
		if mainHeld && destMainFileName == "" {
			destMainFileName = destOtherFileName
		}

		if destMainFileName != "" {
			f, err := NewMediaFile(destMainFileName)

			if err != nil {
				log.Errorf("import: %s in %s", err.Error(), clean.Log(fs.RelName(destMainFileName, imp.originalsPath())))
				continue
			} else if typeErr := f.CheckType(); typeErr != nil {
				log.Warnf("import: %s %s and was not indexed", clean.Log(f.RootRelName()), typeErr)
				continue
			}

			warnInsta360LensNotVideo("import", f)

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
				log.Errorf("import: %s for %s (create preview image)", clean.Error(imgErr), clean.Log(f.RootRelName()))
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

			// Other files in originals whose names are not portable are not indexed with the imported ones. The
			// imported file keeps the name its destination pattern gave it.
			if clean.PortableName(filepath.Base(f.FileName())) {
				if related = PortableRelated(related, f, func(rf *MediaFile) {
					warnUnportable("index", entity.RootOriginals, rf.RootRelName(), "index it")
				}); related.Main == nil {
					continue
				}
			}

			// Ensure every convertible file in the stack has its own preview image, mirroring
			// IndexRelated. Otherwise a stack with more than one such file (e.g. a base .svg plus
			// a .touch.svg variant) gets only a single preview, leaving the file that becomes the
			// indexed primary without a matching sidecar so its photo stays hidden until a rescan.
			if o.Convert {
				for _, rf := range related.Files {
					if rf == nil || !rf.IsMedia() || rf.HasPreviewImage() || rf.CheckType() != nil {
						continue
					} else if insta360ImportedMember(originalName, relatedOriginalNames[rf.FileName()]) {
						// The combined preview of an imported capture is made from its left lens.
						continue
					}

					// The preview gets the file orientation, which may only be readable with ExifTool.
					if jsonErr := rf.CreateExifToolJson(imp.convert); jsonErr != nil {
						log.Warnf("import: %s", clean.Error(jsonErr))
					}

					if img, imgErr := imp.convert.ToImage(rf, false); imgErr != nil {
						log.Warnf("import: could not create preview image for %s (%s)", clean.Log(rf.RootRelName()), clean.Error(imgErr))
					} else if img != nil {
						if thumbsErr := img.GenerateThumbnails(imp.thumbPath(), false); thumbsErr != nil {
							log.Warnf("import: failed to generate thumbnails for %s (%s)", clean.Log(img.RootRelName()), thumbsErr.Error())
						}

						img.SetRelatedMain(related.Main)
						related.Files = append(related.Files, img)
					}
				}
			}

			done := make(map[string]bool)
			ind := imp.index
			photoUID := ""
			archived := false

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
				main.SetRelatedMain(main)
				mainOriginalName := originalName

				if mainHeld {
					mainOriginalName = relatedOriginalNames[main.FileName()]
				}

				res := ind.UserMediaFile(main, o, mainOriginalName, "", opt.UID)

				// Log result.
				log.Infof("import: %s main %s file %s", res, main.FileType(), clean.Log(main.RootRelName()))
				done[main.FileName()] = true

				// Related files of a new main file are added even if its photo is archived, as it was restored from a backup.
				if res.Status == IndexAdded {
					o.SkipArchived = false
				}

				switch {
				case !res.Success():
					// Skip importing related files if the main file was not indexed successfully.
					continue
				case res.Archived():
					// Related files of an archived photo are found by its UID, and skipped like the main file.
					photoUID = res.PhotoUID
					archived = true
				case res.PhotoUID != "":
					photoUID = res.PhotoUID

					// Add photo to album if a list of albums was provided when importing.
					if albumErr := entity.AddPhotoToUserAlbums(photoUID, opt.Albums, imp.conf.Settings().Albums.Order.Album, opt.UID); albumErr != nil {
						log.Warnf("import: %s (add picture to albums)", clean.Error(albumErr))
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

				// Skip related files whose content does not match their extension.
				if typeErr := file.CheckType(); typeErr != nil {
					log.Warnf("import: %s %s and was not indexed", clean.Log(file.RootRelName()), typeErr)
					continue
				}

				// Show warning if sidecar file exceeds size or resolution limit.
				if _, limitErr := file.ExceedsBytes(o.ByteLimit); limitErr != nil {
					log.Warnf("import: %s", limitErr)
				} else if _, limitErr = file.ExceedsResolution(o.ResolutionLimit); limitErr != nil {
					log.Warnf("import: %s", limitErr)
				}

				warnInsta360LensNotVideo("import", file)

				// Extract metadata to a JSON file with Exiftool and add it to the cached metadata, which
				// the resolution check above may already have read.
				if jsonErr := file.CreateExifToolJson(imp.convert); jsonErr != nil {
					log.Warnf("import: %s", clean.Error(jsonErr))
				}

				// Index related media file including its original filename.
				file.SetRelatedMain(related.Main)
				res := ind.UserMediaFile(file, o, relatedOriginalNames[file.FileName()], photoUID, opt.UID)

				// Save file error.
				if fileUid, fileErr := res.FileError(); fileErr != nil {
					query.SetFileError(fileUid, clean.Error(fileErr))
				}

				// Log result.
				log.Infof("import: %s related %s file %s", res, file.FileType(), clean.Log(file.RootRelName()))
			}

			// Renamed capture files are only recognized by their original names once they are indexed, so the
			// preview of the left lens is then made again from both lenses.
			if o.Convert && photoUID != "" && !archived && related.Main != nil {
				imp.updateInsta360Preview(related.Main, o, photoUID, opt.UID)
			}
		}
	}
}

// updateInsta360Preview replaces the preview of an imported capture's left lens with one made from both lenses.
func (imp *Import) updateInsta360Preview(main *MediaFile, o IndexOptions, photoUID, userUID string) {
	if capture := FindInsta360Capture(main); !capture.ValidPair() || capture.Left.FileName() != main.FileName() {
		return
	} else if !imp.conf.FFmpegEnabled() || !imp.convert.FFmpegAllowed(main) {
		return
	}

	img, err := imp.convert.ToImage(main, true)

	if err != nil || img == nil {
		log.Warnf("import: could not create equirectangular preview for %s (%s)", clean.Log(main.RootRelName()), clean.Error(err))
		return
	} else if thumbsErr := img.GenerateThumbnails(imp.thumbPath(), false); thumbsErr != nil {
		log.Warnf("import: failed to generate thumbnails for %s (%s)", clean.Log(img.RootRelName()), thumbsErr.Error())
	}

	img.SetRelatedMain(main)
	res := imp.index.UserMediaFile(img, o, "", photoUID, userUID)
	log.Infof("import: %s related %s file %s", res, img.FileType(), clean.Log(img.RootRelName()))
}
