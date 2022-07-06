package fs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSymlinksSupported(t *testing.T) {
	t.Run("TempDir", func(t *testing.T) {
		ok, err := SymlinksSupported(os.TempDir())
		assert.NoError(t, err)
		assert.True(t, ok)
	})
}
