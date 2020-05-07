package commands

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/urfave/cli"
)

// IndexCommand is used to register the index cli command
var IndexCommand = cli.Command{
	Name:   "index",
	Usage:  "Indexes media files in originals folder",
	Flags:  indexFlags,
	Action: indexAction,
}

var indexFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "all, a",
		Usage: "re-index all originals, including unchanged files",
	},
	cli.BoolFlag{
		Name:  "convert, c",
		Usage: "convert other file types to jpeg so that they can be displayed in a browser",
	},
}

// indexAction indexes all photos in originals directory (photo library)
func indexAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	if err := conf.CreateDirectories(); err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := conf.Init(cctx); err != nil {
		return err
	}

	conf.InitDb()

	// get cli first argument
	subPath := strings.TrimSpace(ctx.Args().First())

	if subPath == "" {
		log.Infof("indexing photos in %s", txt.Quote(conf.OriginalsPath()))
	} else {
		log.Infof("indexing originals folder %s", txt.Quote(filepath.Join(conf.OriginalsPath(), subPath)))
	}

	if conf.ReadOnly() {
		log.Infof("read-only mode enabled")
	}

	ind := service.Index()

	opt := photoprism.IndexOptions{
		Path:    subPath,
		Rescan:  ctx.Bool("all"),
		Convert: ctx.Bool("convert"),
	}

	files := ind.Start(opt)
	elapsed := time.Since(start)

	log.Infof("indexed %d files in %s", len(files), elapsed)

	conf.Shutdown()

	return nil
}
