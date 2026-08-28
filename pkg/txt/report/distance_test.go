package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDistance(t *testing.T) {
	t.Run("Threshold", func(t *testing.T) {
		// Reads against a configured threshold without rounding it away.
		assert.Equal(t, "0.850", Distance(0.85))
		assert.Equal(t, "0.640", Distance(0.64))
	})
	t.Run("Zero", func(t *testing.T) {
		assert.Equal(t, "0.000", Distance(0))
	})
	t.Run("Unset", func(t *testing.T) {
		// The sentinel a marker carries when nothing matched it.
		assert.Equal(t, "-1.000", Distance(-1))
	})
	t.Run("Rounds", func(t *testing.T) {
		assert.Equal(t, "0.929", Distance(0.9294))
		assert.Equal(t, "0.930", Distance(0.9296))
		// Not 0.930: the literal is a hair below the midpoint in binary, so a half-way case is
		// decided by the stored value rather than by the decimal it was written as.
		assert.Equal(t, "0.929", Distance(0.9295))
	})
}
