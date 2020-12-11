package commands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/urfave/cli"
)

// RestoreCommand configures the backup cli command.
var RestoreCommand = cli.Command{
	Name:   "restore",
	Usage:  "Restores the index from a backup",
	Flags:  restoreFlags,
	Action: restoreAction,
}

var restoreFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "force, f",
		Usage: "overwrite existing index",
	},
}

// restoreAction restores a database backup.
func restoreAction(ctx *cli.Context) error {
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
		backupPath := filepath.Join(conf.BackupPath(), conf.DatabaseDriver())

		matches, err := filepath.Glob(filepath.Join(regexp.QuoteMeta(backupPath), "*.sql"))

		if err != nil {
			return err
		}

		if len(matches) == 0 {
			log.Errorf("no backup files found in %s", backupPath)
			return nil
		}

		fileName = matches[len(matches)-1]
	}

	if !fs.FileExists(fileName) {
		log.Errorf("backup file not found: %s", fileName)
		return nil
	}

	counts := struct{ Photos int }{}

	conf.Db().Unscoped().Table("photos").
		Select("COUNT(*) AS photos").
		Take(&counts)

	if counts.Photos == 0 {
		// Do nothing;
	} else if !ctx.Bool("force") {
		return fmt.Errorf("use --force to replace exisisting index with %d photos", counts.Photos)
	} else {
		log.Warnf("replacing existing index with %d photos", counts.Photos)
	}

	log.Infof("restoring index from %s", txt.Quote(fileName))

	sqlBackup, err := ioutil.ReadFile(fileName)

	if err != nil {
		return err
	}

	entity.SetDbProvider(conf)
	tables := entity.Entities

	var cmd *exec.Cmd

	switch conf.DatabaseDriver() {
	case config.MySQL:
		cmd = exec.Command(
			conf.MysqlBin(),
			"-h", conf.DatabaseHost(),
			"-P", conf.DatabasePortString(),
			"-u", conf.DatabaseUser(),
			"-p"+conf.DatabasePassword(),
			"-f",
			conf.DatabaseName(),
		)
	case config.SQLite:
		log.Infoln("dropping existing tables")
		tables.Drop()
		cmd = exec.Command(
			conf.SqliteBin(),
			conf.DatabaseDsn(),
		)
	default:
		return fmt.Errorf("unsupported database type: %s", conf.DatabaseDriver())
	}

	// Fetch command output.
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	stdin, err := cmd.StdinPipe()

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		if _, err := io.WriteString(stdin, string(sqlBackup)); err != nil {
			log.Errorf(err.Error())
		}
	}()

	// Run backup command.
	if err := cmd.Run(); err != nil {
		if stderr.String() != "" {
			log.Debugln(stderr.String())
			log.Warnf("index could not be restored completely")
		}
	}

	log.Infoln("migrating database")

	conf.InitDb()

	elapsed := time.Since(start)

	log.Infof("database restored in %s", elapsed)

	conf.Shutdown()

	return nil
}
