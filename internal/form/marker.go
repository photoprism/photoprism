package form

import "github.com/ulule/deepcopier"

// Marker represents an image marker edit form.
type Marker struct {
	SubjectSrc    string `json:"SubjectSrc"`
	MarkerName    string `json:"Name"`
	MarkerInvalid bool   `json:"Invalid"`
}

func NewMarker(m interface{}) (f Marker, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
