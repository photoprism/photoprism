package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	})
	t.Run("DisabledTarget", func(t *testing.T) {
		_, err := RunWithTestContext(FacesMigrateCommand, []string{"migrate", "--to=none", "--dry-run", "--yes"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot migrate to disabled embeddings")
	})
}
