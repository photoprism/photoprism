package geo

import (
	"fmt"
	"math"
	"time"
)

const Meter = 0.00001

// Position represents a geo coordinate.
type Position struct {
	Name     string    // Optional name
	Time     time.Time // Optional time
	Lat      float64   // In degree
	Lng      float64   // In degree
	Altitude float64   // In meter
	Accuracy int       // In meter
	Estimate bool
}

// String returns the position information as string for logging.
func (p Position) String() string {
	name := p.Name

	if name == "" {
		name = "position"
	}

	return fmt.Sprintf("%s @ %f, %f, alt %f m ± %d m", name, p.Lat, p.Lng, p.Altitude, p.Accuracy)
}

// AltitudeInt returns the altitude as integer.
func (p Position) AltitudeInt() int {
	return int(math.Round(p.Altitude))
}

// Km calculates the distance to another position in km.
func (p Position) Km(other Position) float64 {
	return math.Abs(Km(p, other))
}

// InRange tests if coordinates are within a certain range of the position.
func (p *Position) InRange(lat, lng, r float64) bool {
	switch {
	case lat == 0 && lng == 0:
		return false
	case p.Lat == 0 && p.Lng == 0:
		return false
	case lat < p.Lat-r || lat > p.Lat+r:
		return false
	case lng < p.Lng-r || lng > p.Lng+r:
		return false
	}

	return true
}

// Randomize adds a random offset to the coordinates.
func (p *Position) Randomize(diameter float64) {
	if diameter <= 0 {
		// Nothing to do.
		return
	}

	// Randomize latitude and longitude.
	p.Lat = Randomize(p.Lat, diameter)
	p.Lng = Randomize(p.Lng, diameter)

	// Estimate change in accuracy.
	meter := int(math.Round(diameter / Meter))

	// Increase accuracy if needed.
	if p.Accuracy < meter {
		p.Accuracy = meter
	}
}
