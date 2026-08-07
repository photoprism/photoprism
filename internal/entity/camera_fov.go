package entity

import "strings"

// CameraFisheyeFov returns the per-lens field of view in degrees for dewarping fisheye 360°
// originals from known cameras, or 0 when unknown so the caller can fall back to the configured
// default. Only cameras whose formats are handled (Insta360 .insv/.insp/DNG, Ricoh Theta DNG) are
// listed, so this stays in step with MediaFile.FisheyeDng detection.
//
// A value is the angle the stored fisheye disc spans, not the lens specification: the disc is
// inscribed in its half of the frame, so it covers noticeably less than the optics do. Values are
// measured by rendering a sample at candidate angles and picking the one whose seam columns differ
// least from their neighbors; overshooting drops a wedge of the scene at each seam. Ricoh Theta Z1
// (190) and Insta360 ONE RS (186-194 across two captures) were measured, the remaining models
// follow them because no sample was available.
func CameraFisheyeFov(makeName, modelName string) int {
	maker := strings.ToLower(strings.TrimSpace(makeName))
	model := strings.ToLower(strings.TrimSpace(modelName))

	switch {
	case strings.Contains(model, "insta360 x") || strings.Contains(model, "one x") || strings.Contains(model, "one rs") || strings.Contains(model, "oners"):
		return 190 // Insta360 X-series / ONE X / ONE RS.
	case strings.Contains(model, "insta360") || maker == "insta360" || maker == "arashi vision":
		return 190 // Other Insta360 models identified by model or maker (e.g. ONE, bare model names).
	case strings.Contains(model, "theta"):
		return 190 // Ricoh Theta series (model always carries "THETA").
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
