package commands

import (
	"fmt"
	"os"

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

// VideoTranscodeCommand configures the command name, flags, and action.
var VideoTranscodeCommand = &cli.Command{
	Name:      "transcode",
	Usage:     "Transcodes matching videos to AVC sidecar files",
	ArgsUsage: "[filter]...",
	Flags: []cli.Flag{
		videoCountFlag,
		OffsetFlag,
		videoForceFlag,
		DryRunFlag("prints planned transcode operations without writing files"),
		YesFlag(),
	},
	Action: videoTranscodeAction,
}

// videoTranscodeAction transcodes matching videos into sidecar AVC files.
func videoTranscodeAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if conf.DisableFFmpeg() {
			return fmt.Errorf("ffmpeg is disabled")
		}

		filter := videoNormalizeFilter(ctx.Args().Slice())
		results, err := videoSearchResults(filter, ctx.Int(videoCountFlag.Name), ctx.Int(OffsetFlag.Name))
		if err != nil {
			return err
		}

		plans, preflight, err := videoBuildTranscodePlans(conf, results, ctx.Bool(videoForceFlag.Name))
		if err != nil {
			return err
		}

		if len(plans) == 0 {
			log.Infof("transcode: found no matching videos")
			return nil
		}

		if !ctx.Bool("dry-run") {
			if err = videoCheckFreeSpace(preflight); err != nil {
				return err
			}
		}

		if !ctx.Bool("dry-run") && !RunNonInteractively(ctx.Bool("yes")) {
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Transcode %d video files?", len(plans)),
				IsConfirm: true,
			}
			if _, err = prompt.Run(); err != nil {
				log.Info("transcode: cancelled")
				return nil
			}
		}

		var processed, skipped, failed int
		convert := get.Convert()

		for _, plan := range plans {
			if ctx.Bool("dry-run") {
				log.Infof("transcode: would transcode %s to %s", clean.Log(plan.SrcPath), clean.Log(plan.DestPath))
				skipped++
				continue
			}

			file, err := videoTranscodeFile(conf, convert, plan, ctx.Bool(videoForceFlag.Name))
			if err != nil {
				log.Errorf("transcode: %s", clean.Error(err))
				failed++
				continue
			}

			if file != nil {
				if chmodErr := os.Chmod(file.FileName(), fs.ModeFile); chmodErr != nil {
					log.Warnf("transcode: %s", clean.Error(chmodErr))
				}
			}

			if err = videoReindexRelated(conf, plan.IndexPath); err != nil {
				log.Errorf("transcode: %s", clean.Error(err))
				failed++
				continue
			}

			processed++
		}

		log.Infof("transcode: processed %d, skipped %d, failed %d", processed, skipped, failed)

		if failed > 0 {
			return fmt.Errorf("transcode: %d files failed", failed)
		}

		return nil
	})
}

// videoTranscodePlan holds a resolved transcode operation for a single video file.
type videoTranscodePlan struct {
	IndexPath string
	SrcPath   string
	DestPath  string
	SizeBytes int64
}

// videoBuildTranscodePlans prepares transcode operations and preflight size checks from search results.
func videoBuildTranscodePlans(conf *config.Config, results []search.Photo, force bool) ([]videoTranscodePlan, []videoOutputPlan, error) {
	plans := make([]videoTranscodePlan, 0, len(results))
	preflight := make([]videoOutputPlan, 0, len(results))

	for _, found := range results {
		videoFile, ok := videoPrimaryFile(found)
		if !ok {
			log.Warnf("transcode: missing video file for %s", clean.Log(found.PhotoUID))
			continue
		}

		if videoFile.FileSidecar {
			log.Warnf("transcode: skipping sidecar file %s", clean.Log(videoFile.FileName))
			continue
		}

		if videoFile.MediaType == entity.MediaLive {
			log.Warnf("transcode: skipping live photo video %s", clean.Log(videoFile.FileName))
			continue
		}

		srcPath := photoprism.FileName(videoFile.FileRoot, videoFile.FileName)
		if !fs.FileExistsNotEmpty(srcPath) {
			log.Warnf("transcode: missing file %s", clean.Log(srcPath))
			continue
		}

		if !conf.SidecarWritable() || !fs.PathWritable(conf.SidecarPath()) {
			return nil, nil, config.ErrReadOnly
		}

		destPath, err := videoTranscodeTarget(conf, srcPath)
		if err != nil {
			log.Warnf("transcode: %s", clean.Error(err))
			continue
		}

		if destPath == srcPath {
			log.Warnf("transcode: skipping because output equals source %s", clean.Log(srcPath))
			continue
		}

		if fs.FileExistsNotEmpty(destPath) && !force {
			log.Warnf("transcode: output already exists %s", clean.Log(destPath))
			continue
		}

		plans = append(plans, videoTranscodePlan{
			IndexPath: srcPath,
			SrcPath:   srcPath,
			DestPath:  destPath,
			SizeBytes: videoFile.FileSize,
		})

		preflight = append(preflight, videoOutputPlan{
			Destination: destPath,
			SizeBytes:   videoFile.FileSize,
		})
	}

	return plans, preflight, nil
}

// videoTranscodeTarget computes the sidecar output path for an AVC transcode.
func videoTranscodeTarget(conf *config.Config, srcPath string) (string, error) {
	mediaFile, err := photoprism.NewMediaFile(srcPath)
	if err != nil {
		return "", err
	}

	base := videoSidecarPath(srcPath, conf.OriginalsPath(), conf.SidecarPath())
	if mediaFile.IsAnimatedImage() {
		return fs.StripKnownExt(base) + fs.ExtMp4, nil
	}

	return fs.StripKnownExt(base) + fs.ExtAvc, nil
}

// videoTranscodeFile runs the transcode operation and returns the resulting media file.
func videoTranscodeFile(conf *config.Config, convert *photoprism.Convert, plan videoTranscodePlan, force bool) (*photoprism.MediaFile, error) {
	if convert == nil {
		return nil, fmt.Errorf("transcode: convert service unavailable")
	}

	mediaFile, err := photoprism.NewMediaFile(plan.SrcPath)
	if err != nil {
		return nil, err
	}

	return convert.ToAvc(mediaFile, conf.FFmpegEncoder(), false, force)
}
