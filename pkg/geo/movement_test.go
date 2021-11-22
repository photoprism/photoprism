package geo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMovement(t *testing.T) {
	t.Run("BerlinShanghai", func(t *testing.T) {
		berlin := Position{52.5243700, 13.4105300}
		shanghai := Position{31.2222200, 121.4580600}

		time1 := time.Date(2015, 5, 17, 17, 48, 46, 0, time.UTC)
		time2 := time.Date(2015, 5, 17, 23, 14, 34, 0, time.UTC)

		result := NewMovement(berlin, shanghai, time1, time2)

		assert.Equal(t, 8396, int(result.DistKm))
		assert.Equal(t, 19548, int(time2.Sub(time1).Seconds()))
		assert.Equal(t, 1546, int(result.SpeedKmh))
		assert.Equal(t, -21, int(result.LatDiff))
		assert.Equal(t, 108, int(result.LngDiff))
		assert.Equal(t, 5, int(result.Hours()))
		assert.Equal(t, 19548, int(result.Seconds()))

		timeEst := time.Date(2015, 5, 17, 18, 14, 34, 0, time.UTC)

		posEst := result.Position(timeEst)

		assert.Equal(t, 50, int(posEst.Lat))
		assert.Equal(t, 21, int(posEst.Lng))

		posMid := result.Midpoint()

		assert.Equal(t, 41, int(posMid.Lat))
		assert.Equal(t, 67, int(posMid.Lng))
	})
}
