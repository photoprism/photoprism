package commands

import (
	"bytes"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/testextras"
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

	caller := "internal/commands/commands_test.go/TestMain"
	dbc, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		os.Exit(1)
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	tempDir, err := os.MkdirTemp("", "commands-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	c := config.NewMinimalTestConfigWithDb("commands", tempDir)
	get.SetConfig(c)

	// Keep DB connection open for the duration of this package's tests to
	// avoid late access after CloseDb() in concurrent test runs.

	// Init config and connect to database.
	InitConfig = func(ctx *cli.Context) (*config.Config, error) {
		return c, c.Init()
	}

	// Run unit tests.
	beforeTimestamp := time.Now().UTC()
	code := m.Run()
	code = testextras.ValidateDBErrors(dbc.Db(), log, beforeTimestamp, code)

	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)

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
	return cli.NewContext(app, flagSet, cli.NewContext(app, flagSet, nil))
}

// RunWithTestContext executes a command with a test context and returns its output.
func RunWithTestContext(cmd *cli.Command, args []string) (output string, err error) {
	// Create test context with flags and arguments.
	ctx := NewTestContext(args)

	cmd.HideHelp = false

	// Ensure DB connection is open for each command run (some commands call Shutdown).
	if c := get.Config(); c != nil {
		c.RegisterDb() // (re)register provider
	}

	// Redirect the output from cli to buffer for transfer to output for testing
	var catureOutput bytes.Buffer
	oldWriter := ctx.App.Writer
	ctx.App.Writer = &catureOutput

	// Run command with test context.
	output = capture.Output(func() {
		origExiter := cli.OsExiter
		cli.OsExiter = func(int) {}
		defer func() { cli.OsExiter = origExiter }()
		err = cmd.Run(ctx, args...)
	})
	ctx.App.Writer = oldWriter
	output += catureOutput.String()

	// Re-open the database after the command completed so follow-up checks
	// (potentially issued by the test itself) have an active connection.
	if c := get.Config(); c != nil {
		c.RegisterDb()
	}

	return output, err
}

// NewTestContextWithParse creates a new CLI test context with the flags and arguments provided.
func NewTestContextWithParse(appArgs []string, cmdArgs []string) *cli.Context {
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

	// Parse photoprism command arguments.
	photoprismFlagSet := flag.NewFlagSet("photoprism", flag.ContinueOnError)
	for _, f := range app.Flags {
		f.Apply(photoprismFlagSet)
	}
	LogErr(photoprismFlagSet.Parse(appArgs[1:]))

	// Parse command test arguments.
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	LogErr(flagSet.Parse(cmdArgs))

	// Create and return new test context.
	return cli.NewContext(app, flagSet, cli.NewContext(app, photoprismFlagSet, nil))
}

func RunWithProvidedTestContext(ctx *cli.Context, cmd *cli.Command, args []string) (output string, err error) {
	// Redirect the output from cli to buffer for transfer to output for testing
	var catureOutput bytes.Buffer
	oldWriter := ctx.App.Writer
	ctx.App.Writer = &catureOutput

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
	ctx.App.Writer = oldWriter
	output += catureOutput.String()

	// Re-open the database after the command completed so follow-up checks
	// (potentially issued by the test itself) have an active connection.
	if c := get.Config(); c != nil {
		c.RegisterDb()
	}

	return output, err
}
