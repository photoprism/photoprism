package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/karrick/godirwalk"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// Convert represents a converter that can convert RAW/HEIF images to JPEG.
type Convert struct {
	conf                 *config.Config
	cmdMutex             sync.Mutex
	darktableBlacklist   fs.Blacklist
	rawtherapeeBlacklist fs.Blacklist
}

// NewConvert returns a new converter and expects the config as argument.
func NewConvert(conf *config.Config) *Convert {
	c := &Convert{
		conf:                 conf,
		darktableBlacklist:   fs.NewBlacklist(conf.DarktableBlacklist()),
		rawtherapeeBlacklist: fs.NewBlacklist(conf.RawtherapeeBlacklist()),
	}

	return c
}

// Start converts all files in a directory to JPEG if possible.
func (c *Convert) Start(path string, force bool) (err error) {
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
		log.Infof("convert: ignoring %s", sanitize.Log(filepath.Base(fileName)))
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

			if err != nil || !(f.IsRaw() || f.IsHEIF() || f.IsImageOther() || f.IsVideo()) {
				return nil
			}

			done[fileName] = fs.Processed

			jobs <- ConvertJob{
				force:   force,
				file:    f,
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
		return "", fmt.Errorf("exiftool: file is nil - you might have found a bug")
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

	cmd := exec.Command(c.conf.ExifToolBin(), "-n", "-m", "-api", "LargeFileSupport", "-j", f.FileName())

	// Fetch command output.
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Log exact command for debugging in trace mode.
	log.Trace(cmd.String())

	// Run convert command.
	if err := cmd.Run(); err != nil {
		if stderr.String() != "" {
			return "", errors.New(stderr.String())
		} else {
			return "", err
		}
	}

	// Write output to file.
	if err := os.WriteFile(jsonName, []byte(out.String()), os.ModePerm); err != nil {
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
	if f == nil {
		return result, useMutex, fmt.Errorf("file is nil - you might have found a bug")
	}

	size := strconv.Itoa(c.conf.JpegSize())
	fileExt := f.Extension()

	if f.IsRaw() {
		if c.conf.SipsEnabled() {
			result = exec.Command(c.conf.SipsBin(), "-Z", size, "-s", "format", "jpeg", "--out", jpegName, f.FileName())
		} else if c.conf.DarktableEnabled() && c.darktableBlacklist.Ok(fileExt) {
			var args []string

			// Set RAW, XMP, and JPEG filenames.
			if xmpName != "" {
				args = []string{f.FileName(), xmpName, jpegName}
			} else {
				args = []string{f.FileName(), jpegName}
			}

			// Set RAW to JPEG conversion options.
			if c.conf.RawPresets() {
				useMutex = true // can run one instance only with presets enabled
				args = append(args, "--width", size, "--height", size, "--hq", "true", "--upscale", "false")
			} else {
				useMutex = false // --apply-custom-presets=false disables locking
				args = append(args, "--apply-custom-presets", "false", "--width", size, "--height", size, "--hq", "true", "--upscale", "false")
			}

			// Set Darktable core storage paths.
			args = append(args, "--core", "--configdir", c.conf.DarktableConfigPath(), "--cachedir", c.conf.DarktableCachePath(), "--library", ":memory:")

			result = exec.Command(c.conf.DarktableBin(), args...)
		} else if c.conf.RawtherapeeEnabled() && c.rawtherapeeBlacklist.Ok(fileExt) {
			jpegQuality := fmt.Sprintf("-j%d", c.conf.JpegQuality())
			profile := filepath.Join(conf.AssetsPath(), "profiles", "raw.pp3")

			args := []string{"-o", jpegName, "-p", profile, "-s", "-d", jpegQuality, "-js3", "-b8", "-c", f.FileName()}

			result = exec.Command(c.conf.RawtherapeeBin(), args...)
		} else {
			return nil, useMutex, fmt.Errorf("no suitable converter found")
		}
	} else if f.IsVideo() && c.conf.FFmpegEnabled() {
		result = exec.Command(c.conf.FFmpegBin(), "-y", "-i", f.FileName(), "-ss", "00:00:00.001", "-vframes", "1", jpegName)
	} else if f.IsHEIF() && c.conf.HeifConvertEnabled() {
		result = exec.Command(c.conf.HeifConvertBin(), f.FileName(), jpegName)
	} else {
		return nil, useMutex, fmt.Errorf("file type %s not supported", f.FileType())
	}

	// Log convert command in trace mode only as it exposes server internals.
	if result != nil {
		log.Tracef("convert: %s", result.String())
	}

	return result, useMutex, nil
}

// ToJpeg converts a single image file to JPEG if possible.
func (c *Convert) ToJpeg(f *MediaFile, force bool) (*MediaFile, error) {
	if f == nil {
		return nil, fmt.Errorf("convert: file is nil - you might have found a bug")
	}

	if !f.Exists() {
		return nil, fmt.Errorf("convert: %s not found", f.RelName(c.conf.OriginalsPath()))
	}

	if f.IsJpeg() {
		return f, nil
	}

	jpegName := fs.FormatJpeg.FindFirst(f.FileName(), []string{c.conf.SidecarPath(), fs.HiddenPath}, c.conf.OriginalsPath(), false)

	mediaFile, err := NewMediaFile(jpegName)

	// Replace existing sidecar if "force" is true.
	if err == nil && mediaFile.IsJpeg() {
		if force && mediaFile.InSidecar() {
			if err := mediaFile.Remove(); err != nil {
				return mediaFile, fmt.Errorf("convert: failed removing %s (%s)", mediaFile.RootRelName(), err)
			} else {
				log.Infof("convert: replacing %s", sanitize.Log(mediaFile.RootRelName()))
			}
		} else {
			return mediaFile, nil
		}
	} else {
		jpegName = fs.FileName(f.FileName(), c.conf.SidecarPath(), c.conf.OriginalsPath(), fs.JpegExt)
	}

	if !c.conf.SidecarWritable() {
		return nil, fmt.Errorf("convert: disabled in read only mode (%s)", f.RelName(c.conf.OriginalsPath()))
	}

	fileName := f.RelName(c.conf.OriginalsPath())
	xmpName := fs.FormatXMP.Find(f.FileName(), false)

	event.Publish("index.converting", event.Data{
		"fileType": f.FileType(),
		"fileName": fileName,
		"baseName": filepath.Base(fileName),
		"xmpName":  filepath.Base(xmpName),
	})

	start := time.Now()

	if f.IsImageOther() {
		log.Infof("convert: converting %s to %s (%s)", sanitize.Log(filepath.Base(fileName)), sanitize.Log(filepath.Base(jpegName)), f.FileType())

		_, err = thumb.Jpeg(f.FileName(), jpegName, f.Orientation())

		if err != nil {
			return nil, err
		}

		log.Infof("convert: %s created in %s (%s)", sanitize.Log(filepath.Base(jpegName)), time.Since(start), f.FileType())

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

	log.Infof("convert: converting %s to %s (%s)", sanitize.Log(filepath.Base(fileName)), sanitize.Log(filepath.Base(jpegName)), filepath.Base(cmd.Path))

	// Log exact command for debugging in trace mode.
	log.Trace(cmd.String())

	// Run convert command.
	if err := cmd.Run(); err != nil {
		if stderr.String() != "" {
			return nil, errors.New(stderr.String())
		} else {
			return nil, err
		}
	}

	log.Infof("convert: %s created in %s (%s)", sanitize.Log(filepath.Base(jpegName)), time.Since(start), filepath.Base(cmd.Path))

	return NewMediaFile(jpegName)
}

// AvcBitrate returns the ideal AVC encoding bitrate in megabits per second.
func (c *Convert) AvcBitrate(f *MediaFile) string {
	const defaultBitrate = "8M"

	if f == nil {
		return defaultBitrate
	}

	limit := c.conf.FFmpegBitrate()
	quality := 12

	bitrate := int(math.Ceil(float64(f.Width()*f.Height()*quality) / 1000000))

	if bitrate <= 0 {
		return defaultBitrate
	} else if bitrate > limit {
		bitrate = limit
	}

	return fmt.Sprintf("%dM", bitrate)
}
