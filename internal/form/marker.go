package form

import "github.com/ulule/deepcopier"

// Marker represents an image marker edit form.
type Marker struct {
	MarkerType    string `json:"Type"`
	MarkerSrc     string `json:"Src"`
	MarkerName    string `json:"Name"`
	SubjectUID    string `json:"SubjectUID"`
	SubjectSrc    string `json:"SubjectSrc"`
	FaceID        string `json:"FaceID"`
	Score         int    `json:"Score"`
	MarkerInvalid bool   `json:"Invalid"`
}

func NewMarker(m interface{}) (f Marker, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
