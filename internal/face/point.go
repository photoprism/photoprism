package face

// Points is a list of face landmark coordinates.
type Points []Point

// Markers returns relative marker positions for all face landmark coordinates.
func (pts Points) Markers(r Point, rows, cols float32) (m Markers) {
	for _, p := range pts {
		m = append(m, p.Marker(r, rows, cols))
	}

	return m
}

// Point represents face landmark coordinates.
type Point struct {
	Name  string `json:"name,omitempty"`
	Row   int    `json:"x,omitempty"`
	Col   int    `json:"y,omitempty"`
	Scale int    `json:"size,omitempty"`
}

// NewPoint returns new face landmark coordinates.
func NewPoint(name string, row, col, scale int) Point {
	return Point{
		Name:  name,
		Row:   row,
		Col:   col,
		Scale: scale,
	}
}

// Marker returns a relative marker position for the face landmark coordinates.
func (p Point) Marker(r Point, rows, cols float32) Marker {
	if rows < 1 {
		rows = 1
	}

	if cols < 1 {
		cols = 1
	}

	return NewMarker(
		p.Name,
		float32(p.Col-r.Col)/cols,
		float32(p.Row-r.Row)/rows,
		float32(p.Scale)/rows,
		float32(p.Scale)/cols,
	)
}

// TopLeft returns the top left position of the face.
func (p Point) TopLeft() (int, int) {
	return p.Row - (p.Scale / 2), p.Col - (p.Scale / 2)
}
