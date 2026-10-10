package form

// PhotoLabel contains the editable properties of a photo label assignment.
type PhotoLabel struct {
	Uncertainty *int            `json:"Uncertainty"`
	Label       *PhotoLabelName `json:"Label"`
}

// PhotoLabelName contains an optional shared label name.
type PhotoLabelName struct {
	Name *string `json:"Name"`
}
