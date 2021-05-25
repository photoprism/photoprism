package face

// Points is a list of face landmark coordinates.
type Points []Point

// Markers returns relative marker positions for all face landmark coordinates.
func (pts Points) Markers(r Point, dim float32) (m Markers) {
	for _, p := range pts {
		m = append(m, p.Marker(r, dim))
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
func (p Point) Marker(r Point, dim float32) Marker {
	if dim < 1 {
		dim = 1
	}

	return NewMarker(
		p.Name,
		float32(p.Col-r.Col)/dim,
		float32(p.Row-r.Row)/dim,
		float32(p.Scale)/dim,
		float32(p.Scale)/dim,
	)
}
