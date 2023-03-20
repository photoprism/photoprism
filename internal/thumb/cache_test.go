package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxAge_String(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, "2630000", MaxAge(2630000).String())
	})
}
