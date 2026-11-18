package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestFacesMigrateAction(t *testing.T) {
	conf := get.Config()
	require.NotNil(t, conf)
	options := conf.Options()
	modelsPath, configuredModel := options.ModelsPath, options.FaceModel
	t.Cleanup(func() {
		options.ModelsPath = modelsPath
		options.FaceModel = configuredModel
	})

	options.ModelsPath = t.TempDir()
	options.FaceModel = face.ModelFaceNet
	require.NoError(t, os.MkdirAll(filepath.Join(options.ModelsPath, "facenet"), fs.ModeDir))

	t.Run("DryRun", func(t *testing.T) {
		before, err := query.FaceMigrationCounts(face.ModelFaceNet)
		require.NoError(t, err)

		_, err = RunWithTestContext(FacesMigrateCommand, []string{"migrate", "--to=facenet", "--dry-run", "--yes"})
		require.NoError(t, err)

		after, err := query.FaceMigrationCounts(face.ModelFaceNet)
		require.NoError(t, err)
		assert.Equal(t, before, after)
	})
	t.Run("Confirm", func(t *testing.T) {
		// Without confirmation, the migration neither starts nor changes the configured model, which
		// differs from the target here.
		t.Setenv("PHOTOPRISM_CLI", "")
		options.FaceModel = face.ModelSFace
		t.Cleanup(func() { options.FaceModel = face.ModelFaceNet })

		before, err := query.FaceMigrationCounts(face.ModelFaceNet)
		require.NoError(t, err)

		// The command harness restores the options after each run, but not what the command saved.
		savedBefore, _ := os.ReadFile(conf.OptionsYaml())

		unchanged := func() {
			t.Helper()

			after, countErr := query.FaceMigrationCounts(face.ModelFaceNet)
			require.NoError(t, countErr)
			assert.Equal(t, before, after)
			assert.Equal(t, face.ModelSFace, options.FaceModel)
			assert.Equal(t, face.ModelSFace, conf.FaceModel())

			savedAfter, _ := os.ReadFile(conf.OptionsYaml())
			assert.Equal(t, string(savedBefore), string(savedAfter))
		}

		_, err = RunWithTestContext(FacesMigrateCommand, []string{"migrate", "--to=facenet"})

		var exit cli.ExitCoder
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 2, exit.ExitCode())
		assert.Contains(t, err.Error(), "could not ask for confirmation")
		unchanged()

		pipeResetAnswers(t, "n\n")

		_, err = RunWithTestContext(FacesMigrateCommand, []string{"migrate", "--to=facenet"})
		assert.NoError(t, err)
		unchanged()
	})
	t.Run("TargetNotInstalled", func(t *testing.T) {
		// A target that differs from the configured model is the command's normal input,
		// so what it refuses here is weights it cannot load.
		_, err := RunWithTestContext(FacesMigrateCommand, []string{"migrate", "--to=sface", "--dry-run", "--yes"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "embedding model sface is not installed")

		// Without an exit code a script cannot tell a refused migration from one that ran.
		exitErr, ok := err.(cli.ExitCoder)
		require.True(t, ok, "migration errors must set an exit status")
		assert.Equal(t, 1, exitErr.ExitCode())
	})
	t.Run("GatedTargetWithoutAcceptance", func(t *testing.T) {
		t.Setenv(face.LicenseAcceptanceVar, "")

		_, err := RunWithTestContext(FacesMigrateCommand, []string{"migrate", "--to=arcface_r50", "--dry-run", "--yes"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), face.LicenseAcceptanceVar)
	})
	t.Run("DisabledTarget", func(t *testing.T) {
		_, err := RunWithTestContext(FacesMigrateCommand, []string{"migrate", "--to=none", "--dry-run", "--yes"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot migrate to disabled embeddings")
	})
}

func TestFacesMigrateCommand(t *testing.T) {
	t.Run("DescribesTheRestart", func(t *testing.T) {
		// The lock keeps a running instance off the rows being replaced, so the operator does
		// not have to stop it. Nothing reloads the model the run records, though, so
		// "photoprism help" has to name the restart, which no guard can perform.
		assert.Contains(t, FacesMigrateCommand.Description, "restart the instance afterwards")
		assert.NotContains(t, FacesMigrateCommand.Description, "Stop the server")
		assert.NotEmpty(t, FacesMigrateCommand.Usage)
	})
}

// TestFacesMigrateExit pins that a migration which was applied but lost markers still exits with
// status 1, so a script can tell it from a clean run, and that its message names no restart.
func TestFacesMigrateExit(t *testing.T) {
	t.Run("Incomplete", func(t *testing.T) {
		for _, incomplete := range []*photoprism.FacesMigrateIncompleteError{
			{Failed: 900, Clusterable: 3},
			{Failed: 5, Unreadable: 5},
			{Failed: 9, Clusterable: 4, Unreadable: 5},
			{Failed: 7},
		} {
			err := facesMigrateExit(incomplete)

			var exit cli.ExitCoder
			require.ErrorAs(t, err, &exit)
			assert.Equal(t, 1, exit.ExitCode())
			assert.Equal(t, incomplete.Error(), err.Error())
			assert.True(t, strings.HasPrefix(err.Error(), "faces: migration applied; "))
			// The restart instruction is logged once, with the model the run recorded.
			assert.NotContains(t, err.Error(), "restart")
		}
	})
	t.Run("Aborted", func(t *testing.T) {
		aborted := &photoprism.FacesMigrateAbortedError{Failed: 12, Reason: "nothing could be re-embedded"}
		err := facesMigrateExit(aborted)

		var exit cli.ExitCoder
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 1, exit.ExitCode())
		assert.Equal(t, aborted.Error(), err.Error())
	})
	t.Run("Success", func(t *testing.T) {
		assert.NoError(t, facesMigrateExit(nil))
	})
}

// TestEmbeddingModelCountMessage pins that rows without a recorded embedding model are named as FaceNet,
// apart from the rows that record it, rather than as a model of their own.
func TestEmbeddingModelCountMessage(t *testing.T) {
	t.Run("Unrecorded", func(t *testing.T) {
		assert.Equal(t, "faces: 14246 markers have no recorded embedding model (treated as facenet)", embeddingModelCountMessage("", 14246, "marker"))
		assert.Equal(t, "faces: 399 clusters have no recorded embedding model (treated as facenet)", embeddingModelCountMessage("", 399, "cluster"))
		assert.Equal(t, "faces: 1 marker has no recorded embedding model (treated as facenet)", embeddingModelCountMessage("", 1, "marker"))
	})
	t.Run("Recorded", func(t *testing.T) {
		assert.Equal(t, "faces: embedding model facenet has 2 markers", embeddingModelCountMessage(face.ModelFaceNet, 2, "marker"))
		assert.Equal(t, "faces: embedding model sface has 1 cluster", embeddingModelCountMessage(face.ModelSFace, 1, "cluster"))
	})
}

// TestFaceModelSwitchedMessage pins that the line ending a migration that recorded its target
// carries the restart instruction.
func TestFaceModelSwitchedMessage(t *testing.T) {
	assert.Equal(t, "faces: the configured face model is now sface, restart the instance to load it",
		faceModelSwitchedMessage(face.ModelSFace))
}

func TestPercentOf(t *testing.T) {
	t.Run("Rounds", func(t *testing.T) {
		// A share of 14.69, so truncating would report a percent less.
		assert.Equal(t, 15, percentOf(6062, 41252))
	})
	t.Run("All", func(t *testing.T) {
		assert.Equal(t, 100, percentOf(9, 9))
	})
	t.Run("None", func(t *testing.T) {
		assert.Equal(t, 0, percentOf(0, 9))
	})
	t.Run("NothingToDivideBy", func(t *testing.T) {
		assert.Equal(t, 0, percentOf(3, 0))
	})
}

// TestReportMigrationCropCoverage covers the forecast an operator reads before the prompt: how much
// of the crop detail the cache already holds, how much the run renders for itself, and how much is
// missing from the originals, which is the only part nothing can recover.
func TestReportMigrationCropCoverage(t *testing.T) {
	capture := func(t *testing.T) *test.Hook {
		t.Helper()

		logger, ok := log.(*logrus.Logger)
		require.True(t, ok)

		hook := test.NewLocal(logger)
		system := event.SystemLog
		event.SystemLog = logger

		t.Cleanup(func() {
			event.SystemLog = system
			hook.Reset()
		})

		return hook
	}

	messages := func(hook *test.Hook) string {
		var b strings.Builder

		for _, entry := range hook.AllEntries() {
			b.WriteString(entry.Message)
			b.WriteString("\n")
		}

		return b.String()
	}

	t.Run("ForecastsTheRendering", func(t *testing.T) {
		hook := capture(t)

		reportMigrationCropCoverage(photoprism.FacesMigratePlan{
			CropCoverage: query.FaceMigrationCropCounts{Total: 41252, FullDetail: 19501, Upscaled: 15689, SourceTooSmall: 6062},
			ThumbSize:    thumb.Sizes[thumb.Fit1920],
		})

		out := messages(hook)
		assert.Contains(t, out, "15689 of 41252 markers (38%)")
		assert.Contains(t, out, "1920x1200")
		assert.Contains(t, out, "rendered again from the original")
		assert.Contains(t, out, "6062 of 41252 markers (15%)")

		// Nothing is asked of the operator: the run does this itself, and a warning would send
		// them to regenerate a thumbnail cache they do not need.
		assert.NotContains(t, out, "thumb-size")
	})
	t.Run("SmallOriginalsAreStatedApart", func(t *testing.T) {
		hook := capture(t)

		reportMigrationCropCoverage(photoprism.FacesMigratePlan{
			CropCoverage: query.FaceMigrationCropCounts{Total: 100, FullDetail: 40, SourceTooSmall: 60},
			ThumbSize:    thumb.Sizes[thumb.Fit4096],
		})

		out := messages(hook)
		assert.NotContains(t, out, "rendered again from the original")
		assert.Contains(t, out, "60 of 100 markers (60%)")
		assert.Contains(t, out, "stay upscaled")
	})
	t.Run("NothingMeasured", func(t *testing.T) {
		// A library with no renditions at all is cropped from the originals, so there is no
		// coverage to report on.
		hook := capture(t)

		reportMigrationCropCoverage(photoprism.FacesMigratePlan{})

		assert.Empty(t, messages(hook))
	})
}
