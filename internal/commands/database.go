package commands

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/pkg/clean"
)

// DatabaseCommands configures the index database backup, restore, and transfer subcommands.
var DatabaseCommands = &cli.Command{
	Name:    "database",
	Aliases: []string{"db"},
	Usage:   "Index database backup, restore, and transfer subcommands",
	Subcommands: []*cli.Command{
		DatabaseBackupCommand,
		DatabaseRestoreCommand,
		DatabaseTransferCommand,
	},
}

// databasePathFlags returns the flags that set the database backup path. The hidden alias keeps
// the flag names of the backup and restore commands working.
func databasePathFlags() []cli.Flag {
	return []cli.Flag{
		&cli.PathFlag{
			Name:      "path",
			Usage:     "custom database backup `PATH`",
			TakesFile: true,
		},
		&cli.PathFlag{
			Name:      "database-path",
			Aliases:   []string{"index-path"},
			Usage:     "custom database backup `PATH`",
			TakesFile: true,
			Hidden:    true,
		},
	}
}

// databasePathArg returns the absolute database backup path passed with --path or its hidden alias.
func databasePathArg(ctx *cli.Context) string {
	if p := ctx.String("path"); p != "" {
		return absPathArg(p)
	}

	return absPathArg(ctx.String("database-path"))
}

// databaseBackupArgs returns the backup path and filename of a database backup or restore command, with a
// relative filename resolved in the path. Arguments whose path would be ambiguous, e.g. an empty --path
// or a filename outside it, are refused with a usage error.
func databaseBackupArgs(ctx *cli.Context) (path, fileName string, err error) {
	if err = RejectTrailingFlags(ctx); err != nil {
		return "", "", cli.Exit(err, 2)
	}

	for _, name := range []string{"path", "database-path"} {
		if ctx.IsSet(name) && ctx.String(name) == "" {
			return "", "", cli.Exit(fmt.Errorf("--%s must not be empty", name), 2)
		}
	}

	path = databasePathArg(ctx)
	fileName = ctx.Args().First()

	switch {
	case ctx.Args().Present() && fileName == "":
		return "", "", cli.Exit(errors.New("filename must not be empty"), 2)
	case fileName == "" || fileName == "-" || path == "":
		return path, fileName, nil
	case filepath.IsAbs(fileName):
		return "", "", cli.Exit(fmt.Errorf("%s is an absolute filename, so it cannot be combined with --path", clean.Log(fileName)), 2)
	}

	fileName = filepath.Join(path, fileName)

	if rel, relErr := filepath.Rel(path, fileName); relErr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", "", cli.Exit(fmt.Errorf("%s is not in the --path directory", clean.Log(ctx.Args().First())), 2)
	}

	return path, fileName, nil
}
