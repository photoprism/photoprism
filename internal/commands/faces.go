package commands

import (
	"context"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
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

// FacesResetDescription explains what each scope of the faces reset command removes and keeps.
const FacesResetDescription = "Without flags, this command removes the automatically recognized faces and matches, as well as people who were created from faces and are left without any, and keeps the names you assigned. " +
	"With --all, it also removes the names you assigned and unverified people, keeping the markers. " +
	"With --force, it removes all faces, people, and markers, so faces must be detected again with \"photoprism faces index\". " +
	"With --detector, it removes all faces instead of only the automatically recognized ones, and detects faces in all pictures again, " +
	"which updates the markers the detector finds, adds new ones, and removes the unnamed markers it does not find. " +
	"Run it while the instance is stopped or idle, and set the face detector in the configuration to the same value. " +
	"Afterwards, run \"photoprism faces update\" to recognize faces again."

// FacesCommands configures the command name, flags, and action.
var FacesCommands = &cli.Command{
	Name:  "faces",
	Usage: "Face recognition subcommands",
	// Ordered as an operator meets them: what the instance is doing, the passes that change the
	// index, then the reports that describe it, and last the ones that diagnose or destroy.
	Subcommands: []*cli.Command{
		FacesStatusCommand,
		{
			Name:  "update",
			Usage: "Performs face clustering and matching",
			Flags: []cli.Flag{
				ForceFlag("update all faces"),
			},
			Action: facesUpdateAction,
		},
		{
			Name:      "index",
			Usage:     "Searches originals for faces",
			ArgsUsage: "[subfolder]",
			Action:    facesIndexAction,
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
		FacesMigrateCommand,
		FacesListCommand,
		FacesMarkersCommand,
		FacesSubjectsCommand,
		FacesConflictsCommand,
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
			Name:   "stats",
			Usage:  "Shows stats on face samples",
			Action: facesStatsAction,
		},
		{
			Name:        "reset",
			Usage:       "Removes people and faces after confirmation",
			Description: FacesResetDescription,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "all",
					Aliases: []string{"a"},
					Usage:   "removes all faces, names, and unverified people, keeping the markers",
				},
				ForceFlag("removes all faces, people, and markers, so faces must be detected again"),
				&cli.StringFlag{
					Name:  "detector",
					Usage: "regenerates all markers with the detection model `NAME` (" + face.DetectorUsageString() + ")",
				},
				&cli.StringFlag{
					Name:   "engine",
					Usage:  "regenerate markers using detection engine `NAME` *deprecated*, use --detector",
					Hidden: true,
				},
				&cli.BoolFlag{
					Name:    "trace",
					Aliases: []string{"t"},
					Usage:   "shows trace logs for debugging",
				},
				YesFlag(),
			},
			Action: facesResetAction,
		},
	},
}

