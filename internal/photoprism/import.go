package photoprism

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"sync"

	"github.com/photoprism/photoprism/pkg/media"

	"github.com/karrick/godirwalk"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Import represents an importer that can copy/move MediaFiles to the originals directory.
type Import struct {
	conf    *config.Config
	index   *Index
	convert *Convert
}

// NewImport returns a new importer and expects its dependencies as arguments.
func NewImport(conf *config.Config, index *Index, convert *Convert) *Import {
	instance := &Import{
		conf:    conf,
		index:   index,
		convert: convert,
	}

	return instance
}

// originalsPath returns the original media files path as string.
func (imp *Import) originalsPath() string {
	return imp.conf.OriginalsPath()
}

// thumbPath returns the thumbnails cache path as string.
func (imp *Import) thumbPath() string {
	return imp.conf.ThumbCachePath()
}

// Start imports media files from a directory and converts/indexes them as needed.
func (imp *Import) Start(opt ImportOptions) fs.Done {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("import: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	var directories []string
	done := make(fs.Done)

	if imp.conf == nil {
		log.Errorf("import: config is nil")
		return done
	}

	ind := imp.index
	importPath := opt.Path

	if !fs.PathExists(importPath) {
		event.Error(fmt.Sprintf("import: %s does not exist", importPath))
		return done
	}

	if err := mutex.MainWorker.Start(); err != nil {
		event.Error(fmt.Sprintf("import: %s", err.Error()))
		return done
	}

	defer mutex.MainWorker.Stop()

	if err := ind.tensorFlow.Init(); err != nil {
		log.Errorf("import: %s", err.Error())
		return done
	}

	jobs := make(chan ImportJob)

	// Start a fixed number of goroutines to import files.
	var wg sync.WaitGroup
	var numWorkers = imp.conf.Workers()
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			ImportWorker(jobs)
			wg.Done()
		}()
	}

	filesImported := 0

	settings := imp.conf.Settings()
	convert := settings.Index.Convert && imp.conf.SidecarWritable()
	indexOpt := NewIndexOptions("/", true, convert, true, false, false)
	skipRaw := imp.conf.DisableRaw()
	ignore := fs.NewIgnoreList(fs.IgnoreFile, true, false)

	if err := ignore.Dir(importPath); err != nil {
		log.Infof("import: %s", err)
	}

	ignore.Log = func(fileName string) {
		log.Infof(`import: ignored "%s"`, fs.RelName(fileName, importPath))
	}

	err := godirwalk.Walk(importPath, &godirwalk.Options{
		ErrorCallback: func(fileName string, err error) godirwalk.ErrorAction {
			return godirwalk.SkipNode
		},
		Callback: func(fileName string, info *godirwalk.Dirent) error {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("import: %s (panic)\nstack: %s", r, debug.Stack())
				}
			}()

			if mutex.MainWorker.Canceled() {
				return errors.New("canceled")
			}

			isDir := info.IsDir()
			isSymlink := info.IsSymlink()

			if skip, result := fs.SkipWalk(fileName, isDir, isSymlink, done, ignore); skip {
				if isDir && result != filepath.SkipDir {
					if fileName != importPath {
						directories = append(directories, fileName)
					}

					folder := entity.NewFolder(entity.RootImport, fs.RelName(fileName, imp.conf.ImportPath()), fs.BirthTime(fileName))

					if err := folder.Create(); err == nil {
						log.Infof("import: added folder /%s", folder.Path)
					}
				}

				return result
			}

			done[fileName] = fs.Found

			if !media.MainFile(fileName) {
				return nil
			}

			mf, err := NewMediaFile(fileName)

			// Check if file exists and is not empty.
			if err != nil {
				log.Warnf("import: %s", err)
				return nil
			}

			// Ignore RAW images?
			if mf.IsRaw() && skipRaw {
				log.Infof("import: skipped raw %s", clean.Log(mf.RootRelName()))
				return nil
			}

			// Find related files to import.
			related, err := mf.RelatedFiles(imp.conf.Settings().StackSequences())

			if err != nil {
				event.Error(fmt.Sprintf("import: %s", err.Error()))
				return nil
			}

			var files MediaFiles

			for _, f := range related.Files {
				if f.FileSize() == 0 || done[f.FileName()].Processed() {
					continue
				}

				files = append(files, f)
				filesImported++
				done[f.FileName()] = fs.Processed
			}

			done[fileName] = fs.Processed

			related.Files = files

			jobs <- ImportJob{
				FileName:  fileName,
				Related:   related,
				IndexOpt:  indexOpt,
				ImportOpt: opt,
				Imp:       imp,
			}

			return nil
		},
		Unsorted:            false,
		FollowSymbolicLinks: true,
	})

	close(jobs)
	wg.Wait()

	sort.Slice(directories, func(i, j int) bool {
		return len(directories[i]) > len(directories[j])
	})

	if opt.RemoveEmptyDirectories {
		// Remove empty directories from import path.
		for _, directory := range directories {
			if fs.IsEmpty(directory) {
				if err := os.Remove(directory); err != nil {
					log.Errorf("import: failed deleting empty folder %s (%s)", clean.Log(fs.RelName(directory, importPath)), err)
				} else {
					log.Infof("import: deleted empty folder %s", clean.Log(fs.RelName(directory, importPath)))
				}
			}
		}
	}

	if opt.RemoveDotFiles {
		// Remove hidden .files if option is enabled.
		for _, file := range ignore.Hidden() {
			if !fs.FileExists(file) {
				continue
			}

			if err := os.Remove(file); err != nil {
				log.Errorf("import: failed removing %s (%s)", clean.Log(fs.RelName(file, importPath)), err.Error())
			}
		}
	}

	if err != nil {
		log.Error(err.Error())
	}

	if filesImported > 0 {
		// Run facial recognition if enabled.
		if w := NewFaces(imp.conf); w.Disabled() {
			log.Debugf("import: skipping facial recognition")
		} else if err := w.Start(FacesOptionsDefault()); err != nil {
			log.Errorf("import: %s", err)
		}

		// Update photo counts and visibilities.
		if err := entity.UpdateCounts(); err != nil {
			log.Warnf("index: %s (update counts)", err)
		}
	}

	runtime.GC()

	return done
}

