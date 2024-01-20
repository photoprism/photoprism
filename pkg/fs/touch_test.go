package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTouch(t *testing.T) {
	t.Run("Testdata", func(t *testing.T) {
		fileName := filepath.Join("testdata", PPStorageFilename)
		assert.False(t, FileExists(fileName))
		assert.False(t, FileExistsNotEmpty(fileName))

		// Create "testdata/.ppstorage" file.
		assert.NoError(t, Touch(fileName))

		assert.True(t, FileExists(fileName))
		assert.True(t, FileExistsNotEmpty(fileName))

		// Delete "testdata/.ppstorage" file.
		assert.NoError(t, os.Remove(fileName))

		assert.False(t, FileExists(fileName))
		assert.False(t, FileExistsNotEmpty(fileName))
	})
}
