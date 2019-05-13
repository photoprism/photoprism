package commands

import (
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// Imports photos from path defined in command-line args
var ImportCommand = cli.Command{
	Name:   "import",
	Usage:  "Imports photos",
	Action: importAction,
}

func importAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)

	if err := conf.CreateDirectories(); err != nil {
		return err
	}

	conf.MigrateDb()

	log.Infof("importing photos from %s", conf.ImportPath())

	tensorFlow := photoprism.NewTensorFlow(conf.TensorFlowModelPath())

	indexer := photoprism.NewIndexer(conf, tensorFlow)

	converter := photoprism.NewConverter(conf.DarktableCli())

	importer := photoprism.NewImporter(conf, indexer, converter)

	importer.ImportPhotosFromDirectory(conf.ImportPath())

	log.Info("photo import complete")

	return nil
}
