package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
)

const databaseTransferDescription = `Copies the index database to an empty database of the same or another kind, e.g. from SQLite to MariaDB.
The target is specified with --target-driver and either --target-dsn or the server, name, user, and password flags,
which can also be set with the PHOTOPRISM_TRANSFER_* environment variables. Settings that are not specified get
their defaults, not the values of the configured database. A SQLite target is specified with --target-dsn only.

Stop PhotoPrism before running a transfer, as changes made while it runs are missing in the target and are not
always detected. The configured database is migrated first, then the schema of the target. Every table is copied as stored from one snapshot, including users, sessions, and settings. A target
that holds data or other tables is refused unless --force is passed, in which case its data is replaced and
other tables are left as they are. If the transfer fails, the target is emptied again, and a SQLite file it
created is removed. Once it is complete, change the database options to the target and start PhotoPrism.

A database backup can only be restored into the same kind of database, so use this command to change the kind.`

// DatabaseTransferCommand configures the command name, flags, and action.
var DatabaseTransferCommand = &cli.Command{
	Name:        "transfer",
	Description: databaseTransferDescription,
	Usage:       "Copies the index database to another database",
	Flags:       databaseTransferFlags(),
	Action:      databaseTransferAction,
}

// MigrationsTransferCommand is a hidden alias of DatabaseTransferCommand.
var MigrationsTransferCommand = &cli.Command{
	Name:        "transfer",
	Description: databaseTransferDescription,
	Usage:       DatabaseTransferCommand.Usage,
	Flags:       databaseTransferFlags(),
	Action:      databaseTransferAction,
	Hidden:      true,
}

// databaseTransferFlags returns the flags of the transfer command, as a new slice for each command.
func databaseTransferFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "target-driver",
			Usage:   "target database `DRIVER` (sqlite, mariadb, or postgres)",
			EnvVars: []string{config.EnvVar("transfer-driver")},
		},
		&cli.StringFlag{
			Name:    "target-dsn",
			Usage:   "target database `DSN`, e.g. a SQLite filename or a PostgreSQL connection URI",
			EnvVars: []string{config.EnvVar("transfer-dsn")},
		},
		&cli.StringFlag{
			Name:    "target-server",
			Usage:   "target database `HOST` incl. port, e.g. \"mariadb:3306\" (or socket path)",
			EnvVars: []string{config.EnvVar("transfer-server")},
		},
		&cli.StringFlag{
			Name:    "target-name",
			Usage:   "target database schema `NAME`",
			EnvVars: []string{config.EnvVar("transfer-name")},
		},
		&cli.StringFlag{
			Name:    "target-user",
			Usage:   "target database user `NAME`",
			EnvVars: []string{config.EnvVar("transfer-user")},
		},
		&cli.StringFlag{
			Name:    "target-password",
			Usage:   "target database user `PASSWORD` (the environment variable or a password file keep it out of the process list)",
			EnvVars: []string{config.EnvVar("transfer-password")},
		},
		&cli.PathFlag{
			Name:      "target-password-file",
			Usage:     "`FILENAME` with the target database user password",
			EnvVars:   []string{config.EnvVar("transfer-password-file")},
			TakesFile: true,
		},
		ForceFlag("replaces the data in a target database that is not empty"),
		&cli.BoolFlag{
			Name:  "allow-failed-migrations",
			Usage: "copies the index even if migrations have failed, which the target then records as successful",
		},
		&cli.IntFlag{
			Name:    "batch",
			Aliases: []string{"b"},
			Usage:   "maximum `NUMBER` of rows copied per statement, lower if the target server limits the packet size",
			Value:   entity.TransferBatchSize,
		},
	}
}

