package commands

import (
	"flag"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/get"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	c := config.NewTestConfig("commands")
	get.SetConfig(c)

	InitConfig = func(ctx *cli.Context) (*config.Config, error) {
		return c, c.Init()
	}

	code := m.Run()

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
