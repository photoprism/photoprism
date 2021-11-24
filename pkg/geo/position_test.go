package geo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPosition_InRange(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		pos := Position{Lat: 15.2, Lng: -4.0}
		assert.True(t, pos.InRange(14.2, -3.0, 1.5))
	})
	t.Run("Zero", func(t *testing.T) {
		pos := Position{Lat: 0, Lng: 0, Estimate: true}
		assert.False(t, pos.InRange(0.1, -0.1, 1.5))
		assert.True(t, pos.Estimate)
	})
	t.Run("False", func(t *testing.T) {
		pos := Position{Lat: 15.2, Lng: -4.0}
		assert.False(t, pos.InRange(13.2, -3.0, 1.5))
		assert.False(t, pos.Estimate)
	})
}

func TestPosition_Randomize(t *testing.T) {
	t.Run("RandomizeKm", func(t *testing.T) {
		pos := Position{Lat: 15.2, Lng: 0.0}

		assert.Equal(t, 15.2, pos.Lat)
		assert.Equal(t, 0.0, pos.Lng)
		assert.Equal(t, 0, pos.Accuracy)

		pos.Randomize(Meter * 1000)

		assert.Equal(t, 1000, pos.Accuracy)

		if pos.Lat == 15.2 && pos.Lng == 0.0 {
			t.Errorf("randomized %s should have changed", pos.String())
		} else {
			t.Logf("randomized %s", pos.String())
		}
	})
	t.Run("RandomizeMeter", func(t *testing.T) {
		pos := Position{Lat: 15.2, Lng: 0.0}

		assert.Equal(t, 15.2, pos.Lat)
		assert.Equal(t, 0.0, pos.Lng)
		assert.Equal(t, 0, pos.Accuracy)

		pos.Randomize(Meter)

		assert.Equal(t, 1, pos.Accuracy)

		if pos.Lat == 15.2 && pos.Lng == 0.0 {
			t.Errorf("randomized %s should have changed", pos.String())
		} else {
			t.Logf("randomized %s", pos.String())
		}
	})
}
