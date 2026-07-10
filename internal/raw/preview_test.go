package raw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreviewExtAllowed(t *testing.T) {
	t.Run("Allowed", func(t *testing.T) {
		assert.True(t, PreviewExtAllowed(".cr3"))
		assert.True(t, PreviewExtAllowed(".dng"))
		assert.True(t, PreviewExtAllowed(".nef"))
	})
	t.Run("Unsafe", func(t *testing.T) {
		assert.False(t, PreviewExtAllowed(".mos"))
	})
}
