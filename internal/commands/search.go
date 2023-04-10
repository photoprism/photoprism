package commands

import (
	"context"
	"time"
	"strings"

	"fmt"

	"github.com/urfave/cli"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/search"
	"github.com/photoprism/photoprism/internal/entity"
)

// SearchCommand registers the search cli command.
var SearchCommand = cli.Command{
	Name:      "search",
	Usage:     "Searches in library using filters",
	ArgsUsage: "search-query",
	Action:    searchAction,
}

// searchAction searches all photos in library
func searchAction(ctx *cli.Context) error {
	start := time.Now()

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	form := form.SearchPhotos{Query: strings.TrimSpace(ctx.Args().First()), Primary: false, Merged: false}
	photos, _, err := search.Photos(form)

	for _, photo := range photos {
		p := entity.Photo{ID: photo.ID}
		p.PreloadFiles()
		for _, file := range p.Files {
			fmt.Printf("%s\n", file.FileName)
		}
	}

	elapsed := time.Since(start)

	log.Infof("searched in %s", elapsed)

	return nil
}
