package commands

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// FacesCommand registers the face recognition subcommands.
var FacesCommand = cli.Command{
	Name:  "faces",
	Usage: "Face recognition subcommands",
	Subcommands: []cli.Command{
		{
			Name:   "stats",
			Usage:  "Shows stats on face samples",
			Action: facesStatsAction,
		},
		{
			Name:  "audit",
			Usage: "Scans the index for issues",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "fix, f",
					Usage: "fix discovered issues",
				},
			},
			Action: facesAuditAction,
		},
		{
			Name:  "reset",
			Usage: "Removes people and faces after confirmation",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "remove all people and faces",
				},
			},
			Action: facesResetAction,
		},
		{
			Name:      "index",
			Usage:     "Searches originals for faces",
			ArgsUsage: "[originals folder]",
			Action:    facesIndexAction,
		},
		{
			Name:  "update",
			Usage: "Performs face clustering and matching",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "update all faces",
				},
			},
			Action: facesUpdateAction,
		},
		{
			Name:   "optimize",
			Usage:  "Optimizes face clusters",
			Action: facesOptimizeAction,
		},
	},
}

// facesStatsAction shows stats on face embeddings.
func facesStatsAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()

	w := service.Faces()

	if err := w.Stats(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	conf.Shutdown()

	return nil
}

// facesAuditAction shows stats on face embeddings.
func facesAuditAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()

	w := service.Faces()

	if err := w.Audit(ctx.Bool("fix")); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	conf.Shutdown()

	return nil
}

// facesResetAction resets face clusters and matches.
func facesResetAction(ctx *cli.Context) error {
	if ctx.Bool("force") {
		return facesResetAllAction(ctx)
	}

	actionPrompt := promptui.Prompt{
		Label:     "Remove automatically recognized faces, matches, and dangling subjects?",
		IsConfirm: true,
	}

	if _, err := actionPrompt.Run(); err != nil {
		return nil
	}

	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()

	w := service.Faces()

	if err := w.Reset(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	conf.Shutdown()

	return nil
}

// facesResetAllAction removes all people, faces, and face markers.
func facesResetAllAction(ctx *cli.Context) error {
	actionPrompt := promptui.Prompt{
		Label:     "Permanently remove all people and faces?",
		IsConfirm: true,
	}

	if _, err := actionPrompt.Run(); err != nil {
		return nil
	}

	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()

	if err := query.RemovePeopleAndFaces(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	conf.Shutdown()

	return nil
}

// facesIndexAction searches originals for faces.
func facesIndexAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()

	// Use first argument to limit scope if set.
	subPath := strings.TrimSpace(ctx.Args().First())

	if subPath == "" {
		log.Infof("finding faces in %s", clean.Log(conf.OriginalsPath()))
	} else {
		log.Infof("finding faces in %s", clean.Log(filepath.Join(conf.OriginalsPath(), subPath)))
	}

	if conf.ReadOnly() {
		log.Infof("config: read-only mode enabled")
	}

	var indexed fs.Done

	settings := conf.Settings()

	if w := service.Index(); w != nil {
		convert := settings.Index.Convert && conf.SidecarWritable()
		opt := photoprism.NewIndexOptions(subPath, true, convert, true, true, true)

		indexed = w.Start(opt)
	}

	if w := service.Purge(); w != nil {
		opt := photoprism.PurgeOptions{
			Path:   subPath,
			Ignore: indexed,
		}

		if files, photos, err := w.Start(opt); err != nil {
			log.Error(err)
		} else if len(files) > 0 || len(photos) > 0 {
			log.Infof("purge: removed %s and %s", english.Plural(len(files), "file", "files"), english.Plural(len(photos), "photo", "photos"))
		}
	}

	elapsed := time.Since(start)

	log.Infof("indexed %s in %s", english.Plural(len(indexed), "file", "files"), elapsed)

	conf.Shutdown()

	return nil
}

// facesUpdateAction performs face clustering and matching.
func facesUpdateAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()

	opt := photoprism.FacesOptions{
		Force: ctx.Bool("force"),
	}

	w := service.Faces()

	if err := w.Start(opt); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	conf.Shutdown()

	return nil
}

// facesOptimizeAction optimizes existing face clusters.
func facesOptimizeAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()

	w := service.Faces()

	if res, err := w.Optimize(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("merged %s in %s", english.Plural(res.Merged, "face cluster", "face clusters"), elapsed)
	}

	conf.Shutdown()

	return nil
}
