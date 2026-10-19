package commands

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
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
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
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

// runTestMain initializes shared command fixtures and executes the package tests.
func runTestMain(m *testing.M) int {
	// Runs as a stand-in process for the stop tests, without the shared fixtures.
	if mode := os.Getenv(stopHelperEnv); mode != "" {
		return runStopHelper(mode)
	}

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

// commandTestProcess isolates commands whose shutdown ends the native media library lifetime.
func commandTestProcess(t *testing.T) bool {
	t.Helper()
	if os.Getenv("PHOTOPRISM_TEST_COMMAND_PROCESS") == t.Name() {
		return false
	}
	binary, err := os.Executable()
	require.NoError(t, err)
	fixture, err := filepath.Abs("../../pkg/fs/testdata/directory/example.jpg")
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "-test.run=^"+regexp.QuoteMeta(t.Name())+"$", "-test.count=1", "-test.parallel=1", "-test.cpu=1", "-test.v") //nolint:gosec // Runs this test binary with a fixed test selector.
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "PHOTOPRISM_TEST_COMMAND_PROCESS="+t.Name(),
		"PHOTOPRISM_TEST_IMAGE_FILE="+fixture, "PHOTOPRISM_ASSETS_PATH="+get.Config().AssetsPath())
	if get.Config().DatabaseDriver() == dsn.DriverSQLite3 {
		cmd.Env = append(cmd.Env, "PHOTOPRISM_TEST_DRIVER=sqlite", "PHOTOPRISM_TEST_DSN=")
	} else if get.Config().DatabaseDriver() == dsn.DriverPostgres {
		server, parseErr := url.Parse(get.Config().DatabaseDSN())

		// The parse error is not shown, as it would quote the password.
		if parseErr != nil {
			t.Fatal("invalid postgres test dsn")
		}
		prefix := "testdb_cli_" + rnd.Base36(12)
		child := *server
		child.Path, child.RawPath = "/"+prefix, ""
		cmd.Env = append(cmd.Env, "PHOTOPRISM_TEST_DSN="+child.String())
		t.Cleanup(func() { dropPostgresTestDbs(t, server.String(), prefix+"_") })
	} else {
		child, parseErr := mysql.ParseDSN(get.Config().DatabaseDSN())
		require.NoError(t, parseErr)
		server := child.Clone()
		server.DBName = ""
		prefix := "testdb_cli_" + rnd.Base36(12)
		child.DBName = prefix
		cmd.Env = append(cmd.Env, "PHOTOPRISM_TEST_DSN="+child.FormatDSN())
		t.Cleanup(func() {
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cleanupCancel()
			db, openErr := sql.Open(dsn.DriverMySQL, server.FormatDSN())
			require.NoError(t, openErr)
			defer db.Close()
			rows, queryErr := db.QueryContext(cleanupCtx, "SELECT schema_name FROM information_schema.schemata WHERE LEFT(schema_name, ?) = ?", len(prefix)+1, prefix+"_")
			require.NoError(t, queryErr)
			defer rows.Close()
			var schemas []string
			for rows.Next() {
				var schema string
				require.NoError(t, rows.Scan(&schema))
				require.Regexp(t, "^"+regexp.QuoteMeta(prefix)+"_[a-z0-9_]+$", schema)
				schemas = append(schemas, schema)
			}
			require.NoError(t, rows.Err())
			require.NoError(t, rows.Close())
			for _, schema := range schemas {
				_, dropErr := db.ExecContext(cleanupCtx, fmt.Sprintf("DROP DATABASE `%s`", schema)) //nolint:gosec // The child owns this validated test schema.
				require.NoError(t, dropErr)
			}
		})
	}
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", output)
	require.Contains(t, string(output), "--- PASS: "+t.Name(), "child must execute the selected test")
	return true
}

