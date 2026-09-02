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
	t.Run("DescribesTheServerConstraint", func(t *testing.T) {
		// The worker guards are process-local, so the operator is the only thing that can
		// keep a running instance away from the rows being replaced. "photoprism help" has
		// to say so, because nothing in the code can enforce it.
		assert.Contains(t, FacesMigrateCommand.Description, "Stop the server")
		assert.NotEmpty(t, FacesMigrateCommand.Usage)
	})
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
