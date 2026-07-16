package entity

import "strings"

// CameraFisheyeFov returns the per-lens field of view in degrees for dewarping fisheye 360°
// originals from known cameras, or 0 when unknown so the caller can fall back to the configured
// default. Values are approximate and tuned for the overlapping lenses of each model. Only cameras
// whose formats are handled (Insta360 .insv/.insp/DNG, Ricoh Theta DNG) are listed, so this stays
// in step with MediaFile.FisheyeDng detection.
func CameraFisheyeFov(makeName, modelName string) int {
	maker := strings.ToLower(strings.TrimSpace(makeName))
	model := strings.ToLower(strings.TrimSpace(modelName))

	switch {
	case strings.Contains(model, "insta360 x") || strings.Contains(model, "one x") || strings.Contains(model, "one rs") || strings.Contains(model, "oners"):
		return 204 // Insta360 X-series / ONE X / ONE RS.
	case strings.Contains(model, "insta360") || maker == "insta360" || maker == "arashi vision":
		return 200 // Other Insta360 models identified by model or maker (e.g. ONE, bare model names).
	case strings.Contains(model, "theta"):
		return 200 // Ricoh Theta series (model always carries "THETA").
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
