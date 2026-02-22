package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// VideoTrimCommand configures the command name, flags, and action.
var VideoTrimCommand = &cli.Command{
	Name:      "trim",
	Usage:     "Trims a duration from the start (positive) or end (negative) of matching videos",
	ArgsUsage: "[filter]... <duration>",
	Flags: []cli.Flag{
		videoCountFlag,
		OffsetFlag,
		DryRunFlag("prints planned trim operations without writing files"),
		YesFlag(),
	},
	Action: videoTrimAction,
}

// videoTrimAction trims matching video files in-place or to sidecar outputs when originals are read-only.
func videoTrimAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if conf.DisableFFmpeg() {
			return fmt.Errorf("ffmpeg is disabled")
		}

		filterArgs, durationArg, err := videoSplitTrimArgs(ctx.Args().Slice())
		if err != nil {
			return cli.Exit(err.Error(), 2)
		}

		trimDuration, err := videoParseTrimDuration(durationArg)
		if err != nil {
			return cli.Exit(err.Error(), 2)
		}

		filter := videoNormalizeFilter(filterArgs)
		results, err := videoSearchResults(filter, ctx.Int(videoCountFlag.Name), ctx.Int(OffsetFlag.Name))
		if err != nil {
			return err
		}

		plans, preflight, err := videoBuildTrimPlans(conf, results, trimDuration)
		if err != nil {
			return err
		}

		if len(plans) == 0 {
			log.Infof("trim: found no matching videos")
			return nil
		}

		if !ctx.Bool("dry-run") {
			if err = videoCheckFreeSpace(preflight); err != nil {
				return err
			}
		}

		if !ctx.Bool("dry-run") && !RunNonInteractively(ctx.Bool("yes")) {
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Trim %d video files?", len(plans)),
				IsConfirm: true,
			}
			if _, err = prompt.Run(); err != nil {
				log.Info("trim: cancelled")
				return nil
			}
		}

		var processed, skipped, failed int
		convert := get.Convert()

		for _, plan := range plans {
			if ctx.Bool("dry-run") {
				log.Infof("trim: would trim %s by %s", clean.Log(plan.IndexPath), trimDuration.String())
				skipped++
				continue
			}

			if err = videoTrimFile(conf, convert, plan, trimDuration, true); err != nil {
				log.Errorf("trim: %s", clean.Error(err))
				failed++
				continue
			}

			processed++
		}

		log.Infof("trim: processed %d, skipped %d, failed %d", processed, skipped, failed)

		if failed > 0 {
			return fmt.Errorf("trim: %d files failed", failed)
		}

		return nil
	})
}

// videoTrimPlan holds a resolved trim operation for a single video file.
type videoTrimPlan struct {
	IndexPath string
	SrcPath   string
	DestPath  string
	Duration  time.Duration
	SizeBytes int64
	Sidecar   bool
}

// videoBuildTrimPlans prepares trim operations and preflight size checks from search results.
func videoBuildTrimPlans(conf *config.Config, results []search.Photo, trimDuration time.Duration) ([]videoTrimPlan, []videoOutputPlan, error) {
	plans := make([]videoTrimPlan, 0, len(results))
	preflight := make([]videoOutputPlan, 0, len(results))

	absTrim := trimDuration
	if absTrim < 0 {
		absTrim = -absTrim
	}

	for _, found := range results {
		videoFile, ok := videoPrimaryFile(found)
		if !ok {
			log.Warnf("trim: missing video file for %s", clean.Log(found.PhotoUID))
			continue
		}

		if videoFile.FileSidecar {
			log.Warnf("trim: skipping sidecar file %s", clean.Log(videoFile.FileName))
			continue
		}

		if videoFile.MediaType == entity.MediaLive {
			log.Warnf("trim: skipping live photo video %s", clean.Log(videoFile.FileName))
			continue
		}

		if videoFile.FileDuration <= 0 {
			log.Warnf("trim: missing duration for %s", clean.Log(videoFile.FileName))
			continue
		}

		remaining := videoFile.FileDuration - absTrim
		if remaining < time.Second {
			log.Errorf("trim: duration exceeds available length for %s", clean.Log(videoFile.FileName))
			continue
		}

		srcPath := photoprism.FileName(videoFile.FileRoot, videoFile.FileName)
		if !fs.FileExistsNotEmpty(srcPath) {
			log.Warnf("trim: missing file %s", clean.Log(srcPath))
			continue
		}

		destPath := srcPath
		useSidecar := false

		if conf.ReadOnly() || !fs.PathWritable(filepath.Dir(srcPath)) || !fs.Writable(srcPath) {
			if !conf.SidecarWritable() || !fs.PathWritable(conf.SidecarPath()) {
				return nil, nil, config.ErrReadOnly
			}

			destPath = videoSidecarPath(srcPath, conf.OriginalsPath(), conf.SidecarPath())
			useSidecar = true
		}

		if useSidecar && fs.FileExistsNotEmpty(destPath) {
			log.Warnf("trim: output already exists %s", clean.Log(destPath))
			continue
		}

		plans = append(plans, videoTrimPlan{
			IndexPath: srcPath,
			SrcPath:   srcPath,
			DestPath:  destPath,
			Duration:  videoFile.FileDuration,
			SizeBytes: videoFile.FileSize,
			Sidecar:   useSidecar,
		})

		preflight = append(preflight, videoOutputPlan{
			Destination: destPath,
			SizeBytes:   videoFile.FileSize,
		})
	}

	return plans, preflight, nil
}

