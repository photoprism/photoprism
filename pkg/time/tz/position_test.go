package tz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPosition(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Position(0, 0))
	})
	t.Run("EuropeBerlin", func(t *testing.T) {
		assert.Equal(t, "Europe/Berlin", Position(52.472833, 13.407500))
	})
	t.Run("OpenSea", func(t *testing.T) {
		// Nautical zones on the equator, far from any coast.
		assert.Equal(t, "UTC-10", Position(0.5, -149))
		assert.Equal(t, "UTC-11", Position(0.5, -171.5))
		assert.Equal(t, "UTC-12", Position(0.5, -179))
		assert.Equal(t, "UTC+10", Position(0.5, 151))
		assert.Equal(t, "UTC+11", Position(0.5, 158.5))
		assert.Equal(t, "UTC-6", Position(0.5, -85))
	})
}
