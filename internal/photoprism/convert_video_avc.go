package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/ffmpeg"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// ToAvc converts a single video file to MPEG-4 AVC.
func (c *Convert) ToAvc(f *MediaFile, encoder ffmpeg.AvcEncoder, noMutex, force bool) (file *MediaFile, err error) {
	// Abort if the source media file is nil.
	if f == nil {
		return nil, fmt.Errorf("convert: file is nil - you may have found a bug")
	}

	// Abort if the source media file does not exist.
	if !f.Exists() {
		return nil, fmt.Errorf("convert: %s not found", clean.Log(f.RootRelName()))
	} else if f.Empty() {
		return nil, fmt.Errorf("convert: %s is empty", clean.Log(f.RootRelName()))
	}

	// AVC video filename.
	var avcName string

	// Use .mp4 file extension for animated images and .avi for videos.
	if f.IsAnimatedImage() {
		avcName = fs.VideoMP4.FindFirst(f.FileName(), []string{c.conf.SidecarPath(), fs.PPHiddenPathname}, c.conf.OriginalsPath(), false)
	} else {
		avcName = fs.VideoAVC.FindFirst(f.FileName(), []string{c.conf.SidecarPath(), fs.PPHiddenPathname}, c.conf.OriginalsPath(), false)
	}

	mediaFile, err := NewMediaFile(avcName)

	// Check if AVC file already exists.
	if mediaFile == nil || err != nil {
		// No, transcode video to AVC.
	} else if mediaFile.IsVideo() {
		// Yes, return AVC video file.
		return mediaFile, nil
	}

	// Check if the sidecar path is writeable, so a new AVC file can be created.
	if !c.conf.SidecarWritable() {
		return nil, fmt.Errorf("convert: transcoding disabled in read-only mode (%s)", f.RootRelName())
	}

	// Get relative filename for logging.
	relName := f.RelName(c.conf.OriginalsPath())

	// Use .mp4 file extension for animated images and .avi for videos.
	if f.IsAnimatedImage() {
		avcName, _ = fs.FileName(f.FileName(), c.conf.SidecarPath(), c.conf.OriginalsPath(), fs.ExtMP4)
	} else {
		avcName, _ = fs.FileName(f.FileName(), c.conf.SidecarPath(), c.conf.OriginalsPath(), fs.ExtAVC)
	}

	cmd, useMutex, err := c.AvcConvertCommand(f, avcName, encoder)

	// Return if an error occurred.
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Make sure only one convert command runs at a time.
	if useMutex && !noMutex {
		c.cmdMutex.Lock()
		defer c.cmdMutex.Unlock()
	}

	// Check if target file already exists.
	if fs.FileExists(avcName) {
		avcFile, avcErr := NewMediaFile(avcName)
		if avcErr != nil {
			return avcFile, avcErr
		} else if !force || !avcFile.InSidecar() {
			return avcFile, nil
		} else if err = avcFile.Remove(); err != nil {
			return avcFile, fmt.Errorf("convert: failed removing %s (%s)", clean.Log(avcFile.RootRelName()), err)
		} else {
			log.Infof("convert: replacing %s", clean.Log(avcFile.RootRelName()))
		}
	}

	// Fetch command output.
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Env = []string{fmt.Sprintf("HOME=%s", c.conf.CmdCachePath())}

	event.Publish("index.converting", event.Data{
		"fileType": f.FileType(),
		"fileName": relName,
		"baseName": filepath.Base(relName),
		"xmpName":  "",
	})

	log.Infof("%s: transcoding %s to %s", encoder, relName, fs.VideoAVC)

	// Log exact command for debugging in trace mode.
	log.Trace(cmd.String())

	// Transcode source media file to AVC.
	start := time.Now()
	if err = cmd.Run(); err != nil {
		if stderr.String() != "" {
			err = errors.New(stderr.String())
		}

		// Log ffmpeg output for debugging.
		if err.Error() != "" {
			log.Debug(err)
		}

		// Log filename and transcoding time.
		log.Warnf("%s: failed to transcode %s [%s]", encoder, relName, time.Since(start))

		// Remove broken video file.
		if !fs.FileExists(avcName) {
			// Do nothing.
		} else if err = os.Remove(avcName); err != nil {
			return nil, fmt.Errorf("convert: failed to remove %s (%s)", clean.Log(RootRelName(avcName)), err)
		}

		// Try again using software encoder.
		if encoder != ffmpeg.SoftwareEncoder {
			return c.ToAvc(f, ffmpeg.SoftwareEncoder, true, false)
		} else {
			return nil, err
		}
	}

	// Log filename and transcoding time.
	log.Infof("%s: created %s [%s]", encoder, filepath.Base(avcName), time.Since(start))

	// Return AVC media file.
	return NewMediaFile(avcName)
}

// AvcConvertCommand returns the command for converting video files to MPEG-4 AVC.
func (c *Convert) AvcConvertCommand(f *MediaFile, avcName string, encoder ffmpeg.AvcEncoder) (result *exec.Cmd, useMutex bool, err error) {
	fileExt := f.Extension()
	fileName := f.FileName()

	switch {
	case fileName == "":
		return nil, false, fmt.Errorf("convert: %s video filename is empty - you may have found a bug", f.FileType())
	case !f.IsAnimated():
		return nil, false, fmt.Errorf("convert: file type %s of %s cannot be transcoded", f.FileType(), clean.Log(f.BaseName()))
	}

	// Try to transcode animated WebP images with ImageMagick.
	if c.conf.ImageMagickEnabled() && f.IsWebP() && c.imagemagickBlacklist.Allow(fileExt) {
		return exec.Command(c.conf.ImageMagickBin(), f.FileName(), avcName), false, nil
	}

	// Use FFmpeg to transcode all other media files to AVC.
	var opt ffmpeg.Options
	if opt, err = c.conf.FFmpegOptions(encoder, c.AvcBitrate(f)); err != nil {
		return nil, false, fmt.Errorf("convert: failed to transcode %s (%s)", clean.Log(f.BaseName()), err)
	} else {
		return ffmpeg.AvcConvertCommand(fileName, avcName, opt)
	}
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
