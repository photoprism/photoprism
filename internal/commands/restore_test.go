package commands

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/photoprism/get"
)

func TestMigrateRestoredIndex(t *testing.T) {
	_, check := prepareRestoreTest(t)

	require.NoError(t, migrateRestoredIndex(get.Config()))

	check()
}

func TestRestoreCommand(t *testing.T) {
	t.Run("Database", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "index.sql")
		change, check := prepareRestoreTest(t)

		_, err := RunWithTestContext(BackupCommand, []string{"backup", "-i", fileName})
		require.NoError(t, err)

		change()

		_, err = RunWithTestContext(RestoreCommand, []string{"restore", "-i", "--force", fileName})
		require.NoError(t, err)

		check()
	})
	t.Run("AlbumsOnly", func(t *testing.T) {
		_, check := prepareRestoreTest(t)

		_, err := RunWithTestContext(RestoreCommand, []string{"restore", "--albums-path", t.TempDir()})
		require.NoError(t, err)

		check()
	})
}
