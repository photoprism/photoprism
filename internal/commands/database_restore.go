package commands

import (
	"context"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/photoprism/backup"
)

const databaseRestoreDescription = `A custom filename for the backup (or - to read the backup from stdin) can optionally be passed as argument.
If nothing else is specified, the most recent backup in the database backup path is restored.
A relative filename is read from the path specified with --path, if any.

A backup can only be restored into the same kind of database it was created from.`

// DatabaseRestoreCommand configures the command name, flags, and action.
var DatabaseRestoreCommand = &cli.Command{
	Name:        "restore",
	Description: databaseRestoreDescription,
	Usage:       "Restores the index database from a backup",
	ArgsUsage:   "[filename]",
	Flags: append([]cli.Flag{
		ForceFlag("replaces the index database with the backup, if it already exists"),
	}, databasePathFlags()...),
	Action: databaseRestoreAction,
}

// databaseRestoreAction restores the index database from a backup.
func databaseRestoreAction(ctx *cli.Context) error {
	path, fileName, err := databaseBackupArgs(ctx)

	if err != nil {
		return err
	}

	start := time.Now()

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return cli.Exit(err, 1)
	}

	defer conf.Shutdown()

	// A filename has already been resolved in the path.
	if fileName != "" {
		path = ""
	}

	if err = backup.RestoreDatabase(path, fileName, fileName == "-", ctx.Bool("force")); err != nil {
		return cli.Exit(err, 1)
	}

	if err = migrateRestoredIndex(conf); err != nil {
		return err
	}

	log.Infof("completed in %s", time.Since(start))

	return nil
}
