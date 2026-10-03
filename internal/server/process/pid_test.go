package process

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestID(t *testing.T) {
	t.Run("Matches", func(t *testing.T) {
		assert.Equal(t, os.Getpid(), ID)
	})
}

func TestWritePID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "photoprism.pid")

		require.NoError(t, WritePID(fileName))

		data, err := os.ReadFile(fileName) //nolint:gosec // test file in a temporary directory
		require.NoError(t, err)
		assert.Equal(t, strconv.Itoa(ID), string(data))
	})
	t.Run("EmptyFileName", func(t *testing.T) {
		assert.NoError(t, WritePID(""))
	})
	t.Run("DirectoryNotWritable", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "missing")

		err := WritePID(filepath.Join(dir, "photoprism.pid"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "is not writable")
	})
}
