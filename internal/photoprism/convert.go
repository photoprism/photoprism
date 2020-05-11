package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/karrick/godirwalk"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
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
	if err := mutex.Worker.Start(); err != nil {
		return err
	}

	defer mutex.Worker.Stop()

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
	ignore := fs.NewIgnoreList(IgnoreFile, true, false)

	if err := ignore.Dir(path); err != nil {
		log.Infof("convert: %s", err)
	}

	ignore.Log = func(fileName string) {
		log.Infof(`convert: ignored "%s"`, fs.RelativeName(fileName, path))
	}

	err := godirwalk.Walk(path, &godirwalk.Options{
		Callback: func(fileName string, info *godirwalk.Dirent) error {
			defer func() {
				if err := recover(); err != nil {
					log.Errorf("convert: %s [panic]", err)
				}
			}()

			if mutex.Worker.Canceled() {
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

// ConvertCommand returns the command for converting files to JPEG, depending on the format.
func (c *Convert) ConvertCommand(mf *MediaFile, jpegName string, xmpName string) (result *exec.Cmd, useMutex bool, err error) {
	if mf.IsRaw() {
		if c.conf.SipsBin() != "" {
			result = exec.Command(c.conf.SipsBin(), "-s", "format", "jpeg", "--out", jpegName, mf.FileName())
		} else if c.conf.DarktableBin() != "" {
			// Only one instance of darktable-cli allowed due to locking
			useMutex = true

			if xmpName != "" {
				result = exec.Command(c.conf.DarktableBin(), mf.FileName(), xmpName, jpegName)
			} else {
				result = exec.Command(c.conf.DarktableBin(), mf.FileName(), jpegName)
			}
		} else {
			return nil, useMutex, fmt.Errorf("convert: no raw to jpeg converter installed (%s)", mf.Base(c.conf.Settings().Index.Group))
		}
	} else if mf.IsVideo() {
		result = exec.Command(c.conf.FFmpegBin(), "-i", mf.FileName(), "-ss", "00:00:00.001", "-vframes", "1", jpegName)
	} else if mf.IsHEIF() {
		result = exec.Command(c.conf.HeifConvertBin(), mf.FileName(), jpegName)
	} else {
		return nil, useMutex, fmt.Errorf("convert: file type not supported for conversion (%s)", mf.FileType())
	}

	return result, useMutex, nil
}

// ToJson uses exiftool to export metadata to a json file.
func (c *Convert) ToJson(mf *MediaFile) (*MediaFile, error) {
	jsonName := fs.TypeJson.Find(mf.FileName(), c.conf.Settings().Index.Group)

	result, err := NewMediaFile(jsonName)

	if err == nil {
		return result, nil
	}

	jsonName = mf.AbsBase(c.conf.Settings().Index.Group) + ".json"

	if c.conf.ReadOnly() {
		return nil, fmt.Errorf("convert: metadata export to json disabled in read only mode (%s)", mf.RelativeName(c.conf.OriginalsPath()))
	}

	fileName := mf.RelativeName(c.conf.OriginalsPath())

	log.Infof("convert: %s -> %s", fileName, fs.RelativeName(jsonName, c.conf.OriginalsPath()))

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

// ToJpeg converts a single image file to JPEG if possible.
func (c *Convert) ToJpeg(image *MediaFile) (*MediaFile, error) {
	if c.conf.ReadOnly() {
		return nil, errors.New("convert: disabled in read-only mode")
	}

	if !image.Exists() {
		return nil, fmt.Errorf("convert: can not convert to jpeg, file does not exist (%s)", image.RelativeName(c.conf.OriginalsPath()))
	}

	if image.IsJpeg() {
		return image, nil
	}

	jpegName := fs.TypeJpeg.Find(image.FileName(), c.conf.Settings().Index.Group)

	mediaFile, err := NewMediaFile(jpegName)

	if err == nil {
		return mediaFile, nil
	}

	jpegName = image.AbsBase(c.conf.Settings().Index.Group) + ".jpg"

	if c.conf.ReadOnly() {
		return nil, fmt.Errorf("convert: disabled in read only mode (%s)", image.RelativeName(c.conf.OriginalsPath()))
	}

	fileName := image.RelativeName(c.conf.OriginalsPath())

	log.Infof("convert: %s -> %s", fileName, fs.RelativeName(jpegName, c.conf.OriginalsPath()))

	xmpName := fs.TypeXMP.Find(image.FileName(), c.conf.Settings().Index.Group)

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

	cmd, useMutex, err := c.ConvertCommand(image, jpegName, xmpName)

	if err != nil {
		return nil, err
	}

	// Unclear if this is really necessary here, but safe is safe.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

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
