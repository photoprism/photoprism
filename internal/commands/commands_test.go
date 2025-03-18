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
	"github.com/photoprism/photoprism/pkg/capture"
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
	_ = c.CloseDb()

	os.Exit(code)
}

// NewTestContext creates a new CLI test context with the flags and arguments provided.
func NewTestContext(args []string) *cli.Context {
	// Create new command-line test app.
	app := cli.NewApp()
	app.Name = "photoprism"
	app.Usage = "PhotoPrism®"
	app.Description = ""
	app.Version = "test"
	app.Copyright = "(c) 2018-2025 PhotoPrism UG. All rights reserved."
	app.Flags = config.Flags.Cli()
	app.Commands = PhotoPrism
	app.HelpName = app.Name
	app.CustomAppHelpTemplate = ""
	app.HideHelp = true
	app.HideHelpCommand = true
	app.Action = func(*cli.Context) error { return nil }
	app.EnableBashCompletion = false
	app.Metadata = map[string]interface{}{
		"Name":    "PhotoPrism",
		"About":   "PhotoPrism®",
		"Edition": "ce",
		"Version": "test",
	}

	// Parse command test arguments.
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	LogErr(flagSet.Parse(args))

	// Create and return new test context.
	return cli.NewContext(app, flagSet, nil)
}

// RunWithTestContext executes a command with a test context and returns its output.
func RunWithTestContext(cmd *cli.Command, args []string) (output string, err error) {
	// Create test context with flags and arguments.
	ctx := NewTestContext(args)

	// TODO: Help output can currently not be generated in test mode due to
	//       a nil pointer panic in the "github.com/urfave/cli/v2" package.
	cmd.HideHelp = true

	// Run command with test context.
	output = capture.Output(func() {
		err = cmd.Run(ctx, args...)
	})

	return output, err
}