// Cancel stops the current import operation.
func (imp *Import) Cancel() {
	mutex.MainWorker.Cancel()
}

// DestinationFilename returns the destination filename of a MediaFile to be imported.
func (imp *Import) DestinationFilename(mainFile *MediaFile, mediaFile *MediaFile) (string, error) {
	fileName := mainFile.CanonicalName()
	fileExtension := mediaFile.Extension()
	dateCreated := mainFile.DateCreated()

	if !mediaFile.IsSidecar() {
		if f, err := entity.FirstFileByHash(mediaFile.Hash()); err == nil {
			existingFilename := FileName(f.FileRoot, f.FileName)
			if fs.FileExists(existingFilename) {
				return existingFilename, fmt.Errorf("%s is identical to %s (sha1 %s)", clean.Log(filepath.Base(mediaFile.FileName())), clean.Log(f.FileName), mediaFile.Hash())
			} else {
				return existingFilename, nil
			}
		}
	}

	//	Mon Jan 2 15:04:05 -0700 MST 2006
	pathName := filepath.Join(imp.originalsPath(), dateCreated.Format("2006/01"))

	iteration := 0

	result := filepath.Join(pathName, fileName+fileExtension)

	for fs.FileExists(result) {
		if mediaFile.Hash() == fs.Hash(result) {
			return result, fmt.Errorf("%s already exists", clean.Log(fs.RelName(result, imp.originalsPath())))
		}

		iteration++

		result = filepath.Join(pathName, fileName+"."+fmt.Sprintf("%05d", iteration)+fileExtension)
	}

	return result, nil
}
