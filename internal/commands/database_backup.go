package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

const databaseBackupDescription = `A custom filename for the backup (or - to send the backup to stdout) can optionally be passed as argument.
If nothing else is specified, the backup is saved in the database backup path with the date as filename.
A relative filename is saved in the path specified with --path, if any, and older backups are only removed
according to --retain when no filename is passed.
When using Docker, please run the docker command with the -T flag to prevent log messages from being sent to stdout.

Backups to stdout (-) bypass the insufficient storage check, so dumps can be streamed even when the local storage is full.
A backup file is replaced only once the new dump is complete. The filename must not be a symbolic link, even with --force,
so use - rather than /dev/stdout to send the backup to stdout.

A backup can only be restored into the same kind of database it was created from.`

// DatabaseBackupCommand configures the command name, flags, and action.
var DatabaseBackupCommand = &cli.Command{
	Name:        "backup",
	Description: databaseBackupDescription,
	Usage:       "Creates an index database backup",
	ArgsUsage:   "[filename]",
	Flags: append([]cli.Flag{
		ForceFlag("replaces the backup file, if it exists"),
		&cli.IntFlag{
			Name:        "retain",
			Aliases:     []string{"r"},
			Usage:       "`NUMBER` of database backups to keep (-1 to keep all)",
			DefaultText: "global value",
		},
	}, databasePathFlags()...),
	Action: databaseBackupAction,
}

// databaseBackupAction creates an index database backup.
func databaseBackupAction(ctx *cli.Context) error {
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

	// A path that was specified is created, rather than replaced by the default path if it does not exist.
	if path != "" && fileName == "" {
		if err = fs.MkdirAll(path); err != nil {
			return cli.Exit(fmt.Errorf("failed to create backup path %s: %w", clean.Log(path), err), 1)
		} else if !fs.PathWritable(path) {
			return cli.Exit(fmt.Errorf("backup path %s is not writable", clean.Log(path)), 1)
		}
	}

	if err = backupIndex(conf, path, fileName, ctx.Bool("force"), backupRetain(ctx, conf.BackupRetain())); err != nil {
		return err
	}

	log.Infof("completed in %s", time.Since(start))

	return nil
}