// databaseTransferTarget returns the target database settings passed with the command flags.
func databaseTransferTarget(ctx *cli.Context) (config.DatabaseOptions, error) {
	target := config.DatabaseOptions{
		Driver:   strings.TrimSpace(ctx.String("target-driver")),
		DSN:      strings.TrimSpace(ctx.String("target-dsn")),
		Server:   strings.TrimSpace(ctx.String("target-server")),
		Name:     strings.TrimSpace(ctx.String("target-name")),
		User:     strings.TrimSpace(ctx.String("target-user")),
		Password: ctx.String("target-password"),
	}

	if target.Driver == "" && target.DSN != "" {
		target.Driver = dsn.Parse(target.DSN).Driver
	}

	if target.Driver == "" {
		return target, errors.New("target database driver must be specified with --target-driver")
	}

	// A DSN contains all connection settings, so other flags would be ignored.
	if target.DSN != "" && (target.Server != "" || target.Name != "" || target.User != "" ||
		target.Password != "" && dsn.ParseDriver(target.Driver) != dsn.DriverSQLite3) {
		return target, errors.New("--target-dsn cannot be combined with the server, name, user, or password flags")
	}

	switch dsn.ParseDriver(target.Driver) {
	case dsn.DriverMySQL, dsn.DriverMariaDB:
		target.Driver = dsn.DriverMySQL
	case dsn.DriverPostgres:
		target.Driver = dsn.DriverPostgres
	case dsn.DriverSQLite3:
		target.Driver = dsn.DriverSQLite3

		// A SQLite database is specified by its filename alone and has no password.
		if target.Server != "" || target.Name != "" || target.User != "" {
			return target, errors.New("a SQLite target is specified with --target-dsn only")
		}

		target.Password = ""

		// A relative filename refers to the current directory, not the storage path.
		if target.DSN != "" && !strings.HasPrefix(target.DSN, "file:") {
			if fileName, params, found := strings.Cut(target.DSN, "?"); fileName != "" {
				if abs, err := filepath.Abs(fileName); err == nil {
					target.DSN = abs

					if found {
						target.DSN += "?" + params
					}
				}
			}
		}
	default:
		return target, fmt.Errorf("unsupported target database driver %s, use sqlite, mariadb, or postgres", clean.Log(target.Driver))
	}

	// The password file is only read for a server that is not specified by a DSN.
	if fileName := ctx.String("target-password-file"); fileName == "" || target.Password != "" || target.DSN != "" ||
		target.Driver == dsn.DriverSQLite3 {
		return target, nil
	} else if b, err := os.ReadFile(fileName); err != nil { //nolint:gosec // the file is chosen by the operator
		return target, fmt.Errorf("failed to read the target password file: %w", err)
	} else {
		target.Password = clean.Password(string(b))
	}

	return target, nil
}

