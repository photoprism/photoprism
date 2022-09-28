package commands

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/service"
)

// CopyCommand registers the copy cli command.
var CopyCommand = cli.Command{
	Name:      "cp",
	Aliases:   []string{"copy"},
	Usage:     "Copies media files to originals",
	ArgsUsage: "[path]",
	Action:    copyAction,
}

// copyAction copies photos to originals path. Default import path is used if no path argument provided
func copyAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	// very if copy directory exist and is writable
	if conf.ReadOnly() {
		return config.ErrReadOnly
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	// get cli first argument
	sourcePath := strings.TrimSpace(ctx.Args().First())

	if sourcePath == "" {
		sourcePath = conf.ImportPath()
	} else {
		abs, err := filepath.Abs(sourcePath)

		if err != nil {
			return err
		}

		sourcePath = abs
	}

	if sourcePath == conf.OriginalsPath() {
		return errors.New("import path is identical with originals")
	}

	log.Infof("copying media files from %s to %s", sourcePath, conf.OriginalsPath())

	w := service.Import()
	opt := photoprism.ImportOptionsCopy(sourcePath)

	w.Start(opt)

	elapsed := time.Since(start)

	log.Infof("import completed in %s", elapsed)

	return nil
}
