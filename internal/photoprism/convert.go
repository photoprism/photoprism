package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"sync"

	"github.com/karrick/godirwalk"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Convert represents a converter that can convert RAW/HEIF images to JPEG.
type Convert struct {
	conf     *config.Config
	cmdMutex sync.Mutex
}

// NewConvert returns a new converter and expects the config as argument.
func NewConvert(conf *config.Config) *Convert {
	return &Convert{conf: conf}
}

// Start converts all files in a directory to JPEG if possible.
func (c *Convert) Start(path string) error {
	if err := mutex.MainWorker.Start(); err != nil {
		return err
	}

	defer mutex.MainWorker.Stop()

	jobs := make(chan ConvertJob)

	// Start a fixed number of goroutines to convert files.
	var wg sync.WaitGroup
	var numWorkers = c.conf.Workers()
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			ConvertWorker(jobs)
			wg.Done()
		}()
	}

	done := make(map[string]bool)
	ignore := fs.NewIgnoreList(fs.IgnoreFile, true, false)

	if err := ignore.Dir(path); err != nil {
		log.Infof("convert: %s", err)
	}

	ignore.Log = func(fileName string) {
		log.Infof("convert: ignoring %s", txt.Quote(filepath.Base(fileName)))
	}

	err := godirwalk.Walk(path, &godirwalk.Options{
		Callback: func(fileName string, info *godirwalk.Dirent) error {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("convert: %s (panic)\nstack: %s", r, debug.Stack())
				}
			}()

			if mutex.MainWorker.Canceled() {
				return errors.New("convert: canceled")
			}

			isDir := info.IsDir()
			isSymlink := info.IsSymlink()

			if skip, result := fs.SkipWalk(fileName, isDir, isSymlink, done, ignore); skip {
				return result
			}

			mf, err := NewMediaFile(fileName)

			if err != nil || !(mf.IsRaw() || mf.IsHEIF() || mf.IsImageOther()) {
				return nil
			}

			done[fileName] = true

			jobs <- ConvertJob{
				image:   mf,
				convert: c,
			}

			return nil
		},
		Unsorted:            true,
		FollowSymbolicLinks: true,
	})

	close(jobs)
	wg.Wait()

	return err
}

// ToJson uses exiftool to export metadata to a json file.
func (c *Convert) ToJson(mf *MediaFile) (*MediaFile, error) {
	jsonName := fs.TypeJson.FindFirst(mf.FileName(), []string{c.conf.SidecarPath(), fs.HiddenPath}, c.conf.OriginalsPath(), c.conf.Settings().Index.Sequences)

	result, err := NewMediaFile(jsonName)

	if err == nil {
		return result, nil
	}

	if !c.conf.SidecarWritable() {
		return nil, fmt.Errorf("convert: can't create json sidecar file for %s in read only mode", txt.Quote(mf.BaseName()))
	}

	jsonName = fs.FileName(mf.FileName(), c.conf.SidecarPath(), c.conf.OriginalsPath(), ".json", c.conf.Settings().Index.Sequences)

	fileName := mf.RelName(c.conf.OriginalsPath())

	log.Infof("convert: %s -> %s", fileName, filepath.Base(jsonName))

	cmd := exec.Command(c.conf.ExifToolBin(), "-j", mf.FileName())

	// Fetch command output.
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Run convert command.
	if err := cmd.Run(); err != nil {
		if stderr.String() != "" {
			return nil, errors.New(stderr.String())
		} else {
			return nil, err
		}
	}

	// Write output to file.
	if err := ioutil.WriteFile(jsonName, []byte(out.String()), os.ModePerm); err != nil {
		return nil, err
	}

	// Check if file exists.
	if !fs.FileExists(jsonName) {
		return nil, fmt.Errorf("convert: %s could not be created, check configuration", jsonName)
	}

	return NewMediaFile(jsonName)
}

