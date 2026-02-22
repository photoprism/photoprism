package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/video"
)

// VideoRemuxCommand configures the command name, flags, and action.
var VideoRemuxCommand = &cli.Command{
	Name:      "remux",
	Usage:     "Remuxes AVC videos into an MP4 container",
	ArgsUsage: "[filter]...",
	Flags: []cli.Flag{
		videoCountFlag,
		OffsetFlag,
		videoForceFlag,
		DryRunFlag("prints planned remux operations without writing files"),
		YesFlag(),
	},
	Action: videoRemuxAction,
}

// videoRemuxAction remuxes matching AVC files into MP4 containers.
func videoRemuxAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if conf.DisableFFmpeg() {
			return fmt.Errorf("ffmpeg is disabled")
		}

		filter := videoNormalizeFilter(ctx.Args().Slice())
		results, err := videoSearchResults(filter, ctx.Int(videoCountFlag.Name), ctx.Int(OffsetFlag.Name))
		if err != nil {
			return err
		}

		plans, preflight, err := videoBuildRemuxPlans(conf, results, ctx.Bool(videoForceFlag.Name))
		if err != nil {
			return err
		}

		if len(plans) == 0 {
			log.Infof("remux: found no matching videos")
			return nil
		}

		if !ctx.Bool("dry-run") {
			if err = videoCheckFreeSpace(preflight); err != nil {
				return err
			}
		}

		if !ctx.Bool("dry-run") && !RunNonInteractively(ctx.Bool("yes")) {
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Remux %d video files?", len(plans)),
				IsConfirm: true,
			}
			if _, err = prompt.Run(); err != nil {
				log.Info("remux: cancelled")
				return nil
			}
		}

		var processed, skipped, failed int
		convert := get.Convert()

		for _, plan := range plans {
			if ctx.Bool("dry-run") {
				log.Infof("remux: would remux %s to %s", clean.Log(plan.SrcPath), clean.Log(plan.DestPath))
				skipped++
				continue
			}

			if err = videoRemuxFile(conf, convert, plan, ctx.Bool(videoForceFlag.Name), true); err != nil {
				log.Errorf("remux: %s", clean.Error(err))
				failed++
				continue
			}

			processed++
		}

		log.Infof("remux: processed %d, skipped %d, failed %d", processed, skipped, failed)

		if failed > 0 {
			return fmt.Errorf("remux: %d files failed", failed)
		}

		return nil
	})
}

// videoRemuxPlan holds a resolved remux operation for a single video file.
type videoRemuxPlan struct {
	IndexPath string
	SrcPath   string
	DestPath  string
	SizeBytes int64
	Sidecar   bool
}

