package face

// Markers represents a list of relative marker positions.
type Markers []Marker

// Marker represents a relative marker position.
type Marker struct {
	Name string  `json:"name,omitempty"`
	X    float32 `json:"x,omitempty"`
	Y    float32 `json:"y,omitempty"`
	H    float32 `json:"h,omitempty"`
	W    float32 `json:"w,omitempty"`
}

// NewMarker returns new relative marker position.
func NewMarker(name string, x, y, h, w float32) Marker {
	return Marker{
		Name: name,
		X:    x,
		Y:    y,
		H:    h,
		W:    w,
	}
}
