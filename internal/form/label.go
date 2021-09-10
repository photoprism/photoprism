package form

import "github.com/ulule/deepcopier"

// Label represents a label edit form.
type Label struct {
	LabelName     string `json:"Name"`
	Uncertainty   int    `json:"Uncertainty"`
	LabelPriority int    `json:"Priority"`
	Thumb         string `json:"Thumb"`
	ThumbSrc      string `json:"ThumbSrc"`
}

func NewLabel(m interface{}) (f Label, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
