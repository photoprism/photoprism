package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAltitude(t *testing.T) {
	t.Run("0", func(t *testing.T) {
		assert.Equal(t, 0, Altitude(0.0))
	})
	t.Run("Negative", func(t *testing.T) {
		assert.Equal(t, -1234, Altitude(-1234.2))
	})
	t.Run("Positive", func(t *testing.T) {
		assert.Equal(t, 9234, Altitude(9234.4))
	})
	t.Run("TooLarge", func(t *testing.T) {
		assert.Equal(t, 0, Altitude(4294967284))
	})
	t.Run("TooSmall", func(t *testing.T) {
		assert.Equal(t, 0, Altitude(-4294967284))
	})
}