// JpegConvertCommand returns the command for converting files to JPEG, depending on the format.
func (c *Convert) JpegConvertCommand(mf *MediaFile, jpegName string, xmpName string) (result *exec.Cmd, useMutex bool, err error) {
	size := strconv.Itoa(c.conf.JpegSize())

	if mf.IsRaw() {
		if c.conf.SipsBin() != "" {
			result = exec.Command(c.conf.SipsBin(), "-Z", size, "-s", "format", "jpeg", "--out", jpegName, mf.FileName())
		} else if c.conf.DarktableBin() != "" {
			var args []string

			// Only one instance of darktable-cli allowed due to locking if presets are loaded.
			if c.conf.DarktableUnlock() {
				useMutex = false
				args = []string{"--apply-custom-presets", "false", "--width", size, "--height", size, mf.FileName()}
			} else {
				useMutex = true
				args = []string{"--width", size, "--height", size, mf.FileName()}
			}

			if xmpName != "" {
				args = append(args, xmpName, jpegName)
			} else {
				args = append(args, jpegName)
			}

			result = exec.Command(c.conf.DarktableBin(), args...)
		} else {
			return nil, useMutex, fmt.Errorf("convert: no converter found for %s", txt.Quote(mf.BaseName()))
		}
	} else if mf.IsVideo() {
		result = exec.Command(c.conf.FFmpegBin(), "-i", mf.FileName(), "-ss", "00:00:00.001", "-vframes", "1", jpegName)
	} else if mf.IsHEIF() {
		result = exec.Command(c.conf.HeifConvertBin(), mf.FileName(), jpegName)
	} else {
		return nil, useMutex, fmt.Errorf("convert: file type %s not supported in %s", mf.FileType(), txt.Quote(mf.BaseName()))
	}

	return result, useMutex, nil
}

// ToJpeg converts a single image file to JPEG if possible.
func (c *Convert) ToJpeg(image *MediaFile) (*MediaFile, error) {
	if !image.Exists() {
		return nil, fmt.Errorf("convert: can not convert to jpeg, file does not exist (%s)", image.RelName(c.conf.OriginalsPath()))
	}

	if image.IsJpeg() {
		return image, nil
	}

	jpegName := fs.TypeJpeg.FindFirst(image.FileName(), []string{c.conf.SidecarPath(), fs.HiddenPath}, c.conf.OriginalsPath(), c.conf.Settings().Index.Sequences)

	mediaFile, err := NewMediaFile(jpegName)

	if err == nil {
		return mediaFile, nil
	}

	if !c.conf.SidecarWritable() {
		return nil, fmt.Errorf("convert: disabled in read only mode (%s)", image.RelName(c.conf.OriginalsPath()))
	}

	jpegName = fs.FileName(image.FileName(), c.conf.SidecarPath(), c.conf.OriginalsPath(), fs.JpegExt, c.conf.Settings().Index.Sequences)
	fileName := image.RelName(c.conf.OriginalsPath())

	log.Infof("convert: %s -> %s", fileName, filepath.Base(jpegName))

	xmpName := fs.TypeXMP.Find(image.FileName(), c.conf.Settings().Index.Sequences)

	event.Publish("index.converting", event.Data{
		"fileType": image.FileType(),
		"fileName": fileName,
		"baseName": filepath.Base(fileName),
		"xmpName":  filepath.Base(xmpName),
	})

	if image.IsImageOther() {
		_, err = thumb.Jpeg(image.FileName(), jpegName)

		if err != nil {
			return nil, err
		}

		return NewMediaFile(jpegName)
	}

	cmd, useMutex, err := c.JpegConvertCommand(image, jpegName, xmpName)

	if err != nil {
		return nil, err
	}

	if useMutex {
		// Make sure only one command is executed at a time.
		// See https://photo.stackexchange.com/questions/105969/darktable-cli-fails-because-of-locked-database-file
		c.cmdMutex.Lock()
		defer c.cmdMutex.Unlock()
	}

	// Fetch command output.
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Run convert command.
	if err := cmd.Run(); err != nil {
		if stderr.String() != "" {
			return nil, errors.New(stderr.String())
		} else {
			return nil, err
		}
	}

	return NewMediaFile(jpegName)
}
