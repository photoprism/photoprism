package commands

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/testextras"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	caller := "internal/commands/commands_test.go/TestMain"
	dbc, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		os.Exit(1)
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	c := config.NewTestConfig("commands")
	get.SetConfig(c)

	// Init config and connect to database.
	InitConfig = func(ctx *cli.Context) (*config.Config, error) {
		return c, c.Init()
	}

	// Run unit tests.
	beforeTimestamp := time.Now().UTC()
	code := m.Run()
	code = testextras.ValidateDBErrors(dbc.Db(), log, beforeTimestamp, code)

	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)

	// Close database connection.
	c.CloseDb()

	os.Exit(code)
}

// NewTestContext creates a new CLI test context with the flags and arguments provided.
func NewTestContext(args []string) *cli.Context {
	// Create new command-line app.
	app := cli.NewApp()
	app.Usage = "PhotoPrism®"
	app.Version = "test"
	app.Copyright = "(c) 2018-2024 PhotoPrism UG. All rights reserved."
	app.EnableBashCompletion = true
	app.Flags = config.Flags.Cli()
	app.Metadata = map[string]interface{}{
		"Name":    "PhotoPrism",
		"About":   "PhotoPrism®",
		"Edition": "ce",
		"Version": "test",
	}

	// Parse command arguments.
	flags := flag.NewFlagSet("test", 0)
	LogErr(flags.Parse(args))

	// Create and return new context.
	return cli.NewContext(app, flags, nil)
}