// FacesMigrateCommand configures the face embedding migration command.
var FacesMigrateCommand = &cli.Command{
	Name:  "migrate",
	Usage: "Migrates face embeddings to a supported model",
	Description: "This is how the face embedding model is changed: every marker is re-embedded and " +
		"the target is recorded as the configured model. It defaults to " + face.DefaultModelName() +
		", the model this release supports, so an ordinary migration needs no target. A running " +
		"instance does not have to be stopped, but start this when no indexing or import is under " +
		"way, and restart the instance afterwards to load the model it recorded.",
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
		// Ready counts the markers already on the target model, which a re-run still samples again
		// where they record no extent - the line below carries that number. Unlinked markers are
		// cleared by every run regardless of how the migration goes.
		log.Infof(
			"faces: %d markers already use %s, %d have no file, and %d were identified manually",
			plan.Markers.Ready, clean.Log(plan.Target), plan.Markers.Unlinked, plan.Markers.Manual,
		)
		// The crop is an axis of the embedding space, so a detector change leaves a library in
		// two of them. This is the only run that repairs that, and it is why a re-run to the
		// same model can still have work to do. Every marker indexed before the detector or the
		// sample extent was recorded counts here, which on a first run is all of them.
		if plan.RecropMarkers > 0 {
			log.Infof("faces: %d of those were cropped by another or an unrecorded detector, or record no sample extent, and are re-embedded, keeping their vector if detection cannot find them again",
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
		// A crop is taken from a thumbnail and never from the original, so what the cache holds
		// decides how much detail the vectors rest on. The run renders what it is missing; this
		// says how much of that is coming, and how much of it the originals cannot supply.
		reportMigrationCropCoverage(plan)

		if ctx.Bool("dry-run") {
			log.Infof("faces: dry run completed without changes")
			return nil
		}

		// A running instance reads the lock file this run takes, so it starts no new indexing and
		// refuses people edits. A pass already under way is not interrupted, and what it writes can
		// roll the finalize back, which is why the warning names that as well as the restart.
		event.SystemWarn([]string{"faces", "migrate", "this replaces every face cluster; a running instance " +
			"starts no new indexing and refuses people edits while it runs, but a pass already under way can " +
			"still force a re-run; restart the instance afterwards to load %s"}, clean.Log(plan.Target))

		if !RunNonInteractively(ctx.Bool("yes")) {
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Migrate all face embeddings to %s?", plan.Target),
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
		// The cache is what a crop is taken from, so a run that had to render is the difference
		// between this library's vectors and the ones a pre-generated cache would have produced.
		if result.RenderedThumbs > 0 {
			log.Infof("faces: rendered %d thumbnail(s) from originals so their face crops were not upscaled",
				result.RenderedThumbs)
		}
		// The one part of the run whose cost is not visible in the result afterwards: these
		// markers hold a vector drawn from fewer pixels than their original could supply.
		if result.FailedThumbs > 0 {
			event.SystemWarn([]string{"faces", "migrate", "%d file(s) could not be given the wider thumbnail their face crops need, " +
				"so those markers were embedded from upscaled crops; re-run once the cache volume is writable"},
				result.FailedThumbs)
		}
		// The other cost that leaves no trace in the vectors: an aligned model was trained on
		// pose-normalized faces, and these reached it as a plain box crop instead.
		if result.UnalignedCrops > 0 {
			log.Infof("faces: %d marker(s) could not be aligned and were embedded from a plain box crop",
				result.UnalignedCrops)
		}
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

// reportMigrationCropCoverage states how much of the crop detail the thumbnail cache already holds,
// what the run renders for itself, and what no rendition can supply.
//
// A forecast rather than a warning: the run renders the renditions its crops need as it reaches
// each file, so the only number an operator has to decide anything about is the last one.
func reportMigrationCropCoverage(plan photoprism.FacesMigratePlan) {
	coverage := plan.CropCoverage

	if coverage.Total < 1 {
		return
	}

	// Markers rather than files, because that is the population the buckets count: a file with
	// several of them is rendered for once, so the number of renditions is lower than this.
	if coverage.Upscaled > 0 {
		log.Infof("faces: %d of %d markers (%d%%) need a wider crop than the largest thumbnail this library holds (%dx%d), "+
			"so their files are rendered again from the original as the run reaches them",
			coverage.Upscaled, coverage.Total, percentOf(coverage.Upscaled, coverage.Total),
			plan.ThumbSize.Width, plan.ThumbSize.Height)
	}

	// Stated apart, because this is the part no rendition recovers: their files are re-rendered
	// as well, at the resolution the original holds, and the crop is still upscaled onto the
	// template afterwards.
	if coverage.SourceTooSmall > 0 {
		log.Infof("faces: %d of %d markers (%d%%) have originals too small for a full-detail face crop, so theirs stay upscaled",
			coverage.SourceTooSmall, coverage.Total, percentOf(coverage.SourceTooSmall, coverage.Total))
	}
}

// percentOf returns the share of total in whole percent, and 0 when there is nothing to divide by.
func percentOf(n, total int) int {
	if total < 1 {
		return 0
	}

	return int(math.Round(float64(n) * 100 / float64(total)))
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

// facesResetAction removes faces, matches, and people in the scope the flags select, and regenerates
// the markers when a detector is named.
func facesResetAction(ctx *cli.Context) error {
	if ctx.Bool("force") {
		// The two do not compose: --force removes every person, face and marker, and the names go
		// with them, so a caller who also asked to regenerate would silently get the destructive
		// half alone. Refused rather than reordered, because which of the two they meant is not
		// knowable from the command.
		if ctx.IsSet("detector") || ctx.IsSet("engine") {
			return cli.Exit("faces: --force removes all faces, people, and markers, so it cannot be combined with --detector", 2)
		}

		// Refused rather than treated as the wider of the two: the flags name different outcomes
		// for the markers table, and which one a caller meant is not knowable from the command.
		if ctx.Bool("all") {
			return cli.Exit("faces: --force also removes the markers, so it cannot be combined with --all", 2)
		}

		return facesResetAllAction(ctx)
	}

	all := ctx.Bool("all")
	detector := facesResetDetector(ctx)

	if !face.KnownDetectorName(detector) {
		return cli.Exit(fmt.Sprintf("faces: unsupported face detector %s", clean.Log(detector)), 2)
	}

	regenerate := detector != "" && face.ParseDetectorName(detector) != face.DetectorNone
	detectorName := detector

	// Auto names the configured detector, which only the config knows. It is loaded without the
	// database, so a declined or refused prompt changes nothing.
	if regenerate && face.ParseDetectorName(detector) == face.DetectorAuto {
		coreConf, coreErr := InitCoreConfig(ctx, false)

		if coreErr != nil {
			return coreErr
		} else if detectorName = facesResetDetectorName(coreConf, detector); face.ParseDetectorName(detectorName) == face.DetectorNone {
			return cli.Exit("faces: no face detector can be used, so markers cannot be regenerated", 2)
		}
	}

	// The run cannot see a running instance, so the operator is asked to rule out a concurrent pass.
	if regenerate && !RunNonInteractively(ctx.Bool("yes")) {
		log.Warnf("faces: make sure the instance is stopped or idle before you continue")
	}

	if proceed, confirmErr := ConfirmAction(ctx.Bool("yes"), facesResetLabel(all, detectorName)); confirmErr != nil {
		return confirmErr
	} else if !proceed {
		log.Infof("faces: no faces were removed")
		return nil
	}

	if ctx.Bool("trace") {
		log.SetLevel(logrus.TraceLevel)
		log.Infoln("reset: enabled trace mode")
	}

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	start := time.Now()
	w := get.Faces()

	var resetErr error

	switch {
	case detector != "":
		resetErr = w.ResetAndReindex(detector, get.Index(), all)
	case all:
		resetErr = w.ResetAll()
	default:
		resetErr = w.Reset()
	}

	if resetErr != nil {
		return resetErr
	}

	elapsed := time.Since(start)
	log.Infof("completed in %s", elapsed)

	return nil
}

// facesResetDetector returns the detector the markers are regenerated with, or "" to reset only.
// The deprecated --engine flag names a runtime every detector shares, so it can only ask for the
// configured detector, or for no regeneration at all.
func facesResetDetector(ctx *cli.Context) string {
	if detector := strings.TrimSpace(ctx.String("detector")); detector != "" {
		return detector
	}

	if engine := strings.TrimSpace(ctx.String("engine")); engine != "" && face.ParseEngine(engine) != face.EngineNone {
		return face.DetectorAuto
	}

	return ""
}

// facesResetDetectorName returns the name of the detector a reset regenerates the markers with, which
// is the configured one when auto is requested.
func facesResetDetectorName(conf *config.Config, detector string) string {
	if conf != nil && detector != "" && face.ParseDetectorName(detector) == face.DetectorAuto {
		return string(conf.FaceDetector())
	}

	return detector
}

// facesResetLabel returns the confirmation prompt for the scope and detector a reset was given.
func facesResetLabel(all bool, detector string) string {
	var with string

	switch face.ParseDetectorName(detector) {
	case face.DetectorNone:
	case face.DetectorAuto:
		if strings.TrimSpace(detector) != "" {
			with = "the configured detector"
		}
	default:
		with = clean.Log(detector)
	}

	switch {
	case with != "" && all:
		return fmt.Sprintf("Remove all faces, matches, names, and unverified people, then detect faces in all pictures with %s "+
			"and remove the unnamed markers it does not find again?", with)
	case with != "":
		return fmt.Sprintf("Remove all faces and automatic matches, then detect faces in all pictures with %s "+
			"and remove the unnamed markers it does not find again?", with)
	case all:
		return "Remove all faces and matches, including names and unverified people, keeping the markers?"
	default:
		return "Remove automatically recognized faces, matches, and people left without faces?"
	}
}

// facesResetAllAction removes all people, faces, and face markers.
func facesResetAllAction(ctx *cli.Context) error {
	if proceed, err := ConfirmAction(ctx.Bool("yes"), "Permanently remove all faces, people, and markers?"); err != nil {
		return err
	} else if !proceed {
		log.Infof("faces: no people or faces were removed")
		return nil
	}

	if ctx.Bool("trace") {
		log.SetLevel(logrus.TraceLevel)
		log.Infoln("reset: enabled trace mode")
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
