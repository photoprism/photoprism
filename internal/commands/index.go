package commands

import (
	"fmt"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/urfave/cli"
	"log"
)

var IndexCommand = cli.Command{
	Name:   "index",
	Usage:  "Re-indexes all originals",
	Action: indexAction,
}

// Indexes original photos; called by IndexCommand
func indexAction(context *cli.Context) error {
	conf := photoprism.NewConfig(context)

	if err := conf.CreateDirectories(); err != nil {
		log.Fatal(err)
	}

	conf.MigrateDb()

	fmt.Printf("Indexing photos in %s...\n", conf.OriginalsPath)

	tensorFlow := photoprism.NewTensorFlow(conf.GetTensorFlowModelPath())

	indexer := photoprism.NewIndexer(conf.OriginalsPath, tensorFlow, conf.GetDb())

	indexer.IndexAll()

	fmt.Println("Done.")

	return nil
}
