package commands

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/clean"
)

// CopyCommand configures the command name, flags, and action.
var CopyCommand = cli.Command{
	Name:      "cp",
	Aliases:   []string{"copy"},
	Usage:     "Copies media files to originals",
	ArgsUsage: "[source]",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "dest, d",
			Usage: "relative originals `PATH` to which the files should be imported",
		},
	},
	Action: copyAction,
}

// copyAction copies photos to originals path. Default import path is used if no path argument provided
func copyAction(ctx *cli.Context) error {
	start := time.Now()

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	// very if copy directory exist and is writable
	if conf.ReadOnly() {
		return config.ErrReadOnly
	}

	conf.InitDb()
	defer conf.Shutdown()

	// get cli first argument
	sourcePath := strings.TrimSpace(ctx.Args().First())

	if sourcePath == "" {
		sourcePath = conf.ImportPath()
	} else {
		abs, err := filepath.Abs(sourcePath)

		if err != nil {
			return err
		}

		sourcePath = abs
	}

	if sourcePath == conf.OriginalsPath() {
		return errors.New("import path is identical with originals")
	}

	var destFolder string
	if ctx.IsSet("dest") {
		destFolder = clean.UserPath(ctx.String("dest"))
	} else {
		destFolder = conf.ImportDest()
	}

	log.Infof("copying media files from %s to %s", sourcePath, filepath.Join(conf.OriginalsPath(), destFolder))

	w := get.Import()
	opt := photoprism.ImportOptionsCopy(sourcePath, destFolder)

	w.Start(opt)

	elapsed := time.Since(start)

	log.Infof("completed in %s", elapsed)

	return nil
}