// dropPostgresTestDbs drops the PostgreSQL databases whose names start with prefix.
func dropPostgresTestDbs(t *testing.T, serverDsn, prefix string) {
	t.Helper()

	// Dropping a database waits for databases that other test packages create at the same time.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	db, err := sql.Open("pgx", serverDsn)
	require.NoError(t, err)
	defer db.Close()

	rows, err := db.QueryContext(ctx, "SELECT datname FROM pg_database WHERE LEFT(datname, $1) = $2", len(prefix), prefix)
	require.NoError(t, err)
	defer rows.Close()

	var names []string

	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		require.Regexp(t, "^"+regexp.QuoteMeta(prefix)+"[a-z0-9_]+$", name)
		names = append(names, name)
	}

	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())

	for _, name := range names {
		_, dropErr := db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s WITH (FORCE)", name)) //nolint:gosec // The child owns this validated test database.
		require.NoError(t, dropErr)
	}
}

// resetConfigCleanup captures the active config and registers cleanup for its replacement.
func resetConfigCleanup(t *testing.T) func(*config.Config) {
	t.Helper()
	previous, previousInit := get.Config(), InitConfig
	return func(current *config.Config) {
		t.Cleanup(func() {
			assert.NoError(t, current.CloseDb())
			get.SetConfig(previous)
			previous.RegisterDb()
			// MariaDB and PostgreSQL configs share the package database, including destructive fixture resets.
			if previous.DatabaseDriver() != dsn.DriverSQLite3 {
				previous.InitTestDb()
			}
			InitConfig = previousInit
		})
	}
}

// resetConfigAndDB installs a fixture-backed config for the test.
func resetConfigAndDB(t *testing.T) *config.Config {
	t.Helper()
	restore := resetConfigCleanup(t)
	c := config.NewMinimalTestConfigWithDb(nextResetDbName("commands"), savedPath)
	restore(c)
	get.SetConfig(c)
	entity.SetDbProvider(c)

	InitConfig = func(ctx *cli.Context) (*config.Config, error) {
		return c, c.Init()
	}

	return c
}

