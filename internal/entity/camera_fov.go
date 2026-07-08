package entity

import "strings"

// CameraFisheyeFov returns the per-lens field of view in degrees for dewarping fisheye 360°
// originals from known cameras, or 0 when unknown so the caller can fall back to the configured
// default. Values are approximate and tuned for the overlapping lenses of each model.
func CameraFisheyeFov(makeName, modelName string) int {
	model := strings.ToLower(strings.TrimSpace(modelName))

	switch {
	case model == "":
		return 0
	case strings.Contains(model, "insta360 x") || strings.Contains(model, "one x") || strings.Contains(model, "one rs"):
		return 204 // Insta360 X-series / ONE X / ONE RS.
	case strings.Contains(model, "insta360"):
		return 200 // Other Insta360 models (e.g. ONE).
	case strings.Contains(model, "theta"):
		return 200 // Ricoh Theta series.
	case strings.Contains(model, "gopro max"):
		return 200 // GoPro Max.
	}

	return 0
}
