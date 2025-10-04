package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessed(t *testing.T) {
	t.Run("Jpeg", func(t *testing.T) {
		assert.True(t, Processed.Processed())
		assert.False(t, Found.Processed())
	})
}

func TestDoneProcessedCount(t *testing.T) {
	d := Done{
		"a.jpg": Found,
		"b.jpg": Processed,
		"c.jpg": 0,
	}
	assert.Equal(t, 1, d.Processed())
}
