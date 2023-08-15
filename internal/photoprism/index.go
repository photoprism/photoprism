package photoprism

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/karrick/godirwalk"

	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/nsfw"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// Index represents an indexer that indexes files in the originals directory.
type Index struct {
	conf         *config.Config
	tensorFlow   *classify.TensorFlow
	nsfwDetector *nsfw.Detector
	faceNet      *face.Net
	convert      *Convert
	files        *Files
	photos       *Photos
	lastRun      time.Time
	lastFound    int
	findFaces    bool
	findLabels   bool
}

// NewIndex returns a new indexer and expects its dependencies as arguments.
func NewIndex(conf *config.Config, tensorFlow *classify.TensorFlow, nsfwDetector *nsfw.Detector, faceNet *face.Net, convert *Convert, files *Files, photos *Photos) *Index {
	if conf == nil {
		log.Errorf("index: config is not set")
		return nil
	}

	i := &Index{
		conf:         conf,
		tensorFlow:   tensorFlow,
		nsfwDetector: nsfwDetector,
		faceNet:      faceNet,
		convert:      convert,
		files:        files,
		photos:       photos,
		findFaces:    !conf.DisableFaces(),
		findLabels:   !conf.DisableClassification(),
	}

	return i
}

func (ind *Index) originalsPath() string {
	return ind.conf.OriginalsPath()
}

func (ind *Index) thumbPath() string {
	return ind.conf.ThumbCachePath()
}

// Cancel stops the current indexing operation.
func (ind *Index) Cancel() {
	mutex.MainWorker.Cancel()
}

