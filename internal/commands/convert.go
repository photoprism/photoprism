package commands

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ConvertCommand configures the command name, flags, and action.
var ConvertCommand = cli.Command{
	Name:      "convert",
	Usage:     "Converts files in other formats to JPEG and AVC as needed",
	ArgsUsage: "[subfolder]",
	Flags: []cli.Flag{
		cli.StringSliceFlag{
			Name:  "ext, e",
			Usage: "only process files with the specified extensions, e.g. mp4",
		},
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "replace existing JPEG files in the sidecar folder",
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
	subPath := strings.TrimSpace(ctx.Args().First())

	if subPath != "" {
		convertPath = filepath.Join(convertPath, subPath)
	}

	log.Infof("converting originals in %s", clean.Log(convertPath))

	w := get.Convert()

	// Start file conversion.
	if err := w.Start(convertPath, ctx.StringSlice("ext"), ctx.Bool("force")); err != nil {
		log.Error(err)
	}

	elapsed := time.Since(start)

	log.Infof("completed in %s", elapsed)

	return nil
}
