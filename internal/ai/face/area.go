package face

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/thumb/crop"
)

// Areas is a list of face landmark areas.
type Areas []Area

// Relative returns all areas as offsets from the reference point r, as a fraction of the image.
// The offsets keep their sign, since half the landmarks of a face sit left of or above the eyes
// midpoint that RelativeLandmarks measures them from.
func (pts Areas) Relative(r Area, rows, cols float32) crop.Areas {
	if len(pts) == 0 {
		return nil
	}

	if rows < 1 {
		rows = 1
	}

	if cols < 1 {
		cols = 1
	}

	invRows := 1 / rows
	invCols := 1 / cols

	result := make(crop.Areas, 0, len(pts))

	for _, p := range pts {
		result = append(result, crop.NewOffsetArea(
			p.Name,
			float32(p.Col-r.Col)*invCols,
			float32(p.Row-r.Row)*invRows,
			float32(p.Scale)*invCols,
			float32(p.Scale)*invRows,
		))
	}

	return result
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

// Relative returns the area as an offset from the reference point r, as a fraction of the image.
func (a Area) Relative(r Area, rows, cols float32) crop.Area {
	if rows < 1 {
		rows = 1
	}

	if cols < 1 {
		cols = 1
	}

	return crop.NewOffsetArea(
		a.Name,
		float32(a.Col-r.Col)/cols,
		float32(a.Row-r.Row)/rows,
		float32(a.Scale)/cols,
		float32(a.Scale)/rows,
	)
}

// TopLeft returns the top left position of the area.
func (a Area) TopLeft() (int, int) {
	return a.Row - (a.Scale / 2), a.Col - (a.Scale / 2)
}
