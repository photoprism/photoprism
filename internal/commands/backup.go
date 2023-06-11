package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	Usage:       "Creates an index backup and optionally album YAML files organized by type",
	ArgsUsage:   "[filename]",
	Flags:       backupFlags,
	Action:      backupAction,
}

var backupFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "force, f",
		Usage: "replace existing files",
	},
	cli.BoolFlag{
		Name:  "albums, a",
		Usage: "create album YAML files organized by type",
	},
	cli.StringFlag{
		Name:  "albums-path",
		Usage: "custom album files `PATH`",
	},
	cli.BoolFlag{
		Name:  "index, i",
		Usage: "create index backup",
	},
	cli.StringFlag{
		Name:  "index-path",
		Usage: "custom index backup `PATH`",
	},
}

// backupAction creates a database backup.
func backupAction(ctx *cli.Context) error {
	// Use command argument as backup file name.
	indexFileName := ctx.Args().First()
	indexPath := ctx.String("index-path")
	backupIndex := ctx.Bool("index") || indexFileName != "" || indexPath != ""

	albumsPath := ctx.String("albums-path")

	backupAlbums := ctx.Bool("albums") || albumsPath != ""

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
		if indexFileName == "" {
			if !fs.PathWritable(indexPath) {
				if indexPath != "" {
					log.Warnf("custom index backup path not writable, using default")
				}

				indexPath = filepath.Join(conf.BackupPath(), conf.DatabaseDriver())
			}

			backupFile := time.Now().UTC().Format("2006-01-02") + ".sql"
			indexFileName = filepath.Join(indexPath, backupFile)
		}

		if indexFileName != "-" {
			if _, err := os.Stat(indexFileName); err == nil && !ctx.Bool("force") {
				return fmt.Errorf("%s already exists", clean.Log(indexFileName))
			} else if err == nil {
				log.Warnf("replacing existing backup")
			}

			// Create backup directory if not exists.
			if dir := filepath.Dir(indexFileName); dir != "." {
				if err := os.MkdirAll(dir, fs.ModeDir); err != nil {
					return err
				}
			}
		}

		var cmd *exec.Cmd

		switch conf.DatabaseDriver() {
		case config.MySQL, config.MariaDB:
			cmd = exec.Command(
				conf.MariadbDumpBin(),
				"--protocol", "tcp",
				"-h", conf.DatabaseHost(),
				"-P", conf.DatabasePortString(),
				"-u", conf.DatabaseUser(),
				"-p"+conf.DatabasePassword(),
				conf.DatabaseName(),
			)
		case config.SQLite3:
			cmd = exec.Command(
				conf.SqliteBin(),
				conf.DatabaseFile(),
				".dump",
			)
		default:
			return fmt.Errorf("unsupported database type: %s", conf.DatabaseDriver())
		}

		// Write to stdout or file.
		var f *os.File
		if indexFileName == "-" {
			log.Infof("writing backup to stdout")
			f = os.Stdout
		} else if f, err = os.OpenFile(indexFileName, os.O_TRUNC|os.O_RDWR|os.O_CREATE, fs.ModeFile); err != nil {
			return fmt.Errorf("failed to create %s: %s", clean.Log(indexFileName), err)
		} else {
			log.Infof("writing backup to %s", clean.Log(indexFileName))
			defer f.Close()
		}

		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		cmd.Stdout = f

		// Log exact command for debugging in trace mode.
		log.Trace(cmd.String())

		// Run backup command.
		if err := cmd.Run(); err != nil {
			if stderr.String() != "" {
				return errors.New(stderr.String())
			}
		}
	}

	if backupAlbums {
		if !fs.PathWritable(albumsPath) {
			if albumsPath != "" {
				log.Warnf("album files path not writable, using default")
			}

			albumsPath = conf.AlbumsPath()
		}

		log.Infof("saving albums in %s", clean.Log(albumsPath))

		if count, err := photoprism.BackupAlbums(albumsPath, true); err != nil {
			return err
		} else {
			log.Infof("created %s", english.Plural(count, "YAML album file", "YAML album files"))
		}
	}

	elapsed := time.Since(start)

	log.Infof("completed in %s", elapsed)

	return nil
}
