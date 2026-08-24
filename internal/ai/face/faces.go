package face

// Faces represents a list of faces detected.
type Faces []Face

// Contains returns true if the face conflicts with existing faces.
func (faces Faces) Contains(other Face) bool {
	cropArea := other.CropArea()

	for _, f := range faces {
		if f.CropArea().OverlapPercent(cropArea) > OverlapThresholdFloor {
			return true
		}
	}

	return false
}

// Append adds a face.
func (faces *Faces) Append(f Face) {
	*faces = append(*faces, f)
}

// Count returns the number of faces detected.
func (faces Faces) Count() int {
	return len(faces)
}

// MaxScale returns the largest face scale in pixels.
func (faces Faces) MaxScale() (max int) {
	for _, f := range faces {
		if f.Area.Scale > max {
			max = f.Area.Scale
		}
	}

	return max
}

// Uncertainty returns the max face detection uncertainty in percent.
func (faces Faces) Uncertainty() int {
	if len(faces) < 1 {
		return 100
	}

	maxScore := 0

	for _, f := range faces {
		if f.Score > maxScore {
			maxScore = f.Score
		}
	}

	return ScoreUncertainty(maxScore)
}

// ScoreUncertainty converts a detection score into an uncertainty in percent.
//
// The steps are spread over the 0-100 confidence scale the ONNX detectors report, where the
// cascade detector that preceded them scored into the hundreds. Reading one scale's thresholds
// on the other leaves the confident end unreachable, so nothing ever reports as certain.
func ScoreUncertainty(score int) int {
	switch {
	case score > 95:
		return 1
	case score > 90:
		return 5
	case score > 85:
		return 10
	case score > 80:
		return 15
	case score > 75:
		return 20
	case score > 70:
		return 25
	case score > 65:
		return 30
	case score > 60:
		return 35
	case score > 55:
		return 40
	case score > 50:
		return 45
	}

	return 50
}
