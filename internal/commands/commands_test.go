package commands

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync/atomic"
	"testing"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/log/status"
)

var savedPath string
var initialTestDbDSN string
var resetDbSequence atomic.Uint64

// nextResetDbName returns a distinct alphabetic SQLite name for each config.
func nextResetDbName(prefix string) string {
	n := resetDbSequence.Add(1)
	var suffix [14]byte

	for i := len(suffix) - 1; i >= 0; i-- {
		suffix[i] = 'a' + byte(n%26)
		n /= 26
	}

	return prefix + string(suffix[:])
}

// TODO: Several CLI commands defer conf.Shutdown(), which closes the shared
// database connection. To avoid flakiness, RunWithTestContext re-initializes
// and re-registers the DB provider before each command invocation. If you see
// "config: database not connected" during test runs, consider moving shutdown
// behavior behind an interface or gating it for tests.

// TestMain executes runTestMain returning it's results.  It is done this way so that defer can be used to cleanup.
func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) int {
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
	savedPath = tempDir
	defer os.RemoveAll(tempDir)

	c := config.NewMinimalTestConfigWithDb("commands", tempDir)
	initialTestDbDSN = c.DatabaseDSN()
	defer c.CleanupTestFolder()
	defer func() {
		if err := c.CloseDb(); err != nil {
			log.Warnf("close db: %v", err)
		}
		// Remove temporary SQLite files after running the tests.
		fs.PurgeTestDbFiles(".", false)
	}()

	get.SetConfig(c)

	// Keep DB connection open for the duration of this package's tests to
	// avoid late access after CloseDb() in concurrent test runs.

	// Init config and connect to database.
	InitConfig = func(ctx *cli.Context) (*config.Config, error) {
		return c, c.Init()
	}

	// Use the current test config for core-only commands.
	InitCoreConfig = func(ctx *cli.Context, quiet bool) (*config.Config, error) {
		current := get.Config()
		return current, current.InitCore()
	}

	// Run unit tests.
	return m.Run()
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
	// cli.NewContext(app, flagSet, nil) will cause a Panic if HideHelp = false.  You must provide a context in the OUTER call.
	return cli.NewContext(app, flagSet, cli.NewContext(app, photoprismFlagSet, nil))
}

func RunWithProvidedTestContext(ctx *cli.Context, cmd *cli.Command, args []string) (output string, err error) {
	// Ensure DB connection is open for each command run (some commands call Shutdown).
	_ = reopenConnection()
	conf := get.Config()
	previousOptions := *conf.Options()
	// Redirect the output from cli to buffer for transfer to output for testing
	var captureOutput bytes.Buffer
	oldWriter := ctx.App.Writer
	ctx.App.Writer = &captureOutput

	// Run command via cli.Command.Run but neutralize os.Exit so ExitCoder
	// errors don't terminate the test binary.
	output = capture.Output(func() {
		origExiter := cli.OsExiter
		cli.OsExiter = func(int) {}
		defer func() { cli.OsExiter = origExiter }()
		err = cmd.Run(ctx, args...)
	})
	ctx.App.Writer = oldWriter
	output += captureOutput.String()

	// Reset the config options just in case they have been affected
	*conf.Options() = previousOptions
	// // Re-open the database after the command completed so follow-up checks
	// // (potentially issued by the test itself) have an active connection.
	_ = reopenConnection()

	return output, err
}

// resetConfigAndDB replaces the config with a distinct fixture-backed database.
func resetConfigAndDB() *config.Config {
	c := config.NewMinimalTestConfigWithDb(nextResetDbName("commands"), savedPath)
	get.SetConfig(c)
	entity.SetDbProvider(c)

	InitConfig = func(ctx *cli.Context) (*config.Config, error) {
		return c, c.Init()
	}

	return c
}

