package form

import "github.com/ulule/deepcopier"

// Marker represents an image marker edit form.
type Marker struct {
	Ref           string `json:"Ref"`
	RefSrc        string `json:"RefSrc"`
	MarkerSrc     string `json:"Src"`
	MarkerType    string `json:"Type"`
	MarkerScore   int    `json:"Score"`
	MarkerInvalid bool   `json:"Invalid"`
	MarkerLabel   string `json:"Label"`
}

func NewMarker(m interface{}) (f Marker, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
