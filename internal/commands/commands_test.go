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

	// Init core config (no database) using the shared test config so commands
	// like "show config" and "faces config" don't fall back to a storage path
	// derived from the real originals directory.
	InitCoreConfig = func(ctx *cli.Context, quiet bool) (*config.Config, error) {
		return c, c.InitCore()
	}

	// Run unit tests.
	code := m.Run()

	if err = c.CloseDb(); err != nil {
		log.Warnf("close db: %v", err)
	}

	_ = os.RemoveAll(tempDir)

	// Remove temporary SQLite files after running the tests.
	fs.PurgeTestDbFiles(".", false)

	os.Exit(code)
}

// SetEnvForTest sets an environment variable and restores its original value after the test.
func SetEnvForTest(t *testing.T, key, value string) {
	t.Helper()

	previous, hadPrevious := os.LookupEnv(key)

	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("set env %s: %v", key, err)
	}

	t.Cleanup(func() {
		var restoreErr error

		if hadPrevious {
			restoreErr = os.Setenv(key, previous)
		} else {
			restoreErr = os.Unsetenv(key)
		}

		if restoreErr != nil {
			t.Errorf("restore env %s: %v", key, restoreErr)
		}
	})
}

// NewTestContext creates a new CLI test context with the flags and arguments provided.
func NewTestContext(args []string) *cli.Context {
	// Create new command-line test app.
	app := cli.NewApp()
	app.Name = "photoprism"
	app.Usage = "PhotoPrism®"
	app.Description = ""
	app.Version = "test"
	app.Copyright = "(c) 2018-2026 PhotoPrism UG. All rights reserved."
	app.Flags = config.Flags.Cli()
	app.Commands = PhotoPrism
	app.HelpName = app.Name
	app.CustomAppHelpTemplate = ""
	app.HideHelp = true
	app.HideHelpCommand = true
	app.Action = func(*cli.Context) error { return nil }
	app.EnableBashCompletion = false
	app.Metadata = map[string]any{
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

	// Snapshot the database connection settings before running the command. A
	// command that rewrites them (e.g. cluster register persisting provisioned
	// MySQL credentials) would otherwise make the post-command RegisterDb() below
	// block in connectDb's 60s retry loop reaching for a server that isn't there.
	c := get.Config()
	var o *config.Options
	var driver, dsn, name, server, user, password string
	if c != nil {
		o = c.Options()
		driver, dsn, name = o.DatabaseDriver, o.DatabaseDSN, o.DatabaseName
		server, user, password = o.DatabaseServer, o.DatabaseUser, o.DatabasePassword
		// Ensure DB connection is open for each command run (some commands call Shutdown).
		c.RegisterDb()
	}

	// Run command via cli.Command.Run but neutralize os.Exit so ExitCoder
	// errors don't terminate the test binary.
	output = capture.Output(func() {
		origExiter := cli.OsExiter
		cli.OsExiter = func(int) {}
		defer func() { cli.OsExiter = origExiter }()
		err = cmd.Run(ctx, args...)
	})

	// Restore the original connection settings, then re-open so follow-up checks
	// (potentially issued by the test itself) reconnect to the per-suite SQLite
	// test DB rather than whatever server the command may have configured.
	if c != nil {
		o.DatabaseDriver, o.DatabaseDSN, o.DatabaseName = driver, dsn, name
		o.DatabaseServer, o.DatabaseUser, o.DatabasePassword = server, user, password
		c.RegisterDb()
	}

	return output, err
}
