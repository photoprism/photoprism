package commands

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/nsfw"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/urfave/cli"
)

var MoveCommand = cli.Command{
	Name:   "move",
	Aliases: []string{"mv"},
	Usage:  "Moves files to originals path, converts and indexes them as needed",
	Action: moveAction,
}

// Moves photos to originals path.
func moveAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)

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

	conf.MigrateDb()

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
		return errors.New("import path is identical with originals path")
	}

	log.Infof("moving media files from %s to %s", sourcePath, conf.OriginalsPath())

	tensorFlow := classify.New(conf.ResourcesPath(), conf.TensorFlowDisabled())
	nsfwDetector := nsfw.New(conf.NSFWModelPath())

	ind := photoprism.NewIndex(conf, tensorFlow, nsfwDetector)

	convert := photoprism.NewConvert(conf)

	imp := photoprism.NewImport(conf, ind, convert)
	opt := photoprism.ImportOptionsMove(sourcePath)

	imp.Start(opt)

	elapsed := time.Since(start)

	log.Infof("import completed in %s", elapsed)
	conf.Shutdown()
	return nil
}
