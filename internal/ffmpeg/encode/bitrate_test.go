package encode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBitrate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, 25, ParseBitrate("25M"))
		assert.Equal(t, 1, ParseBitrate("1M"))
		assert.Equal(t, 960, ParseBitrate(" 960M "))
	})
	t.Run("Clamped", func(t *testing.T) {
		assert.Equal(t, MaxBitrateLimit, ParseBitrate("961M"))
		assert.Equal(t, MaxBitrateLimit, ParseBitrate("9223372036854775807M"))
		assert.Equal(t, MaxBitrateLimit, ParseBitrate("99999999999999999999M"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, 0, ParseBitrate(""))
		assert.Equal(t, 0, ParseBitrate("M"))
		assert.Equal(t, 0, ParseBitrate("25"))
		assert.Equal(t, 0, ParseBitrate("25k"))
		assert.Equal(t, 0, ParseBitrate("25m"))
		assert.Equal(t, 0, ParseBitrate("2.5M"))
		assert.Equal(t, 0, ParseBitrate("0M"))
		assert.Equal(t, 0, ParseBitrate("-1M"))
		assert.Equal(t, 0, ParseBitrate("-99999999999999999999M"))
	})
}
