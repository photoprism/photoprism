package form

import (
	"fmt"

	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// Marker represents an image marker edit form.
type Marker struct {
	FileUID       string  `json:"FileUID,omitempty"`
	MarkerType    string  `json:"Type,omitempty"`
	MarkerSrc     string  `json:"Src,omitempty"`
	X             float32 `json:"X"`
	Y             float32 `json:"Y"`
	W             float32 `json:"W,omitempty"`
	H             float32 `json:"H,omitempty"`
	SubjSrc       string  `json:"SubjSrc"`
	MarkerName    string  `json:"Name"`
	MarkerReview  bool    `json:"MarkerReview"`
	MarkerInvalid bool    `json:"Invalid"`
}

// NewMarker creates a new form initialized with model values.
func NewMarker(m interface{}) (f Marker, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}

// Validate returns an error if any form values are invalid.
func (frm *Marker) Validate() error {
	// Check type and src length.
	if len(frm.MarkerType) > 8 || len(frm.MarkerSrc) > 8 || len(frm.SubjSrc) > 8 {
		return fmt.Errorf("invalid type or src")
	}

	if len([]rune(frm.MarkerName)) > 160 {
		return fmt.Errorf("name is too long")
	}

	// Validate file UID.
	if frm.FileUID == "" {
		return fmt.Errorf("missing file uid")
	} else if rnd.InvalidUID(frm.FileUID, 'f') {
		return fmt.Errorf("invalid file uid")
	}

	// Check if the coordinates are within a valid range.
	if frm.X > 1 || frm.Y > 1 || frm.X < 0 || frm.Y < 0 || frm.W < 0 || frm.H < 0 || frm.W > 1 || frm.H > 1 {
		return fmt.Errorf("invalid area")
	}

	return nil
}
