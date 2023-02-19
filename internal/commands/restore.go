package commands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

const restoreDescription = "A user-defined SQL dump FILENAME can be passed as the first argument. " +
	"The -i parameter can be omitted in this case.\n" +
	"   The index backup and album file paths are automatically detected if not specified explicitly."

// RestoreCommand configures the command name, flags, and action.
var RestoreCommand = cli.Command{
	Name:        "restore",
	Description: restoreDescription,
	Usage:       "Restores the index from an SQL dump and optionally albums from YAML files",
	ArgsUsage:   "filename.sql",
	Flags:       restoreFlags,
	Action:      restoreAction,
}

var restoreFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "force, f",
		Usage: "replace existing index",
	},
	cli.BoolFlag{
		Name:  "albums, a",
		Usage: "restore albums from YAML files",
	},
	cli.StringFlag{
		Name:  "albums-path",
		Usage: "custom album files `PATH`",
	},
	cli.BoolFlag{
		Name:  "index, i",
		Usage: "restore index from SQL dump",
	},
	cli.StringFlag{
		Name:  "index-path",
		Usage: "custom index backup `PATH`",
	},
}

// restoreAction restores a database backup.
func restoreAction(ctx *cli.Context) error {
	// Use command argument as backup file name.
	indexFileName := ctx.Args().First()
	indexPath := ctx.String("index-path")
	restoreIndex := ctx.Bool("index") || indexFileName != "" || indexPath != ""

	albumsPath := ctx.String("albums-path")
	restoreAlbums := ctx.Bool("albums") || albumsPath != ""

	if !restoreIndex && !restoreAlbums {
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

	if restoreIndex {
		// If empty, use default backup file name.
		if indexFileName == "" {
			if indexPath == "" {
				indexPath = filepath.Join(conf.BackupPath(), conf.DatabaseDriver())
			}

			matches, err := filepath.Glob(filepath.Join(regexp.QuoteMeta(indexPath), "*.sql"))

			if err != nil {
				return err
			}

			if len(matches) == 0 {
				log.Errorf("no SQL dumps found in %s", indexPath)
				return nil
			}

			indexFileName = matches[len(matches)-1]
		}

		if !fs.FileExists(indexFileName) {
			log.Errorf("SQL dump not found: %s", indexFileName)
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

		log.Infof("restoring index from %s", clean.Log(indexFileName))

		sqlBackup, err := os.ReadFile(indexFileName)

		if err != nil {
			return err
		}

		tables := entity.Entities

		var cmd *exec.Cmd

		switch conf.DatabaseDriver() {
		case config.MySQL, config.MariaDB:
			cmd = exec.Command(
				conf.MysqlBin(),
				"--protocol", "tcp",
				"-h", conf.DatabaseHost(),
				"-P", conf.DatabasePortString(),
				"-u", conf.DatabaseUser(),
				"-p"+conf.DatabasePassword(),
				"-f",
				conf.DatabaseName(),
			)
		case config.SQLite3:
			log.Infoln("dropping existing tables")
			tables.Drop(conf.Db())
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

		// Log exact command for debugging in trace mode.
		log.Trace(cmd.String())

		// Run backup command.
		if err := cmd.Run(); err != nil {
			if stderr.String() != "" {
				log.Debugln(stderr.String())
				log.Warnf("index could not be restored completely")
			}
		}
	}

	log.Infoln("migrating index database schema")

	conf.InitDb()

	if restoreAlbums {
		get.SetConfig(conf)

		if albumsPath == "" {
			albumsPath = conf.AlbumsPath()
		}

		if !fs.PathExists(albumsPath) {
			log.Warnf("album files path %s not found", clean.Log(albumsPath))
		} else {
			log.Infof("restoring albums from %s", clean.Log(albumsPath))

			if count, err := photoprism.RestoreAlbums(albumsPath, true); err != nil {
				return err
			} else {
				log.Infof("restored %s from YAML files", english.Plural(count, "album", "albums"))
			}
		}
	}

	elapsed := time.Since(start)

	log.Infof("restored in %s", elapsed)

	return nil
}
