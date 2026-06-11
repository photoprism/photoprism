package photoprism

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime/debug"
	"sync"

	"github.com/karrick/godirwalk"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/media/video"
)

// Convert represents a file format conversion worker.
type Convert struct {
	conf               *config.Config
	cmdMutex           sync.Mutex
	sipsExclude        fs.ExtList
	darktableExclude   fs.ExtList
	rawTherapeeExclude fs.ExtList
	imageMagickExclude fs.ExtList
	ffmpegExclude      video.Formats
}

// NewConvert returns a new file format conversion worker.
func NewConvert(conf *config.Config) *Convert {
	c := &Convert{
		conf:               conf,
		sipsExclude:        fs.NewExtList(conf.SipsExclude()),
		darktableExclude:   fs.NewExtList(conf.DarktableExclude()),
		rawTherapeeExclude: fs.NewExtList(conf.RawTherapeeExclude()),
		imageMagickExclude: fs.NewExtList(conf.ImageMagickExclude()),
		ffmpegExclude:      ffmpeg.Exclude(),
	}

	return c
}

// FFmpegAllowed reports whether the media file's codec and container are both
// absent from the FFmpegExclude list and the file may therefore be handed to
// FFmpeg. The codec is taken from both the metadata and the built-in video
// probe so either detector can match. A file with unknown codec and unknown
// container passes through.
func (w *Convert) FFmpegAllowed(f *MediaFile) bool {
	if f == nil || len(w.ffmpegExclude) == 0 {
		return true
	}

	return !w.ffmpegExclude.Contains(f.MetaData().Codec, f.VideoInfo().VideoCodec, f.FileType().String())
}

// Cancel stops the current conversion operation.
func (w *Convert) Cancel() {
	mutex.IndexWorker.Cancel()
}

// insufficientStorage reports whether the converter must abort due to quota or low free disk space.
func (w *Convert) insufficientStorage() bool {
	if !w.conf.InsufficientStorage() {
		return false
	}

	log.Errorf("convert: aborting due to insufficient storage")
	event.ErrorMsg(i18n.ErrInsufficientStorage)
	return true
}

// cancelInsufficientStorage logs the insufficient-storage cause once and cancels the run so the
// directory walk stops, instead of failing every remaining file when the disk filled mid-convert.
// It is called from worker goroutines when a conversion leaf reports status.ErrInsufficientStorage.
func (w *Convert) cancelInsufficientStorage() {
	if mutex.IndexWorker.Canceled() {
		return
	}

	log.Errorf("convert: aborting due to insufficient storage")
	event.ErrorMsg(i18n.ErrInsufficientStorage)
	w.Cancel()
}

// Start converts all files in the specified directory based on the current configuration.
func (w *Convert) Start(dir string, ext []string, force bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("convert: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	// Reset the cached disk usage so a freshly freed disk is detected immediately.
	disk.FlushFree()

	if w.insufficientStorage() {
		return status.ErrInsufficientStorage
	}

	if err = mutex.IndexWorker.Start(); err != nil {
		return err
	}

	defer mutex.IndexWorker.Stop()

	jobs := make(chan ConvertJob)

	// Start a fixed number of goroutines to convert files.
	var wg sync.WaitGroup
	var numWorkers = w.conf.IndexWorkers()
	wg.Add(numWorkers)
	for range numWorkers {
		go func() {
			ConvertWorker(jobs)
			wg.Done()
		}()
	}

	done := make(fs.Done)
	ignore := fs.NewIgnoreList(fs.PPIgnoreFilename, true, false)

	if err = ignore.Path(dir); err != nil {
		log.Infof("convert: %s", err)
	}

	ignore.Log = func(fileName string) {
		log.Infof("convert: ignoring %s", clean.Log(filepath.Base(fileName)))
	}

	err = godirwalk.Walk(dir, &godirwalk.Options{
		ErrorCallback: func(fileName string, err error) godirwalk.ErrorAction {
			return godirwalk.SkipNode
		},
		Callback: func(fileName string, info *godirwalk.Dirent) error {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("convert: %s (panic)\nstack: %s", r, debug.Stack())
				}
			}()

			if mutex.IndexWorker.Canceled() {
				return status.ErrCanceled
			}

			// Stop the walk if storage drops below the threshold mid-convert.
			if w.insufficientStorage() {
				w.Cancel()
				return status.ErrInsufficientStorage
			}

			isDir, _ := info.IsDirOrSymlinkToDir()
			isSymlink := info.IsSymlink()

			// Skip file?
			if skip, result := fs.SkipWalk(fileName, isDir, isSymlink, done, ignore); skip {
				return result
			}

			// Process only files with specified extensions?
			if list.Excludes(ext, fs.NormalizedExt(fileName)) {
				return nil
			}

			f, err := NewMediaFile(fileName)

			if err != nil || f.Empty() || f.IsPreviewImage() || !f.IsMedia() {
				return nil
			}

			done[fileName] = fs.Processed

			jobs <- ConvertJob{
				force:   force,
				file:    f,
				convert: w,
			}

			return nil
		},
		Unsorted:            false,
		FollowSymbolicLinks: true,
	})

	close(jobs)
	wg.Wait()

	logWalkResult("convert", err)

	// A user-initiated Ctrl+C is expected; do not surface it as a CLI error.
	if errors.Is(err, status.ErrCanceled) {
		return nil
	}

	return err
}
