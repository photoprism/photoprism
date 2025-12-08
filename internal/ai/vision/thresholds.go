package vision

// Thresholds are expressed as percentages (0-100) and gate label acceptance,
// topicality, and NSFW handling for the configured vision models.
type Thresholds struct {
	Confidence int `yaml:"Confidence,omitempty" json:"confidence,omitempty"`
	Topicality int `yaml:"Topicality,omitempty" json:"topicality,omitempty"`
	NSFW       int `yaml:"NSFW,omitempty" json:"nsfw,omitempty"`
}

// GetConfidence returns the Confidence threshold in percent from 0 to 100.
func (t *Thresholds) GetConfidence() int {
	if t.Confidence < 0 {
		return 0
	} else if t.Confidence > 100 {
		return 1
	}

	return t.Confidence
}

// GetConfidenceFloat32 returns the Confidence threshold as float32 for comparison.
func (t *Thresholds) GetConfidenceFloat32() float32 {
	return float32(t.GetConfidence()) / 100
}

// GetTopicality returns the Topicality threshold in percent from 0 to 100.
func (t *Thresholds) GetTopicality() int {
	if t.Topicality < 0 {
		return 0
	} else if t.Topicality > 100 {
		return 1
	}

	return t.Topicality
}

// GetTopicalityFloat32 returns the Topicality threshold as float32 for comparison.
func (t *Thresholds) GetTopicalityFloat32() float32 {
	return float32(t.GetTopicality()) / 100
}

// GetNSFW returns the NSFW threshold in percent from 0 to 100.
func (t *Thresholds) GetNSFW() int {
	if t.NSFW <= 0 {
		return DefaultThresholds.NSFW
	} else if t.NSFW > 100 {
		return 1
	}

	return t.NSFW
}

// GetNSFWFloat32 returns the NSFW threshold as float32 for comparison.
func (t *Thresholds) GetNSFWFloat32() float32 {
	return float32(t.GetNSFW()) / 100
}
