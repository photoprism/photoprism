package geo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMovement(t *testing.T) {
	t.Run("BerlinShanghai", func(t *testing.T) {
		berlin := Position{
			Name:     "Berlin",
			Time:     time.Date(2015, 5, 17, 11, 48, 46, 0, time.UTC),
			Lat:      52.5243700,
			Lng:      13.4105300,
			Altitude: 34}

		shanghai := Position{
			Name:     "Shanghai",
			Time:     time.Date(2015, 5, 17, 23, 14, 34, 0, time.UTC),
			Lat:      31.2222200,
			Lng:      121.4580600,
			Altitude: 4}

		result := NewMovement(berlin, shanghai)
		t.Log(result.String())

		// movement from 2015-05-17 23:14:34 to 2015-05-17 11:48:46 in 41148.000000 s,
		// Δ lat -21.302150, Δ lng  108.047530, dist 8396.511771 km, speed 734.602955 km/h
		assert.InEpsilon(t, 41148.00, result.Seconds(), 0.01)
		assert.InEpsilon(t, 11.43, result.Hours(), 0.01)
		assert.InEpsilon(t, -21.302150, result.DegLat(), 0.01)
		assert.InEpsilon(t, 108.047530, result.DegLng(), 0.01)
		assert.InEpsilon(t, 8396.511771, result.Km(), 0.01)
		assert.InEpsilon(t, 734.602955, result.Speed(), 0.1)
		assert.InEpsilon(t, 19, result.AverageAltitude(), 0.001)

		posEst1 := result.EstimatePosition(time.Date(2015, 5, 17, 12, 05, 22, 0, time.UTC))
		t.Log(posEst1.String())

		// 2015-05-17 12:05:22 @ 52.008745, 16.025854
		// estimate @ 31.325956, 120.931898, 4.000000 m
		assert.InEpsilon(t, 52.008745, posEst1.Lat, 0.01)
		assert.InEpsilon(t, 16.025854, posEst1.Lng, 0.01)
		assert.True(t, posEst1.Estimate)

		posEst2 := result.EstimatePosition(time.Date(2015, 5, 17, 18, 14, 34, 0, time.UTC))
		t.Log(posEst2.String())

		// 2015-05-17 18:14:34 @ 40.540746, 74.193174
		assert.InEpsilon(t, 40.540746, posEst2.Lat, 0.01)
		assert.InEpsilon(t, 74.193174, posEst2.Lng, 0.01)
		assert.True(t, posEst2.Estimate)

		posMid := result.Midpoint()
		t.Log(posMid.String())

		// midpoint @ 41.873295, 67.434295
		assert.InEpsilon(t, 41.873295, posMid.Lat, 0.01)
		assert.InEpsilon(t, 67.434295, posMid.Lng, 0.01)
	})

	t.Run("PositionBefore", func(t *testing.T) {
		timeEst := time.Date(2019, time.July, 21, 11, 56, 47, 0, time.UTC)

		pos1 := Position{
			Name: "Pos1",
			Time: time.Date(2019, time.July, 21, 15, 37, 44, 0, time.UTC),
			Lat:  48.2992,
			Lng:  8.92953}

		pos2 := Position{
			Name: "Pos2",
			Time: time.Date(2019, time.July, 21, 15, 37, 42, 0, time.UTC),
			Lat:  48.2992,
			Lng:  8.92954}

		result := NewMovement(pos1, pos2)

		t.Log(result.String())

		// from 2019-07-21 15:37:42 to 2019-07-21 15:37:44 in 2.000000 s
		// Δ lat 0.000000, Δ lng -0.000010, dist 0.000740 km, speed 1.331485 km/h
		assert.InEpsilon(t, 2.000000, result.Seconds(), 0.01)
		assert.Equal(t, 0.000000, result.DegLat(), 0.01)
		assert.InEpsilon(t, -0.000010, result.DegLng(), 0.01)
		assert.InEpsilon(t, 0.000740, result.Km(), 0.01)
		assert.InEpsilon(t, 1.331485, result.Speed(), 0.01)

		posEst := result.EstimatePosition(timeEst)

		t.Log(posEst.String())

		// 2019-07-21 11:56:47 @ 48.299200, 8.930116
		assert.InEpsilon(t, 48.299200, posEst.Lat, 0.01)
		assert.InEpsilon(t, 8.930116, posEst.Lng, 0.01)
		assert.True(t, posEst.Estimate)

		posMid := result.Midpoint()

		t.Log(posMid.String())

		// midpoint @ 48.299200, 8.929535
		assert.InEpsilon(t, 48.299200, posMid.Lat, 0.001)
		assert.InEpsilon(t, 8.929535, posMid.Lng, 0.01)
	})

	t.Run("TooFast", func(t *testing.T) {
		timeEst := time.Date(2019, time.July, 21, 11, 56, 47, 0, time.UTC)

		pos1 := Position{Name: "Pos1", Time: time.Date(2019, time.July, 21, 15, 37, 38, 0, time.UTC), Lat: 48.3032, Lng: 8.92772}
		pos2 := Position{Name: "Pos2", Time: time.Date(2019, time.July, 21, 15, 37, 42, 0, time.UTC), Lat: 48.2992, Lng: 8.92954}

		result := NewMovement(pos1, pos2)

		t.Log(result.String())

		// from 2019-07-21 15:37:38 to 2019-07-21 15:37:42 in 4.000000 s
		// Δ lat -0.004000, Δ lng 0.001820, dist 0.464707 km, speed 418.235966 km/h
		assert.InEpsilon(t, 4.0, result.Seconds(), 0.01)
		assert.InEpsilon(t, -0.004000, result.DegLat(), 0.01)
		assert.InEpsilon(t, 0.001820, result.DegLng(), 0.01)
		assert.InEpsilon(t, 0.464707, result.Km(), 0.01)
		assert.InEpsilon(t, 418.23596, result.Speed(), 0.1)

		posEst := result.EstimatePosition(timeEst)

		t.Log(posEst.String())

		// midpoint @ 48.301200, 8.928630
		assert.InEpsilon(t, 48.301200, posEst.Lat, 0.01)
		assert.InEpsilon(t, 8.928630, posEst.Lng, 0.01)
		assert.True(t, posEst.Estimate)

		posMid := result.Midpoint()

		t.Log(posMid.String())

		// midpoint @ 48.301200, 8.928630
		assert.InEpsilon(t, 48.301200, posMid.Lat, 0.01)
		assert.InEpsilon(t, 8.928630, posMid.Lng, 0.01)
	})

	t.Run("PositionBetween", func(t *testing.T) {
		timeEst := time.Date(2019, time.July, 21, 11, 56, 47, 0, time.UTC)

		pos1 := Position{Name: "Pos1", Time: time.Date(2019, time.July, 21, 15, 37, 44, 0, time.UTC), Lat: 48.2992, Lng: 8.92953}
		pos2 := Position{Name: "Pos2", Time: time.Date(2019, time.July, 20, 19, 33, 16, 0, time.UTC), Lat: 48.2994, Lng: 8.92914}

		result := NewMovement(pos1, pos2)

		t.Log(result.String())

		// movement from 2019-07-20 19:33:16 to 2019-07-21 15:37:44 in 72268.000000 s
		// Δ lat -0.000200, Δ lng 0.000390, dist 0.036426 km, speed 0.001815 km/
		assert.InEpsilon(t, 72268.000000, result.Seconds(), 0.01)
		assert.InEpsilon(t, -0.000200, result.DegLat(), 0.01)
		assert.InEpsilon(t, 00.000390, result.DegLng(), 0.01)
		assert.InEpsilon(t, 0.036426, result.Km(), 0.01)
		assert.InEpsilon(t, 0.001815, result.Speed(), 0.1)

		posEst := result.EstimatePosition(timeEst)

		t.Log(posEst.String())

		// 2019-07-21 11:56:47: lat 48.299237, lng 8.929458
		assert.InEpsilon(t, 48.299237, posEst.Lat, 0.01)
		assert.InEpsilon(t, 8.929458, posEst.Lng, 0.01)
		assert.True(t, posEst.Estimate)

		posMid := result.Midpoint()

		t.Log(posMid.String())

		// midpoint: lat 48.299300, lng 8.929335
		assert.InEpsilon(t, 48.299300, posMid.Lat, 0.01)
		assert.InEpsilon(t, 8.929335, posMid.Lng, 0.01)
	})

	t.Run("NotRealistic", func(t *testing.T) {
		timeEst := time.Date(2013, time.August, 10, 00, 05, 37, 0, time.UTC)

		time1 := time.Date(2013, time.August, 9, 17, 9, 0, 0, time.UTC)
		time2 := time.Date(2013, time.August, 9, 17, 8, 44, 0, time.UTC)

		pos1 := Position{Name: "Pos1", Time: time1, Lat: 52.6648, Lng: 13.3387}
		pos2 := Position{Name: "Pos2", Time: time2, Lat: 48.5193, Lng: 9.04933}

		result := NewMovement(pos1, pos2)

		t.Log(result.String())

		// movement from 2013-08-09 17:08:44 to 2013-08-09 17:09:00 in 16.000000 s
		// Δ lat 4.145500, Δ lng  4.289370, dist 551.290399 km, speed 124040.339691 km/h
		assert.InEpsilon(t, 16.000000, result.Seconds(), 0.01)
		assert.InEpsilon(t, 4.145500, result.DegLat(), 0.01)
		assert.InEpsilon(t, 4.289370, result.DegLng(), 0.01)
		assert.InEpsilon(t, 551.290399, result.Km(), 0.01)
		assert.InEpsilon(t, 124040.339691, result.Speed(), 0.1)
		assert.False(t, result.Realistic())

		posEst := result.EstimatePosition(timeEst)

		t.Log(posEst.String())

		// estimate @ 52.664800, 13.338700, alt 0.000000 m ± 275645 m
		assert.InEpsilon(t, 52.664800, posEst.Lat, 0.01)
		assert.InEpsilon(t, 13.338700, posEst.Lng, 0.01)
		assert.InEpsilon(t, 275645, posEst.Accuracy, 0.1)
		assert.True(t, posEst.Estimate)

		posMid := result.Midpoint()

		t.Log(posMid.String())

		// midpoint @ 50.592050, 11.194015, alt 0.000000 m ± 0 m
		assert.InEpsilon(t, 50.592050, posMid.Lat, 0.01)
		assert.InEpsilon(t, 11.194015, posMid.Lng, 0.01)
		assert.Equal(t, 275645, result.EstimateAccuracy(timeEst))
	})
}
