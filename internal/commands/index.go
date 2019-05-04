package commands

import (
	"github.com/photoprism/photoprism/internal/context"
	"github.com/photoprism/photoprism/internal/photoprism"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// Re-indexes all photos in originals directory (photo library)
var IndexCommand = cli.Command{
	Name:   "index",
	Usage:  "Re-indexes all originals",
	Action: indexAction,
}

func indexAction(ctx *cli.Context) error {
	app := context.NewContext(ctx)

	if err := app.CreateDirectories(); err != nil {
		return err
	}

	app.MigrateDb()

	log.Infof("indexing photos in %s", app.OriginalsPath())

	tensorFlow := photoprism.NewTensorFlow(app.TensorFlowModelPath())

	indexer := photoprism.NewIndexer(app.OriginalsPath(), tensorFlow, app.Db())

	files := indexer.IndexAll()

	log.Infof("indexed %d files", len(files))

	app.Shutdown()

	return nil
}