// resetConfigAndOpenDB opens a distinct cached test database for each config.
func resetConfigAndOpenDB() *config.Config {
	c := config.NewIsolatedTestConfig(nextResetDbName("commandsreset"), savedPath, false)
	config.RestoreDBFromCache(c) // With SQLite, NewIsolatedTestConfig removes the database file first.
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

// TestNextResetDbName checks that rollover names stay distinct and valid.
func TestNextResetDbName(t *testing.T) {
	seen := make(map[string]bool)

	for i := 0; i < 27; i++ {
		name := nextResetDbName("commands")
		require.Equal(t, name, config.PkgNameRegexp.ReplaceAllString(name, ""))
		require.False(t, seen[name])
		seen[name] = true
	}
}

// TestResetConfigAndOpenDB checks that the database stays writable when other tests build minimal configs.
func TestResetConfigAndOpenDB(t *testing.T) {
	t.Cleanup(func() { resetConfigAndDB() })

	first := resetConfigAndOpenDB()
	require.Same(t, first, get.Config())
	require.Same(t, first.Db(), entity.Db())
	core, err := InitCoreConfig(nil, true)
	require.NoError(t, err)
	require.Same(t, first, core)

	second := resetConfigAndOpenDB()
	require.Same(t, second, get.Config())
	require.Same(t, second.Db(), entity.Db())
	core, err = InitCoreConfig(nil, true)
	require.NoError(t, err)
	require.Same(t, second, core)

	restored := resetConfigAndDB()
	require.Same(t, restored, get.Config())
	require.Same(t, restored.Db(), entity.Db())
	core, err = InitCoreConfig(nil, true)
	require.NoError(t, err)
	require.Same(t, restored, core)
	restoredAgain := resetConfigAndDB()
	require.Same(t, restoredAgain, get.Config())
	require.Same(t, restoredAgain.Db(), entity.Db())
	core, err = InitCoreConfig(nil, true)
	require.NoError(t, err)
	require.Same(t, restoredAgain, core)

	config.NewMinimalTestConfig(t.TempDir())

	for name, c := range map[string]*config.Config{"first": first, "second": second, "restored": restored, "restored again": restoredAgain} {
		if os.Getenv("PHOTOPRISM_TEST_DSN") == "" {
			require.NotEqual(t, initialTestDbDSN, c.DatabaseDSN(), name)
		}

		label := entity.NewLabel("Reset Config Check "+name, 0)
		require.NoError(t, c.Db().Create(label).Error, name)
		t.Cleanup(func() { _ = c.Db().Unscoped().Delete(label).Error })
	}

	label := entity.NewLabel("Reset Config Current", 0)
	require.NoError(t, label.Create())
	t.Cleanup(func() { _ = entity.UnscopedDb().Delete(label).Error })
}

// requireTestDb reopens the shared database for direct registry or entity access.
// A preceding command may have closed it through conf.Shutdown().
func requireTestDb(t *testing.T) *config.Config {
	t.Helper()

	c := reopenConnection()

	if c == nil {
		t.Fatal("test config is not available")
	}

	return c
}

// reopenConnection returns the current config and opens its database if needed.
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

// TestExitCode covers the status main exits with for an error that app.Run returns.
func TestExitCode(t *testing.T) {
	// runApp returns the error urfave/cli reports for a command with a required flag.
	runApp := func(args ...string) error {
		app := cli.NewApp()
		app.Writer, app.ErrWriter = io.Discard, io.Discard
		app.Commands = []*cli.Command{{
			Name:   "add",
			Flags:  []cli.Flag{&cli.StringFlag{Name: "make", Required: true}},
			Action: func(ctx *cli.Context) error { return errors.New("database unreachable") },
		}}
		return app.Run(append([]string{"photoprism"}, args...))
	}

	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, 0, ExitCode(nil))
	})
	t.Run("PlainError", func(t *testing.T) {
		assert.Equal(t, 1, ExitCode(runApp("add", "--make", "Canon")))
	})
	t.Run("MissingRequiredFlag", func(t *testing.T) {
		err := runApp("add")
		assert.ErrorContains(t, err, "Required flag")
		assert.Equal(t, 2, ExitCode(err))
	})
	t.Run("ExitCoder", func(t *testing.T) {
		assert.Equal(t, 3, ExitCode(cli.Exit("not found", 3)))
	})
	t.Run("WrappedExitCoder", func(t *testing.T) {
		assert.Equal(t, 3, ExitCode(fmt.Errorf("users: %w", cli.Exit("not found", 3))))
	})
	t.Run("OutOfRangeExitCoder", func(t *testing.T) {
		assert.Equal(t, 1, ExitCode(fmt.Errorf("users: %w", cli.Exit("failed", 300))))
	})
	t.Run("ExternalToolStatus", func(t *testing.T) {
		execErr := exec.Command("sh", "-c", "exit 3").Run()
		require.Error(t, execErr)
		assert.Equal(t, 1, ExitCode(fmt.Errorf("convert: %w", execErr)))
	})
	t.Run("Canceled", func(t *testing.T) {
		assert.Equal(t, 0, ExitCode(fmt.Errorf("index: %w", status.ErrCanceled)))
	})
	t.Run("InterruptedPrompt", func(t *testing.T) {
		assert.Equal(t, 0, ExitCode(promptui.ErrInterrupt))
	})
}

// TestShowUsageError covers commands that print their help and exit 2 when the argument is missing.
func TestShowUsageError(t *testing.T) {
	for _, c := range []struct {
		name string
		cmd  *cli.Command
	}{
		{"users show", UsersShowCommand},
		{"users rm", UsersRemoveCommand},
		{"users mod", UsersModCommand},
		{"clients show", ClientsShowCommand},
		{"clients rm", ClientsRemoveCommand},
		{"clients mod", ClientsModCommand},
		{"auth show", AuthShowCommand},
		{"auth rm", AuthRemoveCommand},
		{"passwd", PasswdCommand},
		{"connect", ConnectCommand},
	} {
		t.Run(c.name, func(t *testing.T) {
			output, err := RunWithTestContext(c.cmd, []string{c.cmd.Name})
			assertExitCode(t, err, 2)
			assert.Contains(t, output, "USAGE:")
		})
	}
}

// TestCallWithDependencies covers the exit code reported when the configuration cannot be loaded.
func TestCallWithDependencies(t *testing.T) {
	initConfig := InitConfig
	t.Cleanup(func() { InitConfig = initConfig })

	action := func(conf *config.Config) error {
		t.Fatal("the action must not run without a configuration")
		return nil
	}

	t.Run("InitErrorExitsOne", func(t *testing.T) {
		InitConfig = func(ctx *cli.Context) (*config.Config, error) {
			return nil, errors.New("config not readable")
		}

		err := CallWithDependencies(NewTestContext(nil), action)

		assertExitCode(t, err, 1)
		assert.Contains(t, err.Error(), "config not readable")
	})
	t.Run("InitExitCodeIsKept", func(t *testing.T) {
		InitConfig = func(ctx *cli.Context) (*config.Config, error) {
			return nil, cli.Exit("invalid flag", 2)
		}

		assertExitCode(t, CallWithDependencies(NewTestContext(nil), action), 2)
	})
}
