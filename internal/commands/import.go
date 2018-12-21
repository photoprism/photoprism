package commands

import (
	"fmt"
	"log"

	"github.com/photoprism/photoprism/internal/context"
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
	conf := context.NewConfig(ctx)

	if err := conf.CreateDirectories(); err != nil {
		log.Fatal(err)
	}

	conf.MigrateDb()

	fmt.Printf("Importing photos from %s...\n", conf.GetImportPath())

	tensorFlow := photoprism.NewTensorFlow(conf.GetTensorFlowModelPath())

	indexer := photoprism.NewIndexer(conf.OriginalsPath(), tensorFlow, conf.Db())

	converter := photoprism.NewConverter(conf.GetDarktableCli())

	importer := photoprism.NewImporter(conf.OriginalsPath(), indexer, converter)

	importer.ImportPhotosFromDirectory(conf.GetImportPath())

	fmt.Println("Done.")

	return nil
}