// databaseTransferAction copies the index database to another database.
func databaseTransferAction(ctx *cli.Context) error {
	if ctx.Args().Present() {
		return cli.Exit(errors.New("unexpected argument, the target is specified with flags"), 2)
	}

	batchSize := ctx.Int("batch")

	if batchSize < 1 {
		return cli.Exit(errors.New("--batch must be at least 1"), 2)
	}

	targetOpt, err := databaseTransferTarget(ctx)

	if err != nil {
		return cli.Exit(err, 2)
	}

	start := time.Now()

	conf, err := InitConfig(ctx)

	if err != nil {
		return cli.Exit(err, 1)
	}

	defer conf.Shutdown()

	log.Infof("transfer: migrating the %s index database", conf.DatabaseDriverName())

	if conf.RegisterDb(); conf.DbIfConnected() == nil {
		return cli.Exit(errors.New("failed to connect to the index database"), 1)
	} else if err = conf.MigrateDb(false, nil); err != nil {
		return cli.Exit(fmt.Errorf("failed to migrate the index database: %w", err), 1)
	} else if failed, failedErr := failedMigrations(conf.Db()); failedErr != nil {
		return cli.Exit(fmt.Errorf("failed to check the index database migrations: %w", failedErr), 1)
	} else if len(failed) == 0 {
		// All migrations have been successful.
	} else if !ctx.Bool("allow-failed-migrations") {
		return cli.Exit(fmt.Errorf("index database migrations %s have failed, run \"photoprism migrations run --failed\" first", strings.Join(failed, ", ")), 1)
	} else {
		log.Warnf("transfer: the target database records migrations %s as successful", strings.Join(failed, ", "))
	}

	// Tables of other editions or applications are not copied.
	if unknown, unknownErr := entity.Entities.Unknown(conf.Db()); unknownErr != nil {
		return cli.Exit(fmt.Errorf("failed to list the index database tables: %w", unknownErr), 1)
	} else if len(unknown) > 0 {
		log.Warnf("transfer: %s not copied (%s)", english.PluralWord(len(unknown), "table is", "tables are"), strings.Join(unknown, ", "))
	}

	target := conf.WithDatabase(targetOpt)

	// A SQLite file that the transfer creates is removed again if it does not complete.
	newFile := ""

	if target.DatabaseDriver() == dsn.DriverSQLite3 {
		if _, statErr := os.Lstat(target.DatabaseFile()); os.IsNotExist(statErr) {
			newFile = target.DatabaseFile()
		}
	}

	if err = target.ConnectDb(); err != nil {
		removeSQLiteFiles(newFile)
		return cli.Exit(fmt.Errorf("failed to connect to the target database: %w", err), 1)
	}

	completed := false

	defer func() {
		_ = entity.CloseDb(target.DbIfConnected())

		if !completed {
			removeSQLiteFiles(newFile)
		}
	}()

	if same, sameErr := entity.SameDatabase(conf.Db(), target.Db()); sameErr != nil {
		return cli.Exit(fmt.Errorf("failed to compare the databases: %w", sameErr), 1)
	} else if same {
		return cli.Exit(errors.New("the target is the index database, choose another database"), 2)
	}

	// A target is empty if no registered data table has rows and it holds no other tables.
	if rows, rowsErr := entity.Entities.Rows(target.Db()); rowsErr != nil {
		return cli.Exit(fmt.Errorf("failed to check the target database: %w", rowsErr), 1)
	} else if unknown, unknownErr := entity.Entities.Unknown(target.Db()); unknownErr != nil {
		return cli.Exit(fmt.Errorf("failed to list the target database tables: %w", unknownErr), 1)
	} else if len(rows) > 0 || len(unknown) > 0 {
		tables := unknown

		for name := range rows {
			tables = append(tables, name)
		}

		slices.Sort(tables)

		if !ctx.Bool("force") {
			return cli.Exit(fmt.Errorf("the target database is not empty (%s), use --force to replace its data", strings.Join(tables, ", ")), 2)
		}

		log.Infof("transfer: replacing the data in the target database")
	}

	log.Infof("transfer: migrating the %s target database", target.DatabaseDriverName())

	if err = target.MigrateDbSchema(); err != nil {
		return cli.Exit(fmt.Errorf("failed to migrate the target database: %w", err), 1)
	}

	log.Infof("transfer: copying the index to the %s target database", target.DatabaseDriverName())

	result, err := entity.Entities.Transfer(conf.Db(), target.Db(), entity.TransferOptions{
		BatchSize: batchSize,
		Progress: func(table string, rows int64) {
			log.Infof("transfer: copied %s from %s", english.Plural(int(rows), "row", "rows"), table)
		},
	})

	if err != nil {
		return cli.Exit(fmt.Errorf("failed to transfer the index: %w", err), 1)
	}

	completed = true

	var total int64

	for _, rows := range result {
		total += rows
	}

	log.Infof("transfer: copied %s in %s", english.Plural(int(total), "row", "rows"), time.Since(start))
	log.Infof("transfer: change the database options to the target database and restart PhotoPrism to use it")

	return nil
}

// failedMigrations returns the IDs of the schema migrations of the database's dialect that failed in it, and
// can be run again; other rows, e.g. of migrations that no longer exist, are ignored.
func failedMigrations(db *gorm.DB) (ids []string, err error) {
	var failed []string

	if err = db.Model(&migrate.Migration{}).Where("error <> ''").Order("id").Pluck("id", &failed).Error; err != nil {
		return nil, err
	}

	known := migrate.Dialects[dsn.ParseDriver(db.Name())]

	for _, id := range failed {
		if slices.ContainsFunc(known, func(m migrate.Migration) bool { return m.ID == id }) {
			ids = append(ids, id)
		}
	}

	return ids, nil
}

// removeSQLiteFiles removes a SQLite database file and its journal files, if fileName is not empty. Names that
// are not regular files, e.g. symbolic links or directories, are left as they are.
func removeSQLiteFiles(fileName string) {
	if fileName == "" {
		return
	}

	for _, name := range []string{fileName, fileName + "-journal", fileName + "-wal", fileName + "-shm"} {
		if info, err := os.Lstat(name); err != nil || !info.Mode().IsRegular() {
			continue
		} else if err = os.Remove(name); err != nil {
			log.Warnf("transfer: %s", clean.Error(err))
		}
	}
}
