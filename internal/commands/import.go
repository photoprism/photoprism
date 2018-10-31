package commands

import (
	"fmt"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/urfave/cli"
	"log"
)

var ImportCommand = cli.Command{
	Name:   "import",
	Usage:  "Imports photos",
	Action: importAction,
}

// Imports photos from path defined in context arg; called by ImportCommand;
func importAction(context *cli.Context) error {
	conf := photoprism.NewConfig(context)

	if err := conf.CreateDirectories(); err != nil {
		log.Fatal(err)
	}

	conf.MigrateDb()

	fmt.Printf("Importing photos from %s...\n", conf.ImportPath)

	tensorFlow := photoprism.NewTensorFlow(conf.GetTensorFlowModelPath())

	indexer := photoprism.NewIndexer(conf.OriginalsPath, tensorFlow, conf.GetDb())

	converter := photoprism.NewConverter(conf.DarktableCli)

	importer := photoprism.NewImporter(conf.OriginalsPath, indexer, converter)

	importer.ImportPhotosFromDirectory(conf.ImportPath)

	fmt.Println("Done.")

	return nil
}
