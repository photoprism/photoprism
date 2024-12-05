package commands

import (
	"flag"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	c := config.NewTestConfig("commands")
	get.SetConfig(c)

	// Init config and connect to database.
	InitConfig = func(ctx *cli.Context) (*config.Config, error) {
		return c, c.Init()
	}

	// Run unit tests.
	code := m.Run()

	// Close database connection.
	c.CloseDb()

	os.Exit(code)
}

// NewTestContext creates a new CLI test context with the flags and arguments provided.
func NewTestContext(args []string) *cli.Context {
	// Create new command-line app.
	app := cli.NewApp()
	app.Name = "photoprism"
	app.Usage = "PhotoPrism®"
	app.Description = ""
	app.Version = "test"
	app.Copyright = "(c) 2018-2024 PhotoPrism UG. All rights reserved."
	app.Flags = config.Flags.Cli()
	app.Commands = PhotoPrism
	app.EnableBashCompletion = false
	app.Metadata = map[string]interface{}{
		"Name":    "PhotoPrism",
		"About":   "PhotoPrism®",
		"Edition": "ce",
		"Version": "test",
	}

	// Parse command arguments.
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	LogErr(set.Parse(args))

	// Create and return new context.
	return cli.NewContext(app, set, nil)
}