// Start indexes media files in the "originals" folder.
func (ind *Index) Start(o IndexOptions) (found fs.Done, updated int) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("index: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	found = make(fs.Done)

	if ind.conf == nil {
		log.Errorf("index: config is not set")
		return found, updated
	}

	originalsPath := ind.originalsPath()
	optionsPath := filepath.Join(originalsPath, o.Path)

	if !fs.PathExists(optionsPath) {
		event.Error(fmt.Sprintf("index: directory %s not found", clean.Log(optionsPath)))
		return found, updated
	} else if fs.DirIsEmpty(originalsPath) {
		event.InfoMsg(i18n.ErrOriginalsEmpty)
		return found, updated
	}

	if err := mutex.MainWorker.Start(); err != nil {
		event.Error(fmt.Sprintf("index: %s", err.Error()))
		return found, updated
	}

	defer mutex.MainWorker.Stop()

	if err := ind.tensorFlow.Init(); err != nil {
		log.Errorf("index: %s", err.Error())

		return found, updated
	}

	jobs := make(chan IndexJob)

	// Start a fixed number of goroutines to index files.
	var wg sync.WaitGroup
	var numWorkers = ind.conf.Workers()
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			IndexWorker(jobs) // HLc
			wg.Done()
		}()
	}

	if err := ind.files.Init(); err != nil {
		log.Errorf("index: %s", err)
	}

	defer ind.files.Done()

	skipRaw := ind.conf.DisableRaw()
	ignore := fs.NewIgnoreList(fs.IgnoreFile, true, false)

	if err := ignore.Dir(originalsPath); err != nil {
		log.Infof("index: %s", err)
	}

	ignore.Log = func(fileName string) {
		log.Infof(`index: ignored "%s"`, fs.RelName(fileName, originalsPath))
	}

	err := godirwalk.Walk(optionsPath, &godirwalk.Options{
		ErrorCallback: func(fileName string, err error) godirwalk.ErrorAction {
			return godirwalk.SkipNode
		},
		Callback: func(fileName string, info *godirwalk.Dirent) error {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("index: %s (panic)\nstack: %s", r, debug.Stack())
				}
			}()

			if mutex.MainWorker.Canceled() {
				return errors.New("canceled")
			}

			isDir, _ := info.IsDirOrSymlinkToDir()
			isSymlink := info.IsSymlink()
			relName := fs.RelName(fileName, originalsPath)

			// Skip directories and known files.
			if skip, result := fs.SkipWalk(fileName, isDir, isSymlink, found, ignore); skip {
				if !isDir {
					return result
				}

				if result != filepath.SkipDir {
					folder := entity.NewFolder(entity.RootOriginals, relName, fs.BirthTime(fileName))

					if err := folder.Create(); err == nil {
						log.Infof("index: added folder /%s", folder.Path)
					}
				}

				event.Publish("index.folder", event.Data{
					"uid":      o.UID,
					"filePath": relName,
				})

				return result
			}

			found[fileName] = fs.Found

			if !media.MainFile(fileName) {
				return nil
			}

			var mf *MediaFile
			var err error
			if isSymlink {
				mf, err = NewMediaFile(fileName)
			} else {
				// If the file found while scanning is not a symlink we can
				// skip resolving the fileName, which is resource intensive.
				mf, err = NewMediaFileSkipResolve(fileName, fileName)
			}

			// Check if file exists and is not empty.
			if err != nil {
				log.Warnf("index: %s", err)
				return nil
			} else if mf.Empty() {
				return nil
			}

			// Skip already indexed?
			if ind.files.Indexed(relName, entity.RootOriginals, mf.modTime, o.Rescan) {
				return nil
			}

			// Skip RAW image?
			if mf.IsRaw() && skipRaw {
				log.Infof("index: skipped raw %s", clean.Log(mf.RootRelName()))
				return nil
			}

			// Create JSON sidecar file, if needed.
			if err = mf.CreateExifToolJson(ind.convert); err != nil {
				log.Errorf("index: %s", clean.LogError(err), clean.Log(mf.BaseName()))
			}

			// Find related files to index.
			related, err := mf.RelatedFiles(ind.conf.Settings().StackSequences())

			if err != nil {
				log.Warnf("index: %s", err)
				return nil
			}

			var files MediaFiles

			// Main media file is required to proceed.
			if related.Main == nil {
				return nil
			}

			skip := false

			// Check related files.
			for _, f := range related.Files {
				if found[f.FileName()].Processed() {
					// Ignore already processed files.
					continue
				} else if limitErr, fileSize := f.ExceedsBytes(o.ByteLimit); fileSize == 0 || ind.files.Indexed(f.RootRelName(), f.Root(), f.ModTime(), o.Rescan) {
					// Flag file as found but not processed.
					found[f.FileName()] = fs.Found
					continue
				} else if limitErr == nil {
					// Add to file list.
					files = append(files, f)
				} else if related.Main.FileName() != f.FileName() {
					// Sidecar file is too large, ignore.
					log.Infof("index: %s", limitErr)
				} else {
					// Main file is too large, skip all.
					log.Warnf("index: %s", limitErr)
					skip = true
				}

				found[f.FileName()] = fs.Processed
			}

			found[fileName] = fs.Processed

			// Skip if main file is too large or there are no files left to index.
			if skip || len(files) == 0 {
				return nil
			}

			updated += len(files)
			related.Files = files

			jobs <- IndexJob{
				FileName: mf.FileName(),
				Related:  related,
				IndexOpt: o,
				Ind:      ind,
			}

			return nil
		},
		Unsorted:            false,
		FollowSymbolicLinks: true,
	})

	close(jobs)
	wg.Wait()

	if err != nil {
		log.Error(err.Error())
	}

	if updated > 0 {
		event.Publish("index.updating", event.Data{
			"uid":  o.UID,
			"step": "faces",
		})

		// Run face recognition if enabled.
		if w := NewFaces(ind.conf); w.Disabled() {
			log.Debugf("index: skipping face recognition")
		} else if err := w.Start(FacesOptionsDefault()); err != nil {
			log.Errorf("index: %s", err)
		}

		event.Publish("index.updating", event.Data{
			"uid":  o.UID,
			"step": "counts",
		})

		// Update precalculated photo and file counts.
		if err := entity.UpdateCounts(); err != nil {
			log.Warnf("index: %s (update counts)", err)
		}
	} else {
		log.Infof("index: found no new or modified files")
	}

	runtime.GC()

	ind.lastRun = entity.TimeStamp()
	ind.lastFound = len(found)

	return found, updated
}

// LastRun returns the time when the index was last updated and how many files were found.
func (ind *Index) LastRun() (lastRun time.Time, lastFound int) {
	return ind.lastRun, ind.lastFound
}

// FileName indexes a single file and returns the result.
func (ind *Index) FileName(fileName string, o IndexOptions) (result IndexResult) {
	file, err := NewMediaFile(fileName)

	if err != nil {
		result.Err = err
		result.Status = IndexFailed

		return result
	} else if file.Empty() {
		result.Status = IndexSkipped

		return result
	}

	related, err := file.RelatedFiles(false)

	if err != nil {
		result.Err = err
		result.Status = IndexFailed

		return result
	}

	return IndexRelated(related, ind, o)
}
