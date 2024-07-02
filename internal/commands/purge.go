package commands

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// PurgeCommand configures the command name, flags, and action.
var PurgeCommand = cli.Command{
	Name:   "purge",
	Usage:  "Updates missing files, photo counts, and album covers",
	Flags:  purgeFlags,
	Action: purgeAction,
}

var purgeFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "hard",
		Usage: "permanently remove from index",
	},
	cli.BoolFlag{
		Name:  "dry",
		Usage: "dry run, don't actually remove anything",
	},
}

// purgeAction removes missing files from search results
func purgeAction(ctx *cli.Context) error {
	start := time.Now()

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	// get cli first argument
	subPath := strings.TrimSpace(ctx.Args().First())

	if subPath == "" {
		log.Infof("purge: removing missing files in %s", clean.Log(filepath.Base(conf.OriginalsPath())))
	} else {
		log.Infof("purge: removing missing files in %s", clean.Log(fs.RelName(filepath.Join(conf.OriginalsPath(), subPath), filepath.Dir(conf.OriginalsPath()))))
	}

	if conf.ReadOnly() {
		log.Infof("config: enabled read-only mode")
	}

	w := get.Purge()

	opt := photoprism.PurgeOptions{
		Path:  subPath,
		Dry:   ctx.Bool("dry"),
		Hard:  ctx.Bool("hard"),
		Force: true,
	}

	if files, photos, updated, err := w.Start(opt); err != nil {
		return err
	} else if updated > 0 {
		log.Infof("purged %s and %s in %s", english.Plural(len(files), "file", "files"), english.Plural(len(photos), "photo", "photos"), time.Since(start))
	} else {
		log.Infof("purge completed in %s", time.Since(start))
	}

	return nil
}
