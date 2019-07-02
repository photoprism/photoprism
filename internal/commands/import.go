package commands

import (
	"context"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/urfave/cli"
)

// Imports photos from path defined in command-line args
var ImportCommand = cli.Command{
	Name:   "import",
	Usage:  "Imports photos",
	Action: importAction,
}

func importAction(ctx *cli.Context) error {
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

	log.Infof("importing photos from %s", conf.ImportPath())

	tensorFlow := photoprism.NewTensorFlow(conf)

	indexer := photoprism.NewIndexer(conf, tensorFlow)

	converter := photoprism.NewConverter(conf)

	importer := photoprism.NewImporter(conf, indexer, converter)

	importer.ImportPhotosFromDirectory(conf.ImportPath())

	elapsed := time.Since(start)

	log.Infof("photo import completed in %s", elapsed)
	conf.Shutdown()
	return nil
}
