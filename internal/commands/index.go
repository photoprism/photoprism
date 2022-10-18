package commands

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// IndexCommand registers the index cli command.
var IndexCommand = cli.Command{
	Name:      "index",
	Usage:     "Indexes original media files",
	ArgsUsage: "[sub-folder]",
	Flags:     indexFlags,
	Action:    indexAction,
}

var indexFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "force, f",
		Usage: "rescan all originals, including unchanged files",
	},
	cli.BoolFlag{
		Name:  "archived, a",
		Usage: "do not skip files belonging to archived photos",
	},
	cli.BoolFlag{
		Name:  "cleanup, c",
		Usage: "remove orphan index entries and thumbnails",
	},
}

// indexAction indexes all photos in originals directory (photo library)
func indexAction(ctx *cli.Context) error {
	start := time.Now()

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	// Use first argument to limit scope if set.
	subPath := strings.TrimSpace(ctx.Args().First())

	if subPath == "" {
		log.Infof("indexing originals in %s", clean.Log(conf.OriginalsPath()))
	} else {
		log.Infof("indexing originals in %s", clean.Log(filepath.Join(conf.OriginalsPath(), subPath)))
	}

	if conf.ReadOnly() {
		log.Infof("config: read-only mode enabled")
	}

	var indexed fs.Done

	if w := get.Index(); w != nil {
		convert := conf.Settings().Index.Convert && conf.SidecarWritable()
		opt := photoprism.NewIndexOptions(subPath, ctx.Bool("force"), convert, true, false, !ctx.Bool("archived"))

		indexed = w.Start(opt)
	}

	if w := get.Purge(); w != nil {
		purgeStart := time.Now()
		opt := photoprism.PurgeOptions{
			Path:   subPath,
			Ignore: indexed,
		}

		if files, photos, err := w.Start(opt); err != nil {
			log.Error(err)
		} else if len(files) > 0 || len(photos) > 0 {
			log.Infof("purge: removed %s and %s [%s]", english.Plural(len(files), "file", "files"), english.Plural(len(photos), "photo", "photos"), time.Since(purgeStart))
		}
	}

	if ctx.Bool("cleanup") {
		cleanupStart := time.Now()
		w := get.CleanUp()

		opt := photoprism.CleanUpOptions{
			Dry: false,
		}

		// Start cleanup worker.
		if thumbnails, _, sidecars, err := w.Start(opt); err != nil {
			return err
		} else if total := thumbnails + sidecars; total > 0 {
			log.Infof("cleanup: removed %s in total [%s]", english.Plural(total, "file", "files"), time.Since(cleanupStart))
		}
	}

	elapsed := time.Since(start)

	log.Infof("indexed %s in %s", english.Plural(len(indexed), "file", "files"), elapsed)

	return nil
}
