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
	"github.com/photoprism/photoprism/pkg/txt"
)

// BackupCommand configures the backup cli command.
var BackupCommand = cli.Command{
	Name:      "backup",
	Usage:     "Creates index database dumps and optional YAML album backups",
	UsageText: `A custom database SQL dump FILENAME may be passed as first argument. Use - for stdout. The backup paths will be detected automatically if not provided.`,
	Flags:     backupFlags,
	Action:    backupAction,
}

var backupFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "force, f",
		Usage: "replace existing backup files",
	},
	cli.BoolFlag{
		Name:  "albums, a",
		Usage: "create YAML album backups",
	},
	cli.StringFlag{
		Name:  "albums-path",
		Usage: "custom albums backup `PATH`",
	},
	cli.BoolFlag{
		Name:  "index, i",
		Usage: "create index database SQL dump",
	},
	cli.StringFlag{
		Name:  "index-path",
		Usage: "custom database backup `PATH`",
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
		fmt.Printf("OPTIONS:\n")

		for _, flag := range backupFlags {
			fmt.Printf("   %s\n", flag.String())
		}

		return nil
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
				return fmt.Errorf("backup file already exists: %s", indexFileName)
			} else if err == nil {
				log.Warnf("replacing existing backup file")
			}

			// Create backup directory if not exists.
			if dir := filepath.Dir(indexFileName); dir != "." {
				if err := os.MkdirAll(dir, os.ModePerm); err != nil {
					return err
				}
			}

			log.Infof("backing up database to %s", txt.Quote(indexFileName))
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
				log.Warnf("albums backup path not writable, using default")
			}

			albumsPath = conf.AlbumsPath()
		}

		log.Infof("backing up albums to %s", txt.Quote(albumsPath))

		if count, err := photoprism.BackupAlbums(albumsPath, true); err != nil {
			return err
		} else {
			log.Infof("created %s", english.Plural(count, "YAML album backup", "YAML album backups"))
		}
	}

	elapsed := time.Since(start)

	log.Infof("backup completed in %s", elapsed)

	conf.Shutdown()

	return nil
}
