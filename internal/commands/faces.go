package commands

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// FacesCommands configures the command name, flags, and action.
var FacesCommands = &cli.Command{
	Name:  "faces",
	Usage: "Face recognition subcommands",
	Subcommands: []*cli.Command{
		{
			Name:   "stats",
			Usage:  "Shows stats on face samples",
			Action: facesStatsAction,
		},
		{
			Name:  "audit",
			Usage: "Scans the index for issues",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "fix",
					Aliases: []string{"f"},
					Usage:   "fix discovered issues",
				},
				&cli.StringFlag{
					Name:  "subject",
					Usage: "limit audit to the specific subject UID",
				},
			},
			Action: facesAuditAction,
		},
		{
			Name:  "reset",
			Usage: "Removes people and faces after confirmation",
			Flags: []cli.Flag{
				ForceFlag("removes all people and faces"),
				&cli.StringFlag{
					Name:  "detector",
					Usage: "regenerate markers with the detection model `NAME` (" + face.DetectorUsageString() + ")",
				},
				&cli.StringFlag{
					Name:   "engine",
					Usage:  "regenerate markers using detection engine `NAME` *deprecated*, use --detector",
					Hidden: true,
				},
			},
			Action: facesResetAction,
		},
		FacesMigrateCommand,
		{
			Name:      "index",
			Usage:     "Searches originals for faces",
			ArgsUsage: "[subfolder]",
			Action:    facesIndexAction,
		},
		{
			Name:  "update",
			Usage: "Performs face clustering and matching",
			Flags: []cli.Flag{
				ForceFlag("update all faces"),
			},
			Action: facesUpdateAction,
		},
		{
			Name:  "optimize",
			Usage: "Optimizes face clusters",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "retry",
					Usage: "reset merge retry counters before optimizing",
				},
			},
			Action: facesOptimizeAction,
		},
		FacesStatusCommand,
	},
}

// FacesMigrateCommand configures the face embedding migration command.
var FacesMigrateCommand = &cli.Command{
	Name:  "migrate",
	Usage: "Migrates face embeddings to the supported model",
	Description: "This is how the face embedding model is changed: every marker is re-embedded and " +
		"the target is recorded as the configured model. It defaults to " + face.DefaultModelName() +
		", the model this release supports, so an ordinary migration needs no target. Stop the server " +
		"before running it, as the migration replaces every face cluster in one transaction.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "to",
			Usage: "target embedding `MODEL` (default " + face.DefaultModelName() + ")",
		},
		DryRunFlag("reports the face migration scope without changing the index"),
		ForceFlag("finalizes the migration even when markers could not be re-embedded"),
		YesFlag(),
	},
	Action: facesMigrateAction,
}