// resetConfigAndOpenDB installs a cached test config and opens its database.
func resetConfigAndOpenDB(t *testing.T) *config.Config {
	t.Helper()
	restore := resetConfigCleanup(t)
	c := config.NewIsolatedTestConfig(nextResetDbName("commandsreset"), savedPath, false)
	restore(c)
	config.RestoreDBFromCache(c) // With SQLite, NewIsolatedTestConfig removes the database file first.
	require.NoError(t, c.Init())
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

// TestResetConfigAndOpenDB checks replacement databases and restores their captured test state.
func TestResetConfigAndOpenDB(t *testing.T) {
	previous := requireTestDb(t)
	previousInit := InitConfig
	sentinel := errors.New("captured test initializer")
	InitConfig = func(*cli.Context) (*config.Config, error) { return previous, sentinel }
	t.Cleanup(func() { InitConfig = previousInit })
	var resets []*config.Config
	t.Run("ReplacementConfigs", func(t *testing.T) {
		first := resetConfigAndOpenDB(t)
		require.Same(t, first, get.Config())
		require.Same(t, first.Db(), entity.Db())
		core, err := InitCoreConfig(nil, true)
		require.NoError(t, err)
		require.Same(t, first, core)

		second := resetConfigAndOpenDB(t)
		require.Same(t, second, get.Config())
		require.Same(t, second.Db(), entity.Db())
		core, err = InitCoreConfig(nil, true)
		require.NoError(t, err)
		require.Same(t, second, core)

		restored := resetConfigAndDB(t)
		require.Same(t, restored, get.Config())
		require.Same(t, restored.Db(), entity.Db())
		core, err = InitCoreConfig(nil, true)
		require.NoError(t, err)
		require.Same(t, restored, core)
		restoredAgain := resetConfigAndDB(t)
		require.Same(t, restoredAgain, get.Config())
		require.Same(t, restoredAgain.Db(), entity.Db())
		core, err = InitCoreConfig(nil, true)
		require.NoError(t, err)
		require.Same(t, restoredAgain, core)

		resets = []*config.Config{first, second, restored, restoredAgain}
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
	})
	require.Same(t, previous, get.Config())
	require.NotNil(t, entity.Db())
	require.Same(t, previous.Db(), entity.Db())
	for _, c := range resets {
		assert.False(t, c.IsDbOpen(), "reset configs must be closed")
	}
	restored, err := InitConfig(nil)
	require.ErrorIs(t, err, sentinel)
	require.Same(t, previous, restored)

	fixture := entity.LabelFixtures.Get("landscape")
	var label entity.Label
	require.NoError(t, previous.Db().Where("label_uid = ?", fixture.LabelUID).First(&label).Error)
	t.Run("FixtureRestoration", func(t *testing.T) {
		c := resetConfigAndOpenDB(t)
		result := c.Db().Unscoped().Where("label_uid = ?", fixture.LabelUID).Delete(&entity.Label{})
		require.NoError(t, result.Error)
		require.EqualValues(t, 1, result.RowsAffected)
		var count int64
		require.NoError(t, c.Db().Unscoped().Model(&entity.Label{}).Where("label_uid = ?", fixture.LabelUID).Count(&count).Error)
		require.Zero(t, count)
	})
	label = entity.Label{}
	require.NoError(t, previous.Db().Where("label_uid = ?", fixture.LabelUID).First(&label).Error)
	assert.Equal(t, fixture.LabelName, label.LabelName)
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

// formattedExitErr is a cli.ExitCoder that also implements cli.ErrorFormatter.
type formattedExitErr struct{ msg string }

// Error returns the message.
func (e formattedExitErr) Error() string { return e.msg }

// ExitCode returns 7.
func (e formattedExitErr) ExitCode() int { return 7 }

// Format writes the message with details.
func (e formattedExitErr) Format(s fmt.State, verb rune) {
	_, _ = fmt.Fprint(s, e.msg+": details a\u200Bb")
}

// multiExitErr is a cli.MultiError holding the given errors.
type multiExitErr []error

// Error returns a summary.
func (e multiExitErr) Error() string { return "multiple errors" }

// Errors returns the errors.
func (e multiExitErr) Errors() []error { return e }

// TestHandleExitErr covers how the message of a cli.Exit error is printed and which code it exits with.
func TestHandleExitErr(t *testing.T) {
	// redirectExit redirects the urfave/cli error writer and exit function for one test.
	redirectExit := func(t *testing.T) (*bytes.Buffer, *int) {
		var buf bytes.Buffer
		code := -1
		w, exit := cli.ErrWriter, cli.OsExiter
		cli.ErrWriter = &buf
		cli.OsExiter = func(c int) { code = c }
		t.Cleanup(func() { cli.ErrWriter, cli.OsExiter = w, exit })
		return &buf, &code
	}

	t.Run("CleansMessage", func(t *testing.T) {
		buf, code := redirectExit(t)
		HandleExitErr(nil, cli.Exit("portal: status 503\r\nname: a\x07b\u200Bc", 3))
		assert.Equal(t, "portal: status 503\nname: a[U+0007]b[U+200B]c\n", buf.String())
		assert.Equal(t, 3, *code)
	})
	t.Run("KeepsText", func(t *testing.T) {
		buf, code := redirectExit(t)
		HandleExitErr(nil, cli.Exit("رحلة 2026: not found", 4))
		assert.Equal(t, "رحلة 2026: not found\n", buf.String())
		assert.Equal(t, 4, *code)
	})
	t.Run("EmptyMessage", func(t *testing.T) {
		buf, code := redirectExit(t)
		HandleExitErr(nil, cli.Exit("", 2))
		assert.Empty(t, buf.String())
		assert.Equal(t, 2, *code)
	})
	t.Run("Formatter", func(t *testing.T) {
		buf, code := redirectExit(t)
		HandleExitErr(nil, formattedExitErr{msg: "failed"})
		assert.Equal(t, "failed: details a[U+200B]b\n", buf.String())
		assert.Equal(t, 7, *code)
	})
	t.Run("MultiError", func(t *testing.T) {
		buf, code := redirectExit(t)
		HandleExitErr(nil, multiExitErr{errors.New("first\x07"), nil, multiExitErr{cli.Exit("second\u200B", 5)}})
		assert.Equal(t, "first[U+0007]\nsecond[U+200B]\n", buf.String())
		assert.Equal(t, 5, *code)
	})
	t.Run("PlainError", func(t *testing.T) {
		buf, code := redirectExit(t)
		HandleExitErr(nil, errors.New("failed"))
		HandleExitErr(nil, nil)
		assert.Empty(t, buf.String())
		assert.Equal(t, -1, *code)
	})
}

// TestPrintMultiError covers the output and exit code for the errors of a cli.MultiError.
func TestPrintMultiError(t *testing.T) {
	var buf bytes.Buffer
	w := cli.ErrWriter
	cli.ErrWriter = &buf
	t.Cleanup(func() { cli.ErrWriter = w })

	t.Run("PlainErrors", func(t *testing.T) {
		buf.Reset()
		assert.Equal(t, 1, printMultiError(multiExitErr{errors.New("a\tb"), nil}))
		assert.Equal(t, "a b\n", buf.String())
	})
	t.Run("LastExitCode", func(t *testing.T) {
		buf.Reset()
		assert.Equal(t, 6, printMultiError(multiExitErr{cli.Exit("x", 4), multiExitErr{cli.Exit("y", 6)}, errors.New("z")}))
		assert.Equal(t, "x\ny\nz\n", buf.String())
	})
	t.Run("Empty", func(t *testing.T) {
		buf.Reset()
		assert.Equal(t, 1, printMultiError(multiExitErr{}))
		assert.Empty(t, buf.String())
	})
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
func TestCallWithCoreDependencies(t *testing.T) {
	initCoreConfig := InitCoreConfig
	t.Cleanup(func() { InitCoreConfig = initCoreConfig })

	t.Run("Success", func(t *testing.T) {
		// The configured log level applies, as Config.Propagate is not called without a database.
		prevLevel := event.SystemLog.GetLevel()
		t.Cleanup(func() { config.SetAppLogLevel(prevLevel) })
		conf := config.NewMinimalTestConfig(t.TempDir())
		conf.Options().LogLevel = "error"
		config.SetAppLogLevel(logrus.PanicLevel)
		InitCoreConfig = func(ctx *cli.Context, quiet bool) (*config.Config, error) {
			assert.False(t, quiet)
			return conf, nil
		}

		var got *config.Config
		require.NoError(t, CallWithCoreDependencies(NewTestContext(nil), func(c *config.Config) error {
			got = c
			return nil
		}))
		assert.Same(t, conf, got)
		assert.Equal(t, conf.LogLevel(), event.SystemLog.GetLevel())
	})
	t.Run("InitErrorExitsOne", func(t *testing.T) {
		InitCoreConfig = func(ctx *cli.Context, quiet bool) (*config.Config, error) {
			return nil, errors.New("config not readable")
		}

		err := CallWithCoreDependencies(NewTestContext(nil), func(*config.Config) error {
			t.Fatal("the action must not run without a configuration")
			return nil
		})

		assertExitCode(t, err, 1)
		assert.Contains(t, err.Error(), "config not readable")
	})
	t.Run("InitExitCodeIsKept", func(t *testing.T) {
		InitCoreConfig = func(ctx *cli.Context, quiet bool) (*config.Config, error) {
			return nil, cli.Exit("invalid flag", 2)
		}

		assertExitCode(t, CallWithCoreDependencies(NewTestContext(nil), func(*config.Config) error { return nil }), 2)
	})
}

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
