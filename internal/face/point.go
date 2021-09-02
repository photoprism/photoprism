package face

import (
	"fmt"
)

// Areas is a list of face landmark areas.
type Areas []Area

// Markers returns relative marker positions for all face landmark coordinates.
func (pts Areas) Markers(r Area, rows, cols float32) (m Markers) {
	for _, p := range pts {
		m = append(m, p.Marker(r, rows, cols))
	}

	return m
}

// Area represents a face landmark position.
type Area struct {
	Name  string `json:"name,omitempty"`
	Row   int    `json:"x,omitempty"`
	Col   int    `json:"y,omitempty"`
	Scale int    `json:"size,omitempty"`
}

// String returns the face landmark position as string.
func (a Area) String() string {
	return fmt.Sprintf("%d-%d-%d", a.Row, a.Col, a.Scale)
}

// NewArea returns new face landmark coordinates.
func NewArea(name string, row, col, scale int) Area {
	return Area{
		Name:  name,
		Row:   row,
		Col:   col,
		Scale: scale,
	}
}

// Marker returns a relative marker area for the face landmark coordinates.
func (a Area) Marker(r Area, rows, cols float32) Marker {
	if rows < 1 {
		rows = 1
	}

	if cols < 1 {
		cols = 1
	}

	return NewMarker(
		a.Name,
		float32(a.Col-r.Col)/cols,
		float32(a.Row-r.Row)/rows,
		float32(a.Scale)/rows,
		float32(a.Scale)/cols,
	)
}

// TopLeft returns the top left position of the face.
func (a Area) TopLeft() (int, int) {
	return a.Row - (a.Scale / 2), a.Col - (a.Scale / 2)
}
