package form

import "github.com/ulule/deepcopier"

// Marker represents an image marker edit form.
type Marker struct {
	SubjSrc       string `json:"SubjSrc"`
	MarkerName    string `json:"Name"`
	MarkerReview  bool   `json:"MarkerReview"`
	MarkerInvalid bool   `json:"Invalid"`
}

func NewMarker(m interface{}) (f Marker, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
