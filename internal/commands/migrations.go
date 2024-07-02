package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/report"
)

var MigrationsStatusCommand = cli.Command{
	Name:      "ls",
	Aliases:   []string{"status", "show"},
	Usage:     "Displays the status of schema migrations",
	ArgsUsage: "[migrations...]",
	Flags:     report.CliFlags,
	Action:    migrationsStatusAction,
}

var MigrationsRunCommand = cli.Command{
	Name:      "run",
	Aliases:   []string{"execute", "migrate"},
	Usage:     "Executes database schema migrations",
	ArgsUsage: "[migrations...]",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "failed, f",
			Usage: "run previously failed migrations",
		},
		cli.BoolFlag{
			Name:  "trace, t",
			Usage: "show trace logs for debugging",
		},
	},
	Action: migrationsRunAction,
}

// MigrationsCommands registers the "migrations" CLI command.
var MigrationsCommands = cli.Command{
	Name:  "migrations",
	Usage: "Database schema migration subcommands",
	Subcommands: []cli.Command{
		MigrationsStatusCommand,
		MigrationsRunCommand,
	},
}

// migrationsStatusAction lists the status of schema migration.
func migrationsStatusAction(ctx *cli.Context) error {
	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.RegisterDb()
	defer conf.Shutdown()

	var ids []string

	// Check argument for specific migrations to be run.
	if migrations := strings.TrimSpace(ctx.Args().First()); migrations != "" {
		ids = strings.Fields(migrations)
	}

	db := conf.Db()

	status, err := migrate.Status(db, ids)

	if err != nil {
		return err
	}

	// Report columns.
	cols := []string{"ID", "Dialect", "Stage", "Started At", "Finished At", "Status"}

	// Report rows.
	rows := make([][]string, 0, len(status))

	for _, m := range status {
		var stage, started, finished, info string

		if m.Stage == "" {
			stage = "main"
		} else {
			stage = m.Stage
		}

		if m.StartedAt.IsZero() {
			started = "-"
		} else {
			started = m.StartedAt.Format("2006-01-02 15:04:05")
		}

		if m.Finished() {
			finished = m.FinishedAt.Format("2006-01-02 15:04:05")
		} else {
			finished = "-"
		}

		if m.Error != "" {
			info = m.Error
		} else if m.Finished() {
			info = "OK"
		} else if m.StartedAt.IsZero() {
			info = "-"
		} else if m.Repeat(false) {
			info = "Repeat"
		} else {
			info = "Running?"
		}

		rows = append(rows, []string{m.ID, m.Dialect, stage, started, finished, info})
	}

	// Display report.
	info, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

	if err != nil {
		return err
	}

	fmt.Println(info)

	return nil
}

// migrationsRunAction executes database schema migrations.
func migrationsRunAction(ctx *cli.Context) error {
	if ctx.Args().First() == "ls" {
		return fmt.Errorf("run '%s migrations ls' to display the status of schema migrations", filepath.Base(os.Args[0]))
	}

	start := time.Now()

	conf := config.NewConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.RegisterDb()
	defer conf.Shutdown()

	if ctx.Bool("trace") {
		log.SetLevel(logrus.TraceLevel)
		log.Infoln("migrate: enabled trace mode")
	}

	runFailed := ctx.Bool("failed")

	if runFailed {
		log.Infoln("migrate: running previously failed migrations")
	}

	var ids []string

	// Check argument for specific migrations to be run.
	if migrations := strings.TrimSpace(ctx.Args().First()); migrations != "" {
		ids = strings.Fields(migrations)
	}

	log.Infoln("migrating database schema...")

	// Run migrations.
	conf.MigrateDb(runFailed, ids)

	elapsed := time.Since(start)

	log.Infof("completed in %s", elapsed)

	return nil
}
