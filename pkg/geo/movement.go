package geo

import (
	"fmt"
	"math"
	"time"
)

// Movement represents a position change in degrees per second.
type Movement struct {
	Start Position
	End   Position
}

// NewMovement returns the movement between two positions and points in time.
func NewMovement(pos1, pos2 Position) (m Movement) {
	// Make sure start and end are in the right order.
	if d := pos1.Time.Sub(pos2.Time); d > 0 {
		return Movement{Start: pos2, End: pos1}
	} else {
		return Movement{Start: pos1, End: pos2}
	}
}

// Duration calculates the movement duration.
func (m *Movement) Duration() time.Duration {
	return m.End.Time.Sub(m.Start.Time)
}

// Deg calculates the position change in degrees.
func (m *Movement) Deg() (lat, lng float64) {
	return m.DegLat(), m.DegLng()
}

// DegLng calculates the longitude change in degrees.
func (m *Movement) DegLng() float64 {
	return m.End.Lng - m.Start.Lng
}

// DegLat calculates the latitude change in degrees.
func (m *Movement) DegLat() float64 {
	return m.End.Lat - m.Start.Lat
}

// DegPerSecond returns the position change in degrees per second.
func (m *Movement) DegPerSecond() (latSec, lngSec float64) {
	s := m.Seconds()

	if s < 1 {
		return 0, 0
	}

	latSec = m.DegLat() / s
	lngSec = m.DegLng() / s

	return latSec, lngSec
}

// Km calculates the movement distance in km.
func (m *Movement) Km() float64 {
	return math.Abs(Km(m.Start, m.End))
}

// Meter calculates the movement distance in m.
func (m *Movement) Meter() float64 {
	return m.Km() * 1000
}

// Speed calculates the average movement speed in km/h.
func (m *Movement) Speed() float64 {
	km := m.Km()

	if km == 0 {
		return 0
	}

	h := m.Hours()

	if h == 0 {
		return 0
	}

	return km / h
}

// Midpoint returns the movement midpoint position.
func (m *Movement) Midpoint() Position {
	return Position{
		Name: "midpoint",
		Lat:  (m.Start.Lat + m.End.Lat) / 2,
		Lng:  (m.Start.Lng + m.End.Lng) / 2,
	}
}

// Closest returns the position closest in time, either start or end.
func (m *Movement) Closest(t time.Time) Position {
	delaStart := math.Abs(m.Start.Time.Sub(t).Seconds())
	deltaEnd := math.Abs(m.End.Time.Sub(t).Seconds())

	if delaStart > deltaEnd {
		return m.End
	} else {
		return m.Start
	}
}

// Seconds returns the movement duration in seconds.
func (m *Movement) Seconds() float64 {
	return math.Abs(m.Duration().Seconds())
}

// Hours returns the movement duration in hours.
func (m *Movement) Hours() float64 {
	return math.Abs(m.Duration().Hours())
}

// String returns the movement information as string for logging.
func (m *Movement) String() string {
	lat, lng := m.Deg()

	return fmt.Sprintf("movement from %s to %s in %f s, Δ lat %f, Δ lng  %f, dist %f km, speed %f km/h",
		m.Start.Time.Format("2006-01-02 15:04:05.999999999"),
		m.End.Time.Format("2006-01-02 15:04:05.999999999"),
		m.Seconds(), lat, lng, m.Km(), m.Speed())
}

// Realistic tests if the movement may have happened in the real world.
func (m *Movement) Realistic() bool {
	speed := m.Speed()

	switch {
	case speed > 900:
		return false
	case speed > 200 && m.Seconds() < 60:
		return false
	default:
		return true
	}
}

// AverageAltitude returns the average altitude.
func (m *Movement) AverageAltitude() float64 {
	if m.Start.Altitude != 0 && m.End.Altitude == 0 {
		return m.Start.Altitude
	} else if m.Start.Altitude == 0 && m.End.Altitude != 0 {
		return m.End.Altitude
	} else if m.Start.Altitude != 0 && m.End.Altitude != 0 {
		return (m.Start.Altitude + m.End.Altitude) / 2
	}

	return 0
}

// EstimateAccuracy returns the position estimate accuracy in meter.
func (m *Movement) EstimateAccuracy(t time.Time) int {
	var a float64

	if !m.Realistic() {
		a = m.Meter() / 2
	} else if t.Before(m.Start.Time) {
		d := m.Start.Time.Sub(t).Hours() * 1000
		d = math.Copysign(math.Sqrt(math.Abs(d)), d)
		a = m.Speed() * d
	} else if t.After(m.End.Time) {
		d := t.Sub(m.End.Time).Hours() * 1000
		d = math.Copysign(math.Sqrt(math.Abs(d)), d)
		a = m.Speed() * d
	} else {
		a = m.Meter() / 20
	}

	if meter := math.Round(math.Abs(a)); meter > 5 {
		return int(meter)
	}

	return 5
}

// EstimateAltitude estimates the altitude at a given time.
func (m *Movement) EstimateAltitude(t time.Time) float64 {
	if t.Before(m.Start.Time) {
		return m.Start.Altitude
	} else if t.After(m.End.Time) {
		return m.End.Altitude
	}

	return m.AverageAltitude()
}

// EstimateAltitudeInt returns the estimated altitude as integer.
func (m *Movement) EstimateAltitudeInt(t time.Time) int {
	return int(math.Round(m.EstimateAltitude(t)))
}

// EstimatePosition returns the estimated position at a given time.
func (m *Movement) EstimatePosition(t time.Time) Position {
	t = t.UTC()
	d := t.Sub(m.Start.Time)
	s := d.Seconds()

	estimate := Position{
		Name:     "estimate",
		Time:     t,
		Altitude: m.EstimateAltitude(t),
		Accuracy: m.EstimateAccuracy(t),
		Estimate: true,
	}

	if m.Realistic() {
		if t.Before(m.Start.Time) || t.After(m.End.Time) {
			s = math.Copysign(math.Sqrt(math.Abs(s)), s)
		}

		latSec, lngSec := m.DegPerSecond()

		estimate.Lat = m.Start.Lat + latSec*s
		estimate.Lng = m.Start.Lng + lngSec*s

		return estimate
	} else if km := m.Km(); km < 1 {
		p := m.Midpoint()

		estimate.Lat = p.Lat
		estimate.Lng = p.Lng

		return estimate
	} else {
		p := m.Closest(t)

		estimate.Lat = p.Lat
		estimate.Lng = p.Lng

		return estimate
	}
}
