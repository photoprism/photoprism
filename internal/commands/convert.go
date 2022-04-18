package commands

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ConvertCommand registers the convert cli command.
var ConvertCommand = cli.Command{
	Name:      "convert",
	Usage:     "Converts files in other formats to JPEG and AVC as needed",
	ArgsUsage: "[originals folder]",
	Flags: []cli.Flag{
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

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	if !conf.SidecarWritable() {
		return config.ErrReadOnly
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	convertPath := conf.OriginalsPath()

	// Use first argument to limit scope if set.
	subPath := strings.TrimSpace(ctx.Args().First())

	if subPath != "" {
		convertPath = filepath.Join(convertPath, subPath)
	}

	log.Infof("converting originals in %s", clean.Log(convertPath))

	w := service.Convert()

	// Start file conversion.
	if err := w.Start(convertPath, ctx.Bool("force")); err != nil {
		log.Error(err)
	}

	elapsed := time.Since(start)

	log.Infof("converting completed in %s", elapsed)

	return nil
}
