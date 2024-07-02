package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/backup"
	"github.com/photoprism/photoprism/pkg/fs"
)

const backupDescription = `A custom filename for the database backup (or - to send the backup to stdout) can optionally be passed as argument.
   The --database flag can be omitted in this case. When using Docker, please run the docker command with the -T flag
   to prevent log messages from being sent to stdout. If nothing else is specified, the database and album backup paths
   will be automatically determined based on the current configuration.`

// BackupCommand configures the command name, flags, and action.
var BackupCommand = cli.Command{
	Name:        "backup",
	Description: backupDescription,
	Usage:       "Creates an index database backup and/or album YAML backup files",
	ArgsUsage:   "[filename]",
	Flags:       backupFlags,
	Action:      backupAction,
}

var backupFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "force, f",
		Usage: "replace the index database backup file, if it exists",
	},
	cli.BoolFlag{
		Name:  "albums, a",
		Usage: "create YAML files to back up album metadata (in the standard backup path if no other path is specified)",
	},
	cli.StringFlag{
		Name:  "albums-path",
		Usage: "custom album backup `PATH`",
	},
	cli.BoolFlag{
		Name:  "database, index, i",
		Usage: "create index database backup (in the backup path with the date as filename if no filename is passed, or sent to stdout if - is passed as filename)",
	},
	cli.StringFlag{
		Name:  "database-path, index-path",
		Usage: "custom database backup `PATH`",
	},
	cli.IntFlag{
		Name:  "retain, r",
		Usage: "`NUMBER` of database backups to keep (-1 to keep all)",
		Value: config.DefaultBackupRetain,
	},
}

// backupAction creates a database backup.
func backupAction(ctx *cli.Context) error {
	// Use command argument as backup file name.
	fileName := ctx.Args().First()
	databasePath := ctx.String("database-path")
	backupDatabase := ctx.Bool("database") || fileName != "" || databasePath != ""
	albumsPath := ctx.String("albums-path")
	backupAlbums := ctx.Bool("albums") || albumsPath != ""
	force := ctx.Bool("force")
	retain := ctx.Int("retain")

	if !backupDatabase && !backupAlbums {
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

	if backupDatabase {
		// Use default if no explicit filename was provided.
		if fileName == "" {
			if !fs.PathWritable(databasePath) {
				if databasePath != "" {
					log.Warnf("backup: specified database backup path is not writable, using default directory instead")
				}

				databasePath = conf.BackupDatabasePath()
			}

			backupFile := time.Now().UTC().Format("2006-01-02") + ".sql"
			fileName = filepath.Join(databasePath, backupFile)
		}

		if err = backup.Database(databasePath, fileName, fileName == "-", force, retain); err != nil {
			return fmt.Errorf("failed to create database backup: %w", err)
		}
	}

	if backupAlbums {
		if !fs.PathWritable(albumsPath) {
			if albumsPath != "" {
				log.Warnf("backup: specified albums backup path is not writable, using default directory instead")
			}

			albumsPath = conf.BackupAlbumsPath()
		}

		if count, backupErr := backup.Albums(albumsPath, true); backupErr != nil {
			return backupErr
		} else {
			log.Infof("backup: saved %s", english.Plural(count, "album backup", "album backups"))
		}
	}

	elapsed := time.Since(start)

	log.Infof("completed in %s", elapsed)

	return nil
}
