package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// FFmpegSoftwareEncoder see https://trac.ffmpeg.org/wiki/HWAccelIntro.
const FFmpegSoftwareEncoder = "libx264"

// FFmpegIntelEncoder is the Intel Quick Sync H.264 encoder.
const FFmpegIntelEncoder = "h264_qsv"

// FFmpegAppleEncoder is the Apple Video Toolboar H.264 encoder.
const FFmpegAppleEncoder = "h264_videotoolbox"

// FFmpegVAAPIEncoder is the Video Acceleration API H.264 encoder.
const FFmpegVAAPIEncoder = "h264_vaapi"

// FFmpegNvidiaEncoder is the NVIDIA H.264 encoder.
const FFmpegNvidiaEncoder = "h264_nvenc"

// FFmpegV4L2Encoder is the Video4Linux H.264 encoder.
const FFmpegV4L2Encoder = "h264_v4l2m2m"

// FFmpegAvcEncoders is the list of supported H.264 encoders with aliases.
var FFmpegAvcEncoders = map[string]string{
	"":                    FFmpegSoftwareEncoder,
	"default":             FFmpegSoftwareEncoder,
	"software":            FFmpegSoftwareEncoder,
	FFmpegSoftwareEncoder: FFmpegSoftwareEncoder,
	"intel":               FFmpegIntelEncoder,
	"qsv":                 FFmpegIntelEncoder,
	FFmpegIntelEncoder:    FFmpegIntelEncoder,
	"apple":               FFmpegAppleEncoder,
	"osx":                 FFmpegAppleEncoder,
	"mac":                 FFmpegAppleEncoder,
	"macos":               FFmpegAppleEncoder,
	FFmpegAppleEncoder:    FFmpegAppleEncoder,
	"vaapi":               FFmpegVAAPIEncoder,
	"libva":               FFmpegVAAPIEncoder,
	FFmpegVAAPIEncoder:    FFmpegVAAPIEncoder,
	"nvidia":              FFmpegNvidiaEncoder,
	"nvenc":               FFmpegNvidiaEncoder,
	"cuda":                FFmpegNvidiaEncoder,
	FFmpegNvidiaEncoder:   FFmpegNvidiaEncoder,
	"v4l2":                FFmpegV4L2Encoder,
	"video4linux":         FFmpegV4L2Encoder,
	"rp4":                 FFmpegV4L2Encoder,
	"raspberry":           FFmpegV4L2Encoder,
	"raspberrypi":         FFmpegV4L2Encoder,
	FFmpegV4L2Encoder:     FFmpegV4L2Encoder,
}

// AvcConvertCommand returns the command for converting video files to MPEG-4 AVC.
func (c *Convert) AvcConvertCommand(f *MediaFile, avcName, encoderName string) (result *exec.Cmd, useMutex bool, err error) {
	if f.IsVideo() {
		// Don't transcode more than one video at the same time.
		useMutex = true

		// Display encoder info.
		if encoderName != FFmpegSoftwareEncoder {
			log.Infof("convert: ffmpeg encoder %s selected", encoderName)
		}

		if encoderName == FFmpegIntelEncoder {
			format := "format=rgb32"

			result = exec.Command(
				c.conf.FFmpegBin(),
				"-qsv_device", "/dev/dri/renderD128",
				"-init_hw_device", "qsv=hw",
				"-filter_hw_device", "hw",
				"-i", f.FileName(),
				"-c:a", "aac",
				"-vf", format,
				"-c:v", encoderName,
				"-vsync", "vfr",
				"-r", "30",
				"-b:v", c.AvcBitrate(f),
				"-maxrate", c.AvcBitrate(f),
				"-f", "mp4",
				"-y",
				avcName,
			)
		} else if encoderName == FFmpegAppleEncoder {
			format := "format=yuv420p"

			result = exec.Command(
				c.conf.FFmpegBin(),
				"-i", f.FileName(),
				"-c:v", encoderName,
				"-c:a", "aac",
				"-vf", format,
				"-profile", "high",
				"-level", "51",
				"-vsync", "vfr",
				"-r", "30",
				"-b:v", c.AvcBitrate(f),
				"-f", "mp4",
				"-y",
				avcName,
			)
		} else if encoderName == FFmpegNvidiaEncoder {
			// to show options: ffmpeg -hide_banner -h encoder=h264_nvenc

			result = exec.Command(
				c.conf.FFmpegBin(),
				"-r", "30",
				"-i", f.FileName(),
				"-pix_fmt", "yuv420p",
				"-c:v", encoderName,
				"-c:a", "aac",
				"-preset", "15",
				"-pixel_format", "yuv420p",
				"-gpu", "any",
				"-vf", "format=yuv420p",
				"-rc:v", "constqp",
				"-cq", "0",
				"-tune", "2",
				"-b:v", c.AvcBitrate(f),
				"-profile:v", "1",
				"-level:v", "41",
				"-coder:v", "1",
				"-f", "mp4",
				"-y",
				avcName,
			)
		} else {
			format := "format=yuv420p"

			result = exec.Command(
				c.conf.FFmpegBin(),
				"-i", f.FileName(),
				"-c:v", encoderName,
				"-c:a", "aac",
				"-vf", format,
				"-num_output_buffers", strconv.Itoa(c.conf.FFmpegBuffers()+8),
				"-num_capture_buffers", strconv.Itoa(c.conf.FFmpegBuffers()),
				"-max_muxing_queue_size", "1024",
				"-crf", "23",
				"-vsync", "vfr",
				"-r", "30",
				"-b:v", c.AvcBitrate(f),
				"-f", "mp4",
				"-y",
				avcName,
			)
		}
	} else {
		return nil, useMutex, fmt.Errorf("convert: file type %s not supported in %s", f.FileType(), sanitize.Log(f.BaseName()))
	}

	return result, useMutex, nil
}

