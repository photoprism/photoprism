package commands

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/urfave/cli"
)

// ImportCommand is used to register the import cli command
var ImportCommand = cli.Command{
	Name:    "import",
	Aliases: []string{"mv"},
	Usage:   "Moves files to originals folder, converts and indexes them as needed",
	Action:  importAction,
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

	if err := conf.CreateDirectories(); err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := conf.Init(cctx); err != nil {
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
		return errors.New("import folder is identical with originals")
	}

	log.Infof("moving media files from %s to %s", sourcePath, conf.OriginalsPath())

	imp := service.Import()
	opt := photoprism.ImportOptionsMove(sourcePath)

	imp.Start(opt)

	elapsed := time.Since(start)

	log.Infof("import completed in %s", elapsed)
	conf.Shutdown()
	return nil
}