// facesMigrateAction migrates face embeddings to the configured model.
func facesMigrateAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		w := get.Faces()
		plan, err := w.PlanMigration(ctx.String("to"))
		if err != nil {
			// Plain errors leave the exit status at 0, so a script cannot tell a refused
			// migration from one that ran.
			return cli.Exit(err.Error(), 1)
		}

		log.Infof(
			"faces: migration to %s includes %d valid markers, %d invalid markers, and %d identified people",
			clean.Log(plan.Target), plan.Markers.Valid, plan.Markers.Invalid, plan.Subjects,
		)
		// People keep the faces already assigned to them, so an operator can tell at a
		// glance whether the run is about to touch a well-curated library.
		log.Infof(
			"faces: %d markers are assigned to a person and keep that assignment",
			plan.AssignedMarkers,
		)
		// A face too small or too poorly scored to be clustered cannot define a centroid
		// either, so a library of mostly small faces rebuilds from less than it looks like.
		if plan.LowQualitySamples > 0 {
			log.Infof("faces: %d of those are too small or too low-scoring to seed a cluster",
				plan.LowQualitySamples)
		}
		// Ready tells an operator that a re-run has nothing left to do, and unlinked
		// markers are cleared by every run regardless of how the migration goes.
		log.Infof(
			"faces: %d markers already use %s, %d have no file, and %d were identified manually",
			plan.Markers.Ready, clean.Log(plan.Target), plan.Markers.Unlinked, plan.Markers.Manual,
		)
		// The crop is an axis of the embedding space, so a detector change leaves a library in
		// two of them. This is the only run that repairs that, and it is why a re-run to the
		// same model can still have work to do. Every marker indexed before the detector was
		// recorded counts here, which on a first run is all of them.
		if plan.RecropMarkers > 0 {
			log.Infof("faces: %d of those were cropped by another or an unrecorded detector and are re-embedded, keeping their vector if detection cannot find them again",
				plan.RecropMarkers)
		}
		// Re-embedding reads the file, so a marker whose file the index has already recorded
		// as unreadable is going to fail. Naming them before the prompt is what separates an
		// expected loss from a surprise, since a failed marker keeps no vector at all.
		if plan.Markers.Unreadable > 0 {
			event.SystemWarn([]string{"faces", "migrate", "%d markers cannot be re-embedded because their file is missing or unreadable"},
				plan.Markers.Unreadable)
		}
		// The counts above come from the index, which believes whatever it was told last. An
		// unmounted originals volume leaves them looking clean and then fails every file.
		if plan.OriginalsUnavailable {
			event.SystemWarn([]string{"faces", "migrate", "originals path %s is empty or cannot be read, so no marker can be re-embedded"},
				clean.Log(conf.OriginalsPath()))
		}
		for _, count := range plan.MarkerModels {
			model := count.EmbedModel
			if model == "" {
				model = "legacy"
			}
			log.Infof("faces: embedding model %s has %d markers", clean.Log(model), count.Markers)
		}
		for _, count := range plan.FaceModels {
			model := count.EmbedModel
			if model == "" {
				model = "legacy"
			}
			log.Infof("faces: embedding model %s has %d clusters", clean.Log(model), count.Faces)
		}
		if m := face.FindEmbeddingModel(plan.Target); m != nil {
			log.Infof("faces: %s uses cluster distance %.2f, cluster radius %.2f, and match distance %.2f",
				clean.Log(plan.Target), m.ClusterDist, m.ClusterRadius, m.MatchDist)
		}

		// Finalizing clears the stored vectors of every marker that is not on the target
		// model, so an operator has to see that number before deciding to run this.
		if stale := plan.Markers.Valid - plan.Markers.Ready; stale > 0 {
			event.SystemWarn([]string{"faces", "migrate", "%d markers must be re-embedded and lose their stored vectors if that fails"}, stale)
		}

		if ctx.Bool("dry-run") {
			log.Infof("faces: dry run completed without changes")
			return nil
		}

		// The worker guards in Migrate are process-local, so they cannot see a server that
		// is indexing or matching the same rows. Stopping it is the operator's job, and the
		// prompt is the last point at which saying so still helps.
		event.SystemWarn([]string{"faces", "migrate", "this replaces every face cluster; indexing and vision hold off " +
			"while it runs, but changes made in the app are not covered, so stopping the server is still the safe way"})

		if !RunNonInteractively(ctx.Bool("yes")) {
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Migrate all face embeddings to %s, with the server stopped?", plan.Target),
				IsConfirm: true,
			}
			if _, promptErr := prompt.Run(); promptErr != nil {
				log.Info("faces: migration canceled")
				return nil
			}
		}

		result, migrateErr := w.Migrate(ctx.Context, photoprism.FacesMigrateOptions{
			Target: plan.Target,
			Force:  ctx.Bool("force"),
			Plan:   &plan,
		})
		log.Infof(
			"faces: migrated %d markers, skipped %d, failed %d, %d without a file; preserved %d people, %d assignments and %d hidden clusters, rebuilt %d clusters, %d need attention",
			result.Migrated, result.Skipped, result.Failed, result.Unlinked,
			result.PreservedSubjects, result.PreservedMarkers, result.HiddenClusters,
			result.RebuiltSubjects, result.AttentionSubjects,
		)
		// Reported apart from both, because a retained marker is neither work done nor a loss:
		// detection did not find it again, most often because a person drew it by hand.
		if result.Retained > 0 {
			log.Infof("faces: %d markers kept the vector another detector's crop produced", result.Retained)
		}
		// Excluded assignments keep their person but seed no cluster, so the count is what
		// tells an operator how much of a curated library did not shape its own centroids.
		if result.ExcludedMarkers > 0 || result.LowQualityMarkers > 0 {
			log.Infof("faces: %d assignment(s) were left out of a cluster as outliers, and %d as too low-quality to seed one",
				result.ExcludedMarkers, result.LowQualityMarkers)
		}

		// The setting is written by a run that replaced the clusters, including one that
		// reports failed markers. A write that failed carries its own error, so this line has
		// to follow the file rather than the value this process is holding.
		var settingErr *photoprism.FacesMigrateSettingError

		if !errors.As(migrateErr, &settingErr) && conf.FaceModel() == plan.Target {
			log.Infof("faces: the configured face model is now %s", clean.Log(plan.Target))
		}

		if migrateErr != nil {
			return cli.Exit(migrateErr.Error(), 1)
		}

		return nil
	})
}

