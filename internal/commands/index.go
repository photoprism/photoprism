package commands

import (
	"github.com/photoprism/photoprism/internal/config"
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
	conf := config.NewConfig(ctx)

	if err := conf.CreateDirectories(); err != nil {
		return err
	}

	conf.MigrateDb()

	log.Infof("indexing photos in %s", conf.OriginalsPath())

	tensorFlow := photoprism.NewTensorFlow(conf.TensorFlowModelPath())

	indexer := photoprism.NewIndexer(conf.OriginalsPath(), tensorFlow, conf.Db())

	files := indexer.IndexAll()

	log.Infof("indexed %d files", len(files))

	conf.Shutdown()

	return nil
}
