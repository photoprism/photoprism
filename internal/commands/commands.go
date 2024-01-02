/*
Package commands provides the CLI commands of PhotoPrism.

Copyright (c) 2018 - 2023 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package commands

import (
	"context"
	"os"
	"syscall"

	"github.com/sevlyar/go-daemon"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
)

var log = event.Log

// PhotoPrism contains the photoprism CLI (sub-)commands.
var PhotoPrism = []cli.Command{
	StartCommand,
	StopCommand,
	StatusCommand,
	IndexCommand,
	FindCommand,
	ImportCommand,
	CopyCommand,
	FacesCommands,
	PlacesCommands,
	PurgeCommand,
	CleanUpCommand,
	OptimizeCommand,
	MomentsCommand,
	ConvertCommand,
	ThumbsCommand,
	MigrateCommand,
	MigrationsCommands,
	BackupCommand,
	RestoreCommand,
	ResetCommand,
	PasswdCommand,
	UsersCommands,
	ClientsCommands,
	ShowCommands,
	VersionCommand,
	ShowConfigCommand,
	ConnectCommand,
}

// countFlag represents a CLI flag to limit the number of report rows.
var countFlag = cli.UintFlag{
	Name:  "n",
	Usage: "`LIMIT` number of results",
	Value: 100,
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

// CallWithDependencies calls a command action with initialized dependencies.
func CallWithDependencies(ctx *cli.Context, action func(conf *config.Config) error) (err error) {
	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.RegisterDb()
	defer conf.Shutdown()

	// Run command.
	err = action(conf)

	return err
}
