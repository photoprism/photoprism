package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessed(t *testing.T) {
	t.Run("jpeg", func(t *testing.T) {
		assert.True(t, Processed.Processed())
		assert.False(t, Found.Processed())
	})
}
