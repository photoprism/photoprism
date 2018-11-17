package commands

import (
	"fmt"
	"log"

	"github.com/photoprism/photoprism/internal/context"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/urfave/cli"
)

// Converts RAW files to JPEG images, if no JPEG already exists
var ConvertCommand = cli.Command{
	Name:   "convert",
	Usage:  "Converts RAW originals to JPEG",
	Action: convertAction,
}

func convertAction(ctx *cli.Context) error {
	conf := context.NewConfig(ctx)

	if err := conf.CreateDirectories(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Converting RAW images in %s to JPEG...\n", conf.GetOriginalsPath())

	converter := photoprism.NewConverter(conf.GetDarktableCli())

	converter.ConvertAll(conf.GetOriginalsPath())

	fmt.Println("Done.")

	return nil
}
