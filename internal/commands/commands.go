/*
Package commands provides the CLI commands of PhotoPrism.

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark/>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// NONINTERACTIVE is the CLI environment flag to disable prompts.
const NONINTERACTIVE = "noninteractive"

var log = event.Log

// RunNonInteractively checks if command should run non-interactively.
func RunNonInteractively(confirmed bool) bool {
	return confirmed || strings.ToLower(os.Getenv(config.EnvVar("cli"))) == NONINTERACTIVE
}

// confirmStdin is where confirmation prompts read answers from, or nil for the terminal.
var confirmStdin io.ReadCloser

// SetConfirmInput makes confirmation prompts read their answers from r, so that tests in other packages can
// answer them, and returns a function that restores the previous input.
func SetConfirmInput(r io.ReadCloser) (restore func()) {
	prev := confirmStdin
	confirmStdin = r

	return func() { confirmStdin = prev }
}

// confirmTerminal reports whether confirmation prompts read their answers from a terminal.
var confirmTerminal = func() bool {
	return confirmStdin == nil && isTerminal(os.Stdin)
}

// isTerminal reports whether f is a terminal.
func isTerminal(f *os.File) bool {
	return f != nil && (isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd()))
}

// confirmOutput returns where a confirmation prompt is rendered: on stderr, so that stdout only carries
// what the command outputs, unless only stdout is a terminal and the prompt would otherwise not be seen.
func confirmOutput(stdoutTerminal, stderrTerminal bool) *os.File {
	if stdoutTerminal && !stderrTerminal {
		return os.Stdout
	}

	return os.Stderr
}

// ConfirmAction asks the operator to confirm a destructive action and reports whether it may
// proceed. It returns false with no error when the answer is no, including Ctrl-C and Ctrl-D on
// a terminal, and an error when no answer could be obtained at all, since a caller could not
// tell a missing terminal from a considered refusal.
func ConfirmAction(confirmed bool, label string) (proceed bool, err error) {
	if RunNonInteractively(confirmed) {
		return true, nil
	} else if proceed, err = askConfirmation(label); err == nil {
		return proceed, nil
	}

	// Exit code 2 is the usage error: the command was reached in an environment that cannot
	// answer it, and the caller fixes that by passing --yes.
	return false, cli.Exit(fmt.Errorf("could not ask for confirmation (%w), pass --yes to run non-interactively", err), 2)
}

// ConfirmRestore asks whether to restore a deleted record instead of performing the requested operation.
// PHOTOPRISM_CLI=noninteractive alone does not confirm it, and when no answer can be obtained it returns
// a plain error naming flag, so the command exits 1 as it does when the offer is declined.
func ConfirmRestore(confirmed bool, label, flag string) (restore bool, err error) {
	if confirmed {
		return true, nil
	} else if RunNonInteractively(false) {
		return false, fmt.Errorf("%s requires confirmation, pass %s", confirmLabel(label), flag)
	} else if restore, err = askConfirmation(label); err == nil {
		return restore, nil
	}

	return false, fmt.Errorf("could not ask for confirmation (%w), pass %s", err, flag)
}

// askConfirmation shows a yes/no prompt, and returns an error only when no answer could be obtained.
func askConfirmation(label string) (bool, error) {
	prompt := promptui.Prompt{Label: confirmLabel(label), IsConfirm: true, Stdin: confirmStdin,
		Stdout: confirmOutput(isTerminal(os.Stdout), isTerminal(os.Stderr))}

	if _, err := prompt.Run(); err == nil {
		return true, nil
	} else if errors.Is(err, promptui.ErrAbort) || errors.Is(err, promptui.ErrInterrupt) {
		return false, nil
	} else if !errors.Is(err, promptui.ErrEOF) {
		return false, err
	} else if confirmTerminal() {
		return false, nil
	}

	return false, errors.New("no terminal")
}

// confirmLabel removes a trailing question mark, since the prompt appends its own.
func confirmLabel(label string) string {
	return strings.TrimSpace(strings.TrimRight(strings.TrimSpace(label), "?"))
}

// PhotoPrism contains the photoprism CLI (sub-)commands.
var PhotoPrism = []*cli.Command{
	StartCommand,
	StopCommand,
	StatusCommand,
	IndexCommand,
	FindCommand,
	ImportCommand,
	CopyCommand,
	DownloadCommand,
	VisionCommands,
	FacesCommands,
	CamerasCommand,
	LensesCommand,
	PlacesCommands,
	PurgeCommand,
	CleanUpCommand,
	OptimizeCommand,
	MomentsCommand,
	ConvertCommand,
	ThumbsCommand,
	VideosCommands,
	MigrateCommand,
	MigrationsCommands,
	DatabaseCommands,
	BackupCommand,
	RestoreCommand,
	ResetCommand,
	PasswdCommand,
	UsersCommands,
	ClientsCommands,
	ClusterCommands,
	AuthCommands,
	MCPCommands,
	ShowCommands,
	VersionCommand,
	EditionCommand,
	ShowConfigCommand,
	ConnectCommand,
}

// CountFlag represents a CLI flag to limit the number of report rows.
var CountFlag = &cli.UintFlag{
	Name:    "count",
	Aliases: []string{"n"},
	Usage:   "maximum `NUMBER` of results",
	Value:   100,
}

// LogErr logs an error if the argument is not nil.
func LogErr(err error) {
	if err != nil {
		log.Error(err)
	}
}

// childAlreadyRunning reports whether a running daemon holds the pid file at filePath, and returns the pid in the file.
// A missing or empty pid file means that no daemon is running, as does invalid content no other process holds
// locked, which the next daemon overwrites. A pid file that cannot be read or is too large is an error.
func childAlreadyRunning(filePath string) (pid int, running bool, err error) {
	pid, info, err := readPidFile(filePath)

	switch {
	case errors.Is(err, errPidFileEmpty) || daemonNotRunning(err):
		return 0, false, nil
	case errors.Is(err, errPidFileInvalid):
		if held, _ := pidFileLocked(filePath, info); held {
			return 0, false, fmt.Errorf("%w, and another process holds it locked", err)
		}

		log.Warnf("%s, and will be overwritten", clean.Error(err))

		return 0, false, nil
	case err != nil:
		return 0, false, err
	}

	process, err := os.FindProcess(pid)

	if err != nil {
		return pid, false, err
	}

	running, _, err = daemonHoldsPidFile(process, filePath, info)

	return pid, running, err
}

// ExitCode returns the process exit status for an error that app.Run returned without exiting.
// A canceled operation or interrupted prompt exits 0; the status of an external tool is not passed on.
// urfave/cli reports a missing required flag with an unexported type, so its name is compared.
func ExitCode(err error) int {
	var exit cli.ExitCoder
	var execErr *exec.ExitError

	switch {
	case err == nil:
		return 0
	case errors.As(err, &execErr):
		return 1
	case errors.As(err, &exit):
		if code := exit.ExitCode(); code >= 0 && code <= 125 {
			return code
		}
		return 1
	case errors.Is(err, status.ErrCanceled), errors.Is(err, promptui.ErrInterrupt):
		return 0
	case fmt.Sprintf("%T", err) == "*cli.errRequiredFlags":
		return 2
	default:
		return 1
	}
}

// HandleExitErr prints the message of a cli.Exit error like cli.HandleExitCoder, with the message passed
// through clean.DisplayLines, and exits with its code. A cli.MultiError is handled the same way.
func HandleExitErr(_ *cli.Context, err error) {
	switch e := err.(type) { //nolint:errorlint // matches cli.HandleExitCoder
	case cli.ExitCoder:
		if err.Error() != "" {
			msg := err.Error()

			if _, ok := e.(cli.ErrorFormatter); ok {
				msg = fmt.Sprintf("%+v", err)
			}

			_, _ = fmt.Fprintln(cli.ErrWriter, clean.DisplayLines(msg))
		}

		cli.OsExiter(e.ExitCode())
	case cli.MultiError:
		cli.OsExiter(printMultiError(e))
	}
}

// printMultiError prints the errors of a cli.MultiError like cli.HandleExitCoder and returns the exit code
// of the last cli.ExitCoder, or 1.
func printMultiError(multiErr cli.MultiError) int {
	code := 1

	for _, err := range multiErr.Errors() {
		if nested, ok := err.(cli.MultiError); ok { //nolint:errorlint // matches cli.HandleExitCoder
			code = printMultiError(nested)
		} else if err != nil {
			_, _ = fmt.Fprintln(cli.ErrWriter, clean.DisplayLines(err.Error()))

			if exitErr, ok := err.(cli.ExitCoder); ok { //nolint:errorlint // matches cli.HandleExitCoder
				code = exitErr.ExitCode()
			}
		}
	}

	return code
}

// ShowUsageError prints the command help and returns a usage error, e.g. for a missing argument.
func ShowUsageError(ctx *cli.Context) error {
	if err := cli.ShowSubcommandHelp(ctx); err != nil {
		return err
	}

	return cli.Exit("", 2)
}

// CallWithCoreDependencies calls a command action with a config that is initialized without the node
// bootstrap and a database connection, for commands that only read options and call other services.
func CallWithCoreDependencies(ctx *cli.Context, action func(conf *config.Config) error) error {
	conf, err := InitCoreConfig(ctx, false)

	if err != nil {
		var exit cli.ExitCoder

		if errors.As(err, &exit) {
			return err
		}

		return cli.Exit(err, 1)
	}

	// Apply the configured log level, as Config.Propagate does after a database connection.
	config.SetAppLogLevel(conf.LogLevel())

	return action(conf)
}

// CallWithDependencies calls a command action with initialized dependencies.
func CallWithDependencies(ctx *cli.Context, action func(conf *config.Config) error) (err error) {
	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		var exit cli.ExitCoder

		if errors.As(err, &exit) {
			return err
		}

		return cli.Exit(err, 1)
	}

	defer conf.Shutdown()

	// Run command.
	err = action(conf)

	return err
}
