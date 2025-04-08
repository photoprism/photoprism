package vision

// Thresholds are percentages, e.g. to determine the minimum confidence level
// a model must have for a prediction to be considered valid or "accepted".
type Thresholds struct {
	Confidence int `yaml:"Confidence" json:"confidence"`
}