// videoBuildRemuxPlans prepares remux operations and preflight size checks from search results.
func videoBuildRemuxPlans(conf *config.Config, results []search.Photo, force bool) ([]videoRemuxPlan, []videoOutputPlan, error) {
	plans := make([]videoRemuxPlan, 0, len(results))
	preflight := make([]videoOutputPlan, 0, len(results))

	for _, found := range results {
		videoFile, ok := videoPrimaryFile(found)
		if !ok {
			log.Warnf("remux: missing video file for %s", clean.Log(found.PhotoUID))
			continue
		}

		if videoFile.FileSidecar {
			log.Warnf("remux: skipping sidecar file %s", clean.Log(videoFile.FileName))
			continue
		}

		if videoFile.MediaType == entity.MediaLive {
			log.Warnf("remux: skipping live photo video %s", clean.Log(videoFile.FileName))
			continue
		}

		srcPath := photoprism.FileName(videoFile.FileRoot, videoFile.FileName)
		if !fs.FileExistsNotEmpty(srcPath) {
			log.Warnf("remux: missing file %s", clean.Log(srcPath))
			continue
		}

		if !videoCodecIsAvc(videoFile.FileCodec) && !force {
			if !videoFallbackCodecAvc(srcPath) {
				log.Warnf("remux: skipping non-AVC video %s", clean.Log(videoFile.FileName))
				continue
			}
		}

		destPath := fs.StripKnownExt(srcPath) + fs.ExtMp4
		useSidecar := false
		indexPath := destPath

		if conf.ReadOnly() || !fs.PathWritable(filepath.Dir(srcPath)) || !fs.Writable(srcPath) {
			if !conf.SidecarWritable() || !fs.PathWritable(conf.SidecarPath()) {
				return nil, nil, config.ErrReadOnly
			}

			sidecarBase := videoSidecarPath(srcPath, conf.OriginalsPath(), conf.SidecarPath())
			destPath = fs.StripKnownExt(sidecarBase) + fs.ExtMp4
			useSidecar = true
			indexPath = srcPath
		}

		if destPath != srcPath && fs.FileExistsNotEmpty(destPath) && !force {
			log.Warnf("remux: output already exists %s", clean.Log(destPath))
			continue
		}

		plans = append(plans, videoRemuxPlan{
			IndexPath: indexPath,
			SrcPath:   srcPath,
			DestPath:  destPath,
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

// videoRemuxFile runs ffmpeg remuxing and refreshes previews/thumbnails before reindexing.
func videoRemuxFile(conf *config.Config, convert *photoprism.Convert, plan videoRemuxPlan, force, noBackup bool) error {
	tempDir := filepath.Dir(plan.DestPath)
	tempPath, err := videoTempPath(tempDir, ".remux-*.mp4")
	if err != nil {
		return err
	}

	opt := encode.NewRemuxOptions(conf.FFmpegBin(), fs.VideoMp4, true)
	opt.Force = true

	if err = ffmpeg.RemuxFile(plan.SrcPath, tempPath, opt); err != nil {
		return err
	}

	if !fs.FileExistsNotEmpty(tempPath) {
		_ = os.Remove(tempPath)
		return fmt.Errorf("remux output missing for %s", clean.Log(plan.SrcPath))
	}

	if err = os.Chmod(tempPath, fs.ModeFile); err != nil {
		return err
	}

	if plan.Sidecar {
		if fs.FileExists(plan.DestPath) && !force {
			_ = os.Remove(tempPath)
			return fmt.Errorf("output already exists %s", clean.Log(plan.DestPath))
		}

		if fs.FileExists(plan.DestPath) {
			_ = os.Remove(plan.DestPath)
		}

		if err = os.Rename(tempPath, plan.DestPath); err != nil {
			_ = os.Remove(tempPath)
			return err
		}
	} else {
		if plan.DestPath != plan.SrcPath && fs.FileExists(plan.DestPath) && !force {
			_ = os.Remove(tempPath)
			return fmt.Errorf("output already exists %s", clean.Log(plan.DestPath))
		}

		if noBackup {
			if plan.DestPath != plan.SrcPath {
				_ = os.Remove(plan.DestPath)
			}
		} else {
			backupPath := plan.SrcPath + ".backup"
			if fs.FileExists(backupPath) {
				_ = os.Remove(backupPath)
			}
			if err = os.Rename(plan.SrcPath, backupPath); err != nil {
				_ = os.Remove(tempPath)
				return err
			}
			_ = os.Chmod(backupPath, fs.ModeBackupFile)
		}

		if plan.DestPath != plan.SrcPath && fs.FileExists(plan.DestPath) {
			_ = os.Remove(plan.DestPath)
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
			log.Warnf("remux: %s", clean.Error(imgErr))
		} else if img != nil {
			if thumbsErr := img.GenerateThumbnails(conf.ThumbCachePath(), true); thumbsErr != nil {
				log.Warnf("remux: %s", clean.Error(thumbsErr))
			}
		}
	}

	return videoReindexRelated(conf, plan.IndexPath)
}

// videoCodecIsAvc reports whether a codec string maps to an AVC/H.264 variant.
func videoCodecIsAvc(codec string) bool {
	value := strings.ToLower(strings.TrimSpace(codec))
	if value == "" {
		return false
	}

	if value == "h264" || value == "x264" {
		return true
	}

	switch video.Codecs[value] {
	case video.CodecAvc1, video.CodecAvc2, video.CodecAvc3, video.CodecAvc4:
		return true
	default:
		return false
	}
}

// videoFallbackCodecAvc probes codec metadata when the indexed codec is missing.
func videoFallbackCodecAvc(srcPath string) bool {
	mediaFile, err := photoprism.NewMediaFile(srcPath)
	if err != nil {
		return false
	}

	if info := mediaFile.VideoInfo(); info.VideoCodec != "" {
		return videoCodecIsAvc(info.VideoCodec)
	}

	return mediaFile.MetaData().CodecAvc()
}
