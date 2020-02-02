package form

// Label represents a label edit form.
type Label struct {
	LabelName        string `json:"LabelName"`
	LabelUncertainty int    `json:"LabelUncertainty"`
	LabelPriority    int    `json:"LabelPriority"`
}
