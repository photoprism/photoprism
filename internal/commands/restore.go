package commands

import (
	"context"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

const restoreDescription = "A user-defined filename or - for stdin can be passed as the first argument. " +
	"The -i parameter can be omitted in this case.\n" +
	"   The index backup and album file paths are automatically detected if not specified explicitly."

// RestoreCommand configures the command name, flags, and action.
var RestoreCommand = cli.Command{
	Name:        "restore",
	Description: restoreDescription,
	Usage:       "Restores the index from a database dump and/or album YAML file backups",
	ArgsUsage:   "[filename]",
	Flags:       restoreFlags,
	Action:      restoreAction,
}

var restoreFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "force, f",
		Usage: "replace existing index schema and data",
	},
	cli.BoolFlag{
		Name:  "albums, a",
		Usage: "restore album metadata from YAML backup files",
	},
	cli.StringFlag{
		Name:  "albums-path",
		Usage: "custom album backup `PATH`",
	},
	cli.BoolFlag{
		Name:  "index, i",
		Usage: "restore index from the specified file or the most recent file in the backup path (from stdin if - is passed as first argument)",
	},
	cli.StringFlag{
		Name:  "index-path",
		Usage: "custom index backup `PATH`",
	},
}

// restoreAction restores a database backup.
func restoreAction(ctx *cli.Context) error {
	// Use command argument as backup file name.
	indexFileName := ctx.Args().First()
	indexPath := ctx.String("index-path")
	restoreIndex := ctx.Bool("index") || indexFileName != "" || indexPath != ""
	force := ctx.Bool("force")
	albumsPath := ctx.String("albums-path")
	restoreAlbums := ctx.Bool("albums") || albumsPath != ""

	if !restoreIndex && !restoreAlbums {
		return cli.ShowSubcommandHelp(ctx)
	}

	start := time.Now()

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.RegisterDb()
	defer conf.Shutdown()

	// Restore index from specified file?
	if !restoreIndex {
		// Do nothing.
	} else if err = photoprism.RestoreIndex(indexPath, indexFileName, indexFileName == "-", force); err != nil {
		return err
	}

	log.Infoln("migrating index database schema")

	conf.InitDb()

	if restoreAlbums {
		get.SetConfig(conf)

		if albumsPath == "" {
			albumsPath = conf.BackupAlbumsPath()
		}

		if !fs.PathExists(albumsPath) {
			log.Warnf("album files path %s not found", clean.Log(albumsPath))
		} else {
			log.Infof("restoring albums from %s", clean.Log(albumsPath))

			if count, err := photoprism.RestoreAlbums(albumsPath, true); err != nil {
				return err
			} else {
				log.Infof("restored %s from YAML files", english.Plural(count, "album", "albums"))
			}
		}
	}

	elapsed := time.Since(start)

	log.Infof("restored in %s", elapsed)

	return nil
}
