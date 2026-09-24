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
	"syscall"

	"github.com/manifoldco/promptui"
	"github.com/sevlyar/go-daemon"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// NONINTERACTIVE is the CLI environment flag to disable prompts.
const NONINTERACTIVE = "noninteractive"

var log = event.Log

// RunNonInteractively checks if command should run non-interactively.
func RunNonInteractively(confirmed bool) bool {
	return confirmed || strings.ToLower(os.Getenv(config.EnvVar("cli"))) == NONINTERACTIVE
}

// confirmStdin is where ConfirmAction reads answers from, or nil for the terminal.
var confirmStdin io.ReadCloser

// ConfirmAction asks the operator to confirm a destructive action and reports whether it may
// proceed. It returns false with no error when the answer is no, and an error when no answer
// could be obtained at all - without a terminal there is nothing to report as a decision, and
// treating that as a refusal would tell a caller the action had been considered and declined.
func ConfirmAction(confirmed bool, label string) (proceed bool, err error) {
	if RunNonInteractively(confirmed) {
		return true, nil
	}

	prompt := promptui.Prompt{Label: confirmLabel(label), IsConfirm: true, Stdin: confirmStdin}

	if _, err = prompt.Run(); err == nil {
		return true, nil
	} else if errors.Is(err, promptui.ErrAbort) || errors.Is(err, promptui.ErrInterrupt) {
		return false, nil
	}

	// Exit code 2 is the usage error: the command was reached in an environment that cannot
	// answer it, and the caller fixes that by passing --yes.
	return false, cli.Exit(fmt.Errorf("could not ask for confirmation (%w), pass --yes to run non-interactively", err), 2)
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

// childAlreadyRunning tests if a .pid file at filePath is a running process.
// it returns the pid value and the running status (true or false).
func childAlreadyRunning(filePath string) (pid int, running bool) {
	if !fs.FileExists(filePath) {
		return pid, false
	}

	pid, err := daemon.ReadPidFile(filePath)

	// Failed?
	if err != nil {
		return pid, false
	}

	process, err := os.FindProcess(pid)

	// Failed?
	if err != nil {
		return pid, false
	}

	return pid, process.Signal(syscall.Signal(0)) == nil
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

// ShowUsageError prints the command help and returns a usage error, e.g. for a missing argument.
func ShowUsageError(ctx *cli.Context) error {
	if err := cli.ShowSubcommandHelp(ctx); err != nil {
		return err
	}

	return cli.Exit("", 2)
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
