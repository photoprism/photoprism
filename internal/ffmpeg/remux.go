package ffmpeg

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/video"
	"github.com/photoprism/photoprism/pkg/proc"
)

// RemuxFile changes the file format to the specified container as needed.
func RemuxFile(videoFilePath, destFilePath string, opt encode.Options) (err error) {
	// Return if destination file already exists and force option is not set.
	if !opt.Force && fs.FileExistsNotEmpty(destFilePath) {
		return nil
	}

	// Error if source file does not exist or is empty.
	if !fs.FileExistsNotEmpty(videoFilePath) {
		return errors.New("invalid video file path")
	}

	// Use MP4 as default container format.
	if opt.Container == "" {
		opt.Container = fs.ExtMp4
	}

	// Normalize HEVC sample-entry tag to "hvc1" when remuxing to MP4/MOV so the
	// resulting file plays on macOS, iOS, and Edge/Chrome on Windows. Without
	// this override, "-codec copy" preserves a source "hev1" tag, which those
	// players reject. Skip when the caller already set VideoTag explicitly.
	if opt.VideoTag == "" && (opt.Container == fs.VideoMp4 || opt.Container == fs.VideoMov) && video.IsHEVCFile(videoFilePath) {
		opt.VideoTag = "hvc1"
	}

	videoBaseName := filepath.Base(videoFilePath)

	if destFilePath == "" {
		destFilePath = fs.StripKnownExt(videoFilePath) + opt.Container.DefaultExt()
	}

	destFileBase := filepath.Base(destFilePath)

	// Write through a sibling reserved by this call, so each conversion owns its working file.
	tempFilePath, err := fs.CreateStageFile(destFilePath)

	if err != nil {
		return fmt.Errorf("failed to create temp file for %s (%s)", clean.Log(destFileBase), err)
	}

	tempBaseName := filepath.Base(tempFilePath)
	published := false

	// Remove only the file this call created, on every way out including a panic. A converter that
	// removed its own output leaves nothing to clean up, so that is not a failure.
	defer func() {
		if published {
			return
		}

		if removeErr := os.Remove(tempFilePath); removeErr != nil && !os.IsNotExist(removeErr) {
			err = errors.Join(err, fmt.Errorf("failed to remove temp file %s (%s)", clean.Log(tempBaseName), clean.Error(removeErr)))
		}
	}()

	cmd, err := RemuxCmd(videoFilePath, tempFilePath, opt)

	// Return if an error occurred.
	if err != nil {
		log.Error(err)
		return err
	}

	// Fetch command output.
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Env = append(cmd.Env, []string{
		fmt.Sprintf("HOME=%s", tempFilePath),
	}...)

	log.Infof("ffmpeg: changing container format of %s to %s", clean.Log(videoBaseName), opt.Container)

	// Log exact command for debugging in trace mode.
	log.Trace(clean.Cmd(cmd))

	// Run the remux command within its budget, terminating the whole process tree if it is
	// exceeded, so a source the muxer cannot finish does not hold the caller indefinitely.
	start := time.Now()
	if err = proc.Run(cmd, opt.Timeout); err != nil {
		if s := stderr.String(); s != "" {
			err = fmt.Errorf("%w: %s", err, s)
		}

		// Log ffmpeg output for debugging.
		if err.Error() != "" {
			log.Debug(err)
		}

		// Log filename and remux time.
		log.Warnf("ffmpeg: failed to remux %s [%s]", clean.Log(videoBaseName), time.Since(start))

		return err
	}

	// Abort if the staged file is missing or empty.
	if !fs.FileExistsNotEmpty(tempFilePath) {
		return fmt.Errorf("failed change container format of %s [%s]", clean.Log(videoBaseName), time.Since(start))
	}

	// Publish by renaming, which replaces a regular destination in a single step.
	if err = os.Rename(tempFilePath, destFilePath); err != nil {
		return fmt.Errorf("failed to rename %s to %s (%s)", clean.Log(tempBaseName), clean.Log(destFileBase), err)
	}

	published = true

	// Log filename and remux time.
	if videoBaseName != destFileBase {
		log.Infof("ffmpeg: converted %s to %s [%s]", clean.Log(videoBaseName), clean.Log(destFileBase), time.Since(start))
	} else {
		log.Infof("ffmpeg: converted %s to %s [%s]", clean.Log(videoBaseName), opt.Container.String(), time.Since(start))
	}

	return nil
}

// RemuxCmd returns the FFmpeg command for transferring content from one container format to another without altering the original video or audio stream.
func RemuxCmd(srcName, destName string, opt encode.Options) (cmd *exec.Cmd, err error) {
	switch {
	case srcName == "":
		return nil, fmt.Errorf("empty source filename")
	case !fs.FileExistsNotEmpty(srcName):
		return nil, fmt.Errorf("source file is empty or missing")
	case destName == "":
		return nil, fmt.Errorf("empty destination filename")
	case srcName == destName:
		return nil, fmt.Errorf("source and destination filenames must be different")
	}

	// Use the default binary name if no name is specified.
	if opt.Bin == "" {
		opt.Bin = encode.FFmpegBin
	}

	// Compose "ffmpeg" command flags, see https://ffmpeg.org/ffmpeg-formats.html#Format-Options:
	flags := []string{
		"-hide_banner",
		"-y",
		"-strict", "-2",
		// The "-avoid_negative_ts" flag is commonly used for remuxing, but may cause desync (please report any issues):
		"-avoid_negative_ts", "make_zero",
		"-i", srcName,
		"-map", opt.MapVideo,
		"-map", opt.MapAudio,
		// The "-dn" flag removes data streams, such as subtitles, timecode tracks, and camera motion data:
		"-dn",
		"-ignore_unknown",
		"-codec", "copy",
		"-f", opt.Container.String(),
	}

	// Override the output video sample-entry tag when requested (e.g. "hvc1" for
	// HEVC in MP4/MOV containers). Applied before the container-specific block
	// so it covers both MP4 and MOV outputs.
	if opt.VideoTag != "" {
		flags = append(flags, "-tag:v", opt.VideoTag)
	}

	// Append format specific "ffmpeg" command flags.
	if opt.Container == fs.VideoMp4 {
		// Ensure MP4 compatibility:
		flags = append(flags,
			"-movflags", opt.MovFlags,
			"-map_metadata", opt.MapMetadata, // Copy existing video metadata.
		)

		// If specified, add the following metadata:
		if title := clean.Name(opt.Title); title != "" {
			flags = append(flags, "-metadata", fmt.Sprintf(`title=%s`, title))
		}

		if desc := strings.TrimSpace(opt.Description); desc != "" {
			flags = append(flags, "-metadata", fmt.Sprintf(`description=%s`, desc))
		}

		if comment := strings.TrimSpace(opt.Comment); comment != "" {
			flags = append(flags, "-metadata", fmt.Sprintf(`comment=%s`, comment))
		}

		if author := clean.Name(opt.Author); author != "" {
			flags = append(flags, "-metadata", fmt.Sprintf(`author=%s`, author))
		}

		if !opt.Created.IsZero() {
			flags = append(flags, "-metadata", fmt.Sprintf(`creation_time=%s`, opt.Created.Format(time.DateTime)))
		}
	}

	// Set the destination file name as the last command flag.
	flags = append(flags, destName)

	// #nosec G204 -- filenames and flags are constructed internally and not user-controlled.
	cmd = exec.Command(
		opt.Bin,
		flags...,
	)

	return cmd, nil
}
