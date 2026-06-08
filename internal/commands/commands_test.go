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
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

var savedPath string

// TODO: Several CLI commands defer conf.Shutdown(), which closes the shared
// database connection. To avoid flakiness, RunWithTestContext re-initializes
// and re-registers the DB provider before each command invocation. If you see
// "config: database not connected" during test runs, consider moving shutdown
// behavior behind an interface or gating it for tests.
func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) (code int) {
	_ = os.Setenv("TF_CPP_MIN_LOG_LEVEL", "3")

	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)

	caller := "internal/commands/commands_test.go/TestMain"
	dbc, dbn, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		os.Exit(1)
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	_, dsname := dsn.PhotoPrismTestToDriverDSN(dbn)
	dsn.SetDSNToEnv(dsname)

	tempDir, err := os.MkdirTemp("", "commands-test")
	if err != nil {
		panic(err)
	}
	savedPath = tempDir

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
	beforeTimestamp := time.Now().UTC()
	code = m.Run()
	code = testextras.ValidateDBErrors(c.Db(), log, beforeTimestamp, code)

	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)

	if err = c.CloseDb(); err != nil {
		log.Warnf("close db: %v", err)
	}

	_ = os.RemoveAll(tempDir)

	// Remove temporary SQLite files after running the tests.
	fs.PurgeTestDbFiles(".", false)

	return code
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
	app.HideHelp = false
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
	return cli.NewContext(app, flagSet, cli.NewContext(app, flagSet, nil))
}

// RunWithTestContext executes a command with a test context and returns its output.
func RunWithTestContext(cmd *cli.Command, args []string) (output string, err error) {
	return RunWithProvidedTestContext(NewTestContext(args), cmd, args)
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
	app.HideHelp = false
	app.HideHelpCommand = true
	app.Action = func(*cli.Context) error { return nil }
	app.EnableBashCompletion = false
	app.Metadata = map[string]any{
		"Name":    "PhotoPrism",
		"About":   "PhotoPrism®",
		"Edition": "ce",
		"Version": "test",
	}

	// Parse photoprism command arguments.
	photoprismFlagSet := flag.NewFlagSet("photoprism", flag.ContinueOnError)
	for _, f := range app.Flags {
		LogErr(f.Apply(photoprismFlagSet))
	}
	LogErr(photoprismFlagSet.Parse(appArgs[1:]))

	// Parse command test arguments.
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	LogErr(flagSet.Parse(cmdArgs))

	// Create and return new test context.
	return cli.NewContext(app, flagSet, cli.NewContext(app, photoprismFlagSet, nil))
}

func RunWithProvidedTestContext(ctx *cli.Context, cmd *cli.Command, args []string) (output string, err error) {
	// Ensure DB connection is open for each command run (some commands call Shutdown).
	_ = reopenConnection()
	conf := get.Config()
	previousOptions := *conf.Options()
	// Redirect the output from cli to buffer for transfer to output for testing
	var catureOutput bytes.Buffer
	oldWriter := ctx.App.Writer
	ctx.App.Writer = &catureOutput

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

	// Reset the config options just in case they have been affected
	*conf.Options() = previousOptions
	// // Re-open the database after the command completed so follow-up checks
	// // (potentially issued by the test itself) have an active connection.
	_ = reopenConnection()

	return output, err
}

// resetConfigAndDB replaces the config with a generated minimal config, and may replace the database if it doesn't exist.
// it does call Migrate and TestFixtures for Postgres and MariaDB.  It may call Migrate and TestFixtures for SQLite if the database
// doesn't exist.  That can only happen if you are using PHOTOPRISM_TEST_DSN_NAME="sqlite".
func resetConfigAndDB() *config.Config {
	c := config.NewMinimalTestConfigWithDb("commands", savedPath)
	get.SetConfig(c)
	entity.SetDbProvider(c)

	InitConfig = func(ctx *cli.Context) (*config.Config, error) {
		return c, c.Init()
	}

	return c
}

// resetConfigAndOpenDB replaces the config with a generated minimal config, and opens the configured database.
// it does not call Migrate and TestFixtures if the database has records in auth_users and photos.
func resetConfigAndOpenDB() *config.Config {
	c := config.NewMinimalTestConfig("commands", savedPath)
	config.RestoreDBFromCache(c) // If using sqlite (not sqlitefile) then the db is removed by NewMinimalTestConfig
	if err := c.Init(); err != nil {
		log.Fatalf("config: %s (init)", err.Error())
	}
	get.SetConfig(c)
	entity.SetDbProvider(c)

	InitConfig = func(ctx *cli.Context) (*config.Config, error) {
		return c, c.Init()
	}

	return c
}

// reopenConnection gets the current configured connection and opens it if it is closed.
// It returns the current config to allow queries in tests if needed.
func reopenConnection() *config.Config {
	if c := get.Config(); c != nil {
		if !c.IsDbOpen() {
			c.RegisterDb()
		} else {
			entity.SetDbProvider(c) // entity can get out of sync with c, so make sure it's correct
		}
		InitConfig = func(ctx *cli.Context) (*config.Config, error) {
			return c, c.Init()
		}
		return c
	} else {
		log.Warn("reopenConnection: config is nil")
		return nil
	}
}
