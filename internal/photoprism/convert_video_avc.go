package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/media/projection"
	"github.com/photoprism/photoprism/pkg/proc"
)

// ToAvc converts a single video file to MPEG-4 AVC.
func (w *Convert) ToAvc(f *MediaFile, encoder encode.Encoder, noMutex, force bool) (file *MediaFile, err error) {
	// Abort if the source media file is nil.
	if f == nil {
		return nil, fmt.Errorf("convert: no media file provided for processing - you may have found a bug")
	}

	// Normalize every member of a complete Insta360 capture to its canonical left lens so manual
	// conversion, background conversion, and playback all reuse one equirectangular AVC sidecar.
	if capture := FindInsta360Capture(f); capture != nil && capture.ValidPair() {
		f = capture.Left
	}

	// Sanitized relative filename for use in logs.
	logFileName := clean.Log(f.RootRelName())

	// Abort if the source media file does not exist.
	if !f.Exists() {
		return nil, fmt.Errorf("convert: %s not found", logFileName)
	} else if f.Empty() {
		return nil, fmt.Errorf("convert: %s is empty", logFileName)
	}

	// Skip files whose codec or container is on the FFmpeg exclude list.
	if !w.FFmpegAllowed(f) {
		format := clean.Log(w.ffmpegExclude.Match(f.MetaData().Codec, f.VideoInfo().VideoCodec, f.FileType().String()))
		log.Warnf("convert: skipping %s because format %s is on the FFmpeg exclude list", logFileName, format)
		return nil, fmt.Errorf("convert: format %s is excluded from FFmpeg processing", format)
	}

	// Convert MPEG-2 Transport Stream (M2TS) files to MPEG4 containers. Neither ExifTool nor the
	// video probe reports a codec for a transport stream, so whether one carries AVC is only known
	// after the conversion and cannot be decided from the source up front.
	if !f.IsAnimatedImage() && f.IsM2TS() && w.conf.SidecarWritable() && !w.conf.InsufficientStorage() {
		mp4File, mp4Err := w.avcFromM2TS(f, logFileName)

		if mp4Err != nil {
			return nil, mp4Err
		} else if mp4File != nil {
			return mp4File, nil
		}
	}

	// AVC video filename. Animated images are converted into an MPEG-4 container, videos into AVC.
	var avcName string

	if f.IsAnimatedImage() {
		avcName = fs.VideoMp4.FindFirst(f.FileName(), []string{w.conf.SidecarPath(), fs.PPHiddenPathname}, w.conf.OriginalsPath(), false)
	} else {
		avcName = fs.VideoAvc.FindFirst(f.FileName(), []string{w.conf.SidecarPath(), fs.PPHiddenPathname}, w.conf.OriginalsPath(), false)
	}

	mediaFile, err := NewMediaFile(avcName)

	// Return the AVC-encoded video file if it already exists.
	if mediaFile == nil || err != nil {
		// Do nothing.
	} else if mediaFile.IsVideo() && (!force || !mediaFile.InSidecar()) {
		// Return existing AVC file.
		log.Debugf("convert: %s has already been transcoded to MPEG-4 AVC", logFileName)
		return mediaFile, nil
	}

	// Check if the sidecar path is writable, otherwise no new AVC file can be created.
	if !w.conf.SidecarWritable() {
		return nil, fmt.Errorf("convert: cannot transcode %s because the sidecar path is not writable", logFileName)
	} else if w.conf.InsufficientStorage() {
		return nil, status.ErrInsufficientStorage
	}

	// Get relative filename for logging.
	relName := f.RelName(w.conf.OriginalsPath())

	// Use .mp4 file extension for animated images and .avi for videos.
	if f.IsAnimatedImage() {
		avcName, _ = fs.FileName(f.FileName(), w.conf.SidecarPath(), w.conf.OriginalsPath(), fs.ExtMp4)
	} else {
		avcName, _ = fs.FileName(f.FileName(), w.conf.SidecarPath(), w.conf.OriginalsPath(), fs.ExtAvc)
	}

	cmd, useMutex, err := w.TranscodeToAvcCmd(f, avcName, encoder)

	// Return if an error occurred.
	if err != nil {
		log.Errorf("convert: %s for %s (transcode command)", clean.Error(err), logFileName)
		return nil, err
	}

	// Make sure only one convert command runs at a time.
	if useMutex && !noMutex {
		w.cmdMutex.Lock()
		defer w.cmdMutex.Unlock()
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
	cmd.Env = append(cmd.Env, []string{
		fmt.Sprintf("HOME=%s", w.conf.CmdCachePath()),
	}...)

	event.Publish("index.converting", event.Data{
		"fileType": f.FileType(),
		"fileName": relName,
		"baseName": filepath.Base(relName),
		"xmpName":  "",
	})

	log.Infof("%s: transcoding %s to %s", encoder, clean.Log(relName), fs.VideoAvc)

	// Log exact command for debugging in trace mode.
	log.Trace(cmd.String())

	// Transcode source media file to AVC. Transcoding time tracks the length of the source, so
	// it has a budget of its own, which is unset by default. An animated image is converted
	// rather than transcoded, so it is charged the conversion budget instead.
	budget := w.conf.TranscodeTimeout()

	if cmd.Path == w.conf.ImageMagickBin() {
		budget = w.conf.ConvertTimeout()
	}

	start := time.Now()
	if err = proc.Run(cmd, budget); err != nil {
		// Log the transcoder output for debugging, which is console-only.
		if s := strings.TrimSpace(stderr.String()); s != "" {
			log.Debugf("%s: %s for %s", encoder, s, logFileName)
		}

		// Log filename and transcoding time.
		log.Warnf("%s: failed to transcode %s [%s]", encoder, clean.Log(relName), time.Since(start))

		// Remove broken video file.
		if !fs.FileExists(avcName) {
			// Do nothing.
		} else if err = os.Remove(avcName); err != nil {
			return nil, fmt.Errorf("convert: failed to remove %s (%s)", clean.Log(RootRelName(avcName)), err)
		}

		switch {
		case disk.IsNoSpace(err):
			// Do not retry on a full disk; surface the cause so the worker can abort the run.
			return nil, disk.AsInsufficientStorage(err)
		case encoder != encode.SoftwareAvc:
			// Try again using software encoder.
			return w.ToAvc(f, encode.SoftwareAvc, true, false)
		default:
			return nil, err
		}
	}

	// Log filename and transcoding time.
	log.Infof("%s: created %s [%s]", encoder, filepath.Base(avcName), time.Since(start))

	// Return AVC media file and keep the successful dewarp projection available to the indexer even
	// when ExifTool is disabled. Later reindexes infer the same value from source and sidecar paths.
	avcFile, avcErr := NewMediaFile(avcName)
	if avcErr == nil && f.DewarpableInsv() {
		avcFile.SetVisualProjection(projection.Equirectangular)
	}

	return avcFile, avcErr
}

// avcFromM2TS returns the MPEG-4 container of a transport stream that carries AVC, and nil when the
// source has to be transcoded instead. The container is written under a name this call reserved and
// published only when it is kept, so one already in place is never rewritten or removed.
func (w *Convert) avcFromM2TS(f *MediaFile, logFileName string) (*MediaFile, error) {
	mp4Name, err := fs.FileName(f.FileName(), w.conf.SidecarPath(), w.conf.OriginalsPath(), fs.ExtMp4)

	if err != nil {
		return nil, fmt.Errorf("convert: %s in %s (remux)", err, clean.Log(f.RootRelName()))
	}

	// Reuse a container that is already in place. One holding no playable AVC belongs to whoever
	// wrote it and stays where it is.
	if fs.FileExistsNotEmpty(mp4Name) {
		return w.avcContainer(mp4Name, logFileName), nil
	}

	// Skip the conversion once a transcoding result is already in place, so a source that never
	// yields a playable container is converted once rather than on every request. The search is the
	// one the caller reuses from, and only a regular file counts, so a name that merely resolves to
	// one does not stand in for it.
	avcName := fs.VideoAvc.FindFirst(f.FileName(), []string{w.conf.SidecarPath(), fs.PPHiddenPathname}, w.conf.OriginalsPath(), false)

	if avcName != "" && fs.FileExistsNotEmpty(avcName) && !fs.IsSymlink(avcName) {
		return nil, nil
	}

	stagedName, err := fs.CreateStageFile(mp4Name)

	if err != nil {
		return nil, fmt.Errorf("convert: %s in %s (remux)", err, clean.Log(f.RootRelName()))
	}

	// Remove only the file this call created, on every way out including a panic.
	defer func() {
		if removeErr := os.Remove(stagedName); removeErr != nil && !os.IsNotExist(removeErr) {
			log.Warnf("convert: %s in %s (remove unused mp4)", clean.Error(removeErr), logFileName)
		}
	}()

	if err = ffmpeg.RemuxFile(f.FileName(), stagedName, w.RemuxOptions(fs.VideoMp4, true)); err != nil {
		return nil, fmt.Errorf("convert: %s in %s (remux)", err, clean.Log(f.RootRelName()))
	}

	// A container without playable AVC is never published: the source is transcoded instead, and
	// a transport stream carrying MPEG-2 rather than AVC (e.g. a JVC .tod) always takes this path.
	if w.avcContainer(stagedName, logFileName) == nil {
		return nil, nil
	}

	// A link fails when the name is taken, so a container that appeared while this one was written
	// is used rather than replaced.
	if err = fs.PublishFile(stagedName, mp4Name, false); errors.Is(err, os.ErrExist) {
		return w.avcContainer(mp4Name, logFileName), nil
	} else if err != nil {
		return nil, fmt.Errorf("convert: %s in %s (publish mp4)", clean.Error(err), logFileName)
	}

	mp4File, err := NewMediaFile(mp4Name)

	if err != nil {
		return nil, fmt.Errorf("convert: %s in %s (published mp4)", clean.Error(err), logFileName)
	}

	return mp4File, nil
}

// avcContainer returns the video file when it holds a playable AVC stream, and nil otherwise, since
// whether a converted transport stream carries AVC is only visible in its metadata.
func (w *Convert) avcContainer(fileName, logFileName string) *MediaFile {
	mp4File, err := NewMediaFile(fileName)

	if mp4File == nil || err != nil {
		log.Warnf("convert: %s could not be converted to mp4", logFileName)
	} else if jsonErr := mp4File.CreateExifToolJson(w); jsonErr != nil {
		log.Warnf("convert: %s in %s (create json)", jsonErr, logFileName)
	} else if jsonErr = mp4File.ReadExifToolJson(); jsonErr != nil {
		log.Warnf("convert: %s in %s (read json)", jsonErr, logFileName)
	} else if mp4File.MetaData().CodecAvc() {
		return mp4File
	}

	return nil
}

// TranscodeToAvcCmd returns the command for converting video files to MPEG-4 AVC.
func (w *Convert) TranscodeToAvcCmd(f *MediaFile, avcName string, encoder encode.Encoder) (result *exec.Cmd, useMutex bool, err error) {
	fileExt := f.Extension()
	fileName := f.FileName()

	switch {
	case fileName == "":
		return nil, false, fmt.Errorf("convert: %s video filename is empty - you may have found a bug", f.FileType())
	case !f.IsAnimated():
		return nil, false, fmt.Errorf("convert: file type %s of %s cannot be transcoded", f.FileType(), clean.Log(f.BaseName()))
	}

	// Try to transcode animated WebP images with ImageMagick.
	if w.conf.ImageMagickEnabled() && f.IsWebp() && w.imageMagickExclude.Allow(fileExt) {
		// #nosec G204 -- arguments are built from validated config and file paths.
		return exec.Command(w.conf.ImageMagickBin(), f.FileName(), avcName), false, nil
	}

	// Complete separate-lens captures are combined before dewarping. Single-file INSV originals are
	// dewarped only when their decoded frame is already a side-by-side ~2:1 dual-fisheye layout.
	capture := FindInsta360Capture(f)
	dewarpPair := capture != nil && capture.ValidPair() && capture.Left.FileName() == f.FileName()
	dewarp := dewarpPair || f.IsInsv() && f.DualFisheyeLayout()

	if dewarp {
		encoder = encode.SoftwareAvc
	}

	// Use FFmpeg to transcode all other media files to AVC.
	var opt encode.Options
	if opt, err = w.conf.FFmpegOptions(encoder, w.AvcBitrate(f)); err != nil {
		return nil, false, fmt.Errorf("convert: failed to transcode %s (%s)", clean.Log(f.BaseName()), err)
	}

	if dewarp {
		opt.V360 = ffmpeg.V360DualFisheyeToEquirect(w.fisheyeFov(f), w.fisheyeRoll(f))
	}

	if dewarpPair {
		return ffmpeg.DewarpDualFisheyePairToAvcCmd(capture.Left.FileName(), capture.Right.FileName(), avcName, opt), true, nil
	}

	return ffmpeg.TranscodeCmd(fileName, avcName, opt)
}

// fisheyeFov returns the v360 dewarp field of view in degrees for the given fisheye 360° file,
// preferring a per-camera default and falling back to the configured FFmpegFisheyeFov.
func (w *Convert) fisheyeFov(f *MediaFile) int {
	if f != nil {
		model := f.CameraModel()

		if model == "" && f.DualFisheye() {
			model = f.Insta360CameraModel()
		}

		if fov := entity.CameraFisheyeFov(f.CameraMake(), model); fov > 0 {
			return fov
		}
	}

	return w.conf.FFmpegFisheyeFov()
}

// fisheyeRoll returns a verified spherical roll correction for a compatible Insta360 original.
func (w *Convert) fisheyeRoll(f *MediaFile) int {
	if f == nil || !f.DualFisheye() && !f.FisheyeDng() {
		return 0
	}

	if capture := FindInsta360Capture(f); capture != nil && capture.ValidPair() {
		f = capture.Left
	} else if f.DualFisheye() && !f.DualFisheyeLayout() {
		return 0
	}

	model := f.Insta360CameraModel()
	if f.FisheyeDng() {
		model = f.CameraModel()
	}
	roll := entity.CameraFisheyeRoll(f.CameraMake(), model)

	if roll != 0 {
		log.Debugf("convert: using v360 profile insta360-one-rs (roll %d) for %s", roll, clean.Log(f.BaseName()))
	}

	return roll
}

// AvcBitrate returns the ideal AVC encoding bitrate in megabits per second.
func (w *Convert) AvcBitrate(f *MediaFile) string {
	const defaultBitrate = "8M"

	if f == nil {
		return defaultBitrate
	}

	limit := w.conf.FFmpegBitrate()
	quality := 12

	bitrate := int(math.Ceil(float64(f.Width()*f.Height()*quality) / 1000000))

	if bitrate <= 0 {
		return defaultBitrate
	} else if bitrate > limit {
		bitrate = limit
	}

	return fmt.Sprintf("%dM", bitrate)
}
