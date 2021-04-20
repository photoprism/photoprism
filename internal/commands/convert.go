package commands

import (
	"context"
	"github.com/photoprism/photoprism/pkg/txt"
	"path/filepath"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/urfave/cli"
)

// ConvertCommand registers the convert cli command.
var ConvertCommand = cli.Command{
	Name:      "convert",
	Usage:     "Converts originals in other formats to JPEG and AVC sidecar files",
	UsageText: `To limit scope, a sub folder may be passed as first argument.`,
	Action:    convertAction,
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

	log.Infof("converting originals in %s", txt.Quote(convertPath))

	convert := service.Convert()

	if err := convert.Start(convertPath); err != nil {
		log.Error(err)
	}

	elapsed := time.Since(start)

	log.Infof("converting completed in %s", elapsed)

	return nil
}
