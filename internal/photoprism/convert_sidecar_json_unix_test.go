//go:build unix

package photoprism

import (
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExifToolCacheValid_Pipe(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "cache.json")
	require.NoError(t, syscall.Mkfifo(fileName, 0o600))
	assert.False(t, exifToolCacheValid(fileName))
}
