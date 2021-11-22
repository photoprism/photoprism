package geo

import (
	"math"
	"time"
)

// Movement represents a position change in degrees per second.
type Movement struct {
	Start     Position
	StartTime time.Time
	End       Position
	EndTime   time.Time
	Duration  time.Duration
	LatDiff   float64
	LngDiff   float64
	SpeedKmh  float64
	DistKm    float64
}

// NewMovement returns the movement between two positions and points in time.
func NewMovement(pos1, pos2 Position, time1, time2 time.Time) (m Movement) {
	t1 := time1.UTC()
	t2 := time2.UTC()

	m = Movement{DistKm: Dist(pos1, pos2), Duration: t2.Sub(t1)}

	if m.Duration >= 0 {
		m.Start = pos1
		m.StartTime = time1
		m.End = pos2
		m.EndTime = time2
		m.LatDiff = pos2.Lat - pos1.Lat
		m.LngDiff = pos2.Lng - pos1.Lng
	} else {
		m.Start = pos2
		m.StartTime = time2
		m.End = pos1
		m.EndTime = time1
		m.LatDiff = pos1.Lat - pos2.Lat
		m.LngDiff = pos1.Lng - pos2.Lng
	}

	if m.DistKm > 0.001 && m.Seconds() > 1 {
		m.SpeedKmh = m.DistKm / m.Hours()
	}

	return m
}

// Midpoint returns the movement midpoint position.
func (m *Movement) Midpoint() Position {
	return Position{
		Lat: (m.Start.Lat + m.End.Lat) / 2,
		Lng: (m.Start.Lng + m.End.Lng) / 2,
	}
}

// Seconds returns the movement duration in seconds.
func (m *Movement) Seconds() float64 {
	return math.Abs(m.Duration.Seconds())
}

// Hours returns the movement duration in hours.
func (m *Movement) Hours() float64 {
	return math.Abs(m.Duration.Hours())
}

// DegPerSecond returns the position change in degrees per second.
func (m *Movement) DegPerSecond() (latSec, lngSec float64) {
	s := m.Seconds()

	if s < 1 {
		return 0, 0
	}

	return m.LatDiff / s, m.LngDiff / s
}

// Position returns the absolute position in degrees at a given time.
func (m *Movement) Position(t time.Time) Position {
	t = t.UTC()
	d := t.Sub(m.StartTime)
	s := d.Seconds()

	if m.Seconds() < 1 || math.Abs(s) < 1 {
		return m.Midpoint()
	}

	latSec, lngSec := m.DegPerSecond()

	return Position{
		Lat: m.Start.Lat + latSec*s,
		Lng: m.Start.Lng + lngSec*s,
	}
}
