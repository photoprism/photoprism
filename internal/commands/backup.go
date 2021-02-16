package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/photoprism/photoprism/internal/service"

	"github.com/photoprism/photoprism/internal/photoprism"

	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/urfave/cli"
)

// BackupCommand configures the backup cli command.
var BackupCommand = cli.Command{
	Name:   "backup",
	Usage:  "Creates album and index backups",
	Flags:  backupFlags,
	Action: backupAction,
}

var backupFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "force, f",
		Usage: "overwrite existing backup files",
	},
	cli.BoolFlag{
		Name:  "albums, a",
		Usage: "create album yaml file backups",
	},
	cli.BoolFlag{
		Name:  "index, i",
		Usage: "create index database backup",
	},
}

// backupAction creates a database backup.
func backupAction(ctx *cli.Context) error {
	if !ctx.Bool("index") && !ctx.Bool("albums") {
		for _, flag := range backupFlags {
			fmt.Println(flag.String())
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

	if ctx.Bool("index") {
		// Use command argument as backup file name.
		fileName := ctx.Args().First()

		// If empty, use default backup file name.
		if fileName == "" {
			backupFile := time.Now().UTC().Format("2006-01-02") + ".sql"
			backupPath := filepath.Join(conf.BackupPath(), conf.DatabaseDriver())
			fileName = filepath.Join(backupPath, backupFile)
		}

		if _, err := os.Stat(fileName); err == nil && !ctx.Bool("force") {
			return fmt.Errorf("backup file already exists: %s", fileName)
		} else if err == nil {
			log.Warnf("replacing existing backup file")
		}

		// Create backup directory if not exists.
		if dir := filepath.Dir(fileName); dir != "." {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}
		}

		log.Infof("backing up database to %s", txt.Quote(fileName))

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
		case config.SQLite:
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

		// Write output to file.
		if err := ioutil.WriteFile(fileName, []byte(out.String()), os.ModePerm); err != nil {
			return err
		}
	}

	if ctx.Bool("albums") {
		service.SetConfig(conf)
		conf.InitDb()

		if count, err := photoprism.BackupAlbums(true); err != nil {
			return err
		} else {
			log.Infof("%d albums saved as yaml files", count)
		}
	}

	elapsed := time.Since(start)

	log.Infof("backup completed in %s", elapsed)

	conf.Shutdown()

	return nil
}