// videoTrimFile executes the trim operation and refreshes previews/thumbnails before reindexing.
func videoTrimFile(conf *config.Config, convert *photoprism.Convert, plan videoTrimPlan, trimDuration time.Duration, noBackup bool) error {
	start := time.Duration(0)
	absTrim := trimDuration
	if absTrim < 0 {
		absTrim = -absTrim
	}

	if trimDuration > 0 {
		start = absTrim
	}

	remaining := plan.Duration - absTrim
	if remaining < time.Second {
		return fmt.Errorf("remaining duration too short for %s", clean.Log(plan.SrcPath))
	}

	destDir := filepath.Dir(plan.DestPath)
	ext := filepath.Ext(plan.DestPath)
	if ext == "" {
		ext = filepath.Ext(plan.SrcPath)
	}
	if ext == "" {
		ext = ".tmp"
	}

	tempPath, err := videoTempPath(destDir, ".trim-*"+ext)
	if err != nil {
		return err
	}

	cmd := videoTrimCmd(conf.FFmpegBin(), plan.SrcPath, tempPath, start, remaining)
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", conf.CmdCachePath()))

	log.Debugf("ffmpeg: %s", clean.Log(cmd.String()))

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed for %s: %s", clean.Log(plan.SrcPath), strings.TrimSpace(stderr.String()))
	}

	if !fs.FileExistsNotEmpty(tempPath) {
		_ = os.Remove(tempPath)
		return fmt.Errorf("trim output missing for %s", clean.Log(plan.SrcPath))
	}

	if err = os.Chmod(tempPath, fs.ModeFile); err != nil {
		return err
	}

	if plan.Sidecar {
		if fs.FileExists(plan.DestPath) {
			_ = os.Remove(tempPath)
			return fmt.Errorf("output already exists %s", clean.Log(plan.DestPath))
		}

		if err = os.Rename(tempPath, plan.DestPath); err != nil {
			_ = os.Remove(tempPath)
			return err
		}
	} else {
		if noBackup {
			_ = os.Remove(plan.DestPath)
		} else {
			backupPath := plan.DestPath + ".backup"
			if fs.FileExists(backupPath) {
				_ = os.Remove(backupPath)
			}
			if err = os.Rename(plan.DestPath, backupPath); err != nil {
				_ = os.Remove(tempPath)
				return err
			}
			_ = os.Chmod(backupPath, fs.ModeBackupFile)
		}

		if err = os.Rename(tempPath, plan.DestPath); err != nil {
			_ = os.Remove(tempPath)
			return err
		}
	}

	mediaFile, err := photoprism.NewMediaFile(plan.DestPath)
	if err != nil {
		return err
	}

	if convert != nil {
		if img, imgErr := convert.ToImage(mediaFile, true); imgErr != nil {
			log.Warnf("trim: %s", clean.Error(imgErr))
		} else if img != nil {
			if thumbsErr := img.GenerateThumbnails(conf.ThumbCachePath(), true); thumbsErr != nil {
				log.Warnf("trim: %s", clean.Error(thumbsErr))
			}
		}
	}

	return videoReindexRelated(conf, plan.IndexPath)
}

// videoTrimCmd builds an ffmpeg command that trims a source file with stream copy.
func videoTrimCmd(ffmpegBin, srcName, destName string, start, duration time.Duration) *exec.Cmd {
	args := []string{
		"-hide_banner",
		"-y",
	}

	if start > 0 {
		args = append(args, "-ss", videoFFmpegSeconds(start))
	}

	args = append(args,
		"-i", srcName,
		"-t", videoFFmpegSeconds(duration),
		"-map", "0",
		"-dn",
		"-ignore_unknown",
		"-codec", "copy",
		"-avoid_negative_ts", "make_zero",
	)

	if videoTrimFastStart(destName) {
		args = append(args, "-movflags", "+faststart")
	}

	args = append(args, destName)

	// #nosec G204 -- arguments are built from validated inputs and config.
	return exec.Command(ffmpegBin, args...)
}

// videoTrimFastStart reports whether the trim output should enable faststart for MP4/MOV containers.
func videoTrimFastStart(destName string) bool {
	switch strings.ToLower(filepath.Ext(destName)) {
	case fs.ExtMp4, fs.ExtMov, fs.ExtQT, ".m4v":
		return true
	default:
		return false
	}
}