// ToAvc converts a single video file to MPEG-4 AVC.
func (c *Convert) ToAvc(f *MediaFile, encoderName string) (file *MediaFile, err error) {
	if n := FFmpegAvcEncoders[encoderName]; n != "" {
		encoderName = n
	} else {
		log.Warnf("convert: unsupported ffmpeg encoder %s", encoderName)
		encoderName = FFmpegSoftwareEncoder
	}

	if f == nil {
		return nil, fmt.Errorf("convert: file is nil - you might have found a bug")
	}

	if !f.Exists() {
		return nil, fmt.Errorf("convert: %s not found", f.RootRelName())
	}

	avcName := fs.FormatAvc.FindFirst(f.FileName(), []string{c.conf.SidecarPath(), fs.HiddenPath}, c.conf.OriginalsPath(), false)

	mediaFile, err := NewMediaFile(avcName)

	if err == nil && mediaFile.IsVideo() {
		return mediaFile, nil
	}

	if !c.conf.SidecarWritable() {
		return nil, fmt.Errorf("convert: transcoding disabled in read only mode (%s)", f.RootRelName())
	}

	if c.conf.DisableFFmpeg() {
		return nil, fmt.Errorf("convert: ffmpeg is disabled for transcoding %s", f.RootRelName())
	}

	fileName := f.RelName(c.conf.OriginalsPath())
	avcName = fs.FileName(f.FileName(), c.conf.SidecarPath(), c.conf.OriginalsPath(), fs.AvcExt)

	cmd, useMutex, err := c.AvcConvertCommand(f, avcName, encoderName)

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

	event.Publish("index.converting", event.Data{
		"fileType": f.FileType(),
		"fileName": fileName,
		"baseName": filepath.Base(fileName),
		"xmpName":  "",
	})

	log.Infof("%s: transcoding %s to %s", encoderName, fileName, fs.FormatAvc)

	// Log exact command for debugging in trace mode.
	log.Trace(cmd.String())

	// Run convert command.
	start := time.Now()
	if err = cmd.Run(); err != nil {
		_ = os.Remove(avcName)

		if stderr.String() != "" {
			err = errors.New(stderr.String())
		}

		// Log ffmpeg output for debugging.
		if err.Error() != "" {
			log.Debug(err)
		}

		// Log filename and transcoding time.
		log.Warnf("%s: failed transcoding %s [%s]", encoderName, fileName, time.Since(start))

		if encoderName != FFmpegSoftwareEncoder {
			return c.ToAvc(f, FFmpegSoftwareEncoder)
		} else {
			return nil, err
		}
	}

	// Log transcoding time.
	log.Infof("%s: created %s [%s]", encoderName, filepath.Base(avcName), time.Since(start))

	return NewMediaFile(avcName)
}
