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
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

const backupDescription = "A user-defined SQL dump FILENAME or - for stdout can be passed as the first argument. " +
	"The -i parameter can be omitted in this case.\n" +
	"   Make sure to run the command with exec -T when using Docker to prevent log messages from being sent to stdout.\n" +
	"   The index backup and album file paths are automatically detected if not specified explicitly."

// BackupCommand configures the backup cli command.
var BackupCommand = cli.Command{
	Name:        "backup",
	Description: backupDescription,
	Usage:       "Creates an index SQL dump and optionally album YAML files organized by type",
	ArgsUsage:   "[FILENAME | -]",
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
		Usage: "create index SQL dump",
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

	conf := config.NewConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

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
				return fmt.Errorf("SQL dump already exists: %s", indexFileName)
			} else if err == nil {
				log.Warnf("replacing existing SQL dump")
			}

			// Create backup directory if not exists.
			if dir := filepath.Dir(indexFileName); dir != "." {
				if err := os.MkdirAll(dir, os.ModePerm); err != nil {
					return err
				}
			}

			log.Infof("writing SQL dump to %s", sanitize.Log(indexFileName))
		}

		var cmd *exec.Cmd

		switch conf.DatabaseDriver() {
		case config.MySQL, config.MariaDB:
			cmd = exec.Command(
				conf.MysqldumpBin(),
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
				conf.DatabaseDsn(),
				".dump",
			)
		default:
			return fmt.Errorf("unsupported database type: %s", conf.DatabaseDriver())
		}

		// Fetch command output.
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		// Run backup command.
		if err := cmd.Run(); err != nil {
			if stderr.String() != "" {
				return errors.New(stderr.String())
			}
		}

		if indexFileName == "-" {
			// Return output via stdout.
			fmt.Println(out.String())
		} else {
			// Write output to file.
			if err := os.WriteFile(indexFileName, []byte(out.String()), os.ModePerm); err != nil {
				return err
			}
		}
	}

	if backupAlbums {
		service.SetConfig(conf)
		conf.InitDb()

		if !fs.PathWritable(albumsPath) {
			if albumsPath != "" {
				log.Warnf("album files path not writable, using default")
			}

			albumsPath = conf.AlbumsPath()
		}

		log.Infof("saving albums in %s", sanitize.Log(albumsPath))

		if count, err := photoprism.BackupAlbums(albumsPath, true); err != nil {
			return err
		} else {
			log.Infof("created %s", english.Plural(count, "YAML album file", "YAML album files"))
		}
	}

	elapsed := time.Since(start)

	log.Infof("backup completed in %s", elapsed)

	conf.Shutdown()

	return nil
}
