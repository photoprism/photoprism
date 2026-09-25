package vision

// NSFWThresholdAuto selects the calibrated threshold of the active detector.
const NSFWThresholdAuto = -1

// Thresholds are expressed as percentages and gate label acceptance, topicality,
// and NSFW handling, with zero or -1 selecting automatic NSFW calibration.
type Thresholds struct {
	Confidence int  `yaml:"Confidence,omitempty" json:"confidence,omitempty"`
	Topicality int  `yaml:"Topicality,omitempty" json:"topicality,omitempty"`
	NSFW       int  `yaml:"NSFW,omitempty" json:"nsfw,omitempty"`
	NSFWUpload *int `yaml:"NSFWUpload,omitempty" json:"nsfwUpload,omitempty"`
	NSFWIndex  *int `yaml:"NSFWIndex,omitempty" json:"nsfwIndex,omitempty"`
	NSFWLabels *int `yaml:"NSFWLabels,omitempty" json:"nsfwLabels,omitempty"`
}

// GetConfidence returns the Confidence threshold in percent from 0 to 100.
func (t *Thresholds) GetConfidence() int {
	if t == nil || t.Confidence < 0 {
		return 0
	} else if t.Confidence > 100 {
		return 100
	}

	return t.Confidence
}

// GetConfidenceFloat32 returns the Confidence threshold as float32 for comparison.
func (t *Thresholds) GetConfidenceFloat32() float32 {
	return float32(t.GetConfidence()) / 100
}

// GetTopicality returns the Topicality threshold in percent from 0 to 100.
func (t *Thresholds) GetTopicality() int {
	if t == nil || t.Topicality < 0 {
		return 0
	} else if t.Topicality > 100 {
		return 100
	}

	return t.Topicality
}

// GetTopicalityFloat32 returns the Topicality threshold as float32 for comparison.
func (t *Thresholds) GetTopicalityFloat32() float32 {
	return float32(t.GetTopicality()) / 100
}

// GetNSFWUpload returns the upload-screening threshold in percent.
func (t *Thresholds) GetNSFWUpload() int {
	var override *int
	if t != nil {
		override = t.NSFWUpload
	}
	value, _ := t.nsfwValue(override)
	return value
}

// GetNSFWUploadFloat32 returns the upload-screening threshold as float32.
func (t *Thresholds) GetNSFWUploadFloat32() float32 {
	return float32(t.GetNSFWUpload()) / 100
}

// NSFWUploadIsSet reports whether upload screening has an explicit threshold.
func (t *Thresholds) NSFWUploadIsSet() bool {
	var override *int
	if t != nil {
		override = t.NSFWUpload
	}
	_, configured := t.nsfwValue(override)
	return configured
}

// GetNSFWIndex returns the indexing threshold in percent.
func (t *Thresholds) GetNSFWIndex() int {
	var override *int
	if t != nil {
		override = t.NSFWIndex
	}
	value, _ := t.nsfwValue(override)
	return value
}

// GetNSFWIndexFloat32 returns the indexing threshold as float32.
func (t *Thresholds) GetNSFWIndexFloat32() float32 {
	return float32(t.GetNSFWIndex()) / 100
}

// NSFWIndexIsSet reports whether indexing has an explicit detector threshold.
func (t *Thresholds) NSFWIndexIsSet() bool {
	var override *int
	if t != nil {
		override = t.NSFWIndex
	}
	_, configured := t.nsfwValue(override)
	return configured
}

// GetNSFWLabels returns the label-derived NSFW threshold in percent.
func (t *Thresholds) GetNSFWLabels() int {
	var override *int
	if t != nil {
		override = t.NSFWLabels
	}
	value, _ := t.nsfwValue(override)
	return value
}

// nsfwValue resolves a context override, the shared legacy value, or the fallback.
func (t *Thresholds) nsfwValue(override *int) (int, bool) {
	if override != nil {
		if *override <= 0 {
			return DefaultNSFWThreshold, false
		}

		return min(*override, 100), true
	}

	if t != nil && t.NSFW > 0 {
		return min(t.NSFW, 100), true
	}

	return DefaultNSFWThreshold, false
}
