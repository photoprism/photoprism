package commands

import (
	"github.com/photoprism/photoprism/internal/context"
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
	app := context.NewContext(ctx)

	if err := app.CreateDirectories(); err != nil {
		return err
	}

	app.MigrateDb()

	log.Infof("importing photos from %s", app.ImportPath())

	tensorFlow := photoprism.NewTensorFlow(app.TensorFlowModelPath())

	indexer := photoprism.NewIndexer(app.OriginalsPath(), tensorFlow, app.Db())

	converter := photoprism.NewConverter(app.DarktableCli())

	importer := photoprism.NewImporter(app.OriginalsPath(), indexer, converter)

	importer.ImportPhotosFromDirectory(app.ImportPath())

	log.Info("photo import complete")

	return nil
}
