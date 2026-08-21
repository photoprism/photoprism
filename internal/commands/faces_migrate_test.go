package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
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
	t.Run("TargetMismatch", func(t *testing.T) {
		_, err := RunWithTestContext(FacesMigrateCommand, []string{"migrate", "--to=sface", "--dry-run", "--yes"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "configured model facenet")

		// Without an exit code a script cannot tell a refused migration from one that ran.
		exitErr, ok := err.(cli.ExitCoder)
		require.True(t, ok, "migration errors must set an exit status")
		assert.Equal(t, 1, exitErr.ExitCode())
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
