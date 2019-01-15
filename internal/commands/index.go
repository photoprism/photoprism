package commands

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/context"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/urfave/cli"
)

// Re-indexes all photos in originals directory (photo library)
var IndexCommand = cli.Command{
	Name:   "index",
	Usage:  "Re-indexes all originals",
	Action: indexAction,
}

func indexAction(ctx *cli.Context) error {
	conf := context.NewConfig(ctx)

	if err := conf.CreateDirectories(); err != nil {
		return err
	}

	conf.MigrateDb()

	fmt.Printf("Indexing photos in %s...\n", conf.OriginalsPath())

	tensorFlow := photoprism.NewTensorFlow(conf.TensorFlowModelPath())

	indexer := photoprism.NewIndexer(conf.OriginalsPath(), tensorFlow, conf.Db())

	indexer.IndexAll()

	fmt.Println("Done.")

	return nil
}
