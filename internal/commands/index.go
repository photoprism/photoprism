package commands

import (
	"fmt"
	"log"

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
		log.Fatal(err)
	}

	conf.MigrateDb()

	fmt.Printf("Indexing photos in %s...\n", conf.GetOriginalsPath())

	tensorFlow := photoprism.NewTensorFlow(conf.GetTensorFlowModelPath())

	indexer := photoprism.NewIndexer(conf.GetOriginalsPath(), tensorFlow, conf.Db())

	indexer.IndexAll()

	fmt.Println("Done.")

	return nil
}
