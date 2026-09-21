package commands

import (
	"context"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
)

const convertDescription = `Missing preview images and AVC videos are created for the originals in the specified
subfolder, or for all originals if none is given. A preview is a JPEG, or a PNG when the original is a
vector graphic.

The --force flag replaces an existing preview image, and only one that is in the sidecar folder. Videos
that already have an AVC sidecar are skipped either way, since transcoding them costs far more than
rendering a still image again. Use "photoprism videos transcode" to transcode those.`

// ConvertCommand configures the command name, flags, and action.
var ConvertCommand = &cli.Command{
	Name:        "convert",
	Description: convertDescription,
	Usage:       "Creates missing preview images and AVC sidecar files as needed",
	ArgsUsage:   "[subfolder]",
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "ext",
			Aliases: []string{"e"},
			Usage:   "only process files with the specified extensions, e.g. mp4",
		},
		&cli.BoolFlag{
			Name:    "force",
			Aliases: []string{"f"},
			Usage:   "replace existing preview images in the sidecar folder",
		},
	},
	Action: convertAction,
}

// convertAction converts originals in other formats to JPEG and AVC sidecar files.
func convertAction(ctx *cli.Context) error {
	start := time.Now()

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	if !conf.SidecarWritable() {
		return config.ErrReadOnly
	}

	defer conf.Shutdown()

	convertPath := conf.OriginalsPath()

	// Use first argument to limit scope if set.
	subPath, err := sanitizeSubfolderArg(ctx.Args().First())

	if err != nil {
		return err
	}

	if subPath != "" {
		convertPath = filepath.Join(convertPath, subPath)
	}

	log.Infof("converting originals in %s", clean.Log(convertPath))

	w := get.Convert()

	// Start file conversion.
	if err = w.Start(convertPath, ctx.StringSlice("ext"), ctx.Bool("force")); err != nil {
		log.Error(err)
	}

	elapsed := time.Since(start)

	log.Infof("completed in %s", elapsed)

	return nil
}
