package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

const backupDescription = "A user-defined filename or - for stdout can be passed as the first argument. " +
	"The -i parameter can be omitted in this case.\n" +
	"   Make sure to run the command with exec -T when using Docker to prevent log messages from being sent to stdout.\n" +
	"   The index backup and album file paths are automatically detected if not specified explicitly."

// BackupCommand configures the command name, flags, and action.
var BackupCommand = cli.Command{
	Name:        "backup",
	Description: backupDescription,
	Usage:       "Creates an index database dump and/or album YAML file backups",
	ArgsUsage:   "[filename]",
	Flags:       backupFlags,
	Action:      backupAction,
}

var backupFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "force, f",
		Usage: "replace existing index backup files",
	},
	cli.BoolFlag{
		Name:  "albums, a",
		Usage: "export album metadata to YAML files located in the backup path",
	},
	cli.StringFlag{
		Name:  "albums-path",
		Usage: "custom album backup `PATH`",
	},
	cli.BoolFlag{
		Name:  "index, i",
		Usage: "create index database backup (sent to stdout if - is passed as first argument)",
	},
	cli.StringFlag{
		Name:  "index-path",
		Usage: "custom index backup `PATH`",
	},
	cli.IntFlag{
		Name:  "retain, r",
		Usage: "`NUMBER` of index backups to keep (-1 to keep all)",
		Value: config.DefaultBackupRetain,
	},
}

// backupAction creates a database backup.
func backupAction(ctx *cli.Context) error {
	// Use command argument as backup file name.
	fileName := ctx.Args().First()
	backupPath := ctx.String("index-path")
	backupIndex := ctx.Bool("index") || fileName != "" || backupPath != ""
	albumsPath := ctx.String("albums-path")
	backupAlbums := ctx.Bool("albums") || albumsPath != ""
	force := ctx.Bool("force")
	retain := ctx.Int("retain")

	if !backupIndex && !backupAlbums {
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

	if backupIndex {
		// If empty, use default backup file name.
		if fileName == "" {
			if !fs.PathWritable(backupPath) {
				if backupPath != "" {
					log.Warnf("custom index backup path not writable, using default")
				}

				backupPath = conf.BackupIndexPath()
			}

			backupFile := time.Now().UTC().Format("2006-01-02") + ".sql"
			fileName = filepath.Join(backupPath, backupFile)
		}

		if err = photoprism.BackupIndex(backupPath, fileName, fileName == "-", force, retain); err != nil {
			return fmt.Errorf("failed to create index backup: %w", err)
		}
	}

	if backupAlbums {
		if !fs.PathWritable(albumsPath) {
			if albumsPath != "" {
				log.Warnf("album files path not writable, using default")
			}

			albumsPath = conf.BackupAlbumsPath()
		}

		log.Infof("creating album YAML files in %s", clean.Log(albumsPath))

		if count, backupErr := photoprism.BackupAlbums(albumsPath, true); backupErr != nil {
			return backupErr
		} else {
			log.Infof("created %s", english.Plural(count, "YAML album file", "YAML album files"))
		}
	}

	elapsed := time.Since(start)

	log.Infof("completed in %s", elapsed)

	return nil
}
