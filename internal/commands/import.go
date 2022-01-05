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

// ImportCommand registers the import cli command.
var ImportCommand = cli.Command{
	Name:      "mv",
	Aliases:   []string{"import"},
	Usage:     "Moves media files to originals",
	ArgsUsage: "[PATH]",
	Action:    importAction,
}

// importAction moves photos to originals path. Default import path is used if no path argument provided
func importAction(ctx *cli.Context) error {
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

	log.Infof("moving media files from %s to %s", sourcePath, conf.OriginalsPath())

	w := service.Import()
	opt := photoprism.ImportOptionsMove(sourcePath)

	w.Start(opt)

	elapsed := time.Since(start)

	log.Infof("import completed in %s", elapsed)

	conf.Shutdown()

	return nil
}
