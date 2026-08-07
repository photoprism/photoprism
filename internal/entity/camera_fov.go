package entity

import "strings"

// MeasuredFisheyeFov is the dewarp angle shared by every camera CameraFisheyeFov recognizes.
// It is the angle the stored fisheye disc spans, not a lens specification, since the disc is
// inscribed in its half of the frame; overshooting drops a wedge of the scene at each seam.
const MeasuredFisheyeFov = 190

// CameraFisheyeFov returns the field of view in degrees for a known 360° camera, or 0 so the
// caller falls back to the configured default. Recognized cameras must stay in step with
// MediaFile.FisheyeDng detection.
func CameraFisheyeFov(makeName, modelName string) int {
	maker := strings.ToLower(strings.TrimSpace(makeName))
	model := strings.ToLower(strings.TrimSpace(modelName))

	// Maker and model are matched separately because an original may carry only one of the two.
	// A model must identify the brand, so bare product names are skipped: "one x" would otherwise
	// match "iPhone X".
	switch {
	case maker == "insta360", maker == "arashi vision":
		return MeasuredFisheyeFov
	case strings.Contains(model, "insta360"): // Insta360 X-series, ONE, ONE X, ONE RS, ...
		return MeasuredFisheyeFov
	case strings.Contains(model, "theta"): // Ricoh Theta series (the model always carries "THETA").
		return MeasuredFisheyeFov
	}

	return 0
}

// CameraFisheyeRoll returns the verified spherical roll correction for a known camera profile.
func CameraFisheyeRoll(makeName, modelName string) int {
	maker := strings.ToLower(strings.TrimSpace(makeName))
	model := strings.ToLower(strings.TrimSpace(modelName))

	if model != "insta360 oners" {
		return 0
	}

	switch maker {
	case "", "insta360", "arashi vision":
		return 180
	default:
		return 0
	}
}
