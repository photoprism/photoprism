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

	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/urfave/cli"
)

// BackupCommand configures the backup cli command.
var BackupCommand = cli.Command{
	Name:   "backup",
	Usage:  "Creates an index database backup",
	Flags:  backupFlags,
	Action: backupAction,
}

var backupFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "force, f",
		Usage: "overwrite existing backup files",
	},
}

// backupAction creates a database backup.
func backupAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

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
	case config.MySQL:
		cmd = exec.Command(
			conf.MysqldumpBin(),
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

	elapsed := time.Since(start)

	log.Infof("database backup completed in %s", elapsed)

	conf.Shutdown()

	return nil
}
