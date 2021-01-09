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
	"strings"
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
func (c *Convert) Start(path string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("convert: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

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

	done := make(fs.Done)
	ignore := fs.NewIgnoreList(fs.IgnoreFile, true, false)

	if err := ignore.Dir(path); err != nil {
		log.Infof("convert: %s", err)
	}

	ignore.Log = func(fileName string) {
		log.Infof("convert: ignoring %s", txt.Quote(filepath.Base(fileName)))
	}

	err = godirwalk.Walk(path, &godirwalk.Options{
		ErrorCallback: func(fileName string, err error) godirwalk.ErrorAction {
			log.Errorf("convert: %s", strings.Replace(err.Error(), path, "", 1))
			return godirwalk.SkipNode
		},
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

			f, err := NewMediaFile(fileName)

			if err != nil || !(f.IsRaw() || f.IsHEIF() || f.IsImageOther()) {
				return nil
			}

			done[fileName] = fs.Processed

			jobs <- ConvertJob{
				image:   f,
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
func (c *Convert) ToJson(f *MediaFile) (jsonName string, err error) {
	if f == nil {
		return "", fmt.Errorf("exiftool: file is nil (found a bug?)")
	}

	jsonName, err = f.ExifToolJsonName()

	if err != nil {
		return "", nil
	}

	if fs.FileExists(jsonName) {
		return jsonName, nil
	}

	relName := f.RelName(c.conf.OriginalsPath())

	log.Debugf("exiftool: extracting metadata from %s", relName)

	cmd := exec.Command(c.conf.ExifToolBin(), "-j", f.FileName())

	// Fetch command output.
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Run convert command.
	if err := cmd.Run(); err != nil {
		if stderr.String() != "" {
			return "", errors.New(stderr.String())
		} else {
			return "", err
		}
	}

	// Write output to file.
	if err := ioutil.WriteFile(jsonName, []byte(out.String()), os.ModePerm); err != nil {
		return "", err
	}

	// Check if file exists.
	if !fs.FileExists(jsonName) {
		return "", fmt.Errorf("exiftool: failed creating %s", filepath.Base(jsonName))
	}

	return jsonName, err
}

// JpegConvertCommand returns the command for converting files to JPEG, depending on the format.
func (c *Convert) JpegConvertCommand(f *MediaFile, jpegName string, xmpName string) (result *exec.Cmd, useMutex bool, err error) {
	size := strconv.Itoa(c.conf.JpegSize())

	if f.IsRaw() {
		if c.conf.SipsBin() != "" {
			result = exec.Command(c.conf.SipsBin(), "-Z", size, "-s", "format", "jpeg", "--out", jpegName, f.FileName())
		} else if c.conf.DarktableBin() != "" && f.Extension() != ".cr3" {
			var args []string

			// Only one instance of darktable-cli allowed due to locking if presets are loaded.
			if c.conf.DarktablePresets() {
				useMutex = true
				args = []string{"--width", size, "--height", size, f.FileName()}
			} else {
				useMutex = false
				args = []string{"--apply-custom-presets", "false", "--width", size, "--height", size, f.FileName()}
			}

			if xmpName != "" {
				args = append(args, xmpName, jpegName)
			} else {
				args = append(args, jpegName)
			}

			result = exec.Command(c.conf.DarktableBin(), args...)
		} else if c.conf.RawtherapeeBin() != "" {
			jpegQuality := fmt.Sprintf("-j%d", c.conf.JpegQuality())
			profile := filepath.Join(conf.AssetsPath(), "profiles", "raw.pp3")

			args := []string{"-o", jpegName, "-p", profile, "-d", jpegQuality, "-js3", "-b8", "-c", f.FileName()}

			result = exec.Command(c.conf.RawtherapeeBin(), args...)
		} else {
			return nil, useMutex, fmt.Errorf("convert: no converter found for %s", txt.Quote(f.BaseName()))
		}
	} else if f.IsVideo() {
		result = exec.Command(c.conf.FFmpegBin(), "-y", "-i", f.FileName(), "-ss", "00:00:00.001", "-vframes", "1", jpegName)
	} else if f.IsHEIF() {
		result = exec.Command(c.conf.HeifConvertBin(), f.FileName(), jpegName)
	} else {
		return nil, useMutex, fmt.Errorf("convert: file type %s not supported in %s", f.FileType(), txt.Quote(f.BaseName()))
	}

	return result, useMutex, nil
}

// ToJpeg converts a single image file to JPEG if possible.
func (c *Convert) ToJpeg(f *MediaFile) (*MediaFile, error) {
	if f == nil {
		return nil, fmt.Errorf("convert: file is nil (found a bug?)")
	}

	if !f.Exists() {
		return nil, fmt.Errorf("convert: can not convert to jpeg, file does not exist (%s)", f.RelName(c.conf.OriginalsPath()))
	}

	if f.IsJpeg() {
		return f, nil
	}

	jpegName := fs.FormatJpeg.FindFirst(f.FileName(), []string{c.conf.SidecarPath(), fs.HiddenPath}, c.conf.OriginalsPath(), false)

	mediaFile, err := NewMediaFile(jpegName)

	if err == nil && mediaFile.IsJpeg() {
		return mediaFile, nil
	}

	if !c.conf.SidecarWritable() {
		return nil, fmt.Errorf("convert: disabled in read only mode (%s)", f.RelName(c.conf.OriginalsPath()))
	}

	jpegName = fs.FileName(f.FileName(), c.conf.SidecarPath(), c.conf.OriginalsPath(), fs.JpegExt)
	fileName := f.RelName(c.conf.OriginalsPath())

	log.Debugf("convert: %s -> %s", fileName, filepath.Base(jpegName))

	xmpName := fs.FormatXMP.Find(f.FileName(), false)

	event.Publish("index.converting", event.Data{
		"fileType": f.FileType(),
		"fileName": fileName,
		"baseName": filepath.Base(fileName),
		"xmpName":  filepath.Base(xmpName),
	})

	if f.IsImageOther() {
		_, err = thumb.Jpeg(f.FileName(), jpegName)

		if err != nil {
			return nil, err
		}

		return NewMediaFile(jpegName)
	}

	cmd, useMutex, err := c.JpegConvertCommand(f, jpegName, xmpName)

	if err != nil {
		return nil, err
	}

	if useMutex {
		// Make sure only one command is executed at a time.
		// See https://photo.stackexchange.com/questions/105969/darktable-cli-fails-because-of-locked-database-file
		c.cmdMutex.Lock()
		defer c.cmdMutex.Unlock()
	}

	if fs.FileExists(jpegName) {
		return NewMediaFile(jpegName)
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

// AvcConvertCommand returns the command for converting video files to MPEG-4 AVC.
func (c *Convert) AvcConvertCommand(f *MediaFile, avcName string) (result *exec.Cmd, useMutex bool, err error) {
	if f.IsVideo() {
		// Don't transcode more than one video at the same time.
		useMutex = true
		result = exec.Command(
			c.conf.FFmpegBin(),
			"-i", f.FileName(),
			"-c:v", "libx264",
			"-f", "mp4",
			avcName,
		)
	} else {
		return nil, useMutex, fmt.Errorf("convert: file type %s not supported in %s", f.FileType(), txt.Quote(f.BaseName()))
	}

	return result, useMutex, nil
}

// ToAvc converts a single video file to MPEG-4 AVC.
func (c *Convert) ToAvc(f *MediaFile) (*MediaFile, error) {
	if f == nil {
		return nil, fmt.Errorf("convert: file is nil (found a bug?)")
	}

	if !f.Exists() {
		return nil, fmt.Errorf("convert: can not convert to avc1, file does not exist (%s)", f.RelName(c.conf.OriginalsPath()))
	}

	avcName := fs.FormatAvc.FindFirst(f.FileName(), []string{c.conf.SidecarPath(), fs.HiddenPath}, c.conf.OriginalsPath(), false)

	mediaFile, err := NewMediaFile(avcName)

	if err == nil && mediaFile.IsVideo() {
		return mediaFile, nil
	}

	if !c.conf.SidecarWritable() {
		return nil, fmt.Errorf("convert: disabled in read only mode (%s)", f.RelName(c.conf.OriginalsPath()))
	}

	avcName = fs.FileName(f.FileName(), c.conf.SidecarPath(), c.conf.OriginalsPath(), fs.AvcExt)
	fileName := f.RelName(c.conf.OriginalsPath())

	log.Debugf("convert: %s -> %s", fileName, filepath.Base(avcName))

	event.Publish("index.converting", event.Data{
		"fileType": f.FileType(),
		"fileName": fileName,
		"baseName": filepath.Base(fileName),
		"xmpName":  "",
	})

	cmd, useMutex, err := c.AvcConvertCommand(f, avcName)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	if useMutex {
		// Make sure only one command is executed at a time.
		// See https://photo.stackexchange.com/questions/105969/darktable-cli-fails-because-of-locked-database-file
		c.cmdMutex.Lock()
		defer c.cmdMutex.Unlock()
	}

	if fs.FileExists(avcName) {
		return NewMediaFile(avcName)
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

	return NewMediaFile(avcName)
}
