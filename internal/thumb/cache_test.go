package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxAge_String(t *testing.T) {
	t.Run("Hour", func(t *testing.T) {
		assert.Equal(t, "3600", MaxAge(3600).String())
	})
	t.Run("Month", func(t *testing.T) {
		assert.Equal(t, "2592000", MaxAge(2592000).String())
	})
}
