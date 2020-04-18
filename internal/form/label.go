package form

// Label represents a label edit form.
type Label struct {
	LabelName     string `json:"LabelName"`
	Uncertainty   int    `json:"Uncertainty"`
	LabelPriority int    `json:"LabelPriority"`
}
