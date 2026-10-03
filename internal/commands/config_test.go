package commands

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
)

// initConfigErrorCase is a command whose action initializes the config itself, with the arguments
// that let it reach that call.
type initConfigErrorCase struct {
	name   string
	cmd    *cli.Command
	args   []string
	action string
}

// findSubcommand returns the subcommand with the specified name.
func findSubcommand(t *testing.T, parent *cli.Command, name string) *cli.Command {
	t.Helper()

	for _, c := range parent.Subcommands {
		if c.Name == name {
			return c
		}
	}

	require.Failf(t, "subcommand not found", "%s %s", parent.Name, name)

	return nil
}

// actionName returns the unqualified name of the function a command runs.
func actionName(cmd *cli.Command) string {
	name := runtime.FuncForPC(reflect.ValueOf(cmd.Action).Pointer()).Name()
	return name[strings.LastIndex(name, ".")+1:]
}

// initConfigCallers returns the names of the functions in this package that call InitConfig.
func initConfigCallers(t *testing.T) []string {
	t.Helper()

	files, err := filepath.Glob("*.go")
	require.NoError(t, err)

	var result []string

	fset := token.NewFileSet()

	for _, fileName := range files {
		if strings.HasSuffix(fileName, "_test.go") {
			continue
		}

		file, parseErr := parser.ParseFile(fset, fileName, nil, 0)
		require.NoError(t, parseErr)

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)

			if !ok || fn.Body == nil {
				continue
			}

			calls := false

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				if call, isCall := n.(*ast.CallExpr); isCall {
					if ident, isIdent := call.Fun.(*ast.Ident); isIdent && ident.Name == "InitConfig" {
						calls = true
					}
				}

				return !calls
			})

			if calls {
				result = append(result, fn.Name.Name)
			}
		}
	}

	sort.Strings(result)

	return result
}

// runAction runs a command action and fails the test if it panics or does not return in time.
func runAction(t *testing.T, cmd *cli.Command, ctx *cli.Context) error {
	t.Helper()

	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()

		done <- cmd.Action(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(30 * time.Second):
		t.Fatalf("%s did not return", cmd.Name)
		return nil
	}
}

func TestInitConfigErrors(t *testing.T) {
	SetEnvForTest(t, "PHOTOPRISM_CLI", "noninteractive")

	sentinel := errors.New("init config failed")

	var calls atomic.Int32

	previousInit := InitConfig
	t.Cleanup(func() { InitConfig = previousInit })

	// The returned config is never opened, so a command that uses it despite the error cannot
	// change the test database.
	InitConfig = func(*cli.Context) (*config.Config, error) {
		calls.Add(1)
		return config.NewMinimalTestConfig(t.TempDir()), sentinel
	}

	cases := []initConfigErrorCase{
		{name: "Backup", cmd: BackupCommand, args: []string{"--database"}, action: "backupAction"},
		{name: "CleanUp", cmd: CleanUpCommand, action: "cleanUpAction"},
		{name: "Convert", cmd: ConvertCommand, action: "convertAction"},
		{name: "Copy", cmd: CopyCommand, action: "copyAction"},
		{name: "Download", cmd: DownloadCommand, action: "downloadAction"},
		{name: "FacesAudit", cmd: findSubcommand(t, FacesCommands, "audit"), action: "facesAuditAction"},
		{name: "FacesIndex", cmd: findSubcommand(t, FacesCommands, "index"), action: "facesIndexAction"},
		{name: "FacesOptimize", cmd: findSubcommand(t, FacesCommands, "optimize"), action: "facesOptimizeAction"},
		{name: "FacesReset", cmd: findSubcommand(t, FacesCommands, "reset"), args: []string{"--yes"}, action: "facesResetAction"},
		{name: "FacesResetForce", cmd: findSubcommand(t, FacesCommands, "reset"), args: []string{"--force", "--yes"}, action: "facesResetAllAction"},
		{name: "FacesStats", cmd: findSubcommand(t, FacesCommands, "stats"), action: "facesStatsAction"},
		{name: "FacesUpdate", cmd: findSubcommand(t, FacesCommands, "update"), action: "facesUpdateAction"},
		{name: "Find", cmd: FindCommand, action: "findAction"},
		{name: "Import", cmd: ImportCommand, action: "importAction"},
		{name: "Index", cmd: IndexCommand, action: "indexAction"},
		{name: "MCPServe", cmd: MCPServeCommand, action: "mcpServeAction"},
		{name: "MigrationsRun", cmd: MigrationsRunCommand, action: "migrationsRunAction"},
		{name: "MigrationsStatus", cmd: MigrationsStatusCommand, action: "migrationsStatusAction"},
		{name: "Moments", cmd: MomentsCommand, action: "momentsAction"},
		{name: "Optimize", cmd: OptimizeCommand, action: "optimizeAction"},
		{name: "Passwd", cmd: PasswdCommand, args: []string{"zz-passwd-user"}, action: "passwdAction"},
		{name: "PlacesUpdate", cmd: findSubcommand(t, PlacesCommands, "update"), action: "placesUpdateAction"},
		{name: "Purge", cmd: PurgeCommand, action: "purgeAction"},
		{name: "Reset", cmd: ResetCommand, action: "resetAction"},
		{name: "Restore", cmd: RestoreCommand, args: []string{"--database"}, action: "restoreAction"},
		{name: "Start", cmd: StartCommand, action: "startAction"},
		{name: "Thumbs", cmd: ThumbsCommand, action: "thumbsAction"},
	}

	t.Run("AllCallersCovered", func(t *testing.T) {
		// CallWithDependencies wraps the error itself and is covered by TestCallWithDependencies.
		expected := []string{"CallWithDependencies"}

		for _, c := range cases {
			expected = append(expected, c.action)
		}

		sort.Strings(expected)

		assert.Equal(t, expected, initConfigCallers(t), "add a case for every command that calls InitConfig")
	})

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.action != "facesResetAllAction" {
				require.Equal(t, c.action, actionName(c.cmd))
			}

			calls.Store(0)

			err := runAction(t, c.cmd, newCommandContext(t, c.cmd, c.args...))

			assert.Equal(t, int32(1), calls.Load(), "InitConfig must be called once")
			require.Error(t, err)
			assert.ErrorIs(t, err, sentinel)
			assert.Equal(t, 1, ExitCode(err))
		})
	}
}