// facesStatsAction shows stats on face embeddings.
func facesStatsAction(ctx *cli.Context) error {
	start := time.Now()

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	w := get.Faces()

	if err := w.Stats(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	return nil
}

// facesAuditAction shows stats on face embeddings.
func facesAuditAction(ctx *cli.Context) error {
	start := time.Now()

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	w := get.Faces()

	subject := strings.TrimSpace(ctx.String("subject"))

	if err := w.Audit(ctx.Bool("fix"), subject); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	return nil
}

// facesResetAction resets face clusters and matches.
func facesResetAction(ctx *cli.Context) error {
	if ctx.Bool("force") {
		// The two do not compose: --force removes every person, face and marker, and the names go
		// with them, so a caller who also asked to regenerate would silently get the destructive
		// half alone. Refused rather than reordered, because which of the two they meant is not
		// knowable from the command.
		if ctx.IsSet("detector") || ctx.IsSet("engine") {
			return cli.Exit("faces: --force removes all people and faces, so it cannot be combined with --detector", 1)
		}

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

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	w := get.Faces()

	detector := strings.TrimSpace(ctx.String("detector"))

	// The deprecated flag names a runtime that every detector shares, so it can only ask for the
	// detector already configured, or for no regeneration at all.
	if detector == "" {
		if engine := strings.TrimSpace(ctx.String("engine")); engine != "" {
			if detector = face.DetectorAuto; face.ParseEngine(engine) == face.EngineNone {
				detector = ""
			}
		}
	}

	if detector != "" {
		if err := w.ResetAndReindex(detector, get.Index()); err != nil {
			return err
		}
	} else {
		if err := w.Reset(); err != nil {
			return err
		}
	}

	elapsed := time.Since(start)
	log.Infof("completed in %s", elapsed)

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

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	if err := query.RemovePeopleAndFaces(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	return nil
}

// facesIndexAction searches originals for faces.
func facesIndexAction(ctx *cli.Context) error {
	start := time.Now()

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	// Use first argument to limit scope if set.
	subPath, err := sanitizeSubfolderArg(ctx.Args().First())

	if err != nil {
		return err
	}

	if subPath == "" {
		log.Infof("finding faces in %s", clean.Log(conf.OriginalsPath()))
	} else {
		log.Infof("finding faces in %s", clean.Log(filepath.Join(conf.OriginalsPath(), subPath)))
	}

	if conf.ReadOnly() {
		log.Infof("config: enabled read-only mode")
	}

	var found fs.Done
	var lastFound, indexed int

	settings := conf.Settings()

	if w := get.Index(); w != nil {
		indexStart := time.Now()
		_, lastFound = w.LastRun()
		convert := settings.Index.Convert && conf.SidecarWritable()
		opt := photoprism.NewIndexOptions(subPath, true, convert, true, true, true, conf)

		found, indexed = w.Start(opt)

		log.Infof("index: updated %s [%s]", english.Plural(indexed, "file", "files"), time.Since(indexStart))
	}

	if w := get.Purge(); w != nil {
		opt := photoprism.PurgeOptions{
			Path:   subPath,
			Ignore: found,
			Force:  lastFound != len(found) || indexed > 0,
		}

		if files, photos, updated, err := w.Start(opt); err != nil {
			log.Error(err)
		} else if updated > 0 {
			log.Infof("purge: removed %s and %s", english.Plural(len(files), "file", "files"), english.Plural(len(photos), "photo", "photos"))
		}
	}

	elapsed := time.Since(start)

	log.Infof("indexed %s in %s", english.Plural(len(found), "file", "files"), elapsed)

	return nil
}

// facesUpdateAction performs face clustering and matching.
func facesUpdateAction(ctx *cli.Context) error {
	start := time.Now()

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	opt := photoprism.FacesOptions{
		Force: ctx.Bool("force"),
	}

	w := get.Faces()

	if err := w.Start(opt); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	return nil
}

// facesOptimizeAction optimizes existing face clusters.
func facesOptimizeAction(ctx *cli.Context) error {
	start := time.Now()

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	w := get.Faces()

	if ctx.Bool("retry") {
		if reset, err := query.ResetFaceMergeRetry(""); err != nil {
			return err
		} else if reset > 0 {
			log.Infof("faces: reset merge retry counters for %s", english.Plural(reset, "cluster", "clusters"))
		}
	}

	if res, err := w.Optimize(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("merged %s in %s", english.Plural(res.Merged, "face cluster", "face clusters"), elapsed)
	}

	return nil
}
