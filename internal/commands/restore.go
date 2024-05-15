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

const restoreDescription = `A custom filename for the database backup (or - to read the backup from stdin) can optionally be passed as argument.
   The --database flag can be omitted in this case. If nothing else is specified, the database and album backup paths
   will be automatically determined based on the current configuration.`

// RestoreCommand configures the command name, flags, and action.
var RestoreCommand = cli.Command{
	Name:        "restore",
	Description: restoreDescription,
	Usage:       "Restores the index database and/or album metadata from a backup",
	ArgsUsage:   "[filename]",
	Flags:       restoreFlags,
	Action:      restoreAction,
}

var restoreFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "force, f",
		Usage: "replace the index database with the backup, if it already exists",
	},
	cli.BoolFlag{
		Name:  "albums, a",
		Usage: "restore albums from the YAML backup files found in the album backup path",
	},
	cli.StringFlag{
		Name:  "albums-path",
		Usage: "custom album backup `PATH`",
	},
	cli.BoolFlag{
		Name:  "database, index, i",
		Usage: "restore the index database from the specified file (stdin if - is passed as filename), or the most recent backup found in the database backup path",
	},
	cli.StringFlag{
		Name:  "database-path, index-path",
		Usage: "custom database backup `PATH`",
	},
}

// restoreAction restores a database backup.
func restoreAction(ctx *cli.Context) error {
	// Use command argument as backup file name.
	databaseFile := ctx.Args().First()
	databasePath := ctx.String("database-path")
	restoreDatabase := ctx.Bool("database") || databaseFile != "" || databasePath != ""
	force := ctx.Bool("force")
	albumsPath := ctx.String("albums-path")
	restoreAlbums := ctx.Bool("albums") || albumsPath != ""

	if !restoreDatabase && !restoreAlbums {
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

	// Restore database from backup dump?
	if !restoreDatabase {
		// Do nothing.
	} else if err = photoprism.RestoreDatabase(databasePath, databaseFile, databaseFile == "-", force); err != nil {
		return err
	}

	log.Infoln("restore: migrating index database schema")

	conf.InitDb()

	// Restore albums from YAML backup files?
	if restoreAlbums {
		get.SetConfig(conf)

		if albumsPath == "" {
			albumsPath = conf.BackupAlbumsPath()
		}

		if !fs.PathExists(albumsPath) {
			log.Warnf("restore: failed to open %s, album backups cannot be restored", clean.Log(albumsPath))
		} else {
			log.Infof("restore: restoring album backups from %s", clean.Log(albumsPath))

			if count, restoreErr := photoprism.RestoreAlbums(albumsPath, true); restoreErr != nil {
				return restoreErr
			} else {
				log.Infof("restore: restored %s from YAML files", english.Plural(count, "album", "albums"))
			}
		}
	}

	elapsed := time.Since(start)

	log.Infof("completed in %s", elapsed)

	return nil
}
