package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestDatabase(t *testing.T) {
	t.Run("Force", func(t *testing.T) {
		backupPath := filepath.Join(t.TempDir(), "testdata", "sqlite")
		require.NoError(t, os.MkdirAll(backupPath, fs.ModeDir))
		t.Cleanup(func() { _ = os.RemoveAll(backupPath) })

		require.NoError(t, Database(backupPath, "", false, true, 2))
	})
	t.Run("ForceStdOut", func(t *testing.T) {
		backupPath := filepath.Join(t.TempDir(), "testdata", "sqlite")
		require.NoError(t, os.MkdirAll(backupPath, fs.ModeDir))
		t.Cleanup(func() { _ = os.RemoveAll(backupPath) })

		require.NoError(t, Database(backupPath, "", true, true, 2))
	})
}
