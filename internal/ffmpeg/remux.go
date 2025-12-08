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
)

// RemuxFile changes the file format to the specified container as needed.
func RemuxFile(videoFilePath, destFilePath string, opt encode.Options) error {
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

	videoBaseName := filepath.Base(videoFilePath)

	if destFilePath == "" {
		destFilePath = fs.StripKnownExt(videoFilePath) + opt.Container.DefaultExt()
	}

	destFileBase := filepath.Base(destFilePath)
	destPathName := filepath.Dir(destFilePath)

	tempBaseName := "." + fs.StripKnownExt(clean.FileName(videoBaseName)) + opt.Container.DefaultExt()
	tempFilePath := filepath.Join(destPathName, tempBaseName)

	cmd, err := RemuxCmd(videoFilePath, tempFilePath, opt)

	// Return if an error occurred.
	if err != nil {
		log.Error(err)
		return err
	}

	// Check if target file already exists.
	if fs.FileExists(tempFilePath) {
		if !opt.Force {
			return fmt.Errorf("temp file %s already exists", clean.Log(tempBaseName))
		} else if err = os.Remove(tempFilePath); err != nil {
			return fmt.Errorf("%s (remove temp file)", err)
		}

		log.Infof("ffmpeg: replacing temp file %s", clean.Log(tempBaseName))
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
		log.Warnf("ffmpeg: failed to convert %s [%s]", clean.Log(videoBaseName), time.Since(start))

		// Remove broken video file.
		if !fs.FileExists(tempFilePath) {
			// Do nothing.
		} else if err = os.Remove(tempFilePath); err != nil {
			return fmt.Errorf("failed to remove temp file %s (%s)", clean.Log(tempBaseName), err)
		}

		return err
	}

	// Abort if destination file is missing or empty.
	if !fs.FileExistsNotEmpty(tempFilePath) {
		_ = os.Remove(tempFilePath)
		return fmt.Errorf("failed change container format of %s [%s]", clean.Log(videoBaseName), time.Since(start))
	}

	if !fs.FileExists(destFilePath) {
		// Do nothing.
	} else if err = os.Remove(destFilePath); err != nil {
		_ = os.Remove(tempFilePath)
		return fmt.Errorf("failed to remove %s (%s)", clean.Log(destFileBase), err)
	}

	if err = os.Rename(tempFilePath, destFilePath); err != nil {
		return fmt.Errorf("failed to rename %s to %s (%s)", clean.Log(tempBaseName), clean.Log(destFileBase), err)
	}

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
