package latlng

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRound(t *testing.T) {
	t.Run("Germany", func(t *testing.T) {
		assert.Equal(t, 48.5634483, Round(48.56344833333333))
		assert.Equal(t, 8.9968783, Round(8.996878333333333))
	})
}

func TestRoundCoords(t *testing.T) {
	t.Run("Germany", func(t *testing.T) {
		lat, lng := RoundCoords(48.56344833333333, 8.996878333333333)
		assert.Equal(t, 48.5634483, lat)
		assert.Equal(t, 8.9968783, lng)
	})
}
