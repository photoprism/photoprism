package mutex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTempArchives(t *testing.T) {
	t.Run("DefaultsToTrue", func(t *testing.T) {
		// Starts set so the first sweep after a restart also covers archives left by a previous process.
		assert.True(t, TempArchives.Load())
	})
	t.Run("StoreAndLoad", func(t *testing.T) {
		t.Cleanup(func() { TempArchives.Store(true) })
		TempArchives.Store(false)
		assert.False(t, TempArchives.Load())
		TempArchives.Store(true)
		assert.True(t, TempArchives.Load())
	})
}
