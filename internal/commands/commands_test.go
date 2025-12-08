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
	"github.com/photoprism/photoprism/pkg/fs"
)

// TODO: Several CLI commands defer conf.Shutdown(), which closes the shared
// database connection. To avoid flakiness, RunWithTestContext re-initializes
// and re-registers the DB provider before each command invocation. If you see
// "config: database not connected" during test runs, consider moving shutdown
// behavior behind an interface or gating it for tests.

func TestMain(m *testing.M) {
	_ = os.Setenv("TF_CPP_MIN_LOG_LEVEL", "3")

	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)

	tempDir, err := os.MkdirTemp("", "commands-test")
	if err != nil {
		panic(err)
	}

	c := config.NewMinimalTestConfigWithDb("commands", tempDir)
	get.SetConfig(c)

	// Keep DB connection open for the duration of this package's tests to
	// avoid late access after CloseDb() in concurrent test runs.

	// Init config and connect to database.
	InitConfig = func(ctx *cli.Context) (*config.Config, error) {
		return c, c.Init()
	}

	// Run unit tests.
	code := m.Run()

	if err = c.CloseDb(); err != nil {
		log.Errorf("close db: %v", err)
	}

	_ = os.RemoveAll(tempDir)

	// Remove temporary SQLite files after running the tests.
	fs.PurgeTestDbFiles(".", false)

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

	// Ensure DB connection is open for each command run (some commands call Shutdown).
	if c := get.Config(); c != nil {
		c.RegisterDb() // (re)register provider
	}

	// Run command via cli.Command.Run but neutralize os.Exit so ExitCoder
	// errors don't terminate the test binary.
	output = capture.Output(func() {
		origExiter := cli.OsExiter
		cli.OsExiter = func(int) {}
		defer func() { cli.OsExiter = origExiter }()
		err = cmd.Run(ctx, args...)
	})

	// Re-open the database after the command completed so follow-up checks
	// (potentially issued by the test itself) have an active connection.
	if c := get.Config(); c != nil {
		c.RegisterDb()
	}

	return output, err
}
