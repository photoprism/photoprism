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
	ArgsUsage: "[subfolder]",
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
		log.Infof("config: enabled read-only mode")
	}

	var found fs.Done
	var indexed int

	// Update file index.
	if w := get.Index(); w != nil {
		indexStart := time.Now()
		convert := conf.Settings().Index.Convert && conf.SidecarWritable()
		opt := photoprism.NewIndexOptions(subPath, ctx.Bool("force"), convert, true, false, !ctx.Bool("archived"))

		found, indexed = w.Start(opt)

		log.Infof("index: updated %s [%s]", english.Plural(indexed, "file", "files"), time.Since(indexStart))
	}

	// Remove missing files from search results.
	if w := get.Purge(); w != nil {
		// Purge worker options.
		opt := photoprism.PurgeOptions{
			Path:   subPath,
			Ignore: found,
			Force:  ctx.Bool("force") || ctx.Bool("cleanup") || indexed > 0,
		}

		// Start purge.
		purgeStart := time.Now()
		if files, photos, updated, err := w.Start(opt); err != nil {
			log.Error(err)
		} else if updated > 0 {
			log.Infof("purge: removed %s and %s [%s]", english.Plural(len(files), "file", "files"), english.Plural(len(photos), "photo", "photos"), time.Since(purgeStart))
		}
	}

	// Delete orphaned index entries, sidecar files and thumbnails?
	if ctx.Bool("cleanup") {
		// Get cleanup worker instance.
		w := get.CleanUp()

		// Cleanup worker options.
		opt := photoprism.CleanUpOptions{
			Dry: false,
		}

		// Start index and cache cleanup.
		cleanupStart := time.Now()
		if thumbnails, _, sidecars, err := w.Start(opt); err != nil {
			return err
		} else if total := thumbnails + sidecars; total > 0 {
			log.Infof("cleanup: deleted %s in total [%s]", english.Plural(total, "file", "files"), time.Since(cleanupStart))
		}
	}

	elapsed := time.Since(start)

	log.Infof("indexed %s in %s", english.Plural(len(found), "file", "files"), elapsed)

	return nil
}
